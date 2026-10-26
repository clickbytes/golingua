package synthesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
)

// DecodeGoal strictly decodes one bounded GoalSpec. It checks wire shape and
// resource bounds only; context-dependent types and symbols remain the job of
// ValidateGoal with an explicitly selected declaration context.
func DecodeGoal(data []byte) (GoalSpec, error) {
	if err := algebra.PreflightJSON(data); err != nil {
		return GoalSpec{}, err
	}
	root, err := goalJSONObject(data, "goal", []string{"schema_version", "profile", "context_id", "expected_type", "ports", "references", "allowed_operators", "constraints", "provenance", "limits"}, []string{"schema_version", "profile", "context_id", "expected_type", "ports", "references", "allowed_operators", "constraints", "provenance", "limits"})
	if err != nil {
		return GoalSpec{}, err
	}
	var goal GoalSpec
	goal.SchemaVersion, err = goalJSONInt(root["schema_version"], "schema_version", -1, 1<<31-1)
	if err != nil {
		return GoalSpec{}, err
	}
	if goal.SchemaVersion != GoalSchemaVersion {
		return GoalSpec{}, goalInvalid("INVALID_GOAL_SCHEMA_VERSION", "schema_version", "goal schema version is unsupported")
	}
	goal.Profile, err = goalJSONString(root["profile"], "profile", 256, false)
	if err != nil {
		return GoalSpec{}, err
	}
	if goal.Profile != GoalProfile {
		return GoalSpec{}, goalUnsupported("UNSUPPORTED_GOAL_PROFILE", "profile", "goal profile is unsupported")
	}
	goal.ContextID, err = goalJSONString(root["context_id"], "context_id", 256, false)
	if err != nil {
		return GoalSpec{}, err
	}
	if !goalID(goal.ContextID) {
		return GoalSpec{}, goalInvalid("INVALID_CONTEXT_ID", "context_id", "context ID must be a bounded identifier")
	}
	goal.ExpectedType, err = algebra.DecodeTypeJSON(root["expected_type"], "expected_type")
	if err != nil {
		return GoalSpec{}, err
	}

	portValues, err := goalJSONArray(root["ports"], "ports", MaxGoalPorts)
	if err != nil {
		return GoalSpec{}, err
	}
	goal.Ports = make([]GoalPort, 0, len(portValues))
	for i, raw := range portValues {
		port, decodeErr := decodeGoalPort(raw, fmt.Sprintf("ports[%d]", i))
		if decodeErr != nil {
			return GoalSpec{}, decodeErr
		}
		goal.Ports = append(goal.Ports, port)
	}

	referenceValues, err := goalJSONArray(root["references"], "references", algebra.DefaultMaxSymbols)
	if err != nil {
		return GoalSpec{}, err
	}
	goal.References = make([]algebra.Declaration, 0, len(referenceValues))
	for i, raw := range referenceValues {
		ref, decodeErr := decodeGoalReference(raw, fmt.Sprintf("references[%d]", i))
		if decodeErr != nil {
			return GoalSpec{}, decodeErr
		}
		goal.References = append(goal.References, ref)
	}

	operatorValues, err := goalJSONArray(root["allowed_operators"], "allowed_operators", len(allowedGoalOperators))
	if err != nil {
		return GoalSpec{}, err
	}
	goal.AllowedOperators = make([]string, 0, len(operatorValues))
	for i, raw := range operatorValues {
		operator, stringErr := goalJSONString(raw, fmt.Sprintf("allowed_operators[%d]", i), 256, false)
		if stringErr != nil {
			return GoalSpec{}, stringErr
		}
		if _, supported := allowedGoalOperators[operator]; !supported {
			return GoalSpec{}, goalUnsupported("UNSUPPORTED_GOAL_OPERATOR", fmt.Sprintf("allowed_operators[%d]", i), "operator is outside the bounded goal profile")
		}
		goal.AllowedOperators = append(goal.AllowedOperators, operator)
	}

	constraintValues, err := goalJSONArray(root["constraints"], "constraints", MaxGoalConstraints)
	if err != nil {
		return GoalSpec{}, err
	}
	goal.Constraints = make([]Constraint, 0, len(constraintValues))
	for i, raw := range constraintValues {
		constraint, decodeErr := decodeGoalConstraint(raw, fmt.Sprintf("constraints[%d]", i))
		if decodeErr != nil {
			return GoalSpec{}, decodeErr
		}
		goal.Constraints = append(goal.Constraints, constraint)
	}

	sourceValues, err := goalJSONArray(root["provenance"], "provenance", MaxGoalSources)
	if err != nil {
		return GoalSpec{}, err
	}
	goal.Provenance = make([]provenance.SourceRef, 0, len(sourceValues))
	for i, raw := range sourceValues {
		source, decodeErr := decodeGoalSource(raw, fmt.Sprintf("provenance[%d]", i))
		if decodeErr != nil {
			return GoalSpec{}, decodeErr
		}
		goal.Provenance = append(goal.Provenance, source)
	}
	goal.Limits, err = decodeGoalLimits(root["limits"])
	if err != nil {
		return GoalSpec{}, err
	}
	if _, err := NormalizeGoalLimits(goal.Limits); err != nil {
		return GoalSpec{}, err
	}
	return goal, nil
}

