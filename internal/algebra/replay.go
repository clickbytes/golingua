package algebra

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

// ReplayArtifact validates an artifact envelope against independently selected
// metadata and request context. It reconstructs no metadata from the envelope
// and performs no normalization or fact evaluation.
func ReplayArtifact(expected ArtifactEnvironment, envelope ArtifactEnvelope, requestedLimits Limits) (*Artifact, error) {
	selected, _, err := validateArtifactEnvironment(expected)
	if err != nil {
		return nil, err
	}
	if err := validateReplayHeader(envelope); err != nil {
		return nil, err
	}
	base := artifactEnvelopeBase(selected, selected.document)
	if err := matchReplayEnvironment(envelope, base); err != nil {
		return nil, err
	}
	wireLimits, err := NormalizeLimits(envelope.Limits)
	if err != nil {
		return nil, err
	}
	requestLimits, err := NormalizeLimits(requestedLimits)
	if err != nil {
		return nil, err
	}
	effectiveLimits := intersectLimits(wireLimits, requestLimits)
	if len(envelope.Nodes) > effectiveLimits.MaxNodes {
		return nil, resourceLimit("nodes", "artifact node count exceeds effective replay limits")
	}
	if err := preflightReplayNodes(envelope.Nodes, effectiveLimits); err != nil {
		return nil, err
	}
	if err := validateReplayDeclarations(envelope); err != nil {
		return nil, err
	}
	context := selected.document
	context.Parameters = envelope.Parameters
	context.References = envelope.References
	context.Closed = envelope.Closed
	context.Limits = effectiveLimits
	context.Expression = nil
	arena, err := NewArena(context, envelope.Root, envelope.Nodes)
	if err != nil {
		return nil, err
	}

	// The graph is now bounded and fully checked, so recursive copies and wire
	// canonicalization below operate only on admitted structures.
	ownedEnvelope, err := cloneArtifactEnvelope(envelope)
	if err != nil {
		return nil, err
	}
	resourceDescriptor, err := artifactResourceBytes(ownedEnvelope)
	if err != nil {
		return nil, err
	}
	if len(resourceDescriptor) > MaxArtifactBytes {
		return nil, resourceLimit("artifact", "canonical artifact resource descriptor exceeds 8 MiB")
	}
	validatedDocument, err := arena.Document()
	if err != nil {
		return nil, err
	}
	meaningDescriptor, err := artifactMeaningBytes(ownedEnvelope, validatedDocument.Expression)
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
	if envelope.ResourceFingerprint != resourceFingerprint {
		return nil, invalid("RESOURCE_FINGERPRINT_MISMATCH", "resource_fingerprint", "artifact resource fingerprint does not match its canonical envelope body")
	}
	if envelope.MeaningFingerprint != meaningFingerprint {
		return nil, invalid("MEANING_FINGERPRINT_MISMATCH", "meaning_fingerprint", "artifact meaning fingerprint does not match the validated expression and selected metadata")
	}
	if envelope.CacheKey != cacheKey {
		return nil, invalid("CACHE_KEY_MISMATCH", "cache_key", "artifact cache key does not match its canonical semantic descriptor")
	}
	fullPayload, err := artifactEnvelopeBytes(ownedEnvelope)
	if err != nil {
		return nil, err
	}
	if len(fullPayload) > MaxArtifactBytes {
		return nil, resourceLimit("artifact", "complete artifact envelope exceeds 8 MiB")
	}
	return &Artifact{
		arena: arena, envelope: ownedEnvelope,
		resourceDescriptor: append([]byte(nil), resourceDescriptor...), registryDescriptor: append([]byte(nil), selected.registryStructure...),
		meaningDescriptor: append([]byte(nil), meaningDescriptor...), cacheDescriptor: append([]byte(nil), cacheDescriptor...),
		resourceFingerprint: resourceFingerprint, meaningFingerprint: meaningFingerprint, cacheKey: cacheKey,
	}, nil
}

