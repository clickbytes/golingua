package facts

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
)

const (
	DefaultNaryMaxFacts     = 4096
	DefaultNaryMaxReferents = 512
)

// NaryValue is one explicit argument identity. Only Ref is a ground identity;
// the other algebra term kinds remain representable so the validator can
// refuse them with a precise input location.
type NaryValue struct {
	Kind algebra.TermKind `json:"kind"`
	ID   string           `json:"id"`
}

// NaryArgument assigns a value to a named relation role. Argument order is
// canonicalized against the selected signature by the validator.
type NaryArgument struct {
	Role  string    `json:"role"`
	Value NaryValue `json:"value"`
}

// NaryAtom is a grounded or proposed n-ary relation application. Its symbol
// identifies a relation declaration, never a concept definition.
type NaryAtom struct {
	Symbol string         `json:"symbol"`
	Args   []NaryArgument `json:"args"`
}

// NaryFact is explicit evidence for one atom. The existing evidence, polarity,
// time, and scope types retain their established meanings.
type NaryFact struct {
	ID             string         `json:"id"`
	Atom           NaryAtom       `json:"atom"`
	Polarity       Polarity       `json:"polarity"`
	ValidTime      ValidTime      `json:"valid_time"`
	Source         SourceEvidence `json:"source"`
	AssertionScope AssertionScope `json:"assertion_scope"`
}

// NaryInput is a request-scoped set of assertions over explicitly declared
// referents. SnapshotID is an optional expected pin; it does not load a schema.
type NaryInput struct {
	SchemaVersion int                   `json:"schema_version"`
	SnapshotID    string                `json:"snapshot_id,omitempty"`
	ContextID     string                `json:"context_id"`
	Referents     []algebra.Declaration `json:"referents"`
	Facts         []NaryFact            `json:"facts"`
	At            string                `json:"at,omitempty"`
	Limits        NaryLimits            `json:"limits,omitempty"`
}

// NarySchema contains the actual selected relation declarations. Expression,
// parameters, and references are not part of this registry view.
type NarySchema struct {
	Document   algebra.Document `json:"document"`
	SnapshotID string           `json:"snapshot_id"`
}

// NaryLimits bounds one request's fact and referent inventories.
type NaryLimits struct {
	MaxFacts     int `json:"max_facts,omitempty"`
	MaxReferents int `json:"max_referents,omitempty"`
}

// NormalizeNaryLimits applies conservative defaults and fixed package caps.
func NormalizeNaryLimits(requested NaryLimits) (NaryLimits, error) {
	if requested.MaxFacts == 0 {
		requested.MaxFacts = DefaultNaryMaxFacts
	}
	if requested.MaxReferents == 0 {
		requested.MaxReferents = DefaultNaryMaxReferents
	}
	if requested.MaxFacts < 1 || requested.MaxFacts > DefaultNaryMaxFacts {
		return NaryLimits{}, factError("INVALID_FACT_BUDGET", "limits.max_facts", "must be between 1 and the package ceiling")
	}
	if requested.MaxReferents < 1 || requested.MaxReferents > DefaultNaryMaxReferents {
		return NaryLimits{}, factError("INVALID_FACT_BUDGET", "limits.max_referents", "must be between 1 and the package ceiling")
	}
	return requested, nil
}

// NaryKnowledge describes the state of supplied evidence, not world truth.
type NaryKnowledge string

const (
	NaryUnknown  NaryKnowledge = "unknown"
	NaryPositive NaryKnowledge = "positive"
	NaryNegative NaryKnowledge = "negative"
	NaryBoth     NaryKnowledge = "both"
)

// NaryEvidence preserves both explicit polarities and their source records.
type NaryEvidence struct {
	State    NaryKnowledge `json:"state"`
	Positive []NaryFact    `json:"positive"`
	Negative []NaryFact    `json:"negative"`
}

// NaryConflict names evidence with opposing polarity for one atom. It is a
// data conflict, not a logical explosion or an inferred result.
type NaryConflict struct {
	Atom        NaryAtom `json:"atom"`
	PositiveIDs []string `json:"positive_ids"`
	NegativeIDs []string `json:"negative_ids"`
}

// NaryView is an immutable request-local snapshot created only after a later
// validator has checked the complete fact and role set.
type NaryView struct {
	snapshotID string
	contextID  string
	referents  []algebra.Declaration
	facts      []NaryFact
	byID       map[string]int
	schema     algebra.Document
	at         string
	limits     NaryLimits
}

