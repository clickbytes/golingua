package language

import (
	"context"
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type AnalysisCandidate struct {
	SenseID        ids.SenseID
	CoreRef        ids.ConceptID
	ValencyFrameID string
	FamilyID       ids.ConstructionID
	VariantID      ids.ConstructionID
}

type LanguageProvider interface {
	Language() features.LanguageID
	Version() string
	Evidence() []CapabilityEvidence
}

type Analyzer interface {
	Analyze(context.Context, string) ([]AnalysisCandidate, error)
}
type Realizer interface {
	Realize(context.Context, meaning.Meaning, meaning.NodeID, construction.CompiledConstructionPlan) (string, error)
}

type CapabilityEvidence struct {
	ID      CapabilityID
	Analyze bool
	Realize bool
	CaseIDs []string
}
type VerifiedCapability struct {
	ID              CapabilityID `json:"id"`
	Analyze         bool         `json:"analyze"`
	Realize         bool         `json:"realize"`
	EvidenceCaseIDs []string     `json:"evidence_case_ids"`
}
type Language struct {
	Manifest     Manifest             `json:"manifest"`
	Bindings     []Binding            `json:"bindings"`
	Capabilities []VerifiedCapability `json:"capabilities"`
}

type Registry struct {
	languages map[string]Language
	providers map[string]LanguageProvider
}

func BuildRegistry(packages []Package, providers []LanguageProvider) (*Registry, error) {
	if len(packages) > 256 {
		return nil, resource("language_packages", len(packages), 256)
	}
	providerByLanguage := make(map[string]LanguageProvider, len(providers))
	if len(providers) > 256 {
		return nil, resource("language_providers", len(providers), 256)
	}
	for _, provider := range providers {
		if provider == nil {
			return nil, invalid("NIL_LANGUAGE_PROVIDER", "providers", "provider cannot be nil")
		}
		language := provider.Language()
		if !validLanguage(language) || !validVersion(provider.Version()) {
			return nil, invalid("INVALID_LANGUAGE_PROVIDER", "providers", "provider requires canonical language and version")
		}
		key := string(language)
		if _, ok := providerByLanguage[key]; ok {
			return nil, invalid("DUPLICATE_LANGUAGE_PROVIDER", "providers", "language provider is duplicated")
		}
		providerByLanguage[key] = provider
	}
	registry := &Registry{languages: make(map[string]Language), providers: providerByLanguage}
	for _, pack := range packages {
		if err := pack.Manifest.Validate(); err != nil {
			return nil, err
		}
		if len(pack.Bindings) > MaxBindings {
			return nil, resource("bindings", len(pack.Bindings), MaxBindings)
		}
		key := string(pack.Manifest.Language)
		if _, ok := registry.languages[key]; ok {
			return nil, invalid("DUPLICATE_LANGUAGE_PACKAGE", "packages", "language package is duplicated")
		}
		provider, hasProvider := providerByLanguage[key]
		if hasProvider && provider.Version() != pack.Manifest.Version {
			return nil, unsupported("LANGUAGE_PROVIDER_VERSION_MISMATCH", "manifest.version", "provider and package versions differ")
		}
		providerEvidence := map[CapabilityID]CapabilityEvidence{}
		if hasProvider {
			evidenceList := provider.Evidence()
			if len(evidenceList) > MaxCapabilities {
				return nil, resource("provider_capabilities", len(evidenceList), MaxCapabilities)
			}
			for _, evidence := range evidenceList {
				if !evidence.ID.Valid() || !evidence.Analyze && !evidence.Realize {
					return nil, invalid("INVALID_PROVIDER_EVIDENCE", "provider.evidence", "provider evidence requires a known capability and operation direction")
				}
				if _, err := canonicalEvidence(evidence.CaseIDs); err != nil {
					return nil, invalid("INVALID_PROVIDER_EVIDENCE", "provider.evidence", err.Error())
				}
				if _, ok := providerEvidence[evidence.ID]; ok {
					return nil, invalid("DUPLICATE_PROVIDER_EVIDENCE", "provider.evidence", "provider evidence capability is duplicated")
				}
				providerEvidence[evidence.ID] = evidence
			}
		}
		capabilities := make([]VerifiedCapability, 0, len(pack.Manifest.Capabilities))
		for _, declaration := range pack.Manifest.Capabilities {
			if !declaration.Implemented {
				continue
			}
			if !hasProvider {
				return nil, unsupported("LANGUAGE_CAPABILITY_WITHOUT_PROVIDER", "manifest.capabilities", "implemented capability has no compiled Go provider")
			}
			evidence, ok := providerEvidence[declaration.ID]
			if !ok {
				return nil, unsupported("LANGUAGE_CAPABILITY_WITHOUT_EVIDENCE", "manifest.capabilities", "provider does not attest the declared capability")
			}
			declaredCases, _ := canonicalEvidence(declaration.EvidenceCaseIDs)
			providerCases, err := canonicalEvidence(evidence.CaseIDs)
			if err != nil {
				return nil, invalid("INVALID_PROVIDER_EVIDENCE", "provider.evidence", "provider evidence IDs are invalid")
			}
			if !sameStrings(declaredCases, providerCases) || declaration.Analyze != evidence.Analyze || declaration.Realize != evidence.Realize {
				return nil, unsupported("LANGUAGE_CAPABILITY_EVIDENCE_MISMATCH", "manifest.capabilities", "manifest and Go provider capability evidence differ")
			}
			if evidence.Analyze {
				if _, ok := provider.(Analyzer); !ok {
					return nil, unsupported("ANALYZER_IMPLEMENTATION_MISSING", "provider", "provider declares analysis without implementing the typed Analyzer interface")
				}
			}
			if evidence.Realize {
				if _, ok := provider.(Realizer); !ok {
					return nil, unsupported("REALIZER_IMPLEMENTATION_MISSING", "provider", "provider declares realization without implementing the typed Realizer interface")
				}
			}
			capabilities = append(capabilities, VerifiedCapability{ID: declaration.ID, Analyze: evidence.Analyze, Realize: evidence.Realize, EvidenceCaseIDs: providerCases})
		}
		seenBindings := map[string]struct{}{}
		for _, binding := range pack.Bindings {
			if err := binding.Validate(); err != nil {
				return nil, err
			}
			bindingKey := binding.SenseID.String() + "\x00" + binding.ValencyFrameID + "\x00" + binding.FamilyID.String()
			if _, ok := seenBindings[bindingKey]; ok {
				return nil, invalid("DUPLICATE_LANGUAGE_BINDING", "bindings", "sense/frame/family binding is duplicated")
			}
			seenBindings[bindingKey] = struct{}{}
			if binding.Analyze && !supportsDirection(capabilities, true) {
				return nil, unsupported("BINDING_ANALYZE_UNSUPPORTED", "bindings", "analyze binding requires verified analyze capability")
			}
			if binding.Realize && !supportsDirection(capabilities, false) {
				return nil, unsupported("BINDING_REALIZE_UNSUPPORTED", "bindings", "realize binding requires verified realize capability")
			}
		}
		sort.Slice(capabilities, func(i, j int) bool { return capabilities[i].ID < capabilities[j].ID })
		registry.languages[key] = Language{Manifest: cloneManifest(pack.Manifest), Bindings: cloneBindings(pack.Bindings), Capabilities: capabilities}
	}
	return registry, nil
}

func (registry *Registry) Language(id features.LanguageID) (Language, bool) {
	if registry == nil || !validLanguage(id) {
		return Language{}, false
	}
	language, ok := registry.languages[string(id)]
	if !ok {
		return Language{}, false
	}
	return cloneLanguage(language), true
}
func (registry *Registry) Languages() []Language {
	if registry == nil {
		return []Language{}
	}
	keys := make([]string, 0, len(registry.languages))
	for key := range registry.languages {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Language, 0, len(keys))
	for _, key := range keys {
		result = append(result, cloneLanguage(registry.languages[key]))
	}
	return result
}
func (registry *Registry) Capabilities(id features.LanguageID) []VerifiedCapability {
	language, ok := registry.Language(id)
	if !ok {
		return []VerifiedCapability{}
	}
	result := append([]VerifiedCapability(nil), language.Capabilities...)
	for i := range result {
		result[i].EvidenceCaseIDs = append([]string(nil), result[i].EvidenceCaseIDs...)
	}
	return result
}
func (registry *Registry) Provider(id features.LanguageID) (LanguageProvider, bool) {
	if registry == nil || !validLanguage(id) {
		return nil, false
	}
	provider, ok := registry.providers[string(id)]
	return provider, ok
}

func supportsDirection(capabilities []VerifiedCapability, analyze bool) bool {
	for _, capability := range capabilities {
		if analyze && capability.Analyze || !analyze && capability.Realize {
			return true
		}
	}
	return false
}
func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func cloneManifest(manifest Manifest) Manifest {
	result := manifest
	result.Capabilities = append([]Capability(nil), manifest.Capabilities...)
	for i := range result.Capabilities {
		result.Capabilities[i].EvidenceCaseIDs = append([]string(nil), manifest.Capabilities[i].EvidenceCaseIDs...)
	}
	result.Modules = append([]Module(nil), manifest.Modules...)
	sort.Slice(result.Capabilities, func(i, j int) bool { return result.Capabilities[i].ID < result.Capabilities[j].ID })
	sort.Slice(result.Modules, func(i, j int) bool { return result.Modules[i].ID < result.Modules[j].ID })
	return result
}
func cloneBindings(bindings []Binding) []Binding {
	result := append([]Binding(nil), bindings...)
	for i := range result {
		result[i].VariantIDs = append([]ids.ConstructionID(nil), bindings[i].VariantIDs...)
		sort.Slice(result[i].VariantIDs, func(a, b int) bool { return result[i].VariantIDs[a].String() < result[i].VariantIDs[b].String() })
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.SenseID != right.SenseID {
			return left.SenseID.String() < right.SenseID.String()
		}
		if left.ValencyFrameID != right.ValencyFrameID {
			return left.ValencyFrameID < right.ValencyFrameID
		}
		return left.FamilyID.String() < right.FamilyID.String()
	})
	return result
}
func cloneLanguage(language Language) Language {
	result := language
	result.Manifest = cloneManifest(language.Manifest)
	result.Bindings = cloneBindings(language.Bindings)
	result.Capabilities = append([]VerifiedCapability(nil), language.Capabilities...)
	for i := range result.Capabilities {
		result.Capabilities[i].EvidenceCaseIDs = append([]string(nil), language.Capabilities[i].EvidenceCaseIDs...)
	}
	return result
}
