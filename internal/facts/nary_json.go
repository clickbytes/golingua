package facts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
)

const MaxNaryInputBytes = algebra.MaxDocumentBytes

// DecodeNaryInput strictly decodes one bounded request. It validates wire
// shape only; BuildNary must still validate the input against selected facts
// metadata before the records are usable as evidence.
func DecodeNaryInput(data []byte) (NaryInput, error) {
	if err := algebra.PreflightJSON(data); err != nil {
		return NaryInput{}, err
	}
	fields, err := strictNaryObject(data, "request", []string{"schema_version", "snapshot_id", "context_id", "referents", "facts", "at", "limits"}, []string{"schema_version", "context_id", "referents", "facts"})
	if err != nil {
		return NaryInput{}, err
	}
	version, err := naryJSONInt(fields["schema_version"], "request.schema_version", 1, 1)
	if err != nil {
		return NaryInput{}, err
	}
	contextID, err := naryJSONString(fields["context_id"], "request.context_id", 256, false)
	if err != nil || !validNaryID(contextID) {
		if err != nil {
			return NaryInput{}, err
		}
		return NaryInput{}, factError("INVALID_FACT_CONTEXT", "request.context_id", "context ID must be a bounded identifier")
	}
	input := NaryInput{SchemaVersion: version, ContextID: contextID}
	if raw, exists := fields["snapshot_id"]; exists {
		input.SnapshotID, err = naryJSONString(raw, "request.snapshot_id", 256, false)
		if err != nil || !validNaryID(input.SnapshotID) {
			if err != nil {
				return NaryInput{}, err
			}
			return NaryInput{}, factError("INVALID_FACT_SNAPSHOT", "request.snapshot_id", "snapshot ID must be a bounded identifier")
		}
	}
	if raw, exists := fields["at"]; exists {
		input.At, err = naryJSONString(raw, "request.at", 128, true)
		if err != nil {
			return NaryInput{}, err
		}
	}
	if raw, exists := fields["limits"]; exists {
		input.Limits, err = decodeNaryLimits(raw, "request.limits")
		if err != nil {
			return NaryInput{}, err
		}
	}
	limits, err := NormalizeNaryLimits(input.Limits)
	if err != nil {
		return NaryInput{}, err
	}
	referentRaw, err := strictNaryArray(fields["referents"], "request.referents", limits.MaxReferents)
	if err != nil {
		return NaryInput{}, err
	}
	factRaw, err := strictNaryArray(fields["facts"], "request.facts", limits.MaxFacts)
	if err != nil {
		return NaryInput{}, err
	}
	input.Referents = make([]algebra.Declaration, 0, len(referentRaw))
	for index, raw := range referentRaw {
		declaration, decodeErr := decodeNaryReferent(raw, fmt.Sprintf("request.referents[%d]", index))
		if decodeErr != nil {
			return NaryInput{}, decodeErr
		}
		input.Referents = append(input.Referents, declaration)
	}
	input.Facts = make([]NaryFact, 0, len(factRaw))
	for index, raw := range factRaw {
		fact, decodeErr := decodeNaryFact(raw, fmt.Sprintf("request.facts[%d]", index))
		if decodeErr != nil {
			return NaryInput{}, decodeErr
		}
		input.Facts = append(input.Facts, fact)
	}
	return input, nil
}

