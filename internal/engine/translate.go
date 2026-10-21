package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// Request is one bounded language-pair translation request. PreviewID, when
// set, explicitly selects a retained candidate for display; it never changes
// the candidate Meaning or the request's ambiguity/incompleteness status.
type Request struct {
	From      features.LanguageID
	To        features.LanguageID
	Text      string
	PreviewID string
}

// Translate runs the same source-analysis and target-realization sequence for
// every language pair. Domain outcomes are returned in TranslationResult with
// a nil error so partial alternatives and diagnostics remain serializable.
func (engine *Engine) Translate(ctx context.Context, request Request) (TranslationResult, error) {
	if !validLanguage(string(request.From)) || !validLanguage(string(request.To)) || request.From == request.To {
		return invalidRequestResult(request, "INVALID_LANGUAGE_PAIR", "language", "translation requires two distinct canonical language identifiers")
	}
	if request.Text == "" || !utf8.ValidString(request.Text) || strings.TrimSpace(request.Text) == "" {
		return invalidRequestResult(request, "INVALID_TRANSLATION_INPUT", "text", "translation text must be non-empty valid UTF-8")
	}
	if len(request.Text) > engineBudget(engine).MaxInputBytes {
		return emptyResult(request, status.OutcomeResourceLimit, diagnostic("RESOURCE_LIMIT", "text", "translation input exceeds the request byte limit")), nil
	}
	if request.PreviewID != "" && !validID(request.PreviewID) {
		return invalidRequestResult(request, "INVALID_PREVIEW_ID", "preview_id", "preview identity must be non-empty bounded UTF-8")
	}
	if ctx == nil {
		return invalidRequestResult(request, "MISSING_REQUEST_CONTEXT", "context", "translation requires a request context")
	}
	if engine == nil {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("ENGINE_UNAVAILABLE", "engine", "translation runtime is not configured")), nil
	}
	budget := engineBudget(engine)
	requestContext, cancel := context.WithTimeout(ctx, budget.RequestTimeout)
	defer cancel()

	if engine.languages == nil {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("LANGUAGE_REGISTRY_UNAVAILABLE", "languages", "no verified language registry is configured")), nil
	}
	sourceLanguage, sourceRegistered := engine.languages.Language(request.From)
	if !sourceRegistered {
		return emptyResult(request, status.OutcomeUnknown, diagnostic("UNKNOWN_SOURCE_LANGUAGE", "from", "source language is not registered")), nil
	}
	targetLanguage, targetRegistered := engine.languages.Language(request.To)
	if !targetRegistered {
		return emptyResult(request, status.OutcomeUnknown, diagnostic("UNKNOWN_TARGET_LANGUAGE", "to", "target language is not registered")), nil
	}
	if !supportsCapability(sourceLanguage.Capabilities, true) || !supportsCapability(targetLanguage.Capabilities, false) {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("LANGUAGE_PAIR_UNSUPPORTED", "language.capabilities", "verified source analysis and target realization capabilities are required")), nil
	}
	sourceAnalyzer := engine.analyzers[request.From]
	if sourceAnalyzer == nil {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("SOURCE_RUNTIME_UNAVAILABLE", "from", "no source analysis and semantic-composition adapter is configured")), nil
	}
	targetRealizer := engine.realizers[request.To]
	if targetRealizer == nil {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("TARGET_RUNTIME_UNAVAILABLE", "to", "no target selection, planning, and realization adapter is configured")), nil
	}
	if engine.core == nil || engine.snapshot == nil {
		return emptyResult(request, status.OutcomeUnknown, diagnostic("RUNTIME_SNAPSHOT_UNAVAILABLE", "snapshot", "validated CORE and pack snapshots are required")), nil
	}
	if !reflect.DeepEqual(engine.snapshot.Concepts(), engine.core.Concepts()) {
		return emptyResult(request, status.OutcomeInvalid, diagnostic("RUNTIME_SNAPSHOT_MISMATCH", "snapshot", "CORE concept inventory does not match the configured pack snapshot")), nil
	}

	source, err := sourceAnalyzer.Analyze(requestContext, request.Text)
	if err != nil {
		return stageFailure(request, err, "SOURCE_ANALYSIS_FAILED"), nil
	}
	if len(source.Candidates) > budget.MaxSourceCandidates {
		return emptyResult(request, status.OutcomeResourceLimit, diagnostic("RESOURCE_LIMIT", "source.candidates", "source candidates exceed the engine retention limit")), nil
	}
	validated, validationErr := semantic.Disambiguate(
		source.Candidates, source.Outcome, source.Diagnostics, engine.core, semantic.DefaultEvidenceBudget(),
	)
	if validationErr != nil {
		return stageFailure(request, validationErr, "SOURCE_RESULT_INVALID"), nil
	}
	result := TranslationResult{
		Status: validated.Outcome, From: request.From, To: request.To,
		Candidates:  []TranslationAlternative{},
		Diagnostics: append([]status.Diagnostic(nil), validated.Diagnostics...),
	}
	if !validated.Outcome.Valid() {
		return emptyResult(request, status.OutcomeInvalid, diagnostic("INVALID_SOURCE_OUTCOME", "source.outcome", "source analyzer returned an unsupported status")), nil
	}
	if len(validated.Candidates) == 0 {
		if result.Status == status.OutcomeOK {
			result.Status = status.OutcomeUnknown
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("NO_SOURCE_MEANING", "source.candidates", "source analysis returned no semantic candidate"))
		}
		return finalizeResult(result)
	}

	surfaceCount := 0
	var candidateTextBytes int
	for _, sourceCandidate := range validated.Candidates {
		if len(result.Candidates) >= MaxTranslationAlternatives {
			result.Status = status.OutcomeResourceLimit
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("RESOURCE_LIMIT", "candidates", "translation alternative retention limit was exhausted"))
			break
		}
		if sourceCandidate.Status != status.OutcomeOK || sourceCandidate.Meaning == nil {
			result.Candidates = append(result.Candidates, unresolvedAlternative(sourceCandidate))
			continue
		}
		if !hasVerifiedBindingEvidence(sourceLanguage, sourceCandidate.Evidence, sourceCandidate.Meaning, true) {
			unresolved := unresolvedAlternative(sourceCandidate)
			unresolved.Outcome = status.OutcomeUnsupported
			unresolved.Diagnostics = appendDiagnostic(unresolved.Diagnostics, diagnostic("SOURCE_BINDING_EVIDENCE_UNLICENSED", "source.evidence", "source candidate has no exact registry binding for its sense, frame, family, variant, and CORE concept"))
			result.Candidates = append(result.Candidates, unresolved)
			result.Status = mergeEngineOutcome(result.Status, status.OutcomeUnsupported)
			continue
		}
		target, targetErr := targetRealizer.Realize(requestContext, sourceCandidate)
		if targetErr != nil {
			failed := stageFailure(request, targetErr, "TARGET_REALIZATION_FAILED")
			result.Diagnostics = appendDiagnostics(result.Diagnostics, failed.Diagnostics)
			result.Status = mergeEngineOutcome(result.Status, failed.Status)
			result.Candidates = append(result.Candidates, unresolvedTargetAlternative(sourceCandidate, failed.Status, failed.Diagnostics))
			continue
		}
		if !target.Outcome.Valid() || len(target.Diagnostics) > status.MaxDiagnostics || len(target.Alternatives) > budget.MaxSurfaceAlternatives {
			result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("INVALID_TARGET_RESULT", "target", "target adapter returned an invalid outcome or exceeded its declared bound"))
			continue
		}
		result.Diagnostics = appendDiagnostics(result.Diagnostics, target.Diagnostics)
		targetHasComplete := false
		for _, surface := range target.Alternatives {
			if surface.Outcome == status.OutcomeOK && surface.Text != "" && hasVerifiedBindingEvidence(targetLanguage, surface.Evidence, sourceCandidate.Meaning, false) {
				targetHasComplete = true
				break
			}
		}
		targetOutcome := target.Outcome
		if targetHasComplete && (targetOutcome == status.OutcomeUnsupported || targetOutcome == status.OutcomeUnknown) {
			// A complete licensed output coexisting with unresolved target
			// alternatives is partial coverage, never a confident success or a
			// reason to hide the explicit candidate preview.
			targetOutcome = status.OutcomeIncomplete
		}
		if len(target.Alternatives) == 0 {
			if target.Outcome == status.OutcomeOK {
				target.Outcome = status.OutcomeIncomplete
				target.Diagnostics = appendDiagnostic(target.Diagnostics, diagnostic("NO_TARGET_ALTERNATIVE", "target.alternatives", "target adapter returned OK without a surface alternative"))
			}
			result.Status = mergeEngineOutcome(result.Status, target.Outcome)
			result.Candidates = append(result.Candidates, unresolvedTargetAlternative(sourceCandidate, target.Outcome, target.Diagnostics))
			continue
		}
		result.Status = mergeEngineOutcome(result.Status, targetOutcome)
		for _, surface := range target.Alternatives {
			surfaceCount++
			if surfaceCount > budget.MaxSurfaceAlternatives || len(result.Candidates) >= MaxTranslationAlternatives {
				result.Status = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("RESOURCE_LIMIT", "target.alternatives", "translation alternative retention limit was exhausted"))
				break
			}
			if !surface.Outcome.Valid() || !validID(surface.ID) || len(surface.Diagnostics) > status.MaxDiagnostics || len(surface.Losses) > MaxTranslationLosses {
				result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
				result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("INVALID_SURFACE_ALTERNATIVE", "target.alternatives", "target adapter returned malformed alternative metadata"))
				continue
			}
			if surface.Text != "" {
				candidateTextBytes += len(surface.Text)
				if len(surface.Text) > MaxTranslationTextBytes || candidateTextBytes > budget.MaxCandidateTextBytes || !utf8.ValidString(surface.Text) {
					result.Status = status.OutcomeResourceLimit
					result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("RESOURCE_LIMIT", "target.text", "surface alternatives exceed the bounded result text limit"))
					break
				}
			}
			outcome := mergeEngineOutcome(sourceCandidate.Status, surface.Outcome)
			licensed := true
			if surface.Outcome == status.OutcomeOK && surface.Text != "" {
				licensed = hasVerifiedBindingEvidence(targetLanguage, surface.Evidence, sourceCandidate.Meaning, false)
				if !licensed {
					outcome = mergeEngineOutcome(outcome, status.OutcomeUnsupported)
				}
			}
			for _, loss := range surface.Losses {
				if loss.Outcome != status.OutcomeOK {
					outcome = mergeEngineOutcome(outcome, loss.Outcome)
				}
			}
			if outcome == status.OutcomeOK && surface.Text == "" {
				outcome = status.OutcomeIncomplete
			}
			candidate := TranslationAlternative{
				ID:                translationAlternativeID(sourceCandidate.ID, surface.ID),
				SourceCandidateID: sourceCandidate.ID, SemanticID: sourceCandidate.SemanticID,
				Outcome: outcome, Text: surface.Text, Meaning: sourceCandidate.Meaning,
				Evidence:    appendEvidence(sourceCandidate.Evidence, surface.Evidence),
				Losses:      append([]realizer.FieldLoss(nil), surface.Losses...),
				Diagnostics: append([]status.Diagnostic(nil), surface.Diagnostics...),
			}
			if !licensed {
				candidate.Diagnostics = appendDiagnostic(candidate.Diagnostics, diagnostic("TARGET_BINDING_EVIDENCE_UNLICENSED", "target.evidence", "target output has no exact registry binding for its sense, frame, family, variant, and CORE concept"))
			}
			if candidate.Outcome != status.OutcomeOK && len(candidate.Diagnostics) == 0 {
				candidate.Diagnostics = append(candidate.Diagnostics, diagnostic("TARGET_ALTERNATIVE_NOT_OK", "target.alternatives", "target alternative is not a complete supported realization"))
			}
			result.Candidates = append(result.Candidates, candidate)
		}
	}

	complete := 0
	unresolved := false
	for _, candidate := range result.Candidates {
		if candidate.Outcome == status.OutcomeOK && candidate.Text != "" {
			complete++
		} else {
			unresolved = true
		}
	}
	if result.Status == status.OutcomeOK {
		switch {
		case complete > 1:
			result.Status = status.OutcomeAmbiguous
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("MULTIPLE_TRANSLATION_ALTERNATIVES", "candidates", "multiple complete translations remain; no candidate was selected"))
		case unresolved:
			result.Status = status.OutcomeIncomplete
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("UNRESOLVED_TRANSLATION_ALTERNATIVE", "candidates", "at least one source or target alternative remains unresolved"))
		case complete == 0:
			result.Status = status.OutcomeUnknown
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("NO_COMPLETE_TRANSLATION", "candidates", "no complete licensed translation was produced"))
		}
	}
	if result.Status == status.OutcomeOK && complete == 1 {
		for _, candidate := range result.Candidates {
			if candidate.Outcome == status.OutcomeOK && candidate.Text != "" {
				result.Text = candidate.Text
				break
			}
		}
	}
	if request.PreviewID != "" {
		var selected *TranslationAlternative
		for i := range result.Candidates {
			if result.Candidates[i].ID == request.PreviewID {
				selected = &result.Candidates[i]
				break
			}
		}
		if selected == nil || selected.Outcome != status.OutcomeOK || selected.Text == "" ||
			(result.Status != status.OutcomeAmbiguous && result.Status != status.OutcomeIncomplete) {
			return invalidResultFrom(result, diagnostic("INVALID_PREVIEW_ID", "preview_id", "preview ID must name a complete retained candidate in an ambiguous or incomplete result"))
		}
		result.PreviewID = selected.ID
		result.Text = selected.Text
	}
	return finalizeResult(result)
}

