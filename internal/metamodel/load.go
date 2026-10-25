package metamodel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

// Decode strictly reads the manifest wire shape. It does not resolve IDs,
// validate graph relationships, approve symbols, or create a Registry.
func Decode(data []byte) (Document, error) {
	if err := algebra.PreflightJSON(data); err != nil {
		return Document{}, err
	}
	root, err := object(data, "document")
	if err != nil {
		return Document{}, err
	}
	if err := fields(root, "document", []string{"schema_version", "id", "version", "profile", "provenance", "sorts", "symbols", "dependencies", "typical_edges", "axioms"}, []string{"schema_version", "id", "version", "profile", "provenance", "sorts", "symbols", "dependencies", "typical_edges", "axioms"}); err != nil {
		return Document{}, err
	}
	version, err := integer(root["schema_version"], "schema_version")
	if err != nil {
		return Document{}, err
	}
	if version != DocumentSchemaVersion {
		return Document{}, failure(status.OutcomeInvalid, "INVALID_SCHEMA_VERSION", "schema_version", "unsupported manifest version")
	}
	id, err := text(root["id"], "id", 256, true)
	if err != nil {
		return Document{}, err
	}
	manifestVersion, err := text(root["version"], "version", 256, true)
	if err != nil {
		return Document{}, err
	}
	profile, err := text(root["profile"], "profile", 64, true)
	if err != nil {
		return Document{}, err
	}
	if profile != PositiveFactsProfile {
		return Document{}, failure(status.OutcomeUnsupported, "UNSUPPORTED_PROFILE", "profile", "manifest profile is not supported")
	}
	var result Document
	result.SchemaVersion, result.ID, result.Version, result.Profile = version, id, manifestVersion, profile
	if result.Provenance, err = decodeProvenanceArray(root["provenance"], "provenance", true); err != nil {
		return Document{}, err
	}
	if result.Sorts, err = decodeSorts(root["sorts"]); err != nil {
		return Document{}, err
	}
	if result.Symbols, err = decodeSymbols(root["symbols"]); err != nil {
		return Document{}, err
	}
	if result.Dependencies, err = decodeDependencies(root["dependencies"]); err != nil {
		return Document{}, err
	}
	if result.TypicalEdges, err = decodeTypicalEdges(root["typical_edges"]); err != nil {
		return Document{}, err
	}
	axioms, err := array(root["axioms"], "axioms", 512)
	if err != nil {
		return Document{}, err
	}
	if len(axioms) != 0 {
		return Document{}, failure(status.OutcomeUnsupported, "UNSUPPORTED_AXIOM", "axioms", "positive-facts-v1 does not execute axioms")
	}
	result.Axioms = make([]json.RawMessage, 0)
	return result, nil
}

func decodeSorts(raw json.RawMessage) ([]algebra.SortDeclaration, error) {
	items, err := array(raw, "sorts", 512)
	if err != nil {
		return nil, err
	}
	out := make([]algebra.SortDeclaration, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("sorts[%d]", i)
		fieldsByName, err := object(item, path)
		if err != nil {
			return nil, err
		}
		if err := fields(fieldsByName, path, []string{"id", "label"}, []string{"id"}); err != nil {
			return nil, err
		}
		id, err := text(fieldsByName["id"], path+".id", 256, true)
		if err != nil {
			return nil, err
		}
		label, err := optionalText(fieldsByName, "label", path, 1024)
		if err != nil {
			return nil, err
		}
		out = append(out, algebra.SortDeclaration{ID: id, Label: label})
	}
	return out, nil
}

