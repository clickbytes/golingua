package translation_test

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionTranslation(t *testing.T) {
	runtime := matrix.Runtime(t)
	cases := []struct {
		id, enInput, plInput, subject, object string
	}{
		{"composition.black-cat-sees-dog", "The black cat sees the dog.", "Czarny kot widzi psa.", "cat", "dog"},
		{"composition.black-dog-sees-cat", "The black dog sees the cat.", "Czarny pies widzi kota.", "dog", "cat"},
		{"composition.black-bird-sees-dog", "The black bird sees the dog.", "Czarny ptak widzi psa.", "bird", "dog"},
		{"composition.black-cat-sees-cat", "The black cat sees the cat.", "Czarny kot widzi kota.", "cat", "cat"},
	}
	for _, tc := range cases {
		t.Run(tc.id+" EN-to-PL", func(t *testing.T) {
			checkCompositionTranslation(t, runtime, tc.id, "en", "pl", tc.enInput, status.OutcomeOK, []string{tc.plInput})
		})
		t.Run(tc.id+" PL-to-EN", func(t *testing.T) {
			checkCompositionTranslation(t, runtime, tc.id, "pl", "en", tc.plInput, status.OutcomeAmbiguous, englishCompositionTextSet(tc.subject, tc.object))
		})
	}
}

func checkCompositionTranslation(t *testing.T, runtime *engine.Engine, fixtureID, from, to, input string, expectedOutcome status.OutcomeKind, expectedTexts []string) {
	t.Helper()
	want, found, err := interlingua.Composition(fixtureID)
	if err != nil || !found {
		t.Fatalf("load independent Composition fixture %q: found=%t error=%v", fixtureID, found, err)
	}
	result, translateErr := runtime.Translate(context.Background(), engine.Request{
		From: features.LanguageID(from), To: features.LanguageID(to), Text: input,
	})
	logTranslationAttributes(t, input, result, translateErr)
	if translateErr != nil {
		t.Fatalf("translation error: %v", translateErr)
	}
	if result.Status != expectedOutcome {
		t.Fatalf("unexpected overall status %s; expected %s; diagnostics=%+v", result.Status, expectedOutcome, result.Diagnostics)
	}
	if len(result.Candidates) == 0 {
		t.Fatalf("translation has no candidates: status=%s diagnostics=%+v", result.Status, result.Diagnostics)
	}
	texts := make(map[string]bool, len(result.Candidates))
	wantCandidateOutcome := status.OutcomeIncomplete
	if from == "en" {
		wantCandidateOutcome = status.OutcomeOK
	}
	for _, candidate := range result.Candidates {
		if candidate.Meaning == nil || candidate.Text == "" {
			t.Errorf("candidate %q lacks a complete Meaning or surface (outcome=%s)", candidate.ID, candidate.Outcome)
			continue
		}
		if candidate.Outcome != wantCandidateOutcome {
			t.Errorf("candidate %q has unexpected outcome %s; want %s", candidate.ID, candidate.Outcome, wantCandidateOutcome)
		}
		matches, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
		if compareErr != nil || !matches {
			t.Errorf("candidate %q intermediate Meaning differs from fixture %q (matches=%t error=%v)", candidate.ID, fixtureID, matches, compareErr)
		}
		if len(candidate.Losses) != 0 {
			t.Errorf("candidate %q loses Meaning fields: %+v", candidate.ID, candidate.Losses)
		}
		if from == "pl" {
			assertCompositionDeterminerAssumptions(t, candidate)
		} else {
			if len(candidate.Assumptions) != 0 {
				t.Errorf("EN->PL candidate %q has unexpected assumptions: %+v", candidate.ID, candidate.Assumptions)
			}
			if len(candidate.Diagnostics) != 0 {
				t.Errorf("EN->PL candidate %q has unexpected diagnostics: %+v", candidate.ID, candidate.Diagnostics)
			}
		}
		texts[candidate.Text] = true
	}
	if from == "en" {
		if len(result.Diagnostics) != 0 {
			t.Errorf("EN->PL result has unexpected diagnostics: %+v", result.Diagnostics)
		}
	}
	gotTexts := make([]string, 0, len(texts))
	for text := range texts {
		gotTexts = append(gotTexts, text)
	}
	sort.Strings(gotTexts)
	if !equalStringsAttributes(gotTexts, expectedTexts) {
		t.Errorf("surface set=%v; expected exactly %v", gotTexts, expectedTexts)
	}
}

