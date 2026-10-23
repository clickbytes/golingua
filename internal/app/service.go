// Package app exposes the typed in-process operations shared by command-line
// and interactive clients. It owns application orchestration, while callers
// own argument parsing, presentation, and process behavior.
package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/clickbytes/golingua/internal/bundled"
	"github.com/clickbytes/golingua/internal/catalog"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/staging"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

// EngineLoader supplies an engine for an explicitly selected runtime config.
type EngineLoader func(string) (*engine.Engine, error)

// Morphology provides the bundled language-local analyze and generate API.
type Morphology interface {
	Generate(string, []features.FeatureValue) (morph.GenerationResult, error)
	Analyze(string) ([]morph.Reading, error)
}

// SnapshotStore is the mutation boundary used by snapshot preview/apply.
type SnapshotStore interface {
	Preview(staging.Candidate) (staging.Preview, error)
	Apply(staging.Candidate) (staging.Preview, error)
}

type MorphologyLoader func(features.LanguageID) (Morphology, error)
type CatalogLoader func() (CatalogData, error)
type CandidateLoader func() (staging.Candidate, error)
type SnapshotStoreOpener func(string) (SnapshotStore, error)

// CatalogData is an immutable compiled view of the authored CORE and PL/EN
// catalog. The service returns defensive copies of mutable inventory data.
type CatalogData struct {
	Graph                *core.ConceptGraph
	Inventory            catalog.Inventory
	CoreRelationTypes    int
	ValencyFrames        ValencyFrameCounts
	ConstructionBindings ConstructionBindingCounts
}

type ValencyFrameCounts struct {
	PL int `json:"pl"`
	EN int `json:"en"`
}

type ConstructionBindingCounts struct {
	PLPendingCandidates int `json:"pl_pending_candidates"`
	ENPendingCandidates int `json:"en_pending_candidates"`
	PLApproved          int `json:"pl_approved"`
	ENApproved          int `json:"en_approved"`
}

// Config allows the executable to provide trusted runtime and data adapters.
// Missing adapters use the compiled Go catalogs and local snapshot store.
type Config struct {
	LoadEngine     EngineLoader
	LoadMorphology MorphologyLoader
	LoadCatalog    CatalogLoader
	LoadCandidate  CandidateLoader
	OpenStore      SnapshotStoreOpener
}

// Service is the shared application facade. Catalog compilation is lazy and
// cached because the authored graph and language catalogs are immutable.
type Service struct {
	loadEngine     EngineLoader
	loadMorphology MorphologyLoader
	loadCatalog    CatalogLoader
	loadCandidate  CandidateLoader
	openStore      SnapshotStoreOpener
	catalogOnce    sync.Once
	catalogData    CatalogData
	catalogErr     error
}

func NewService(config Config) *Service {
	if config.LoadMorphology == nil {
		config.LoadMorphology = loadMorphology
	}
	if config.LoadCatalog == nil {
		config.LoadCatalog = loadCatalog
	}
	if config.LoadCandidate == nil {
		config.LoadCandidate = bundled.Candidate
	}
	if config.OpenStore == nil {
		config.OpenStore = func(path string) (SnapshotStore, error) { return staging.OpenStore(path) }
	}
	return &Service{
		loadEngine: config.LoadEngine, loadMorphology: config.LoadMorphology,
		loadCatalog: config.LoadCatalog, loadCandidate: config.LoadCandidate,
		openStore: config.OpenStore,
	}
}

// Translate runs the configured engine and returns the stable typed result.
// It never writes output or chooses a process exit code.
type TranslationResponse struct {
	Result  *engine.TranslationResult
	Failure *status.Result
}

