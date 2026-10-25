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

func TestCompositionExpansionCatalogueValidatesAndHasUniqueVariants(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("INFRA: compile CORE: %v", err)
	}
	seen := map[string]map[string]string{}
	ids := interlingua.CompositionExpansionIDs()
	if len(ids) != 42 {
		t.Fatalf("catalogue contains %d IDs; want 42", len(ids))
	}
	for _, id := range ids {
		fixture, ok, err := interlingua.CompositionExpansion(id)
		if err != nil || !ok {
			t.Fatalf("fixture %s: ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(fixture.Graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %s invalid against CORE: %v", id, err)
		}
		if err := meaning.ValidateReferences(fixture.Graph, fixture.Context, meaning.DefaultBudget()); err != nil {
			t.Errorf("fixture %s has invalid request context: %v", id, err)
		}
		parts := strings.Split(id, ".")
		group := strings.Join(parts[:5], ".")
		if seen[group] == nil {
			seen[group] = map[string]string{}
		}
		for otherID, otherGraph := range seen[group] {
			if otherGraph == fmt.Sprintf("%#v", fixture.Graph) {
				t.Errorf("catalogue aliases one graph for %s and %s", otherID, id)
			}
		}
		seen[group][id] = fmt.Sprintf("%#v", fixture.Graph)
	}
	for _, index := range []int{12, 13, 14, 15, 16, 17, 18, 22, 25, 27, 29, 30, 32} {
		for _, language := range []string{"en", "pl"} {
			id := fmt.Sprintf("a1.13.comp.%03d.%s", index, language)
			if _, ok, err := interlingua.CompositionExpansion(id); err != nil || ok {
				t.Errorf("full gap %s must have no fixture: ok=%t error=%v", id, ok, err)
			}
			if gap, ok := interlingua.CompositionExpansionGap(id); !ok || gap == "" {
				t.Errorf("full gap %s lacks its precise missing relation/operator", id)
			}
		}
	}
	if _, ok := interlingua.CompositionExpansionGap("a1.13.comp.001.en"); ok {
		t.Error("represented composition must not be labelled as a gap")
	}
	if _, ok, err := interlingua.CompositionExpansion("a1.13.comp.004.en.age"); err != nil || ok {
		t.Errorf("noncanonical age alias must be rejected: ok=%t error=%v", ok, err)
	}
}

