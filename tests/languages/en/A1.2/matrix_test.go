package a1en12_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en12 "github.com/clickbytes/golingua/tests/languages/en/A1.2"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestContrastAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en12.ContrastAnalysisCases)
}

func TestContrastRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en12.ContrastRealizationCases)
}