func (service *Service) Translate(ctx context.Context, request TranslateRequest) TranslationResponse {
	if ctx == nil {
		ctx = context.Background()
	}
	var translator *engine.Engine
	var err error
	if request.RuntimeConfig != "" {
		if service == nil || service.loadEngine == nil {
			return translationFailure(failure(status.OutcomeUnsupported, "RUNTIME_LOADER_UNAVAILABLE", "runtime_config", "this executable cannot load configured language runtimes"))
		}
		translator, err = service.loadEngine(request.RuntimeConfig)
	} else {
		translator, err = engine.New(engine.Config{})
	}
	if err != nil {
		return translationFailure(errorResult(err, "ENGINE_SETUP_ERROR", "translate", "translation engine setup failed unexpectedly"))
	}
	if translator == nil {
		return translationFailure(failure(status.OutcomeError, "ENGINE_SETUP_ERROR", "translate", "translation engine setup failed unexpectedly"))
	}
	result, err := translator.Translate(ctx, engine.Request{
		From: features.LanguageID(request.From), To: features.LanguageID(request.To),
		Text: request.Text, PreviewID: request.PreviewID,
	})
	if err != nil {
		return translationFailure(errorResult(err, "TRANSLATION_RUNTIME_ERROR", "translate", "translation failed unexpectedly"))
	}
	return TranslationResponse{Result: &result}
}

func translationFailure(result status.Result) TranslationResponse {
	return TranslationResponse{Failure: &result}
}

type TranslateRequest struct {
	From          string
	To            string
	Text          string
	PreviewID     string
	RuntimeConfig string
}

// FormsRequest selects one explicit analysis or generation operation.
type FormsRequest struct {
	Language string
	LexemeID string
	Surface  string
	Features []features.FeatureValue
}

// FormsResult is the stable, typed morphology response shared with CLI/TUI.
type FormsResult struct {
	Status      status.OutcomeKind    `json:"status"`
	Language    string                `json:"language,omitempty"`
	LexemeID    string                `json:"lexeme_id,omitempty"`
	Surface     string                `json:"surface,omitempty"`
	State       morph.GenerationState `json:"state,omitempty"`
	Forms       []morph.GeneratedForm `json:"forms,omitempty"`
	Readings    []morph.Reading       `json:"readings,omitempty"`
	Diagnostics []status.Diagnostic   `json:"diagnostics,omitempty"`
}

func (service *Service) Forms(ctx context.Context, request FormsRequest) FormsResult {
	if ctx == nil {
		ctx = context.Background()
	}
	response := FormsResult{Status: status.OutcomeOK, Language: request.Language, LexemeID: request.LexemeID, Surface: request.Surface}
	if request.Language != "pl" && request.Language != "en" || (request.LexemeID == "") == (request.Surface == "") {
		return formsFailure(response, status.OutcomeInvalid, "INVALID_FORMS_REQUEST", "forms request requires language pl|en and exactly one of lexeme ID or surface")
	}
	if request.LexemeID != "" && len(request.Features) == 0 {
		return formsFailure(response, status.OutcomeInvalid, "INVALID_FORMS_REQUEST", "generation requires at least one feature value")
	}
	if request.Surface != "" && len(request.Features) > 0 {
		return formsFailure(response, status.OutcomeInvalid, "INVALID_FORMS_REQUEST", "analysis does not accept generation features")
	}
	if err := ctx.Err(); err != nil {
		return formsFailure(response, status.OutcomeError, "MORPHOLOGY_CANCELED", "morphology operation was canceled")
	}
	if service == nil || service.loadMorphology == nil {
		return formsFailure(response, status.OutcomeError, "MORPHOLOGY_LOAD_FAILED", "could not load bundled morphology")
	}
	seed, err := service.loadMorphology(features.LanguageID(request.Language))
	if err != nil || seed == nil {
		return formsFailure(response, status.OutcomeError, "MORPHOLOGY_LOAD_FAILED", "could not load bundled morphology")
	}
	if request.LexemeID != "" {
		result, generateErr := seed.Generate(request.LexemeID, append([]features.FeatureValue(nil), request.Features...))
		if generateErr != nil {
			return formsError(response, generateErr)
		}
		response.State, response.Forms, response.Diagnostics = result.State, result.Forms, result.Diagnostics
		switch result.State {
		case morph.GenerationGenerated:
			response.Status = status.OutcomeOK
		case morph.GenerationIncomplete:
			response.Status = status.OutcomeIncomplete
		case morph.GenerationBlocked, morph.GenerationUnavailable:
			response.Status = status.OutcomeUnsupported
		default:
			response.Status = status.OutcomeUnknown
		}
	} else {
		response.Readings, err = seed.Analyze(request.Surface)
		if err != nil {
			return formsError(response, err)
		}
		if len(response.Readings) == 0 {
			response.Status = status.OutcomeUnknown
		}
	}
	if len(response.Forms) > morph.MaxGeneratedForms || len(response.Readings) > 4096 || len(response.Diagnostics) > status.MaxDiagnostics {
		return FormsResult{Status: status.OutcomeResourceLimit, Diagnostics: []status.Diagnostic{diagnostic("RESOURCE_LIMIT", "morphology output exceeds limit", "forms")}}
	}
	if response.Status != status.OutcomeOK && len(response.Diagnostics) == 0 {
		response.Diagnostics = []status.Diagnostic{diagnostic("FORM_UNAVAILABLE", "requested morphology is not available", "forms")}
	}
	return response
}

