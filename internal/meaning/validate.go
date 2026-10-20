package meaning

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/status"
)

// SchemaVersion is the current closed JSON Meaning schema version.
const (
	SchemaVersion     = 1
	MaxMeaningBytes   = 4 << 20
	MaxMeaningNodes   = 10_000
	MaxMeaningEdges   = 40_000
	MaxMeaningDepth   = 128
	MaxParticipants   = 64
	MaxDiscourseItems = 1_024
)

// Budget bounds the number of Meaning nodes, references and reference depth.
type Budget struct {
	MaxNodes int
	MaxEdges int
	MaxDepth int
}

// DefaultBudget returns the conservative per-Meaning graph limits.
func DefaultBudget() Budget {
	return Budget{MaxNodes: MaxMeaningNodes, MaxEdges: MaxMeaningEdges, MaxDepth: MaxMeaningDepth}
}

// Validate rejects zero, negative and above-ceiling graph budgets.
func (budget Budget) Validate() error {
	if budget.MaxNodes <= 0 || budget.MaxNodes > MaxMeaningNodes {
		return invalid("INVALID_MEANING_BUDGET", "budget.max_nodes", fmt.Sprintf("must be between 1 and %d", MaxMeaningNodes))
	}
	if budget.MaxEdges <= 0 || budget.MaxEdges > MaxMeaningEdges {
		return invalid("INVALID_MEANING_BUDGET", "budget.max_edges", fmt.Sprintf("must be between 1 and %d", MaxMeaningEdges))
	}
	if budget.MaxDepth <= 0 || budget.MaxDepth > MaxMeaningDepth {
		return invalid("INVALID_MEANING_BUDGET", "budget.max_depth", fmt.Sprintf("must be between 1 and %d", MaxMeaningDepth))
	}
	return nil
}

// StatusError preserves the typed outcome and bounded diagnostic of an operation.
type StatusError struct{ result status.Result }

