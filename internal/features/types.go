// Package features defines language-local feature profiles, request bundles
// and non-mutating unification. Feature inventories do not belong in CORE.
package features

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/status"
)

const SchemaVersion = 1

type FeatureID string
type FeatureValueID string
type VariableID string
type VariableScopeID string
type LanguageID string
type ProfileVersion string

type ValueState string

const (
	StateAbsent  ValueState = "absent"
	StateKnown   ValueState = "known"
	StateUnknown ValueState = "unknown"
)

// VariableRef identifies an unresolved value only within its explicit scope.
type VariableRef struct {
	ID    VariableID      `json:"id"`
	Scope VariableScopeID `json:"scope"`
}

// FeatureValue explicitly separates absent, known and unresolved values.
// Value is present only for StateKnown; Variable is optional and applies only
// to StateUnknown.
type FeatureValue struct {
	Feature  FeatureID      `json:"feature"`
	State    ValueState     `json:"state"`
	Value    FeatureValueID `json:"value,omitempty"`
	Variable *VariableRef   `json:"variable,omitempty"`
}

// FeatureBundle carries values under one language profile version.
type FeatureBundle struct {
	Language       LanguageID     `json:"language"`
	ProfileVersion ProfileVersion `json:"profile_version"`
	Values         []FeatureValue `json:"values"`
}

// FeatureConstraint marks whether a feature may be omitted during unification.
// Optional affects omission only; an explicit absent value is still distinct.
type FeatureConstraint struct {
	Feature  FeatureID `json:"feature"`
	Optional bool      `json:"optional,omitempty"`
}

type FeatureError struct{ result status.Result }

func (failure *FeatureError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "feature operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *FeatureError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func failure(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &FeatureError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

func invalid(code, path, message string) error {
	return failure(status.OutcomeInvalid, code, path, message)
}

func unsupported(code, path, message string) error {
	return failure(status.OutcomeUnsupported, code, path, message)
}

func incomplete(code, path, message string) error {
	return failure(status.OutcomeIncomplete, code, path, message)
}

func failureUnknown(code, path, message string) error {
	return failure(status.OutcomeUnknown, code, path, message)
}

func resourceLimit(name string, value, maximum int) error {
	return failure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+name, fmt.Sprintf("feature %s limit exceeded: %d > %d", name, value, maximum))
}
