package a1pl_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1pl "github.com/clickbytes/golingua/tests/languages/pl/A1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestSeeingAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl.ActionAnalysisCases)
}

func TestSeeingRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl.ActionRealizationCases)
}

func TestIdentificationAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl.IdentificationAnalysisCases)
}

func TestIdentificationRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl.IdentificationRealizationCases)
}

func TestNegationAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl.NegationAnalysisCases)
}

func TestNegationRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl.NegationRealizationCases)
	matrix.RunRealization(t, features.LanguageID("pl"), []matrix.RealizationCase{{
		ID: "I-PL-negative-classification-to", MeaningID: "negation.classification.to",
		ProposalTexts: []string{"To nie jest kot."}, AllowedTexts: []string{"To nie jest kot."}, RequiredTexts: []string{"To nie jest kot."},
		AllowedAssumptionFields: []string{"tense", "person", "number"},
		ReferenceOrigin:         "authored interlingua fixture", OwnerAcceptance: "NOT_EVALUATED",
	}})
}

func TestOtherAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), append(append(append([]matrix.AnalysisCase{}, a1pl.LocationQuantityAnalysisCases...), a1pl.PossessionAnalysisCases...), a1pl.QuestionAnalysisCases...))
}

func TestOtherRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), append(append(append([]matrix.RealizationCase{}, a1pl.LocationQuantityRealizationCases...), a1pl.PossessionRealizationCases...), a1pl.QuestionRealizationCases...))
}