func decodeGoalPort(data []byte, path string) (GoalPort, error) {
	fields, err := goalJSONObject(data, path, []string{"id", "type", "binding"}, []string{"id", "type"})
	if err != nil {
		return GoalPort{}, err
	}
	var port GoalPort
	port.ID, err = goalJSONString(fields["id"], path+".id", 256, false)
	if err != nil {
		return GoalPort{}, err
	}
	port.Type, err = algebra.DecodeTypeJSON(fields["type"], path+".type")
	if err != nil {
		return GoalPort{}, err
	}
	if raw, ok := fields["binding"]; ok {
		if goalNull(raw) {
			return GoalPort{}, goalInvalid("INVALID_JSON_VALUE", path+".binding", "binding must be omitted or an object, not null")
		}
		bindingFields, objectErr := goalJSONObject(raw, path+".binding", []string{"kind", "id", "type"}, []string{"kind", "id", "type"})
		if objectErr != nil {
			return GoalPort{}, objectErr
		}
		binding := &GoalValue{}
		kind, stringErr := goalJSONString(bindingFields["kind"], path+".binding.kind", 32, false)
		if stringErr != nil {
			return GoalPort{}, stringErr
		}
		binding.Kind = algebra.TermKind(kind)
		if binding.Kind != algebra.TermRef && binding.Kind != algebra.TermHole {
			return GoalPort{}, goalInvalid("INVALID_GOAL_BINDING_KIND", path+".binding.kind", "binding kind must be Ref or Hole")
		}
		binding.ID, err = goalJSONString(bindingFields["id"], path+".binding.id", 256, false)
		if err != nil {
			return GoalPort{}, err
		}
		binding.Type, err = decodeGoalValueType(bindingFields["type"], path+".binding.type")
		if err != nil {
			return GoalPort{}, err
		}
		port.Binding = binding
	}
	return port, nil
}

// decodeGoalValueType keeps the GoalValue pointer shape while using the shared
// algebra type decoder for its strict tagged union.
func decodeGoalValueType(data []byte, path string) (*algebra.Type, error) {
	typ, err := algebra.DecodeTypeJSON(data, path)
	if err != nil {
		return nil, err
	}
	return &typ, nil
}

func decodeGoalReference(data []byte, path string) (algebra.Declaration, error) {
	fields, err := goalJSONObject(data, path, []string{"id", "type"}, []string{"id", "type"})
	if err != nil {
		return algebra.Declaration{}, err
	}
	id, err := goalJSONString(fields["id"], path+".id", 256, false)
	if err != nil {
		return algebra.Declaration{}, err
	}
	typ, err := algebra.DecodeTypeJSON(fields["type"], path+".type")
	if err != nil {
		return algebra.Declaration{}, err
	}
	return algebra.Declaration{ID: id, Type: typ}, nil
}

