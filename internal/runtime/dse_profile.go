package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/metamodel"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxDSERegistryPackages  = 16
	MaxDSEProfileLanguages  = 2
	maxDSERegistryFileBytes = 1 << 20
)

// DSEResourcePin identifies one bounded runtime resource. Fingerprints for
// language packages and plans are SHA-256 of the exact file bytes; metamodel
// packages use their validated semantic resource fingerprint.
type DSEResourcePin struct {
	Path        string `json:"path"`
	ID          string `json:"id"`
	Version     string `json:"version"`
	Fingerprint string `json:"fingerprint"`
}

type DSELanguageConfig struct {
	SourcePackage DSEResourcePin `json:"source_package"`
	TargetPackage DSEResourcePin `json:"target_package"`
	SourcePlan    DSEResourcePin `json:"source_plan"`
	SourceContext DSEResourcePin `json:"source_context"`
}

// DSEProfileConfig is an explicit, separately pinned DSE selection. It does
// not merge its metamodel registry into the CORE/lexicon Pack.
type DSEProfileConfig struct {
	ID                  string                       `json:"id"`
	Version             string                       `json:"version"`
	RegistryFingerprint string                       `json:"registry_fingerprint"`
	Registries          []DSEResourcePin             `json:"registries"`
	Languages           map[string]DSELanguageConfig `json:"languages"`
	Budget              semantic.DSEBudget           `json:"budget,omitempty"`
}

func (profile *DSEProfileConfig) resolvePaths(base string) {
	if profile == nil {
		return
	}
	for index := range profile.Registries {
		profile.Registries[index].Path = resolvePath(base, profile.Registries[index].Path)
	}
	for languageID, config := range profile.Languages {
		config.SourcePackage.Path = resolvePath(base, config.SourcePackage.Path)
		config.TargetPackage.Path = resolvePath(base, config.TargetPackage.Path)
		config.SourcePlan.Path = resolvePath(base, config.SourcePlan.Path)
		config.SourceContext.Path = resolvePath(base, config.SourceContext.Path)
		profile.Languages[languageID] = config
	}
}

func (profile *DSEProfileConfig) validate() error {
	if profile == nil || strings.TrimSpace(profile.ID) == "" || strings.TrimSpace(profile.Version) == "" || strings.TrimSpace(profile.RegistryFingerprint) == "" {
		return runtimeError(status.OutcomeInvalid, "INVALID_DSE_PROFILE_CONFIG", "dse_profile", "DSE profile requires an ID, version and pinned registry fingerprint")
	}
	if len(profile.Registries) == 0 || len(profile.Registries) > MaxDSERegistryPackages {
		return runtimeError(status.OutcomeInvalid, "INVALID_DSE_REGISTRY_PATHS", "dse_profile.registries", "DSE profile requires a bounded, non-empty metamodel package set")
	}
	for i, pin := range profile.Registries {
		if err := validateDSEPin(pin); err != nil {
			return runtimeError(status.OutcomeInvalid, "INVALID_DSE_RESOURCE_PIN", fmt.Sprintf("dse_profile.registries[%d]", i), "metamodel package pin requires path, ID, version and fingerprint")
		}
	}
	if len(profile.Languages) != MaxDSEProfileLanguages {
		return runtimeError(status.OutcomeInvalid, "INVALID_DSE_LANGUAGE_PROFILES", "dse_profile.languages", "DSE runtime profile must explicitly configure both en and pl")
	}
	for _, languageID := range []string{"en", "pl"} {
		config, exists := profile.Languages[languageID]
		if !exists {
			return runtimeError(status.OutcomeInvalid, "DSE_LANGUAGE_PROFILE_REQUIRED", "dse_profile.languages", "DSE runtime profile must explicitly configure en and pl")
		}
		for path, pin := range map[string]DSEResourcePin{
			"source_package": config.SourcePackage, "target_package": config.TargetPackage,
			"source_plan": config.SourcePlan, "source_context": config.SourceContext,
		} {
			if err := validateDSEPin(pin); err != nil {
				return runtimeError(status.OutcomeInvalid, "INVALID_DSE_RESOURCE_PIN", "dse_profile.languages."+languageID+"."+path, "resource pin requires path, ID, version and fingerprint")
			}
		}
	}
	for languageID := range profile.Languages {
		if languageID != "en" && languageID != "pl" {
			return runtimeError(status.OutcomeInvalid, "INVALID_DSE_LANGUAGE", "dse_profile.languages", "only the configured en and pl language IDs are supported")
		}
	}
	if _, err := semantic.NormalizeDSEBudget(profile.Budget); err != nil {
		return preserveError(err, "INVALID_DSE_BUDGET", "dse_profile.budget")
	}
	if profile.Budget.MaxOutputBytes > engine.MaxTranslationPayloadBytes/4 {
		return runtimeError(status.OutcomeInvalid, "DSE_OUTPUT_BUDGET_TOO_LARGE", "dse_profile.budget.max_output_bytes", "DSE composition output budget cannot exceed one quarter of the public translation result limit")
	}
	return nil
}

