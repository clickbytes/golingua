package a1pl12_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1pl12 "github.com/clickbytes/golingua/tests/languages/pl/A1.2"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestContrastAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl12.ContrastAnalysisCases)
}

func TestContrastRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl12.ContrastRealizationCases)
}
