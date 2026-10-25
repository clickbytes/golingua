package translation

import (
	"strings"
	"testing"

	a1en "github.com/clickbytes/golingua/tests/languages/en/A1"
	a1en11 "github.com/clickbytes/golingua/tests/languages/en/A1.1"
	a1pl "github.com/clickbytes/golingua/tests/languages/pl/A1"
	a1pl11 "github.com/clickbytes/golingua/tests/languages/pl/A1.1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestInventory(t *testing.T) {
	if len(Cases) == 0 {
		t.Fatal("declarative language matrix inventory is empty")
	}
	analysisLinks := make(map[string]int)
	realizationLinks := make(map[string]int)
	integrationIDs := make(map[string]bool, len(Cases))
	for _, c := range Cases {
		t.Logf("inventory ID=%s direction=%s-%s coverage_declared=%s", c.ID, c.From, c.To, strings.Join(c.CoverageTags, ","))
		if c.ID == "" || !validCaseID(c.ID, c.From, c.To) {
			t.Errorf("invalid integration ID/direction: %+v", c)
		}
		if strings.TrimSuffix(c.AnalysisID, ":analyze") != c.ID || strings.TrimSuffix(c.RealizationID, ":realize") != c.ID {
			t.Errorf("case %q links analysis/realization from a different record", c.ID)
		}
		if integrationIDs[c.ID] {
			t.Errorf("duplicate integration ID %q", c.ID)
		}
		integrationIDs[c.ID] = true
		analysis, realization, err := Resolve(c)
		if err != nil {
			t.Errorf("case %q reference resolution: %v", c.ID, err)
			continue
		}
		analysisLinks[analysis.ID]++
		realizationLinks[realization.ID]++
		if !strings.HasSuffix(analysis.ID, ":analyze") || !strings.HasSuffix(realization.ID, ":realize") {
			t.Errorf("case %q has malformed analysis/realization links", c.ID)
		}
	}
	checkCases(t, "pl", a1pl.AnalysisCases, a1pl.RealizationCases, analysisLinks, realizationLinks)
	checkCases(t, "en", a1en.AnalysisCases, a1en.RealizationCases, analysisLinks, realizationLinks)
	checkCases(t, "pl-A1.1", a1pl11.CorpusAnalysisCases, a1pl11.CorpusRealizationCases, analysisLinks, realizationLinks)
	checkCases(t, "en-A1.1", a1en11.CorpusAnalysisCases, a1en11.CorpusRealizationCases, analysisLinks, realizationLinks)
}

func validCaseID(id, from, to string) bool {
	if !((from == "en" && to == "pl") || (from == "pl" && to == "en")) {
		return false
	}
	parts := strings.Split(id, "-")
	if len(parts) != 5 || len(parts[0]) < 2 || !strings.ContainsRune("ABC", rune(parts[0][0])) || len(parts[1]) == 0 || parts[3] != strings.ToUpper(from) || parts[4] != strings.ToUpper(to) {
		return false
	}
	for _, level := range strings.Split(parts[0][1:], ".") {
		if !digits(level) {
			return false
		}
	}
	for i, r := range parts[1] {
		if (i == 0 && (r < 'A' || r > 'Z')) || (i > 0 && !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))) {
			return false
		}
	}
	return digits(parts[2])
}

func digits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func TestInventoryIdentifierSupportsBatches(t *testing.T) {
	for _, test := range []struct {
		id   string
		from string
		to   string
		want bool
	}{
		{id: "A1-AC-001-EN-PL", from: "en", to: "pl", want: true},
		{id: "A1.1-NEG-012-EN-PL", from: "en", to: "pl", want: true},
		{id: "B1-NEG-012-PL-EN", from: "pl", to: "en", want: true},
		{id: "C2.1-Q-009-EN-PL", from: "en", to: "pl", want: true},
		{id: "A1..1-NEG-012-EN-PL", from: "en", to: "pl"},
		{id: "A1.1-NEG-012-EN-PL", from: "pl", to: "en"},
	} {
		if got := validCaseID(test.id, test.from, test.to); got != test.want {
			t.Errorf("validCaseID(%q, %q, %q) = %t, want %t", test.id, test.from, test.to, got, test.want)
		}
	}
}

