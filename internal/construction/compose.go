package construction

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// ConstructionPlan is one bounded, ordered composition. It represents one
// path only; Compose never materializes a cartesian product of families.
type ConstructionPlan struct {
	ID             ConstructionPlanID  `json:"id"`
	Language       features.LanguageID `json:"language"`
	SenseID        ids.SenseID         `json:"sense_id"`
	ValencyFrameID valency.FrameID     `json:"valency_frame_id"`
	InputFamilyID  ids.ConstructionID  `json:"input_family_id"`
	OutputFamilyID ids.ConstructionID  `json:"output_family_id"`
	Steps          []Operator          `json:"steps"`
}

// Compose applies an explicitly ordered list of licensed operators to one
// initial family and one Meaning graph. The provided order is preserved; no
// commutativity or alternate combinations are inferred.
func Compose(registry *Registry, frame valency.ValencyFrame, initialFamilyID ids.ConstructionID, semanticMeaning meaning.Meaning, operators []Operator) (ConstructionPlan, error) {
	if registry == nil {
		return ConstructionPlan{}, unknown("CONSTRUCTION_REGISTRY_UNAVAILABLE", "registry", "construction registry is unavailable")
	}
	if err := frame.Validate(); err != nil {
		return ConstructionPlan{}, err
	}
	if len(operators) > MaxOperatorsPerPlan {
		return ConstructionPlan{}, resourceLimit("operators", len(operators), MaxOperatorsPerPlan)
	}
	initial, exists := registry.ByID(initialFamilyID)
	if !exists {
		return ConstructionPlan{}, unknown("UNKNOWN_CONSTRUCTION_FAMILY", "input_family_id", "initial construction family is not registered")
	}
	if !frameAllowsFamily(frame, initial.Family.ID) {
		return ConstructionPlan{}, unsupported("FRAME_FAMILY_NOT_LICENSED", "input_family_id", "selected sense frame does not license the initial family")
	}
	initial, err := BindFrame(initial, frame)
	if err != nil {
		return ConstructionPlan{}, err
	}
	if err := validateFamilyFrameRoles(initial.Family, frame); err != nil {
		return ConstructionPlan{}, err
	}
	if len(operators) > 0 {
		if err := meaning.ValidateStructure(semanticMeaning, meaning.DefaultBudget()); err != nil {
			return ConstructionPlan{}, err
		}
	}
	positions := make(map[OperatorID]int, len(operators))
	for index, operator := range operators {
		if err := operator.Validate(); err != nil {
			return ConstructionPlan{}, err
		}
		if _, exists := positions[operator.ID]; exists {
			return ConstructionPlan{}, invalid("DUPLICATE_OPERATOR_ID", "operators", "operator IDs must be unique within one composition")
		}
		positions[operator.ID] = index
	}
	current := initial
	steps := make([]Operator, 0, len(operators))
	operatorStates := make(map[string]OperatorValue, len(operators))
	for index, operator := range operators {
		path := "operators[" + strconv.Itoa(index) + "]"
		if operator.InputFamilyID != current.Family.ID {
			return ConstructionPlan{}, unsupported("OPERATOR_INPUT_FAMILY_MISMATCH", path+".input_family_id", "operator input family does not match the current plan family")
		}
		if !containsOperator(current.Family.Operators, operator.ID) {
			return ConstructionPlan{}, unsupported("OPERATOR_NOT_LICENSED", path+".id", "current family does not license this operator")
		}
		for _, capability := range operator.Capabilities {
			if !containsCapability(current.Family.Capabilities, capability) {
				return ConstructionPlan{}, unsupported("OPERATOR_CAPABILITY_UNSUPPORTED", path+".capabilities", "input family lacks a required operator capability")
			}
		}
		output, exists := registry.ByID(operator.OutputFamilyID)
		if !exists {
			return ConstructionPlan{}, unknown("UNKNOWN_OPERATOR_OUTPUT_FAMILY", path+".output_family_id", "operator output family is not registered")
		}
		if output.Family.Language != initial.Family.Language {
			return ConstructionPlan{}, unsupported("OPERATOR_LANGUAGE_MISMATCH", path+".output_family_id", "operator cannot compose families across languages")
		}
		if !frameAllowsFamily(frame, output.Family.ID) {
			return ConstructionPlan{}, unsupported("FRAME_FAMILY_NOT_LICENSED", path+".output_family_id", "selected sense frame does not license the operator output family")
		}
		output, err = BindFrame(output, frame)
		if err != nil {
			return ConstructionPlan{}, err
		}
		if err := validateFamilyFrameRoles(output.Family, frame); err != nil {
			return ConstructionPlan{}, err
		}
		for _, capability := range operator.Capabilities {
			if !containsCapability(output.Family.Capabilities, capability) {
				return ConstructionPlan{}, unsupported("OUTPUT_OPERATOR_CAPABILITY_UNSUPPORTED", path+".capabilities", "output family lacks a required operator capability")
			}
		}
		if err := validateOperatorScope(operator, semanticMeaning); err != nil {
			return ConstructionPlan{}, err
		}
		for _, dependency := range operator.Dependencies {
			dependencyPosition, exists := positions[dependency]
			if !exists {
				return ConstructionPlan{}, incomplete("MISSING_OPERATOR_DEPENDENCY", path+".dependencies", "required operator dependency is absent from the composition")
			}
			if dependencyPosition >= index {
				return ConstructionPlan{}, unsupported("OPERATOR_ORDER_CONFLICT", path+".dependencies", "operator dependency must occur earlier in the explicit composition order")
			}
		}
		stateKey := string(operator.Kind) + "\x00" + string(operator.SemanticScope.TargetNodeID)
		if previous, exists := operatorStates[stateKey]; exists {
			if previous != operator.Value {
				return ConstructionPlan{}, unsupported("CONFLICTING_OPERATOR_STATE", path, "composition assigns incompatible values to the same semantic target")
			}
			return ConstructionPlan{}, invalid("DUPLICATE_OPERATOR_STATE", path, "composition repeats the same operator state for one target")
		}
		operatorStates[stateKey] = operator.Value
		steps = append(steps, cloneOperator(operator))
		current = output
	}
	return ConstructionPlan{
		ID:             composedPlanID(initial, current, frame, steps),
		Language:       initial.Family.Language,
		SenseID:        frame.SenseID,
		ValencyFrameID: frame.ID,
		InputFamilyID:  initial.Family.ID,
		OutputFamilyID: current.Family.ID,
		Steps:          steps,
	}, nil
}

