package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	langruntime "github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestContextMatchesBindsAnchorsToGraphRoles(t *testing.T) {
	wantGraph := contextGraph(t, "event", "speaker-node", "addressee-node")
	want := contextFixture(wantGraph, "speaker-node", "addressee-node")
	gotGraph := contextGraph(t, "renamed-event", "renamed-speaker", "renamed-addressee")
	gotGraph.Roots = []meaning.NodeID{"renamed-addressee", "renamed-event", "renamed-speaker"}
	gotGraph.Nodes[0], gotGraph.Nodes[2] = gotGraph.Nodes[2], gotGraph.Nodes[0]
	gotGraph, err := meaning.Build(gotGraph.Roots, gotGraph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	candidateGraph := contextGraph(t, "independent-event", "independent-speaker", "independent-addressee")
	got := semantic.MeaningCandidate{Meaning: &candidateGraph, Contexts: []semantic.MeaningContext{{Meaning: gotGraph, Context: contextWithMentions("candidate-context", []meaning.Mention{
		contextMention("m-addressee", "renamed-addressee", "addressee"),
		contextMention("m-speaker", "renamed-speaker", "speaker"),
	})}}}
	if matches, err := ContextMatches(want, got); err != nil || !matches {
		t.Fatalf("renamed nodes, reordered graph nodes and reordered mentions should match: matches=%t err=%v", matches, err)
	}

	wrongRole := semantic.MeaningCandidate{Meaning: &candidateGraph, Contexts: []semantic.MeaningContext{{Meaning: gotGraph, Context: contextWithMentions("wrong-role", []meaning.Mention{
		contextMention("m-speaker", "renamed-speaker", "addressee"),
		contextMention("m-addressee", "renamed-addressee", "speaker"),
	})}}}
	if matches, err := ContextMatches(want, wrongRole); err == nil || matches {
		t.Fatalf("same anchor values attached to wrong participant roles must fail, matches=%t err=%v", matches, err)
	}
}

