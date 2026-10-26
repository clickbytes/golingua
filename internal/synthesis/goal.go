// Package synthesis contains bounded, declarative goal types. A goal describes
// a construction problem; it is not itself a term or a construction recipe.
package synthesis

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	GoalSchemaVersion = 1
	GoalProfile       = "positive-facts-v1"

	DefaultGoalMaxAtoms = 16
	MaxGoalAtoms        = 16
	DefaultGoalMaxSteps = 64
	MaxGoalSteps        = 64
	MaxGoalConstraints  = 32
	MaxGoalPorts        = 16
	MaxGoalSources      = 16
	MaxGoalInputBytes   = 1 << 20
)

const (
	GoalOperatorRelation = "relation"
	GoalOperatorJoin     = "join"
	GoalOperatorRestrict = "restrict"
	GoalOperatorBind     = "bind"
	GoalOperatorPermute  = "permute"
	GoalOperatorProject  = "project"
)

type ConstraintKind string

const (
	ConstraintPath  ConstraintKind = "path"
	ConstraintUnary ConstraintKind = "unary"
)

// GoalValue is a typed, request-local constant or unresolved typed hole.
// Later validation accepts only Ref or Hole here. A Hole remains unresolved
// and cannot be treated as an operator input.
type GoalValue struct {
	Kind algebra.TermKind `json:"kind"`
	ID   string           `json:"id"`
	Type *algebra.Type    `json:"type"`
}

// GoalPort declares one ordered candidate output port. A nil Binding leaves
// the port open unless GoalSpec.ProjectPorts closes it; a Ref binding closes
// it with an explicitly declared reference.
type GoalPort struct {
	ID      string       `json:"id"`
	Type    algebra.Type `json:"type"`
	Binding *GoalValue   `json:"binding,omitempty"`
}

// Constraint describes one closed-profile structural obligation. A path
// constraint counts relation occurrences along one binary connection; a unary
// constraint restricts one port. These fields do not represent arbitrary logic.
type Constraint struct {
	ID                string         `json:"id"`
	Kind              ConstraintKind `json:"kind"`
	Symbol            string         `json:"symbol,omitempty"`
	FromPort          string         `json:"from_port,omitempty"`
	ToPort            string         `json:"to_port,omitempty"`
	FromRole          string         `json:"from_role,omitempty"`
	ToRole            string         `json:"to_role,omitempty"`
	AllowedAtomCounts []int          `json:"allowed_atom_counts,omitempty"`
	Port              string         `json:"port,omitempty"`
}

// GoalLimits bounds the construction profile and the algebra checker used for
// its eventual result. Atom count includes path relation occurrences and
// unary atoms; it does not increase the separate algebra term limits.
type GoalLimits struct {
	MaxAtoms   int            `json:"max_atoms,omitempty"`
	MaxSteps   int            `json:"max_steps,omitempty"`
	TermLimits algebra.Limits `json:"term_limits,omitempty"`
}

// GoalSpec states desired type, open ports, permitted generic operators, and
// explicit obligations. It deliberately contains no final AST or named
// derived-concept selection.
type GoalSpec struct {
	SchemaVersion int                   `json:"schema_version"`
	Profile       string                `json:"profile"`
	ContextID     string                `json:"context_id"`
	ExpectedType  algebra.Type          `json:"expected_type"`
	Ports         []GoalPort            `json:"ports"`
	References    []algebra.Declaration `json:"references"`
	// ProjectPorts is an optional set of ports to existentially close after
	// constraints and bindings have been applied.
	ProjectPorts     []string               `json:"project_ports,omitempty"`
	AllowedOperators []string               `json:"allowed_operators"`
	Constraints      []Constraint           `json:"constraints"`
	Provenance       []provenance.SourceRef `json:"provenance"`
	Limits           GoalLimits             `json:"limits"`
}

// NormalizeGoalLimits applies finite defaults and hard caps. Detailed goal
// structure and operator admission are validated by a later synthesis stage.
func NormalizeGoalLimits(requested GoalLimits) (GoalLimits, error) {
	maxAtoms, err := normalizeGoalBound(requested.MaxAtoms, DefaultGoalMaxAtoms, MaxGoalAtoms, "limits.max_atoms")
	if err != nil {
		return GoalLimits{}, err
	}
	maxSteps, err := normalizeGoalBound(requested.MaxSteps, DefaultGoalMaxSteps, MaxGoalSteps, "limits.max_steps")
	if err != nil {
		return GoalLimits{}, err
	}
	termLimits, err := algebra.NormalizeLimits(requested.TermLimits)
	if err != nil {
		return GoalLimits{}, err
	}
	return GoalLimits{MaxAtoms: maxAtoms, MaxSteps: maxSteps, TermLimits: termLimits}, nil
}

func normalizeGoalBound(value int, fallback, cap int, path string) (int, error) {
	if value == 0 {
		return fallback, nil
	}
	if value < 0 || value > cap {
		return 0, &algebra.Error{
			Status:  status.OutcomeInvalid,
			Code:    "INVALID_GOAL_LIMIT",
			Path:    path,
			Message: fmt.Sprintf("must be zero (default) or between 1 and %d", cap),
		}
	}
	return value, nil
}
