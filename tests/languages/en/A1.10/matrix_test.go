package a1en110_test

import (
	"testing"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/tests/matrix"
	"github.com/clickbytes/golingua/tests/languages/en/A1.10"
)

func TestTimePersonAnalysis(t *testing.T) { matrix.RunAnalysis(t, features.LanguageID("en"), a1en110.TimePersonAnalysisCases) }
func TestTimePersonRealization(t *testing.T) { matrix.RunRealization(t, features.LanguageID("en"), a1en110.TimePersonRealizationCases) }
