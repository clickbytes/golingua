package realizer

import "github.com/clickbytes/golingua/internal/construction"

// explicitSurfaceNegation reports whether this plan's selected family and
// compiled variant carry the effect of exactly one identity negation operator.
// Other operators still require a dedicated surface realization path.
func explicitSurfaceNegation(plan RealizationPlan, family construction.ConstructionFamily) bool {
	if len(plan.Construction.Steps) != 1 || plan.FamilyID != family.ID || plan.Language != family.Language ||
		plan.Construction.InputFamilyID != family.ID || plan.Construction.OutputFamilyID != family.ID {
		return false
	}
	operator := plan.Construction.Steps[0]
	if operator.Kind != construction.OperatorNegation || operator.Value != construction.ValueNegated ||
		operator.InputFamilyID != family.ID || operator.OutputFamilyID != family.ID ||
		operator.SemanticScope.TargetNodeID != plan.NodeID || operator.SemanticScope.ScopeNodeID == "" {
		return false
	}
	if !hasConstructionOperator(family, operator.ID) || !hasConstructionCapability(family, "negation") ||
		!hasConstructionCapability(family, "surface_negation") {
		return false
	}
	return len(plan.CompiledVariant.Variant.Sequence) > 0
}

// explicitSurfaceQuestion reports the one request-scoped WH operator whose
// surface effect is already carried by the selected language construction
// sequence. Earlier plan and Execute validation bind the operator to one
// unique Meaning question wrapper and revalidate its exact answer role.
func explicitSurfaceQuestion(plan RealizationPlan, family construction.ConstructionFamily) bool {
	answer := plan.QuestionAnswer
	if answer == nil || answer.VariableID == "" || answer.Sort != "Entity" || answer.Role == "" ||
		answer.TargetNodeID == "" || answer.TargetNodeID != plan.NodeID ||
		len(plan.Construction.Steps) != 1 || plan.FamilyID != family.ID || plan.Language != family.Language ||
		plan.Construction.InputFamilyID != family.ID || plan.Construction.OutputFamilyID != family.ID {
		return false
	}
	operator := plan.Construction.Steps[0]
	if operator.Kind != construction.OperatorQuestion || operator.Value != construction.ValueWH ||
		operator.InputFamilyID != family.ID || operator.OutputFamilyID != family.ID ||
		operator.SemanticScope.TargetNodeID != plan.NodeID || operator.SemanticScope.ScopeNodeID == "" ||
		!hasConstructionOperator(family, operator.ID) || !hasConstructionCapability(family, "question-wh") {
		return false
	}
	roleSlots := make(map[string]string, len(family.SlotSpecs))
	for _, slot := range family.SlotSpecs {
		if slot.ArgumentBinding == nil {
			continue
		}
		roleSlots[string(slot.ArgumentBinding.SemanticRole)] = slot.ID
	}
	answerSlotID, exists := roleSlots[string(answer.Role)]
	if !exists {
		return false
	}
	var answerSlot *construction.SlotSpec
	for index := range family.SlotSpecs {
		if family.SlotSpecs[index].ID == answerSlotID {
			answerSlot = &family.SlotSpecs[index]
			break
		}
	}
	if answerSlot == nil || answerSlot.Category != construction.CategoryNP ||
		answerSlot.Optionality != construction.SlotOptional || answerSlot.Cardinality.Minimum != 0 || answerSlot.Cardinality.Maximum != 1 {
		return false
	}
	sequenceSlots := make(map[string]struct{}, len(plan.CompiledVariant.Variant.Sequence))
	for _, token := range plan.CompiledVariant.Variant.Sequence {
		if token.Kind == construction.TokenSlot {
			sequenceSlots[token.SlotID] = struct{}{}
		}
	}
	if _, included := sequenceSlots[answerSlotID]; included {
		return false
	}
	for _, role := range plan.Roles {
		slotID, bound := roleSlots[string(role)]
		if !bound {
			return false
		}
		if _, included := sequenceSlots[slotID]; !included {
			return false
		}
	}
	return len(plan.CompiledVariant.Variant.Sequence) > 0
}

func hasConstructionOperator(family construction.ConstructionFamily, wanted construction.OperatorID) bool {
	for _, operator := range family.Operators {
		if operator == wanted {
			return true
		}
	}
	return false
}

func hasConstructionCapability(family construction.ConstructionFamily, wanted construction.Capability) bool {
	for _, capability := range family.Capabilities {
		if capability == wanted {
			return true
		}
	}
	return false
}
