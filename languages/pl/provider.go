package pl

import (
	"context"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/valency"
)

const sentenceProviderVersion = "v2026.10.23-g10-20-reading-past"

type sentenceProvider struct {
	source   *runtime.SourcePipeline
	realizer *runtime.ConstructionPlanRealizer
	senseIDs []ids.SenseID
}

var _ language.LanguageProvider = (*sentenceProvider)(nil)
var _ language.Analyzer = (*sentenceProvider)(nil)
var _ language.Realizer = (*sentenceProvider)(nil)

// NewSentenceProvider creates the exact Polish seeing, hearing, receiving,
// looking, and present/past reading package and its trusted provider against
// one validated snapshot.
func NewSentenceProvider(snapshot *pack.Pack) (language.Package, language.LanguageProvider, error) {
	if snapshot == nil {
		return language.Package{}, nil, fmt.Errorf("Polish sentence provider requires a validated snapshot")
	}
	seed, err := Load()
	if err != nil {
		return language.Package{}, nil, err
	}
	return NewSentenceProviderWithSeed(snapshot, seed)
}

// NewSentenceProviderWithSeed builds analysis and realization over the same
// caller-supplied Polish morphology seed.
func NewSentenceProviderWithSeed(snapshot *pack.Pack, seed *Seed) (language.Package, language.LanguageProvider, error) {
	if snapshot == nil {
		return language.Package{}, nil, fmt.Errorf("Polish sentence provider requires a validated snapshot")
	}
	if seed == nil {
		return language.Package{}, nil, fmt.Errorf("Polish sentence provider requires a Polish morphology seed")
	}
	source, err := NewSentenceSourceWithSeed(snapshot, seed)
	if err != nil {
		return language.Package{}, nil, err
	}
	realizer, valencyIndex, err := newSentencePlanRealizerWithSeed(snapshot, seed)
	if err != nil {
		return language.Package{}, nil, err
	}
	senseIDs, err := sentenceSenseIDs(valencyIndex, snapshot.Lexicon())
	if err != nil {
		return language.Package{}, nil, err
	}
	languagePackage, err := runtime.BuildSentenceLanguagePackage("pl", sentenceProviderVersion, valencyIndex, snapshot.Lexicon(), senseIDs...)
	if err != nil {
		return language.Package{}, nil, err
	}
	return languagePackage, &sentenceProvider{source: source, realizer: realizer, senseIDs: senseIDs}, nil
}

func (provider *sentenceProvider) Language() features.LanguageID { return "pl" }
func (provider *sentenceProvider) Version() string               { return sentenceProviderVersion }
func (provider *sentenceProvider) Evidence() []language.CapabilityEvidence {
	return runtime.SentenceCapabilityEvidence("pl")
}

func (provider *sentenceProvider) Analyze(ctx context.Context, surface string) ([]language.AnalysisCandidate, error) {
	if provider == nil || provider.source == nil {
		return nil, fmt.Errorf("Polish sentence analyzer is unavailable")
	}
	result, err := provider.source.Analyze(ctx, surface)
	if err != nil {
		return nil, err
	}
	return runtime.SentenceAnalysisCandidates(result, provider.senseIDs...), nil
}

func sentenceSenseIDs(valencyIndex *valency.Index, lexicalIndex *lexicon.Index) ([]ids.SenseID, error) {
	if valencyIndex == nil || lexicalIndex == nil {
		return nil, fmt.Errorf("Polish sentence provider requires compiled valency and lexical indexes")
	}
	senseSet := make(map[ids.SenseID]struct{})
	for _, binding := range valencyIndex.Bindings() {
		if !binding.Approved {
			continue
		}
		for _, transfer := range lexicalIndex.TransferBindings(binding.SenseID) {
			if transfer.TransferAuthorized() {
				senseSet[binding.SenseID] = struct{}{}
				break
			}
		}
	}
	senseIDs := make([]ids.SenseID, 0, len(senseSet))
	for senseID := range senseSet {
		senseIDs = append(senseIDs, senseID)
	}
	sort.Slice(senseIDs, func(i, j int) bool { return senseIDs[i].String() < senseIDs[j].String() })
	if len(senseIDs) == 0 {
		return nil, fmt.Errorf("Polish sentence provider requires at least one approved lexical route")
	}
	return senseIDs, nil
}

func (provider *sentenceProvider) Realize(ctx context.Context, graph meaning.Meaning, targetID meaning.NodeID, plan construction.CompiledConstructionPlan) (language.RealizationResult, error) {
	if provider == nil || provider.realizer == nil {
		return language.RealizationResult{}, fmt.Errorf("Polish sentence realizer is unavailable")
	}
	return provider.realizer.Realize(ctx, graph, targetID, plan)
}

func newSentencePlanRealizerWithSeed(snapshot *pack.Pack, seed *Seed) (*runtime.ConstructionPlanRealizer, *valency.Index, error) {
	if snapshot == nil || seed == nil {
		return nil, nil, fmt.Errorf("Polish sentence realizer requires a snapshot and Polish morphology seed")
	}
	profile, _, _, _ := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "pl" {
		return nil, nil, fmt.Errorf("Polish sentence realizer requires a Polish morphology seed")
	}
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
	execution, err := NewSentenceExecutionBuilderWithSeed(seed)
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
