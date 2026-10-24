package a1en_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	runtimepkg "github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionBoundaries(t *testing.T) {
	runtime := matrix.Runtime(t)
	target := targetEN(t)
	base, ok, err := interlingua.Composition("composition.black-cat-sees-dog")
	if err != nil || !ok {
		t.Fatalf("load independent composition fixture: ok=%t error=%v", ok, err)
	}
	positive := candidateEN(base, "composition-positive", meaning.RequestContext{})
	validateCandidateEN(t, positive)
	positiveResult, positiveErr := target.Realize(context.Background(), positive)
	logEN(t, "positive independent composition I", positiveResult, positiveErr)
	positiveTexts := []string{
		"A black cat sees a dog.", "A black cat sees the dog.",
		"The black cat sees a dog.", "The black cat sees the dog.",
	}
	if positiveErr != nil || !sameTextEN(textsEN(positiveResult.Alternatives), positiveTexts) {
		t.Fatalf("positive independent I did not realize its four licensed texts: outcome=%s texts=%v error=%v", positiveResult.Outcome, textsEN(positiveResult.Alternatives), positiveErr)
	}

	// A simple event without a property root remains a valid route.
	simple, ok, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("load simple seeing fixture: ok=%t error=%v", ok, err)
	}
	simpleCandidate := candidateEN(simple, "simple-seeing-no-property", meaning.RequestContext{})
	validateCandidateEN(t, simpleCandidate)
	simpleResult, simpleErr := target.Realize(context.Background(), simpleCandidate)
	logEN(t, "simple seeing without q", simpleResult, simpleErr)
	if simpleErr != nil || len(textsEN(simpleResult.Alternatives)) == 0 {
		t.Errorf("simple seeing without q was refused: outcome=%s diagnostics=%+v error=%v", simpleResult.Outcome, simpleResult.Diagnostics, simpleErr)
	}

	catCat, ok, err := interlingua.Composition("composition.black-cat-sees-cat")
	if err != nil || !ok {
		t.Fatalf("load cat/cat composition fixture: ok=%t error=%v", ok, err)
	}
	cases := []struct {
		name string
		edit func(meaning.Meaning) (meaning.Meaning, meaning.RequestContext)
	}{
		{"bearer is object", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return retargetBlackEN(t, g, "r2"), meaning.RequestContext{}
		}},
		{"bearer is second cat despite matching concept", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return retargetBlackEN(t, catCat, "r2"), meaning.RequestContext{}
		}},
		{"event time absent", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeFrameEN(&g, "e", "")
			return buildBoundaryEN(t, g), meaning.RequestContext{}
		}},
		{"event past", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeFrameEN(&g, "e", meaning.TemporalFramePast)
			return buildBoundaryEN(t, g), meaning.RequestContext{}
		}},
		{"subject many", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeMultiplicityEN(&g, "r1", meaning.MultiplicityMany)
			return buildBoundaryEN(t, g), meaning.RequestContext{}
		}},
		{"object multiplicity absent", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeMultiplicityEN(&g, "r2", "")
			return buildBoundaryEN(t, g), meaning.RequestContext{}
		}},
		{"property is timed", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			setNodeFrameEN(&g, "q", meaning.TemporalFramePresent)
			return buildBoundaryEN(t, g), meaning.RequestContext{}
		}},
		{"additional property root", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return addSecondBlackEN(t, g), meaning.RequestContext{}
		}},
		{"negation scopes event alongside property", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return scopeEventEN(t, g), meaning.RequestContext{}
		}},
		{"anchored subject context", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return g, compositionContextEN("r1", true)
		}},
		{"context graph has a different bearer", func(g meaning.Meaning) (meaning.Meaning, meaning.RequestContext) {
			return g, compositionContextEN("r1", false)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			graph, requestContext := tc.edit(cloneMeaningEN(t, base))
			candidate := candidateEN(graph, "composition-boundary-"+tc.name, requestContext)
			if tc.name == "context graph has a different bearer" {
				contextGraph := retargetBlackEN(t, cloneMeaningEN(t, graph), "r2")
				candidate.Contexts = []semantic.MeaningContext{{Meaning: contextGraph, Context: requestContext}}
			}
			validateCandidateEN(t, candidate)
			result, realizeErr := target.Realize(context.Background(), candidate)
			logEN(t, tc.name, result, realizeErr)
			assertNoTextEN(t, result, realizeErr)
		})
	}

	t.Run("offspring modifiers and article licensing", func(t *testing.T) {
		analyzer, exists := runtime.SourceAnalyzerFor("en")
		if !exists || analyzer == nil {
			t.Fatal("INFRA: English analyzer unavailable")
		}
		offspring, err := compositionRoleConceptEN(base, "experiencer", "child_of_person")
		if err != nil {
			t.Fatalf("prepare child_of_person I: %v", err)
		}
		if err := meaning.ValidateMeaning(offspring, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
			t.Fatalf("offspring I invalid against CORE: %v", err)
		}
		candidate := candidateEN(offspring, "composition-child-of-person", meaning.RequestContext{})
		validateCandidateEN(t, candidate)
		surface, realizeErr := target.Realize(context.Background(), candidate)
		logEN(t, "I->EN offspring subject", surface, realizeErr)
		want := []string{
			"A black offspring sees a dog.", "A black offspring sees the dog.",
			"The black offspring sees a dog.", "The black offspring sees the dog.",
		}
		if realizeErr != nil || !sameTextEN(textsEN(surface.Alternatives), want) {
			t.Errorf("FAIL: offspring target texts=%v want=%v outcome=%s error=%v", textsEN(surface.Alternatives), want, surface.Outcome, realizeErr)
		}

		result, analyzeErr := analyzer.Analyze(context.Background(), "A black offspring sees the dog.")
		logEN(t, "A black offspring sees the dog.", result, analyzeErr)
		checkCompositionEN(t, runtime, offspring, result, analyzeErr)
		for _, sourceCandidate := range result.Candidates {
			targetResult, targetErr := target.Realize(context.Background(), sourceCandidate)
			logEN(t, "direct target from A black offspring candidate", targetResult, targetErr)
			if targetErr != nil || targetResult.Outcome != "ok" || !sameTextEN(textsEN(targetResult.Alternatives), []string{"A black offspring sees the dog."}) {
				t.Errorf("FAIL: analyzed offspring did not select its one matching surface: outcome=%s texts=%v error=%v", targetResult.Outcome, textsEN(targetResult.Alternatives), targetErr)
			}
		}

		bad, badErr := analyzer.Analyze(context.Background(), "An black offspring sees the dog.")
		logEN(t, "An black offspring sees the dog.", bad, badErr)
		if badErr != nil {
			t.Errorf("unexpected analyzer error for rejected article: %v", badErr)
		} else {
			for _, sourceCandidate := range bad.Candidates {
				if sourceCandidate.Status == "ok" && sourceCandidate.Meaning != nil {
					t.Errorf("FAIL: unlicensed An black source produced OK Meaning %q", sourceCandidate.ID)
				}
			}
			if !hasCompositionDiagnostic(bad) {
				t.Errorf("FAIL: rejected An black input has no explicit diagnostic (outcome=%s)", bad.Outcome)
			}
		}

		articleResult, articleErr := analyzer.Analyze(context.Background(), "The black cat sees an offspring.")
		logEN(t, "The black cat sees an offspring.", articleResult, articleErr)
		objectOffspring, err := compositionRoleConceptEN(base, "stimulus", "child_of_person")
		if err != nil {
			t.Fatalf("prepare offspring object I: %v", err)
		}
		if err := meaning.ValidateMeaning(objectOffspring, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
			t.Fatalf("offspring object I invalid against CORE: %v", err)
		}
		if checkCompositionEN(t, runtime, objectOffspring, articleResult, articleErr) {
			for _, sourceCandidate := range articleResult.Candidates {
				targetResult, targetErr := target.Realize(context.Background(), sourceCandidate)
				logEN(t, "direct target preserving an offspring", targetResult, targetErr)
				if targetErr != nil || !sameTextEN(textsEN(targetResult.Alternatives), []string{"The black cat sees an offspring."}) {
					t.Errorf("FAIL: source an was not preserved: outcome=%s texts=%v error=%v", targetResult.Outcome, textsEN(targetResult.Alternatives), targetErr)
				}
			}
		}
	})

	t.Run("pending modifier mappings do not silently compose", func(t *testing.T) {
		assertCachedENComposition(t, runtime)
		for _, tc := range []struct {
			name, code string
			mutate     func(*lexicon.Document)
		}{
			{"black equivalent is pending", "MODIFIER_TRANSFER_UNRESOLVED", func(document *lexicon.Document) {
				for index := range document.Bindings {
					if document.Bindings[index].SenseID.String() == "sense:en.black.color" && document.Bindings[index].CoreRef.String() == "concept:black" {
						document.Bindings[index].Status = lexicon.BindingPending
						return
					}
				}
				t.Fatal("authored black binding not found")
			}},
			{"competing pending equivalent exists", "MODIFIER_TRANSFER_AMBIGUOUS", func(document *lexicon.Document) {
				for _, binding := range document.Bindings {
					if binding.SenseID.String() == "sense:en.black.color" && binding.CoreRef.String() == "concept:black" {
						seeing, err := ids.NewConceptID("seeing")
						if err != nil {
							t.Fatal(err)
						}
						binding.CoreRef = seeing
						binding.Status = lexicon.BindingPending
						document.Bindings = append(document.Bindings, binding)
						return
					}
				}
				t.Fatal("authored black binding not found")
			}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				probe := enProbePipelineEN(t, tc.mutate)
				result, analyzeErr := probe.Analyze(context.Background(), "The black cat sees the dog.")
				logEN(t, "probe "+tc.name, result, analyzeErr)
				if analyzeErr != nil {
					t.Fatalf("probe analysis failed: %v", analyzeErr)
				}
				if !hasCompositionDiagnosticCodeEN(result, tc.code) {
					t.Errorf("FAIL: expected diagnostic %s; outcome=%s diagnostics=%+v", tc.code, result.Outcome, result.Diagnostics)
				}
				if result.Outcome == "ok" {
					t.Error("FAIL: unresolved modifier mapping reported OK")
				}
				for _, sourceCandidate := range result.Candidates {
					if sourceCandidate.Meaning != nil {
						t.Errorf("FAIL: unresolved modifier mapping materialized Meaning %q", sourceCandidate.ID)
					}
				}
			})
			assertCachedENComposition(t, runtime)
		}
	})
}

