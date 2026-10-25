package interlingua_test

import (
	"fmt"
	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
	"reflect"
	"strings"
	"testing"
)

func TestCaseOrderExpansionIndependentGraphsValidate(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range interlingua.CaseOrderExpansionIDs() {
		f, ok, err := interlingua.CaseOrderExpansion(id)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%t err=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(f.Graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s CORE: %v", id, err)
		}
		if err := meaning.ValidateReferences(f.Graph, f.Context, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s refs: %v", id, err)
		}
		seen := map[meaning.NodeID]bool{}
		refs := map[meaning.NodeID]meaning.Node{}
		for _, n := range f.Graph.Nodes {
			if n.Kind == meaning.KindReferent {
				if seen[n.ID] {
					t.Errorf("%s duplicate mention node %s", id, n.ID)
				}
				seen[n.ID] = true
				refs[n.ID] = n
			}
		}
		events, states := 0, 0
		for _, n := range f.Graph.Nodes {
			if n.Kind == meaning.KindEvent {
				events++
				if n.TemporalFrame != meaning.TemporalFramePresent {
					t.Errorf("%s event frame=%q", id, n.TemporalFrame)
				}
				for _, p := range n.Participants {
					if _, ok := refs[p.Target]; !ok {
						t.Errorf("%s role %s missing referent", id, p.Role)
					}
				}
			}
			if n.Kind == meaning.KindState {
				states++
				if n.TemporalFrame.Valid() {
					t.Errorf("%s black attribute has frame %q", id, n.TemporalFrame)
				}
				if len(n.Participants) != 1 || n.Participants[0].Role != "bearer" {
					t.Errorf("%s malformed black state: %+v", id, n)
				} else if refs[n.Participants[0].Target].ID == "" {
					t.Errorf("%s black state target is not a referent", id)
				}
			}
		}
		if events != 1 {
			t.Errorf("%s event count=%d", id, events)
		}
		base := strings.TrimSuffix(id, ".offspring")
		if strings.HasSuffix(id, ".offspring") {
			age, ok1, e1 := interlingua.CaseOrderExpansion(base)
			off, ok2, e2 := interlingua.CaseOrderExpansion(id)
			if e1 != nil || e2 != nil || !ok1 || !ok2 {
				t.Errorf("child variants unavailable for %s", id)
			} else if reflect.DeepEqual(age.Graph, off.Graph) {
				t.Errorf("%s and offspring alias same graph", id)
			}
		}
	}
	for _, tc := range []struct {
		id   string
		want meaning.Multiplicity
	}{{"a1.11.case.006.pl", ""}, {"a1.11.case.006.en", meaning.MultiplicityOne}} {
		f, ok, err := interlingua.CaseOrderExpansion(tc.id)
		if err != nil || !ok {
			t.Fatalf("%s: %t %v", tc.id, ok, err)
		}
		var door meaning.Node
		for _, n := range f.Graph.Nodes {
			if n.Kind == meaning.KindReferent && n.Concept == "concept:door" {
				door = n
			}
		}
		if door.ID == "" || door.Multiplicity != tc.want {
			t.Errorf("%s door multiplicity=%q want %q", tc.id, door.Multiplicity, tc.want)
		}
	}
	blackRoles := map[int]string{1: "experiencer", 3: "stimulus", 4: "stimulus", 5: "stimulus", 6: "stimulus", 7: "stimulus", 8: "experiencer", 9: "stimulus", 10: "experiencer", 11: "stimulus", 12: "stimulus", 13: "experiencer", 21: "stimulus", 22: "stimulus", 23: "stimulus"}
	for row, role := range blackRoles {
		f, ok, err := interlingua.CaseOrderExpansion(fmt.Sprintf("a1.11.case.%03d.en", row))
		if err != nil || !ok {
			t.Fatalf("black row %d: %v", row, err)
		}
		var target meaning.NodeID
		for _, n := range f.Graph.Nodes {
			if n.Kind == meaning.KindEvent {
				for _, part := range n.Participants {
					if string(part.Role) == role {
						target = part.Target
					}
				}
			}
		}
		found := false
		for _, n := range f.Graph.Nodes {
			if n.Kind == meaning.KindState && len(n.Participants) == 1 && n.Participants[0].Target == target && n.Concept == "concept:black" {
				found = true
			}
		}
		if target == "" || !found {
			t.Errorf("row %d black state not bound to %s target %s", row, role, target)
		}
	}
	for _, n := range []int{2, 14, 37, 38, 39, 40, 41} {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.11.case.%03d.%s", n, lang)
			if _, ok, err := interlingua.CaseOrderExpansion(id); err != nil || ok {
				t.Errorf("gap %s returned graph: ok=%t err=%v", id, ok, err)
			}
		}
	}
}
