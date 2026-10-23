package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/status"
)

type catalogResponse = app.CatalogResult
type catalogFrameCounts = app.ValencyFrameCounts
type catalogConstructionBindingCounts = app.ConstructionBindingCounts

// runCatalog reports only compiled Go catalog data. --details includes one
// status per stable CORE ID; --summary omits that list.
func runCatalog(args []string, stdout, stderr io.Writer, service *app.Service) int {
	if len(args) != 2 || args[1] != "--json" || args[0] != "--summary" && args[0] != "--details" {
		return writeCatalog(stdout, stderr, catalogResponse{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{Code: "INVALID_CATALOG_ARGUMENTS", Message: "use catalog --summary|--details --json", Path: "catalog", Severity: status.SeverityError}}})
	}
	return writeCatalog(stdout, stderr, service.Catalog(context.Background(), app.CatalogRequest{Details: args[0] == "--details"}))
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
