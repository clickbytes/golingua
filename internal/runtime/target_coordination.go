package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func (pipeline *TargetPipeline) realizeCoordinated(ctx context.Context, source semantic.MeaningCandidate) (engine.SurfaceResult, bool, error) {
	if source.Meaning == nil || pipeline.ClauseCoordination == nil {
		return engine.SurfaceResult{}, false, nil
	}
	discourse, events, ok := coordinationEvents(*source.Meaning)
	if !ok {
		return engine.SurfaceResult{}, false, nil
	}
	if source.QuestionAnswer != nil || len(source.Evidence.Records) != 0 {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeUnsupported, "COORDINATION_SOURCE_EVIDENCE_UNSUPPORTED", "source", "coordinated target requires a statement Meaning without query answer or source-aligned evidence")
	}
	_ = discourse
	if err := meaning.ValidateMeaning(*source.Meaning, pipeline.Core, meaning.DefaultBudget()); err != nil {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeInvalid, "INVALID_COORDINATION_MEANING", "meaning", err.Error())
	}
	if len(source.Contexts) != 1 {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeInvalid, "COORDINATION_CONTEXT_COUNT", "source.contexts", "exactly one matching request context is required")
	}
	equal, eqErr := meaning.CanonicalEqual(source.Contexts[0].Meaning, *source.Meaning, meaning.DefaultBudget())
	if eqErr != nil || !equal {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeInvalid, "COORDINATION_CONTEXT_MISMATCH", "source.contexts", "request context meaning must exactly match the target meaning")
	}
	allowed := map[meaning.NodeID]bool{discourse.ID: true, events[0]: true, events[1]: true}
	nodeMap := map[meaning.NodeID]meaning.Node{}
	for _, node := range source.Meaning.Nodes {
		nodeMap[node.ID] = node
	}
	for _, id := range events {
		n := nodeMap[id]
		for _, p := range n.Participants {
			allowed[p.Target] = true
		}
	}
	if len(allowed) != len(source.Meaning.Nodes) {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeUnsupported, "COORDINATION_EXTRA_MEANING", "meaning", "coordinated statement contains unsupported extra nodes")
	}
	for _, n := range source.Meaning.Nodes {
		if !allowed[n.ID] || n.Kind != meaning.KindDiscourse && n.Kind != meaning.KindEvent && n.Kind != meaning.KindReferent {
			return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeUnsupported, "COORDINATION_EXTRA_MEANING", "meaning", "coordinated statement contains unsupported nodes or operators")
		}
	}
	if pipeline.ClauseCoordination.Language != pipeline.Language || pipeline.ClauseCoordination.ID == "" {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeUnsupported, "COORDINATION_PLAN_UNAVAILABLE", "target.coordination", "language-local coordination plan is invalid")
	}
	if events[0] == events[1] {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeAmbiguous, "REPEATED_EVENT_IDENTITY", "meaning.discourse.content", "coordinated clauses must have distinct events")
	}
	views := make([]semantic.MeaningCandidate, 2)
	for i, eventID := range events {
		graph, ctxMeaning, err := coordinationClauseView(*source.Meaning, source.Contexts, eventID)
		if err != nil {
			return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeInvalid, "INVALID_COORDINATION_MEANING", "meaning", "coordinated clause projection failed validation")
		}
		views[i] = source
		views[i].Meaning = &graph
		views[i].Contexts = []semantic.MeaningContext{{Meaning: graph, Context: ctxMeaning}}
		views[i].QuestionAnswer = nil
		views[i].Evidence.Records = []semantic.EvidenceRecord{}
	}
	left, err := pipeline.Realize(ctx, views[0])
	if err != nil {
		return engine.SurfaceResult{}, true, err
	}
	right, err := pipeline.Realize(ctx, views[1])
	if err != nil {
		return engine.SurfaceResult{}, true, err
	}
	result := engine.SurfaceResult{Outcome: combineTargetOutcomes(left.Outcome, right.Outcome), Diagnostics: appendTargetDiagnostics(left.Diagnostics, right.Diagnostics...), Alternatives: []engine.SurfaceAlternative{}}
	for _, a := range left.Alternatives {
		if a.Text == "" || a.Outcome != status.OutcomeOK && a.Outcome != status.OutcomeIncomplete {
			continue
		}
		for _, b := range right.Alternatives {
			if b.Text == "" || b.Outcome != status.OutcomeOK && b.Outcome != status.OutcomeIncomplete {
				continue
			}
			if err := ctx.Err(); err != nil {
				return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
			}
			if len(result.Alternatives) >= engine.MaxTranslationAlternatives {
				return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "target.coordination.alternatives", "coordinated cross-product exceeds the alternative limit")
			}
			text := joinCoordinationClauses(a.Text, b.Text, *pipeline.ClauseCoordination)
			if text == "" {
				continue
			}
			outcome := combineTargetOutcomes(a.Outcome, b.Outcome)
			id := coordinatedAlternativeID(pipeline.ClauseCoordination.ID, a.ID, b.ID)
			records := append([]semantic.EvidenceRecord{}, a.Evidence.Records...)
			records = append(records, b.Evidence.Records...)
			trace := semantic.EvidenceTrace{AlternativeID: id, Outcome: outcome, Records: records}
			if outcome != status.OutcomeOK {
				trace.Diagnostics = appendTargetDiagnostics(a.Evidence.Diagnostics, b.Evidence.Diagnostics...)
			}
			result.Alternatives = append(result.Alternatives, engine.SurfaceAlternative{ID: id, Outcome: outcome, Text: text, Evidence: trace, Assumptions: append(append([]realizer.FieldAssumption{}, a.Assumptions...), b.Assumptions...), Losses: append(append([]realizer.FieldLoss{}, a.Losses...), b.Losses...), Diagnostics: appendTargetDiagnostics(a.Diagnostics, b.Diagnostics...)})
		}
	}
	if len(result.Alternatives) == 0 {
		return engine.SurfaceResult{}, true, targetPipelineError(status.OutcomeUnsupported, "COORDINATED_CLAUSE_INCOMPLETE", "target.coordination", "both clauses must have text-bearing complete alternatives")
	}
	return result, true, nil
}

