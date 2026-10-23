package runtime

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// SourceFeatures returns the agreed, known morphology evidence for one exact
// CORE referent or predicate. A conflict across retained source records is
// reported as unavailable so target morphology cannot silently choose a form.
func SourceFeatures(source semantic.MeaningCandidate, role meaning.Role, concept ids.ConceptID) (map[features.FeatureID]features.FeatureValueID, bool) {
	if concept.String() == "" {
		return nil, false
	}
	values := make(map[features.FeatureID]features.FeatureValueID)
	found := false
	for _, record := range source.Evidence.Records {
		if record.Outcome != status.OutcomeOK || record.CoreConceptID != concept || record.Role != role {
			continue
		}
		found = true
		for _, feature := range record.Features {
			if feature.State != features.StateKnown {
				continue
			}
			if previous, exists := values[feature.Feature]; exists && previous != feature.Value {
				return nil, false
			}
			values[feature.Feature] = feature.Value
		}
	}
	return values, found
}
