package a1pl11_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1pl11 "github.com/clickbytes/golingua/tests/languages/pl/A1.1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCorpusAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl11.CorpusAnalysisCases)
}

func TestCorpusRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl11.CorpusRealizationCases)
}