func decodeGoalConstraint(data []byte, path string) (Constraint, error) {
	base, err := goalJSONObject(data, path, []string{"id", "kind", "symbol", "from_port", "to_port", "from_role", "to_role", "allowed_atom_counts", "port"}, []string{"id", "kind"})
	if err != nil {
		return Constraint{}, err
	}
	id, err := goalJSONString(base["id"], path+".id", 256, false)
	if err != nil {
		return Constraint{}, err
	}
	kindText, err := goalJSONString(base["kind"], path+".kind", 32, false)
	if err != nil {
		return Constraint{}, err
	}
	constraint := Constraint{ID: id, Kind: ConstraintKind(kindText)}
	var allowed, required []string
	switch constraint.Kind {
	case ConstraintPath:
		allowed = []string{"id", "kind", "symbol", "from_port", "to_port", "from_role", "to_role", "allowed_atom_counts"}
		required = []string{"id", "kind", "from_port", "to_port"}
	case ConstraintUnary:
		allowed = []string{"id", "kind", "symbol", "port"}
		required = []string{"id", "kind"}
	default:
		return Constraint{}, goalUnsupported("UNSUPPORTED_CONSTRAINT_KIND", path+".kind", "constraint kind is outside the bounded goal profile")
	}
	fields, err := goalJSONObject(data, path, allowed, required)
	if err != nil {
		return Constraint{}, err
	}
	readOptional := func(name string, max int) (string, error) {
		raw, ok := fields[name]
		if !ok {
			return "", nil
		}
		return goalJSONString(raw, path+"."+name, max, true)
	}
	constraint.Symbol, err = readOptional("symbol", 256)
	if err != nil {
		return Constraint{}, err
	}
	if constraint.Kind == ConstraintPath {
		constraint.FromPort, err = goalJSONString(fields["from_port"], path+".from_port", 256, false)
		if err != nil {
			return Constraint{}, err
		}
		constraint.ToPort, err = goalJSONString(fields["to_port"], path+".to_port", 256, false)
		if err != nil {
			return Constraint{}, err
		}
		constraint.FromRole, err = readOptional("from_role", 256)
		if err != nil {
			return Constraint{}, err
		}
		constraint.ToRole, err = readOptional("to_role", 256)
		if err != nil {
			return Constraint{}, err
		}
		if raw, ok := fields["allowed_atom_counts"]; ok {
			values, arrayErr := goalJSONArray(raw, path+".allowed_atom_counts", MaxGoalAtoms)
			if arrayErr != nil {
				return Constraint{}, arrayErr
			}
			constraint.AllowedAtomCounts = make([]int, 0, len(values))
			for i, value := range values {
				count, intErr := goalJSONInt(value, fmt.Sprintf("%s.allowed_atom_counts[%d]", path, i), -MaxGoalAtoms, MaxGoalAtoms)
				if intErr != nil {
					return Constraint{}, intErr
				}
				constraint.AllowedAtomCounts = append(constraint.AllowedAtomCounts, count)
			}
		}
	} else {
		constraint.Port, err = readOptional("port", 256)
		if err != nil {
			return Constraint{}, err
		}
	}
	return constraint, nil
}

func decodeGoalSource(data []byte, path string) (provenance.SourceRef, error) {
	fields, err := goalJSONObject(data, path, []string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}, []string{"id", "version", "license"})
	if err != nil {
		return provenance.SourceRef{}, err
	}
	var source provenance.SourceRef
	if source.ID, err = goalJSONString(fields["id"], path+".id", provenance.MaxSourceIDBytes, false); err != nil {
		return provenance.SourceRef{}, err
	}
	if source.Version, err = goalJSONString(fields["version"], path+".version", provenance.MaxSourceVersionBytes, false); err != nil {
		return provenance.SourceRef{}, err
	}
	if source.License, err = goalJSONString(fields["license"], path+".license", provenance.MaxSourceLicenseBytes, false); err != nil {
		return provenance.SourceRef{}, err
	}
	if source.Path, err = goalOptionalSourceString(fields, "path", path, provenance.MaxSourceLocationBytes); err != nil {
		return provenance.SourceRef{}, err
	}
	if source.URI, err = goalOptionalSourceString(fields, "uri", path, provenance.MaxSourceLocationBytes); err != nil {
		return provenance.SourceRef{}, err
	}
	if source.Author, err = goalOptionalSourceString(fields, "author", path, provenance.MaxSourceAuthorBytes); err != nil {
		return provenance.SourceRef{}, err
	}
	if source.ChecksumSHA256, err = goalOptionalSourceString(fields, "checksum_sha256", path, provenance.SHA256HexBytes); err != nil {
		return provenance.SourceRef{}, err
	}
	return source, nil
}

func goalOptionalSourceString(fields map[string]json.RawMessage, name, path string, max int) (string, error) {
	if raw, ok := fields[name]; ok {
		return goalJSONString(raw, path+"."+name, max, true)
	}
	return "", nil
}

func decodeGoalLimits(data []byte) (GoalLimits, error) {
	fields, err := goalJSONObject(data, "limits", []string{"max_atoms", "max_steps", "term_limits"}, nil)
	if err != nil {
		return GoalLimits{}, err
	}
	var limits GoalLimits
	if raw, ok := fields["max_atoms"]; ok {
		limits.MaxAtoms, err = goalJSONInt(raw, "limits.max_atoms", -1, MaxGoalAtoms+1)
		if err != nil {
			return GoalLimits{}, err
		}
	}
	if raw, ok := fields["max_steps"]; ok {
		limits.MaxSteps, err = goalJSONInt(raw, "limits.max_steps", -1, MaxGoalSteps+1)
		if err != nil {
			return GoalLimits{}, err
		}
	}
	if raw, ok := fields["term_limits"]; ok {
		if goalNull(raw) {
			return GoalLimits{}, goalInvalid("INVALID_JSON_VALUE", "limits.term_limits", "term_limits must be an object, not null")
		}
		termFields, objectErr := goalJSONObject(raw, "limits.term_limits", []string{"max_nodes", "max_depth", "max_type_depth", "max_symbols", "max_ports"}, nil)
		if objectErr != nil {
			return GoalLimits{}, objectErr
		}
		values := []struct {
			name string
			dest *int
			cap  int
		}{
			{"max_nodes", &limits.TermLimits.MaxNodes, algebra.DefaultMaxNodes},
			{"max_depth", &limits.TermLimits.MaxDepth, algebra.DefaultMaxDepth},
			{"max_type_depth", &limits.TermLimits.MaxTypeDepth, algebra.DefaultMaxTypeDepth},
			{"max_symbols", &limits.TermLimits.MaxSymbols, algebra.DefaultMaxSymbols},
			{"max_ports", &limits.TermLimits.MaxPorts, algebra.DefaultMaxPorts},
		}
		for _, item := range values {
			if rawValue, exists := termFields[item.name]; exists {
				*item.dest, err = goalJSONInt(rawValue, "limits.term_limits."+item.name, -1, item.cap+1)
				if err != nil {
					return GoalLimits{}, err
				}
			}
		}
	}
	return limits, nil
}