func decodeSymbols(raw json.RawMessage) ([]Symbol, error) {
	items, err := array(raw, "symbols", 512)
	if err != nil {
		return nil, err
	}
	out := make([]Symbol, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("symbols[%d]", i)
		fieldsByName, err := object(item, path)
		if err != nil {
			return nil, err
		}
		allowed := []string{"id", "kind", "label", "description", "review_status", "provenance", "ports", "constraints", "alias"}
		if err := fields(fieldsByName, path, allowed, []string{"id", "kind", "review_status", "provenance"}); err != nil {
			return nil, err
		}
		id, err := text(fieldsByName["id"], path+".id", 256, true)
		if err != nil {
			return nil, err
		}
		kind, err := text(fieldsByName["kind"], path+".kind", 64, true)
		if err != nil {
			return nil, err
		}
		label, err := optionalText(fieldsByName, "label", path, 1024)
		if err != nil {
			return nil, err
		}
		description, err := optionalText(fieldsByName, "description", path, 2048)
		if err != nil {
			return nil, err
		}
		review, err := text(fieldsByName["review_status"], path+".review_status", 32, true)
		if err != nil {
			return nil, err
		}
		if review != string(ReviewPending) && review != string(ReviewApproved) {
			return nil, failure(status.OutcomeInvalid, "INVALID_REVIEW_STATUS", path+".review_status", "review_status must be pending or approved")
		}
		provenanceRecords, err := decodeProvenanceArray(fieldsByName["provenance"], path+".provenance", true)
		if err != nil {
			return nil, err
		}
		symbol := Symbol{ID: id, Kind: SymbolKind(kind), Label: label, Description: description, ReviewStatus: ReviewStatus(review), Provenance: provenanceRecords}
		switch symbol.Kind {
		case SymbolPrimitive:
			if _, exists := fieldsByName["alias"]; exists {
				return nil, failure(status.OutcomeInvalid, "IRRELEVANT_FIELD", path+".alias", "primitive symbol cannot carry alias")
			}
			if _, exists := fieldsByName["ports"]; !exists {
				return nil, failure(status.OutcomeInvalid, "MISSING_FIELD", path+".ports", "primitive symbol requires ports")
			}
			if symbol.Ports, err = decodePorts(fieldsByName["ports"], path+".ports"); err != nil {
				return nil, err
			}
			if rawConstraints, exists := fieldsByName["constraints"]; exists {
				if symbol.Constraints, err = decodeConstraints(rawConstraints, path+".constraints"); err != nil {
					return nil, err
				}
			}
		case SymbolAlias:
			if _, exists := fieldsByName["ports"]; !exists {
				return nil, failure(status.OutcomeInvalid, "MISSING_FIELD", path+".ports", "alias symbol requires ports")
			}
			if symbol.Ports, err = decodePorts(fieldsByName["ports"], path+".ports"); err != nil {
				return nil, err
			}
			if _, exists := fieldsByName["alias"]; !exists {
				return nil, failure(status.OutcomeInvalid, "MISSING_FIELD", path+".alias", "alias symbol requires alias expression")
			}
			if symbol.Alias, err = decodeAlias(fieldsByName["alias"], path+".alias"); err != nil {
				return nil, err
			}
			if rawConstraints, exists := fieldsByName["constraints"]; exists {
				if symbol.Constraints, err = decodeConstraints(rawConstraints, path+".constraints"); err != nil {
					return nil, err
				}
			}
		case SymbolOpaqueUnformalized:
			for _, forbidden := range []string{"ports", "constraints", "alias"} {
				if _, exists := fieldsByName[forbidden]; exists {
					return nil, failure(status.OutcomeInvalid, "IRRELEVANT_FIELD", path+"."+forbidden, "opaque symbol cannot carry executable structure")
				}
			}
		default:
			return nil, failure(status.OutcomeUnsupported, "UNSUPPORTED_SYMBOL_KIND", path+".kind", "symbol kind is unsupported")
		}
		out = append(out, symbol)
	}
	return out, nil
}

func decodePorts(raw json.RawMessage, path string) ([]algebra.Port, error) {
	items, err := array(raw, path, 16)
	if err != nil {
		return nil, err
	}
	out := make([]algebra.Port, 0, len(items))
	for i, item := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		values, err := object(item, itemPath)
		if err != nil {
			return nil, err
		}
		if err := fields(values, itemPath, []string{"role", "type"}, []string{"role", "type"}); err != nil {
			return nil, err
		}
		role, err := text(values["role"], itemPath+".role", 256, true)
		if err != nil {
			return nil, err
		}
		typeValue, err := algebra.DecodeTypeJSON(values["type"], itemPath+".type")
		if err != nil {
			return nil, err
		}
		if typeValue.Kind == algebra.TypeArrow {
			return nil, failure(status.OutcomeInvalid, "INVALID_RELATION_PORT_TYPE", itemPath+".type", "relation port must have a base sort type")
		}
		out = append(out, algebra.Port{Role: role, Type: typeValue})
	}
	return out, nil
}

