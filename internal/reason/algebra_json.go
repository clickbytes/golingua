package reason

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// DecodeAlgebraSelections strictly decodes a bounded list of selected output
// variables. It performs wire validation only; plan compilation checks that
// the selected binder or parameter exists and has the claimed scope.
func DecodeAlgebraSelections(data []byte) ([]AlgebraSelection, error) {
	if err := algebra.PreflightJSON(data); err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" || !strings.HasPrefix(trimmed, "[") {
		return nil, selectionError("INVALID_SELECTIONS", "select", "selection input must be a JSON array")
	}
	var records []json.RawMessage
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, selectionError("INVALID_SELECTIONS", "select", err.Error())
	}
	if records == nil {
		return nil, selectionError("INVALID_SELECTIONS", "select", "selection array must not be null")
	}
	if len(records) > MaxAlgebraVariables {
		return nil, &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: "select", Message: "selection count exceeds the algebra variable limit"}
	}
	selections := make([]AlgebraSelection, 0, len(records))
	for i, record := range records {
		path := fmt.Sprintf("select[%d]", i)
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(record, &fields); err != nil || fields == nil {
			return nil, selectionError("INVALID_SELECTION", path, "selection must be an object")
		}
		kind, err := selectionString(fields, "kind", path+".kind")
		if err != nil {
			return nil, err
		}
		as, err := selectionString(fields, "as", path+".as")
		if err != nil {
			return nil, err
		}
		selection := AlgebraSelection{Kind: kind, As: as}
		switch kind {
		case AlgebraSelectionLambda:
			if err := selectionKeys(fields, []string{"kind", "index", "as"}, path); err != nil {
				return nil, err
			}
			raw, ok := fields["index"]
			if !ok || string(raw) == "null" {
				return nil, selectionError("INVALID_SELECTION", path+".index", "lambda selection requires a non-null integer index")
			}
			value := strings.TrimSpace(string(raw))
			if value == "" {
				return nil, selectionError("INVALID_SELECTION", path+".index", "lambda index must be an integer")
			}
			for _, digit := range value {
				if digit < '0' || digit > '9' {
					return nil, selectionError("INVALID_SELECTION", path+".index", "lambda index must be a nonnegative decimal integer")
				}
			}
			index, parseErr := strconv.Atoi(value)
			if parseErr != nil {
				return nil, selectionError("INVALID_SELECTION", path+".index", "lambda index is outside the supported integer range")
			}
			selection.Index = &index
		case AlgebraSelectionParameter:
			if err := selectionKeys(fields, []string{"kind", "id", "as"}, path); err != nil {
				return nil, err
			}
			selection.ID, err = selectionString(fields, "id", path+".id")
			if err != nil {
				return nil, err
			}
		default:
			return nil, selectionError("INVALID_SELECTION_KIND", path+".kind", "selection kind must be lambda or parameter")
		}
		selections = append(selections, selection)
	}
	return selections, nil
}

func selectionString(fields map[string]json.RawMessage, key, path string) (string, error) {
	raw, ok := fields[key]
	if !ok || string(raw) == "null" {
		return "", selectionError("INVALID_SELECTION", path, "required string field is missing or null")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", selectionError("INVALID_SELECTION", path, "field must be a JSON string")
	}
	if value == "" || len(value) > 256 || !utf8.ValidString(value) || strings.TrimSpace(value) == "" {
		return "", selectionError("INVALID_SELECTION", path, "string must be nonempty bounded UTF-8 text")
	}
	return value, nil
}

func selectionKeys(fields map[string]json.RawMessage, allowed []string, path string) error {
	known := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		known[key] = struct{}{}
	}
	for key := range fields {
		if _, ok := known[key]; !ok {
			return selectionError("UNKNOWN_SELECTION_FIELD", path+"."+key, "field is not valid for this selection kind")
		}
	}
	for _, key := range allowed {
		if _, ok := fields[key]; !ok {
			return selectionError("INVALID_SELECTION", path+"."+key, "required field is missing")
		}
	}
	return nil
}

func selectionError(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}
