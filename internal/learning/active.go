package learning

import (
	"fmt"
	"reflect"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/morph"
)

// ActiveRules loads active corrections only when each recorded baseline can
// be reproduced exactly by the caller's generator without applying overlays.
// On any mismatch it returns no partial rule set.
func ActiveRules(
	path, datasetID, datasetVersion, baselineDigest string,
	generate func(features.LanguageID, string, []features.FeatureValue) (morph.GenerationResult, error),
) (map[features.LanguageID][]morph.ExceptionRule, error) {
	if !validDigest(baselineDigest) {
		return nil, fmt.Errorf("baseline digest must be a SHA-256 digest")
	}
	if !validIdentifier(datasetID, MaxIDBytes) || !validDatasetVersion(datasetVersion) {
		return nil, fmt.Errorf("dataset ID and version are required bounded identifiers")
	}
	if generate == nil {
		return nil, fmt.Errorf("base morphology generator is required")
	}
	store, err := Load(path)
	if err != nil {
		return nil, err
	}

	active := make(map[features.LanguageID][]morph.ExceptionRule)
	for index, record := range store.Records {
		if !record.Active {
			continue
		}
		if digestValue(record.BaselineDigest) != digestValue(baselineDigest) {
			return nil, fmt.Errorf("records[%d] (%s): baseline digest does not match requested snapshot", index, record.ID)
		}
		if record.DatasetID != datasetID || record.DatasetVersion != datasetVersion {
			return nil, fmt.Errorf("records[%d] (%s): dataset identity does not match requested snapshot", index, record.ID)
		}
		generated, err := generate(record.Language, record.Rule.Selector.LexemeID, append([]features.FeatureValue(nil), record.Features...))
		if err != nil {
			return nil, fmt.Errorf("records[%d] (%s): regenerate baseline: %w", index, record.ID, err)
		}
		if generated.State != morph.GenerationGenerated || len(generated.Forms) != 1 {
			return nil, fmt.Errorf("records[%d] (%s): base generator must return generated state with exactly one form", index, record.ID)
		}
		if !reflect.DeepEqual(generated.Forms[0], record.BaselineForm) {
			return nil, fmt.Errorf("records[%d] (%s): regenerated form differs from recorded baseline", index, record.ID)
		}
		active[record.Language] = append(active[record.Language], cloneRule(record.Rule))
	}
	return active, nil
}

func cloneRule(rule morph.ExceptionRule) morph.ExceptionRule {
	rule.Selector.Features = append([]morph.ExceptionFeatureMatch(nil), rule.Selector.Features...)
	rule.Effect.Forms = append([]string(nil), rule.Effect.Forms...)
	rule.OverrideOf = append([]morph.ExceptionID(nil), rule.OverrideOf...)
	return rule
}
