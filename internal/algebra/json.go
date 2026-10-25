package algebra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"unicode/utf8"
)

// MaxDocumentBytes is the hard wire-format size ceiling for one algebra
// document. It is checked before parsing.
const MaxDocumentBytes = 1 << 20

// MaxJSONNesting is a pre-decode cap on nested JSON arrays and objects.
const MaxJSONNesting = 256

// Decode strictly decodes and structurally checks one version-1 expression
// document. It does not evaluate the expression or infer truth.
func Decode(data []byte) (Document, error) {
	document, err := DecodeDocumentJSON(data)
	if err != nil {
		return Document{}, err
	}
	if _, err := Check(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

// DecodeDocumentJSON validates the strict document wire shape and returns the
// typed document without performing declaration or expression semantics.
func DecodeDocumentJSON(data []byte) (Document, error) {
	if err := PreflightJSON(data); err != nil {
		return Document{}, err
	}
	if err := validateDocumentJSON(data); err != nil {
		return Document{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document Document
	if err := decoder.Decode(&document); err != nil {
		return Document{}, invalid("INVALID_JSON_DOCUMENT", "document", err.Error())
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Document{}, invalid("TRAILING_DOCUMENT", "document", "expected exactly one JSON document")
	}
	return document, nil
}

// PreflightJSON applies the wire-level byte, UTF-8, surrogate, nesting,
// duplicate-key, and single-document guards without interpreting a schema.
func PreflightJSON(data []byte) error {
	if len(data) > MaxDocumentBytes {
		return resourceLimit("document", fmt.Sprintf("JSON document is %d bytes; maximum is %d", len(data), MaxDocumentBytes))
	}
	if !utf8.Valid(data) {
		return invalid("INVALID_UTF8", "document", "JSON document is not valid UTF-8")
	}
	if err := validateUnicodeEscapes(data); err != nil {
		return err
	}
	return scanJSON(data)
}

// DecodeTypeJSON strictly decodes one type value after applying the shared
// wire guards. It does not resolve nominal sorts; use CheckDeclarations for
// declaration-context validation.
func DecodeTypeJSON(data []byte, path string) (Type, error) {
	if err := PreflightJSON(data); err != nil {
		return Type{}, err
	}
	if err := validateTypeJSON(json.RawMessage(data), path); err != nil {
		return Type{}, err
	}
	var value Type
	if err := strictTypedUnmarshal(data, &value, path); err != nil {
		return Type{}, err
	}
	return value, nil
}

// DecodeTermJSON strictly decodes one term value after applying shared wire
// guards. Declaration, scope, and type checking remain the caller's job.
func DecodeTermJSON(data []byte, path string) (*Term, error) {
	if err := PreflightJSON(data); err != nil {
		return nil, err
	}
	if err := validateTermJSON(json.RawMessage(data), path); err != nil {
		return nil, err
	}
	var value Term
	if err := strictTypedUnmarshal(data, &value, path); err != nil {
		return nil, err
	}
	return &value, nil
}

func strictTypedUnmarshal(data []byte, target any, path string) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return invalid("INVALID_JSON_VALUE", path, err.Error())
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return invalid("TRAILING_DOCUMENT", path, "expected exactly one JSON value")
	}
	return nil
}

// Encode validates and encodes a document without changing it. Incomplete
// expressions containing typed holes remain encodable; no truth is evaluated.
func Encode(document Document) ([]byte, error) {
	if _, err := Check(document); err != nil {
		return nil, err
	}
	wire, err := documentWire(document)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("ENCODE_ERROR", "document", err.Error())
	}
	if len(data) > MaxDocumentBytes {
		return nil, resourceLimit("document", fmt.Sprintf("encoded JSON document is %d bytes; maximum is %d", len(data), MaxDocumentBytes))
	}
	if err := validateDocumentJSON(data); err != nil {
		return nil, err
	}
	return data, nil
}

func documentWire(document Document) (map[string]any, error) {
	wire := map[string]any{"schema_version": document.SchemaVersion}
	if len(document.Sorts) != 0 {
		wire["sorts"] = document.Sorts
	}
	if len(document.Signatures) != 0 {
		signatures := make([]map[string]any, 0, len(document.Signatures))
		for _, signature := range document.Signatures {
			ports := make([]Port, len(signature.Ports))
			copy(ports, signature.Ports)
			entry := map[string]any{"id": signature.ID, "ports": ports}
			if signature.Label != "" {
				entry["label"] = signature.Label
			}
			signatures = append(signatures, entry)
		}
		wire["signatures"] = signatures
	}
	if len(document.Parameters) != 0 {
		wire["parameters"] = document.Parameters
	}
	if len(document.References) != 0 {
		wire["references"] = document.References
	}
	if document.Closed {
		wire["closed"] = true
	}
	if document.Limits != (Limits{}) {
		wire["limits"] = document.Limits
	}
	expression, err := termWire(document.Expression)
	if err != nil {
		return nil, err
	}
	wire["expression"] = expression
	return wire, nil
}

func termWire(term *Term) (map[string]any, error) {
	if term == nil {
		return nil, invalid("MISSING_TERM", "expression", "term is required")
	}
	wire := map[string]any{"kind": term.Kind}
	if term.Label != "" {
		wire["label"] = term.Label
	}
	switch term.Kind {
	case TermVar:
		wire["id"] = term.ID
	case TermParam, TermRef, TermHole:
		wire["id"], wire["type"] = term.ID, term.Type
	case TermAtom:
		wire["symbol"] = term.Symbol
		args := make([]map[string]any, 0, len(term.Args))
		for _, arg := range term.Args {
			value, err := termWire(arg.Value)
			if err != nil {
				return nil, err
			}
			args = append(args, map[string]any{"role": arg.Role, "value": value})
		}
		wire["args"] = args
	case TermAnd:
		operands := make([]map[string]any, 0, len(term.Operands))
		for _, operand := range term.Operands {
			value, err := termWire(operand)
			if err != nil {
				return nil, err
			}
			operands = append(operands, value)
		}
		wire["operands"] = operands
	case TermExists, TermLambda:
		body, err := termWire(term.Body)
		if err != nil {
			return nil, err
		}
		wire["id"], wire["type"], wire["body"] = term.ID, term.Type, body
	case TermApply:
		function, err := termWire(term.Function)
		if err != nil {
			return nil, err
		}
		argument, err := termWire(term.Argument)
		if err != nil {
			return nil, err
		}
		wire["function"], wire["argument"] = function, argument
	default:
		return nil, invalid("UNKNOWN_TERM_KIND", "expression.kind", "term kind is unsupported")
	}
	return wire, nil
}

func scanJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder, "document", 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return invalid("INVALID_JSON", "document", err.Error())
		}
		return invalid("TRAILING_DOCUMENT", "document", "expected exactly one JSON value")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder, path string, depth int) error {
	token, err := decoder.Token()
	if err != nil {
		return invalid("INVALID_JSON", path, err.Error())
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	if depth >= MaxJSONNesting {
		return resourceLimit(path, fmt.Sprintf("JSON nesting exceeds limit %d", MaxJSONNesting))
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return invalid("INVALID_JSON", path, err.Error())
			}
			key, ok := keyToken.(string)
			if !ok {
				return invalid("INVALID_JSON", path, "object key must be a string")
			}
			if _, exists := seen[key]; exists {
				return invalid("DUPLICATE_JSON_KEY", appendJSONPath(path, key), "object key is duplicated")
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder, appendJSONPath(path, key), depth+1); err != nil {
				return err
			}
		}
		closeToken, err := decoder.Token()
		if err != nil || closeToken != json.Delim('}') {
			return invalid("INVALID_JSON", path, "object is not properly closed")
		}
	case '[':
		index := 0
		for decoder.More() {
			if err := scanJSONValue(decoder, fmt.Sprintf("%s[%d]", path, index), depth+1); err != nil {
				return err
			}
			index++
		}
		closeToken, err := decoder.Token()
		if err != nil || closeToken != json.Delim(']') {
			return invalid("INVALID_JSON", path, "array is not properly closed")
		}
	default:
		return invalid("INVALID_JSON", path, "unexpected JSON delimiter")
	}
	return nil
}