func validateReplayHeader(envelope ArtifactEnvelope) error {
	if envelope.Kind != ArtifactKind {
		return invalid("INVALID_ARTIFACT_KIND", "kind", "artifact kind is not supported")
	}
	if envelope.SchemaVersion != ArtifactFormatVersion {
		return invalid("INVALID_ARTIFACT_VERSION", "schema_version", "artifact envelope version is not supported")
	}
	if envelope.NormalizerVersion != NormalizerVersion {
		return &Error{Status: status.OutcomeUnsupported, Code: "UNKNOWN_NORMALIZER_VERSION", Path: "normalizer_version", Message: "artifact uses an unknown normalizer version"}
	}
	if !validText(envelope.RegistrySnapshotID, 256) || !validText(envelope.ContextID, 256) {
		return invalid("INVALID_ARTIFACT_HEADER", "artifact", "registry snapshot and context IDs must be bounded")
	}
	if !validText(envelope.ResourceFingerprint, 256) || !validText(envelope.MeaningFingerprint, 256) || !validText(envelope.CacheKey, 256) {
		return invalid("INVALID_ARTIFACT_FINGERPRINT", "artifact", "all three artifact receipts must be bounded nonempty strings")
	}
	if len(envelope.Registries) == 0 {
		return invalid("MISSING_REGISTRY_PINS", "registries", "artifact must carry the selected registry pins")
	}
	if len(envelope.Registries) > 16 || len(envelope.ContextDependencies) > DefaultMaxSymbols || len(envelope.Provenance) > DefaultMaxSymbols ||
		len(envelope.Sorts) > DefaultMaxSymbols || len(envelope.Signatures) > DefaultMaxSymbols ||
		len(envelope.Parameters) > DefaultMaxSymbols || len(envelope.References) > DefaultMaxSymbols ||
		len(envelope.Nodes) > DefaultMaxNodes {
		return resourceLimit("artifact", "artifact declaration or node count exceeds hard replay bounds")
	}
	if len(envelope.Nodes) == 0 {
		return invalid("EMPTY_ARTIFACT_NODES", "nodes", "artifact must contain at least one syntax node")
	}
	for i, sortDeclaration := range envelope.Sorts {
		if !validText(sortDeclaration.ID, 256) || !validOptionalText(sortDeclaration.Label, 1024) {
			return invalid("INVALID_ARTIFACT_SORT", fmt.Sprintf("sorts[%d]", i), "artifact sort fields are malformed")
		}
	}
	for i, signature := range envelope.Signatures {
		path := fmt.Sprintf("signatures[%d]", i)
		if !validText(signature.ID, 256) || !validOptionalText(signature.Label, 1024) {
			return invalid("INVALID_ARTIFACT_SIGNATURE", path, "artifact signature fields are malformed")
		}
		if len(signature.Ports) > DefaultMaxPorts {
			return resourceLimit(path+".ports", "artifact signature port count exceeds the hard limit")
		}
		for j, port := range signature.Ports {
			if !validText(port.Role, 256) {
				return invalid("INVALID_ARTIFACT_PORT", fmt.Sprintf("%s.ports[%d].role", path, j), "artifact role must be bounded")
			}
			if _, err := cloneStandaloneType(&port.Type, DefaultLimits()); err != nil {
				return err
			}
		}
	}
	if err := ValidateRewriteProfile(envelope.Profile); err != nil {
		return err
	}
	return nil
}

func matchReplayEnvironment(actual, expected ArtifactEnvelope) error {
	if actual.RegistrySnapshotID != expected.RegistrySnapshotID {
		return invalid("REGISTRY_SNAPSHOT_MISMATCH", "registry_snapshot_id", "artifact names a different selected registry snapshot")
	}
	if !sameRegistryPins(actual.Registries, expected.Registries) {
		return invalid("REGISTRY_PIN_MISMATCH", "registries", "artifact registry pins differ from the independently selected metadata")
	}
	if !sameRewriteProfile(actual.Profile, expected.Profile) {
		return invalid("REWRITE_PROFILE_MISMATCH", "profile", "artifact rewrite profile differs from the selected profile")
	}
	if actual.ContextID != expected.ContextID {
		return invalid("CONTEXT_MISMATCH", "context_id", "artifact context differs from the selected request context")
	}
	if !sameContextDependencies(actual.ContextDependencies, expected.ContextDependencies) {
		return invalid("CONTEXT_DEPENDENCY_MISMATCH", "context_dependencies", "artifact dependencies differ from the selected request context")
	}
	if !sameProvenance(actual.Provenance, expected.Provenance) {
		return invalid("PROVENANCE_MISMATCH", "provenance", "artifact provenance differs from the selected metadata")
	}
	if !sameSortDeclarations(actual.Sorts, expected.Sorts) {
		return invalid("SORT_INVENTORY_MISMATCH", "sorts", "artifact sort inventory differs from the selected metadata")
	}
	if !sameSignaturesExact(actual.Signatures, expected.Signatures) {
		return invalid("SIGNATURE_INVENTORY_MISMATCH", "signatures", "artifact signature inventory or labels differ from the selected metadata")
	}
	if !strictlySortedDeclarations(actual.Parameters) || !strictlySortedDeclarations(actual.References) {
		return invalid("NONCANONICAL_DECLARATIONS", "artifact", "parameter and reference declarations must be in canonical ID order")
	}
	return nil
}

