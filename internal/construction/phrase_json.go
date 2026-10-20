package construction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/features"
)

const MaxPhraseProfileBytes = 1 << 20

// DecodePhraseProfile validates a bounded language-local structural profile.
// Profiles declare reusable plan shapes, not lexical data or realization rules.
func DecodePhraseProfile(data []byte, featureProfile *features.LanguageFeatureProfile) (PhraseProfile, error) {
	if len(data) == 0 || len(data) > MaxPhraseProfileBytes {
		return PhraseProfile{}, invalid("INVALID_PHRASE_PROFILE_SIZE", "profile", "phrase profile must be non-empty and within the byte limit")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return PhraseProfile{}, invalid("DUPLICATE_JSON_KEY", "profile", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var profile PhraseProfile
	if err := decoder.Decode(&profile); err != nil {
		return PhraseProfile{}, invalid("INVALID_PHRASE_PROFILE_JSON", "profile", err.Error())
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return PhraseProfile{}, invalid("TRAILING_PHRASE_PROFILE_DATA", "profile", "profile must contain exactly one JSON value")
	}
	if profile.SchemaVersion != 1 {
		return PhraseProfile{}, invalid("PHRASE_PROFILE_SCHEMA_VERSION", "schema_version", "unsupported phrase profile schema")
	}
	if featureProfile == nil || profile.Language != featureProfile.Language() {
		return PhraseProfile{}, unsupported("PHRASE_PROFILE_LANGUAGE_MISMATCH", "language", "phrase and feature profiles must have the same language")
	}
	if !validToken(profile.Version) || len(profile.Templates) == 0 || len(profile.Templates) > MaxPhrasePlans {
		return PhraseProfile{}, invalid("INVALID_PHRASE_PROFILE", "profile", "profile requires a bounded version and non-empty template list")
	}
	seen := map[string]struct{}{}
	for i, template := range profile.Templates {
		path := fmt.Sprintf("templates[%d]", i)
		if !validToken(template.ID) || (template.Kind != PhraseNP && template.Kind != PhraseVP && template.Kind != PhrasePP) {
			return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE", path, "template requires a canonical ID and supported phrase kind")
		}
		if _, ok := seen[template.ID]; ok {
			return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE", path+".id", "template ID is duplicated")
		}
		seen[template.ID] = struct{}{}
		roles := map[string]struct{}{}
		for _, input := range template.Inputs {
			if !validToken(input.Role) || (input.Kind != PhraseInputLexical && input.Kind != PhraseInputNP && input.Kind != PhraseInputVP && input.Kind != PhraseInputPP) {
				return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE_INPUT", path+".inputs", "input role and kind must be supported")
			}
			if _, ok := roles[input.Role]; ok {
				return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE_INPUT", path+".inputs", "input role is duplicated")
			}
			roles[input.Role] = struct{}{}
		}
		headRole := "head"
		if template.Kind == PhraseVP {
			headRole = "verb"
		}
		if template.Kind == PhrasePP {
			headRole = "preposition"
		}
		hasHead := false
		for _, input := range template.Inputs {
			if input.Role == headRole && input.Kind == PhraseInputLexical && input.Required {
				hasHead = true
			}
		}
		if !hasHead {
			return PhraseProfile{}, invalid("PHRASE_TEMPLATE_MISSING_HEAD", path+".inputs", "each phrase template requires its typed lexical head input")
		}
		childRoles := map[string]struct{}{}
		for _, child := range template.Children {
			if !validToken(child.Role) || (child.Kind != PhraseNP && child.Kind != PhraseVP && child.Kind != PhrasePP) {
				return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE_CHILD", path+".children", "child role and kind must be supported")
			}
			if _, ok := childRoles[child.Role]; ok {
				return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE_CHILD", path+".children", "child role is duplicated")
			}
			childRoles[child.Role] = struct{}{}
			if template.Kind == PhrasePP && (child.Role != "complement" || child.Kind != PhraseNP || !child.Required) {
				return PhraseProfile{}, invalid("INVALID_PP_TEMPLATE_COMPLEMENT", path+".children", "PP template requires one required NP complement")
			}
		}
		if template.Kind == PhrasePP {
			if _, ok := childRoles["complement"]; !ok {
				return PhraseProfile{}, invalid("MISSING_PP_TEMPLATE_COMPLEMENT", path+".children", "PP template requires a complement child")
			}
		}
		featureIDs := map[features.FeatureID]struct{}{}
		for _, id := range template.FeatureIDs {
			if _, ok := featureProfile.Lookup(id); !ok {
				return PhraseProfile{}, unknown("UNKNOWN_PHRASE_TEMPLATE_FEATURE", path+".feature_ids", "template feature is absent from the language feature profile")
			}
			if _, exists := featureIDs[id]; exists {
				return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE_FEATURE", path+".feature_ids", "template feature is duplicated")
			}
			featureIDs[id] = struct{}{}
		}
	}
	return profile, nil
}

// ValidatePhraseProfile rechecks an in-memory profile through the same schema
// validation used for JSON input.
func ValidatePhraseProfile(profile PhraseProfile, featureProfile *features.LanguageFeatureProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return invalid("INVALID_PHRASE_PROFILE", "profile", err.Error())
	}
	_, err = DecodePhraseProfile(data, featureProfile)
	return err
}
