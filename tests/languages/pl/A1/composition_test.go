package a1pl_test

import (
	"context"
	"sort"
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionAnalysis(t *testing.T) {
	engine := matrix.Runtime(t)
	analyzer, ok := engine.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: Polish source analyzer is unavailable")
	}
	cases := []struct {
		id, input string
	}{
		{"composition.black-cat-sees-dog", "Czarny kot widzi psa."},
		{"composition.black-dog-sees-cat", "Czarny pies widzi kota."},
		{"composition.black-bird-sees-dog", "Czarny ptak widzi psa."},
		{"composition.black-cat-sees-cat", "Czarny kot widzi kota."},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			want, found, err := interlingua.Composition(tc.id)
			if err != nil || !found {
				t.Fatalf("load independent Composition fixture: found=%t error=%v", found, err)
			}
			if err := meaning.ValidateMeaning(want, engine.CoreGraph(), meaning.DefaultBudget()); err != nil {
				t.Fatalf("INFRA: authored Meaning does not validate against CORE: %v", err)
			}
			result, analyzeErr := analyzer.Analyze(context.Background(), tc.input)
			logPL(t, tc.input, result, analyzeErr)
			if analyzeErr != nil {
				t.Fatalf("GAP_ANALYZE: Polish composition analysis failed: %v", analyzeErr)
			}
			if len(result.Candidates) == 0 || result.Outcome == status.OutcomeUnsupported {
				t.Fatalf("GAP_ANALYZE: no supported Polish composition candidates (status=%s)", result.Outcome)
			}
			for _, candidate := range result.Candidates {
				if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
					t.Errorf("FAIL: candidate %q has no complete Meaning (status=%s)", candidate.ID, candidate.Status)
					continue
				}
				matches, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
				if compareErr != nil || !matches {
					t.Errorf("FAIL: candidate %q differs from authored full graph (matches=%t error=%v)", candidate.ID, matches, compareErr)
				}
				for index, candidateContext := range candidate.Contexts {
					if err := meaning.ValidateReferences(candidateContext.Meaning, candidateContext.Context, meaning.DefaultBudget()); err != nil {
						t.Errorf("FAIL: candidate %q context %d is invalid: %v", candidate.ID, index, err)
						continue
					}
					equal, err := meaning.CanonicalEqual(candidateContext.Meaning, *candidate.Meaning, meaning.DefaultBudget())
					if err != nil || !equal || len(candidateContext.Context.Referents) != 0 {
						t.Errorf("FAIL: candidate %q context %d differs from unresolved source Meaning: equal=%t error=%v", candidate.ID, index, equal, err)
					}
					for _, mention := range candidateContext.Context.Mentions {
						if mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor != nil {
							t.Errorf("FAIL: candidate %q introduced resolved identity or deixis: %+v", candidate.ID, mention)
						}
					}
				}
			}
		})
	}
}

func TestCompositionRealization(t *testing.T) {
	realizer := targetPL(t)
	cases := []struct {
		id, text string
	}{
		{"composition.black-cat-sees-dog", "Czarny kot widzi psa."},
		{"composition.black-dog-sees-cat", "Czarny pies widzi kota."},
		{"composition.black-bird-sees-dog", "Czarny ptak widzi psa."},
		{"composition.black-cat-sees-cat", "Czarny kot widzi kota."},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			graph, found, err := interlingua.Composition(tc.id)
			if err != nil || !found {
				t.Fatalf("load independent Composition fixture: found=%t error=%v", found, err)
			}
			graph = reorderedCompositionMeaning(t, graph)
			candidate := meaningCandidatePL(graph, "I-PL-"+tc.id, meaning.RequestContext{})
			validateCandidatePL(t, candidate)
			result, realizeErr := realizer.Realize(context.Background(), candidate)
			logPL(t, "I->PL "+tc.id, result, realizeErr)
			if realizeErr != nil {
				t.Fatalf("GAP_REALIZE: Polish target realization failed: %v", realizeErr)
			}
			if result.Outcome != status.OutcomeOK && result.Outcome != status.OutcomeIncomplete || !sameStrings(textSet(result.Alternatives), []string{tc.text}) {
				t.Errorf("FAIL: got outcome=%s texts=%v, want exactly %q", result.Outcome, textSet(result.Alternatives), tc.text)
			}
			assertNoAttributeAssumptions(t, result.Alternatives)
		})
	}
}

func reorderedCompositionMeaning(t *testing.T, source meaning.Meaning) meaning.Meaning {
	t.Helper()
	names := map[meaning.NodeID]meaning.NodeID{
		"e": "m_event", "q": "a_property", "r1": "z_subject", "r2": "a_object",
	}
	graph := meaning.Meaning{SchemaVersion: source.SchemaVersion, Roots: make([]meaning.NodeID, len(source.Roots)), Nodes: make([]meaning.Node, len(source.Nodes))}
	for index, root := range source.Roots {
		graph.Roots[index] = names[root]
	}
	graph.Roots[0], graph.Roots[1] = graph.Roots[1], graph.Roots[0]
	for index, node := range source.Nodes {
		node.ID = names[node.ID]
		for participantIndex := range node.Participants {
			node.Participants[participantIndex].Target = names[node.Participants[participantIndex].Target]
		}
		graph.Nodes[index] = node
	}
	sort.Slice(graph.Nodes, func(left, right int) bool { return graph.Nodes[left].ID < graph.Nodes[right].ID })
	if graph.Nodes[0].ID != "a_object" || graph.Nodes[len(graph.Nodes)-1].ID != "z_subject" {
		t.Fatalf("test setup did not put the object ID before the subject ID: %v", graph.Nodes)
	}
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		t.Fatalf("reordered composition graph is invalid: %v", err)
	}
	return graph
}
