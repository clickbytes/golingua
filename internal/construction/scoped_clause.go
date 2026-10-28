package construction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	ScopedClauseSchemaVersion       = 1
	maxScopedClausePlanBytes        = 64 * 1024
	maxScopedClauseVariants         = 8
	maxScopedClauseParts            = 128
	maxScopedClauseLiteralBytes     = 256
	maxScopedClauseConstraints      = 32
	maxScopedClauseConstraintValues = 16
	maxScopedClauseIdentifierBytes  = 64
)

// ScopedClausePlan describes bounded, language-local surface patterns while
// leaving lexical meaning and interpretation to other engine layers.
type ScopedClausePlan struct {
	SchemaVersion int                   `json:"schema_version"`
	Language      features.LanguageID   `json:"language"`
	Provenance    provenance.SourceRef  `json:"provenance"`
	Variants      []ScopedClauseVariant `json:"variants"`
}

type ScopedClauseVariant struct {
	ID            string                          `json:"id"`
	Negation      string                          `json:"negation"`
	SubjectBinder string                          `json:"subject_binder"`
	CasePolicy    string                          `json:"case_policy"`
	Parts         []ScopedClausePart              `json:"parts"`
	Constraints   []ScopedClauseVariantConstraint `json:"constraints,omitempty"`
}

// ScopedClauseVariantConstraint is a declarative restriction. It carries no
// inference behavior; consumers must enforce it against the bound feature.
type ScopedClauseVariantConstraint struct {
	Binder  string   `json:"binder"`
	Feature string   `json:"feature"`
	Values  []string `json:"values"`
}

// Exactly one of Literal and Slot is populated. Case applies to noun slots.
type ScopedClausePart struct {
	Literal string `json:"literal,omitempty"`
	Slot    string `json:"slot,omitempty"`
	Case    string `json:"case,omitempty"`
}

