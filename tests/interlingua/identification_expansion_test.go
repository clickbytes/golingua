package interlingua_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestIdentificationExpansionFixturesAndMissingColorGap(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	fixtureIDs := interlingua.IdentificationExpansionFixtureIDs()
	if len(fixtureIDs) != 68 {
		t.Fatalf("fixture IDs = %d; want 68", len(fixtureIDs))
	}
	manyPairs := map[int]bool{21: true, 22: true, 31: true, 32: true, 33: true, 34: true}
	for _, id := range fixtureIDs {
		fixture, ok, err := interlingua.IdentificationExpansion(id)
		if err != nil || !ok {
			t.Fatalf("fixture %q: ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(fixture.Graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %q invalid against CORE: %v", id, err)
			continue
		}
		if err := meaning.ValidateReferences(fixture.Graph, fixture.Context, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %q has invalid context: %v", id, err)
		}
		if len(fixture.Graph.Roots) != 1 || fixture.Graph.Roots[0] != "s" || len(fixture.Graph.Nodes) != 2 {
			t.Errorf("fixture %q must have one predicate and one referent: %+v", id, fixture.Graph)
			continue
		}
		state, ref := fixture.Graph.Nodes[0], fixture.Graph.Nodes[1]
		if state.Kind != meaning.KindState || state.TemporalFrame != meaning.TemporalFramePresent || len(state.Participants) != 1 || ref.Kind != meaning.KindReferent || state.Participants[0].Target != ref.ID {
			t.Errorf("fixture %q has invalid identification structure: %+v", id, fixture.Graph)
			continue
		}
		wantRole := meaning.Role("member")
		if state.Concept == "concept:black" {
			wantRole = "bearer"
		}
		if state.Participants[0].Role != wantRole {
			t.Errorf("fixture %q role=%q want %q", id, state.Participants[0].Role, wantRole)
		}
		parts := strings.Split(id, ".")
		pair, parseErr := strconv.Atoi(parts[3])
		if parseErr != nil {
			t.Fatalf("fixture id %q: %v", id, parseErr)
		}
		wantMultiplicity := meaning.MultiplicityOne
		if manyPairs[pair] {
			wantMultiplicity = meaning.MultiplicityMany
		}
		if ref.Multiplicity != wantMultiplicity {
			t.Errorf("fixture %q multiplicity=%q want %q", id, ref.Multiplicity, wantMultiplicity)
		}
		if len(fixture.Context.Referents) != 0 || len(fixture.Context.Mentions) != 1 || fixture.Context.Mentions[0].NodeID != ref.ID || fixture.Context.Mentions[0].ReferentID != "" || fixture.Context.Mentions[0].EvidenceID != "" || fixture.Context.Mentions[0].Anchor == nil {
			t.Errorf("fixture %q must contain one unresolved anchored mention only: %+v", id, fixture.Context)
		}
	}
	for _, predicate := range []string{"red", "white", "blue", "green"} {
		predicateID, _ := ids.NewConceptID(predicate)
		if _, ok := core.Lookup(predicateID); ok {
			t.Errorf("unexpected CORE predicate %q", predicate)
		}
	}
	blackID, _ := ids.NewConceptID("black")
	if _, ok := core.Lookup(blackID); !ok {
		t.Fatal("CORE black predicate missing")
	}
	colorID, _ := ids.NewConceptID("color")
	color, ok := core.Lookup(colorID)
	if !ok {
		t.Fatal("CORE generic color missing")
	}
	for _, role := range color.RoleTypes {
		if role.Name == "bearer" {
			t.Fatal("generic color unexpectedly declares bearer role")
		}
	}
	bookID, _ := ids.NewConceptID("book")
	book, _ := meaning.NewReferent("r", bookID)
	book.Multiplicity = meaning.MultiplicityOne
	state, _ := meaning.NewState("s", colorID, []meaning.Participant{{Role: "bearer", Target: "r"}})
	state.TemporalFrame = meaning.TemporalFramePresent
	graph, err := meaning.Build([]meaning.NodeID{"s"}, []meaning.Node{state, book}, meaning.DefaultBudget())
	if err != nil {
		t.Fatal(err)
	}
	err = meaning.ValidateMeaning(graph, core, meaning.DefaultBudget())
	var statusError interface{ Result() status.Result }
	if !errors.As(err, &statusError) {
		t.Fatalf("generic color with invented bearer role error=%v; want typed UNKNOWN_ROLE", err)
	}
	found := false
	for _, diagnostic := range statusError.Result().Diagnostics {
		if diagnostic.Code == "UNKNOWN_ROLE" {
			found = true
		}
	}
	if !found {
		t.Fatalf("generic color error diagnostics=%+v; want UNKNOWN_ROLE", statusError.Result().Diagnostics)
	}
	for pair := 35; pair <= 38; pair++ {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.5.identification.%03d.%s", pair, lang)
			if _, ok, err := interlingua.IdentificationExpansion(id); err != nil || ok {
				t.Errorf("missing-color ID %q must not fabricate a fixture: ok=%t error=%v", id, ok, err)
			}
		}
	}
}