func buildBoundaryEN(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	built, err := meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build valid boundary Meaning: %v", err)
	}
	return built
}

func setNodeFrameEN(graph *meaning.Meaning, id meaning.NodeID, value meaning.TemporalFrame) {
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == id {
			graph.Nodes[index].TemporalFrame = value
		}
	}
}

func setNodeMultiplicityEN(graph *meaning.Meaning, id meaning.NodeID, value meaning.Multiplicity) {
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == id {
			graph.Nodes[index].Multiplicity = value
		}
	}
}

func retargetBlackEN(t *testing.T, graph meaning.Meaning, referent meaning.NodeID) meaning.Meaning {
	t.Helper()
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == "q" {
			graph.Nodes[index].Participants[0].Target = referent
		}
	}
	return buildBoundaryEN(t, graph)
}

func addSecondBlackEN(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	black, err := ids.NewConceptID("black")
	if err != nil {
		t.Fatal(err)
	}
	state, err := meaning.NewState("q2", black, []meaning.Participant{{Role: "bearer", Target: "r1"}})
	if err != nil {
		t.Fatal(err)
	}
	graph.Nodes = append(graph.Nodes, state)
	graph.Roots = append(graph.Roots, "q2")
	return buildBoundaryEN(t, graph)
}

func scopeEventEN(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	negation, err := meaning.NewNegation("neg", "e")
	if err != nil {
		t.Fatal(err)
	}
	graph.Nodes = append(graph.Nodes, negation)
	graph.Roots = []meaning.NodeID{"neg", "q"}
	return buildBoundaryEN(t, graph)
}

