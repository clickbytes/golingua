package status

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MaxDiagnosticCodeBytes    = 64
	MaxDiagnosticMessageBytes = 2048
	MaxDiagnosticPathBytes    = 512
)

// Severity communicates the diagnostic's weight without changing the outcome
// category of the enclosing Result.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

var ErrInvalidDiagnostic = errors.New("invalid diagnostic")

// Diagnostic is a bounded machine-readable explanation. Code is stable for
// callers; Message is intended for people; Path identifies the relevant input
// or runtime field when one exists.
type Diagnostic struct {
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Path     string   `json:"path,omitempty"`
	Severity Severity `json:"severity"`
}

// Validate checks the stable code format and bounds all user-facing strings.
func (d Diagnostic) Validate() error {
	if len(d.Code) == 0 || len(d.Code) > MaxDiagnosticCodeBytes || !validCode(d.Code) {
		return fmt.Errorf("%w: code must use uppercase letters, digits, and underscores", ErrInvalidDiagnostic)
	}
	if len(d.Message) == 0 || len(d.Message) > MaxDiagnosticMessageBytes || !utf8.ValidString(d.Message) {
		return fmt.Errorf("%w: message is empty, invalid UTF-8, or exceeds %d bytes", ErrInvalidDiagnostic, MaxDiagnosticMessageBytes)
	}
	if len(d.Path) > MaxDiagnosticPathBytes || !utf8.ValidString(d.Path) {
		return fmt.Errorf("%w: path is invalid UTF-8 or exceeds %d bytes", ErrInvalidDiagnostic, MaxDiagnosticPathBytes)
	}
	switch d.Severity {
	case SeverityInfo, SeverityWarning, SeverityError:
		return nil
	default:
		return fmt.Errorf("%w: unknown severity %q", ErrInvalidDiagnostic, d.Severity)
	}
}

func validCode(code string) bool {
	for _, r := range code {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return !strings.HasPrefix(code, "_") && !strings.HasSuffix(code, "_")
}
