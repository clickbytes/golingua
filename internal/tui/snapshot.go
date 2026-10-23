package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/internal/status"
)

type snapshotState struct {
	revision       uint64
	pendingPreview uint64
	pendingApply   uint64
	path           string
	preview        *app.SnapshotResult
	result         *app.SnapshotResult
	dialog         bool
	confirmApply   bool
}

type snapshotDone struct {
	id, revision uint64
	path         string
	apply        bool
	reconfirm    bool
	result       app.SnapshotResult
}

func (s *snapshotState) invalidate() {
	s.revision++
	s.pendingPreview = 0
	s.preview = nil
	s.result = nil
	s.dialog = false
	s.confirmApply = false
}

func normalizeSnapshotPath(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("snapshot store path is required")
	}
	path, err := filepath.Abs(text)
	if err != nil {
		return "", fmt.Errorf("snapshot store path: %w", err)
	}
	return path, nil
}

func snapshotPreviewCmd(service *app.Service, ctx context.Context, path string, id, revision uint64) tea.Cmd {
	return func() tea.Msg {
		return snapshotDone{
			id: id, revision: revision, path: path,
			result: service.Snapshot(ctx, app.SnapshotRequest{StorePath: path}),
		}
	}
}

func snapshotApplyCmd(service *app.Service, ctx context.Context, path, expected string, id, revision uint64) tea.Cmd {
	return func() tea.Msg {
		preview := service.Snapshot(ctx, app.SnapshotRequest{StorePath: path})
		if preview.Status != status.OutcomeOK || preview.Preview == nil {
			return snapshotDone{id: id, revision: revision, path: path, apply: true, result: preview}
		}
		if preview.Preview.Digest != expected {
			return snapshotDone{
				id: id, revision: revision, path: path, apply: true, reconfirm: true,
				result: app.SnapshotResult{
					Status:  status.OutcomeInvalid,
					Preview: preview.Preview,
					Diagnostics: []status.Diagnostic{{
						Code: "SNAPSHOT_PREVIEW_CHANGED", Path: "snapshot.expected_digest", Severity: status.SeverityError,
						Message: "snapshot changed since confirmation; review the new digest and confirm again",
					}},
				},
			}
		}
		result := service.Snapshot(ctx, app.SnapshotRequest{
			StorePath: path, Apply: true, Confirmed: true, ExpectedDigest: expected,
		})
		reconfirm := false
		if result.Status != status.OutcomeOK {
			for _, diagnostic := range result.Diagnostics {
				if diagnostic.Code == "SNAPSHOT_DIGEST_MISMATCH" {
					reconfirm = true
					break
				}
			}
		}
		return snapshotDone{id: id, revision: revision, path: path, apply: true, reconfirm: reconfirm, result: result}
	}
}

func (s snapshotState) body() string {
	var b strings.Builder
	b.WriteString("F5 Snapshot · Preview does not activate a snapshot\n")
	if s.path == "" {
		b.WriteString("Snapshot store path: (not selected)\n")
	} else {
		fmt.Fprintf(&b, "Snapshot store path: %s\n", plain(s.path))
	}
	if s.pendingPreview != 0 {
		b.WriteString("Preview: checking\n")
	}
	if s.pendingApply != 0 {
		b.WriteString("Apply: in progress\n")
	}
	if s.preview == nil {
		b.WriteString("Preview result: (not returned)\n")
	} else {
		fmt.Fprintf(&b, "Preview status: %s\n", plain(string(s.preview.Status)))
		writeSnapshotPreview(&b, s.preview.Preview)
		b.WriteString(diagnostics(s.preview.Diagnostics) + "\n")
	}
	if s.result == nil {
		b.WriteString("Latest operation result: (not returned)\n")
	} else {
		fmt.Fprintf(&b, "Latest operation status: %s\n", plain(string(s.result.Status)))
		writeSnapshotPreview(&b, s.result.Preview)
		b.WriteString(diagnostics(s.result.Diagnostics) + "\n")
	}
	b.WriteString("Preview may create store directories. Apply does not select a runtime or save preferences.")
	return b.String()
}

func writeSnapshotPreview(b *strings.Builder, preview *staging.Preview) {
	if preview == nil {
		b.WriteString("Snapshot details: (not returned)\n")
		return
	}
	if preview.DatasetID == "" {
		b.WriteString("Dataset ID: (not returned)\n")
	} else {
		fmt.Fprintf(b, "Dataset ID: %s\n", plain(preview.DatasetID))
	}
	if preview.Version == "" {
		b.WriteString("Version: (not returned)\n")
	} else {
		fmt.Fprintf(b, "Version: %s\n", plain(preview.Version))
	}
	if preview.Digest == "" {
		b.WriteString("Digest: (not returned)\n")
	} else {
		fmt.Fprintf(b, "Digest: %s\n", plain(preview.Digest))
	}
	if preview.ActiveDigest == "" {
		b.WriteString("Active digest: (none returned)\n")
	} else {
		fmt.Fprintf(b, "Active digest: %s\n", plain(preview.ActiveDigest))
	}
	fmt.Fprintf(b, "Changed: %t · Applied: %t · Durability confirmed: %t\n", preview.Changed, preview.Applied, preview.DurabilityConfirmed)
	fmt.Fprintf(b, "Summary · files %d · concepts %d · relations %d · lexemes %d · senses %d · bindings %d\n",
		preview.Summary.Files, preview.Summary.Concepts, preview.Summary.RelationTypes,
		preview.Summary.Lexemes, preview.Summary.Senses, preview.Summary.Bindings)
	if preview.Warning == "" {
		b.WriteString("Warning: (none returned)\n")
	} else {
		fmt.Fprintf(b, "Warning: %s\n", plain(preview.Warning))
	}
}

func (s snapshotState) dialogBody() string {
	var b strings.Builder
	b.WriteString("Confirm snapshot Apply\n")
	if s.path == "" {
		b.WriteString("Store path: (not selected)\n")
	} else {
		fmt.Fprintf(&b, "Store path: %s\n", plain(s.path))
	}
	digest := "(not returned)"
	if s.preview != nil && s.preview.Preview != nil && s.preview.Preview.Digest != "" {
		digest = s.preview.Preview.Digest
	}
	fmt.Fprintf(&b, "Digest: %s\n", plain(digest))
	if s.confirmApply {
		b.WriteString("   Cancel\n>> Apply\n")
	} else {
		b.WriteString(">> Cancel\n   Apply\n")
	}
	b.WriteString("Tab/Left/Right changes selection · Enter confirms · default is Cancel")
	return b.String()
}
