package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionMeaning(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	fixtureIDs := []string{
		"composition.black-cat-sees-dog",
		"composition.black-dog-sees-cat",
		"composition.black-bird-sees-dog",
		"composition.black-cat-sees-cat",
	}
	expectedConcepts := map[string][2]meaning.ConceptRef{
		"composition.black-cat-sees-dog":  {"concept:cat", "concept:dog"},
		"composition.black-dog-sees-cat":  {"concept:dog", "concept:cat"},
		"composition.black-bird-sees-dog": {"concept:bird", "concept:dog"},
		"composition.black-cat-sees-cat":  {"concept:cat", "concept:cat"},
	}
	graphs := make(map[string]meaning.Meaning, len(fixtureIDs))
	for _, id := range fixtureIDs {
		graph, ok, err := interlingua.Composition(id)
		if err != nil || !ok {
			t.Fatalf("INFRA: load Composition fixture %q: ok=%t error=%v", id, ok, err)
		}
		assertCompositionValid(t, core, id, graph)
		assertAuthoredCompositionShape(t, id, graph)
		subject, _ := compositionNode(graph, "r1")
		object, _ := compositionNode(graph, "r2")
		if subject.Concept != expectedConcepts[id][0] || object.Concept != expectedConcepts[id][1] {
			t.Errorf("fixture %q has unexpected subject/object concepts: %s/%s", id, subject.Concept, object.Concept)
		}
		if id == "composition.black-cat-sees-cat" && subject.ID == object.ID {
			t.Errorf("same-concept cat referents must remain separate graph nodes")
		}
		graphs[id] = graph
	}

	catDog := graphs["composition.black-cat-sees-dog"]
	renamed := renameMeaning(catDog, map[meaning.NodeID]meaning.NodeID{
		"e": "event", "q": "property", "r1": "subject", "r2": "object",
	})
	renamed.Roots[0], renamed.Roots[1] = renamed.Roots[1], renamed.Roots[0]
	if matches, err := matrix.GraphMatches(catDog, renamed); err != nil || !matches {
		t.Fatalf("alpha-renamed graph with independently reordered roots should match: matches=%t error=%v", matches, err)
	}

	withoutProperty := cloneMeaning(catDog)
	withoutProperty.Roots = []meaning.NodeID{"e"}
	withoutProperty.Nodes = filterCompositionNodes(withoutProperty.Nodes, "q")
	withoutProperty = mustBuildMeaning(t, withoutProperty)
	assertCompositionValid(t, core, "without property", withoutProperty)
	assertCompositionMismatch(t, "property state omitted", catDog, withoutProperty)

	wrongBearer := cloneMeaning(catDog)
	for index := range wrongBearer.Nodes {
		if wrongBearer.Nodes[index].ID == "q" {
			wrongBearer.Nodes[index].Participants[0].Target = "r2"
		}
	}
	wrongBearer = mustBuildMeaning(t, wrongBearer)
	assertCompositionValid(t, core, "wrong bearer", wrongBearer)
	assertCompositionMismatch(t, "black bearer points to object", catDog, wrongBearer)

	swappedRoles := cloneMeaning(catDog)
	for index := range swappedRoles.Nodes {
		if swappedRoles.Nodes[index].ID == "e" {
			swappedRoles.Nodes[index].Participants[0].Target = "r2"
			swappedRoles.Nodes[index].Participants[1].Target = "r1"
		}
	}
	swappedRoles = mustBuildMeaning(t, swappedRoles)
	assertCompositionValid(t, core, "swapped roles", swappedRoles)
	assertCompositionMismatch(t, "seeing roles swapped", catDog, swappedRoles)

	catAsDog := cloneMeaning(catDog)
	for index := range catAsDog.Nodes {
		if catAsDog.Nodes[index].ID == "r1" {
			catAsDog.Nodes[index].Concept = "concept:dog"
		}
	}
	catAsDog = mustBuildMeaning(t, catAsDog)
	assertCompositionValid(t, core, "cat changed to dog", catAsDog)
	assertCompositionMismatch(t, "subject concept changed", catDog, catAsDog)

	sameConcept := graphs["composition.black-cat-sees-cat"]
	wrongSameConceptBearer := cloneMeaning(sameConcept)
	for index := range wrongSameConceptBearer.Nodes {
		if wrongSameConceptBearer.Nodes[index].ID == "q" {
			wrongSameConceptBearer.Nodes[index].Participants[0].Target = "r2"
		}
	}
	wrongSameConceptBearer = mustBuildMeaning(t, wrongSameConceptBearer)
	assertCompositionValid(t, core, "same-concept wrong bearer", wrongSameConceptBearer)
	assertCompositionMismatch(t, "same-concept bearer changed to other referent", sameConcept, wrongSameConceptBearer)

	collapsed := cloneMeaning(sameConcept)
	for index := range collapsed.Nodes {
		switch collapsed.Nodes[index].ID {
		case "e":
			collapsed.Nodes[index].Participants[1].Target = "r1"
		}
	}
	collapsed.Nodes = filterCompositionNodes(collapsed.Nodes, "r2")
	collapsed = mustBuildMeaning(t, collapsed)
	assertCompositionValid(t, core, "collapsed cat referent", collapsed)
	assertCompositionMismatch(t, "same-concept referents collapsed", sameConcept, collapsed)
}

