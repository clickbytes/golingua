package status

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	DefaultMaxInputBytes = 16 * 1024
	DefaultMaxGraphDepth = 64
	DefaultMaxAnalyses   = 256
)

var ErrInvalidBounds = errors.New("invalid runtime bounds")
var ErrInvalidInput = errors.New("invalid runtime input")
var ErrInvalidMeasurement = errors.New("invalid runtime measurement")

// Bounds defines finite limits for work performed on one request. Limits are
// explicit and positive; zero never means unlimited.
type Bounds struct {
	MaxInputBytes int
	MaxGraphDepth int
	MaxAnalyses   int
}

// DefaultBounds returns conservative bootstrap limits for future analysis and
// realization operations.
func DefaultBounds() Bounds {
	return Bounds{
		MaxInputBytes: DefaultMaxInputBytes,
		MaxGraphDepth: DefaultMaxGraphDepth,
		MaxAnalyses:   DefaultMaxAnalyses,
	}
}

// Validate rejects missing or negative limits so a zero value cannot silently
// disable resource protection.
func (b Bounds) Validate() error {
	if b.MaxInputBytes <= 0 || b.MaxGraphDepth <= 0 || b.MaxAnalyses <= 0 {
		return fmt.Errorf("%w: all limits must be positive", ErrInvalidBounds)
	}
	return nil
}

// CheckInputBytes returns a LimitExceededError if input exceeds MaxInputBytes.
func (b Bounds) CheckInputBytes(size int) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("%w: input size cannot be negative", ErrInvalidMeasurement)
	}
	return b.check("input_bytes", size, b.MaxInputBytes)
}

// CheckInputText validates UTF-8 and applies MaxInputBytes to the encoded
// request size. Invalid encoding is not reported as a resource limit.
func (b Bounds) CheckInputText(text string) error {
	if !utf8.ValidString(text) {
		return fmt.Errorf("%w: text is not valid UTF-8", ErrInvalidInput)
	}
	return b.CheckInputBytes(len(text))
}

// CheckGraphDepth returns a LimitExceededError if depth exceeds MaxGraphDepth.
func (b Bounds) CheckGraphDepth(depth int) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if depth < 0 {
		return fmt.Errorf("%w: graph depth cannot be negative", ErrInvalidMeasurement)
	}
	return b.check("graph_depth", depth, b.MaxGraphDepth)
}

// CheckAnalyses returns a LimitExceededError if count exceeds MaxAnalyses.
func (b Bounds) CheckAnalyses(count int) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if count < 0 {
		return fmt.Errorf("%w: analysis count cannot be negative", ErrInvalidMeasurement)
	}
	return b.check("analyses", count, b.MaxAnalyses)
}

func (b Bounds) check(name string, value, maximum int) error {
	if value > maximum {
		return &LimitExceededError{Name: name, Value: value, Maximum: maximum}
	}
	return nil
}

// LimitExceededError represents a controlled resource limit, distinct from an
// unsupported operation or malformed input.
type LimitExceededError struct {
	Name    string
	Value   int
	Maximum int
}

func (e *LimitExceededError) Error() string {
	if e == nil {
		return "resource limit exceeded"
	}
	if !validLimitName(e.Name) || e.Value < 0 || e.Maximum <= 0 || e.Value <= e.Maximum {
		return "invalid resource limit error"
	}
	return fmt.Sprintf("%s limit exceeded: %d > %d", e.Name, e.Value, e.Maximum)
}

// Result converts the limit error to the typed resource_limit outcome.
func (e *LimitExceededError) Result() Result {
	if e == nil || !validLimitName(e.Name) || e.Value < 0 || e.Maximum <= 0 || e.Value <= e.Maximum {
		return Result{
			Status: OutcomeError,
			Diagnostics: []Diagnostic{{
				Code:     "INVALID_LIMIT_ERROR",
				Message:  "invalid resource limit error",
				Path:     "budget",
				Severity: SeverityError,
			}},
		}
	}
	diagnostic := Diagnostic{
		Code:     "RESOURCE_LIMIT",
		Message:  e.Error(),
		Path:     "budget." + e.Name,
		Severity: SeverityError,
	}
	return Result{
		Status:      OutcomeResourceLimit,
		Diagnostics: []Diagnostic{diagnostic},
	}
}

func validLimitName(name string) bool {
	switch name {
	case "input_bytes", "graph_depth", "analyses":
		return true
	default:
		return false
	}
}
