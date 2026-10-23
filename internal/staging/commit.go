package staging

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/importer"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

// Apply validates a private copy of candidate data, writes and fsyncs a
// complete content-addressed snapshot, validates the bytes from disk, then
// atomically replaces only the dataset's active pointer.
func (store *Store) Apply(candidate Candidate) (Preview, error) {
	if store == nil {
		return Preview{}, storeError(status.OutcomeUnknown, "STORE_UNAVAILABLE", "store", "snapshot store is unavailable")
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	stable, dataset, err := validateCandidate(candidate)
	if err != nil {
		return Preview{}, err
	}
	datasetID, digest := stable.Manifest.DatasetID, stable.Digest
	if err := ensureExistingDirectory(filepath.Join(store.root, "snapshots")); err != nil {
		return Preview{}, storeInfrastructure("SNAPSHOT_DIRECTORY_UNAVAILABLE", "snapshots", err)
	}
	datasetDir := filepath.Join(store.root, "snapshots", datasetID)
	if err := ensureDirectory(datasetDir); err != nil {
		return Preview{}, storeInfrastructure("SNAPSHOT_DIRECTORY_UNAVAILABLE", "snapshots", err)
	}
	if err := syncDirectory(filepath.Join(store.root, "snapshots")); err != nil {
		return Preview{}, storeInfrastructure("SNAPSHOT_DIRECTORY_SYNC_FAILED", "snapshots", err)
	}
	snapshotDir := filepath.Join(datasetDir, digest)
	if info, statErr := os.Lstat(snapshotDir); statErr == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return Preview{}, storeError(status.OutcomeInvalid, "UNSAFE_SNAPSHOT_PATH", "snapshot", "content-addressed snapshot path is not a real directory")
		}
		if _, err := store.loadSnapshot(datasetID, digest); err != nil {
			return Preview{}, err
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return Preview{}, storeInfrastructure("SNAPSHOT_DIRECTORY_UNAVAILABLE", "snapshot", statErr)
	} else {
		if err := store.writeSnapshot(datasetDir, snapshotDir, stable, dataset); err != nil {
			return Preview{}, err
		}
	}

	activeDigest, err := store.activeDigest(datasetID)
	if err != nil {
		return Preview{}, err
	}
	changed := activeDigest != digest
	durabilityConfirmed := true
	warning := ""
	if changed {
		durable, err := store.replacePointer(storePointer{SchemaVersion: storeSchemaVersion, DatasetID: datasetID, Digest: digest})
		if err != nil {
			return Preview{}, err
		}
		durabilityConfirmed = durable
		if !durable {
			warning = "active pointer was replaced atomically, but directory sync failed; crash durability is unconfirmed"
		}
	}
	return Preview{
		DatasetID: datasetID, Version: stable.Manifest.Version,
		Digest: digest, ActiveDigest: digest, Changed: changed, Applied: true,
		DurabilityConfirmed: durabilityConfirmed, Warning: warning, Summary: stable.Summary,
	}, nil
}

func validateCandidate(candidate Candidate) (Candidate, importer.Dataset, error) {
	if err := candidate.Manifest.Validate(); err != nil {
		return Candidate{}, importer.Dataset{}, err
	}
	if len(candidate.Files) != len(candidate.Manifest.Files) {
		return Candidate{}, importer.Dataset{}, storeError(status.OutcomeInvalid, "INVALID_CANDIDATE", "candidate.files", "candidate file count does not match manifest")
	}
	var total int64
	for index, file := range candidate.Files {
		declared := candidate.Manifest.Files[index]
		if file.ID != declared.ID || file.Path != declared.Path || file.SourceID != declared.SourceID {
			return Candidate{}, importer.Dataset{}, storeError(status.OutcomeInvalid, "INVALID_CANDIDATE", fmt.Sprintf("candidate.files[%d]", index), "candidate files must match manifest order and metadata")
		}
		if len(file.Contents) > importer.MaxManifestFileBytes {
			return Candidate{}, importer.Dataset{}, storeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", fmt.Sprintf("candidate.files[%d]", index), "candidate file exceeds the per-file byte limit")
		}
		total += int64(len(file.Contents))
		if total > importer.MaxDatasetBytes {
			return Candidate{}, importer.Dataset{}, storeError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "candidate.files", "candidate aggregate bytes exceed the dataset limit")
		}
	}
	stable := cloneCandidate(candidate)
	for _, file := range stable.Manifest.Files {
		if reservedManifestPath(file.Path) {
			return Candidate{}, importer.Dataset{}, storeError(status.OutcomeInvalid, "RESERVED_SNAPSHOT_PATH", file.Path, "dataset file collides with the snapshot manifest")
		}
	}
	digest, err := ComputeDigest(stable.Manifest, stable.Files)
	if err != nil {
		return Candidate{}, importer.Dataset{}, storeError(status.OutcomeInvalid, "INVALID_CANDIDATE", "candidate", err.Error())
	}
	if !validDigest(stable.Digest) || digest != stable.Digest {
		return Candidate{}, importer.Dataset{}, storeError(status.OutcomeInvalid, "CANDIDATE_DIGEST_MISMATCH", "candidate.digest", "candidate digest does not match its manifest and exact file bytes")
	}
	root, err := os.MkdirTemp("", "golingua-candidate-")
	if err != nil {
		return Candidate{}, importer.Dataset{}, storeInfrastructure("CANDIDATE_VALIDATION_TEMP_FAILED", "candidate", err)
	}
	defer os.RemoveAll(root)
	if err := writeCandidateFiles(root, stable.Files); err != nil {
		return Candidate{}, importer.Dataset{}, storeInfrastructure("CANDIDATE_VALIDATION_WRITE_FAILED", "candidate", err)
	}
	dataset, err := importer.LoadDataset(root, stable.Manifest)
	if err != nil {
		return Candidate{}, importer.Dataset{}, preserveImportFailure(err, status.OutcomeInvalid, "CANDIDATE_VALIDATION_FAILED", "candidate")
	}
	stable.Summary = summarizeDataset(dataset)
	return stable, dataset, nil
}

