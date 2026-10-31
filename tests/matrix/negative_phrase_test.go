package matrix

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	langruntime "github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// These graphs are authored independently of the source adapters. Participant
// identity, modifier attachment and scope are explicit even for repeated concepts.
func negativePhraseMeaning(t *testing.T, tense meaning.TemporalFrame, subject, object meaning.Multiplicity) semantic.MeaningCandidate {
	t.Helper()
	g := meaning.Meaning{SchemaVersion: 1, Roots: []meaning.NodeID{"neg", "color"}, Nodes: []meaning.Node{
		{ID: "teacher", Kind: meaning.KindReferent, Concept: "concept:teacher", Multiplicity: subject},
		{ID: "cat", Kind: meaning.KindReferent, Concept: "concept:cat", Multiplicity: object},
		{ID: "seeing", Kind: meaning.KindEvent, Concept: "concept:seeing", TemporalFrame: tense, Participants: []meaning.Participant{{Role: "experiencer", Target: "teacher"}, {Role: "stimulus", Target: "cat"}}},
		{ID: "color", Kind: meaning.KindState, Concept: "concept:black", Participants: []meaning.Participant{{Role: "bearer", Target: "cat"}}},
		{ID: "neg", Kind: meaning.KindScope, Operator: meaning.OperatorNegation, Target: "seeing"},
	}}
	if err := meaning.ValidateMeaning(g, Runtime(t).CoreGraph(), meaning.DefaultBudget()); err != nil {
		t.Fatal(err)
	}
	return semantic.MeaningCandidate{ID: "independent-negative-phrase", SemanticID: "independent-negative-phrase", Meaning: &g, Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: "independent-negative-phrase", Outcome: status.OutcomeOK}}
}

func negativePhraseTexts(result engine.SurfaceResult) []string {
	var texts []string
	for _, alternative := range result.Alternatives {
		if alternative.Text != "" {
			texts = append(texts, alternative.Text)
		}
	}
	sort.Strings(texts)
	return texts
}

func TestNegativePhraseIndependentComposition(t *testing.T) {
	rt := Runtime(t)
	for _, tc := range []struct {
		name            string
		tense           meaning.TemporalFrame
		subject, object meaning.Multiplicity
		english, polish []string
	}{
		{"singular-present", meaning.TemporalFramePresent, meaning.MultiplicityOne, meaning.MultiplicityOne,
			[]string{"A teacher does not see a black cat.", "A teacher does not see the black cat.", "The teacher does not see a black cat.", "The teacher does not see the black cat."},
			[]string{"Nauczyciel nie widzi czarnego kota.", "Czarnego kota nie widzi nauczyciel."}},
		{"plural-past", meaning.TemporalFramePast, meaning.MultiplicityMany, meaning.MultiplicityMany,
			[]string{"Teachers did not see black cats.", "Teachers did not see the black cats.", "The teachers did not see black cats.", "The teachers did not see the black cats."},
			[]string{"Nauczyciele nie widzieli czarnych kotów.", "Czarnych kotów nie widzieli nauczyciele."}},
		{"mixed-present", meaning.TemporalFramePresent, meaning.MultiplicityMany, meaning.MultiplicityOne,
			[]string{"Teachers do not see a black cat.", "Teachers do not see the black cat.", "The teachers do not see a black cat.", "The teachers do not see the black cat."},
			[]string{"Nauczyciele nie widzą czarnego kota.", "Czarnego kota nie widzą nauczyciele."}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := negativePhraseMeaning(t, tc.tense, tc.subject, tc.object)
			before, _ := json.Marshal(input)
			for _, lang := range []features.LanguageID{"en", "pl"} {
				target, _ := rt.TargetRealizerFor(lang)
				result, err := target.Realize(context.Background(), input)
				if err != nil {
					t.Fatal(err)
				}
				want := append([]string(nil), tc.english...)
				if lang == "pl" {
					want = append([]string(nil), tc.polish...)
				}
				sort.Strings(want)
				got := negativePhraseTexts(result)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s texts=%q want=%q; status=%s diagnostics=%+v", lang, got, want, result.Outcome, result.Diagnostics)
				}
				ids := map[string]bool{}
				for _, alt := range result.Alternatives {
					if ids[alt.ID] {
						t.Errorf("duplicate surface identity: %s", alt.ID)
					}
					ids[alt.ID] = true
					for _, record := range alt.Evidence.Records {
						if strings.Contains(record.VariantID.String(), ".source.") {
							t.Error("analysis-only variant entered realization")
						}
					}
					for _, loss := range alt.Losses {
						if loss.Field == "determiner" {
							t.Error("independent Meaning acquired an invented source article loss")
						}
					}
				}
			}
			after, _ := json.Marshal(input)
			if string(before) != string(after) {
				t.Error("realization mutated independent IR or evidence")
			}
		})
	}
}

