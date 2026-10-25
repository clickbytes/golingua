package a1en18_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1en18 "github.com/clickbytes/golingua/tests/languages/en/A1.8"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestNegationExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en18.NegationAnalysisCases)
}
func TestNegationExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en18.NegationRealizationCases)
}
