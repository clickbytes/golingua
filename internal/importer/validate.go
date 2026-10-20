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
	Manifest Manifest
	Concepts []ConceptFile
	Lexicons []LexiconFile
}

// LoadDataset reads only declared files rooted beneath root, rejects symlink
// escape, enforces aggregate/per-file bounds, then validates all references.
func LoadDataset(root string, manifest Manifest) (Dataset, error) {
	if err := manifest.Validate(); err != nil {
		return Dataset{}, err
	}
	if len(manifest.Files) > MaxDatasetFiles {
		return Dataset{}, resourceLimit("dataset_files", len(manifest.Files), MaxDatasetFiles)
	}
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return Dataset{}, infrastructure("DATA_ROOT_UNAVAILABLE", err)
	}
	rootReal, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		return Dataset{}, infrastructure("DATA_ROOT_UNAVAILABLE", err)
	}
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
		candidate := filepath.Join(rootPath, filepath.FromSlash(file.Path))
		resolved, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return Dataset{}, infrastructure("DATA_FILE_UNAVAILABLE", err)
		}
		relative, err := filepath.Rel(rootReal, resolved)
		if err != nil || relative == ".." || filepath.IsAbs(relative) || len(relative) >= 3 && relative[:3] == ".."+string(filepath.Separator) {
			return Dataset{}, invalid("DATA_PATH_ESCAPE", file.ID, "resolved data path escapes the configured root")
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return Dataset{}, infrastructure("DATA_FILE_UNAVAILABLE", err)
		}
		if !info.Mode().IsRegular() {
			return Dataset{}, invalid("INVALID_DATA_FILE", file.ID, "data file must be a regular file")
		}
		if info.Size() > MaxManifestFileBytes || total > MaxDatasetBytes-info.Size() {
			return Dataset{}, resourceLimit("dataset_bytes", MaxDatasetBytes+1, MaxDatasetBytes)
		}
		if file.SizeBytes > 0 && file.SizeBytes != info.Size() {
			return Dataset{}, invalid("DATA_SIZE_MISMATCH", file.ID, "actual file size differs from manifest")
		}
		contents, err := readBounded(resolved, MaxManifestFileBytes)
		if err != nil {
			if isLimitError(err) {
				return Dataset{}, resourceLimit("file_bytes", MaxManifestFileBytes+1, MaxManifestFileBytes)
			}
			return Dataset{}, infrastructure("DATA_FILE_READ_FAILED", err)
		}
		total += int64(len(contents))
		if file.ChecksumSHA256 != "" {
			digest := sha256.Sum256(contents)
			if hex.EncodeToString(digest[:]) != file.ChecksumSHA256 {
				return Dataset{}, invalid("DATA_CHECKSUM_MISMATCH", file.ID, "actual file checksum differs from manifest")
			}
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

func readBounded(path string, maximum int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	contents, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(contents)) > maximum {
		return nil, errSizeLimit
	}
	return contents, nil
}
