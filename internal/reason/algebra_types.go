package reason

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	AlgebraQuerySchemaVersion = 1
	AlgebraQueryProfile       = "positive-facts-v1"

	AlgebraSelectionLambda    = "lambda"
	AlgebraSelectionParameter = "parameter"
)

// AlgebraSelection selects one explicit query answer variable. Lambda indices
// refer to the leading-lambda order after bounded beta lowering; parameter IDs
// are looked up only in the parameter namespace.
type AlgebraSelection struct {
	Kind  string `json:"kind"`
	Index *int   `json:"index,omitempty"`
	ID    string `json:"id,omitempty"`
	As    string `json:"as"`
}

// AlgebraQueryRequest is the finite request envelope for evaluating an
// expression over one caller-supplied, request-scoped fact view. It declares
// selection and context, not truth or a query plan.
type AlgebraQueryRequest struct {
	SchemaVersion int                `json:"schema_version"`
	Profile       string             `json:"profile"`
	ContextID     string             `json:"context_id"`
	At            string             `json:"at,omitempty"`
	Select        []AlgebraSelection `json:"select"`
}

const (
	DefaultAlgebraMaxPlanAtoms        = 64
	MaxAlgebraPlanAtoms               = 128
	DefaultAlgebraMaxVariables        = 128
	MaxAlgebraVariables               = 512
	DefaultAlgebraMaxIntermediateRows = 256
	MaxAlgebraIntermediateRows        = 4096
	DefaultAlgebraMaxResults          = 64
	MaxAlgebraResults                 = 512
	DefaultAlgebraMaxProofNodes       = 4096
	MaxAlgebraProofNodes              = 16384
	DefaultAlgebraMaxProofsPerAnswer  = 8
	MaxAlgebraProofsPerAnswer         = 64
	DefaultAlgebraMaxProofDepth       = 128
	MaxAlgebraProofDepth              = 128
	DefaultAlgebraMaxWorkUnits        = 1_048_576
	MaxAlgebraWorkUnits               = 16_777_216
	DefaultAlgebraMaxOutputBytes      = 8 << 20
	MaxAlgebraOutputBytes             = 16 << 20
	DefaultAlgebraTimeoutMillis       = 2000
	MaxAlgebraTimeoutMillis           = 30_000
)

// AlgebraEvaluationLimits bounds plan compilation, joins, result/proof
// materialization, and cooperative cancellation. Zero selects a finite
// default; it never means unlimited.
type AlgebraEvaluationLimits struct {
	MaxPlanAtoms        int            `json:"max_plan_atoms,omitempty"`
	MaxVariables        int            `json:"max_variables,omitempty"`
	MaxIntermediateRows int            `json:"max_intermediate_rows,omitempty"`
	MaxResults          int            `json:"max_results,omitempty"`
	MaxProofNodes       int            `json:"max_proof_nodes,omitempty"`
	MaxProofsPerAnswer  int            `json:"max_proofs_per_answer,omitempty"`
	MaxProofDepth       int            `json:"max_proof_depth,omitempty"`
	MaxWorkUnits        int            `json:"max_work_units,omitempty"`
	MaxOutputBytes      int            `json:"max_output_bytes,omitempty"`
	TimeoutMillis       int            `json:"timeout_millis,omitempty"`
	TermLimits          algebra.Limits `json:"term_limits,omitempty"`
}

// DefaultAlgebraEvaluationLimits returns the finite evaluation defaults.
func DefaultAlgebraEvaluationLimits() AlgebraEvaluationLimits {
	return AlgebraEvaluationLimits{
		MaxPlanAtoms: DefaultAlgebraMaxPlanAtoms, MaxVariables: DefaultAlgebraMaxVariables,
		MaxIntermediateRows: DefaultAlgebraMaxIntermediateRows, MaxResults: DefaultAlgebraMaxResults,
		MaxProofNodes: DefaultAlgebraMaxProofNodes, MaxProofsPerAnswer: DefaultAlgebraMaxProofsPerAnswer,
		MaxProofDepth: DefaultAlgebraMaxProofDepth, MaxWorkUnits: DefaultAlgebraMaxWorkUnits,
		MaxOutputBytes: DefaultAlgebraMaxOutputBytes, TimeoutMillis: DefaultAlgebraTimeoutMillis,
		TermLimits: algebra.DefaultLimits(),
	}
}

