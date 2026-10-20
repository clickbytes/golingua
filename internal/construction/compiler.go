package construction

import (
	"sort"
	"strconv"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	MaxCompiledPlans = 100_000
	MaxCompileItems  = 1_000_000
)

type ConstructionProfile struct {
	Language   features.LanguageID         `json:"language"`
	Version    features.ProfileVersion     `json:"version"`
	Dimensions []features.FeatureDimension `json:"dimensions"`
	Provenance provenance.SourceRef        `json:"provenance"`
}

type CompileInput struct {
	Family       ConstructionFamily
	Frame        valency.ValencyFrame
	Profile      ConstructionProfile
	Variant      Variant
	Constraints  []CompiledConstraint
	ProcedureIDs []string
}

type PlanIndex struct {
	plans       []CompiledConstructionPlan
	byID        map[ConstructionPlanID]int
	bySignature map[string]ConstructionPlanID
}

func CompileConstruction(input CompileInput) (CompiledConstructionPlan, error) {
	family, frame, profile, variant := input.Family, input.Frame, input.Profile, input.Variant
	if err := family.Validate(); err != nil {
		return CompiledConstructionPlan{}, err
	}
	if err := frame.Validate(); err != nil {
		return CompiledConstructionPlan{}, err
	}
	if err := profile.Provenance.Validate(); err != nil {
		return CompiledConstructionPlan{}, invalid("INVALID_PROFILE_PROVENANCE", "profile.provenance", "profile requires valid provenance")
	}
	if !validLanguage(profile.Language) || !validToken(string(profile.Version)) {
		return CompiledConstructionPlan{}, invalid("INVALID_CONSTRUCTION_PROFILE", "profile", "profile requires canonical language and version")
	}
	if profile.Language != family.Language {
		return CompiledConstructionPlan{}, unsupported("PROFILE_LANGUAGE_MISMATCH", "profile.language", "profile and family languages must match")
	}
	if profile.Version == "" {
		return CompiledConstructionPlan{}, invalid("INVALID_CONSTRUCTION_PROFILE", "profile.version", "profile version is required")
	}
	if len(profile.Dimensions) == 0 || len(profile.Dimensions) > features.MaxDimensions {
		return CompiledConstructionPlan{}, invalid("INVALID_CONSTRUCTION_PROFILE", "profile.dimensions", "profile must contain a bounded feature inventory")
	}
	profileDimensions := make(map[features.FeatureID]map[features.FeatureValueID]struct{}, len(profile.Dimensions))
	for _, dimension := range profile.Dimensions {
		if !validToken(string(dimension.ID)) {
			return CompiledConstructionPlan{}, invalid("INVALID_PROFILE_FEATURE", "profile.dimensions", "feature ID must be a bounded token")
		}
		if _, exists := profileDimensions[dimension.ID]; exists {
			return CompiledConstructionPlan{}, invalid("DUPLICATE_PROFILE_FEATURE", "profile.dimensions", "feature dimension is duplicated")
		}
		if len(dimension.Values) == 0 || len(dimension.Values) > features.MaxValuesPerDim {
			return CompiledConstructionPlan{}, invalid("INVALID_PROFILE_FEATURE_VALUES", "profile.dimensions."+string(dimension.ID), "feature dimension requires bounded values")
		}
		values := make(map[features.FeatureValueID]struct{}, len(dimension.Values))
		for _, value := range dimension.Values {
			if !validToken(string(value)) {
				return CompiledConstructionPlan{}, invalid("INVALID_PROFILE_FEATURE_VALUE", "profile.dimensions."+string(dimension.ID), "feature value must be a bounded token")
			}
			if _, exists := values[value]; exists {
				return CompiledConstructionPlan{}, invalid("DUPLICATE_PROFILE_FEATURE_VALUE", "profile.dimensions."+string(dimension.ID), "feature value is duplicated")
			}
			values[value] = struct{}{}
		}
		profileDimensions[dimension.ID] = values
	}
	compiled, err := ValidateVariantForFrame(variant, family, frame)
	if err != nil {
		return CompiledConstructionPlan{}, err
	}
	if err := validateConstraints(family, argumentIDs(frame), input.Constraints); err != nil {
		return CompiledConstructionPlan{}, err
	}
	for i, constraint := range input.Constraints {
		featuresToCheck := append([]features.FeatureID(nil), constraint.Features...)
		if constraint.Kind == ConstraintAgree {
			featuresToCheck = append(featuresToCheck, constraint.Target.Feature, constraint.Controller.Feature)
		}
		if constraint.Kind == ConstraintRequire && constraint.Condition != nil && constraint.Condition.Kind == PredicateFeatureEquals {
			featuresToCheck = append(featuresToCheck, constraint.Condition.Source.Feature)
		}
		for _, feature := range featuresToCheck {
			if _, exists := profileDimensions[feature]; !exists {
				return CompiledConstructionPlan{}, unknown("UNKNOWN_PROFILE_FEATURE", "constraints["+strconv.Itoa(i)+"].features", "agreement feature is not declared by the language profile")
			}
		}
		if constraint.Kind == ConstraintRequire && constraint.Condition != nil && constraint.Condition.Kind == PredicateFeatureEquals {
			values := profileDimensions[constraint.Condition.Source.Feature]
			if _, ok := values[constraint.Condition.Value]; !ok {
				return CompiledConstructionPlan{}, unknown("UNKNOWN_PROFILE_FEATURE_VALUE", "constraints["+strconv.Itoa(i)+"].condition.value", "predicate value is not declared by the language profile")
			}
		}
	}
	if len(input.ProcedureIDs) > MaxFamilyItems {
		return CompiledConstructionPlan{}, resourceLimit("procedure_ids", len(input.ProcedureIDs), MaxFamilyItems)
	}
	procedureIDs := append([]string(nil), input.ProcedureIDs...)
	seenProcedures := map[string]struct{}{}
	for i, id := range procedureIDs {
		id = normalizeProcedureID(id)
		procedureIDs[i] = id
		if !validToken(id) {
			return CompiledConstructionPlan{}, invalid("INVALID_PROCEDURE_ID", "procedure_ids", "procedure ID must be a bounded token")
		}
		if _, ok := seenProcedures[id]; ok {
			return CompiledConstructionPlan{}, invalid("DUPLICATE_PROCEDURE_ID", "procedure_ids", "procedure ID is duplicated")
		}
		seenProcedures[id] = struct{}{}
	}
	slots := make(map[string]SlotSpec, len(family.SlotSpecs))
	roleSlots := make(map[string]string)
	for _, slot := range family.SlotSpecs {
		slots[slot.ID] = slot
		if slot.ArgumentBinding != nil {
			roleSlots[string(slot.ArgumentBinding.SemanticRole)] = slot.ID
		}
	}
	bindings := make([]CompiledArgumentBinding, 0, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		slotID, ok := roleSlots[string(argument.Role)]
		if !ok {
			if argument.Required {
				return CompiledConstructionPlan{}, unsupported("REQUIRED_ARGUMENT_UNBOUND", "family.slot_specs", "required frame argument has no construction slot")
			}
			continue
		}
		slot := slots[slotID]
		if slot.Case != "" && argument.Case != "" && slot.Case != argument.Case {
			return CompiledConstructionPlan{}, unsupported("CASE_POLICY_CONFLICT", "family.slot_specs."+slotID, "family case and valency case policies conflict")
		}
		if slot.Adposition != "" && argument.Adposition != "" && slot.Adposition != argument.Adposition {
			return CompiledConstructionPlan{}, unsupported("ADPOSITION_POLICY_CONFLICT", "family.slot_specs."+slotID, "family and valency adposition policies conflict")
		}
		policy := argument.Case
		if policy == "" {
			policy = slot.Case
		}
		adposition := argument.Adposition
		if adposition == "" {
			adposition = slot.Adposition
		}
		bindings = append(bindings, CompiledArgumentBinding{ArgumentID: argument.ID, Role: string(argument.Role), Required: argument.Required, SlotID: slotID, CasePolicy: policy, Adposition: adposition})
	}
	plan := CompiledConstructionPlan{Language: string(family.Language), ProfileVersion: profile.Version, ProfileProvenance: profile.Provenance, Permissions: []SensePermission{{SenseID: frame.SenseID, FrameID: frame.ID}}, FamilyID: family.ID, VariantID: compiled.Variant.ID, ArgumentBindings: bindings, Constraints: cloneConstraints(input.Constraints), Linearization: append([]SequenceToken(nil), compiled.Variant.Sequence...), ProcedureIDs: procedureIDs}
	plan.ID = canonicalPlanID(plan)
	plan.PermissionDigest = permissionDigest(plan.Permissions)
	return plan, nil
}

