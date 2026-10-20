// Package language defines bounded, data-only language packs and their trusted
// Go providers. Pack data never contains executable code.
package language

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	SchemaVersion         = 1
	MaxModules            = 64
	MaxBindings           = 100_000
	MaxVariantsPerBinding = 256
	MaxCapabilities       = 64
	MaxEvidenceIDs        = 256
	MaxIdentifierBytes    = 128
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var versionPattern = regexp.MustCompile(`^v?[0-9]+(\.[0-9]+){0,2}(-[a-z0-9][a-z0-9.-]*)?$`)

type CapabilityID string

const (
	CapabilityAnalyze    CapabilityID = "analyze"
	CapabilityRealize    CapabilityID = "realize"
	CapabilityMorphology CapabilityID = "morphology"
)

type Capability struct {
	ID              CapabilityID `json:"id"`
	Implemented     bool         `json:"implemented"`
	Analyze         bool         `json:"analyze,omitempty"`
	Realize         bool         `json:"realize,omitempty"`
	EvidenceCaseIDs []string     `json:"evidence_case_ids,omitempty"`
}

type Module struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	SHA256   string `json:"sha256"`
	Optional bool   `json:"optional,omitempty"`
}

type Manifest struct {
	SchemaVersion int                 `json:"schema_version"`
	Language      features.LanguageID `json:"language"`
	Version       string              `json:"version"`
	Capabilities  []Capability        `json:"capabilities"`
	Modules       []Module            `json:"modules,omitempty"`
}

type Binding struct {
	SenseID        ids.SenseID          `json:"sense_id"`
	ValencyFrameID string               `json:"valency_frame_id"`
	FamilyID       ids.ConstructionID   `json:"family_id"`
	VariantIDs     []ids.ConstructionID `json:"variant_ids"`
	CoreRef        ids.ConceptID        `json:"core_ref"`
	Mapping        lexicon.MappingKind  `json:"mapping_kind"`
	Approved       bool                 `json:"approved"`
	Provenance     provenance.SourceRef `json:"provenance"`
	Analyze        bool                 `json:"analyze"`
	Realize        bool                 `json:"realize"`
}

type Package struct {
	Manifest Manifest  `json:"manifest"`
	Bindings []Binding `json:"bindings"`
}

