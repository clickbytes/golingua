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

func TestRequiredTextGroups(t *testing.T) {
	groups := [][]string{{"The cat is on the table.", "The table has the cat."}}
	allowed := map[string]bool{"The cat is on the table.": true, "The table has the cat.": true}
	if invalid := invalidRequiredTextGroups(groups, allowed); len(invalid) != 0 {
		t.Fatalf("valid style group rejected: %v", invalid)
	}
	if missing := missingRequiredTextGroups(groups, map[string]bool{"The table has the cat.": true}); len(missing) != 0 {
		t.Fatalf("one licensed style should satisfy group: %v", missing)
	}
	if missing := missingRequiredTextGroups(groups, map[string]bool{}); len(missing) != 1 || missing[0] != 0 {
		t.Fatalf("empty group output should fail: %v", missing)
	}
	if invalid := invalidRequiredTextGroups(groups, map[string]bool{"The cat is on the table.": true}); len(invalid) != 1 {
		t.Fatalf("unlicensed declared alternative should fail: %v", invalid)
	}
}
