package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCorpusEventFixturesValidateAndPreserveAuthoredStructure(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	tests := []struct {
		id, event, role1, concept1, role2, concept2 string
		past                                        bool
	}{
		{"S003", "concept:seeing", "experiencer", "concept:teacher", "stimulus", "concept:book", false},
		{"S010", "concept:seeing", "experiencer", "concept:doctor", "stimulus", "concept:car", false},
		{"S016", "concept:seeing", "experiencer", "concept:father", "stimulus", "concept:dog", false},
		{"S023", "concept:seeing", "experiencer", "concept:child", "stimulus", "concept:book", false},
		{"S023.offspring", "concept:seeing", "experiencer", "concept:child_of_person", "stimulus", "concept:book", false},
		{"S028", "concept:hearing", "experiencer", "concept:teacher", "stimulus", "concept:bird", false},
		{"S034", "concept:hearing", "experiencer", "concept:doctor", "stimulus", "concept:mother", false},
		{"S040", "concept:hearing", "experiencer", "concept:child", "stimulus", "concept:father", false},
		{"S040.offspring", "concept:hearing", "experiencer", "concept:child_of_person", "stimulus", "concept:father", false},
		{"S043", "concept:hearing", "experiencer", "concept:neighbor", "stimulus", "concept:bird", false},
		{"S051", "concept:looking", "agent", "concept:teacher", "target", "concept:book", false},
		{"S055", "concept:looking", "agent", "concept:teacher", "target", "concept:tree", false},
		{"S063", "concept:looking", "agent", "concept:mother", "target", "concept:dog", false},
		{"S074", "concept:looking", "agent", "concept:child", "target", "concept:car", false},
		{"S074.offspring", "concept:looking", "agent", "concept:child_of_person", "target", "concept:car", false},
		{"S076", "concept:receiving", "recipient", "concept:teacher", "theme", "concept:book", true},
		{"S081", "concept:receiving", "recipient", "concept:doctor", "theme", "concept:book", true},
		{"S086", "concept:receiving", "recipient", "concept:mother", "theme", "concept:book", true},
		{"S099", "concept:receiving", "recipient", "concept:child", "theme", "concept:message", true},
		{"S099.offspring", "concept:receiving", "recipient", "concept:child_of_person", "theme", "concept:message", true},
	}
	for _, test := range tests {
		graph, ok, err := interlingua.CorpusEvent(test.id)
		if err != nil || !ok {
			t.Fatalf("load authored corpus fixture %s: ok=%t error=%v", test.id, ok, err)
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %s is invalid against CORE: %v", test.id, err)
			continue
		}
		if len(graph.Roots) != 1 || graph.Roots[0] != "e" {
			t.Errorf("fixture %s has unexpected root: %v", test.id, graph.Roots)
			continue
		}
		var event *meaning.Node
		nodes := map[meaning.NodeID]meaning.Node{}
		for i := range graph.Nodes {
			node := graph.Nodes[i]
			nodes[node.ID] = node
			if node.ID == "e" {
				event = &graph.Nodes[i]
			}
		}
		if event == nil || event.Kind != meaning.KindEvent || event.Concept != meaning.ConceptRef(test.event) || len(event.Participants) != 2 {
			t.Errorf("fixture %s has unexpected event: %+v", test.id, event)
			continue
		}
		wantFrame := meaning.TemporalFramePresent
		if test.past {
			wantFrame = meaning.TemporalFramePast
		}
		if event.TemporalFrame != wantFrame {
			t.Errorf("fixture %s temporal frame=%s, want %s", test.id, event.TemporalFrame, wantFrame)
		}
		participants := map[string]string{}
		for _, participant := range event.Participants {
			participants[string(participant.Role)] = string(nodes[participant.Target].Concept)
		}
		if participants[test.role1] != test.concept1 || participants[test.role2] != test.concept2 {
			t.Errorf("fixture %s roles/concepts=%v", test.id, participants)
		}
	}
}

func TestCorpusMeaningAlternativesRequireEachAllowedReading(t *testing.T) {
	primary, ok, err := interlingua.CorpusEvent("S023")
	if err != nil || !ok {
		t.Fatalf("load primary child fixture: ok=%t error=%v", ok, err)
	}
	offspring, ok, err := interlingua.CorpusEvent("S023.offspring")
	if err != nil || !ok {
		t.Fatalf("load offspring fixture: ok=%t error=%v", ok, err)
	}
	allowedIDs := []string{"corpus.S023", "corpus.S023.offspring"}
	fixtures := map[string]meaning.Meaning{
		allowedIDs[0]: primary,
		allowedIDs[1]: offspring,
	}
	for id, graph := range fixtures {
		matchedID, matches, err := matrix.MatchAllowedMeaning(allowedIDs, fixtures, graph)
		if err != nil || !matches || matchedID != id {
			t.Errorf("fixture %q matched %q (matches=%t, error=%v)", id, matchedID, matches, err)
		}
	}
	missing := matrix.MissingRequiredMeaningIDs(allowedIDs, map[string]bool{allowedIDs[0]: true})
	if len(missing) != 1 || missing[0] != allowedIDs[1] {
		t.Errorf("missing required readings=%v, want only %q", missing, allowedIDs[1])
	}
	if missing := matrix.MissingRequiredMeaningIDs(allowedIDs, map[string]bool{allowedIDs[0]: true, allowedIDs[1]: true}); len(missing) != 0 {
		t.Errorf("complete required readings still missing: %v", missing)
	}
	bad := primary
	bad.Nodes = append([]meaning.Node(nil), primary.Nodes...)
	for i := range bad.Nodes {
		if bad.Nodes[i].ID == "r1" {
			bad.Nodes[i].Concept = "concept:teacher"
		}
	}
	if matchedID, matches, err := matrix.MatchAllowedMeaning(allowedIDs, fixtures, bad); err != nil || matches {
		t.Errorf("unlicensed graph matched %q (matches=%t, error=%v)", matchedID, matches, err)
	}
}
