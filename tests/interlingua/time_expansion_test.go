package interlingua_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/tests/interlingua"
)

func TestTimeExpansionFixturesValidateAndPreserveSourceAttributes(t *testing.T) {
	core, err := catalog.Compile()
	if err != nil {
		t.Fatalf("compile CORE: %v", err)
	}
	ids := interlingua.TimeExpansionIDs()
	if len(ids) < 96 {
		t.Fatalf("licensed fixture enumeration unexpectedly small: %d", len(ids))
	}
	for _, id := range ids {
		fixture, ok, err := interlingua.TimeExpansion(id)
		if err != nil || !ok {
			t.Fatalf("TimeExpansion(%q): ok=%t error=%v", id, ok, err)
		}
		if err := meaning.ValidateMeaning(fixture.Graph, core, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s does not validate against CORE: %v", id, err)
		}
		if err := meaning.ValidateReferences(fixture.Graph, fixture.Context, meaning.DefaultBudget()); err != nil {
			t.Errorf("%s references invalid: %v", id, err)
		}
		var pred *meaning.Node
		refs := map[meaning.NodeID]meaning.Node{}
		for i := range fixture.Graph.Nodes {
			n := &fixture.Graph.Nodes[i]
			if n.Kind == meaning.KindEvent || n.Kind == meaning.KindState {
				pred = n
			}
			if n.Kind == meaning.KindReferent {
				refs[n.ID] = *n
			}
		}
		if pred == nil || !pred.TemporalFrame.Valid() {
			t.Errorf("%s lacks a supported temporal frame", id)
			continue
		}
		parts := strings.Split(id, ".")
		index, _ := strconv.Atoi(parts[3])
		wantFrame := meaning.TemporalFramePresent
		if index >= 17 && index <= 32 || index >= 41 && index <= 48 || hasSuffix(id, ".past") {
			wantFrame = meaning.TemporalFramePast
		}
		if index == 22 && strings.HasSuffix(id, ".en") {
			wantFrame = meaning.TemporalFramePast
		}
		if hasSuffix(id, ".present") {
			wantFrame = meaning.TemporalFramePresent
		}
		if pred.TemporalFrame != wantFrame {
			t.Errorf("%s frame=%q, want %q", id, pred.TemporalFrame, wantFrame)
		}
		for _, participant := range pred.Participants {
			ref, exists := refs[participant.Target]
			if !exists {
				t.Errorf("%s role %s has no referent", id, participant.Role)
				continue
			}
			if strings.HasPrefix(string(ref.Concept), "concept:") && (ref.Concept == "concept:coffee_beverage" || ref.Concept == "concept:tea_beverage" || ref.Concept == "concept:water" || ref.Concept == "concept:milk_food" || ref.Concept == "concept:bread") {
				if ref.Multiplicity != "" {
					t.Errorf("%s mass referent %s has multiplicity %q", id, ref.ID, ref.Multiplicity)
				}
			}
		}
	}
}

func TestTimeExpansionPersonAnchorsAndLanguageSpecificCounts(t *testing.T) {
	for _, tc := range []struct {
		id     string
		count  meaning.Multiplicity
		anchor string
	}{
		{"a1.10.time.001.en", meaning.MultiplicityOne, "speaker"},
		{"a1.10.time.003.pl", meaning.MultiplicityMany, "speaker_inclusive_group"},
		{"a1.10.time.004.en", "", "addressee"},
		{"a1.10.time.004.pl", meaning.MultiplicityMany, "addressee"},
		{"a1.10.time.002.en", "", "addressee"},
	} {
		fixture, ok, err := interlingua.TimeExpansion(tc.id)
		if err != nil || !ok {
			t.Fatalf("TimeExpansion(%q): ok=%t error=%v", tc.id, ok, err)
		}
		var person meaning.Node
		for _, n := range fixture.Graph.Nodes {
			if n.Kind == meaning.KindReferent && n.Concept == "concept:person" {
				person = n
				break
			}
		}
		if person.ID == "" || person.Multiplicity != tc.count {
			t.Errorf("%s person multiplicity=%q, want %q", tc.id, person.Multiplicity, tc.count)
		}
		if len(fixture.Context.Referents) != 0 || len(fixture.Context.Mentions) != 1 || fixture.Context.Mentions[0].ReferentID != "" || fixture.Context.Mentions[0].EvidenceID != "" || fixture.Context.Mentions[0].Anchor == nil || fixture.Context.Mentions[0].Anchor.Value != tc.anchor {
			t.Errorf("%s has unexpected participant anchor context: %+v", tc.id, fixture.Context)
		}
	}
}

