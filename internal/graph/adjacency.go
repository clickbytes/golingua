package graph

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
)

// NodeID is a dense index local to one immutable GraphReader snapshot.
type NodeID = ids.NodeID

// Edge is a required/typical definition edge or an implicit taxonomy edge.
// Taxonomy edges have IsParent=true and a zero RelationID.
type Edge struct {
	From     NodeID
	To       NodeID
	Relation core.RelationID
	Strength core.EdgeStrength
	Role     string
	IsParent bool
}

// NodeInfo is a read-only view of a graph node and its semantic kind.
type NodeInfo struct {
	ID        ids.ConceptID
	Kind      core.ConceptKind
	Parent    NodeID
	HasParent bool
}

// GraphReader is an immutable dense adjacency snapshot. Its accessors return
// values or defensive copies, never internal slices.
type GraphReader struct {
	conceptIDs   []ids.ConceptID
	byConcept    map[ids.ConceptID]NodeID
	kinds        []core.ConceptKind
	parents      []NodeID
	hasParent    []bool
	edges        []Edge
	outgoing     [][]int
	incoming     [][]int
	byRelation   map[core.RelationID][]int
	relationType map[core.RelationID]core.RelationType
}

// Compile builds a deterministic, bounded adjacency index from a validated
// ConceptGraph. Required and typical strengths are preserved without inference.
func Compile(source *core.ConceptGraph) (*GraphReader, error) {
	if source == nil {
		return nil, invalid("INVALID_GRAPH", "graph", "compiled concept graph is nil")
	}
	nodeCount, definitionEdgeCount, parentCount := source.GraphSize()
	if nodeCount > MaxGraphNodes {
		return nil, resourceLimit("graph_nodes", nodeCount, MaxGraphNodes)
	}
	totalEdges := definitionEdgeCount + parentCount
	if totalEdges > MaxGraphEdges {
		return nil, resourceLimit("graph_edges", totalEdges, MaxGraphEdges)
	}
	concepts := source.GraphNodes()

	builder := ids.NewSymbolTableBuilder()
	for _, concept := range concepts {
		if err := builder.Add(concept.ID.Symbol()); err != nil {
			return nil, invalid("INVALID_GRAPH", "concepts", err.Error())
		}
	}
	symbols, err := builder.Freeze()
	if err != nil {
		return nil, invalid("INVALID_GRAPH", "concepts", err.Error())
	}
	reader := &GraphReader{
		conceptIDs:   make([]ids.ConceptID, symbols.Len()),
		byConcept:    make(map[ids.ConceptID]NodeID, symbols.Len()),
		kinds:        make([]core.ConceptKind, symbols.Len()),
		parents:      make([]NodeID, symbols.Len()),
		hasParent:    make([]bool, symbols.Len()),
		edges:        make([]Edge, 0, totalEdges),
		outgoing:     make([][]int, symbols.Len()),
		incoming:     make([][]int, symbols.Len()),
		byRelation:   make(map[core.RelationID][]int),
		relationType: make(map[core.RelationID]core.RelationType),
	}
	for _, relation := range source.RelationTypes() {
		reader.relationType[relation.ID] = relation
	}
	for _, concept := range concepts {
		id, ok := symbols.Resolve(concept.ID.Symbol())
		if !ok {
			return nil, invalid("INVALID_GRAPH", "concepts", "compiled concept ID is absent from dense symbol table")
		}
		reader.conceptIDs[id] = concept.ID
		reader.byConcept[concept.ID] = id
		reader.kinds[id] = concept.Kind
		if concept.Parent != nil {
			parent, ok := symbols.Resolve(concept.Parent.Symbol())
			if !ok {
				return nil, invalid("INVALID_GRAPH", "concepts."+concept.ID.String()+".parent", "parent ID is absent from dense symbol table")
			}
			reader.parents[id], reader.hasParent[id] = parent, true
			reader.edges = append(reader.edges, Edge{From: id, To: parent, Strength: core.StrengthRequired, IsParent: true})
		}
		for edgeIndex, definition := range concept.DefinitionEdges {
			target, ok := symbols.Resolve(definition.Target.Symbol())
			if !ok {
				return nil, invalid("INVALID_GRAPH", fmt.Sprintf("concepts.%s.definition_edges[%d].target", concept.ID.String(), edgeIndex), "target ID is absent from dense symbol table")
			}
			reader.edges = append(reader.edges, Edge{From: id, To: target, Relation: definition.Relation, Strength: definition.Strength, Role: definition.Role})
		}
	}
	sort.Slice(reader.edges, func(i, j int) bool { return edgeLess(reader.edges[i], reader.edges[j]) })
	for index, edge := range reader.edges {
		if index > 0 && edgeIdentityEqual(reader.edges[index-1], edge) {
			return nil, invalid("DUPLICATE_EDGE", fmt.Sprintf("edges[%d]", index), "duplicate relation or taxonomy edge")
		}
		reader.outgoing[edge.From] = append(reader.outgoing[edge.From], index)
		reader.incoming[edge.To] = append(reader.incoming[edge.To], index)
		if !edge.IsParent {
			key := edge.Relation
			reader.byRelation[key] = append(reader.byRelation[key], index)
		}
	}
	for node := range reader.incoming {
		sort.Slice(reader.incoming[node], func(i, j int) bool {
			left, right := reader.edges[reader.incoming[node][i]], reader.edges[reader.incoming[node][j]]
			if left.IsParent != right.IsParent {
				return left.IsParent
			}
			if left.Relation != right.Relation {
				return left.Relation.Symbol().Name < right.Relation.Symbol().Name
			}
			if left.From != right.From {
				return left.From < right.From
			}
			return left.Role < right.Role
		})
	}
	for key := range reader.relationType {
		if _, ok := reader.byRelation[key]; !ok {
			reader.byRelation[key] = nil
		}
	}
	return reader, nil
}