func loadMorphology(language features.LanguageID) (Morphology, error) {
	switch language {
	case "pl":
		return pl.Load()
	case "en":
		return en.Load()
	default:
		return nil, fmt.Errorf("unsupported morphology language %q", language)
	}
}

type CatalogRequest struct {
	Details bool
}

// CatalogResult is the shared inventory response used by catalog screens and
// the scriptable catalog command.
type CatalogResult struct {
	Status               status.OutcomeKind         `json:"status"`
	Inventory            *catalog.Inventory         `json:"inventory,omitempty"`
	CoreRelationTypes    int                        `json:"core_relation_types"`
	ValencyFrames        *ValencyFrameCounts        `json:"valency_frames,omitempty"`
	ConstructionBindings *ConstructionBindingCounts `json:"construction_bindings,omitempty"`
	Diagnostics          []status.Diagnostic        `json:"diagnostics,omitempty"`
}

func (service *Service) Catalog(ctx context.Context, request CatalogRequest) CatalogResult {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return CatalogResult{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic("CATALOG_CANCELED", "catalog operation was canceled", "catalog")}}
	}
	data, err := service.catalog()
	if err != nil {
		return CatalogResult{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic("CATALOG_LOAD_FAILED", "could not compile the bundled Go catalog", "catalog")}}
	}
	inventory := cloneInventory(data.Inventory)
	if !request.Details {
		inventory.PerConcept = []catalog.ConceptCoverage{}
	}
	frames, bindings := data.ValencyFrames, data.ConstructionBindings
	return CatalogResult{
		Status: status.OutcomeOK, Inventory: &inventory, CoreRelationTypes: data.CoreRelationTypes,
		ValencyFrames: &frames, ConstructionBindings: &bindings,
	}
}

// LookupConcept returns a defensive ConceptGraph value for a stable concept ID.
func (service *Service) LookupConcept(ctx context.Context, id ids.ConceptID) (core.Concept, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return core.Concept{}, false, err
	}
	data, err := service.catalog()
	if err != nil {
		return core.Concept{}, false, err
	}
	concept, found := data.Graph.Lookup(id)
	return concept, found, nil
}

func (service *Service) catalog() (CatalogData, error) {
	if service == nil {
		return CatalogData{}, errors.New("application service is unavailable")
	}
	service.catalogOnce.Do(func() {
		service.catalogData, service.catalogErr = service.loadCatalog()
	})
	return service.catalogData, service.catalogErr
}

func loadCatalog() (CatalogData, error) {
	graph, err := catalog.Compile()
	if err != nil {
		return CatalogData{}, err
	}
	polish, err := pl.CompileLexicon(graph)
	if err != nil {
		return CatalogData{}, err
	}
	english, err := en.CompileLexicon(graph)
	if err != nil {
		return CatalogData{}, err
	}
	inventory, err := catalog.Inspect(graph, polish, english)
	if err != nil {
		return CatalogData{}, err
	}
	polishValency, err := pl.CompileValency()
	if err != nil {
		return CatalogData{}, err
	}
	englishValency, err := en.CompileValency()
	if err != nil {
		return CatalogData{}, err
	}
	return CatalogData{
		Graph: graph, Inventory: inventory, CoreRelationTypes: len(graph.RelationTypes()),
		ValencyFrames: ValencyFrameCounts{PL: len(polishValency.Frames()), EN: len(englishValency.Frames())},
		ConstructionBindings: ConstructionBindingCounts{
			PLPendingCandidates: len(polishValency.Candidates()), ENPendingCandidates: len(englishValency.Candidates()),
			PLApproved: len(polishValency.Bindings()), ENApproved: len(englishValency.Bindings()),
		},
	}, nil
}