func sameRegistryPins(left, right []RegistryPin) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] || !validText(left[i].ID, 256) || !validText(left[i].Version, 256) ||
			!validText(left[i].ResourceFingerprint, 256) || !validText(left[i].FormalDenotationID, 256) {
			return false
		}
	}
	return true
}

func sameRewriteProfile(left, right RewriteProfile) bool {
	if left.ID != right.ID || len(left.Rules) != len(right.Rules) {
		return false
	}
	for i := range left.Rules {
		if left.Rules[i] != right.Rules[i] {
			return false
		}
	}
	return true
}

func sameContextDependencies(left, right []ContextDependency) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] || !validText(left[i].ID, 256) || !validText(left[i].Fingerprint, 256) {
			return false
		}
	}
	return true
}

func sameProvenance(left, right []provenance.SourceRef) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] || left[i].Validate() != nil {
			return false
		}
	}
	return true
}

func sameSortDeclarations(left, right []SortDeclaration) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] || !validText(left[i].ID, 256) || !validOptionalText(left[i].Label, 1024) {
			return false
		}
	}
	return true
}

func sameSignaturesExact(left, right []Signature) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy, err := cloneArtifactSignatures(left)
	if err != nil {
		return false
	}
	for i := range leftCopy {
		a, b := leftCopy[i], right[i]
		if a.ID != b.ID || a.Label != b.Label || len(a.Ports) != len(b.Ports) ||
			!validText(a.ID, 256) || !validOptionalText(a.Label, 1024) {
			return false
		}
		for j := range a.Ports {
			if a.Ports[j].Role != b.Ports[j].Role || !validText(a.Ports[j].Role, 256) || !TypeEqual(a.Ports[j].Type, b.Ports[j].Type) {
				return false
			}
		}
	}
	return true
}

func strictlySortedDeclarations(values []Declaration) bool {
	for i, declaration := range values {
		if !validText(declaration.ID, 256) || (i > 0 && values[i-1].ID >= declaration.ID) {
			return false
		}
	}
	return true
}

func validateReplayDeclarations(envelope ArtifactEnvelope) error {
	for _, declarations := range [][]Declaration{envelope.Parameters, envelope.References} {
		for _, declaration := range declarations {
			if !validText(declaration.ID, 256) {
				return invalid("INVALID_ARTIFACT_DECLARATION", "declarations", "parameter and reference IDs must be bounded")
			}
		}
	}
	return nil
}

func preflightReplayNodes(nodes []Node, limits Limits) error {
	totalEdges := 0
	for i, node := range nodes {
		path := fmt.Sprintf("nodes[%d]", i)
		if len(node.Args) > limits.MaxNodes || len(node.Operands) > limits.MaxNodes || len(node.Args)+len(node.Operands) > limits.MaxNodes {
			return resourceLimit(path, "node edge count exceeds effective replay limits")
		}
		totalEdges += len(node.Args) + len(node.Operands)
		if node.Body != 0 {
			totalEdges++
		}
		if node.Function != 0 {
			totalEdges++
		}
		if node.Argument != 0 {
			totalEdges++
		}
		if totalEdges > limits.MaxNodes {
			return resourceLimit("nodes", "artifact edge count exceeds effective replay limits")
		}
		if !validOptionalText(node.Label, 1024) {
			return invalid("INVALID_NODE_LABEL", path+".label", "node label is invalid or exceeds its bound")
		}
		if node.Type != nil {
			if _, err := cloneStandaloneType(node.Type, limits); err != nil {
				return err
			}
		}
	}
	return nil
}

func intersectLimits(left, right Limits) Limits {
	return Limits{
		MaxNodes: minInt(left.MaxNodes, right.MaxNodes), MaxDepth: minInt(left.MaxDepth, right.MaxDepth),
		MaxTypeDepth: minInt(left.MaxTypeDepth, right.MaxTypeDepth), MaxSymbols: minInt(left.MaxSymbols, right.MaxSymbols),
		MaxPorts: minInt(left.MaxPorts, right.MaxPorts),
	}
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
