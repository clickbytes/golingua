package a1en11_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en11 "github.com/clickbytes/golingua/tests/languages/en/A1.1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestCorpusAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en11.CorpusAnalysisCases)
}

func TestCorpusRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en11.CorpusRealizationCases)
}
