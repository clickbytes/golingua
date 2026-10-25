package a1pl13_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1pl13 "github.com/clickbytes/golingua/tests/languages/pl/A1.3"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCompositionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl13.CompositionAnalysisCases)
}

func TestCompositionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl13.CompositionRealizationCases)
}
