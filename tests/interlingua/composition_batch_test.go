package interlingua_test

import (
	"fmt"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestCompositionBatchFixturesValidateAndPreserveReferents(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	for i := 1; i <= 12; i++ {
		id := fmt.Sprintf("composition.batch.%02d", i)
		graph, ok, err := interlingua.CompositionBatch(id)
		if err != nil || !ok {
			t.Fatalf("fixture %s: ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %s invalid against CORE: %v", id, err)
		}
	}
	separate, _, err := interlingua.CompositionBatch("composition.batch.03")
	if err != nil {
		t.Fatal(err)
	}
	if len(separate.Nodes) != 6 {
		t.Errorf("four-participant pair must have six graph nodes, got %d", len(separate.Nodes))
	}
	for _, tc := range []struct {
		id                  string
		event, role, target string
	}{
		{"composition.batch.01", "e1", "experiencer", "r1"}, {"composition.batch.01", "e2", "experiencer", "r1"},
		{"composition.batch.02", "e1", "experiencer", "r1"}, {"composition.batch.02", "e2", "experiencer", "r1"},
		{"composition.batch.03", "e1", "experiencer", "r1"}, {"composition.batch.03", "e2", "experiencer", "r3"},
		{"composition.batch.12", "e1", "experiencer", "r1"}, {"composition.batch.12", "e2", "agent", "r1"},
	} {
		graph, _, err := interlingua.CompositionBatch(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		event := findNode(graph.Nodes, meaning.NodeID(tc.event))
		found := false
		for _, participant := range event.Participants {
			if string(participant.Role) == tc.role && string(participant.Target) == tc.target {
				found = true
			}
		}
		if !found {
			t.Errorf("%s %s missing %s→%s participant: %+v", tc.id, tc.event, tc.role, tc.target, event.Participants)
		}
	}
	catPair, _, err := interlingua.CompositionBatch("composition.batch.08")
	if err != nil {
		t.Fatal(err)
	}
	catRefs := 0
	catIDs := map[meaning.NodeID]bool{}
	for _, node := range catPair.Nodes {
		if node.Kind == meaning.KindReferent && node.Concept == "concept:cat" {
			catRefs++
			catIDs[node.ID] = true
		}
	}
	if catRefs != 2 || len(catIDs) != 2 {
		t.Errorf("C08 must retain two distinct cat referents: nodes=%v", catPair.Nodes)
	}
	for _, tc := range []struct {
		id      string
		bearers []string
		roots   int
	}{
		{"composition.batch.04", []string{"r1", "r2"}, 3}, {"composition.batch.05", []string{"r2"}, 2},
		{"composition.batch.10", []string{"r2"}, 2}, {"composition.batch.11", []string{"r1"}, 2},
	} {
		graph, _, err := interlingua.CompositionBatch(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if len(graph.Roots) != tc.roots {
			t.Errorf("%s roots=%v; want %d", tc.id, graph.Roots, tc.roots)
		}
		foundBearers := map[string]bool{}
		for _, node := range graph.Nodes {
			if node.Kind == meaning.KindState && node.Concept == "concept:black" {
				if len(node.Participants) == 1 && node.Participants[0].Role == "bearer" {
					foundBearers[string(node.Participants[0].Target)] = true
				} else {
					t.Errorf("%s black state has invalid bearer roles: %+v", tc.id, node.Participants)
				}
			}
		}
		if len(foundBearers) != len(tc.bearers) {
			t.Errorf("%s black bearers=%v; want %v", tc.id, foundBearers, tc.bearers)
		}
		for _, bearer := range tc.bearers {
			if !foundBearers[bearer] {
				t.Errorf("%s missing black bearer %s", tc.id, bearer)
			}
		}
	}
	for _, id := range []string{"composition.batch.10", "composition.batch.11"} {
		graph, _, err := interlingua.CompositionBatch(id)
		if err != nil {
			t.Fatal(err)
		}
		for _, root := range graph.Roots {
			node := findNode(graph.Nodes, root)
			if node.Kind == meaning.KindScope && (node.Operator != meaning.OperatorNegation || node.Target != "e1") {
				t.Errorf("%s negation must scope only over seeing event: %+v", id, node)
			}
		}
	}
}

func findNode(nodes []meaning.Node, id meaning.NodeID) meaning.Node {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}
