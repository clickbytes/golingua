package morph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

const MaxParadigmJSONBytes = 1 << 20

func (paradigm *Paradigm) UnmarshalJSON(data []byte) error {
	type wireParadigm Paradigm
	var decoded wireParadigm
	if err := decodeMorphObject(data,
		map[string]bool{"schema_version": true, "id": true, "language": true, "profile_version": true, "pos": true, "inherits": true, "principal_part_requirements": true, "cells": true, "rules": true},
		[]string{"schema_version", "id", "language", "profile_version", "pos", "principal_part_requirements", "cells", "rules"},
		[]string{"inherits", "principal_part_requirements", "cells", "rules"}, &decoded); err != nil {
		return err
	}
	*paradigm = Paradigm(decoded)
	return nil
}

func (cell *ParadigmCell) UnmarshalJSON(data []byte) error {
	type wireCell ParadigmCell
	var decoded wireCell
	if err := decodeMorphObject(data, map[string]bool{"id": true, "features": true}, []string{"id", "features"}, []string{"features"}, &decoded); err != nil {
		return err
	}
	*cell = ParadigmCell(decoded)
	return nil
}

func (rule *CellRule) UnmarshalJSON(data []byte) error {
	type wireRule CellRule
	var decoded wireRule
	if err := decodeMorphObject(data, map[string]bool{"cell_id": true, "steps": true}, []string{"cell_id", "steps"}, []string{"steps"}, &decoded); err != nil {
		return err
	}
	*rule = CellRule(decoded)
	return nil
}

func (requirement *PrincipalPartRequirement) UnmarshalJSON(data []byte) error {
	type wireRequirement PrincipalPartRequirement
	var decoded wireRequirement
	if err := decodeMorphObject(data, map[string]bool{"stem_id": true, "part_id": true, "required": true}, []string{"stem_id", "part_id", "required"}, nil, &decoded); err != nil {
		return err
	}
	*requirement = PrincipalPartRequirement(decoded)
	return nil
}

func (part *PrincipalPart) UnmarshalJSON(data []byte) error {
	type wirePart PrincipalPart
	var decoded wirePart
	if err := decodeMorphObject(data, map[string]bool{"id": true, "form": true}, []string{"id", "form"}, nil, &decoded); err != nil {
		return err
	}
	*part = PrincipalPart(decoded)
	return nil
}

func DecodeParadigm(data []byte) (Paradigm, error) {
	if len(data) > MaxParadigmJSONBytes {
		return Paradigm{}, fmt.Errorf("paradigm JSON exceeds %d bytes", MaxParadigmJSONBytes)
	}
	var paradigm Paradigm
	if err := json.Unmarshal(data, &paradigm); err != nil {
		return Paradigm{}, err
	}
	return paradigm, nil
}

func DecodeParadigms(data []byte) ([]Paradigm, error) {
	if len(data) > MaxParadigmJSONBytes {
		return nil, fmt.Errorf("paradigm JSON exceeds %d bytes", MaxParadigmJSONBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var paradigms []Paradigm
	if err := decoder.Decode(&paradigms); err != nil {
		return nil, err
	}
	if paradigms == nil {
		return nil, fmt.Errorf("paradigms must be a JSON array, not null")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("expected exactly one JSON value")
	}
	if len(paradigms) > MaxParadigms {
		return nil, fmt.Errorf("paradigm count exceeds %d", MaxParadigms)
	}
	return paradigms, nil
}

func decodeMorphObject(data []byte, allowed map[string]bool, required, nonNull []string, target any) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("expected a morphology JSON object")
	}
	for field := range fields {
		if !allowed[field] {
			return fmt.Errorf("unknown or mis-cased field %q", field)
		}
	}
	for _, field := range required {
		if _, exists := fields[field]; !exists {
			return fmt.Errorf("required field %q is missing", field)
		}
	}
	for _, field := range nonNull {
		if value, exists := fields[field]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("field %q cannot be null", field)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}
