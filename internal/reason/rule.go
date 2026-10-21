// Package reason evaluates a finite set of declared, positive CORE relation
// rules over one request FactSet. It does not maintain world memory.
package reason

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
)

type RuleKind string

const (
	RuleSymmetry     RuleKind = "symmetry"
	RuleTransitivity RuleKind = "transitivity"
)

type Monotonicity string

const MonotonePositive Monotonicity = "positive"

// Rule activates one safe schema that the pinned CORE relation explicitly
// declares. There is no executable expression or arbitrary rule language.
type Rule struct {
	ID           string       `json:"id"`
	Kind         RuleKind     `json:"kind"`
	Predicate    string       `json:"predicate"`
	Monotonicity Monotonicity `json:"monotonicity"`
}

const MaxRules = 256

func compileRules(ontology *core.ConceptGraph, input []Rule) ([]Rule, error) {
	if len(input) > MaxRules {
		return nil, reasonLimit("rules", len(input))
	}
	result := append([]Rule(nil), input...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	seen := map[string]struct{}{}
	for index, rule := range result {
		path := fmt.Sprintf("rules[%d]", index)
		if !validID(rule.ID) {
			return nil, reasonInvalid("INVALID_RULE_ID", path+".id", "rule ID must be bounded ASCII")
		}
		if _, exists := seen[rule.ID]; exists {
			return nil, reasonInvalid("DUPLICATE_RULE_ID", path+".id", "rule ID is duplicated")
		}
		seen[rule.ID] = struct{}{}
		predicate, err := core.ParseRelationID(rule.Predicate)
		if err != nil || predicate.String() != rule.Predicate {
			return nil, reasonInvalid("INVALID_RULE_PREDICATE", path+".predicate", "rule predicate must be a canonical CORE relation ID")
		}
		relation, exists := ontology.Relation(predicate)
		if !exists {
			return nil, reasonInvalid("UNKNOWN_RULE_PREDICATE", path+".predicate", "rule predicate does not exist in CORE")
		}
		if rule.Monotonicity != MonotonePositive {
			return nil, reasonInvalid("INVALID_RULE_MONOTONICITY", path+".monotonicity", "only positive monotone rules are supported")
		}
		switch rule.Kind {
		case RuleSymmetry:
			if !relation.Symmetric {
				return nil, reasonInvalid("UNDECLARED_SYMMETRY", path+".kind", "CORE relation does not declare symmetry")
			}
		case RuleTransitivity:
			if !relation.Transitive {
				return nil, reasonInvalid("UNDECLARED_TRANSITIVITY", path+".kind", "CORE relation does not declare transitivity")
			}
		default:
			return nil, reasonInvalid("INVALID_RULE_KIND", path+".kind", "rule schema is unsupported")
		}
	}
	return result, nil
}
