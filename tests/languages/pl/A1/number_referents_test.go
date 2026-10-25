package a1pl_test

import (
	"context"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestNeuterBlackAttributionUsesTo(t *testing.T) {
	runtime := matrix.Runtime(t)
	realizer, ok := runtime.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || realizer == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	for _, tc := range []struct{ id, surface string }{
		{"a1.5.identification.029.en", "To okno jest czarne."},
		{"a1.5.identification.030.en", "To jabłko jest czarne."},
	} {
		t.Run(tc.id, func(t *testing.T) {
			fixture, ok, err := interlingua.IdentificationExpansion(tc.id)
			if err != nil || !ok {
				t.Fatalf("load authored I fixture: ok=%t error=%v", ok, err)
			}
			candidate := semantic.MeaningCandidate{
				ID: tc.id, SemanticID: tc.id, Meaning: &fixture.Graph, Status: status.OutcomeOK,
				Evidence: semantic.EvidenceTrace{AlternativeID: tc.id, Outcome: status.OutcomeOK},
				Contexts: []semantic.MeaningContext{{Meaning: fixture.Graph, Context: fixture.Context}},
			}
			result, err := realizer.Realize(context.Background(), candidate)
			if err != nil {
				t.Fatalf("realize: %v", err)
			}
			t.Logf("outcome=%s alternatives=%d", result.Outcome, len(result.Alternatives))
			for _, alternative := range result.Alternatives {
				if alternative.Text == tc.surface {
					return
				}
				if alternative.Text == "Ten okno jest czarne." || alternative.Text == "Ten jabłko jest czarne." {
					t.Errorf("masculine demonstrative was selected for neuter bearer: %q", alternative.Text)
				}
			}
			t.Errorf("expected %q, got %+v", tc.surface, result.Alternatives)
		})
	}
}

func TestCountNounReferentsInOrdinaryClauses(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	for _, tc := range []struct {
		input, predicate, role, concept string
		count                           meaning.Multiplicity
	}{
		{"Nauczyciel widzi okno.", "seeing", "stimulus", "window", meaning.MultiplicityOne},
		{"Nauczyciel widzi okna.", "seeing", "stimulus", "window", meaning.MultiplicityMany},
		{"Nauczyciel widzi jabłko.", "seeing", "stimulus", "apple_fruit", meaning.MultiplicityOne},
		{"Nauczyciel widzi jabłka.", "seeing", "stimulus", "apple_fruit", meaning.MultiplicityMany},
		{"Nauczyciel czyta list.", "reading", "material", "letter_document", meaning.MultiplicityOne},
		{"Nauczyciel czyta listy.", "reading", "material", "letter_document", meaning.MultiplicityMany},
	} {
		t.Run(tc.input, func(t *testing.T) {
			result, err := analyzer.Analyze(context.Background(), tc.input)
			if err != nil {
				t.Fatalf("analyze: %v", err)
			}
			t.Logf("outcome=%s candidate_count=%d", result.Outcome, len(result.Candidates))
			if !hasCountedParticipant(result.Candidates, tc.predicate, tc.role, tc.concept, tc.count, "") {
				t.Errorf("no candidate has %s=%s with multiplicity %s", tc.role, tc.concept, tc.count)
			}
		})
	}
}

func TestMotherReadLetterPastMeaning(t *testing.T) {
	fixture, ok, err := interlingua.TimeExpansion("a1.10.time.021.en")
	if err != nil || !ok {
		t.Fatalf("load authored past reading I: ok=%t error=%v", ok, err)
	}
	runtime := matrix.Runtime(t)
	realizer, ok := runtime.TargetRealizerFor(features.LanguageID("pl"))
	if !ok || realizer == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	candidate := semantic.MeaningCandidate{
		ID: "a1.10.time.021.en", SemanticID: "a1.10.time.021.en", Meaning: &fixture.Graph, Status: status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: "a1.10.time.021.en", Outcome: status.OutcomeOK},
		Contexts: []semantic.MeaningContext{{Meaning: fixture.Graph, Context: fixture.Context}},
	}
	result, err := realizer.Realize(context.Background(), candidate)
	if err != nil {
		t.Fatalf("realize: %v", err)
	}
	t.Logf("outcome=%s alternatives=%d", result.Outcome, len(result.Alternatives))
	for _, alternative := range result.Alternatives {
		if alternative.Text == "Matka czytała list." {
			return
		}
	}
	t.Errorf("expected past reading with singular letter: %+v", result.Alternatives)
}