func checkCases(t *testing.T, lang string, analyses []matrix.AnalysisCase, realizations []matrix.RealizationCase, analysisLinks, realizationLinks map[string]int) {
	t.Helper()
	analysisIDs := make(map[string]bool, len(analyses))
	realizationIDs := make(map[string]bool, len(realizations))
	for _, c := range analyses {
		if c.ID == "" || analysisIDs[c.ID] {
			t.Errorf("%s analysis ID is empty or duplicated: %q", lang, c.ID)
		}
		analysisIDs[c.ID] = true
		if analysisLinks[c.ID] != 1 {
			t.Errorf("%s analysis %q is linked %d times; want exactly once", lang, c.ID, analysisLinks[c.ID])
		}
		if c.Input == "" || len(c.MustPreserve) == 0 || c.InputOrigin == "" || c.OwnerAcceptance == "" {
			t.Errorf("%s analysis %q lacks input, requirements, provenance, or owner status", lang, c.ID)
		}
		allowedMeanings := map[string]bool{c.MeaningID: true}
		for _, meaningID := range c.AllowedMeaningIDs {
			if meaningID == "" || allowedMeanings[meaningID] {
				t.Errorf("%s analysis %q has an empty or duplicate allowed MeaningID %q", lang, c.ID, meaningID)
			}
			allowedMeanings[meaningID] = true
		}
		requiredMeanings := map[string]bool{c.MeaningID: true}
		for _, meaningID := range c.RequiredMeaningIDs {
			if meaningID == "" || requiredMeanings[meaningID] || !allowedMeanings[meaningID] {
				t.Errorf("%s analysis %q has an empty, duplicate, or disallowed required MeaningID %q", lang, c.ID, meaningID)
			}
			requiredMeanings[meaningID] = true
		}
	}
	for _, c := range realizations {
		if c.ID == "" || realizationIDs[c.ID] {
			t.Errorf("%s realization ID is empty or duplicated: %q", lang, c.ID)
		}
		realizationIDs[c.ID] = true
		if realizationLinks[c.ID] != 1 {
			t.Errorf("%s realization %q is linked %d times; want exactly once", lang, c.ID, realizationLinks[c.ID])
		}
		if len(c.ProposalTexts) == 0 || c.ReferenceOrigin == "" || c.AmbiguityNote == "" || c.OwnerAcceptance == "" {
			t.Errorf("%s realization %q lacks proposal, provenance, or owner status", lang, c.ID)
		}
		allowed := make(map[string]bool, len(c.AllowedTexts))
		for _, text := range c.AllowedTexts {
			allowed[text] = true
		}
		for _, rule := range c.RequiredTextAssumptions {
			if !allowed[rule.Text] || rule.Field == "" || rule.Value == "" {
				t.Errorf("%s realization %q has invalid per-text assumption rule: %+v", lang, c.ID, rule)
			}
		}
		variantIDs := map[string]bool{c.MeaningID: true}
		for _, variant := range c.MeaningVariants {
			if variant.MeaningID == "" || variantIDs[variant.MeaningID] || len(variant.AllowedTexts) == 0 {
				t.Errorf("%s realization %q has an empty/duplicate Meaning variant or no allowed texts: %+v", lang, c.ID, variant)
			}
			variantIDs[variant.MeaningID] = true
			variantAllowed := make(map[string]bool, len(variant.AllowedTexts))
			for _, text := range variant.AllowedTexts {
				variantAllowed[text] = true
			}
			for _, rule := range variant.RequiredTextAssumptions {
				if !variantAllowed[rule.Text] || rule.Field == "" || rule.Value == "" {
					t.Errorf("%s realization %q variant %q has invalid assumption rule: %+v", lang, c.ID, variant.MeaningID, rule)
				}
			}
			variantRequired := make(map[string]bool, len(variant.RequiredTexts))
			for _, text := range variant.RequiredTexts {
				if variantRequired[text] || !variantAllowed[text] {
					t.Errorf("%s realization %q variant %q requires a repeated or disallowed text %q", lang, c.ID, variant.MeaningID, text)
				}
				variantRequired[text] = true
			}
		}
		required := make(map[string]bool, len(c.RequiredTexts))
		for _, text := range c.RequiredTexts {
			if required[text] {
				t.Errorf("%s realization %q repeats required text %q", lang, c.ID, text)
			}
			if !allowed[text] {
				t.Errorf("%s realization %q requires text that is not allowed: %q", lang, c.ID, text)
			}
			required[text] = true
		}
	}
}
