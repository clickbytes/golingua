package features

import "sort"

// NewBundle validates and copies feature values into deterministic feature
// order. Unknown and absent states are preserved exactly.
func NewBundle(profile *LanguageFeatureProfile, values []FeatureValue) (FeatureBundle, error) {
	if profile == nil {
		return FeatureBundle{}, failureUnknownProfile()
	}
	if len(values) > MaxBundleValues {
		return FeatureBundle{}, resourceLimit("bundle_values", len(values), MaxBundleValues)
	}
	bundle := FeatureBundle{Language: profile.language, ProfileVersion: profile.version, Values: cloneValues(values)}
	sort.Slice(bundle.Values, func(i, j int) bool { return bundle.Values[i].Feature < bundle.Values[j].Feature })
	if err := ValidateBundle(profile, bundle); err != nil {
		return FeatureBundle{}, err
	}
	return cloneBundle(bundle), nil
}

// ValidateBundle validates explicitly supplied values without requiring every
// profile dimension. ValidateFeatureRequest additionally checks required
// dimensions for a concrete morphology request.
func ValidateBundle(profile *LanguageFeatureProfile, bundle FeatureBundle) error {
	if profile == nil {
		return failureUnknownProfile()
	}
	if !validLanguage(string(bundle.Language)) {
		return invalid("INVALID_LANGUAGE_ID", "bundle.language", "language tag must use canonical lowercase ASCII")
	}
	if !validToken(string(bundle.ProfileVersion)) {
		return invalid("INVALID_PROFILE_VERSION", "bundle.profile_version", "profile version is malformed or exceeds its bound")
	}
	if bundle.Language != profile.language || bundle.ProfileVersion != profile.version {
		return incomplete("FEATURE_PROFILE_MISMATCH", "bundle", "feature bundle and profile language/version differ")
	}
	if len(bundle.Values) > MaxBundleValues {
		return resourceLimit("bundle_values", len(bundle.Values), MaxBundleValues)
	}
	seen := make(map[FeatureID]struct{}, len(bundle.Values))
	for index, value := range bundle.Values {
		path := "values[" + itoa(index) + "]"
		if !validToken(string(value.Feature)) {
			return invalid("INVALID_FEATURE_ID", path+".feature", "feature ID is malformed or exceeds its bound")
		}
		dimension, exists := profile.Lookup(value.Feature)
		if !exists {
			return unsupported("UNSUPPORTED_FEATURE_DIMENSION", path+".feature", "feature dimension is not declared by this language profile")
		}
		if _, exists := seen[value.Feature]; exists {
			return invalid("DUPLICATE_FEATURE_ID", path+".feature", "feature bundle contains a duplicate dimension")
		}
		seen[value.Feature] = struct{}{}
		switch value.State {
		case StateAbsent:
			if !dimension.AllowAbsent {
				return unsupported("ABSENT_FEATURE_UNSUPPORTED", path+".state", "profile does not allow an explicit absent state for this feature")
			}
			if value.Value != "" || value.Variable != nil {
				return invalid("INVALID_ABSENT_FEATURE", path, "absent feature cannot carry a value or variable")
			}
		case StateKnown:
			if value.Value == "" || value.Variable != nil {
				return invalid("INVALID_KNOWN_FEATURE", path, "known feature requires one value and cannot carry a variable")
			}
			if !validToken(string(value.Value)) {
				return invalid("INVALID_FEATURE_VALUE_ID", path+".value", "feature value ID is malformed or exceeds its bound")
			}
			if !containsValue(dimension.Values, value.Value) {
				return unsupported("UNSUPPORTED_FEATURE_VALUE", path+".value", "feature value is not declared by this language profile")
			}
		case StateUnknown:
			if !dimension.AllowUnknown {
				return unsupported("UNKNOWN_FEATURE_UNSUPPORTED", path+".state", "profile does not allow unknown state for this feature")
			}
			if value.Value != "" {
				return invalid("INVALID_UNKNOWN_FEATURE", path, "unknown feature cannot carry a concrete value")
			}
			if value.Variable != nil && (!validToken(string(value.Variable.ID)) || !validToken(string(value.Variable.Scope))) {
				return invalid("INVALID_FEATURE_VARIABLE", path+".variable", "variable ID and scope must be bounded local tokens")
			}
		default:
			return invalid("INVALID_FEATURE_STATE", path+".state", "feature state must be absent, known or unknown")
		}
	}
	return nil
}

// ValidateFeatureRequest validates a bundle and requires every profile
// dimension marked Required to have a known supported value.
func ValidateFeatureRequest(profile *LanguageFeatureProfile, bundle FeatureBundle) error {
	if err := ValidateBundle(profile, bundle); err != nil {
		return err
	}
	values := make(map[FeatureID]FeatureValue, len(bundle.Values))
	for _, value := range bundle.Values {
		values[value.Feature] = value
	}
	for _, dimension := range profile.dimensions {
		if !dimension.Required {
			continue
		}
		value, exists := values[dimension.ID]
		if !exists || value.State == StateAbsent {
			return invalid("MISSING_REQUIRED_FEATURE", "features."+string(dimension.ID), "required feature is absent")
		}
		if value.State == StateUnknown {
			return incomplete("UNKNOWN_REQUIRED_FEATURE", "features."+string(dimension.ID), "required feature value is unresolved")
		}
	}
	return nil
}

func containsValue(values []FeatureValueID, candidate FeatureValueID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func cloneValues(values []FeatureValue) []FeatureValue {
	result := make([]FeatureValue, len(values))
	for index, value := range values {
		result[index] = value
		if value.Variable != nil {
			variable := *value.Variable
			result[index].Variable = &variable
		}
	}
	return result
}

func cloneBundle(source FeatureBundle) FeatureBundle {
	return FeatureBundle{Language: source.Language, ProfileVersion: source.ProfileVersion, Values: cloneValues(source.Values)}
}

func failureUnknownProfile() error {
	return failureUnknown("FEATURE_PROFILE_UNAVAILABLE", "profile", "language feature profile is unavailable")
}