// EncodeNaryInput emits a strict, bounded wire representation. It preserves
// explicit identities and always writes the required empty arrays.
func EncodeNaryInput(input NaryInput) ([]byte, error) {
	copyOf, err := cloneNaryInputForWire(input)
	if err != nil {
		return nil, err
	}
	fields := make(map[string]json.RawMessage, 7)
	put := func(name string, value any) error {
		encoded, marshalErr := json.Marshal(value)
		if marshalErr != nil {
			return factError("INVALID_FACT_WIRE_VALUE", "request."+name, "value cannot be encoded in the fact request format")
		}
		fields[name] = encoded
		return nil
	}
	if err := put("schema_version", copyOf.SchemaVersion); err != nil {
		return nil, err
	}
	if copyOf.SnapshotID != "" {
		if err := put("snapshot_id", copyOf.SnapshotID); err != nil {
			return nil, err
		}
	}
	if err := put("context_id", copyOf.ContextID); err != nil {
		return nil, err
	}
	referents, err := marshalNaryArray(copyOf.Referents, "referents")
	if err != nil {
		return nil, err
	}
	fields["referents"] = referents
	facts, err := marshalNaryArray(copyOf.Facts, "facts")
	if err != nil {
		return nil, err
	}
	fields["facts"] = facts
	if copyOf.At != "" {
		if err := put("at", copyOf.At); err != nil {
			return nil, err
		}
	}
	if err := put("limits", copyOf.Limits); err != nil {
		return nil, err
	}
	projectedSize := 2 // outer braces
	for key, value := range fields {
		projectedSize += len(key) + 3 + len(value) // quoted key and colon
	}
	if len(fields) > 1 {
		projectedSize += len(fields) - 1 // commas
	}
	if projectedSize > MaxNaryInputBytes {
		return nil, factLimit("request.bytes", projectedSize)
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, factError("INVALID_FACT_WIRE_VALUE", "request", "request cannot be encoded")
	}
	if len(encoded) > MaxNaryInputBytes {
		return nil, factLimit("request.bytes", len(encoded))
	}
	if _, err := DecodeNaryInput(encoded); err != nil {
		return nil, err
	}
	return encoded, nil
}

func decodeNaryLimits(data []byte, path string) (NaryLimits, error) {
	fields, err := strictNaryObject(data, path, []string{"max_facts", "max_referents"}, nil)
	if err != nil {
		return NaryLimits{}, err
	}
	var limits NaryLimits
	if raw, exists := fields["max_facts"]; exists {
		limits.MaxFacts, err = naryJSONInt(raw, path+".max_facts", 0, DefaultNaryMaxFacts)
		if err != nil {
			return NaryLimits{}, err
		}
	}
	if raw, exists := fields["max_referents"]; exists {
		limits.MaxReferents, err = naryJSONInt(raw, path+".max_referents", 0, DefaultNaryMaxReferents)
		if err != nil {
			return NaryLimits{}, err
		}
	}
	return limits, nil
}

func decodeNaryReferent(data []byte, path string) (algebra.Declaration, error) {
	fields, err := strictNaryObject(data, path, []string{"id", "type"}, []string{"id", "type"})
	if err != nil {
		return algebra.Declaration{}, err
	}
	id, err := naryJSONString(fields["id"], path+".id", 256, false)
	if err != nil || !validNaryID(id) {
		if err != nil {
			return algebra.Declaration{}, err
		}
		return algebra.Declaration{}, factError("INVALID_REFERENT_ID", path+".id", "referent ID must be a bounded identifier")
	}
	typ, err := algebra.DecodeTypeJSON(fields["type"], path+".type")
	if err != nil {
		return algebra.Declaration{}, err
	}
	return algebra.Declaration{ID: id, Type: typ}, nil
}