func TestNegativePhraseSourceAgreementAndDerivations(t *testing.T) {
	rt := Runtime(t)
	source, _ := rt.SourceAnalyzerFor("en")
	pipeline := source.(*langruntime.SourcePipeline)
	target, _ := rt.TargetRealizerFor("en")
	for _, tc := range []struct {
		input           string
		tense           meaning.TemporalFrame
		subject, object meaning.Multiplicity
	}{
		{"The teachers did not see the black cats.", meaning.TemporalFramePast, meaning.MultiplicityMany, meaning.MultiplicityMany},
		{"Teachers do not see black cats.", meaning.TemporalFramePresent, meaning.MultiplicityMany, meaning.MultiplicityMany},
		{"A teacher does not see a black cat.", meaning.TemporalFramePresent, meaning.MultiplicityOne, meaning.MultiplicityOne},
	} {
		t.Run(tc.input, func(t *testing.T) {
			want := negativePhraseMeaning(t, tc.tense, tc.subject, tc.object)
			syntax, err := pipeline.AnalyzeSyntax(context.Background(), tc.input)
			if err != nil {
				t.Fatal(err)
			}
			for _, node := range syntax.Forest.Nodes {
				ids := map[string]bool{}
				for _, alt := range node.Alternatives {
					if ids[alt.DerivationID] {
						t.Fatalf("same complete derivation retained twice: %s", alt.DerivationID)
					}
					ids[alt.DerivationID] = true
				}
			}
			result, err := source.Analyze(context.Background(), tc.input)
			if err != nil || len(result.Candidates) == 0 {
				t.Fatalf("source unavailable: %+v err=%v", result, err)
			}
			for _, candidate := range result.Candidates {
				if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
					t.Fatalf("not a complete Meaning: %+v", candidate)
				}
				matches, err := GraphMatches(*want.Meaning, *candidate.Meaning)
				if err != nil || !matches {
					t.Fatalf("wrong independent roles/modifier/scope: %+v err=%v", candidate.Meaning, err)
				}
				realized, err := target.Realize(context.Background(), candidate)
				if err != nil || !reflect.DeepEqual(negativePhraseTexts(realized), []string{tc.input}) {
					t.Errorf("known articles did not preserve exactly one surface: %q err=%v", negativePhraseTexts(realized), err)
				}
				for _, alt := range realized.Alternatives {
					for _, a := range alt.Assumptions {
						if a.Field == "determiner" {
							t.Error("known source article was assumed again")
						}
					}
				}
			}
		})
	}
	for _, input := range []string{
		"The teachers does not see the black cats.", "The teacher do not see the black cat.",
		"A teachers do not see black cats.", "Teacher does not see the black cat.",
		"The teacher does not see an black cat.", "The teacher did not sees the black cat.",
	} {
		t.Run("refuse/"+input, func(t *testing.T) {
			result, err := source.Analyze(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			for _, c := range result.Candidates {
				if c.Status == status.OutcomeOK && c.Meaning != nil {
					t.Errorf("invalid agreement/article received an OK Meaning: %q", input)
				}
			}
		})
	}
}

