package staging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/status"
)

const storeSchemaVersion = 1

type storePointer struct {
	SchemaVersion int    `json:"schema_version"`
	DatasetID     string `json:"dataset_id"`
	Digest        string `json:"digest"`
}

// Preview describes the result of checking or applying a candidate. Applied
// is true only after a complete content-addressed snapshot has been published
// and its active pointer has been atomically replaced.
type Preview struct {
	DatasetID           string  `json:"dataset_id"`
	Version             string  `json:"version"`
	Digest              string  `json:"digest"`
	ActiveDigest        string  `json:"active_digest,omitempty"`
	Changed             bool    `json:"changed"`
	Applied             bool    `json:"applied"`
	DurabilityConfirmed bool    `json:"durability_confirmed"`
	Warning             string  `json:"warning,omitempty"`
	Summary             Summary `json:"summary"`
}

// Store manages immutable local snapshots and one active pointer per dataset.
// A Store serializes its own operations; separate processes may race to update
// a pointer, but every visible pointer names a complete immutable snapshot.
type Store struct {
	root string
	mu   sync.Mutex
}

// OpenStore creates or opens a local snapshot store. The store root is an
// explicitly selected location; its managed child directories must not be
// symlinks.
func OpenStore(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		return nil, storeError(status.OutcomeInvalid, "INVALID_STORE_ROOT", "store", "snapshot store root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, storeInfrastructure("STORE_ROOT_UNAVAILABLE", "store", err)
	}
	if err := os.MkdirAll(abs, 0o700); err != nil {
		return nil, storeInfrastructure("STORE_ROOT_UNAVAILABLE", "store", err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, storeInfrastructure("STORE_ROOT_UNAVAILABLE", "store", err)
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		if err == nil {
			err = errors.New("store root is not a directory")
		}
		return nil, storeInfrastructure("STORE_ROOT_UNAVAILABLE", "store", err)
	}
	for _, child := range []string{"active", "snapshots"} {
		if err := ensureDirectory(filepath.Join(resolved, child)); err != nil {
			return nil, storeInfrastructure("STORE_DIRECTORY_UNAVAILABLE", child, err)
		}
	}
	if err := syncDirectory(resolved); err != nil {
		return nil, storeInfrastructure("STORE_DIRECTORY_SYNC_FAILED", "store", err)
	}
	return &Store{root: resolved}, nil
}

// Preview validates an immutable copy of the candidate and compares it to the
// current active digest without writing snapshots or changing the pointer.
func (store *Store) Preview(candidate Candidate) (Preview, error) {
	if store == nil {
		return Preview{}, storeError(status.OutcomeUnknown, "STORE_UNAVAILABLE", "store", "snapshot store is unavailable")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	stable, _, err := validateCandidate(candidate)
	if err != nil {
		return Preview{}, err
	}
	activeDigest, err := store.activeDigest(stable.Manifest.DatasetID)
	if err != nil {
		return Preview{}, err
	}
	return Preview{
		DatasetID: stable.Manifest.DatasetID, Version: stable.Manifest.Version,
		Digest: stable.Digest, ActiveDigest: activeDigest,
		Changed: activeDigest != stable.Digest, DurabilityConfirmed: true, Summary: stable.Summary,
	}, nil
}

// Active loads and revalidates the complete snapshot named by a dataset's
// active pointer. It never follows a symlink within the managed store tree.
func (store *Store) Active(datasetID string) (Candidate, error) {
	if store == nil {
		return Candidate{}, storeError(status.OutcomeUnknown, "STORE_UNAVAILABLE", "store", "snapshot store is unavailable")
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if !validDatasetID(datasetID) {
		return Candidate{}, storeError(status.OutcomeInvalid, "INVALID_DATASET_ID", "dataset_id", "dataset ID is malformed")
	}
	pointer, err := store.readPointer(datasetID)
	if err != nil {
		return Candidate{}, err
	}
	return store.loadSnapshot(pointer.DatasetID, pointer.Digest)
}

func (store *Store) activeDigest(datasetID string) (string, error) {
	if !validDatasetID(datasetID) {
		return "", storeError(status.OutcomeInvalid, "INVALID_DATASET_ID", "dataset_id", "dataset ID is malformed")
	}
	activeDir := filepath.Join(store.root, "active")
	if err := ensureExistingDirectory(activeDir); err != nil {
		return "", storeInfrastructure("ACTIVE_DIRECTORY_UNAVAILABLE", "active", err)
	}
	pointerPath := filepath.Join(store.root, "active", datasetID+".json")
	info, err := os.Lstat(pointerPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", storeInfrastructure("ACTIVE_POINTER_READ_FAILED", "active", err)
	}
	if !info.Mode().IsRegular() {
		return "", storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer must be a regular file")
	}
	pointer, err := store.readPointer(datasetID)
	if err != nil {
		return "", err
	}
	// Fail closed if the pointer names a missing or corrupt snapshot.
	if _, err := store.loadSnapshot(pointer.DatasetID, pointer.Digest); err != nil {
		return "", err
	}
	return pointer.Digest, nil
}

func (store *Store) readPointer(datasetID string) (storePointer, error) {
	activeDir := filepath.Join(store.root, "active")
	if err := ensureExistingDirectory(activeDir); err != nil {
		return storePointer{}, storeInfrastructure("ACTIVE_DIRECTORY_UNAVAILABLE", "active", err)
	}
	path := filepath.Join(activeDir, datasetID+".json")
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return storePointer{}, storeError(status.OutcomeUnknown, "NO_ACTIVE_SNAPSHOT", "active", "dataset has no active snapshot")
		}
		return storePointer{}, storeInfrastructure("ACTIVE_POINTER_READ_FAILED", "active", err)
	}
	if !info.Mode().IsRegular() || info.Size() > 4096 {
		return storePointer{}, storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer must be a bounded regular file")
	}
	data, err := readRegularFile(path, 4096)
	if err != nil {
		return storePointer{}, storeInfrastructure("ACTIVE_POINTER_READ_FAILED", "active", err)
	}
	if hasDuplicateJSONKeys(data) {
		return storePointer{}, storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer contains duplicate JSON keys")
	}
	var pointer storePointer
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&pointer); err != nil {
		return storePointer{}, storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer JSON is malformed")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return storePointer{}, storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer must contain exactly one JSON value")
	}
	if pointer.SchemaVersion != storeSchemaVersion || pointer.DatasetID != datasetID || !validDigest(pointer.Digest) {
		return storePointer{}, storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer version, dataset ID, or digest is invalid")
	}
	canonical, err := json.Marshal(pointer)
	if err != nil || string(canonical) != string(data) {
		return storePointer{}, storeError(status.OutcomeInvalid, "INVALID_ACTIVE_POINTER", "active", "active pointer is not in canonical form")
	}
	return pointer, nil
}

func hasDuplicateJSONKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var scan func() error
	scan = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := make(map[string]struct{})
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return fmt.Errorf("object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return errDuplicatePointerKey
				}
				seen[key] = struct{}{}
				if err := scan(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := scan(); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter")
		}
	}
	if err := scan(); err != nil {
		return errors.Is(err, errDuplicatePointerKey)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return false
	}
	return false
}

var errDuplicatePointerKey = errors.New("duplicate pointer JSON key")

func (store *Store) loadSnapshot(datasetID, digest string) (Candidate, error) {
	if !validDatasetID(datasetID) || !validDigest(digest) {
		return Candidate{}, storeError(status.OutcomeInvalid, "INVALID_SNAPSHOT_IDENTITY", "snapshot", "snapshot identity is malformed")
	}
	snapshotDir := filepath.Join(store.root, "snapshots", datasetID, digest)
	if err := ensureExistingDirectory(filepath.Join(store.root, "snapshots")); err != nil {
		return Candidate{}, storeError(status.OutcomeInvalid, "UNSAFE_SNAPSHOT_PATH", "snapshots", "snapshot root is missing or unsafe")
	}
	if err := ensureExistingDirectory(filepath.Join(store.root, "snapshots", datasetID)); err != nil {
		return Candidate{}, storeError(status.OutcomeInvalid, "SNAPSHOT_UNAVAILABLE", "snapshot", "dataset snapshot directory is missing or unsafe")
	}
	if err := ensureExistingDirectory(snapshotDir); err != nil {
		return Candidate{}, storeError(status.OutcomeInvalid, "SNAPSHOT_UNAVAILABLE", "snapshot", "active snapshot directory is missing or unsafe")
	}
	manifestPath := filepath.Join(snapshotDir, "manifest.json")
	manifestInfo, err := os.Lstat(manifestPath)
	if err != nil || !manifestInfo.Mode().IsRegular() || manifestInfo.Size() > importer.MaxManifestBytes {
		return Candidate{}, storeError(status.OutcomeInvalid, "INVALID_SNAPSHOT_MANIFEST", "snapshot.manifest", "snapshot manifest is missing, unsafe, or oversized")
	}
	manifestBytes, err := readRegularFile(manifestPath, importer.MaxManifestBytes)
	if err != nil {
		return Candidate{}, storeInfrastructure("SNAPSHOT_READ_FAILED", "snapshot.manifest", err)
	}
	manifest, err := importer.DecodeManifest(manifestBytes)
	if err != nil {
		return Candidate{}, storeError(status.OutcomeInvalid, "INVALID_SNAPSHOT_MANIFEST", "snapshot.manifest", err.Error())
	}
	if manifest.DatasetID != datasetID {
		return Candidate{}, storeError(status.OutcomeInvalid, "SNAPSHOT_DATASET_MISMATCH", "snapshot.manifest.dataset_id", "snapshot manifest dataset does not match its directory")
	}
	for _, file := range manifest.Files {
		if reservedManifestPath(file.Path) {
			return Candidate{}, storeError(status.OutcomeInvalid, "RESERVED_SNAPSHOT_PATH", file.Path, "dataset file collides with the snapshot manifest")
		}
		if err := ensureNoSymlinkPath(snapshotDir, filepath.Join(snapshotDir, filepath.FromSlash(file.Path))); err != nil {
			return Candidate{}, storeError(status.OutcomeInvalid, "UNSAFE_SNAPSHOT_PATH", file.Path, "snapshot data path contains a symlink or escapes the snapshot")
		}
	}
	if err := verifySnapshotTree(snapshotDir, manifest); err != nil {
		return Candidate{}, storeError(status.OutcomeInvalid, "INVALID_SNAPSHOT_TREE", "snapshot", err.Error())
	}
	candidate, err := BuildCandidate(snapshotDir, manifest)
	if err != nil {
		return Candidate{}, preserveImportFailure(err, status.OutcomeInvalid, "INVALID_SNAPSHOT_DATA", "snapshot")
	}
	if candidate.Digest != digest {
		return Candidate{}, storeError(status.OutcomeInvalid, "SNAPSHOT_DIGEST_MISMATCH", "snapshot.digest", "snapshot contents do not match their content address")
	}
	return candidate, nil
}

