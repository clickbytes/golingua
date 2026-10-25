package translation

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/tests/matrix"
)

func TestInventory(t *testing.T) {
	if len(Cases) == 0 {
		t.Fatal("declarative language matrix inventory is empty")
	}
	analysisLinks := make(map[string]int)
	realizationLinks := make(map[string]int)
	integrationIDs := make(map[string]bool, len(Cases))
	pairs := make(map[string]struct {
		group      string
		directions map[string]int
	})
	logInventorySnapshot(t)
	for _, c := range Cases {
		t.Logf("inventory ID=%s direction=%s-%s coverage_declared=%s", c.ID, c.From, c.To, strings.Join(c.CoverageTags, ","))
		if c.PairID == "" || c.PrimaryGroup == "" {
			t.Errorf("case %q lacks pair/group metadata", c.ID)
		}
		if c.PairID != "" {
			pair := pairs[c.PairID]
			if pair.directions == nil {
				pair.group = c.PrimaryGroup
				pair.directions = map[string]int{}
			}
			if pair.group != c.PrimaryGroup {
				t.Errorf("pair %q has inconsistent groups %q and %q", c.PairID, pair.group, c.PrimaryGroup)
			}
			pair.directions[c.From+"-"+c.To]++
			pairs[c.PairID] = pair
		}
		if c.ID == "" || !validCaseID(c.ID, c.From, c.To) {
			t.Errorf("invalid integration ID/direction: %+v", c)
		}
		if strings.TrimSuffix(c.AnalysisID, ":analyze") != c.ID || strings.TrimSuffix(c.RealizationID, ":realize") != c.ID {
			t.Errorf("case %q links analysis/realization from a different record", c.ID)
		}
		if integrationIDs[c.ID] {
			t.Errorf("duplicate integration ID %q", c.ID)
		}
		integrationIDs[c.ID] = true
		analysis, realization, err := Resolve(c)
		if err != nil {
			t.Errorf("case %q reference resolution: %v", c.ID, err)
			continue
		}
		analysisLinks[analysis.ID]++
		realizationLinks[realization.ID]++
		logInventoryCase(t, c, analysis)
		if !strings.HasSuffix(analysis.ID, ":analyze") || !strings.HasSuffix(realization.ID, ":realize") {
			t.Errorf("case %q has malformed analysis/realization links", c.ID)
		}
	}
	for pairID, pair := range pairs {
		if pair.directions["en-pl"] != 1 || pair.directions["pl-en"] != 1 || len(pair.directions) != 2 {
			t.Errorf("pair %q must have exactly one en-pl and one pl-en case; got %v", pairID, pair.directions)
		}
	}
	for _, batch := range LanguageBatches {
		checkCases(t, "en-"+batch.ID, batch.EnglishAnalysis, batch.EnglishRealizations, analysisLinks, realizationLinks)
		checkCases(t, "pl-"+batch.ID, batch.PolishAnalysis, batch.PolishRealizations, analysisLinks, realizationLinks)
	}
}

func logInventorySnapshot(t *testing.T) {
	t.Helper()
	sourceSHA := matrixSourceSHA()
	snapshot := matrix.Runtime(t).PackSnapshot()
	if snapshot == nil {
		t.Fatalf("INFRA: matrix runtime has no pack snapshot")
	}
	logMatrixJSON(t, "matrix_snapshot", map[string]any{
		"schema_version": 1,
		"source_sha":     sourceSHA,
		"pack_identity":  snapshot.Identity(),
	})
	document := snapshot.Lexicon().Document()
	lexemes := make(map[string]any, len(document.Lexemes))
	for _, lexeme := range document.Lexemes {
		paradigmID := lexeme.ParadigmID
		if paradigmID == "" {
			paradigmID = "not-declared"
		}
		lexemes[lexeme.ID.String()] = map[string]any{
			"language":    lexeme.Language,
			"pos":         lexeme.POS,
			"paradigm_id": paradigmID,
		}
	}
	for _, sense := range document.Senses {
		metadata := map[string]any{
			"schema_version":      1,
			"source_sha":          sourceSHA,
			"metadata_kind":       "catalogue_association",
			"morphology_executed": false,
			"sense_id":            sense.ID.String(),
			"lexeme_id":           sense.LexemeID.String(),
		}
		if lexeme, exists := lexemes[sense.LexemeID.String()]; exists {
			metadata["association_status"] = "associated"
			metadata["lexeme"] = lexeme
		} else {
			metadata["association_status"] = "unknown"
		}
		logMatrixJSON(t, "matrix_sense_association", metadata)
	}
}

