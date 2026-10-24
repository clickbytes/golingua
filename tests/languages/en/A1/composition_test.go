package a1en_test

import (
	"context"
	"sort"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionAnalysis(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("en"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: English source analyzer is unavailable")
	}
	cases := []struct{ id, input string }{
		{"composition.black-cat-sees-dog", "The black cat sees the dog."},
		{"composition.black-dog-sees-cat", "The black dog sees the cat."},
		{"composition.black-bird-sees-dog", "The black bird sees the dog."},
		{"composition.black-cat-sees-cat", "The black cat sees the cat."},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			checkCompositionSource(t, runtime, analyzer, tc.id, tc.input, nil)
		})
	}
	t.Run("indefinite black subject and object", func(t *testing.T) {
		checkCompositionSource(t, runtime, analyzer, "composition.black-cat-sees-dog", "A black cat sees a dog.", []string{"indefinite", "indefinite"})
	})
	t.Run("analyzed source candidate realizes without ambiguity", func(t *testing.T) {
		result, err := analyzer.Analyze(context.Background(), "The black cat sees the dog.")
		logEN(t, "The black cat sees the dog. -> target", result, err)
		if err != nil || len(result.Candidates) == 0 {
			t.Fatalf("GAP_ANALYZE: source candidate unavailable: %s %v", result.Outcome, err)
		}
		target := targetEN(t)
		for _, candidate := range result.Candidates {
			if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
				t.Errorf("FAIL: source candidate %q is not a complete meaning", candidate.ID)
				continue
			}
			surface, realizeErr := target.Realize(context.Background(), candidate)
			logEN(t, "direct target for source candidate "+candidate.ID, surface, realizeErr)
			if realizeErr != nil || surface.Outcome != status.OutcomeOK || !sameTextEN(textsEN(surface.Alternatives), []string{"The black cat sees the dog."}) {
				t.Errorf("FAIL: source determiner features did not select one exact surface: outcome=%s texts=%v error=%v", surface.Outcome, textsEN(surface.Alternatives), realizeErr)
			}
			assertAllowedAssumptionsEN(t, surface.Alternatives)
			assertNoTenseNumberEN(t, surface.Alternatives)
		}
	})
	for _, input := range []string{"This is a water", "This is a money"} {
		t.Run("refuse "+input, func(t *testing.T) {
			result, err := analyzer.Analyze(context.Background(), input)
			logEN(t, input, result, err)
			if err != nil {
				t.Fatalf("unexpected analyzer error instead of a diagnosed source refusal: %v", err)
			}
			for _, candidate := range result.Candidates {
				if candidate.Status == status.OutcomeOK && candidate.Meaning != nil {
					t.Errorf("FAIL: indefinite classification unexpectedly produced an OK Meaning for %q", input)
				}
			}
			if !hasCompositionDiagnostic(result) {
				t.Errorf("FAIL: unsupported indefinite classification has no explicit diagnostic (outcome=%s)", result.Outcome)
			}
		})
	}
}

func checkCompositionSource(t *testing.T, runtime *engine.Engine, analyzer engine.SourceAnalyzer, id, input string, wantedDeterminers []string) {
	t.Helper()
	want, ok, err := interlingua.Composition(id)
	if err != nil || !ok {
		t.Fatalf("load authored Composition fixture: ok=%t error=%v", ok, err)
	}
	if err := meaning.ValidateMeaning(want, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
		t.Fatalf("INFRA: fixture does not validate against CORE: %v", err)
	}
	result, analyzeErr := analyzer.Analyze(context.Background(), input)
	logEN(t, input, result, analyzeErr)
	if analyzeErr != nil || len(result.Candidates) == 0 || result.Outcome == status.OutcomeUnsupported {
		t.Fatalf("GAP_ANALYZE: no supported composition candidate: %s %v", result.Outcome, analyzeErr)
	}
	for _, candidate := range result.Candidates {
		if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
			t.Errorf("FAIL: candidate %q has no complete Meaning (status=%s)", candidate.ID, candidate.Status)
			continue
		}
		matches, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
		if compareErr != nil || !matches {
			t.Errorf("FAIL: candidate %q differs from authored complete graph (matches=%t error=%v)", candidate.ID, matches, compareErr)
		}
		for index, candidateContext := range candidate.Contexts {
			if err := meaning.ValidateReferences(candidateContext.Meaning, candidateContext.Context, meaning.DefaultBudget()); err != nil {
				t.Errorf("FAIL: candidate %q context %d is invalid: %v", candidate.ID, index, err)
				continue
			}
			equal, err := meaning.CanonicalEqual(candidateContext.Meaning, *candidate.Meaning, meaning.DefaultBudget())
			if err != nil || !equal || len(candidateContext.Context.Referents) != 0 {
				t.Errorf("FAIL: candidate %q context %d differs from unresolved Meaning: equal=%t error=%v", candidate.ID, index, equal, err)
			}
			for _, mention := range candidateContext.Context.Mentions {
				if mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor != nil {
					t.Errorf("FAIL: candidate %q introduced identity or deixis: %+v", candidate.ID, mention)
				}
			}
		}
		if wantedDeterminers != nil {
			got := englishSourceDeterminers(candidate)
			if !sameTextEN(got, wantedDeterminers) {
				t.Errorf("FAIL: explicit source determiner evidence=%v, want %v", got, wantedDeterminers)
			}
		}
	}
}

