// Package matrix contains the shared, production-backed runtime used by the
// language behavior matrix.
package matrix

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/bundled"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

// NewRuntime compiles the bundled data and the same trusted PL/EN production
// adapters used by the CLI. It does not analyze or realize any case.
func NewRuntime() (*engine.Engine, error) {
	candidate, err := bundled.Candidate()
	if err != nil {
		return nil, fmt.Errorf("build bundled candidate: %w", err)
	}
	snapshot, err := pack.Compile(candidate)
	if err != nil {
		return nil, fmt.Errorf("compile bundled pack: %w", err)
	}
	polishPackage, polishProvider, err := pl.NewSentenceProvider(snapshot)
	if err != nil {
		return nil, fmt.Errorf("compile Polish sentence provider: %w", err)
	}
	englishPackage, englishProvider, err := en.NewSentenceProvider(snapshot)
	if err != nil {
		return nil, fmt.Errorf("compile English sentence provider: %w", err)
	}
	registry, err := language.BuildRegistry(
		[]language.Package{polishPackage, englishPackage},
		[]language.LanguageProvider{polishProvider, englishProvider},
	)
	if err != nil {
		return nil, fmt.Errorf("build verified language registry: %w", err)
	}
	polishSource, err := pl.NewSentenceSource(snapshot)
	if err != nil {
		return nil, fmt.Errorf("compile Polish source pipeline: %w", err)
	}
	englishSource, err := en.NewSentenceSource(snapshot)
	if err != nil {
		return nil, fmt.Errorf("compile English source pipeline: %w", err)
	}
	polishTarget, err := pl.NewSentenceTarget(snapshot, registry)
	if err != nil {
		return nil, fmt.Errorf("compile Polish target pipeline: %w", err)
	}
	englishTarget, err := en.NewSentenceTarget(snapshot, registry)
	if err != nil {
		return nil, fmt.Errorf("compile English target pipeline: %w", err)
	}
	return engine.New(engine.Config{
		Core: snapshot.CoreGraph(), LanguageRegistry: registry, PackSnapshot: snapshot,
		Analyzers: map[features.LanguageID]engine.SourceAnalyzer{
			"pl": polishSource, "en": englishSource,
		},
		Realizers: map[features.LanguageID]engine.TargetRealizer{
			"pl": polishTarget, "en": englishTarget,
		},
	})
}
