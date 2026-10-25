package a1en15_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en15 "github.com/clickbytes/golingua/tests/languages/en/A1.5"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestIdentificationExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en15.IdentificationAnalysisCases)
}
func TestIdentificationExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en15.IdentificationRealizationCases)
}
