package matrix

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// The two cats fill the same role of two distinct events. A concept/role key
// cannot identify either occurrence, even when only one source record exists.
func TestSourceFeaturesDoNotGuessRepeatedOccurrences(t *testing.T) {
	graph := meaning.Meaning{SchemaVersion: 1, Roots: []meaning.NodeID{"left", "right"}, Nodes: []meaning.Node{
		{ID: "left", Kind: meaning.KindEvent, Concept: "concept:seeing", TemporalFrame: meaning.TemporalFramePresent, Participants: []meaning.Participant{{Role: "experiencer", Target: "teacher"}, {Role: "stimulus", Target: "cat1"}}},
		{ID: "right", Kind: meaning.KindEvent, Concept: "concept:seeing", TemporalFrame: meaning.TemporalFramePast, Participants: []meaning.Participant{{Role: "experiencer", Target: "teacher"}, {Role: "stimulus", Target: "cat2"}}},
		{ID: "teacher", Kind: meaning.KindReferent, Concept: "concept:teacher", Multiplicity: meaning.MultiplicityOne},
		{ID: "cat1", Kind: meaning.KindReferent, Concept: "concept:cat", Multiplicity: meaning.MultiplicityOne},
		{ID: "cat2", Kind: meaning.KindReferent, Concept: "concept:cat", Multiplicity: meaning.MultiplicityOne},
	}}
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		t.Fatal(err)
	}
	cat, _ := ids.ParseConceptID("concept:cat")
	candidate := semantic.MeaningCandidate{Meaning: &graph, Evidence: semantic.EvidenceTrace{Records: []semantic.EvidenceRecord{
		{CoreConceptID: cat, Role: "stimulus", Outcome: status.OutcomeOK, Features: []features.FeatureValue{{Feature: "determiner", State: features.StateKnown, Value: "definite"}}},
	}}}
	if values, known := runtime.SourceFeatures(candidate, "stimulus", cat); known || values != nil {
		t.Fatalf("concept/role lookup guessed a repeated occurrence: values=%v known=%t", values, known)
	}
}

func occurrenceSource() semantic.MeaningCandidate {
	graph := meaning.Meaning{SchemaVersion: 1, Roots: []meaning.NodeID{"left", "right"}, Nodes: []meaning.Node{
		{ID: "left", Kind: meaning.KindEvent, Concept: "concept:seeing", TemporalFrame: meaning.TemporalFramePresent, Participants: []meaning.Participant{{Role: "experiencer", Target: "teacher"}, {Role: "stimulus", Target: "cat1"}}},
		{ID: "right", Kind: meaning.KindEvent, Concept: "concept:seeing", TemporalFrame: meaning.TemporalFramePast, Participants: []meaning.Participant{{Role: "experiencer", Target: "teacher"}, {Role: "stimulus", Target: "cat2"}}},
		{ID: "teacher", Kind: meaning.KindReferent, Concept: "concept:teacher", Multiplicity: meaning.MultiplicityOne},
		{ID: "cat1", Kind: meaning.KindReferent, Concept: "concept:cat", Multiplicity: meaning.MultiplicityOne},
		{ID: "cat2", Kind: meaning.KindReferent, Concept: "concept:cat", Multiplicity: meaning.MultiplicityOne},
	}}
	return semantic.MeaningCandidate{ID: "authored-occurrences", SemanticID: "authored-occurrences", Meaning: &graph, Status: status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: "authored-occurrences", Outcome: status.OutcomeOK, Records: []semantic.EvidenceRecord{
			occurrenceRecord("cat1", "left", "stimulus", "cat", "determiner", "definite"),
			occurrenceRecord("cat2", "right", "stimulus", "cat", "determiner", "indefinite"),
			occurrenceRecord("teacher", "left", "experiencer", "teacher", "determiner", "indefinite"),
			occurrenceRecord("teacher", "right", "experiencer", "teacher", "determiner", "definite"),
			occurrenceRecord("left", "", "", "seeing", "tense", "present"),
			occurrenceRecord("right", "", "", "seeing", "tense", "past"),
		}}}
}

