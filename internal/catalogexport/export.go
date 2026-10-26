// Package catalogexport provides a bounded, inert export of the authored CORE
// catalog. It does not infer formal relations from concept definitions.
package catalogexport

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/metamodel"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	FormatID              = "catalog-dse-inventory-v1"
	ExportVersion         = "2026-10-26.1"
	maxSourceRecords      = 4096
	maxPackageRecords     = 400
	maxPackages           = 16
	maxPackageBytes       = 1 << 20
	maxCombinedExportSize = 16 << 20
)

const (
	ClassificationOpaque = "opaque_unformalized"
	MappingUnformalized  = "unformalized"
)

// SourceRecord retains one typed CORE definition and its provenance. The
// classification is an inventory decision, not a claim that the description
// has a formal denotation.
type SourceRecord struct {
	Definition     core.ConceptDefinition `json:"definition"`
	Source         provenance.SourceRef   `json:"source"`
	Classification string                 `json:"classification"`
	ReviewStatus   metamodel.ReviewStatus `json:"review_status"`
	MappingStatus  string                 `json:"mapping_status"`
}

// Counts distinguish the source inventory from inert metadata symbols.
type Counts struct {
	SourceRecords int `json:"source_records"`
	Pending       int `json:"pending"`
	Opaque        int `json:"opaque_unformalized"`
}

// ExportDocument contains the full source inventory and bounded metadata
// packages. Packages intentionally classify every catalog item as opaque and
// pending until it receives a separate formal and editorial review.
type ExportDocument struct {
	SchemaVersion int                  `json:"schema_version"`
	Format        string               `json:"format"`
	Version       string               `json:"version"`
	Counts        Counts               `json:"counts"`
	SourceRecords []SourceRecord       `json:"source_records"`
	Packages      []metamodel.Document `json:"packages"`
}

// Export builds a deterministic inventory from the current catalog entries.
// Each emitted package is validated by the production metadata compiler before
// this all-or-nothing result is returned.
func Export() (ExportDocument, error) {
	entries, err := catalog.Entries()
	if err != nil {
		return ExportDocument{}, fmt.Errorf("read catalog entries: %w", err)
	}
	if len(entries) > maxSourceRecords {
		return ExportDocument{}, exportError(status.OutcomeResourceLimit, "CATALOG_EXPORT_LIMIT", "source_records", "catalog entry count exceeds the export limit")
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Definition.ID.String() < entries[j].Definition.ID.String()
	})

	result := ExportDocument{
		SchemaVersion: 1,
		Format:        FormatID,
		Version:       ExportVersion,
		SourceRecords: make([]SourceRecord, 0, len(entries)),
		Packages:      make([]metamodel.Document, 0),
	}
	result.Counts = Counts{SourceRecords: len(entries), Pending: len(entries), Opaque: len(entries)}
	for _, entry := range entries {
		result.SourceRecords = append(result.SourceRecords, SourceRecord{
			Definition:     cloneDefinition(entry.Definition),
			Source:         entry.Source,
			Classification: ClassificationOpaque,
			ReviewStatus:   metamodel.ReviewPending,
			MappingStatus:  MappingUnformalized,
		})
	}

	for start, part := 0, 1; start < len(entries); part++ {
		if part > maxPackages {
			return ExportDocument{}, exportError(status.OutcomeResourceLimit, "CATALOG_EXPORT_PACKAGE_LIMIT", "packages", "catalog requires more than the export package limit")
		}
		count := len(entries) - start
		if count > maxPackageRecords {
			count = maxPackageRecords
		}
		var selected metamodel.Document
		selectedCount := 0
		for count > 0 {
			candidate := makePackage(entries[start:start+count], part)
			encoded, marshalErr := json.Marshal(candidate)
			if marshalErr != nil {
				return ExportDocument{}, fmt.Errorf("encode catalog package %d: %w", part, marshalErr)
			}
			if len(encoded) <= maxPackageBytes {
				if _, compileErr := metamodel.Compile(candidate, nil); compileErr != nil {
					return ExportDocument{}, fmt.Errorf("validate catalog package %d: %w", part, compileErr)
				}
				selected, selectedCount = candidate, count
				break
			}
			count /= 2
		}
		if selectedCount == 0 {
			return ExportDocument{}, exportError(status.OutcomeResourceLimit, "CATALOG_EXPORT_PACKAGE_BYTES", fmt.Sprintf("packages[%d]", part-1), "a catalog record cannot fit in a package within the byte limit")
		}
		result.Packages = append(result.Packages, selected)
		start += selectedCount
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return ExportDocument{}, fmt.Errorf("encode catalog export: %w", err)
	}
	if len(encoded) > maxCombinedExportSize {
		return ExportDocument{}, exportError(status.OutcomeResourceLimit, "CATALOG_EXPORT_BYTES", "export", "combined catalog export exceeds the byte limit")
	}
	return result, nil
}

func makePackage(entries []catalog.Entry, part int) metamodel.Document {
	sources := make(map[string]provenance.SourceRef)
	symbols := make([]metamodel.Symbol, 0, len(entries))
	for _, entry := range entries {
		source := entry.Source
		sources[source.ID] = source
		symbols = append(symbols, metamodel.Symbol{
			ID:           entry.Definition.ID.String(),
			Kind:         metamodel.SymbolOpaqueUnformalized,
			Description:  entry.Definition.Definition,
			ReviewStatus: metamodel.ReviewPending,
			Provenance:   []provenance.SourceRef{source},
		})
	}
	sourceIDs := make([]string, 0, len(sources))
	for id := range sources {
		sourceIDs = append(sourceIDs, id)
	}
	sort.Strings(sourceIDs)
	packageSources := make([]provenance.SourceRef, 0, len(sourceIDs))
	for _, id := range sourceIDs {
		packageSources = append(packageSources, sources[id])
	}
	return metamodel.Document{
		SchemaVersion: metamodel.DocumentSchemaVersion,
		ID:            fmt.Sprintf("golingua.catalog.inventory.2026-10-26.part-%03d", part),
		Version:       ExportVersion,
		Profile:       metamodel.PositiveFactsProfile,
		Provenance:    packageSources,
		Sorts:         make([]algebra.SortDeclaration, 0),
		Symbols:       symbols,
		Dependencies:  make([]metamodel.Dependency, 0),
		TypicalEdges:  make([]metamodel.TypicalEdge, 0),
		Axioms:        make([]json.RawMessage, 0),
	}
}

func cloneDefinition(definition core.ConceptDefinition) core.ConceptDefinition {
	copyOf := definition
	if definition.Parent != nil {
		parent := *definition.Parent
		copyOf.Parent = &parent
	}
	if definition.RoleTypes != nil {
		copyOf.RoleTypes = make([]core.RoleType, len(definition.RoleTypes))
		for index, role := range definition.RoleTypes {
			copyOf.RoleTypes[index] = core.RoleType{
				Name:         role.Name,
				AllowedKinds: append([]core.ConceptKind(nil), role.AllowedKinds...),
			}
		}
	}
	if definition.DefinitionEdges != nil {
		copyOf.DefinitionEdges = append([]core.DefinitionEdge{}, definition.DefinitionEdges...)
	}
	if definition.SemanticEffects != nil {
		copyOf.SemanticEffects = append([]core.SemanticEffect{}, definition.SemanticEffects...)
	}
	return copyOf
}

func exportError(kind status.OutcomeKind, code, path, message string) error {
	return &algebra.Error{Status: kind, Code: code, Path: path, Message: message}
}