func decodeNaryFact(data []byte, path string) (NaryFact, error) {
	fields, err := strictNaryObject(data, path, []string{"id", "atom", "polarity", "valid_time", "source", "assertion_scope"}, []string{"id", "atom", "polarity", "source", "assertion_scope"})
	if err != nil {
		return NaryFact{}, err
	}
	id, err := naryJSONString(fields["id"], path+".id", 128, false)
	if err != nil || !validLocalID(id) {
		if err != nil {
			return NaryFact{}, err
		}
		return NaryFact{}, factError("INVALID_FACT_ID", path+".id", "fact ID must be a bounded identifier")
	}
	atom, err := decodeNaryAtom(fields["atom"], path+".atom")
	if err != nil {
		return NaryFact{}, err
	}
	polarity, err := naryJSONString(fields["polarity"], path+".polarity", 16, false)
	if err != nil {
		return NaryFact{}, err
	}
	if Polarity(polarity) != Positive && Polarity(polarity) != Negative {
		return NaryFact{}, factError("INVALID_FACT_POLARITY", path+".polarity", "polarity must be positive or negative")
	}
	source, err := decodeNarySource(fields["source"], path+".source")
	if err != nil {
		return NaryFact{}, err
	}
	scope, err := naryJSONString(fields["assertion_scope"], path+".assertion_scope", 32, false)
	if err != nil {
		return NaryFact{}, err
	}
	switch AssertionScope(scope) {
	case ScopeAsserted, ScopeQuestion, ScopeHypothetical, ScopeNegated:
	default:
		return NaryFact{}, factError("INVALID_FACT_SCOPE", path+".assertion_scope", "assertion scope is not recognized")
	}
	var validTime ValidTime
	if raw, exists := fields["valid_time"]; exists {
		validTime, err = decodeNaryValidTime(raw, path+".valid_time")
		if err != nil {
			return NaryFact{}, err
		}
	}
	return NaryFact{ID: id, Atom: atom, Polarity: Polarity(polarity), ValidTime: validTime, Source: source, AssertionScope: AssertionScope(scope)}, nil
}

func decodeNaryAtom(data []byte, path string) (NaryAtom, error) {
	fields, err := strictNaryObject(data, path, []string{"symbol", "args"}, []string{"symbol", "args"})
	if err != nil {
		return NaryAtom{}, err
	}
	symbol, err := naryJSONString(fields["symbol"], path+".symbol", 256, false)
	if err != nil || !validNaryID(symbol) {
		if err != nil {
			return NaryAtom{}, err
		}
		return NaryAtom{}, factError("INVALID_FACT_SYMBOL", path+".symbol", "symbol ID must be a bounded identifier")
	}
	args, err := strictNaryArray(fields["args"], path+".args", algebra.DefaultMaxPorts)
	if err != nil {
		return NaryAtom{}, err
	}
	atom := NaryAtom{Symbol: symbol, Args: make([]NaryArgument, 0, len(args))}
	for index, raw := range args {
		argument, decodeErr := decodeNaryArgument(raw, fmt.Sprintf("%s.args[%d]", path, index))
		if decodeErr != nil {
			return NaryAtom{}, decodeErr
		}
		atom.Args = append(atom.Args, argument)
	}
	return atom, nil
}

func decodeNaryArgument(data []byte, path string) (NaryArgument, error) {
	fields, err := strictNaryObject(data, path, []string{"role", "value"}, []string{"role", "value"})
	if err != nil {
		return NaryArgument{}, err
	}
	role, err := naryJSONString(fields["role"], path+".role", 256, false)
	if err != nil || !validNaryID(role) {
		if err != nil {
			return NaryArgument{}, err
		}
		return NaryArgument{}, factError("INVALID_FACT_ROLE", path+".role", "role ID must be a bounded identifier")
	}
	valueFields, err := strictNaryObject(fields["value"], path+".value", []string{"kind", "id"}, []string{"kind", "id"})
	if err != nil {
		return NaryArgument{}, err
	}
	kind, err := naryJSONString(valueFields["kind"], path+".value.kind", 32, false)
	if err != nil {
		return NaryArgument{}, err
	}
	valueKind := algebra.TermKind(kind)
	switch valueKind {
	case algebra.TermRef, algebra.TermVar, algebra.TermParam, algebra.TermHole:
	default:
		return NaryArgument{}, factError("INVALID_FACT_VALUE_KIND", path+".value.kind", "value kind must be Ref, Var, Param, or Hole")
	}
	id, err := naryJSONString(valueFields["id"], path+".value.id", 256, false)
	if err != nil || !validNaryID(id) {
		if err != nil {
			return NaryArgument{}, err
		}
		return NaryArgument{}, factError("INVALID_FACT_REFERENT", path+".value.id", "value ID must be a bounded identifier")
	}
	return NaryArgument{Role: role, Value: NaryValue{Kind: valueKind, ID: id}}, nil
}

