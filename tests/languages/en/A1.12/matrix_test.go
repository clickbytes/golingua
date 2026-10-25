package a1en112_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1 "github.com/clickbytes/golingua/tests/languages/en/A1.12"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestAmbiguityAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1.AmbiguityAnalysisCases)
}
func TestAmbiguityRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1.AmbiguityRealizationCases)
}
