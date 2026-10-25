package matrix

import (
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/realizer"
)

func TestRequiredSurfaceContracts(t *testing.T) {
	c := RealizationCase{
		AllowedTexts:  []string{"This is a cat.", "That is a cat."},
		RequiredTexts: []string{"This is a cat."},
		RequiredTextAssumptions: []TextAssumptionRequirement{
			{Text: "This is a cat.", Field: "reference", Value: "this"},
			{Text: "That is a cat.", Field: "reference", Value: "that"},
		},
	}
	seen := map[string]bool{"This is a cat.": true}
	if missing := missingRequiredTexts(c.RequiredTexts, seen); len(missing) != 0 {
		t.Fatalf("optional allowed variant was incorrectly required: %v", missing)
	}
	if missing := missingRequiredTexts([]string{"This is a cat.", "That is a cat."}, seen); len(missing) != 1 || missing[0] != "That is a cat." {
		t.Fatalf("missing required variant was not found: %v", missing)
	}
	candidate := engine.SurfaceAlternative{Text: "That is a cat.", Assumptions: []realizer.FieldAssumption{{Field: "reference", Value: "this"}}}
	if failures := surfaceRequirementFailures(c, candidate); len(failures) != 1 {
		t.Fatalf("wrong variant assumption should fail: %v", failures)
	}
	candidate.Assumptions[0].Value = "that"
	if failures := surfaceRequirementFailures(c, candidate); len(failures) != 0 {
		t.Fatalf("matching variant assumption should pass: %v", failures)
	}
}
