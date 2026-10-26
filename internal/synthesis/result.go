package synthesis

import (
	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// GoalObligation reports which bounded construction steps were associated
// with one input constraint. Its status records construction provenance only;
// it is not an independent proof or a fact about the world.
type GoalObligation struct {
	ConstraintID string         `json:"constraint_id"`
	Kind         ConstraintKind `json:"kind"`
	Status       string         `json:"status"`
	StepIDs      []string       `json:"step_ids"`
	Fragments    []string       `json:"fragments"`
}

// Result contains a checked representation only when the goal was fully
// specified and construction completed. An incomplete result has no
// expression, checked document, or construction trace.
type Result struct {
	Status            status.OutcomeKind         `json:"status"`
	Expression        *algebra.Term              `json:"expression,omitempty"`
	Checked           *algebra.CheckedExpression `json:"checked,omitempty"`
	Interface         []algebra.PortBinding      `json:"interface,omitempty"`
	CanonicalDocument *algebra.Document          `json:"canonical_document,omitempty"`
	Trace             []algebra.BuildTrace       `json:"trace"`
	Unresolved        []Unresolved               `json:"unresolved"`
	Obligations       []GoalObligation           `json:"obligations"`
}
