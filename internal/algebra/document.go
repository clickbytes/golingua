package algebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	NormalizerVersion     = "dse-normalizer-v1"
	ArtifactFormatVersion = 1
	ArtifactKind          = "typed_expression_dag"
	MaxArtifactBytes      = 8 << 20
)

// RegistryPin records both resource and formal registry identities. The two
// fingerprints serve different purposes and are never substituted for one
// another.
type RegistryPin struct {
	ID                  string `json:"id"`
	Version             string `json:"version"`
	ResourceFingerprint string `json:"resource_fingerprint"`
	FormalDenotationID  string `json:"formal_denotation_id"`
}

// ContextDependency pins an explicitly selected semantic context dependency.
type ContextDependency struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
}

// ArtifactEnvironment contains the declarations and request metadata selected
// by a trusted caller. Algebra validates structure; it does not validate
// metadata provenance or approval status.
type ArtifactEnvironment struct {
	Document           Document
	RegistrySnapshotID string
	Registries         []RegistryPin
	// RegistryStructure is optional trusted canonical registry content retained
	// only for full collision comparison. It is not part of the public envelope
	// or the semantic cache identity.
	RegistryStructure   []byte
	Profile             RewriteProfile
	ContextID           string
	RequestContext      *RequestContext
	ContextDependencies []ContextDependency
	Provenance          []provenance.SourceRef
}

// ArtifactEnvelope is the detached, versioned public representation of one
// checked expression DAG. Fingerprints are descriptive identities, not proof.
type ArtifactEnvelope struct {
	Kind                      string                 `json:"kind"`
	SchemaVersion             int                    `json:"schema_version"`
	NormalizerVersion         string                 `json:"normalizer_version"`
	RegistrySnapshotID        string                 `json:"registry_snapshot_id"`
	Registries                []RegistryPin          `json:"registries"`
	Profile                   RewriteProfile         `json:"profile"`
	ContextID                 string                 `json:"context_id"`
	RequestContext            *RequestContext        `json:"request_context,omitempty"`
	RequestContextFingerprint string                 `json:"request_context_fingerprint,omitempty"`
	ContextDependencies       []ContextDependency    `json:"context_dependencies"`
	Sorts                     []SortDeclaration      `json:"sorts"`
	Signatures                []Signature            `json:"signatures"`
	Parameters                []Declaration          `json:"parameters"`
	References                []Declaration          `json:"references"`
	Closed                    bool                   `json:"closed"`
	Limits                    Limits                 `json:"limits"`
	Root                      NodeID                 `json:"root"`
	Nodes                     []Node                 `json:"nodes"`
	Provenance                []provenance.SourceRef `json:"provenance"`
	ResourceFingerprint       string                 `json:"resource_fingerprint"`
	MeaningFingerprint        string                 `json:"meaning_fingerprint"`
	CacheKey                  string                 `json:"cache_key"`
}

// Artifact holds an immutable arena and the complete descriptors required by
// future collision-safe cache lookup. It does not execute or replay rewrites.
type Artifact struct {
	arena               *Arena
	envelope            ArtifactEnvelope
	resourceDescriptor  []byte
	registryDescriptor  []byte
	meaningDescriptor   []byte
	cacheDescriptor     []byte
	resourceFingerprint string
	meaningFingerprint  string
	cacheKey            string
}

// NewArtifact checks a real source document against the explicitly selected
// declaration inventory, builds its immutable arena, and computes versioned
// resource and conservative semantic identities.
func NewArtifact(environment ArtifactEnvironment, sourceDocument Document, useInterning bool) (*Artifact, error) {
	prepared, err := prepareArtifactSource(environment, sourceDocument)
	if err != nil {
		return nil, err
	}
	return buildArtifactFromPrepared(prepared, useInterning)
}

type artifactPreparation struct {
	selected validatedArtifactEnvironment
	document Document
	envelope ArtifactEnvelope
}

func prepareArtifactSource(environment ArtifactEnvironment, sourceDocument Document) (artifactPreparation, error) {
	selected, _, err := validateArtifactEnvironment(environment)
	if err != nil {
		return artifactPreparation{}, err
	}
	if sourceDocument.Expression == nil {
		return artifactPreparation{}, invalid("MISSING_EXPRESSION", "document.expression", "artifact source expression is required")
	}
	if sourceDocument.SchemaVersion != SchemaVersion {
		return artifactPreparation{}, invalid("INVALID_SCHEMA_VERSION", "document.schema_version", "unsupported algebra document version")
	}
	if err := matchInlineArtifactInventory(sourceDocument, selected.document); err != nil {
		return artifactPreparation{}, err
	}
	if (sourceDocument.RequestContext == nil) != (selected.requestContext == nil) {
		return artifactPreparation{}, invalid("REQUEST_CONTEXT_MISMATCH", "document.request_context", "source and independently selected request context must both be present or absent")
	}
	if sourceDocument.RequestContext != nil {
		if sourceDocument.RequestContext.ID != selected.contextID {
			return artifactPreparation{}, invalid("REQUEST_CONTEXT_ID_MISMATCH", "document.request_context.id", "source request context ID must match the selected context ID")
		}
	}

	document := sourceDocument
	document.Sorts = append([]SortDeclaration{}, selected.document.Sorts...)
	document.Signatures, err = cloneArtifactSignatures(selected.document.Signatures)
	if err != nil {
		return artifactPreparation{}, err
	}
	if sourceDocument.RequestContext != nil {
		canonical, sourceFP, contextErr := CanonicalRequestContext(*sourceDocument.RequestContext, document)
		if contextErr != nil {
			return artifactPreparation{}, contextErr
		}
		_, selectedFP, contextErr := CanonicalRequestContext(*selected.requestContext, document)
		if contextErr != nil {
			return artifactPreparation{}, contextErr
		}
		if sourceFP != selectedFP {
			return artifactPreparation{}, invalid("REQUEST_CONTEXT_MISMATCH", "document.request_context", "source context differs from the independently selected request context")
		}
		document.RequestContext = &canonical
	}
	// Validate the source declarations together with the selected inventory
	// before copying the source expression or constructing the arena.
	declarationContext := document
	declarationContext.Expression = nil
	if err := CheckDeclarations(declarationContext); err != nil {
		return artifactPreparation{}, err
	}
	if _, err := Check(document); err != nil {
		return artifactPreparation{}, err
	}
	envelope := artifactEnvelopeBase(selected, document)
	return artifactPreparation{selected: selected, document: document, envelope: envelope}, nil
}