func decodeAlias(raw json.RawMessage, path string) (*Alias, error) {
	values, err := object(raw, path)
	if err != nil {
		return nil, err
	}
	if err := fields(values, path, []string{"expression", "interface"}, []string{"expression", "interface"}); err != nil {
		return nil, err
	}
	expression, err := algebra.DecodeTermJSON(values["expression"], path+".expression")
	if err != nil {
		return nil, err
	}
	items, err := array(values["interface"], path+".interface", 16)
	if err != nil {
		return nil, err
	}
	bindings := make([]PortBinding, 0, len(items))
	for i, item := range items {
		itemPath := fmt.Sprintf("%s.interface[%d]", path, i)
		binding, err := object(item, itemPath)
		if err != nil {
			return nil, err
		}
		if err := fields(binding, itemPath, []string{"port", "binder"}, []string{"port", "binder"}); err != nil {
			return nil, err
		}
		port, err := text(binding["port"], itemPath+".port", 256, true)
		if err != nil {
			return nil, err
		}
		binder, err := text(binding["binder"], itemPath+".binder", 256, true)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, PortBinding{Port: port, Binder: binder})
	}
	return &Alias{Expression: expression, Interface: bindings}, nil
}

func decodeConstraints(raw json.RawMessage, path string) ([]Constraint, error) {
	items, err := array(raw, path, 512)
	if err != nil {
		return nil, err
	}
	out := make([]Constraint, 0, len(items))
	for i, item := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		base, err := object(item, itemPath)
		if err != nil {
			return nil, err
		}
		if err := fields(base, itemPath, []string{"kind", "count", "role", "type"}, []string{"kind"}); err != nil {
			return nil, err
		}
		kind, err := text(base["kind"], itemPath+".kind", 64, true)
		if err != nil {
			return nil, err
		}
		constraint := Constraint{Kind: ConstraintKind(kind)}
		switch constraint.Kind {
		case ConstraintArity:
			if err := fields(base, itemPath, []string{"kind", "count"}, []string{"kind", "count"}); err != nil {
				return nil, err
			}
			count, err := integer(base["count"], itemPath+".count")
			if err != nil {
				return nil, err
			}
			if count < 0 || count > 16 {
				return nil, failure(status.OutcomeInvalid, "INVALID_CONSTRAINT_VALUE", itemPath+".count", "arity count must be between 0 and 16")
			}
			constraint.Count = &count
		case ConstraintPortType:
			if err := fields(base, itemPath, []string{"kind", "role", "type"}, []string{"kind", "role", "type"}); err != nil {
				return nil, err
			}
			constraint.Role, err = text(base["role"], itemPath+".role", 256, true)
			if err != nil {
				return nil, err
			}
			typeValue, err := algebra.DecodeTypeJSON(base["type"], itemPath+".type")
			if err != nil {
				return nil, err
			}
			if typeValue.Kind == algebra.TypeArrow {
				return nil, failure(status.OutcomeInvalid, "INVALID_RELATION_PORT_TYPE", itemPath+".type", "port constraint type must be a base sort")
			}
			constraint.Type = &typeValue
		default:
			return nil, failure(status.OutcomeUnsupported, "UNSUPPORTED_CONSTRAINT", itemPath+".kind", "constraint kind is unsupported")
		}
		out = append(out, constraint)
	}
	return out, nil
}

func decodeProvenanceArray(raw json.RawMessage, path string, nonempty bool) ([]provenance.SourceRef, error) {
	items, err := array(raw, path, 512)
	if err != nil {
		return nil, err
	}
	if nonempty && len(items) == 0 {
		return nil, failure(status.OutcomeInvalid, "EMPTY_PROVENANCE", path, "provenance must contain at least one source")
	}
	out := make([]provenance.SourceRef, 0, len(items))
	for i, item := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		values, err := object(item, itemPath)
		if err != nil {
			return nil, err
		}
		allowed := []string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}
		if err := fields(values, itemPath, allowed, []string{"id", "version", "license"}); err != nil {
			return nil, err
		}
		var source provenance.SourceRef
		if source.ID, err = text(values["id"], itemPath+".id", provenance.MaxSourceIDBytes, true); err != nil {
			return nil, err
		}
		if source.Version, err = text(values["version"], itemPath+".version", provenance.MaxSourceVersionBytes, true); err != nil {
			return nil, err
		}
		if source.License, err = text(values["license"], itemPath+".license", provenance.MaxSourceLicenseBytes, true); err != nil {
			return nil, err
		}
		if source.Path, err = optionalText(values, "path", itemPath, provenance.MaxSourceLocationBytes); err != nil {
			return nil, err
		}
		if source.URI, err = optionalText(values, "uri", itemPath, provenance.MaxSourceLocationBytes); err != nil {
			return nil, err
		}
		if source.Author, err = optionalText(values, "author", itemPath, provenance.MaxSourceAuthorBytes); err != nil {
			return nil, err
		}
		if source.ChecksumSHA256, err = optionalText(values, "checksum_sha256", itemPath, provenance.SHA256HexBytes); err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	return out, nil
}