// newValidatedNaryView is the private activation boundary for the n-ary
// validator. It performs bounded structural copying, not fact semantics.
func newValidatedNaryView(schema NarySchema, input NaryInput, canonicalFacts []NaryFact) (*NaryView, error) {
	if input.SchemaVersion != SchemaVersion {
		return nil, factError("INVALID_FACT_SCHEMA_VERSION", "schema_version", "unsupported n-ary fact input schema version")
	}
	if !validNaryID(schema.SnapshotID) {
		return nil, factError("INVALID_FACT_SNAPSHOT", "schema.snapshot_id", "selected schema requires a bounded snapshot ID")
	}
	if input.SnapshotID != "" && (!validNaryID(input.SnapshotID) || input.SnapshotID != schema.SnapshotID) {
		return nil, factError("FACT_SNAPSHOT_MISMATCH", "snapshot_id", "request snapshot pin does not match the selected schema")
	}
	if !validNaryID(input.ContextID) {
		return nil, factError("INVALID_FACT_CONTEXT", "context_id", "context ID must be bounded valid UTF-8 text")
	}
	if len(input.At) > 128 || !utf8.ValidString(input.At) {
		return nil, factError("INVALID_FACT_QUERY_TIME", "at", "query instant must be valid UTF-8 and bounded")
	}
	limits, err := NormalizeNaryLimits(input.Limits)
	if err != nil {
		return nil, err
	}
	facts := canonicalFacts
	if facts == nil {
		facts = input.Facts
	}
	if len(facts) > limits.MaxFacts || len(input.Referents) > limits.MaxReferents {
		return nil, factLimit("facts_or_referents", len(facts)+len(input.Referents))
	}
	if len(schema.Document.Sorts)+len(schema.Document.Signatures)+len(input.Referents) > algebra.DefaultMaxSymbols {
		return nil, factLimit("schema_or_referents", len(schema.Document.Sorts)+len(schema.Document.Signatures)+len(input.Referents))
	}
	if schema.Document.SchemaVersion != algebra.SchemaVersion || schema.Document.Expression != nil || len(schema.Document.Parameters) != 0 || len(schema.Document.References) != 0 {
		return nil, factError("INVALID_FACT_SCHEMA_DOCUMENT", "schema.document", "schema must contain declarations only, without expression, parameters, or references")
	}
	if len(schema.Document.Sorts) > algebra.DefaultMaxSymbols || len(schema.Document.Signatures) > algebra.DefaultMaxSymbols {
		return nil, factLimit("schema.document", len(schema.Document.Sorts)+len(schema.Document.Signatures))
	}
	for i, declaration := range input.Referents {
		if !validNaryID(declaration.ID) {
			return nil, factError("INVALID_REFERENT_ID", fmt.Sprintf("referents[%d].id", i), "referent ID must be bounded valid UTF-8 text")
		}
	}
	// Check raw types before any defensive type clone, so cyclic or deeply
	// nested programmatic Type values are rejected by the bounded checker.
	declarations := schema.Document
	declarations.References = input.Referents
	if err := algebra.CheckDeclarations(declarations); err != nil {
		return nil, err
	}
	schemaCopy, err := cloneNarySchemaDocument(schema.Document)
	if err != nil {
		return nil, err
	}
	referents, err := cloneNaryDeclarations(input.Referents)
	if err != nil {
		return nil, err
	}
	copyFacts := make([]NaryFact, len(facts))
	byID := make(map[string]int, len(facts))
	for i, fact := range facts {
		if !validLocalID(fact.ID) {
			return nil, factError("INVALID_FACT_ID", fmt.Sprintf("facts[%d].id", i), "fact ID must be a bounded explicit identifier")
		}
		if !validNaryID(fact.Atom.Symbol) || len(fact.Atom.Args) > algebra.DefaultMaxPorts {
			return nil, factError("INVALID_FACT_ATOM", fmt.Sprintf("facts[%d].atom", i), "atom symbol or argument count exceeds its bound")
		}
		for j, argument := range fact.Atom.Args {
			if !validNaryID(argument.Role) || !validNaryID(argument.Value.ID) {
				return nil, factError("INVALID_FACT_ARGUMENT", fmt.Sprintf("facts[%d].atom.args[%d]", i, j), "role and value identifiers must be bounded valid UTF-8 text")
			}
		}
		if _, exists := byID[fact.ID]; exists {
			return nil, factError("DUPLICATE_FACT_ID", fmt.Sprintf("facts[%d].id", i), "fact ID is duplicated")
		}
		byID[fact.ID] = i
		copyFacts[i] = cloneNaryFact(fact)
	}
	return &NaryView{
		snapshotID: schema.SnapshotID, contextID: input.ContextID,
		referents: referents, facts: copyFacts, byID: byID,
		schema: schemaCopy, at: input.At, limits: limits,
	}, nil
}

// SnapshotID returns the immutable selected registry snapshot pin.
func (view *NaryView) SnapshotID() string {
	if view == nil {
		return ""
	}
	return view.snapshotID
}

