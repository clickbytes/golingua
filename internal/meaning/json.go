package meaning

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// UnmarshalJSON enforces the tagged union at the wire boundary, including
// fields whose explicit value would otherwise decode to a zero value.
func (node *Node) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("node must be a JSON object")
	}
	kindData, exists := fields["kind"]
	if !exists {
		return fmt.Errorf("node.kind is required")
	}
	var kind NodeKind
	if err := json.Unmarshal(kindData, &kind); err != nil || !kind.Valid() {
		return fmt.Errorf("node.kind is invalid")
	}
	for _, field := range []string{"temporal_frame", "multiplicity"} {
		value, present := fields[field]
		if !present {
			continue
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("node.%s cannot be null", field)
		}
		var decoded string
		if err := json.Unmarshal(value, &decoded); err != nil || decoded == "" {
			return fmt.Errorf("node.%s must be a non-empty string", field)
		}
	}
	allowed := nodeFieldsForKind(kind)
	type wireNode Node
	var decoded wireNode
	if err := decodeExactObject(data, allowed, &decoded); err != nil {
		return err
	}
	*node = Node(decoded)
	return nil
}

// UnmarshalJSON rejects unknown or case-mismatched participant fields.
func (participant *Participant) UnmarshalJSON(data []byte) error {
	type wireParticipant Participant
	var decoded wireParticipant
	if err := decodeExactObject(data, map[string]bool{"role": true, "target": true}, &decoded); err != nil {
		return err
	}
	*participant = Participant(decoded)
	return nil
}

// UnmarshalJSON enforces exact quantifier fields and rejects explicit null counts.
func (quantifier *Quantifier) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("quantifier must be a JSON object")
	}
	if count, exists := fields["count"]; exists && bytes.Equal(bytes.TrimSpace(count), []byte("null")) {
		return fmt.Errorf("quantifier.count cannot be null")
	}
	type wireQuantifier Quantifier
	var decoded wireQuantifier
	if err := decodeExactObject(data, map[string]bool{"kind": true, "count": true}, &decoded); err != nil {
		return err
	}
	*quantifier = Quantifier(decoded)
	return nil
}

// UnmarshalJSON requires the exact top-level schema field spellings.
func (meaning *Meaning) UnmarshalJSON(data []byte) error {
	type wireMeaning Meaning
	var decoded wireMeaning
	if err := decodeExactObject(data, map[string]bool{"schema_version": true, "roots": true, "nodes": true}, &decoded); err != nil {
		return err
	}
	*meaning = Meaning(decoded)
	return nil
}

func decodeExactObject(data []byte, allowed map[string]bool, target any) error {
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

// DecodeMeaning strictly decodes one versioned Meaning document. Unknown
// fields and trailing JSON values are rejected.
func DecodeMeaning(data []byte, budget Budget) (Meaning, error) {
	if len(data) > MaxMeaningBytes {
		return Meaning{}, resourceLimit("input_bytes", len(data), MaxMeaningBytes)
	}
	if err := budget.Validate(); err != nil {
		return Meaning{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var meaning Meaning
	if err := decoder.Decode(&meaning); err != nil {
		return Meaning{}, invalid("INVALID_MEANING_JSON", "document", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Meaning{}, invalid("TRAILING_JSON", "document", "expected exactly one JSON value")
	}
	if err := ValidateStructure(meaning, budget); err != nil {
		return Meaning{}, err
	}
	return cloneMeaning(meaning), nil
}

// EncodeMeaning emits the versioned JSON representation in caller order after
// validating structure and resource bounds.
func EncodeMeaning(meaning Meaning, budget Budget) ([]byte, error) {
	if err := ValidateStructure(meaning, budget); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(meaning)
	if err != nil {
		return nil, invalid("INVALID_MEANING_JSON", "document", err.Error())
	}
	if len(encoded) > MaxMeaningBytes {
		return nil, resourceLimit("input_bytes", len(encoded), MaxMeaningBytes)
	}
	return encoded, nil
}

// CanonicalJSON sorts graph nodes by local ID and role assignments by role.
// Role names remain attached to targets, so swapping semantic roles changes
// the canonical form. Ordered discourse content is preserved.
func CanonicalJSON(meaning Meaning, budget Budget) ([]byte, error) {
	if err := ValidateStructure(meaning, budget); err != nil {
		return nil, err
	}
	canonical := cloneMeaning(meaning)
	sort.Slice(canonical.Roots, func(i, j int) bool { return canonical.Roots[i] < canonical.Roots[j] })
	sort.Slice(canonical.Nodes, func(i, j int) bool { return canonical.Nodes[i].ID < canonical.Nodes[j].ID })
	for index := range canonical.Nodes {
		if canonical.Nodes[index].Kind == KindEvent || canonical.Nodes[index].Kind == KindState {
			sort.Slice(canonical.Nodes[index].Participants, func(i, j int) bool {
				left, right := canonical.Nodes[index].Participants[i], canonical.Nodes[index].Participants[j]
				if left.Role != right.Role {
					return left.Role < right.Role
				}
				return left.Target < right.Target
			})
		}
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return nil, invalid("INVALID_MEANING_JSON", "document", err.Error())
	}
	if len(encoded) > MaxMeaningBytes {
		return nil, resourceLimit("input_bytes", len(encoded), MaxMeaningBytes)
	}
	return encoded, nil
}

// CanonicalHash returns a SHA-256 digest of CanonicalJSON.
func CanonicalHash(meaning Meaning, budget Budget) (string, error) {
	encoded, err := CanonicalJSON(meaning, budget)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(encoded)
	return hex.EncodeToString(hash[:]), nil
}

// CanonicalEqual compares meanings after stable node and role ordering.
func CanonicalEqual(left, right Meaning, budget Budget) (bool, error) {
	leftJSON, err := CanonicalJSON(left, budget)
	if err != nil {
		return false, err
	}
	rightJSON, err := CanonicalJSON(right, budget)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftJSON, rightJSON), nil
}

func cloneMeaning(source Meaning) Meaning {
	result := Meaning{SchemaVersion: source.SchemaVersion, Roots: append([]NodeID(nil), source.Roots...), Nodes: make([]Node, len(source.Nodes))}
	for index, node := range source.Nodes {
		result.Nodes[index] = node
		result.Nodes[index].Participants = append([]Participant(nil), node.Participants...)
		result.Nodes[index].Content = append([]NodeID(nil), node.Content...)
		if node.Quantifier != nil {
			quantifier := cloneQuantifier(*node.Quantifier)
			result.Nodes[index].Quantifier = &quantifier
		}
	}
	return result
}
