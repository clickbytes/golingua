package lexicon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/synthesis"
)

const (
	SemanticTargetSchemaVersion = 1
	SemanticExecutionProfile    = synthesis.GoalProfile
	MaxSemanticRolePorts        = synthesis.MaxGoalPorts
	MaxSemanticTargetBytes      = synthesis.MaxGoalInputBytes
)

type SemanticTargetKind string

const (
	SemanticTargetSymbol SemanticTargetKind = "symbol"
	SemanticTargetGoal   SemanticTargetKind = "goal"
)

// SymbolRef is an opaque selected-registry symbol ID. It carries no behavior.
type SymbolRef string

// SemanticRolePort explicitly connects one sense-specific frame role to a
// typed target port. Frame IDs and roles are opaque here; the valency package
// validates their declarations without introducing a package cycle.
type SemanticRolePort struct {
	FrameID string       `json:"frame_id"`
	Role    string       `json:"role"`
	PortID  string       `json:"port_id"`
	Type    algebra.Type `json:"type"`
}

// SemanticTarget is a closed tagged data reference. Exactly one of SymbolRef
// and GoalSpec is present according to Kind. Profile is an explicit bounded
// execution profile, not a callback or a fallback selector.
type SemanticTarget struct {
	SchemaVersion int                 `json:"schema_version"`
	Kind          SemanticTargetKind  `json:"kind"`
	Profile       string              `json:"profile"`
	SymbolRef     SymbolRef           `json:"symbol_ref,omitempty"`
	GoalSpec      *synthesis.GoalSpec `json:"goal_spec,omitempty"`
	RolePorts     []SemanticRolePort  `json:"role_ports"`
}

func (target *SemanticTarget) UnmarshalJSON(data []byte) error {
	if err := algebra.PreflightJSONLimit(data, MaxSemanticTargetBytes); err != nil {
		return fmt.Errorf("semantic target: %w", err)
	}
	fields, err := strictTargetObject(data, "target", []string{"schema_version", "kind", "profile", "symbol_ref", "goal_spec", "role_ports"}, []string{"schema_version", "kind", "profile", "role_ports"})
	if err != nil {
		return err
	}
	var out SemanticTarget
	if err := json.Unmarshal(fields["schema_version"], &out.SchemaVersion); err != nil {
		return targetFormatError("target.schema_version", err)
	}
	if err := json.Unmarshal(fields["kind"], &out.Kind); err != nil {
		return targetFormatError("target.kind", err)
	}
	if err := json.Unmarshal(fields["profile"], &out.Profile); err != nil {
		return targetFormatError("target.profile", err)
	}
	if raw, ok := fields["symbol_ref"]; ok {
		if isJSONNull(raw) {
			return targetError("INVALID_TARGET_FIELD", "target.symbol_ref", "symbol_ref cannot be null")
		}
		var symbol string
		if err := json.Unmarshal(raw, &symbol); err != nil {
			return targetFormatError("target.symbol_ref", err)
		}
		out.SymbolRef = SymbolRef(symbol)
	}
	if raw, ok := fields["goal_spec"]; ok {
		if isJSONNull(raw) {
			return targetError("INVALID_TARGET_FIELD", "target.goal_spec", "goal_spec cannot be null")
		}
		goal, err := synthesis.DecodeGoal(raw)
		if err != nil {
			return targetFormatError("target.goal_spec", err)
		}
		out.GoalSpec = &goal
	}
	roleRows, err := strictTargetArray(fields["role_ports"], "target.role_ports", MaxSemanticRolePorts)
	if err != nil {
		return err
	}
	out.RolePorts = make([]SemanticRolePort, 0, len(roleRows))
	for i, raw := range roleRows {
		path := fmt.Sprintf("target.role_ports[%d]", i)
		rowFields, err := strictTargetObject(raw, path, []string{"frame_id", "role", "port_id", "type"}, []string{"frame_id", "role", "port_id", "type"})
		if err != nil {
			return err
		}
		var row SemanticRolePort
		for key, dest := range map[string]*string{"frame_id": &row.FrameID, "role": &row.Role, "port_id": &row.PortID} {
			if err := json.Unmarshal(rowFields[key], dest); err != nil {
				return targetFormatError(path+"."+key, err)
			}
		}
		row.Type, err = algebra.DecodeTypeJSON(rowFields["type"], path+".type")
		if err != nil {
			return err
		}
		out.RolePorts = append(out.RolePorts, row)
	}
	if out.SchemaVersion != SemanticTargetSchemaVersion {
		return targetError("UNSUPPORTED_TARGET_VERSION", "target.schema_version", "semantic target version is unsupported")
	}
	switch out.Kind {
	case SemanticTargetSymbol:
		if _, ok := fields["symbol_ref"]; !ok || fields["goal_spec"] != nil || !validTargetText(string(out.SymbolRef), 256) {
			return targetError("INVALID_TARGET_VARIANT", "target", "symbol target requires a bounded symbol_ref and forbids goal_spec")
		}
	case SemanticTargetGoal:
		if fields["symbol_ref"] != nil || out.GoalSpec == nil {
			return targetError("INVALID_TARGET_VARIANT", "target", "goal target requires goal_spec and forbids symbol_ref")
		}
	default:
		return targetError("UNSUPPORTED_TARGET_KIND", "target.kind", "semantic target kind is unsupported")
	}
	if !validTargetText(out.Profile, 256) || out.Profile != SemanticExecutionProfile {
		return targetError("UNSUPPORTED_TARGET_PROFILE", "target.profile", "target requires an explicitly supported execution profile")
	}
	*target = out
	return nil
}

