package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestA12ContrastGraphsValidateRolesCountsAndScope(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	tests := []struct {
		id, event, role1, concept1, role2, concept2 string
		count1, count2                              meaning.Multiplicity
		negative, blackObject                       bool
	}{
		{"m5.seeing.teachers-dog", "concept:seeing", "experiencer", "concept:teacher", "stimulus", "concept:dog", meaning.MultiplicityMany, meaning.MultiplicityOne, false, false},
		{"m5.neg.hearing.teachers-dog", "concept:hearing", "experiencer", "concept:teacher", "stimulus", "concept:dog", meaning.MultiplicityMany, meaning.MultiplicityOne, true, false},
		{"m5.black.dog-sees-black-cat", "concept:seeing", "experiencer", "concept:dog", "stimulus", "concept:cat", meaning.MultiplicityOne, meaning.MultiplicityOne, false, true},
		{"m5.neg.seeing.cat-dog", "concept:seeing", "experiencer", "concept:cat", "stimulus", "concept:dog", meaning.MultiplicityOne, meaning.MultiplicityOne, true, false},
		{"m5.neg.seeing.teacher-book", "concept:seeing", "experiencer", "concept:teacher", "stimulus", "concept:book", meaning.MultiplicityOne, meaning.MultiplicityOne, true, false},
		{"m5.hearing.dogs-cat", "concept:hearing", "experiencer", "concept:dog", "stimulus", "concept:cat", meaning.MultiplicityMany, meaning.MultiplicityOne, false, false},
		{"m5.black.teacher-sees-cats", "concept:seeing", "experiencer", "concept:teacher", "stimulus", "concept:cat", meaning.MultiplicityOne, meaning.MultiplicityMany, false, true},
	}
	for _, test := range tests {
		graph, ok, err := interlingua.Contrast(test.id)
		if err != nil || !ok {
			t.Fatalf("load %s: ok=%t error=%v", test.id, ok, err)
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s fails CORE validation: %v", test.id, err)
			continue
		}
		nodes := map[meaning.NodeID]meaning.Node{}
		for _, node := range graph.Nodes {
			nodes[node.ID] = node
		}
		event, exists := nodes["e"]
		if !exists || event.Kind != meaning.KindEvent || event.Concept != meaning.ConceptRef(test.event) || event.TemporalFrame != meaning.TemporalFramePresent {
			t.Errorf("%s has an unexpected present event: %+v", test.id, event)
			continue
		}
		roles := map[meaning.Role]meaning.NodeID{}
		for _, participant := range event.Participants {
			roles[participant.Role] = participant.Target
		}
		if nodes[roles[meaning.Role(test.role1)]].Concept != meaning.ConceptRef(test.concept1) ||
			nodes[roles[meaning.Role(test.role2)]].Concept != meaning.ConceptRef(test.concept2) {
			t.Errorf("%s event roles do not preserve the authored participants", test.id)
		}
		if nodes["r1"].Multiplicity != test.count1 || nodes["r2"].Multiplicity != test.count2 {
			t.Errorf("%s multiplicities=%s/%s", test.id, nodes["r1"].Multiplicity, nodes["r2"].Multiplicity)
		}
		if test.negative {
			if len(graph.Roots) != 1 || graph.Roots[0] != "neg" || nodes["neg"].Kind != meaning.KindScope || nodes["neg"].Target != "e" {
				t.Errorf("%s does not have one negation root over its event", test.id)
			}
		}
		if test.blackObject {
			if nodes["q"].Kind != meaning.KindState || nodes["q"].Concept != "concept:black" ||
				len(nodes["q"].Participants) != 1 || nodes["q"].Participants[0].Role != "bearer" ||
				nodes["q"].Participants[0].Target != "r2" {
				t.Errorf("%s black property is not attached to the object referent", test.id)
			}
		}
	}
	if _, ok, err := interlingua.Contrast("m5.possession.have-book"); err != nil || ok {
		t.Errorf("possession GAP must not fabricate a graph: ok=%t error=%v", ok, err)
	}
}
