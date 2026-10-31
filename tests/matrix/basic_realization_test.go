package matrix

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

// Independently authored contracts: no source parse supplies the reference IR.
// Polish lexical plural drzwi intentionally leaves semantic multiplicity open.
func basicRealizationMeaning(t *testing.T, predicate, subject, object string, tense meaning.TemporalFrame, count meaning.Multiplicity, negative bool) semantic.MeaningCandidate {
	t.Helper()
	graph := meaning.Meaning{SchemaVersion: 1, Roots: []meaning.NodeID{"p"}, Nodes: []meaning.Node{
		{ID: "s", Kind: meaning.KindReferent, Concept: meaning.ConceptRef("concept:" + subject), Multiplicity: count},
		{ID: "p", Kind: meaning.KindEvent, Concept: meaning.ConceptRef("concept:" + predicate), TemporalFrame: tense},
	}}
	role := meaning.Role("agent")
	if predicate == "flying_motion" {
		role = "mover"
	}
	if predicate == "black" {
		graph.Nodes[1].Kind = meaning.KindState
		role = "bearer"
	}
	graph.Nodes[1].Participants = []meaning.Participant{{Role: role, Target: "s"}}
	if object != "" {
		multiplicity := meaning.MultiplicityOne
		if object == "door" {
			multiplicity = ""
		}
		graph.Nodes = append(graph.Nodes, meaning.Node{ID: "o", Kind: meaning.KindReferent, Concept: meaning.ConceptRef("concept:" + object), Multiplicity: multiplicity})
		graph.Nodes[1].Participants = append(graph.Nodes[1].Participants, meaning.Participant{Role: "object", Target: "o"})
	}
	if negative {
		graph.Roots = []meaning.NodeID{"not"}
		graph.Nodes = append(graph.Nodes, meaning.Node{ID: "not", Kind: meaning.KindScope, Operator: meaning.OperatorNegation, Target: "p"})
	}
	if err := meaning.ValidateMeaning(graph, Runtime(t).CoreGraph(), meaning.DefaultBudget()); err != nil {
		t.Fatal(err)
	}
	return semantic.MeaningCandidate{ID: "independent-basic-realization", SemanticID: "independent-basic-realization", Meaning: &graph, Contexts: []semantic.MeaningContext{{Meaning: graph, Context: meaning.RequestContext{ID: "independent-basic-realization"}}}, Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: "independent-basic-realization", Outcome: status.OutcomeOK}}
}

func TestBasicRealizationIndependentGaps(t *testing.T) {
	cases := []struct {
		name, predicate, subject, object string
		tense                            meaning.TemporalFrame
		count                            meaning.Multiplicity
		negative                         bool
		target                           features.LanguageID
		want                             []string
	}{
		{"open-door", "opening_object", "doctor", "door", meaning.TemporalFramePresent, meaning.MultiplicityOne, true, "en", []string{"A doctor does not open a door.", "A doctor does not open the door.", "A doctor does not open doors.", "A doctor does not open the doors.", "The doctor does not open a door.", "The doctor does not open the door.", "The doctor does not open doors.", "The doctor does not open the doors."}},
		{"close-window", "closing_object", "teacher", "window", meaning.TemporalFramePresent, meaning.MultiplicityOne, true, "en", []string{"A teacher does not close a window.", "A teacher does not close the window.", "The teacher does not close a window.", "The teacher does not close the window."}},
		{"close-door-past", "closing_object", "doctor", "door", meaning.TemporalFramePast, meaning.MultiplicityOne, true, "en", []string{"A doctor did not close a door.", "A doctor did not close the door.", "A doctor did not close doors.", "A doctor did not close the doors.", "The doctor did not close a door.", "The doctor did not close the door.", "The doctor did not close doors.", "The doctor did not close the doors."}},
		{"door-attribution-past", "black", "door", "", meaning.TemporalFramePast, "", false, "en", []string{"A door was black.", "The door was black.", "The doors were black.", "Doors were black."}},
		{"flying-past-one", "flying_motion", "bird", "", meaning.TemporalFramePast, meaning.MultiplicityOne, false, "pl", []string{"Ptak latał."}},
		{"flying-past-many", "flying_motion", "bird", "", meaning.TemporalFramePast, meaning.MultiplicityMany, false, "pl", []string{"Ptaki latały."}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := basicRealizationMeaning(t, tc.predicate, tc.subject, tc.object, tc.tense, tc.count, tc.negative)
			target, _ := Runtime(t).TargetRealizerFor(tc.target)
			result, err := target.Realize(context.Background(), input)
			want := append([]string(nil), tc.want...)
			sort.Strings(want)
			got := negativePhraseTexts(result)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("texts=%q want=%q outcome=%s diagnostics=%+v err=%v", got, want, result.Outcome, result.Diagnostics, err)
			}
			if result.Outcome == status.OutcomeInvalid || result.Outcome == status.OutcomeError || result.Outcome == status.OutcomeResourceLimit || result.Outcome == status.OutcomeUnsupported {
				t.Fatalf("required translation refused: %+v", result)
			}
			for _, alt := range result.Alternatives {
				for _, loss := range alt.Losses {
					if loss.Field == "determiner" || loss.Field == "aspect" {
						t.Errorf("independent IR acquired source loss: %+v", loss)
					}
				}
			}
		})
	}
}