func hasCountedParticipant(candidates []semantic.MeaningCandidate, predicate, role, concept string, count meaning.Multiplicity, tense string) bool {
	for _, candidate := range candidates {
		if candidate.Meaning == nil {
			continue
		}
		for _, node := range candidate.Meaning.Nodes {
			if node.Kind != meaning.KindEvent || string(node.Concept) != "concept:"+predicate || tense != "" && node.TemporalFrame != meaning.TemporalFrame(tense) {
				continue
			}
			for _, participant := range node.Participants {
				if string(participant.Role) != role {
					continue
				}
				for _, referent := range candidate.Meaning.Nodes {
					if referent.ID == participant.Target && referent.Kind == meaning.KindReferent && string(referent.Concept) == "concept:"+concept && referent.Multiplicity == count {
						return true
					}
				}
			}
		}
	}
	return false
}

func TestNeuterAttributionSourceAndTranslation(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	for _, input := range []string{"To okno jest czarne.", "To jabłko jest czarne."} {
		result, err := analyzer.Analyze(context.Background(), input)
		if err != nil {
			t.Fatalf("analyze %q: %v", input, err)
		}
		if !hasBlackBearer(result.Candidates) {
			t.Errorf("%q produced no black bearer Meaning", input)
		}
	}
	for _, input := range []string{"Ten okno jest czarne.", "To kot jest czarny."} {
		result, err := analyzer.Analyze(context.Background(), input)
		if err != nil {
			t.Fatalf("analyze mismatch %q: %v", input, err)
		}
		if hasBlackBearer(result.Candidates) {
			t.Errorf("mismatched demonstrative/gender unexpectedly formed black Meaning for %q", input)
		}
	}
	for _, tc := range []struct {
		from, to       features.LanguageID
		input, surface string
	}{
		{"en", "pl", "This window is black.", "To okno jest czarne."},
		{"pl", "en", "To okno jest czarne.", "This window is black."},
	} {
		result, err := runtime.Translate(context.Background(), engine.Request{From: tc.from, To: tc.to, Text: tc.input})
		if err != nil {
			t.Fatalf("translate %q: %v", tc.input, err)
		}
		texts := make([]string, 0, len(result.Candidates))
		for _, candidate := range result.Candidates {
			texts = append(texts, candidate.Text)
		}
		t.Logf("%s->%s status=%s texts=%v", tc.from, tc.to, result.Status, texts)
		found := false
		for _, text := range texts {
			if text == tc.surface {
				found = true
			}
		}
		if !found {
			t.Errorf("%s->%s did not produce %q: status=%s texts=%v", tc.from, tc.to, tc.surface, result.Status, texts)
		}
	}
}

func hasBlackBearer(candidates []semantic.MeaningCandidate) bool {
	for _, candidate := range candidates {
		if candidate.Meaning == nil {
			continue
		}
		anchored := false
		for _, requestContext := range candidate.Contexts {
			for _, mention := range requestContext.Context.Mentions {
				if mention.Anchor != nil && mention.Anchor.Kind == "deixis" && mention.Anchor.Value == "demonstrative" {
					anchored = true
				}
			}
		}
		if !anchored {
			continue
		}
		for _, state := range candidate.Meaning.Nodes {
			if state.Kind != meaning.KindState || string(state.Concept) != "concept:black" || len(state.Participants) != 1 || state.Participants[0].Role != "bearer" {
				continue
			}
			for _, referent := range candidate.Meaning.Nodes {
				if referent.ID == state.Participants[0].Target && referent.Kind == meaning.KindReferent && referent.Multiplicity == meaning.MultiplicityOne {
					return true
				}
			}
		}
	}
	return false
}
