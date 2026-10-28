package runtime

import (
	"context"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
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
	Coordination        *SourceCoordination
	OriginalSourceBytes int
}

// SourceCoordination keeps the two independently parsed source clauses and
// their evidence in original-input token coordinates.
type SourceCoordination struct {
	Connector string               `json:"connector"`
	Span      text.Span            `json:"span"`
	Clauses   []SourceClauseSyntax `json:"clauses"`
}

type SourceClauseSyntax struct {
	TokenStart int                 `json:"token_start"`
	TokenEnd   int                 `json:"token_end"`
	Forest     parser.Forest       `json:"forest"`
	Evidence   CompositionEvidence `json:"composition_evidence"`
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
	if pipeline.ClauseConnector != "" {
		attestConstructionTerminals(&lattice, map[string]bool{pipeline.ClauseConnector: true})
	}
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
	var coordination *SourceCoordination
	if pipeline.ClauseConnector != "" {
		coordination, err = pipeline.analyzeTwoClauseCoordination(ctx, lattice, phraseCandidates, parseContext, parseBudget)
		if err != nil {
			return SourceSyntax{}, err
		}
	}
	return SourceSyntax{
		Lattice: lattice, PhraseCandidates: phraseCandidates, ParseContext: parseContext,
		Forest: forest, CompositionEvidence: compositionEvidence, Coordination: coordination, OriginalSourceBytes: len(input),
	}, nil
}

func (pipeline *SourcePipeline) analyzeTwoClauseCoordination(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, parseContext parser.ParseContext, budget parser.Budget) (*SourceCoordination, error) {
	connectorIndexes := make([]int, 0, 2)
	for index, token := range lattice.Tokens {
		if token.Token.Folded == pipeline.ClauseConnector {
			connectorIndexes = append(connectorIndexes, index)
		}
	}
	if len(connectorIndexes) != 1 {
		return nil, nil
	}
	connectorIndex := connectorIndexes[0]
	leftEnd := connectorIndex
	if connectorIndex > 0 && lattice.Tokens[connectorIndex-1].Token.Kind == text.TokenPunctuation && lattice.Tokens[connectorIndex-1].Token.Normalized == "," {
		leftEnd--
	}
	rightStart := connectorIndex + 1
	if leftEnd == 0 || rightStart >= len(lattice.Tokens) {
		return nil, nil
	}
	left, err := parser.ParseClauseRange(lattice, candidates, parseContext, pipeline.ParserIndex, budget, 0, leftEnd, true)
	if err != nil {
		return nil, err
	}
	right, err := parser.ParseClauseRange(lattice, candidates, parseContext, pipeline.ParserIndex, budget, rightStart, len(lattice.Tokens), false)
	if err != nil {
		return nil, err
	}
	if !oneCompleteDeclarativeRoot(left) || !oneCompleteDeclarativeRoot(right) {
		return nil, nil
	}
	leftEvidence, err := pipeline.clauseEvidence(ctx, lattice, candidates, parseContext, left)
	if err != nil {
		return nil, err
	}
	rightEvidence, err := pipeline.clauseEvidence(ctx, lattice, candidates, parseContext, right)
	if err != nil {
		return nil, err
	}
	if !leftEvidence.OperatorCoverageComplete || !rightEvidence.OperatorCoverageComplete || len(leftEvidence.Operators) != 0 || len(rightEvidence.Operators) != 0 {
		return nil, nil
	}
	return &SourceCoordination{Connector: pipeline.ClauseConnector, Span: lattice.Tokens[connectorIndex].Token.Span, Clauses: []SourceClauseSyntax{
		{TokenStart: 0, TokenEnd: leftEnd, Forest: left, Evidence: leftEvidence},
		{TokenStart: rightStart, TokenEnd: len(lattice.Tokens), Forest: right, Evidence: rightEvidence},
	}}, nil
}

func (pipeline *SourcePipeline) clauseEvidence(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, parseContext parser.ParseContext, forest parser.Forest) (CompositionEvidence, error) {
	if pipeline.EvidenceBuilder == nil || len(forest.Roots) != 1 {
		return CompositionEvidence{}, nil
	}
	rootStart, rootEnd := -1, -1
	for _, node := range forest.Nodes {
		if node.ID == forest.Roots[0] {
			rootStart, rootEnd = node.Start, node.End
			break
		}
	}
	if rootStart < 0 || rootEnd <= rootStart {
		return CompositionEvidence{}, nil
	}
	clauseCandidates := make([]parser.PhraseCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.TokenStart >= rootStart && candidate.TokenEnd <= rootEnd {
			clauseCandidates = append(clauseCandidates, candidate)
		}
	}
	return pipeline.EvidenceBuilder.Build(ctx, lattice, clauseCandidates, forest)
}

func oneCompleteDeclarativeRoot(forest parser.Forest) bool {
	if forest.Outcome != status.OutcomeOK || len(forest.Roots) != 1 || len(forest.Unmatched) != 0 {
		return false
	}
	rootID := forest.Roots[0]
	for _, node := range forest.Nodes {
		if node.ID != rootID || string(node.Category) != "clause" || len(node.Alternatives) != 1 {
			continue
		}
		alternative := node.Alternatives[0]
		if len(alternative.Unresolved) != 0 || alternative.VariantID == "" {
			return false
		}
		for _, slot := range alternative.Slots {
			if slot.WHGap != nil {
				return false
			}
		}
		return true
	}
	return false
}