func decodeNarySource(data []byte, path string) (SourceEvidence, error) {
	fields, err := strictNaryObject(data, path, []string{"kind", "id", "reference"}, []string{"kind", "id"})
	if err != nil {
		return SourceEvidence{}, err
	}
	kind, err := naryJSONString(fields["kind"], path+".kind", 32, false)
	if err != nil {
		return SourceEvidence{}, err
	}
	if SourceKind(kind) != SourceUser && SourceKind(kind) != SourceDocument {
		return SourceEvidence{}, factError("INVALID_FACT_SOURCE", path+".kind", "source kind must be user or document")
	}
	id, err := naryJSONString(fields["id"], path+".id", 128, false)
	if err != nil || !validLocalID(id) {
		if err != nil {
			return SourceEvidence{}, err
		}
		return SourceEvidence{}, factError("INVALID_FACT_SOURCE", path+".id", "source ID must be a bounded identifier")
	}
	source := SourceEvidence{Kind: SourceKind(kind), ID: id}
	if raw, exists := fields["reference"]; exists {
		reference, decodeErr := decodeNarySourceRef(raw, path+".reference")
		if decodeErr != nil {
			return SourceEvidence{}, decodeErr
		}
		source.Reference = &reference
	}
	return source, nil
}

func decodeNarySourceRef(data []byte, path string) (provenance.SourceRef, error) {
	fields, err := strictNaryObject(data, path, []string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}, []string{"id", "version", "license"})
	if err != nil {
		return provenance.SourceRef{}, err
	}
	var result provenance.SourceRef
	if result.ID, err = naryJSONString(fields["id"], path+".id", provenance.MaxSourceIDBytes, false); err != nil {
		return provenance.SourceRef{}, err
	}
	if result.Version, err = naryJSONString(fields["version"], path+".version", provenance.MaxSourceVersionBytes, false); err != nil {
		return provenance.SourceRef{}, err
	}
	if result.License, err = naryJSONString(fields["license"], path+".license", provenance.MaxSourceLicenseBytes, false); err != nil {
		return provenance.SourceRef{}, err
	}
	if raw, exists := fields["path"]; exists {
		if result.Path, err = naryJSONString(raw, path+".path", provenance.MaxSourceLocationBytes, true); err != nil {
			return provenance.SourceRef{}, err
		}
	}
	if raw, exists := fields["uri"]; exists {
		if result.URI, err = naryJSONString(raw, path+".uri", provenance.MaxSourceLocationBytes, true); err != nil {
			return provenance.SourceRef{}, err
		}
	}
	if raw, exists := fields["author"]; exists {
		if result.Author, err = naryJSONString(raw, path+".author", provenance.MaxSourceAuthorBytes, true); err != nil {
			return provenance.SourceRef{}, err
		}
	}
	if raw, exists := fields["checksum_sha256"]; exists {
		if result.ChecksumSHA256, err = naryJSONString(raw, path+".checksum_sha256", provenance.SHA256HexBytes, true); err != nil {
			return provenance.SourceRef{}, err
		}
	}
	return result, nil
}

func decodeNaryValidTime(data []byte, path string) (ValidTime, error) {
	fields, err := strictNaryObject(data, path, []string{"from", "to"}, nil)
	if err != nil {
		return ValidTime{}, err
	}
	var result ValidTime
	if raw, exists := fields["from"]; exists {
		result.From, err = naryJSONString(raw, path+".from", 128, true)
		if err != nil {
			return ValidTime{}, err
		}
	}
	if raw, exists := fields["to"]; exists {
		result.To, err = naryJSONString(raw, path+".to", 128, true)
		if err != nil {
			return ValidTime{}, err
		}
	}
	return result, nil
}