func strictTargetObject(data []byte, path string, allowed, required []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, targetError("INVALID_TARGET_OBJECT", path, "expected a JSON object")
	}
	allow := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		allow[key] = true
	}
	for key := range fields {
		if !allow[key] {
			return nil, targetError("UNKNOWN_TARGET_FIELD", path+"."+key, "unknown or mis-cased target field")
		}
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return nil, targetError("MISSING_TARGET_FIELD", path+"."+key, "required target field is missing")
		}
	}
	return fields, nil
}

func strictTargetArray(raw json.RawMessage, path string, maximum int) ([]json.RawMessage, error) {
	if len(raw) == 0 || isJSONNull(raw) {
		return nil, targetError("INVALID_TARGET_ARRAY", path, "required field must be a non-null array")
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil || rows == nil {
		return nil, targetError("INVALID_TARGET_ARRAY", path, "required field must be a non-null array")
	}
	if len(rows) > maximum {
		return nil, targetResource(path, len(rows), maximum)
	}
	return rows, nil
}

func isJSONNull(raw []byte) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }

func (target SemanticTarget) ExecutionAuthorized(mapping MappingKind, status BindingStatus, provenanceValid bool) bool {
	if mapping != MappingEquivalent || status != BindingApproved || !provenanceValid || target.SchemaVersion != SemanticTargetSchemaVersion || target.Profile != SemanticExecutionProfile || len(target.RolePorts) == 0 {
		return false
	}
	switch target.Kind {
	case SemanticTargetSymbol:
		return target.SymbolRef != "" && target.GoalSpec == nil
	case SemanticTargetGoal:
		return target.SymbolRef == "" && target.GoalSpec != nil
	default:
		return false
	}
}

func cloneSemanticTarget(source *SemanticTarget) *SemanticTarget {
	if source == nil {
		return nil
	}
	result := *source
	result.RolePorts = append([]SemanticRolePort(nil), source.RolePorts...)
	for i := range result.RolePorts {
		result.RolePorts[i].Type = cloneAlgebraType(result.RolePorts[i].Type)
	}
	sort.Slice(result.RolePorts, func(i, j int) bool {
		left, right := result.RolePorts[i], result.RolePorts[j]
		if left.FrameID != right.FrameID {
			return left.FrameID < right.FrameID
		}
		if left.Role != right.Role {
			return left.Role < right.Role
		}
		return left.PortID < right.PortID
	})
	if source.GoalSpec != nil {
		goal := cloneGoalSpec(*source.GoalSpec)
		result.GoalSpec = &goal
	}
	return &result
}

func canonicalSemanticTarget(source *SemanticTarget) *SemanticTarget {
	return cloneSemanticTarget(source)
}