func TestContextMatchesRejectsMissingExtraWrongAndResolvedMentions(t *testing.T) {
	graph := contextGraph(t, "e", "r-speaker", "r-addressee")
	want := contextFixture(graph, "r-speaker", "r-addressee")
	cases := []struct {
		name     string
		mentions []meaning.Mention
		referent []meaning.Referent
	}{
		{name: "missing", mentions: []meaning.Mention{contextMention("m1", "r-speaker", "speaker")}},
		{name: "extra", mentions: []meaning.Mention{contextMention("m1", "r-speaker", "speaker"), contextMention("m2", "r-addressee", "addressee"), contextMention("m3", "r-speaker", "speaker")}},
		{name: "wrong-anchor", mentions: []meaning.Mention{contextMention("m1", "r-speaker", "proximal"), contextMention("m2", "r-addressee", "addressee")}},
		{name: "resolved-id", mentions: []meaning.Mention{{ID: "m1", NodeID: "r-speaker", ReferentID: "person-1", EvidenceID: "evidence-1", Anchor: roleAnchor("speaker")}, contextMention("m2", "r-addressee", "addressee")}, referent: []meaning.Referent{{ID: "person-1", Concept: "concept:teacher"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := contextWithMentions("actual-context", tc.mentions)
			ctx.Referents = tc.referent
			candidate := contextCandidate(graph, ctx)
			if matches, err := ContextMatches(want, candidate); err == nil || matches {
				t.Fatalf("invalid context should be rejected, matches=%t err=%v", matches, err)
			}
		})
	}
}

func TestContextMatchesIgnoresOnlyUnanchoredUnresolvedMentions(t *testing.T) {
	want, ok, err := interlingua.TimeExpansion("a1.10.time.001.en")
	if err != nil || !ok {
		t.Fatalf("load time expansion: ok=%t err=%v", ok, err)
	}
	var objectNode meaning.NodeID
	for _, node := range want.Graph.Nodes {
		if node.Kind == meaning.KindEvent {
			for _, participant := range node.Participants {
				if participant.Role == "stimulus" {
					objectNode = participant.Target
				}
			}
		}
	}
	if objectNode == "" {
		t.Fatal("fixture has no stimulus referent")
	}
	withExtra := func(mutate func(*meaning.Mention)) semantic.MeaningCandidate {
		context := want.Context
		context.Mentions = append([]meaning.Mention(nil), want.Context.Mentions...)
		extra := meaning.Mention{ID: "extra-object-mention", NodeID: objectNode}
		if mutate != nil {
			mutate(&extra)
		}
		context.Mentions = append(context.Mentions, extra)
		return contextCandidate(want.Graph, context)
	}
	if matches, err := ContextMatches(want, withExtra(nil)); err != nil || !matches {
		t.Fatalf("valid unanchored unresolved object mention should be ignored: matches=%t err=%v", matches, err)
	}
	if matches, err := ContextMatches(want, withExtra(func(m *meaning.Mention) { m.Anchor = roleAnchor("speaker") })); err == nil || matches {
		t.Fatalf("extra anchored mention should fail: matches=%t err=%v", matches, err)
	}
	if matches, err := ContextMatches(want, withExtra(func(m *meaning.Mention) { m.EvidenceID = "evidence-extra" })); err == nil || matches {
		t.Fatalf("extra evidence-linked mention should fail: matches=%t err=%v", matches, err)
	}
	resolved := withExtra(func(m *meaning.Mention) {
		m.ReferentID = "resolved-extra"
		m.EvidenceID = "evidence-extra"
	})
	resolved.Contexts[0].Context.Referents = []meaning.Referent{{ID: "resolved-extra", Concept: "concept:dog"}}
	if err := meaning.ValidateReferences(resolved.Contexts[0].Meaning, resolved.Contexts[0].Context, meaning.DefaultBudget()); err != nil {
		t.Fatalf("resolved control must be a valid graph/context pair: %v", err)
	}
	if matches, err := ContextMatches(want, resolved); err == nil || matches {
		t.Fatalf("extra resolved mention should fail: matches=%t err=%v", matches, err)
	}
}

func TestMatchAllowedCandidateUsesMatchingVariantContext(t *testing.T) {
	graph := contextGraph(t, "event", "left", "right")
	wrong := contextFixture(graph, "left", "right")
	wrong.Context.Mentions[0].Anchor = roleAnchor("addressee")
	correct := contextFixture(graph, "left", "right")
	candidate := contextCandidate(graph, contextWithMentions("actual", []meaning.Mention{
		contextMention("a", "left", "speaker"), contextMention("b", "right", "addressee"),
	}))
	id, matches, err := MatchAllowedCandidate(
		[]string{"primary-wrong-context", "secondary-correct-context"},
		map[string]meaning.Meaning{"primary-wrong-context": graph, "secondary-correct-context": graph},
		map[string]*interlingua.IdentificationMeaning{"primary-wrong-context": &wrong, "secondary-correct-context": &correct}, candidate)
	if err != nil || !matches || id != "secondary-correct-context" {
		t.Fatalf("candidate should match only the variant with its own context: id=%q matches=%t err=%v", id, matches, err)
	}
	boundID, matches, err := MatchSourceCandidateMeaning("source-secondary", graph, map[string]string{"source-secondary": id}, map[string]meaning.Meaning{"primary-wrong-context": graph, "secondary-correct-context": graph})
	if err != nil || !matches || boundID != "secondary-correct-context" {
		t.Fatalf("translation candidate should select source-candidate variant before graph matching: id=%q matches=%t err=%v", boundID, matches, err)
	}
	for _, sourceID := range []string{"", "unknown-source"} {
		if _, matches, err := MatchSourceCandidateMeaning(sourceID, graph, map[string]string{"source-secondary": id}, map[string]meaning.Meaning{id: graph}); err == nil || matches {
			t.Errorf("empty or unknown source ID %q must not bind: matches=%t err=%v", sourceID, matches, err)
		}
	}
	wrongGraph := contextGraph(t, "wrong-event", "left", "right")
	wrongEvent := wrongGraph.Nodes[0]
	wrongEvent.Participants[0].Role = "agent"
	wrongGraph.Nodes[0] = wrongEvent
	wrongGraph, err = meaning.Build(wrongGraph.Roots, wrongGraph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	if _, matches, err := MatchSourceCandidateMeaning("source-secondary", wrongGraph, map[string]string{"source-secondary": id}, map[string]meaning.Meaning{id: graph}); err != nil || matches {
		t.Errorf("source-ID binding must still reject a different graph: matches=%t err=%v", matches, err)
	}
}

func TestMatchAllowedCandidatePropagatesInvalidExpectedContextAndBudget(t *testing.T) {
	graph := contextGraph(t, "event", "left", "right")
	invalid := contextFixture(graph, "left", "right")
	invalid.Context.Mentions[0].NodeID = "missing-node"
	valid := contextFixture(graph, "left", "right")
	candidate := contextCandidate(graph, contextWithMentions("valid", []meaning.Mention{
		contextMention("a", "left", "speaker"), contextMention("b", "right", "addressee"),
	}))
	if _, matches, err := MatchAllowedCandidate([]string{"invalid", "valid"}, map[string]meaning.Meaning{"invalid": graph, "valid": graph}, map[string]*interlingua.IdentificationMeaning{"invalid": &invalid, "valid": &valid}, candidate); err == nil || matches {
		t.Fatalf("invalid expected context must not be hidden by another matching variant, matches=%t err=%v", matches, err)
	}

	large := contextGraphWithParticipants(t, 32)
	largeContext := contextFixture(large, "r0", "r1")
	largeCandidate := contextCandidate(large, largeContext.Context)
	if _, matches, err := MatchAllowedCandidate([]string{"large"}, map[string]meaning.Meaning{"large": large}, map[string]*interlingua.IdentificationMeaning{"large": &largeContext}, largeCandidate); err == nil || matches {
		t.Fatalf("comparison-budget error must propagate, matches=%t err=%v", matches, err)
	}
}

func contextGraph(t *testing.T, eventID, firstID, secondID meaning.NodeID) meaning.Meaning {
	t.Helper()
	seeing, err := ids.NewConceptID("seeing")
	if err != nil {
		t.Fatal(err)
	}
	teacher, err := ids.NewConceptID("teacher")
	if err != nil {
		t.Fatal(err)
	}
	dog, err := ids.NewConceptID("dog")
	if err != nil {
		t.Fatal(err)
	}
	event, err := meaning.NewEvent(eventID, seeing, []meaning.Participant{{Role: "experiencer", Target: firstID}, {Role: "stimulus", Target: secondID}})
	if err != nil {
		t.Fatal(err)
	}
	first, err := meaning.NewReferent(firstID, teacher)
	if err != nil {
		t.Fatal(err)
	}
	second, err := meaning.NewReferent(secondID, dog)
	if err != nil {
		t.Fatal(err)
	}
	first.Multiplicity, second.Multiplicity = meaning.MultiplicityOne, meaning.MultiplicityOne
	graph, err := meaning.Build([]meaning.NodeID{eventID, firstID, secondID}, []meaning.Node{event, first, second}, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func contextGraphWithParticipants(t *testing.T, extra int) meaning.Meaning {
	t.Helper()
	graph := contextGraph(t, "large-event", "r0", "r1")
	participants := []meaning.Participant{{Role: "experiencer", Target: "r0"}, {Role: "stimulus", Target: "r1"}}
	for i := 0; i < extra; i++ {
		id, err := ids.NewConceptID("cat")
		if err != nil {
			t.Fatal(err)
		}
		nodeID := meaning.NodeID(fmt.Sprintf("extra-%d", i))
		ref, err := meaning.NewReferent(nodeID, id)
		if err != nil {
			t.Fatal(err)
		}
		graph.Nodes = append(graph.Nodes, ref)
		participants = append(participants, meaning.Participant{Role: meaning.Role(fmt.Sprintf("extra%d", i)), Target: nodeID})
		graph.Roots = append(graph.Roots, nodeID)
	}
	seeing, _ := ids.NewConceptID("seeing")
	event, err := meaning.NewEvent("large-event", seeing, participants)
	if err != nil {
		t.Fatal(err)
	}
	graph.Nodes[0] = event
	large, err := meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	return large
}

func contextFixture(graph meaning.Meaning, first, second meaning.NodeID) interlingua.IdentificationMeaning {
	return interlingua.IdentificationMeaning{Graph: graph, Context: contextWithMentions("expected-context", []meaning.Mention{
		contextMention("speaker-mention", first, "speaker"), contextMention("addressee-mention", second, "addressee"),
	})}
}

func contextCandidate(graph meaning.Meaning, context meaning.RequestContext) semantic.MeaningCandidate {
	return semantic.MeaningCandidate{Meaning: &graph, Contexts: []semantic.MeaningContext{{Meaning: graph, Context: context}}}
}

func contextWithMentions(id string, mentions []meaning.Mention) meaning.RequestContext {
	return meaning.RequestContext{ID: id, Mentions: mentions}
}

func contextMention(id string, node meaning.NodeID, value string) meaning.Mention {
	return meaning.Mention{ID: meaning.MentionID(id), NodeID: node, Anchor: roleAnchor(value)}
}

func roleAnchor(value string) *meaning.DeicticAnchor {
	return &meaning.DeicticAnchor{Kind: "participant", Value: value}
}

func TestQuestionFixtureLoaderRoutesOnlyAnchoredContexts(t *testing.T) {
	plainGraph, plainContext, ok, err := loadFixture("a1.9.q.005.en")
	if err != nil || !ok {
		t.Fatalf("non-deictic question fixture: ok=%t err=%v", ok, err)
	}
	if plainContext != nil {
		t.Fatalf("non-deictic fixture was routed as a context contract: %+v", plainContext.Context)
	}
	if len(plainGraph.Nodes) == 0 {
		t.Fatal("non-deictic question fixture graph is empty")
	}

	anchoredGraph, anchoredContext, ok, err := loadFixture("a1.9.q.001.en")
	if err != nil || !ok {
		t.Fatalf("deictic question fixture: ok=%t err=%v", ok, err)
	}
	if anchoredContext == nil || len(anchoredContext.Context.Mentions) != 1 {
		t.Fatalf("deictic fixture lost its anchor context: %+v", anchoredContext)
	}
	if len(anchoredGraph.Nodes) == 0 {
		t.Fatal("deictic question fixture graph is empty")
	}
}

// A supported surface must not disappear because a different, licensed
// lexicalization has a typed coverage gap. Invalid data remains fatal.
func TestTargetAlternativeFailureIsolation(t *testing.T) {
	shared := Runtime(t)
	target, ok := shared.TargetRealizerFor("en")
	if !ok {
		t.Fatal("English target unavailable")
	}
	pipeline, ok := target.(*langruntime.TargetPipeline)
	if !ok {
		t.Fatal("production target is not the construction pipeline")
	}
	graph, found, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !found {
		t.Fatalf("independent seeing fixture: found=%t err=%v", found, err)
	}
	source := semantic.MeaningCandidate{
		ID: "independent-failure-isolation", SemanticID: "independent-failure-isolation", Meaning: &graph,
		Status:   status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: "independent-failure-isolation", Outcome: status.OutcomeOK},
	}
	baseline, err := pipeline.Realize(context.Background(), source)
	if err != nil || len(baseline.Alternatives) != 4 {
		t.Fatalf("expected four existing determiner alternatives: result=%+v err=%v", baseline, err)
	}
	allowed := map[string]bool{}
	for _, alternative := range baseline.Alternatives {
		allowed[alternative.Text] = true
	}
	before, _ := json.Marshal(source)
	for _, tc := range []struct {
		name                              string
		at                                int
		kind                              status.OutcomeKind
		cancel, fatal, malformed, untyped bool
		wantSurfaces                      int
	}{
		{name: "first unsupported", at: 1, kind: status.OutcomeUnsupported, wantSurfaces: 3},
		{name: "late incomplete", at: 2, kind: status.OutcomeIncomplete, wantSurfaces: 3},
		{name: "late unknown", at: 2, kind: status.OutcomeUnknown, wantSurfaces: 3},
		{name: "late ambiguous", at: 2, kind: status.OutcomeAmbiguous, wantSurfaces: 3},
		{name: "all unsupported", at: 0, kind: status.OutcomeUnsupported, wantSurfaces: 0},
		{name: "resource limit", at: 2, kind: status.OutcomeResourceLimit, wantSurfaces: 1},
		{name: "context canceled", at: 2, cancel: true, kind: status.OutcomeResourceLimit, wantSurfaces: 1},
		{name: "invalid remains fatal", at: 2, kind: status.OutcomeInvalid, fatal: true},
		{name: "error remains fatal", at: 2, kind: status.OutcomeError, fatal: true},
		{name: "untyped remains fatal", at: 2, untyped: true, fatal: true},
		{name: "malformed remains fatal", at: 2, kind: status.OutcomeUnsupported, malformed: true, fatal: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			copy := *pipeline
			injected := &targetFailureInjection{inner: pipeline.ExecutionBuilder, at: tc.at, kind: tc.kind, malformed: tc.malformed, untyped: tc.untyped}
			if tc.cancel {
				injected.cancel = cancel
			}
			copy.ExecutionBuilder = injected
			result, realizeErr := copy.Realize(ctx, source)
			logObservation(t, map[string]any{"check_class": "target_failure_isolation", "case_id": tc.name, "stage": "independent_realization", "result": result, "error": errorText(realizeErr)})
			if tc.fatal {
				if realizeErr == nil || len(result.Alternatives) != 0 {
					t.Fatalf("fatal error was converted to a usable partial result: %+v err=%v", result, realizeErr)
				}
				return
			}
			if realizeErr != nil || len(result.Alternatives) != tc.wantSurfaces || len(result.Diagnostics) == 0 {
				t.Fatalf("want %d retained alternatives and diagnostic, got %+v err=%v", tc.wantSurfaces, result, realizeErr)
			}
			if result.Outcome == status.OutcomeOK {
				t.Error("partial coverage was reported as a complete success")
			}
			if tc.kind == status.OutcomeResourceLimit && (result.Outcome != status.OutcomeResourceLimit || injected.calls != 2) {
				t.Errorf("resource stop must preserve the first alternative and terminate: outcome=%s calls=%d", result.Outcome, injected.calls)
			}
			for _, alternative := range result.Alternatives {
				if !allowed[alternative.Text] {
					t.Errorf("injection fabricated a surface: %q", alternative.Text)
				}
			}
			after, _ := json.Marshal(source)
			if string(before) != string(after) {
				t.Error("failure handling changed the independent meaning or context")
			}
		})
	}
}

type targetFailureInjection struct {
	inner              langruntime.TargetExecutionBuilder
	calls, at          int
	kind               status.OutcomeKind
	cancel             context.CancelFunc
	malformed, untyped bool
}

func (injection *targetFailureInjection) Build(ctx context.Context, source semantic.MeaningCandidate, candidates []realizer.TargetCandidate, selected realizer.TargetCandidate, plan realizer.RealizationPlan) (realizer.ExecuteInput, error) {
	injection.calls++
	if injection.at == 0 || injection.calls == injection.at {
		if injection.cancel != nil {
			injection.cancel()
			return realizer.ExecuteInput{}, ctx.Err()
		}
		if injection.untyped {
			return realizer.ExecuteInput{}, errors.New("injected untyped failure")
		}
		result := status.Result{Status: injection.kind}
		if !injection.malformed {
			result.Diagnostics = []status.Diagnostic{{Code: "INJECTED_TARGET_REFUSAL", Path: "target.slots", Message: "bounded fault injection", Severity: status.SeverityWarning}}
		}
		return realizer.ExecuteInput{}, fmt.Errorf("wrapped candidate failure: %w", targetInjectedFailure{result: result})
	}
	return injection.inner.Build(ctx, source, candidates, selected, plan)
}

type targetInjectedFailure struct{ result status.Result }

func (targetInjectedFailure) Error() string                 { return "injected target failure" }
func (failure targetInjectedFailure) Result() status.Result { return failure.result }

// Missing morphology must be a typed coverage gap, not an internal error or
// permission to use present. Future licensed past forms may satisfy this test;
// it deliberately does not freeze today's absence of a paradigm cell.
func TestTargetMissingPastParadigmIsDiagnosed(t *testing.T) {
	rt := Runtime(t)
	for _, text := range []string{"The bird flew.", "The birds flew."} {
		t.Run(text, func(t *testing.T) {
			result, err := rt.Translate(context.Background(), engine.Request{From: "en", To: "pl", Text: text})
			logObservation(t, map[string]any{"check_class": "past_morphology_gap", "case_id": text, "stage": "integration", "result": result, "error": errorText(err)})
			if err != nil || result.Status == status.OutcomeError || result.Status == status.OutcomeInvalid || result.Status == status.OutcomeResourceLimit {
				t.Fatalf("morphology coverage must not cause a runtime failure: status=%s err=%v", result.Status, err)
			}
			surfaces := 0
			for _, candidate := range result.Candidates {
				if candidate.Text == "" {
					continue
				}
				surfaces++
				past := false
				if candidate.Meaning != nil {
					for _, node := range candidate.Meaning.Nodes {
						if node.Concept == "concept:flying_motion" && node.Kind == meaning.KindEvent {
							past = node.TemporalFrame == meaning.TemporalFramePast
						}
					}
				}
				if !past {
					t.Error("realized candidate did not preserve the source past event")
				}
				if candidate.Outcome != status.OutcomeOK && candidate.Outcome != status.OutcomeIncomplete && candidate.Outcome != status.OutcomeAmbiguous {
					t.Errorf("refused candidate contains a surface: %+v", candidate)
				}
			}
			if surfaces == 0 {
				found := false
				for _, diagnostic := range result.Diagnostics {
					found = found || diagnostic.Code == "TARGET_PAST_PARADIGM_UNSUPPORTED"
				}
				if result.Status != status.OutcomeUnsupported || !found {
					t.Errorf("unrealized past must retain its specific coverage diagnostic: %+v", result)
				}
			}
		})
	}
}
