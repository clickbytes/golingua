package graph

import (
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
)

// TraversalOptions bounds work and controls whether typical edges are visible.
type TraversalOptions struct {
	Budget         Budget
	IncludeTypical bool
}

// Ancestors returns parent taxonomy nodes from nearest parent to root.
func (reader *GraphReader) Ancestors(root ids.ConceptID, budget Budget) ([]NodeID, error) {
	if err := budget.Validate(); err != nil {
		return nil, err
	}
	start, ok := reader.Resolve(root)
	if !ok {
		return nil, unknown("UNKNOWN_NODE", "root", "concept root does not exist in this graph snapshot")
	}
	result := make([]NodeID, 0, minInt(budget.MaxDepth, 32))
	current := start
	depth, edges := 0, 0
	for reader.hasParent[current] {
		parent := reader.parents[current]
		depth++
		edges++
		if depth > budget.MaxDepth {
			return nil, resourceLimit("graph_depth", depth, budget.MaxDepth)
		}
		if edges > budget.MaxEdges {
			return nil, resourceLimit("graph_edges", edges, budget.MaxEdges)
		}
		if len(result)+1 >= budget.MaxNodes {
			return nil, resourceLimit("graph_nodes", len(result)+2, budget.MaxNodes)
		}
		result = append(result, parent)
		current = parent
	}
	return result, nil
}

// IsA reports whether ancestor is the same concept as, or a taxonomy ancestor
// of, child. It follows only the acyclic typed parent hierarchy.
func (reader *GraphReader) IsA(child, ancestor ids.ConceptID, budget Budget) (bool, error) {
	if err := budget.Validate(); err != nil {
		return false, err
	}
	childNode, ok := reader.Resolve(child)
	if !ok {
		return false, unknown("UNKNOWN_NODE", "child", "child concept does not exist in this graph snapshot")
	}
	ancestorNode, ok := reader.Resolve(ancestor)
	if !ok {
		return false, unknown("UNKNOWN_NODE", "ancestor", "ancestor concept does not exist in this graph snapshot")
	}
	if childNode == ancestorNode {
		if err := budget.Validate(); err != nil {
			return false, err
		}
		return true, nil
	}
	ancestors, err := reader.Ancestors(child, budget)
	if err != nil {
		return false, err
	}
	for _, candidate := range ancestors {
		if candidate == ancestorNode {
			return true, nil
		}
	}
	return false, nil
}

// Targets returns direct targets of a relation, or its transitive closure only
// when the relation schema explicitly declares Transitive=true. Typical edges
// are skipped unless IncludeTypical is set.
func (reader *GraphReader) Targets(root ids.ConceptID, relationID core.RelationID, options TraversalOptions) ([]NodeID, error) {
	if err := options.Budget.Validate(); err != nil {
		return nil, err
	}
	if !relationID.Valid() {
		return nil, invalid("INVALID_RELATION_ID", "relation", "relation ID is empty or invalid")
	}
	start, ok := reader.Resolve(root)
	if !ok {
		return nil, unknown("UNKNOWN_NODE", "root", "concept root does not exist in this graph snapshot")
	}
	relation, ok := reader.relation(relationID)
	if !ok {
		return nil, unknown("UNKNOWN_RELATION", "relation", "relation schema does not exist in this graph snapshot")
	}
	if !relation.Transitive {
		return reader.directTargets(start, relationID, options)
	}
	return reader.transitiveTargets(start, relationID, options)
}

func (reader *GraphReader) directTargets(start NodeID, relation core.RelationID, options TraversalOptions) ([]NodeID, error) {
	indexes := reader.edgeIndexes(start, relation)
	targetSet := make(map[NodeID]struct{})
	nodeSet := map[NodeID]struct{}{start: {}}
	considered := 0
	for _, index := range indexes {
		edge := reader.edges[index]
		if !includeEdge(edge, options.IncludeTypical) {
			continue
		}
		considered++
		if considered > options.Budget.MaxEdges {
			return nil, resourceLimit("graph_edges", considered, options.Budget.MaxEdges)
		}
		if _, exists := nodeSet[edge.To]; !exists && len(nodeSet) >= options.Budget.MaxNodes {
			return nil, resourceLimit("graph_nodes", len(nodeSet)+1, options.Budget.MaxNodes)
		}
		nodeSet[edge.To] = struct{}{}
		targetSet[edge.To] = struct{}{}
	}
	result := make([]NodeID, 0, len(targetSet))
	for target := range targetSet {
		result = append(result, target)
	}
	return sortedUniqueNodes(result), nil
}