func (failure *StatusError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "meaning graph operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed status result.
func (failure *StatusError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func invalid(code, path, message string) error {
	return meaningError(status.OutcomeInvalid, code, path, message)
}

func unknown(code, path, message string) error {
	return meaningError(status.OutcomeUnknown, code, path, message)
}

func resourceLimit(name string, value, maximum int) error {
	return meaningError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+name, fmt.Sprintf("meaning %s limit exceeded: %d > %d", name, value, maximum))
}

func meaningError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &StatusError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

// ValidateStructure checks the closed node union, references, cycles and
// resource limits without consulting any ontology or inferring facts.
func ValidateStructure(meaning Meaning, budget Budget) error {
	if err := budget.Validate(); err != nil {
		return err
	}
	if meaning.SchemaVersion != SchemaVersion {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported meaning schema version")
	}
	if len(meaning.Nodes) == 0 || len(meaning.Roots) == 0 {
		return invalid("EMPTY_MEANING", "nodes", "meaning requires at least one node and root")
	}
	if len(meaning.Nodes) > budget.MaxNodes {
		return resourceLimit("nodes", len(meaning.Nodes), budget.MaxNodes)
	}
	if len(meaning.Nodes) > MaxMeaningNodes {
		return resourceLimit("nodes", len(meaning.Nodes), MaxMeaningNodes)
	}
	if len(meaning.Roots) > budget.MaxNodes {
		return resourceLimit("roots", len(meaning.Roots), budget.MaxNodes)
	}
	nodes := make(map[NodeID]Node, len(meaning.Nodes))
	for index, node := range meaning.Nodes {
		path := fmt.Sprintf("nodes[%d]", index)
		if !validNodeID(node.ID) {
			return invalid("INVALID_NODE_ID", path+".id", "node ID must use the bounded ASCII identifier form")
		}
		if _, exists := nodes[node.ID]; exists {
			return invalid("DUPLICATE_NODE_ID", path+".id", "node ID is duplicated")
		}
		if !node.Kind.Valid() {
			return invalid("INVALID_NODE_KIND", path+".kind", "node kind is empty or unsupported")
		}
		if err := validateNodeShape(node, path); err != nil {
			return err
		}
		nodes[node.ID] = node
	}
	rootSet := make(map[NodeID]struct{}, len(meaning.Roots))
	for index, root := range meaning.Roots {
		if !validNodeID(root) {
			return invalid("INVALID_NODE_ID", fmt.Sprintf("roots[%d]", index), "root ID must use the bounded ASCII identifier form")
		}
		if _, exists := nodes[root]; !exists {
			return unknown("UNKNOWN_NODE", fmt.Sprintf("roots[%d]", index), "meaning root does not resolve")
		}
		if _, duplicate := rootSet[root]; duplicate {
			return invalid("DUPLICATE_ROOT", fmt.Sprintf("roots[%d]", index), "root ID is duplicated")
		}
		rootSet[root] = struct{}{}
	}
	adjacency := make(map[NodeID][]NodeID, len(nodes))
	edgeCount := 0
	for _, node := range meaning.Nodes {
		targets := nodeTargets(node)
		edgeCount += len(targets)
		if edgeCount > budget.MaxEdges || edgeCount > MaxMeaningEdges {
			limit := budget.MaxEdges
			if limit > MaxMeaningEdges {
				limit = MaxMeaningEdges
			}
			return resourceLimit("edges", edgeCount, limit)
		}
		for _, target := range targets {
			if _, exists := nodes[target]; !exists {
				return unknown("UNKNOWN_NODE", "nodes."+string(node.ID), "meaning reference does not resolve: "+string(target))
			}
			adjacency[node.ID] = append(adjacency[node.ID], target)
		}
	}
	reachable := make(map[NodeID]struct{}, len(nodes))
	queue := append([]NodeID(nil), meaning.Roots...)
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if _, visited := reachable[current]; visited {
			continue
		}
		reachable[current] = struct{}{}
		queue = append(queue, adjacency[current]...)
	}
	if len(reachable) != len(nodes) {
		return invalid("UNREACHABLE_NODE", "nodes", "every meaning node must be reachable from a declared root")
	}
	return validateAcyclic(adjacency, nodes, budget.MaxDepth)
}

