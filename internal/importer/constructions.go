package importer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	ConstructionDatasetSchemaVersion = 1
	MaxConstructionDatasetBytes      = 16 << 20
	MaxConstructionDatasetPackages   = 1_024
	MaxConstructionDatasetOperators  = 32_768
	MaxConstructionCompileInputs     = 100_000
)

// ConstructionDataset contains declarative language-local construction and
// sense-specific valency source records. It does not activate translation.
type ConstructionDataset struct {
	SchemaVersion int
	Language      features.LanguageID
	Packages      []construction.ConstructionPackage
	Frames        []valency.ValencyFrame
	Operators     []construction.Operator
	ProfileSource provenance.SourceRef
}

type CompiledConstructionDataset struct {
	Dataset   ConstructionDataset
	Variants  []*construction.VariantIndex
	Valencies *valency.Index
	Plans     *construction.PlanIndex
}

type constructionDatasetWire struct {
	SchemaVersion        int                     `json:"schema_version"`
	Language             features.LanguageID     `json:"language"`
	ProfileVersion       features.ProfileVersion `json:"profile_version"`
	ProfileProvenance    provenance.SourceRef    `json:"profile_provenance"`
	ConstructionPackages []json.RawMessage       `json:"construction_packages"`
	ValencyFrames        []valencyFrameWire      `json:"valency_frames"`
	Operators            []operatorWire          `json:"operators"`
}

type valencyFrameWire struct {
	ID                     string                      `json:"id"`
	SenseID                string                      `json:"sense_id"`
	SemanticRoles          []meaning.Role              `json:"semantic_roles"`
	Arguments              []valency.ArgumentSlot      `json:"arguments"`
	SelectionalConstraints []selectionalConstraintWire `json:"selectional_constraints,omitempty"`
	AllowedFamilyIDs       []string                    `json:"allowed_family_ids"`
	Provenance             provenance.SourceRef        `json:"provenance"`
}

type selectionalConstraintWire struct {
	ID                string                     `json:"id"`
	Role              meaning.Role               `json:"role"`
	Strength          valency.ConstraintStrength `json:"strength"`
	AllowedConceptIDs []string                   `json:"allowed_concept_ids,omitempty"`
	AllowedKinds      []string                   `json:"allowed_kinds,omitempty"`
}

type operatorWire struct {
	ID             string                     `json:"id"`
	Kind           construction.OperatorKind  `json:"kind"`
	Value          construction.OperatorValue `json:"value"`
	InputFamilyID  string                     `json:"input_family_id"`
	OutputFamilyID string                     `json:"output_family_id"`
	SemanticScope  struct {
		TargetNodeID string `json:"target_node_id"`
		ScopeNodeID  string `json:"scope_node_id,omitempty"`
	} `json:"semantic_scope"`
	Dependencies []string             `json:"dependencies,omitempty"`
	Capabilities []string             `json:"capabilities,omitempty"`
	Provenance   provenance.SourceRef `json:"provenance"`
}