func cloneCandidate(candidate Candidate) Candidate {
	result := candidate
	result.Manifest.Sources = append([]provenance.SourceRef(nil), candidate.Manifest.Sources...)
	result.Manifest.Files = append([]importer.ManifestFile(nil), candidate.Manifest.Files...)
	result.Files = make([]File, len(candidate.Files))
	for index, file := range candidate.Files {
		result.Files[index] = File{ID: file.ID, Path: file.Path, SourceID: file.SourceID, Contents: append([]byte(nil), file.Contents...)}
	}
	return result
}

func summarizeDataset(dataset importer.Dataset) Summary {
	summary := Summary{DatasetID: dataset.Manifest.DatasetID, Version: dataset.Manifest.Version, Files: len(dataset.Manifest.Files)}
	for _, file := range dataset.Concepts {
		summary.Concepts += len(file.Concepts)
		summary.RelationTypes += len(file.Relations)
	}
	for _, file := range dataset.Lexicons {
		summary.Lexemes += len(file.Lexemes)
		summary.Senses += len(file.Senses)
		summary.Bindings += len(file.Bindings)
	}
	return summary
}

func writeCandidateFiles(root string, files []File) error {
	for _, file := range files {
		path := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := ensureParents(root, filepath.Dir(path)); err != nil {
			return err
		}
		if err := writeNewFile(path, file.Contents); err != nil {
			return err
		}
	}
	return syncDirectories(root)
}

func (store *Store) writeSnapshot(datasetDir, snapshotDir string, candidate Candidate, dataset importer.Dataset) (err error) {
	temporary, err := os.MkdirTemp(datasetDir, ".snapshot-")
	if err != nil {
		return storeInfrastructure("SNAPSHOT_PREPARE_FAILED", "snapshot", err)
	}
	defer func() {
		if temporary != "" {
			_ = os.RemoveAll(temporary)
		}
	}()
	if err := os.Chmod(temporary, 0o700); err != nil {
		return storeInfrastructure("SNAPSHOT_PREPARE_FAILED", "snapshot", err)
	}
	if err := writeCandidateFiles(temporary, candidate.Files); err != nil {
		return storeInfrastructure("SNAPSHOT_WRITE_FAILED", "snapshot", err)
	}
	// Validate the exact bytes now present in the candidate snapshot before it
	// receives its content-addressed name.
	validated, err := importer.LoadDataset(temporary, candidate.Manifest)
	if err != nil {
		return preserveImportFailure(err, status.OutcomeInvalid, "SNAPSHOT_VALIDATION_FAILED", "snapshot")
	}
	if summarizeDataset(validated) != summarizeDataset(dataset) {
		return storeError(status.OutcomeInvalid, "SNAPSHOT_VALIDATION_MISMATCH", "snapshot", "snapshot summary changed while preparing the stored files")
	}
	manifestBytes, err := json.Marshal(candidate.Manifest)
	if err != nil {
		return storeInfrastructure("SNAPSHOT_MANIFEST_WRITE_FAILED", "snapshot.manifest", err)
	}
	if err := writeNewFile(filepath.Join(temporary, "manifest.json"), manifestBytes); err != nil {
		return storeInfrastructure("SNAPSHOT_MANIFEST_WRITE_FAILED", "snapshot.manifest", err)
	}
	if err := syncDirectories(temporary); err != nil {
		return storeInfrastructure("SNAPSHOT_SYNC_FAILED", "snapshot", err)
	}
	if err := os.Rename(temporary, snapshotDir); err != nil {
		// Another process may have installed the same digest concurrently. Its
		// bytes must validate before this candidate can use that path.
		if _, statErr := os.Lstat(snapshotDir); statErr == nil {
			if _, loadErr := store.loadSnapshot(candidate.Manifest.DatasetID, candidate.Digest); loadErr == nil {
				return syncDirectory(datasetDir)
			}
		}
		return storeInfrastructure("SNAPSHOT_PUBLISH_FAILED", "snapshot", err)
	}
	temporary = ""
	if err := syncDirectory(datasetDir); err != nil {
		return storeInfrastructure("SNAPSHOT_SYNC_FAILED", "snapshots", err)
	}
	return nil
}