func TestNegativeAnalysisOnlyFrameDoesNotLicenseCompactTarget(t *testing.T) {
	index, err := en.CompileValency()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []string{"opening_object", "closing_object"} {
		sourceOnly, compactTarget := false, false
		for _, b := range index.Bindings() {
			if b.SenseID.String() != "sense:en."+item+".event" {
				continue
			}
			if string(b.ValencyFrameID) == "en."+item+".transitive" && b.FamilyID.String() == "construction:en.transitive_negative" {
				sourceOnly = true
				if b.VariantID.String() == "construction:en.transitive_negative.svo.vp_auxiliary" {
					t.Errorf("source-only extension silently licensed target for %s", b.ValencyFrameID)
				}
			}
			if string(b.ValencyFrameID) == "en."+item+".transitive.negative" && b.VariantID.String() == "construction:en.transitive_negative.svo.vp_auxiliary" {
				compactTarget = true
			}
		}
		if !sourceOnly || !compactTarget {
			t.Errorf("missing explicit source/target routes for %s", item)
		}
	}
}

func TestBasicRealizationTranslationsPreserveOperators(t *testing.T) {
	for _, input := range []string{"Lekarz nie otwiera drzwi.", "Nauczyciel nie zamyka okna.", "Lekarz nie zamykał drzwi."} {
		t.Run(input, func(t *testing.T) {
			result, err := Runtime(t).Translate(context.Background(), engine.Request{From: "pl", To: "en", Text: input})
			if err != nil {
				t.Fatal(err)
			}
			texts := 0
			for _, c := range result.Candidates {
				if c.Text == "" {
					continue
				}
				texts++
				if strings.Contains(c.Text, "open door.") || strings.Contains(c.Text, "close door.") {
					t.Errorf("bare singular count noun: %s", c.Text)
				}
				if c.Meaning == nil {
					t.Fatal("text without Meaning")
				}
				found := false
				for _, n := range c.Meaning.Nodes {
					if n.Kind == meaning.KindScope && n.Operator == meaning.OperatorNegation {
						found = true
					}
				}
				if !found {
					t.Errorf("negative source lost Not: %s", c.Text)
				}
			}
			wantCount := 4
			if strings.Contains(input, "drzwi") {
				wantCount = 8
			}
			if texts != wantCount {
				t.Fatalf("got %d candidates, want %d independent article/number choices: %s %+v", texts, wantCount, result.Status, result.Diagnostics)
			}
		})
	}
}

