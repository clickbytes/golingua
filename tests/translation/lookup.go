package translation

import (
	"fmt"

	"github.com/clickbytes/golingua/tests/languages/en/A1"
	a1en11 "github.com/clickbytes/golingua/tests/languages/en/A1.1"
	a1en110 "github.com/clickbytes/golingua/tests/languages/en/A1.10"
	a1en12 "github.com/clickbytes/golingua/tests/languages/en/A1.2"
	a1en13 "github.com/clickbytes/golingua/tests/languages/en/A1.3"
	a1en14 "github.com/clickbytes/golingua/tests/languages/en/A1.4"
	a1en15 "github.com/clickbytes/golingua/tests/languages/en/A1.5"
	a1en16 "github.com/clickbytes/golingua/tests/languages/en/A1.6"
	a1en17 "github.com/clickbytes/golingua/tests/languages/en/A1.7"
	a1en18 "github.com/clickbytes/golingua/tests/languages/en/A1.8"
	a1en19 "github.com/clickbytes/golingua/tests/languages/en/A1.9"
	"github.com/clickbytes/golingua/tests/languages/pl/A1"
	a1pl11 "github.com/clickbytes/golingua/tests/languages/pl/A1.1"
	a1pl110 "github.com/clickbytes/golingua/tests/languages/pl/A1.10"
	a1pl12 "github.com/clickbytes/golingua/tests/languages/pl/A1.2"
	a1pl13 "github.com/clickbytes/golingua/tests/languages/pl/A1.3"
	a1pl14 "github.com/clickbytes/golingua/tests/languages/pl/A1.4"
	a1pl15 "github.com/clickbytes/golingua/tests/languages/pl/A1.5"
	a1pl16 "github.com/clickbytes/golingua/tests/languages/pl/A1.6"
	a1pl17 "github.com/clickbytes/golingua/tests/languages/pl/A1.7"
	a1pl18 "github.com/clickbytes/golingua/tests/languages/pl/A1.8"
	a1pl19 "github.com/clickbytes/golingua/tests/languages/pl/A1.9"
	"github.com/clickbytes/golingua/tests/matrix"
)

type LanguageBatch struct {
	ID                                      string
	EnglishAnalysis, PolishAnalysis         []matrix.AnalysisCase
	EnglishRealizations, PolishRealizations []matrix.RealizationCase
}

// LanguageBatches is the declarative registry used for lookup and inventory.
var LanguageBatches = []LanguageBatch{
	{ID: "A1", EnglishAnalysis: a1en.AnalysisCases, PolishAnalysis: a1pl.AnalysisCases, EnglishRealizations: a1en.RealizationCases, PolishRealizations: a1pl.RealizationCases},
	{ID: "A1.1", EnglishAnalysis: a1en11.CorpusAnalysisCases, PolishAnalysis: a1pl11.CorpusAnalysisCases, EnglishRealizations: a1en11.CorpusRealizationCases, PolishRealizations: a1pl11.CorpusRealizationCases},
	{ID: "A1.2", EnglishAnalysis: a1en12.ContrastAnalysisCases, PolishAnalysis: a1pl12.ContrastAnalysisCases, EnglishRealizations: a1en12.ContrastRealizationCases, PolishRealizations: a1pl12.ContrastRealizationCases},
	{ID: "A1.3", EnglishAnalysis: a1en13.CompositionAnalysisCases, PolishAnalysis: a1pl13.CompositionAnalysisCases, EnglishRealizations: a1en13.CompositionRealizationCases, PolishRealizations: a1pl13.CompositionRealizationCases},
	{ID: "A1.4", EnglishAnalysis: a1en14.ActionAnalysisCases, PolishAnalysis: a1pl14.ActionAnalysisCases, EnglishRealizations: a1en14.ActionRealizationCases, PolishRealizations: a1pl14.ActionRealizationCases},
	{ID: "A1.5", EnglishAnalysis: a1en15.IdentificationAnalysisCases, PolishAnalysis: a1pl15.IdentificationAnalysisCases, EnglishRealizations: a1en15.IdentificationRealizationCases, PolishRealizations: a1pl15.IdentificationRealizationCases},
	{ID: "A1.6", EnglishAnalysis: a1en16.PossessionAnalysisCases, PolishAnalysis: a1pl16.PossessionAnalysisCases, EnglishRealizations: a1en16.PossessionRealizationCases, PolishRealizations: a1pl16.PossessionRealizationCases},
	{ID: "A1.7", EnglishAnalysis: a1en17.LocationQuantityAnalysisCases, PolishAnalysis: a1pl17.LocationQuantityAnalysisCases, EnglishRealizations: a1en17.LocationQuantityRealizationCases, PolishRealizations: a1pl17.LocationQuantityRealizationCases},
	{ID: "A1.8", EnglishAnalysis: a1en18.NegationAnalysisCases, PolishAnalysis: a1pl18.NegationAnalysisCases, EnglishRealizations: a1en18.NegationRealizationCases, PolishRealizations: a1pl18.NegationRealizationCases},
	{ID: "A1.9", EnglishAnalysis: a1en19.QuestionAnalysisCases, PolishAnalysis: a1pl19.QuestionAnalysisCases, EnglishRealizations: a1en19.QuestionRealizationCases, PolishRealizations: a1pl19.QuestionRealizationCases},
	{ID: "A1.10-person", EnglishAnalysis: a1en110.TimePersonAnalysisCases, PolishAnalysis: a1pl110.TimePersonAnalysisCases, EnglishRealizations: a1en110.TimePersonRealizationCases, PolishRealizations: a1pl110.TimePersonRealizationCases},
	{ID: "A1.10-other", EnglishAnalysis: a1en110.TimeOtherAnalysisCases, PolishAnalysis: a1pl110.TimeOtherAnalysisCases, EnglishRealizations: a1en110.TimeOtherRealizationCases, PolishRealizations: a1pl110.TimeOtherRealizationCases},
}

func AnalysisCasesFor(language string) []matrix.AnalysisCase {
	var cases []matrix.AnalysisCase
	for _, batch := range LanguageBatches {
		if language == "en" {
			cases = append(cases, batch.EnglishAnalysis...)
		} else if language == "pl" {
			cases = append(cases, batch.PolishAnalysis...)
		}
	}
	return cases
}
func RealizationCasesFor(language string) []matrix.RealizationCase {
	var cases []matrix.RealizationCase
	for _, batch := range LanguageBatches {
		if language == "en" {
			cases = append(cases, batch.EnglishRealizations...)
		} else if language == "pl" {
			cases = append(cases, batch.PolishRealizations...)
		}
	}
	return cases
}

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
	if !((c.From == "en" && c.To == "pl") || (c.From == "pl" && c.To == "en")) {
		return matrix.AnalysisCase{}, matrix.RealizationCase{}, fmt.Errorf("unsupported translation direction %q to %q", c.From, c.To)
	}
	analyses := AnalysisCasesFor(c.From)
	realizations := RealizationCasesFor(c.To)
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
