package a1en113_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1 "github.com/clickbytes/golingua/tests/languages/en/A1.13"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestCompositionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1.CompositionAnalysisCases)
}
func TestCompositionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1.CompositionRealizationCases)
}
