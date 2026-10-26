package language

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/synthesis"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	SemanticLanguagePackageSchemaVersion = 1
	MaxSemanticLanguagePackageBytes      = 8 << 20
	MaxSemanticLanguageFrameCount        = 128
)

var semanticPackageVersionRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

// SemanticLanguagePackage is an immutable, explicitly loaded language-owned
// semantic snapshot. It contains no callbacks or executable Go code.
type SemanticLanguagePackage struct {
	id         string
	version    string
	language   lexicon.LanguageID
	provenance provenance.SourceRef
	lexicon    *lexicon.Index
	valency    *valency.Index
	frames     []valency.ValencyFrame
	frameByID  map[string]valency.ValencyFrame
}

type semanticLanguagePackageWire struct {
	SchemaVersion int                  `json:"schema_version"`
	ID            string               `json:"id"`
	Version       string               `json:"version"`
	Language      string               `json:"language"`
	Provenance    provenance.SourceRef `json:"provenance"`
	Lexicon       json.RawMessage      `json:"lexicon"`
	Frames        []semanticFrameWire  `json:"frames"`
}

type semanticFrameWire struct {
	ID                     string                   `json:"id"`
	SenseID                string                   `json:"sense_id"`
	SemanticRoles          []string                 `json:"semantic_roles"`
	Arguments              []semanticArgumentWire   `json:"arguments"`
	SelectionalConstraints []semanticConstraintWire `json:"selectional_constraints"`
	AllowedFamilyIDs       []string                 `json:"allowed_family_ids"`
	Provenance             provenance.SourceRef     `json:"provenance"`
}

type semanticArgumentWire struct {
	ID                  string `json:"id"`
	Role                string `json:"role"`
	Required            bool   `json:"required"`
	GrammaticalFunction string `json:"grammatical_function,omitempty"`
	Case                string `json:"case,omitempty"`
	Adposition          string `json:"adposition,omitempty"`
}

type semanticConstraintWire struct {
	ID                string   `json:"id"`
	Role              string   `json:"role"`
	Strength          string   `json:"strength"`
	AllowedConceptIDs []string `json:"allowed_concept_ids,omitempty"`
	AllowedKinds      []string `json:"allowed_kinds,omitempty"`
}

// LoadSemanticPackage decodes and builds one caller-selected local package.
// A CORE graph is optional and is consulted only for legacy CORE bindings.
// No filesystem or network access is performed.
func LoadSemanticPackage(data []byte, ontology *core.ConceptGraph) (*SemanticLanguagePackage, error) {
	if err := algebra.PreflightJSONLimit(data, MaxSemanticLanguagePackageBytes); err != nil {
		return nil, err
	}
	root, err := languageStrictObject(data, "package", []string{"schema_version", "id", "version", "language", "provenance", "lexicon", "frames"}, []string{"schema_version", "id", "version", "language", "provenance", "lexicon", "frames"})
	if err != nil {
		return nil, err
	}
	var version int
	if err := languageStrictValue(root["schema_version"], &version); err != nil || version != SemanticLanguagePackageSchemaVersion {
		return nil, invalid("UNSUPPORTED_SEMANTIC_PACKAGE_VERSION", "schema_version", "semantic language package schema version must be 1")
	}
	var packageID, packageVersion, languageTag string
	for key, target := range map[string]*string{"id": &packageID, "version": &packageVersion, "language": &languageTag} {
		if err := languageStrictValue(root[key], target); err != nil {
			return nil, invalid("INVALID_SEMANTIC_PACKAGE_FIELD", key, "field must be a string")
		}
	}
	if !validSemanticPackageText(packageID, 256) || !semanticPackageVersionRE.MatchString(packageVersion) || !validLanguageTag(languageTag) {
		return nil, invalid("INVALID_SEMANTIC_PACKAGE_IDENTITY", "package", "package ID, version, and canonical language tag are required")
	}
	packageSource, err := decodePackageSource(root["provenance"], "provenance")
	if err != nil {
		return nil, err
	}
	if err := packageSource.Validate(); err != nil {
		return nil, invalid("INVALID_SEMANTIC_PACKAGE_PROVENANCE", "provenance", "package provenance is malformed")
	}
	if len(root["lexicon"]) == 0 || bytes.Equal(bytes.TrimSpace(root["lexicon"]), []byte("null")) {
		return nil, invalid("MISSING_SEMANTIC_PACKAGE_LEXICON", "lexicon", "lexicon document is required")
	}
	if len(root["lexicon"]) > lexicon.MaxDocumentBytes {
		return nil, resource("lexicon_bytes", len(root["lexicon"]), lexicon.MaxDocumentBytes)
	}
	lexDoc, err := lexicon.Decode(root["lexicon"], ontology, lexicon.DefaultBudget())
	if err != nil {
		return nil, err
	}
	lexIndex, err := lexicon.Build(lexDoc, ontology, lexicon.DefaultBudget())
	if err != nil {
		return nil, err
	}
	frames, err := decodeSemanticFrames(root["frames"])
	if err != nil {
		return nil, err
	}
	if len(frames) == 0 {
		return nil, invalid("EMPTY_SEMANTIC_FRAMES", "frames", "semantic language package requires at least one authored frame")
	}
	if err := validateSemanticPackageFrames(lexIndex, lexicon.LanguageID(languageTag), frames); err != nil {
		return nil, err
	}
	frameIndex, err := valency.BuildIndex(frames, nil)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]valency.ValencyFrame, len(frames))
	for _, frame := range frameIndex.Frames() {
		byID[string(frame.ID)] = frame
	}
	return &SemanticLanguagePackage{
		id: packageID, version: packageVersion, language: lexicon.LanguageID(languageTag),
		provenance: packageSource, lexicon: lexIndex, valency: frameIndex,
		frames: frameIndex.Frames(), frameByID: byID,
	}, nil
}