// DecodeScopedClausePlan strictly decodes one bounded plan for the requested
// language, rejecting ambiguous or unsupported structures before use.
func DecodeScopedClausePlan(data []byte, language features.LanguageID) (ScopedClausePlan, error) {
	if len(data) == 0 || len(data) > maxScopedClausePlanBytes {
		return ScopedClausePlan{}, fmt.Errorf("invalid scoped clause plan size")
	}
	if language == "" || strings.TrimSpace(string(language)) != string(language) {
		return ScopedClausePlan{}, fmt.Errorf("invalid requested language")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return ScopedClausePlan{}, fmt.Errorf("invalid scoped clause JSON keys: %w", err)
	}
	var plan ScopedClausePlan
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return ScopedClausePlan{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ScopedClausePlan{}, fmt.Errorf("scoped clause plan must contain one JSON value")
	}
	if plan.SchemaVersion != ScopedClauseSchemaVersion || plan.Language != language || len(plan.Variants) == 0 || len(plan.Variants) > maxScopedClauseVariants {
		return ScopedClausePlan{}, fmt.Errorf("invalid scoped clause plan header")
	}
	if err := plan.Provenance.Validate(); err != nil {
		return ScopedClausePlan{}, fmt.Errorf("invalid scoped clause provenance: %w", err)
	}
	seenIDs := make(map[string]struct{}, len(plan.Variants))
	seenPairs := make(map[string]struct{}, len(plan.Variants))
	for index := range plan.Variants {
		variant := &plan.Variants[index]
		id, err := ids.ParseConstructionID(variant.ID)
		if err != nil || id.String() != variant.ID {
			return ScopedClausePlan{}, fmt.Errorf("variant %d has invalid canonical construction id", index)
		}
		if _, exists := seenIDs[variant.ID]; exists {
			return ScopedClausePlan{}, fmt.Errorf("duplicate scoped clause construction id")
		}
		seenIDs[variant.ID] = struct{}{}
		if (variant.Negation != "outer" && variant.Negation != "inner") || (variant.SubjectBinder != "x" && variant.SubjectBinder != "y") {
			return ScopedClausePlan{}, fmt.Errorf("variant %d has invalid scope or subject binder", index)
		}
		if variant.CasePolicy != "marked" && variant.CasePolicy != "unmarked" {
			return ScopedClausePlan{}, fmt.Errorf("variant %d has invalid or missing case policy", index)
		}
		pair := variant.Negation + ":" + variant.SubjectBinder
		if _, exists := seenPairs[pair]; exists {
			return ScopedClausePlan{}, fmt.Errorf("duplicate scoped clause negation and subject pair")
		}
		seenPairs[pair] = struct{}{}
		if len(variant.Parts) == 0 || len(variant.Parts) > maxScopedClauseParts {
			return ScopedClausePlan{}, fmt.Errorf("variant %d has invalid part count", index)
		}
		slots := map[string]int{"x": 0, "y": 0, "event": 0}
		for partIndex, part := range variant.Parts {
			if (part.Literal != "") == (part.Slot != "") {
				return ScopedClausePlan{}, fmt.Errorf("variant %d part %d must contain exactly one literal or slot", index, partIndex)
			}
			if part.Literal != "" {
				if len(part.Literal) > maxScopedClauseLiteralBytes || !validScopedClauseLiteral(part.Literal) || strings.TrimSpace(part.Literal) != part.Literal {
					return ScopedClausePlan{}, fmt.Errorf("variant %d part %d has invalid literal", index, partIndex)
				}
				if part.Case != "" {
					return ScopedClausePlan{}, fmt.Errorf("literal part cannot carry a case")
				}
				continue
			}
			switch part.Slot {
			case "x", "y":
				if part.Case != "nominative" && part.Case != "genitive" && part.Case != "accusative" {
					return ScopedClausePlan{}, fmt.Errorf("noun slot must carry a supported case")
				}
				slots[part.Slot]++
			case "event":
				if part.Case != "" {
					return ScopedClausePlan{}, fmt.Errorf("event slot cannot carry a case")
				}
				slots[part.Slot]++
			default:
				return ScopedClausePlan{}, fmt.Errorf("variant %d part %d has invalid slot", index, partIndex)
			}
		}
		if slots["x"] == 0 || slots["y"] == 0 || slots["event"] != 1 {
			return ScopedClausePlan{}, fmt.Errorf("variant %d must contain x, y, and exactly one event slot", index)
		}
		if err := validateScopedClauseConstraints(variant.Constraints); err != nil {
			return ScopedClausePlan{}, fmt.Errorf("variant %d: %w", index, err)
		}
	}
	return plan, nil
}

func validateScopedClauseConstraints(constraints []ScopedClauseVariantConstraint) error {
	if len(constraints) > maxScopedClauseConstraints {
		return fmt.Errorf("too many variant constraints")
	}
	seen := make(map[string]struct{}, len(constraints))
	for _, constraint := range constraints {
		if (constraint.Binder != "x" && constraint.Binder != "y" && constraint.Binder != "event") || !validScopedClauseIdentifier(constraint.Feature) || len(constraint.Values) == 0 || len(constraint.Values) > maxScopedClauseConstraintValues {
			return fmt.Errorf("invalid variant constraint")
		}
		key := constraint.Binder + ":" + constraint.Feature
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate constraint for binder and feature")
		}
		seen[key] = struct{}{}
		values := make(map[string]struct{}, len(constraint.Values))
		for _, value := range constraint.Values {
			if !validScopedClauseIdentifier(value) {
				return fmt.Errorf("invalid constraint value")
			}
			if _, exists := values[value]; exists {
				return fmt.Errorf("duplicate constraint value")
			}
			values[value] = struct{}{}
		}
	}
	return nil
}

func validScopedClauseIdentifier(value string) bool {
	if len(value) == 0 || len(value) > maxScopedClauseIdentifierBytes || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}

func validScopedClauseLiteral(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