func assertCompositionDeterminerAssumptions(t *testing.T, candidate engine.TranslationAlternative) {
	t.Helper()
	if candidate.Meaning == nil {
		return
	}
	var event meaning.Node
	for _, node := range candidate.Meaning.Nodes {
		if node.Kind == meaning.KindEvent && node.Concept == "concept:seeing" {
			if event.ID != "" {
				t.Errorf("candidate %q has multiple seeing events", candidate.ID)
				return
			}
			event = node
		}
	}
	roles := make(map[meaning.Role]meaning.NodeID, 2)
	for _, participant := range event.Participants {
		if participant.Role == "experiencer" || participant.Role == "stimulus" {
			if _, duplicate := roles[participant.Role]; duplicate {
				t.Errorf("candidate %q has duplicate role %q", candidate.ID, participant.Role)
				return
			}
			roles[participant.Role] = participant.Target
		}
	}
	if event.ID == "" || roles["experiencer"] == "" || roles["stimulus"] == "" || roles["experiencer"] == roles["stimulus"] {
		t.Errorf("candidate %q lacks distinct experiencer/stimulus node references", candidate.ID)
		return
	}
	for _, id := range roles {
		found := false
		for _, node := range candidate.Meaning.Nodes {
			if node.ID == id && node.Kind == meaning.KindReferent {
				found = true
			}
		}
		if !found {
			t.Errorf("candidate %q role target %q is not a referent", candidate.ID, id)
			return
		}
	}
	parts := strings.Fields(candidate.Text)
	if len(parts) != 6 || parts[1] != "black" || parts[3] != "sees" || !strings.HasSuffix(parts[5], ".") {
		t.Errorf("candidate %q does not have the licensed composition surface form %q", candidate.ID, candidate.Text)
		return
	}
	articleValue := func(article string) string {
		switch article {
		case "A", "a":
			return "indefinite"
		case "The", "the":
			return "definite"
		default:
			return ""
		}
	}
	expected := map[meaning.NodeID]string{
		roles["experiencer"]: articleValue(parts[0]),
		roles["stimulus"]:    articleValue(parts[4]),
	}
	if expected[roles["experiencer"]] == "" || expected[roles["stimulus"]] == "" {
		t.Errorf("candidate %q uses an unknown article in %q", candidate.ID, candidate.Text)
		return
	}
	seen := make(map[meaning.NodeID]bool, 2)
	if len(candidate.Assumptions) != 2 {
		t.Errorf("PL->EN candidate %q has %d assumptions; want exactly two determiner choices", candidate.ID, len(candidate.Assumptions))
	}
	for _, assumption := range candidate.Assumptions {
		if assumption.Field != "determiner" || len(assumption.Path) != 2 || assumption.Path[0] != event.ID {
			t.Errorf("PL->EN candidate %q has an unlicensed assumption: %+v", candidate.ID, assumption)
			continue
		}
		nodeID := assumption.Path[1]
		wantValue, roleExists := expected[nodeID]
		if !roleExists || seen[nodeID] || assumption.Value != wantValue || assumption.Reason == "" {
			t.Errorf("PL->EN candidate %q determiner assumption is not tied to the article/role: %+v; expected=%q", candidate.ID, assumption, wantValue)
			continue
		}
		seen[nodeID] = true
	}
	if len(seen) != 2 {
		t.Errorf("PL->EN candidate %q does not license exactly one determiner assumption per role: %+v", candidate.ID, candidate.Assumptions)
	}
}

func englishCompositionTextSet(subject, object string) []string {
	texts := make([]string, 0, 4)
	for _, subjectArticle := range []string{"A", "The"} {
		for _, objectArticle := range []string{"a", "the"} {
			texts = append(texts, subjectArticle+" black "+subject+" sees "+objectArticle+" "+object+".")
		}
	}
	sort.Strings(texts)
	return texts
}
