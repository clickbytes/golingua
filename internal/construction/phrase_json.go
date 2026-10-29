package construction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

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
	if profile.SchemaVersion != 1 && profile.SchemaVersion != 2 && profile.SchemaVersion != 3 {
		return PhraseProfile{}, invalid("PHRASE_PROFILE_SCHEMA_VERSION", "schema_version", "unsupported phrase profile schema")
	}
	if featureProfile == nil || profile.Language != featureProfile.Language() {
		return PhraseProfile{}, unsupported("PHRASE_PROFILE_LANGUAGE_MISMATCH", "language", "phrase and feature profiles must have the same language")
	}
	if !validToken(profile.Version) || len(profile.Templates) == 0 || len(profile.Templates) > MaxPhrasePlans {
		return PhraseProfile{}, invalid("INVALID_PHRASE_PROFILE", "profile", "profile requires a bounded version and non-empty template list")
	}
	seen := map[string]struct{}{}
	if len(profile.DeterminerForms) > 4096 {
		return PhraseProfile{}, resourceLimit("determiner_forms", len(profile.DeterminerForms), 4096)
	}
	if profile.SchemaVersion == 1 && len(profile.DeterminerForms) != 0 {
		return PhraseProfile{}, invalid("PHRASE_PROFILE_SCHEMA_VERSION", "schema_version", "determiner forms require phrase schema version 2")
	}
	seenForms := map[string]struct{}{}
	for i, rule := range profile.DeterminerForms {
		path := fmt.Sprintf("determiner_forms[%d]", i)
		if rule.Surface == "" || strings.TrimSpace(rule.Surface) != rule.Surface || utf8.RuneCountInString(rule.Surface) > 128 || strings.ContainsAny(rule.Surface, "\r\n\t") || (rule.Onset != "" && !validToken(rule.Onset)) || len(rule.HeadFeatures) == 0 || len(rule.HeadFeatures) > 16 || len(rule.DeterminerFeatures) == 0 || len(rule.DeterminerFeatures) > 16 {
			return PhraseProfile{}, invalid("INVALID_DETERMINER_FORM", path, "determiner rule requires a bounded surface and feature tuple")
		}
		key := rule.Surface + "\x00" + rule.Onset + "\x00" + featureTupleKey(rule.HeadFeatures) + "\x00" + featureTupleKey(rule.DeterminerFeatures)
		if _, ok := seenForms[key]; ok {
			return PhraseProfile{}, invalid("DUPLICATE_DETERMINER_FORM", path, "determiner feature tuple is duplicated")
		}
		seenForms[key] = struct{}{}
		for _, values := range [][]features.FeatureValue{rule.HeadFeatures, rule.DeterminerFeatures} {
			bundle, err := features.NewBundle(featureProfile, values)
			if err != nil {
				return PhraseProfile{}, err
			}
			if err := features.ValidateBundle(featureProfile, bundle); err != nil {
				return PhraseProfile{}, err
			}
		}
	}
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
			if !validToken(input.Role) || (input.Kind != PhraseInputLexical && input.Kind != PhraseInputNP && input.Kind != PhraseInputAP && input.Kind != PhraseInputVP && input.Kind != PhraseInputPP) {
				return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE_INPUT", path+".inputs", "input role and kind must be supported")
			}
			if !validTemplateInputShape(template.Kind, input) {
				return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE_INPUT_SHAPE", path+".inputs", "input role, kind, and requiredness are incompatible with this phrase kind")
			}
			if _, ok := roles[input.Role]; ok {
				return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE_INPUT", path+".inputs", "input role is duplicated")
			}
			roles[input.Role] = struct{}{}
			inputFeatures := map[features.FeatureID]struct{}{}
			for _, id := range input.FeatureIDs {
				if _, ok := featureProfile.Lookup(id); !ok {
					return PhraseProfile{}, unknown("UNKNOWN_PHRASE_INPUT_FEATURE", path+".inputs."+input.Role+".feature_ids", "input feature is absent from the language feature profile")
				}
				if _, exists := inputFeatures[id]; exists {
					return PhraseProfile{}, invalid("DUPLICATE_PHRASE_INPUT_FEATURE", path+".inputs."+input.Role+".feature_ids", "input feature is duplicated")
				}
				inputFeatures[id] = struct{}{}
				if len(template.FeatureIDs) > 0 && !containsFeatureID(template.FeatureIDs, id) {
					return PhraseProfile{}, unsupported("PHRASE_INPUT_FEATURE_NOT_IN_TEMPLATE", path+".inputs."+input.Role+".feature_ids", "input feature is not declared by the phrase template")
				}
			}
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
			if template.Kind == PhraseNP && ((child.Role != "modifier" && child.Role != "quantity") || (child.Kind != PhraseNP && child.Kind != PhrasePP)) {
				return PhraseProfile{}, invalid("INVALID_NP_TEMPLATE_CHILD", path+".children", "NP child must be a modifier or quantity phrase")
			}
			if template.Kind == PhraseVP && (child.Role != "auxiliary" || child.Kind != PhraseVP || child.Required) {
				return PhraseProfile{}, invalid("INVALID_VP_TEMPLATE_CHILD", path+".children", "VP auxiliary placeholder must be optional and VP typed")
			}
		}
		if len(template.Order) > 0 {
			ordered := map[string]struct{}{}
			declaredRoles := make(map[string]struct{}, len(roles)+len(childRoles))
			for role := range roles {
				declaredRoles[role] = struct{}{}
			}
			for role := range childRoles {
				declaredRoles[role] = struct{}{}
			}
			for _, role := range template.Order {
				if !validToken(role) {
					return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE_ORDER", path+".order", "template order contains an invalid role")
				}
				if _, exists := ordered[role]; exists {
					return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE_ORDER_ROLE", path+".order", "template order role is duplicated")
				}
				if _, isInput := roles[role]; !isInput {
					if _, isChild := childRoles[role]; !isChild {
						return PhraseProfile{}, unknown("UNKNOWN_PHRASE_TEMPLATE_ORDER_ROLE", path+".order", "template order role is not declared as an input or child")
					}
				}
				ordered[role] = struct{}{}
			}
			if len(ordered) != len(declaredRoles) {
				return PhraseProfile{}, incomplete("INCOMPLETE_PHRASE_TEMPLATE_ORDER", path+".order", "template order must include every input and child role exactly once")
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
		if len(template.Agreements) > MaxTemplateAgreements {
			return PhraseProfile{}, resourceLimit(path+".agreements", len(template.Agreements), MaxTemplateAgreements)
		}
		if profile.SchemaVersion < 3 && len(template.Agreements) > 0 {
			return PhraseProfile{}, invalid("PHRASE_PROFILE_SCHEMA_VERSION", path+".agreements", "template agreements require phrase schema version 3")
		}
		if len(template.Agreements) > 0 && template.Kind != PhraseNP {
			return PhraseProfile{}, unsupported("PHRASE_TEMPLATE_AGREEMENTS_UNSUPPORTED", path+".agreements", "cross-role agreement declarations are currently supported for NP templates")
		}
		inputByRole := make(map[string]PhraseTemplateInput, len(template.Inputs))
		for _, input := range template.Inputs {
			inputByRole[input.Role] = input
		}
		seenAgreements := map[string]struct{}{}
		for j, agreement := range template.Agreements {
			agreementPath := fmt.Sprintf("%s.agreements[%d]", path, j)
			if !validToken(agreement.ControllerRole) || !validToken(agreement.TargetRole) || agreement.ControllerRole == agreement.TargetRole {
				return PhraseProfile{}, invalid("INVALID_PHRASE_TEMPLATE_AGREEMENT", agreementPath, "agreement requires two distinct input roles")
			}
			controller, controllerOK := inputByRole[agreement.ControllerRole]
			target, targetOK := inputByRole[agreement.TargetRole]
			if !controllerOK || !targetOK {
				return PhraseProfile{}, unknown("UNKNOWN_PHRASE_TEMPLATE_AGREEMENT_ROLE", agreementPath, "agreement roles must be declared template inputs")
			}
			if _, ok := featureProfile.Lookup(agreement.Feature); !ok {
				return PhraseProfile{}, unknown("UNKNOWN_PHRASE_TEMPLATE_AGREEMENT_FEATURE", agreementPath+".feature", "agreement feature is absent from the language feature profile")
			}
			if _, ok := featureIDs[agreement.Feature]; !ok {
				return PhraseProfile{}, unsupported("PHRASE_TEMPLATE_AGREEMENT_FEATURE_NOT_IN_TEMPLATE", agreementPath+".feature", "agreement feature must be declared by the phrase template")
			}
			controllerFeatures := controller.FeatureIDs
			if len(controllerFeatures) == 0 {
				controllerFeatures = template.FeatureIDs
			}
			targetFeatures := target.FeatureIDs
			if len(targetFeatures) == 0 {
				targetFeatures = template.FeatureIDs
			}
			if !containsFeatureID(controllerFeatures, agreement.Feature) || !containsFeatureID(targetFeatures, agreement.Feature) {
				return PhraseProfile{}, unsupported("PHRASE_TEMPLATE_AGREEMENT_FEATURE_NOT_IN_INPUT", agreementPath+".feature", "agreement feature must be declared for both input roles")
			}
			left, right := agreement.ControllerRole, agreement.TargetRole
			if right < left {
				left, right = right, left
			}
			key := left + "\x00" + right + "\x00" + string(agreement.Feature)
			if _, exists := seenAgreements[key]; exists {
				return PhraseProfile{}, invalid("DUPLICATE_PHRASE_TEMPLATE_AGREEMENT", agreementPath, "role and feature agreement is duplicated")
			}
			seenAgreements[key] = struct{}{}
		}
	}
	if len(profile.DeterminerForms) > 0 {
		var np *PhraseTemplate
		for i := range profile.Templates {
			if profile.Templates[i].ID == "np-attributive" && profile.Templates[i].Kind == PhraseNP {
				np = &profile.Templates[i]
				break
			}
		}
		if np == nil {
			return PhraseProfile{}, invalid("DETERMINER_TEMPLATE_UNAVAILABLE", "templates", "determiner forms require an np-attributive template")
		}
		declared := map[string]map[features.FeatureID]bool{}
		for _, input := range np.Inputs {
			declared[input.Role] = map[features.FeatureID]bool{}
			for _, id := range input.FeatureIDs {
				declared[input.Role][id] = true
			}
		}
		for i, rule := range profile.DeterminerForms {
			for _, check := range []struct {
				role   string
				values []features.FeatureValue
			}{{"head", rule.HeadFeatures}, {"determiner", rule.DeterminerFeatures}} {
				for _, value := range check.values {
					if !declared[check.role][value.Feature] {
						return PhraseProfile{}, unsupported("DETERMINER_FEATURE_NOT_IN_TEMPLATE", fmt.Sprintf("determiner_forms[%d].%s_features.%s", i, check.role, value.Feature), "determiner rule feature must be declared by its np-attributive input")
					}
				}
			}
		}
	}
	return profile, nil
}

func featureTupleKey(values []features.FeatureValue) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, string(value.Feature)+"="+string(value.State)+":"+string(value.Value))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func validTemplateInputShape(kind PhraseKind, input PhraseTemplateInput) bool {
	switch kind {
	case PhraseNP:
		switch input.Role {
		case "head":
			return input.Kind == PhraseInputLexical && input.Required
		case "determiner", "quantity":
			return input.Kind == PhraseInputLexical && !input.Required
		case "modifier":
			return (input.Kind == PhraseInputNP || input.Kind == PhraseInputAP) && !input.Required
		}
	case PhraseVP:
		switch input.Role {
		case "verb":
			return input.Kind == PhraseInputLexical && input.Required
		case "auxiliary":
			return input.Kind == PhraseInputLexical && !input.Required
		}
	case PhrasePP:
		return input.Role == "preposition" && input.Kind == PhraseInputLexical && input.Required
	}
	return false
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
