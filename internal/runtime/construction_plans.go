package runtime

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

// CompileSentencePlans compiles only already approved exact sense/frame/family/
// variant routes. The frame's authored argument IDs bind reusable family slots
// to semantic roles; a shared family never imposes one sense's roles on another.
func CompileSentencePlans(profile *features.LanguageFeatureProfile, packages []construction.CompiledConstructionPackage, valencyIndex *valency.Index) (*parser.Index, map[ids.ConstructionID]*construction.VariantIndex, error) {
	if profile == nil || valencyIndex == nil {
		return nil, nil, fmt.Errorf("sentence plans require a feature profile and valency index")
	}
	byFamily := make(map[ids.ConstructionID]construction.CompiledConstructionPackage, len(packages))
	variants := make(map[ids.ConstructionID]*construction.VariantIndex, len(packages))
	for _, pkg := range packages {
		family := pkg.Package.Family
		if family.Language != profile.Language() || pkg.Index == nil {
			return nil, nil, fmt.Errorf("construction package has no matching language or variant index: %s", family.ID)
		}
		if _, duplicate := byFamily[family.ID]; duplicate {
			return nil, nil, fmt.Errorf("duplicate sentence construction family: %s", family.ID)
		}
		byFamily[family.ID] = pkg
		variants[family.ID] = pkg.Index
	}
	frames := make(map[valency.FrameID]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[frame.ID] = frame
	}
	profileSource := provenance.SourceRef{
		ID: "golingua-compiled-sentence-profile", Path: "internal/runtime/construction_plans.go",
		Version: "2026-10-23", License: "internal", Author: "golingua contributors",
	}
	inputs := make([]parser.PlanInput, 0, len(valencyIndex.Bindings()))
	for _, binding := range valencyIndex.Bindings() {
		if !binding.Approved {
			continue
		}
		frame, exists := frames[binding.ValencyFrameID]
		if !exists || frame.SenseID != binding.SenseID {
			return nil, nil, fmt.Errorf("approved binding has no exact valency frame: %s", binding.ValencyFrameID)
		}
		pkg, exists := byFamily[binding.FamilyID]
		if !exists {
			return nil, nil, fmt.Errorf("approved binding has no construction family: %s", binding.FamilyID)
		}
		family, err := construction.BindFrame(construction.FamilyRecord{Family: pkg.Package.Family}, frame)
		if err != nil {
			return nil, nil, fmt.Errorf("bind construction %s to %s: %w", binding.FamilyID, frame.ID, err)
		}
		var variant construction.Variant
		for _, candidate := range pkg.Package.Variants {
			if candidate.ID.String() == binding.VariantID.String() {
				variant = candidate
				break
			}
		}
		if variant.ID == "" {
			return nil, nil, fmt.Errorf("approved binding has no construction variant: %s", binding.VariantID)
		}
		if variant.DirectionPolicy != construction.DirectionBoth && variant.DirectionPolicy != construction.DirectionAnalyze {
			continue
		}
		plan, err := construction.CompileConstruction(construction.CompileInput{
			Family: family.Family, Frame: frame, Variant: variant,
			Profile: construction.ConstructionProfile{
				Language: profile.Language(), Version: profile.Version(), Dimensions: profile.Dimensions(), Provenance: profileSource,
			},
			Constraints: sentenceFamilyConstraints(family.Family),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("compile construction %s for %s: %w", binding.VariantID, frame.ID, err)
		}
		inputs = append(inputs, parser.PlanInput{Plan: plan, Family: family.Family, Variant: variant, Frame: frame, Binding: binding})
	}
	if len(inputs) == 0 {
		return nil, nil, fmt.Errorf("no approved analysis-capable construction bindings")
	}
	index, err := parser.Compile(inputs, parser.DefaultBudget())
	if err != nil {
		return nil, nil, err
	}
	return index, variants, nil
}

func sentenceFamilyConstraints(family construction.ConstructionFamily) []construction.CompiledConstraint {
	var constraints []construction.CompiledConstraint
	for _, source := range family.Constraints {
		if source.Kind != construction.ConstraintAgreement {
			continue
		}
		for _, feature := range source.Features {
			constraints = append(constraints, construction.CompiledConstraint{
				ID:       source.ID + "_" + string(feature),
				Kind:     construction.ConstraintAgree,
				Strength: source.Strength,
				Target:   construction.SourcePath{SlotID: source.TargetSlot, Feature: feature},
				Controller: construction.SourcePath{
					SlotID: source.ControllerSlot, Feature: feature,
				},
				Features: []features.FeatureID{feature},
			})
		}
	}
	return constraints
}
