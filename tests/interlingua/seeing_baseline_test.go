package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestSeeingMeaning(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	teacherDog, ok, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("INFRA: load teacher-dog fixture: ok=%t error=%v", ok, err)
	}
	dogTeacher, ok, err := interlingua.Seeing("seeing.dog-teacher")
	if err != nil || !ok {
		t.Fatalf("INFRA: load dog-teacher fixture: ok=%t error=%v", ok, err)
	}
	for id, graph := range map[string]meaning.Meaning{"seeing.teacher-dog": teacherDog, "seeing.dog-teacher": dogTeacher} {
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %q is invalid: %v", id, err)
		}
	}

	renamed := renameMeaning(teacherDog, map[meaning.NodeID]meaning.NodeID{"e": "event", "t": "person", "d": "animal"})
	if matches, err := matrix.GraphMatches(teacherDog, renamed); err != nil || !matches {
		t.Errorf("renamed/reordered graph should match: matches=%t error=%v", matches, err)
	}
	if matches, err := matrix.GraphMatches(teacherDog, dogTeacher); err != nil || matches {
		t.Errorf("reversed semantic roles must differ: matches=%t error=%v", matches, err)
	}

	teacher, _ := ids.NewConceptID("teacher")
	seeing, _ := ids.NewConceptID("seeing")
	distinct, err := buildGraph(seeing, []meaning.Participant{
		{Role: "experiencer", Target: "t1"}, {Role: "stimulus", Target: "t2"},
	}, []meaning.NodeID{"t1", "t2"}, teacher)
	if err != nil {
		t.Fatalf("build distinct teacher referents: %v", err)
	}
	collapsed, err := buildGraph(seeing, []meaning.Participant{
		{Role: "experiencer", Target: "t1"}, {Role: "stimulus", Target: "t1"},
	}, []meaning.NodeID{"t1"}, teacher)
	if err != nil {
		t.Fatalf("build shared teacher referent: %v", err)
	}
	if matches, err := matrix.GraphMatches(distinct, collapsed); err != nil || matches {
		t.Errorf("graphs with distinct same-concept referents must not collapse: matches=%t error=%v", matches, err)
	}
}

func renameMeaning(graph meaning.Meaning, names map[meaning.NodeID]meaning.NodeID) meaning.Meaning {
	roots := make([]meaning.NodeID, len(graph.Roots))
	for i, id := range graph.Roots {
		roots[i] = names[id]
	}
	nodes := make([]meaning.Node, len(graph.Nodes))
	for i, node := range graph.Nodes {
		node.ID = names[node.ID]
		node.Participants = append([]meaning.Participant(nil), node.Participants...)
		node.Content = append([]meaning.NodeID(nil), node.Content...)
		for j := range node.Participants {
			node.Participants[j].Target = names[node.Participants[j].Target]
		}
		if node.Target != "" {
			node.Target = names[node.Target]
		}
		for j := range node.Content {
			node.Content[j] = names[node.Content[j]]
		}
		nodes[len(nodes)-1-i] = node
	}
	result, err := meaning.Build(roots, nodes, meaning.DefaultBudget())
	if err != nil {
		panic(err)
	}
	return result
}

func buildGraph(seeing ids.ConceptID, participants []meaning.Participant, referentIDs []meaning.NodeID, teacher ids.ConceptID) (meaning.Meaning, error) {
	event, err := meaning.NewEvent("e", seeing, participants)
	if err != nil {
		return meaning.Meaning{}, err
	}
	nodes := []meaning.Node{event}
	for _, id := range referentIDs {
		referent, err := meaning.NewReferent(id, teacher)
		if err != nil {
			return meaning.Meaning{}, err
		}
		nodes = append(nodes, referent)
	}
	return meaning.Build([]meaning.NodeID{"e"}, nodes, meaning.DefaultBudget())
}
