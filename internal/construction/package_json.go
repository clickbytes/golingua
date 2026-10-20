package construction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	ConstructionPackageSchemaVersion = 1
	MaxConstructionPackageBytes      = 1 << 20
)

// ConstructionPackage is a bounded data package containing one language-local
// family and its variants. Compilation validates the templates only; it does
// not license any sense or activate analysis or realization.
type ConstructionPackage struct {
	SchemaVersion int                 `json:"schema_version"`
	Language      features.LanguageID `json:"language"`
	Family        ConstructionFamily  `json:"family"`
	Variants      []Variant           `json:"variants"`
}

// CompiledConstructionPackage contains the validated package and its stable
// variant index. The returned package is copied from the decoded input.
type CompiledConstructionPackage struct {
	Package ConstructionPackage
	Index   *VariantIndex
}

// DecodeConstructionPackage rejects oversized, ambiguous, or unknown JSON
// and compiles each variant against the package's exact family.
func DecodeConstructionPackage(data []byte) (CompiledConstructionPackage, error) {
	if len(data) == 0 || len(data) > MaxConstructionPackageBytes {
		return CompiledConstructionPackage{}, invalid("INVALID_CONSTRUCTION_PACKAGE_SIZE", "package", "construction package must be non-empty and within the byte limit")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return CompiledConstructionPackage{}, invalid("DUPLICATE_JSON_KEY", "package", err.Error())
	}
	var wire constructionPackageWire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return CompiledConstructionPackage{}, invalid("INVALID_CONSTRUCTION_PACKAGE_JSON", "package", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return CompiledConstructionPackage{}, invalid("TRAILING_CONSTRUCTION_PACKAGE_DATA", "package", "construction package must contain exactly one JSON value")
	}
	if wire.SchemaVersion != ConstructionPackageSchemaVersion {
		return CompiledConstructionPackage{}, invalid("CONSTRUCTION_PACKAGE_SCHEMA_VERSION", "schema_version", "unsupported construction package schema version")
	}
	familyID, err := ids.ParseConstructionID(wire.Family.ID)
	if err != nil {
		return CompiledConstructionPackage{}, invalid("INVALID_CONSTRUCTION_FAMILY_ID", "family.id", err.Error())
	}
	family := ConstructionFamily{ID: familyID, Language: wire.Family.Language, SlotSpecs: wire.Family.SlotSpecs, Constraints: wire.Family.Constraints, Operators: wire.Family.Operators, Capabilities: wire.Family.Capabilities, Provenance: wire.Family.Provenance}
	if wire.Language != family.Language {
		return CompiledConstructionPackage{}, unsupported("CONSTRUCTION_PACKAGE_LANGUAGE_MISMATCH", "language", "package and family languages must match")
	}
	variants := make([]Variant, len(wire.Variants))
	for i, value := range wire.Variants {
		id, parseErr := ParseVariantID(value.ID)
		if parseErr != nil {
			return CompiledConstructionPackage{}, invalid("INVALID_VARIANT_ID", fmt.Sprintf("variants[%d].id", i), parseErr.Error())
		}
		variantFamily, parseErr := ids.ParseConstructionID(value.FamilyID)
		if parseErr != nil {
			return CompiledConstructionPackage{}, invalid("INVALID_VARIANT_FAMILY_ID", fmt.Sprintf("variants[%d].family_id", i), parseErr.Error())
		}
		variants[i] = Variant{ID: id, FamilyID: variantFamily, Sequence: value.Sequence, Conditions: value.Conditions, Markedness: value.Markedness, Priority: value.Priority, DirectionPolicy: value.DirectionPolicy, Provenance: value.Provenance}
	}
	index, err := CompileVariantIndex(family, variants)
	if err != nil {
		return CompiledConstructionPackage{}, err
	}
	return CompiledConstructionPackage{Package: ConstructionPackage{SchemaVersion: wire.SchemaVersion, Language: wire.Language, Family: family, Variants: variants}, Index: index}, nil
}

type constructionPackageWire struct {
	SchemaVersion int                       `json:"schema_version"`
	Language      features.LanguageID       `json:"language"`
	Family        constructionFamilyWire    `json:"family"`
	Variants      []constructionVariantWire `json:"variants"`
}

type constructionFamilyWire struct {
	ID           string              `json:"id"`
	Language     features.LanguageID `json:"language"`
	SlotSpecs    []SlotSpec          `json:"slot_specs"`
	Constraints  []FamilyConstraint  `json:"constraints,omitempty"`
	Operators    []OperatorID        `json:"operators,omitempty"`
	Capabilities []Capability        `json:"capabilities,omitempty"`
	Provenance   provenanceSourceRef `json:"provenance"`
}

type constructionVariantWire struct {
	ID              string              `json:"id"`
	FamilyID        string              `json:"family_id"`
	Sequence        []SequenceToken     `json:"sequence"`
	Conditions      []Condition         `json:"conditions,omitempty"`
	Markedness      Markedness          `json:"markedness"`
	Priority        int                 `json:"priority"`
	DirectionPolicy DirectionPolicy     `json:"direction_policy"`
	Provenance      provenanceSourceRef `json:"provenance"`
}

// Alias avoids defining custom ID JSON methods on graph identity types.
type provenanceSourceRef = provenance.SourceRef

// rejectDuplicateJSONKeys walks the complete JSON token stream, including
// nested objects, before normal unmarshalling (which otherwise keeps the last
// duplicate key).
func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() error
	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return keyErr
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate object key %q", key)
				}
				seen[key] = struct{}{}
				if err := walk(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := walk(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter")
		}
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}
