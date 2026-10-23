package tui

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/app"
)

// languagesState keeps catalog inventory separate from runtime readiness.
type languagesState struct {
	catalog           *app.CatalogResult
	validation        app.RuntimeValidationResult
	path              string
	pendingCatalog    uint64
	pendingValidation uint64
}

func (s languagesState) body() string {
	var b strings.Builder
	b.WriteString("F4 Languages · catalog inventory and runtime readiness\n")
	if s.catalog == nil {
		if s.pendingCatalog != 0 {
			b.WriteString("Loading catalog inventory…\n")
		} else {
			b.WriteString("Catalog inventory has not been loaded.\n")
		}
	} else if s.catalog.Inventory == nil {
		fmt.Fprintf(&b, "Catalog status: %s\n", plain(string(s.catalog.Status)))
		if len(s.catalog.Diagnostics) == 0 {
			b.WriteString("Inventory is unavailable.\n")
		}
		b.WriteString(diagnostics(s.catalog.Diagnostics))
	} else {
		inventory := s.catalog.Inventory
		fmt.Fprintf(&b, "Catalog status: %s · concepts: %d (target %d)\n", plain(string(s.catalog.Status)), inventory.Concepts, inventory.Target)
		fmt.Fprintf(&b, "PL · lexemes %d · senses %d · bindings %d\n", inventory.Polish.Lexemes, inventory.Polish.Senses, inventory.Polish.Bindings)
		fmt.Fprintf(&b, "   concepts · approved %d · pending %d · non-transferable %d · absent %d\n", inventory.Polish.ApprovedConcepts, inventory.Polish.PendingConcepts, inventory.Polish.OtherConcepts, inventory.Polish.AbsentConcepts)
		fmt.Fprintf(&b, "EN · lexemes %d · senses %d · bindings %d\n", inventory.English.Lexemes, inventory.English.Senses, inventory.English.Bindings)
		fmt.Fprintf(&b, "   concepts · approved %d · pending %d · non-transferable %d · absent %d\n", inventory.English.ApprovedConcepts, inventory.English.PendingConcepts, inventory.English.OtherConcepts, inventory.English.AbsentConcepts)
		if s.pendingCatalog != 0 {
			b.WriteString("Refreshing catalog inventory…\n")
		}
		b.WriteString(diagnostics(s.catalog.Diagnostics))
	}
	b.WriteString("\nRuntime readiness\n")
	if s.path == "" {
		b.WriteString("Selected runtime path: (none)\n")
	} else {
		fmt.Fprintf(&b, "Selected runtime path: %s\n", plain(s.path))
	}
	if s.pendingValidation != 0 {
		b.WriteString("Validation: checking\n")
	} else if s.validation.Result.Status == "" {
		b.WriteString("Validation: not verified\n")
	} else {
		fmt.Fprintf(&b, "Validation: %s · ready: %t\n", plain(string(s.validation.Result.Status)), s.validation.Ready)
		b.WriteString(diagnostics(s.validation.Result.Diagnostics))
	}
	b.WriteString("\nA loadable configuration does not promise support for every sentence.\n")
	b.WriteString("Enter on Refresh rechecks inventory and the selected runtime.")
	return b.String()
}
