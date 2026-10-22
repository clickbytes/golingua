package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/clickbytes/golingua/internal/bundled"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/internal/status"
)

type snapshotResponse struct {
	Status      status.OutcomeKind  `json:"status"`
	Preview     *staging.Preview    `json:"preview,omitempty"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

// runSnapshot explicitly previews or activates the already authored Go data.
// Activation changes only the caller-selected local snapshot store.
func runSnapshot(args []string, stdout, stderr io.Writer) int {
	if len(args) != 4 || args[0] != "--preview" && args[0] != "--apply" || args[1] != "--store" || args[2] == "" || args[3] != "--json" {
		return writeSnapshot(stdout, stderr, snapshotResponse{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{Code: "INVALID_SNAPSHOT_ARGUMENTS", Path: "snapshot", Message: "use snapshot --preview|--apply --store <path> --json", Severity: status.SeverityError}}})
	}
	candidate, err := bundled.Candidate()
	if err != nil {
		return snapshotFailure(stdout, stderr, "BUNDLED_SNAPSHOT_INVALID", err)
	}
	store, err := staging.OpenStore(args[2])
	if err != nil {
		return snapshotFailure(stdout, stderr, "SNAPSHOT_STORE_UNAVAILABLE", err)
	}
	var preview staging.Preview
	if args[0] == "--apply" {
		preview, err = store.Apply(candidate)
	} else {
		preview, err = store.Preview(candidate)
	}
	if err != nil {
		return snapshotFailure(stdout, stderr, "SNAPSHOT_OPERATION_FAILED", err)
	}
	return writeSnapshot(stdout, stderr, snapshotResponse{Status: status.OutcomeOK, Preview: &preview})
}

func snapshotFailure(stdout, stderr io.Writer, code string, err error) int {
	result := snapshotResponse{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{{Code: code, Path: "snapshot", Message: err.Error(), Severity: status.SeverityError}}}
	if source, ok := err.(interface{ Result() status.Result }); ok {
		outcome := source.Result()
		result.Status, result.Diagnostics = outcome.Status, outcome.Diagnostics
	}
	return writeSnapshot(stdout, stderr, result)
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
