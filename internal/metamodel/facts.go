package metamodel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/facts"
)

type factRegistryReceipt struct {
	ID                  string `json:"id"`
	Version             string `json:"version"`
	ResourceFingerprint string `json:"resource_fingerprint"`
	FormalDenotationID  string `json:"formal_denotation_id"`
}

// FactsSchema builds a detached, request-independent declaration view for
// explicit n-ary evidence from selected validated registry snapshots. It
// exposes only approved primitive relation signatures.
func FactsSchema(registries []*Registry) (facts.NarySchema, error) {
	if len(registries) == 0 {
		return facts.NarySchema{}, invalid("EMPTY_FACT_REGISTRY_SET", "registries", "facts require at least one explicitly selected registry")
	}
	if len(registries) > maxPackages {
		return facts.NarySchema{}, resource("RESOURCE_LIMIT", "registries", "selected registry count exceeds limit 16")
	}
	seen := make(map[string]struct{}, len(registries))
	receipts := make([]factRegistryReceipt, len(registries))
	for index, registry := range registries {
		if registry == nil {
			return facts.NarySchema{}, invalid("NIL_FACT_REGISTRY", fmt.Sprintf("registries[%d]", index), "selected registry is nil")
		}
		id, version := registry.ID(), registry.Version()
		resourceFingerprint, formalID := registry.ResourceFingerprint(), registry.FormalDenotationID()
		if id == "" || version == "" || resourceFingerprint == "" || formalID == "" {
			return facts.NarySchema{}, invalid("UNVALIDATED_FACT_REGISTRY", fmt.Sprintf("registries[%d]", index), "selected registry lacks a validated identity and fingerprints")
		}
		if _, exists := seen[id]; exists {
			return facts.NarySchema{}, invalid("DUPLICATE_FACT_REGISTRY", fmt.Sprintf("registries[%d]", index), "selected registry ID is duplicated")
		}
		seen[id] = struct{}{}
		receipts[index] = factRegistryReceipt{ID: id, Version: version, ResourceFingerprint: resourceFingerprint, FormalDenotationID: formalID}
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].ID < receipts[j].ID })
	serialized, err := json.Marshal(receipts)
	if err != nil {
		return facts.NarySchema{}, invalid("FACT_SNAPSHOT_ENCODING_FAILED", "registries", "selected registry snapshot could not be fingerprinted")
	}
	digest := sha256.Sum256(serialized)
	snapshotID := "facts-snapshot-v1:" + hex.EncodeToString(digest[:])

	declarations, err := DeclarationContext(algebra.Document{SchemaVersion: algebra.SchemaVersion}, registries)
	if err != nil {
		return facts.NarySchema{}, err
	}
	// DeclarationContext returns only actual sorts and approved primitive
	// signatures. Keep the fact schema as a declaration-only document.
	declarations.Expression = nil
	declarations.Parameters = nil
	declarations.References = nil
	declarations.Closed = false
	return facts.NarySchema{Document: declarations, SnapshotID: snapshotID}, nil
}
