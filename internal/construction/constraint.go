package construction

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/features"
)

const (
	MaxPlanConstraints = 512
	MaxConditionDepth  = 16
	MaxConditionNodes  = 256
)

type CompiledConstraintKind string

const (
	ConstraintAgree    CompiledConstraintKind = "agree"
	ConstraintGovern   CompiledConstraintKind = "govern"
	ConstraintRequire  CompiledConstraintKind = "require"
	ConstraintPrecede  CompiledConstraintKind = "precede"
	ConstraintOptional CompiledConstraintKind = "optional"
)

// SourcePath identifies one typed value in the compiled family/frame input.
// Exactly one of SlotID or ArgumentID is set.
type SourcePath struct {
	SlotID     string             `json:"slot_id,omitempty"`
	ArgumentID string             `json:"argument_id,omitempty"`
	Feature    features.FeatureID `json:"feature,omitempty"`
}

type PredicateKind string

const (
	PredicateFeatureEquals PredicateKind = "feature_equals"
	PredicateSlotPresent   PredicateKind = "slot_present"
)

// Predicate is a closed, bounded condition tree. It has no executable text.
type Predicate struct {
	Kind   PredicateKind           `json:"kind"`
	Source SourcePath              `json:"source,omitempty"`
	Value  features.FeatureValueID `json:"value,omitempty"`
	SlotID string                  `json:"slot_id,omitempty"`
}

type CompiledConstraint struct {
	ID           string                 `json:"id"`
	Kind         CompiledConstraintKind `json:"kind"`
	Strength     ConstraintStrength     `json:"strength,omitempty"`
	Target       SourcePath             `json:"target,omitempty"`
	Controller   SourcePath             `json:"controller,omitempty"`
	Features     []features.FeatureID   `json:"features,omitempty"`
	CasePolicy   string                 `json:"case_policy,omitempty"`
	Condition    *Predicate             `json:"condition,omitempty"`
	BeforeSlot   string                 `json:"before_slot,omitempty"`
	AfterSlot    string                 `json:"after_slot,omitempty"`
	OptionalSlot string                 `json:"optional_slot,omitempty"`
}

func validateSourcePath(source SourcePath, slots map[string]SlotSpec, arguments map[string]struct{}, path string) error {
	if (source.SlotID == "") == (source.ArgumentID == "") || !validToken(string(source.Feature)) {
		return invalid("INVALID_CONSTRAINT_SOURCE", path, "source requires exactly one known slot or argument and a feature")
	}
	if source.SlotID != "" {
		if _, exists := slots[source.SlotID]; !exists {
			return unknown("UNKNOWN_CONSTRAINT_SLOT", path+".slot_id", "constraint source slot does not exist")
		}
	}
	if source.ArgumentID != "" {
		if _, exists := arguments[source.ArgumentID]; !exists {
			return unknown("UNKNOWN_CONSTRAINT_ARGUMENT", path+".argument_id", "constraint source argument does not exist")
		}
	}
	return nil
}

func validatePredicate(predicate Predicate, slots map[string]SlotSpec, arguments map[string]struct{}, depth int, budget *int, path string) error {
	(*budget)++
	if *budget > MaxConditionNodes {
		return resourceLimit("condition_nodes", *budget, MaxConditionNodes)
	}
	if depth > MaxConditionDepth {
		return resourceLimit("condition_depth", depth, MaxConditionDepth)
	}
	switch predicate.Kind {
	case PredicateFeatureEquals:
		if predicate.SlotID != "" || predicate.Value == "" {
			return invalid("INVALID_FEATURE_PREDICATE", path, "feature predicate requires only source and explicit value")
		}
		return validateSourcePath(predicate.Source, slots, arguments, path+".source")
	case PredicateSlotPresent:
		if _, exists := slots[predicate.SlotID]; !exists || predicate.Source != (SourcePath{}) || predicate.Value != "" {
			return invalid("INVALID_SLOT_PREDICATE", path, "slot predicate requires only one known slot ID")
		}
		return nil
	default:
		return invalid("INVALID_PREDICATE_KIND", path+".kind", "predicate kind is unsupported")
	}
}