func (binding *Binding) UnmarshalJSON(data []byte) error {
	var wire struct {
		SenseID        string               `json:"sense_id"`
		ValencyFrameID string               `json:"valency_frame_id"`
		FamilyID       string               `json:"family_id"`
		VariantIDs     []string             `json:"variant_ids"`
		CoreRef        string               `json:"core_ref"`
		Mapping        lexicon.MappingKind  `json:"mapping_kind"`
		Approved       bool                 `json:"approved"`
		Provenance     provenance.SourceRef `json:"provenance"`
		Analyze        bool                 `json:"analyze"`
		Realize        bool                 `json:"realize"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("binding JSON contains trailing data")
	}
	sense, err := ids.ParseSenseID(wire.SenseID)
	if err != nil {
		return err
	}
	family, err := ids.ParseConstructionID(wire.FamilyID)
	if err != nil {
		return err
	}
	core, err := ids.ParseConceptID(wire.CoreRef)
	if err != nil {
		return err
	}
	variants := make([]ids.ConstructionID, len(wire.VariantIDs))
	for i, value := range wire.VariantIDs {
		variants[i], err = ids.ParseConstructionID(value)
		if err != nil {
			return err
		}
	}
	*binding = Binding{SenseID: sense, ValencyFrameID: wire.ValencyFrameID, FamilyID: family, VariantIDs: variants, CoreRef: core, Mapping: wire.Mapping, Approved: wire.Approved, Provenance: wire.Provenance, Analyze: wire.Analyze, Realize: wire.Realize}
	return nil
}

func (binding Binding) MarshalJSON() ([]byte, error) {
	variants := make([]string, len(binding.VariantIDs))
	for i, value := range binding.VariantIDs {
		variants[i] = value.String()
	}
	return json.Marshal(struct {
		SenseID        string               `json:"sense_id"`
		ValencyFrameID string               `json:"valency_frame_id"`
		FamilyID       string               `json:"family_id"`
		VariantIDs     []string             `json:"variant_ids"`
		CoreRef        string               `json:"core_ref"`
		Mapping        lexicon.MappingKind  `json:"mapping_kind"`
		Approved       bool                 `json:"approved"`
		Provenance     provenance.SourceRef `json:"provenance"`
		Analyze        bool                 `json:"analyze"`
		Realize        bool                 `json:"realize"`
	}{binding.SenseID.String(), binding.ValencyFrameID, binding.FamilyID.String(), variants, binding.CoreRef.String(), binding.Mapping, binding.Approved, binding.Provenance, binding.Analyze, binding.Realize})
}

func (manifest Manifest) Validate() error {
	if manifest.SchemaVersion != SchemaVersion {
		return invalid("LANGUAGE_SCHEMA_VERSION", "manifest.schema_version", "unsupported language package schema version")
	}
	if !validLanguage(manifest.Language) || !validVersion(manifest.Version) {
		return invalid("INVALID_LANGUAGE_IDENTITY", "manifest", "manifest requires canonical language and bounded version")
	}
	if len(manifest.Capabilities) > MaxCapabilities {
		return resource("capabilities", len(manifest.Capabilities), MaxCapabilities)
	}
	if len(manifest.Modules) > MaxModules {
		return resource("modules", len(manifest.Modules), MaxModules)
	}
	seen := map[CapabilityID]struct{}{}
	for i, capability := range manifest.Capabilities {
		path := fmt.Sprintf("manifest.capabilities[%d]", i)
		if !capability.ID.Valid() {
			return invalid("INVALID_LANGUAGE_CAPABILITY", path+".id", "capability ID is unsupported")
		}
		if capability.Implemented && (capability.ID == CapabilityAnalyze && (!capability.Analyze || capability.Realize) || capability.ID == CapabilityRealize && (!capability.Realize || capability.Analyze) || capability.ID == CapabilityMorphology && !capability.Analyze && !capability.Realize) {
			return invalid("LANGUAGE_CAPABILITY_DIRECTION_MISMATCH", path, "capability directions do not match the declared operation")
		}
		if _, ok := seen[capability.ID]; ok {
			return invalid("DUPLICATE_LANGUAGE_CAPABILITY", path+".id", "capability is duplicated")
		}
		seen[capability.ID] = struct{}{}
		evidence, err := canonicalEvidence(capability.EvidenceCaseIDs)
		if err != nil {
			return invalid("INVALID_CAPABILITY_EVIDENCE", path+".evidence_case_ids", err.Error())
		}
		if capability.Implemented {
			if !capability.Analyze && !capability.Realize {
				return invalid("IMPLEMENTED_WITHOUT_DIRECTION", path, "implemented capability requires analyze or realize provider direction")
			}
			if len(evidence) == 0 {
				return invalid("IMPLEMENTED_WITHOUT_EVIDENCE", path, "implemented capability requires reviewed evidence case IDs")
			}
		} else if capability.Analyze || capability.Realize || len(evidence) > 0 {
			return invalid("UNIMPLEMENTED_CAPABILITY_HAS_EVIDENCE", path, "unimplemented capability cannot claim directions or evidence")
		}
	}
	moduleIDs := map[string]struct{}{}
	files := map[string]struct{}{}
	for i, module := range manifest.Modules {
		path := fmt.Sprintf("manifest.modules[%d]", i)
		if !validIdentifier(module.ID) || !validModuleFile(module.File) || !digestPattern.MatchString(module.SHA256) {
			return invalid("INVALID_LANGUAGE_MODULE", path, "module requires a bounded ID, local JSON filename, and lowercase SHA-256")
		}
		if _, ok := moduleIDs[module.ID]; ok {
			return invalid("DUPLICATE_LANGUAGE_MODULE", path+".id", "module ID is duplicated")
		}
		if _, ok := files[module.File]; ok {
			return invalid("DUPLICATE_LANGUAGE_MODULE_FILE", path+".file", "module file is duplicated")
		}
		moduleIDs[module.ID] = struct{}{}
		files[module.File] = struct{}{}
	}
	return nil
}

func (id CapabilityID) Valid() bool {
	return id == CapabilityAnalyze || id == CapabilityRealize || id == CapabilityMorphology
}
func canonicalEvidence(values []string) ([]string, error) {
	if len(values) > MaxEvidenceIDs {
		return nil, fmt.Errorf("evidence list exceeds %d", MaxEvidenceIDs)
	}
	result := append([]string(nil), values...)
	sort.Strings(result)
	for i, value := range result {
		if !validIdentifier(value) {
			return nil, fmt.Errorf("evidence case ID must be a bounded identifier")
		}
		if i > 0 && value == result[i-1] {
			return nil, fmt.Errorf("evidence case ID is duplicated")
		}
	}
	return result, nil
}
func validIdentifier(value string) bool {
	if len(value) == 0 || len(value) > MaxIdentifierBytes || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':') {
			return false
		}
	}
	return true
}
func validVersion(value string) bool { return len(value) <= 64 && versionPattern.MatchString(value) }
func validModuleFile(value string) bool {
	if len(value) <= 5 || len(value) > MaxIdentifierBytes || !strings.HasSuffix(value, ".json") {
		return false
	}
	base := strings.TrimSuffix(value, ".json")
	return validToken(base) && !strings.ContainsAny(value, "/\\")
}
func validToken(value string) bool {
	if len(value) == 0 || len(value) > MaxIdentifierBytes {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
func validLanguage(value features.LanguageID) bool {
	if len(value) < 2 || len(value) > 35 || strings.ToLower(string(value)) != string(value) {
		return false
	}
	parts := strings.Split(string(value), "-")
	for i, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, r := range part {
			if !(r >= 'a' && r <= 'z' || i > 0 && r >= '0' && r <= '9') {
				return false
			}
		}
	}
	return len(parts[0]) >= 2
}

func (binding Binding) Validate() error {
	if _, err := ids.ParseSenseID(binding.SenseID.String()); err != nil {
		return invalid("INVALID_LANGUAGE_BINDING_SENSE", "binding.sense_id", "binding requires canonical SenseID")
	}
	if !validIdentifier(binding.ValencyFrameID) {
		return invalid("INVALID_LANGUAGE_BINDING_FRAME", "binding.valency_frame_id", "binding requires bounded frame ID")
	}
	if _, err := ids.ParseConstructionID(binding.FamilyID.String()); err != nil {
		return invalid("INVALID_LANGUAGE_BINDING_FAMILY", "binding.family_id", "binding requires canonical family ID")
	}
	if _, err := ids.ParseConceptID(binding.CoreRef.String()); err != nil {
		return invalid("INVALID_LANGUAGE_BINDING_CORE", "binding.core_ref", "binding requires canonical CORE ConceptID")
	}
	if len(binding.VariantIDs) == 0 || len(binding.VariantIDs) > MaxVariantsPerBinding {
		return invalid("INVALID_LANGUAGE_BINDING_VARIANTS", "binding.variant_ids", "binding requires bounded variant IDs")
	}
	seen := map[string]struct{}{}
	for _, variant := range binding.VariantIDs {
		if _, err := ids.ParseConstructionID(variant.String()); err != nil {
			return invalid("INVALID_LANGUAGE_BINDING_VARIANT", "binding.variant_ids", "binding requires canonical variant IDs")
		}
		if _, ok := seen[variant.String()]; ok {
			return invalid("DUPLICATE_LANGUAGE_BINDING_VARIANT", "binding.variant_ids", "variant ID is duplicated")
		}
		seen[variant.String()] = struct{}{}
	}
	if !binding.Analyze && !binding.Realize {
		return invalid("LANGUAGE_BINDING_NO_DIRECTION", "binding", "binding requires explicit analyze or realize direction")
	}
	if binding.Mapping != lexicon.MappingEquivalent {
		return unsupported("LANGUAGE_BINDING_NOT_EQUIVALENT", "binding.mapping_kind", "language bridge requires an explicitly equivalent Sense-to-CORE mapping")
	}
	if !binding.Approved {
		return unsupported("LANGUAGE_BINDING_NOT_APPROVED", "binding.approved", "language binding must be explicitly approved before activation")
	}
	if err := binding.Provenance.Validate(); err != nil {
		return invalid("LANGUAGE_BINDING_PROVENANCE_INVALID", "binding.provenance", "approved language binding requires valid provenance")
	}
	return nil
}
