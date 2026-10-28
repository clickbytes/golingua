package main

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/bundled"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/learning"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

// learningRuntimeOptions explicitly opts into the local correction store. It
// pins overlay construction and both runtime factories to the bundled source
// identity; the default compiledRuntimeOptions remains unchanged.
func learningRuntimeOptions(storePath string) (runtime.Options, error) {
	_, _, options, err := prepareLearningRuntime(storePath)
	return options, err
}

func prepareLearningRuntime(storePath string) (*pl.Seed, *en.Seed, runtime.Options, error) {
	candidate, err := bundled.Candidate()
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("build bundled baseline candidate: %w", err)
	}
	baseline, err := pack.Compile(candidate)
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("compile bundled baseline snapshot: %w", err)
	}
	expectedIdentity := baseline.Identity()

	basePolish, err := pl.Load()
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("load base Polish seed: %w", err)
	}
	baseEnglish, err := en.Load()
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("load base English seed: %w", err)
	}
	rules, err := learning.ActiveRules(
		storePath, expectedIdentity.DatasetID, expectedIdentity.DatasetVersion, expectedIdentity.Digest,
		func(languageID features.LanguageID, lexemeID string, values []features.FeatureValue) (morph.GenerationResult, error) {
			switch languageID {
			case "pl":
				return basePolish.Generate(lexemeID, values)
			case "en":
				return baseEnglish.Generate(lexemeID, values)
			default:
				return morph.GenerationResult{}, fmt.Errorf("unsupported learning language %q", languageID)
			}
		},
	)
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("validate active morphology corrections: %w", err)
	}
	polishSeed, err := pl.LoadWithExceptions(rules[features.LanguageID("pl")])
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("compile Polish learning overlay: %w", err)
	}
	englishSeed, err := en.LoadWithExceptions(rules[features.LanguageID("en")])
	if err != nil {
		return nil, nil, runtime.Options{}, fmt.Errorf("compile English learning overlay: %w", err)
	}

	checkSnapshot := func(snapshot *pack.Pack) error {
		if snapshot == nil {
			return fmt.Errorf("learning runtime requires a validated snapshot")
		}
		if snapshot.Identity() != expectedIdentity {
			return fmt.Errorf("runtime snapshot identity does not match the learning overlay baseline")
		}
		return nil
	}

	options := runtime.Options{
		LanguageFactory: func(snapshot *pack.Pack) ([]language.Package, []language.LanguageProvider, error) {
			if err := checkSnapshot(snapshot); err != nil {
				return nil, nil, err
			}
			polishPackage, polishProvider, err := pl.NewSentenceProviderWithSeed(snapshot, polishSeed)
			if err != nil {
				return nil, nil, fmt.Errorf("compile Polish sentence provider with learning seed: %w", err)
			}
			englishPackage, englishProvider, err := en.NewSentenceProviderWithSeed(snapshot, englishSeed)
			if err != nil {
				return nil, nil, fmt.Errorf("compile English sentence provider with learning seed: %w", err)
			}
			return []language.Package{polishPackage, englishPackage}, []language.LanguageProvider{polishProvider, englishProvider}, nil
		},
		StageFactory: func(snapshot *pack.Pack, registry *language.Registry) (runtime.StageAdapters, error) {
			if err := checkSnapshot(snapshot); err != nil {
				return runtime.StageAdapters{}, err
			}
			if registry == nil {
				return runtime.StageAdapters{}, fmt.Errorf("learning runtime requires a verified language registry")
			}
			polishSource, err := pl.NewSentenceSourceWithSeed(snapshot, polishSeed)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile Polish source pipeline with learning seed: %w", err)
			}
			englishSource, err := en.NewSentenceSourceWithSeed(snapshot, englishSeed)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile English source pipeline with learning seed: %w", err)
			}
			polishTarget, err := pl.NewSentenceTargetWithSeed(snapshot, registry, polishSeed)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile Polish target pipeline with learning seed: %w", err)
			}
			englishTarget, err := en.NewSentenceTargetWithSeed(snapshot, registry, englishSeed)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile English target pipeline with learning seed: %w", err)
			}
			return runtime.StageAdapters{
				Analyzers: map[features.LanguageID]engine.SourceAnalyzer{"pl": polishSource, "en": englishSource},
				Realizers: map[features.LanguageID]engine.TargetRealizer{"pl": polishTarget, "en": englishTarget},
			}, nil
		},
	}
	return polishSeed, englishSeed, options, nil
}
