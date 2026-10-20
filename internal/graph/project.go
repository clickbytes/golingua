// Package graph builds bounded immutable indexes over compiled CORE snapshots.
package graph

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxGraphNodes = core.MaxConcepts
	MaxGraphEdges = 1 << 18
	MaxGraphDepth = core.MaxConcepts

	DefaultMaxNodes = 10_000
	DefaultMaxEdges = 20_000
	DefaultMaxDepth = 128
)

// Budget bounds one traversal or projection. All limits are positive and have
// hard ceilings so caller values cannot trigger unbounded preallocation.
type Budget struct {
	MaxNodes int
	MaxEdges int
	MaxDepth int
}

// DefaultBudget returns conservative per-operation traversal limits.
func DefaultBudget() Budget {
	return Budget{MaxNodes: DefaultMaxNodes, MaxEdges: DefaultMaxEdges, MaxDepth: DefaultMaxDepth}
}

// Validate rejects zero, negative and above-ceiling limits.
func (budget Budget) Validate() error {
	if budget.MaxNodes <= 0 || budget.MaxNodes > MaxGraphNodes {
		return invalid("INVALID_GRAPH_BUDGET", "budget.max_nodes", fmt.Sprintf("must be between 1 and %d", MaxGraphNodes))
	}
	if budget.MaxEdges <= 0 || budget.MaxEdges > MaxGraphEdges {
		return invalid("INVALID_GRAPH_BUDGET", "budget.max_edges", fmt.Sprintf("must be between 1 and %d", MaxGraphEdges))
	}
	if budget.MaxDepth <= 0 || budget.MaxDepth > MaxGraphDepth {
		return invalid("INVALID_GRAPH_BUDGET", "budget.max_depth", fmt.Sprintf("must be between 1 and %d", MaxGraphDepth))
	}
	return nil
}

// ProjectionOptions controls which source edges are included and bounds the
// complete projection. Typical edges are omitted unless explicitly requested.
type ProjectionOptions struct {
	Budget         Budget
	IncludeTypical bool
}

// Projection is a complete bounded subgraph rooted at Root. Edges preserve
// required/typical strength; IsParent marks taxonomy edges with no relation ID.
type Projection struct {
	Root  ids.NodeID
	Nodes []ids.NodeID
	Edges []Edge
}

// StatusError preserves the typed outcome and diagnostics associated with a
// graph operation. Resource exhaustion never returns a partial graph as full.
type StatusError struct {
	result status.Result
}

func (failure *StatusError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "graph operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed status result.
func (failure *StatusError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func invalid(code, path, message string) error {
	return statusError(status.OutcomeInvalid, code, message, path)
}

func unknown(code, path, message string) error {
	return statusError(status.OutcomeUnknown, code, message, path)
}

func resourceLimit(name string, value, maximum int) error {
	return statusError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("%s limit exceeded: %d > %d", name, value, maximum), "budget."+name)
}

func statusError(kind status.OutcomeKind, code, message, path string) error {
	diagnostic := status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityError}
	return &StatusError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func edgeLess(left, right Edge) bool {
	if left.From != right.From {
		return left.From < right.From
	}
	if left.IsParent != right.IsParent {
		return left.IsParent
	}
	if left.Relation != right.Relation {
		return left.Relation.Symbol().Name < right.Relation.Symbol().Name
	}
	if left.To != right.To {
		return left.To < right.To
	}
	if left.Role != right.Role {
		return left.Role < right.Role
	}
	return left.Strength < right.Strength
}

func edgeIdentityEqual(left, right Edge) bool {
	return left.From == right.From && left.To == right.To && left.IsParent == right.IsParent && left.Relation == right.Relation && left.Role == right.Role
}

func sortedUniqueNodes(nodes []ids.NodeID) []ids.NodeID {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
	if len(nodes) < 2 {
		return nodes
	}
	write := 1
	for read := 1; read < len(nodes); read++ {
		if nodes[read] != nodes[write-1] {
			nodes[write] = nodes[read]
			write++
		}
	}
	return nodes[:write]
}