func finalizeResult(result TranslationResult) (TranslationResult, error) {
	result.Diagnostics = trimDiagnostics(result.Diagnostics)
	if result.Status != status.OutcomeOK && len(result.Diagnostics) == 0 {
		result.Diagnostics = []status.Diagnostic{diagnostic("TRANSLATION_NOT_OK", "translation", "translation did not establish one complete supported output")}
	}
	if err := result.Validate(); err != nil {
		return TranslationResult{}, engineError(status.OutcomeInvalid, "INVALID_TRANSLATION_RESULT", "result", err.Error())
	}
	return result.Clone(), nil
}

func emptyResult(request Request, outcome status.OutcomeKind, issue status.Diagnostic) TranslationResult {
	return TranslationResult{Status: outcome, From: request.From, To: request.To, Candidates: []TranslationAlternative{}, Diagnostics: []status.Diagnostic{issue}}
}

func invalidRequestResult(request Request, code, path, message string) (TranslationResult, error) {
	result := TranslationResult{
		Status: status.OutcomeInvalid, From: safeLanguageForResult(request.From, features.LanguageID("und")), To: safeLanguageForResult(request.To, features.LanguageID("und")),
		Candidates: []TranslationAlternative{}, Diagnostics: []status.Diagnostic{diagnostic(code, path, message)},
	}
	return finalizeResult(result)
}

