package main

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

// compiledRuntimeOptions is the sole compile-time registration point for
// trusted language providers and stage factories in the CLI. Runtime package
// bindings are compiled from the exact selected snapshot and authored seed.
func compiledRuntimeOptions() runtime.Options {
	return runtime.Options{
		LanguageFactory: func(snapshot *pack.Pack) ([]language.Package, []language.LanguageProvider, error) {
			polishPackage, polishProvider, err := pl.NewSentenceProvider(snapshot)
			if err != nil {
				return nil, nil, fmt.Errorf("compile Polish sentence provider: %w", err)
			}
			englishPackage, englishProvider, err := en.NewSentenceProvider(snapshot)
			if err != nil {
				return nil, nil, fmt.Errorf("compile English sentence provider: %w", err)
			}
			return []language.Package{polishPackage, englishPackage}, []language.LanguageProvider{polishProvider, englishProvider}, nil
		},
		StageFactory: func(snapshot *pack.Pack, registry *language.Registry) (runtime.StageAdapters, error) {
			polishSource, err := pl.NewSentenceSource(snapshot)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile Polish source pipeline: %w", err)
			}
			englishSource, err := en.NewSentenceSource(snapshot)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile English source pipeline: %w", err)
			}
			polishTarget, err := pl.NewSentenceTarget(snapshot, registry)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile Polish target pipeline: %w", err)
			}
			englishTarget, err := en.NewSentenceTarget(snapshot, registry)
			if err != nil {
				return runtime.StageAdapters{}, fmt.Errorf("compile English target pipeline: %w", err)
			}
			return runtime.StageAdapters{
				Analyzers: map[features.LanguageID]engine.SourceAnalyzer{
					"pl": polishSource, "en": englishSource,
				},
				Realizers: map[features.LanguageID]engine.TargetRealizer{
					"pl": polishTarget, "en": englishTarget,
				},
			}, nil
		},
	}
}