// Package identity accessors return scalar values from the immutable snapshot.
func (pack *SemanticLanguagePackage) ID() string {
	if pack == nil {
		return ""
	}
	return pack.id
}
func (pack *SemanticLanguagePackage) Version() string {
	if pack == nil {
		return ""
	}
	return pack.version
}
func (pack *SemanticLanguagePackage) Language() lexicon.LanguageID {
	if pack == nil {
		return ""
	}
	return pack.language
}
func (pack *SemanticLanguagePackage) Provenance() provenance.SourceRef {
	if pack == nil {
		return provenance.SourceRef{}
	}
	return pack.provenance
}
func (pack *SemanticLanguagePackage) LexiconIndex() *lexicon.Index {
	if pack == nil {
		return nil
	}
	return pack.lexicon
}
func (pack *SemanticLanguagePackage) ValencyIndex() *valency.Index {
	if pack == nil {
		return nil
	}
	return pack.valency
}
func (pack *SemanticLanguagePackage) Frames() []valency.ValencyFrame {
	if pack == nil {
		return []valency.ValencyFrame{}
	}
	return pack.valency.Frames()
}

// TargetValidation is a structural/selected-metadata report. Authorized is a
// separate review gate; this report does not establish truth or owner approval.
type TargetValidation struct {
	SenseID    string                     `json:"sense_id"`
	Kind       lexicon.SemanticTargetKind `json:"kind"`
	Status     lexicon.BindingStatus      `json:"status"`
	Mapping    lexicon.MappingKind        `json:"mapping_kind"`
	Authorized bool                       `json:"authorized"`
	Unresolved []synthesis.Unresolved     `json:"unresolved"`
}

