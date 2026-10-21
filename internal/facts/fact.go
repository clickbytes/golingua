// Package facts holds explicit, request-scoped assertions about individual
// referents. CORE definitions and utterance meanings are never world facts.
package facts

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	SchemaVersion = 1
	MaxFacts      = 4096
)

// Polarity is evidence for a proposition or its explicit denial. Missing
// evidence has neither polarity and cannot establish a negative fact.
type Polarity string

const (
	Positive Polarity = "positive"
	Negative Polarity = "negative"
)

// AssertionScope must be asserted. A question, hypothesis, or scoped
// negation cannot be silently promoted to a positive world assertion.
type AssertionScope string

const (
	ScopeAsserted     AssertionScope = "asserted"
	ScopeQuestion     AssertionScope = "question"
	ScopeHypothetical AssertionScope = "hypothetical"
	ScopeNegated      AssertionScope = "negated"
)

// SourceKind identifies the origin of one explicitly supplied assertion.
type SourceKind string

const (
	SourceUser     SourceKind = "user"
	SourceDocument SourceKind = "document"
)

// SourceEvidence identifies the request input or document that supplied a
// fact. External documents additionally carry SourceRef.
type SourceEvidence struct {
	Kind      SourceKind            `json:"kind"`
	ID        string                `json:"id"`
	Reference *provenance.SourceRef `json:"reference,omitempty"`
}

// ValidTime is a closed interval in UTC. Empty endpoints are unbounded.
// RFC3339 timestamps are required when an endpoint is present.
type ValidTime struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// Fact is a binary relation assertion over two request-local referents.
// Predicate names a CORE relation schema; it is not a lexical sense or a
// ConceptDefinition. The source is evidence, not automatic truth approval.
type Fact struct {
	ID             string                `json:"id"`
	Predicate      string                `json:"predicate"`
	ArgReferentIDs [2]meaning.ReferentID `json:"arg_referent_ids"`
	Polarity       Polarity              `json:"polarity"`
	ValidTime      ValidTime             `json:"valid_time"`
	Source         SourceEvidence        `json:"source"`
	AssertionScope AssertionScope        `json:"assertion_scope"`
}

// Input is the closed, versioned shape accepted by Build. The referenced
// CORE graph is supplied separately and cannot be selected by a fact claim.
type Input struct {
	SchemaVersion int                    `json:"schema_version"`
	Context       meaning.RequestContext `json:"context"`
	Facts         []Fact                 `json:"facts"`
}
