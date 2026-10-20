// Package analyzer builds bounded lexical alternatives over source text.
package analyzer

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

const (
	defaultMaxTokens     = 256
	defaultMaxCandidates = 256
	defaultMaxMWEPlans   = 256
)

// Budget bounds input scanning and the alternatives retained in a lexical
// lattice. Values must be positive and cannot exceed the package ceilings.
type Budget struct {
	MaxInputBytes int
	MaxTokens     int
	MaxCandidates int
	MaxMWEPlans   int
}

// DefaultBudget returns the finite default limits for one lexical scan.
func DefaultBudget() Budget {
	return Budget{
		MaxInputBytes: status.DefaultBounds().MaxInputBytes,
		MaxTokens:     defaultMaxTokens,
		MaxCandidates: defaultMaxCandidates,
		MaxMWEPlans:   defaultMaxMWEPlans,
	}
}

// Validate rejects zero, negative, or above-ceiling limits so callers cannot
// accidentally disable the protections provided by the default budget.
func (budget Budget) Validate() error {
	defaults := DefaultBudget()
	limits := []struct {
		name           string
		value, maximum int
	}{
		{"max_input_bytes", budget.MaxInputBytes, defaults.MaxInputBytes},
		{"max_tokens", budget.MaxTokens, defaults.MaxTokens},
		{"max_candidates", budget.MaxCandidates, defaults.MaxCandidates},
		{"max_mwe_plans", budget.MaxMWEPlans, defaults.MaxMWEPlans},
	}
	for _, limit := range limits {
		if limit.value <= 0 || limit.value > limit.maximum {
			return newLatticeSetupError("budget."+limit.name, fmt.Sprintf("must be between 1 and %d", limit.maximum))
		}
	}
	return nil
}

// LexicalLattice retains all licensed morphology and sense alternatives found
// over one input. A non-OK Outcome may still carry partial results when a
// candidate budget truncated the inventory.
type LexicalLattice struct {
	Language    features.LanguageID
	Tokens      []TokenNode
	MWEs        []MWEEdge
	Outcome     status.OutcomeKind
	Diagnostics []status.Diagnostic
}

// TokenNode anchors alternatives to a token's source span and source surface.
type TokenNode struct {
	Index    int
	Token    text.Token
	Outcome  status.OutcomeKind
	Readings []ReadingCandidate
}

// ReadingCandidate joins one morphological analysis to its lexical record and
// all known senses without treating missing semantic approval as absence.
type ReadingCandidate struct {
	Reading   morph.Reading
	Lexeme    lexicon.MorphLexeme
	HasLexeme bool
	State     CandidateState
	Senses    []SenseCandidate
}

// CandidateState distinguishes unavailable lexical identity from missing
// lexical sense data.
type CandidateState string

const (
	CandidateKnown         CandidateState = "known"
	CandidateUnknownLexeme CandidateState = "unknown_lexeme"
	CandidateUnknownSense  CandidateState = "unknown_sense"
)

// SenseCandidate retains a lexical sense and all of its semantic bindings,
// including bindings that do not authorize transfer.
type SenseCandidate struct {
	Sense    lexicon.LexicalSense
	Bindings []lexicon.SemanticBinding
}

// MWEEdge connects a complete MWE lexeme/sense plan to a contiguous token
// range while retaining component-level source matches.
type MWEEdge struct {
	PlanID     string
	LexemeID   ids.LexemeID
	SenseID    ids.SenseID
	TokenStart int
	TokenEnd   int
	Span       text.Span
	Components []MWEComponentMatch
}

// MWEComponentMatch records one matched component and every morphology reading
// that supports that component at its token position.
type MWEComponentMatch struct {
	ComponentID string
	TokenIndex  int
	ReadingIDs  []string
}

// LatticeError wraps a typed result for invalid lattice setup.
type LatticeError struct {
	result status.Result
}

// NewLatticeError wraps a status result for a lattice configuration failure.
func NewLatticeError(result status.Result) *LatticeError {
	return &LatticeError{result: cloneStatusResult(result)}
}

func (err *LatticeError) Error() string {
	if err == nil {
		return "lexical lattice setup error"
	}
	if len(err.result.Diagnostics) != 0 {
		return err.result.Diagnostics[0].Message
	}
	return "lexical lattice setup error"
}

// Result returns the wrapped result with a defensive diagnostics copy.
func (err *LatticeError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	return cloneStatusResult(err.result)
}

func newLatticeSetupError(path, message string) *LatticeError {
	return NewLatticeError(status.Result{
		Status: status.OutcomeInvalid,
		Diagnostics: []status.Diagnostic{{
			Code:     "INVALID_LATTICE_SETUP",
			Message:  message,
			Path:     path,
			Severity: status.SeverityError,
		}},
	})
}

func cloneStatusResult(result status.Result) status.Result {
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}
