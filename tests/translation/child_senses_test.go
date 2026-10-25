package translation_test

import (
	"fmt"
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en112 "github.com/clickbytes/golingua/tests/languages/en/A1.12"
	a1en18 "github.com/clickbytes/golingua/tests/languages/en/A1.8"
	a1pl112 "github.com/clickbytes/golingua/tests/languages/pl/A1.12"
	a1pl18 "github.com/clickbytes/golingua/tests/languages/pl/A1.8"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestChildAgeAndOffspringRoutes(t *testing.T) {
	_ = matrix.Runtime(t)

	ambiguityIDs := []string{
		"A1.12-AMB-1229-EN-PL:analyze",
		"A1.12-AMB-1231-EN-PL:analyze",
		"A1.12-AMB-1233-EN-PL:analyze",
	}
	negationENID := "A1.8-NEG-823-EN-PL:analyze"
	negationPLID := "A1.8-NEG-824-PL-EN:analyze"
	enAnalyses := pickChildAnalyses(t, "R6-child-source-en", a1en112.AmbiguityAnalysisCases, ambiguityIDs...)
	enAnalyses = append(enAnalyses, pickChildAnalyses(t, "R6-child-negative-source-en", a1en18.NegationAnalysisCases, negationENID)...)
	plAnalyses := pickChildAnalyses(t, "R6-child-source-pl", a1pl112.AmbiguityAnalysisCases,
		"A1.12-AMB-1230-PL-EN:analyze",
		"A1.12-AMB-1232-PL-EN:analyze",
		"A1.12-AMB-1234-PL-EN:analyze",
	)
	plAnalyses = append(plAnalyses, pickChildAnalyses(t, "R6-child-negative-source-pl", a1pl18.NegationAnalysisCases, negationPLID)...)
	matrix.RunAnalysis(t, features.LanguageID("en"), enAnalyses)
	matrix.RunAnalysis(t, features.LanguageID("pl"), plAnalyses)

	// These authored targets cover singular and plural child referents for both
	// meanings, including English invariant offspring, plus negated object use.
	plTargets := pickChildRealizations(t, "R6-child-target-pl", a1pl112.AmbiguityRealizationCases,
		"A1.12-AMB-1229-EN-PL:realize",
		"A1.12-AMB-1231-EN-PL:realize",
		"A1.12-AMB-1233-EN-PL:realize",
	)
	plTargets = append(plTargets, pickChildRealizations(t, "R6-child-negative-target-pl", a1pl18.NegationRealizationCases, "A1.8-NEG-823-EN-PL:realize")...)
	enTargets := pickChildRealizations(t, "R6-child-target-en", a1en112.AmbiguityRealizationCases,
		"A1.12-AMB-1230-PL-EN:realize",
		"A1.12-AMB-1232-PL-EN:realize",
		"A1.12-AMB-1234-PL-EN:realize",
	)
	enTargets = append(enTargets, pickChildRealizations(t, "R6-child-negative-target-en", a1en18.NegationRealizationCases, "A1.8-NEG-824-PL-EN:realize")...)
	matrix.RunRealization(t, features.LanguageID("pl"), plTargets)
	matrix.RunRealization(t, features.LanguageID("en"), enTargets)
}

func pickChildAnalyses(t *testing.T, idPrefix string, cases []matrix.AnalysisCase, ids ...string) []matrix.AnalysisCase {
	t.Helper()
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	var selected []matrix.AnalysisCase
	for _, testCase := range cases {
		if wanted[testCase.ID] {
			delete(wanted, testCase.ID)
			testCase.ID = fmt.Sprintf("%s-%02d", idPrefix, len(selected)+1)
			selected = append(selected, testCase)
		}
	}
	for id := range wanted {
		t.Errorf("authored child analysis case %q was not found", id)
	}
	return selected
}

func pickChildRealizations(t *testing.T, idPrefix string, cases []matrix.RealizationCase, ids ...string) []matrix.RealizationCase {
	t.Helper()
	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	var selected []matrix.RealizationCase
	for _, testCase := range cases {
		if wanted[testCase.ID] {
			delete(wanted, testCase.ID)
			testCase.ID = fmt.Sprintf("%s-%02d", idPrefix, len(selected)+1)
			selected = append(selected, testCase)
		}
	}
	for id := range wanted {
		t.Errorf("authored child realization case %q was not found", id)
	}
	return selected
}
