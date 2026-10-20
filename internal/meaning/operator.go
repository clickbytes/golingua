package meaning

import "fmt"

// OperatorKind is the semantic classification of one explicit wrapper node.
type OperatorKind string

const (
	OperatorKindNegation     OperatorKind = "negation"
	OperatorKindQuestion     OperatorKind = "question"
	OperatorKindHypothetical OperatorKind = "hypothetical"
	OperatorKindPossibility  OperatorKind = "possibility"
	OperatorKindNecessity    OperatorKind = "necessity"
	OperatorKindLikelihood   OperatorKind = "likelihood"
	OperatorKindQuantifier   OperatorKind = "quantifier"
)

// OperatorNode is a typed view over the operator wrapper already stored in a
// Meaning graph. ID identifies its scope and Operand identifies its target.
type OperatorNode struct {
	ID         NodeID
	Kind       OperatorKind
	Operand    NodeID
	Quantifier *Quantifier
}

// QuantifierKind expresses semantic quantity independently of grammatical
// number or surface form.
type QuantifierKind string

const (
	QuantifierSome    QuantifierKind = "some"
	QuantifierAll     QuantifierKind = "all"
	QuantifierExactly QuantifierKind = "exactly"
	QuantifierAtLeast QuantifierKind = "at_least"
	QuantifierAtMost  QuantifierKind = "at_most"
	// MaxQuantifierCount bounds cardinality values in the request-scoped IR.
	MaxQuantifierCount = 1_000_000
)

// Quantifier is an explicitly scoped semantic cardinality operator.
type Quantifier struct {
	Kind  QuantifierKind `json:"kind"`
	Count *int           `json:"count,omitempty"`
}

// Validate rejects unknown quantifier types, unbounded counts and invalid
// count/operator combinations.
func (quantifier Quantifier) Validate() error {
	switch quantifier.Kind {
	case QuantifierSome, QuantifierAll:
		if quantifier.Count != nil {
			return invalid("INVALID_QUANTIFIER", "quantifier.count", "some/all quantifiers do not carry a numeric count")
		}
	case QuantifierExactly, QuantifierAtMost:
		if quantifier.Count == nil {
			return invalid("INVALID_QUANTIFIER", "quantifier.count", "numeric quantifier requires an explicit count")
		}
		if *quantifier.Count < 0 || *quantifier.Count > MaxQuantifierCount {
			return invalid("INVALID_QUANTIFIER", "quantifier.count", fmt.Sprintf("must be between 0 and %d", MaxQuantifierCount))
		}
	case QuantifierAtLeast:
		if quantifier.Count == nil {
			return invalid("INVALID_QUANTIFIER", "quantifier.count", "numeric quantifier requires an explicit count")
		}
		if *quantifier.Count <= 0 || *quantifier.Count > MaxQuantifierCount {
			return invalid("INVALID_QUANTIFIER", "quantifier.count", fmt.Sprintf("must be between 1 and %d", MaxQuantifierCount))
		}
	default:
		return invalid("INVALID_QUANTIFIER", "quantifier.kind", "quantifier kind is empty or unsupported")
	}
	return nil
}

// NewQuantifier creates an explicit quantifier scope over one Meaning node.
func NewQuantifier(id, target NodeID, quantifier Quantifier) (Node, error) {
	if !validNodeID(id) || !validNodeID(target) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node and target IDs must use the bounded ASCII identifier form")
	}
	if err := quantifier.Validate(); err != nil {
		return Node{}, err
	}
	copy := cloneQuantifier(quantifier)
	return Node{ID: id, Kind: KindQuantifier, Target: target, Quantifier: &copy}, nil
}

func cloneQuantifier(source Quantifier) Quantifier {
	result := source
	if source.Count != nil {
		count := *source.Count
		result.Count = &count
	}
	return result
}

// OperatorForNode returns the typed operator representation for a Meaning
// wrapper node. Non-operator nodes return ok=false.
func OperatorForNode(node Node) (OperatorNode, bool) {
	operator := OperatorNode{ID: node.ID, Operand: node.Target}
	switch node.Kind {
	case KindQuestion:
		operator.Kind = OperatorKindQuestion
	case KindScope:
		switch node.Operator {
		case OperatorNegation:
			operator.Kind = OperatorKindNegation
		case OperatorQuestion:
			operator.Kind = OperatorKindQuestion
		case OperatorHypothetical:
			operator.Kind = OperatorKindHypothetical
		default:
			return OperatorNode{}, false
		}
	case KindModality:
		switch node.Modality {
		case ModalityPossible:
			operator.Kind = OperatorKindPossibility
		case ModalityNecessary:
			operator.Kind = OperatorKindNecessity
		case ModalityLikely:
			operator.Kind = OperatorKindLikelihood
		default:
			return OperatorNode{}, false
		}
	case KindQuantifier:
		if node.Quantifier == nil {
			return OperatorNode{}, false
		}
		operator.Kind = OperatorKindQuantifier
		quantifier := cloneQuantifier(*node.Quantifier)
		operator.Quantifier = &quantifier
	default:
		return OperatorNode{}, false
	}
	return operator, true
}

func isOperatorNode(node Node) bool {
	_, ok := OperatorForNode(node)
	return ok
}
