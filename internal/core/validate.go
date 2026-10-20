package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/status"
)

const MaxGraphDiagnostics = status.MaxDiagnostics
const MaxConceptDocumentBytes = 16 << 20

// GraphDiagnostic is a stable, bounded validation result with an input path.
type GraphDiagnostic struct {
	Code    string
	Message string
	Path    string
}

// GraphValidationError carries all bounded invalid-input diagnostics.
type GraphValidationError struct{ Diagnostics []GraphDiagnostic }

var ErrInvalidGraph = errors.New("invalid concept graph")

func (err *GraphValidationError) Error() string {
	if err == nil || len(err.Diagnostics) == 0 {
		return "invalid concept graph"
	}
	return fmt.Sprintf("invalid concept graph: %s at %s", err.Diagnostics[0].Message, err.Diagnostics[0].Path)
}

func (err *GraphValidationError) Unwrap() error { return ErrInvalidGraph }

func (err *GraphValidationError) add(code, path, message string) {
	if len(err.Diagnostics) >= MaxGraphDiagnostics {
		return
	}
	err.Diagnostics = append(err.Diagnostics, GraphDiagnostic{
		Code:    boundedText(code, status.MaxDiagnosticCodeBytes),
		Path:    boundedText(path, status.MaxDiagnosticPathBytes),
		Message: boundedText(message, status.MaxDiagnosticMessageBytes),
	})
}