func cloneInventory(source catalog.Inventory) catalog.Inventory {
	result := source
	result.PerConcept = append([]catalog.ConceptCoverage(nil), source.PerConcept...)
	return result
}

type SnapshotRequest struct {
	StorePath      string
	Apply          bool
	Confirmed      bool
	ExpectedDigest string
}

type SnapshotResult struct {
	Status      status.OutcomeKind  `json:"status"`
	Preview     *staging.Preview    `json:"preview,omitempty"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

func (service *Service) Snapshot(ctx context.Context, request SnapshotRequest) SnapshotResult {
	if ctx == nil {
		ctx = context.Background()
	}
	if request.Apply && !request.Confirmed {
		return SnapshotResult{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{diagnostic("SNAPSHOT_APPLY_CONFIRMATION_REQUIRED", "snapshot apply requires explicit confirmation", "snapshot.confirmed")}}
	}
	if request.StorePath == "" {
		return SnapshotResult{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{diagnostic("INVALID_SNAPSHOT_REQUEST", "snapshot store path is required", "snapshot.store_path")}}
	}
	if err := ctx.Err(); err != nil {
		return SnapshotResult{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic("SNAPSHOT_CANCELED", "snapshot operation was canceled", "snapshot")}}
	}
	if service == nil {
		return SnapshotResult{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic("BUNDLED_SNAPSHOT_INVALID", "could not compile the bundled Go catalog", "snapshot")}}
	}
	candidate, err := service.loadCandidate()
	if err != nil {
		return snapshotError("BUNDLED_SNAPSHOT_INVALID", err)
	}
	store, err := service.openStore(request.StorePath)
	if err != nil {
		return snapshotError("SNAPSHOT_STORE_UNAVAILABLE", err)
	}
	var preview staging.Preview
	if request.Apply {
		if request.ExpectedDigest != "" {
			preview, err = store.Preview(candidate)
			if err != nil {
				return snapshotError("SNAPSHOT_OPERATION_FAILED", err)
			}
			if preview.Digest != request.ExpectedDigest {
				return SnapshotResult{
					Status:  status.OutcomeInvalid,
					Preview: &preview,
					Diagnostics: []status.Diagnostic{diagnostic(
						"SNAPSHOT_DIGEST_MISMATCH",
						"snapshot changed since preview; obtain a new preview and confirmation before applying",
						"snapshot.expected_digest",
					)},
				}
			}
		}
		if err := ctx.Err(); err != nil {
			return SnapshotResult{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic("SNAPSHOT_CANCELED", "snapshot operation was canceled", "snapshot")}}
		}
		preview, err = store.Apply(candidate)
	} else {
		preview, err = store.Preview(candidate)
	}
	if err != nil {
		return snapshotError("SNAPSHOT_OPERATION_FAILED", err)
	}
	return SnapshotResult{Status: status.OutcomeOK, Preview: &preview}
}

func formsFailure(response FormsResult, kind status.OutcomeKind, code, message string) FormsResult {
	response.Status = kind
	response.Diagnostics = []status.Diagnostic{diagnostic(code, message, "forms")}
	return response
}

func formsError(response FormsResult, err error) FormsResult {
	var source interface{ Result() status.Result }
	if errors.As(err, &source) {
		result := source.Result()
		response.Status, response.Diagnostics = result.Status, result.Diagnostics
		return response
	}
	return formsFailure(response, status.OutcomeError, "MORPHOLOGY_ERROR", "morphology operation failed")
}

func snapshotError(code string, err error) SnapshotResult {
	var source interface{ Result() status.Result }
	if errors.As(err, &source) {
		result := source.Result()
		return SnapshotResult{Status: result.Status, Diagnostics: result.Diagnostics}
	}
	return SnapshotResult{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic(code, err.Error(), "snapshot")}}
}

func errorResult(err error, code, path, message string) status.Result {
	var source interface{ Result() status.Result }
	if errors.As(err, &source) {
		return source.Result()
	}
	return status.Result{Status: status.OutcomeError, Diagnostics: []status.Diagnostic{diagnostic(code, message, path)}}
}

func failure(kind status.OutcomeKind, code, path, message string) status.Result {
	return status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic(code, message, path)}}
}

func diagnostic(code, message, path string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityError}
}