func TestCompositionExpansionPreservesFramesRolesNegationAndAttributes(t *testing.T) {
	row6, _, _ := interlingua.CompositionExpansion("a1.13.comp.006.en")
	if compNode(row6.Graph.Nodes, "e1").TemporalFrame != meaning.TemporalFramePast || compNode(row6.Graph.Nodes, "e2").TemporalFrame != meaning.TemporalFramePresent {
		t.Errorf("row 6 must retain reading=past and seeing=present: %+v", row6.Graph.Nodes)
	}
	for _, tc := range []struct {
		row, event, role, target string
	}{
		{"003", "e1", "reader", "rteacher"}, {"003", "e2", "experiencer", "rteacher"},
		{"004", "e1", "drinker", "rchild"}, {"004", "e2", "eater", "rchild"},
		{"007", "e1", "experiencer", "rmother"}, {"007", "e2", "experiencer", "rmother"},
		{"008", "e1", "experiencer", "rteacher"}, {"008", "e2", "experiencer", "rmother"},
		{"009", "e1", "experiencer", "rdoctor"}, {"009", "e2", "experiencer", "rdoctor"},
		{"033", "e1", "experiencer", "rmother"}, {"033", "e1", "stimulus", "rmother"},
	} {
		fixture, ok, err := interlingua.CompositionExpansion("a1.13.comp." + tc.row + ".en")
		if err != nil || !ok {
			t.Fatalf("fixture row %s: ok=%t error=%v", tc.row, ok, err)
		}
		event := compNode(fixture.Graph.Nodes, meaning.NodeID(tc.event))
		if !hasParticipant(event, tc.role, tc.target) {
			t.Errorf("row %s %s lacks %s→%s: %+v", tc.row, tc.event, tc.role, tc.target, event.Participants)
		}
	}
	for _, id := range []string{"a1.13.comp.007.en", "a1.13.comp.008.en", "a1.13.comp.009.en", "a1.13.comp.019.en", "a1.13.comp.020.en", "a1.13.comp.031.en"} {
		fixture, _, _ := interlingua.CompositionExpansion(id)
		for _, n := range fixture.Graph.Nodes {
			if n.Kind == meaning.KindScope && (n.Operator != meaning.OperatorNegation || n.Target == "") {
				t.Errorf("%s has malformed negative scope: %+v", id, n)
			}
		}
	}
	for _, id := range []string{"a1.13.comp.019.en", "a1.13.comp.020.en"} {
		fixture, _, _ := interlingua.CompositionExpansion(id)
		if len(fixture.Graph.Roots) != 2 {
			t.Errorf("%s must retain the independent black-state and negated-event roots: %v", id, fixture.Graph.Roots)
		}
		for _, rootID := range fixture.Graph.Roots {
			root := compNode(fixture.Graph.Nodes, rootID)
			if root.Kind == meaning.KindScope && root.Target != "e1" {
				t.Errorf("%s negation must scope over event only: %+v", id, root)
			}
		}
	}
	for _, row := range []string{"004", "010", "011", "019", "020"} {
		fixture, _, _ := interlingua.CompositionExpansion("a1.13.comp." + row + ".en")
		for _, node := range fixture.Graph.Nodes {
			if node.Kind == meaning.KindState && node.Concept == "concept:black" && node.TemporalFrame != "" {
				t.Errorf("row %s black attribute must be untimed: %+v", row, node)
			}
		}
	}
	for _, pair := range []struct{ base, offspring string }{
		{"a1.13.comp.004.en", "a1.13.comp.004.en.offspring"},
		{"a1.13.comp.004.pl", "a1.13.comp.004.pl.offspring"},
	} {
		age, ok1, err1 := interlingua.CompositionExpansion(pair.base)
		offspring, ok2, err2 := interlingua.CompositionExpansion(pair.offspring)
		if err1 != nil || err2 != nil || !ok1 || !ok2 {
			t.Fatalf("row 4 child variants missing: %v / %v", err1, err2)
		}
		if !hasConcept(age.Graph, "concept:child") || !hasConcept(offspring.Graph, "concept:child_of_person") {
			t.Errorf("row 4 must distinguish age child from offspring")
		}
	}
	if ids := interlingua.CompositionExpansionIDs(); reflect.DeepEqual(ids, []string{}) {
		t.Fatal("composition catalogue unexpectedly empty")
	}
}