// Len returns the number of indexed concepts.
func (reader *GraphReader) Len() int {
	if reader == nil {
		return 0
	}
	return len(reader.conceptIDs)
}

// EdgeCount returns definition edges plus implicit parent taxonomy edges.
func (reader *GraphReader) EdgeCount() int {
	if reader == nil {
		return 0
	}
	return len(reader.edges)
}

// Resolve maps a stable ConceptID to its snapshot-local dense NodeID.
func (reader *GraphReader) Resolve(id ids.ConceptID) (NodeID, bool) {
	if reader == nil {
		return 0, false
	}
	node, ok := reader.byConcept[id]
	return node, ok
}

// Node returns stable identity, kind and parent metadata for a dense ID.
func (reader *GraphReader) Node(id NodeID) (NodeInfo, bool) {
	if reader == nil || uint64(id) >= uint64(len(reader.conceptIDs)) {
		return NodeInfo{}, false
	}
	return NodeInfo{ID: reader.conceptIDs[id], Kind: reader.kinds[id], Parent: reader.parents[id], HasParent: reader.hasParent[id]}, true
}

// Nodes returns all dense IDs in deterministic ascending order.
func (reader *GraphReader) Nodes() []NodeID {
	if reader == nil {
		return nil
	}
	result := make([]NodeID, len(reader.conceptIDs))
	for index := range result {
		result[index] = NodeID(index)
	}
	return result
}

// Outgoing returns a defensive copy of all outgoing definition and parent edges.
func (reader *GraphReader) Outgoing(id NodeID) ([]Edge, bool) {
	if reader == nil || uint64(id) >= uint64(len(reader.outgoing)) {
		return nil, false
	}
	return reader.copyEdges(reader.outgoing[id]), true
}

// Incoming returns a defensive copy of all incoming definition and parent edges.
func (reader *GraphReader) Incoming(id NodeID) ([]Edge, bool) {
	if reader == nil || uint64(id) >= uint64(len(reader.incoming)) {
		return nil, false
	}
	return reader.copyEdges(reader.incoming[id]), true
}

// EdgesByRelation returns relation edges in deterministic source/target order.
// The bool is false only when the relation schema is unknown.
func (reader *GraphReader) EdgesByRelation(relation core.RelationID) ([]Edge, bool) {
	if reader == nil || !relation.Valid() {
		return nil, false
	}
	key := relation
	if _, exists := reader.relationType[key]; !exists {
		return nil, false
	}
	return reader.copyEdges(reader.byRelation[key]), true
}

func (reader *GraphReader) copyEdges(indexes []int) []Edge {
	result := make([]Edge, len(indexes))
	for index, edgeIndex := range indexes {
		result[index] = reader.edges[edgeIndex]
	}
	return result
}

func (reader *GraphReader) relation(id core.RelationID) (core.RelationType, bool) {
	if reader == nil || !id.Valid() {
		return core.RelationType{}, false
	}
	relation, ok := reader.relationType[id]
	return relation, ok
}

func (reader *GraphReader) edgeIndexes(node NodeID, relation core.RelationID) []int {
	if reader == nil || uint64(node) >= uint64(len(reader.outgoing)) {
		return nil
	}
	indexes := reader.outgoing[node]
	relationName := relation.Symbol().Name
	first := sort.Search(len(indexes), func(i int) bool { return reader.edges[indexes[i]].Relation.Symbol().Name >= relationName })
	last := first + sort.Search(len(indexes)-first, func(i int) bool { return reader.edges[indexes[first+i]].Relation.Symbol().Name > relationName })
	return indexes[first:last]
}
