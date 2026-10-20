package core

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/ids"
)

const (
	MaxConcepts              = 100_000
	MaxRelations             = 1_024
	MaxEdgesPerConcept       = 256
	MaxRolesPerConcept       = 64
	MaxEffectsPerConcept     = 64
	MaxDefinitionStringBytes = 1_024
)

// DefinitionEdge is an explicit relation between concepts. Strength is kept
// attached to source data so typical assertions cannot silently entail facts.
type DefinitionEdge struct {
	Relation RelationID
	Target   ids.ConceptID
	Strength EdgeStrength
	Role     string
}

// ConceptDefinition is the closed source form accepted by CompileConceptGraph.
// Slices are copied during compilation; callers may freely reuse their input.
type ConceptDefinition struct {
	ID              ids.ConceptID
	Kind            ConceptKind
	Parent          *ids.ConceptID
	RoleTypes       []RoleType
	DefinitionEdges []DefinitionEdge
	SemanticEffects []SemanticEffect
	Definition      string
}

// Concept is a value view of an immutable compiled concept.
type Concept struct {
	ID              ids.ConceptID
	Kind            ConceptKind
	Parent          *ids.ConceptID
	RoleTypes       []RoleType
	DefinitionEdges []DefinitionEdge
	SemanticEffects []SemanticEffect
	Definition      string
}

// GraphNode is the minimal defensive view needed by graph index builders.
type GraphNode struct {
	ID              ids.ConceptID
	Kind            ConceptKind
	Parent          *ids.ConceptID
	DefinitionEdges []DefinitionEdge
}

type compiledConcept struct {
	id         ids.ConceptID
	kind       ConceptKind
	parent     *ids.ConceptID
	roles      []RoleType
	edges      []DefinitionEdge
	effects    []SemanticEffect
	definition string
}

// ConceptGraph is an immutable compiled ontology snapshot.
type ConceptGraph struct {
	concepts  []compiledConcept
	byID      map[string]int
	relations map[string]RelationType
}

// Lookup returns a defensive copy of a concept. Mutating the returned slices
// cannot change the immutable graph.
func (graph *ConceptGraph) Lookup(id ids.ConceptID) (Concept, bool) {
	if graph == nil {
		return Concept{}, false
	}
	index, ok := graph.byID[id.String()]
	if !ok {
		return Concept{}, false
	}
	return graph.concepts[index].view(), true
}

// Len returns the number of concepts in the snapshot.
func (graph *ConceptGraph) Len() int {
	if graph == nil {
		return 0
	}
	return len(graph.concepts)
}

// Concepts returns all concept values in canonical ID order. Every value is a
// defensive copy, so callers can build immutable indexes without accessing the
// graph's private storage.
func (graph *ConceptGraph) Concepts() []Concept {
	if graph == nil {
		return nil
	}
	result := make([]Concept, len(graph.concepts))
	for index, concept := range graph.concepts {
		result[index] = concept.view()
	}
	return result
}

// GraphSize reports concept and stored definition-edge counts without making
// copies. Parent links are taxonomy edges and are counted separately by indexers.
func (graph *ConceptGraph) GraphSize() (concepts, definitionEdges, parentLinks int) {
	if graph == nil {
		return 0, 0, 0
	}
	for _, concept := range graph.concepts {
		definitionEdges += len(concept.edges)
		if concept.parent != nil {
			parentLinks++
		}
	}
	return len(graph.concepts), definitionEdges, parentLinks
}

// GraphNodes returns minimal nodes in canonical ID order without copying
// unrelated role or semantic-effect data.
func (graph *ConceptGraph) GraphNodes() []GraphNode {
	if graph == nil {
		return nil
	}
	result := make([]GraphNode, len(graph.concepts))
	for index, concept := range graph.concepts {
		node := GraphNode{ID: concept.id, Kind: concept.kind, DefinitionEdges: append([]DefinitionEdge(nil), concept.edges...)}
		if concept.parent != nil {
			parent := *concept.parent
			node.Parent = &parent
		}
		result[index] = node
	}
	return result
}

