package algebra

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/status"
)

// Substitution maps declared parameter IDs to terms. Returned maps and terms
// are detached copies; applying a substitution never mutates its inputs.
type Substitution map[string]*Term

// Unify computes a bounded, first-order unifier for the positive structural
// fragment (Params, Refs, Atoms, and And). It is not fact evaluation.
func Unify(context Document, left, right *Term) (Substitution, error) {
	if left == nil || right == nil {
		return nil, invalid("MISSING_TERM", "terms", "both terms are required")
	}
	context.Closed = false // parameters are the declared unification variables
	if err := checkUnificationTerm(context, left); err != nil {
		return nil, err
	}
	if err := checkUnificationTerm(context, right); err != nil {
		return nil, err
	}
	limits, err := NormalizeLimits(context.Limits)
	if err != nil {
		return nil, err
	}
	leftType, err := supportedTermType(left, context, limits)
	if err != nil {
		return nil, err
	}
	rightType, err := supportedTermType(right, context, limits)
	if err != nil {
		return nil, err
	}
	if !TypeEqual(leftType, rightType) {
		return nil, invalid("UNIFICATION_TYPE_MISMATCH", "terms", "terms have different types")
	}
	parameterTypes := make(map[string]Type, len(context.Parameters))
	referenceTypes := make(map[string]Type, len(context.References))
	for _, declaration := range context.Parameters {
		parameterTypes[declaration.ID] = declaration.Type
	}
	for _, declaration := range context.References {
		referenceTypes[declaration.ID] = declaration.Type
	}
	state := unifier{limits: limits, parameterTypes: parameterTypes, referenceTypes: referenceTypes, substitution: make(Substitution)}
	if err := state.unify(left, right, "terms", 1); err != nil {
		return nil, err
	}
	result := make(Substitution, len(state.substitution))
	for id := range state.substitution {
		value, err := state.resolve(&Term{Kind: TermParam, ID: id, Type: typePointer(parameterTypes[id])}, "substitution."+id, 1, make(map[string]bool))
		if err != nil {
			return nil, err
		}
		result[id] = value
	}
	return result, nil
}

func checkUnificationTerm(context Document, term *Term) error {
	document := context
	document.Expression = term
	document.Closed = false
	checked, err := Check(document)
	if err != nil {
		return err
	}
	if checked.Status == status.OutcomeIncomplete {
		return &Error{Status: status.OutcomeIncomplete, Code: "UNIFICATION_INCOMPLETE", Path: "expression", Message: "unification does not assign or interpret holes"}
	}
	return nil
}

func typePointer(value Type) *Type { copyValue := value; return &copyValue }

func supportedTermType(term *Term, context Document, limits Limits) (Type, error) {
	parameters := make(map[string]Type, len(context.Parameters))
	references := make(map[string]Type, len(context.References))
	for _, item := range context.Parameters {
		parameters[item.ID] = item.Type
	}
	for _, item := range context.References {
		references[item.ID] = item.Type
	}
	active := make(map[*Term]bool)
	nodes := 0
	var visit func(*Term, string, int) (Type, error)
	visit = func(value *Term, path string, depth int) (Type, error) {
		if value == nil {
			return Type{}, invalid("MISSING_TERM", path, "term is required")
		}
		if depth > limits.MaxDepth {
			return Type{}, resourceLimit(path, fmt.Sprintf("expression depth exceeds limit %d", limits.MaxDepth))
		}
		if active[value] {
			return Type{}, invalid("CYCLIC_TERM", path, "expression contains a pointer cycle")
		}
		nodes++
		if nodes > limits.MaxNodes {
			return Type{}, resourceLimit(path, "unification term exceeds node limit")
		}
		active[value] = true
		defer delete(active, value)
		switch value.Kind {
		case TermParam:
			t, ok := parameters[value.ID]
			if !ok {
				return Type{}, invalid("UNDECLARED_PARAMETER", path+".id", "parameter is not declared")
			}
			if t.Kind == TypeArrow {
				return Type{}, &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_ARROW_VARIABLE", Path: path, Message: "function-typed parameters are outside first-order unification"}
			}
			return t, nil
		case TermRef:
			t, ok := references[value.ID]
			if !ok {
				return Type{}, invalid("UNDECLARED_REFERENCE", path+".id", "reference is not declared")
			}
			if t.Kind == TypeArrow {
				return Type{}, &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_ARROW_REFERENCE", Path: path, Message: "function-typed references are outside first-order unification"}
			}
			return t, nil
		case TermAtom:
			for i, arg := range value.Args {
				if _, err := visit(arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1); err != nil {
					return Type{}, err
				}
			}
			return Type{Kind: TypeProp}, nil
		case TermAnd:
			if len(value.Operands) == 0 {
				return Type{}, invalid("EMPTY_CONJUNCTION", path+".operands", "And requires operands")
			}
			for i, child := range value.Operands {
				childType, err := visit(child, fmt.Sprintf("%s.operands[%d]", path, i), depth+1)
				if err != nil {
					return Type{}, err
				}
				if childType.Kind != TypeProp {
					return Type{}, invalid("EXPECTED_PROP", path, "And operands must have type Prop")
				}
			}
			return Type{Kind: TypeProp}, nil
		case TermHole:
			return Type{}, &Error{Status: status.OutcomeIncomplete, Code: "UNIFICATION_INCOMPLETE", Path: path, Message: "unification does not interpret holes"}
		case TermLambda, TermApply, TermExists, TermVar:
			return Type{}, &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_UNIFICATION_TERM", Path: path, Message: "this term kind is outside the bounded first-order unification profile"}
		default:
			return Type{}, invalid("UNKNOWN_TERM_KIND", path+".kind", "unknown term kind")
		}
	}
	return visit(term, "expression", 1)
}

