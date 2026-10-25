package a1en14_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en14 "github.com/clickbytes/golingua/tests/languages/en/A1.4"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestActionExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en14.ActionAnalysisCases)
}
func TestActionExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en14.ActionRealizationCases)
}
