package engine

import "github.com/clickbytes/golingua/internal/algebra"

// algebraRelationEventPathPlan extends the anonymous binary relation path
// evidence with two independently bound Event witnesses. Temporal frames are
// copied verbatim from the checked input and are never inferred or rewritten.
type algebraRelationEventPathPlan struct {
	algebraRelationPathPlan
	EventPort                string
	EventPortPosition        int
	EntityPortPositions      [2]int
	FirstEvent               *algebra.Term
	FirstEventPath           string
	FirstEventBinder         string
	FirstEventTemporalFrame  string
	FirstEventAtom           *algebra.Term
	SecondEvent              *algebra.Term
	SecondEventPath          string
	SecondEventBinder        string
	SecondEventTemporalFrame string
	SecondEventAtom          *algebra.Term
}

// matchAlgebraEventRelationPath recognizes exactly
// λx:Entity.λy:Entity.∃z:Entity.And(∃e1:Event.r(e1,x,z),∃e2:Event.r(e2,z,y))
// after the caller has run algebra.Check. It resolves neither aliases nor
// truth and accepts either operand order for the conjunction.
func matchAlgebraEventRelationPath(root *algebra.Term, signatures map[string]algebra.Signature) (algebraRelationEventPathPlan, bool) {
	var out algebraRelationEventPathPlan
	if root == nil || root.Kind != algebra.TermLambda || !termHasKind(root, algebra.TypeEntity) || root.ID == "" || root.Body == nil {
		return out, false
	}
	out.RootPath = "input.expression"
	out.FirstLambda, out.FirstPath, out.FirstBinder = root, out.RootPath, root.ID

	second := root.Body
	if second.Kind != algebra.TermLambda || !termHasKind(second, algebra.TypeEntity) || second.ID == "" || second.ID == root.ID || second.Body == nil {
		return algebraRelationEventPathPlan{}, false
	}
	out.SecondLambda, out.SecondPath, out.SecondBinder = second, out.RootPath+".body", second.ID

	witness := second.Body
	if witness.Kind != algebra.TermExists || !termHasKind(witness, algebra.TypeEntity) || witness.ID == "" || binderConflicts(witness.ID, root.ID, second.ID) || witness.Body == nil {
		return algebraRelationEventPathPlan{}, false
	}
	out.Witness, out.WitnessPath, out.WitnessBinder = witness, out.SecondPath+".body", witness.ID
	conjunction := witness.Body
	if conjunction.Kind != algebra.TermAnd || len(conjunction.Operands) != 2 || conjunction.Operands[0] == nil || conjunction.Operands[1] == nil {
		return algebraRelationEventPathPlan{}, false
	}
	out.Conjunction, out.ConjunctionPath = conjunction, out.WitnessPath+".body"

	left, leftOK := eventRelationOperand(conjunction.Operands[0], out.ConjunctionPath+pathIndex(0), signatures, root.ID, second.ID, witness.ID)
	right, rightOK := eventRelationOperand(conjunction.Operands[1], out.ConjunctionPath+pathIndex(1), signatures, root.ID, second.ID, witness.ID)
	if !leftOK || !rightOK || left.signature.ID != right.signature.ID || left.eventID == right.eventID || left.pattern == right.pattern {
		return algebraRelationEventPathPlan{}, false
	}
	first, secondPath := left, right
	if left.pattern == relationEventZY && right.pattern == relationEventXZ {
		first, secondPath = right, left
	} else if left.pattern != relationEventXZ || right.pattern != relationEventZY {
		return algebraRelationEventPathPlan{}, false
	}
	if first.signature.ID == "" || len(first.signature.Ports) != 3 || first.entityPositions != secondPath.entityPositions || first.eventPosition != secondPath.eventPosition || first.eventRole != secondPath.eventRole || first.entityRoles != secondPath.entityRoles {
		return algebraRelationEventPathPlan{}, false
	}
	if first.entityRoles[0] == first.entityRoles[1] || first.eventRole == first.entityRoles[0] || first.eventRole == first.entityRoles[1] {
		return algebraRelationEventPathPlan{}, false
	}

	out.Symbol, out.Signature = first.signature.ID, first.signature
	out.FirstPorts, out.SecondPorts = first.entityRoles, secondPath.entityRoles
	out.EntityPortPositions, out.EventPortPosition = first.entityPositions, first.eventPosition
	out.EventPort = first.eventRole
	out.FirstEvent, out.FirstEventPath, out.FirstEventBinder = first.eventTerm, first.eventPath, first.eventID
	out.FirstEventTemporalFrame, out.FirstEventAtom = first.eventTerm.TemporalFrame, first.atom
	out.FirstAtom, out.FirstAtomPath = first.atom, first.atomPath
	out.SecondEvent, out.SecondEventPath, out.SecondEventBinder = secondPath.eventTerm, secondPath.eventPath, secondPath.eventID
	out.SecondEventTemporalFrame, out.SecondEventAtom = secondPath.eventTerm.TemporalFrame, secondPath.atom
	out.SecondAtom, out.SecondAtomPath = secondPath.atom, secondPath.atomPath
	return out, true
}

