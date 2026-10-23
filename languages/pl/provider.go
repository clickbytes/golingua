package pl

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/valency"
)

const sentenceProviderVersion = "v2026.10.23"

type sentenceProvider struct {
	source   *runtime.SourcePipeline
	realizer *runtime.ConstructionPlanRealizer
}

var _ language.LanguageProvider = (*sentenceProvider)(nil)
var _ language.Analyzer = (*sentenceProvider)(nil)
var _ language.Realizer = (*sentenceProvider)(nil)

// NewSentenceProvider creates the exact Polish seeing capability package and
// its trusted provider against one validated snapshot.
func NewSentenceProvider(snapshot *pack.Pack) (language.Package, language.LanguageProvider, error) {
	if snapshot == nil {
		return language.Package{}, nil, fmt.Errorf("Polish sentence provider requires a validated snapshot")
	}
	source, err := NewSentenceSource(snapshot)
	if err != nil {
		return language.Package{}, nil, err
	}
	realizer, valencyIndex, err := newSentencePlanRealizer(snapshot)
	if err != nil {
		return language.Package{}, nil, err
	}
	senseID, err := ids.NewSenseID("pl.widziec.visual")
	if err != nil {
		return language.Package{}, nil, err
	}
	languagePackage, err := runtime.BuildSentenceLanguagePackage("pl", sentenceProviderVersion, senseID, valencyIndex, snapshot.Lexicon())
	if err != nil {
		return language.Package{}, nil, err
	}
	return languagePackage, &sentenceProvider{source: source, realizer: realizer}, nil
}

func (provider *sentenceProvider) Language() features.LanguageID { return "pl" }
func (provider *sentenceProvider) Version() string               { return sentenceProviderVersion }
func (provider *sentenceProvider) Evidence() []language.CapabilityEvidence {
	return runtime.SentenceCapabilityEvidence()
}

func (provider *sentenceProvider) Analyze(ctx context.Context, surface string) ([]language.AnalysisCandidate, error) {
	if provider == nil || provider.source == nil {
		return nil, fmt.Errorf("Polish sentence analyzer is unavailable")
	}
	result, err := provider.source.Analyze(ctx, surface)
	if err != nil {
		return nil, err
	}
	senseID, err := ids.NewSenseID("pl.widziec.visual")
	if err != nil {
		return nil, err
	}
	return runtime.SentenceAnalysisCandidates(result, senseID), nil
}

func (provider *sentenceProvider) Realize(ctx context.Context, graph meaning.Meaning, targetID meaning.NodeID, plan construction.CompiledConstructionPlan) (language.RealizationResult, error) {
	if provider == nil || provider.realizer == nil {
		return language.RealizationResult{}, fmt.Errorf("Polish sentence realizer is unavailable")
	}
	return provider.realizer.Realize(ctx, graph, targetID, plan)
}

func newSentencePlanRealizer(snapshot *pack.Pack) (*runtime.ConstructionPlanRealizer, *valency.Index, error) {
	seed, err := Load()
	if err != nil {
		return nil, nil, err
	}
	profile, _, _, _ := seed.RuntimeMorphology()
	valencyIndex, err := CompileValency()
	if err != nil {
		return nil, nil, err
	}
	packages, constructions, err := SentenceConstructions(valencyIndex)
	if err != nil {
		return nil, nil, err
	}
	_, variants, err := runtime.CompileSentencePlans(profile, packages, valencyIndex)
	if err != nil {
		return nil, nil, err
	}
	execution, err := NewSentenceExecutionBuilder()
	if err != nil {
		return nil, nil, err
	}
	adapter := &runtime.ConstructionPlanRealizer{
		Language: "pl", Lexicon: snapshot.Lexicon(), Valency: valencyIndex,
		Constructions: constructions, Variants: variants, Core: snapshot.CoreGraph(),
		ExecutionBuilder: execution,
	}
	return adapter, valencyIndex, nil
}