func TestNegativePhraseKeepsRepeatedConceptsDistinct(t *testing.T) {
	rt := Runtime(t)
	want := negativePhraseMeaning(t, meaning.TemporalFramePast, meaning.MultiplicityMany, meaning.MultiplicityMany)
	want.Meaning.Nodes[0].Concept = "concept:cat"
	if err := meaning.ValidateMeaning(*want.Meaning, rt.CoreGraph(), meaning.DefaultBudget()); err != nil {
		t.Fatal(err)
	}
	source, _ := rt.SourceAnalyzerFor("en")
	result, err := source.Analyze(context.Background(), "Cats did not see black cats.")
	if err != nil || len(result.Candidates) == 0 {
		t.Fatalf("source unavailable: %+v err=%v", result, err)
	}
	for _, c := range result.Candidates {
		if c.Meaning == nil || c.Status != status.OutcomeOK {
			t.Fatalf("source must retain a complete reading: %+v", c)
		}
		matches, err := GraphMatches(*want.Meaning, *c.Meaning)
		if err != nil || !matches {
			t.Errorf("same ConceptID merged two participants or moved the modifier: %+v err=%v", c.Meaning, err)
		}
	}
	// Swapping participants while keeping their own modifier links cannot be
	// repaired by reattaching the property just to obtain a surface candidate.
	want.Meaning.Nodes[2].Participants[0].Target = "cat"
	want.Meaning.Nodes[2].Participants[1].Target = "teacher"
	for _, lang := range []features.LanguageID{"en", "pl"} {
		target, _ := rt.TargetRealizerFor(lang)
		r, e := target.Realize(context.Background(), want)
		if e == nil {
			for _, text := range negativePhraseTexts(r) {
				if strings.Contains(text, "see black cats") || strings.Contains(text, "see the black cats") || strings.Contains(text, "czarnych kotów") {
					t.Errorf("subject attachment projected to the object: %q", text)
				}
			}
		}
	}
}

func TestNegativePhraseDoesNotReattachScopeOrModifier(t *testing.T) {
	rt := Runtime(t)
	for _, mutation := range []string{"modifier-on-subject", "negate-property", "missing-object"} {
		t.Run(mutation, func(t *testing.T) {
			input := negativePhraseMeaning(t, meaning.TemporalFramePast, meaning.MultiplicityMany, meaning.MultiplicityMany)
			// Do not freeze a historical unsupported boundary. Future valid
			// realizations may appear, but not by reusing the original meaning's
			// surfaces after changing its participants, attachment or scope.
			forbidden := map[features.LanguageID]map[string]bool{}
			for _, lang := range []features.LanguageID{"en", "pl"} {
				target, _ := rt.TargetRealizerFor(lang)
				base, err := target.Realize(context.Background(), input)
				if err != nil || len(negativePhraseTexts(base)) == 0 {
					t.Fatalf("control realization unavailable: %+v err=%v", base, err)
				}
				forbidden[lang] = map[string]bool{}
				for _, text := range negativePhraseTexts(base) {
					forbidden[lang][text] = true
				}
			}
			switch mutation {
			case "modifier-on-subject":
				input.Meaning.Nodes[3].Participants[0].Target = "teacher"
			case "negate-property":
				input.Meaning.Nodes[4].Target = "color"
				input.Meaning.Roots = []meaning.NodeID{"neg", "seeing"}
			case "missing-object":
				input.Meaning.Nodes[2].Participants = input.Meaning.Nodes[2].Participants[:1]
			}
			for _, lang := range []features.LanguageID{"en", "pl"} {
				target, _ := rt.TargetRealizerFor(lang)
				r, err := target.Realize(context.Background(), input)
				if err == nil {
					for _, text := range negativePhraseTexts(r) {
						if forbidden[lang][text] {
							t.Errorf("changed meaning silently projected to the original text: %q", text)
						}
					}
				}
			}
		})
	}
}

