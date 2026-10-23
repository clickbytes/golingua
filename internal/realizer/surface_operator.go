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
