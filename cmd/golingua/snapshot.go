package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/status"
)

type snapshotResponse = app.SnapshotResult

// runSnapshot explicitly previews or activates the already authored Go data.
// Activation changes only the caller-selected local snapshot store.
func runSnapshot(args []string, stdout, stderr io.Writer, service *app.Service) int {
	if len(args) != 4 || args[0] != "--preview" && args[0] != "--apply" || args[1] != "--store" || args[2] == "" || args[3] != "--json" {
		return writeSnapshot(stdout, stderr, snapshotResponse{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{Code: "INVALID_SNAPSHOT_ARGUMENTS", Path: "snapshot", Message: "use snapshot --preview|--apply --store <path> --json", Severity: status.SeverityError}}})
	}
	apply := args[0] == "--apply"
	return writeSnapshot(stdout, stderr, service.Snapshot(context.Background(), app.SnapshotRequest{
		StorePath: args[2], Apply: apply, Confirmed: apply,
	}))
}

func writeSnapshot(stdout, stderr io.Writer, result snapshotResponse) int {
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize snapshot result: %v\n", err)
		return 1
	}
	if result.Status == status.OutcomeOK {
		return 0
	}
	if result.Status == status.OutcomeInvalid {
		return 2
	}
	return 1
}