func decodeDependencies(raw json.RawMessage) ([]Dependency, error) {
	items, err := array(raw, "dependencies", 16)
	if err != nil {
		return nil, err
	}
	out := make([]Dependency, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("dependencies[%d]", i)
		values, err := object(item, path)
		if err != nil {
			return nil, err
		}
		if err := fields(values, path, []string{"id", "version", "fingerprint"}, []string{"id", "version", "fingerprint"}); err != nil {
			return nil, err
		}
		id, err := text(values["id"], path+".id", 256, true)
		if err != nil {
			return nil, err
		}
		version, err := text(values["version"], path+".version", 256, true)
		if err != nil {
			return nil, err
		}
		fingerprint, err := text(values["fingerprint"], path+".fingerprint", 128, true)
		if err != nil {
			return nil, err
		}
		out = append(out, Dependency{ID: id, Version: version, Fingerprint: fingerprint})
	}
	return out, nil
}

func decodeTypicalEdges(raw json.RawMessage) ([]TypicalEdge, error) {
	items, err := array(raw, "typical_edges", 512)
	if err != nil {
		return nil, err
	}
	out := make([]TypicalEdge, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("typical_edges[%d]", i)
		values, err := object(item, path)
		if err != nil {
			return nil, err
		}
		if err := fields(values, path, []string{"from", "to", "kind"}, []string{"from", "to", "kind"}); err != nil {
			return nil, err
		}
		from, err := text(values["from"], path+".from", 256, true)
		if err != nil {
			return nil, err
		}
		to, err := text(values["to"], path+".to", 256, true)
		if err != nil {
			return nil, err
		}
		kind, err := text(values["kind"], path+".kind", 64, true)
		if err != nil {
			return nil, err
		}
		out = append(out, TypicalEdge{From: from, To: to, Kind: kind})
	}
	return out, nil
}

func object(raw []byte, path string) (map[string]json.RawMessage, error) {
	var result map[string]json.RawMessage
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		return nil, failure(status.OutcomeInvalid, "EXPECTED_OBJECT", path, "expected a non-null JSON object")
	}
	keys := make([]string, 0, len(result))
	for key := range result {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := result[key]
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, failure(status.OutcomeInvalid, "NULL_FIELD", path+"."+key, "field must not be null")
		}
	}
	return result, nil
}

func fields(values map[string]json.RawMessage, path string, allowed, required []string) error {
	allow := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allow[name] = struct{}{}
	}
	keys := make([]string, 0, len(values))
	for name := range values {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		if _, ok := allow[name]; !ok {
			return failure(status.OutcomeInvalid, "UNKNOWN_FIELD", path+"."+name, "field is not allowed")
		}
	}
	for _, name := range required {
		if _, ok := values[name]; !ok {
			return failure(status.OutcomeInvalid, "MISSING_FIELD", path+"."+name, "required field is missing")
		}
	}
	return nil
}

func array(raw []byte, path string, maximum int) ([]json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, failure(status.OutcomeInvalid, "MISSING_FIELD", path, "required array is missing")
	}
	var result []json.RawMessage
	if err := json.Unmarshal(raw, &result); err != nil || result == nil {
		return nil, failure(status.OutcomeInvalid, "EXPECTED_ARRAY", path, "expected a non-null JSON array")
	}
	if len(result) > maximum {
		return nil, failure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", path, fmt.Sprintf("array has %d items; maximum is %d", len(result), maximum))
	}
	return result, nil
}

func text(raw []byte, path string, maximum int, nonempty bool) (string, error) {
	var result string
	if err := json.Unmarshal(raw, &result); err != nil || !utf8.ValidString(result) {
		return "", failure(status.OutcomeInvalid, "EXPECTED_STRING", path, "expected a non-null UTF-8 string")
	}
	if len(result) > maximum || (nonempty && (result == "" || strings.TrimSpace(result) == "")) {
		return "", failure(status.OutcomeInvalid, "INVALID_TEXT", path, fmt.Sprintf("string must be non-empty and at most %d UTF-8 bytes", maximum))
	}
	return result, nil
}

func optionalText(values map[string]json.RawMessage, name, parent string, maximum int) (string, error) {
	raw, ok := values[name]
	if !ok {
		return "", nil
	}
	return text(raw, parent+"."+name, maximum, false)
}

func integer(raw []byte, path string) (int, error) {
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, failure(status.OutcomeInvalid, "EXPECTED_INTEGER", path, "expected an integer")
	}
	return value, nil
}

func failure(kind status.OutcomeKind, code, path, message string) error {
	return &algebra.Error{Status: kind, Code: code, Path: path, Message: message}
}