func TestTimeExpansionChildAndDoorDistinctionsAndExplicitGaps(t *testing.T) {
	for _, tc := range []struct{ id, concept string }{
		{"a1.10.time.025.en", "child"}, {"a1.10.time.025.pl.offspring", "child_of_person"},
		{"a1.10.time.008.en", "child"}, {"a1.10.time.008.pl.offspring", "child_of_person"},
		{"a1.10.time.045.pl", "door"}, {"a1.10.time.045.en", "door"},
	} {
		fixture, ok, err := interlingua.TimeExpansion(tc.id)
		if err != nil || !ok {
			t.Fatalf("TimeExpansion(%q): ok=%t error=%v", tc.id, ok, err)
		}
		for _, ref := range fixture.Graph.Nodes {
			if ref.Kind != meaning.KindReferent {
				continue
			}
			if tc.id == "a1.10.time.045.pl" && ref.Concept == "concept:door" && ref.Multiplicity != "" {
				t.Errorf("PL plural-only door inferred count %q", ref.Multiplicity)
			}
			if tc.id == "a1.10.time.045.en" && ref.Concept == "concept:door" && ref.Multiplicity != meaning.MultiplicityOne {
				t.Errorf("EN singular door multiplicity=%q", ref.Multiplicity)
			}
			if tc.id != "a1.10.time.045.pl" && tc.id != "a1.10.time.045.en" && (ref.Concept == "concept:child" || ref.Concept == "concept:child_of_person") {
				if ref.Concept != meaning.ConceptRef("concept:"+tc.concept) {
					t.Errorf("%s child concept=%q, want %q", tc.id, ref.Concept, tc.concept)
				}
			}
		}
	}
	for index := 49; index <= 56; index++ {
		for _, language := range []string{"en", "pl"} {
			id := fmtTimeID(index, language)
			if _, ok, err := interlingua.TimeExpansion(id); err != nil || ok {
				t.Errorf("gap %s produced graph: ok=%t error=%v", id, ok, err)
			}
		}
	}
	for _, id := range []string{"a1.10.time.025.en.age", "a1.10.time.009.pl.past", "a1.10.time.033.en.past", "a1.10.time.045.pl.age", "a1.10.time.022.en.past"} {
		if _, ok, err := interlingua.TimeExpansion(id); err != nil || ok {
			t.Errorf("unlicensed variant %s: ok=%t error=%v", id, ok, err)
		}
	}
}

func TestTimeExpansionEnumeratedVariantsAreUniqueAndReadIsAmbiguous(t *testing.T) {
	seenBySource := map[string]map[string]string{}
	ids := interlingua.TimeExpansionIDs()
	listed := map[string]bool{}
	for _, id := range ids {
		listed[id] = true
		fixture, ok, err := interlingua.TimeExpansion(id)
		if err != nil || !ok {
			t.Fatalf("TimeExpansion(%q): ok=%t error=%v", id, ok, err)
		}
		hash, err := meaning.CanonicalHash(fixture.Graph, meaning.DefaultBudget())
		if err != nil {
			t.Fatalf("hash %s: %v", id, err)
		}
		parts := strings.Split(id, ".")
		source := strings.Join(parts[:5], ".")
		if seenBySource[source] == nil {
			seenBySource[source] = map[string]string{}
		}
		if previous, exists := seenBySource[source][hash]; exists {
			t.Errorf("%s and %s enumerate duplicate graphs for source %s", previous, id, source)
		}
		seenBySource[source][hash] = id
	}
	for _, index := range []int{9, 10, 11, 12, 22, 34} {
		base := fmtTimeID(index, "en")
		alternate := base + ".past"
		if index == 22 {
			alternate = base + ".present"
		}
		if !listed[base] || !listed[alternate] {
			t.Errorf("English bare read row %d lacks both licensed frame IDs: %q and %q", index, base, alternate)
		}
		if index == 22 && listed[base+".past"] {
			t.Errorf("row22 English past primary was redundantly enumerated")
		}
	}
}

func hasSuffix(value, suffix string) bool {
	return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
}
func fmtTimeID(index int, language string) string {
	digits := "000" + string(rune('0'+index/10)) + string(rune('0'+index%10))
	return "a1.10.time." + digits[len(digits)-3:] + "." + language
}
