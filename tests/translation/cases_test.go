package translation

import (
	"fmt"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/tests/languages/en/A1"
	"github.com/clickbytes/golingua/tests/languages/pl/A1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestInventory(t *testing.T) {
	expected := expectedIDs()
	counts := make(map[string]int, len(Cases))
	analysisLinks := make(map[string]int)
	realizationLinks := make(map[string]int)
	for _, c := range Cases {
		counts[c.ID]++
		t.Logf("inventory ID=%s", c.ID)
		if _, exists := expected[c.ID]; !exists {
			t.Errorf("unexpected integration ID %q", c.ID)
		}
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
	for id := range expected {
		if counts[id] != 1 {
			t.Errorf("expected integration ID %q exactly once, got %d", id, counts[id])
		}
	}
	checkCases(t, "pl", a1pl.AnalysisCases, a1pl.RealizationCases, analysisLinks, realizationLinks)
	checkCases(t, "en", a1en.AnalysisCases, a1en.RealizationCases, analysisLinks, realizationLinks)
}

func checkCases(t *testing.T, lang string, analyses []matrixAnalysisCase, realizations []matrixRealizationCase, analysisLinks, realizationLinks map[string]int) {
	t.Helper()
	if len(analyses) != 12 {
		t.Errorf("%s analysis inventory has %d cases; want 12", lang, len(analyses))
	}
	if len(realizations) != 12 {
		t.Errorf("%s realization inventory has %d cases; want 12", lang, len(realizations))
	}
	for _, c := range analyses {
		if analysisLinks[c.ID] != 1 {
			t.Errorf("%s analysis %q is linked %d times; want exactly once", lang, c.ID, analysisLinks[c.ID])
		}
		if c.Input == "" || len(c.MustPreserve) == 0 || c.InputOrigin == "" || c.OwnerAcceptance == "" {
			t.Errorf("%s analysis %q lacks input, requirements, provenance, or owner status", lang, c.ID)
		}
	}
	for _, c := range realizations {
		if realizationLinks[c.ID] != 1 {
			t.Errorf("%s realization %q is linked %d times; want exactly once", lang, c.ID, realizationLinks[c.ID])
		}
		if len(c.ProposalTexts) == 0 || c.ReferenceOrigin == "" || c.AmbiguityNote == "" || c.OwnerAcceptance == "" {
			t.Errorf("%s realization %q lacks proposal, provenance, or owner status", lang, c.ID)
		}
	}
}

type matrixAnalysisCase = matrix.AnalysisCase
type matrixRealizationCase = matrix.RealizationCase

func expectedIDs() map[string]struct{} {
	result := make(map[string]struct{}, 24)
	for _, domain := range []string{"AC", "ID", "LQ", "NEG", "PO", "Q"} {
		for n := 1; n <= 4; n++ {
			direction := "PL-EN"
			if n%2 == 1 {
				direction = "EN-PL"
			}
			result[fmt.Sprintf("A1-%s-%03d-%s", domain, n, direction)] = struct{}{}
		}
	}
	return result
}
