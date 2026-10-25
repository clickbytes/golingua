package interlingua_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestAmbiguityExpansionCatalogValidatesAndHasNoPerSourceAliases(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatal(err)
	}
	ids := interlingua.AmbiguityExpansionIDs()
	if len(ids) < 100 {
		t.Fatalf("canonical variant catalogue unexpectedly small: %d", len(ids))
	}
	graphs := map[string]meaning.Meaning{}
	seenIDs := map[string]bool{}
	for _, id := range ids {
		if seenIDs[id] {
			t.Errorf("duplicate canonical ID %s", id)
		}
		seenIDs[id] = true
		f, ok, err := interlingua.AmbiguityExpansion(id)
		if err != nil || !ok {
			t.Fatalf("%s: ok=%t err=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(f.Graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s CORE: %v", id, err)
		}
		if err := meaning.ValidateReferences(f.Graph, f.Context, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s refs: %v", id, err)
		}
		if len(f.Context.Referents) != 0 {
			t.Errorf("%s has resolved request referents", id)
		}
		parts := strings.Split(id, ".")
		source := strings.Join(parts[:5], ".")
		for prior, g := range graphs {
			if strings.HasPrefix(prior, source+"|") && reflect.DeepEqual(g, f.Graph) {
				t.Errorf("%s aliases graph from %s", id, strings.TrimPrefix(prior, source+"|"))
			}
		}
		graphs[source+"|"+id] = f.Graph
		var event *meaning.Node
		for i := range f.Graph.Nodes {
			n := &f.Graph.Nodes[i]
			if n.Kind == meaning.KindEvent {
				if event != nil {
					t.Errorf("%s has multiple events", id)
				}
				event = n
			}
		}
		if event == nil {
			t.Errorf("%s lacks event", id)
			continue
		}
		want := meaning.TemporalFramePresent
		if strings.Contains("."+id+".", ".past.") {
			want = meaning.TemporalFramePast
		}
		if event.TemporalFrame != want {
			t.Errorf("%s frame=%q want %q", id, event.TemporalFrame, want)
		}
	}
}

func TestAmbiguityExpansionReadFramesAndChildVariants(t *testing.T) {
	for row := 9; row <= 14; row++ {
		for _, lang := range []string{"en", "pl"} {
			base := fmt.Sprintf("a1.12.amb.%03d.%s", row, lang)
			if _, ok, err := interlingua.AmbiguityExpansion(base); err != nil || !ok {
				t.Fatalf("%s: %v", base, err)
			}
			_, past, err := interlingua.AmbiguityExpansion(base + ".past")
			if lang == "en" {
				if err != nil || !past {
					t.Errorf("%s lacks English past reading: %v", base, err)
				}
			} else if err != nil || past {
				t.Errorf("%s incorrectly licenses Polish past variant", base)
			}
		}
	}
	children := []int{12, 15, 16, 17, 18, 19, 20, 31, 32, 41}
	for _, row := range children {
		for _, lang := range []string{"en", "pl"} {
			base := fmt.Sprintf("a1.12.amb.%03d.%s", row, lang)
			age, ok, e := interlingua.AmbiguityExpansion(base)
			if e != nil || !ok {
				t.Fatalf("%s age: %v", base, e)
			}
			offID := base + ".offspring"
			off, ok, e := interlingua.AmbiguityExpansion(offID)
			if e != nil || !ok {
				t.Fatalf("%s: %v", offID, e)
			}
			if reflect.DeepEqual(age.Graph, off.Graph) {
				t.Errorf("%s age and offspring graphs alias", base)
			}
			if _, ok, _ := interlingua.AmbiguityExpansion(base + ".age"); ok {
				t.Errorf("%s accepted noncanonical .age", base)
			}
		}
	}
	for _, id := range []string{"a1.12.amb.012.en.past", "a1.12.amb.012.en.past.offspring"} {
		if _, ok, err := interlingua.AmbiguityExpansion(id); err != nil || !ok {
			t.Errorf("crossed read/child variant %s unavailable: %v", id, err)
		}
	}
	if _, ok, _ := interlingua.AmbiguityExpansion("a1.12.amb.012.en.offspring.past"); ok {
		t.Error("accepted noncanonical suffix order alias")
	}
}