func invalidResultFrom(previous TranslationResult, issue status.Diagnostic) (TranslationResult, error) {
	previous.Status = status.OutcomeInvalid
	previous.Text = ""
	previous.PreviewID = ""
	previous.Diagnostics = appendDiagnostic(previous.Diagnostics, issue)
	return finalizeResult(previous)
}

func safeLanguageForResult(language, fallback features.LanguageID) features.LanguageID {
	value := string(language)
	if validLanguage(value) {
		return language
	}
	return fallback
}

func stageFailure(request Request, err error, fallbackCode string) TranslationResult {
	if errors.Is(err, context.DeadlineExceeded) {
		return emptyResult(request, status.OutcomeResourceLimit, diagnostic("REQUEST_TIMEOUT", "context", "translation request exceeded its time budget"))
	}
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		result := typed.Result()
		if result.Validate() == nil {
			outcome := result.Status
			if outcome == status.OutcomeOK {
				outcome = status.OutcomeError
			}
			diagnostics := appendDiagnostics(nil, result.Diagnostics)
			if len(diagnostics) == 0 {
				diagnostics = appendDiagnostic(diagnostics, diagnostic(fallbackCode, "runtime", "a translation stage returned a typed refusal or failure"))
			}
			return TranslationResult{Status: outcome, From: request.From, To: request.To, Candidates: []TranslationAlternative{}, Diagnostics: diagnostics}
		}
	}
	return emptyResult(request, status.OutcomeError, diagnostic(fallbackCode, "runtime", "a translation stage failed unexpectedly"))
}

