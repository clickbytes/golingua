package importer

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
)

func DeduplicateUniMorph(records []UniMorphRecord) ([]UniMorphRecord, int, error) {
	seen := make(map[string]struct{}, len(records))
	result := make([]UniMorphRecord, 0, len(records))
	duplicates := 0
	for _, record := range records {
		key, err := uniMorphRecordKey(record)
		if err != nil {
			return nil, 0, err
		}
		if _, exists := seen[key]; exists {
			duplicates++
			continue
		}
		seen[key] = struct{}{}
		result = append(result, record)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Lemma != result[j].Lemma {
			return result[i].Lemma < result[j].Lemma
		}
		if result[i].POS != result[j].POS {
			return result[i].POS < result[j].POS
		}
		if result[i].Surface != result[j].Surface {
			return result[i].Surface < result[j].Surface
		}
		if featureKey(result[i].Features) != featureKey(result[j].Features) {
			return featureKey(result[i].Features) < featureKey(result[j].Features)
		}
		return strings.Join(result[i].Tags, ";") < strings.Join(result[j].Tags, ";")
	})
	return result, duplicates, nil
}

func featureKey(bundle features.FeatureBundle) string {
	encoded, _ := json.Marshal(bundle.Values)
	return string(encoded)
}
