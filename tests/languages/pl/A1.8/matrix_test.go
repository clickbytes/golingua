package a1pl18_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1pl18 "github.com/clickbytes/golingua/tests/languages/pl/A1.8"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestNegationExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl18.NegationAnalysisCases)
}
func TestNegationExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl18.NegationRealizationCases)
}
