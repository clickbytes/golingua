package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

type catalogResponse struct {
	Status      status.OutcomeKind  `json:"status"`
	Inventory   *catalog.Inventory  `json:"inventory,omitempty"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

// runCatalog reports only compiled Go catalog data. --details includes one
// status per stable CORE ID; --summary omits that list.
func runCatalog(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[1] != "--json" || args[0] != "--summary" && args[0] != "--details" {
		return writeCatalog(stdout, stderr, catalogResponse{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{Code: "INVALID_CATALOG_ARGUMENTS", Message: "use catalog --summary|--details --json", Path: "catalog", Severity: status.SeverityError}}})
	}
	graph, err := catalog.Compile()
	if err != nil {
		return catalogLoadError(stdout, stderr)
	}
	polish, err := pl.CompileLexicon(graph)
	if err != nil {
		return catalogLoadError(stdout, stderr)
	}
	english, err := en.CompileLexicon(graph)
	if err != nil {
		return catalogLoadError(stdout, stderr)
	}
	inventory, err := catalog.Inspect(graph, polish, english)
	if err != nil {
		return catalogLoadError(stdout, stderr)
	}
	if args[0] == "--summary" {
		inventory.PerConcept = []catalog.ConceptCoverage{}
	}
	return writeCatalog(stdout, stderr, catalogResponse{Status: status.OutcomeOK, Inventory: &inventory})
}

func catalogLoadError(stdout, stderr io.Writer) int {
	return writeCatalog(stdout, stderr, catalogResponse{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{{Code: "CATALOG_LOAD_FAILED", Message: "could not compile the bundled Go catalog", Path: "catalog", Severity: status.SeverityError}}})
}

func writeCatalog(stdout, stderr io.Writer, response catalogResponse) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(response); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize catalog: %v\n", err)
		return 1
	}
	switch response.Status {
	case status.OutcomeOK:
		return 0
	case status.OutcomeInvalid:
		return 2
	default:
		return 1
	}
}
