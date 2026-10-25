package translation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestMetamorphicProperties(t *testing.T) {
	tests := []struct {
		name, lang, baseText, changedText string
		baseFixture, changedFixture       func() (meaning.Meaning, error)
		normalizeChanged                  func(meaning.Meaning) (meaning.Meaning, error)
		mustChange                        bool
	}{
		{"role_swap_PL", "pl", "Nauczyciel widzi psa.", "Pies widzi nauczyciela.", seeingFixture("seeing.teacher-dog"), seeingFixture("seeing.dog-teacher"), swapRoles, true},
		{"role_swap_EN", "en", "The teacher sees the dog.", "The dog sees the teacher.", seeingFixture("seeing.teacher-dog"), seeingFixture("seeing.dog-teacher"), swapRoles, true},
		{"number_PL", "pl", "Nauczyciel widzi psa.", "Nauczyciele widzą psa.", seeingFixture("seeing.teacher-dog"), changedSeeingNumber, normalizeTeacherMultiplicity, true},
		{"number_EN", "en", "The teacher sees the dog.", "The teachers see the dog.", seeingFixture("seeing.teacher-dog"), changedSeeingNumber, normalizeTeacherMultiplicity, true},
		{"tense_PL", "pl", "Nauczyciel czyta wiadomość.", "Nauczyciel czytał wiadomość.", readingFixture("reading.teacher-message.present"), readingFixture("reading.teacher-message.past"), normalizeEventFrame, true},
		{"tense_EN", "en", "The teacher reads the message.", "The teacher read the message.", readingFixture("reading.teacher-message.present"), readingFixture("reading.teacher-message.past"), normalizeEventFrame, true},
		{"polarity_PL", "pl", "Nauczyciel widzi psa.", "Nauczyciel nie widzi psa.", seeingFixture("seeing.teacher-dog"), negativeSeeingFixture, unwrapNegation, true},
		{"polarity_EN", "en", "The teacher sees the dog.", "The teacher does not see the dog.", seeingFixture("seeing.teacher-dog"), negativeSeeingFixture, unwrapNegation, true},
		{"SVO_to_OVS_PL", "pl", "Pies widzi nauczyciela.", "Nauczyciela widzi pies.", seeingFixture("seeing.dog-teacher"), seeingFixture("seeing.dog-teacher"), identityMeaning, false},
	}
	runtime := matrix.Runtime(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseWant, err := tc.baseFixture()
			if err != nil {
				t.Fatalf("INFRA: build authored base fixture: %v", err)
			}
			changedWant, err := tc.changedFixture()
			if err != nil {
				t.Fatalf("INFRA: build authored changed fixture: %v", err)
			}
			baseGot := analyzeMetamorphicFixture(t, runtime, features.LanguageID(tc.lang), tc.baseText, baseWant)
			changedGot := analyzeMetamorphicFixture(t, runtime, features.LanguageID(tc.lang), tc.changedText, changedWant)
			if baseGot == nil || changedGot == nil {
				return
			}
			normalized, err := tc.normalizeChanged(*changedGot)
			if err != nil {
				t.Errorf("FAIL: normalize only changed feature: %v", err)
				return
			}
			unchanged, compareErr := matrix.GraphMatches(*baseGot, normalized)
			if compareErr != nil || !unchanged {
				t.Errorf("FAIL: unchanged graph properties differ after normalization (matches=%t error=%v)", unchanged, compareErr)
			}
			changed, compareErr := matrix.GraphMatches(*baseGot, *changedGot)
			if compareErr != nil {
				t.Errorf("FAIL: compare source graphs: %v", compareErr)
				return
			}
			if tc.mustChange && changed {
				t.Errorf("FAIL: requested semantic property did not change the MeaningGraph")
			}
			if !tc.mustChange && !changed {
				t.Errorf("FAIL: word-order change altered MeaningGraph")
			}
		})
	}
}

func analyzeMetamorphicFixture(t *testing.T, runtime *engine.Engine, lang features.LanguageID, input string, want meaning.Meaning) *meaning.Meaning {
	t.Helper()
	analyzer, ok := runtime.SourceAnalyzerFor(lang)
	if !ok || analyzer == nil {
		t.Errorf("INFRA: %s source analyzer unavailable", lang)
		return nil
	}
	result, err := analyzer.Analyze(context.Background(), input)
	encoded, _ := json.Marshal(result)
	sha := os.Getenv("GOLINGUA_MATRIX_SHA")
	if sha == "" {
		sha = "UNRECORDED"
	}
	t.Logf("SHA=%s metamorphic source=%s input=%q result=%s error=%v", sha, lang, input, encoded, err)
	if err != nil {
		t.Errorf("GAP_ANALYZE: %s analysis failed: %v", lang, err)
		return nil
	}
	if result.Outcome == status.OutcomeUnsupported || len(result.Candidates) == 0 {
		t.Errorf("GAP_ANALYZE: %s returned %s without candidates", lang, result.Outcome)
		return nil
	}
	if result.Outcome != status.OutcomeOK && result.Outcome != status.OutcomeAmbiguous {
		t.Errorf("GAP_ANALYZE: %s returned incomplete outcome %s", lang, result.Outcome)
	}
	var matched *meaning.Meaning
	for _, candidate := range result.Candidates {
		if candidate.Meaning == nil || candidate.Status != status.OutcomeOK {
			t.Errorf("GAP_ANALYZE: candidate %q has no supported complete Meaning (status=%s)", candidate.ID, candidate.Status)
			continue
		}
		same, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
		if compareErr != nil || !same {
			t.Errorf("FAIL: candidate %q does not match independent fixture (matches=%t error=%v)", candidate.ID, same, compareErr)
			continue
		}
		copy := *candidate.Meaning
		matched = &copy
	}
	if matched == nil {
		t.Errorf("FAIL: no candidate matched independent fixture")
		return nil
	}
	return matched
}

