// Package runtime loads an explicitly selected, validated runtime snapshot.
// Language providers and stage adapters remain compiled Go dependencies; data
// files can configure them but cannot supply executable code.
package runtime

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	ConfigSchemaVersion = 1
	MaxConfigBytes      = 64 << 10
	MaxPackagePaths     = 32
)

// Config selects one already active CORE/lexicon snapshot and its trusted
// language package directories. Relative paths are resolved from the config
// file's directory. Provider and analyzer implementations are supplied by the
// compiled Go application, never loaded from data.
type Config struct {
	SchemaVersion       int               `json:"schema_version"`
	DatasetStore        string            `json:"dataset_store"`
	DatasetID           string            `json:"dataset_id"`
	LanguagePackageDirs []string          `json:"language_package_dirs"`
	SemanticProfile     string            `json:"semantic_profile,omitempty"`
	DSEProfile          *DSEProfileConfig `json:"dse_profile,omitempty"`
}

// Options supplies trusted compiled providers and the engine stage adapters
// that are available in this application build.
type Options struct {
	Providers       []language.LanguageProvider
	LanguageFactory LanguageFactory
	Analyzers       map[features.LanguageID]engine.SourceAnalyzer
	Realizers       map[features.LanguageID]engine.TargetRealizer
	StageFactory    StageFactory
	Budget          engine.Budget
}

// LanguageFactory compiles data packages and their matching trusted Go
// providers against the exact selected snapshot before registry validation.
type LanguageFactory func(*pack.Pack) ([]language.Package, []language.LanguageProvider, error)

// StageAdapters are compiled, language-specific engine stages constructed
// against the exact pack and verified language registry selected by Config.
type StageAdapters struct {
	Analyzers map[features.LanguageID]engine.SourceAnalyzer
	Realizers map[features.LanguageID]engine.TargetRealizer
}

// StageFactory runs only after the configured data and language packages have
// been verified. It is supplied by the application binary, never by JSON.
type StageFactory func(*pack.Pack, *language.Registry) (StageAdapters, error)

// LoadConfigFile reads a bounded regular JSON config. It does not create or
// activate datasets and does not load code from configured paths.
func LoadConfigFile(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, runtimeError(status.OutcomeInvalid, "RUNTIME_CONFIG_PATH_REQUIRED", "runtime_config", "runtime config path is required")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return Config{}, runtimeError(status.OutcomeUnknown, "RUNTIME_CONFIG_UNAVAILABLE", "runtime_config", "runtime config file could not be accessed")
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > MaxConfigBytes {
		return Config{}, runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_CONFIG_FILE", "runtime_config", "runtime config must be a non-empty regular file within the byte limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, runtimeError(status.OutcomeUnknown, "RUNTIME_CONFIG_UNAVAILABLE", "runtime_config", "runtime config file could not be read")
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil || !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return Config{}, runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_CONFIG_FILE", "runtime_config", "runtime config changed while it was being opened")
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxConfigBytes+1))
	if err != nil {
		return Config{}, runtimeError(status.OutcomeUnknown, "RUNTIME_CONFIG_UNAVAILABLE", "runtime_config", "runtime config file could not be read")
	}
	if len(data) == 0 || len(data) > MaxConfigBytes {
		return Config{}, runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_CONFIG_FILE", "runtime_config", "runtime config must be non-empty and within the byte limit")
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return Config{}, runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_CONFIG_JSON", "runtime_config", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_CONFIG_JSON", "runtime_config", err.Error())
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Config{}, runtimeError(status.OutcomeInvalid, "TRAILING_RUNTIME_CONFIG_DATA", "runtime_config", "runtime config must contain exactly one JSON value")
	}
	if err := config.validateFields(); err != nil {
		return Config{}, err
	}
	base, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return Config{}, runtimeError(status.OutcomeUnknown, "RUNTIME_CONFIG_UNAVAILABLE", "runtime_config", "runtime config directory could not be resolved")
	}
	config.DatasetStore = resolvePath(base, config.DatasetStore)
	for index, directory := range config.LanguagePackageDirs {
		config.LanguagePackageDirs[index] = resolvePath(base, directory)
	}
	if config.DSEProfile != nil {
		config.DSEProfile.resolvePaths(base)
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

// Validate rejects unbounded or ambiguous runtime selections. Language
// packages are optional for callers that only need the selected CORE snapshot.
func (config Config) Validate() error {
	if err := config.validateFields(); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(config.LanguagePackageDirs))
	for index, directory := range config.LanguagePackageDirs {
		canonical, err := filepath.Abs(directory)
		if err != nil {
			return runtimeError(status.OutcomeInvalid, "INVALID_LANGUAGE_PACKAGE_PATH", fmt.Sprintf("language_package_dirs[%d]", index), "language package path could not be resolved")
		}
		if _, exists := seen[canonical]; exists {
			return runtimeError(status.OutcomeInvalid, "DUPLICATE_LANGUAGE_PACKAGE_PATH", fmt.Sprintf("language_package_dirs[%d]", index), "language package path is duplicated")
		}
		seen[canonical] = struct{}{}
	}
	return nil
}