func artifactEnvelopeBase(selected validatedArtifactEnvironment, document Document) ArtifactEnvelope {
	envelope := ArtifactEnvelope{
		Kind: ArtifactKind, SchemaVersion: ArtifactFormatVersion,
		NormalizerVersion: NormalizerVersion, RegistrySnapshotID: selected.registrySnapshotID,
		Registries: append([]RegistryPin{}, selected.registries...), Profile: cloneArtifactProfile(selected.profile),
		ContextID: selected.contextID, ContextDependencies: append([]ContextDependency{}, selected.contextDependencies...),
		Sorts: append([]SortDeclaration{}, document.Sorts...), Signatures: cloneArtifactSignaturesTrusted(document.Signatures),
		Parameters: append([]Declaration{}, document.Parameters...), References: append([]Declaration{}, document.References...),
		Closed: document.Closed, Limits: document.Limits, Provenance: append([]provenance.SourceRef{}, selected.provenance...),
	}
	if selected.requestContext != nil {
		envelope.RequestContext = selected.requestContext
		envelope.RequestContextFingerprint = selected.requestContextFingerprint
	}
	canonicalizeArtifactEnvelope(&envelope)
	return envelope
}

// artifactSourceResourceBytes captures the fully selected, hydrated source
// request before arena construction. It is used only for exact cache collision
// checks; the public artifact resource fingerprint remains based on the
// canonical envelope produced after arena construction.
func artifactSourceResourceBytes(prepared artifactPreparation, useInterning bool) ([]byte, error) {
	documentWireValue, err := documentWire(prepared.document)
	if err != nil {
		return nil, err
	}
	encodedDocument, err := json.Marshal(documentWireValue)
	if err != nil {
		return nil, invalid("ARTIFACT_SOURCE_ENCODING", "source", err.Error())
	}
	type sourceDescriptor struct {
		NormalizerVersion   string                 `json:"normalizer_version"`
		RegistrySnapshotID  string                 `json:"registry_snapshot_id"`
		Registries          []RegistryPin          `json:"registries"`
		Profile             RewriteProfile         `json:"profile"`
		ContextID           string                 `json:"context_id"`
		ContextDependencies []ContextDependency    `json:"context_dependencies"`
		Provenance          []provenance.SourceRef `json:"provenance"`
		Document            json.RawMessage        `json:"document"`
		UseInterning        bool                   `json:"use_interning"`
	}
	body := sourceDescriptor{
		NormalizerVersion: NormalizerVersion, RegistrySnapshotID: prepared.envelope.RegistrySnapshotID,
		Registries: append([]RegistryPin{}, prepared.envelope.Registries...), Profile: cloneArtifactProfile(prepared.envelope.Profile),
		ContextID: prepared.envelope.ContextID, ContextDependencies: append([]ContextDependency{}, prepared.envelope.ContextDependencies...),
		Provenance: append([]provenance.SourceRef{}, prepared.envelope.Provenance...), Document: encodedDocument,
		UseInterning: useInterning,
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, invalid("ARTIFACT_SOURCE_ENCODING", "source", err.Error())
	}
	return encoded, nil
}

func cloneArtifactSignaturesTrusted(values []Signature) []Signature {
	result := make([]Signature, len(values))
	for i, signature := range values {
		result[i] = Signature{ID: signature.ID, Label: signature.Label, Ports: append([]Port{}, signature.Ports...)}
	}
	return result
}