// DecodeConstructionDataset decodes a strict top-level dataset, reuses the
// construction package decoder, types valency references, and compiles plans
// against the exact matching local feature profile.
func DecodeConstructionDataset(data []byte, featureProfile *features.LanguageFeatureProfile) (CompiledConstructionDataset, error) {
	if len(data) == 0 {
		return CompiledConstructionDataset{}, invalid("EMPTY_CONSTRUCTION_DATASET", "dataset", "construction dataset must not be empty")
	}
	if len(data) > MaxConstructionDatasetBytes {
		return CompiledConstructionDataset{}, resourceLimit("construction_dataset_bytes", len(data), MaxConstructionDatasetBytes)
	}
	if featureProfile == nil {
		return CompiledConstructionDataset{}, unknown("FEATURE_PROFILE_UNAVAILABLE", "feature_profile", "construction import requires a language feature profile")
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return CompiledConstructionDataset{}, invalid("DUPLICATE_JSON_KEY", "dataset", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire constructionDatasetWire
	if err := decoder.Decode(&wire); err != nil {
		return CompiledConstructionDataset{}, invalid("INVALID_CONSTRUCTION_DATASET_JSON", "dataset", err.Error())
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return CompiledConstructionDataset{}, invalid("TRAILING_CONSTRUCTION_DATASET_DATA", "dataset", "dataset must contain exactly one JSON value")
	}
	if wire.SchemaVersion != ConstructionDatasetSchemaVersion {
		return CompiledConstructionDataset{}, invalid("CONSTRUCTION_DATASET_SCHEMA_VERSION", "schema_version", "unsupported construction dataset schema version")
	}
	if wire.Language != featureProfile.Language() || wire.ProfileVersion != featureProfile.Version() {
		return CompiledConstructionDataset{}, incomplete("CONSTRUCTION_FEATURE_PROFILE_MISMATCH", "profile_version", "dataset language and version must exactly match the supplied feature profile")
	}
	if err := wire.ProfileProvenance.Validate(); err != nil {
		return CompiledConstructionDataset{}, invalid("INVALID_CONSTRUCTION_PROFILE_PROVENANCE", "profile_provenance", err.Error())
	}
	if len(wire.ConstructionPackages) == 0 || len(wire.ConstructionPackages) > MaxConstructionDatasetPackages {
		if len(wire.ConstructionPackages) > MaxConstructionDatasetPackages {
			return CompiledConstructionDataset{}, resourceLimit("construction_packages", len(wire.ConstructionPackages), MaxConstructionDatasetPackages)
		}
		return CompiledConstructionDataset{}, invalid("EMPTY_CONSTRUCTION_PACKAGES", "construction_packages", "dataset requires construction packages")
	}
	if len(wire.ValencyFrames) == 0 {
		return CompiledConstructionDataset{}, invalid("EMPTY_VALENCY_FRAMES", "valency_frames", "dataset requires at least one valency frame")
	}
	if len(wire.ValencyFrames) > valency.MaxFrames {
		return CompiledConstructionDataset{}, resourceLimit("valency_frames", len(wire.ValencyFrames), valency.MaxFrames)
	}
	if len(wire.Operators) > MaxConstructionDatasetOperators {
		return CompiledConstructionDataset{}, resourceLimit("construction_operators", len(wire.Operators), MaxConstructionDatasetOperators)
	}
	result := CompiledConstructionDataset{Dataset: ConstructionDataset{SchemaVersion: wire.SchemaVersion, Language: wire.Language, ProfileSource: wire.ProfileProvenance}}
	families := make(map[string]int, len(wire.ConstructionPackages))
	variantIDs := make(map[string]struct{})
	result.Variants = make([]*construction.VariantIndex, len(wire.ConstructionPackages))
	for index, rawPackage := range wire.ConstructionPackages {
		compiledPackage, err := construction.DecodeConstructionPackage(rawPackage)
		if err != nil {
			return CompiledConstructionDataset{}, err
		}
		packageValue := compiledPackage.Package
		if packageValue.Language != wire.Language {
			return CompiledConstructionDataset{}, unsupported("CONSTRUCTION_DATASET_LANGUAGE_MISMATCH", fmt.Sprintf("construction_packages[%d].language", index), "all construction packages must match the dataset language")
		}
		familyKey := packageValue.Family.ID.String()
		if _, exists := families[familyKey]; exists {
			return CompiledConstructionDataset{}, invalid("DUPLICATE_CONSTRUCTION_FAMILY_ID", fmt.Sprintf("construction_packages[%d].family.id", index), "family ID is duplicated across packages")
		}
		families[familyKey] = index
		for variantIndex, variant := range packageValue.Variants {
			variantKey := variant.ID.String()
			if _, exists := variantIDs[variantKey]; exists {
				return CompiledConstructionDataset{}, invalid("DUPLICATE_CONSTRUCTION_VARIANT_ID", fmt.Sprintf("construction_packages[%d].variants[%d].id", index, variantIndex), "variant ID is duplicated across packages")
			}
			variantIDs[variantKey] = struct{}{}
			if err := validateLocalFeatures(packageValue.Family, variant, featureProfile); err != nil {
				return CompiledConstructionDataset{}, err
			}
		}
		result.Dataset.Packages = append(result.Dataset.Packages, packageValue)
		result.Variants[index] = compiledPackage.Index
	}
	for index, frameWire := range wire.ValencyFrames {
		frame, err := decodeValencyFrame(frameWire)
		if err != nil {
			return CompiledConstructionDataset{}, invalid("INVALID_VALENCY_FRAME", fmt.Sprintf("valency_frames[%d]", index), err.Error())
		}
		result.Dataset.Frames = append(result.Dataset.Frames, frame)
	}
	result.Dataset.Operators = make([]construction.Operator, len(wire.Operators))
	operatorIDs := make(map[construction.OperatorID]struct{}, len(wire.Operators))
	for index, operatorWire := range wire.Operators {
		operator, err := decodeOperator(operatorWire)
		if err != nil {
			return CompiledConstructionDataset{}, invalid("INVALID_CONSTRUCTION_OPERATOR", fmt.Sprintf("operators[%d]", index), err.Error())
		}
		if err := operator.Validate(); err != nil {
			return CompiledConstructionDataset{}, err
		}
		if _, exists := operatorIDs[operator.ID]; exists {
			return CompiledConstructionDataset{}, invalid("DUPLICATE_CONSTRUCTION_OPERATOR_ID", fmt.Sprintf("operators[%d].id", index), "operator ID is duplicated")
		}
		operatorIDs[operator.ID] = struct{}{}
		if _, ok := families[operator.InputFamilyID.String()]; !ok {
			return CompiledConstructionDataset{}, unknown("UNKNOWN_OPERATOR_INPUT_FAMILY", fmt.Sprintf("operators[%d].input_family_id", index), "operator input family does not resolve")
		}
		if _, ok := families[operator.OutputFamilyID.String()]; !ok {
			return CompiledConstructionDataset{}, unknown("UNKNOWN_OPERATOR_OUTPUT_FAMILY", fmt.Sprintf("operators[%d].output_family_id", index), "operator output family does not resolve")
		}
		result.Dataset.Operators[index] = operator
	}
	if err := validateValencyFrames(result.Dataset.Frames, families, result.Dataset.Packages, string(wire.Language)); err != nil {
		return CompiledConstructionDataset{}, err
	}
	if err := validateOperatorReferences(result.Dataset.Packages, result.Dataset.Operators); err != nil {
		return CompiledConstructionDataset{}, err
	}
	index, err := valency.BuildIndex(result.Dataset.Frames, nil)
	if err != nil {
		return CompiledConstructionDataset{}, err
	}
	result.Valencies = index
	planIndex, err := compileStructuralPlans(result.Dataset, featureProfile)
	if err != nil {
		return CompiledConstructionDataset{}, err
	}
	result.Plans = planIndex
	return result, nil
}

func validateLocalFeatures(family construction.ConstructionFamily, variant construction.Variant, profile *features.LanguageFeatureProfile) error {
	for constraintIndex, constraint := range family.Constraints {
		for featureIndex, feature := range constraint.Features {
			if _, exists := profile.Lookup(feature); !exists {
				return unknown("UNKNOWN_FAMILY_FEATURE", fmt.Sprintf("family.constraints[%d].features[%d]", constraintIndex, featureIndex), "family feature is not declared by the supplied language profile")
			}
		}
	}
	for conditionIndex, condition := range variant.Conditions {
		if condition.Kind != construction.ConditionFeatureEquals {
			continue
		}
		dimension, exists := profile.Lookup(condition.Feature)
		if !exists {
			return unknown("UNKNOWN_VARIANT_FEATURE", fmt.Sprintf("variant.conditions[%d].feature", conditionIndex), "variant feature is not declared by the supplied language profile")
		}
		if !containsFeatureValue(dimension.Values, condition.Value) {
			return unknown("UNKNOWN_VARIANT_FEATURE_VALUE", fmt.Sprintf("variant.conditions[%d].value", conditionIndex), "variant feature value is not declared by the supplied language profile")
		}
	}
	return nil
}

func decodeValencyFrame(wire valencyFrameWire) (valency.ValencyFrame, error) {
	senseID, err := ids.ParseSenseID(wire.SenseID)
	if err != nil {
		return valency.ValencyFrame{}, fmt.Errorf("sense_id: %w", err)
	}
	familyIDs := make([]ids.ConstructionID, len(wire.AllowedFamilyIDs))
	for index, rawID := range wire.AllowedFamilyIDs {
		familyIDs[index], err = ids.ParseConstructionID(rawID)
		if err != nil {
			return valency.ValencyFrame{}, fmt.Errorf("allowed_family_ids[%d]: %w", index, err)
		}
	}
	constraints := make([]valency.SelectionalConstraint, len(wire.SelectionalConstraints))
	for index, rawConstraint := range wire.SelectionalConstraints {
		conceptIDs := make([]ids.ConceptID, len(rawConstraint.AllowedConceptIDs))
		for conceptIndex, rawID := range rawConstraint.AllowedConceptIDs {
			conceptIDs[conceptIndex], err = ids.ParseConceptID(rawID)
			if err != nil {
				return valency.ValencyFrame{}, fmt.Errorf("selectional_constraints[%d].allowed_concept_ids[%d]: %w", index, conceptIndex, err)
			}
		}
		parsedKinds := make([]core.ConceptKind, len(rawConstraint.AllowedKinds))
		for kindIndex, kind := range rawConstraint.AllowedKinds {
			parsedKinds[kindIndex] = core.ConceptKind(kind)
		}
		constraints[index] = valency.SelectionalConstraint{ID: rawConstraint.ID, Role: rawConstraint.Role, Strength: rawConstraint.Strength, AllowedConceptIDs: conceptIDs, AllowedKinds: parsedKinds}
	}
	return valency.ValencyFrame{ID: valency.FrameID(wire.ID), SenseID: senseID, SemanticRoles: wire.SemanticRoles, Arguments: wire.Arguments, SelectionalConstraints: constraints, AllowedFamilyIDs: familyIDs, Provenance: wire.Provenance}, nil
}

func decodeOperator(wire operatorWire) (construction.Operator, error) {
	if !validMeaningNodeID(wire.SemanticScope.TargetNodeID) || wire.SemanticScope.ScopeNodeID != "" && !validMeaningNodeID(wire.SemanticScope.ScopeNodeID) {
		return construction.Operator{}, fmt.Errorf("semantic_scope node IDs are malformed")
	}
	inputFamily, err := ids.ParseConstructionID(wire.InputFamilyID)
	if err != nil {
		return construction.Operator{}, fmt.Errorf("input_family_id: %w", err)
	}
	outputFamily, err := ids.ParseConstructionID(wire.OutputFamilyID)
	if err != nil {
		return construction.Operator{}, fmt.Errorf("output_family_id: %w", err)
	}
	dependencies := make([]construction.OperatorID, len(wire.Dependencies))
	for index, value := range wire.Dependencies {
		dependencies[index] = construction.OperatorID(value)
	}
	capabilities := make([]construction.Capability, len(wire.Capabilities))
	for index, value := range wire.Capabilities {
		capabilities[index] = construction.Capability(value)
	}
	return construction.Operator{ID: construction.OperatorID(wire.ID), Kind: wire.Kind, Value: wire.Value, InputFamilyID: inputFamily, OutputFamilyID: outputFamily, SemanticScope: construction.SemanticScope{TargetNodeID: meaning.NodeID(wire.SemanticScope.TargetNodeID), ScopeNodeID: meaning.NodeID(wire.SemanticScope.ScopeNodeID)}, Dependencies: dependencies, Capabilities: capabilities, Provenance: wire.Provenance}, nil
}

func validMeaningNodeID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}