func (config Config) validateFields() error {
	if config.SchemaVersion != ConfigSchemaVersion {
		return runtimeError(status.OutcomeInvalid, "RUNTIME_CONFIG_SCHEMA_VERSION", "schema_version", "unsupported runtime config schema version")
	}
	if !safeIdentifier(config.DatasetID) || strings.TrimSpace(config.DatasetStore) != config.DatasetStore || config.DatasetStore == "" {
		return runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_DATASET", "dataset", "runtime config requires a dataset ID and store path")
	}
	if len(config.LanguagePackageDirs) > MaxPackagePaths {
		return runtimeError(status.OutcomeInvalid, "INVALID_LANGUAGE_PACKAGE_PATHS", "language_package_dirs", "runtime config contains too many language package paths")
	}
	if config.SemanticProfile != "" && config.SemanticProfile != "meaning-v1" && config.SemanticProfile != "dse" {
		return runtimeError(status.OutcomeInvalid, "UNSUPPORTED_SEMANTIC_PROFILE", "semantic_profile", "semantic_profile must be meaning-v1 or dse")
	}
	if (config.SemanticProfile == "dse") != (config.DSEProfile != nil) {
		return runtimeError(status.OutcomeInvalid, "DSE_PROFILE_CONFIG_MISMATCH", "dse_profile", "DSE activation requires exactly one explicit dse_profile configuration")
	}
	if config.DSEProfile != nil {
		if err := config.DSEProfile.validate(); err != nil {
			return err
		}
	}
	for index, directory := range config.LanguagePackageDirs {
		if strings.TrimSpace(directory) == "" || strings.TrimSpace(directory) != directory {
			return runtimeError(status.OutcomeInvalid, "INVALID_LANGUAGE_PACKAGE_PATH", fmt.Sprintf("language_package_dirs[%d]", index), "language package path must be non-empty and trimmed")
		}
	}
	return nil
}

// LoadEngineFile compiles the selected active pack and verified package
// registry, then creates an immutable request engine. Missing stage adapters
// are valid configuration and remain request-level unsupported outcomes.
func LoadEngineFile(path string, options Options) (*engine.Engine, error) {
	config, err := LoadConfigFile(path)
	if err != nil {
		return nil, err
	}
	snapshot, err := loadActivePack(config)
	if err != nil {
		return nil, err
	}
	if len(config.LanguagePackageDirs) == 0 && options.LanguageFactory == nil {
		return nil, runtimeError(status.OutcomeInvalid, "LANGUAGE_PACKAGES_REQUIRED", "language_package_dirs", "translation runtime requires at least one language package")
	}
	var registry *language.Registry
	if options.LanguageFactory != nil {
		if len(config.LanguagePackageDirs) != 0 {
			return nil, runtimeError(status.OutcomeInvalid, "AMBIGUOUS_LANGUAGE_PACKAGE_SOURCE", "language_package_dirs", "runtime config cannot combine compiled language packages with external language package directories")
		}
		packages, providers, factoryErr := options.LanguageFactory(snapshot)
		if factoryErr != nil {
			return nil, preserveError(factoryErr, "RUNTIME_LANGUAGE_FACTORY_FAILED", "language_factory")
		}
		registry, err = language.BuildRegistry(packages, providers)
	} else {
		registry, err = language.LoadRegistry(config.LanguagePackageDirs, options.Providers)
	}
	if err != nil {
		return nil, preserveError(err, "RUNTIME_LANGUAGE_REGISTRY_INVALID", "language_package_dirs")
	}
	if options.StageFactory != nil {
		stages, err := options.StageFactory(snapshot, registry)
		if err != nil {
			return nil, preserveError(err, "RUNTIME_STAGE_FACTORY_FAILED", "stage_factory")
		}
		if err := mergeStageAdapters(&options, stages); err != nil {
			return nil, err
		}
	}
	analyzers, realizers := providerStages(registry, options)
	var dseProfile *engine.DSEProfile
	if config.SemanticProfile == "dse" {
		dseProfile, err = loadDSEProfile(*config.DSEProfile, snapshot, analyzers)
		if err != nil {
			return nil, err
		}
	}
	return engine.New(engine.Config{
		Core: snapshot.CoreGraph(), LanguageRegistry: registry, PackSnapshot: snapshot,
		Analyzers: analyzers, Realizers: realizers, DSEProfile: dseProfile, Budget: options.Budget,
	})
}

// LoadCoreFile loads only the active CORE snapshot selected by a runtime
// configuration. Querying request facts does not require language packages.
func LoadCoreFile(path string) (*core.ConceptGraph, error) {
	config, err := LoadConfigFile(path)
	if err != nil {
		return nil, err
	}
	snapshot, err := loadActivePack(config)
	if err != nil {
		return nil, err
	}
	return snapshot.CoreGraph(), nil
}