func buildArtifactFromPrepared(prepared artifactPreparation, useInterning bool) (*Artifact, error) {
	selected, document := prepared.selected, prepared.document
	arena, err := BuildArena(document, useInterning)
	if err != nil {
		return nil, err
	}
	ownedDocument, err := arena.Document()
	if err != nil {
		return nil, err
	}
	ownedSignatures, err := cloneArtifactSignatures(ownedDocument.Signatures)
	if err != nil {
		return nil, err
	}
	ownedEnvelope := prepared.envelope
	ownedEnvelope.Sorts = append([]SortDeclaration{}, ownedDocument.Sorts...)
	ownedEnvelope.Signatures = ownedSignatures
	ownedEnvelope.Parameters = append([]Declaration{}, ownedDocument.Parameters...)
	ownedEnvelope.References = append([]Declaration{}, ownedDocument.References...)
	ownedEnvelope.Closed, ownedEnvelope.Limits = ownedDocument.Closed, ownedDocument.Limits
	ownedEnvelope.Root, ownedEnvelope.Nodes = arena.Root(), arena.Nodes()
	canonicalizeArtifactEnvelope(&ownedEnvelope)

	resourceDescriptor, err := artifactResourceBytes(ownedEnvelope)
	if err != nil {
		return nil, err
	}
	if len(resourceDescriptor) > MaxArtifactBytes {
		return nil, resourceLimit("artifact", "canonical artifact resource descriptor exceeds 8 MiB")
	}
	registryDescriptor := append([]byte(nil), selected.registryStructure...)
	if len(registryDescriptor) > 16<<20 {
		return nil, resourceLimit("artifact.registry_structure", "trusted registry structure exceeds 16 MiB")
	}
	meaningDescriptor, err := artifactMeaningBytes(ownedEnvelope, ownedDocument.Expression)
	if err != nil {
		return nil, err
	}
	if len(meaningDescriptor) > MaxArtifactBytes {
		return nil, resourceLimit("artifact", "canonical artifact meaning descriptor exceeds 8 MiB")
	}
	resourceFingerprint := artifactDigest("golingua-artifact-resource-v1\x00", resourceDescriptor)
	meaningFingerprint := artifactDigest("golingua-artifact-meaning-v1\x00", meaningDescriptor)
	cacheDescriptor := append([]byte("golingua-artifact-cache-descriptor-v1\x00"), meaningDescriptor...)
	cacheKey := artifactDigest("golingua-artifact-cache-v1\x00", meaningDescriptor)
	ownedEnvelope.ResourceFingerprint = resourceFingerprint
	ownedEnvelope.MeaningFingerprint = meaningFingerprint
	ownedEnvelope.CacheKey = cacheKey

	// Bound the complete public payload too, including the three receipts.
	fullPayload, err := artifactEnvelopeBytes(ownedEnvelope)
	if err != nil {
		return nil, err
	}
	if len(fullPayload) > MaxArtifactBytes {
		return nil, resourceLimit("artifact", "complete artifact envelope exceeds 8 MiB")
	}
	return &Artifact{
		arena: arena, envelope: ownedEnvelope,
		resourceDescriptor:  append([]byte(nil), resourceDescriptor...),
		registryDescriptor:  registryDescriptor,
		meaningDescriptor:   append([]byte(nil), meaningDescriptor...),
		cacheDescriptor:     append([]byte(nil), cacheDescriptor...),
		resourceFingerprint: resourceFingerprint, meaningFingerprint: meaningFingerprint, cacheKey: cacheKey,
	}, nil
}

// Arena returns the artifact's immutable arena. Arena exposes only detached
// getters and no mutation operations.
func (artifact *Artifact) Arena() *Arena {
	if artifact == nil {
		return nil
	}
	return artifact.arena
}

// Envelope returns a detached copy of the public artifact record.
func (artifact *Artifact) Envelope() ArtifactEnvelope {
	if artifact == nil {
		return ArtifactEnvelope{}
	}
	copyValue, err := cloneArtifactEnvelope(artifact.envelope)
	if err != nil {
		return ArtifactEnvelope{}
	}
	return copyValue
}

// Document reconstructs a detached, checked algebra document.
func (artifact *Artifact) Document() (Document, error) {
	if artifact == nil || artifact.arena == nil {
		return Document{}, invalid("UNINITIALIZED_ARTIFACT", "artifact", "a validated artifact is required")
	}
	return artifact.arena.Document()
}

func (artifact *Artifact) ResourceFingerprint() string {
	if artifact == nil {
		return ""
	}
	return artifact.resourceFingerprint
}

func (artifact *Artifact) MeaningFingerprint() string {
	if artifact == nil {
		return ""
	}
	return artifact.meaningFingerprint
}

func (artifact *Artifact) CacheKey() string {
	if artifact == nil {
		return ""
	}
	return artifact.cacheKey
}

type validatedArtifactEnvironment struct {
	document                  Document
	limits                    Limits
	registrySnapshotID        string
	registries                []RegistryPin
	registryStructure         []byte
	profile                   RewriteProfile
	contextID                 string
	requestContext            *RequestContext
	requestContextFingerprint string
	contextDependencies       []ContextDependency
	provenance                []provenance.SourceRef
}