func boundedText(value string, maximum int) string {
	if !utf8.ValidString(value) {
		value = strings.ToValidUTF8(value, "�")
	}
	if len(value) <= maximum {
		return value
	}
	value = value[:maximum]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

// RelationDefinition is the source declaration for one relation schema.
type RelationDefinition struct{ RelationType }

// ConceptDocument is the versioned JSON source envelope described by the schema.
type ConceptDocument struct {
	SchemaVersion int                  `json:"schema_version"`
	Relations     []RelationDefinition `json:"relations"`
	Concepts      []ConceptDefinition  `json:"concepts"`
}

// DecodeConceptDocument rejects oversized input, unknown fields, trailing JSON
// values and unsupported schema versions before compiling the graph.
func DecodeConceptDocument(data []byte) (ConceptDocument, error) {
	if len(data) > MaxConceptDocumentBytes {
		return ConceptDocument{}, fmt.Errorf("document: exceeds %d bytes", MaxConceptDocumentBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var source struct {
		SchemaVersion *int                  `json:"schema_version"`
		Relations     *[]RelationDefinition `json:"relations"`
		Concepts      *[]ConceptDefinition  `json:"concepts"`
	}
	if err := decoder.Decode(&source); err != nil {
		return ConceptDocument{}, fmt.Errorf("document: invalid JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ConceptDocument{}, fmt.Errorf("document: expected one JSON value")
	}
	if source.SchemaVersion == nil {
		return ConceptDocument{}, fmt.Errorf("schema_version: required field is missing")
	}
	if *source.SchemaVersion != 1 {
		return ConceptDocument{}, fmt.Errorf("schema_version: unsupported version %d", *source.SchemaVersion)
	}
	if source.Relations == nil {
		return ConceptDocument{}, fmt.Errorf("relations: required array is missing or null")
	}
	if source.Concepts == nil {
		return ConceptDocument{}, fmt.Errorf("concepts: required array is missing or null")
	}
	return ConceptDocument{SchemaVersion: *source.SchemaVersion, Relations: *source.Relations, Concepts: *source.Concepts}, nil
}

// CompileConceptDocument decodes, validates and compiles one versioned JSON document.
func CompileConceptDocument(data []byte) (*ConceptGraph, error) {
	document, err := DecodeConceptDocument(data)
	if err != nil {
		validation := &GraphValidationError{}
		validation.add("INVALID_DOCUMENT", "document", err.Error())
		return nil, validation
	}
	return CompileConceptGraph(document.Concepts, document.Relations)
}

// CompileConceptGraph validates and freezes new language-neutral definitions.
// It does not infer facts from typical edges or mutate any caller-owned data.
func CompileConceptGraph(definitions []ConceptDefinition, relationDefinitions []RelationDefinition) (*ConceptGraph, error) {
	if len(definitions) > MaxConcepts {
		validation := &GraphValidationError{}
		validation.add("CONCEPT_LIMIT", "concepts", fmt.Sprintf("exceeds limit %d", MaxConcepts))
		return nil, validation
	}
	if len(relationDefinitions) > MaxRelations {
		validation := &GraphValidationError{}
		validation.add("RELATION_LIMIT", "relations", fmt.Sprintf("exceeds limit %d", MaxRelations))
		return nil, validation
	}
	validation := &GraphValidationError{}
	conceptByKey := make(map[string]ConceptDefinition, len(definitions))
	for index, definition := range definitions {
		path := fmt.Sprintf("concepts[%d]", index)
		if err := validateDefinition(definition, path); err != nil {
			validation.add("INVALID_CONCEPT", path, err.Error())
			continue
		}
		key := canonicalConceptKey(definition.ID)
		if _, exists := conceptByKey[key]; exists {
			validation.add("DUPLICATE_CONCEPT", path+".id", "duplicate concept ID "+key)
			continue
		}
		conceptByKey[key] = cloneDefinition(definition)
	}
	relations := make(map[string]RelationType, len(relationDefinitions))
	for index, definition := range relationDefinitions {
		relation := cloneRelationType(definition.RelationType)
		key := relation.ID.String()
		path := fmt.Sprintf("relations[%d]", index)
		if !relation.ID.valid() {
			validation.add("INVALID_RELATION", path+".id", "relation ID must use relation namespace")
			continue
		}
		if _, exists := relations[key]; exists {
			validation.add("DUPLICATE_RELATION", path+".id", "duplicate relation ID "+key)
			continue
		}
		if err := validateKinds(relation.SubjectKind, path+".subject_kinds"); err != nil {
			validation.add("INVALID_RELATION", path+".subject_kinds", err.Error())
			continue
		}
		if err := validateKinds(relation.ObjectKind, path+".object_kinds"); err != nil {
			validation.add("INVALID_RELATION", path+".object_kinds", err.Error())
			continue
		}
		if (relation.Symmetric || relation.Transitive) && !sameKinds(relation.SubjectKind, relation.ObjectKind) {
			validation.add("INVALID_RELATION", path, "symmetric or transitive relation must have matching endpoint kind signatures")
			continue
		}
		relations[key] = relation
	}
	for _, key := range sortedKeys(conceptByKey) {
		definition := conceptByKey[key]
		path := "concepts[" + key + "]"
		if definition.Parent != nil {
			parent, exists := conceptByKey[definition.Parent.String()]
			if !exists {
				validation.add("DANGLING_PARENT", path+".parent", "parent concept does not resolve: "+definition.Parent.String())
			} else if parent.Kind != definition.Kind && parent.Kind != KindAbstract {
				validation.add("PARENT_KIND", path+".parent", "parent must have the same kind or be abstract")
			}
		}
		for index, edge := range definition.DefinitionEdges {
			edgePath := fmt.Sprintf("%s.definition_edges[%d]", path, index)
			relation, exists := relations[edge.Relation.String()]
			if !exists {
				validation.add("UNKNOWN_RELATION", edgePath+".relation", "relation schema does not resolve: "+edge.Relation.String())
				continue
			}
			target, exists := conceptByKey[edge.Target.String()]
			if !exists {
				validation.add("DANGLING_TARGET", edgePath+".target", "target concept does not resolve: "+edge.Target.String())
				continue
			}
			if !containsKind(relation.SubjectKind, definition.Kind) {
				validation.add("RELATION_SUBJECT_KIND", edgePath, "subject kind is not allowed by relation schema")
			}
			if !containsKind(relation.ObjectKind, target.Kind) {
				validation.add("RELATION_OBJECT_KIND", edgePath, "object kind is not allowed by relation schema")
			}
			if edge.Role != "" {
				role, ok := findRole(definition.RoleTypes, edge.Role)
				if !ok {
					validation.add("UNKNOWN_ROLE", edgePath+".role", "role is not declared on subject concept")
					continue
				}
				if !containsKind(role.AllowedKinds, target.Kind) {
					validation.add("ROLE_TARGET_KIND", edgePath+".target", "target kind is not allowed by role signature")
				}
			}
		}
	}
	validateParentCycles(conceptByKey, validation)
	if len(validation.Diagnostics) > 0 {
		return nil, validation
	}
	keys := sortedKeys(conceptByKey)
	graph := &ConceptGraph{concepts: make([]compiledConcept, 0, len(keys)), byID: make(map[string]int, len(keys)), relations: make(map[string]RelationType, len(relations))}
	for _, key := range keys {
		definition := conceptByKey[key]
		compiled := compiledConcept{id: definition.ID, kind: definition.Kind, roles: cloneRoles(definition.RoleTypes), edges: append([]DefinitionEdge(nil), definition.DefinitionEdges...), effects: append([]SemanticEffect(nil), definition.SemanticEffects...), definition: definition.Definition}
		if definition.Parent != nil {
			parent := *definition.Parent
			compiled.parent = &parent
		}
		graph.byID[key] = len(graph.concepts)
		graph.concepts = append(graph.concepts, compiled)
	}
	for key, relation := range relations {
		graph.relations[key] = cloneRelationType(relation)
	}
	return graph, nil
}

// ValidateConceptGraph is a convenience validator that discards the compiled snapshot.
func ValidateConceptGraph(definitions []ConceptDefinition, relationDefinitions []RelationDefinition) []GraphDiagnostic {
	_, err := CompileConceptGraph(definitions, relationDefinitions)
	if err == nil {
		return nil
	}
	if validation, ok := err.(*GraphValidationError); ok {
		return append([]GraphDiagnostic(nil), validation.Diagnostics...)
	}
	validation := &GraphValidationError{}
	validation.add("INVALID_GRAPH", "", err.Error())
	return validation.Diagnostics
}

// DiagnosticsAsStatus maps invalid graph input to the repository's typed outcome.
func DiagnosticsAsStatus(diagnostics []GraphDiagnostic) []status.Diagnostic {
	if len(diagnostics) > status.MaxDiagnostics {
		diagnostics = diagnostics[:status.MaxDiagnostics]
	}
	result := make([]status.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		result = append(result, status.Diagnostic{
			Code:     boundedText(diagnostic.Code, status.MaxDiagnosticCodeBytes),
			Message:  boundedText(diagnostic.Message, status.MaxDiagnosticMessageBytes),
			Path:     boundedText(diagnostic.Path, status.MaxDiagnosticPathBytes),
			Severity: status.SeverityError,
		})
	}
	return result
}

func cloneRelationType(relation RelationType) RelationType {
	relation.SubjectKind = append([]ConceptKind(nil), relation.SubjectKind...)
	relation.ObjectKind = append([]ConceptKind(nil), relation.ObjectKind...)
	return relation
}

func cloneDefinition(definition ConceptDefinition) ConceptDefinition {
	if definition.Parent != nil {
		parent := *definition.Parent
		definition.Parent = &parent
	}
	definition.RoleTypes = cloneRoles(definition.RoleTypes)
	definition.DefinitionEdges = append([]DefinitionEdge(nil), definition.DefinitionEdges...)
	definition.SemanticEffects = append([]SemanticEffect(nil), definition.SemanticEffects...)
	return definition
}

func findRole(roles []RoleType, name string) (RoleType, bool) {
	for _, role := range roles {
		if role.Name == name {
			return role, true
		}
	}
	return RoleType{}, false
}

func sameKinds(left, right []ConceptKind) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy, rightCopy := append([]ConceptKind(nil), left...), append([]ConceptKind(nil), right...)
	sort.Slice(leftCopy, func(i, j int) bool { return leftCopy[i] < leftCopy[j] })
	sort.Slice(rightCopy, func(i, j int) bool { return rightCopy[i] < rightCopy[j] })
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}

func validateParentCycles(concepts map[string]ConceptDefinition, validation *GraphValidationError) {
	state := make(map[string]uint8, len(concepts))
	keys := make([]string, 0, len(concepts))
	for key := range concepts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	path := make([]string, 0)
	positions := make(map[string]int)
	for _, key := range keys {
		if len(validation.Diagnostics) >= MaxGraphDiagnostics {
			return
		}
		if state[key] == 2 {
			continue
		}
		path = path[:0]
		current := key
		for state[current] != 2 {
			if position, exists := positions[current]; exists {
				cycle := append(append([]string(nil), path[position:]...), current)
				validation.add("PARENT_CYCLE", "concepts["+current+"].parent", "parent cycle detected: "+fmt.Sprint(cycle))
				break
			}
			positions[current] = len(path)
			path = append(path, current)
			state[current] = 1
			definition := concepts[current]
			if definition.Parent == nil {
				break
			}
			parent := definition.Parent.String()
			if _, exists := concepts[parent]; !exists {
				break
			}
			current = parent
		}
		for _, visited := range path {
			state[visited] = 2
			delete(positions, visited)
		}
	}
}