func TestCompositionExpansionRequestActsAndParticipantAnchors(t *testing.T) {
	for _, row := range []string{"021", "023", "024", "026", "028"} {
		for _, language := range []string{"en", "pl"} {
			id := "a1.13.comp." + row + "." + language
			fixture, ok, err := interlingua.CompositionExpansion(id)
			if err != nil || !ok {
				t.Fatalf("request fixture %s: ok=%t error=%v", id, ok, err)
			}
			root := compNode(fixture.Graph.Nodes, "request")
			if root.Kind != meaning.KindDiscourse || root.Act != meaning.ActRequest {
				t.Errorf("%s must be an ActRequest: %+v", id, root)
			}
			for _, n := range fixture.Graph.Nodes {
				if (n.Kind == meaning.KindEvent || n.Kind == meaning.KindState) && n.TemporalFrame != "" {
					t.Errorf("%s request content must not assert time: %+v", id, n)
				}
			}
			for _, mention := range fixture.Context.Mentions {
				if mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil {
					t.Errorf("%s participant anchor must remain unresolved: %+v", id, mention)
				}
			}
			actor := compNode(fixture.Graph.Nodes, "rperson")
			wantCount := meaning.MultiplicityOne
			if language == "en" {
				wantCount = ""
			}
			if actor.Multiplicity != wantCount {
				t.Errorf("%s addressee count=%q; want %q", id, actor.Multiplicity, wantCount)
			}
			if !hasAnchor(fixture.Context, "rperson", "addressee") {
				t.Errorf("%s actor is missing unresolved addressee anchor", id)
			}
		}
	}
	fixture, _, _ := interlingua.CompositionExpansion("a1.13.comp.026.en")
	speaker := compNode(fixture.Graph.Nodes, "rspeaker")
	if speaker.Concept != "concept:person" || speaker.Multiplicity != meaning.MultiplicityOne || !hasAnchor(fixture.Context, "rspeaker", "speaker") {
		t.Errorf("row 26 recipient must be one speaker-anchored person: %+v / %+v", speaker, fixture.Context)
	}
	if !hasParticipant(compNode(fixture.Graph.Nodes, "e1"), "recipient", "rspeaker") {
		t.Errorf("row 26 giving role must target the speaker recipient")
	}
	for _, row := range []string{"024", "028"} {
		fixture, _, _ := interlingua.CompositionExpansion("a1.13.comp." + row + ".en")
		if len(compNode(fixture.Graph.Nodes, "request").Content) != 2 {
			t.Errorf("row %s must have one request act containing two ordered requested actions", row)
		}
	}
	for _, row := range []string{"021", "023", "024"} {
		for _, language := range []string{"en", "pl"} {
			id := "a1.13.comp." + row + "." + language
			fixture, ok, err := interlingua.CompositionExpansion(id)
			if err != nil || !ok {
				t.Fatalf("door request fixture %s: ok=%t error=%v", id, ok, err)
			}
			door := compNode(fixture.Graph.Nodes, "rdoor")
			wantDoor := meaning.MultiplicityOne
			if language == "pl" {
				wantDoor = ""
			}
			if door.Multiplicity != wantDoor {
				t.Errorf("%s door count=%q; want %q", id, door.Multiplicity, wantDoor)
			}
			if len(fixture.Graph.Roots) != 1 || fixture.Graph.Roots[0] != "request" {
				t.Errorf("%s must have only the request act as a graph root: %v", id, fixture.Graph.Roots)
			}
			request := compNode(fixture.Graph.Nodes, "request")
			wantContents := 1
			if row == "024" {
				wantContents = 2
			}
			if request.Kind != meaning.KindDiscourse || request.Act != meaning.ActRequest || len(request.Content) != wantContents {
				t.Errorf("%s request must contain %d requested event(s): %+v", id, wantContents, request)
			}
			actor := compNode(fixture.Graph.Nodes, "rperson")
			wantActor := meaning.MultiplicityOne
			if language == "en" {
				wantActor = ""
			}
			if actor.Multiplicity != wantActor || !hasAnchor(fixture.Context, "rperson", "addressee") {
				t.Errorf("%s addressee must retain source-specific count and unresolved anchor: %+v / %+v", id, actor, fixture.Context)
			}
			for _, node := range fixture.Graph.Nodes {
				if node.Kind == meaning.KindEvent && node.TemporalFrame != "" {
					t.Errorf("%s request event must not assert a finite frame: %+v", id, node)
				}
			}
		}
	}
}

func hasParticipant(node meaning.Node, role, target string) bool {
	for _, part := range node.Participants {
		if string(part.Role) == role && string(part.Target) == target {
			return true
		}
	}
	return false
}

func compNode(nodes []meaning.Node, id meaning.NodeID) meaning.Node {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}

func hasAnchor(context meaning.RequestContext, nodeID meaning.NodeID, value string) bool {
	for _, mention := range context.Mentions {
		if mention.NodeID == nodeID && mention.Anchor != nil && mention.Anchor.Value == value {
			return true
		}
	}
	return false
}
