package importer

import (
	"errors"
	"fmt"

	"github.com/clickbytes/golingua/internal/status"
)

var ErrInvalidDataset = errors.New("invalid dataset")

type DatasetError struct{ result status.Result }

func (err *DatasetError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "dataset import failed"
	}
	return err.result.Diagnostics[0].Message
}

func (err *DatasetError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

func infrastructure(code string, err error) error {
	return datasetError(status.OutcomeError, code, "dataset", fmt.Sprintf("local data access failed: %v", err))
}

func datasetError(kind status.OutcomeKind, code, path, message string) error {
	return &DatasetError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: boundedDiagnosticText(code, status.MaxDiagnosticCodeBytes), Path: boundedDiagnosticText(path, status.MaxDiagnosticPathBytes), Message: boundedDiagnosticText(message, status.MaxDiagnosticMessageBytes), Severity: status.SeverityError}}}}
}