type unifier struct {
	limits         Limits
	parameterTypes map[string]Type
	referenceTypes map[string]Type
	substitution   Substitution
	active         map[*Term]bool
	nodes          int
}

func (u *unifier) unify(left, right *Term, path string, depth int) error {
	if depth > u.limits.MaxDepth {
		return resourceLimit(path, "unification depth exceeds limit")
	}
	u.nodes++
	if u.nodes > u.limits.MaxNodes {
		return resourceLimit(path, "unification work exceeds node limit")
	}
	left, err := u.deref(left, path+".left")
	if err != nil {
		return err
	}
	right, err = u.deref(right, path+".right")
	if err != nil {
		return err
	}
	if left.Kind == TermParam {
		if right.Kind == TermParam && left.ID == right.ID {
			return nil
		}
		return u.bind(left.ID, right, path)
	}
	if right.Kind == TermParam {
		return u.bind(right.ID, left, path)
	}
	if left.Kind != right.Kind {
		return invalid("UNIFICATION_CONFLICT", path, "term constructors differ")
	}
	switch left.Kind {
	case TermRef:
		if left.ID != right.ID {
			return invalid("UNIFICATION_CONFLICT", path, "references are distinct and cannot be equated")
		}
		return nil
	case TermAtom:
		if left.Symbol != right.Symbol || len(left.Args) != len(right.Args) {
			return invalid("UNIFICATION_CONFLICT", path, "relation symbols or arities differ")
		}
		for i := range left.Args {
			if left.Args[i].Role != right.Args[i].Role {
				return invalid("UNIFICATION_CONFLICT", fmt.Sprintf("%s.args[%d].role", path, i), "ordered relation roles differ")
			}
			if err := u.unify(left.Args[i].Value, right.Args[i].Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1); err != nil {
				return err
			}
		}
		return nil
	case TermAnd:
		if len(left.Operands) != len(right.Operands) {
			return invalid("UNIFICATION_CONFLICT", path, "conjunction sizes differ")
		}
		for i := range left.Operands {
			if err := u.unify(left.Operands[i], right.Operands[i], fmt.Sprintf("%s.operands[%d]", path, i), depth+1); err != nil {
				return err
			}
		}
		return nil
	default:
		return &Error{Status: status.OutcomeUnsupported, Code: "UNSUPPORTED_UNIFICATION_TERM", Path: path, Message: "term kind is outside the bounded first-order profile"}
	}
}

func (u *unifier) deref(term *Term, path string) (*Term, error) {
	seen := make(map[string]bool)
	for term != nil && term.Kind == TermParam {
		replacement, ok := u.substitution[term.ID]
		if !ok {
			return term, nil
		}
		if seen[term.ID] {
			return nil, invalid("UNIFICATION_CYCLE", path, "substitution chain contains a cycle")
		}
		seen[term.ID] = true
		term = replacement
	}
	if term == nil {
		return nil, invalid("MISSING_TERM", path, "substitution produced a nil term")
	}
	return term, nil
}

