package a1pl15_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1pl15 "github.com/clickbytes/golingua/tests/languages/pl/A1.5"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestIdentificationExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl15.IdentificationAnalysisCases)
}
func TestIdentificationExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl15.IdentificationRealizationCases)
}
