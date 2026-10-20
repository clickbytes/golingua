package features

import "sort"

// UnifyFeatureBundles merges two compatible bundles into a new immutable
// value. Optional constraints govern only omitted dimensions. Explicit absent
// and unknown states are never treated as missing or as concrete matches.
func UnifyFeatureBundles(profile *LanguageFeatureProfile, left, right FeatureBundle, constraints []FeatureConstraint) (FeatureBundle, error) {
	if err := ValidateBundle(profile, left); err != nil {
		return FeatureBundle{}, err
	}
	if err := ValidateBundle(profile, right); err != nil {
		return FeatureBundle{}, err
	}
	if len(constraints) > MaxDimensions {
		return FeatureBundle{}, resourceLimit("constraints", len(constraints), MaxDimensions)
	}
	constraintByID := make(map[FeatureID]FeatureConstraint, len(constraints))
	for index, constraint := range constraints {
		path := "constraints[" + itoa(index) + "]"
		if !validToken(string(constraint.Feature)) {
			return FeatureBundle{}, invalid("INVALID_FEATURE_ID", path+".feature", "constraint feature ID is malformed")
		}
		if _, exists := profile.Lookup(constraint.Feature); !exists {
			return FeatureBundle{}, unsupported("UNSUPPORTED_FEATURE_DIMENSION", path+".feature", "constraint feature is not declared by this profile")
		}
		if _, exists := constraintByID[constraint.Feature]; exists {
			return FeatureBundle{}, invalid("DUPLICATE_FEATURE_CONSTRAINT", path+".feature", "feature constraint is duplicated")
		}
		constraintByID[constraint.Feature] = constraint
	}
	leftValues := indexValues(left.Values)
	rightValues := indexValues(right.Values)
	featureSet := make(map[FeatureID]struct{}, len(leftValues)+len(rightValues)+len(constraintByID))
	for feature := range leftValues {
		featureSet[feature] = struct{}{}
	}
	for feature := range rightValues {
		featureSet[feature] = struct{}{}
	}
	for feature := range constraintByID {
		featureSet[feature] = struct{}{}
	}
	if len(featureSet) > MaxBundleValues {
		return FeatureBundle{}, resourceLimit("bundle_values", len(featureSet), MaxBundleValues)
	}
	features := make([]FeatureID, 0, len(featureSet))
	for feature := range featureSet {
		features = append(features, feature)
	}
	sort.Slice(features, func(i, j int) bool { return features[i] < features[j] })
	result := FeatureBundle{Language: profile.language, ProfileVersion: profile.version, Values: make([]FeatureValue, 0, len(features))}
	for _, feature := range features {
		leftValue, hasLeft := leftValues[feature]
		rightValue, hasRight := rightValues[feature]
		constraint, constrained := constraintByID[feature]
		if constrained && !constraint.Optional && (!hasLeft || !hasRight) {
			return FeatureBundle{}, invalid("MISSING_REQUIRED_FEATURE", "features."+string(feature), "required unification feature must be present in both bundles")
		}
		if !hasLeft && !hasRight {
			continue
		}
		if !hasLeft {
			result.Values = append(result.Values, cloneValues([]FeatureValue{rightValue})[0])
			continue
		}
		if !hasRight {
			result.Values = append(result.Values, cloneValues([]FeatureValue{leftValue})[0])
			continue
		}
		merged, err := unifyValue(leftValue, rightValue)
		if err != nil {
			return FeatureBundle{}, err
		}
		result.Values = append(result.Values, merged)
	}
	if err := ValidateBundle(profile, result); err != nil {
		return FeatureBundle{}, err
	}
	return cloneBundle(result), nil
}

func indexValues(values []FeatureValue) map[FeatureID]FeatureValue {
	result := make(map[FeatureID]FeatureValue, len(values))
	for _, value := range values {
		result[value.Feature] = value
	}
	return result
}

func unifyValue(left, right FeatureValue) (FeatureValue, error) {
	if left.State == StateAbsent || right.State == StateAbsent {
		if left.State == StateAbsent && right.State == StateAbsent {
			return FeatureValue{Feature: left.Feature, State: StateAbsent}, nil
		}
		return FeatureValue{}, unsupported("FEATURE_STATE_CONFLICT", "features."+string(left.Feature), "explicit absent state conflicts with a present or unknown value")
	}
	if left.State == StateKnown && right.State == StateKnown {
		if left.Value != right.Value {
			return FeatureValue{}, unsupported("FEATURE_CONFLICT", "features."+string(left.Feature), "known feature values conflict")
		}
		return FeatureValue{Feature: left.Feature, State: StateKnown, Value: left.Value}, nil
	}
	if left.State == StateUnknown && right.State == StateKnown {
		return cloneValues([]FeatureValue{left})[0], nil
	}
	if left.State == StateKnown && right.State == StateUnknown {
		return cloneValues([]FeatureValue{right})[0], nil
	}
	if left.State == StateUnknown && right.State == StateUnknown {
		if left.Variable != nil && right.Variable != nil && *left.Variable != *right.Variable {
			return FeatureValue{}, unsupported("FEATURE_VARIABLE_SCOPE_CONFLICT", "features."+string(left.Feature), "unknown values refer to distinct scoped variables")
		}
		result := FeatureValue{Feature: left.Feature, State: StateUnknown}
		if left.Variable != nil {
			variable := *left.Variable
			result.Variable = &variable
		} else if right.Variable != nil {
			variable := *right.Variable
			result.Variable = &variable
		}
		return result, nil
	}
	return FeatureValue{}, unsupported("FEATURE_CONFLICT", "features."+string(left.Feature), "feature values cannot be unified")
}
