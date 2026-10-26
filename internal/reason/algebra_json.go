package reason

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// DecodeAlgebraEvaluationProof strictly decodes one bounded evaluator result
// for later independent witness verification. The optional Plan description is
// ignored: verification uses the immutable compiled plan supplied separately.
func DecodeAlgebraEvaluationProof(data []byte) (AlgebraEvaluationResult, error) {
	if err := algebra.PreflightJSONLimit(data, MaxAlgebraOutputBytes); err != nil {
		return AlgebraEvaluationResult{}, err
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil || root == nil {
		return AlgebraEvaluationResult{}, planInvalid("INVALID_EVALUATION_PROOF_JSON", "proof", "expected one evaluation-result object")
	}
	if err := validateEvaluationProofWire(root); err != nil {
		return AlgebraEvaluationResult{}, err
	}
	// The supplied plan description is never an authority and is not traversed.
	// Its bytes remain covered by the global 16 MiB preflight and this smaller
	// field bound before it is discarded from the typed claim.
	if raw, exists := root["plan"]; exists && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		trimmed := bytes.TrimSpace(raw)
		if len(trimmed) > algebra.MaxDocumentBytes || len(trimmed) == 0 || trimmed[0] != '{' {
			return AlgebraEvaluationResult{}, planResource("proof.plan", "untrusted plan description exceeds its bound or is not an object")
		}
		root["plan"] = json.RawMessage("null")
	}
	bounded, err := json.Marshal(root)
	if err != nil || len(bounded) > MaxAlgebraOutputBytes {
		return AlgebraEvaluationResult{}, planResource("proof", "evaluation result exceeds the bounded decoding envelope")
	}
	decoder := json.NewDecoder(bytes.NewReader(bounded))
	decoder.DisallowUnknownFields()
	var result AlgebraEvaluationResult
	if err := decoder.Decode(&result); err != nil {
		return AlgebraEvaluationResult{}, planInvalid("INVALID_EVALUATION_PROOF_JSON", "proof", err.Error())
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return AlgebraEvaluationResult{}, planInvalid("TRAILING_EVALUATION_PROOF", "proof", "expected exactly one JSON document")
	}
	return result, nil
}