func TestAmbiguityExpansionPolishSyncreticRoleSwap(t *testing.T) {
	for row := 21; row <= 26; row++ {
		base := fmt.Sprintf("a1.12.amb.%03d.pl", row)
		a, ok, err := interlingua.AmbiguityExpansion(base)
		if err != nil || !ok {
			t.Fatalf("%s: %v", base, err)
		}
		b, ok, err := interlingua.AmbiguityExpansion(base + ".swap")
		if err != nil || !ok {
			t.Fatalf("%s.swap: %v", base, err)
		}
		if _, ok, _ := interlingua.AmbiguityExpansion(fmt.Sprintf("a1.12.amb.%03d.en.swap", row)); ok {
			t.Errorf("English row %d licensed syncretic swap", row)
		}
		roles := func(g meaning.Meaning) map[string]meaning.NodeID {
			out := map[string]meaning.NodeID{}
			for _, n := range g.Nodes {
				if n.Kind == meaning.KindEvent {
					for _, p := range n.Participants {
						out[string(p.Role)] = p.Target
					}
				}
			}
			return out
		}
		ra, rb := roles(a.Graph), roles(b.Graph)
		if ra["experiencer"] != rb["stimulus"] || ra["stimulus"] != rb["experiencer"] {
			t.Errorf("row %d did not swap role bindings: %v vs %v", row, ra, rb)
		}
		refs := func(g meaning.Meaning) map[meaning.NodeID]meaning.Node {
			m := map[meaning.NodeID]meaning.Node{}
			for _, n := range g.Nodes {
				if n.Kind == meaning.KindReferent {
					m[n.ID] = n
				}
			}
			return m
		}
		ar, br := refs(a.Graph), refs(b.Graph)
		for id, x := range ar {
			y := br[id]
			if x.Concept != y.Concept || x.Multiplicity != y.Multiplicity {
				t.Errorf("row %d swap changed referent %s", row, id)
			}
		}
	}
}