func occurrenceRecord(node, owner meaning.NodeID, role meaning.Role, concept, feature, value string) semantic.EvidenceRecord {
	core, _ := ids.ParseConceptID("concept:" + concept)
	sense, _ := ids.ParseSenseID("sense:en." + concept + ".authored")
	return semantic.EvidenceRecord{MeaningNodeID: node, PredicateNodeID: owner, Role: role, CoreConceptID: core, SenseID: sense, BindingKey: semantic.BindingKeyFor(sense, core),
		SourceSpan: text.Span{Start: 0, End: 1}, Outcome: status.OutcomeOK,
		Provenance: provenance.SourceRef{ID: "authored-occurrence-test", Path: "tests/matrix/source_features_test.go", Version: "1", License: "unknown", Author: "test"},
		Features:   []features.FeatureValue{{Feature: features.FeatureID(feature), State: features.StateKnown, Value: features.FeatureValueID(value)}}}
}

func TestSourceFeaturesAddressedParticipantsAndPredicates(t *testing.T) {
	source := occurrenceSource()
	if err := source.Evidence.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := runtime.ValidateSourceFeatureEvidence(source); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		owner                  meaning.NodeID
		role                   meaning.Role
		concept, feature, want string
	}{
		{"left", "stimulus", "cat", "determiner", "definite"}, {"right", "stimulus", "cat", "determiner", "indefinite"},
		{"left", "experiencer", "teacher", "determiner", "indefinite"}, {"right", "experiencer", "teacher", "determiner", "definite"},
		{"left", "", "seeing", "tense", "present"}, {"right", "", "seeing", "tense", "past"},
	} {
		t.Run(string(tc.owner)+"-"+string(tc.role), func(t *testing.T) {
			concept, _ := ids.ParseConceptID("concept:" + tc.concept)
			values, found := runtime.SourceFeaturesAt(source, tc.owner, tc.role, concept)
			if !found || string(values[features.FeatureID(tc.feature)]) != tc.want {
				t.Fatalf("occurrence features=%v found=%t want=%s", values, found, tc.want)
			}
			values[features.FeatureID(tc.feature)] = "mutated"
			again, _ := runtime.SourceFeaturesAt(source, tc.owner, tc.role, concept)
			if string(again[features.FeatureID(tc.feature)]) != tc.want {
				t.Fatal("lookup leaked mutable values")
			}
		})
	}
	cat, _ := ids.ParseConceptID("concept:cat")
	source.Evidence.Records = append(source.Evidence.Records[:1], source.Evidence.Records[2:]...)
	if values, found := runtime.SourceFeaturesAt(source, "right", "stimulus", cat); found {
		t.Fatalf("borrowed other occurrence: %v", values)
	}
	if runtime.SourceFeaturesPresentAt(source, "right", "stimulus", cat) {
		t.Fatal("missing occurrence was reported present")
	}
}

func TestSourceFeaturesRejectStaleAndConflictingEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		edit       func(*semantic.MeaningCandidate)
	}{
		{"wrong cat with same concept", "EVIDENCE_ARGUMENT_MISMATCH", func(c *semantic.MeaningCandidate) { c.Evidence.Records[0].MeaningNodeID = "cat2" }},
		{"wrong owning predicate", "EVIDENCE_ARGUMENT_MISMATCH", func(c *semantic.MeaningCandidate) { c.Evidence.Records[0].PredicateNodeID = "right" }},
		{"missing node", "EVIDENCE_NODE_MISMATCH", func(c *semantic.MeaningCandidate) { c.Evidence.Records[0].MeaningNodeID = "absent" }},
		{"wrong concept", "EVIDENCE_NODE_MISMATCH", func(c *semantic.MeaningCandidate) {
			c.Evidence.Records[0].CoreConceptID, _ = ids.ParseConceptID("concept:dog")
		}},
		{"missing owner", "EVIDENCE_ARGUMENT_OWNER_MISSING", func(c *semantic.MeaningCandidate) { c.Evidence.Records[0].PredicateNodeID = "" }},
		{"legacy ambiguity", "SOURCE_FEATURE_OCCURRENCE_AMBIGUOUS", func(c *semantic.MeaningCandidate) {
			c.Evidence.Records[0].MeaningNodeID = ""
			c.Evidence.Records[0].PredicateNodeID = ""
		}},
		{"same occurrence disagreement", "SOURCE_FEATURE_CONFLICT", func(c *semantic.MeaningCandidate) {
			r := c.Evidence.Records[0]
			r.Features = []features.FeatureValue{{Feature: "determiner", State: features.StateKnown, Value: "indefinite"}}
			c.Evidence.Records = append(c.Evidence.Records, r)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := occurrenceSource()
			tc.edit(&source)
			err := runtime.ValidateSourceFeatureEvidence(source)
			var typed interface{ Result() status.Result }
			if !errors.As(err, &typed) {
				t.Fatalf("expected typed %s, got %v", tc.code, err)
			}
			result := typed.Result()
			found := false
			for _, d := range result.Diagnostics {
				found = found || d.Code == tc.code
			}
			if !found {
				t.Fatalf("expected %s, got %+v", tc.code, result)
			}
			target, _ := Runtime(t).TargetRealizerFor("en")
			surface, realizeErr := target.Realize(context.Background(), source)
			if realizeErr == nil || len(negativePhraseTexts(surface)) != 0 {
				t.Fatalf("invalid source produced text: %+v err=%v", surface, realizeErr)
			}
		})
	}
}

