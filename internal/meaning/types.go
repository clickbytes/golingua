// Package meaning defines language-neutral, request-scoped meaning graphs.
// It stores interpretations of utterances, not ontology definitions or facts.
package meaning

import "github.com/clickbytes/golingua/internal/ids"

// NodeID identifies one node within a Meaning graph. It is not a persistent
// CORE ConceptID and is never a dense runtime ID.
type NodeID string

// ConceptRef carries the stable external CORE ID form in a Meaning node.
type ConceptRef string

func ConceptRefFromID(id ids.ConceptID) ConceptRef { return ConceptRef(id.String()) }

// Role is a language-neutral semantic role declared by a CORE concept.
type Role string

// NodeKind selects one variant of the Meaning node tagged union.
type NodeKind string

const (
	KindEvent      NodeKind = "event"
	KindState      NodeKind = "state"
	KindReferent   NodeKind = "referent"
	KindQuestion   NodeKind = "question"
	KindScope      NodeKind = "scope"
	KindModality   NodeKind = "modality"
	KindQuantifier NodeKind = "quantifier"
	KindDiscourse  NodeKind = "discourse"
)

// Valid reports whether kind is a supported Meaning node variant.
func (kind NodeKind) Valid() bool {
	switch kind {
	case KindEvent, KindState, KindReferent, KindQuestion, KindScope, KindModality, KindQuantifier, KindDiscourse:
		return true
	default:
		return false
	}
}

func nodeFieldsForKind(kind NodeKind) map[string]bool {
	fields := map[string]bool{"id": true, "kind": true}
	var specific []string
	switch kind {
	case KindEvent, KindState:
		specific = []string{"concept", "participants", "temporal_frame"}
		if kind == KindState {
			specific = append(specific, "relation_predicate")
		}
	case KindReferent:
		specific = []string{"concept", "multiplicity"}
	case KindQuestion:
		specific = []string{"target"}
	case KindScope:
		specific = []string{"target", "operator"}
	case KindModality:
		specific = []string{"target", "modality"}
	case KindQuantifier:
		specific = []string{"target", "quantifier"}
	case KindDiscourse:
		specific = []string{"act", "content"}
	}
	for _, field := range specific {
		fields[field] = true
	}
	return fields
}

// Operator values affect interpretation only inside their explicit target
// scope. They never create asserted facts by themselves.
type Operator string

const (
	OperatorNegation     Operator = "negation"
	OperatorQuestion     Operator = "question"
	OperatorHypothetical Operator = "hypothetical"
)

// Valid reports whether operator is supported by the v1 scope node.
func (operator Operator) Valid() bool {
	switch operator {
	case OperatorNegation, OperatorQuestion, OperatorHypothetical:
		return true
	default:
		return false
	}
}

type Modality string

const (
	ModalityPossible  Modality = "possible"
	ModalityNecessary Modality = "necessary"
	ModalityLikely    Modality = "likely"
)

// Valid reports whether modality is supported by the v1 Meaning IR.
func (modality Modality) Valid() bool {
	switch modality {
	case ModalityPossible, ModalityNecessary, ModalityLikely:
		return true
	default:
		return false
	}
}

type DialogueAct string

const (
	ActStatement   DialogueAct = "statement"
	ActQuestion    DialogueAct = "question"
	ActRequest     DialogueAct = "request"
	ActRefusal     DialogueAct = "refusal"
	ActExplanation DialogueAct = "explanation"
)

// Valid reports whether act is supported by the v1 discourse node.
func (act DialogueAct) Valid() bool {
	switch act {
	case ActStatement, ActQuestion, ActRequest, ActRefusal, ActExplanation:
		return true
	default:
		return false
	}
}

// Participant assigns one semantic role to a target node. Role is declared
// by the event/state Concept in CORE; its target is a Meaning node.
type Participant struct {
	Role   Role   `json:"role"`
	Target NodeID `json:"target"`
}

// Node is a closed tagged union. Only fields belonging to Kind may be set:
// Concept for event/state/referent; Participants and TemporalFrame for
// event/state; Multiplicity for referent; Target for question/scope/modality;
// Operator for scope; Modality for modality; Act and Content for discourse.
// ValidateMeaning enforces the union at runtime.
type Node struct {
	ID                NodeID        `json:"id"`
	Kind              NodeKind      `json:"kind"`
	Concept           ConceptRef    `json:"concept,omitempty"`
	Participants      []Participant `json:"participants,omitempty"`
	TemporalFrame     TemporalFrame `json:"temporal_frame,omitempty"`
	RelationPredicate bool          `json:"relation_predicate,omitempty"`
	Multiplicity      Multiplicity  `json:"multiplicity,omitempty"`
	Target            NodeID        `json:"target,omitempty"`
	Operator          Operator      `json:"operator,omitempty"`
	Modality          Modality      `json:"modality,omitempty"`
	Quantifier        *Quantifier   `json:"quantifier,omitempty"`
	Act               DialogueAct   `json:"act,omitempty"`
	Content           []NodeID      `json:"content,omitempty"`
}

// Meaning is one versioned semantic interpretation. Roots are entry points;
// nodes are request-scoped and have no assertion/truth field.
type Meaning struct {
	SchemaVersion int      `json:"schema_version"`
	Roots         []NodeID `json:"roots"`
	Nodes         []Node   `json:"nodes"`
}