func validateArtifactEnvironment(environment ArtifactEnvironment) (validatedArtifactEnvironment, Limits, error) {
	if environment.Document.Expression != nil {
		return validatedArtifactEnvironment{}, Limits{}, invalid("ARTIFACT_ENVIRONMENT_HAS_EXPRESSION", "environment.document.expression", "artifact environment must contain declarations only")
	}
	if len(environment.Document.Parameters) != 0 || len(environment.Document.References) != 0 {
		return validatedArtifactEnvironment{}, Limits{}, invalid("ARTIFACT_ENVIRONMENT_HAS_REQUEST_DECLARATIONS", "environment.document", "parameters and references belong to the source document")
	}
	if environment.Document.Closed {
		return validatedArtifactEnvironment{}, Limits{}, invalid("ARTIFACT_ENVIRONMENT_HAS_REQUEST_MODE", "environment.document.closed", "closed mode belongs to the source document")
	}
	if environment.Document.RequestContext != nil {
		return validatedArtifactEnvironment{}, Limits{}, invalid("ARTIFACT_ENVIRONMENT_HAS_REQUEST_CONTEXT", "environment.document.request_context", "request context is supplied separately from declaration metadata")
	}
	if err := CheckDeclarations(environment.Document); err != nil {
		return validatedArtifactEnvironment{}, Limits{}, err
	}
	limits, err := NormalizeLimits(environment.Document.Limits)
	if err != nil {
		return validatedArtifactEnvironment{}, Limits{}, err
	}
	if !validText(environment.RegistrySnapshotID, 256) {
		return validatedArtifactEnvironment{}, Limits{}, invalid("INVALID_REGISTRY_SNAPSHOT", "environment.registry_snapshot_id", "registry snapshot ID is required and bounded")
	}
	if !validText(environment.ContextID, 256) {
		return validatedArtifactEnvironment{}, Limits{}, invalid("INVALID_CONTEXT_ID", "environment.context_id", "context ID is required and bounded")
	}
	if len(environment.Registries) == 0 || len(environment.Registries) > 16 {
		return validatedArtifactEnvironment{}, Limits{}, invalid("INVALID_REGISTRY_PINS", "environment.registries", "one to sixteen registry pins are required")
	}
	if err := ValidateRewriteProfile(environment.Profile); err != nil {
		return validatedArtifactEnvironment{}, Limits{}, err
	}
	if len(environment.RegistryStructure) > 16<<20 {
		return validatedArtifactEnvironment{}, Limits{}, resourceLimit("environment.registry_structure", "trusted registry structure exceeds 16 MiB")
	}
	registries := append([]RegistryPin{}, environment.Registries...)
	seenRegistry := make(map[string]struct{}, len(registries))
	for i, pin := range registries {
		path := fmt.Sprintf("environment.registries[%d]", i)
		if !validText(pin.ID, 256) || !validText(pin.Version, 256) || !validText(pin.ResourceFingerprint, 256) || !validText(pin.FormalDenotationID, 256) {
			return validatedArtifactEnvironment{}, Limits{}, invalid("INVALID_REGISTRY_PIN", path, "registry pin fields must be nonempty and bounded")
		}
		if _, exists := seenRegistry[pin.ID]; exists {
			return validatedArtifactEnvironment{}, Limits{}, invalid("DUPLICATE_REGISTRY_PIN", path+".id", "registry pin ID is duplicated")
		}
		seenRegistry[pin.ID] = struct{}{}
	}
	sort.Slice(registries, func(i, j int) bool { return registries[i].ID < registries[j].ID })
	if len(environment.ContextDependencies) > DefaultMaxSymbols {
		return validatedArtifactEnvironment{}, Limits{}, resourceLimit("environment.context_dependencies", "context dependency count exceeds 512")
	}
	dependencies := append([]ContextDependency{}, environment.ContextDependencies...)
	seenDependencies := make(map[string]struct{}, len(dependencies))
	for i, dependency := range dependencies {
		path := fmt.Sprintf("environment.context_dependencies[%d]", i)
		if !validText(dependency.ID, 256) || !validText(dependency.Fingerprint, 256) {
			return validatedArtifactEnvironment{}, Limits{}, invalid("INVALID_CONTEXT_DEPENDENCY", path, "dependency ID and fingerprint must be nonempty and bounded")
		}
		if _, exists := seenDependencies[dependency.ID]; exists {
			return validatedArtifactEnvironment{}, Limits{}, invalid("DUPLICATE_CONTEXT_DEPENDENCY", path+".id", "context dependency ID is duplicated")
		}
		seenDependencies[dependency.ID] = struct{}{}
	}
	sort.Slice(dependencies, func(i, j int) bool { return dependencies[i].ID < dependencies[j].ID })
	if len(environment.Provenance) > DefaultMaxSymbols {
		return validatedArtifactEnvironment{}, Limits{}, resourceLimit("environment.provenance", "provenance count exceeds 512")
	}
	provenanceRefs := append([]provenance.SourceRef{}, environment.Provenance...)
	seenProvenance := make(map[string]struct{}, len(provenanceRefs))
	for i, source := range provenanceRefs {
		if err := source.Validate(); err != nil {
			return validatedArtifactEnvironment{}, Limits{}, invalid("INVALID_PROVENANCE", fmt.Sprintf("environment.provenance[%d]", i), err.Error())
		}
		if _, exists := seenProvenance[source.ID]; exists {
			return validatedArtifactEnvironment{}, Limits{}, invalid("DUPLICATE_PROVENANCE_ID", fmt.Sprintf("environment.provenance[%d].id", i), "provenance ID is duplicated")
		}
		seenProvenance[source.ID] = struct{}{}
	}
	sort.Slice(provenanceRefs, func(i, j int) bool { return provenanceRefs[i].ID < provenanceRefs[j].ID })
	copyDocument, err := cloneArenaContext(environment.Document, limits)
	if err != nil {
		return validatedArtifactEnvironment{}, Limits{}, err
	}
	copyDocument.Limits = environment.Document.Limits
	copyDocument.Sorts = append([]SortDeclaration{}, environment.Document.Sorts...)
	copyDocument.Signatures, err = cloneArtifactSignatures(environment.Document.Signatures)
	if err != nil {
		return validatedArtifactEnvironment{}, Limits{}, err
	}
	var requestContext *RequestContext
	requestContextFingerprint := ""
	if environment.RequestContext != nil {
		if environment.RequestContext.ID != environment.ContextID {
			return validatedArtifactEnvironment{}, Limits{}, invalid("REQUEST_CONTEXT_ID_MISMATCH", "environment.request_context.id", "request context ID must match the selected context ID")
		}
		canonical, fingerprint, contextErr := CanonicalRequestContext(*environment.RequestContext, copyDocument)
		if contextErr != nil {
			return validatedArtifactEnvironment{}, Limits{}, contextErr
		}
		requestContext = &canonical
		requestContextFingerprint = fingerprint
	}
	return validatedArtifactEnvironment{
		document: copyDocument, limits: limits, registrySnapshotID: environment.RegistrySnapshotID,
		registries: registries, registryStructure: append([]byte(nil), environment.RegistryStructure...),
		profile: cloneArtifactProfile(environment.Profile), contextID: environment.ContextID,
		contextDependencies: dependencies, provenance: provenanceRefs,
		requestContext: requestContext, requestContextFingerprint: requestContextFingerprint,
	}, limits, nil
}

