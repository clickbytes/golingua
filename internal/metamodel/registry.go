package metamodel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
)

// Registry is an immutable snapshot of a document and its signature index.
// It can only be created inside this package after a future validator has
// accepted the input; no public unchecked activation function is provided.
type Registry struct {
	document         Document
	symbols          map[string]Symbol
	signatures       map[string]algebra.Signature
	resourceHash     string
	formalDenotation string
}

// newValidatedRegistry snapshots a document after validation. Dependency
// resolution is deliberately outside this data container.
func newValidatedRegistry(document Document, dependencySnapshots []*Registry) (*Registry, error) {
	copyDoc, err := cloneDocument(document)
	if err != nil {
		return nil, err
	}
	resolvedDependencies, err := resolveDependencyReceipts(copyDoc.Dependencies, dependencySnapshots)
	if err != nil {
		return nil, err
	}
	resourceBytes, err := canonicalResource(copyDoc)
	if err != nil {
		return nil, err
	}
	formalBytes, err := canonicalFormal(copyDoc, resolvedDependencies)
	if err != nil {
		return nil, err
	}
	registry := &Registry{
		document:         copyDoc,
		symbols:          make(map[string]Symbol, len(copyDoc.Symbols)),
		signatures:       make(map[string]algebra.Signature, len(copyDoc.Symbols)),
		resourceHash:     digest("metamodel-resource-v1", resourceBytes),
		formalDenotation: digest("metamodel-formal-v1", formalBytes),
	}
	for _, symbol := range copyDoc.Symbols {
		registry.symbols[symbol.ID] = symbol
		if symbol.Kind == SymbolPrimitive && symbol.ReviewStatus == ReviewApproved {
			ports := append([]algebra.Port(nil), symbol.Ports...)
			registry.signatures[symbol.ID] = algebra.Signature{ID: symbol.ID, Label: symbol.Label, Ports: ports}
		}
	}
	return registry, nil
}

// Document returns a detached copy; callers cannot mutate the registry.
func (registry *Registry) Document() Document {
	if registry == nil {
		return Document{}
	}
	copyDoc, _ := cloneDocument(registry.document)
	return copyDoc
}

// Symbol returns a detached symbol record.
func (registry *Registry) Symbol(id string) (Symbol, bool) {
	if registry == nil {
		return Symbol{}, false
	}
	symbol, ok := registry.symbols[id]
	if !ok {
		return Symbol{}, false
	}
	copySymbol, err := cloneSymbol(symbol)
	return copySymbol, err == nil
}

// Signature returns the algebra interface for a primitive symbol. Alias and
// opaque records are not offered as executable primitive relations here.
func (registry *Registry) Signature(id string) (algebra.Signature, bool) {
	if registry == nil {
		return algebra.Signature{}, false
	}
	symbol, ok := registry.symbols[id]
	if !ok || symbol.Kind != SymbolPrimitive || symbol.ReviewStatus != ReviewApproved {
		return algebra.Signature{}, false
	}
	signature := registry.signatures[id]
	signature.Ports = append([]algebra.Port(nil), signature.Ports...)
	return signature, true
}

func (registry *Registry) ID() string {
	if registry == nil {
		return ""
	}
	return registry.document.ID
}

func (registry *Registry) Version() string {
	if registry == nil {
		return ""
	}
	return registry.document.Version
}

// ResourceFingerprint hashes the complete manifest projection. Metadata-only
// edits therefore change this digest.
func (registry *Registry) ResourceFingerprint() string {
	if registry == nil {
		return ""
	}
	return registry.resourceHash
}

// FormalDenotationID hashes only IDs and formal structure. It is conservative
// structural identity, not a proof of arbitrary expression equivalence.
func (registry *Registry) FormalDenotationID() string {
	if registry == nil {
		return ""
	}
	return registry.formalDenotation
}

func cloneDocument(document Document) (Document, error) {
	encoded, err := json.Marshal(document)
	if err != nil {
		return Document{}, err
	}
	var result Document
	err = json.Unmarshal(encoded, &result)
	return result, err
}

func cloneSymbol(symbol Symbol) (Symbol, error) {
	encoded, err := json.Marshal(symbol)
	if err != nil {
		return Symbol{}, err
	}
	var result Symbol
	err = json.Unmarshal(encoded, &result)
	return result, err
}

func canonicalResource(document Document) ([]byte, error) {
	sort.Slice(document.Provenance, func(i, j int) bool { return document.Provenance[i].ID < document.Provenance[j].ID })
	sort.Slice(document.Sorts, func(i, j int) bool { return document.Sorts[i].ID < document.Sorts[j].ID })
	sort.Slice(document.Symbols, func(i, j int) bool { return document.Symbols[i].ID < document.Symbols[j].ID })
	for i := range document.Symbols {
		symbol := &document.Symbols[i]
		sort.Slice(symbol.Provenance, func(a, b int) bool { return symbol.Provenance[a].ID < symbol.Provenance[b].ID })
		sort.Slice(symbol.Constraints, func(a, b int) bool {
			left, _ := json.Marshal(symbol.Constraints[a])
			right, _ := json.Marshal(symbol.Constraints[b])
			return string(left) < string(right)
		})
	}
	sort.Slice(document.Dependencies, func(i, j int) bool {
		if document.Dependencies[i].ID != document.Dependencies[j].ID {
			return document.Dependencies[i].ID < document.Dependencies[j].ID
		}
		return document.Dependencies[i].Version < document.Dependencies[j].Version
	})
	sort.Slice(document.TypicalEdges, func(i, j int) bool {
		a, b := document.TypicalEdges[i], document.TypicalEdges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.Kind < b.Kind
	})
	return json.Marshal(document)
}

