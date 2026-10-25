package a1pl110_test

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/tests/languages/pl/A1.10"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestTimePersonAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl110.TimePersonAnalysisCases)
}
func TestTimePersonRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl110.TimePersonRealizationCases)
}
