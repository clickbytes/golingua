package interlingua

import (
	"fmt"
	"github.com/clickbytes/golingua/internal/bundled"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"testing"
)

func TestQuestionExpansionFixturesValidate(t *testing.T) {
	candidate, err := bundled.Candidate()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := pack.Compile(candidate)
	if err != nil {
		t.Fatal(err)
	}
	for id := range questionSpecs {
		for _, lang := range []string{"en", "pl"} {
			variants := []string{fmt.Sprintf("a1.9.q.%03d.%s", id, lang)}
			if id == 9 {
				variants = append(variants, fmt.Sprintf("a1.9.q.%03d.%s.offspring", id, lang))
			}
			for _, fixtureID := range variants {
				fixture, ok, err := QuestionExpansion(fixtureID)
				if err != nil || !ok {
					t.Fatalf("fixture %s: ok=%t err=%v", fixtureID, ok, err)
				}
				if err := meaning.ValidateMeaning(fixture.Graph, snapshot.CoreGraph(), meaning.DefaultBudget()); err != nil {
					t.Errorf("fixture %s CORE validation: %v", fixtureID, err)
				}
				if err := meaning.ValidateReferences(fixture.Graph, fixture.Context, meaning.DefaultBudget()); err != nil {
					t.Errorf("fixture %s references: %v", fixtureID, err)
				}
				for _, mention := range fixture.Context.Mentions {
					if mention.ReferentID != "" || mention.EvidenceID != "" {
						t.Errorf("fixture %s resolves identity without evidence", fixtureID)
					}
				}
				root := findQuestionFixtureNode(fixture.Graph, fixture.Graph.Roots[0])
				if root.Kind != meaning.KindDiscourse || root.Act != meaning.ActQuestion {
					t.Errorf("fixture %s root is not an ActQuestion", fixtureID)
				}
			}
		}
	}
}

func findQuestionFixtureNode(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}

func TestQuestionExpansionDoesNotInventWHVariableFixtures(t *testing.T) {
	for i := 11; i <= 44; i++ {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.9.q.%03d.%s", i, lang)
			if _, ok, err := QuestionExpansion(id); err != nil || ok {
				t.Errorf("WH gap %s unexpectedly has a fixture: ok=%t err=%v", id, ok, err)
			}
		}
	}
	for _, i := range []int{51, 52, 53, 55, 56, 57, 58} {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.9.q.%03d.%s", i, lang)
			if _, ok, err := QuestionExpansion(id); err != nil || ok {
				t.Errorf("representation gap %s unexpectedly has a fixture: ok=%t err=%v", id, ok, err)
			}
		}
	}
}
