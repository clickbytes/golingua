package interlingua

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
)

func TestActionExpansionFixturesPreserveAuthoredRolesAndMultiplicity(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	for _, spec := range actionExpansionSpecs {
		graph, ok, err := ActionExpansion(spec.id)
		if err != nil || !ok {
			t.Fatalf("fixture %q: ok=%t error=%v", spec.id, ok, err)
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %q invalid against CORE: %v", spec.id, err)
			continue
		}
		if len(graph.Roots) != 1 || graph.Roots[0] != "e" || len(graph.Nodes) != len(spec.participants)+1 {
			t.Errorf("fixture %q has unexpected roots/nodes: %+v", spec.id, graph)
			continue
		}
		event := nodeByID(graph.Nodes, "e")
		if event.Kind != meaning.KindEvent || string(event.Concept) != "concept:"+spec.event || len(event.Participants) != len(spec.participants) || event.TemporalFrame != meaning.TemporalFramePresent {
			t.Errorf("fixture %q has unexpected event structure: %+v", spec.id, event)
			continue
		}
		for i, expected := range spec.participants {
			participant := event.Participants[i]
			target := nodeByID(graph.Nodes, participant.Target)
			if string(participant.Role) != expected.role || string(target.Concept) != "concept:"+expected.concept {
				t.Errorf("fixture %q participant %d mismatch: %+v target=%+v", spec.id, i, participant, target)
			}
			want := meaning.Multiplicity("")
			if expected.multiplicity == "one" {
				want = meaning.MultiplicityOne
			}
			if target.Multiplicity != want {
				t.Errorf("fixture %q participant %d multiplicity=%q want %q", spec.id, i, target.Multiplicity, want)
			}
		}
	}
}

func nodeByID(nodes []meaning.Node, id meaning.NodeID) meaning.Node {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}
