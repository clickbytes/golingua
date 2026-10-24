package translation

import (
	"fmt"

	"github.com/clickbytes/golingua/tests/languages/en/A1"
	"github.com/clickbytes/golingua/tests/languages/pl/A1"
	"github.com/clickbytes/golingua/tests/matrix"
)

// CasesFor returns the registered integration cases for one exact direction.
func CasesFor(from, to string) []matrix.TranslationCase {
	result := make([]matrix.TranslationCase, 0)
	for _, c := range Cases {
		if c.From == from && c.To == to {
			result = append(result, c)
		}
	}
	return result
}

// Resolve finds the exact source analysis and target realization linked by c.
func Resolve(c matrix.TranslationCase) (matrix.AnalysisCase, matrix.RealizationCase, error) {
	var analyses []matrix.AnalysisCase
	var realizations []matrix.RealizationCase
	switch {
	case c.From == "en" && c.To == "pl":
		analyses, realizations = a1en.AnalysisCases, a1pl.RealizationCases
	case c.From == "pl" && c.To == "en":
		analyses, realizations = a1pl.AnalysisCases, a1en.RealizationCases
	default:
		return matrix.AnalysisCase{}, matrix.RealizationCase{}, fmt.Errorf("unsupported translation direction %q to %q", c.From, c.To)
	}
	analysis, analysisCount := findAnalysis(analyses, c.AnalysisID)
	realization, realizationCount := findRealization(realizations, c.RealizationID)
	if analysisCount != 1 || realizationCount != 1 {
		return matrix.AnalysisCase{}, matrix.RealizationCase{}, fmt.Errorf("case %q resolves analysis %d times and realization %d times", c.ID, analysisCount, realizationCount)
	}
	return analysis, realization, nil
}

func findAnalysis(cases []matrix.AnalysisCase, id string) (matrix.AnalysisCase, int) {
	var found matrix.AnalysisCase
	count := 0
	for _, c := range cases {
		if c.ID == id {
			found, count = c, count+1
		}
	}
	return found, count
}

func findRealization(cases []matrix.RealizationCase, id string) (matrix.RealizationCase, int) {
	var found matrix.RealizationCase
	count := 0
	for _, c := range cases {
		if c.ID == id {
			found, count = c, count+1
		}
	}
	return found, count
}
