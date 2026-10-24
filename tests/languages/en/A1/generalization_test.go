package a1en_test

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
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("en"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: English analyzer unavailable")
	}
	cases := []struct{ input, fixture string }{
		{"This is a dog", "identification.classification.this"},
		{"This dog is black.", "identification.black.this"},
		{"This is not a dog.", "negation.classification.this"},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			want := dogIdentificationFixtureEN(t, tc.fixture)
			if err := meaning.ValidateMeaning(want.Graph, runtime.CoreGraph(), meaning.DefaultBudget()); err != nil {
				t.Fatalf("test fixture invalid against CORE: %v", err)
			}
			if err := meaning.ValidateReferences(want.Graph, want.Context, meaning.DefaultBudget()); err != nil {
				t.Fatalf("test fixture context invalid: %v", err)
			}
			result, err := analyzer.Analyze(context.Background(), tc.input)
			logEN(t, tc.input, result, err)
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
		{"identification.classification.this", "This is a dog."},
		{"identification.black.this", "This dog is black."},
		{"negation.classification.this", "This is not a dog."},
	}
	target := targetEN(t)
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			fixture := dogIdentificationFixtureEN(t, tc.fixture)
			candidate := candidateEN(fixture.Graph, "I-EN-"+tc.fixture, fixture.Context)
			validateCandidateEN(t, candidate)
			result, err := target.Realize(context.Background(), candidate)
			logEN(t, "independent I->EN "+tc.fixture, result, err)
			if err != nil || !sameTextEN(textsEN(result.Alternatives), []string{tc.text}) {
				t.Fatalf("FAIL: expected only %q; outcome=%s texts=%v error=%v", tc.text, result.Outcome, textsEN(result.Alternatives), err)
			}
			for _, alternative := range result.Alternatives {
				for _, assumption := range alternative.Assumptions {
					if assumption.Field != "person" {
						t.Errorf("FAIL: unexpected assumption for known proximal dog Meaning: %+v", assumption)
					}
				}
				if len(alternative.Losses) != 0 {
					t.Errorf("FAIL: native English fixture incurred a loss: %+v", alternative.Losses)
				}
			}
		})
	}
}

func TestGeneralizationCountAndArticleRefusal(t *testing.T) {
	target := targetEN(t)
	for _, tc := range []struct {
		name, concept string
		multiplicity  meaning.Multiplicity
	}{
		{"mass water with unknown multiplicity", "water", ""},
		{"mass money with unknown multiplicity", "money", ""},
		{"count eye without licensed a article", "eye", meaning.MultiplicityOne},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := classificationConceptEN(t, tc.concept, tc.multiplicity)
			candidate := candidateEN(fixture.Graph, "I-EN-refusal-"+tc.concept, fixture.Context)
			validateCandidateEN(t, candidate)
			result, err := target.Realize(context.Background(), candidate)
			logEN(t, "I->EN refusal "+tc.name, result, err)
			assertNoTextEN(t, result, err)
		})
	}
}

func dogIdentificationFixtureEN(t *testing.T, id string) interlingua.IdentificationMeaning {
	t.Helper()
	var fixture interlingua.IdentificationMeaning
	if id == "negation.classification.this" {
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
	graph, err := graphConceptEN(fixture.Graph, "dog", meaning.MultiplicityOne)
	if err != nil {
		t.Fatalf("replace fixture cat with dog: %v", err)
	}
	fixture.Graph = graph
	return fixture
}

func classificationConceptEN(t *testing.T, concept string, multiplicity meaning.Multiplicity) interlingua.IdentificationMeaning {
	t.Helper()
	fixture, ok, err := interlingua.Identification("identification.classification.this")
	if err != nil || !ok {
		t.Fatalf("load proximal classification fixture: ok=%t error=%v", ok, err)
	}
	graph, err := graphConceptEN(fixture.Graph, concept, multiplicity)
	if err != nil {
		t.Fatalf("replace classification concept with %s: %v", concept, err)
	}
	fixture.Graph = graph
	return fixture
}

func graphConceptEN(source meaning.Meaning, concept string, multiplicity meaning.Multiplicity) (meaning.Meaning, error) {
	conceptID, err := ids.NewConceptID(concept)
	if err != nil {
		return meaning.Meaning{}, err
	}
	nodes := append([]meaning.Node(nil), source.Nodes...)
	found := false
	for index := range nodes {
		if nodes[index].Kind == meaning.KindReferent {
			nodes[index].Concept = meaning.ConceptRefFromID(conceptID)
			nodes[index].Multiplicity = multiplicity
			found = true
		}
	}
	if !found {
		return meaning.Meaning{}, fmt.Errorf("fixture has no referent to update")
	}
	return meaning.Build(source.Roots, nodes, meaning.DefaultBudget())
}