// ValidateTargets checks every tagged target, including pending records,
// against the explicitly supplied selected primitive declaration context.
func (pack *SemanticLanguagePackage) ValidateTargets(context algebra.Document) ([]TargetValidation, error) {
	if pack == nil || pack.lexicon == nil {
		return nil, invalid("MISSING_SEMANTIC_PACKAGE", "package", "a built semantic language package is required")
	}
	if err := validateSelectedPrimitiveContext(context); err != nil {
		return nil, err
	}
	bindings := pack.lexicon.Document().Bindings
	rows := make([]TargetValidation, 0)
	for i, binding := range bindings {
		if binding.Target == nil {
			continue
		}
		frameIDs := make([]string, 0)
		seenFrames := make(map[string]bool)
		for _, mapping := range binding.Target.RolePorts {
			if !seenFrames[mapping.FrameID] {
				seenFrames[mapping.FrameID] = true
				frameIDs = append(frameIDs, mapping.FrameID)
			}
		}
		sort.Strings(frameIDs)
		var unresolved []synthesis.Unresolved
		unresolvedKeys := make(map[string]bool)
		for _, frameID := range frameIDs {
			frame, ok := pack.frameByIDForTarget(binding, frameID)
			if !ok {
				return nil, unknown("UNKNOWN_TARGET_FRAME", fmt.Sprintf("bindings[%d].target.role_ports", i), "target frame is absent from the package")
			}
			items, err := validateSemanticTargetEnvironment(context, binding.Target, frame)
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				key := item.Code + "\x00" + item.Path + "\x00" + item.Message
				if !unresolvedKeys[key] {
					unresolvedKeys[key] = true
					unresolved = append(unresolved, item)
				}
			}
		}
		rows = append(rows, TargetValidation{SenseID: binding.SenseID.String(), Kind: binding.Target.Kind, Status: binding.Status, Mapping: binding.Mapping, Authorized: binding.TargetExecutionAuthorized(), Unresolved: unresolved})
	}
	return rows, nil
}

func (pack *SemanticLanguagePackage) frameByIDForTarget(binding lexicon.SemanticBinding, requested string) (valency.ValencyFrame, bool) {
	if requested != "" {
		frame, ok := pack.frameByID[requested]
		return frame, ok && frame.SenseID == binding.SenseID
	}
	for _, mapping := range binding.Target.RolePorts {
		frame, ok := pack.frameByID[mapping.FrameID]
		if ok && frame.SenseID == binding.SenseID {
			return frame, true
		}
	}
	return valency.ValencyFrame{}, false
}

func validateSelectedPrimitiveContext(context algebra.Document) error {
	if context.SchemaVersion != algebra.SchemaVersion || context.Expression != nil || len(context.Parameters) != 0 || len(context.References) != 0 || context.RequestContext != nil {
		return invalid("INVALID_SELECTED_CONTEXT", "context", "selected context must be a declaration-only algebra document")
	}
	return algebra.CheckDeclarations(context)
}

func validateSemanticTargetEnvironment(context algebra.Document, target *lexicon.SemanticTarget, frame valency.ValencyFrame) ([]synthesis.Unresolved, error) {
	if target == nil {
		return nil, invalid("MISSING_SEMANTIC_TARGET", "target", "target is required")
	}
	rolePorts := make(map[string]lexicon.SemanticRolePort, len(target.RolePorts))
	for i, mapping := range target.RolePorts {
		if mapping.FrameID != string(frame.ID) {
			continue
		}
		if !frameHasRole(frame, mapping.Role) {
			return nil, unknown("UNKNOWN_FRAME_ROLE", fmt.Sprintf("target.role_ports[%d].role", i), "mapped role is not declared by the selected frame")
		}
		rolePorts[mapping.PortID] = mapping
	}
	switch target.Kind {
	case lexicon.SemanticTargetSymbol:
		var signature *algebra.Signature
		for i := range context.Signatures {
			if context.Signatures[i].ID == string(target.SymbolRef) {
				signature = &context.Signatures[i]
				break
			}
		}
		if signature == nil {
			return nil, unknown("UNKNOWN_TARGET_SYMBOL", "target.symbol_ref", "symbol is not an approved primitive in the selected context")
		}
		ports := make(map[string]algebra.Type, len(signature.Ports))
		for _, port := range signature.Ports {
			ports[port.Role] = port.Type
		}
		for portID, mapping := range rolePorts {
			typ, ok := ports[portID]
			if !ok {
				return nil, unknown("UNKNOWN_TARGET_PORT", "target.role_ports", "mapped port is absent from the selected primitive")
			}
			if !algebra.TypeEqual(typ, mapping.Type) {
				return nil, invalid("TARGET_PORT_TYPE_MISMATCH", "target.role_ports", "mapped role type differs from the selected primitive port type")
			}
		}
		return []synthesis.Unresolved{}, nil
	case lexicon.SemanticTargetGoal:
		if target.GoalSpec == nil {
			return nil, invalid("MISSING_GOAL_TEMPLATE", "target.goal_spec", "goal target requires a goal template")
		}
		unresolved, err := synthesis.ValidateGoal(context, *target.GoalSpec)
		if err != nil {
			return nil, err
		}
		ports := make(map[string]algebra.Type, len(target.GoalSpec.Ports))
		for _, port := range target.GoalSpec.Ports {
			ports[port.ID] = port.Type
		}
		for portID, mapping := range rolePorts {
			typ, ok := ports[portID]
			if !ok {
				return nil, unknown("UNKNOWN_TARGET_PORT", "target.role_ports", "mapped port is absent from the goal template")
			}
			if !algebra.TypeEqual(typ, mapping.Type) {
				return nil, invalid("TARGET_PORT_TYPE_MISMATCH", "target.role_ports", "mapped role type differs from the goal port type")
			}
		}
		return unresolved, nil
	default:
		return nil, unsupported("UNSUPPORTED_TARGET_KIND", "target.kind", "semantic target kind is not executable")
	}
}