func englishSourceDeterminers(candidate semantic.MeaningCandidate) []string {
	result := []string{"", ""}
	for _, record := range candidate.Evidence.Records {
		index := -1
		switch string(record.Role) {
		case "experiencer":
			index = 0
		case "stimulus":
			index = 1
		}
		if index < 0 {
			continue
		}
		for _, feature := range record.Features {
			if string(feature.Feature) == "determiner" {
				result[index] = string(feature.Value)
			}
		}
	}
	return result
}

func hasCompositionDiagnostic(result semantic.Result) bool {
	if len(result.Diagnostics) != 0 {
		return true
	}
	for _, candidate := range result.Candidates {
		if len(candidate.Evidence.Diagnostics) != 0 {
			return true
		}
	}
	return false
}

func TestCompositionRealization(t *testing.T) {
	target := targetEN(t)
	cases := []struct{ id string }{
		{"composition.black-cat-sees-dog"},
		{"composition.black-dog-sees-cat"},
		{"composition.black-bird-sees-dog"},
		{"composition.black-cat-sees-cat"},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			graph, ok, err := interlingua.Composition(tc.id)
			if err != nil || !ok {
				t.Fatalf("load independent Composition fixture: ok=%t error=%v", ok, err)
			}
			graph = reorderedEnglishComposition(t, graph)
			candidate := candidateEN(graph, "I-EN-"+tc.id, meaning.RequestContext{})
			validateCandidateEN(t, candidate)
			result, realizeErr := target.Realize(context.Background(), candidate)
			logEN(t, "I->EN "+tc.id, result, realizeErr)
			want := expectedEnglishCompositionTexts(tc.id)
			if realizeErr != nil || !sameTextEN(textsEN(result.Alternatives), want) {
				t.Errorf("FAIL: got texts=%v, want exactly %v outcome=%s error=%v", textsEN(result.Alternatives), want, result.Outcome, realizeErr)
			}
			assertAllowedAssumptionsEN(t, result.Alternatives)
			assertNoTenseNumberEN(t, result.Alternatives)
		})
	}
}

func reorderedEnglishComposition(t *testing.T, source meaning.Meaning) meaning.Meaning {
	t.Helper()
	names := map[meaning.NodeID]meaning.NodeID{"e": "m_event", "q": "a_property", "r1": "z_subject", "r2": "a_object"}
	graph := meaning.Meaning{SchemaVersion: source.SchemaVersion, Roots: make([]meaning.NodeID, len(source.Roots)), Nodes: make([]meaning.Node, len(source.Nodes))}
	for index, root := range source.Roots {
		graph.Roots[index] = names[root]
	}
	graph.Roots[0], graph.Roots[1] = graph.Roots[1], graph.Roots[0]
	for index, node := range source.Nodes {
		node.ID = names[node.ID]
		node.Participants = append([]meaning.Participant(nil), node.Participants...)
		for participantIndex := range node.Participants {
			node.Participants[participantIndex].Target = names[node.Participants[participantIndex].Target]
		}
		graph.Nodes[index] = node
	}
	sort.Slice(graph.Nodes, func(left, right int) bool { return graph.Nodes[left].ID < graph.Nodes[right].ID })
	if graph.Nodes[0].ID != "a_object" || graph.Nodes[len(graph.Nodes)-1].ID != "z_subject" {
		t.Fatalf("test setup did not put object ID before subject ID")
	}
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		t.Fatalf("reordered authored graph is invalid: %v", err)
	}
	return graph
}

func expectedEnglishCompositionTexts(id string) []string {
	nouns := map[string][2]string{
		"composition.black-cat-sees-dog":  {"cat", "dog"},
		"composition.black-dog-sees-cat":  {"dog", "cat"},
		"composition.black-bird-sees-dog": {"bird", "dog"},
		"composition.black-cat-sees-cat":  {"cat", "cat"},
	}
	pair := nouns[id]
	result := make([]string, 0, 4)
	for _, subjectArticle := range []string{"The", "A"} {
		for _, objectArticle := range []string{"the", "a"} {
			result = append(result, subjectArticle+" black "+pair[0]+" sees "+objectArticle+" "+pair[1]+".")
		}
	}
	sort.Strings(result)
	return result
}