func loadActivePack(config Config) (*pack.Pack, error) {
	storeInfo, err := os.Stat(config.DatasetStore)
	if err != nil {
		return nil, runtimeError(status.OutcomeUnknown, "RUNTIME_DATASET_STORE_UNAVAILABLE", "dataset_store", "configured dataset store does not exist")
	}
	if !storeInfo.IsDir() {
		return nil, runtimeError(status.OutcomeInvalid, "INVALID_RUNTIME_DATASET_STORE", "dataset_store", "configured dataset store must be a directory")
	}
	store, err := staging.OpenStore(config.DatasetStore)
	if err != nil {
		return nil, preserveError(err, "RUNTIME_DATASET_STORE_UNAVAILABLE", "dataset_store")
	}
	snapshot, err := pack.LoadActive(store, config.DatasetID)
	if err != nil {
		return nil, preserveError(err, "RUNTIME_DATASET_UNAVAILABLE", "dataset_id")
	}
	return snapshot, nil
}

func mergeStageAdapters(options *Options, stages StageAdapters) error {
	analyzers := make(map[features.LanguageID]engine.SourceAnalyzer, len(options.Analyzers)+len(stages.Analyzers))
	for id, adapter := range options.Analyzers {
		analyzers[id] = adapter
	}
	for id, adapter := range stages.Analyzers {
		if _, exists := analyzers[id]; exists {
			return runtimeError(status.OutcomeInvalid, "DUPLICATE_SOURCE_STAGE", "stage_factory.analyzers", "source stage is already registered for this language")
		}
		analyzers[id] = adapter
	}
	realizers := make(map[features.LanguageID]engine.TargetRealizer, len(options.Realizers)+len(stages.Realizers))
	for id, adapter := range options.Realizers {
		realizers[id] = adapter
	}
	for id, adapter := range stages.Realizers {
		if _, exists := realizers[id]; exists {
			return runtimeError(status.OutcomeInvalid, "DUPLICATE_TARGET_STAGE", "stage_factory.realizers", "target stage is already registered for this language")
		}
		realizers[id] = adapter
	}
	options.Analyzers = analyzers
	options.Realizers = realizers
	return nil
}

// providerStages connects compiled language providers that implement the
// engine's full source/target stage contracts. Language data alone cannot
// install either interface. Explicit application adapters in Options take
// precedence for the same language.
func providerStages(registry *language.Registry, options Options) (map[features.LanguageID]engine.SourceAnalyzer, map[features.LanguageID]engine.TargetRealizer) {
	analyzers := make(map[features.LanguageID]engine.SourceAnalyzer, len(options.Analyzers))
	for id, adapter := range options.Analyzers {
		analyzers[id] = adapter
	}
	realizers := make(map[features.LanguageID]engine.TargetRealizer, len(options.Realizers))
	for id, adapter := range options.Realizers {
		realizers[id] = adapter
	}
	if registry == nil {
		return analyzers, realizers
	}
	for _, registered := range registry.Languages() {
		id := registered.Manifest.Language
		provider, exists := registry.Provider(id)
		if !exists {
			continue
		}
		if _, configured := analyzers[id]; !configured {
			if adapter, implements := provider.(engine.SourceAnalyzer); implements {
				analyzers[id] = adapter
			}
		}
		if _, configured := realizers[id]; !configured {
			if adapter, implements := provider.(engine.TargetRealizer); implements {
				realizers[id] = adapter
			}
		}
	}
	return analyzers, realizers
}

func resolvePath(base, value string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(base, value))
}

func safeIdentifier(value string) bool {
	if value == "" || len(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for index, character := range value {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || (index > 0 && (character == '-' || character == '_' || character == '.'))) {
			return false
		}
	}
	return true
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var visit func() error
	visit = func() error {
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
					return errors.New("object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate object key %q", key)
				}
				seen[key] = struct{}{}
				if err := visit(); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim('}') {
				return errors.New("object is not closed")
			}
		case '[':
			for decoder.More() {
				if err := visit(); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				return errors.New("array is not closed")
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
		}
		return nil
	}
	if err := visit(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

type runtimeErrorValue struct{ result status.Result }

func (failure *runtimeErrorValue) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "runtime configuration error"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *runtimeErrorValue) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func runtimeError(kind status.OutcomeKind, code, path, message string) error {
	result, err := status.NewResult(kind, "", []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}, nil)
	if err != nil {
		return &runtimeErrorValue{result: status.Result{Status: status.OutcomeError}}
	}
	return &runtimeErrorValue{result: result}
}

func preserveError(err error, code, path string) error {
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		result := typed.Result()
		if result.Validate() == nil && result.Status != status.OutcomeOK {
			return &runtimeErrorValue{result: result}
		}
	}
	return runtimeError(status.OutcomeUnknown, code, path, "configured runtime resource could not be loaded")
}