func goalJSONObject(data []byte, path string, allowed, required []string) (map[string]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, goalInvalid("INVALID_JSON_OBJECT", path, "expected a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	fields := make(map[string]json.RawMessage, len(allowed))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, goalInvalid("INVALID_JSON_OBJECT", path, "expected a JSON object")
	}
	for decoder.More() {
		keyToken, keyErr := decoder.Token()
		name, isString := keyToken.(string)
		if keyErr != nil || !isString {
			return nil, goalInvalid("INVALID_JSON_OBJECT", path, "object field name must be a string")
		}
		if _, ok := allowedSet[name]; !ok {
			return nil, goalInvalid("UNKNOWN_JSON_FIELD", path+"."+name, "field is not allowed in this goal object")
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, goalInvalid("INVALID_JSON_VALUE", path+"."+name, "field value is malformed")
		}
		fields[name] = raw
	}
	closeToken, err := decoder.Token()
	if err != nil || closeToken != json.Delim('}') {
		return nil, goalInvalid("INVALID_JSON_OBJECT", path, "object is unterminated")
	}
	var trailing json.RawMessage
	if decoder.Decode(&trailing) != io.EOF {
		return nil, goalInvalid("TRAILING_DOCUMENT", path, "expected exactly one JSON value")
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return nil, goalInvalid("MISSING_JSON_FIELD", path+"."+name, "required field is missing")
		}
	}
	return fields, nil
}

func goalJSONArray(data []byte, path string, maximum int) ([]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil, goalInvalid("INVALID_JSON_ARRAY", path, "expected a JSON array")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('[') {
		return nil, goalInvalid("INVALID_JSON_ARRAY", path, "expected a non-null JSON array")
	}
	values := make([]json.RawMessage, 0, maximum)
	for decoder.More() {
		if len(values) >= maximum {
			return nil, goalResource(path, fmt.Sprintf("array count exceeds limit %d", maximum))
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, goalInvalid("INVALID_JSON_ARRAY", path, "array value is malformed")
		}
		values = append(values, raw)
	}
	closeToken, err := decoder.Token()
	if err != nil || closeToken != json.Delim(']') {
		return nil, goalInvalid("INVALID_JSON_ARRAY", path, "array is unterminated")
	}
	var trailing json.RawMessage
	if decoder.Decode(&trailing) != io.EOF {
		return nil, goalInvalid("TRAILING_DOCUMENT", path, "expected exactly one JSON value")
	}
	return values, nil
}

func goalJSONString(data []byte, path string, maximum int, allowEmpty bool) (string, error) {
	var value string
	if goalNull(data) || json.Unmarshal(data, &value) != nil {
		return "", goalInvalid("INVALID_JSON_STRING", path, "expected a non-null JSON string")
	}
	if len(value) > maximum || !utf8.ValidString(value) || (!allowEmpty && value == "") {
		return "", goalInvalid("INVALID_JSON_STRING", path, "string is empty, invalid, or exceeds its byte limit")
	}
	return value, nil
}

func goalJSONInt(data []byte, path string, minimum, maximum int) (int, error) {
	var value json.Number
	if goalNull(data) || json.Unmarshal(data, &value) != nil {
		return 0, goalInvalid("INVALID_JSON_INTEGER", path, "expected a non-null JSON integer")
	}
	parsed, err := strconv.ParseInt(value.String(), 10, 64)
	if err != nil {
		return 0, goalInvalid("INVALID_JSON_INTEGER", path, "expected a bounded base-10 integer")
	}
	if parsed > int64(maximum) {
		return 0, goalResource(path, "integer exceeds the bounded goal limit")
	}
	if parsed < int64(minimum) {
		return 0, goalInvalid("INVALID_JSON_INTEGER", path, "integer is below the accepted lower bound")
	}
	return int(parsed), nil
}

func goalNull(data []byte) bool { return bytes.Equal(bytes.TrimSpace(data), []byte("null")) }
