package runtime

import (
	"context"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
)

// SourceSyntax is the output of lexical analysis and bounded phrase parsing.
// It preserves the original source byte count even when the scanner omits an
// allowed terminal period from its token input.
type SourceSyntax struct {
	Lattice             analyzer.LexicalLattice
	PhraseCandidates    []parser.PhraseCandidate
	ParseContext        parser.ParseContext
	Forest              parser.Forest
	CompositionEvidence CompositionEvidence
	OriginalSourceBytes int
}

// AnalyzeSyntax runs the shared lexical, phrase, parser, and evidence stages
// without composing or disambiguating semantic meanings. It does not require a
// CORE graph, valency index, or construction registry.
func (pipeline *SourcePipeline) AnalyzeSyntax(ctx context.Context, input string) (SourceSyntax, error) {
	if pipeline == nil || pipeline.PhraseBuilder == nil || pipeline.ParserIndex == nil || pipeline.Lexicon == nil {
		return SourceSyntax{}, pipelineError(status.OutcomeUnsupported, "SOURCE_SYNTAX_UNAVAILABLE", "source_pipeline", "source syntax requires lexical and parser indexes and a language phrase evidence builder")
	}
	if ctx == nil {
		return SourceSyntax{}, pipelineError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "source syntax analysis requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return SourceSyntax{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
	}

	latticeBudget := pipeline.LatticeBudget
	if latticeBudget == (analyzer.Budget{}) {
		latticeBudget = analyzer.DefaultBudget()
	}
	scanInput := input
	if pipeline.AllowFinalPeriod && strings.HasSuffix(input, ".") && len(input) > 1 && !strings.ContainsAny(input[:len(input)-1], ".?!") {
		scanInput = input[:len(input)-1]
	}
	lattice, err := analyzer.BuildLattice(scanInput, pipeline.Language, pipeline.Tokenize, pipeline.Morphology, pipeline.Lexicon, pipeline.MWEPlans, latticeBudget)
	if err != nil {
		return SourceSyntax{}, err
	}
	attestConstructionTerminals(&lattice, pipeline.AllowedUnlexicalizedTerminals)
	if err := ctx.Err(); err != nil {
		return SourceSyntax{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
	}
	phraseCandidates, parseContext, err := pipeline.PhraseBuilder.Build(ctx, lattice)
	if err != nil {
		return SourceSyntax{}, err
	}
	if err := ctx.Err(); err != nil {
		return SourceSyntax{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
	}
	parseBudget := pipeline.ParseBudget
	if parseBudget == (parser.Budget{}) {
		parseBudget = parser.DefaultBudget()
	}
	forest, err := parser.Parse(lattice, phraseCandidates, parseContext, pipeline.ParserIndex, parseBudget)
	if err != nil {
		return SourceSyntax{}, err
	}
	if err := ctx.Err(); err != nil {
		return SourceSyntax{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
	}
	compositionEvidence := CompositionEvidence{}
	if pipeline.EvidenceBuilder != nil {
		compositionEvidence, err = pipeline.EvidenceBuilder.Build(ctx, lattice, phraseCandidates, forest)
		if err != nil {
			return SourceSyntax{}, err
		}
		if err := ctx.Err(); err != nil {
			return SourceSyntax{}, pipelineError(status.OutcomeResourceLimit, "SOURCE_ANALYSIS_CANCELED", "context", "source analysis was canceled or timed out")
		}
	}
	return SourceSyntax{
		Lattice: lattice, PhraseCandidates: phraseCandidates, ParseContext: parseContext,
		Forest: forest, CompositionEvidence: compositionEvidence, OriginalSourceBytes: len(input),
	}, nil
}
