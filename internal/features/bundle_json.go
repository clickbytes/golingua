package features

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
)

const MaxBundleBytes = 64 << 10

func (bundle *FeatureBundle) UnmarshalJSON(data []byte) error {
	if err := requireFields(data, []string{"language", "profile_version", "values"}, []string{"values"}); err != nil {
		return err
	}
	type wireBundle FeatureBundle
	var decoded wireBundle
	if err := decodeExact(data, map[string]bool{"language": true, "profile_version": true, "values": true}, &decoded); err != nil {
		return err
	}
	*bundle = FeatureBundle(decoded)
	return nil
}

func (value *FeatureValue) UnmarshalJSON(data []byte) error {
	if err := requireFields(data, []string{"feature", "state"}, nil); err != nil {
		return err
	}
	if err := rejectNull(data, "value", "variable"); err != nil {
		return err
	}
	type wireValue FeatureValue
	var decoded wireValue
	allowed := map[string]bool{"feature": true, "state": true, "value": true, "variable": true}
	if err := decodeExact(data, allowed, &decoded); err != nil {
		return err
	}
	*value = FeatureValue(decoded)
	return nil
}

func (variable *VariableRef) UnmarshalJSON(data []byte) error {
	if err := requireFields(data, []string{"id", "scope"}, nil); err != nil {
		return err
	}
	type wireVariable VariableRef
	var decoded wireVariable
	if err := decodeExact(data, map[string]bool{"id": true, "scope": true}, &decoded); err != nil {
		return err
	}
	*variable = VariableRef(decoded)
	return nil
}

// DecodeBundle decodes one strict, bounded FeatureBundle and validates it
// against its declared language profile.
func DecodeBundle(data []byte, profile *LanguageFeatureProfile) (FeatureBundle, error) {
	if len(data) > MaxBundleBytes {
		return FeatureBundle{}, resourceLimit("bundle_bytes", len(data), MaxBundleBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var bundle FeatureBundle
	if err := decoder.Decode(&bundle); err != nil {
		return FeatureBundle{}, invalid("INVALID_FEATURE_BUNDLE_JSON", "document", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return FeatureBundle{}, invalid("TRAILING_JSON", "document", "expected exactly one JSON value")
	}
	if err := ValidateBundle(profile, bundle); err != nil {
		return FeatureBundle{}, err
	}
	return cloneBundle(bundle), nil
}

// EncodeBundle emits a canonical bounded bundle representation.
func EncodeBundle(profile *LanguageFeatureProfile, bundle FeatureBundle) ([]byte, error) {
	if err := ValidateBundle(profile, bundle); err != nil {
		return nil, err
	}
	canonical := cloneBundle(bundle)
	sort.Slice(canonical.Values, func(i, j int) bool { return canonical.Values[i].Feature < canonical.Values[j].Feature })
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return nil, invalid("INVALID_FEATURE_BUNDLE_JSON", "document", err.Error())
	}
	if len(encoded) > MaxBundleBytes {
		return nil, resourceLimit("bundle_bytes", len(encoded), MaxBundleBytes)
	}
	return encoded, nil
}
