package a1pl_test

import (
	"context"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"

	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestPastSeeingIndependentIToPolish(t *testing.T) {
	runtime := matrix.Runtime(t)
	realizer, ok := runtime.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || realizer == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	for _, tc := range []struct{ id, want string }{
		{"a1.10.time.017.en", "Nauczyciel widział psa."},
		{"a1.10.time.018.en", "Nauczyciele widzieli psa."},
	} {
		t.Run(tc.id, func(t *testing.T) {
			fixture, ok, err := interlingua.TimeExpansion(tc.id)
			if err != nil || !ok {
				t.Fatalf("load I fixture: ok=%t error=%v", ok, err)
			}
			candidate := semantic.MeaningCandidate{
				ID: tc.id, SemanticID: tc.id, Meaning: &fixture.Graph, Status: status.OutcomeOK,
				Evidence: semantic.EvidenceTrace{AlternativeID: tc.id, Outcome: status.OutcomeOK},
				Contexts: []semantic.MeaningContext{{Meaning: fixture.Graph, Context: fixture.Context}},
			}
			result, err := realizer.Realize(context.Background(), candidate)
			if err != nil {
				t.Fatalf("realize I: %v", err)
			}
			texts := realizationTexts(result.Alternatives)
			t.Logf("outcome=%s texts=%v", result.Outcome, texts)
			if !containsText(texts, tc.want) {
				t.Errorf("expected %q, got %v", tc.want, texts)
			}
		})
	}
	for _, tc := range []struct {
		who, whoConcept string
		number          meaning.Multiplicity
		stimulus        string
		stimulusCount   meaning.Multiplicity
		want            string
	}{
		{"Matka", "mother", meaning.MultiplicityOne, "dog", meaning.MultiplicityOne, "Matka widziała psa."},
		{"Pies", "dog", meaning.MultiplicityOne, "cat", meaning.MultiplicityOne, "Pies widział kota."},
		{"Psy", "dog", meaning.MultiplicityMany, "cat", meaning.MultiplicityOne, "Psy widziały kota."},
	} {
		graph := pastSeeingGraph(t, tc.whoConcept, tc.number, tc.stimulus, tc.stimulusCount)
		candidate := semantic.MeaningCandidate{ID: tc.want, SemanticID: tc.want, Meaning: &graph, Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: tc.want, Outcome: status.OutcomeOK}}
		result, err := realizer.Realize(context.Background(), candidate)
		if err != nil {
			t.Fatalf("realize %s: %v", tc.who, err)
		}
		if !containsText(realizationTexts(result.Alternatives), tc.want) {
			t.Errorf("expected %q, got %v", tc.want, realizationTexts(result.Alternatives))
		}
	}
}

func TestPastSeeingPolishSourceAndAgreement(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	for _, tc := range []struct {
		input, experiencer string
		multiplicity       meaning.Multiplicity
	}{
		{"Nauczyciel widział psa.", "teacher", meaning.MultiplicityOne},
		{"Nauczyciele widzieli psa.", "teacher", meaning.MultiplicityMany},
		{"Matka widziała psa.", "mother", meaning.MultiplicityOne},
		{"Psy widziały kota.", "dog", meaning.MultiplicityMany},
	} {
		t.Run(tc.input, func(t *testing.T) {
			result, err := analyzer.Analyze(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("analyze: %v", err)
			}
			t.Logf("outcome=%s candidates=%d", result.Outcome, len(result.Candidates))
			if !hasPastSeeing(result.Candidates, tc.experiencer, tc.multiplicity) {
				t.Errorf("no past seeing Meaning with experiencer %s multiplicity %s", tc.experiencer, tc.multiplicity)
			}
		})
	}
	for _, input := range []string{"Matka widział psa.", "Nauczyciele widziała psa.", "Psy widzieli kota."} {
		result, err := analyzer.Analyze(context.Background(), input)
		if err != nil {
			t.Fatalf("analyze known agreement conflict: %v", err)
		}
		if result.Outcome == status.OutcomeOK {
			t.Errorf("known past agreement conflict produced top-level OK: %q", input)
		}
		for _, candidate := range result.Candidates {
			if candidate.Status == status.OutcomeOK && hasAnyPastSeeing(candidate) {
				t.Errorf("known past agreement conflict produced an OK past seeing Meaning: %q", input)
			}
		}
	}
}