func frameHasRole(frame valency.ValencyFrame, role string) bool {
	for _, argument := range frame.Arguments {
		if string(argument.Role) == role {
			return true
		}
	}
	return false
}

func validateTargetForFrame(target *lexicon.SemanticTarget, frame valency.ValencyFrame) error {
	if target == nil {
		return invalid("MISSING_SEMANTIC_TARGET", "target", "target is required")
	}
	count := 0
	seenRoles, seenPorts := map[string]bool{}, map[string]bool{}
	for i, mapping := range target.RolePorts {
		if mapping.FrameID != string(frame.ID) {
			continue
		}
		count++
		if !frameHasRole(frame, mapping.Role) {
			return unknown("UNKNOWN_FRAME_ROLE", fmt.Sprintf("target.role_ports[%d].role", i), "mapped role is not declared by the exact selected frame")
		}
		if seenRoles[mapping.Role] || seenPorts[mapping.PortID] {
			return invalid("DUPLICATE_FRAME_PORT_MAPPING", fmt.Sprintf("target.role_ports[%d]", i), "frame role and target port mappings must be one-to-one")
		}
		seenRoles[mapping.Role], seenPorts[mapping.PortID] = true, true
	}
	if count == 0 {
		return unknown("FRAME_NOT_MAPPED_TO_TARGET", "target.role_ports", "selected frame has no explicit role-to-port mappings in this target")
	}
	return nil
}

func languageStrictObject(data []byte, path string, allowed, required []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, invalid("INVALID_SEMANTIC_PACKAGE_OBJECT", path, "expected an object")
	}
	allow := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allow[name] = true
	}
	for name := range fields {
		if !allow[name] {
			return nil, invalid("UNKNOWN_SEMANTIC_PACKAGE_FIELD", path+"."+name, "unknown or mis-cased field")
		}
	}
	for _, name := range required {
		if _, ok := fields[name]; !ok {
			return nil, invalid("MISSING_SEMANTIC_PACKAGE_FIELD", path+"."+name, "required field is missing")
		}
	}
	return fields, nil
}

func languageStrictValue(raw []byte, target any) error {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("missing or null")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing value")
	}
	return nil
}

func decodePackageSource(raw []byte, path string) (provenance.SourceRef, error) {
	fields, err := languageStrictObject(raw, path, []string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}, []string{"id", "version", "license"})
	if err != nil {
		return provenance.SourceRef{}, err
	}
	var source provenance.SourceRef
	for key, target := range map[string]*string{"id": &source.ID, "path": &source.Path, "uri": &source.URI, "version": &source.Version, "license": &source.License, "author": &source.Author, "checksum_sha256": &source.ChecksumSHA256} {
		rawValue, ok := fields[key]
		if !ok {
			continue
		}
		if err := languageStrictValue(rawValue, target); err != nil {
			return provenance.SourceRef{}, invalid("INVALID_PROVENANCE_FIELD", path+"."+key, "field must be a non-null string")
		}
	}
	return source, nil
}

