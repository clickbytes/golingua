package construction

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxOperatorsPerPlan     = 32
	MaxOperatorDependencies = 32
	MaxOperatorCapabilities = 64
)

// OperatorKind is a closed set of semantic operations supported by this API.
type OperatorKind string

const (
	OperatorTense    OperatorKind = "tense"
	OperatorNegation OperatorKind = "negation"
	OperatorQuestion OperatorKind = "question"
	OperatorVoice    OperatorKind = "voice"
	OperatorModality OperatorKind = "modality"
)

// OperatorValue is an explicit, bounded semantic or grammatical value.
type OperatorValue string

const (
	ValuePast      OperatorValue = "past"
	ValuePresent   OperatorValue = "present"
	ValueFuture    OperatorValue = "future"
	ValueNegated   OperatorValue = "negated"
	ValueYesNo     OperatorValue = "yes_no"
	ValueWH        OperatorValue = "wh"
	ValueActive    OperatorValue = "active"
	ValuePassive   OperatorValue = "passive"
	ValuePossible  OperatorValue = "possible"
	ValueNecessary OperatorValue = "necessary"
	ValueLikely    OperatorValue = "likely"
)

// SemanticScope binds an operation to one explicit target node and, when the
// Meaning model represents it, a node that carries the scope/operator itself.
type SemanticScope struct {
	TargetNodeID meaning.NodeID `json:"target_node_id"`
	ScopeNodeID  meaning.NodeID `json:"scope_node_id,omitempty"`
}

// Operator is a declarative, licensed family transformation. It stores no
// executable script; compose preserves the caller's exact operator order.
type Operator struct {
	ID             OperatorID           `json:"id"`
	Kind           OperatorKind         `json:"kind"`
	Value          OperatorValue        `json:"value"`
	InputFamilyID  ids.ConstructionID   `json:"input_family_id"`
	OutputFamilyID ids.ConstructionID   `json:"output_family_id"`
	SemanticScope  SemanticScope        `json:"semantic_scope"`
	Dependencies   []OperatorID         `json:"dependencies,omitempty"`
	Capabilities   []Capability         `json:"capabilities,omitempty"`
	Provenance     provenance.SourceRef `json:"provenance"`
}

func (operator Operator) Validate() error {
	if !validToken(string(operator.ID)) {
		return invalid("INVALID_OPERATOR_ID", "operator.id", "operator ID must be a bounded lowercase token")
	}
	if !validFamilyID(operator.InputFamilyID) || !validFamilyID(operator.OutputFamilyID) {
		return invalid("INVALID_OPERATOR_FAMILY_ID", "operator", "operator input and output require canonical family IDs")
	}
	if len(operator.Dependencies) > MaxOperatorDependencies {
		return resourceLimit("operator_dependencies", len(operator.Dependencies), MaxOperatorDependencies)
	}
	if len(operator.Capabilities) > MaxOperatorCapabilities {
		return resourceLimit("operator_capabilities", len(operator.Capabilities), MaxOperatorCapabilities)
	}
	if !validOperatorValue(operator.Kind, operator.Value) {
		return invalid("INVALID_OPERATOR_VALUE", "operator.value", "operator kind and explicit value are not a supported pair")
	}
	if !containsCapability(operator.Capabilities, requiredCapability(operator)) {
		return invalid("MISSING_OPERATOR_CAPABILITY", "operator.capabilities", "operator must declare the capability required by its kind and value")
	}
	if len(operator.SemanticScope.TargetNodeID) == 0 {
		return incomplete("MISSING_OPERATOR_SCOPE", "operator.semantic_scope.target_node_id", "operator requires an explicit semantic target")
	}
	seenDependencies := make(map[OperatorID]struct{}, len(operator.Dependencies))
	for index, dependency := range operator.Dependencies {
		if !validToken(string(dependency)) {
			return invalid("INVALID_OPERATOR_DEPENDENCY", fmt.Sprintf("operator.dependencies[%d]", index), "dependency ID must be a bounded lowercase token")
		}
		if _, exists := seenDependencies[dependency]; exists {
			return invalid("DUPLICATE_OPERATOR_DEPENDENCY", "operator.dependencies", "dependency ID is duplicated")
		}
		seenDependencies[dependency] = struct{}{}
	}
	seenCapabilities := make(map[Capability]struct{}, len(operator.Capabilities))
	for index, capability := range operator.Capabilities {
		if !validToken(string(capability)) {
			return invalid("INVALID_OPERATOR_CAPABILITY", fmt.Sprintf("operator.capabilities[%d]", index), "capability must be a bounded lowercase token")
		}
		if _, exists := seenCapabilities[capability]; exists {
			return invalid("DUPLICATE_OPERATOR_CAPABILITY", "operator.capabilities", "capability is duplicated")
		}
		seenCapabilities[capability] = struct{}{}
	}
	if err := operator.Provenance.Validate(); err != nil {
		return invalid("INVALID_OPERATOR_PROVENANCE", "operator.provenance", "operator requires valid provenance")
	}
	return nil
}

