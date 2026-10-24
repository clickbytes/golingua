package enpl_test

import (
	"strings"
	"testing"

	"github.com/clickbytes/golingua/tests/matrix"
	"github.com/clickbytes/golingua/tests/translation"
)

func TestSeeing(t *testing.T)         { runDomain(t, "A1-AC-") }
func TestIdentification(t *testing.T) { runDomain(t, "A1-ID-") }
func TestNegation(t *testing.T)       { runDomain(t, "A1-NEG-") }
func TestOtherDomains(t *testing.T)   { runDomain(t, "other") }

func runDomain(t *testing.T, prefix string) {
	t.Helper()
	selected := 0
	for _, c := range translation.CasesFor("en", "pl") {
		if prefix == "other" {
			if strings.HasPrefix(c.ID, "A1-AC-") || strings.HasPrefix(c.ID, "A1-ID-") || strings.HasPrefix(c.ID, "A1-NEG-") {
				continue
			}
		} else if !strings.HasPrefix(c.ID, prefix) {
			continue
		}
		selected++
		analysis, realization, err := translation.Resolve(c)
		if err != nil {
			t.Fatalf("SPEC_PENDING: translation %q references: %v", c.ID, err)
		}
		matrix.RunTranslation(t, c, analysis, realization)
	}
	if selected == 0 {
		t.Errorf("SPEC_PENDING: no translation cases selected for %q", prefix)
	}
}