func TestNegativePhraseTransferLocalArticles(t *testing.T) {
	rt := Runtime(t)
	g := meaning.Meaning{SchemaVersion: 1, Roots: []meaning.NodeID{"neg"}, Nodes: []meaning.Node{
		{ID: "giver", Kind: meaning.KindReferent, Concept: "concept:teacher", Multiplicity: meaning.MultiplicityMany},
		{ID: "recipient", Kind: meaning.KindReferent, Concept: "concept:doctor", Multiplicity: meaning.MultiplicityOne},
		{ID: "theme", Kind: meaning.KindReferent, Concept: "concept:apple_fruit", Multiplicity: meaning.MultiplicityOne},
		{ID: "event", Kind: meaning.KindEvent, Concept: "concept:giving", TemporalFrame: meaning.TemporalFramePast, Participants: []meaning.Participant{{Role: "giver", Target: "giver"}, {Role: "recipient", Target: "recipient"}, {Role: "theme", Target: "theme"}}},
		{ID: "neg", Kind: meaning.KindScope, Operator: meaning.OperatorNegation, Target: "event"},
	}}
	if err := meaning.ValidateMeaning(g, rt.CoreGraph(), meaning.DefaultBudget()); err != nil {
		t.Fatal(err)
	}
	c := semantic.MeaningCandidate{ID: "independent-negative-transfer", SemanticID: "independent-negative-transfer", Meaning: &g, Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: "independent-negative-transfer", Outcome: status.OutcomeOK}}
	target, _ := rt.TargetRealizerFor("en")
	r, err := target.Realize(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, subject := range []string{"Teachers", "The teachers"} {
		for _, recipient := range []string{"a doctor", "the doctor"} {
			for _, theme := range []string{"an apple", "the apple"} {
				want = append(want, subject+" did not give "+recipient+" "+theme+".", subject+" did not give "+theme+" to "+recipient+".")
			}
		}
	}
	sort.Strings(want)
	got := negativePhraseTexts(r)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("three-role phrase composition: got=%q want=%q diagnostics=%+v", got, want, r.Diagnostics)
	}
	for _, a := range r.Alternatives {
		if len(a.Assumptions) != 3 {
			t.Errorf("want one article assumption per exact role, got %+v", a.Assumptions)
		}
	}
	// Both clause orders and bare plurals are consumed by the same approved frame.
	source, _ := rt.SourceAnalyzerFor("en")
	for _, input := range []string{"Teachers did not give an apple to a doctor.", "Teachers did not give a doctor an apple."} {
		r, err := source.Analyze(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		matched := false
		for _, c := range r.Candidates {
			if c.Meaning != nil && (c.Status == status.OutcomeOK || c.Status == status.OutcomeIncomplete) {
				// Apple has a retained, unbound competing lexical sense. Keep that
				// incompleteness visible; it must not erase the licensed giving reading.
				if c.Status == status.OutcomeIncomplete {
					lexicalGap := false
					for _, d := range c.Evidence.Diagnostics {
						lexicalGap = lexicalGap || d.Code == "UNRESOLVED_ROLE_SENSE"
					}
					if !lexicalGap {
						t.Errorf("unexpected partial transfer: %+v", c.Evidence.Diagnostics)
					}
				}
				same, e := GraphMatches(g, *c.Meaning)
				matched = matched || e == nil && same
			}
		}
		if !matched {
			t.Errorf("transfer did not preserve independently authored roles: %s status=%s", input, r.Outcome)
		}
	}
}

func TestNegativePhraseAspectLossRequiresSourceEvidence(t *testing.T) {
	rt := Runtime(t)
	source, _ := rt.SourceAnalyzerFor("pl")
	target, _ := rt.TargetRealizerFor("en")
	for _, input := range []string{"Matka nie piła wody.", "Lekarze nie czytali listu.", "Pies nie widział kota."} {
		t.Run(input, func(t *testing.T) {
			analyzed, err := source.Analyze(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			checked := 0
			for _, c := range analyzed.Candidates {
				if c.Meaning == nil || c.Status != status.OutcomeOK {
					continue
				}
				var event meaning.Node
				for _, n := range c.Meaning.Nodes {
					if n.Kind == meaning.KindEvent {
						event = n
					}
				}
				if event.ID == "" {
					t.Fatal("source has no event")
				}
				checked++
				r, err := target.Realize(context.Background(), c)
				if err != nil || len(negativePhraseTexts(r)) == 0 {
					t.Fatalf("target unavailable: %+v err=%v", r, err)
				}
				for _, a := range r.Alternatives {
					if a.Text == "" {
						continue
					}
					count := 0
					for _, loss := range a.Losses {
						if loss.Field == "aspect" {
							count++
							if len(loss.Path) == 0 || loss.Path[len(loss.Path)-1] != event.ID || loss.Outcome != status.OutcomeIncomplete {
								t.Errorf("aspect loss not on the exact past event: %+v", loss)
							}
						}
					}
					if count != 1 {
						t.Errorf("want exactly one explicit aspect loss, got %d: %+v", count, a.Losses)
					}
				}
				// An IR graph and its context remain independently realizable, but
				// carry no aspect merely because they originated in a Polish test.
				c.Evidence = semantic.EvidenceTrace{AlternativeID: c.ID, Outcome: status.OutcomeOK}
				without, err := target.Realize(context.Background(), c)
				if err != nil || len(negativePhraseTexts(without)) == 0 {
					t.Fatalf("standalone IR unavailable: %+v err=%v", without, err)
				}
				for _, a := range without.Alternatives {
					for _, loss := range a.Losses {
						if loss.Field == "aspect" {
							t.Errorf("invented aspect loss without source evidence: %+v", loss)
						}
					}
				}
			}
			if checked == 0 {
				t.Error("no complete source checked")
			}
		})
	}
}
