package a1en_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en "github.com/clickbytes/golingua/tests/languages/en/A1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestSeeingAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en.ActionAnalysisCases)
}

func TestSeeingRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en.ActionRealizationCases)
}

func TestIdentificationAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en.IdentificationAnalysisCases)
}

func TestIdentificationRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en.IdentificationRealizationCases)
}

func TestOtherAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), append(append(append(append([]matrix.AnalysisCase{}, a1en.LocationQuantityAnalysisCases...), a1en.NegationAnalysisCases...), a1en.PossessionAnalysisCases...), a1en.QuestionAnalysisCases...))
}

func TestOtherRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), append(append(append(append([]matrix.RealizationCase{}, a1en.LocationQuantityRealizationCases...), a1en.NegationRealizationCases...), a1en.PossessionRealizationCases...), a1en.QuestionRealizationCases...))
}