func preserveImportFailure(err error, fallback status.OutcomeKind, code, path string) error {
	var source interface{ Result() status.Result }
	if errors.As(err, &source) {
		result := source.Result()
		if result.Validate() == nil && result.Status != status.OutcomeOK {
			return &StoreError{result: result}
		}
	}
	return storeError(fallback, code, path, err.Error())
}

func (store *Store) replacePointer(pointer storePointer) (bool, error) {
	activeDir := filepath.Join(store.root, "active")
	if err := ensureExistingDirectory(activeDir); err != nil {
		return false, storeInfrastructure("ACTIVE_DIRECTORY_UNAVAILABLE", "active", err)
	}
	data, err := json.Marshal(pointer)
	if err != nil {
		return false, storeInfrastructure("ACTIVE_POINTER_ENCODE_FAILED", "active", err)
	}
	temporary, err := os.CreateTemp(activeDir, ".active-")
	if err != nil {
		return false, storeInfrastructure("ACTIVE_POINTER_WRITE_FAILED", "active", err)
	}
	tempPath := temporary.Name()
	defer os.Remove(tempPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return false, storeInfrastructure("ACTIVE_POINTER_WRITE_FAILED", "active", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return false, storeInfrastructure("ACTIVE_POINTER_WRITE_FAILED", "active", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return false, storeInfrastructure("ACTIVE_POINTER_SYNC_FAILED", "active", err)
	}
	if err := temporary.Close(); err != nil {
		return false, storeInfrastructure("ACTIVE_POINTER_WRITE_FAILED", "active", err)
	}
	if err := ensureReplaceablePointer(filepath.Join(activeDir, pointer.DatasetID+".json")); err != nil {
		return false, storeError(status.OutcomeInvalid, "UNSAFE_ACTIVE_POINTER", "active", err.Error())
	}
	if err := os.Rename(tempPath, filepath.Join(activeDir, pointer.DatasetID+".json")); err != nil {
		return false, storeInfrastructure("ACTIVE_POINTER_PUBLISH_FAILED", "active", err)
	}
	// The rename is the activation boundary. Syncing the containing directory
	// makes that atomic name update durable on filesystems that support fsync.
	if err := syncDirectory(activeDir); err != nil {
		return false, nil
	}
	return true, nil
}

func ensureReplaceablePointer(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("active pointer path must be a regular file")
	}
	return nil
}

func ensureDirectory(path string) error {
	if err := os.Mkdir(path, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	return ensureExistingDirectory(path)
}

func ensureExistingDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s is not a real directory", path)
	}
	return nil
}

func ensureParents(root, parent string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	parentAbs, err := filepath.Abs(parent)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(rootAbs, parentAbs)
	if err != nil || relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("data path escapes the candidate root")
	}
	current := rootAbs
	if err := ensureExistingDirectory(current); err != nil {
		return err
	}
	if relative == "." {
		return nil
	}
	for _, element := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, element)
		if err := ensureDirectory(current); err != nil {
			return err
		}
	}
	return nil
}

func ensureNoSymlinkPath(root, path string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(rootAbs, pathAbs)
	if err != nil || relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes snapshot root")
	}
	if err := ensureExistingDirectory(rootAbs); err != nil {
		return err
	}
	current := rootAbs
	if relative == "." {
		return nil
	}
	parts := strings.Split(relative, string(filepath.Separator))
	for index, element := range parts {
		current = filepath.Join(current, element)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("snapshot path contains symlink")
		}
		if index < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("snapshot path parent is not a directory")
		}
	}
	return nil
}

func writeNewFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func readRegularFile(path string, maximum int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maximum {
		return nil, fmt.Errorf("file is not regular or exceeds %d bytes", maximum)
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, fmt.Errorf("file exceeds %d bytes", maximum)
	}
	return data, nil
}

func syncDirectories(root string) error {
	var directories []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("snapshot tree contains a symlink")
		}
		if entry.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(directories, func(i, j int) bool { return len(directories[i]) > len(directories[j]) })
	for _, directory := range directories {
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
