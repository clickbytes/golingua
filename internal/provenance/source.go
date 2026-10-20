// Package provenance defines source references for manually authored or
// optionally imported data. It does not fetch, approve, or activate sources.
package provenance

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MaxSourceIDBytes       = 128
	MaxSourceLocationBytes = 2048
	MaxSourceVersionBytes  = 128
	MaxSourceLicenseBytes  = 256
)

var ErrInvalidSourceRef = errors.New("invalid source reference")

// SourceRef records where data came from and the version and license statement
// applicable to that source. Exactly one of Path or URI must be set. A valid
// reference is attribution metadata only; it does not approve redistribution.
type SourceRef struct {
	ID      string `json:"id"`
	Path    string `json:"path,omitempty"`
	URI     string `json:"uri,omitempty"`
	Version string `json:"version"`
	License string `json:"license"`
}

// Validate checks the reference's structure and bounded fields. It does not
// access the filesystem or network and does not interpret license terms.
func (source SourceRef) Validate() error {
	if !validID(source.ID) {
		return fmt.Errorf("%w: id must be a bounded stable identifier", ErrInvalidSourceRef)
	}
	if (source.Path == "") == (source.URI == "") {
		return fmt.Errorf("%w: exactly one of path or uri is required", ErrInvalidSourceRef)
	}
	if source.Path != "" {
		if err := validatePath(source.Path); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidSourceRef, err)
		}
	}
	if source.URI != "" {
		if err := validateURI(source.URI); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidSourceRef, err)
		}
	}
	if !validMetadata(source.Version, MaxSourceVersionBytes) {
		return fmt.Errorf("%w: version is empty, invalid, or too long", ErrInvalidSourceRef)
	}
	if !validMetadata(source.License, MaxSourceLicenseBytes) {
		return fmt.Errorf("%w: license is empty, invalid, or too long", ErrInvalidSourceRef)
	}
	return nil
}

func validID(value string) bool {
	if len(value) == 0 || len(value) > MaxSourceIDBytes {
		return false
	}
	for i, r := range value {
		if i == 0 && !asciiAlphaNumeric(r) {
			return false
		}
		if !asciiAlphaNumeric(r) && r != '_' && r != '-' && r != '.' {
			return false
		}
	}
	return true
}

func asciiAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func validatePath(value string) error {
	if !validMetadata(value, MaxSourceLocationBytes) {
		return errors.New("path is empty, invalid, or too long")
	}
	if strings.ContainsRune(value, '\\') || path.IsAbs(value) || isWindowsAbsolute(value) {
		return errors.New("path must be relative and use slash separators")
	}
	clean := path.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != value {
		return errors.New("path must be normalized and stay within the data root")
	}
	return nil
}

func validateURI(value string) error {
	if !validMetadata(value, MaxSourceLocationBytes) {
		return errors.New("uri is empty, invalid, or too long")
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" {
		return errors.New("uri must be absolute and include a scheme")
	}
	if parsed.User != nil {
		return errors.New("uri must not contain user credentials")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		if parsed.Host == "" {
			return errors.New("http(s) uri must include a host")
		}
	case "urn", "doi", "ark":
		if parsed.Opaque == "" && parsed.Path == "" {
			return errors.New("uri identifier must not be empty")
		}
	default:
		return errors.New("uri scheme must be http, https, urn, doi, or ark")
	}
	return nil
}

func isWindowsAbsolute(value string) bool {
	return len(value) >= 2 && ((value[0] >= 'a' && value[0] <= 'z') || (value[0] >= 'A' && value[0] <= 'Z')) && value[1] == ':'
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
