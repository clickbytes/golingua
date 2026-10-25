package a1pl19_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1pl19 "github.com/clickbytes/golingua/tests/languages/pl/A1.9"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestQuestionExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("pl"), a1pl19.QuestionAnalysisCases)
}
func TestQuestionExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), a1pl19.QuestionRealizationCases)
}
