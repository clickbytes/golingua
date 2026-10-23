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
}

type settingsLoaded struct {
	settings settings
	path     string
	err      error
}
type settingsSaved struct {
	id  uint64
	err error
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
	msg := settingsLoaded{settings: settings{SchemaVersion: 1}, path: path, err: err}
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
	var value settings
	if err = dec.Decode(&value); err == nil {
		var extra any
		if dec.Decode(&extra) != io.EOF {
			err = errors.New("settings contain trailing data")
		}
	}
	if err == nil && (value.SchemaVersion != 1 || value.RuntimeConfig != "" && !filepath.IsAbs(value.RuntimeConfig)) {
		err = errors.New("invalid settings version or non-absolute runtime path")
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
	if value.SchemaVersion != 1 || value.RuntimeConfig != "" && !filepath.IsAbs(value.RuntimeConfig) {
		return errors.New("settings require version 1 and an absolute runtime path")
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