func TestPastSeeingEnglishSourceAndCrossLanguageTranslation(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("en"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: English source analyzer unavailable")
	}
	for _, tc := range []struct {
		input, experiencer string
		multiplicity       meaning.Multiplicity
	}{
		{"The teacher saw the dog.", "teacher", meaning.MultiplicityOne},
		{"The teachers saw the dog.", "teacher", meaning.MultiplicityMany},
	} {
		result, err := analyzer.Analyze(context.Background(), tc.input)
		if err != nil {
			t.Fatalf("analyze %q: %v", tc.input, err)
		}
		t.Logf("English %q outcome=%s candidates=%d", tc.input, result.Outcome, len(result.Candidates))
		if !hasPastSeeing(result.Candidates, tc.experiencer, tc.multiplicity) {
			t.Errorf("no English source past seeing Meaning for %q", tc.input)
		}
	}
	for _, tc := range []struct {
		from, to    features.LanguageID
		input, want string
	}{
		{"en", "pl", "The teacher saw the dog.", "Nauczyciel widział psa."},
		{"en", "pl", "The teachers saw the dog.", "Nauczyciele widzieli psa."},
		{"pl", "en", "Nauczyciel widział psa.", "The teacher saw the dog."},
		{"pl", "en", "Nauczyciele widzieli psa.", "The teachers saw the dog."},
	} {
		result, err := runtime.Translate(context.Background(), engine.Request{From: tc.from, To: tc.to, Text: tc.input})
		if err != nil {
			t.Fatalf("translate %q: %v", tc.input, err)
		}
		texts := make([]string, 0, len(result.Candidates))
		for _, alternative := range result.Candidates {
			texts = append(texts, alternative.Text)
		}
		t.Logf("%s->%s status=%s texts=%v", tc.from, tc.to, result.Status, texts)
		if !containsText(texts, tc.want) {
			t.Errorf("%s->%s expected %q, got %v", tc.from, tc.to, tc.want, texts)
		}
	}
}

func pastSeeingGraph(t *testing.T, experiencer string, experiencerCount meaning.Multiplicity, stimulus string, stimulusCount meaning.Multiplicity) meaning.Meaning {
	t.Helper()
	seeingID, err := ids.NewConceptID("seeing")
	if err != nil {
		t.Fatal(err)
	}
	experiencerID, err := ids.NewConceptID(experiencer)
	if err != nil {
		t.Fatal(err)
	}
	stimulusID, err := ids.NewConceptID(stimulus)
	if err != nil {
		t.Fatal(err)
	}
	event, err := meaning.NewEvent("e", seeingID, []meaning.Participant{{Role: "experiencer", Target: "a"}, {Role: "stimulus", Target: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	event.TemporalFrame = meaning.TemporalFramePast
	a, err := meaning.NewReferent("a", experiencerID)
	if err != nil {
		t.Fatal(err)
	}
	a.Multiplicity = experiencerCount
	b, err := meaning.NewReferent("b", stimulusID)
	if err != nil {
		t.Fatal(err)
	}
	b.Multiplicity = stimulusCount
	graph, err := meaning.Build([]meaning.NodeID{"e"}, []meaning.Node{event, a, b}, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	return graph
}

func hasPastSeeing(candidates []semantic.MeaningCandidate, experiencer string, count meaning.Multiplicity) bool {
	for _, candidate := range candidates {
		if candidate.Meaning == nil {
			continue
		}
		for _, event := range candidate.Meaning.Nodes {
			if event.Kind != meaning.KindEvent || string(event.Concept) != "concept:seeing" || event.TemporalFrame != meaning.TemporalFramePast {
				continue
			}
			for _, participant := range event.Participants {
				if participant.Role != "experiencer" {
					continue
				}
				for _, ref := range candidate.Meaning.Nodes {
					if ref.ID == participant.Target && ref.Kind == meaning.KindReferent && string(ref.Concept) == "concept:"+experiencer && ref.Multiplicity == count {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasAnyPastSeeing(candidate semantic.MeaningCandidate) bool {
	if candidate.Meaning == nil {
		return false
	}
	for _, node := range candidate.Meaning.Nodes {
		if node.Kind == meaning.KindEvent && string(node.Concept) == "concept:seeing" && node.TemporalFrame == meaning.TemporalFramePast {
			return true
		}
	}
	return false
}

func realizationTexts(alternatives []engine.SurfaceAlternative) []string {
	texts := make([]string, 0, len(alternatives))
	for _, alternative := range alternatives {
		texts = append(texts, alternative.Text)
	}
	return texts
}

func containsText(texts []string, want string) bool {
	for _, text := range texts {
		if text == want {
			return true
		}
	}
	return false
}
