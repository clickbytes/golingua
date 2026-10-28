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

const sentenceProviderVersion = "v2026.10.23-g10-20-reading-past"

type sentenceProvider struct {
	source   *runtime.SourcePipeline
	realizer *runtime.ConstructionPlanRealizer
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
	senseIDs, err := sentenceSenseIDs()
	if err != nil {
		return language.Package{}, nil, err
	}
	languagePackage, err := runtime.BuildSentenceLanguagePackage("pl", sentenceProviderVersion, valencyIndex, snapshot.Lexicon(), senseIDs...)
	if err != nil {
		return language.Package{}, nil, err
	}
	return languagePackage, &sentenceProvider{source: source, realizer: realizer}, nil
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
	senseIDs, err := sentenceSenseIDs()
	if err != nil {
		return nil, err
	}
	return runtime.SentenceAnalysisCandidates(result, senseIDs...), nil
}

func sentenceSenseIDs() ([]ids.SenseID, error) {
	senseIDs := make([]ids.SenseID, 0, 26)
	for _, value := range []string{"pl.widziec.visual", "pl.slyszec.auditory", "pl.dostac.receive", "pl.patrzec.directed", "pl.czytac.interpret", "pl.byc.class_membership", "pl.czarny.black", "pl.jesc.eat", "pl.pic.drink", "pl.kupowac.buy", "pl.sprzedawac.sell", "pl.otwierac_obiekt.open", "pl.zamykac_obiekt.close", "pl.niesc.carry", "pl.trzymac.hold", "pl.ciagnac.pull", "pl.pchac.push", "pl.ciac.cut", "pl.skladac.fold", "pl.rozdzierac.tear", "pl.spac.sleep", "pl.oddychac.breathe", "pl.kichac.sneeze", "pl.biec.run", "pl.kaszlec.cough", "pl.chodzic.walk", "pl.zyc.live", "pl.latac.fly", "pl.zasypiac.sleep", "pl.budzic_sie.wake", "pl.pracowac.activity", "pl.isc.move", "pl.dawac.transfer"} {
		senseID, err := ids.NewSenseID(value)
		if err != nil {
			return nil, err
		}
		senseIDs = append(senseIDs, senseID)
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
