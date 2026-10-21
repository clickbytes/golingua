// Package planner chooses response content without selecting surface language.
package planner

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/reason"
	"github.com/clickbytes/golingua/internal/status"
)

type DialogueGoal string

const (
	GoalAnswerYesNo DialogueGoal = "answer_yes_no"
	GoalAnswerWH    DialogueGoal = "answer_wh"
	GoalClarify     DialogueGoal = "clarify"
	GoalRefuse      DialogueGoal = "refuse"
	GoalExplain     DialogueGoal = "explain"
)

type ClarificationKind string

const (
	ClarifyReference ClarificationKind = "unresolved_reference"
	ClarifyChoice    ClarificationKind = "answer_choice"
)

// ClarificationGap is explicit evidence that a user response can resolve one
// known missing identity or answer choice. The planner never invents one.
type ClarificationGap struct {
	Kind   ClarificationKind `json:"kind"`
	NodeID string            `json:"node_id,omitempty"`
	Role   string            `json:"role,omitempty"`
}

// DecisionInput is request-scoped. Query must have been produced by the typed
// reasoner; Clarification is optional and must come from an explicit resolver.
type DecisionInput struct {
	Query         reason.Question
	Result        *reason.QuestionResult
	Clarification *ClarificationGap
}

type Decision struct {
	Goal    DialogueGoal
	Act     meaning.DialogueAct
	Outcome status.OutcomeKind
}