func unresolvedAlternative(candidate semantic.MeaningCandidate) TranslationAlternative {
	diagnostics := append([]status.Diagnostic(nil), candidate.Evidence.Diagnostics...)
	if len(diagnostics) == 0 {
		diagnostics = []status.Diagnostic{diagnostic("SOURCE_MEANING_UNRESOLVED", "source.candidates", "source candidate does not contain one complete supported Meaning")}
	}
	var evidence []semantic.EvidenceTrace
	if candidate.Evidence.AlternativeID != "" {
		evidence = []semantic.EvidenceTrace{candidate.Evidence.Clone()}
	}
	return TranslationAlternative{
		ID: candidate.ID, SourceCandidateID: candidate.ID, SemanticID: candidate.SemanticID,
		Outcome: candidate.Status, Meaning: candidate.Meaning,
		Evidence: evidence, Diagnostics: diagnostics,
	}
}

func unresolvedTargetAlternative(candidate semantic.MeaningCandidate, outcome status.OutcomeKind, diagnostics []status.Diagnostic) TranslationAlternative {
	if len(diagnostics) == 0 {
		diagnostics = []status.Diagnostic{diagnostic("TARGET_REALIZATION_UNAVAILABLE", "target", "target runtime returned no complete realization")}
	}
	var evidence []semantic.EvidenceTrace
	if candidate.Evidence.AlternativeID != "" {
		evidence = []semantic.EvidenceTrace{candidate.Evidence.Clone()}
	}
	return TranslationAlternative{
		ID: translationAlternativeID(candidate.ID, "unrealized"), SourceCandidateID: candidate.ID,
		SemanticID: candidate.SemanticID, Outcome: outcome, Meaning: candidate.Meaning,
		Evidence: evidence, Diagnostics: append([]status.Diagnostic(nil), diagnostics...),
	}
}

