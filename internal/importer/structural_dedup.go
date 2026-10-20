package importer

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
)

// compileStructuralPlans compiles only frame/family/variant combinations
// explicitly licensed by each frame. CompilePlanIndex performs deterministic
// structural deduplication while preserving each sense/frame permission.
func compileStructuralPlans(dataset ConstructionDataset, featureProfile *features.LanguageFeatureProfile) (*construction.PlanIndex, error) {
	families := make(map[string]construction.ConstructionPackage, len(dataset.Packages))
	for _, packageValue := range dataset.Packages {
		families[packageValue.Family.ID.String()] = packageValue
	}
	profile := construction.ConstructionProfile{Language: featureProfile.Language(), Version: featureProfile.Version(), Dimensions: featureProfile.Dimensions(), Provenance: dataset.ProfileSource}
	inputs := make([]construction.CompileInput, 0)
	for frameIndex, frame := range dataset.Frames {
		for _, familyID := range frame.AllowedFamilyIDs {
			packageValue, exists := families[familyID.String()]
			if !exists {
				return nil, unknown("UNKNOWN_ALLOWED_FAMILY", fmt.Sprintf("valency_frames[%d].allowed_family_ids", frameIndex), "valency frame refers to no construction package family")
			}
			if len(inputs)+len(packageValue.Variants) > MaxConstructionCompileInputs {
				return nil, resourceLimit("construction_compile_inputs", len(inputs)+len(packageValue.Variants), MaxConstructionCompileInputs)
			}
			for _, variant := range packageValue.Variants {
				inputs = append(inputs, construction.CompileInput{Family: packageValue.Family, Frame: frame, Profile: profile, Variant: variant})
			}
		}
	}
	if len(inputs) == 0 {
		return nil, invalid("EMPTY_CONSTRUCTION_COMBINATIONS", "valency_frames", "no frame explicitly licenses a family with variants")
	}
	return construction.CompilePlanIndex(inputs)
}

func validateOperatorReferences(packages []construction.ConstructionPackage, operators []construction.Operator) error {
	byID := make(map[construction.OperatorID]construction.Operator, len(operators))
	listedBy := make(map[construction.OperatorID]map[string]struct{})
	for _, operator := range operators {
		byID[operator.ID] = operator
	}
	for packageIndex, packageValue := range packages {
		for _, operatorID := range packageValue.Family.Operators {
			if _, exists := byID[operatorID]; !exists {
				return unknown("UNKNOWN_FAMILY_OPERATOR", fmt.Sprintf("construction_packages[%d].family.operators", packageIndex), "family references an operator that is absent from the dataset")
			}
			if listedBy[operatorID] == nil {
				listedBy[operatorID] = make(map[string]struct{})
			}
			listedBy[operatorID][packageValue.Family.ID.String()] = struct{}{}
		}
	}
	dependents := make(map[construction.OperatorID][]construction.OperatorID, len(operators))
	indegree := make(map[construction.OperatorID]int, len(operators))
	for operatorIndex, operator := range operators {
		owners := listedBy[operator.ID]
		if len(owners) == 0 {
			return invalid("UNATTACHED_CONSTRUCTION_OPERATOR", fmt.Sprintf("operators[%d].id", operatorIndex), "operator must be explicitly listed by an owning family")
		}
		_, inputOwner := owners[operator.InputFamilyID.String()]
		if !inputOwner {
			return invalid("OPERATOR_INPUT_FAMILY_NOT_LICENSED", fmt.Sprintf("operators[%d].id", operatorIndex), "operator must be listed by its input family")
		}
		for ownerID := range owners {
			if ownerID != operator.InputFamilyID.String() && ownerID != operator.OutputFamilyID.String() {
				return invalid("OPERATOR_FAMILY_ATTACHMENT_MISMATCH", fmt.Sprintf("operators[%d].id", operatorIndex), "operator is listed by a family outside its declared input/output pair")
			}
		}
		for _, dependency := range operator.Dependencies {
			if _, exists := byID[dependency]; !exists {
				return unknown("UNKNOWN_OPERATOR_DEPENDENCY", fmt.Sprintf("operators[%d].dependencies", operatorIndex), "operator dependency does not resolve in this dataset")
			}
			dependents[dependency] = append(dependents[dependency], operator.ID)
			indegree[operator.ID]++
		}
	}
	ready := make([]construction.OperatorID, 0, len(operators))
	for _, operator := range operators {
		if indegree[operator.ID] == 0 {
			ready = append(ready, operator.ID)
		}
	}
	visited := 0
	for readyIndex := 0; readyIndex < len(ready); readyIndex++ {
		current := ready[readyIndex]
		visited++
		for _, dependent := range dependents[current] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
	}
	if visited != len(operators) {
		return invalid("CONSTRUCTION_OPERATOR_DEPENDENCY_CYCLE", "operators", "operator dependencies must be acyclic")
	}
	return nil
}
