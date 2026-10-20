package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/ids"
)

type definitionEdgeJSON struct {
	Relation string       `json:"relation"`
	Target   string       `json:"target"`
	Strength EdgeStrength `json:"strength"`
	Role     string       `json:"role,omitempty"`
}

type conceptDefinitionJSON struct {
	ID              string               `json:"id"`
	Kind            ConceptKind          `json:"kind"`
	Parent          *string              `json:"parent,omitempty"`
	RoleTypes       []RoleType           `json:"role_types,omitempty"`
	DefinitionEdges []definitionEdgeJSON `json:"definition_edges,omitempty"`
	SemanticEffects []SemanticEffect     `json:"semantic_effects,omitempty"`
	Definition      string               `json:"definition,omitempty"`
}

type roleJSON struct {
	Name         *string        `json:"name"`
	AllowedKinds *[]ConceptKind `json:"allowed_kinds"`
}

func (role *RoleType) UnmarshalJSON(data []byte) error {
	var source roleJSON
	if err := decodeStrict(data, &source); err != nil {
		return err
	}
	if source.Name == nil || source.AllowedKinds == nil {
		return fmt.Errorf("name and allowed_kinds are required")
	}
	role.Name = *source.Name
	role.AllowedKinds = *source.AllowedKinds
	return nil
}

type semanticEffectJSON struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

func (effect *SemanticEffect) UnmarshalJSON(data []byte) error {
	var source semanticEffectJSON
	if err := decodeStrict(data, &source); err != nil {
		return err
	}
	if source.Name == nil || source.Value == nil {
		return fmt.Errorf("name and value are required")
	}
	effect.Name = *source.Name
	effect.Value = *source.Value
	return nil
}

func (definition *ConceptDefinition) UnmarshalJSON(data []byte) error {
	var source conceptDefinitionJSON
	if err := decodeStrict(data, &source); err != nil {
		return err
	}
	id, err := ids.ParseConceptID(source.ID)
	if err != nil {
		return fmt.Errorf("id: %w", err)
	}
	definition.ID = id
	definition.Kind = source.Kind
	definition.RoleTypes = source.RoleTypes
	definition.SemanticEffects = source.SemanticEffects
	definition.Definition = source.Definition
	if source.Parent != nil {
		parent, err := ids.ParseConceptID(*source.Parent)
		if err != nil {
			return fmt.Errorf("parent: %w", err)
		}
		definition.Parent = &parent
	}
	definition.DefinitionEdges = make([]DefinitionEdge, len(source.DefinitionEdges))
	for index, edge := range source.DefinitionEdges {
		relation, err := ParseRelationID(edge.Relation)
		if err != nil {
			return fmt.Errorf("definition_edges[%d].relation: %w", index, err)
		}
		target, err := ids.ParseConceptID(edge.Target)
		if err != nil {
			return fmt.Errorf("definition_edges[%d].target: %w", index, err)
		}
		definition.DefinitionEdges[index] = DefinitionEdge{Relation: relation, Target: target, Strength: edge.Strength, Role: edge.Role}
	}
	return nil
}

func (definition ConceptDefinition) MarshalJSON() ([]byte, error) {
	source := conceptDefinitionJSON{ID: definition.ID.String(), Kind: definition.Kind, RoleTypes: definition.RoleTypes, SemanticEffects: definition.SemanticEffects, Definition: definition.Definition}
	if definition.Parent != nil {
		parent := definition.Parent.String()
		source.Parent = &parent
	}
	source.DefinitionEdges = make([]definitionEdgeJSON, len(definition.DefinitionEdges))
	for index, edge := range definition.DefinitionEdges {
		source.DefinitionEdges[index] = definitionEdgeJSON{Relation: edge.Relation.String(), Target: edge.Target.String(), Strength: edge.Strength, Role: edge.Role}
	}
	return json.Marshal(source)
}

type relationDefinitionJSON struct {
	ID          string         `json:"id"`
	SubjectKind *[]ConceptKind `json:"subject_kinds"`
	ObjectKind  *[]ConceptKind `json:"object_kinds"`
	Symmetric   bool           `json:"symmetric,omitempty"`
	Transitive  bool           `json:"transitive,omitempty"`
}

func (definition *RelationDefinition) UnmarshalJSON(data []byte) error {
	var source relationDefinitionJSON
	if err := decodeStrict(data, &source); err != nil {
		return err
	}
	id, err := ParseRelationID(source.ID)
	if err != nil {
		return fmt.Errorf("id: %w", err)
	}
	if source.SubjectKind == nil || source.ObjectKind == nil {
		return fmt.Errorf("subject_kinds and object_kinds are required")
	}
	definition.RelationType = RelationType{ID: id, SubjectKind: *source.SubjectKind, ObjectKind: *source.ObjectKind, Symmetric: source.Symmetric, Transitive: source.Transitive}
	return nil
}

func (definition RelationDefinition) MarshalJSON() ([]byte, error) {
	relation := definition.RelationType
	return json.Marshal(relationDefinitionJSON{ID: relation.ID.String(), SubjectKind: &relation.SubjectKind, ObjectKind: &relation.ObjectKind, Symmetric: relation.Symmetric, Transitive: relation.Transitive})
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}