func appendEvidence(source semantic.EvidenceTrace, target semantic.EvidenceTrace) []semantic.EvidenceTrace {
	result := make([]semantic.EvidenceTrace, 0, 2)
	if source.AlternativeID != "" {
		result = append(result, source.Clone())
	}
	if target.AlternativeID != "" {
		result = append(result, target.Clone())
	}
	return result
}

func supportsCapability(capabilities []language.VerifiedCapability, analyze bool) bool {
	for _, capability := range capabilities {
		if analyze && capability.Analyze || !analyze && capability.Realize {
			return true
		}
	}
	return false
}

func hasVerifiedBindingEvidence(pack language.Language, trace semantic.EvidenceTrace, graph *meaning.Meaning, analyze bool) bool {
	if trace.Outcome != status.OutcomeOK || graph == nil {
		return false
	}
	concepts := make(map[meaning.ConceptRef]struct{})
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindEvent || node.Kind == meaning.KindState || node.Kind == meaning.KindReferent {
			if node.Concept != "" {
				concepts[node.Concept] = struct{}{}
			}
		}
	}
	for _, record := range trace.Records {
		if record.SenseID.String() == "" || record.FrameID == "" || record.FamilyID.String() == "" || record.VariantID == "" || record.CoreConceptID.String() == "" {
			continue
		}
		if record.Outcome != status.OutcomeOK {
			continue
		}
		if _, referenced := concepts[meaning.ConceptRefFromID(record.CoreConceptID)]; !referenced {
			continue
		}
		for _, binding := range pack.Bindings {
			if binding.SenseID != record.SenseID || binding.ValencyFrameID != string(record.FrameID) ||
				binding.FamilyID != record.FamilyID || binding.CoreRef != record.CoreConceptID ||
				binding.Mapping != "equivalent" || !binding.Approved ||
				analyze && !binding.Analyze || !analyze && !binding.Realize ||
				!containsVariantID(binding.VariantIDs, string(record.VariantID)) {
				continue
			}
			return true
		}
	}
	return false
}

