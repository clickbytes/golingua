package interlingua_test

import (
	"errors"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestPendingDomainRepresentableMeaning(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	for _, id := range []string{"quantity.exactly-two-cats", "question.classification.this", "question.classification.to"} {
		t.Run(id, func(t *testing.T) {
			graph, fixture, ok, err := interlingua.PendingDomain(id)
			if err != nil || !ok {
				t.Fatalf("build independent fixture: ok=%t err=%v", ok, err)
			}
			if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
				t.Fatalf("fixture does not validate against CORE: %v", err)
			}
			if fixture != nil {
				if err := meaning.ValidateReferences(graph, fixture.Context, meaning.DefaultBudget()); err != nil {
					t.Fatalf("fixture context does not validate: %v", err)
				}
			}
			if id == "quantity.exactly-two-cats" {
				assertExactlyTwoCats(t, graph)
				return
			}
			if fixture == nil {
				t.Fatal("question fixture lacks reference context")
			}
			assertClassificationQuestion(t, graph, fixture.Context, id == "question.classification.this")
		})
	}
}

func assertExactlyTwoCats(t *testing.T, graph meaning.Meaning) {
	t.Helper()
	if len(graph.Roots) != 1 {
		t.Fatalf("exact-two fixture roots = %v, want one", graph.Roots)
	}
	nodes := nodeMap(graph)
	root := nodes[graph.Roots[0]]
	if root.Kind != meaning.KindQuantifier || root.Quantifier == nil || root.Quantifier.Kind != meaning.QuantifierExactly || root.Quantifier.Count == nil || *root.Quantifier.Count != 2 {
		t.Fatalf("root does not assert exactly two: %+v", root)
	}
	cat := nodes[root.Target]
	if cat.Kind != meaning.KindReferent || cat.Concept != "concept:cat" || cat.Multiplicity != meaning.MultiplicityMany {
		t.Fatalf("quantifier target does not denote a many-cat set: %+v", cat)
	}
}

func assertClassificationQuestion(t *testing.T, graph meaning.Meaning, context meaning.RequestContext, proximal bool) {
	t.Helper()
	if len(graph.Roots) != 1 {
		t.Fatalf("question roots = %v, want one discourse root", graph.Roots)
	}
	nodes := nodeMap(graph)
	root := nodes[graph.Roots[0]]
	if root.Kind != meaning.KindDiscourse || root.Act != meaning.ActQuestion || len(root.Content) != 1 {
		t.Fatalf("root is not a single question discourse act: %+v", root)
	}
	question := nodes[root.Content[0]]
	if question.Kind != meaning.KindQuestion {
		t.Fatalf("discourse content is not a question: %+v", question)
	}
	state := nodes[question.Target]
	if state.Kind != meaning.KindState || state.Concept != "concept:class_membership" || len(state.Participants) != 1 || state.Participants[0].Role != "member" {
		t.Fatalf("question does not ask for cat class membership: %+v", state)
	}
	member := nodes[state.Participants[0].Target]
	if member.Kind != meaning.KindReferent || member.Concept != "concept:cat" {
		t.Fatalf("class member is not a cat referent: %+v", member)
	}
	if len(context.Mentions) != 1 || context.Mentions[0].NodeID != member.ID || context.Mentions[0].Anchor == nil {
		t.Fatalf("question context lacks one anchored, unresolved mention: %+v", context)
	}
	want := meaning.DeicticAnchor{Kind: "deixis", Value: "unresolved"}
	if proximal {
		want.Value = "proximal"
	}
	if *context.Mentions[0].Anchor != want || context.Mentions[0].ReferentID != "" || context.Mentions[0].EvidenceID != "" {
		t.Fatalf("question reference = %+v, want unresolved identity anchored at %+v", context.Mentions[0], want)
	}
}

func TestPendingDomainRelationsAreNotMeaningStates(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	for _, concept := range []string{"located_in", "temporary_possession"} {
		t.Run(concept, func(t *testing.T) {
			id, err := ids.NewConceptID(concept)
			if err != nil {
				t.Fatal(err)
			}
			roles := []meaning.Participant{{Role: "holder", Target: "a"}, {Role: "object", Target: "b"}}
			leftConcept, rightConcept := "teacher", "book"
			if concept == "located_in" {
				roles = []meaning.Participant{{Role: "occupant", Target: "a"}, {Role: "location", Target: "b"}}
				leftConcept, rightConcept = "cat", "chair"
			}
			state, err := meaning.NewState("s", id, roles)
			if err != nil {
				t.Fatal(err)
			}
			a, _ := ids.NewConceptID(leftConcept)
			b, _ := ids.NewConceptID(rightConcept)
			left, _ := meaning.NewReferent("a", a)
			right, _ := meaning.NewReferent("b", b)
			graph, err := meaning.Build([]meaning.NodeID{"s"}, []meaning.Node{state, left, right}, meaning.DefaultBudget())
			if err != nil {
				t.Fatal(err)
			}
			err = meaning.ValidateMeaning(graph, core, meaning.DefaultBudget())
			assertDiagnostic(t, err, "CONCEPT_KIND_MISMATCH")
		})
	}
}

func nodeMap(graph meaning.Meaning) map[meaning.NodeID]meaning.Node {
	result := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		result[node.ID] = node
	}
	return result
}

func assertDiagnostic(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s validation error", code)
	}
	var resultError interface{ Result() status.Result }
	if !errors.As(err, &resultError) {
		t.Fatalf("expected typed validation error %s, got %T: %v", code, err, err)
	}
	for _, diagnostic := range resultError.Result().Diagnostics {
		if diagnostic.Code == code {
			return
		}
	}
	t.Fatalf("missing %s diagnostic: %+v", code, resultError.Result().Diagnostics)
}