// verifySnapshotTree rejects content that is not covered by the manifest and
// digest. The only permitted directories are parents required by declared
// files; every file must be regular and either the manifest or declared data.
func verifySnapshotTree(root string, manifest importer.Manifest) error {
	expectedFiles := map[string]struct{}{"manifest.json": {}}
	expectedDirectories := map[string]struct{}{".": {}}
	for _, file := range manifest.Files {
		relative := filepath.ToSlash(filepath.FromSlash(file.Path))
		expectedFiles[relative] = struct{}{}
		parent := filepath.ToSlash(filepath.Dir(filepath.FromSlash(relative)))
		for parent != "." && parent != "" {
			expectedDirectories[parent] = struct{}{}
			parent = filepath.ToSlash(filepath.Dir(filepath.FromSlash(parent)))
		}
	}
	seenFiles := make(map[string]struct{}, len(expectedFiles))
	seenDirectories := make(map[string]struct{}, len(expectedDirectories))
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			seenDirectories["."] = struct{}{}
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("snapshot contains symlink %q", name)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("snapshot contains symlink %q", name)
		}
		if info.IsDir() {
			if _, ok := expectedDirectories[name]; !ok {
				return fmt.Errorf("snapshot contains undeclared directory %q", name)
			}
			seenDirectories[name] = struct{}{}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("snapshot contains non-regular file %q", name)
		}
		if _, ok := expectedFiles[name]; !ok {
			return fmt.Errorf("snapshot contains undeclared file %q", name)
		}
		seenFiles[name] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	for name := range expectedFiles {
		if _, ok := seenFiles[name]; !ok {
			return fmt.Errorf("snapshot is missing required file %q", name)
		}
	}
	for name := range expectedDirectories {
		if _, ok := seenDirectories[name]; !ok {
			return fmt.Errorf("snapshot is missing required directory %q", name)
		}
	}
	return nil
}

func validDatasetID(value string) bool {
	if len(value) == 0 || len(value) > importer.MaxManifestIDBytes {
		return false
	}
	for index, character := range value {
		if index == 0 && !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9')) {
			return false
		}
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}

func reservedManifestPath(value string) bool {
	return strings.EqualFold(filepath.ToSlash(value), "manifest.json")
}

func storeError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: boundedDiagnostic(code, status.MaxDiagnosticCodeBytes), Path: boundedDiagnostic(path, status.MaxDiagnosticPathBytes), Message: boundedDiagnostic(message, status.MaxDiagnosticMessageBytes), Severity: status.SeverityError}
	return &StoreError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

func boundedDiagnostic(value string, maximum int) string {
	value = strings.ToValidUTF8(value, "�")
	if len(value) <= maximum {
		return value
	}
	value = value[:maximum]
	for len(value) > 0 && !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func storeInfrastructure(code, path string, cause error) error {
	return storeError(status.OutcomeError, code, path, fmt.Sprintf("local snapshot store operation failed: %v", cause))
}

// StoreError is a bounded CLI-serializable store failure.
type StoreError struct{ result status.Result }

func (err *StoreError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "snapshot store operation failed"
	}
	return err.result.Diagnostics[0].Message
}

func (err *StoreError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}