// NormalizeAlgebraEvaluationLimits applies finite defaults and hard caps.
// Shared algebra term-limit errors are returned unchanged.
func NormalizeAlgebraEvaluationLimits(requested AlgebraEvaluationLimits) (AlgebraEvaluationLimits, error) {
	defaults := DefaultAlgebraEvaluationLimits()
	values := []struct {
		name  string
		value *int
		def   int
		cap   int
	}{
		{"max_plan_atoms", &requested.MaxPlanAtoms, defaults.MaxPlanAtoms, MaxAlgebraPlanAtoms},
		{"max_variables", &requested.MaxVariables, defaults.MaxVariables, MaxAlgebraVariables},
		{"max_intermediate_rows", &requested.MaxIntermediateRows, defaults.MaxIntermediateRows, MaxAlgebraIntermediateRows},
		{"max_results", &requested.MaxResults, defaults.MaxResults, MaxAlgebraResults},
		{"max_proof_nodes", &requested.MaxProofNodes, defaults.MaxProofNodes, MaxAlgebraProofNodes},
		{"max_proofs_per_answer", &requested.MaxProofsPerAnswer, defaults.MaxProofsPerAnswer, MaxAlgebraProofsPerAnswer},
		{"max_proof_depth", &requested.MaxProofDepth, defaults.MaxProofDepth, MaxAlgebraProofDepth},
		{"max_work_units", &requested.MaxWorkUnits, defaults.MaxWorkUnits, MaxAlgebraWorkUnits},
		{"max_output_bytes", &requested.MaxOutputBytes, defaults.MaxOutputBytes, MaxAlgebraOutputBytes},
		{"timeout_millis", &requested.TimeoutMillis, defaults.TimeoutMillis, MaxAlgebraTimeoutMillis},
	}
	for _, item := range values {
		if *item.value == 0 {
			*item.value = item.def
			continue
		}
		if *item.value < 0 || *item.value > item.cap {
			return AlgebraEvaluationLimits{}, &algebra.Error{
				Status: status.OutcomeInvalid, Code: "INVALID_EVALUATION_LIMIT",
				Path:    "limits." + item.name,
				Message: fmt.Sprintf("must be zero (default) or between 1 and %d", item.cap),
			}
		}
	}
	termLimits, err := algebra.NormalizeLimits(requested.TermLimits)
	if err != nil {
		return AlgebraEvaluationLimits{}, err
	}
	requested.TermLimits = termLimits
	return requested, nil
}

// AlgebraVariableKey keeps variable namespaces distinct even when their IDs
// have identical spellings.
type AlgebraVariableKey struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// AlgebraAnswerPort describes one explicitly selected, typed output variable.
type AlgebraAnswerPort struct {
	As       string             `json:"as"`
	Variable AlgebraVariableKey `json:"variable"`
	Type     algebra.Type       `json:"type"`
}

// AlgebraEvaluationUsage counts bounded logical work and retained output. It
// is not a measurement of CPU time, energy, or wall-clock cost.
type AlgebraEvaluationUsage struct {
	WorkUnits        int `json:"work_units"`
	JoinAttempts     int `json:"join_attempts"`
	IndexedFacts     int `json:"indexed_facts"`
	IntermediateRows int `json:"intermediate_rows"`
	Results          int `json:"results"`
	ProofNodes       int `json:"proof_nodes"`
	Proofs           int `json:"proofs"`
	OutputBytes      int `json:"output_bytes"`
}

// AlgebraEvaluationStopReason names an operational stop, never a truth value.
type AlgebraEvaluationStopReason string

const (
	AlgebraStopMaxWorkUnits        AlgebraEvaluationStopReason = "max_work_units"
	AlgebraStopMaxIntermediateRows AlgebraEvaluationStopReason = "max_intermediate_rows"
	AlgebraStopMaxResults          AlgebraEvaluationStopReason = "max_results"
	AlgebraStopMaxProofNodes       AlgebraEvaluationStopReason = "max_proof_nodes"
	AlgebraStopMaxProofsPerAnswer  AlgebraEvaluationStopReason = "max_proofs_per_answer"
	AlgebraStopMaxProofDepth       AlgebraEvaluationStopReason = "max_proof_depth"
	AlgebraStopMaxOutputBytes      AlgebraEvaluationStopReason = "max_output_bytes"
	AlgebraStopTimeout             AlgebraEvaluationStopReason = "timeout"
	AlgebraStopCancelled           AlgebraEvaluationStopReason = "cancelled"
	AlgebraStopLoweringBudget      AlgebraEvaluationStopReason = "lowering_budget"
)
