// Package importer defines bounded metadata envelopes for optional language
// data imports. It never fetches sources or executes imported content.
package importer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	ManifestSchemaVersion   = 1
	MaxManifestBytes        = 1 << 20
	MaxSources              = 256
	MaxFiles                = 100_000
	MaxManifestIDBytes      = 128
	MaxManifestVersionBytes = 128
	MaxManifestFileBytes    = 32 << 20
)

var errDuplicateJSONKey = errors.New("duplicate JSON object key")

type Manifest struct {
	SchemaVersion int                    `json:"schema_version"`
	DatasetID     string                 `json:"dataset_id"`
	Version       string                 `json:"version"`
	Sources       []provenance.SourceRef `json:"sources"`
	Files         []ManifestFile         `json:"files"`
}

type ManifestFile struct {
	ID             string `json:"id"`
	Path           string `json:"path"`
	SourceID       string `json:"source_id"`
	ChecksumSHA256 string `json:"checksum_sha256,omitempty"`
	SizeBytes      int64  `json:"size_bytes,omitempty"`
}

// DecodeManifest validates a single bounded JSON manifest. It performs no
// filesystem or network access; callers must separately read files only after
// resolving them beneath their configured data root.
func DecodeManifest(data []byte) (Manifest, error) {
	if len(data) == 0 {
		return Manifest{}, invalid("EMPTY_MANIFEST", "manifest", "manifest must not be empty")
	}
	if len(data) > MaxManifestBytes {
		return Manifest{}, resourceLimit("manifest_bytes", len(data), MaxManifestBytes)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		if errors.Is(err, errDuplicateJSONKey) {
			return Manifest{}, invalid("DUPLICATE_JSON_KEY", "manifest", err.Error())
		}
		return Manifest{}, invalid("INVALID_MANIFEST_JSON", "manifest", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, invalid("INVALID_MANIFEST_JSON", "manifest", err.Error())
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return Manifest{}, invalid("TRAILING_MANIFEST_DATA", "manifest", "manifest must contain exactly one JSON value")
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}
	return cloneManifest(manifest), nil
}

func (manifest Manifest) Validate() error {
	if manifest.SchemaVersion != ManifestSchemaVersion {
		return invalid("MANIFEST_SCHEMA_VERSION", "schema_version", "unsupported importer manifest schema")
	}
	if !validIdentifier(manifest.DatasetID, MaxManifestIDBytes) || !validMetadata(manifest.Version, MaxManifestVersionBytes) {
		return invalid("INVALID_MANIFEST_IDENTITY", "manifest", "manifest requires bounded dataset ID and version")
	}
	if len(manifest.Sources) == 0 {
		return invalid("EMPTY_MANIFEST_SOURCES", "sources", "manifest requires at least one source")
	}
	if len(manifest.Sources) > MaxSources {
		return resourceLimit("sources", len(manifest.Sources), MaxSources)
	}
	if len(manifest.Files) == 0 {
		return invalid("EMPTY_MANIFEST_FILES", "files", "manifest requires at least one file")
	}
	if len(manifest.Files) > MaxFiles {
		return resourceLimit("files", len(manifest.Files), MaxFiles)
	}
	sources := make(map[string]struct{}, len(manifest.Sources))
	for i, source := range manifest.Sources {
		if err := source.Validate(); err != nil {
			return invalid("INVALID_MANIFEST_SOURCE", fmt.Sprintf("sources[%d]", i), err.Error())
		}
		if _, exists := sources[source.ID]; exists {
			return invalid("DUPLICATE_MANIFEST_SOURCE", fmt.Sprintf("sources[%d].id", i), "source ID is duplicated")
		}
		sources[source.ID] = struct{}{}
	}
	fileIDs, paths := map[string]struct{}{}, map[string]struct{}{}
	for i, file := range manifest.Files {
		path := fmt.Sprintf("files[%d]", i)
		if !validIdentifier(file.ID, MaxManifestIDBytes) {
			return invalid("INVALID_MANIFEST_FILE_ID", path+".id", "file ID must be bounded ASCII identifier")
		}
		if err := provenance.ValidateRelativeDataPath(file.Path); err != nil {
			return invalid("INVALID_MANIFEST_FILE_PATH", path+".path", err.Error())
		}
		if _, exists := fileIDs[file.ID]; exists {
			return invalid("DUPLICATE_MANIFEST_FILE_ID", path+".id", "manifest file ID is duplicated")
		}
		if _, exists := paths[file.Path]; exists {
			return invalid("DUPLICATE_MANIFEST_FILE_PATH", path+".path", "manifest file path is duplicated")
		}
		if _, exists := sources[file.SourceID]; !exists {
			return unknown("UNKNOWN_MANIFEST_SOURCE", path+".source_id", "file refers to no source in this manifest")
		}
		if file.ChecksumSHA256 != "" && !provenance.ValidSHA256(file.ChecksumSHA256) {
			return invalid("INVALID_MANIFEST_CHECKSUM", path+".checksum_sha256", "checksum must be lowercase hexadecimal SHA-256")
		}
		if file.SizeBytes < 0 {
			return invalid("INVALID_MANIFEST_FILE_SIZE", path+".size_bytes", "declared file size must not be negative")
		}
		if file.SizeBytes > MaxManifestFileBytes {
			return resourceLimit("file_bytes", int(file.SizeBytes), MaxManifestFileBytes)
		}
		fileIDs[file.ID], paths[file.Path] = struct{}{}, struct{}{}
	}
	return nil
}

func validIdentifier(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum {
		return false
	}
	for i, r := range value {
		if i == 0 && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

func validMetadata(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func cloneManifest(source Manifest) Manifest {
	result := source
	result.Sources = append([]provenance.SourceRef(nil), source.Sources...)
	result.Files = append([]ManifestFile(nil), source.Files...)
	return result
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() error
	walk = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]struct{}{}
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
					return fmt.Errorf("%w %q", errDuplicateJSONKey, key)
				}
				seen[key] = struct{}{}
				if err := walk(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := walk(); err != nil {
					return err
				}
			}
			_, err := decoder.Token()
			return err
		default:
			return fmt.Errorf("unexpected JSON delimiter")
		}
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON data")
	}
	return nil
}

func invalid(code, path, message string) error {
	return newManifestError(status.OutcomeInvalid, code, path, message)
}
func unknown(code, path, message string) error {
	return newManifestError(status.OutcomeUnknown, code, path, message)
}

func resourceLimit(resource string, value, maximum int) error {
	return newManifestError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+resource, fmt.Sprintf("manifest %s limit exceeded: %d > %d", resource, value, maximum))
}

func newManifestError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: boundedDiagnosticText(code, status.MaxDiagnosticCodeBytes), Path: boundedDiagnosticText(path, status.MaxDiagnosticPathBytes), Message: boundedDiagnosticText(message, status.MaxDiagnosticMessageBytes), Severity: status.SeverityError}
	return &ManifestError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

func boundedDiagnosticText(value string, maximum int) string {
	if !utf8.ValidString(value) {
		value = strings.ToValidUTF8(value, "�")
	}
	if len(value) <= maximum {
		return value
	}
	value = value[:maximum]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

type ManifestError struct {
	result status.Result
}

func (err *ManifestError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "manifest error"
	}
	return err.result.Diagnostics[0].Message
}

func (err *ManifestError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}
