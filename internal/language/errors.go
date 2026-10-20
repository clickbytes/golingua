package language

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/status"
)

type Error struct{ result status.Result }

func (failure *Error) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "language package operation failed"
	}
	return failure.result.Diagnostics[0].Message
}
func (failure *Error) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}
func languageError(kind status.OutcomeKind, code, path, message string) error {
	return &Error{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}}
}
func invalid(code, path, message string) error {
	return languageError(status.OutcomeInvalid, code, path, message)
}
func unknown(code, path, message string) error {
	return languageError(status.OutcomeUnknown, code, path, message)
}
func unsupported(code, path, message string) error {
	return languageError(status.OutcomeUnsupported, code, path, message)
}
func incomplete(code, path, message string) error {
	return languageError(status.OutcomeIncomplete, code, path, message)
}
func infra(code, path, message string) error {
	return languageError(status.OutcomeError, code, path, message)
}
func resource(name string, value, maximum int) error {
	return languageError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+name, fmt.Sprintf("language %s limit exceeded: %d > %d", name, value, maximum))
}