type relationEventPattern uint8

const (
	relationEventUnknown relationEventPattern = iota
	relationEventXZ
	relationEventZY
)

type relationEventOperand struct {
	pattern         relationEventPattern
	eventTerm       *algebra.Term
	eventPath       string
	eventID         string
	atom            *algebra.Term
	atomPath        string
	signature       algebra.Signature
	eventRole       string
	eventPosition   int
	entityRoles     [2]string
	entityPositions [2]int
}

func eventRelationOperand(term *algebra.Term, path string, signatures map[string]algebra.Signature, xID, yID, zID string) (relationEventOperand, bool) {
	var out relationEventOperand
	if term == nil || term.Kind != algebra.TermExists || !termHasKind(term, algebra.TypeEvent) || term.ID == "" || binderConflicts(term.ID, xID, yID, zID) || term.Body == nil || term.Body.Kind != algebra.TermAtom {
		return out, false
	}
	atom := term.Body
	signature, ok := signatures[atom.Symbol]
	if !ok || signature.ID != atom.Symbol || len(signature.Ports) != 3 || len(atom.Args) != 3 {
		return out, false
	}
	if atom.Symbol == "" {
		return out, false
	}
	eventPosition := -1
	entityPositions := [2]int{-1, -1}
	entityRoles := [2]string{}
	seenRoles := make(map[string]bool, len(signature.Ports))
	for index, port := range signature.Ports {
		if port.Role == "" || seenRoles[port.Role] {
			return out, false
		}
		seenRoles[port.Role] = true
		if atom.Args[index].Role != port.Role || atom.Args[index].Value == nil || atom.Args[index].Value.Kind != algebra.TermVar {
			return out, false
		}
		switch port.Type.Kind {
		case algebra.TypeEvent:
			if eventPosition >= 0 || port.Type.SortID != "" || port.Type.From != nil || port.Type.To != nil || atom.Args[index].Value.ID != term.ID {
				return out, false
			}
			eventPosition = index
		case algebra.TypeEntity:
			if entityPositions[1] >= 0 || port.Type.SortID != "" || port.Type.From != nil || port.Type.To != nil {
				return out, false
			}
			if entityPositions[0] < 0 {
				entityPositions[0], entityRoles[0] = index, port.Role
			} else {
				entityPositions[1], entityRoles[1] = index, port.Role
			}
		default:
			return out, false
		}
	}
	if eventPosition < 0 || entityPositions[0] < 0 || entityPositions[1] < 0 {
		return out, false
	}
	firstID, secondID := atom.Args[entityPositions[0]].Value.ID, atom.Args[entityPositions[1]].Value.ID
	pattern := relationEventUnknown
	switch {
	case firstID == xID && secondID == zID:
		pattern = relationEventXZ
	case firstID == zID && secondID == yID:
		pattern = relationEventZY
	default:
		return out, false
	}
	out = relationEventOperand{
		pattern: pattern, eventTerm: term, eventPath: path, eventID: term.ID,
		atom: atom, atomPath: path + ".body", signature: signature,
		eventRole: signature.Ports[eventPosition].Role, eventPosition: eventPosition,
		entityRoles: entityRoles, entityPositions: entityPositions,
	}
	return out, true
}

func termHasKind(term *algebra.Term, kind algebra.TypeKind) bool {
	return term != nil && term.Type != nil && term.Type.Kind == kind && term.Type.SortID == "" && term.Type.From == nil && term.Type.To == nil
}

func binderConflicts(candidate string, existing ...string) bool {
	if candidate == "" {
		return true
	}
	for _, id := range existing {
		if candidate == id {
			return true
		}
	}
	return false
}
