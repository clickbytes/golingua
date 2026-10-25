package translation_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	a1en11 "github.com/clickbytes/golingua/tests/languages/en/A1.1"
	a1pl11 "github.com/clickbytes/golingua/tests/languages/pl/A1.1"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestRefusalTranslation(t *testing.T) {
	matrix.RunRefusalTranslation(t, features.LanguageID("en"), features.LanguageID("pl"), a1en11.RefusalCases)
	matrix.RunRefusalTranslation(t, features.LanguageID("pl"), features.LanguageID("en"), a1pl11.RefusalCases)
}
