package construction

import "github.com/clickbytes/golingua/internal/valency"

// BindFrame specializes a reusable family for one explicitly authored frame.
// ArgumentSlot.ID must equal the construction slot ID: no semantic role is
// inferred from subject/object position or from the CORE concept.
func BindFrame(record FamilyRecord, frame valency.ValencyFrame) (FamilyRecord, error) {
	if err := record.Family.Validate(); err != nil {
		return FamilyRecord{}, err
	}
	if err := frame.Validate(); err != nil {
		return FamilyRecord{}, err
	}
	if !frameAllowsFamily(frame, record.Family.ID) {
		return FamilyRecord{}, unsupported("FRAME_FAMILY_NOT_LICENSED", "family.id", "selected frame does not license this construction family")
	}
	family := cloneFamily(record.Family)
	arguments := make(map[string]valency.ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		arguments[argument.ID] = argument
	}
	bound := make(map[string]bool, len(frame.Arguments))
	for index := range family.SlotSpecs {
		slot := &family.SlotSpecs[index]
		argument, exists := arguments[slot.ID]
		if !exists {
			if slot.ArgumentBinding != nil {
				return FamilyRecord{}, unsupported("FAMILY_SLOT_WITHOUT_FRAME_ARGUMENT", "family.slot_specs."+slot.ID, "bound construction slot has no exact frame argument")
			}
			continue
		}
		if slot.Category != CategoryNP && slot.Category != CategoryPP {
			return FamilyRecord{}, unsupported("ARGUMENT_SLOT_CATEGORY_UNSUPPORTED", "family.slot_specs."+slot.ID, "frame argument requires an NP or PP construction slot")
		}
		if slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole != argument.Role {
			return FamilyRecord{}, unsupported("FRAME_FAMILY_ROLE_CONFLICT", "family.slot_specs."+slot.ID, "construction role conflicts with the exact frame argument")
		}
		if slot.Case != "" && argument.Case != "" && slot.Case != argument.Case {
			return FamilyRecord{}, unsupported("CASE_POLICY_CONFLICT", "family.slot_specs."+slot.ID, "construction and frame case policies conflict")
		}
		if slot.Adposition != "" && argument.Adposition != "" && slot.Adposition != argument.Adposition {
			return FamilyRecord{}, unsupported("ADPOSITION_POLICY_CONFLICT", "family.slot_specs."+slot.ID, "construction and frame adposition policies conflict")
		}
		slot.ArgumentBinding = &ArgumentBinding{SemanticRole: argument.Role}
		if slot.Case == "" {
			slot.Case = argument.Case
		}
		if slot.Adposition == "" {
			slot.Adposition = argument.Adposition
		}
		bound[argument.ID] = true
	}
	for _, argument := range frame.Arguments {
		if argument.Required && !bound[argument.ID] {
			return FamilyRecord{}, unsupported("REQUIRED_ARGUMENT_UNBOUND", "frame.arguments."+argument.ID, "required frame argument has no construction slot with the same ID")
		}
	}
	if err := family.Validate(); err != nil {
		return FamilyRecord{}, err
	}
	return FamilyRecord{Family: family, PlanID: planID(family)}, nil
}
