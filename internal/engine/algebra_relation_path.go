package engine

import "github.com/clickbytes/golingua/internal/algebra"

// algebraRelationPathPlan records the exact checked-tree locations and
// signature ports for λx:Entity. λy:Entity. ∃z:Entity.
// And(r(x,z), r(z,y)). It describes structure only; it makes no claim about
// the truth of either relation atom.
type algebraRelationPathPlan struct {
	RootPath        string
	FirstLambda     *algebra.Term
	FirstPath       string
	FirstBinder     string
	SecondLambda    *algebra.Term
	SecondPath      string
	SecondBinder    string
	Witness         *algebra.Term
	WitnessPath     string
	WitnessBinder   string
	Conjunction     *algebra.Term
	ConjunctionPath string
	FirstAtom       *algebra.Term
	FirstAtomPath   string
	SecondAtom      *algebra.Term
	SecondAtomPath  string
	Symbol          string
	Signature       algebra.Signature
	FirstPorts      [2]string
	SecondPorts     [2]string
}

// matchAlgebraRelationPath recognizes one exact binary relation path in an IR
// that has already passed algebra.Check. Every name and symbol is read from
// the supplied expression and selected signatures; none is inferred or
// hard-coded here.
func matchAlgebraRelationPath(root *algebra.Term, signatures map[string]algebra.Signature) (algebraRelationPathPlan, bool) {
	var plan algebraRelationPathPlan
	if root == nil || root.Kind != algebra.TermLambda || root.Type == nil || root.Type.Kind != algebra.TypeEntity || root.Body == nil || root.ID == "" {
		return plan, false
	}
	plan.RootPath = "input.expression"
	plan.FirstLambda, plan.FirstPath, plan.FirstBinder = root, plan.RootPath, root.ID

	second := root.Body
	if second.Kind != algebra.TermLambda || second.Type == nil || second.Type.Kind != algebra.TypeEntity || second.Body == nil || second.ID == "" || second.ID == root.ID {
		return algebraRelationPathPlan{}, false
	}
	plan.SecondLambda, plan.SecondPath, plan.SecondBinder = second, plan.RootPath+".body", second.ID

	witness := second.Body
	if witness.Kind != algebra.TermExists || witness.Type == nil || witness.Type.Kind != algebra.TypeEntity || witness.ID == "" || witness.ID == root.ID || witness.ID == second.ID || witness.Body == nil {
		return algebraRelationPathPlan{}, false
	}
	plan.Witness, plan.WitnessPath, plan.WitnessBinder = witness, plan.SecondPath+".body", witness.ID

	conjunction := witness.Body
	if conjunction.Kind != algebra.TermAnd || len(conjunction.Operands) != 2 || conjunction.Operands[0] == nil || conjunction.Operands[1] == nil {
		return algebraRelationPathPlan{}, false
	}
	plan.Conjunction, plan.ConjunctionPath = conjunction, plan.WitnessPath+".body"

	left, right := conjunction.Operands[0], conjunction.Operands[1]
	if left.Kind != algebra.TermAtom || right.Kind != algebra.TermAtom || left.Symbol == "" || left.Symbol != right.Symbol {
		return algebraRelationPathPlan{}, false
	}
	sig, ok := signatures[left.Symbol]
	if !ok || sig.ID != left.Symbol || len(sig.Ports) != 2 || len(left.Args) != 2 || len(right.Args) != 2 {
		return algebraRelationPathPlan{}, false
	}
	if sig.Ports[0].Type.Kind != algebra.TypeEntity || sig.Ports[1].Type.Kind != algebra.TypeEntity || sig.Ports[0].Role == "" || sig.Ports[1].Role == "" || sig.Ports[0].Role == sig.Ports[1].Role {
		return algebraRelationPathPlan{}, false
	}
	for i := 0; i < 2; i++ {
		if left.Args[i].Role != sig.Ports[i].Role || right.Args[i].Role != sig.Ports[i].Role || left.Args[i].Value == nil || right.Args[i].Value == nil || left.Args[i].Value.Kind != algebra.TermVar || right.Args[i].Value.Kind != algebra.TermVar {
			return algebraRelationPathPlan{}, false
		}
	}
	first, secondAtom := left, right
	firstIndex, secondIndex := 0, 1
	if left.Args[0].Value.ID == witness.ID && left.Args[1].Value.ID == second.ID && right.Args[0].Value.ID == root.ID && right.Args[1].Value.ID == witness.ID {
		first, secondAtom = right, left
		firstIndex, secondIndex = 1, 0
	} else if !(left.Args[0].Value.ID == root.ID && left.Args[1].Value.ID == witness.ID && right.Args[0].Value.ID == witness.ID && right.Args[1].Value.ID == second.ID) {
		return algebraRelationPathPlan{}, false
	}

	plan.FirstAtom, plan.FirstAtomPath = first, plan.ConjunctionPath+pathIndex(firstIndex)
	plan.SecondAtom, plan.SecondAtomPath = secondAtom, plan.ConjunctionPath+pathIndex(secondIndex)
	plan.Symbol, plan.Signature = first.Symbol, sig
	plan.FirstPorts = [2]string{first.Args[0].Role, first.Args[1].Role}
	plan.SecondPorts = [2]string{secondAtom.Args[0].Role, secondAtom.Args[1].Role}
	return plan, true
}

func pathIndex(index int) string {
	if index == 0 {
		return ".operands[0]"
	}
	return ".operands[1]"
}