type formalDependency struct {
	ID               string `json:"id"`
	FormalDenotation string `json:"formal_denotation_id"`
}

func canonicalFormal(document Document, dependencies []formalDependency) ([]byte, error) {
	type formalSymbol struct {
		ID          string         `json:"id"`
		Kind        SymbolKind     `json:"kind"`
		Ports       []algebra.Port `json:"ports,omitempty"`
		Constraints []Constraint   `json:"constraints,omitempty"`
		Alias       *Alias         `json:"alias,omitempty"`
	}
	type formalDocument struct {
		SchemaVersion int                `json:"schema_version"`
		ID            string             `json:"id"`
		Profile       string             `json:"profile"`
		SortIDs       []string           `json:"sort_ids"`
		Symbols       []formalSymbol     `json:"symbols"`
		Dependencies  []formalDependency `json:"dependencies"`
	}
	formal := formalDocument{SchemaVersion: document.SchemaVersion, ID: document.ID, Profile: document.Profile, Dependencies: append([]formalDependency(nil), dependencies...)}
	for _, sortDecl := range document.Sorts {
		formal.SortIDs = append(formal.SortIDs, sortDecl.ID)
	}
	sort.Strings(formal.SortIDs)
	for _, symbol := range document.Symbols {
		constraints := append([]Constraint(nil), symbol.Constraints...)
		sort.Slice(constraints, func(i, j int) bool {
			left, _ := json.Marshal(constraints[i])
			right, _ := json.Marshal(constraints[j])
			return string(left) < string(right)
		})
		entry := formalSymbol{ID: symbol.ID, Kind: symbol.Kind, Ports: append([]algebra.Port(nil), symbol.Ports...), Constraints: constraints}
		if symbol.Alias != nil {
			alias := *symbol.Alias
			alias.Expression = stripTermLabels(alias.Expression)
			alias.Interface = append([]PortBinding(nil), alias.Interface...)
			entry.Alias = &alias
		}
		formal.Symbols = append(formal.Symbols, entry)
	}
	sort.Slice(formal.Symbols, func(i, j int) bool { return formal.Symbols[i].ID < formal.Symbols[j].ID })
	return json.Marshal(formal)
}

func resolveDependencyReceipts(declarations []Dependency, snapshots []*Registry) ([]formalDependency, error) {
	if len(declarations) != len(snapshots) {
		return nil, fmt.Errorf("resolved dependency snapshot count %d does not match manifest dependency count %d", len(snapshots), len(declarations))
	}
	byID := make(map[string]*Registry, len(snapshots))
	for _, snapshot := range snapshots {
		if snapshot == nil {
			return nil, fmt.Errorf("resolved dependency snapshot is nil")
		}
		if _, exists := byID[snapshot.ID()]; exists {
			return nil, fmt.Errorf("duplicate resolved dependency snapshot %q", snapshot.ID())
		}
		byID[snapshot.ID()] = snapshot
	}
	receipts := make([]formalDependency, 0, len(declarations))
	seen := make(map[string]struct{}, len(declarations))
	for _, declaration := range declarations {
		if _, exists := seen[declaration.ID]; exists {
			return nil, fmt.Errorf("duplicate dependency declaration %q", declaration.ID)
		}
		seen[declaration.ID] = struct{}{}
		snapshot, exists := byID[declaration.ID]
		if !exists || snapshot.Version() != declaration.Version || snapshot.ResourceFingerprint() != declaration.Fingerprint {
			return nil, fmt.Errorf("dependency %q is unresolved or does not match its pinned version and fingerprint", declaration.ID)
		}
		receipts = append(receipts, formalDependency{ID: snapshot.ID(), FormalDenotation: snapshot.FormalDenotationID()})
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i].ID < receipts[j].ID })
	return receipts, nil
}

func stripTermLabels(term *algebra.Term) *algebra.Term {
	if term == nil {
		return nil
	}
	copyTerm := *term
	copyTerm.Label = ""
	copyTerm.Type = cloneType(term.Type)
	copyTerm.Body = stripTermLabels(term.Body)
	copyTerm.Function = stripTermLabels(term.Function)
	copyTerm.Argument = stripTermLabels(term.Argument)
	copyTerm.Operands = make([]*algebra.Term, len(term.Operands))
	for i, operand := range term.Operands {
		copyTerm.Operands[i] = stripTermLabels(operand)
	}
	copyTerm.Args = append([]algebra.Argument(nil), term.Args...)
	for i := range copyTerm.Args {
		copyTerm.Args[i].Value = stripTermLabels(term.Args[i].Value)
	}
	return &copyTerm
}

func cloneType(value *algebra.Type) *algebra.Type {
	if value == nil {
		return nil
	}
	copyType := *value
	copyType.From = cloneType(value.From)
	copyType.To = cloneType(value.To)
	return &copyType
}

func digest(domain string, value []byte) string {
	sum := sha256.Sum256(append(append([]byte(domain), 0), value...))
	return "sha256:" + hex.EncodeToString(sum[:])
}