func validateEvaluationProofWire(root map[string]json.RawMessage) error {
	required := []string{"schema_version", "status", "profile", "knowledge", "context_id", "facts_snapshot_id", "formal_snapshot_id", "evidence_fingerprint", "plan_fingerprint", "columns", "answers", "proof", "input_evidence", "complete", "diagnostics", "usage", "limits", "assumptions"}
	for _, field := range required {
		raw, ok := root[field]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return planInvalid("MISSING_EVALUATION_PROOF_FIELD", "proof."+field, "required evaluation-result field is missing or null")
		}
	}
	columns, err := wireArray(root["columns"], "proof.columns", 512)
	if err != nil {
		return err
	}
	answers, err := wireArray(root["answers"], "proof.answers", MaxAlgebraResults)
	if err != nil {
		return err
	}
	proof, err := wireArray(root["proof"], "proof.proof", MaxAlgebraProofNodes)
	if err != nil {
		return err
	}
	diagnostics, err := wireArray(root["diagnostics"], "proof.diagnostics", 16)
	if err != nil {
		return err
	}
	assumptions, err := wireArray(root["assumptions"], "proof.assumptions", 32)
	if err != nil {
		return err
	}
	_ = diagnostics
	_ = assumptions
	for i, raw := range columns {
		object, err := wireObject(raw, fmt.Sprintf("proof.columns[%d]", i))
		if err != nil {
			return err
		}
		if err := requireWireFields(object, fmt.Sprintf("proof.columns[%d]", i), "as", "variable", "type"); err != nil {
			return err
		}
		if _, err := algebra.DecodeTypeJSON(object["type"], fmt.Sprintf("proof.columns[%d].type", i)); err != nil {
			return err
		}
	}
	for i, raw := range answers {
		object, err := wireObject(raw, fmt.Sprintf("proof.answers[%d]", i))
		if err != nil {
			return err
		}
		if err := requireWireFields(object, fmt.Sprintf("proof.answers[%d]", i), "id", "values", "proof_ids", "has_conflict"); err != nil {
			return err
		}
		if _, err := wireArray(object["values"], fmt.Sprintf("proof.answers[%d].values", i), 512); err != nil {
			return err
		}
		if _, err := wireArray(object["proof_ids"], fmt.Sprintf("proof.answers[%d].proof_ids", i), MaxAlgebraProofsPerAnswer); err != nil {
			return err
		}
	}
	for i, raw := range proof {
		object, err := wireObject(raw, fmt.Sprintf("proof.proof[%d]", i))
		if err != nil {
			return err
		}
		path := fmt.Sprintf("proof.proof[%d]", i)
		if err := requireWireFields(object, path, "id", "rule_id", "plan_node_id", "premises", "fact_ids", "sources", "bindings", "opposing_negative_fact_ids", "depth"); err != nil {
			return err
		}
		for field, maximum := range map[string]int{"premises": MaxAlgebraProofNodes, "fact_ids": factsWireMax, "sources": factsWireMax, "bindings": 512, "opposing_negative_fact_ids": factsWireMax} {
			if _, err := wireArray(object[field], path+"."+field, maximum); err != nil {
				return err
			}
		}
		if reference, ok := object["reference"]; ok && !bytes.Equal(bytes.TrimSpace(reference), []byte("null")) {
			declaration, err := wireObject(reference, path+".reference")
			if err != nil {
				return err
			}
			if err := requireWireFields(declaration, path+".reference", "id", "type"); err != nil {
				return err
			}
			if _, err := algebra.DecodeTypeJSON(declaration["type"], path+".reference.type"); err != nil {
				return err
			}
		}
	}
	evidence, err := wireObject(root["input_evidence"], "proof.input_evidence")
	if err != nil {
		return err
	}
	if err := requireWireFields(evidence, "proof.input_evidence", "stored_facts", "active_facts", "inactive_facts", "positive_records", "negative_records", "conflicts"); err != nil {
		return err
	}
	if _, err := wireArray(evidence["conflicts"], "proof.input_evidence.conflicts", factsWireMax); err != nil {
		return err
	}
	if ground, ok := root["ground_atom_evidence"]; ok && !bytes.Equal(bytes.TrimSpace(ground), []byte("null")) {
		object, err := wireObject(ground, "proof.ground_atom_evidence")
		if err != nil {
			return err
		}
		if err := requireWireFields(object, "proof.ground_atom_evidence", "state", "positive", "negative"); err != nil {
			return err
		}
		if _, err := wireArray(object["positive"], "proof.ground_atom_evidence.positive", factsWireMax); err != nil {
			return err
		}
		if _, err := wireArray(object["negative"], "proof.ground_atom_evidence.negative", factsWireMax); err != nil {
			return err
		}
	}
	return nil
}

const factsWireMax = 4096

func wireObject(raw json.RawMessage, path string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, planInvalid("INVALID_EVALUATION_PROOF_SHAPE", path, "expected a non-null object")
	}
	return object, nil
}

func wireArray(raw json.RawMessage, path string, maximum int) ([]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, planInvalid("INVALID_EVALUATION_PROOF_SHAPE", path, "expected a non-null array")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(trimmed, &values); err != nil {
		return nil, planInvalid("INVALID_EVALUATION_PROOF_SHAPE", path, "array shape is invalid")
	}
	if len(values) > maximum {
		return nil, planResource(path, "array exceeds the bounded evaluation-proof limit")
	}
	return values, nil
}

func requireWireFields(object map[string]json.RawMessage, path string, fields ...string) error {
	for _, field := range fields {
		raw, ok := object[field]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return planInvalid("MISSING_EVALUATION_PROOF_FIELD", path+"."+field, "required proof field is missing or null")
		}
	}
	return nil
}

// DecodeAlgebraSelections strictly decodes a bounded list of selected output
// variables. Plan compilation checks that each selected binder or parameter
// exists and has the claimed scope.
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
