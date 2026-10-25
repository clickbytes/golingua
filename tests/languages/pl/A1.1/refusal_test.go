package a1pl11_test

import (
	"github.com/clickbytes/golingua/internal/features"
	a1pl11 "github.com/clickbytes/golingua/tests/languages/pl/A1.1"
	"github.com/clickbytes/golingua/tests/matrix"
	"testing"
)

func TestRefusalAnalysis(t *testing.T) {
	matrix.RunRefusalAnalysis(t, features.LanguageID("pl"), a1pl11.RefusalCases)
}
