package status

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	MaxTraceSteps   = 32
	MaxTraceIDBytes = 128
)

var ErrInvalidTrace = errors.New("invalid trace")

// Trace is an optional, deliberately short provenance trail. Truncated marks
// that additional steps were omitted at the configured trace bound.
type Trace struct {
	Steps     []TraceStep `json:"steps"`
	Truncated bool        `json:"truncated,omitempty"`
}

// TraceStep identifies the rule and, when applicable, the lexical sense and
// construction used by an operation. It does not carry arbitrary log text.
type TraceStep struct {
	RuleID         string `json:"rule_id,omitempty"`
	SenseID        string `json:"sense_id,omitempty"`
	ConstructionID string `json:"construction_id,omitempty"`
}

// Validate enforces the trace length and requires each step to identify at
// least one known rule, sense, or construction.
func (t Trace) Validate() error {
	if len(t.Steps) > MaxTraceSteps {
		return fmt.Errorf("%w: %d steps exceeds limit %d", ErrInvalidTrace, len(t.Steps), MaxTraceSteps)
	}
	for i, step := range t.Steps {
		ids := [...]string{step.RuleID, step.SenseID, step.ConstructionID}
		hasID := false
		for _, id := range ids {
			if id == "" {
				continue
			}
			hasID = true
			if len(id) > MaxTraceIDBytes || !utf8.ValidString(id) {
				return fmt.Errorf("%w: step %d has invalid or oversized ID", ErrInvalidTrace, i)
			}
		}
		if !hasID {
			return fmt.Errorf("%w: step %d has no provenance ID", ErrInvalidTrace, i)
		}
	}
	return nil
}