func TestBasicRealizationSourceGraphs(t *testing.T) {
	for _, tc := range []struct {
		language                         features.LanguageID
		text, predicate, subject, object string
		tense                            meaning.TemporalFrame
		subjectCount, objectCount        meaning.Multiplicity
		negative                         bool
	}{
		{"pl", "Lekarz nie otwiera drzwi.", "opening_object", "doctor", "door", meaning.TemporalFramePresent, meaning.MultiplicityOne, "", true},
		{"en", "The doctor does not open a door.", "opening_object", "doctor", "door", meaning.TemporalFramePresent, meaning.MultiplicityOne, meaning.MultiplicityOne, true},
		{"en", "The doctor does not open doors.", "opening_object", "doctor", "door", meaning.TemporalFramePresent, meaning.MultiplicityOne, meaning.MultiplicityMany, true},
		{"pl", "Nauczyciel nie zamyka okna.", "closing_object", "teacher", "window", meaning.TemporalFramePresent, meaning.MultiplicityOne, meaning.MultiplicityOne, true},
		{"en", "The teacher does not close a window.", "closing_object", "teacher", "window", meaning.TemporalFramePresent, meaning.MultiplicityOne, meaning.MultiplicityOne, true},
		{"pl", "Lekarz nie zamykał drzwi.", "closing_object", "doctor", "door", meaning.TemporalFramePast, meaning.MultiplicityOne, "", true},
		{"en", "The doctor did not close doors.", "closing_object", "doctor", "door", meaning.TemporalFramePast, meaning.MultiplicityOne, meaning.MultiplicityMany, true},
		{"pl", "Drzwi były czarne.", "black", "door", "", meaning.TemporalFramePast, "", "", false},
		{"en", "The doors were black.", "black", "door", "", meaning.TemporalFramePast, meaning.MultiplicityMany, "", false},
		{"pl", "Ptak latał.", "flying_motion", "bird", "", meaning.TemporalFramePast, meaning.MultiplicityOne, "", false},
		{"pl", "Ptaki latały.", "flying_motion", "bird", "", meaning.TemporalFramePast, meaning.MultiplicityMany, "", false},
		{"en", "The bird flew.", "flying_motion", "bird", "", meaning.TemporalFramePast, meaning.MultiplicityOne, "", false},
		{"en", "The birds flew.", "flying_motion", "bird", "", meaning.TemporalFramePast, meaning.MultiplicityMany, "", false},
	} {
		t.Run(tc.text, func(t *testing.T) {
			want := basicRealizationMeaning(t, tc.predicate, tc.subject, tc.object, tc.tense, tc.subjectCount, tc.negative)
			if tc.object != "" {
				want.Meaning.Nodes[2].Multiplicity = tc.objectCount
			}
			source, _ := Runtime(t).SourceAnalyzerFor(tc.language)
			got, err := source.Analyze(context.Background(), tc.text)
			if err != nil {
				t.Fatal(err)
			}
			complete := 0
			for _, c := range got.Candidates {
				if c.Meaning == nil {
					continue
				}
				// English doctor has a separately recorded academic reading;
				// keep that unresolved sense visible rather than forcing medical.
				if tc.language == "en" && tc.subject == "doctor" {
					unresolved := false
					for _, d := range c.Evidence.Diagnostics {
						unresolved = unresolved || d.Code == "UNRESOLVED_ROLE_SENSE"
					}
					if c.Status != status.OutcomeIncomplete || !unresolved {
						t.Errorf("doctor ambiguity must remain explicit: %+v", c.Evidence)
					}
				} else if c.Status != status.OutcomeOK {
					t.Errorf("source has unexpected incomplete reading: %+v", c.Evidence)
				}
				complete++
				match, err := GraphMatches(*want.Meaning, *c.Meaning)
				if err != nil || !match {
					t.Errorf("source differs from independent IR: %+v err=%v", c.Meaning, err)
				}
			}
			if complete == 0 {
				t.Fatalf("no complete source graph: %+v", got)
			}
		})
	}
}

func TestBasicRealizationRejectsInvalidAgreement(t *testing.T) {
	for _, tc := range []struct {
		language features.LanguageID
		text     string
	}{
		{"en", "The doctor do not open a door."},
		{"en", "The doctors does not close the door."},
		{"en", "The doctor did not closed the door."},
		{"en", "The doctor does not open door."},
		{"en", "The doors was black."},
		{"en", "A doors were black."},
		{"pl", "Ptaki latał."},
		{"pl", "Ptak latały."},
		{"pl", "Drzwi był czarny."},
	} {
		t.Run(tc.text, func(t *testing.T) {
			source, _ := Runtime(t).SourceAnalyzerFor(tc.language)
			got, err := source.Analyze(context.Background(), tc.text)
			if err != nil {
				return
			}
			for _, c := range got.Candidates {
				if c.Status == status.OutcomeOK && c.Meaning != nil {
					t.Errorf("invalid agreement/article received complete Meaning: %q", tc.text)
				}
			}
		})
	}
}

