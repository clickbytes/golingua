package language

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/clickbytes/golingua/internal/ids"
)

const (
	MaxPackageBytes     = 16 << 20
	MaxModuleBytes      = 8 << 20
	MaxTotalModuleBytes = 32 << 20
)

// LoadRegistry loads each trusted local package and compiles the registry
// against the caller's already compiled Go providers.
func LoadRegistry(directories []string, providers []LanguageProvider) (*Registry, error) {
	if len(directories) == 0 {
		return nil, invalid("EMPTY_LANGUAGE_PACKAGE_PATHS", "packages", "at least one trusted package directory is required")
	}
	if len(directories) > 256 {
		return nil, resource("language_packages", len(directories), 256)
	}
	packages := make([]Package, 0, len(directories))
	for _, directory := range directories {
		pack, err := LoadPackage(directory)
		if err != nil {
			return nil, err
		}
		packages = append(packages, pack)
	}
	return BuildRegistry(packages, providers)
}

// LoadPackage reads package.json and its declared local JSON modules. The
// caller chooses a trusted directory; this function performs no network IO.
func LoadPackage(directory string) (Package, error) {
	if directory == "" {
		return Package{}, invalid("LANGUAGE_PACKAGE_PATH", "path", "trusted package directory is required")
	}
	root, err := filepath.Abs(directory)
	if err != nil {
		return Package{}, infra("LANGUAGE_PACKAGE_PATH", "path", "could not resolve package directory")
	}
	packagePath := filepath.Join(root, "package.json")
	data, err := readBoundedRegular(packagePath, MaxPackageBytes)
	if err != nil {
		if os.IsNotExist(unwrapOS(err)) {
			return Package{}, unknown("LANGUAGE_PACKAGE_NOT_FOUND", "package.json", "language package file does not exist")
		}
		if isFileError(err) {
			return Package{}, infra("LANGUAGE_PACKAGE_READ", "package.json", "could not read language package file")
		}
		return Package{}, err
	}
	var pack Package
	if err := decodeStrict(data, &pack); err != nil {
		return Package{}, invalid("INVALID_LANGUAGE_PACKAGE_JSON", "package.json", err.Error())
	}
	if err := pack.Manifest.Validate(); err != nil {
		return Package{}, err
	}
	if len(pack.Bindings) > MaxBindings {
		return Package{}, resource("bindings", len(pack.Bindings), MaxBindings)
	}
	seen := map[string]struct{}{}
	for i, binding := range pack.Bindings {
		if err := binding.Validate(); err != nil {
			return Package{}, err
		}
		key := binding.SenseID.String() + "\x00" + binding.ValencyFrameID + "\x00" + binding.FamilyID.String()
		if _, ok := seen[key]; ok {
			return Package{}, invalid("DUPLICATE_LANGUAGE_BINDING", fmt.Sprintf("bindings[%d]", i), "sense/frame/family binding is duplicated")
		}
		seen[key] = struct{}{}
	}
	totalModuleBytes := 0
	for _, module := range pack.Manifest.Modules {
		path := filepath.Join(root, module.File)
		remaining := MaxTotalModuleBytes - totalModuleBytes
		if remaining < 0 {
			remaining = 0
		}
		moduleLimit := MaxModuleBytes
		if remaining < moduleLimit {
			moduleLimit = remaining
		}
		moduleData, err := readBoundedRegular(path, moduleLimit)
		if err != nil {
			if os.IsNotExist(unwrapOS(err)) {
				if module.Optional {
					continue
				}
				return Package{}, unknown("LANGUAGE_MODULE_NOT_FOUND", module.File, "required language module does not exist")
			}
			if isFileError(err) {
				return Package{}, infra("LANGUAGE_MODULE_READ", module.File, "could not read language module")
			}
			return Package{}, err
		}
		if !json.Valid(moduleData) || rejectDuplicateJSONKeys(moduleData) != nil {
			return Package{}, invalid("INVALID_LANGUAGE_MODULE_JSON", module.File, "module must contain one valid JSON value")
		}
		totalModuleBytes += len(moduleData)
		digest := sha256.Sum256(moduleData)
		if hex.EncodeToString(digest[:]) != module.SHA256 {
			return Package{}, invalid("LANGUAGE_MODULE_DIGEST_MISMATCH", module.File, "module content does not match declared SHA-256")
		}
	}
	return clonePackage(pack), nil
}

func decodeStrict(data []byte, target any) error {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}
func scanJSONValue(decoder *json.Decoder) error {
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
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return fmt.Errorf("object is not closed")
		}
		return nil
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return fmt.Errorf("array is not closed")
		}
		return nil
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}

type fileError struct{ cause error }

func (e *fileError) Error() string { return e.cause.Error() }
func (e *fileError) Unwrap() error { return e.cause }
func unwrapOS(err error) error {
	if value, ok := err.(*fileError); ok {
		return value.cause
	}
	return err
}
func isFileError(err error) bool { _, ok := err.(*fileError); return ok }
func readBoundedRegular(path string, maximum int) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, &fileError{cause: err}
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, invalid("LANGUAGE_PACKAGE_FILE_TYPE", filepath.Base(path), "package file must be a regular non-symlink file")
	}
	if info.Size() > int64(maximum) {
		return nil, resource("package_bytes", int(info.Size()), maximum)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, &fileError{cause: err}
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(maximum)+1))
	if err != nil {
		return nil, infra("LANGUAGE_PACKAGE_READ", "file", "could not read trusted package file")
	}
	if len(data) > maximum {
		return nil, resource("package_bytes", len(data), maximum)
	}
	return data, nil
}
func clonePackage(pack Package) Package {
	result := pack
	result.Manifest.Capabilities = append([]Capability(nil), pack.Manifest.Capabilities...)
	for i := range result.Manifest.Capabilities {
		result.Manifest.Capabilities[i].EvidenceCaseIDs = append([]string(nil), pack.Manifest.Capabilities[i].EvidenceCaseIDs...)
	}
	result.Manifest.Modules = append([]Module(nil), pack.Manifest.Modules...)
	result.Bindings = append([]Binding(nil), pack.Bindings...)
	for i := range result.Bindings {
		result.Bindings[i].VariantIDs = append([]ids.ConstructionID(nil), pack.Bindings[i].VariantIDs...)
	}
	return result
}