func cloneGoalSpec(source synthesis.GoalSpec) synthesis.GoalSpec {
	result := source
	result.ExpectedType = cloneAlgebraType(source.ExpectedType)
	result.Ports = append([]synthesis.GoalPort(nil), source.Ports...)
	for i := range result.Ports {
		result.Ports[i].Type = cloneAlgebraType(source.Ports[i].Type)
		if source.Ports[i].Binding != nil {
			binding := *source.Ports[i].Binding
			if binding.Type != nil {
				cloned := cloneAlgebraType(*binding.Type)
				binding.Type = &cloned
			}
			result.Ports[i].Binding = &binding
		}
	}
	result.References = append([]algebra.Declaration(nil), source.References...)
	for i := range result.References {
		result.References[i].Type = cloneAlgebraType(source.References[i].Type)
	}
	result.ProjectPorts = append([]string(nil), source.ProjectPorts...)
	result.AllowedOperators = append([]string(nil), source.AllowedOperators...)
	result.Constraints = append([]synthesis.Constraint(nil), source.Constraints...)
	for i := range result.Constraints {
		result.Constraints[i].AllowedAtomCounts = append([]int(nil), source.Constraints[i].AllowedAtomCounts...)
	}
	result.Provenance = append([]provenance.SourceRef(nil), source.Provenance...)
	if result.Ports == nil {
		result.Ports = []synthesis.GoalPort{}
	}
	if result.References == nil {
		result.References = []algebra.Declaration{}
	}
	if result.AllowedOperators == nil {
		result.AllowedOperators = []string{}
	}
	if result.Constraints == nil {
		result.Constraints = []synthesis.Constraint{}
	}
	if result.Provenance == nil {
		result.Provenance = []provenance.SourceRef{}
	}
	if source.RequestContext != nil {
		if contextCopy, err := algebra.CloneRequestContext(*source.RequestContext); err == nil {
			result.RequestContext = &contextCopy
		} else {
			result.RequestContext = nil
		}
	}
	return result
}

func cloneAlgebraType(source algebra.Type) algebra.Type {
	result := source
	if source.From != nil {
		from := cloneAlgebraType(*source.From)
		result.From = &from
	}
	if source.To != nil {
		to := cloneAlgebraType(*source.To)
		result.To = &to
	}
	return result
}