func validateDSEPin(pin DSEResourcePin) error {
	if strings.TrimSpace(pin.Path) == "" || strings.TrimSpace(pin.ID) == "" || strings.TrimSpace(pin.Version) == "" || strings.TrimSpace(pin.Fingerprint) == "" {
		return fmt.Errorf("incomplete resource pin")
	}
	return nil
}

type dseRuntimeSource interface {
	AnalyzeDSE(context.Context, string, DSEOptions) (DSEAnalysis, error)
}

type configuredDSESource struct {
	pipeline dseRuntimeSource
	options  DSEOptions
}

func (source configuredDSESource) AnalyzeDSE(ctx context.Context, input string) (semantic.DSEResult, error) {
	analysis, err := source.pipeline.AnalyzeDSE(ctx, input, source.options)
	if err != nil {
		return semantic.DSEResult{}, err
	}
	return analysis.Composition, nil
}

func loadDSEProfile(config DSEProfileConfig, snapshot *pack.Pack, analyzers map[features.LanguageID]engine.SourceAnalyzer) (*engine.DSEProfile, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	registryInputs := make([][]byte, len(config.Registries))
	for i, pin := range config.Registries {
		data, err := readDSEFile(pin.Path, maxDSERegistryFileBytes)
		if err != nil {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_REGISTRY_UNAVAILABLE", fmt.Sprintf("dse_profile.registries[%d]", i), "configured metamodel package could not be read")
		}
		registryInputs[i] = data
	}
	registries, err := metamodel.LoadPackages(registryInputs)
	if err != nil {
		return nil, preserveError(err, "DSE_REGISTRY_INVALID", "dse_profile.registries")
	}
	if len(registries) != len(config.Registries) {
		return nil, runtimeError(status.OutcomeInvalid, "DSE_REGISTRY_SET_INCOMPLETE", "dse_profile.registries", "every activated registry, including dependencies, must have an explicit pin")
	}
	registryPins := make([]engine.DSERegistryPin, 0, len(registries))
	seenRegistryPins := make(map[string]struct{}, len(config.Registries))
	for _, configured := range config.Registries {
		if _, duplicate := seenRegistryPins[configured.ID]; duplicate {
			return nil, runtimeError(status.OutcomeInvalid, "DUPLICATE_DSE_REGISTRY_PIN", "dse_profile.registries", "metamodel registry IDs must have one pin each")
		}
		seenRegistryPins[configured.ID] = struct{}{}
		var selected *metamodel.Registry
		for _, registry := range registries {
			if registry.ID() == configured.ID {
				if selected != nil {
					return nil, runtimeError(status.OutcomeInvalid, "DUPLICATE_DSE_REGISTRY_ID", "dse_profile.registries", "configured registry ID is duplicated")
				}
				selected = registry
			}
		}
		if selected == nil || selected.Version() != configured.Version || selected.ResourceFingerprint() != configured.Fingerprint {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_REGISTRY_PIN_MISMATCH", "dse_profile.registries", "metamodel registry identity or resource fingerprint differs from the configured pin")
		}
		registryPins = append(registryPins, engine.DSERegistryPin{ID: selected.ID(), Version: selected.Version(), Fingerprint: selected.ResourceFingerprint()})
	}
	sort.Slice(registryPins, func(i, j int) bool { return registryPins[i].ID < registryPins[j].ID })
	if combinedDSERegistryFingerprint(registryPins) != config.RegistryFingerprint {
		return nil, runtimeError(status.OutcomeInvalid, "DSE_REGISTRY_SET_FINGERPRINT_MISMATCH", "dse_profile.registry_fingerprint", "selected metamodel registry set does not match its configured fingerprint")
	}
	selected, err := metamodel.DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return nil, preserveError(err, "DSE_METADATA_CONTEXT_INVALID", "dse_profile.registries")
	}
	rewrite, err := metamodel.NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return nil, preserveError(err, "DSE_REWRITE_PROFILE_INVALID", "dse_profile.registries")
	}
	requestedBudget := config.Budget
	if requestedBudget.MaxOutputBytes == 0 {
		requestedBudget.MaxOutputBytes = engine.MaxTranslationPayloadBytes / 4
	}
	budget, err := semantic.NormalizeDSEBudget(requestedBudget)
	if err != nil {
		return nil, preserveError(err, "INVALID_DSE_BUDGET", "dse_profile.budget")
	}
	profile := &engine.DSEProfile{
		ID: config.ID, Version: config.Version, RegistryFingerprint: config.RegistryFingerprint,
		Registries: registryPins, SelectedMetadata: selected,
		Languages: make(map[features.LanguageID]engine.DSELanguageResources, len(config.Languages)),
	}
	for _, languageID := range []features.LanguageID{"en", "pl"} {
		configured := config.Languages[string(languageID)]
		sourceAnalyzer, ok := analyzers[languageID].(dseRuntimeSource)
		if !ok || sourceAnalyzer == nil {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_STAGE_UNAVAILABLE", "dse_profile.languages."+string(languageID), "configured language has no compiled source pipeline with DSE support")
		}
		sourceBytes, err := readDSEFile(configured.SourcePackage.Path, language.MaxSemanticLanguagePackageBytes)
		if err != nil || rawDSEFingerprint(sourceBytes) != configured.SourcePackage.Fingerprint {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_PACKAGE_PIN_MISMATCH", "dse_profile.languages."+string(languageID)+".source_package", "source semantic package is missing or differs from its fingerprint pin")
		}
		sourcePackage, err := language.LoadSemanticPackage(sourceBytes, snapshot.CoreGraph())
		if err != nil || sourcePackage.ID() != configured.SourcePackage.ID || sourcePackage.Version() != configured.SourcePackage.Version || string(sourcePackage.Language()) != string(languageID) {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_PACKAGE_IDENTITY_MISMATCH", "dse_profile.languages."+string(languageID)+".source_package", "source semantic package identity, version or language differs from its pin")
		}
		targetBytes, err := readDSEFile(configured.TargetPackage.Path, language.MaxSemanticLanguagePackageBytes)
		if err != nil || rawDSEFingerprint(targetBytes) != configured.TargetPackage.Fingerprint {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_TARGET_PACKAGE_PIN_MISMATCH", "dse_profile.languages."+string(languageID)+".target_package", "target semantic package is missing or differs from its fingerprint pin")
		}
		targetPackage, err := language.LoadSemanticPackage(targetBytes, snapshot.CoreGraph())
		if err != nil || targetPackage.ID() != configured.TargetPackage.ID || targetPackage.Version() != configured.TargetPackage.Version || string(targetPackage.Language()) != string(languageID) {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_TARGET_PACKAGE_IDENTITY_MISMATCH", "dse_profile.languages."+string(languageID)+".target_package", "target semantic package identity, version or language differs from its pin")
		}
		targetValidations, err := targetPackage.ValidateTargets(selected)
		if err != nil {
			return nil, preserveError(err, "DSE_TARGET_BINDINGS_INVALID", "dse_profile.languages."+string(languageID)+".target_package")
		}
		targetBindingsUsable := false
		for _, validation := range targetValidations {
			if validation.Authorized && string(validation.Status) == "approved" && string(validation.Mapping) == "equivalent" && len(validation.Unresolved) == 0 {
				targetBindingsUsable = true
				break
			}
		}
		if !targetBindingsUsable {
			return nil, runtimeError(status.OutcomeUnsupported, "DSE_TARGET_BINDING_UNAVAILABLE", "dse_profile.languages."+string(languageID)+".target_package", "target package has no approved usable binding under selected metamodel")
		}
		planBytes, err := readDSEFile(configured.SourcePlan.Path, language.MaxSourcePlanBytes)
		if err != nil || rawDSEFingerprint(planBytes) != configured.SourcePlan.Fingerprint {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_PLAN_PIN_MISMATCH", "dse_profile.languages."+string(languageID)+".source_plan", "source plan is missing or differs from its fingerprint pin")
		}
		plan, err := language.LoadSourcePlan(planBytes, sourcePackage, selected)
		if err != nil || plan.ID() != configured.SourcePlan.ID || plan.Version() != configured.SourcePlan.Version {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_PLAN_IDENTITY_MISMATCH", "dse_profile.languages."+string(languageID)+".source_plan", "source plan identity or version differs from its pin or package declarations")
		}
		contextBytes, err := readDSEFile(configured.SourceContext.Path, language.MaxSemanticBindingRequestBytes)
		if err != nil || rawDSEFingerprint(contextBytes) != configured.SourceContext.Fingerprint {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_CONTEXT_PIN_MISMATCH", "dse_profile.languages."+string(languageID)+".source_context", "source reference context is missing or differs from its fingerprint pin")
		}
		references, err := language.DecodeSourceReferenceContext(contextBytes, selected)
		if err != nil || references.ContextID != configured.SourceContext.ID || fmt.Sprint(references.SchemaVersion) != configured.SourceContext.Version {
			return nil, runtimeError(status.OutcomeInvalid, "DSE_SOURCE_CONTEXT_IDENTITY_MISMATCH", "dse_profile.languages."+string(languageID)+".source_context", "source context identity or schema version differs from its pin")
		}
		profile.Languages[languageID] = engine.DSELanguageResources{
			Source: configuredDSESource{pipeline: sourceAnalyzer, options: DSEOptions{
				Package: sourcePackage, Plan: plan, Selected: selected, Rewrite: rewrite, References: references, Budget: budget,
			}},
			SourcePackage:   sourcePackage,
			SourcePackageID: sourcePackage.ID(), SourcePackageVersion: sourcePackage.Version(), SourcePackageFingerprint: configured.SourcePackage.Fingerprint,
			TargetPackage: targetPackage, TargetPackageFingerprint: configured.TargetPackage.Fingerprint,
			TargetBindingsUsable: targetBindingsUsable,
			SourcePlanID:         plan.ID(), SourcePlanVersion: plan.Version(), SourcePlanFingerprint: configured.SourcePlan.Fingerprint,
			SourceContextID: references.ContextID, SourceContextVersion: configured.SourceContext.Version, SourceContextFingerprint: configured.SourceContext.Fingerprint,
		}
	}
	return profile, nil
}

func readDSEFile(path string, maximum int64) ([]byte, error) {
	if strings.TrimSpace(path) == "" || maximum <= 0 {
		return nil, fmt.Errorf("invalid DSE file path or limit")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maximum {
		return nil, fmt.Errorf("DSE file is unavailable or outside its size limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("DSE file changed while opening")
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil || len(data) == 0 || int64(len(data)) > maximum {
		return nil, fmt.Errorf("DSE file could not be read within its limit")
	}
	return data, nil
}

func rawDSEFingerprint(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func combinedDSERegistryFingerprint(registries []engine.DSERegistryPin) string {
	copyPins := append([]engine.DSERegistryPin(nil), registries...)
	sort.Slice(copyPins, func(i, j int) bool { return copyPins[i].ID < copyPins[j].ID })
	encoded, err := json.Marshal(copyPins)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(append([]byte("golingua-dse-registry-set-v1\x00"), encoded...))
	return "dse-registry-set-v1:" + hex.EncodeToString(digest[:])
}