func strictNaryObject(data []byte, path string, allowed, required []string) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return nil, factError("EXPECTED_OBJECT", path, "expected a non-null JSON object")
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	fields := make(map[string]json.RawMessage, len(allowed))
	for decoder.More() {
		token, tokenErr := decoder.Token()
		key, isString := token.(string)
		if tokenErr != nil || !isString {
			return nil, factError("INVALID_OBJECT_KEY", path, "object property name must be a string")
		}
		if _, ok := allowedSet[key]; !ok {
			return nil, factError("UNKNOWN_FIELD", path+"."+key, "field is not allowed in this object")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, factError("INVALID_JSON_FIELD", path+"."+key, "field value could not be decoded")
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, factError("NULL_FIELD", path+"."+key, "field cannot be null")
		}
		fields[key] = value
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return nil, factError("INVALID_JSON_OBJECT", path, "object is not properly terminated")
	}
	if err := requireNaryJSONEOF(decoder, path); err != nil {
		return nil, err
	}
	for _, key := range required {
		if _, exists := fields[key]; !exists {
			return nil, factError("MISSING_FIELD", path+"."+key, "required field is missing")
		}
	}
	return fields, nil
}

func strictNaryArray(data []byte, path string, maximum int) ([]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('[') {
		return nil, factError("EXPECTED_ARRAY", path, "expected a non-null JSON array")
	}
	values := make([]json.RawMessage, 0)
	for decoder.More() {
		if len(values) >= maximum {
			return nil, factLimit(path, len(values)+1)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, factError("INVALID_JSON_ARRAY", fmt.Sprintf("%s[%d]", path, len(values)), "array item could not be decoded")
		}
		values = append(values, value)
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim(']') {
		return nil, factError("INVALID_JSON_ARRAY", path, "array is not properly terminated")
	}
	if err := requireNaryJSONEOF(decoder, path); err != nil {
		return nil, err
	}
	return values, nil
}

func requireNaryJSONEOF(decoder *json.Decoder, path string) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return factError("TRAILING_DOCUMENT", path, "expected exactly one JSON value")
	}
	return nil
}

func naryJSONString(data []byte, path string, maximum int, allowEmpty bool) (string, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return "", factError("NULL_FIELD", path, "string cannot be null")
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", factError("INVALID_STRING", path, "expected a JSON string")
	}
	if !utf8.ValidString(value) || len(value) > maximum || (!allowEmpty && value == "") {
		return "", factError("INVALID_STRING", path, "string is empty or exceeds its byte limit")
	}
	return value, nil
}

func naryJSONInt(data []byte, path string, minimum, maximum int) (int, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return 0, factError("INVALID_INTEGER", path, "integer cannot be null")
	}
	var value int
	if err := json.Unmarshal(data, &value); err != nil || value < minimum || value > maximum {
		return 0, factError("INVALID_INTEGER", path, fmt.Sprintf("integer must be between %d and %d", minimum, maximum))
	}
	return value, nil
}