func compositionContextEN(subject meaning.NodeID, anchored bool) meaning.RequestContext {
	request := meaning.RequestContext{ID: "composition-boundary-context", Mentions: []meaning.Mention{
		{ID: "subject", NodeID: subject}, {ID: "object", NodeID: "r2"},
	}}
	if anchored {
		request.Mentions[0].Anchor = &meaning.DeicticAnchor{Kind: "deixis", Value: "proximal"}
	}
	return request
}

func compositionRoleConceptEN(source meaning.Meaning, role meaning.Role, concept string) (meaning.Meaning, error) {
	conceptID, err := ids.NewConceptID(concept)
	if err != nil {
		return meaning.Meaning{}, err
	}
	graph, err := meaning.Build(source.Roots, source.Nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, err
	}
	var subjectID meaning.NodeID
	for _, node := range graph.Nodes {
		if node.ID == "e" {
			for _, participant := range node.Participants {
				if participant.Role == role {
					subjectID = participant.Target
				}
			}
		}
	}
	if subjectID == "" {
		return meaning.Meaning{}, fmt.Errorf("fixture has no %s", role)
	}
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == subjectID {
			graph.Nodes[index].Concept = meaning.ConceptRefFromID(conceptID)
		}
	}
	return meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget())
}

func checkCompositionEN(t *testing.T, runtime *engine.Engine, want meaning.Meaning, result semantic.Result, err error) bool {
	t.Helper()
	if err != nil || result.Outcome != "ok" || len(result.Candidates) == 0 {
		t.Errorf("GAP_ANALYZE: composition candidate missing or not OK: outcome=%s error=%v", result.Outcome, err)
		return false
	}
	found := false
	for _, candidate := range result.Candidates {
		if candidate.Status != "ok" || candidate.Meaning == nil {
			t.Errorf("FAIL: composition candidate %q has no complete Meaning", candidate.ID)
			continue
		}
		matches, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
		if compareErr != nil || !matches {
			t.Errorf("FAIL: candidate %q differs from expected complete graph: matches=%t error=%v", candidate.ID, matches, compareErr)
			continue
		}
		found = true
		for _, context := range candidate.Contexts {
			if meaning.ValidateReferences(context.Meaning, context.Context, meaning.DefaultBudget()) != nil {
				t.Errorf("FAIL: candidate %q has invalid reference context", candidate.ID)
			}
		}
	}
	return found
}