func (u *unifier) resolve(term *Term, path string, depth int, chain map[string]bool) (*Term, error) {
	if depth > u.limits.MaxDepth {
		return nil, resourceLimit(path, "substitution resolution depth exceeds limit")
	}
	if term == nil {
		return nil, invalid("MISSING_TERM", path, "cannot resolve a nil term")
	}
	u.nodes++
	if u.nodes > u.limits.MaxNodes {
		return nil, resourceLimit(path, "substitution resolution exceeds node limit")
	}
	if term.Kind == TermParam {
		value, ok := u.substitution[term.ID]
		if !ok {
			return copyTermDetached(term, u.limits)
		}
		if chain[term.ID] {
			return nil, invalid("UNIFICATION_CYCLE", path, "substitution chain contains a cycle")
		}
		chain[term.ID] = true
		resolved, err := u.resolve(value, path, depth+1, chain)
		delete(chain, term.ID)
		return resolved, err
	}
	copyValue, err := copyTermDetached(term, u.limits)
	if err != nil {
		return nil, err
	}
	if term.Kind == TermAtom {
		for i, arg := range term.Args {
			copyValue.Args[i].Value, err = u.resolve(arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1, chain)
			if err != nil {
				return nil, err
			}
		}
	} else if term.Kind == TermAnd {
		for i, child := range term.Operands {
			copyValue.Operands[i], err = u.resolve(child, fmt.Sprintf("%s.operands[%d]", path, i), depth+1, chain)
			if err != nil {
				return nil, err
			}
		}
	}
	return copyValue, nil
}

func (u *unifier) bind(id string, term *Term, path string) error {
	if _, declared := u.parameterTypes[id]; !declared {
		return invalid("UNDECLARED_PARAMETER", path, "only declared parameters may be assigned")
	}
	resolved, err := u.deref(term, path)
	if err != nil {
		return err
	}
	if resolved.Kind == TermParam && resolved.ID == id {
		return nil
	}
	if !TypeEqual(u.parameterTypes[id], u.termTypeKnown(resolved)) {
		return invalid("UNIFICATION_TYPE_MISMATCH", path, "parameter and replacement types differ")
	}
	occurs, err := u.occurs(id, resolved, path, 1, make(map[*Term]bool), make(map[string]bool))
	if err != nil {
		return err
	}
	if occurs {
		return invalid("UNIFICATION_OCCURS_CHECK", path, "parameter occurs in its proposed replacement")
	}
	copyValue, err := copyTermDetached(resolved, u.limits)
	if err != nil {
		return err
	}
	u.substitution[id] = copyValue
	return nil
}

func (u *unifier) termTypeKnown(term *Term) Type {
	switch term.Kind {
	case TermParam:
		return u.parameterTypes[term.ID]
	case TermRef:
		return u.referenceTypes[term.ID]
	case TermAtom, TermAnd:
		return Type{Kind: TypeProp}
	default:
		return Type{}
	}
}

func (u *unifier) occurs(id string, term *Term, path string, depth int, active map[*Term]bool, chain map[string]bool) (bool, error) {
	if depth > u.limits.MaxDepth {
		return false, resourceLimit(path, "occurs-check depth exceeds limit")
	}
	if term == nil {
		return false, invalid("MISSING_TERM", path, "occurs-check encountered nil term")
	}
	if active[term] {
		return false, invalid("CYCLIC_TERM", path, "occurs-check found a pointer cycle")
	}
	active[term] = true
	defer delete(active, term)
	if term.Kind == TermParam {
		if term.ID == id {
			return true, nil
		}
		if replacement, ok := u.substitution[term.ID]; ok {
			if chain[term.ID] {
				return false, invalid("UNIFICATION_CYCLE", path, "substitution chain contains a cycle")
			}
			chain[term.ID] = true
			found, err := u.occurs(id, replacement, path, depth+1, active, chain)
			delete(chain, term.ID)
			return found, err
		}
		return false, nil
	}
	if term.Kind == TermAtom {
		for i, arg := range term.Args {
			found, err := u.occurs(id, arg.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1, active, chain)
			if err != nil || found {
				return found, err
			}
		}
	}
	if term.Kind == TermAnd {
		for i, child := range term.Operands {
			found, err := u.occurs(id, child, fmt.Sprintf("%s.operands[%d]", path, i), depth+1, active, chain)
			if err != nil || found {
				return found, err
			}
		}
	}
	return false, nil
}
