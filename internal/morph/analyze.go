package morph

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/status"
)

// BuildInverseIndexFromReadings canonicalizes and compiles readings before
// creating the inverse lookup structure.
func BuildInverseIndexFromReadings(readings []Reading, profile *features.LanguageFeatureProfile, budget FiniteTableBudget) (*InverseIndex, error) {
	table, err := BuildFiniteRelationTable(readings, profile, budget)
	if err != nil {
		return nil, err
	}
	return BuildInverseIndex(table)
}

// Analyze normalizes only to NFC. It preserves case, diacritics, and every
// licensed homographic/syncretic reading in deterministic table order.
func (index *InverseIndex) Analyze(surface string) ([]Reading, error) {
	if index == nil || index.table == nil {
		return nil, morphError(status.OutcomeUnknown, "INVERSE_INDEX_UNAVAILABLE", "index", "inverse index is unavailable")
	}
	normalized, err := NormalizeSurface(surface, DefaultNormalizationBudget())
	if err != nil {
		return nil, err
	}
	positions := index.bySurface[normalized]
	if len(positions) == 0 {
		return nil, morphError(status.OutcomeUnknown, "UNKNOWN_SURFACE_FORM", "surface", "no licensed reading matches the normalized surface")
	}
	readings := make([]Reading, len(positions))
	for offset, position := range positions {
		readings[offset] = cloneReading(index.table.payload.Readings[position])
	}
	return readings, nil
}
