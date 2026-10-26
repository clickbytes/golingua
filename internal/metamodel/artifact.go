package metamodel

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
)

// ArtifactEnvironment builds the trusted, fully selected algebra environment
// for artifact export or replay. Artifact claims are never inputs to this
// factory; callers provide actual validated Registry snapshots and request
// metadata independently.
func ArtifactEnvironment(registries []*Registry, profile algebra.RewriteProfile, contextID string, dependencies []algebra.ContextDependency, sources []provenance.SourceRef, requestContexts ...*algebra.RequestContext) (algebra.ArtifactEnvironment, error) {
	if len(requestContexts) > 1 {
		return algebra.ArtifactEnvironment{}, invalid("INVALID_REQUEST_CONTEXT", "request_context", "at most one explicit request context may be supplied")
	}
	if len(registries) == 0 {
		return algebra.ArtifactEnvironment{}, invalid("EMPTY_REGISTRY_SET", "registries", "artifact construction requires explicitly selected registries")
	}
	if len(registries) > maxPackages {
		return algebra.ArtifactEnvironment{}, resource("RESOURCE_LIMIT", "registries", "selected registry count exceeds 16")
	}
	if err := algebra.ValidateRewriteProfile(profile); err != nil {
		return algebra.ArtifactEnvironment{}, err
	}
	if !validText(contextID, maxIDBytes) {
		return algebra.ArtifactEnvironment{}, invalid("INVALID_CONTEXT_ID", "context_id", "context ID must be bounded nonempty text")
	}
	dependencyCopy, err := validateArtifactContextDependencies(dependencies)
	if err != nil {
		return algebra.ArtifactEnvironment{}, err
	}
	sourceCopy, err := validateArtifactSources(sources)
	if err != nil {
		return algebra.ArtifactEnvironment{}, err
	}

	// The normalization environment performs the shared validated-registry
	// selection, approved-primitive/alias inventory, and formal snapshot logic.
	selected, err := NormalizationEnvironment(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return algebra.ArtifactEnvironment{}, err
	}
	ordered := append([]*Registry(nil), registries...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID() < ordered[j].ID() })
	pins := make([]algebra.RegistryPin, len(ordered))
	structureRecords := make([]artifactRegistryRecord, len(ordered))
	for i, registry := range ordered {
		if registry == nil || !validText(registry.ID(), maxIDBytes) || !validText(registry.Version(), maxIDBytes) ||
			!validText(registry.ResourceFingerprint(), 256) || !validText(registry.FormalDenotationID(), 256) {
			return algebra.ArtifactEnvironment{}, invalid("UNVALIDATED_REGISTRY", fmt.Sprintf("registries[%d]", i), "selected registry has no validated identity and fingerprints")
		}
		pins[i] = algebra.RegistryPin{
			ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint(),
			FormalDenotationID: registry.FormalDenotationID(),
		}
		document := registry.Document()
		encoded, encodeErr := canonicalResource(document)
		if encodeErr != nil {
			return algebra.ArtifactEnvironment{}, invalid("REGISTRY_STRUCTURE_ENCODING", fmt.Sprintf("registries[%d]", i), encodeErr.Error())
		}
		structureRecords[i] = artifactRegistryRecord{
			ID: registry.ID(), Version: registry.Version(), ResourceFingerprint: registry.ResourceFingerprint(),
			FormalDenotationID: registry.FormalDenotationID(), Document: encoded,
		}
	}
	structure, err := json.Marshal(structureRecords)
	if err != nil {
		return algebra.ArtifactEnvironment{}, invalid("REGISTRY_STRUCTURE_ENCODING", "registries", err.Error())
	}
	if len(structure) > 16<<20 {
		return algebra.ArtifactEnvironment{}, resource("RESOURCE_LIMIT", "registries", "complete selected registry structure exceeds 16 MiB")
	}
	var requestContext *algebra.RequestContext
	if len(requestContexts) == 1 && requestContexts[0] != nil {
		copyContext, cloneErr := algebra.CloneRequestContext(*requestContexts[0])
		if cloneErr != nil {
			return algebra.ArtifactEnvironment{}, cloneErr
		}
		requestContext = &copyContext
	}
	return algebra.ArtifactEnvironment{
		Document:            selected.Document,
		RegistrySnapshotID:  selected.SnapshotID,
		Registries:          pins,
		RegistryStructure:   append([]byte(nil), structure...),
		Profile:             algebra.RewriteProfile{ID: profile.ID, Rules: append([]algebra.RewriteRule{}, profile.Rules...)},
		ContextID:           contextID,
		RequestContext:      requestContext,
		ContextDependencies: dependencyCopy,
		Provenance:          sourceCopy,
	}, nil
}

type artifactRegistryRecord struct {
	ID                  string          `json:"id"`
	Version             string          `json:"version"`
	ResourceFingerprint string          `json:"resource_fingerprint"`
	FormalDenotationID  string          `json:"formal_denotation_id"`
	Document            json.RawMessage `json:"document"`
}

func validateArtifactContextDependencies(values []algebra.ContextDependency) ([]algebra.ContextDependency, error) {
	if len(values) > maxDeclarations {
		return nil, resource("RESOURCE_LIMIT", "context_dependencies", "context dependency count exceeds 512")
	}
	result := append([]algebra.ContextDependency{}, values...)
	seen := make(map[string]struct{}, len(result))
	for i, dependency := range result {
		if !validText(dependency.ID, maxIDBytes) || !validText(dependency.Fingerprint, maxIDBytes) {
			return nil, invalid("INVALID_CONTEXT_DEPENDENCY", fmt.Sprintf("context_dependencies[%d]", i), "dependency ID and fingerprint must be bounded nonempty text")
		}
		if _, exists := seen[dependency.ID]; exists {
			return nil, invalid("DUPLICATE_CONTEXT_DEPENDENCY", fmt.Sprintf("context_dependencies[%d].id", i), "dependency ID is duplicated")
		}
		seen[dependency.ID] = struct{}{}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func validateArtifactSources(values []provenance.SourceRef) ([]provenance.SourceRef, error) {
	if len(values) > maxDeclarations {
		return nil, resource("RESOURCE_LIMIT", "provenance", "provenance count exceeds 512")
	}
	result := append([]provenance.SourceRef{}, values...)
	seen := make(map[string]struct{}, len(result))
	for i, source := range result {
		if err := source.Validate(); err != nil {
			return nil, invalid("INVALID_PROVENANCE", fmt.Sprintf("provenance[%d]", i), err.Error())
		}
		if _, exists := seen[source.ID]; exists {
			return nil, invalid("DUPLICATE_PROVENANCE_ID", fmt.Sprintf("provenance[%d].id", i), "provenance ID is duplicated")
		}
		seen[source.ID] = struct{}{}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
