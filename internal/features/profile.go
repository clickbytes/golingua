package features

import (
	"sort"
	"strings"
)

const (
	MaxProfileBytes = 1 << 20
	MaxDimensions   = 64
	MaxValuesPerDim = 128
	MaxBundleValues = 64
	MaxTokenBytes   = 64
)

type FeatureDimension struct {
	ID           FeatureID        `json:"id"`
	Required     bool             `json:"required,omitempty"`
	AllowAbsent  bool             `json:"allow_absent,omitempty"`
	AllowUnknown bool             `json:"allow_unknown,omitempty"`
	Values       []FeatureValueID `json:"values"`
}

type ProfileDocument struct {
	SchemaVersion int                `json:"schema_version"`
	Language      LanguageID         `json:"language"`
	Version       ProfileVersion     `json:"version"`
	Dimensions    []FeatureDimension `json:"dimensions"`
}

// LanguageFeatureProfile is an immutable per-language inventory.
type LanguageFeatureProfile struct {
	language   LanguageID
	version    ProfileVersion
	dimensions []FeatureDimension
	byID       map[FeatureID]int
}

// BuildProfile validates and freezes one language-local dimension inventory.
func BuildProfile(document ProfileDocument) (*LanguageFeatureProfile, error) {
	if document.SchemaVersion != SchemaVersion {
		return nil, invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported feature profile version")
	}
	if !validLanguage(string(document.Language)) {
		return nil, invalid("INVALID_LANGUAGE_ID", "language", "language tag must use canonical lowercase ASCII")
	}
	if !validToken(string(document.Version)) {
		return nil, invalid("INVALID_PROFILE_VERSION", "version", "profile version is malformed or exceeds its bound")
	}
	if len(document.Dimensions) == 0 {
		return nil, invalid("EMPTY_FEATURE_PROFILE", "dimensions", "profile must declare at least one dimension")
	}
	if len(document.Dimensions) > MaxDimensions {
		return nil, resourceLimit("dimensions", len(document.Dimensions), MaxDimensions)
	}
	canonical := make([]FeatureDimension, len(document.Dimensions))
	byID := make(map[FeatureID]int, len(document.Dimensions))
	for index, dimension := range document.Dimensions {
		path := "dimensions[" + itoa(index) + "]"
		if !validToken(string(dimension.ID)) {
			return nil, invalid("INVALID_FEATURE_ID", path+".id", "feature ID is malformed or exceeds its bound")
		}
		if _, exists := byID[dimension.ID]; exists {
			return nil, invalid("DUPLICATE_FEATURE_ID", path+".id", "feature dimension is duplicated")
		}
		if len(dimension.Values) == 0 {
			return nil, invalid("EMPTY_FEATURE_VALUES", path+".values", "feature dimension must declare at least one value")
		}
		if len(dimension.Values) > MaxValuesPerDim {
			return nil, resourceLimit("values_per_dimension", len(dimension.Values), MaxValuesPerDim)
		}
		values := append([]FeatureValueID(nil), dimension.Values...)
		seen := make(map[FeatureValueID]struct{}, len(values))
		for valueIndex, value := range values {
			if !validToken(string(value)) {
				return nil, invalid("INVALID_FEATURE_VALUE_ID", path+".values["+itoa(valueIndex)+"]", "feature value ID is malformed or exceeds its bound")
			}
			if _, exists := seen[value]; exists {
				return nil, invalid("DUPLICATE_FEATURE_VALUE_ID", path+".values", "feature value is duplicated")
			}
			seen[value] = struct{}{}
		}
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		canonical[index] = FeatureDimension{ID: dimension.ID, Required: dimension.Required, AllowAbsent: dimension.AllowAbsent, AllowUnknown: dimension.AllowUnknown, Values: values}
		byID[dimension.ID] = index
	}
	sort.Slice(canonical, func(i, j int) bool { return canonical[i].ID < canonical[j].ID })
	byID = make(map[FeatureID]int, len(canonical))
	for index, dimension := range canonical {
		byID[dimension.ID] = index
	}
	return &LanguageFeatureProfile{language: document.Language, version: document.Version, dimensions: canonical, byID: byID}, nil
}

func (profile *LanguageFeatureProfile) Language() LanguageID {
	if profile == nil {
		return ""
	}
	return profile.language
}

func (profile *LanguageFeatureProfile) Version() ProfileVersion {
	if profile == nil {
		return ""
	}
	return profile.version
}

func (profile *LanguageFeatureProfile) Dimensions() []FeatureDimension {
	if profile == nil {
		return nil
	}
	result := make([]FeatureDimension, len(profile.dimensions))
	for index, dimension := range profile.dimensions {
		result[index] = cloneDimension(dimension)
	}
	return result
}

func (profile *LanguageFeatureProfile) Lookup(id FeatureID) (FeatureDimension, bool) {
	if profile == nil {
		return FeatureDimension{}, false
	}
	index, exists := profile.byID[id]
	if !exists {
		return FeatureDimension{}, false
	}
	return cloneDimension(profile.dimensions[index]), true
}

func cloneDimension(source FeatureDimension) FeatureDimension {
	result := source
	result.Values = append([]FeatureValueID(nil), source.Values...)
	return result
}

func validLanguage(value string) bool {
	if len(value) == 0 || len(value) > 35 || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, character := range part {
			if !(character >= 'a' && character <= 'z') && !(index > 0 && character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

func validToken(value string) bool {
	if len(value) == 0 || len(value) > MaxTokenBytes || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	position := len(buffer)
	for value > 0 {
		position--
		buffer[position] = byte('0' + value%10)
		value /= 10
	}
	return string(buffer[position:])
}