func decodeSemanticFrames(raw []byte) ([]valency.ValencyFrame, error) {
	items, err := languageStrictArray(raw, "frames", MaxSemanticLanguageFrameCount)
	if err != nil {
		return nil, err
	}
	frames := make([]valency.ValencyFrame, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("frames[%d]", i)
		f, err := languageStrictObject(item, path, []string{"id", "sense_id", "semantic_roles", "arguments", "selectional_constraints", "allowed_family_ids", "provenance"}, []string{"id", "sense_id", "semantic_roles", "arguments", "allowed_family_ids", "provenance"})
		if err != nil {
			return nil, err
		}
		var row semanticFrameWire
		for key, target := range map[string]*string{"id": &row.ID, "sense_id": &row.SenseID} {
			if err := languageStrictValue(f[key], target); err != nil {
				return nil, invalid("INVALID_FRAME_FIELD", path+"."+key, "field must be a string")
			}
		}
		if err := languageStrictArrayInto(f["semantic_roles"], &row.SemanticRoles, path+".semantic_roles", valency.MaxSemanticRoles); err != nil {
			return nil, err
		}
		if err := decodeSemanticArguments(f["arguments"], &row.Arguments, path+".arguments"); err != nil {
			return nil, err
		}
		if rawConstraints, ok := f["selectional_constraints"]; ok {
			if err := decodeSemanticConstraints(rawConstraints, &row.SelectionalConstraints, path+".selectional_constraints"); err != nil {
				return nil, err
			}
		} else {
			row.SelectionalConstraints = []semanticConstraintWire{}
		}
		if err := languageStrictArrayInto(f["allowed_family_ids"], &row.AllowedFamilyIDs, path+".allowed_family_ids", valency.MaxFamiliesPerFrame); err != nil {
			return nil, err
		}
		row.Provenance, err = decodePackageSource(f["provenance"], path+".provenance")
		if err != nil {
			return nil, err
		}
		senseID, parseErr := ids.ParseSenseID(row.SenseID)
		if parseErr != nil {
			return nil, invalid("INVALID_SENSE_ID", path+".sense_id", "frame sense ID must be a typed sense ID")
		}
		frame := valency.ValencyFrame{ID: valency.FrameID(row.ID), SenseID: senseID, SemanticRoles: make([]meaning.Role, len(row.SemanticRoles)), Arguments: make([]valency.ArgumentSlot, len(row.Arguments)), SelectionalConstraints: make([]valency.SelectionalConstraint, len(row.SelectionalConstraints)), AllowedFamilyIDs: make([]ids.ConstructionID, len(row.AllowedFamilyIDs)), Provenance: row.Provenance}
		for j, role := range row.SemanticRoles {
			frame.SemanticRoles[j] = meaning.Role(role)
		}
		for j, arg := range row.Arguments {
			frame.Arguments[j] = valency.ArgumentSlot{ID: arg.ID, Role: meaning.Role(arg.Role), Required: arg.Required, GrammaticalFunction: arg.GrammaticalFunction, Case: arg.Case, Adposition: arg.Adposition}
		}
		for j, value := range row.AllowedFamilyIDs {
			family, parseErr := ids.ParseConstructionID(value)
			if parseErr != nil {
				return nil, invalid("INVALID_CONSTRUCTION_ID", fmt.Sprintf("%s.allowed_family_ids[%d]", path, j), "family ID must be a typed construction ID")
			}
			frame.AllowedFamilyIDs[j] = family
		}
		for j, constraint := range row.SelectionalConstraints {
			item := valency.SelectionalConstraint{ID: constraint.ID, Role: meaning.Role(constraint.Role), Strength: valency.ConstraintStrength(constraint.Strength), AllowedConceptIDs: make([]ids.ConceptID, len(constraint.AllowedConceptIDs)), AllowedKinds: make([]core.ConceptKind, len(constraint.AllowedKinds))}
			for k, value := range constraint.AllowedConceptIDs {
				concept, parseErr := ids.ParseConceptID(value)
				if parseErr != nil {
					return nil, invalid("INVALID_CONSTRAINT_CONCEPT_ID", fmt.Sprintf("%s.selectional_constraints[%d].allowed_concept_ids[%d]", path, j, k), "concept ID must be typed")
				}
				item.AllowedConceptIDs[k] = concept
			}
			for k, value := range constraint.AllowedKinds {
				item.AllowedKinds[k] = core.ConceptKind(value)
			}
			frame.SelectionalConstraints[j] = item
		}
		if err := frame.Validate(); err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

func decodeSemanticArguments(raw []byte, out *[]semanticArgumentWire, path string) error {
	items, err := languageStrictArray(raw, path, valency.MaxArgumentsPerFrame)
	if err != nil {
		return err
	}
	rows := make([]semanticArgumentWire, 0, len(items))
	for i, item := range items {
		p := fmt.Sprintf("%s[%d]", path, i)
		f, err := languageStrictObject(item, p, []string{"id", "role", "required", "grammatical_function", "case", "adposition"}, []string{"id", "role", "required"})
		if err != nil {
			return err
		}
		var row semanticArgumentWire
		for key, target := range map[string]*string{"id": &row.ID, "role": &row.Role, "grammatical_function": &row.GrammaticalFunction, "case": &row.Case, "adposition": &row.Adposition} {
			if value, ok := f[key]; ok {
				if err := languageStrictValue(value, target); err != nil {
					return invalid("INVALID_FRAME_ARGUMENT", p+"."+key, "field must be a non-null string")
				}
			}
		}
		if err := languageStrictValue(f["required"], &row.Required); err != nil {
			return invalid("INVALID_FRAME_ARGUMENT", p+".required", "required must be a non-null boolean")
		}
		rows = append(rows, row)
	}
	*out = rows
	return nil
}

func decodeSemanticConstraints(raw []byte, out *[]semanticConstraintWire, path string) error {
	items, err := languageStrictArray(raw, path, valency.MaxConstraintsPerFrame)
	if err != nil {
		return err
	}
	rows := make([]semanticConstraintWire, 0, len(items))
	for i, item := range items {
		p := fmt.Sprintf("%s[%d]", path, i)
		f, err := languageStrictObject(item, p, []string{"id", "role", "strength", "allowed_concept_ids", "allowed_kinds"}, []string{"id", "role", "strength"})
		if err != nil {
			return err
		}
		var row semanticConstraintWire
		for key, target := range map[string]*string{"id": &row.ID, "role": &row.Role, "strength": &row.Strength} {
			if err := languageStrictValue(f[key], target); err != nil {
				return invalid("INVALID_FRAME_CONSTRAINT", p+"."+key, "field must be a non-null string")
			}
		}
		if v, ok := f["allowed_concept_ids"]; ok {
			if err := languageStrictArrayInto(v, &row.AllowedConceptIDs, p+".allowed_concept_ids", valency.MaxConceptsPerConstraint); err != nil {
				return err
			}
		}
		if v, ok := f["allowed_kinds"]; ok {
			if err := languageStrictArrayInto(v, &row.AllowedKinds, p+".allowed_kinds", valency.MaxKindsPerConstraint); err != nil {
				return err
			}
		}
		rows = append(rows, row)
	}
	*out = rows
	return nil
}

func languageStrictArray(raw []byte, path string, maximum int) ([][]byte, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, invalid("INVALID_SEMANTIC_PACKAGE_ARRAY", path, "field must be a non-null array")
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil || rows == nil {
		return nil, invalid("INVALID_SEMANTIC_PACKAGE_ARRAY", path, "field must be an array")
	}
	if len(rows) > maximum {
		return nil, resource("semantic_package_items", len(rows), maximum)
	}
	out := make([][]byte, len(rows))
	for i := range rows {
		out[i] = rows[i]
	}
	return out, nil
}

func languageStrictArrayInto(raw []byte, target any, path string, maximum int) error {
	rows, err := languageStrictArray(raw, path, maximum)
	if err != nil {
		return err
	}
	if target == nil {
		return invalid("INVALID_SEMANTIC_PACKAGE_ARRAY", path, "array target is required")
	}
	// Decode each scalar individually so null elements cannot silently become zero values.
	switch output := target.(type) {
	case *[]string:
		values := make([]string, len(rows))
		for i, row := range rows {
			if err := languageStrictValue(row, &values[i]); err != nil {
				return invalid("INVALID_SEMANTIC_PACKAGE_ARRAY_ITEM", fmt.Sprintf("%s[%d]", path, i), "array item must be a non-null string")
			}
		}
		*output = values
	default:
		return invalid("INVALID_SEMANTIC_PACKAGE_ARRAY", path, "unsupported array type")
	}
	return nil
}

func validateSemanticPackageFrames(index *lexicon.Index, packageLanguage lexicon.LanguageID, frames []valency.ValencyFrame) error {
	if index == nil {
		return invalid("MISSING_LEXICON_INDEX", "lexicon", "built lexicon index is required")
	}
	doc := index.Document()
	senses := make(map[string]lexicon.LexicalSense, len(doc.Senses))
	lexemes := make(map[string]lexicon.MorphLexeme, len(doc.Lexemes))
	for _, lexeme := range doc.Lexemes {
		lexemes[lexeme.ID.String()] = lexeme
	}
	for i, lexeme := range doc.Lexemes {
		if lexeme.Language != packageLanguage {
			return invalid("PACKAGE_LEXEME_LANGUAGE_MISMATCH", fmt.Sprintf("lexicon.lexemes[%d].language", i), "every lexeme in the package must use the declared package language")
		}
	}
	for _, sense := range doc.Senses {
		senses[sense.ID.String()] = sense
	}
	frameIDs := make(map[string]bool, len(frames))
	for i, frame := range frames {
		id := string(frame.ID)
		if frameIDs[id] {
			return invalid("DUPLICATE_VALENCY_FRAME_ID", fmt.Sprintf("frames[%d].id", i), "frame ID is duplicated")
		}
		frameIDs[id] = true
		sense, ok := senses[frame.SenseID.String()]
		if !ok {
			return unknown("UNKNOWN_FRAME_SENSE", fmt.Sprintf("frames[%d].sense_id", i), "frame sense does not resolve in the package lexicon")
		}
		lexeme, ok := lexemes[sense.LexemeID.String()]
		if !ok || lexeme.Language != packageLanguage {
			return invalid("INVALID_FRAME_LANGUAGE", fmt.Sprintf("frames[%d].sense_id", i), "frame sense lexeme must belong to the package language")
		}
		declared := false
		for _, frameID := range sense.ValencyIDs {
			if string(frameID) == id {
				declared = true
				break
			}
		}
		if !declared {
			return invalid("FRAME_NOT_LINKED_FROM_SENSE", fmt.Sprintf("frames[%d].id", i), "frame ID must be explicitly listed by its lexical sense")
		}
	}
	for i, sense := range doc.Senses {
		for j, frameID := range sense.ValencyIDs {
			frame, ok := findFrame(frames, string(frameID))
			if !ok || frame.SenseID != sense.ID {
				return unknown("UNKNOWN_SENSE_VALENCY_FRAME", fmt.Sprintf("lexicon.senses[%d].valency_ids[%d]", i, j), "each linked frame ID must resolve to an authored frame for this exact sense")
			}
		}
	}
	// Every tagged binding is checked against its exact sense/frame role map,
	// regardless of pending/approved status.
	for i, binding := range doc.Bindings {
		if binding.Target == nil {
			continue
		}
		for j, mapping := range binding.Target.RolePorts {
			frame, ok := findFrame(frames, mapping.FrameID)
			if !ok || frame.SenseID != binding.SenseID {
				return unknown("UNKNOWN_TARGET_FRAME", fmt.Sprintf("bindings[%d].target.role_ports[%d].frame_id", i, j), "target frame must be present and belong to the exact linked sense")
			}
			if !frameHasRole(frame, mapping.Role) {
				return unknown("UNKNOWN_FRAME_ROLE", fmt.Sprintf("bindings[%d].target.role_ports[%d].role", i, j), "target role must be declared by the linked frame")
			}
		}
	}
	return nil
}

func findFrame(frames []valency.ValencyFrame, id string) (valency.ValencyFrame, bool) {
	for _, frame := range frames {
		if string(frame.ID) == id {
			return frame, true
		}
	}
	return valency.ValencyFrame{}, false
}

func validSemanticPackageText(value string, max int) bool {
	return value != "" && len(value) <= max && utf8.ValidString(value) && strings.TrimSpace(value) == value
}
func validLanguageTag(value string) bool {
	if value == "" || len(value) > 35 || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for i, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, r := range part {
			if !(r >= 'a' && r <= 'z') && !(i > 0 && r >= '0' && r <= '9') {
				return false
			}
		}
	}
	return true
}
