package interlingua_test

import (
	"errors"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

// These rejected candidates prove that CORE relation concepts cannot be used
// as full Meaning states; they are not fixtures for the sentence matrix.
func TestPossessionAndKinshipRelationsAreNotMeaningStates(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	cases := []struct {
		predicate string
		roles     []meaning.Participant
		concepts  []string
	}{
		{"temporary_possession", []meaning.Participant{{Role: "holder", Target: "a"}, {Role: "object", Target: "b"}}, []string{"teacher", "book"}},
		{"legal_ownership", []meaning.Participant{{Role: "owner", Target: "a"}, {Role: "asset", Target: "b"}}, []string{"teacher", "car"}},
		{"siblinghood", []meaning.Participant{{Role: "sibling_a", Target: "a"}, {Role: "sibling_b", Target: "b"}}, []string{"person", "person"}},
		{"friendship_relationship", []meaning.Participant{{Role: "friend_a", Target: "a"}, {Role: "friend_b", Target: "b"}}, []string{"person", "person"}},
	}
	for _, tc := range cases {
		t.Run(tc.predicate, func(t *testing.T) {
			predicate, err := ids.NewConceptID(tc.predicate)
			if err != nil {
				t.Fatal(err)
			}
			state, err := meaning.NewState("s", predicate, tc.roles)
			if err != nil {
				t.Fatal(err)
			}
			nodes := []meaning.Node{state}
			for i, concept := range tc.concepts {
				id, err := ids.NewConceptID(concept)
				if err != nil {
					t.Fatal(err)
				}
				ref, err := meaning.NewReferent(meaning.NodeID(string(rune('a'+i))), id)
				if err != nil {
					t.Fatal(err)
				}
				nodes = append(nodes, ref)
			}
			graph, err := meaning.Build([]meaning.NodeID{"s"}, nodes, meaning.DefaultBudget())
			if err != nil {
				t.Fatal(err)
			}
			if err := meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()); err == nil {
				t.Fatal("expected KindRelation to be rejected as a Meaning state")
			} else {
				var resultError interface{ Result() status.Result }
				if !errors.As(err, &resultError) {
					t.Fatalf("expected typed CONCEPT_KIND_MISMATCH, got %v", err)
				}
				found := false
				for _, diagnostic := range resultError.Result().Diagnostics {
					if diagnostic.Code == "CONCEPT_KIND_MISMATCH" {
						found = true
					}
				}
				if !found {
					t.Fatalf("expected CONCEPT_KIND_MISMATCH, got %+v", resultError.Result().Diagnostics)
				}
			}
		})
	}
}