func validateConstraints(family ConstructionFamily, frameArguments []string, constraints []CompiledConstraint) error {
	if len(constraints) > MaxPlanConstraints {
		return resourceLimit("constraints", len(constraints), MaxPlanConstraints)
	}
	slots := make(map[string]SlotSpec, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		slots[slot.ID] = slot
	}
	arguments := make(map[string]struct{}, len(frameArguments))
	for _, id := range frameArguments {
		arguments[id] = struct{}{}
	}
	seen := make(map[string]struct{}, len(constraints))
	for index, constraint := range constraints {
		path := fmt.Sprintf("constraints[%d]", index)
		if !validToken(constraint.ID) {
			return invalid("INVALID_COMPILED_CONSTRAINT_ID", path+".id", "constraint ID must be a bounded token")
		}
		if _, exists := seen[constraint.ID]; exists {
			return invalid("DUPLICATE_COMPILED_CONSTRAINT_ID", path+".id", "constraint ID is duplicated")
		}
		seen[constraint.ID] = struct{}{}
		switch constraint.Kind {
		case ConstraintAgree:
			if constraint.Strength != ConstraintHard && constraint.Strength != ConstraintSoft || len(constraint.Features) == 0 || constraint.CasePolicy != "" || constraint.Condition != nil || constraint.BeforeSlot != "" || constraint.AfterSlot != "" || constraint.OptionalSlot != "" {
				return invalid("INVALID_AGREEMENT_CONSTRAINT", path, "agreement requires strength and feature dimensions only")
			}
			if err := validateSourcePath(constraint.Target, slots, arguments, path+".target"); err != nil {
				return err
			}
			if err := validateSourcePath(constraint.Controller, slots, arguments, path+".controller"); err != nil {
				return err
			}
			features := make(map[features.FeatureID]struct{}, len(constraint.Features))
			for _, feature := range constraint.Features {
				if !validToken(string(feature)) {
					return invalid("INVALID_AGREEMENT_FEATURE", path+".features", "feature must be a bounded token")
				}
				if _, exists := features[feature]; exists {
					return invalid("DUPLICATE_AGREEMENT_FEATURE", path+".features", "agreement feature is duplicated")
				}
				features[feature] = struct{}{}
			}
			if constraint.Target.Feature == "" || constraint.Controller.Feature == "" {
				return invalid("MISSING_AGREEMENT_FEATURE", path, "target and controller paths must each name a feature")
			}
			if constraint.Target == constraint.Controller {
				return invalid("SELF_AGREEMENT", path, "agreement target and controller must be distinct")
			}
		case ConstraintGovern:
			if constraint.Target.SlotID == "" || constraint.Target.ArgumentID != "" || constraint.Target.Feature != "" || constraint.CasePolicy == "" || !validToken(constraint.CasePolicy) || constraint.Controller != (SourcePath{}) || len(constraint.Features) != 0 || constraint.Condition != nil || constraint.BeforeSlot != "" || constraint.AfterSlot != "" || constraint.OptionalSlot != "" || constraint.Strength != "" {
				return invalid("INVALID_GOVERN_CONSTRAINT", path, "govern requires one slot and explicit case policy only")
			}
			if _, ok := slots[constraint.Target.SlotID]; !ok {
				return unknown("UNKNOWN_GOVERNED_SLOT", path+".target.slot_id", "governed slot does not exist")
			}
		case ConstraintRequire:
			if constraint.Condition == nil || constraint.Target != (SourcePath{}) || constraint.Controller != (SourcePath{}) || len(constraint.Features) != 0 || constraint.CasePolicy != "" || constraint.BeforeSlot != "" || constraint.AfterSlot != "" || constraint.OptionalSlot != "" {
				return invalid("INVALID_REQUIRE_CONSTRAINT", path, "require needs one predicate only")
			}
			if constraint.Strength != "" {
				return invalid("INVALID_REQUIRE_CONSTRAINT", path+".strength", "require has no strength")
			}
			budget := 0
			if err := validatePredicate(*constraint.Condition, slots, arguments, 1, &budget, path+".condition"); err != nil {
				return err
			}
		case ConstraintPrecede:
			if constraint.BeforeSlot == "" || constraint.AfterSlot == "" || constraint.BeforeSlot == constraint.AfterSlot || constraint.Target != (SourcePath{}) || constraint.Controller != (SourcePath{}) || len(constraint.Features) != 0 || constraint.CasePolicy != "" || constraint.Condition != nil || constraint.OptionalSlot != "" {
				return invalid("INVALID_PRECEDE_CONSTRAINT", path, "precede requires two distinct slots only")
			}
			if constraint.Strength != "" {
				return invalid("INVALID_PRECEDE_CONSTRAINT", path+".strength", "precedence has no strength")
			}
			if _, ok := slots[constraint.BeforeSlot]; !ok {
				return unknown("UNKNOWN_PRECEDE_SLOT", path+".before_slot", "precedence slot does not exist")
			}
			if _, ok := slots[constraint.AfterSlot]; !ok {
				return unknown("UNKNOWN_PRECEDE_SLOT", path+".after_slot", "precedence slot does not exist")
			}
		case ConstraintOptional:
			if constraint.OptionalSlot == "" || constraint.Target != (SourcePath{}) || constraint.Controller != (SourcePath{}) || len(constraint.Features) != 0 || constraint.CasePolicy != "" || constraint.Condition != nil || constraint.BeforeSlot != "" || constraint.AfterSlot != "" {
				return invalid("INVALID_OPTIONAL_CONSTRAINT", path, "optional requires one slot only")
			}
			slot, ok := slots[constraint.OptionalSlot]
			if !ok {
				return unknown("UNKNOWN_OPTIONAL_SLOT", path+".optional_slot", "optional slot does not exist")
			}
			if constraint.Strength != "" {
				return invalid("INVALID_OPTIONAL_CONSTRAINT", path+".strength", "optional has no strength")
			}
			if slot.Optionality != SlotOptional {
				return unsupported("REQUIRED_SLOT_CANNOT_BE_OPTIONAL", path+".optional_slot", "constraint cannot make a required family slot optional")
			}
		default:
			return invalid("INVALID_COMPILED_CONSTRAINT_KIND", path+".kind", "constraint kind is unsupported")
		}
	}
	if err := validatePrecedenceGraph(constraints, slots); err != nil {
		return err
	}
	return nil
}

func validatePrecedenceGraph(constraints []CompiledConstraint, slots map[string]SlotSpec) error {
	graph := make(map[string][]string)
	indegree := make(map[string]int, len(slots))
	for id := range slots {
		indegree[id] = 0
	}
	for _, constraint := range constraints {
		if constraint.Kind != ConstraintPrecede {
			continue
		}
		graph[constraint.BeforeSlot] = append(graph[constraint.BeforeSlot], constraint.AfterSlot)
		indegree[constraint.AfterSlot]++
	}
	ready := make([]string, 0, len(indegree))
	for id, degree := range indegree {
		if degree == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	visited := 0
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		visited++
		for _, target := range graph[id] {
			indegree[target]--
			if indegree[target] == 0 {
				ready = append(ready, target)
				sort.Strings(ready)
			}
		}
	}
	if visited != len(slots) {
		return unsupported("PRECEDENCE_CYCLE", "constraints", "precedence constraints form a cycle")
	}
	return nil
}