func TestSourceFeatureAddressesOnActualAnalysis(t *testing.T) {
	rt := Runtime(t)
	for _, tc := range []struct {
		lang  features.LanguageID
		input string
	}{
		{"en", "The cat sees a cat."}, {"en", "The teachers did not see the black cats."}, {"pl", "Czytałam list."},
	} {
		t.Run(tc.input, func(t *testing.T) {
			analyzer, _ := rt.SourceAnalyzerFor(tc.lang)
			analyzed, err := analyzer.Analyze(context.Background(), tc.input)
			if err != nil {
				t.Fatal(err)
			}
			found := 0
			for _, candidate := range analyzed.Candidates {
				if candidate.Status != status.OutcomeOK || candidate.Meaning == nil {
					continue
				}
				found++
				if err := runtime.ValidateSourceFeatureEvidence(candidate); err != nil {
					t.Fatal(err)
				}
				addresses := 0
				for _, record := range candidate.Evidence.Records {
					if record.Outcome == status.OutcomeOK && record.CoreConceptID.String() != "" && len(record.Features) > 0 {
						if record.MeaningNodeID == "" {
							t.Fatalf("source morphology lacks occurrence address: %+v", record)
						}
						addresses++
					}
				}
				if addresses < 2 {
					t.Fatalf("expected predicate and argument features, got %d", addresses)
				}
				encoded, _ := json.Marshal(candidate.Evidence)
				if !strings.Contains(string(encoded), "meaning_node_id") || !strings.Contains(string(encoded), "predicate_node_id") {
					t.Fatal("JSON dropped source addresses")
				}
				target, _ := rt.TargetRealizerFor(tc.lang)
				before, _ := json.Marshal(candidate)
				result, err := target.Realize(context.Background(), candidate)
				if err != nil {
					t.Fatal(err)
				}
				texts := negativePhraseTexts(result)
				if !slices.Contains(texts, tc.input) {
					t.Fatalf("round-trip lost source features: %q %+v", texts, result.Diagnostics)
				}
				if tc.lang == "pl" && slices.Contains(texts, "Czytałem list.") {
					t.Fatal("known feminine evidence was replaced")
				}
				after, _ := json.Marshal(candidate)
				if string(before) != string(after) {
					t.Fatal("realization mutated source or evidence")
				}
			}
			if found == 0 {
				t.Fatalf("no complete source candidate: %s %+v", analyzed.Outcome, analyzed.Diagnostics)
			}
		})
	}
}

