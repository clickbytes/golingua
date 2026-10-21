// Package engine coordinates bounded, language-independent translation stages.
// Language-specific parsing and realization remain injected runtime adapters.
package engine

import (
	"context"
	"time"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	defaultMaxInputBytes          = 64 * 1024
	maxInputBytes                 = 1 << 20
	defaultMaxSourceCandidates    = 256
	maxSourceCandidates           = 4096
	defaultMaxSurfaceAlternatives = MaxTranslationAlternatives
	maxSurfaceAlternatives        = MaxTranslationAlternatives
	defaultMaxCandidateTextBytes  = 64 * 1024
	maxCandidateTextBytes         = 1 << 20
	defaultRequestTimeout         = 5 * time.Second
	maxRequestTimeout             = 30 * time.Second
)

// Budget bounds the engine request independently of parser, semantic and
// language-runtime budgets. Each stage remains responsible for its own limits.
type Budget struct {
	MaxInputBytes          int
	MaxSourceCandidates    int
	MaxSurfaceAlternatives int
	MaxCandidateTextBytes  int
	RequestTimeout         time.Duration
}

func DefaultBudget() Budget {
	return Budget{
		MaxInputBytes: defaultMaxInputBytes, MaxSourceCandidates: defaultMaxSourceCandidates,
		MaxSurfaceAlternatives: defaultMaxSurfaceAlternatives,
		MaxCandidateTextBytes:  defaultMaxCandidateTextBytes,
		RequestTimeout:         defaultRequestTimeout,
	}
}

func (budget Budget) Validate() error {
	if budget.MaxInputBytes <= 0 || budget.MaxInputBytes > maxInputBytes ||
		budget.MaxSourceCandidates <= 0 || budget.MaxSourceCandidates > maxSourceCandidates ||
		budget.MaxSurfaceAlternatives <= 0 || budget.MaxSurfaceAlternatives > maxSurfaceAlternatives ||
		budget.MaxCandidateTextBytes <= 0 || budget.MaxCandidateTextBytes > maxCandidateTextBytes ||
		budget.RequestTimeout <= 0 || budget.RequestTimeout > maxRequestTimeout {
		return engineError(status.OutcomeInvalid, "INVALID_ENGINE_BUDGET", "budget", "engine limits must be positive and within package ceilings")
	}
	return nil
}

// SourceAnalyzer runs the source-language analysis and semantic composition
// stages. It must retain all supported and unresolved alternatives; it may not
// select by score, insertion order, or a language-pair special case.
type SourceAnalyzer interface {
	Analyze(context.Context, string) (semantic.Result, error)
}

// SurfaceAlternative is one target-side lexicalization/plan/realization
// result for exactly one source Meaning candidate.
type SurfaceAlternative struct {
	ID          string
	Outcome     status.OutcomeKind
	Text        string
	Evidence    semantic.EvidenceTrace
	Losses      []realizer.FieldLoss
	Diagnostics []status.Diagnostic
}

// SurfaceResult retains target-side alternatives and their stage outcome.
type SurfaceResult struct {
	Outcome      status.OutcomeKind
	Alternatives []SurfaceAlternative
	Diagnostics  []status.Diagnostic
}

// TargetRealizer performs target candidate selection, planning and surface
// generation for one exact Meaning candidate. It must expose every viable
// output and never choose an unrequested candidate.
type TargetRealizer interface {
	Realize(context.Context, semantic.MeaningCandidate) (SurfaceResult, error)
}

// Config contains immutable shared snapshots and explicit per-language stage
// adapters. Maps are copied by New; callers must not mutate objects behind the
// adapter interfaces after construction.
type Config struct {
	Core             *core.ConceptGraph
	LanguageRegistry *language.Registry
	PackSnapshot     *pack.Pack
	Analyzers        map[features.LanguageID]SourceAnalyzer
	Realizers        map[features.LanguageID]TargetRealizer
	Budget           Budget
}

// Engine is an immutable request coordinator over one CORE and pack snapshot.
// Language-specific behavior is supplied by verified stage adapters.
type Engine struct {
	core      *core.ConceptGraph
	languages *language.Registry
	snapshot  *pack.Pack
	analyzers map[features.LanguageID]SourceAnalyzer
	realizers map[features.LanguageID]TargetRealizer
	budget    Budget
}

// New validates finite limits and snapshots the configured adapter maps.
// Missing runtime data is allowed at setup so callers can receive a typed,
// request-level unsupported/unknown result from Translate.
func New(config Config) (*Engine, error) {
	budget := config.Budget
	if budget == (Budget{}) {
		budget = DefaultBudget()
	}
	if err := budget.Validate(); err != nil {
		return nil, err
	}
	engine := &Engine{
		core: config.Core, languages: config.LanguageRegistry, snapshot: config.PackSnapshot,
		analyzers: make(map[features.LanguageID]SourceAnalyzer, len(config.Analyzers)),
		realizers: make(map[features.LanguageID]TargetRealizer, len(config.Realizers)), budget: budget,
	}
	for languageID, adapter := range config.Analyzers {
		if adapter == nil || !validLanguage(string(languageID)) {
			return nil, engineError(status.OutcomeInvalid, "INVALID_ANALYZER_ADAPTER", "analyzers", "source analyzers require a canonical language and non-nil adapter")
		}
		engine.analyzers[languageID] = adapter
	}
	for languageID, adapter := range config.Realizers {
		if adapter == nil || !validLanguage(string(languageID)) {
			return nil, engineError(status.OutcomeInvalid, "INVALID_REALIZER_ADAPTER", "realizers", "target realizers require a canonical language and non-nil adapter")
		}
		engine.realizers[languageID] = adapter
	}
	return engine, nil
}

// CoreGraph returns the immutable CORE graph configured for this engine.
func (engine *Engine) CoreGraph() *core.ConceptGraph {
	if engine == nil {
		return nil
	}
	return engine.core
}

// LanguageRegistry returns the immutable verified language registry.
func (engine *Engine) LanguageRegistry() *language.Registry {
	if engine == nil {
		return nil
	}
	return engine.languages
}

// PackSnapshot returns the immutable data snapshot pinned to this engine.
func (engine *Engine) PackSnapshot() *pack.Pack {
	if engine == nil {
		return nil
	}
	return engine.snapshot
}

type engineErrorValue struct{ result status.Result }

func (failure *engineErrorValue) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "translation engine error"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *engineErrorValue) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func engineError(kind status.OutcomeKind, code, path, message string) error {
	return &engineErrorValue{result: status.Result{
		Status: kind,
		Diagnostics: []status.Diagnostic{{
			Code: code, Path: path, Message: message, Severity: status.SeverityError,
		}},
	}}
}
