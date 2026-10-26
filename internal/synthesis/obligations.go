package synthesis

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
)

type atomOccurrence struct {
	symbol string
	path   string
}

// mapConstructionFragments attaches each validated constraint to the exact
// Atom occurrence in the final checked representation. This is provenance
// mapping, not an independent proof or semantic match.
func mapConstructionFragments(goal GoalSpec, expression *algebra.Term, existing []GoalObligation) ([]GoalObligation, error) {
	limits, err := NormalizeGoalLimits(goal.Limits)
	if err != nil {
		return nil, err
	}
	if expression == nil {
		return nil, goalInvalid("MISSING_CONSTRUCTED_EXPRESSION", "expression", "fragment mapping requires the final checked expression")
	}
	atoms := make([]atomOccurrence, 0, limits.MaxAtoms)
	active := make(map[*algebra.Term]bool)
	visited := 0
	var walk func(*algebra.Term, string, int) error
	walk = func(term *algebra.Term, path string, depth int) error {
		if term == nil {
			return goalInvalid("INVALID_CONSTRUCTED_TREE", path, "constructed expression contains a missing term")
		}
		if depth > limits.TermLimits.MaxDepth {
			return goalResource(path, "constructed expression exceeds the fragment traversal depth")
		}
		if active[term] {
			return goalInvalid("INVALID_CONSTRUCTED_TREE", path, "constructed expression contains a pointer cycle")
		}
		visited++
		if visited > limits.TermLimits.MaxNodes {
			return goalResource(path, "constructed expression exceeds the fragment traversal node budget")
		}
		active[term] = true
		defer delete(active, term)
		switch term.Kind {
		case algebra.TermAtom:
			if len(atoms) >= limits.MaxAtoms {
				return goalResource(path, "constructed atom count exceeds the goal atom budget")
			}
			atoms = append(atoms, atomOccurrence{symbol: term.Symbol, path: path})
			for i, argument := range term.Args {
				if err := walk(argument.Value, fmt.Sprintf("%s.args[%d].value", path, i), depth+1); err != nil {
					return err
				}
			}
		case algebra.TermAnd:
			for i, operand := range term.Operands {
				if err := walk(operand, fmt.Sprintf("%s.operands[%d]", path, i), depth+1); err != nil {
					return err
				}
			}
		case algebra.TermExists, algebra.TermLambda:
			return walk(term.Body, path+".body", depth+1)
		case algebra.TermApply:
			if err := walk(term.Function, path+".function", depth+1); err != nil {
				return err
			}
			return walk(term.Argument, path+".argument", depth+1)
		case algebra.TermVar, algebra.TermParam, algebra.TermRef, algebra.TermHole:
			return nil
		default:
			return goalInvalid("UNKNOWN_CONSTRUCTED_TERM", path+".kind", "constructed expression contains an unsupported term kind")
		}
		return nil
	}
	if err := walk(expression, "expression", 1); err != nil {
		return nil, err
	}
	pathConstraint, err := uniquePath(goal.Constraints)
	if err != nil {
		return nil, err
	}
	if len(pathConstraint.AllowedAtomCounts) != 1 {
		return nil, goalInvalid("INVALID_PATH_COUNT", "constraints", "fragment mapping requires one selected path atom count")
	}
	pathCount := pathConstraint.AllowedAtomCounts[0]
	if pathCount <= 0 || pathCount > limits.MaxAtoms {
		return nil, goalInvalid("INVALID_PATH_COUNT", "constraints", "path atom count is outside the validated goal bounds")
	}
	unaryConstraints := make([]Constraint, 0, len(goal.Constraints))
	for _, constraint := range goal.Constraints {
		if constraint.Kind == ConstraintUnary {
			unaryConstraints = append(unaryConstraints, constraint)
		}
	}
	if len(atoms) != pathCount+len(unaryConstraints) {
		return nil, goalInvalid("CONSTRUCTED_ATOM_COUNT_MISMATCH", "expression", "atom occurrences do not match the validated path and unary constraints")
	}
	for i := 0; i < pathCount; i++ {
		if atoms[i].symbol != pathConstraint.Symbol {
			return nil, goalInvalid("CONSTRUCTED_PATH_SYMBOL_MISMATCH", atoms[i].path+".symbol", "path occurrence symbol differs from its source constraint")
		}
	}
	for i, constraint := range unaryConstraints {
		if atoms[pathCount+i].symbol != constraint.Symbol {
			return nil, goalInvalid("CONSTRUCTED_UNARY_SYMBOL_MISMATCH", atoms[pathCount+i].path+".symbol", "unary occurrence symbol differs from its source constraint")
		}
	}

	byID := make(map[string]GoalObligation, len(existing))
	for _, obligation := range existing {
		if obligation.ConstraintID == "" {
			return nil, goalInvalid("INVALID_CONSTRUCTION_OBLIGATION", "obligations", "obligation is missing its source constraint ID")
		}
		if _, exists := byID[obligation.ConstraintID]; exists {
			return nil, goalInvalid("DUPLICATE_CONSTRUCTION_OBLIGATION", "obligations", "constraint obligation is duplicated")
		}
		byID[obligation.ConstraintID] = obligation
	}
	if len(byID) != len(goal.Constraints) {
		return nil, goalInvalid("CONSTRUCTION_OBLIGATION_COVERAGE", "obligations", "each goal constraint must have exactly one construction obligation")
	}
	result := make([]GoalObligation, 0, len(goal.Constraints))
	for _, constraint := range goal.Constraints {
		obligation, exists := byID[constraint.ID]
		if !exists || obligation.Kind != constraint.Kind {
			return nil, goalInvalid("CONSTRUCTION_OBLIGATION_COVERAGE", "obligations", "obligation IDs and kinds must match the source constraints")
		}
		obligation.StepIDs = append([]string(nil), obligation.StepIDs...)
		obligation.Fragments = []string{}
		if constraint.Kind == ConstraintPath {
			for i := 0; i < pathCount; i++ {
				obligation.Fragments = append(obligation.Fragments, atoms[i].path)
			}
		} else {
			for i, unary := range unaryConstraints {
				if unary.ID == constraint.ID {
					obligation.Fragments = append(obligation.Fragments, atoms[pathCount+i].path)
					break
				}
			}
			if len(obligation.Fragments) != 1 {
				return nil, goalInvalid("CONSTRUCTION_OBLIGATION_COVERAGE", "obligations", "each unary constraint must map to one distinct Atom occurrence")
			}
		}
		obligation.Status = "constructed_provenance"
		result = append(result, obligation)
	}
	return result, nil
}