func argumentIDs(frame valency.ValencyFrame) []string {
	result := make([]string, 0, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		result = append(result, argument.ID)
	}
	return result
}
func cloneConstraints(values []CompiledConstraint) []CompiledConstraint {
	result := append([]CompiledConstraint(nil), values...)
	for i := range result {
		result[i].Features = append([]features.FeatureID(nil), values[i].Features...)
		if values[i].Condition != nil {
			condition := *values[i].Condition
			result[i].Condition = &condition
		}
	}
	return result
}

func CompilePlanIndex(inputs []CompileInput) (*PlanIndex, error) {
	if len(inputs) == 0 {
		return nil, invalid("EMPTY_PLAN_INPUTS", "inputs", "compiler requires at least one construction input")
	}
	if len(inputs) > MaxCompiledPlans {
		return nil, resourceLimit("plans", len(inputs), MaxCompiledPlans)
	}
	itemCount := 0
	for _, input := range inputs {
		itemCount += len(input.Family.SlotSpecs) + len(input.Family.Constraints) + len(input.Frame.Arguments) + len(input.Constraints) + len(input.Variant.Sequence) + len(input.ProcedureIDs)
		if itemCount > MaxCompileItems {
			return nil, resourceLimit("compile_items", itemCount, MaxCompileItems)
		}
	}
	plans := make([]CompiledConstructionPlan, 0, len(inputs))
	byID := make(map[ConstructionPlanID]int, len(inputs))
	bySignature := make(map[string]ConstructionPlanID, len(inputs))
	for _, input := range inputs {
		plan, err := CompileConstruction(input)
		if err != nil {
			return nil, err
		}
		signature := string(plan.ID)
		if position, ok := byID[plan.ID]; ok {
			plans[position].Permissions = append(plans[position].Permissions, plan.Permissions...)
			continue
		}
		bySignature[signature] = plan.ID
		byID[plan.ID] = len(plans)
		plans = append(plans, plan)
	}
	for i := range plans {
		sort.Slice(plans[i].Permissions, func(a, b int) bool {
			left, right := plans[i].Permissions[a], plans[i].Permissions[b]
			if left.SenseID != right.SenseID {
				return left.SenseID.String() < right.SenseID.String()
			}
			return left.FrameID < right.FrameID
		})
		plans[i].Permissions = deduplicatePermissions(plans[i].Permissions)
		plans[i].PermissionDigest = permissionDigest(plans[i].Permissions)
	}
	sort.Slice(plans, func(i, j int) bool { return plans[i].ID < plans[j].ID })
	byID = make(map[ConstructionPlanID]int, len(plans))
	for i, plan := range plans {
		byID[plan.ID] = i
	}
	return &PlanIndex{plans: plans, byID: byID, bySignature: bySignature}, nil
}

