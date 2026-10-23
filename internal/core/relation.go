// Package core contains language-neutral concept definitions and compiled ontology.
package core

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/ids"
)

// ConceptKind classifies concepts without introducing language-specific POS.
type ConceptKind string

const (
	KindEntity   ConceptKind = "entity"
	KindEvent    ConceptKind = "event"
	KindProperty ConceptKind = "property"
	KindRelation ConceptKind = "relation"
	KindAbstract ConceptKind = "abstract"
)

func (kind ConceptKind) Valid() bool {
	switch kind {
	case KindEntity, KindEvent, KindProperty, KindRelation, KindAbstract:
		return true
	default:
		return false
	}
}

// RelationID is the stable typed identity of a relation schema.
type RelationID struct{ symbol ids.Symbol }

func NewRelationID(name string) (RelationID, error) {
	symbol, err := ids.NewSymbol(ids.NamespaceRelation, name)
	return RelationID{symbol: symbol}, err
}

func ParseRelationID(value string) (RelationID, error) {
	symbol, err := ids.Parse(value)
	if err != nil {
		return RelationID{}, err
	}
	if symbol.Namespace != ids.NamespaceRelation {
		return RelationID{}, fmt.Errorf("relation ID namespace must be relation")
	}
	return RelationID{symbol: symbol}, nil
}

func (id RelationID) String() string { value, _ := id.symbol.Canonical(); return value }

// Symbol returns the stable value used by typed indexes.
func (id RelationID) Symbol() ids.Symbol { return id.symbol }
func (id RelationID) IsZero() bool       { return id.symbol.Name == "" }
func (id RelationID) valid() bool {
	canonical, err := id.symbol.Canonical()
	return err == nil && id.symbol.Namespace == ids.NamespaceRelation && canonical != ""
}

// Valid reports whether id has a canonical relation namespace and name.
func (id RelationID) Valid() bool { return id.valid() }

// EdgeStrength preserves the distinction between asserted required structure
// and typical descriptions. Typical edges never entail a required fact.
type EdgeStrength string

const (
	StrengthRequired EdgeStrength = "required"
	StrengthTypical  EdgeStrength = "typical"
)

func (strength EdgeStrength) Valid() bool {
	return strength == StrengthRequired || strength == StrengthTypical
}

// RelationType defines allowed concept kinds at either end of an edge.
// Empty kind lists mean any valid ConceptKind. Symmetric requires endpoint
// signatures to be equivalent and is enforced during graph compilation.
type RelationType struct {
	ID             RelationID
	SubjectKind    []ConceptKind
	ObjectKind     []ConceptKind
	EventConceptID ids.ConceptID
	ArgumentRoles  [2]string
	Symmetric      bool
	Transitive     bool
}

// Relation returns a defensive copy of a relation schema from a compiled graph.
func (graph *ConceptGraph) Relation(id RelationID) (RelationType, bool) {
	if graph == nil || !id.valid() {
		return RelationType{}, false
	}
	relation, ok := graph.relations[id.String()]
	if !ok {
		return RelationType{}, false
	}
	return cloneRelationType(relation), true
}

// RelationTypes returns relation schemas in canonical ID order with copies of
// their kind signatures.
func (graph *ConceptGraph) RelationTypes() []RelationType {
	if graph == nil {
		return nil
	}
	keys := make([]string, 0, len(graph.relations))
	for key := range graph.relations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]RelationType, len(keys))
	for index, key := range keys {
		result[index] = cloneRelationType(graph.relations[key])
	}
	return result
}

// RoleType is a named semantic role and its allowed target kinds.
type RoleType struct {
	Name         string        `json:"name"`
	AllowedKinds []ConceptKind `json:"allowed_kinds"`
}

// SemanticEffect is a declared, language-neutral effect label. It does not
// describe morphology, surface order, or a specific utterance's truth.
type SemanticEffect struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func containsKind(kinds []ConceptKind, kind ConceptKind) bool {
	if len(kinds) == 0 {
		return true
	}
	for _, allowed := range kinds {
		if allowed == kind {
			return true
		}
	}
	return false
}