func matchInlineArtifactInventory(source Document, selected Document) error {
	sorts := make(map[string]SortDeclaration, len(selected.Sorts))
	for _, declaration := range selected.Sorts {
		sorts[declaration.ID] = declaration
	}
	if len(source.Sorts) > DefaultMaxSymbols || len(source.Signatures) > DefaultMaxSymbols {
		return resourceLimit("document", "inline inventory exceeds the declaration limit")
	}
	seenSorts := make(map[string]struct{}, len(source.Sorts))
	for i, declaration := range source.Sorts {
		if !validText(declaration.ID, 256) || !validOptionalText(declaration.Label, 1024) {
			return invalid("INVALID_INLINE_SORT", fmt.Sprintf("document.sorts[%d]", i), "inline sort fields are malformed")
		}
		if _, exists := seenSorts[declaration.ID]; exists {
			return invalid("DUPLICATE_INLINE_SORT", fmt.Sprintf("document.sorts[%d].id", i), "inline sort ID is duplicated")
		}
		seenSorts[declaration.ID] = struct{}{}
		actual, exists := sorts[declaration.ID]
		if !exists {
			return invalid("UNKNOWN_INLINE_SORT", fmt.Sprintf("document.sorts[%d].id", i), "inline sort is absent from the selected inventory")
		}
		if actual.ID != declaration.ID {
			return invalid("INLINE_SORT_MISMATCH", fmt.Sprintf("document.sorts[%d]", i), "inline sort does not match selected inventory")
		}
	}
	signatures := make(map[string]Signature, len(selected.Signatures))
	for _, signature := range selected.Signatures {
		signatures[signature.ID] = signature
	}
	seenSignatures := make(map[string]struct{}, len(source.Signatures))
	for i, signature := range source.Signatures {
		if !validText(signature.ID, 256) || !validOptionalText(signature.Label, 1024) || len(signature.Ports) > DefaultMaxPorts {
			return invalid("INVALID_INLINE_SIGNATURE", fmt.Sprintf("document.signatures[%d]", i), "inline signature fields exceed their bounds")
		}
		for j, port := range signature.Ports {
			if !validText(port.Role, 256) {
				return invalid("INVALID_INLINE_PORT", fmt.Sprintf("document.signatures[%d].ports[%d].role", i, j), "inline port role is malformed")
			}
			if _, err := cloneStandaloneType(&port.Type, DefaultLimits()); err != nil {
				return err
			}
		}
		if _, exists := seenSignatures[signature.ID]; exists {
			return invalid("DUPLICATE_INLINE_SIGNATURE", fmt.Sprintf("document.signatures[%d].id", i), "inline signature ID is duplicated")
		}
		seenSignatures[signature.ID] = struct{}{}
		actual, exists := signatures[signature.ID]
		if !exists {
			return invalid("UNKNOWN_INLINE_SIGNATURE", fmt.Sprintf("document.signatures[%d].id", i), "inline signature is absent from the selected inventory")
		}
		if !artifactSignatureEqual(actual, signature) {
			return invalid("INLINE_SIGNATURE_MISMATCH", fmt.Sprintf("document.signatures[%d]", i), "inline signature ports or types differ from selected inventory")
		}
	}
	return nil
}

func artifactSignatureEqual(left, right Signature) bool {
	if left.ID != right.ID || len(left.Ports) != len(right.Ports) {
		return false
	}
	for i, port := range left.Ports {
		if port.Role != right.Ports[i].Role || !TypeEqual(port.Type, right.Ports[i].Type) {
			return false
		}
	}
	return true
}

func cloneArtifactSignatures(values []Signature) ([]Signature, error) {
	result := make([]Signature, len(values))
	for i, signature := range values {
		result[i] = Signature{ID: signature.ID, Label: signature.Label, Ports: make([]Port, len(signature.Ports))}
		for j, port := range signature.Ports {
			typeCopy, err := cloneStandaloneType(&port.Type, DefaultLimits())
			if err != nil {
				return nil, err
			}
			result[i].Ports[j] = Port{Role: port.Role, Type: typeCopy}
		}
	}
	return result, nil
}

func cloneArtifactProfile(profile RewriteProfile) RewriteProfile {
	return RewriteProfile{ID: profile.ID, Rules: append([]RewriteRule{}, profile.Rules...)}
}

