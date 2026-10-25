package a1en110_test

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/tests/languages/en/A1.10"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestTimeOtherAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en110.TimeOtherAnalysisCases)
}
func TestTimeOtherRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en110.TimeOtherRealizationCases)
}
