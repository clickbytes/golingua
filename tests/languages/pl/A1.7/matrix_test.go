package a1pl17_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1 "github.com/clickbytes/golingua/tests/languages/pl/A1.7"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestLocationQuantityExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1.LocationQuantityAnalysisCases)
}
func TestLocationQuantityExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1.LocationQuantityRealizationCases)
}