func cloneNaryInputForWire(input NaryInput) (NaryInput, error) {
	if input.SchemaVersion != SchemaVersion {
		return NaryInput{}, factError("INVALID_FACT_SCHEMA_VERSION", "request.schema_version", "schema version must be 1")
	}
	if !validNaryID(input.ContextID) || (input.SnapshotID != "" && !validNaryID(input.SnapshotID)) {
		return NaryInput{}, factError("INVALID_FACT_CONTEXT", "request", "context and optional snapshot IDs must be bounded identifiers")
	}
	if len(input.At) > 128 || !utf8.ValidString(input.At) {
		return NaryInput{}, factError("INVALID_FACT_QUERY_TIME", "request.at", "query instant must be bounded valid UTF-8")
	}
	limits, err := NormalizeNaryLimits(input.Limits)
	if err != nil {
		return NaryInput{}, err
	}
	if len(input.Referents) > limits.MaxReferents || len(input.Facts) > limits.MaxFacts {
		return NaryInput{}, factLimit("request", len(input.Referents)+len(input.Facts))
	}
	result := NaryInput{SchemaVersion: input.SchemaVersion, SnapshotID: input.SnapshotID, ContextID: input.ContextID, At: input.At, Limits: input.Limits}
	result.Referents = make([]algebra.Declaration, len(input.Referents))
	for index, declaration := range input.Referents {
		if !validNaryID(declaration.ID) {
			return NaryInput{}, factError("INVALID_REFERENT_ID", fmt.Sprintf("request.referents[%d].id", index), "referent ID must be a bounded identifier")
		}
		typ, typeErr := cloneNaryWireType(declaration.Type, fmt.Sprintf("request.referents[%d].type", index), 1, make(map[*algebra.Type]bool))
		if typeErr != nil {
			return NaryInput{}, typeErr
		}
		result.Referents[index] = algebra.Declaration{ID: declaration.ID, Type: typ}
	}
	result.Facts = make([]NaryFact, len(input.Facts))
	for index, fact := range input.Facts {
		path := fmt.Sprintf("request.facts[%d]", index)
		if !validLocalID(fact.ID) {
			return NaryInput{}, factError("INVALID_FACT_ID", path+".id", "fact ID must be a bounded identifier")
		}
		if !validNaryID(fact.Atom.Symbol) || len(fact.Atom.Args) > algebra.DefaultMaxPorts {
			return NaryInput{}, factError("INVALID_FACT_ATOM", path+".atom", "atom symbol or argument count exceeds its bound")
		}
		fact.Atom.Args = append([]NaryArgument{}, fact.Atom.Args...)
		for argumentIndex, argument := range fact.Atom.Args {
			if !validNaryID(argument.Role) || !validNaryID(argument.Value.ID) {
				return NaryInput{}, factError("INVALID_FACT_ARGUMENT", fmt.Sprintf("%s.atom.args[%d]", path, argumentIndex), "role and value IDs must be bounded identifiers")
			}
			switch argument.Value.Kind {
			case algebra.TermRef, algebra.TermVar, algebra.TermParam, algebra.TermHole:
			default:
				return NaryInput{}, factError("INVALID_FACT_VALUE_KIND", fmt.Sprintf("%s.atom.args[%d].value.kind", path, argumentIndex), "value kind must be Ref, Var, Param, or Hole")
			}
		}
		if fact.Polarity != Positive && fact.Polarity != Negative {
			return NaryInput{}, factError("INVALID_FACT_POLARITY", path+".polarity", "polarity must be positive or negative")
		}
		switch fact.AssertionScope {
		case ScopeAsserted, ScopeQuestion, ScopeHypothetical, ScopeNegated:
		default:
			return NaryInput{}, factError("INVALID_FACT_SCOPE", path+".assertion_scope", "assertion scope is not recognized")
		}
		if err := validateNarySourceShape(fact.Source, path+".source"); err != nil {
			return NaryInput{}, err
		}
		if len(fact.ValidTime.From) > 128 || len(fact.ValidTime.To) > 128 || !utf8.ValidString(fact.ValidTime.From) || !utf8.ValidString(fact.ValidTime.To) {
			return NaryInput{}, factError("INVALID_FACT_VALID_TIME", path+".valid_time", "valid-time endpoints must be bounded valid UTF-8")
		}
		result.Facts[index] = cloneNaryFact(fact)
	}
	if result.Referents == nil {
		result.Referents = make([]algebra.Declaration, 0)
	}
	if result.Facts == nil {
		result.Facts = make([]NaryFact, 0)
	}
	return result, nil
}