func validOperatorValue(kind OperatorKind, value OperatorValue) bool {
	switch kind {
	case OperatorTense:
		return value == ValuePast || value == ValuePresent || value == ValueFuture
	case OperatorNegation:
		return value == ValueNegated
	case OperatorQuestion:
		return value == ValueYesNo || value == ValueWH
	case OperatorVoice:
		return value == ValueActive || value == ValuePassive
	case OperatorModality:
		return value == ValuePossible || value == ValueNecessary || value == ValueLikely
	default:
		return false
	}
}

func validateOperatorScope(operator Operator, graph meaning.Meaning) error {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	target, exists := nodes[operator.SemanticScope.TargetNodeID]
	if !exists {
		return unknown("UNKNOWN_OPERATOR_TARGET", "operator.semantic_scope.target_node_id", "operator target node does not exist in Meaning")
	}
	if target.Kind != meaning.KindEvent && target.Kind != meaning.KindState {
		return unsupported("OPERATOR_TARGET_UNSUPPORTED", "operator.semantic_scope.target_node_id", "operator target must be an event or state")
	}
	if (operator.Kind == OperatorTense || operator.Kind == OperatorVoice) && target.Kind != meaning.KindEvent {
		return unsupported("OPERATOR_TARGET_UNSUPPORTED", "operator.semantic_scope.target_node_id", "tense and voice require an event target")
	}
	scopeID := operator.SemanticScope.ScopeNodeID
	switch operator.Kind {
	case OperatorNegation:
		scope, ok := nodes[scopeID]
		if !ok {
			return incomplete("MISSING_NEGATION_SCOPE", "operator.semantic_scope.scope_node_id", "negation requires an explicit Meaning scope node")
		}
		if scope.Kind != meaning.KindScope || scope.Operator != meaning.OperatorNegation || scope.Target != target.ID {
			return unsupported("NEGATION_SCOPE_MISMATCH", "operator.semantic_scope.scope_node_id", "Meaning scope must explicitly negate the selected target")
		}
	case OperatorQuestion:
		scope, ok := nodes[scopeID]
		if !ok {
			return incomplete("MISSING_QUESTION_SCOPE", "operator.semantic_scope.scope_node_id", "question requires an explicit Meaning question or scope node")
		}
		validQuestion := scope.Kind == meaning.KindQuestion && scope.Target == target.ID
		validQuestion = validQuestion || (scope.Kind == meaning.KindScope && scope.Operator == meaning.OperatorQuestion && scope.Target == target.ID)
		if !validQuestion {
			return unsupported("QUESTION_SCOPE_MISMATCH", "operator.semantic_scope.scope_node_id", "Meaning question scope must target the selected event or state")
		}
	case OperatorModality:
		scope, ok := nodes[scopeID]
		if !ok {
			return incomplete("MISSING_MODALITY_SCOPE", "operator.semantic_scope.scope_node_id", "modality requires an explicit Meaning modality node")
		}
		if scope.Kind != meaning.KindModality || scope.Target != target.ID || string(scope.Modality) != string(operator.Value) {
			return unsupported("MODALITY_SCOPE_MISMATCH", "operator.semantic_scope.scope_node_id", "Meaning modality must explicitly match the operator value and target")
		}
	case OperatorTense, OperatorVoice:
		if scopeID != "" {
			return invalid("UNEXPECTED_OPERATOR_SCOPE_NODE", "operator.semantic_scope.scope_node_id", "tense and voice use their explicit value on the selected event")
		}
	}
	return nil
}

func cloneOperator(operator Operator) Operator {
	result := operator
	result.Dependencies = append([]OperatorID(nil), operator.Dependencies...)
	result.Capabilities = append([]Capability(nil), operator.Capabilities...)
	return result
}

func requiredCapability(operator Operator) Capability {
	return Capability(string(operator.Kind) + "-" + string(operator.Value))
}

func containsCapability(capabilities []Capability, required Capability) bool {
	for _, capability := range capabilities {
		if capability == required {
			return true
		}
	}
	return false
}

func operatorCapabilityList(operator Operator) []Capability {
	result := append([]Capability(nil), operator.Capabilities...)
	sortCapabilities(result)
	return result
}

func sortCapabilities(values []Capability) {
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
}

func incomplete(code, path, message string) error {
	return constructionError(status.OutcomeIncomplete, code, path, message)
}
