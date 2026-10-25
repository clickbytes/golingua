package a1en16_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1 "github.com/clickbytes/golingua/tests/languages/en/A1.6"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestPossessionExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1.PossessionAnalysisCases)
}
func TestPossessionExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1.PossessionRealizationCases)
}
