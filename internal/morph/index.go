package morph

import (
	"github.com/clickbytes/golingua/internal/status"
)

// InverseIndex is an immutable surface-to-reading projection of one finite
// language/profile relation table.
type InverseIndex struct {
	table     *FiniteRelationTable
	bySurface map[string][]int
}

func BuildInverseIndex(table *FiniteRelationTable) (*InverseIndex, error) {
	if table == nil || table.profile == nil {
		return nil, morphError(status.OutcomeUnknown, "FINITE_TABLE_UNAVAILABLE", "table", "inverse index requires a compiled finite table")
	}
	index := &InverseIndex{table: table, bySurface: make(map[string][]int)}
	for readingIndex, reading := range table.payload.Readings {
		index.bySurface[reading.Surface] = append(index.bySurface[reading.Surface], readingIndex)
	}
	return index, nil
}

func (index *InverseIndex) Table() *FiniteRelationTable {
	if index == nil {
		return nil
	}
	return index.table
}
