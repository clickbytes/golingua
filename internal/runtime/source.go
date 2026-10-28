package runtime

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

// PhraseEvidenceBuilder supplies language-local phrase candidates and parser
// context. It is required because lexical readings do not determine phrase
// categories, feature labels, or role structure on their own.
type PhraseEvidenceBuilder interface {
	Build(context.Context, analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error)
}

// CompositionEvidence is request-scoped evidence that cannot be inferred from
// lexical identity, including referent identity and construction operators.
type CompositionEvidence struct {
	References               []semantic.ReferenceEvidence
	ImplicitRoles            []semantic.ImplicitRoleEvidence
	NodeAttributes           []semantic.NodeAttributesEvidence
	RequestContextID         string
	Operators                []semantic.OperatorEvidence
	OperatorCoverageComplete bool
}

// CompositionEvidenceBuilder may attach explicit discourse and operator
// evidence to one parse. A nil builder means no such evidence is supplied.
type CompositionEvidenceBuilder interface {
	Build(context.Context, analyzer.LexicalLattice, []parser.PhraseCandidate, parser.Forest) (CompositionEvidence, error)
}

// SourcePipeline adapts the existing bounded source stages to engine.SourceAnalyzer.
// Language-specific phrase analysis remains an explicit compiled dependency.
type SourcePipeline struct {
	Language features.LanguageID
	// ClauseConnector is one exact, language-local source terminal that may
	// join two bounded declarative clauses in DSE syntax analysis.
	ClauseConnector  string
	Tokenize         text.TokenizeOptions
	AllowFinalPeriod bool
	Morphology       *morph.InverseIndex
	Lexicon          *lexicon.Index
	MWEPlans         []construction.CompiledMWEPlan
	LatticeBudget    analyzer.Budget
	PhraseBuilder    PhraseEvidenceBuilder
	ParserIndex      *parser.Index
	ParseBudget      parser.Budget
	Valency          *valency.Index
	Core             *core.ConceptGraph
	Constructions    *construction.Registry
	SemanticBudget   semantic.Budget
	EvidenceBuilder  CompositionEvidenceBuilder
	// AllowedUnlexicalizedTerminals names exact function words whose evidence
	// comes from a matched construction terminal rather than a lexical reading.
	AllowedUnlexicalizedTerminals map[string]bool
}

var _ engine.SourceAnalyzer = (*SourcePipeline)(nil)

// Analyze executes lexical analysis, phrase parsing, semantic composition and
// semantic validation without selecting among competing meanings.
func (pipeline *SourcePipeline) Analyze(ctx context.Context, input string) (semantic.Result, error) {
	if pipeline == nil || pipeline.PhraseBuilder == nil || pipeline.ParserIndex == nil || pipeline.Core == nil || pipeline.Lexicon == nil || pipeline.Valency == nil || pipeline.Constructions == nil {
		return semantic.Result{}, pipelineError(status.OutcomeUnsupported, "SOURCE_PIPELINE_UNAVAILABLE", "source_pipeline", "source pipeline requires verified lexical, parser, semantic indexes and a language phrase evidence builder")
	}
	if ctx == nil {
		return semantic.Result{}, pipelineError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "source analysis requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return semantic.Result{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
	}
	syntax, err := pipeline.AnalyzeSyntax(ctx, input)
	if err != nil {
		return semantic.Result{}, err
	}
	semanticBudget := pipeline.SemanticBudget
	if semanticBudget == (semantic.Budget{}) {
		semanticBudget = semantic.DefaultBudget()
	}
	composed, err := semantic.Compose(semantic.CompositionInput{
		Forest: syntax.Forest, Candidates: syntax.PhraseCandidates, Lattice: syntax.Lattice, SourceBytes: syntax.OriginalSourceBytes,
		Lexicon: pipeline.Lexicon, Valency: pipeline.Valency, Core: pipeline.Core,
		Constructions: pipeline.Constructions, References: syntax.CompositionEvidence.References,
		ImplicitRoles:            syntax.CompositionEvidence.ImplicitRoles,
		NodeAttributes:           syntax.CompositionEvidence.NodeAttributes,
		RequestContextID:         syntax.CompositionEvidence.RequestContextID,
		Operators:                syntax.CompositionEvidence.Operators,
		OperatorCoverageComplete: syntax.CompositionEvidence.OperatorCoverageComplete,
		Budget:                   semanticBudget,
	})
	if err != nil {
		return semantic.Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return semantic.Result{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
	}
	validated, err := semantic.Disambiguate(composed.Candidates, composed.Outcome, composed.Diagnostics, pipeline.Core, semanticBudget.Evidence)
	if err != nil {
		return semantic.Result{}, fmt.Errorf("validate source semantic alternatives: %w", err)
	}
	return validated, nil
}

func attestConstructionTerminals(lattice *analyzer.LexicalLattice, allowed map[string]bool) {
	if lattice == nil || len(allowed) == 0 {
		return
	}
	removedUnknownPaths := make(map[string]bool)
	for index := range lattice.Tokens {
		token := &lattice.Tokens[index]
		if token.Outcome != status.OutcomeUnknown || len(token.Readings) != 0 || token.Token.Kind != text.TokenWord {
			continue
		}
		terminal := token.Token.Folded
		if terminal == "" {
			terminal = token.Token.Normalized
		}
		if !allowed[terminal] {
			continue
		}
		token.Outcome = status.OutcomeOK
		removedUnknownPaths[fmt.Sprintf("tokens[%d].span[%d:%d]", index, token.Token.Span.Start, token.Token.Span.End)] = true
	}
	if len(removedUnknownPaths) == 0 {
		return
	}
	remainingUnknown := false
	for _, token := range lattice.Tokens {
		if token.Outcome == status.OutcomeUnknown || token.Outcome == status.OutcomeIncomplete {
			remainingUnknown = true
			break
		}
	}
	if lattice.Outcome == status.OutcomeUnknown && !remainingUnknown {
		lattice.Outcome = status.OutcomeOK
	}
	diagnostics := lattice.Diagnostics[:0]
	for _, diagnostic := range lattice.Diagnostics {
		if diagnostic.Code == "UNKNOWN_SURFACE" && removedUnknownPaths[diagnostic.Path] {
			continue
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	lattice.Diagnostics = diagnostics
}

type sourcePipelineError struct{ result status.Result }

func (failure *sourcePipelineError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "source pipeline unavailable"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *sourcePipelineError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func pipelineError(kind status.OutcomeKind, code, path, message string) error {
	result, err := status.NewResult(kind, "", []status.Diagnostic{{
		Code: code, Path: path, Message: message, Severity: status.SeverityError,
	}}, nil)
	if err != nil {
		return &sourcePipelineError{result: status.Result{Status: status.OutcomeError}}
	}
	return &sourcePipelineError{result: result}
}