func validateSemanticTarget(target *SemanticTarget, sense LexicalSense, path string) error {
	if target == nil || target.SchemaVersion != SemanticTargetSchemaVersion {
		return targetError("UNSUPPORTED_TARGET_VERSION", path+".schema_version", "semantic target version is unsupported")
	}
	if target.Profile != SemanticExecutionProfile {
		return targetError("UNSUPPORTED_TARGET_PROFILE", path+".profile", "target requires an explicitly supported execution profile")
	}
	if len(target.RolePorts) == 0 {
		return targetError("MISSING_ROLE_PORT_MAPPING", path+".role_ports", "semantic target requires an explicit frame-role-to-port mapping")
	}
	if len(target.RolePorts) > MaxSemanticRolePorts {
		return targetResource(path+".role_ports", len(target.RolePorts), MaxSemanticRolePorts)
	}
	frames := make(map[string]struct{}, len(sense.ValencyIDs))
	for _, frameID := range sense.ValencyIDs {
		frames[string(frameID)] = struct{}{}
	}
	seenRoles := make(map[string]struct{}, len(target.RolePorts))
	seenPorts := make(map[string]struct{}, len(target.RolePorts))
	types := make([]algebra.Type, 0, len(target.RolePorts))
	for i, mapping := range target.RolePorts {
		mapPath := fmt.Sprintf("%s.role_ports[%d]", path, i)
		if !validTargetText(mapping.FrameID, 256) {
			return targetError("INVALID_FRAME_ID", mapPath+".frame_id", "frame ID is malformed")
		}
		if _, ok := frames[mapping.FrameID]; !ok {
			return unknown("UNKNOWN_VALENCY_FRAME", mapPath+".frame_id", "frame ID is not declared by the linked sense")
		}
		if !validLabel(mapping.Role, 256) {
			return targetError("INVALID_SEMANTIC_ROLE", mapPath+".role", "semantic role is malformed")
		}
		if !validTargetID(mapping.PortID) {
			return targetError("INVALID_TARGET_PORT_ID", mapPath+".port_id", "target port ID is malformed")
		}
		roleKey := mapping.FrameID + "\x00" + mapping.Role
		if _, ok := seenRoles[roleKey]; ok {
			return targetError("DUPLICATE_FRAME_ROLE_MAPPING", mapPath, "frame-role mapping is duplicated")
		}
		seenRoles[roleKey] = struct{}{}
		portKey := mapping.FrameID + "\x00" + mapping.PortID
		if _, ok := seenPorts[portKey]; ok {
			return targetError("DUPLICATE_FRAME_PORT_MAPPING", mapPath+".port_id", "one frame cannot map multiple roles to the same target port")
		}
		seenPorts[portKey] = struct{}{}
		if mapping.Type.Kind == algebra.TypeArrow {
			return targetError("INVALID_ROLE_PORT_TYPE", mapPath+".type", "semantic role mapping requires a data sort type")
		}
		types = append(types, mapping.Type)
	}
	if err := validateTargetTypes(types, nil, nil, path+".role_ports"); err != nil {
		return err
	}
	switch target.Kind {
	case SemanticTargetSymbol:
		if !validTargetID(string(target.SymbolRef)) || target.GoalSpec != nil {
			return targetError("INVALID_SYMBOL_TARGET", path+".symbol_ref", "symbol target requires exactly one bounded symbol reference")
		}
	case SemanticTargetGoal:
		if target.SymbolRef != "" || target.GoalSpec == nil {
			return targetError("INVALID_GOAL_TARGET", path, "goal target requires exactly one goal specification")
		}
		if err := validateGoalTemplateShape(target.GoalSpec, path+".goal_spec"); err != nil {
			return err
		}
		goalPorts := make(map[string]synthesis.GoalPort, len(target.GoalSpec.Ports))
		for _, port := range target.GoalSpec.Ports {
			goalPorts[port.ID] = port
		}
		for i, mapping := range target.RolePorts {
			port, ok := goalPorts[mapping.PortID]
			if !ok {
				return unknown("UNKNOWN_GOAL_PORT", fmt.Sprintf("%s.role_ports[%d].port_id", path, i), "mapped target port is absent from the goal template")
			}
			if !algebra.TypeEqual(mapping.Type, port.Type) {
				return targetError("ROLE_PORT_TYPE_MISMATCH", fmt.Sprintf("%s.role_ports[%d].type", path, i), "frame role mapping type must exactly match its goal port type")
			}
		}
	default:
		return targetError("UNSUPPORTED_TARGET_KIND", path+".kind", "semantic target kind is unsupported")
	}
	return nil
}

