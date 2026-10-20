package features

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

func (document *ProfileDocument) UnmarshalJSON(data []byte) error {
	if err := requireFields(data, []string{"schema_version", "language", "version", "dimensions"}, []string{"schema_version", "language", "version", "dimensions"}); err != nil {
		return err
	}
	type wireProfile ProfileDocument
	var decoded wireProfile
	if err := decodeExact(data, map[string]bool{"schema_version": true, "language": true, "version": true, "dimensions": true}, &decoded); err != nil {
		return err
	}
	*document = ProfileDocument(decoded)
	return nil
}

func (dimension *FeatureDimension) UnmarshalJSON(data []byte) error {
	if err := rejectNull(data, "values", "required", "allow_absent", "allow_unknown"); err != nil {
		return err
	}
	type wireDimension FeatureDimension
	var decoded wireDimension
	allowed := map[string]bool{"id": true, "required": true, "allow_absent": true, "allow_unknown": true, "values": true}
	if err := decodeExact(data, allowed, &decoded); err != nil {
		return err
	}
	*dimension = FeatureDimension(decoded)
	return nil
}

// DecodeProfile strictly decodes and freezes one bounded profile document.
func DecodeProfile(data []byte) (*LanguageFeatureProfile, error) {
	if len(data) > MaxProfileBytes {
		return nil, resourceLimit("profile_bytes", len(data), MaxProfileBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document ProfileDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, invalid("INVALID_FEATURE_PROFILE_JSON", "document", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, invalid("TRAILING_JSON", "document", "expected exactly one JSON value")
	}
	return BuildProfile(document)
}

// EncodeProfile returns a deterministic validated profile JSON document.
func EncodeProfile(document ProfileDocument) ([]byte, error) {
	profile, err := BuildProfile(document)
	if err != nil {
		return nil, err
	}
	canonical := canonicalizeProfile(ProfileDocument{SchemaVersion: SchemaVersion, Language: profile.language, Version: profile.version, Dimensions: profile.Dimensions()})
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return nil, invalid("INVALID_FEATURE_PROFILE_JSON", "document", err.Error())
	}
	if len(encoded) > MaxProfileBytes {
		return nil, resourceLimit("profile_bytes", len(encoded), MaxProfileBytes)
	}
	return encoded, nil
}

func decodeExact(data []byte, allowed map[string]bool, target any) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	for field := range fields {
		if !allowed[field] {
			return fmt.Errorf("unknown or mis-cased field %q", field)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func rejectNull(data []byte, names ...string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	for _, name := range names {
		if value, exists := fields[name]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("field %q cannot be null", name)
		}
	}
	return nil
}

func requireFields(data []byte, required, arrays []string) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	for _, name := range required {
		value, exists := fields[name]
		if !exists {
			return fmt.Errorf("required field %q is missing", name)
		}
		for _, arrayName := range arrays {
			if name == arrayName && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return fmt.Errorf("array field %q cannot be null", name)
			}
		}
	}
	return nil
}

func canonicalizeProfile(document ProfileDocument) ProfileDocument {
	result := document
	result.Dimensions = make([]FeatureDimension, len(document.Dimensions))
	for index, dimension := range document.Dimensions {
		result.Dimensions[index] = cloneDimension(dimension)
		sort.Slice(result.Dimensions[index].Values, func(i, j int) bool { return result.Dimensions[index].Values[i] < result.Dimensions[index].Values[j] })
	}
	sort.Slice(result.Dimensions, func(i, j int) bool { return result.Dimensions[i].ID < result.Dimensions[j].ID })
	return result
}