// ValidateMeaning checks the structure and cross-validates concept references
// and participant role signatures against the immutable CORE snapshot.
func ValidateMeaning(meaning Meaning, ontology *core.ConceptGraph, budget Budget) error {
	if err := ValidateStructure(meaning, budget); err != nil {
		return err
	}
	if ontology == nil {
		return invalid("INVALID_CORE_GRAPH", "core", "ontology snapshot is nil")
	}
	nodes := make(map[NodeID]Node, len(meaning.Nodes))
	for _, node := range meaning.Nodes {
		nodes[node.ID] = node
	}
	for _, node := range meaning.Nodes {
		if node.Kind != KindEvent && node.Kind != KindState && node.Kind != KindReferent {
			continue
		}
		conceptID, err := ids.ParseConceptID(string(node.Concept))
		if err != nil {
			return invalid("INVALID_CONCEPT_ID", "nodes."+string(node.ID)+".concept", "concept ID is malformed")
		}
		concept, exists := ontology.Lookup(conceptID)
		if !exists {
			return unknown("UNKNOWN_CONCEPT", "nodes."+string(node.ID)+".concept", "meaning concept does not exist in CORE")
		}
		switch node.Kind {
		case KindEvent:
			if concept.Kind != core.KindEvent && concept.Kind != core.KindAbstract {
				return invalid("CONCEPT_KIND_MISMATCH", "nodes."+string(node.ID)+".concept", "event node requires an event or abstract concept")
			}
		case KindState:
			if concept.Kind != core.KindProperty && concept.Kind != core.KindAbstract {
				return invalid("CONCEPT_KIND_MISMATCH", "nodes."+string(node.ID)+".concept", "state node requires a property or abstract concept")
			}
		case KindReferent:
			if concept.Kind != core.KindEntity && concept.Kind != core.KindAbstract {
				return invalid("CONCEPT_KIND_MISMATCH", "nodes."+string(node.ID)+".concept", "referent node requires an entity or abstract concept")
			}
		}
		roles := make(map[string]core.RoleType, len(concept.RoleTypes))
		for _, role := range concept.RoleTypes {
			roles[role.Name] = role
		}
		assigned := make(map[string]struct{}, len(node.Participants))
		for _, participant := range node.Participants {
			role, exists := roles[string(participant.Role)]
			if !exists {
				return invalid("UNKNOWN_ROLE", "nodes."+string(node.ID)+".participants."+string(participant.Role), "role is not declared for this concept")
			}
			if _, duplicate := assigned[string(participant.Role)]; duplicate {
				return invalid("DUPLICATE_ROLE", "nodes."+string(node.ID)+".participants."+string(participant.Role), "role may be assigned only once")
			}
			assigned[string(participant.Role)] = struct{}{}
			target := nodes[participant.Target]
			if target.Kind != KindReferent && target.Kind != KindEvent && target.Kind != KindState {
				return invalid("INVALID_ROLE_TARGET", "nodes."+string(node.ID)+".participants."+string(participant.Role), "participant target must be a referent, event or state")
			}
			targetID, _ := ids.ParseConceptID(string(target.Concept))
			targetConcept, _ := ontology.Lookup(targetID)
			if len(role.AllowedKinds) > 0 && !containsConceptKind(role.AllowedKinds, targetConcept.Kind) {
				return invalid("ROLE_TARGET_KIND", "nodes."+string(node.ID)+".participants."+string(participant.Role), "participant concept kind is not allowed by CORE role signature")
			}
		}
	}
	return nil
}

func validateNodeShape(node Node, path string) error {
	if node.Concept != "" && (len(node.Concept) > ids.MaxSymbolBytes || !utf8.ValidString(string(node.Concept))) {
		return invalid("INVALID_CONCEPT_ID", path+".concept", "concept ID is invalid or too long")
	}
	allowed := nodeFieldsForKind(node.Kind)
	switch node.Kind {
	case KindEvent, KindState:
		if node.Concept == "" {
			return invalid("MISSING_CONCEPT", path+".concept", "event/state node requires a concept ID")
		}
		if len(node.Participants) > MaxParticipants {
			return resourceLimit("participants", len(node.Participants), MaxParticipants)
		}
		seenRoles := make(map[string]struct{}, len(node.Participants))
		for index, participant := range node.Participants {
			if !validName(string(participant.Role)) || !validNodeID(participant.Target) {
				return invalid("INVALID_PARTICIPANT", fmt.Sprintf("%s.participants[%d]", path, index), "participant role or target is invalid")
			}
			if _, exists := seenRoles[string(participant.Role)]; exists {
				return invalid("DUPLICATE_ROLE", fmt.Sprintf("%s.participants[%d].role", path, index), "role may be assigned only once")
			}
			seenRoles[string(participant.Role)] = struct{}{}
		}
	case KindReferent:
		if node.Concept == "" {
			return invalid("MISSING_CONCEPT", path+".concept", "referent node requires a concept ID")
		}
	case KindQuestion:
		if !validNodeID(node.Target) {
			return invalid("INVALID_SCOPE_TARGET", path+".target", "question requires one valid scoped target")
		}
	case KindScope:
		if !node.Operator.Valid() || !validNodeID(node.Target) {
			return invalid("INVALID_SCOPE", path, "scope requires a supported operator and one valid target")
		}
	case KindModality:
		if !node.Modality.Valid() || !validNodeID(node.Target) {
			return invalid("INVALID_MODALITY", path, "modality requires a supported value and one valid target")
		}
	case KindQuantifier:
		if !validNodeID(node.Target) || node.Quantifier == nil {
			return invalid("INVALID_QUANTIFIER", path, "quantifier requires one valid target and typed quantity")
		}
		if err := node.Quantifier.Validate(); err != nil {
			return err
		}
	case KindDiscourse:
		if !node.Act.Valid() || len(node.Content) == 0 {
			return invalid("INVALID_DISCOURSE", path, "discourse node requires a supported act and content")
		}
		if len(node.Content) > MaxDiscourseItems {
			return resourceLimit("discourse_items", len(node.Content), MaxDiscourseItems)
		}
		for index, target := range node.Content {
			if !validNodeID(target) {
				return invalid("INVALID_DISCOURSE_TARGET", fmt.Sprintf("%s.content[%d]", path, index), "discourse content target is invalid")
			}
		}
	}
	if node.Concept != "" {
		if _, err := ids.ParseConceptID(string(node.Concept)); err != nil {
			return invalid("INVALID_CONCEPT_ID", path+".concept", "concept ID is malformed")
		}
	}
	if node.Target != "" {
		if _, allowedField := allowed["target"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".target", "field does not belong to this node kind")
		}
	}
	if node.Operator != "" {
		if _, allowedField := allowed["operator"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".operator", "field does not belong to this node kind")
		}
	}
	if node.Modality != "" {
		if _, allowedField := allowed["modality"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".modality", "field does not belong to this node kind")
		}
	}
	if node.Quantifier != nil {
		if _, allowedField := allowed["quantifier"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".quantifier", "field does not belong to this node kind")
		}
	}
	if node.Act != "" {
		if _, allowedField := allowed["act"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".act", "field does not belong to this node kind")
		}
	}
	if node.Participants != nil {
		if _, allowedField := allowed["participants"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".participants", "field does not belong to this node kind")
		}
	}
	if node.Content != nil {
		if _, allowedField := allowed["content"]; !allowedField {
			return invalid("UNEXPECTED_NODE_FIELD", path+".content", "field does not belong to this node kind")
		}
	}
	return nil
}