func assertCompositionValid(t *testing.T, graphCore *core.ConceptGraph, name string, graph meaning.Meaning) {
	t.Helper()
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		t.Fatalf("%s has invalid structure: %v", name, err)
	}
	if err := meaning.ValidateMeaning(graph, graphCore, meaning.DefaultBudget()); err != nil {
		t.Fatalf("%s does not validate against CORE: %v", name, err)
	}
}

func assertAuthoredCompositionShape(t *testing.T, name string, graph meaning.Meaning) {
	t.Helper()
	if len(graph.Roots) != 2 || graph.Roots[0] != "e" || graph.Roots[1] != "q" || len(graph.Nodes) != 4 {
		t.Errorf("%s must have event/property roots and four nodes: roots=%v nodes=%d", name, graph.Roots, len(graph.Nodes))
	}
	event, eventOK := compositionNode(graph, "e")
	state, stateOK := compositionNode(graph, "q")
	subject, subjectOK := compositionNode(graph, "r1")
	object, objectOK := compositionNode(graph, "r2")
	if !eventOK || !stateOK || !subjectOK || !objectOK {
		t.Fatalf("%s is missing an authored event, property state, or distinct referent node", name)
	}
	if event.Kind != meaning.KindEvent || event.Concept != "concept:seeing" || event.TemporalFrame != meaning.TemporalFramePresent || len(event.Participants) != 2 ||
		event.Participants[0] != (meaning.Participant{Role: "experiencer", Target: "r1"}) || event.Participants[1] != (meaning.Participant{Role: "stimulus", Target: "r2"}) {
		t.Errorf("%s event differs from authored seeing contract: %+v", name, event)
	}
	if state.Kind != meaning.KindState || state.Concept != "concept:black" || state.TemporalFrame != "" || len(state.Participants) != 1 ||
		state.Participants[0] != (meaning.Participant{Role: "bearer", Target: "r1"}) {
		t.Errorf("%s property state differs from authored black-bearer contract: %+v", name, state)
	}
	if subject.Kind != meaning.KindReferent || subject.Multiplicity != meaning.MultiplicityOne || object.Kind != meaning.KindReferent || object.Multiplicity != meaning.MultiplicityOne || subject.ID == object.ID {
		t.Errorf("%s referents must remain separate and one: subject=%+v object=%+v", name, subject, object)
	}
}

func assertCompositionMismatch(t *testing.T, name string, want, got meaning.Meaning) {
	t.Helper()
	matches, err := matrix.GraphMatches(want, got)
	if err != nil {
		t.Errorf("compare %s: %v", name, err)
		return
	}
	if matches {
		t.Errorf("%s unexpectedly matches the authored composition", name)
	}
}

func compositionNode(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func filterCompositionNodes(nodes []meaning.Node, omit meaning.NodeID) []meaning.Node {
	result := make([]meaning.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.ID != omit {
			result = append(result, node)
		}
	}
	return result
}
