package a1pl_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestGeneralizationDogAnalysis(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: Polish analyzer unavailable")
	}
	cases := []struct{ input, fixture string }{
		{"To jest pies.", "identification.classification.to"},
		{"Ten pies jest czarny.", "identification.black.ten"},
		{"To nie jest pies.", "negation.classification.to"},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			want := dogIdentificationFixturePL(t, tc.fixture)
			if err := meaning.ValidateMeaning(want.Graph, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
				t.Fatalf("test fixture invalid against CORE: %v", err)
			}
			if err := meaning.ValidateReferences(want.Graph, want.Context, meaning.DefaultBudget()); err != nil {
				t.Fatalf("test fixture context invalid: %v", err)
			}
			result, err := analyzer.Analyze(context.Background(), tc.input)
			logPL(t, tc.input, result, err)
			if err != nil || result.Outcome != status.OutcomeOK || len(result.Candidates) == 0 {
				t.Fatalf("GAP_ANALYZE: expected an OK analysis with candidates, got %s (%v)", result.Outcome, err)
			}
			for _, candidate := range result.Candidates {
				if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
					t.Errorf("FAIL: unlicensed/non-OK candidate %q status=%s", candidate.ID, candidate.Status)
					continue
				}
				matches, compareErr := matrix.IdentificationMatches(want, candidate)
				if compareErr != nil || !matches {
					t.Errorf("FAIL: candidate %q differs from authored dog fixture (matches=%t error=%v)", candidate.ID, matches, compareErr)
				}
			}
		})
	}
}

func TestGeneralizationDogRealization(t *testing.T) {
	cases := []struct{ fixture, text string }{
		{"identification.classification.to", "To jest pies."},
		{"identification.black.ten", "Ten pies jest czarny."},
		{"negation.classification.to", "To nie jest pies."},
	}
	target := targetPL(t)
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixture := dogIdentificationFixturePL(t, tc.fixture)
			candidate := meaningCandidatePL(fixture.Graph, "I-PL-"+tc.fixture, fixture.Context)
			validateCandidatePL(t, candidate)
			result, err := target.Realize(context.Background(), candidate)
			logPL(t, "independent I->PL "+tc.fixture, result, err)
			if err != nil || !sameStrings(textSet(result.Alternatives), []string{tc.text}) {
				t.Fatalf("FAIL: expected only %q; outcome=%s texts=%v error=%v", tc.text, result.Outcome, textSet(result.Alternatives), err)
			}
			for _, alternative := range result.Alternatives {
				for _, assumption := range alternative.Assumptions {
					if assumption.Field != "person" {
						t.Errorf("FAIL: unexpected assumption for known dog Meaning: %+v", assumption)
					}
				}
				if len(alternative.Losses) != 0 {
					t.Errorf("FAIL: native Polish fixture incurred a loss: %+v", alternative.Losses)
				}
			}
		})
	}
}

func dogIdentificationFixturePL(t *testing.T, id string) interlingua.IdentificationMeaning {
	t.Helper()
	var fixture interlingua.IdentificationMeaning
	if id == "negation.classification.to" {
		graph, contextFixture, ok, err := interlingua.Negation(id)
		if err != nil || !ok || contextFixture == nil {
			t.Fatalf("load %s fixture: ok=%t context=%t error=%v", id, ok, contextFixture != nil, err)
		}
		fixture = *contextFixture
		fixture.Graph = graph
	} else {
		var ok bool
		var err error
		fixture, ok, err = interlingua.Identification(id)
		if err != nil || !ok {
			t.Fatalf("load %s fixture: ok=%t error=%v", id, ok, err)
		}
	}
	graph, err := dogGraphPL(fixture.Graph)
	if err != nil {
		t.Fatalf("replace fixture cat with dog: %v", err)
	}
	fixture.Graph = graph
	return fixture
}

func dogGraphPL(source meaning.Meaning) (meaning.Meaning, error) {
	dog, err := ids.NewConceptID("dog")
	if err != nil {
		return meaning.Meaning{}, err
	}
	nodes := append([]meaning.Node(nil), source.Nodes...)
	found := false
	for index := range nodes {
		if nodes[index].Kind == meaning.KindReferent {
			nodes[index].Concept = meaning.ConceptRefFromID(dog)
			found = true
		}
	}
	if !found {
		return meaning.Meaning{}, fmt.Errorf("fixture has no referent to generalize")
	}
	return meaning.Build(source.Roots, nodes, meaning.DefaultBudget())
}