func validateNarySourceShape(source SourceEvidence, path string) error {
	if source.Kind != SourceUser && source.Kind != SourceDocument {
		return factError("INVALID_FACT_SOURCE", path+".kind", "source kind must be user or document")
	}
	if !validLocalID(source.ID) {
		return factError("INVALID_FACT_SOURCE", path+".id", "source ID must be a bounded identifier")
	}
	if source.Reference != nil {
		value := source.Reference
		if value.ID == "" || len(value.ID) > provenance.MaxSourceIDBytes || value.Version == "" || len(value.Version) > provenance.MaxSourceVersionBytes || value.License == "" || len(value.License) > provenance.MaxSourceLicenseBytes || len(value.Path) > provenance.MaxSourceLocationBytes || len(value.URI) > provenance.MaxSourceLocationBytes || len(value.Author) > provenance.MaxSourceAuthorBytes || len(value.ChecksumSHA256) > provenance.SHA256HexBytes {
			return factError("INVALID_FACT_SOURCE", path+".reference", "source reference fields are missing or exceed their wire bounds")
		}
		for _, text := range []string{value.ID, value.Path, value.URI, value.Version, value.License, value.Author, value.ChecksumSHA256} {
			if !utf8.ValidString(text) {
				return factError("INVALID_FACT_SOURCE", path+".reference", "source reference contains invalid UTF-8")
			}
		}
	}
	return nil
}

func cloneNaryWireType(value algebra.Type, path string, depth int, active map[*algebra.Type]bool) (algebra.Type, error) {
	if depth > algebra.DefaultMaxTypeDepth {
		return algebra.Type{}, factLimit(path, depth)
	}
	switch value.Kind {
	case algebra.TypeEntity, algebra.TypeEvent, algebra.TypeInterval, algebra.TypeProp:
		if value.SortID != "" || value.From != nil || value.To != nil {
			return algebra.Type{}, factError("INVALID_TYPE_SHAPE", path, "built-in sort type has irrelevant fields")
		}
		return value, nil
	case algebra.TypeNominal:
		if !validNaryID(value.SortID) || value.From != nil || value.To != nil {
			return algebra.Type{}, factError("INVALID_TYPE_SHAPE", path, "nominal type requires only a bounded sort ID")
		}
		return value, nil
	case algebra.TypeArrow:
		if value.SortID != "" || value.From == nil || value.To == nil {
			return algebra.Type{}, factError("INVALID_TYPE_SHAPE", path, "arrow type requires from and to types")
		}
		if active[value.From] || active[value.To] {
			return algebra.Type{}, factError("INVALID_TYPE_CYCLE", path, "type graph contains a cycle")
		}
		active[value.From] = true
		from, err := cloneNaryWireType(*value.From, path+".from", depth+1, active)
		delete(active, value.From)
		if err != nil {
			return algebra.Type{}, err
		}
		active[value.To] = true
		to, err := cloneNaryWireType(*value.To, path+".to", depth+1, active)
		delete(active, value.To)
		if err != nil {
			return algebra.Type{}, err
		}
		return algebra.Type{Kind: algebra.TypeArrow, From: &from, To: &to}, nil
	default:
		return algebra.Type{}, factError("INVALID_TYPE_KIND", path+".kind", "type kind is not recognized")
	}
}

func marshalNaryArray[T any](values []T, path string) (json.RawMessage, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('[')
	for index, value := range values {
		if index > 0 {
			buffer.WriteByte(',')
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, factError("INVALID_FACT_WIRE_VALUE", fmt.Sprintf("request.%s[%d]", path, index), "array item cannot be encoded")
		}
		if buffer.Len()+len(encoded)+1 > MaxNaryInputBytes {
			return nil, factLimit("request.bytes", buffer.Len()+len(encoded)+1)
		}
		buffer.Write(encoded)
	}
	buffer.WriteByte(']')
	return json.RawMessage(buffer.Bytes()), nil
}