// ContextID returns the request-local fact context ID.
func (view *NaryView) ContextID() string {
	if view == nil {
		return ""
	}
	return view.contextID
}

// Schema returns detached selected sort and relation declarations.
func (view *NaryView) Schema() algebra.Document {
	if view == nil {
		return algebra.Document{}
	}
	copyOf, err := cloneNarySchemaDocument(view.schema)
	if err != nil {
		return algebra.Document{}
	}
	return copyOf
}

// Referents returns detached declarations for request-local referents.
func (view *NaryView) Referents() []algebra.Declaration {
	if view == nil {
		return nil
	}
	result, err := cloneNaryDeclarations(view.referents)
	if err != nil {
		return nil
	}
	return result
}

// Facts returns detached records in canonical order.
func (view *NaryView) Facts() []NaryFact {
	if view == nil {
		return nil
	}
	return cloneNaryFacts(view.facts)
}

// ByID returns a detached record by its request-local fact ID.
func (view *NaryView) ByID(id string) (NaryFact, bool) {
	if view == nil {
		return NaryFact{}, false
	}
	index, ok := view.byID[id]
	if !ok {
		return NaryFact{}, false
	}
	return cloneNaryFact(view.facts[index]), true
}

// Count reports the number of retained records.
func (view *NaryView) Count() int {
	if view == nil {
		return 0
	}
	return len(view.facts)
}

// At returns the optional query instant carried by the request.
func (view *NaryView) At() string {
	if view == nil {
		return ""
	}
	return view.at
}

func cloneNarySchemaDocument(document algebra.Document) (algebra.Document, error) {
	copyOf := algebra.Document{SchemaVersion: document.SchemaVersion, Closed: document.Closed, Limits: document.Limits}
	copyOf.Sorts = append([]algebra.SortDeclaration{}, document.Sorts...)
	copyOf.Signatures = make([]algebra.Signature, len(document.Signatures))
	for i, signature := range document.Signatures {
		copyOf.Signatures[i] = algebra.Signature{ID: signature.ID, Label: signature.Label, Ports: make([]algebra.Port, len(signature.Ports))}
		for j, port := range signature.Ports {
			typ, err := cloneNaryType(port.Type)
			if err != nil {
				return algebra.Document{}, err
			}
			copyOf.Signatures[i].Ports[j] = algebra.Port{Role: port.Role, Type: typ}
		}
	}
	return copyOf, nil
}

func cloneNaryDeclarations(values []algebra.Declaration) ([]algebra.Declaration, error) {
	result := make([]algebra.Declaration, len(values))
	for i, value := range values {
		typ, err := cloneNaryType(value.Type)
		if err != nil {
			return nil, err
		}
		result[i] = algebra.Declaration{ID: value.ID, Type: typ}
	}
	return result, nil
}

func cloneNaryType(value algebra.Type) (algebra.Type, error) {
	return cloneNaryTypeNode(&value, 1, make(map[*algebra.Type]bool))
}

func cloneNaryTypeNode(value *algebra.Type, depth int, active map[*algebra.Type]bool) (algebra.Type, error) {
	if value == nil {
		return algebra.Type{}, nil
	}
	if depth > algebra.DefaultMaxTypeDepth {
		return algebra.Type{}, factLimit("schema.type", depth)
	}
	if active[value] {
		return algebra.Type{}, factError("INVALID_FACT_SCHEMA_TYPE", "schema.type", "cyclic type declaration")
	}
	active[value] = true
	defer delete(active, value)
	result := algebra.Type{Kind: value.Kind, SortID: value.SortID}
	from, err := cloneNaryTypeNode(value.From, depth+1, active)
	if err != nil {
		return algebra.Type{}, err
	}
	to, err := cloneNaryTypeNode(value.To, depth+1, active)
	if err != nil {
		return algebra.Type{}, err
	}
	if value.From != nil {
		result.From = &from
	}
	if value.To != nil {
		result.To = &to
	}
	return result, nil
}

func cloneNaryFact(fact NaryFact) NaryFact {
	fact.Atom.Args = append([]NaryArgument{}, fact.Atom.Args...)
	if fact.Source.Reference != nil {
		reference := *fact.Source.Reference
		fact.Source.Reference = &reference
	}
	return fact
}

func cloneNaryFacts(facts []NaryFact) []NaryFact {
	result := make([]NaryFact, len(facts))
	for i, fact := range facts {
		result[i] = cloneNaryFact(fact)
	}
	return result
}

func validNaryID(value string) bool {
	if len(value) == 0 || len(value) > 256 || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.' || character == ':') {
			return false
		}
	}
	return true
}
