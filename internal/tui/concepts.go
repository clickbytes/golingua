package tui

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/core"
)

// conceptsState owns the local F3 filter, selection, catalog and lookup view.
// Service calls and their request tokens are owned by the root TUI model.
type conceptsState struct {
	filter         editor
	catalog        *app.CatalogResult
	pendingCatalog uint64
	pendingLookup  uint64
	selectedID     string
	detail         bool
	lookup         *core.Concept
	lookupErr      string
	lookupFound    bool
	revision       uint64
	listScroll     int
	detailScroll   int
}

// rows filters the service's coverage list by a literal, case-sensitive
// substring of stable ConceptIDs. It preserves the service's catalog order.
func (s conceptsState) rows() []catalog.ConceptCoverage {
	if s.catalog == nil || s.catalog.Inventory == nil {
		return nil
	}
	rows := s.catalog.Inventory.PerConcept
	if s.filter.text == "" {
		return rows
	}
	filtered := make([]catalog.ConceptCoverage, 0)
	for _, row := range rows {
		if strings.Contains(row.ConceptID, s.filter.text) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

// reconcile retains the selected ID while it remains visible, otherwise it
// selects the first visible row or clears selection when there are no rows.
func (s *conceptsState) reconcile() {
	rows := s.rows()
	for _, row := range rows {
		if row.ConceptID == s.selectedID {
			return
		}
	}
	s.selectedID = ""
	if len(rows) > 0 {
		s.selectedID = rows[0].ConceptID
	}
}

func (s *conceptsState) move(delta int) {
	rows := s.rows()
	if len(rows) == 0 || delta == 0 {
		return
	}
	index := 0
	for i, row := range rows {
		if row.ConceptID == s.selectedID {
			index = i
			break
		}
	}
	index = max(0, min(len(rows)-1, index+delta))
	s.selectedID = rows[index].ConceptID
	s.detail = false
	s.lookup = nil
	s.lookupErr = ""
	s.lookupFound = false
	s.pendingLookup = 0
	s.detailScroll = 0
}

// edited invalidates only in-flight lookup state from the prior filter and
// reconciles selection without discarding the catalog or its ordering.
func (s *conceptsState) edited() {
	s.revision++
	s.pendingLookup = 0
	s.detail = false
	s.lookup = nil
	s.lookupErr = ""
	s.lookupFound = false
	s.listScroll = 0
	s.detailScroll = 0
	s.reconcile()
}

func (s conceptsState) body() string {
	if s.detail {
		return s.detailBody()
	}
	var b strings.Builder
	b.WriteString("F3 Concepts · literal ConceptID filter\n")
	if s.catalog == nil {
		if s.pendingCatalog != 0 {
			b.WriteString("Loading catalog…\n")
		} else {
			b.WriteString("Catalog has not been loaded.\n")
		}
		return b.String()
	}
	if s.catalog.Inventory == nil {
		fmt.Fprintf(&b, "Catalog status: %s\n", plain(string(s.catalog.Status)))
		if len(s.catalog.Diagnostics) == 0 {
			b.WriteString("Inventory is unavailable.\n")
		}
		b.WriteString(diagnostics(s.catalog.Diagnostics))
		return b.String()
	}
	inv := s.catalog.Inventory
	fmt.Fprintf(&b, "Catalog status: %s · concepts: %d (target %d)\n", plain(string(s.catalog.Status)), inv.Concepts, inv.Target)
	b.WriteString("Filter is a case-sensitive substring of ConceptID.\n")
	if s.pendingCatalog != 0 {
		b.WriteString("Refreshing catalog…\n")
	}
	rows := s.rows()
	if len(rows) == 0 {
		b.WriteString("No concepts match this filter.\n")
		return b.String() + diagnostics(s.catalog.Diagnostics)
	}
	for _, row := range rows {
		mark := "  "
		if row.ConceptID == s.selectedID {
			mark = ">>"
		}
		fmt.Fprintf(&b, "%s %s · PL %s · EN %s\n", mark, plain(row.ConceptID), plain(string(row.Polish)), plain(string(row.English)))
	}
	fmt.Fprintf(&b, "\nUp/Down select · Enter details · Esc back · matching concepts: %d", len(rows))
	return b.String() + diagnostics(s.catalog.Diagnostics)
}

func (s conceptsState) detailBody() string {
	var b strings.Builder
	fmt.Fprintf(&b, "F3 Concept details · %s\n", plain(s.selectedID))
	if s.pendingLookup != 0 {
		b.WriteString("Loading concept details…\n")
		return b.String()
	}
	if s.lookupErr != "" {
		b.WriteString("Lookup error: ")
		b.WriteString(plain(s.lookupErr))
		b.WriteByte('\n')
		return b.String()
	}
	if !s.lookupFound || s.lookup == nil {
		b.WriteString("Concept details were not returned.\n")
		return b.String()
	}
	c := s.lookup
	fmt.Fprintf(&b, "ID: %s\nKind: %s\n", plain(c.ID.String()), plain(string(c.Kind)))
	if c.Definition == "" {
		b.WriteString("Definition: (not provided)\n")
	} else {
		fmt.Fprintf(&b, "Definition: %s\n", plain(c.Definition))
	}
	if c.Parent == nil {
		b.WriteString("Parent: (not provided)\n")
	} else {
		fmt.Fprintf(&b, "Parent: %s\n", plain(c.Parent.String()))
	}
	if len(c.RoleTypes) == 0 {
		b.WriteString("Roles: (not provided)\n")
	} else {
		b.WriteString("Roles:\n")
		for _, role := range c.RoleTypes {
			kinds := make([]string, len(role.AllowedKinds))
			for i, kind := range role.AllowedKinds {
				kinds[i] = plain(string(kind))
			}
			fmt.Fprintf(&b, "  %s · allowed kinds: %s\n", plain(role.Name), strings.Join(kinds, ", "))
		}
	}
	if len(c.DefinitionEdges) == 0 {
		b.WriteString("Relations: (not provided)\n")
	} else {
		b.WriteString("Relations:\n")
		for _, edge := range c.DefinitionEdges {
			fmt.Fprintf(&b, "  %s → %s · strength %s", plain(edge.Relation.String()), plain(edge.Target.String()), plain(string(edge.Strength)))
			if edge.Role != "" {
				fmt.Fprintf(&b, " · role %s", plain(edge.Role))
			}
			b.WriteByte('\n')
		}
	}
	if len(c.SemanticEffects) == 0 {
		b.WriteString("Effects: (not provided)\n")
	} else {
		b.WriteString("Effects:\n")
		for _, effect := range c.SemanticEffects {
			fmt.Fprintf(&b, "  %s = %s\n", plain(effect.Name), plain(effect.Value))
		}
	}
	b.WriteString("Esc returns to concepts.\n")
	return b.String()
}