func containsVariantID(variants []ids.ConstructionID, value string) bool {
	for _, variant := range variants {
		if variant.String() == value {
			return true
		}
	}
	return false
}

func translationAlternativeID(sourceID, targetID string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%d:", len(sourceID)))
	builder.WriteString(sourceID)
	builder.WriteByte('|')
	builder.WriteString(fmt.Sprintf("%d:", len(targetID)))
	builder.WriteString(targetID)
	digest := sha256.Sum256([]byte(builder.String()))
	return "translation:" + hex.EncodeToString(digest[:])
}

func appendDiagnostic(diagnostics []status.Diagnostic, item status.Diagnostic) []status.Diagnostic {
	if len(diagnostics) >= status.MaxDiagnostics {
		return diagnostics
	}
	return append(diagnostics, item)
}

func appendDiagnostics(target, source []status.Diagnostic) []status.Diagnostic {
	for _, item := range source {
		target = appendDiagnostic(target, item)
	}
	return target
}

func trimDiagnostics(diagnostics []status.Diagnostic) []status.Diagnostic {
	if len(diagnostics) > status.MaxDiagnostics {
		diagnostics = diagnostics[:status.MaxDiagnostics]
	}
	return diagnostics
}

func mergeEngineOutcome(left, right status.OutcomeKind) status.OutcomeKind {
	if left == "" || left == status.OutcomeOK {
		return right
	}
	if right == "" || right == status.OutcomeOK || left == right {
		return left
	}
	if left == status.OutcomeError || right == status.OutcomeError {
		return status.OutcomeError
	}
	if left == status.OutcomeInvalid || right == status.OutcomeInvalid {
		return status.OutcomeInvalid
	}
	if left == status.OutcomeResourceLimit || right == status.OutcomeResourceLimit {
		return status.OutcomeResourceLimit
	}
	if left == status.OutcomeAmbiguous || right == status.OutcomeAmbiguous {
		return status.OutcomeAmbiguous
	}
	if left == status.OutcomeIncomplete || right == status.OutcomeIncomplete {
		return status.OutcomeIncomplete
	}
	if left == status.OutcomeUnsupported || right == status.OutcomeUnsupported {
		return status.OutcomeUnsupported
	}
	return status.OutcomeUnknown
}

func engineBudget(engine *Engine) Budget {
	if engine == nil || engine.budget == (Budget{}) {
		return DefaultBudget()
	}
	return engine.budget
}

func diagnostic(code, path, message string) status.Diagnostic {
	return status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
}
