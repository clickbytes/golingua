package runtime

import (
	"context"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/semantic"
)

// DSEOptions contains caller-selected semantic snapshots and explicit request
// identity. The runtime wrapper adds only evidence produced by AnalyzeSyntax.
type DSEOptions struct {
	Package    *language.SemanticLanguagePackage
	Plan       *language.SourcePlan
	Selected   algebra.Document
	Rewrite    algebra.RewriteEnvironment
	References language.SourceReferenceContext
	Budget     semantic.DSEBudget
}

// DSEAnalysis preserves the parser output alongside its generic typed
// composition. No legacy CORE composition or candidate selection is run.
type DSEAnalysis struct {
	Syntax      SourceSyntax       `json:"syntax"`
	Composition semantic.DSEResult `json:"composition"`
}

// AnalyzeDSE runs the actual source pipeline exactly once, then composes the
// resulting forest and attached evidence under the explicitly selected DSE
// metadata and request context.
func (pipeline *SourcePipeline) AnalyzeDSE(ctx context.Context, source string, options DSEOptions) (DSEAnalysis, error) {
	syntax, err := pipeline.AnalyzeSyntax(ctx, source)
	if err != nil {
		return DSEAnalysis{}, err
	}
	composition, err := semantic.ComposeDSE(semantic.DSEInput{
		Context:                  ctx,
		Forest:                   syntax.Forest,
		Candidates:               syntax.PhraseCandidates,
		Lattice:                  syntax.Lattice,
		SourceBytes:              syntax.OriginalSourceBytes,
		References:               syntax.CompositionEvidence.References,
		NodeAttributes:           syntax.CompositionEvidence.NodeAttributes,
		Operators:                syntax.CompositionEvidence.Operators,
		OperatorCoverageComplete: syntax.CompositionEvidence.OperatorCoverageComplete,
		Package:                  options.Package,
		Plan:                     options.Plan,
		Selected:                 options.Selected,
		Rewrite:                  options.Rewrite,
		ReferenceContext:         options.References,
		Language:                 pipeline.Language,
		Budget:                   options.Budget,
	})
	if err != nil {
		return DSEAnalysis{}, err
	}
	return DSEAnalysis{Syntax: syntax, Composition: composition}, nil
}
