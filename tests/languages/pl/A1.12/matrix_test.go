package a1pl112_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1 "github.com/clickbytes/golingua/tests/languages/pl/A1.12"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestAmbiguityAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1.AmbiguityAnalysisCases)
}
func TestAmbiguityRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1.AmbiguityRealizationCases)
}
