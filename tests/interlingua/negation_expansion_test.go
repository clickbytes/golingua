package interlingua_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestNegationExpansionFixturesValidateAndRetainScope(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	for i := 1; i <= 45; i++ {
		id := fmt.Sprintf("a1.8.neg.%03d", i)
		graph, ok, err := interlingua.NegationExpansion(id)
		if err != nil {
			t.Fatalf("fixture %s: %v", id, err)
		}
		if !ok {
			continue
		}
		if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %s invalid: %v", id, err)
			continue
		}
		if len(graph.Roots) != 1 {
			t.Errorf("fixture %s must have one root", id)
			continue
		}
		root := nodeByID(graph, graph.Roots[0])
		if root.Kind != meaning.KindScope || root.Operator != meaning.OperatorNegation {
			t.Errorf("fixture %s root is not explicit negation", id)
			continue
		}
		target := nodeByID(graph, root.Target)
		if target.Kind != meaning.KindEvent && target.Kind != meaning.KindState {
			t.Errorf("fixture %s negation must target complete predicate", id)
		}
		for _, n := range graph.Nodes {
			if n.Kind != meaning.KindReferent {
				continue
			}
			if n.Concept == "concept:milk_food" || n.Concept == "concept:water" || n.Concept == "concept:bread" {
				if n.Multiplicity != "" {
					t.Errorf("fixture %s mass referent %s has multiplicity %q", id, n.ID, n.Multiplicity)
				}
			}
		}
	}
	for _, i := range []int{3, 12, 14, 37} {
		age, ok1, err1 := interlingua.NegationExpansion(fmt.Sprintf("a1.8.neg.%03d.age", i))
		offspring, ok2, err2 := interlingua.NegationExpansion(fmt.Sprintf("a1.8.neg.%03d.offspring", i))
		if err1 != nil || err2 != nil || !ok1 || !ok2 {
			t.Errorf("child ambiguity %d missing authored variants: %v / %v", i, err1, err2)
			continue
		}
		if !hasConcept(age, "concept:child") || !hasConcept(offspring, "concept:child_of_person") {
			t.Errorf("child ambiguity %d must retain age-child and offspring readings", i)
		}
	}
}

func TestNegationExpansionGapLabelsHaveNoFixture(t *testing.T) {
	for _, i := range []int{5, 28, 29, 30, 41, 42, 43, 44, 45} {
		id := fmt.Sprintf("a1.8.neg.%03d", i)
		if _, ok, err := interlingua.NegationExpansion(id); err != nil || ok {
			t.Errorf("gap %s must not claim a complete graph: ok=%t err=%v", id, ok, err)
		}
	}
}

func TestNegationExpansionPluralOnlyDoorCountVariants(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	for _, index := range []int{9, 25, 38} {
		baseID := fmt.Sprintf("a1.8.neg.%03d", index)
		polishID := baseID + ".pl"
		known, ok, err := interlingua.NegationExpansion(baseID)
		if err != nil || !ok {
			t.Fatalf("known-count fixture %s: ok=%t error=%v", baseID, ok, err)
		}
		unknown, ok, err := interlingua.NegationExpansion(polishID)
		if err != nil || !ok {
			t.Fatalf("PL plural-only fixture %s: ok=%t error=%v", polishID, ok, err)
		}
		for id, graph := range map[string]meaning.Meaning{baseID: known, polishID: unknown} {
			if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err != nil {
				t.Errorf("fixture %s does not validate against CORE: %v", id, err)
			}
			root := nodeByID(graph, graph.Roots[0])
			if root.Kind != meaning.KindScope || root.Operator != meaning.OperatorNegation || root.Target != "pred" {
				t.Errorf("fixture %s changed negation scope: %+v", id, root)
			}
		}
		knownPred, unknownPred := nodeByID(known, "pred"), nodeByID(unknown, "pred")
		if knownPred.Kind != unknownPred.Kind || knownPred.Concept != unknownPred.Concept || knownPred.TemporalFrame != unknownPred.TemporalFrame || !reflect.DeepEqual(knownPred.Participants, unknownPred.Participants) {
			t.Errorf("%s PL variant changed predicate, frame, or roles", baseID)
		}
		var knownDoor, unknownDoor meaning.Node
		for _, node := range known.Nodes {
			if node.Kind == meaning.KindReferent && node.Concept == "concept:door" {
				knownDoor = node
			}
		}
		for _, node := range unknown.Nodes {
			if node.Kind == meaning.KindReferent && node.Concept == "concept:door" {
				unknownDoor = node
			}
		}
		if knownDoor.ID == "" || knownDoor.Multiplicity != meaning.MultiplicityOne {
			t.Errorf("%s English source must retain door=one: %+v", baseID, knownDoor)
		}
		if unknownDoor.ID == "" || unknownDoor.Multiplicity != "" {
			t.Errorf("%s PL plural-only door must retain unknown count: %+v", polishID, unknownDoor)
		}
		if len(known.Nodes) != len(unknown.Nodes) {
			t.Errorf("%s PL variant changed graph node count", baseID)
		}
		for _, knownNode := range known.Nodes {
			unknownNode := nodeByID(unknown, knownNode.ID)
			if knownNode.Kind == meaning.KindReferent && knownNode.Concept == "concept:door" {
				knownNode.Multiplicity, unknownNode.Multiplicity = "", ""
			}
			if !reflect.DeepEqual(knownNode, unknownNode) {
				t.Errorf("%s PL variant changed node %s beyond door multiplicity", baseID, knownNode.ID)
			}
		}
	}
	for index := 1; index <= 45; index++ {
		if index == 9 || index == 25 || index == 38 {
			continue
		}
		id := fmt.Sprintf("a1.8.neg.%03d.pl", index)
		if _, ok, err := interlingua.NegationExpansion(id); err != nil || ok {
			t.Errorf("non-door PL alias %s must not be accepted: ok=%t error=%v", id, ok, err)
		}
	}
}

func nodeByID(g meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n
		}
	}
	return meaning.Node{}
}
func hasConcept(g meaning.Meaning, c string) bool {
	for _, n := range g.Nodes {
		if string(n.Concept) == c {
			return true
		}
	}
	return false
}

func TestNoCatSentenceNeedsParticipantQuantification(t *testing.T) {
	// Current operators quantify a node target; they have no typed link to a
	// participant slot. A quantifier over an event therefore cannot stand in
	// for “no cat” by itself. This row intentionally has no complete fixture.
	if _, ok, err := interlingua.NegationExpansion("a1.8.neg.044"); err != nil || ok {
		t.Fatalf("No cat drinks milk must remain an explicit gap, ok=%t err=%v", ok, err)
	}
	quantifier := reflect.TypeOf(meaning.Quantifier{})
	for i := 0; i < quantifier.NumField(); i++ {
		if quantifier.Field(i).Name == "Participant" || quantifier.Field(i).Name == "Role" {
			t.Fatal("quantifier unexpectedly has a participant-role binding; revisit the representation gap")
		}
	}
}