func coordinationEvents(graph meaning.Meaning) (meaning.Node, [2]meaning.NodeID, bool) {
	nodes := map[meaning.NodeID]meaning.Node{}
	for _, n := range graph.Nodes {
		nodes[n.ID] = n
	}
	if len(graph.Roots) != 1 {
		return meaning.Node{}, [2]meaning.NodeID{}, false
	}
	d, ok := nodes[graph.Roots[0]]
	if !ok || d.Kind != meaning.KindDiscourse || d.Act != meaning.ActStatement || len(d.Content) != 2 {
		return meaning.Node{}, [2]meaning.NodeID{}, false
	}
	var ids [2]meaning.NodeID
	for i, id := range d.Content {
		n, exists := nodes[id]
		if !exists || n.Kind != meaning.KindEvent {
			return meaning.Node{}, ids, false
		}
		ids[i] = id
	}
	return d, ids, true
}

func coordinationClauseView(graph meaning.Meaning, contexts []semantic.MeaningContext, eventID meaning.NodeID) (meaning.Meaning, meaning.RequestContext, error) {
	nodes := map[meaning.NodeID]meaning.Node{}
	for _, n := range graph.Nodes {
		nodes[n.ID] = n
	}
	event, ok := nodes[eventID]
	if !ok {
		return meaning.Meaning{}, meaning.RequestContext{}, fmt.Errorf("event node is missing")
	}
	keep := map[meaning.NodeID]bool{eventID: true}
	for _, p := range event.Participants {
		keep[p.Target] = true
	}
	out := []meaning.Node{}
	for _, n := range graph.Nodes {
		if n.Kind == meaning.KindDiscourse {
			n.Content = []meaning.NodeID{eventID}
			keep[n.ID] = true
		}
		if keep[n.ID] || n.Kind == meaning.KindDiscourse {
			out = append(out, n)
		}
	}
	view, err := meaning.Build([]meaning.NodeID{graph.Roots[0]}, out, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, meaning.RequestContext{}, err
	}
	var sourceCtx meaning.RequestContext
	for _, c := range contexts {
		equal, eqErr := meaning.CanonicalEqual(c.Meaning, graph, meaning.DefaultBudget())
		if eqErr == nil && equal {
			sourceCtx = c.Context
			break
		}
	}
	if sourceCtx.ID == "" {
		return meaning.Meaning{}, meaning.RequestContext{}, fmt.Errorf("validated request context does not exactly match coordinated meaning")
	}
	filtered := meaning.RequestContext{ID: sourceCtx.ID, Referents: []meaning.Referent{}, Mentions: []meaning.Mention{}}
	refIDs := map[meaning.ReferentID]bool{}
	for _, m := range sourceCtx.Mentions {
		if keep[m.NodeID] {
			filtered.Mentions = append(filtered.Mentions, m)
			if m.ReferentID != "" {
				refIDs[m.ReferentID] = true
			}
		}
	}
	for _, r := range sourceCtx.Referents {
		if refIDs[r.ID] {
			filtered.Referents = append(filtered.Referents, r)
		}
	}
	if err := meaning.ValidateReferences(view, filtered, meaning.DefaultBudget()); err != nil {
		return meaning.Meaning{}, meaning.RequestContext{}, err
	}
	return view, filtered, nil
}

func joinCoordinationClauses(a, b string, plan construction.ClauseCoordinationPlan) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if !strings.HasSuffix(a, ".") || !strings.HasSuffix(b, ".") {
		return ""
	}
	a = strings.TrimSuffix(a, ".")
	b = strings.TrimSuffix(b, ".")
	if plan.LowerSecond {
		runes := []rune(b)
		if len(runes) > 0 {
			runes[0] = unicode.ToLower(runes[0])
			b = string(runes)
		}
	}
	return a + " " + plan.Connector + " " + b + plan.FinalPeriod
}

func coordinatedAlternativeID(plan, left, right string) string {
	sum := sha256.Sum256([]byte(plan + "\x00" + left + "\x00" + right))
	return "coord:" + hex.EncodeToString(sum[:])
}
