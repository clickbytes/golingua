// Package staging builds validated, immutable-by-convention candidate data for
// review before a caller stores or activates a snapshot.
package staging

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"

	"github.com/clickbytes/golingua/internal/importer"
)

// Candidate is a fully validated source dataset together with the exact input
// bytes read during validation. Digest identifies both its manifest metadata
// and file contents; it is not an approval or activation token.
type Candidate struct {
	Manifest importer.Manifest
	Files    []File
	Digest   string
	Summary  Summary
}

// File is one manifest-declared data file. Contents are the bytes captured by
// importer.LoadDataset, not a later re-read of the source path.
type File struct {
	ID       string
	Path     string
	SourceID string
	Contents []byte
}

// Summary contains aggregate counts from the validated CORE and lexicon data.
type Summary struct {
	DatasetID string
	Version   string
	Files     int
	Concepts  int
	Lexemes   int
	Senses    int
	Bindings  int
}

var ErrInvalidCandidateFiles = errors.New("candidate files do not match manifest")

// BuildCandidate validates the manifest and its declared CORE/lexicon files,
// preserving the exact bytes read by the importer for later preview or store
// operations. It performs no writes, activation, or license-term evaluation.
func BuildCandidate(dataRoot string, manifest importer.Manifest) (Candidate, error) {
	dataset, err := importer.LoadDatasetWithFiles(dataRoot, manifest)
	if err != nil {
		return Candidate{}, err
	}

	candidate := Candidate{
		Manifest: dataset.Manifest,
		Files:    make([]File, 0, len(dataset.LoadedFiles)),
		Summary: Summary{
			DatasetID: dataset.Manifest.DatasetID,
			Version:   dataset.Manifest.Version,
			Files:     len(dataset.LoadedFiles),
		},
	}
	for _, conceptFile := range dataset.Concepts {
		candidate.Summary.Concepts += len(conceptFile.Concepts)
	}
	for _, lexiconFile := range dataset.Lexicons {
		candidate.Summary.Lexemes += len(lexiconFile.Lexemes)
		candidate.Summary.Senses += len(lexiconFile.Senses)
		candidate.Summary.Bindings += len(lexiconFile.Bindings)
	}
	for _, loaded := range dataset.LoadedFiles {
		candidate.Files = append(candidate.Files, File{
			ID:       loaded.ManifestFile.ID,
			Path:     loaded.ManifestFile.Path,
			SourceID: loaded.ManifestFile.SourceID,
			Contents: loaded.Contents,
		})
	}
	candidate.Digest, err = ComputeDigest(candidate.Manifest, candidate.Files)
	if err != nil {
		return Candidate{}, err
	}
	return candidate, nil
}

// ComputeDigest returns a stable SHA-256 digest for a manifest and its exact
// file bytes. Files must match the manifest's declared order and metadata.
func ComputeDigest(manifest importer.Manifest, files []File) (string, error) {
	if err := manifest.Validate(); err != nil {
		return "", err
	}
	if len(files) != len(manifest.Files) {
		return "", fmt.Errorf("%w: got %d files for %d manifest entries", ErrInvalidCandidateFiles, len(files), len(manifest.Files))
	}
	var total int64
	for index, declared := range manifest.Files {
		file := files[index]
		if file.ID != declared.ID || file.Path != declared.Path || file.SourceID != declared.SourceID {
			return "", fmt.Errorf("%w: files[%d] is missing, duplicated, or out of manifest order", ErrInvalidCandidateFiles, index)
		}
		if len(file.Contents) > importer.MaxManifestFileBytes {
			return "", fmt.Errorf("%w: files[%d] exceeds the per-file byte limit", ErrInvalidCandidateFiles, index)
		}
		total += int64(len(file.Contents))
		if total > importer.MaxDatasetBytes {
			return "", fmt.Errorf("%w: aggregate file bytes exceed the dataset limit", ErrInvalidCandidateFiles)
		}
		if declared.SizeBytes > 0 && declared.SizeBytes != int64(len(file.Contents)) {
			return "", fmt.Errorf("%w: files[%d] size differs from its manifest entry", ErrInvalidCandidateFiles, index)
		}
		if declared.ChecksumSHA256 != "" {
			digest := sha256.Sum256(file.Contents)
			if hex.EncodeToString(digest[:]) != declared.ChecksumSHA256 {
				return "", fmt.Errorf("%w: files[%d] checksum differs from its manifest entry", ErrInvalidCandidateFiles, index)
			}
		}
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("encode candidate manifest: %w", err)
	}
	digest := sha256.New()
	writeBytes(digest, manifestBytes)
	writeUint64(digest, uint64(len(files)))
	for _, file := range files {
		writeString(digest, file.ID)
		writeString(digest, file.Path)
		writeString(digest, file.SourceID)
		writeBytes(digest, file.Contents)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func writeString(digest hash.Hash, value string) {
	writeBytes(digest, []byte(value))
}

func writeBytes(digest hash.Hash, value []byte) {
	writeUint64(digest, uint64(len(value)))
	_, _ = digest.Write(value)
}

func writeUint64(digest hash.Hash, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = digest.Write(encoded[:])
}