func frameAllowsFamily(frame valency.ValencyFrame, familyID ids.ConstructionID) bool {
	for _, allowed := range frame.AllowedFamilyIDs {
		if allowed == familyID {
			return true
		}
	}
	return false
}

func validateFamilyFrameRoles(family ConstructionFamily, frame valency.ValencyFrame) error {
	frameRoles := make(map[string]valency.ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		frameRoles[string(argument.Role)] = argument
	}
	familyRoles := make(map[string]struct{})
	for _, slot := range family.SlotSpecs {
		if slot.ArgumentBinding == nil {
			continue
		}
		role := string(slot.ArgumentBinding.SemanticRole)
		if _, exists := frameRoles[role]; !exists {
			return unsupported("FAMILY_ROLE_NOT_LICENSED_BY_FRAME", "family.slot_specs."+slot.ID, "family binds a role not licensed by the selected sense frame")
		}
		familyRoles[role] = struct{}{}
	}
	for role, argument := range frameRoles {
		if argument.Required {
			if _, exists := familyRoles[role]; !exists {
				return unsupported("REQUIRED_FRAME_ROLE_UNBOUND", "family.slot_specs", "family has no slot for a required role in the selected sense frame")
			}
		}
	}
	return nil
}

func containsOperator(values []OperatorID, candidate OperatorID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func composedPlanID(initial, output FamilyRecord, frame valency.ValencyFrame, steps []Operator) ConstructionPlanID {
	var builder strings.Builder
	writePlanPart(&builder, "compose-v1")
	writePlanPart(&builder, string(initial.PlanID))
	writePlanPart(&builder, string(output.PlanID))
	writePlanPart(&builder, frame.SenseID.String())
	writePlanPart(&builder, string(frame.ID))
	writePlanPart(&builder, strconv.Itoa(len(steps)))
	for _, operator := range steps {
		writePlanPart(&builder, string(operator.ID))
		writePlanPart(&builder, string(operator.Kind))
		writePlanPart(&builder, string(operator.Value))
		writePlanPart(&builder, operator.InputFamilyID.String())
		writePlanPart(&builder, operator.OutputFamilyID.String())
		writePlanPart(&builder, string(operator.SemanticScope.TargetNodeID))
		writePlanPart(&builder, string(operator.SemanticScope.ScopeNodeID))
		dependencies := append([]OperatorID(nil), operator.Dependencies...)
		sort.Slice(dependencies, func(i, j int) bool { return dependencies[i] < dependencies[j] })
		writePlanPart(&builder, strconv.Itoa(len(dependencies)))
		for _, dependency := range dependencies {
			writePlanPart(&builder, string(dependency))
		}
		capabilities := operatorCapabilityList(operator)
		writePlanPart(&builder, strconv.Itoa(len(capabilities)))
		for _, capability := range capabilities {
			writePlanPart(&builder, string(capability))
		}
	}
	digest := sha256.Sum256([]byte(builder.String()))
	return ConstructionPlanID("plan:" + hex.EncodeToString(digest[:]))
}
