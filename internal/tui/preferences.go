package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/clickbytes/golingua/internal/app"
)

// preferencesState keeps editable preferences separate from values that have
// been explicitly saved and applied.
type preferencesState struct {
	active      settings
	level       string
	defaultFrom string
	store       editor
}

func newPreferencesState() preferencesState {
	state := preferencesState{}
	state.reset(defaultSettings())
	return state
}

func (s *preferencesState) reset(value settings) {
	s.active = value
	s.level = value.DisplayLevel
	s.defaultFrom = value.DefaultFrom
	s.store.set(value.SnapshotStore)
}

// value builds a candidate settings value without changing active preferences.
// Non-empty paths are normalized to absolute paths before validation.
func (s preferencesState) value(runtimeDraft string) (settings, error) {
	value := settings{
		SchemaVersion: 1,
		RuntimeConfig: runtimeDraft,
		DisplayLevel:  s.level,
		DefaultFrom:   s.defaultFrom,
		SnapshotStore: s.store.text,
	}
	if value.RuntimeConfig != "" {
		absolute, err := filepath.Abs(value.RuntimeConfig)
		if err != nil {
			return settings{}, fmt.Errorf("runtime_config: %w", err)
		}
		value.RuntimeConfig = absolute
	}
	if value.SnapshotStore != "" {
		absolute, err := filepath.Abs(value.SnapshotStore)
		if err != nil {
			return settings{}, fmt.Errorf("snapshot_store: %w", err)
		}
		value.SnapshotStore = absolute
	}
	if err := validSettings(value); err != nil {
		return settings{}, err
	}
	return value, nil
}

func (s preferencesState) body(settingsFile, notice string, runtime app.RuntimeValidationResult, pending bool) string {
	var b strings.Builder
	b.WriteString("F5 Preferences · edits are drafts until Save succeeds\n")
	fmt.Fprintf(&b, "Settings file: %s\n", plain(settingsFile))
	fmt.Fprintf(&b, "Display level · active: %s · draft: %s\n", plain(s.active.DisplayLevel), plain(s.level))
	fmt.Fprintf(&b, "Default direction · active from: %s · draft from: %s\n", plain(s.active.DefaultFrom), plain(s.defaultFrom))
	activeStore := s.active.SnapshotStore
	if activeStore == "" {
		activeStore = "(empty)"
	}
	draftStore := s.store.text
	if draftStore == "" {
		draftStore = "(empty)"
	}
	fmt.Fprintf(&b, "Snapshot store · active: %s · draft: %s\n", plain(activeStore), plain(draftStore))
	fmt.Fprintf(&b, "Runtime config · active: %s\n", plain(s.active.RuntimeConfig))
	if pending {
		b.WriteString("Runtime validation: checking\n")
	} else if runtime.Result.Status == "" {
		b.WriteString("Runtime validation: not verified\n")
	} else {
		fmt.Fprintf(&b, "Runtime validation: %s · ready: %t\n", plain(string(runtime.Result.Status)), runtime.Ready)
		b.WriteString(diagnostics(runtime.Result.Diagnostics))
	}
	b.WriteString("\nDisplay: Enter cycles Simple / Normal / Technical · Direction: Enter switches EN→PL / PL→EN\n")
	b.WriteString("Snapshot store path: edit the prompt · Save: Ctrl+S or Enter on Save\n")
	if notice != "" {
		b.WriteString(plain(notice) + "\n")
	}
	return b.String()
}
