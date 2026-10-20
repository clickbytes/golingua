package meaning

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
)

// NewEvent creates a checked event node with defensive participant storage.
func NewEvent(id NodeID, concept ids.ConceptID, participants []Participant) (Node, error) {
	return newPredicateNode(id, KindEvent, concept, participants)
}

// Build validates and defensively copies a complete request-scoped Meaning.
func Build(roots []NodeID, nodes []Node, budget Budget) (Meaning, error) {
	meaning := Meaning{SchemaVersion: SchemaVersion, Roots: roots, Nodes: nodes}
	if err := ValidateStructure(meaning, budget); err != nil {
		return Meaning{}, err
	}
	meaning = Meaning{SchemaVersion: SchemaVersion, Roots: append([]NodeID(nil), roots...), Nodes: make([]Node, len(nodes))}
	for index, node := range nodes {
		meaning.Nodes[index] = node
		meaning.Nodes[index].Participants = cloneParticipants(node.Participants)
		meaning.Nodes[index].Content = append([]NodeID(nil), node.Content...)
		if node.Quantifier != nil {
			quantifier := cloneQuantifier(*node.Quantifier)
			meaning.Nodes[index].Quantifier = &quantifier
		}
	}
	return meaning, nil
}

// NewState creates a checked state node with defensive participant storage.
func NewState(id NodeID, concept ids.ConceptID, participants []Participant) (Node, error) {
	return newPredicateNode(id, KindState, concept, participants)
}

// NewReferent creates a checked referent node for a stable CORE concept.
func NewReferent(id NodeID, concept ids.ConceptID) (Node, error) {
	if !validNodeID(id) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node ID must use the bounded ASCII identifier form")
	}
	if _, err := ids.ParseConceptID(concept.String()); err != nil {
		return Node{}, invalid("INVALID_CONCEPT_ID", "node.concept", "concept ID is malformed")
	}
	return Node{ID: id, Kind: KindReferent, Concept: ConceptRefFromID(concept)}, nil
}

// NewQuestion creates a question wrapper around one meaning node.
func NewQuestion(id, target NodeID) (Node, error) {
	if !validNodeID(id) || !validNodeID(target) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node and target IDs must use the bounded ASCII identifier form")
	}
	return Node{ID: id, Kind: KindQuestion, Target: target}, nil
}

// NewNegation creates an explicit negative scope; it does not assert a fact.
func NewNegation(id, target NodeID) (Node, error) {
	return NewScope(id, OperatorNegation, target)
}

// NewScope creates a checked operator scope over one target node.
func NewScope(id NodeID, operator Operator, target NodeID) (Node, error) {
	node := Node{ID: id, Kind: KindScope, Operator: operator, Target: target}
	if !operator.Valid() {
		return Node{}, invalid("INVALID_OPERATOR", "node.operator", "scope operator is empty or unsupported")
	}
	if !validNodeID(id) || !validNodeID(target) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node and target IDs must use the bounded ASCII identifier form")
	}
	return node, nil
}

// NewModalityNode creates a checked modality wrapper around one target node.
func NewModalityNode(id NodeID, modality Modality, target NodeID) (Node, error) {
	node := Node{ID: id, Kind: KindModality, Modality: modality, Target: target}
	if !modality.Valid() {
		return Node{}, invalid("INVALID_MODALITY", "node.modality", "modality is empty or unsupported")
	}
	if !validNodeID(id) || !validNodeID(target) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node and target IDs must use the bounded ASCII identifier form")
	}
	return node, nil
}

// NewDiscourse creates a checked discourse node and copies its ordered content.
func NewDiscourse(id NodeID, act DialogueAct, content []NodeID) (Node, error) {
	if !act.Valid() {
		return Node{}, invalid("INVALID_DIALOGUE_ACT", "node.act", "dialogue act is empty or unsupported")
	}
	if !validNodeID(id) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node ID must use the bounded ASCII identifier form")
	}
	if len(content) == 0 {
		return Node{}, invalid("INVALID_DISCOURSE", "node.content", "discourse content must contain at least one node")
	}
	if len(content) > MaxDiscourseItems {
		return Node{}, resourceLimit("discourse_items", len(content), MaxDiscourseItems)
	}
	for _, target := range content {
		if !validNodeID(target) {
			return Node{}, invalid("INVALID_DISCOURSE_TARGET", "node.content", "discourse content contains an invalid node ID")
		}
	}
	return Node{ID: id, Kind: KindDiscourse, Act: act, Content: append([]NodeID(nil), content...)}, nil
}

func cloneParticipants(participants []Participant) []Participant {
	return append([]Participant(nil), participants...)
}

func newPredicateNode(id NodeID, kind NodeKind, concept ids.ConceptID, participants []Participant) (Node, error) {
	if !validNodeID(id) {
		return Node{}, invalid("INVALID_NODE_ID", "node.id", "node ID must use the bounded ASCII identifier form")
	}
	if _, err := ids.ParseConceptID(concept.String()); err != nil {
		return Node{}, invalid("INVALID_CONCEPT_ID", "node.concept", "concept ID is malformed")
	}
	if len(participants) > MaxParticipants {
		return Node{}, resourceLimit("participants", len(participants), MaxParticipants)
	}
	for index, participant := range participants {
		if !validName(string(participant.Role)) || !validNodeID(participant.Target) {
			return Node{}, invalid("INVALID_PARTICIPANT", "node.participants", fmt.Sprintf("participant %d has an invalid role or target", index))
		}
	}
	return Node{ID: id, Kind: kind, Concept: ConceptRefFromID(concept), Participants: cloneParticipants(participants)}, nil
}
