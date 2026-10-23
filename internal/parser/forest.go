// Package parser contains bounded construction parsing data structures.
//
// This file defines the shared packed-forest contract. Parsing and index
// compilation are implemented separately so these values remain immutable
// input/output data and carry no language-specific defaults.
package parser

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	defaultMaxNodes        = 4096
	defaultMaxAlternatives = 16384
	defaultMaxExpansions   = 65536
)

// Budget bounds packed nodes, retained parse alternatives, and search work.
// Each value must be positive and no greater than the package ceiling.
type Budget struct {
	MaxNodes        int
	MaxAlternatives int
	MaxExpansions   int
}

// NodeID identifies one packed forest node within a parse result.
type NodeID string

// DefaultBudget returns finite limits for one parse operation.
func DefaultBudget() Budget {
	return Budget{
		MaxNodes:        defaultMaxNodes,
		MaxAlternatives: defaultMaxAlternatives,
		MaxExpansions:   defaultMaxExpansions,
	}
}

// Validate rejects disabled or above-ceiling limits.
func (budget Budget) Validate() error {
	limits := []struct {
		name  string
		value int
		max   int
	}{
		{"max_nodes", budget.MaxNodes, defaultMaxNodes},
		{"max_alternatives", budget.MaxAlternatives, defaultMaxAlternatives},
		{"max_expansions", budget.MaxExpansions, defaultMaxExpansions},
	}
	for _, limit := range limits {
		if limit.value <= 0 || limit.value > limit.max {
			return newSetupError("budget."+limit.name, fmt.Sprintf("must be between 1 and %d", limit.max))
		}
	}
	return nil
}

// PlanInput bundles the mutually validated construction and valency records
// needed to interpret one compiled plan during parsing.
type PlanInput struct {
	Plan    construction.CompiledConstructionPlan
	Family  construction.ConstructionFamily
	Variant construction.Variant
	Frame   valency.ValencyFrame
	Binding valency.ConstructionBinding
}

// Index is an immutable collection of precompiled construction plans.
// compile.go owns construction and population of the private records.
type Index struct {
	plans         []compiledPlan
	byLanguage    map[features.LanguageID][]int
	byFirstSymbol map[planLookupKey][]int
}

// planLookupKey identifies a possible first realized symbol. Exactly one of
// Category or Terminal is populated; Language keeps parser inventories local.
type planLookupKey struct {
	Language features.LanguageID
	Category construction.SlotCategory
	Terminal string
}

type compiledPlan struct {
	plan       construction.CompiledConstructionPlan
	family     construction.ConstructionFamily
	variant    construction.Variant
	frame      valency.ValencyFrame
	permission construction.SensePermission
	headSlotID string
	slots      map[string]construction.SlotSpec
	sequence   []compiledSymbol
}

type compiledSymbol struct {
	Position       int
	Slot           *construction.SlotSpec
	Terminal       string
	Optional       bool
	AcceptedInputs []string
}

// PhraseCandidate is lexical or phrase-level evidence that may fill a
// construction slot. Token ranges are half-open indexes into the input
// lattice; Span is a half-open UTF-8 byte range in the original source.
type PhraseCandidate struct {
	ID         string
	Category   construction.SlotCategory
	TokenStart int
	TokenEnd   int
	Span       text.Span
	// HeadTokenIndex identifies the token whose readings supply the predicate
	// sense for VP candidates. It must fall within the candidate's token range.
	HeadTokenIndex int
	ReadingIDs     []string
	LexemeID       ids.LexemeID
	SenseIDs       []ids.SenseID
	Features       map[features.FeatureID]features.FeatureValueID
	Case           string
	CaseKnown      bool
}

// WHGapEvidence explicitly licenses an omitted semantic argument for a WH
// question. SlotID identifies the construction slot and Role its frame role.
type WHGapEvidence struct {
	SlotID      string
	Role        meaning.Role
	ScopeNodeID meaning.NodeID
}

// ParseContext carries request-scoped evidence used by variant conditions.
type ParseContext struct {
	WHGaps               []WHGapEvidence
	InformationStructure map[meaning.Role]construction.InformationValue
}

// SlotBinding records which candidate or explicit WH gap fills a plan slot.
// CandidateIDs can contain multiple candidates when slot cardinality permits.
type SlotBinding struct {
	SlotID       string
	Role         meaning.Role
	CandidateIDs []string
	NodeIDs      []NodeID
	WHGap        *WHGapEvidence
}

// ParseAlternative is one surviving plan interpretation. Unresolved contains
// slot IDs whose evidence was unavailable, preserving partial parses without
// treating them as complete.
type ParseAlternative struct {
	PlanID       construction.ConstructionPlanID
	FamilyID     ids.ConstructionID
	VariantID    construction.VariantID
	Permission   construction.SensePermission
	Slots        []SlotBinding
	Unresolved   []string
	DerivationID string
}

// PackedNode shares a token subspan across candidate categories and parse
// alternatives rather than materializing duplicate trees.
type PackedNode struct {
	ID           NodeID
	Start        int
	End          int
	Span         text.Span
	Category     construction.SlotCategory
	CandidateIDs []string
	Alternatives []ParseAlternative
}

// UnmatchedCandidate explains why a supplied phrase candidate is not used by
// any retained parse alternative.
type UnmatchedCandidate struct {
	CandidateID string
	TokenStart  int
	TokenEnd    int
	Reason      string
}

// Forest is a bounded packed set of construction parses over one source.
// Roots reference packed node IDs; an incomplete outcome means search or
// retention stopped before all licensed alternatives were considered.
type Forest struct {
	Language    features.LanguageID
	Nodes       []PackedNode
	Roots       []NodeID
	Unmatched   []UnmatchedCandidate
	Outcome     status.OutcomeKind
	Diagnostics []status.Diagnostic
}

// ParseError reports invalid parser setup as a typed status result.
type ParseError struct {
	result status.Result
}

func (err *ParseError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "parser setup error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns the typed status result with a defensive diagnostics copy.
func (err *ParseError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}

func newParserError(path, message string) *ParseError {
	return &ParseError{result: status.Result{
		Status: status.OutcomeInvalid,
		Diagnostics: []status.Diagnostic{{
			Code:     "INVALID_PARSER_SETUP",
			Message:  message,
			Path:     path,
			Severity: status.SeverityError,
		}},
	}}
}

func newSetupError(path, message string) *ParseError { return newParserError(path, message) }