func canonicalizeArtifactEnvelope(envelope *ArtifactEnvelope) {
	sort.Slice(envelope.Registries, func(i, j int) bool { return envelope.Registries[i].ID < envelope.Registries[j].ID })
	sort.Slice(envelope.ContextDependencies, func(i, j int) bool { return envelope.ContextDependencies[i].ID < envelope.ContextDependencies[j].ID })
	sort.Slice(envelope.Sorts, func(i, j int) bool { return envelope.Sorts[i].ID < envelope.Sorts[j].ID })
	sort.Slice(envelope.Signatures, func(i, j int) bool { return envelope.Signatures[i].ID < envelope.Signatures[j].ID })
	sort.Slice(envelope.Parameters, func(i, j int) bool { return envelope.Parameters[i].ID < envelope.Parameters[j].ID })
	sort.Slice(envelope.References, func(i, j int) bool { return envelope.References[i].ID < envelope.References[j].ID })
	sort.Slice(envelope.Provenance, func(i, j int) bool { return envelope.Provenance[i].ID < envelope.Provenance[j].ID })
	if envelope.RequestContext != nil {
		context := envelope.RequestContext
		sort.Slice(context.Referents, func(i, j int) bool { return context.Referents[i].ID < context.Referents[j].ID })
		sort.Slice(context.Mentions, func(i, j int) bool { return context.Mentions[i].ID < context.Mentions[j].ID })
		sort.Slice(context.ResolutionEvidence, func(i, j int) bool { return context.ResolutionEvidence[i].ID < context.ResolutionEvidence[j].ID })
		sort.Slice(context.Conditions, func(i, j int) bool { return context.Conditions[i].ID < context.Conditions[j].ID })
		for i := range context.Conditions {
			sort.Strings(context.Conditions[i].TargetReferentIDs)
		}
	}
}

type artifactBody struct {
	Kind                      string                 `json:"kind"`
	SchemaVersion             int                    `json:"schema_version"`
	NormalizerVersion         string                 `json:"normalizer_version"`
	RegistrySnapshotID        string                 `json:"registry_snapshot_id"`
	Registries                []RegistryPin          `json:"registries"`
	Profile                   RewriteProfile         `json:"profile"`
	ContextID                 string                 `json:"context_id"`
	RequestContext            json.RawMessage        `json:"request_context,omitempty"`
	RequestContextFingerprint string                 `json:"request_context_fingerprint,omitempty"`
	ContextDependencies       []ContextDependency    `json:"context_dependencies"`
	Sorts                     []SortDeclaration      `json:"sorts"`
	Signatures                []Signature            `json:"signatures"`
	Parameters                []Declaration          `json:"parameters"`
	References                []Declaration          `json:"references"`
	Closed                    bool                   `json:"closed"`
	Limits                    Limits                 `json:"limits"`
	Root                      NodeID                 `json:"root"`
	Nodes                     []json.RawMessage      `json:"nodes"`
	Provenance                []provenance.SourceRef `json:"provenance"`
}

func artifactResourceBytes(envelope ArtifactEnvelope) ([]byte, error) {
	nodes := make([]json.RawMessage, len(envelope.Nodes))
	for i, node := range envelope.Nodes {
		wire, err := artifactNodeWire(node, true)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(wire)
		if err != nil {
			return nil, invalid("ARTIFACT_NODE_ENCODING", fmt.Sprintf("nodes[%d]", i), err.Error())
		}
		nodes[i] = encoded
	}
	signatures, err := cloneArtifactSignatures(envelope.Signatures)
	if err != nil {
		return nil, err
	}
	var requestContextWire json.RawMessage
	if envelope.RequestContext != nil {
		requestContextWire, err = requestContextRaw(*envelope.RequestContext, false)
		if err != nil {
			return nil, err
		}
	}
	body := artifactBody{
		Kind: envelope.Kind, SchemaVersion: envelope.SchemaVersion, NormalizerVersion: envelope.NormalizerVersion,
		RegistrySnapshotID: envelope.RegistrySnapshotID, Registries: append([]RegistryPin{}, envelope.Registries...),
		Profile: cloneArtifactProfile(envelope.Profile), ContextID: envelope.ContextID,
		RequestContext:            requestContextWire,
		RequestContextFingerprint: envelope.RequestContextFingerprint,
		ContextDependencies:       append([]ContextDependency{}, envelope.ContextDependencies...),
		Sorts:                     append([]SortDeclaration{}, envelope.Sorts...), Signatures: signatures,
		Parameters: append([]Declaration{}, envelope.Parameters...), References: append([]Declaration{}, envelope.References...),
		Closed: envelope.Closed, Limits: envelope.Limits, Root: envelope.Root,
		Nodes: nodes, Provenance: append([]provenance.SourceRef{}, envelope.Provenance...),
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, invalid("ARTIFACT_ENCODING", "artifact", err.Error())
	}
	return encoded, nil
}

func artifactEnvelopeBytes(envelope ArtifactEnvelope) ([]byte, error) {
	body, err := artifactResourceBytes(envelope)
	if err != nil {
		return nil, err
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, invalid("ARTIFACT_ENCODING", "artifact", err.Error())
	}
	for key, field := range map[string]string{
		"resource_fingerprint": envelope.ResourceFingerprint,
		"meaning_fingerprint":  envelope.MeaningFingerprint,
		"cache_key":            envelope.CacheKey,
	} {
		encoded, err := json.Marshal(field)
		if err != nil {
			return nil, invalid("ARTIFACT_ENCODING", "artifact."+key, err.Error())
		}
		value[key] = encoded
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, invalid("ARTIFACT_ENCODING", "artifact", err.Error())
	}
	return encoded, nil
}