func validateGoalTemplateShape(goal *synthesis.GoalSpec, path string) error {
	if goal == nil || goal.SchemaVersion != synthesis.GoalSchemaVersion {
		return targetError("INVALID_GOAL_SCHEMA_VERSION", path+".schema_version", "goal template schema version is invalid")
	}
	if goal.Profile != synthesis.GoalProfile || !validTargetID(goal.ContextID) {
		return targetError("INVALID_GOAL_PROFILE_OR_CONTEXT", path, "goal template requires the supported profile and a bounded context ID")
	}
	if goal.Ports == nil || goal.References == nil || goal.AllowedOperators == nil || goal.Constraints == nil || goal.Provenance == nil {
		return targetError("MISSING_GOAL_ARRAY", path, "goal template arrays must be present, including empty arrays")
	}
	if len(goal.Ports) == 0 || len(goal.Ports) > synthesis.MaxGoalPorts {
		return targetResource(path+".ports", len(goal.Ports), synthesis.MaxGoalPorts)
	}
	if len(goal.References) > algebra.DefaultMaxSymbols {
		return targetResource(path+".references", len(goal.References), algebra.DefaultMaxSymbols)
	}
	if len(goal.Constraints) > synthesis.MaxGoalConstraints {
		return targetResource(path+".constraints", len(goal.Constraints), synthesis.MaxGoalConstraints)
	}
	if len(goal.Provenance) > synthesis.MaxGoalSources {
		return targetResource(path+".provenance", len(goal.Provenance), synthesis.MaxGoalSources)
	}
	if len(goal.AllowedOperators) > 6 {
		return targetResource(path+".allowed_operators", len(goal.AllowedOperators), 6)
	}
	if len(goal.ProjectPorts) > synthesis.MaxGoalPorts {
		return targetResource(path+".project_ports", len(goal.ProjectPorts), synthesis.MaxGoalPorts)
	}
	if _, err := synthesis.NormalizeGoalLimits(goal.Limits); err != nil {
		return err
	}
	types := []algebra.Type{goal.ExpectedType}
	goalPortIDs := make(map[string]struct{}, len(goal.Ports))
	for i, port := range goal.Ports {
		if !validTargetID(port.ID) {
			return targetError("INVALID_GOAL_PORT_ID", fmt.Sprintf("%s.ports[%d].id", path, i), "goal port ID is malformed")
		}
		if _, exists := goalPortIDs[port.ID]; exists {
			return targetError("DUPLICATE_GOAL_PORT_ID", fmt.Sprintf("%s.ports[%d].id", path, i), "goal port ID is duplicated")
		}
		goalPortIDs[port.ID] = struct{}{}
		if port.Type.Kind == algebra.TypeArrow {
			return targetError("INVALID_GOAL_PORT_TYPE", fmt.Sprintf("%s.ports[%d].type", path, i), "goal output ports require a base data sort")
		}
		types = append(types, port.Type)
		if port.Binding != nil {
			if port.Binding.Type == nil {
				return targetError("MISSING_GOAL_BINDING_TYPE", fmt.Sprintf("%s.ports[%d].binding.type", path, i), "goal binding type is required")
			}
			types = append(types, *port.Binding.Type)
		}
	}
	for i, ref := range goal.References {
		if !validTargetID(ref.ID) {
			return targetError("INVALID_GOAL_REFERENCE_ID", fmt.Sprintf("%s.references[%d].id", path, i), "goal reference ID is malformed")
		}
		types = append(types, ref.Type)
	}
	if goal.RequestContext != nil {
		if len(goal.RequestContext.Referents) > algebra.MaxContextReferents {
			return targetResource(path+".request_context.referents", len(goal.RequestContext.Referents), algebra.MaxContextReferents)
		}
		for _, ref := range goal.RequestContext.Referents {
			types = append(types, ref.Type)
		}
	}
	if err := validateTargetTypes(types, goal.References, goal.RequestContext, path); err != nil {
		return err
	}
	projected := make(map[string]struct{}, len(goal.ProjectPorts))
	for i, id := range goal.ProjectPorts {
		if !validTargetID(id) {
			return targetError("INVALID_PROJECT_PORT", fmt.Sprintf("%s.project_ports[%d]", path, i), "projected port ID is malformed")
		}
		if _, exists := goalPortIDs[id]; !exists {
			return unknown("UNKNOWN_PROJECT_PORT", fmt.Sprintf("%s.project_ports[%d]", path, i), "projected port is not declared by the goal")
		}
		if _, exists := projected[id]; exists {
			return targetError("DUPLICATE_PROJECT_PORT", fmt.Sprintf("%s.project_ports[%d]", path, i), "projected port is duplicated")
		}
		projected[id] = struct{}{}
		for _, port := range goal.Ports {
			if port.ID == id && port.Binding != nil {
				return targetError("BOUND_PROJECT_PORT", fmt.Sprintf("%s.project_ports[%d]", path, i), "a bound goal port cannot also be projected")
			}
		}
	}
	openTypes := make([]algebra.Type, 0, len(goal.Ports))
	for _, port := range goal.Ports {
		if _, isProjected := projected[port.ID]; isProjected || port.Binding != nil {
			continue
		}
		openTypes = append(openTypes, port.Type)
	}
	derivedType := algebra.Type{Kind: algebra.TypeProp}
	for i := len(openTypes) - 1; i >= 0; i-- {
		argument := openTypes[i]
		body := derivedType
		derivedType = algebra.Type{Kind: algebra.TypeArrow, From: &argument, To: &body}
	}
	if !algebra.TypeEqual(derivedType, goal.ExpectedType) {
		return targetError("GOAL_EXPECTED_TYPE_MISMATCH", path+".expected_type", "goal type must match its ordered open ports")
	}
	for i, source := range goal.Provenance {
		if err := source.Validate(); err != nil {
			return targetError("INVALID_GOAL_PROVENANCE", fmt.Sprintf("%s.provenance[%d]", path, i), "goal template provenance is malformed")
		}
	}
	seenOperators := make(map[string]struct{}, len(goal.AllowedOperators))
	for i, operator := range goal.AllowedOperators {
		if !validTargetText(operator, 256) {
			return targetError("INVALID_GOAL_OPERATOR", fmt.Sprintf("%s.allowed_operators[%d]", path, i), "goal operator is malformed")
		}
		if _, exists := seenOperators[operator]; exists {
			return targetError("DUPLICATE_GOAL_OPERATOR", fmt.Sprintf("%s.allowed_operators[%d]", path, i), "goal operator is duplicated")
		}
		seenOperators[operator] = struct{}{}
	}
	encoded, err := json.Marshal(goal)
	if err != nil {
		return targetError("INVALID_GOAL_TEMPLATE", path, "goal template cannot be encoded safely")
	}
	if len(encoded) > synthesis.MaxGoalInputBytes {
		return targetResource(path, len(encoded), synthesis.MaxGoalInputBytes)
	}
	if _, err := synthesis.DecodeGoal(encoded); err != nil {
		return targetFormatError(path, err)
	}
	return nil
}

