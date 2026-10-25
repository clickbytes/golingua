package a1en11_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1en11 "github.com/clickbytes/golingua/tests/languages/en/A1.1"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestRefusalAnalysis(t *testing.T) {
	matrix.RunRefusalAnalysis(t, features.LanguageID("en"), a1en11.RefusalCases)
}