func validateUnicodeEscapes(data []byte) error {
	inString := false
	for i := 0; i < len(data); i++ {
		if !inString {
			if data[i] == '"' {
				inString = true
			}
			continue
		}
		if data[i] == '"' {
			inString = false
			continue
		}
		if data[i] != '\\' {
			continue
		}
		if i+1 >= len(data) {
			return invalid("INVALID_JSON", "document", "trailing escape in JSON string")
		}
		if data[i+1] != 'u' {
			i++
			continue
		}
		code, ok := parseHexRune(data, i+2)
		if !ok {
			return invalid("INVALID_JSON", "document", "malformed Unicode escape")
		}
		switch {
		case code >= 0xD800 && code <= 0xDBFF:
			if i+12 > len(data) || data[i+6] != '\\' || data[i+7] != 'u' {
				return invalid("INVALID_UTF8", "document", "high surrogate is not followed by a low surrogate")
			}
			low, valid := parseHexRune(data, i+8)
			if !valid || low < 0xDC00 || low > 0xDFFF {
				return invalid("INVALID_UTF8", "document", "high surrogate is not followed by a low surrogate")
			}
			i += 11
		case code >= 0xDC00 && code <= 0xDFFF:
			return invalid("INVALID_UTF8", "document", "unpaired low surrogate")
		default:
			i += 5
		}
	}
	return nil
}

