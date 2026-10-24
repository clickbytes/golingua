// Package construction models bounded language-local construction families.
package construction

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxSlotsPerFamily        = 64
	MaxConstraintsPerFamily  = 256
	MaxFeaturesPerConstraint = 16
	MaxFamilyItems           = 4_096
	MaxFamilyTextBytes       = 256
	MaxSlotCardinality       = 32
)

// SlotCategory identifies a structural phrase slot, not a semantic role.
type SlotCategory string

const (
	CategoryNP SlotCategory = "NP"
	CategoryVP SlotCategory = "VP"
	CategoryAP SlotCategory = "AP"
	CategoryPP SlotCategory = "PP"
)

// IsPredicate reports whether a category may carry the construction predicate.
func (category SlotCategory) IsPredicate() bool {
	return category == CategoryVP || category == CategoryAP
}

func (category SlotCategory) Valid() bool {
	return category == CategoryNP || category.IsPredicate() || category == CategoryPP
}

// Optionality states whether a slot must be occupied or may be omitted.
type Optionality string

const (
	SlotRequired Optionality = "required"
	SlotOptional Optionality = "optional"
)

// Cardinality bounds how many phrases a slot may contain.
type Cardinality struct {
	Minimum int `json:"minimum"`
	Maximum int `json:"maximum"`
}

// ArgumentBinding links one phrase slot to a semantic role. The selected
// sense-specific frame supplies that role's argument policy; no
// grammatical-function default is inferred from the role.
type ArgumentBinding struct {
	SemanticRole meaning.Role `json:"semantic_role"`
}

// SlotSpec describes one ordered NP, VP, or PP position in a family.
type SlotSpec struct {
	ID              string           `json:"id"`
	Position        int              `json:"position"`
	Category        SlotCategory     `json:"category"`
	Cardinality     Cardinality      `json:"cardinality"`
	Optionality     Optionality      `json:"optionality"`
	ArgumentBinding *ArgumentBinding `json:"argument_binding,omitempty"`
	Case            string           `json:"case,omitempty"`
	Adposition      string           `json:"adposition,omitempty"`
}

// ConstraintStrength distinguishes an obligatory structural rule from a
// preference. This package stores constraints; a later compiler evaluates them.
type ConstraintStrength string

const (
	ConstraintHard ConstraintStrength = "hard"
	ConstraintSoft ConstraintStrength = "soft"
)

// ConstraintKind is a closed family-level relation between phrase slots.
type ConstraintKind string

const (
	ConstraintPrecedes  ConstraintKind = "precedes"
	ConstraintAgreement ConstraintKind = "agreement"
)

// FamilyConstraint records a bounded, typed relationship between slots.
type FamilyConstraint struct {
	ID             string               `json:"id"`
	Kind           ConstraintKind       `json:"kind"`
	Strength       ConstraintStrength   `json:"strength"`
	ControllerSlot string               `json:"controller_slot"`
	TargetSlot     string               `json:"target_slot"`
	Features       []features.FeatureID `json:"features,omitempty"`
}

func validToken(value string) bool {
	if value == "" || len(value) > MaxFamilyTextBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func validateSlot(slot SlotSpec, index int) error {
	path := fmt.Sprintf("family.slots[%d]", index)
	if !validToken(slot.ID) {
		return invalid("INVALID_SLOT_ID", path+".id", "slot ID must be a bounded lowercase token")
	}
	if !slot.Category.Valid() {
		return invalid("INVALID_SLOT_CATEGORY", path+".category", "slot category must be NP, VP, AP, or PP")
	}
	if slot.Position < 0 || slot.Position >= MaxSlotsPerFamily {
		return invalid("INVALID_SLOT_POSITION", path+".position", "slot position is outside the supported range")
	}
	if slot.Cardinality.Minimum < 0 || slot.Cardinality.Maximum < 1 || slot.Cardinality.Maximum > MaxSlotCardinality || slot.Cardinality.Minimum > slot.Cardinality.Maximum {
		return invalid("INVALID_SLOT_CARDINALITY", path+".cardinality", "slot cardinality must satisfy 0 <= minimum <= maximum <= package limit")
	}
	switch slot.Optionality {
	case SlotRequired:
		if slot.Cardinality.Minimum == 0 {
			return invalid("REQUIRED_SLOT_HAS_ZERO_MINIMUM", path+".optionality", "required slot must have a positive minimum cardinality")
		}
	case SlotOptional:
		if slot.Cardinality.Minimum != 0 {
			return invalid("OPTIONAL_SLOT_HAS_POSITIVE_MINIMUM", path+".optionality", "optional slot must allow zero occurrences")
		}
	default:
		return invalid("INVALID_SLOT_OPTIONALITY", path+".optionality", "optionality must be required or optional")
	}
	if slot.Case != "" && !validToken(slot.Case) {
		return invalid("INVALID_SLOT_CASE", path+".case", "case label must be a bounded lowercase token")
	}
	if slot.Adposition != "" && !validToken(slot.Adposition) {
		return invalid("INVALID_SLOT_ADPOSITION", path+".adposition", "adposition label must be a bounded lowercase token")
	}
	if slot.ArgumentBinding != nil {
		if !validToken(string(slot.ArgumentBinding.SemanticRole)) {
			return invalid("INVALID_ARGUMENT_BINDING", path+".argument_binding", "argument binding requires a bounded semantic role")
		}
	}
	return nil
}

func invalid(code, path, message string) error {
	return constructionError(status.OutcomeInvalid, code, path, message)
}

func unknown(code, path, message string) error {
	return constructionError(status.OutcomeUnknown, code, path, message)
}

func unsupported(code, path, message string) error {
	return constructionError(status.OutcomeUnsupported, code, path, message)
}

func resourceLimit(resource string, value, maximum int) error {
	return constructionError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+resource, fmt.Sprintf("construction %s limit exceeded: %d > %d", resource, value, maximum))
}

func constructionError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &Error{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

// Error carries a typed outcome and bounded diagnostic.
type Error struct{ result status.Result }

func (failure *Error) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "construction operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed outcome.
func (failure *Error) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}
