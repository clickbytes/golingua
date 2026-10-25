package translation

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestAmbiguityCatalogueLinks(t *testing.T) {
	catalogue := make(map[string]map[string]bool)
	for _, id := range interlingua.AmbiguityExpansionIDs() {
		fixture, ok, err := interlingua.AmbiguityExpansion(id)
		if err != nil || !ok {
			t.Fatalf("catalogue ID %q: ok=%t err=%v", id, ok, err)
		}
		parts := strings.Split(id, ".")
		row, err := strconv.Atoi(parts[3])
		if err != nil {
			t.Fatalf("catalogue ID %q: %v", id, err)
		}
		key := fmt.Sprintf("%03d.%s", row, parts[4])
		if catalogue[key] == nil {
			catalogue[key] = map[string]bool{}
		}
		if catalogue[key][id] {
			t.Errorf("duplicate canonical catalogue ID %q", id)
		}
		catalogue[key][id] = true
		_ = fixture
	}
	for _, lang := range []string{"en", "pl"} {
		for row := 1; row <= 43; row++ {
			idTag := "PL-EN"
			if lang == "en" {
				idTag = "EN-PL"
			}
			num := 1200 + 2*row - 1
			if lang == "pl" {
				num++
			}
			caseID := fmt.Sprintf("A1.12-AMB-%04d-%s:analyze", num, idTag)
			var found *matrix.AnalysisCase
			for _, c := range AnalysisCasesFor(lang) {
				if c.ID == caseID {
					if found != nil {
						t.Fatalf("duplicate analysis %s", caseID)
					}
					copy := c
					found = &copy
				}
			}
			if found == nil {
				t.Fatalf("missing analysis %s", caseID)
			}
			primary := fmt.Sprintf("a1.12.amb.%03d.%s", row, lang)
			key := fmt.Sprintf("%03d.%s", row, lang)
			expected := catalogue[key]
			if len(expected) == 0 {
				if len(found.AllowedMeaningIDs) != 0 || len(found.RequiredMeaningIDs) != 0 {
					t.Errorf("gap %s lists fake I variants: allowed=%v required=%v", caseID, found.AllowedMeaningIDs, found.RequiredMeaningIDs)
				}
				if _, ok, err := interlingua.AmbiguityExpansion(primary); err != nil || ok {
					t.Errorf("gap %s unexpectedly has a primary fixture", caseID)
				}
				continue
			}
			if _, ok, err := interlingua.AmbiguityExpansion(primary); err != nil || !ok {
				t.Errorf("%s primary fixture absent: %v", caseID, err)
			}
			allowed := map[string]bool{primary: true}
			for _, id := range found.AllowedMeaningIDs {
				if allowed[id] {
					t.Errorf("%s duplicate/primary in allowed IDs: %s", caseID, id)
				}
				allowed[id] = true
			}
			required := map[string]bool{primary: true}
			for _, id := range found.RequiredMeaningIDs {
				if required[id] {
					t.Errorf("%s duplicate/primary in required IDs: %s", caseID, id)
				}
				required[id] = true
			}
			for id := range expected {
				if !allowed[id] {
					t.Errorf("%s catalogue meaning missing from allowed set: %s", caseID, id)
				}
				if !required[id] {
					t.Errorf("%s catalogue meaning missing from required set: %s", caseID, id)
				}
			}
			for id := range allowed {
				if !expected[id] {
					t.Errorf("%s has undeclared/noncanonical allowed meaning %s", caseID, id)
				}
			}
			for id := range required {
				if !expected[id] {
					t.Errorf("%s has undeclared/noncanonical required meaning %s", caseID, id)
				}
			}
		}
	}
}