func parseHexRune(data []byte, start int) (uint16, bool) {
	if start+4 > len(data) {
		return 0, false
	}
	var value uint16
	for _, char := range data[start : start+4] {
		value <<= 4
		switch {
		case char >= '0' && char <= '9':
			value |= uint16(char - '0')
		case char >= 'a' && char <= 'f':
			value |= uint16(char-'a') + 10
		case char >= 'A' && char <= 'F':
			value |= uint16(char-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func validateDocumentJSON(data []byte) error {
	root, err := strictObject(data, "document", []string{"schema_version", "sorts", "signatures", "parameters", "references", "closed", "expression", "limits"}, []string{"schema_version", "expression"})
	if err != nil {
		return err
	}
	var version int
	if err := json.Unmarshal(root["schema_version"], &version); err != nil {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "schema_version must be an integer")
	}
	if version != SchemaVersion {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported algebra document version")
	}
	if raw, ok := root["sorts"]; ok {
		items, err := strictArray(raw, "sorts", 512)
		if err != nil {
			return err
		}
		for i, item := range items {
			path := fmt.Sprintf("sorts[%d]", i)
			if _, err := strictObject(item, path, []string{"id", "label"}, []string{"id"}); err != nil {
				return err
			}
		}
	}
	if raw, ok := root["signatures"]; ok {
		items, err := strictArray(raw, "signatures", 512)
		if err != nil {
			return err
		}
		for i, item := range items {
			path := fmt.Sprintf("signatures[%d]", i)
			fields, err := strictObject(item, path, []string{"id", "label", "ports"}, []string{"id", "ports"})
			if err != nil {
				return err
			}
			ports, err := strictArray(fields["ports"], path+".ports", DefaultMaxPorts)
			if err != nil {
				return err
			}
			for j, port := range ports {
				portPath := fmt.Sprintf("%s.ports[%d]", path, j)
				portFields, err := strictObject(port, portPath, []string{"role", "type"}, []string{"role", "type"})
				if err != nil {
					return err
				}
				if err := validateSortTypeJSON(portFields["type"], portPath+".type"); err != nil {
					return err
				}
			}
		}
	}
	for _, name := range []string{"parameters", "references"} {
		if raw, ok := root[name]; ok {
			items, err := strictArray(raw, name, 512)
			if err != nil {
				return err
			}
			for i, item := range items {
				path := fmt.Sprintf("%s[%d]", name, i)
				fields, err := strictObject(item, path, []string{"id", "type"}, []string{"id", "type"})
				if err != nil {
					return err
				}
				if err := validateTypeJSON(fields["type"], path+".type"); err != nil {
					return err
				}
			}
		}
	}
	if raw, ok := root["limits"]; ok {
		if _, err := strictObject(raw, "limits", []string{"max_nodes", "max_depth", "max_type_depth", "max_symbols", "max_ports"}, nil); err != nil {
			return err
		}
	}
	return validateTermJSON(root["expression"], "expression")
}

func validateTypeJSON(data json.RawMessage, path string) error {
	base, err := strictObject(data, path, []string{"kind", "sort_id", "from", "to"}, []string{"kind"})
	if err != nil {
		return err
	}
	var kind TypeKind
	if err := json.Unmarshal(base["kind"], &kind); err != nil {
		return invalid("UNKNOWN_TYPE_KIND", path+".kind", "type kind must be a string")
	}
	var allowed, required []string
	switch kind {
	case TypeEntity, TypeEvent, TypeInterval, TypeProp:
		allowed, required = []string{"kind"}, []string{"kind"}
	case TypeNominal:
		allowed, required = []string{"kind", "sort_id"}, []string{"kind", "sort_id"}
	case TypeArrow:
		allowed, required = []string{"kind", "from", "to"}, []string{"kind", "from", "to"}
	default:
		return invalid("UNKNOWN_TYPE_KIND", path+".kind", "type kind is unsupported")
	}
	fields, err := strictObject(data, path, allowed, required)
	if err != nil {
		return err
	}
	if kind == TypeArrow {
		if err := validateTypeJSON(fields["from"], path+".from"); err != nil {
			return err
		}
		return validateTypeJSON(fields["to"], path+".to")
	}
	return nil
}

func validateSortTypeJSON(data json.RawMessage, path string) error {
	if err := validateTypeJSON(data, path); err != nil {
		return err
	}
	var shape struct {
		Kind TypeKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &shape); err != nil {
		return invalid("MALFORMED_TYPE", path, "port type must be an object")
	}
	if shape.Kind == TypeArrow {
		return invalid("INVALID_RELATION_PORT_TYPE", path, "relation port must have a base sort type")
	}
	return nil
}

func validateTermJSON(data json.RawMessage, path string) error {
	base, err := strictObject(data, path, []string{"kind", "id", "label", "type", "symbol", "args", "operands", "body", "function", "argument"}, []string{"kind"})
	if err != nil {
		return err
	}
	var kind TermKind
	if err := json.Unmarshal(base["kind"], &kind); err != nil {
		return invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind must be a string")
	}
	var allowed, required []string
	switch kind {
	case TermVar:
		allowed, required = []string{"kind", "id", "label"}, []string{"kind", "id"}
	case TermParam, TermRef, TermHole:
		allowed, required = []string{"kind", "id", "label", "type"}, []string{"kind", "id", "type"}
	case TermAtom:
		allowed, required = []string{"kind", "label", "symbol", "args"}, []string{"kind", "symbol", "args"}
	case TermAnd:
		allowed, required = []string{"kind", "label", "operands"}, []string{"kind", "operands"}
	case TermExists, TermLambda:
		allowed, required = []string{"kind", "id", "label", "type", "body"}, []string{"kind", "id", "type", "body"}
	case TermApply:
		allowed, required = []string{"kind", "label", "function", "argument"}, []string{"kind", "function", "argument"}
	default:
		return invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind is unsupported")
	}
	fields, err := strictObject(data, path, allowed, required)
	if err != nil {
		return err
	}
	switch kind {
	case TermParam, TermRef, TermHole, TermExists, TermLambda:
		if err := validateTypeJSON(fields["type"], path+".type"); err != nil {
			return err
		}
	}
	switch kind {
	case TermAtom:
		args, err := strictArray(fields["args"], path+".args", DefaultMaxPorts)
		if err != nil {
			return err
		}
		for i, item := range args {
			argPath := fmt.Sprintf("%s.args[%d]", path, i)
			argument, err := strictObject(item, argPath, []string{"role", "value"}, []string{"role", "value"})
			if err != nil {
				return err
			}
			if err := validateTermJSON(argument["value"], argPath+".value"); err != nil {
				return err
			}
		}
	case TermAnd:
		operands, err := strictArray(fields["operands"], path+".operands", DefaultMaxNodes)
		if err != nil {
			return err
		}
		for i, operand := range operands {
			if err := validateTermJSON(operand, fmt.Sprintf("%s.operands[%d]", path, i)); err != nil {
				return err
			}
		}
	case TermExists, TermLambda:
		return validateTermJSON(fields["body"], path+".body")
	case TermApply:
		if err := validateTermJSON(fields["function"], path+".function"); err != nil {
			return err
		}
		return validateTermJSON(fields["argument"], path+".argument")
	}
	return nil
}

func strictObject(data []byte, path string, allowed, required []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, invalid("EXPECTED_OBJECT", path, "expected a non-null JSON object")
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := fields[key]
		if _, ok := allowedSet[key]; !ok {
			return nil, invalid("UNKNOWN_FIELD", appendJSONPath(path, key), "field is not allowed in this object variant")
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, invalid("NULL_FIELD", appendJSONPath(path, key), "field cannot be null")
		}
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return nil, invalid("MISSING_FIELD", appendJSONPath(path, key), "required field is missing")
		}
	}
	return fields, nil
}

func strictArray(data []byte, path string, maxItems int) ([]json.RawMessage, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil || values == nil {
		return nil, invalid("EXPECTED_ARRAY", path, "expected a non-null JSON array")
	}
	if len(values) > maxItems {
		return nil, resourceLimit(path, fmt.Sprintf("array has %d items; maximum is %d", len(values), maxItems))
	}
	return values, nil
}

func appendJSONPath(path, key string) string {
	return path + "." + key
}
