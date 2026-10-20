package construction

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
)

const MaxQuantityPlanBytes = 1 << 20

// QuantityTimeMWEProfile distinguishes semantic quantity types from the
// grammatical feature dimensions a language may use to express agreement.
type QuantityTimeMWEProfile struct {
	SchemaVersion         int                      `json:"schema_version"`
	Language              features.LanguageID      `json:"language"`
	Version               string                   `json:"version"`
	QuantifierKinds       []meaning.QuantifierKind `json:"quantifier_kinds"`
	GrammaticalFeatureIDs []features.FeatureID     `json:"grammatical_feature_ids"`
	TemporalKinds         []TemporalKind           `json:"temporal_kinds"`
	MWEComponentKinds     []MWEComponentKind       `json:"mwe_component_kinds"`
	MaxMWEComponents      int                      `json:"max_mwe_components"`
}

func DecodeQuantityTimeMWEProfile(data []byte, featureProfile *features.LanguageFeatureProfile) (QuantityTimeMWEProfile, error) {
	if len(data) == 0 || len(data) > MaxQuantityPlanBytes {
		return QuantityTimeMWEProfile{}, invalid("INVALID_QUANTITY_PROFILE_SIZE", "profile", "quantity/time/MWE profile is empty or exceeds its byte limit")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return QuantityTimeMWEProfile{}, invalid("DUPLICATE_JSON_KEY", "profile", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var profile QuantityTimeMWEProfile
	if err := decoder.Decode(&profile); err != nil {
		return QuantityTimeMWEProfile{}, invalid("INVALID_QUANTITY_PROFILE_JSON", "profile", err.Error())
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return QuantityTimeMWEProfile{}, invalid("TRAILING_QUANTITY_PROFILE_DATA", "profile", "profile must contain exactly one JSON value")
	}
	if err := validateQuantityTimeMWEProfile(profile, featureProfile); err != nil {
		return QuantityTimeMWEProfile{}, err
	}
	return profile, nil
}

func ValidateQuantityTimeMWEProfile(profile QuantityTimeMWEProfile, featureProfile *features.LanguageFeatureProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return invalid("INVALID_QUANTITY_PROFILE", "profile", err.Error())
	}
	_, err = DecodeQuantityTimeMWEProfile(data, featureProfile)
	return err
}

func validateQuantityTimeMWEProfile(profile QuantityTimeMWEProfile, featureProfile *features.LanguageFeatureProfile) error {
	if profile.SchemaVersion != 1 || !validToken(profile.Version) || featureProfile == nil || profile.Language != featureProfile.Language() {
		return unsupported("QUANTITY_PROFILE_IDENTITY_MISMATCH", "profile", "profile requires schema 1 and a matching language feature profile")
	}
	if len(profile.QuantifierKinds) == 0 || len(profile.TemporalKinds) == 0 || len(profile.MWEComponentKinds) == 0 || profile.MaxMWEComponents < 2 || profile.MaxMWEComponents > MaxSequenceTokens {
		return invalid("INVALID_QUANTITY_PROFILE", "profile", "profile requires bounded quantity, temporal, and MWE inventories")
	}
	if err := uniqueProfileValues("quantifier_kinds", profile.QuantifierKinds); err != nil {
		return err
	}
	if err := uniqueProfileValues("temporal_kinds", profile.TemporalKinds); err != nil {
		return err
	}
	if err := uniqueProfileValues("mwe_component_kinds", profile.MWEComponentKinds); err != nil {
		return err
	}
	for _, kind := range profile.QuantifierKinds {
		if kind != meaning.QuantifierSome && kind != meaning.QuantifierAll && kind != meaning.QuantifierExactly && kind != meaning.QuantifierAtLeast && kind != meaning.QuantifierAtMost {
			return invalid("UNSUPPORTED_PROFILE_QUANTIFIER", "quantifier_kinds", "profile contains an unsupported semantic quantifier")
		}
	}
	for _, kind := range profile.TemporalKinds {
		if kind != TemporalDateKind && kind != TemporalClockKind && kind != TemporalDurationKind && kind != TemporalRelativeDayKind {
			return invalid("UNSUPPORTED_PROFILE_TEMPORAL_KIND", "temporal_kinds", "profile contains an unsupported temporal kind")
		}
	}
	for _, kind := range profile.MWEComponentKinds {
		if kind != MWEFixed && kind != MWEInflected && kind != MWEInput {
			return invalid("UNSUPPORTED_PROFILE_MWE_COMPONENT", "mwe_component_kinds", "profile contains an unsupported MWE component kind")
		}
	}
	seen := map[features.FeatureID]struct{}{}
	for _, id := range profile.GrammaticalFeatureIDs {
		if _, ok := featureProfile.Lookup(id); !ok {
			return unknown("UNKNOWN_QUANTITY_FEATURE", "grammatical_feature_ids", "grammatical feature is not declared by the language profile")
		}
		if _, ok := seen[id]; ok {
			return invalid("DUPLICATE_QUANTITY_FEATURE", "grammatical_feature_ids", "grammatical feature is duplicated")
		}
		seen[id] = struct{}{}
	}
	return nil
}

type QuantityPlan struct {
	ID                  string                 `json:"id"`
	QuantifierNodeID    meaning.NodeID         `json:"quantifier_node_id"`
	TargetMeaningNodeID meaning.NodeID         `json:"target_meaning_node_id"`
	TargetPhraseID      PhraseID               `json:"target_phrase_id"`
	GrammaticalFeatures features.FeatureBundle `json:"grammatical_features,omitempty"`
}
type CompiledQuantityPlan struct{ Plan QuantityPlan }

func CompileQuantity(plan QuantityPlan, dataProfile QuantityTimeMWEProfile, featureProfile *features.LanguageFeatureProfile, graph meaning.Meaning, phrases *PhraseIndex) (CompiledQuantityPlan, error) {
	if !validToken(plan.ID) || !validMeaningNodeID(plan.QuantifierNodeID) || !validMeaningNodeID(plan.TargetMeaningNodeID) {
		return CompiledQuantityPlan{}, invalid("INVALID_QUANTITY_PLAN", "id", "quantity plan requires bounded IDs")
	}
	if err := ValidateQuantityTimeMWEProfile(dataProfile, featureProfile); err != nil {
		return CompiledQuantityPlan{}, err
	}
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		return CompiledQuantityPlan{}, err
	}
	var quantifier *meaning.Node
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == plan.QuantifierNodeID {
			quantifier = &graph.Nodes[i]
			break
		}
	}
	if quantifier == nil {
		return CompiledQuantityPlan{}, unknown("UNKNOWN_QUANTITY_MEANING_NODE", "quantifier_node_id", "quantity node is absent from Meaning")
	}
	if quantifier.Kind != meaning.KindQuantifier || quantifier.Quantifier == nil {
		return CompiledQuantityPlan{}, unsupported("MEANING_NODE_NOT_QUANTIFIER", "quantifier_node_id", "quantity must be licensed by an explicit Meaning quantifier node")
	}
	if quantifier.Target != plan.TargetMeaningNodeID {
		return CompiledQuantityPlan{}, unsupported("QUANTITY_TARGET_MISMATCH", "target_meaning_node_id", "quantity plan target must match the explicit Meaning quantifier target")
	}
	var targetNode *meaning.Node
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == plan.TargetMeaningNodeID {
			targetNode = &graph.Nodes[i]
			break
		}
	}
	if targetNode == nil || targetNode.Kind != meaning.KindReferent {
		return CompiledQuantityPlan{}, unsupported("QUANTITY_TARGET_NOT_REFERENT", "target_meaning_node_id", "quantity plan must target an explicit Meaning referent")
	}
	if err := quantifier.Quantifier.Validate(); err != nil {
		return CompiledQuantityPlan{}, err
	}
	if !containsQuantifierKind(dataProfile.QuantifierKinds, quantifier.Quantifier.Kind) {
		return CompiledQuantityPlan{}, unsupported("QUANTIFIER_KIND_UNSUPPORTED", "quantifier.kind", "language profile does not declare this semantic quantity kind")
	}
	phrase, ok := phrases.ByID(plan.TargetPhraseID)
	if !ok {
		return CompiledQuantityPlan{}, unknown("UNKNOWN_QUANTITY_TARGET_PHRASE", "target_phrase_id", "quantity target phrase is not compiled")
	}
	if phrase.Plan.Kind != PhraseNP || phrase.Plan.Language != dataProfile.Language {
		return CompiledQuantityPlan{}, unsupported("QUANTITY_TARGET_UNSUPPORTED", "target_phrase_id", "quantity must target a noun phrase in the same language")
	}
	if len(plan.GrammaticalFeatures.Values) > 0 {
		if plan.GrammaticalFeatures.Language != featureProfile.Language() || plan.GrammaticalFeatures.ProfileVersion != featureProfile.Version() {
			return CompiledQuantityPlan{}, unsupported("QUANTITY_FEATURE_PROFILE_MISMATCH", "grammatical_features", "grammatical bundle must use the language feature profile")
		}
		if err := features.ValidateBundle(featureProfile, plan.GrammaticalFeatures); err != nil {
			return CompiledQuantityPlan{}, err
		}
		for _, value := range plan.GrammaticalFeatures.Values {
			if !containsFeatureID(dataProfile.GrammaticalFeatureIDs, value.Feature) {
				return CompiledQuantityPlan{}, unsupported("QUANTITY_GRAMMAR_FEATURE_UNSUPPORTED", "grammatical_features", "feature is not declared for quantity agreement by this language profile")
			}
		}
	} else if plan.GrammaticalFeatures.Language != "" || plan.GrammaticalFeatures.ProfileVersion != "" {
		return CompiledQuantityPlan{}, invalid("EMPTY_QUANTITY_GRAMMAR_BUNDLE", "grammatical_features", "an empty grammatical bundle must be omitted")
	}
	result := plan
	result.GrammaticalFeatures.Values = append([]features.FeatureValue(nil), plan.GrammaticalFeatures.Values...)
	for i := range result.GrammaticalFeatures.Values {
		if result.GrammaticalFeatures.Values[i].Variable != nil {
			variable := *result.GrammaticalFeatures.Values[i].Variable
			result.GrammaticalFeatures.Values[i].Variable = &variable
		}
	}
	return CompiledQuantityPlan{Plan: result}, nil
}

func containsQuantifierKind(values []meaning.QuantifierKind, value meaning.QuantifierKind) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
func uniqueProfileValues[T ~string](path string, values []T) error {
	seen := map[T]struct{}{}
	for _, value := range values {
		if value == "" {
			return invalid("INVALID_QUANTITY_PROFILE_VALUE", path, "profile value must not be empty")
		}
		if _, ok := seen[value]; ok {
			return invalid("DUPLICATE_QUANTITY_PROFILE_VALUE", path, "profile value is duplicated")
		}
		seen[value] = struct{}{}
	}
	return nil
}