func logInventoryCase(t *testing.T, c matrix.TranslationCase, analysis matrix.AnalysisCase) {
	t.Helper()
	allowed := append([]string{analysis.MeaningID}, analysis.AllowedMeaningIDs...)
	required := append([]string{analysis.MeaningID}, analysis.RequiredMeaningIDs...)
	semantics := analysis.SemanticRequirement
	logMatrixJSON(t, "matrix_inventory", map[string]any{
		"schema_version":       1,
		"source_sha":           matrixSourceSHA(),
		"case_id":              c.ID,
		"pair_id":              c.PairID,
		"primary_group":        c.PrimaryGroup,
		"from":                 c.From,
		"to":                   c.To,
		"coverage_tags":        c.CoverageTags,
		"input":                analysis.Input,
		"meaning_id":           analysis.MeaningID,
		"allowed_meaning_ids":  allowed,
		"required_meaning_ids": required,
		"source_semantics": map[string]any{
			"predicate": semantics.RequiredPredicate,
			"roles":     semantics.RequiredRoles,
			"assertion": semantics.RequiredAssertion,
			"person":    semantics.RequiredPerson,
			"quantity":  semantics.RequiredQuantity,
			"reference": semantics.RequiredReference,
			"act":       semantics.RequiredAct,
			"gap":       semantics.RepresentationGap,
		},
		"owner_status": analysis.OwnerAcceptance,
	})
}

func logMatrixJSON(t *testing.T, marker string, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Errorf("INFRA: encode %s metadata: %v", marker, err)
		return
	}
	t.Logf("%s=%s", marker, encoded)
}

func matrixSourceSHA() string {
	if sourceSHA := strings.TrimSpace(os.Getenv("GOLINGUA_MATRIX_SHA")); sourceSHA != "" {
		return sourceSHA
	}
	return "UNRECORDED"
}

func validCaseID(id, from, to string) bool {
	if !((from == "en" && to == "pl") || (from == "pl" && to == "en")) {
		return false
	}
	parts := strings.Split(id, "-")
	if len(parts) != 5 || len(parts[0]) < 2 || !strings.ContainsRune("ABC", rune(parts[0][0])) || len(parts[1]) == 0 || parts[3] != strings.ToUpper(from) || parts[4] != strings.ToUpper(to) {
		return false
	}
	for _, level := range strings.Split(parts[0][1:], ".") {
		if !digits(level) {
			return false
		}
	}
	for i, r := range parts[1] {
		if (i == 0 && (r < 'A' || r > 'Z')) || (i > 0 && !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'))) {
			return false
		}
	}
	return digits(parts[2])
}

func digits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func TestInventoryIdentifierSupportsBatches(t *testing.T) {
	for _, test := range []struct {
		id   string
		from string
		to   string
		want bool
	}{
		{id: "A1-AC-001-EN-PL", from: "en", to: "pl", want: true},
		{id: "A1.1-NEG-012-EN-PL", from: "en", to: "pl", want: true},
		{id: "B1-NEG-012-PL-EN", from: "pl", to: "en", want: true},
		{id: "C2.1-Q-009-EN-PL", from: "en", to: "pl", want: true},
		{id: "A1..1-NEG-012-EN-PL", from: "en", to: "pl"},
		{id: "A1.1-NEG-012-EN-PL", from: "pl", to: "en"},
	} {
		if got := validCaseID(test.id, test.from, test.to); got != test.want {
			t.Errorf("validCaseID(%q, %q, %q) = %t, want %t", test.id, test.from, test.to, got, test.want)
		}
	}
}