func (concept compiledConcept) view() Concept {
	result := Concept{ID: concept.id, Kind: concept.kind, Definition: concept.definition}
	if concept.parent != nil {
		parent := *concept.parent
		result.Parent = &parent
	}
	result.RoleTypes = cloneRoles(concept.roles)
	result.DefinitionEdges = append([]DefinitionEdge(nil), concept.edges...)
	result.SemanticEffects = append([]SemanticEffect(nil), concept.effects...)
	return result
}

func cloneRoles(roles []RoleType) []RoleType {
	result := make([]RoleType, len(roles))
	for i, role := range roles {
		result[i] = RoleType{Name: role.Name, AllowedKinds: append([]ConceptKind(nil), role.AllowedKinds...)}
	}
	return result
}

func canonicalConceptKey(id ids.ConceptID) string { return id.String() }

func validateDefinition(definition ConceptDefinition, path string) error {
	if !definition.idValid() {
		return fmt.Errorf("%s.id: concept ID is empty or invalid", path)
	}
	if !definition.Kind.Valid() {
		return fmt.Errorf("%s.kind: unknown concept kind %q", path, definition.Kind)
	}
	if !utf8.ValidString(definition.Definition) || len(definition.Definition) > MaxDefinitionStringBytes {
		return fmt.Errorf("%s.definition: exceeds %d bytes", path, MaxDefinitionStringBytes)
	}
	if len(definition.RoleTypes) > MaxRolesPerConcept {
		return fmt.Errorf("%s.role_types: exceeds %d", path, MaxRolesPerConcept)
	}
	if len(definition.DefinitionEdges) > MaxEdgesPerConcept {
		return fmt.Errorf("%s.definition_edges: exceeds %d", path, MaxEdgesPerConcept)
	}
	if len(definition.SemanticEffects) > MaxEffectsPerConcept {
		return fmt.Errorf("%s.semantic_effects: exceeds %d", path, MaxEffectsPerConcept)
	}
	seen := make(map[string]struct{}, len(definition.RoleTypes))
	for i, role := range definition.RoleTypes {
		rolePath := fmt.Sprintf("%s.role_types[%d]", path, i)
		if !validName(role.Name) {
			return fmt.Errorf("%s.name: invalid role name", rolePath)
		}
		if _, exists := seen[role.Name]; exists {
			return fmt.Errorf("%s.name: duplicate role %q", rolePath, role.Name)
		}
		seen[role.Name] = struct{}{}
		if err := validateKinds(role.AllowedKinds, rolePath+".allowed_kinds"); err != nil {
			return err
		}
	}
	for i, edge := range definition.DefinitionEdges {
		if edge.Relation.IsZero() {
			return fmt.Errorf("%s.definition_edges[%d].relation: empty relation ID", path, i)
		}
		if !edge.Strength.Valid() {
			return fmt.Errorf("%s.definition_edges[%d].strength: unknown strength %q", path, i, edge.Strength)
		}
		if edge.Role != "" && !validName(edge.Role) {
			return fmt.Errorf("%s.definition_edges[%d].role: invalid role name", path, i)
		}
	}
	for i, effect := range definition.SemanticEffects {
		if !validName(effect.Name) || !utf8.ValidString(effect.Value) || len(effect.Value) > MaxDefinitionStringBytes {
			return fmt.Errorf("%s.semantic_effects[%d]: invalid name or value bound", path, i)
		}
	}
	return nil
}

func (definition ConceptDefinition) idValid() bool {
	symbol := definition.ID.Symbol()
	canonical, err := symbol.Canonical()
	return err == nil && symbol.Namespace == ids.NamespaceConcept && canonical != ""
}

func validName(name string) bool {
	if name == "" || len(name) > 64 || strings.TrimSpace(name) != name {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func validateKinds(kinds []ConceptKind, path string) error {
	seen := map[ConceptKind]bool{}
	for i, kind := range kinds {
		if !kind.Valid() {
			return fmt.Errorf("%s[%d]: unknown concept kind %q", path, i, kind)
		}
		if seen[kind] {
			return fmt.Errorf("%s[%d]: duplicate concept kind %q", path, i, kind)
		}
		seen[kind] = true
	}
	return nil
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
