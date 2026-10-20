package meaning

import (
	"sort"

	"github.com/clickbytes/golingua/internal/status"
)

// UnresolvedReason explains why content was not projected as an assertion.
type UnresolvedReason string

const (
	UnresolvedNegated        UnresolvedReason = "negated_scope"
	UnresolvedQuestion       UnresolvedReason = "question_scope"
	UnresolvedHypothetical   UnresolvedReason = "hypothetical_scope"
	UnresolvedModality       UnresolvedReason = "modal_scope"
	UnresolvedQuantifier     UnresolvedReason = "quantified_scope"
	UnresolvedDialogueAct    UnresolvedReason = "non_assertive_dialogue_act"
	UnresolvedNoAssertionAct UnresolvedReason = "no_assertive_dialogue_act"
	UnresolvedNonProposition UnresolvedReason = "not_a_proposition"
)

// Assertion is a proposition explicitly wrapped by a statement discourse act.
// It retains the source path that licensed projection; it is not world truth.
type Assertion struct {
	NodeID       NodeID        `json:"node_id"`
	Kind         NodeKind      `json:"kind"`
	Concept      ConceptRef    `json:"concept"`
	Participants []Participant `json:"participants,omitempty"`
	SourcePath   []NodeID      `json:"source_path"`
}

// Unresolved records a semantic scope that cannot be projected as a positive
// assertion and the path through the Meaning graph that led to it.
type Unresolved struct {
	NodeID     NodeID           `json:"node_id"`
	Reason     UnresolvedReason `json:"reason"`
	SourcePath []NodeID         `json:"source_path"`
}

// AssertedProjection keeps licensed positive assertions separate from
// question, negative, hypothetical and otherwise unresolved content.
type AssertedProjection struct {
	SchemaVersion int          `json:"schema_version"`
	Asserted      []Assertion  `json:"asserted"`
	Unresolved    []Unresolved `json:"unresolved"`
}

type projectionFrame struct {
	node     NodeID
	path     []NodeID
	asserted bool
}

// ProjectAsserted projects only event/state nodes beneath an explicit
// statement act. Scoped operators produce Unresolved records and never emit
// their operands as positive assertions. Unsupported double negation fails
// closed with an unsupported status.
func ProjectAsserted(meaning Meaning, budget Budget) (AssertedProjection, error) {
	if err := ValidateStructure(meaning, budget); err != nil {
		return AssertedProjection{}, err
	}
	nodes := make(map[NodeID]Node, len(meaning.Nodes))
	for _, node := range meaning.Nodes {
		nodes[node.ID] = node
	}
	if err := validateOperatorComposition(nodes, budget.MaxDepth); err != nil {
		return AssertedProjection{}, err
	}
	roots := append([]NodeID(nil), meaning.Roots...)
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	queue := make([]projectionFrame, 0, len(roots))
	for _, root := range roots {
		queue = append(queue, projectionFrame{node: root, path: []NodeID{root}})
	}
	projection := AssertedProjection{SchemaVersion: SchemaVersion, Asserted: []Assertion{}, Unresolved: []Unresolved{}}
	visited, edges := 0, 0
	for head := 0; head < len(queue); head++ {
		frame := queue[head]
		visited++
		if visited > budget.MaxNodes {
			return AssertedProjection{}, resourceLimit("nodes", visited, budget.MaxNodes)
		}
		if len(frame.path) > budget.MaxDepth {
			return AssertedProjection{}, resourceLimit("depth", len(frame.path), budget.MaxDepth)
		}
		node := nodes[frame.node]
		appendUnresolved := func(reason UnresolvedReason) {
			projection.Unresolved = append(projection.Unresolved, Unresolved{NodeID: node.ID, Reason: reason, SourcePath: append([]NodeID(nil), frame.path...)})
		}
		appendScopedUnresolved := func(reason UnresolvedReason) {
			path := operatorScopePath(node, frame.path, nodes, budget.MaxDepth)
			projection.Unresolved = append(projection.Unresolved, Unresolved{NodeID: node.ID, Reason: reason, SourcePath: path})
		}
		switch node.Kind {
		case KindDiscourse:
			if node.Act != ActStatement {
				appendUnresolved(UnresolvedDialogueAct)
				continue
			}
			for _, target := range node.Content {
				edges++
				if edges > budget.MaxEdges {
					return AssertedProjection{}, resourceLimit("edges", edges, budget.MaxEdges)
				}
				queue = append(queue, projectionFrame{node: target, path: appendPath(frame.path, target), asserted: true})
			}
		case KindQuestion:
			appendScopedUnresolved(UnresolvedQuestion)
		case KindScope:
			switch node.Operator {
			case OperatorNegation:
				appendScopedUnresolved(UnresolvedNegated)
			case OperatorQuestion:
				appendScopedUnresolved(UnresolvedQuestion)
			case OperatorHypothetical:
				appendScopedUnresolved(UnresolvedHypothetical)
			}
		case KindModality:
			appendScopedUnresolved(UnresolvedModality)
		case KindQuantifier:
			appendScopedUnresolved(UnresolvedQuantifier)
		case KindEvent, KindState:
			if !frame.asserted {
				appendUnresolved(UnresolvedNoAssertionAct)
				continue
			}
			participants := append([]Participant(nil), node.Participants...)
			sort.Slice(participants, func(i, j int) bool {
				if participants[i].Role != participants[j].Role {
					return participants[i].Role < participants[j].Role
				}
				return participants[i].Target < participants[j].Target
			})
			projection.Asserted = append(projection.Asserted, Assertion{
				NodeID: node.ID, Kind: node.Kind, Concept: node.Concept,
				Participants: participants,
				SourcePath:   append([]NodeID(nil), frame.path...),
			})
		case KindReferent:
			appendUnresolved(UnresolvedNonProposition)
		}
	}
	return projection, nil
}

func validateOperatorComposition(nodes map[NodeID]Node, maxDepth int) error {
	ids := make([]NodeID, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		start := nodes[id]
		if start.Kind != KindScope || start.Operator != OperatorNegation {
			continue
		}
		current := start
		for depth := 1; depth <= maxDepth; depth++ {
			target, exists := nodes[current.Target]
			if !exists || !isOperatorNode(target) {
				break
			}
			if target.Kind == KindScope && target.Operator == OperatorNegation {
				return unsupported("UNLICENSED_DOUBLE_NEGATION", "nodes."+string(target.ID), "double negation has no licensed semantic construction")
			}
			current = target
			if depth == maxDepth {
				return resourceLimit("depth", depth+1, maxDepth)
			}
		}
	}
	return nil
}

func appendPath(path []NodeID, target NodeID) []NodeID {
	result := make([]NodeID, len(path)+1)
	copy(result, path)
	result[len(path)] = target
	return result
}

func operatorScopePath(start Node, sourcePath []NodeID, nodes map[NodeID]Node, maximum int) []NodeID {
	path := append([]NodeID(nil), sourcePath...)
	current := start
	for depth := 0; depth < maximum && isOperatorNode(current); depth++ {
		target, exists := nodes[current.Target]
		if !exists {
			break
		}
		path = appendPath(path, target.ID)
		current = target
	}
	return path
}

func unsupported(code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &StatusError{result: status.Result{Status: status.OutcomeUnsupported, Diagnostics: []status.Diagnostic{diagnostic}}}
}