func artifactNodeWire(node Node, includeNodeID bool) (map[string]any, error) {
	wire := map[string]any{"kind": node.Kind}
	if includeNodeID {
		wire["node_id"] = node.NodeID
	}
	if node.Label != "" {
		wire["label"] = node.Label
	}
	switch node.Kind {
	case TermVar:
		wire["identity"] = node.Identity
	case TermParam, TermRef, TermHole:
		wire["identity"], wire["type"] = node.Identity, node.Type
	case TermAtom:
		args := make([]map[string]any, len(node.Args))
		for i, argument := range node.Args {
			args[i] = map[string]any{"role": argument.Role, "value": argument.Value}
		}
		wire["symbol"], wire["args"] = node.Symbol, args
	case TermAnd:
		operands := append([]NodeID{}, node.Operands...)
		wire["operands"] = operands
	case TermExists, TermLambda:
		wire["identity"], wire["type"], wire["body"] = node.Identity, node.Type, node.Body
		if node.Kind == TermExists && node.TemporalFrame != "" {
			wire["temporal_frame"] = node.TemporalFrame
		}
	case TermNot:
		wire["body"] = node.Body
	case TermModal:
		wire["modality"], wire["body"] = node.Modality, node.Body
		if node.ModalProfile != "" {
			wire["modal_profile"] = node.ModalProfile
		}
	case TermQuantifier:
		wire["identity"], wire["type"], wire["quantifier"], wire["body"] = node.Identity, node.Type, node.Quantifier, node.Body
		if node.Count != nil {
			wire["count"] = *node.Count
		}
		if node.Restrictor != 0 {
			wire["restrictor"] = node.Restrictor
		}
	case TermAsk:
		wire["answers"], wire["body"] = node.Answers, node.Body
		if node.Restrictor != 0 {
			wire["restrictor"] = node.Restrictor
		}
	case TermApply:
		wire["function"], wire["argument"] = node.Function, node.Argument
	default:
		return nil, invalid("UNKNOWN_TERM_KIND", "node.kind", "artifact contains an unsupported term kind")
	}
	return wire, nil
}

type artifactMeaningBody struct {
	NormalizerVersion         string              `json:"normalizer_version"`
	RegistrySnapshotID        string              `json:"registry_snapshot_id"`
	FormalRegistryPins        []formalRegistryPin `json:"formal_registry_pins"`
	Profile                   RewriteProfile      `json:"profile"`
	ContextID                 string              `json:"context_id"`
	RequestContext            json.RawMessage     `json:"request_context,omitempty"`
	RequestContextFingerprint string              `json:"request_context_fingerprint,omitempty"`
	ContextDependencies       []ContextDependency `json:"context_dependencies"`
	Sorts                     []string            `json:"sorts"`
	Signatures                []formalSignature   `json:"signatures"`
	Parameters                []Declaration       `json:"parameters"`
	References                []Declaration       `json:"references"`
	Closed                    bool                `json:"closed"`
	Expression                map[string]any      `json:"expression"`
}

type formalRegistryPin struct {
	ID                 string `json:"id"`
	FormalDenotationID string `json:"formal_denotation_id"`
}

type formalSignature struct {
	ID    string `json:"id"`
	Ports []Port `json:"ports"`
}

func artifactMeaningBytes(envelope ArtifactEnvelope, expression *Term) ([]byte, error) {
	formalPins := make([]formalRegistryPin, len(envelope.Registries))
	for i, pin := range envelope.Registries {
		formalPins[i] = formalRegistryPin{ID: pin.ID, FormalDenotationID: pin.FormalDenotationID}
	}
	sorts := make([]string, len(envelope.Sorts))
	for i, sortDeclaration := range envelope.Sorts {
		sorts[i] = sortDeclaration.ID
	}
	signatures := make([]formalSignature, len(envelope.Signatures))
	for i, signature := range envelope.Signatures {
		signatures[i] = formalSignature{ID: signature.ID, Ports: append([]Port{}, signature.Ports...)}
	}
	expressionWire, err := artifactMeaningTerm(expression, 1, make(map[*Term]bool))
	if err != nil {
		return nil, err
	}
	var requestContextWire json.RawMessage
	if envelope.RequestContext != nil {
		requestContextWire, err = requestContextRaw(*envelope.RequestContext, true)
		if err != nil {
			return nil, err
		}
	}
	body := artifactMeaningBody{
		NormalizerVersion: NormalizerVersion, RegistrySnapshotID: envelope.RegistrySnapshotID,
		FormalRegistryPins: formalPins, Profile: cloneArtifactProfile(envelope.Profile),
		ContextID: envelope.ContextID, ContextDependencies: append([]ContextDependency{}, envelope.ContextDependencies...),
		RequestContext:            requestContextWire,
		RequestContextFingerprint: envelope.RequestContextFingerprint,
		Sorts:                     sorts, Signatures: signatures,
		Parameters: append([]Declaration{}, envelope.Parameters...), References: append([]Declaration{}, envelope.References...),
		Closed: envelope.Closed, Expression: expressionWire,
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, invalid("ARTIFACT_MEANING_ENCODING", "artifact.meaning", err.Error())
	}
	return encoded, nil
}

