package meaning

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/status"
)

// StructuralEquivalence compares canonical graph structure, including local
// node IDs. It does not infer entity identity from concepts.
func StructuralEquivalence(left, right Meaning, budget Budget) (bool, error) {
	return CanonicalEqual(left, right, budget)
}

// SemanticEquivalence compares graph meaning while ignoring local node IDs
// and input node order. Referents require explicit resolution in the same
// request context; missing or conflicting identity evidence returns unknown
// or ambiguous rather than a guessed comparison.
func SemanticEquivalence(left, right Meaning, leftContext, rightContext RequestContext, budget Budget) (bool, error) {
	if err := ValidateStructure(left, budget); err != nil {
		return false, err
	}
	if err := ValidateStructure(right, budget); err != nil {
		return false, err
	}
	if err := ValidateReferences(left, leftContext, budget); err != nil {
		return false, err
	}
	if err := ValidateReferences(right, rightContext, budget); err != nil {
		return false, err
	}
	if leftContext.ID != rightContext.ID {
		return false, unknown("CONTEXT_MISMATCH", "context.id", "semantic identity cannot be compared across request contexts")
	}
	leftSignatures, err := semanticRoots(left, leftContext, budget)
	if err != nil {
		return false, err
	}
	rightSignatures, err := semanticRoots(right, rightContext, budget)
	if err != nil {
		return false, err
	}
	if len(leftSignatures) != len(rightSignatures) {
		return false, nil
	}
	for index := range leftSignatures {
		if leftSignatures[index] != rightSignatures[index] {
			return false, nil
		}
	}
	return true, nil
}

// ValidateReferences checks a request context against one Meaning graph,
// including mention targets and referent class consistency.
func ValidateReferences(meaning Meaning, context RequestContext, budget Budget) error {
	if err := ValidateStructure(meaning, budget); err != nil {
		return err
	}
	if err := context.Validate(); err != nil {
		return err
	}
	nodes := make(map[NodeID]Node, len(meaning.Nodes))
	for _, node := range meaning.Nodes {
		nodes[node.ID] = node
	}
	referents := make(map[ReferentID]Referent, len(context.Referents))
	for _, referent := range context.Referents {
		referents[referent.ID] = referent
	}
	for _, mention := range context.Mentions {
		node, exists := nodes[mention.NodeID]
		if !exists {
			return unknown("UNKNOWN_NODE", "context.mentions."+string(mention.ID)+".node_id", "mention node does not resolve in this meaning graph")
		}
		if node.Kind != KindReferent {
			return invalid("INVALID_MENTION_TARGET", "context.mentions."+string(mention.ID)+".node_id", "mention must identify a referent node")
		}
		if mention.ReferentID != "" && string(node.Concept) != string(referents[mention.ReferentID].Concept) {
			return invalid("REFERENT_CONCEPT_MISMATCH", "context.mentions."+string(mention.ID)+".referent_id", "referent class differs from the linked Meaning node")
		}
	}
	return nil
}

// SurfaceEquivalence compares two exact surface fingerprints produced by
// SurfaceFingerprint. It does not compare meanings or normalize text.
func SurfaceEquivalence(leftFingerprint, rightFingerprint string) (bool, error) {
	if !validFingerprint(leftFingerprint) || !validFingerprint(rightFingerprint) {
		return false, invalid("INVALID_SURFACE_FINGERPRINT", "surface", "surface fingerprint must be a lowercase SHA-256 hex digest")
	}
	return leftFingerprint == rightFingerprint, nil
}

func validFingerprint(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func semanticRoots(meaning Meaning, context RequestContext, budget Budget) ([]string, error) {
	nodes := make(map[NodeID]Node, len(meaning.Nodes))
	for _, node := range meaning.Nodes {
		nodes[node.ID] = node
	}
	bindings := make(map[NodeID]map[ReferentID]struct{})
	for _, mention := range context.Mentions {
		if mention.ReferentID == "" {
			continue
		}
		set := bindings[mention.NodeID]
		if set == nil {
			set = make(map[ReferentID]struct{})
			bindings[mention.NodeID] = set
		}
		set[mention.ReferentID] = struct{}{}
	}
	if len(bindings) > budget.MaxNodes {
		return nil, resourceLimit("references", len(bindings), budget.MaxNodes)
	}
	memo := make(map[NodeID]string, len(nodes))
	active := make(map[NodeID]bool, len(nodes))
	work := 0
	var signature func(NodeID, int) (string, error)
	signature = func(id NodeID, depth int) (string, error) {
		work++
		if work > budget.MaxEdges+budget.MaxNodes {
			return "", resourceLimit("semantic_comparisons", work, budget.MaxEdges+budget.MaxNodes)
		}
		if depth > budget.MaxDepth {
			return "", resourceLimit("depth", depth, budget.MaxDepth)
		}
		if value, exists := memo[id]; exists {
			return value, nil
		}
		if active[id] {
			return "", invalid("CYCLIC_MEANING", "nodes."+string(id), "semantic comparison requires an acyclic meaning graph")
		}
		active[id] = true
		node := nodes[id]
		children := nodeTargets(node)
		childSignatures := make([]string, len(children))
		for index, child := range children {
			value, err := signature(child, depth+1)
			if err != nil {
				return "", err
			}
			childSignatures[index] = value
		}
		value := fmt.Sprintf("kind=%s;concept=%s;act=%s;operator=%s;modality=%s;temporal_frame=%s;relation_predicate=%t;multiplicity=%s;", node.Kind, node.Concept, node.Act, node.Operator, node.Modality, node.TemporalFrame, node.RelationPredicate, node.Multiplicity)
		switch node.Kind {
		case KindReferent:
			set := bindings[id]
			if len(set) == 0 {
				return "", unknown("UNRESOLVED_ANAPHORA", "nodes."+string(id), "referent has no explicit identity binding in this request context")
			}
			if len(set) > 1 {
				return "", meaningError(status.OutcomeAmbiguous, "AMBIGUOUS_REFERENT", "nodes."+string(id), "referent has conflicting mention identity bindings")
			}
			for referent := range set {
				value += "identity=" + string(referent) + ";"
			}
		case KindQuantifier:
			value += fmt.Sprintf("quantifier=%s;", node.Quantifier.Kind)
			if node.Quantifier.Count != nil {
				value += fmt.Sprintf("count=%d;", *node.Quantifier.Count)
			}
		case KindEvent, KindState:
			participants := append([]Participant(nil), node.Participants...)
			sort.Slice(participants, func(i, j int) bool { return participants[i].Role < participants[j].Role })
			for _, participant := range participants {
				childIndex := 0
				for _, candidate := range nodeTargets(node) {
					if candidate == participant.Target {
						break
					}
					childIndex++
				}
				value += fmt.Sprintf("role=%s:%s;", participant.Role, childSignatures[childIndex])
			}
		case KindDiscourse:
			for _, child := range childSignatures {
				value += "content=" + child + ";"
			}
		default:
			for _, child := range childSignatures {
				value += "target=" + child + ";"
			}
		}
		delete(active, id)
		hash := sha256.Sum256([]byte(value))
		memo[id] = hex.EncodeToString(hash[:])
		return memo[id], nil
	}
	roots := append([]NodeID(nil), meaning.Roots...)
	rootSignatures := make([]string, len(roots))
	for index, root := range roots {
		value, err := signature(root, 1)
		if err != nil {
			return nil, err
		}
		rootSignatures[index] = value
	}
	sort.Strings(rootSignatures)
	return rootSignatures, nil
}
