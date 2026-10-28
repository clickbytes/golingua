package engine

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
)

// algebraScopePlan is a checked structural description of the two recognized
// scope patterns. It retains source nodes and source paths so a consumer
// cannot accidentally erase scope.
type algebraScopePlan struct {
	OuterAll        *algebra.Term
	OuterPath       string
	OuterNegation   *algebra.Term
	OuterNegPath    string
	Restrictor      *algebra.Term
	RestrictorPath  string
	InnerExists     *algebra.Term
	InnerPath       string
	Conjunction     *algebra.Term
	ConjunctionPath string
	InnerNegation   *algebra.Term
	InnerNegPath    string
	Predicate       *algebra.Term
	PredicatePath   string
	EventExists     *algebra.Term
	EventPath       string
	EventAtom       *algebra.Term
	EventAtomPath   string
	EventFrame      string
	OuterBinderID   string
	InnerBinderID   string
	EventBinderID   string
	OuterPort       string
	InnerPort       string
	EventPort       string
	EventOuterPort  string
	EventInnerPort  string
	OuterNegated    bool
	InnerNegated    bool
}

// matchAlgebraScope recognizes only the checked quantified scope pattern.
// Callers must run algebra.Check first; this function still fails closed on
// malformed shapes and verifies every atom against the selected signatures.
func matchAlgebraScope(root *algebra.Term, signatures map[string]algebra.Signature) (algebraScopePlan, bool) {
	var plan algebraScopePlan
	if root == nil {
		return plan, false
	}
	path := "input.expression"
	outerNegated := root.Kind == algebra.TermNot
	outer := root
	if outerNegated {
		if root.Body == nil {
			return plan, false
		}
		plan.OuterNegation, plan.OuterNegPath = root, path
		outer, path = root.Body, path+".body"
	}
	if outer.Kind != algebra.TermQuantifier || outer.Quantifier != "all" || outer.Type == nil || outer.Type.Kind != algebra.TypeEntity || outer.Restrictor == nil || outer.Body == nil {
		return algebraScopePlan{}, false
	}
	plan.OuterAll, plan.OuterPath, plan.OuterBinderID = outer, path, outer.ID
	plan.OuterNegated = outerNegated
	plan.Restrictor, plan.RestrictorPath = outer.Restrictor, path+".restrictor"
	if !scopeUnaryAtom(outer.Restrictor, signatures, outer.ID, &plan.OuterPort) {
		return algebraScopePlan{}, false
	}

	body, bodyPath := outer.Body, path+".body"
	innerNegated := body.Kind == algebra.TermNot
	if innerNegated {
		if body.Body == nil {
			return algebraScopePlan{}, false
		}
		plan.InnerNegation, plan.InnerNegPath = body, bodyPath
		body, bodyPath = body.Body, bodyPath+".body"
	}
	if body.Kind != algebra.TermExists || body.Type == nil || body.Type.Kind != algebra.TypeEntity || body.ID == outer.ID || body.Body == nil {
		return algebraScopePlan{}, false
	}
	// The two accepted placements are outer negation or inner negation.
	if outerNegated == innerNegated {
		return algebraScopePlan{}, false
	}
	plan.InnerExists, plan.InnerPath, plan.InnerBinderID = body, bodyPath, body.ID
	plan.InnerNegated = innerNegated
	conjunction, conjunctionPath := body.Body, bodyPath+".body"
	if conjunction.Kind != algebra.TermAnd || len(conjunction.Operands) != 2 {
		return algebraScopePlan{}, false
	}
	plan.Conjunction, plan.ConjunctionPath = conjunction, conjunctionPath
	var predicate, eventExists *algebra.Term
	var predicatePath, eventPath string
	for i, operand := range conjunction.Operands {
		opath := fmt.Sprintf("%s.operands[%d]", conjunctionPath, i)
		if operand == nil {
			return algebraScopePlan{}, false
		}
		if operand.Kind == algebra.TermExists {
			if eventExists != nil {
				return algebraScopePlan{}, false
			}
			eventExists, eventPath = operand, opath
		} else {
			if predicate != nil || operand.Kind != algebra.TermAtom {
				return algebraScopePlan{}, false
			}
			predicate, predicatePath = operand, opath
		}
	}
	if predicate == nil || eventExists == nil || eventExists.Type == nil || eventExists.Type.Kind != algebra.TypeEvent || eventExists.ID == outer.ID || eventExists.ID == body.ID || eventExists.Body == nil {
		return algebraScopePlan{}, false
	}
	plan.Predicate, plan.PredicatePath = predicate, predicatePath
	if !scopeUnaryAtom(predicate, signatures, body.ID, &plan.InnerPort) {
		return algebraScopePlan{}, false
	}
	plan.EventExists, plan.EventPath, plan.EventBinderID = eventExists, eventPath, eventExists.ID
	plan.EventFrame = eventExists.TemporalFrame
	atomPath := eventPath + ".body"
	atom := eventExists.Body
	if atom == nil || atom.Kind != algebra.TermAtom {
		return algebraScopePlan{}, false
	}
	sig, ok := signatures[atom.Symbol]
	if !ok || len(sig.Ports) != 3 || len(atom.Args) != 3 {
		return algebraScopePlan{}, false
	}
	used := map[string]int{}
	for i, port := range sig.Ports {
		arg := atom.Args[i]
		if arg.Role != port.Role || arg.Value == nil || arg.Value.Kind != algebra.TermVar {
			return algebraScopePlan{}, false
		}
		want := ""
		switch port.Type.Kind {
		case algebra.TypeEvent:
			want = eventExists.ID
			plan.EventPort = port.Role
		case algebra.TypeEntity:
			if arg.Value.ID == outer.ID {
				want = outer.ID
				plan.EventOuterPort = port.Role
			} else if arg.Value.ID == body.ID {
				want = body.ID
				plan.EventInnerPort = port.Role
			}
		default:
			return algebraScopePlan{}, false
		}
		if arg.Value.ID != want {
			return algebraScopePlan{}, false
		}
		used[arg.Value.ID]++
	}
	if used[outer.ID] != 1 || used[body.ID] != 1 || used[eventExists.ID] != 1 || plan.EventOuterPort == "" || plan.EventInnerPort == "" || plan.EventOuterPort == plan.EventInnerPort || plan.EventPort == "" {
		return algebraScopePlan{}, false
	}
	plan.EventAtom, plan.EventAtomPath = atom, atomPath
	return plan, true
}

func scopeUnaryAtom(atom *algebra.Term, signatures map[string]algebra.Signature, binder string, portID *string) bool {
	if atom == nil || atom.Kind != algebra.TermAtom || len(atom.Args) != 1 {
		return false
	}
	sig, ok := signatures[atom.Symbol]
	if !ok || len(sig.Ports) != 1 || sig.Ports[0].Type.Kind != algebra.TypeEntity || atom.Args[0].Role != sig.Ports[0].Role || atom.Args[0].Value == nil || atom.Args[0].Value.Kind != algebra.TermVar || atom.Args[0].Value.ID != binder {
		return false
	}
	*portID = sig.Ports[0].Role
	return true
}
