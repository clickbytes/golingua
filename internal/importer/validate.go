package importer

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	MaxDatasetFiles = 100_000
	MaxDatasetBytes = 64 << 20
)

// Dataset is a fully decoded and validated offline source dataset.
type Dataset struct {
	Manifest    Manifest
	Concepts    []ConceptFile
	Lexicons    []LexiconFile
	LoadedFiles []LoadedFile
}

// LoadedFile contains the exact bytes read and validated from one declared
// manifest file. Consumers can stage or persist the validated input without
// reopening its path and introducing a time-of-check/time-of-use race.
type LoadedFile struct {
	ManifestFile ManifestFile
	Contents     []byte
}

// LoadDataset reads only declared files rooted beneath root, rejects symlink
// escape, enforces aggregate/per-file bounds, then validates all references.
func LoadDataset(root string, manifest Manifest) (Dataset, error) {
	return loadDataset(root, manifest, false)
}

// LoadDatasetWithFiles is LoadDataset with exact validated file bytes retained
// for callers that need to stage the input without reopening source paths.
func LoadDatasetWithFiles(root string, manifest Manifest) (Dataset, error) {
	return loadDataset(root, manifest, true)
}

func loadDataset(root string, manifest Manifest, retainFiles bool) (Dataset, error) {
	if err := manifest.Validate(); err != nil {
		return Dataset{}, err
	}
	if len(manifest.Files) > MaxDatasetFiles {
		return Dataset{}, resourceLimit("dataset_files", len(manifest.Files), MaxDatasetFiles)
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return Dataset{}, infrastructure("DATA_ROOT_UNAVAILABLE", err)
	}
	defer rootHandle.Close()
	conceptBytes, lexiconBytes := []byte(nil), []byte(nil)
	conceptFiles, lexiconFiles := 0, 0
	result := Dataset{Manifest: cloneManifest(manifest)}
	var total int64
	for _, file := range manifest.Files {
		if file.SizeBytes > 0 && file.SizeBytes > MaxManifestFileBytes {
			return Dataset{}, resourceLimit("file_bytes", int(file.SizeBytes), MaxManifestFileBytes)
		}
		if err := provenance.ValidateRelativeDataPath(file.Path); err != nil {
			return Dataset{}, invalid("INVALID_DATA_PATH", file.ID, err.Error())
		}
		contents, err := readManifestFile(rootHandle, file, total)
		if err != nil {
			return Dataset{}, err
		}
		total += int64(len(contents))
		if file.ChecksumSHA256 != "" {
			digest := sha256.Sum256(contents)
			if hex.EncodeToString(digest[:]) != file.ChecksumSHA256 {
				return Dataset{}, invalid("DATA_CHECKSUM_MISMATCH", file.ID, "actual file checksum differs from manifest")
			}
		}
		if retainFiles {
			result.LoadedFiles = append(result.LoadedFiles, LoadedFile{
				ManifestFile: file,
				Contents:     contents,
			})
		}
		if file.ID == "concepts" || filepath.Ext(file.Path) == ".concepts.json" {
			conceptFiles++
			if conceptFiles > 1 {
				return Dataset{}, invalid("DUPLICATE_CONCEPT_FILE", file.ID, "dataset may contain only one concepts definition file")
			}
			conceptBytes = contents
		} else if file.ID == "lexicon" || filepath.Ext(file.Path) == ".lexicon.json" {
			lexiconFiles++
			if lexiconFiles > 1 {
				return Dataset{}, invalid("DUPLICATE_LEXICON_FILE", file.ID, "dataset may contain only one lexicon definition file")
			}
			lexiconBytes = contents
		} else {
			return Dataset{}, invalid("UNKNOWN_DATA_FILE_KIND", file.ID, "file ID or suffix must identify a concept or lexicon definition file")
		}
	}
	if conceptBytes == nil {
		return Dataset{}, invalid("MISSING_CONCEPT_FILE", "files", "dataset must include one concepts definition file")
	}
	concepts, err := DecodeConceptFile(conceptBytes)
	if err != nil {
		return Dataset{}, err
	}
	graph, err := core.CompileConceptGraph(concepts.Concepts, concepts.Relations)
	if err != nil {
		return Dataset{}, invalid("INVALID_CONCEPT_FILE", "concepts", err.Error())
	}
	result.Concepts = []ConceptFile{concepts}
	if lexiconBytes != nil {
		lexiconFile, err := DecodeLexiconFile(lexiconBytes, graph)
		if err != nil {
			return Dataset{}, err
		}
		result.Lexicons = []LexiconFile{lexiconFile}
	}
	if len(result.Lexicons) == 0 {
		return Dataset{}, invalid("MISSING_LEXICON_FILE", "files", "dataset must include one lexicon definition file")
	}
	return result, nil
}

// readManifestFile resolves and opens the declared path inside one open root.
// Stat and read use the same descriptor, so replacing a path after validation
// cannot redirect the read to a different file.
func readManifestFile(root *os.Root, declared ManifestFile, total int64) ([]byte, error) {
	file, err := root.Open(filepath.FromSlash(declared.Path))
	if err != nil {
		return nil, infrastructure("DATA_FILE_UNAVAILABLE", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, infrastructure("DATA_FILE_UNAVAILABLE", err)
	}
	if !info.Mode().IsRegular() {
		return nil, invalid("INVALID_DATA_FILE", declared.ID, "data file must be a regular file")
	}
	if info.Size() > MaxManifestFileBytes || total > MaxDatasetBytes-info.Size() {
		return nil, resourceLimit("dataset_bytes", MaxDatasetBytes+1, MaxDatasetBytes)
	}
	if declared.SizeBytes > 0 && declared.SizeBytes != info.Size() {
		return nil, invalid("DATA_SIZE_MISMATCH", declared.ID, "actual file size differs from manifest")
	}
	contents, err := io.ReadAll(io.LimitReader(file, MaxManifestFileBytes+1))
	if err != nil {
		return nil, infrastructure("DATA_FILE_READ_FAILED", err)
	}
	if len(contents) > MaxManifestFileBytes {
		return nil, resourceLimit("file_bytes", len(contents), MaxManifestFileBytes)
	}
	if total > MaxDatasetBytes-int64(len(contents)) {
		return nil, resourceLimit("dataset_bytes", MaxDatasetBytes+1, MaxDatasetBytes)
	}
	if declared.SizeBytes > 0 && declared.SizeBytes != int64(len(contents)) {
		return nil, invalid("DATA_SIZE_MISMATCH", declared.ID, "read file size differs from manifest")
	}
	return contents, nil
}
