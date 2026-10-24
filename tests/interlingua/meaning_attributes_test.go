package interlingua_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestMeaningAttributes(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	readingIDs := []string{
		"reading.teacher-message.present",
		"reading.teacher-message.past",
		"reading.teachers-message.present",
		"reading.teachers-message.past",
	}
	graphs := make(map[string]meaning.Meaning, len(readingIDs))
	for _, id := range readingIDs {
		graph, ok, err := interlingua.Reading(id)
		if err != nil || !ok {
			t.Fatalf("INFRA: load Reading fixture %q: ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Fatalf("Reading fixture %q does not validate against CORE: %v", id, err)
		}
		graphs[id] = graph
	}

	assertMeaningAttributesWireAndCanonical(t, graphs)
	assertMeaningAttributesAlphaGraphMatching(t, graphs)
	assertMeaningAttributesSemanticEquivalence(t, graphs)
	assertMeaningAttributesProjection(t, graphs)
	assertExistingMeaningAttributes(t)
}

func assertMeaningAttributesWireAndCanonical(t *testing.T, graphs map[string]meaning.Meaning) {
	t.Helper()
	onePresent := graphs["reading.teacher-message.present"]
	past := graphs["reading.teacher-message.past"]
	manyPresent := graphs["reading.teachers-message.present"]
	withoutFrame := cloneMeaning(onePresent)
	withoutMultiplicity := cloneMeaning(onePresent)
	for index := range withoutFrame.Nodes {
		if withoutFrame.Nodes[index].Kind == meaning.KindEvent {
			withoutFrame.Nodes[index].TemporalFrame = ""
		}
		if withoutMultiplicity.Nodes[index].Kind == meaning.KindReferent && withoutMultiplicity.Nodes[index].ID == "t" {
			withoutMultiplicity.Nodes[index].Multiplicity = ""
		}
	}
	for name, pair := range map[string][2]meaning.Meaning{
		"present versus past":            {onePresent, past},
		"present versus absent frame":    {onePresent, withoutFrame},
		"one versus many":                {onePresent, manyPresent},
		"one versus absent multiplicity": {onePresent, withoutMultiplicity},
	} {
		t.Run(name, func(t *testing.T) {
			equal, err := meaning.CanonicalEqual(pair[0], pair[1], meaning.DefaultBudget())
			if err != nil || equal {
				t.Fatalf("CanonicalEqual = %t, error=%v; want unequal", equal, err)
			}
			left, err := meaning.CanonicalHash(pair[0], meaning.DefaultBudget())
			if err != nil {
				t.Fatalf("hash left Meaning: %v", err)
			}
			right, err := meaning.CanonicalHash(pair[1], meaning.DefaultBudget())
			if err != nil {
				t.Fatalf("hash right Meaning: %v", err)
			}
			if left == right {
				t.Errorf("canonical hashes unexpectedly match: %s", left)
			}
		})
	}

	encoded, err := meaning.EncodeMeaning(onePresent, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("encode Meaning with attributes: %v", err)
	}
	decoded, err := meaning.DecodeMeaning(encoded, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("decode Meaning with attributes: %v", err)
	}
	if equal, err := meaning.CanonicalEqual(onePresent, decoded, meaning.DefaultBudget()); err != nil || !equal {
		t.Errorf("Meaning JSON round trip changed attributes: equal=%t error=%v", equal, err)
	}

	jsonBytes, err := json.Marshal(onePresent)
	if err != nil {
		t.Fatalf("marshal Meaning for malformed-wire checks: %v", err)
	}
	malformed := []struct {
		name string
		from string
		to   string
	}{
		{"temporal null", `"temporal_frame":"present"`, `"temporal_frame":null`},
		{"temporal empty", `"temporal_frame":"present"`, `"temporal_frame":""`},
		{"temporal unknown", `"temporal_frame":"present"`, `"temporal_frame":"future"`},
		{"multiplicity null", `"multiplicity":"one"`, `"multiplicity":null`},
		{"multiplicity empty", `"multiplicity":"one"`, `"multiplicity":""`},
		{"multiplicity unknown", `"multiplicity":"one"`, `"multiplicity":"few"`},
	}
	for _, test := range malformed {
		t.Run("reject "+test.name, func(t *testing.T) {
			wire := strings.Replace(string(jsonBytes), test.from, test.to, 1)
			if wire == string(jsonBytes) {
				t.Fatalf("wire did not contain %s field to replace", test.name)
			}
			if _, err := meaning.DecodeMeaning([]byte(wire), meaning.DefaultBudget()); err == nil {
				t.Errorf("DecodeMeaning accepted %s", test.name)
			}
		})
	}

	wrongKindTemporal := cloneMeaning(onePresent)
	wrongKindMultiplicity := cloneMeaning(onePresent)
	for index := range wrongKindTemporal.Nodes {
		if wrongKindTemporal.Nodes[index].Kind == meaning.KindReferent {
			wrongKindTemporal.Nodes[index].TemporalFrame = meaning.TemporalFramePresent
			break
		}
	}
	for index := range wrongKindMultiplicity.Nodes {
		if wrongKindMultiplicity.Nodes[index].Kind == meaning.KindEvent {
			wrongKindMultiplicity.Nodes[index].Multiplicity = meaning.MultiplicityMany
			break
		}
	}
	for name, graph := range map[string]meaning.Meaning{
		"temporal frame on referent": wrongKindTemporal,
		"multiplicity on predicate":  wrongKindMultiplicity,
	} {
		t.Run("reject "+name, func(t *testing.T) {
			if _, err := meaning.EncodeMeaning(graph, meaning.DefaultBudget()); err == nil {
				t.Errorf("EncodeMeaning accepted %s", name)
			}
		})
	}
}

func assertMeaningAttributesAlphaGraphMatching(t *testing.T, graphs map[string]meaning.Meaning) {
	t.Helper()
	want := graphs["reading.teachers-message.present"]
	renamed := renameMeaning(want, map[meaning.NodeID]meaning.NodeID{"e": "event", "t": "readers", "m": "thing"})
	if matches, err := matrix.GraphMatches(want, renamed); err != nil || !matches {
		t.Errorf("alpha-renamed/reordered Reading graph should match: matches=%t error=%v", matches, err)
	}

	changedFrame := cloneMeaning(want)
	missingFrame := cloneMeaning(want)
	for index := range changedFrame.Nodes {
		if changedFrame.Nodes[index].Kind == meaning.KindEvent {
			changedFrame.Nodes[index].TemporalFrame = meaning.TemporalFramePast
			missingFrame.Nodes[index].TemporalFrame = ""
		}
	}
	changedMultiplicity := cloneMeaning(want)
	for index := range changedMultiplicity.Nodes {
		if changedMultiplicity.Nodes[index].ID == "t" {
			changedMultiplicity.Nodes[index].Multiplicity = meaning.MultiplicityOne
		}
	}
	missingMultiplicity := cloneMeaning(want)
	for index := range missingMultiplicity.Nodes {
		if missingMultiplicity.Nodes[index].ID == "t" {
			missingMultiplicity.Nodes[index].Multiplicity = ""
		}
	}
	swappedRoles := cloneMeaning(want)
	for index := range swappedRoles.Nodes {
		if swappedRoles.Nodes[index].Kind != meaning.KindEvent {
			continue
		}
		for participant := range swappedRoles.Nodes[index].Participants {
			switch swappedRoles.Nodes[index].Participants[participant].Role {
			case "reader":
				swappedRoles.Nodes[index].Participants[participant].Target = "m"
			case "material":
				swappedRoles.Nodes[index].Participants[participant].Target = "t"
			}
		}
	}
	swappedRoles = mustBuildMeaning(t, swappedRoles)
	for name, got := range map[string]meaning.Meaning{
		"temporal frame changed":    changedFrame,
		"temporal frame omitted":    missingFrame,
		"multiplicity changed":      changedMultiplicity,
		"multiplicity omitted":      missingMultiplicity,
		"participant roles swapped": swappedRoles,
	} {
		t.Run("GraphMatches rejects "+name, func(t *testing.T) {
			if matches, err := matrix.GraphMatches(want, got); err != nil || matches {
				t.Errorf("changed Meaning should differ: matches=%t error=%v", matches, err)
			}
		})
	}
}

func assertMeaningAttributesSemanticEquivalence(t *testing.T, graphs map[string]meaning.Meaning) {
	t.Helper()
	base := graphs["reading.teacher-message.present"]
	past, many := cloneMeaning(base), cloneMeaning(base)
	for i := range base.Nodes {
		if base.Nodes[i].Kind == meaning.KindEvent {
			past.Nodes[i].TemporalFrame = meaning.TemporalFramePast
		}
		if base.Nodes[i].ID == "t" {
			many.Nodes[i].Multiplicity = meaning.MultiplicityMany
		}
	}
	context := meaning.RequestContext{
		ID: "attribute-equivalence",
		Referents: []meaning.Referent{
			{ID: "teacher-1", Concept: "concept:teacher"},
			{ID: "message-1", Concept: "concept:message"},
		},
		Mentions: []meaning.Mention{
			{ID: "teacher-mention", NodeID: "t", ReferentID: "teacher-1", EvidenceID: "teacher-evidence"},
			{ID: "message-mention", NodeID: "m", ReferentID: "message-1", EvidenceID: "message-evidence"},
		},
	}
	for name, pair := range map[string][2]meaning.Meaning{
		"temporal frames": {
			base, past,
		},
		"referent multiplicity": {
			base, many,
		},
	} {
		t.Run(name, func(t *testing.T) {
			if same, err := meaning.SemanticEquivalence(base, cloneMeaning(base), context, context, meaning.DefaultBudget()); err != nil || !same {
				t.Fatalf("resolved positive baseline: same=%t error=%v", same, err)
			}
			equivalent, err := meaning.SemanticEquivalence(pair[0], pair[1], context, context, meaning.DefaultBudget())
			if err != nil || equivalent {
				t.Errorf("resolved graphs differing in %s should not be equivalent: equivalent=%t error=%v", name, equivalent, err)
			}
		})
	}
}

func assertMeaningAttributesProjection(t *testing.T, graphs map[string]meaning.Meaning) {
	t.Helper()
	graph := cloneMeaning(graphs["reading.teachers-message.past"])
	discourse, err := meaning.NewDiscourse("statement", meaning.ActStatement, []meaning.NodeID{"e"})
	if err != nil {
		t.Fatalf("build statement discourse: %v", err)
	}
	graph.Nodes = append(graph.Nodes, discourse)
	graph.Roots = []meaning.NodeID{"statement"}
	graph = mustBuildMeaning(t, graph)
	projection, err := meaning.ProjectAsserted(graph, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("project attributed statement: %v", err)
	}
	if len(projection.Asserted) != 1 {
		t.Fatalf("assertions = %d, want 1", len(projection.Asserted))
	}
	assertion := projection.Asserted[0]
	if assertion.TemporalFrame != meaning.TemporalFramePast || len(assertion.Referents) != 2 {
		t.Fatalf("projection lost temporal frame or direct referents: %+v", assertion)
	}
	multiplicity := make(map[meaning.NodeID]meaning.Multiplicity)
	for _, referent := range assertion.Referents {
		multiplicity[referent.ID] = referent.Multiplicity
	}
	if multiplicity["t"] != meaning.MultiplicityMany || multiplicity["m"] != meaning.MultiplicityOne {
		t.Errorf("projected referent multiplicity = %v, want teacher many/message one", multiplicity)
	}

	negative, _, ok, err := interlingua.Negation("negation.seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("load negative Seeing: ok=%t error=%v", ok, err)
	}
	negativeStatement, err := meaning.NewDiscourse("utterance", meaning.ActStatement, negative.Roots)
	if err != nil {
		t.Fatal(err)
	}
	negative.Nodes = append(negative.Nodes, negativeStatement)
	negative.Roots = []meaning.NodeID{"utterance"}
	negative = mustBuildMeaning(t, negative)
	negativeProjection, err := meaning.ProjectAsserted(negative, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("project negative Seeing: %v", err)
	}
	if len(negativeProjection.Asserted) != 0 || len(negativeProjection.Unresolved) != 1 || negativeProjection.Unresolved[0].Reason != meaning.UnresolvedNegated {
		t.Errorf("negative predicate must not project as a positive assertion: %+v", negativeProjection)
	}
}

func assertExistingMeaningAttributes(t *testing.T) {
	t.Helper()
	seeing, ok, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("load Seeing baseline: ok=%t error=%v", ok, err)
	}
	assertPredicateAndReferentsHaveAttributes(t, seeing, meaning.TemporalFramePresent, meaning.MultiplicityOne)

	for _, id := range []string{
		"identification.classification.this", "identification.classification.to",
		"identification.black.this", "identification.black.ten",
	} {
		fixture, ok, err := interlingua.Identification(id)
		if err != nil || !ok {
			t.Fatalf("load Identification fixture %q: ok=%t error=%v", id, ok, err)
		}
		assertPredicateAndReferentsHaveAttributes(t, fixture.Graph, meaning.TemporalFramePresent, meaning.MultiplicityOne)
	}
	for _, id := range []string{"negation.classification.this", "negation.classification.to", "negation.seeing.teacher-dog"} {
		graph, _, ok, err := interlingua.Negation(id)
		if err != nil || !ok {
			t.Fatalf("load Negation fixture %q: ok=%t error=%v", id, ok, err)
		}
		for _, node := range graph.Nodes {
			if node.Kind == meaning.KindEvent || node.Kind == meaning.KindState {
				if node.TemporalFrame != meaning.TemporalFramePresent {
					t.Errorf("negation %q failed to inherit present frame: %+v", id, node)
				}
			}
			if node.Kind == meaning.KindReferent && node.Multiplicity != meaning.MultiplicityOne {
				t.Errorf("negation %q failed to inherit one multiplicity: %+v", id, node)
			}
		}
	}
}

func assertPredicateAndReferentsHaveAttributes(t *testing.T, graph meaning.Meaning, frame meaning.TemporalFrame, multiplicity meaning.Multiplicity) {
	t.Helper()
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindEvent, meaning.KindState:
			if node.TemporalFrame != frame {
				t.Errorf("predicate %q frame = %q, want %q", node.ID, node.TemporalFrame, frame)
			}
		case meaning.KindReferent:
			if node.Multiplicity != multiplicity {
				t.Errorf("referent %q multiplicity = %q, want %q", node.ID, node.Multiplicity, multiplicity)
			}
		}
	}
}

func mustBuildMeaning(t *testing.T, graph meaning.Meaning) meaning.Meaning {
	t.Helper()
	result, err := meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build Meaning graph: %v", err)
	}
	return result
}
