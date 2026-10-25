package interlingua_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
)

// The supported quantifier fixture is a cardinality component, not a full
// finite existential Meaning.
func TestQuantifierAndReferentCannotCarryExistenceTime(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	base, _, ok, err := interlingua.PendingDomain("quantity.exactly-two-cats.component")
	if err != nil || !ok {
		t.Fatalf("load cardinality component: ok=%t err=%v", ok, err)
	}
	for _, kind := range []meaning.NodeKind{meaning.KindQuantifier, meaning.KindReferent} {
		t.Run(string(kind), func(t *testing.T) {
			candidate := base
			candidate.Nodes = append([]meaning.Node(nil), base.Nodes...)
			found := false
			for i := range candidate.Nodes {
				if candidate.Nodes[i].Kind == kind {
					candidate.Nodes[i].TemporalFrame = meaning.TemporalFramePresent
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("component has no %s node", kind)
			}
			assertDiagnostic(t, meaning.ValidateMeaning(candidate, core, meaning.DefaultBudget()), "UNEXPECTED_NODE_FIELD")
		})
	}
}

func TestCoreExistenceCannotBindCountedReferent(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	existenceID, _ := ids.NewConceptID("existence")
	existence, ok := core.Lookup(existenceID)
	if !ok {
		t.Fatal("CORE existence concept is unavailable")
	}
	if len(existence.RoleTypes) != 0 {
		t.Fatalf("CORE existence unexpectedly declares participant roles: %+v", existence.RoleTypes)
	}
	catID, _ := ids.NewConceptID("cat")
	cat, err := meaning.NewReferent("cats", catID)
	if err != nil {
		t.Fatal(err)
	}
	cat.Multiplicity = meaning.MultiplicityMany
	state, err := meaning.NewState("existence", existenceID, []meaning.Participant{{Role: "existent", Target: cat.ID}})
	if err != nil {
		t.Fatal(err)
	}
	state.TemporalFrame = meaning.TemporalFramePresent
	graph, err := meaning.Build([]meaning.NodeID{state.ID}, []meaning.Node{state, cat}, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	assertDiagnostic(t, meaning.ValidateMeaning(graph, core, meaning.DefaultBudget()), "UNKNOWN_ROLE")
}

func TestWomanConceptIsNotAvailableForCountContracts(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	womanID, err := ids.NewConceptID("woman")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.Lookup(womanID); ok {
		t.Fatal("unexpected CORE woman concept; review the explicit count-gap contract")
	}
}
