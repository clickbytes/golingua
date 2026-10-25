package a1en13_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en13 "github.com/clickbytes/golingua/tests/languages/en/A1.3"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en13.CompositionAnalysisCases)
}

func TestCompositionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en13.CompositionRealizationCases)
}
