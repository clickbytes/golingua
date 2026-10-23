package tui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type settings struct {
	SchemaVersion int    `json:"schema_version"`
	RuntimeConfig string `json:"runtime_config"`
	DisplayLevel  string `json:"display_level"`
	DefaultFrom   string `json:"default_from"`
	SnapshotStore string `json:"snapshot_store"`
}

type settingsLoaded struct {
	settings settings
	path     string
	err      error
}
type settingsSaved struct {
	id    uint64
	value settings
	err   error
}

func defaultSettings() settings {
	return settings{SchemaVersion: 1, DisplayLevel: "normal", DefaultFrom: "en"}
}

func validSettings(value settings) error {
	if value.SchemaVersion != 1 {
		return errors.New("settings schema version must be 1")
	}
	switch value.DisplayLevel {
	case "simple", "normal", "technical":
	default:
		return errors.New("display_level must be simple, normal, or technical")
	}
	if value.DefaultFrom != "en" && value.DefaultFrom != "pl" {
		return errors.New("default_from must be en or pl")
	}
	if value.RuntimeConfig != "" && !filepath.IsAbs(value.RuntimeConfig) {
		return errors.New("runtime_config must be an absolute path")
	}
	if value.SnapshotStore != "" && !filepath.IsAbs(value.SnapshotStore) {
		return errors.New("snapshot_store must be empty or an absolute path")
	}
	return nil
}

func settingsPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "golingua", "settings.json"), nil
}

func loadSettings() settingsLoaded {
	path, err := settingsPath()
	msg := settingsLoaded{settings: defaultSettings(), path: path, err: err}
	if err != nil {
		return msg
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return msg
	}
	if err != nil {
		msg.err = err
		return msg
	}
	if !info.Mode().IsRegular() || info.Size() > 64<<10 {
		msg.err = errors.New("settings must be a regular file within 64 KiB")
		return msg
	}
	data, err := os.ReadFile(path)
	if err != nil {
		msg.err = err
		return msg
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	value := defaultSettings()
	value.SchemaVersion = 0 // The version remains mandatory, including for older files.
	if err = dec.Decode(&value); err == nil {
		var extra any
		if dec.Decode(&extra) != io.EOF {
			err = errors.New("settings contain trailing data")
		}
	}
	if err == nil {
		err = validSettings(value)
	}
	if err != nil {
		msg.err = fmt.Errorf("invalid settings (not overwritten): %w", err)
		return msg
	}
	msg.settings = value
	return msg
}

func saveSettings(path string, value settings) error {
	if path == "" {
		return errors.New("settings path unavailable")
	}
	if err := validSettings(value); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".settings-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(append(data, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(name, path)
}