func (index *PlanIndex) ByID(id ConstructionPlanID) (CompiledConstructionPlan, bool) {
	if index == nil {
		return CompiledConstructionPlan{}, false
	}
	position, ok := index.byID[id]
	if !ok {
		return CompiledConstructionPlan{}, false
	}
	return clonePlan(index.plans[position]), true
}
func (index *PlanIndex) Plans() []CompiledConstructionPlan {
	if index == nil {
		return []CompiledConstructionPlan{}
	}
	result := make([]CompiledConstructionPlan, len(index.plans))
	for i, plan := range index.plans {
		result[i] = clonePlan(plan)
	}
	return result
}
func clonePlan(plan CompiledConstructionPlan) CompiledConstructionPlan {
	result := plan
	result.Permissions = append([]SensePermission(nil), plan.Permissions...)
	result.ArgumentBindings = append([]CompiledArgumentBinding(nil), plan.ArgumentBindings...)
	result.Constraints = cloneConstraints(plan.Constraints)
	result.Linearization = append([]SequenceToken(nil), plan.Linearization...)
	result.ProcedureIDs = append([]string(nil), plan.ProcedureIDs...)
	return result
}

func deduplicatePermissions(values []SensePermission) []SensePermission {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func (index *PlanIndex) SignatureCount() int {
	if index == nil {
		return 0
	}
	return len(index.bySignature)
}