func TestAmbiguityExpansionAnchorsQuantitiesAndDistinctMentions(t *testing.T) {
	for _, tc := range []struct {
		row    int
		en, pl []string
	}{
		{27, []string{"proximal", "distal"}, []string{"demonstrative", "distal"}},
		{28, []string{"distal", "proximal"}, []string{"distal", "demonstrative"}},
		{29, []string{"proximal", "distal"}, []string{"demonstrative", "distal"}},
		{30, []string{"distal", "proximal"}, []string{"distal", "demonstrative"}},
		{31, []string{"proximal", "distal"}, []string{"demonstrative", "distal"}},
		{32, []string{"distal", "proximal"}, []string{"distal", "demonstrative"}},
	} {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.12.amb.%03d.%s", tc.row, lang)
			f, ok, err := interlingua.AmbiguityExpansion(id)
			if err != nil || !ok {
				t.Fatalf("%s: %v", id, err)
			}
			want := tc.en
			if lang == "pl" {
				want = tc.pl
			}
			if len(f.Context.Mentions) != 2 {
				t.Fatalf("%s has %d mentions", id, len(f.Context.Mentions))
			}
			byNode := map[meaning.NodeID]string{}
			roleNodes := map[string]meaning.NodeID{}
			for _, n := range f.Graph.Nodes {
				if n.Kind == meaning.KindEvent {
					for _, part := range n.Participants {
						roleNodes[string(part.Role)] = part.Target
					}
				}
			}
			if roleNodes["experiencer"] != "r1" || roleNodes["stimulus"] != "r2" {
				t.Errorf("%s actor/target role binding=%v", id, roleNodes)
			}
			for _, m := range f.Context.Mentions {
				if m.ReferentID != "" || m.EvidenceID != "" || m.Anchor == nil {
					t.Errorf("%s resolved/unanchored mention %+v", id, m)
				}
				byNode[m.NodeID] = m.Anchor.Value
			}
			if byNode["r1"] != want[0] || byNode["r2"] != want[1] {
				t.Errorf("%s anchors by role node=%v want %v", id, byNode, want)
			}
		}
	}
	for _, tc := range []struct {
		row  int
		role string
		ref  string
	}{{35, "stimulus", "r2"}, {36, "stimulus", "r2"}, {41, "experiencer", "r1"}} {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.12.amb.%03d.%s", tc.row, lang)
			f, ok, err := interlingua.AmbiguityExpansion(id)
			if err != nil || !ok {
				t.Fatalf("%s: %v", id, err)
			}
			var event *meaning.Node
			var q *meaning.Node
			refs := map[meaning.NodeID]meaning.Node{}
			for i := range f.Graph.Nodes {
				n := &f.Graph.Nodes[i]
				if n.Kind == meaning.KindEvent {
					event = n
				}
				if n.Kind == meaning.KindQuantifier {
					q = n
				}
				if n.Kind == meaning.KindReferent {
					refs[n.ID] = *n
				}
			}
			if event == nil || event.TemporalFrame != meaning.TemporalFramePresent || q == nil || q.Target != meaning.NodeID(tc.ref) || q.Quantifier == nil || q.Quantifier.Kind != meaning.QuantifierExactly || q.Quantifier.Count == nil || *q.Quantifier.Count != 2 {
				t.Errorf("%s event/quantifier invalid", id)
				continue
			}
			if refs[q.Target].Multiplicity != meaning.MultiplicityMany {
				t.Errorf("%s counted referent not many", id)
			}
			hasRole := false
			for _, p := range event.Participants {
				if string(p.Role) == tc.role && p.Target == q.Target {
					hasRole = true
				}
			}
			if !hasRole {
				t.Errorf("%s quantifier target is not participant role %s", id, tc.role)
			}
		}
	}
	f, ok, err := interlingua.AmbiguityExpansion("a1.12.amb.042.en")
	if err != nil || !ok {
		t.Fatal(err)
	}
	var targets []meaning.NodeID
	var concepts []meaning.ConceptRef
	for _, n := range f.Graph.Nodes {
		if n.Kind == meaning.KindReferent {
			concepts = append(concepts, n.Concept)
		}
	}
	for _, n := range f.Graph.Nodes {
		if n.Kind == meaning.KindEvent {
			for _, p := range n.Participants {
				targets = append(targets, p.Target)
			}
		}
	}
	if len(concepts) != 2 || concepts[0] != concepts[1] || len(targets) != 2 || targets[0] == targets[1] {
		t.Errorf("same-class mentions merged: concepts=%v role targets=%v", concepts, targets)
	}
	for _, row := range []int{33, 34, 37, 38, 39, 40, 43} {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.12.amb.%03d.%s", row, lang)
			if _, ok, err := interlingua.AmbiguityExpansion(id); err != nil || ok {
				t.Errorf("gap %s fabricated graph: ok=%t err=%v", id, ok, err)
			}
		}
	}
	for _, row := range []int{5, 6} {
		for _, lang := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.12.amb.%03d.%s", row, lang)
			f, ok, err := interlingua.AmbiguityExpansion(id)
			if err != nil || !ok {
				t.Fatalf("%s: %v", id, err)
			}
			for _, n := range f.Graph.Nodes {
				if n.Kind == meaning.KindReferent && n.Multiplicity != "" && n.Concept == "concept:water" || n.Kind == meaning.KindReferent && n.Multiplicity != "" && n.Concept == "concept:bread" {
					t.Errorf("%s mass referent has multiplicity %q", id, n.Multiplicity)
				}
			}
		}
	}
	for _, lang := range []string{"en", "pl"} {
		we, ok, err := interlingua.AmbiguityExpansion("a1.12.amb.014." + lang)
		if err != nil || !ok {
			t.Fatal(err)
		}
		var person meaning.Node
		for _, n := range we.Graph.Nodes {
			if n.Kind == meaning.KindReferent && n.Concept == "concept:person" {
				person = n
			}
		}
		if person.ID != "r1" || person.Multiplicity != meaning.MultiplicityMany || len(we.Context.Mentions) != 1 || we.Context.Mentions[0].NodeID != "r1" || we.Context.Mentions[0].Anchor == nil || we.Context.Mentions[0].Anchor.Value != "speaker_inclusive_group" || we.Context.Mentions[0].ReferentID != "" || we.Context.Mentions[0].EvidenceID != "" {
			t.Errorf("we %s context/actor invalid: %+v %+v", lang, person, we.Context)
		}
	}
}