func validateTargetTypes(types []algebra.Type, references []algebra.Declaration, requestContext *algebra.RequestContext, path string) error {
	sortSet := make(map[string]struct{})
	for i, typ := range types {
		if err := collectTargetSorts(typ, 1, make(map[*algebra.Type]bool), sortSet); err != nil {
			return targetFormatError(fmt.Sprintf("%s.types[%d]", path, i), err)
		}
	}
	for i, ref := range references {
		if err := collectTargetSorts(ref.Type, 1, make(map[*algebra.Type]bool), sortSet); err != nil {
			return targetFormatError(fmt.Sprintf("%s.references[%d].type", path, i), err)
		}
	}
	if requestContext != nil {
		for i, ref := range requestContext.Referents {
			if err := collectTargetSorts(ref.Type, 1, make(map[*algebra.Type]bool), sortSet); err != nil {
				return targetFormatError(fmt.Sprintf("%s.request_context.referents[%d].type", path, i), err)
			}
		}
	}
	sorts := make([]algebra.SortDeclaration, 0, len(sortSet))
	for id := range sortSet {
		sorts = append(sorts, algebra.SortDeclaration{ID: id})
	}
	sort.Slice(sorts, func(i, j int) bool { return sorts[i].ID < sorts[j].ID })
	doc := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: sorts, References: references, RequestContext: requestContext}
	if err := algebra.CheckDeclarations(doc); err != nil {
		return targetFormatError(path, err)
	}
	return nil
}

func collectTargetSorts(value algebra.Type, depth int, active map[*algebra.Type]bool, sorts map[string]struct{}) error {
	if depth > algebra.DefaultMaxTypeDepth {
		return fmt.Errorf("type depth exceeds %d", algebra.DefaultMaxTypeDepth)
	}
	switch value.Kind {
	case algebra.TypeEntity, algebra.TypeEvent, algebra.TypeInterval, algebra.TypeProp, algebra.TypeQuestion:
		if value.SortID != "" || value.From != nil || value.To != nil {
			return fmt.Errorf("built-in type has irrelevant fields")
		}
	case algebra.TypeNominal:
		if !validTargetText(value.SortID, 256) || value.From != nil || value.To != nil {
			return fmt.Errorf("nominal type requires only a bounded sort_id")
		}
		sorts[value.SortID] = struct{}{}
	case algebra.TypeArrow:
		if value.SortID != "" || value.From == nil || value.To == nil {
			return fmt.Errorf("arrow type requires from and to only")
		}
		for _, child := range []*algebra.Type{value.From, value.To} {
			if active[child] {
				return fmt.Errorf("cyclic type graph")
			}
			active[child] = true
			if err := collectTargetSorts(*child, depth+1, active, sorts); err != nil {
				return err
			}
			delete(active, child)
		}
	default:
		return fmt.Errorf("unsupported type kind %q", value.Kind)
	}
	return nil
}

func validTargetText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\x00\r\n\t")
}

func validTargetID(value string) bool {
	if !validTargetText(value, 256) {
		return false
	}
	for _, r := range value {
		if r == ' ' {
			return false
		}
	}
	return true
}

func targetError(code, path, message string) error {
	return invalid(code, path, message)
}

func targetResource(path string, count, maximum int) error {
	return resourceLimit(path, count, maximum)
}

func targetFormatError(path string, err error) error {
	return fmt.Errorf("%s: %w", path, err)
}