func renameOccurrenceCandidate(source semantic.MeaningCandidate, prefix string, remapEvidence bool) semantic.MeaningCandidate {
	rename := func(id meaning.NodeID) meaning.NodeID {
		if id == "" {
			return ""
		}
		return meaning.NodeID(prefix + string(id))
	}
	g := *source.Meaning
	g.Roots = append([]meaning.NodeID(nil), g.Roots...)
	g.Nodes = append([]meaning.Node(nil), g.Nodes...)
	for i := range g.Roots {
		g.Roots[i] = rename(g.Roots[i])
	}
	for i := range g.Nodes {
		n := &g.Nodes[i]
		n.ID = rename(n.ID)
		n.Target = rename(n.Target)
		n.Participants = append([]meaning.Participant(nil), n.Participants...)
		n.Content = append([]meaning.NodeID(nil), n.Content...)
		for j := range n.Participants {
			n.Participants[j].Target = rename(n.Participants[j].Target)
		}
		for j := range n.Content {
			n.Content[j] = rename(n.Content[j])
		}
	}
	slices.Reverse(g.Nodes)
	slices.Reverse(g.Roots)
	source.Meaning = &g
	source.Contexts = append([]semantic.MeaningContext(nil), source.Contexts...)
	for i := range source.Contexts {
		ctx := &source.Contexts[i]
		ctx.Meaning = g
		ctx.Context.Mentions = append([]meaning.Mention(nil), ctx.Context.Mentions...)
		for j := range ctx.Context.Mentions {
			ctx.Context.Mentions[j].NodeID = rename(ctx.Context.Mentions[j].NodeID)
		}
	}
	source.Evidence = source.Evidence.Clone()
	if remapEvidence {
		for i := range source.Evidence.Records {
			r := &source.Evidence.Records[i]
			r.MeaningNodeID = rename(r.MeaningNodeID)
			r.PredicateNodeID = rename(r.PredicateNodeID)
		}
	}
	return source
}

func TestSourceFeatureAddressesRequireExplicitRemapping(t *testing.T) {
	rt := Runtime(t)
	analyzer, _ := rt.SourceAnalyzerFor("en")
	target, _ := rt.TargetRealizerFor("en")
	result, err := analyzer.Analyze(context.Background(), "The cat sees a cat.")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, source := range result.Candidates {
		if source.Status != status.OutcomeOK || source.Meaning == nil {
			continue
		}
		found = true
		base, err := target.Realize(context.Background(), source)
		if err != nil {
			t.Fatal(err)
		}
		remapped := renameOccurrenceCandidate(source, "renamed-", true)
		got, err := target.Realize(context.Background(), remapped)
		if err != nil || !reflect.DeepEqual(negativePhraseTexts(base), negativePhraseTexts(got)) {
			t.Fatalf("explicit remap changed output: %q err=%v", negativePhraseTexts(got), err)
		}
		stale := renameOccurrenceCandidate(source, "renamed-", false)
		got, err = target.Realize(context.Background(), stale)
		if err == nil || len(negativePhraseTexts(got)) != 0 {
			t.Fatalf("stale trace was rebound by concept: %+v %v", got, err)
		}
		independent := source
		independent.Evidence = semantic.EvidenceTrace{AlternativeID: source.ID, Outcome: status.OutcomeOK}
		got, err = target.Realize(context.Background(), independent)
		if err != nil || len(negativePhraseTexts(got)) != 4 {
			t.Fatalf("independent IR requires four determiner choices: %q err=%v", negativePhraseTexts(got), err)
		}
	}
	if !found {
		t.Fatal("source analysis had no complete candidate")
	}
}

func TestSourceFeatureAddressesPreventCrossGraphEvidenceMerge(t *testing.T) {
	source := occurrenceSource()
	ctx := meaning.RequestContext{ID: "same-request"}
	for _, node := range source.Meaning.Nodes {
		if node.Kind == meaning.KindReferent {
			id := meaning.ReferentID("identity-" + string(node.ID))
			ctx.Referents = append(ctx.Referents, meaning.Referent{ID: id, Concept: node.Concept})
			ctx.Mentions = append(ctx.Mentions, meaning.Mention{ID: meaning.MentionID("mention-" + string(node.ID)), NodeID: node.ID, ReferentID: id, EvidenceID: "explicit-test-resolution"})
		}
	}
	source.Contexts = []semantic.MeaningContext{{Meaning: *source.Meaning, Context: ctx}}
	other := renameOccurrenceCandidate(source, "other-", true)
	other.ID = "other-derivation"
	other.Evidence.AlternativeID = other.ID
	equivalent, err := meaning.SemanticEquivalence(*source.Meaning, *other.Meaning, source.Contexts[0].Context, other.Contexts[0].Context, meaning.DefaultBudget())
	if err != nil || !equivalent {
		t.Fatalf("test meanings must be equivalent: %t %v", equivalent, err)
	}
	got, err := semantic.Disambiguate([]semantic.MeaningCandidate{source, other}, status.OutcomeOK, nil, Runtime(t).CoreGraph(), semantic.DefaultEvidenceBudget())
	if err != nil || len(got.Candidates) != 2 {
		t.Fatalf("traces from distinct local namespaces were merged: %d %v", len(got.Candidates), err)
	}
	for _, c := range got.Candidates {
		if err := semantic.ValidateEvidenceMeaning(c.Evidence, *c.Meaning); err != nil {
			t.Fatal(err)
		}
	}
	other = source
	other.ID = "same-namespace"
	other.Evidence = source.Evidence.Clone()
	other.Evidence.AlternativeID = other.ID
	got, err = semantic.Disambiguate([]semantic.MeaningCandidate{source, other}, status.OutcomeOK, nil, Runtime(t).CoreGraph(), semantic.DefaultEvidenceBudget())
	if err != nil || len(got.Candidates) != 1 {
		t.Fatalf("same namespace no longer merges: %d %v", len(got.Candidates), err)
	}
}

