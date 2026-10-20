package procedure

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (operation *Operation) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("operation must be a JSON object")
	}
	kindData, exists := fields["kind"]
	if !exists {
		return fmt.Errorf("operation.kind is required")
	}
	var kind OperationKind
	if err := json.Unmarshal(kindData, &kind); err != nil {
		return fmt.Errorf("operation.kind must be a string")
	}
	allowed := map[string]bool{"kind": true, "when": true}
	required := []string{"kind"}
	switch kind {
	case OperationSelectStem:
		allowed["stem_id"], required = true, append(required, "stem_id")
	case OperationAppend, OperationPrepend:
		allowed["text"], required = true, append(required, "text")
	case OperationStripSuffix:
		allowed["suffix"], required = true, append(required, "suffix")
	case OperationReplaceSuffix:
		allowed["suffix"], allowed["replacement"] = true, true
		required = append(required, "suffix", "replacement")
	case OperationAlternation:
		allowed["from"], allowed["to"] = true, true
		required = append(required, "from", "to")
	}
	for _, field := range required {
		if _, exists := fields[field]; !exists {
			return fmt.Errorf("operation field %q is required", field)
		}
	}
	for _, field := range []string{"when", "replacement", "to"} {
		if value, exists := fields[field]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("operation field %q cannot be null", field)
		}
	}
	type wireOperation Operation
	var decoded wireOperation
	if err := decodeExactOperation(data, allowed, &decoded); err != nil {
		return err
	}
	*operation = Operation(decoded)
	return nil
}

func (condition *FeatureCondition) UnmarshalJSON(data []byte) error {
	type wireCondition FeatureCondition
	var decoded wireCondition
	if err := decodeExactOperation(data, map[string]bool{"feature": true, "value": true}, &decoded); err != nil {
		return err
	}
	*condition = FeatureCondition(decoded)
	return nil
}

func decodeExactOperation(data []byte, allowed map[string]bool, target any) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a JSON object")
	}
	for field := range fields {
		if !allowed[field] {
			return fmt.Errorf("unknown or mis-cased field %q", field)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