func (reader *GraphReader) transitiveTargets(start NodeID, relation core.RelationID, options TraversalOptions) ([]NodeID, error) {
	visited := map[NodeID]struct{}{start: {}}
	targetSet := make(map[NodeID]struct{})
	depths := map[NodeID]int{start: 0}
	queue := make([]NodeID, 1, minInt(options.Budget.MaxNodes, 1024))
	queue[0] = start
	considered, head := 0, 0
	for head < len(queue) {
		current := queue[head]
		head++
		for _, index := range reader.edgeIndexes(current, relation) {
			edge := reader.edges[index]
			if !includeEdge(edge, options.IncludeTypical) {
				continue
			}
			considered++
			if considered > options.Budget.MaxEdges {
				return nil, resourceLimit("graph_edges", considered, options.Budget.MaxEdges)
			}
			targetSet[edge.To] = struct{}{}
			if _, exists := visited[edge.To]; exists {
				continue
			}
			depth := depths[current] + 1
			if depth > options.Budget.MaxDepth {
				return nil, resourceLimit("graph_depth", depth, options.Budget.MaxDepth)
			}
			if len(visited) >= options.Budget.MaxNodes {
				return nil, resourceLimit("graph_nodes", len(visited)+1, options.Budget.MaxNodes)
			}
			visited[edge.To] = struct{}{}
			depths[edge.To] = depth
			queue = append(queue, edge.To)
		}
	}
	result := make([]NodeID, 0, len(targetSet))
	for target := range targetSet {
		result = append(result, target)
	}
	return sortedUniqueNodes(result), nil
}

// ProjectConcept returns the full reachable projection of root. Parent links
// and selected relation edges are included. On any budget exhaustion, it
// returns no projection and a typed resource_limit error.
func (reader *GraphReader) ProjectConcept(root ids.ConceptID, options ProjectionOptions) (Projection, error) {
	if err := options.Budget.Validate(); err != nil {
		return Projection{}, err
	}
	start, ok := reader.Resolve(root)
	if !ok {
		return Projection{}, unknown("UNKNOWN_NODE", "root", "concept root does not exist in this graph snapshot")
	}
	visited := map[NodeID]struct{}{start: {}}
	depths := map[NodeID]int{start: 0}
	queue := make([]NodeID, 1, minInt(options.Budget.MaxNodes, 1024))
	queue[0] = start
	projectedEdges := make([]Edge, 0, minInt(len(reader.edges), options.Budget.MaxEdges))
	considered, head := 0, 0
	for head < len(queue) {
		current := queue[head]
		head++
		for _, index := range reader.outgoing[current] {
			edge := reader.edges[index]
			if !includeEdge(edge, options.IncludeTypical) {
				continue
			}
			considered++
			if considered > options.Budget.MaxEdges {
				return Projection{}, resourceLimit("graph_edges", considered, options.Budget.MaxEdges)
			}
			projectedEdges = append(projectedEdges, edge)
			if _, exists := visited[edge.To]; exists {
				continue
			}
			depth := depths[current] + 1
			if depth > options.Budget.MaxDepth {
				return Projection{}, resourceLimit("graph_depth", depth, options.Budget.MaxDepth)
			}
			if len(visited) >= options.Budget.MaxNodes {
				return Projection{}, resourceLimit("graph_nodes", len(visited)+1, options.Budget.MaxNodes)
			}
			visited[edge.To] = struct{}{}
			depths[edge.To] = depth
			queue = append(queue, edge.To)
		}
	}
	nodes := make([]NodeID, 0, len(visited))
	for node := range visited {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
	sort.Slice(projectedEdges, func(i, j int) bool { return edgeLess(projectedEdges[i], projectedEdges[j]) })
	return Projection{Root: start, Nodes: nodes, Edges: projectedEdges}, nil
}

func includeEdge(edge Edge, includeTypical bool) bool {
	return includeTypical || edge.Strength != core.StrengthTypical
}
