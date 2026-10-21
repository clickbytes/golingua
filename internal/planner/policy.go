package planner

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/reason"
	"github.com/clickbytes/golingua/internal/status"
)

// ChooseGoal is deterministic and uses only typed query outcomes plus an
// explicit clarification gap. Missing facts never license a positive answer.
func ChooseGoal(input DecisionInput) (Decision, error) {
	if input.Result == nil {
		return Decision{}, plannerError(status.OutcomeInvalid, "QUERY_RESULT_REQUIRED", "query_result", "content planning requires a typed query result")
	}
	result := input.Result
	if result.Outcome == status.OutcomeInvalid || result.Outcome == status.OutcomeError || result.Outcome == status.OutcomeResourceLimit {
		return Decision{}, plannerError(result.Outcome, "QUERY_NOT_COMPLETE", "query_result", "query result cannot be used as a complete content plan")
	}
	switch {
	case result.Outcome == status.OutcomeUnsupported:
		return decision(GoalRefuse, meaning.ActRefusal, status.OutcomeUnsupported), nil
	case result.Truth == reason.QuestionConflict:
		return decision(GoalExplain, meaning.ActExplanation, status.OutcomeAmbiguous), nil
	case result.Truth == reason.QuestionAmbiguous:
		if validClarification(input.Clarification, ClarifyChoice) {
			return decision(GoalClarify, meaning.ActQuestion, status.OutcomeAmbiguous), nil
		}
		return decision(GoalExplain, meaning.ActExplanation, status.OutcomeAmbiguous), nil
	case result.Truth == reason.QuestionUnknown:
		if validClarification(input.Clarification, ClarifyReference) {
			return decision(GoalClarify, meaning.ActQuestion, status.OutcomeUnknown), nil
		}
		return decision(GoalRefuse, meaning.ActRefusal, status.OutcomeUnknown), nil
	case result.Kind == reason.QuestionYesNo && (result.Truth == reason.QuestionTrue || result.Truth == reason.QuestionFalse):
		return decision(GoalAnswerYesNo, meaning.ActStatement, status.OutcomeOK), nil
	case result.Kind == reason.QuestionWH && result.Truth == reason.QuestionTrue && len(result.Candidates) == 1 && result.Candidates[0].Truth == reason.QuestionTrue:
		return decision(GoalAnswerWH, meaning.ActStatement, status.OutcomeOK), nil
	default:
		return decision(GoalRefuse, meaning.ActRefusal, status.OutcomeUnsupported), nil
	}
}

func decision(goal DialogueGoal, act meaning.DialogueAct, outcome status.OutcomeKind) Decision {
	return Decision{Goal: goal, Act: act, Outcome: outcome}
}

func validClarification(gap *ClarificationGap, kind ClarificationKind) bool {
	if gap == nil || gap.Kind != kind {
		return false
	}
	if kind == ClarifyReference {
		return validIdentifier(gap.NodeID) && validRole(gap.Role)
	}
	return kind == ClarifyChoice && validRole(gap.Role)
}

func validIdentifier(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func validRole(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
