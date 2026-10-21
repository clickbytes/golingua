// Package dialogue connects planned response meanings to licensed realization.
package dialogue

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/planner"
	"github.com/clickbytes/golingua/internal/status"
)

// SupportedAct reports whether this bridge can currently pass an act through
// the lexical construction pipeline without changing its dialogue force.
func SupportedAct(plan planner.ContentPlan) bool {
	if plan.Goal != planner.GoalAnswerYesNo && plan.Goal != planner.GoalAnswerWH {
		return false
	}
	if plan.Act != meaning.ActStatement || len(plan.ResponseMeaning.Roots) != 1 {
		return false
	}
	root, ok := findNode(plan.ResponseMeaning, plan.ResponseMeaning.Roots[0])
	return ok && root.Kind == meaning.KindDiscourse && root.Act == meaning.ActStatement && len(root.Content) == 1
}

func unsupportedAct() status.Diagnostic {
	return status.Diagnostic{
		Code: "DIALOGUE_ACT_UNSUPPORTED", Path: "response.act",
		Message:  "target language has no verified dialogue construction for this response act",
		Severity: status.SeverityError,
	}
}