func hasCompositionDiagnosticCodeEN(result semantic.Result, code string) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	for _, candidate := range result.Candidates {
		for _, diagnostic := range candidate.Evidence.Diagnostics {
			if diagnostic.Code == code {
				return true
			}
		}
	}
	return false
}

func enProbePipelineEN(t *testing.T, mutate func(*lexicon.Document)) *runtimepkg.SourcePipeline {
	t.Helper()
	analyzer, ok := matrix.Runtime(t).SourceAnalyzerFor("en")
	if !ok {
		t.Fatal("INFRA: English analyzer unavailable")
	}
	original, ok := analyzer.(*runtimepkg.SourcePipeline)
	if !ok || original == nil {
		t.Fatalf("INFRA: source adapter has unexpected implementation %T", analyzer)
	}
	probe := *original
	document := original.Lexicon.Document()
	mutate(&document)
	index, err := lexicon.Build(document, probe.Core, lexicon.DefaultBudget())
	if err != nil {
		t.Fatalf("build probe lexicon: %v", err)
	}
	probe.Lexicon = index
	return &probe
}

func assertCachedENComposition(t *testing.T, runtime *engine.Engine) {
	t.Helper()
	analyzer, ok := runtime.SourceAnalyzerFor("en")
	if !ok || analyzer == nil {
		t.Fatal("INFRA: cached English analyzer unavailable")
	}
	want, found, err := interlingua.Composition("composition.black-cat-sees-dog")
	if err != nil || !found {
		t.Fatalf("load positive control fixture: found=%t error=%v", found, err)
	}
	result, analyzeErr := analyzer.Analyze(context.Background(), "The black cat sees the dog.")
	logEN(t, "cached runtime positive control", result, analyzeErr)
	if analyzeErr != nil || result.Outcome != "ok" || len(result.Candidates) == 0 {
		t.Fatalf("cached runtime changed or failed: outcome=%s error=%v", result.Outcome, analyzeErr)
	}
	matched := false
	for _, candidate := range result.Candidates {
		if candidate.Meaning == nil || candidate.Status != "ok" {
			t.Errorf("cached runtime candidate is incomplete: %q", candidate.ID)
			continue
		}
		equal, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
		if compareErr == nil && equal {
			matched = true
		}
	}
	if !matched {
		t.Error("cached runtime no longer matches the positive composition fixture")
	}
}