func TestBasicRealizationDoesNotInventDoorCardinality(t *testing.T) {
	target, _ := Runtime(t).TargetRealizerFor("en")
	for _, count := range []meaning.Multiplicity{"", meaning.MultiplicityOne, meaning.MultiplicityMany} {
		t.Run(string(count), func(t *testing.T) {
			input := basicRealizationMeaning(t, "black", "door", "", meaning.TemporalFramePast, count, false)
			result, err := target.Realize(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			texts := negativePhraseTexts(result)
			want := 2
			if count == "" {
				want = 4
			}
			if len(texts) != want {
				t.Fatalf("texts=%q expected %d", texts, want)
			}
			for _, a := range result.Alternatives {
				plural := strings.Contains(a.Text, "doors ") || strings.HasPrefix(a.Text, "Doors ")
				if count == meaning.MultiplicityOne && plural || count == meaning.MultiplicityMany && !plural {
					t.Errorf("contradicted explicit count: %q", a.Text)
				}
				for _, v := range a.Assumptions {
					if v.Field == "person" {
						t.Errorf("nominal agreement is not an assumed speech participant: %+v", v)
					}
				}
				if count == "" {
					value := "singular"
					if plural {
						value = "plural"
					}
					found := false
					for _, v := range a.Assumptions {
						if v.Field == "number" && v.Value == value && len(v.Path) > 0 && v.Path[len(v.Path)-1] == "s" {
							found = true
						}
					}
					if !found {
						t.Errorf("missing referent-local number choice for %q", a.Text)
					}
				}
			}
		})
	}
}

// A source-conditioned loss is not permission to attach it indiscriminately:
// the same independent graph must not invent an English article or Polish aspect.
func TestBasicRealizationLossesFollowActualEvidence(t *testing.T) {
	for _, tc := range []struct {
		from, to    features.LanguageID
		text, field string
	}{
		{"en", "pl", "The bird flew.", "determiner"},
		{"en", "pl", "The birds flew.", "determiner"},
		{"pl", "en", "Ptak latał.", "aspect"},
		{"pl", "en", "Ptaki latały.", "aspect"},
		{"pl", "en", "Lekarz nie zamykał drzwi.", "aspect"},
	} {
		t.Run(tc.text, func(t *testing.T) {
			source, _ := Runtime(t).SourceAnalyzerFor(tc.from)
			target, _ := Runtime(t).TargetRealizerFor(tc.to)
			analyzed, err := source.Analyze(context.Background(), tc.text)
			if err != nil {
				t.Fatal(err)
			}
			checked := 0
			for _, c := range analyzed.Candidates {
				if c.Meaning == nil || c.Status != status.OutcomeOK {
					continue
				}
				checked++
				for _, withEvidence := range []bool{true, false} {
					input := c
					if !withEvidence {
						input.Evidence = semantic.EvidenceTrace{AlternativeID: c.ID, Outcome: status.OutcomeOK}
					}
					result, err := target.Realize(context.Background(), input)
					if err != nil || len(negativePhraseTexts(result)) == 0 {
						t.Fatalf("realization unavailable: %+v err=%v", result, err)
					}
					for _, a := range result.Alternatives {
						if a.Text == "" {
							continue
						}
						losses := 0
						for _, loss := range a.Losses {
							if loss.Field != tc.field {
								continue
							}
							losses++
							validPath := false
							for _, n := range c.Meaning.Nodes {
								if len(loss.Path) > 0 && loss.Path[len(loss.Path)-1] == n.ID {
									validPath = tc.field == "aspect" && n.Kind == meaning.KindEvent || tc.field == "determiner" && n.Kind == meaning.KindReferent
								}
							}
							if !validPath || loss.Outcome != status.OutcomeIncomplete {
								t.Errorf("wrong loss address/status: %+v", loss)
							}
						}
						want := 0
						if withEvidence {
							want = 1
						}
						if losses != want {
							t.Errorf("withEvidence=%t %q: got %d %s losses, want %d", withEvidence, a.Text, losses, tc.field, want)
						}
						if tc.to == "pl" {
							found := false
							for _, v := range a.Assumptions {
								if v.Field == "aspect" && v.Value == "imperfective" && len(v.Path) > 0 {
									for _, n := range c.Meaning.Nodes {
										if n.ID == v.Path[len(v.Path)-1] && n.Kind == meaning.KindEvent {
											found = true
										}
									}
								}
							}
							if !found {
								t.Errorf("missing explicit lexical aspect choice for %q", a.Text)
							}
						}
					}
				}
			}
			if checked == 0 {
				t.Fatal("no independently checked source")
			}
		})
	}
}

func TestBasicRealizationKeepsChangedRolesAndScope(t *testing.T) {
	target, _ := Runtime(t).TargetRealizerFor("en")
	for _, mutation := range []string{"swap-roles", "missing-object", "negate-referent", "change-time"} {
		t.Run(mutation, func(t *testing.T) {
			input := basicRealizationMeaning(t, "closing_object", "teacher", "window", meaning.TemporalFramePresent, meaning.MultiplicityOne, true)
			base, err := target.Realize(context.Background(), input)
			if err != nil || len(negativePhraseTexts(base)) == 0 {
				t.Fatal("base realization missing", err)
			}
			forbidden := map[string]bool{}
			for _, s := range negativePhraseTexts(base) {
				forbidden[s] = true
			}
			switch mutation {
			case "swap-roles":
				input.Meaning.Nodes[1].Participants[0].Target = "o"
				input.Meaning.Nodes[1].Participants[1].Target = "s"
			case "missing-object":
				input.Meaning.Nodes[1].Participants = input.Meaning.Nodes[1].Participants[:1]
			case "negate-referent":
				input.Meaning.Nodes[3].Target = "o"
			case "change-time":
				input.Meaning.Nodes[1].TemporalFrame = meaning.TemporalFramePast
			}
			result, err := target.Realize(context.Background(), input)
			if err == nil {
				for _, s := range negativePhraseTexts(result) {
					if forbidden[s] {
						t.Errorf("changed Meaning reused original surface: %q", s)
					}
				}
			}
			if mutation == "change-time" && (err != nil || len(negativePhraseTexts(result)) == 0) {
				t.Fatal("licensed past missing", err)
			}
		})
	}
}

func TestBasicRealizationFlyingPastMorphology(t *testing.T) {
	seed, err := pl.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ number, gender, form string }{
		{"singular", "masculine_personal", "latał"},
		{"singular", "feminine", "latała"},
		{"singular", "neuter", "latało"},
		{"plural", "masculine_personal", "latali"},
		{"plural", "non_masculine_personal", "latały"},
	} {
		t.Run(tc.form, func(t *testing.T) {
			values := []features.FeatureValue{}
			for _, entry := range [][2]string{{"aspect", "imperfective"}, {"mood", "indicative"}, {"voice", "active"}, {"verb_form", "finite"}, {"tense", "past"}, {"person", "third"}, {"number", tc.number}, {"gender", tc.gender}} {
				values = append(values, features.FeatureValue{Feature: features.FeatureID(entry[0]), State: features.StateKnown, Value: features.FeatureValueID(entry[1])})
			}
			generated, err := seed.Generate("lexeme:pl.latac.v", values)
			if err != nil || len(generated.Forms) != 1 || generated.Forms[0].Surface != tc.form {
				t.Fatalf("generation=%+v err=%v", generated, err)
			}
			readings, err := seed.Analyze(tc.form)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, r := range readings {
				if r.LexemeID != "lexeme:pl.latac.v" {
					continue
				}
				exact := true
				for _, v := range values {
					matched := false
					for _, f := range r.Features.Values {
						matched = matched || f.Feature == v.Feature && f.State == v.State && f.Value == v.Value
					}
					exact = exact && matched
				}
				found = found || exact
			}
			if !found {
				t.Errorf("inverse morphology did not recover requested features: %+v", readings)
			}
		})
	}
}