func nodeTargets(node Node) []NodeID {
	switch node.Kind {
	case KindEvent, KindState:
		result := make([]NodeID, len(node.Participants))
		for index, participant := range node.Participants {
			result[index] = participant.Target
		}
		return result
	case KindQuestion, KindScope, KindModality, KindQuantifier:
		return []NodeID{node.Target}
	case KindDiscourse:
		return append([]NodeID(nil), node.Content...)
	default:
		return nil
	}
}

func validateAcyclic(adjacency map[NodeID][]NodeID, nodes map[NodeID]Node, maxDepth int) error {
	indegree := make(map[NodeID]int, len(nodes))
	depth := make(map[NodeID]int, len(nodes))
	for id := range nodes {
		indegree[id] = 0
		depth[id] = 1
	}
	for _, targets := range adjacency {
		for _, target := range targets {
			indegree[target]++
		}
	}
	queue := make([]NodeID, 0, len(nodes))
	for id := range nodes {
		if indegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	processed := 0
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		processed++
		for _, target := range adjacency[current] {
			candidateDepth := depth[current] + 1
			if candidateDepth > depth[target] {
				depth[target] = candidateDepth
				if candidateDepth > maxDepth {
					return resourceLimit("depth", candidateDepth, maxDepth)
				}
			}
			indegree[target]--
			if indegree[target] == 0 {
				queue = append(queue, target)
			}
		}
	}
	if processed != len(nodes) {
		return invalid("MEANING_CYCLE", "nodes", "meaning graph references must be acyclic")
	}
	return nil
}

func validNodeID(id NodeID) bool {
	value := string(id)
	if len(value) == 0 || len(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

func validName(value string) bool {
	if value == "" || len(value) > 64 || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '_' || character == '-') {
			return false
		}
	}
	return true
}

func containsConceptKind(kinds []core.ConceptKind, target core.ConceptKind) bool {
	if len(kinds) == 0 {
		return true
	}
	for _, kind := range kinds {
		if kind == target {
			return true
		}
	}
	return false
}
