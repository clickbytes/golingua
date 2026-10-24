package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestNegation(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	fixtures := make(map[string]meaning.Meaning)
	var proximal, unresolved *interlingua.IdentificationMeaning
	for _, id := range []string{
		"negation.classification.this",
		"negation.classification.to",
		"negation.seeing.teacher-dog",
	} {
		graph, contextFixture, ok, err := interlingua.Negation(id)
		if err != nil || !ok {
			t.Fatalf("INFRA: load negation fixture %q: ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %q has invalid structure: %v", id, err)
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %q does not validate against CORE: %v", id, err)
		}
		if contextFixture != nil {
			if contextFixture.Graph.SchemaVersion != graph.SchemaVersion || !mustGraphMatch(t, graph, contextFixture.Graph) {
				t.Errorf("fixture %q context wrapper must carry the negated graph", id)
			}
			if err := meaning.ValidateReferences(contextFixture.Graph, contextFixture.Context, meaning.DefaultBudget()); err != nil {
				t.Errorf("fixture %q reference context is invalid: %v", id, err)
			}
			if len(contextFixture.Context.Referents) != 0 || len(contextFixture.Context.Mentions) != 1 || contextFixture.Context.Mentions[0].ReferentID != "" || contextFixture.Context.Mentions[0].EvidenceID != "" {
				t.Errorf("fixture %q must preserve the unresolved reference context", id)
			}
			if id == "negation.classification.this" {
				proximal = contextFixture
			}
			if id == "negation.classification.to" {
				unresolved = contextFixture
			}
		}
		root := findFixtureNode(graph, graph.Roots[0])
		if root.Kind != meaning.KindScope || root.Operator != meaning.OperatorNegation || len(graph.Roots) != 1 {
			t.Errorf("fixture %q must have one negative scope root", id)
		}
		predicateID, predicateKind, predicateConcept := meaning.NodeID("s"), meaning.KindState, meaning.ConceptRef("concept:class_membership")
		if id == "negation.seeing.teacher-dog" {
			predicateID, predicateKind, predicateConcept = "e", meaning.KindEvent, "concept:seeing"
		}
		predicate := findFixtureNode(graph, root.Target)
		if root.Target != predicateID || predicate.Kind != predicateKind || predicate.Concept != predicateConcept {
			t.Errorf("fixture %q negation must target its complete predicate: root=%+v target=%+v", id, root, predicate)
		}
		fixtures[id] = graph
	}

	proximalGraph := fixtures["negation.classification.this"]
	if proximal == nil || unresolved == nil {
		t.Fatal("classification fixtures must both include request context")
	}
	if proximal.Context.Mentions[0].Anchor.Value != "proximal" || unresolved.Context.Mentions[0].Anchor.Value != "unresolved" {
		t.Errorf("classification negation must preserve distinct anchor modes: proximal=%+v unresolved=%+v", proximal.Context.Mentions, unresolved.Context.Mentions)
	}

	renamed := renameMeaning(proximalGraph, map[meaning.NodeID]meaning.NodeID{"neg": "scope", "s": "state", "r": "object"})
	if matches, err := matrix.GraphMatches(proximalGraph, renamed); err != nil || !matches {
		t.Errorf("renamed/reordered negative graph should match: matches=%t error=%v", matches, err)
	}
	positive := removeNegationRoot(proximalGraph)
	if matches, err := matrix.GraphMatches(proximalGraph, positive); err != nil || matches {
		t.Errorf("removing the negation scope must change meaning: matches=%t error=%v", matches, err)
	}
	wrongTarget := cloneNegationGraph(proximalGraph)
	for i := range wrongTarget.Nodes {
		if wrongTarget.Nodes[i].Kind == meaning.KindScope {
			wrongTarget.Nodes[i].Target = "r"
		}
	}
	discourse, err := meaning.NewDiscourse("d", meaning.ActStatement, []meaning.NodeID{"neg", "s"})
	if err != nil {
		t.Fatalf("build target comparison discourse: %v", err)
	}
	wrongTarget.Nodes = append(wrongTarget.Nodes, discourse)
	wrongTarget.Roots = []meaning.NodeID{"d"}
	wrongTarget, err = meaning.Build(wrongTarget.Roots, wrongTarget.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build negation with referent target: %v", err)
	}
	expectedTarget := cloneNegationGraph(proximalGraph)
	expectedDiscourse, err := meaning.NewDiscourse("d", meaning.ActStatement, []meaning.NodeID{"neg", "s"})
	if err != nil {
		t.Fatalf("build expected target comparison discourse: %v", err)
	}
	expectedTarget.Nodes = append(expectedTarget.Nodes, expectedDiscourse)
	expectedTarget.Roots = []meaning.NodeID{"d"}
	expectedTarget, err = meaning.Build(expectedTarget.Roots, expectedTarget.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build expected scope target graph: %v", err)
	}
	if matches, err := matrix.GraphMatches(expectedTarget, wrongTarget); err != nil || matches {
		t.Errorf("negation targeting the referent must differ from predication negation: matches=%t error=%v", matches, err)
	}

	negatedSeeing, _, ok, err := interlingua.Negation("negation.seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("reload negated seeing fixture: ok=%t error=%v", ok, err)
	}
	negatedDogTeacher, _, ok, err := interlingua.Negation("negation.seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("load negated seeing fixture for role check: ok=%t error=%v", ok, err)
	}
	for index := range negatedDogTeacher.Nodes {
		if negatedDogTeacher.Nodes[index].Kind != meaning.KindEvent {
			continue
		}
		for participant := range negatedDogTeacher.Nodes[index].Participants {
			if negatedDogTeacher.Nodes[index].Participants[participant].Role == "experiencer" {
				negatedDogTeacher.Nodes[index].Participants[participant].Target = "d"
			} else if negatedDogTeacher.Nodes[index].Participants[participant].Role == "stimulus" {
				negatedDogTeacher.Nodes[index].Participants[participant].Target = "t"
			}
		}
	}
	negatedDogTeacher, err = meaning.Build(negatedDogTeacher.Roots, negatedDogTeacher.Nodes, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("build reversed-role negation fixture: %v", err)
	}
	if matches, err := matrix.GraphMatches(negatedSeeing, negatedDogTeacher); err != nil || matches {
		t.Errorf("negated seeing with swapped roles must differ: matches=%t error=%v", matches, err)
	}

	projection, err := meaning.ProjectAsserted(proximalGraph, meaning.DefaultBudget())
	if err != nil {
		t.Fatalf("project negated classification: %v", err)
	}
	if len(projection.Asserted) != 0 || len(projection.Unresolved) != 1 || projection.Unresolved[0].Reason != meaning.UnresolvedNegated {
		t.Errorf("negated content must remain unresolved and not project as a positive assertion: %+v", projection)
	}
}

func mustGraphMatch(t *testing.T, left, right meaning.Meaning) bool {
	t.Helper()
	matches, err := matrix.GraphMatches(left, right)
	if err != nil {
		t.Errorf("compare graphs: %v", err)
	}
	return matches
}

func findFixtureNode(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}

func cloneNegationGraph(graph meaning.Meaning) meaning.Meaning {
	clone, err := meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget())
	if err != nil {
		panic(err)
	}
	return clone
}

func removeNegationRoot(graph meaning.Meaning) meaning.Meaning {
	clone := cloneNegationGraph(graph)
	clone.Roots = []meaning.NodeID{"s"}
	clone.Nodes = clone.Nodes[1:]
	positive, err := meaning.Build(clone.Roots, clone.Nodes, meaning.DefaultBudget())
	if err != nil {
		panic(err)
	}
	return positive
}