func TestSourceFeatureLegacyCompatibilityAndPresence(t *testing.T) {
	source := occurrenceSource()
	graph := *source.Meaning
	graph.Roots = []meaning.NodeID{"left"}
	graph.Nodes = slices.DeleteFunc(append([]meaning.Node(nil), graph.Nodes...), func(n meaning.Node) bool { return n.ID == "right" || n.ID == "cat2" })
	source.Meaning = &graph
	r := source.Evidence.Records[0]
	r.MeaningNodeID, r.PredicateNodeID = "", ""
	source.Evidence.Records = []semantic.EvidenceRecord{r}
	cat, _ := ids.ParseConceptID("concept:cat")
	if err := runtime.ValidateSourceFeatureEvidence(source); err != nil {
		t.Fatal(err)
	}
	old, oldOK := runtime.SourceFeatures(source, "stimulus", cat)
	addressed, newOK := runtime.SourceFeaturesAt(source, "left", "stimulus", cat)
	if !oldOK || !newOK || !reflect.DeepEqual(old, addressed) || addressed["determiner"] != "definite" {
		t.Fatalf("unique legacy evidence no longer works: %v %v", old, addressed)
	}
	r.Features = []features.FeatureValue{{Feature: "determiner", State: features.StateKnown, Value: "indefinite"}}
	source.Evidence.Records = append(source.Evidence.Records, r)
	if values, ok := runtime.SourceFeaturesAt(source, "left", "stimulus", cat); ok || values != nil {
		t.Fatal("conflicting values were silently selected")
	}
	if !runtime.SourceFeaturesPresentAt(source, "left", "stimulus", cat) {
		t.Fatal("conflicting evidence was mistaken for missing information")
	}
	source.Evidence.Records = nil
	if runtime.SourceFeaturesPresentAt(source, "left", "stimulus", cat) {
		t.Fatal("independent IR acquired source evidence")
	}
}

func TestSourceFeatureAddressBudgetAndOrigin(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		edit       func(*semantic.MeaningCandidate)
	}{
		{"record limit", "RESOURCE_LIMIT", func(c *semantic.MeaningCandidate) {
			c.Evidence.Records = make([]semantic.EvidenceRecord, semantic.DefaultEvidenceBudget().MaxRecords+1)
		}},
		{"feature limit", "RESOURCE_LIMIT", func(c *semantic.MeaningCandidate) {
			c.Evidence.Records[0].Features = make([]features.FeatureValue, features.MaxBundleValues+1)
		}},
		{"generated address", "INVALID_GENERATED_EVIDENCE", func(c *semantic.MeaningCandidate) { c.Evidence.Records[0].Origin = semantic.EvidenceOriginGenerated }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := occurrenceSource()
			tc.edit(&c)
			err := runtime.ValidateSourceFeatureEvidence(c)
			var typed interface{ Result() status.Result }
			if !errors.As(err, &typed) {
				t.Fatalf("expected typed %s, got %v", tc.code, err)
			}
			for _, diagnostic := range typed.Result().Diagnostics {
				if diagnostic.Code == tc.code {
					return
				}
			}
			t.Fatalf("expected %s, got %+v", tc.code, typed.Result())
		})
	}
}