func artifactMeaningTerm(term *Term, depth int, active map[*Term]bool) (map[string]any, error) {
	if term == nil {
		return nil, invalid("MISSING_TERM", "artifact.meaning.expression", "meaning descriptor encountered a missing term")
	}
	if depth > DefaultMaxDepth {
		return nil, resourceLimit("artifact.meaning.expression", "meaning descriptor depth exceeds limit")
	}
	if active[term] {
		return nil, invalid("CYCLIC_TERM", "artifact.meaning.expression", "meaning descriptor encountered a term cycle")
	}
	active[term] = true
	defer delete(active, term)
	wire := map[string]any{"kind": term.Kind}
	switch term.Kind {
	case TermVar:
		wire["id"] = term.ID
	case TermParam, TermRef, TermHole:
		wire["id"], wire["type"] = term.ID, term.Type
	case TermAtom:
		args := make([]map[string]any, len(term.Args))
		for i, argument := range term.Args {
			value, err := artifactMeaningTerm(argument.Value, depth+1, active)
			if err != nil {
				return nil, err
			}
			args[i] = map[string]any{"role": argument.Role, "value": value}
		}
		wire["symbol"], wire["args"] = term.Symbol, args
	case TermAnd:
		operands := make([]map[string]any, len(term.Operands))
		for i, operand := range term.Operands {
			value, err := artifactMeaningTerm(operand, depth+1, active)
			if err != nil {
				return nil, err
			}
			operands[i] = value
		}
		wire["operands"] = operands
	case TermExists, TermLambda:
		body, err := artifactMeaningTerm(term.Body, depth+1, active)
		if err != nil {
			return nil, err
		}
		wire["id"], wire["type"], wire["body"] = term.ID, term.Type, body
		if term.Kind == TermExists && term.TemporalFrame != "" {
			wire["temporal_frame"] = term.TemporalFrame
		}
	case TermNot:
		body, err := artifactMeaningTerm(term.Body, depth+1, active)
		if err != nil {
			return nil, err
		}
		wire["body"] = body
	case TermModal:
		body, err := artifactMeaningTerm(term.Body, depth+1, active)
		if err != nil {
			return nil, err
		}
		wire["modality"], wire["body"] = term.Modality, body
		if term.ModalProfile != "" {
			wire["modal_profile"] = term.ModalProfile
		}
	case TermQuantifier:
		body, err := artifactMeaningTerm(term.Body, depth+1, active)
		if err != nil {
			return nil, err
		}
		wire["id"], wire["type"], wire["quantifier"], wire["body"] = term.ID, term.Type, term.Quantifier, body
		if term.Count != nil {
			wire["count"] = *term.Count
		}
		if term.Restrictor != nil {
			restrictor, err := artifactMeaningTerm(term.Restrictor, depth+1, active)
			if err != nil {
				return nil, err
			}
			wire["restrictor"] = restrictor
		}
	case TermAsk:
		body, err := artifactMeaningTerm(term.Body, depth+1, active)
		if err != nil {
			return nil, err
		}
		answers := make([]Declaration, len(term.Answers))
		copy(answers, term.Answers)
		wire["answers"], wire["body"] = answers, body
		if term.Restrictor != nil {
			restrictor, err := artifactMeaningTerm(term.Restrictor, depth+1, active)
			if err != nil {
				return nil, err
			}
			wire["restrictor"] = restrictor
		}
	case TermApply:
		function, err := artifactMeaningTerm(term.Function, depth+1, active)
		if err != nil {
			return nil, err
		}
		argument, err := artifactMeaningTerm(term.Argument, depth+1, active)
		if err != nil {
			return nil, err
		}
		wire["function"], wire["argument"] = function, argument
	default:
		return nil, invalid("UNKNOWN_TERM_KIND", "artifact.meaning.expression.kind", "meaning descriptor contains an unsupported term kind")
	}
	return wire, nil
}

func artifactDigest(domain string, descriptor []byte) string {
	input := append([]byte(domain), descriptor...)
	digest := sha256.Sum256(input)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func cloneArtifactEnvelope(envelope ArtifactEnvelope) (ArtifactEnvelope, error) {
	copyValue := envelope
	var err error
	copyValue.Registries = append([]RegistryPin{}, envelope.Registries...)
	copyValue.ContextDependencies = append([]ContextDependency{}, envelope.ContextDependencies...)
	copyValue.Profile = cloneArtifactProfile(envelope.Profile)
	copyValue.Sorts = append([]SortDeclaration{}, envelope.Sorts...)
	copyValue.Signatures, err = cloneArtifactSignatures(envelope.Signatures)
	if err != nil {
		return ArtifactEnvelope{}, err
	}
	copyValue.Parameters = make([]Declaration, len(envelope.Parameters))
	for i, declaration := range envelope.Parameters {
		typeCopy, err := cloneStandaloneType(&declaration.Type, DefaultLimits())
		if err != nil {
			return ArtifactEnvelope{}, err
		}
		copyValue.Parameters[i] = Declaration{ID: declaration.ID, Type: typeCopy, TemporalFrame: declaration.TemporalFrame, ReferenceTimeID: declaration.ReferenceTimeID, Multiplicity: declaration.Multiplicity}
	}
	copyValue.References = make([]Declaration, len(envelope.References))
	for i, declaration := range envelope.References {
		typeCopy, err := cloneStandaloneType(&declaration.Type, DefaultLimits())
		if err != nil {
			return ArtifactEnvelope{}, err
		}
		copyValue.References[i] = Declaration{ID: declaration.ID, Type: typeCopy, TemporalFrame: declaration.TemporalFrame, ReferenceTimeID: declaration.ReferenceTimeID, Multiplicity: declaration.Multiplicity}
	}
	copyValue.Nodes, err = cloneArenaNodes(envelope.Nodes, DefaultLimits())
	if err != nil {
		return ArtifactEnvelope{}, err
	}
	copyValue.Provenance = append([]provenance.SourceRef{}, envelope.Provenance...)
	if envelope.RequestContext != nil {
		contextCopy, err := CloneRequestContext(*envelope.RequestContext)
		if err != nil {
			return ArtifactEnvelope{}, err
		}
		copyValue.RequestContext = &contextCopy
	}
	return copyValue, nil
}
