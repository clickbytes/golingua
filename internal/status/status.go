// Package status defines bounded, typed outcomes shared by the CLI and future
// language runtime packages.
package status

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// OutcomeKind describes what the operation established. The values are
// intentionally distinct: missing knowledge is not lack of support, and
// either is different from invalid input or an exhausted resource budget.
type OutcomeKind string

const (
	OutcomeOK            OutcomeKind = "ok"
	OutcomeUnsupported   OutcomeKind = "unsupported"
	OutcomeUnknown       OutcomeKind = "unknown"
	OutcomeAmbiguous     OutcomeKind = "ambiguous"
	OutcomeIncomplete    OutcomeKind = "incomplete"
	OutcomeInvalid       OutcomeKind = "invalid"
	OutcomeResourceLimit OutcomeKind = "resource_limit"
	OutcomeError         OutcomeKind = "error"
)

const (
	MaxResultTextBytes = 64 * 1024
	MaxDiagnostics     = 16
)

var ErrInvalidResult = errors.New("invalid status result")

// Result is the stable outer shape for an operation outcome. Text may contain
// a completed or partial human-readable value; diagnostics explain refusals
// and failures. A non-success outcome must carry at least one diagnostic.
type Result struct {
	Status      OutcomeKind  `json:"status"`
	Text        string       `json:"text,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
	Trace       *Trace       `json:"trace,omitempty"`
}

// NewResult constructs a result and validates every bounded field before it
// can be serialized or handed to another package.
func NewResult(kind OutcomeKind, text string, diagnostics []Diagnostic, trace *Trace) (Result, error) {
	result := Result{
		Status:      kind,
		Text:        text,
		Diagnostics: append([]Diagnostic(nil), diagnostics...),
		Trace:       trace,
	}
	if err := result.Validate(); err != nil {
		return Result{}, err
	}
	return result, nil
}

// Validate rejects unknown outcomes, unbounded payloads, and failure results
// that omit their explanation.
func (r Result) Validate() error {
	if !r.Status.Valid() {
		return fmt.Errorf("%w: unknown outcome %q", ErrInvalidResult, r.Status)
	}
	if !utf8.ValidString(r.Text) || len(r.Text) > MaxResultTextBytes {
		return fmt.Errorf("%w: result text is invalid UTF-8 or exceeds %d bytes", ErrInvalidResult, MaxResultTextBytes)
	}
	if len(r.Diagnostics) > MaxDiagnostics {
		return fmt.Errorf("%w: %d diagnostics exceeds limit %d", ErrInvalidResult, len(r.Diagnostics), MaxDiagnostics)
	}
	if r.Status != OutcomeOK && len(r.Diagnostics) == 0 {
		return fmt.Errorf("%w: outcome %q requires a diagnostic", ErrInvalidResult, r.Status)
	}
	for i, diagnostic := range r.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return fmt.Errorf("%w: diagnostic %d: %v", ErrInvalidResult, i, err)
		}
	}
	if r.Trace != nil {
		if err := r.Trace.Validate(); err != nil {
			return fmt.Errorf("%w: trace: %v", ErrInvalidResult, err)
		}
	}
	return nil
}

// Valid reports whether kind is one of the explicitly supported outcomes.
func (kind OutcomeKind) Valid() bool {
	switch kind {
	case OutcomeOK, OutcomeUnsupported, OutcomeUnknown, OutcomeAmbiguous,
		OutcomeIncomplete, OutcomeInvalid, OutcomeResourceLimit, OutcomeError:
		return true
	default:
		return false
	}
}
