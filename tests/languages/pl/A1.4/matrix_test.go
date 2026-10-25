package a1pl14_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1pl14 "github.com/clickbytes/golingua/tests/languages/pl/A1.4"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestActionExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl14.ActionAnalysisCases)
}
func TestActionExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl14.ActionRealizationCases)
}