func seeingFixture(id string) func() (meaning.Meaning, error) {
	return func() (meaning.Meaning, error) {
		graph, ok, err := interlingua.Seeing(id)
		if err != nil {
			return graph, err
		}
		if !ok {
			return graph, fmt.Errorf("missing Seeing fixture %q", id)
		}
		return graph, nil
	}
}

func readingFixture(id string) func() (meaning.Meaning, error) {
	return func() (meaning.Meaning, error) {
		graph, ok, err := interlingua.Reading(id)
		if err != nil {
			return graph, err
		}
		if !ok {
			return graph, fmt.Errorf("missing Reading fixture %q", id)
		}
		return graph, nil
	}
}

func changedSeeingNumber() (meaning.Meaning, error) {
	graph, ok, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !ok {
		return graph, fmt.Errorf("load Seeing: %t %w", ok, err)
	}
	return setMultiplicity(graph, "t", meaning.MultiplicityMany), nil
}

func negativeSeeingFixture() (meaning.Meaning, error) {
	graph, _, ok, err := interlingua.Negation("negation.seeing.teacher-dog")
	if err != nil || !ok {
		return graph, fmt.Errorf("load negated Seeing: %t %w", ok, err)
	}
	return graph, nil
}

func setMultiplicity(graph meaning.Meaning, id meaning.NodeID, value meaning.Multiplicity) meaning.Meaning {
	graph = cloneMetamorphicGraph(graph)
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == id {
			graph.Nodes[i].Multiplicity = value
		}
	}
	return graph
}

func normalizeTeacherMultiplicity(graph meaning.Meaning) (meaning.Meaning, error) {
	graph = cloneMetamorphicGraph(graph)
	for i := range graph.Nodes {
		if graph.Nodes[i].Kind == meaning.KindReferent && graph.Nodes[i].Concept == "concept:teacher" {
			graph.Nodes[i].Multiplicity = meaning.MultiplicityOne
		}
	}
	return graph, nil
}

func normalizeEventFrame(graph meaning.Meaning) (meaning.Meaning, error) {
	graph = cloneMetamorphicGraph(graph)
	for i := range graph.Nodes {
		if graph.Nodes[i].Kind == meaning.KindEvent {
			graph.Nodes[i].TemporalFrame = meaning.TemporalFramePresent
		}
	}
	return graph, nil
}

func swapRoles(graph meaning.Meaning) (meaning.Meaning, error) {
	graph = cloneMetamorphicGraph(graph)
	teacher, dog := meaning.NodeID(""), meaning.NodeID("")
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindReferent && node.Concept == "concept:teacher" {
			teacher = node.ID
		}
		if node.Kind == meaning.KindReferent && node.Concept == "concept:dog" {
			dog = node.ID
		}
	}
	if teacher == "" || dog == "" {
		return meaning.Meaning{}, fmt.Errorf("missing teacher/dog referent")
	}
	for i := range graph.Nodes {
		if graph.Nodes[i].Kind == meaning.KindEvent {
			for j := range graph.Nodes[i].Participants {
				p := &graph.Nodes[i].Participants[j]
				if p.Role == "experiencer" {
					p.Target = teacher
				} else if p.Role == "stimulus" {
					p.Target = dog
				}
			}
		}
	}
	return graph, nil
}

func unwrapNegation(graph meaning.Meaning) (meaning.Meaning, error) {
	graph = cloneMetamorphicGraph(graph)
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindScope && node.Operator == meaning.OperatorNegation {
			graph.Roots = []meaning.NodeID{node.Target}
			kept := graph.Nodes[:0]
			for _, candidate := range graph.Nodes {
				if candidate.ID != node.ID {
					kept = append(kept, candidate)
				}
			}
			graph.Nodes = kept
			return graph, nil
		}
	}
	return meaning.Meaning{}, fmt.Errorf("no negation scope to unwrap")
}

func identityMeaning(graph meaning.Meaning) (meaning.Meaning, error) { return graph, nil }

func cloneMetamorphicGraph(graph meaning.Meaning) meaning.Meaning {
	clone := graph
	clone.Roots = append([]meaning.NodeID(nil), graph.Roots...)
	clone.Nodes = append([]meaning.Node(nil), graph.Nodes...)
	for i := range clone.Nodes {
		clone.Nodes[i].Participants = append([]meaning.Participant(nil), graph.Nodes[i].Participants...)
	}
	return clone
}
