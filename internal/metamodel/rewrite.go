package metamodel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
)

// NormalizationEnvironment builds a trusted rewrite context from explicitly
// selected, previously validated registry snapshots. The returned aliases are
// typing definitions for normalization only; they are not operational
// primitive signatures and are never admitted by DeclarationContext.
func NormalizationEnvironment(document algebra.Document, registries []*Registry) (algebra.RewriteEnvironment, error) {
	if len(registries) == 0 {
		return algebra.RewriteEnvironment{}, invalid("EMPTY_REGISTRY_SET", "registries", "normalization requires at least one selected registry")
	}
	hydrated, symbols, err := declarationContext(document, registries)
	if err != nil {
		return algebra.RewriteEnvironment{}, err
	}

	selected := append([]*Registry(nil), registries...)
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].ID() != selected[j].ID() {
			return selected[i].ID() < selected[j].ID()
		}
		return selected[i].FormalDenotationID() < selected[j].FormalDenotationID()
	})
	type formalSnapshotEntry struct {
		RegistryID         string `json:"registry_id"`
		FormalDenotationID string `json:"formal_denotation_id"`
	}
	entries := make([]formalSnapshotEntry, 0, len(selected))
	seenRegistryIDs := make(map[string]bool, len(selected))
	for i, registry := range selected {
		if seenRegistryIDs[registry.ID()] {
			return algebra.RewriteEnvironment{}, invalid("DUPLICATE_REGISTRY", fmt.Sprintf("registries[%d]", i), "selected registry IDs must be unique")
		}
		seenRegistryIDs[registry.ID()] = true
		entries = append(entries, formalSnapshotEntry{RegistryID: registry.ID(), FormalDenotationID: registry.FormalDenotationID()})
	}
	formalBytes, err := json.Marshal(entries)
	if err != nil {
		return algebra.RewriteEnvironment{}, invalid("REWRITE_SNAPSHOT_ENCODING", "registries", err.Error())
	}
	snapshotInput := append([]byte("golingua-rewrite-formal-snapshot-v1\x00"), formalBytes...)
	digest := sha256.Sum256(snapshotInput)
	snapshotID := "rewrite-formal-v1:" + hex.EncodeToString(digest[:])

	aliasIDs := make([]string, 0)
	for id, symbol := range symbols {
		if symbol.Kind == SymbolAlias && symbol.ReviewStatus == ReviewApproved {
			aliasIDs = append(aliasIDs, id)
		}
	}
	sort.Strings(aliasIDs)
	aliases := make([]algebra.AliasDefinition, 0, len(aliasIDs))
	for _, id := range aliasIDs {
		symbol := symbols[id]
		if symbol.Alias == nil || symbol.Alias.Expression == nil {
			return algebra.RewriteEnvironment{}, invalid("APPROVED_ALIAS_DEFINITION_MISSING", "registries.symbols."+id, "approved alias has no formal definition")
		}
		copySymbol, cloneErr := cloneSymbol(symbol)
		if cloneErr != nil {
			return algebra.RewriteEnvironment{}, invalid("ALIAS_DEFINITION_COPY", "registries.symbols."+id, cloneErr.Error())
		}
		ports := append([]algebra.Port(nil), copySymbol.Ports...)
		bindings := make([]algebra.PortBinding, len(copySymbol.Alias.Interface))
		for i, binding := range copySymbol.Alias.Interface {
			bindings[i] = algebra.PortBinding{Port: binding.Port, Binder: binding.Binder}
		}
		aliases = append(aliases, algebra.AliasDefinition{
			ID: id, Ports: ports, Expression: copySymbol.Alias.Expression, Interface: bindings,
		})
		hydrated.Signatures = append(hydrated.Signatures, algebra.Signature{ID: id, Label: copySymbol.Label, Ports: append([]algebra.Port(nil), copySymbol.Ports...)})
	}
	sort.Slice(hydrated.Signatures, func(i, j int) bool { return hydrated.Signatures[i].ID < hydrated.Signatures[j].ID })
	environment := algebra.RewriteEnvironment{Document: hydrated, Aliases: aliases, SnapshotID: snapshotID}
	if err := algebra.ValidateRewriteEnvironment(environment); err != nil {
		return algebra.RewriteEnvironment{}, err
	}
	return environment, nil
}
