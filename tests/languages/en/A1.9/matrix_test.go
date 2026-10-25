package a1en19_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1en19 "github.com/clickbytes/golingua/tests/languages/en/A1.9"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestQuestionExpansionAnalysis(t *testing.T) {
	matrix.RunAnalysis(t, features.LanguageID("en"), a1en19.QuestionAnalysisCases)
}
func TestQuestionExpansionRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), a1en19.QuestionRealizationCases)
}