func checkCases(t *testing.T, lang string, analyses []matrix.AnalysisCase, realizations []matrix.RealizationCase, analysisLinks, realizationLinks map[string]int) {
	t.Helper()
	analysisIDs := make(map[string]bool, len(analyses))
	realizationIDs := make(map[string]bool, len(realizations))
	for _, c := range analyses {
		if c.ID == "" || analysisIDs[c.ID] {
			t.Errorf("%s analysis ID is empty or duplicated: %q", lang, c.ID)
		}
		analysisIDs[c.ID] = true
		if analysisLinks[c.ID] != 1 {
			t.Errorf("%s analysis %q is linked %d times; want exactly once", lang, c.ID, analysisLinks[c.ID])
		}
		if c.Input == "" || len(c.MustPreserve) == 0 || c.InputOrigin == "" || c.OwnerAcceptance == "" {
			t.Errorf("%s analysis %q lacks input, requirements, provenance, or owner status", lang, c.ID)
		}
		allowedMeanings := map[string]bool{c.MeaningID: true}
		for _, meaningID := range c.AllowedMeaningIDs {
			if meaningID == "" || allowedMeanings[meaningID] {
				t.Errorf("%s analysis %q has an empty or duplicate allowed MeaningID %q", lang, c.ID, meaningID)
			}
			allowedMeanings[meaningID] = true
		}
		requiredMeanings := map[string]bool{c.MeaningID: true}
		for _, meaningID := range c.RequiredMeaningIDs {
			if meaningID == "" || requiredMeanings[meaningID] || !allowedMeanings[meaningID] {
				t.Errorf("%s analysis %q has an empty, duplicate, or disallowed required MeaningID %q", lang, c.ID, meaningID)
			}
			requiredMeanings[meaningID] = true
		}
	}
	for _, c := range realizations {
		if c.ID == "" || realizationIDs[c.ID] {
			t.Errorf("%s realization ID is empty or duplicated: %q", lang, c.ID)
		}
		realizationIDs[c.ID] = true
		if realizationLinks[c.ID] != 1 {
			t.Errorf("%s realization %q is linked %d times; want exactly once", lang, c.ID, realizationLinks[c.ID])
		}
		if len(c.ProposalTexts) == 0 || c.ReferenceOrigin == "" || c.AmbiguityNote == "" || c.OwnerAcceptance == "" {
			t.Errorf("%s realization %q lacks proposal, provenance, or owner status", lang, c.ID)
		}
		allowed := make(map[string]bool, len(c.AllowedTexts))
		for _, text := range c.AllowedTexts {
			allowed[text] = true
		}
		allowedAssumptions := make(map[string]bool, len(c.AllowedAssumptionFields))
		for _, field := range c.AllowedAssumptionFields {
			allowedAssumptions[field] = true
		}
		for _, field := range c.RequiredAssumptionFields {
			if !allowedAssumptions[field] {
				t.Errorf("%s realization %q requires unallowed assumption field %q", lang, c.ID, field)
			}
		}
		for _, rule := range c.RequiredTextAssumptions {
			if !allowed[rule.Text] || rule.Field == "" || rule.Value == "" {
				t.Errorf("%s realization %q has invalid per-text assumption rule: %+v", lang, c.ID, rule)
			}
			if !allowedAssumptions[rule.Field] {
				t.Errorf("%s realization %q requires per-text assumption field %q outside AllowedAssumptionFields", lang, c.ID, rule.Field)
			}
		}
		if invalid := matrixInvalidRequiredTextGroups(c.RequiredTextGroups, allowed); len(invalid) > 0 {
			t.Errorf("%s realization %q has invalid required text groups %v", lang, c.ID, invalid)
		}
		variantIDs := map[string]bool{c.MeaningID: true}
		for _, variant := range c.MeaningVariants {
			if variant.MeaningID == "" || variantIDs[variant.MeaningID] || len(variant.AllowedTexts) == 0 {
				t.Errorf("%s realization %q has an empty/duplicate Meaning variant or no allowed texts: %+v", lang, c.ID, variant)
			}
			variantIDs[variant.MeaningID] = true
			variantAllowed := make(map[string]bool, len(variant.AllowedTexts))
			for _, text := range variant.AllowedTexts {
				variantAllowed[text] = true
			}
			if invalid := matrixInvalidRequiredTextGroups(variant.RequiredTextGroups, variantAllowed); len(invalid) > 0 {
				t.Errorf("%s realization %q variant %q has invalid required text groups %v", lang, c.ID, variant.MeaningID, invalid)
			}
			for _, rule := range variant.RequiredTextAssumptions {
				if !variantAllowed[rule.Text] || rule.Field == "" || rule.Value == "" {
					t.Errorf("%s realization %q variant %q has invalid assumption rule: %+v", lang, c.ID, variant.MeaningID, rule)
				}
				if !allowedAssumptions[rule.Field] {
					t.Errorf("%s realization %q variant %q requires per-text assumption field %q outside AllowedAssumptionFields", lang, c.ID, variant.MeaningID, rule.Field)
				}
			}
			variantRequired := make(map[string]bool, len(variant.RequiredTexts))
			for _, text := range variant.RequiredTexts {
				if variantRequired[text] || !variantAllowed[text] {
					t.Errorf("%s realization %q variant %q requires a repeated or disallowed text %q", lang, c.ID, variant.MeaningID, text)
				}
				variantRequired[text] = true
			}
		}
		required := make(map[string]bool, len(c.RequiredTexts))
		for _, text := range c.RequiredTexts {
			if required[text] {
				t.Errorf("%s realization %q repeats required text %q", lang, c.ID, text)
			}
			if !allowed[text] {
				t.Errorf("%s realization %q requires text that is not allowed: %q", lang, c.ID, text)
			}
			required[text] = true
		}
	}
}

func matrixInvalidRequiredTextGroups(groups [][]string, allowed map[string]bool) []int {
	var invalid []int
	for i, group := range groups {
		valid := len(group) > 0
		for _, text := range group {
			if text == "" || !allowed[text] {
				valid = false
			}
		}
		if !valid {
			invalid = append(invalid, i)
		}
	}
	return invalid
}

func TestA18NegationLinksHaveMatchingMeaningContracts(t *testing.T) {
	for _, c := range Cases {
		if !strings.HasPrefix(c.ID, "A1.8-NEG-") {
			continue
		}
		analysis, realization, err := Resolve(c)
		if err != nil {
			t.Fatalf("%s resolve: %v", c.ID, err)
		}
		if analysis.MeaningID == "" || analysis.MeaningID != realization.MeaningID {
			t.Errorf("%s source and target MeaningIDs must be the same nonempty ID; got %q and %q", c.ID, analysis.MeaningID, realization.MeaningID)
		}
		if !reflect.DeepEqual(analysis.SemanticRequirement, realization.SemanticRequirement) {
			t.Errorf("%s source and target semantic requirements differ", c.ID)
		}
	}
}
