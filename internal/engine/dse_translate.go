package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func (engine *Engine) translateDSE(ctx context.Context, request Request) (TranslationResult, error) {
	if ctx == nil {
		return invalidRequestResult(request, "MISSING_REQUEST_CONTEXT", "context", "translation requires a request context")
	}
	if engine == nil || engine.dseProfile == nil {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("DSE_PROFILE_UNAVAILABLE", "dse_profile", "the explicitly selected DSE runtime profile is unavailable")), nil
	}
	if engine.core == nil || engine.snapshot == nil || !engine.snapshotMatchesCore {
		return emptyResult(request, status.OutcomeUnsupported, diagnostic("RUNTIME_SNAPSHOT_UNAVAILABLE", "snapshot", "DSE translation requires one validated CORE and pack snapshot")), nil
	}
	profile := engine.dseProfile
	result := TranslationResult{Status: status.OutcomeOK, From: request.From, To: request.To, Candidates: []TranslationAlternative{}, Diagnostics: []status.Diagnostic{}}
	trace := &DSETranslationTrace{
		ProfileID: profile.ID, ProfileVersion: profile.Version,
		RegistryFingerprint: profile.RegistryFingerprint,
		Registries:          append([]DSERegistryPin(nil), profile.Registries...),
		Composition:         semantic.DSEResult{Status: status.OutcomeUnknown, Candidates: []semantic.DSECandidate{}, Residuals: []semantic.DSEResidual{}, Diagnostics: []status.Diagnostic{}},
		Realizations:        []DSETranslationRealization{},
	}
	if engine.snapshot != nil {
		trace.PackIdentity = engine.snapshot.Identity()
	}
	result.DSETrace = trace

	if err := ctx.Err(); err != nil {
		return emptyResult(request, status.OutcomeResourceLimit, diagnostic("REQUEST_TIMEOUT", "context", "translation request was canceled or timed out")), nil
	}
	sourceResources, sourceOK := profile.Languages[request.From]
	targetResources, targetOK := profile.Languages[request.To]
	if !sourceOK || sourceResources.Source == nil {
		return dseFailure(result, status.OutcomeUnsupported, "DSE_SOURCE_PROFILE_UNAVAILABLE", "from", "no explicitly pinned DSE source profile is configured for this language"), nil
	}
	if !targetOK || targetResources.TargetPackage == nil {
		return dseFailure(result, status.OutcomeUnsupported, "DSE_TARGET_PROFILE_UNAVAILABLE", "to", "no explicitly pinned DSE target package is configured for this language"), nil
	}
	budget := engineBudget(engine)
	requestContext, cancel := context.WithTimeout(ctx, budget.RequestTimeout)
	defer cancel()
	composition, err := sourceResources.Source.AnalyzeDSE(requestContext, request.Text)
	if err != nil {
		return stageFailure(request, err, "DSE_SOURCE_ANALYSIS_FAILED"), nil
	}
	trace.Composition = composition
	if !composition.Status.Valid() || len(composition.Candidates) > budget.MaxSourceCandidates {
		return dseFailure(result, status.OutcomeInvalid, "INVALID_DSE_SOURCE_RESULT", "composition", "DSE source returned an invalid status or exceeded the candidate limit"), nil
	}
	result.Status = composition.Status
	result.Diagnostics = append([]status.Diagnostic(nil), composition.Diagnostics...)
	if len(composition.Candidates) == 0 && result.Status == status.OutcomeOK {
		result.Status = status.OutcomeUnknown
		result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("NO_DSE_CANDIDATE", "composition.candidates", "DSE composition returned no candidate"))
	}

	for _, candidate := range composition.Candidates {
		if len(result.Candidates) >= MaxTranslationAlternatives {
			result.Status = status.OutcomeResourceLimit
			result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("RESOURCE_LIMIT", "composition.candidates", "DSE result exceeds the translation alternative limit"))
			break
		}
		realizationReceipt := DSETranslationRealization{
			CandidateID: candidate.ID, Status: candidate.Status,
			Coverage: []AlgebraRealizationCoverage{}, Residuals: []AlgebraRealizationResidual{}, SourceResiduals: append([]semantic.DSEResidual(nil), candidate.Residuals...), Diagnostics: []status.Diagnostic{}, Mappings: []DSEReferenceConceptMapping{}, ConsumedConditions: []DSEConditionConsumption{},
			Source:        DSEResourceReceipt{ID: sourceResources.SourcePackageID, Version: sourceResources.SourcePackageVersion, Fingerprint: sourceResources.SourcePackageFingerprint},
			SourcePlan:    DSEResourceReceipt{ID: sourceResources.SourcePlanID, Version: sourceResources.SourcePlanVersion, Fingerprint: sourceResources.SourcePlanFingerprint},
			SourceContext: DSEResourceReceipt{ID: sourceResources.SourceContextID, Version: sourceResources.SourceContextVersion, Fingerprint: sourceResources.SourceContextFingerprint},
			Target:        DSEResourceReceipt{ID: targetResources.TargetPackage.ID(), Version: targetResources.TargetPackage.Version(), Fingerprint: targetResources.TargetPackageFingerprint},
		}
		refusalStatus, refusalCode, refusalMessage := dseSourceRefusal(candidate)
		if refusalCode != "" {
			refusal := diagnostic(refusalCode, "composition.candidates."+candidate.ID, refusalMessage)
			realizationReceipt.Status = refusalStatus
			realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, refusalStatus)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, refusalStatus, []status.Diagnostic{refusal}))
			continue
		}
		if len(candidate.CanonicalDocument) == 0 || candidate.Checked == nil {
			refusal := diagnostic("DSE_CANONICAL_DOCUMENT_REQUIRED", "composition.candidates."+candidate.ID, "candidate has no checked canonical algebra document")
			realizationReceipt.Status = status.OutcomeIncomplete
			realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, status.OutcomeIncomplete)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeIncomplete, []status.Diagnostic{refusal}))
			continue
		}
		document, decodeErr := algebra.Decode(candidate.CanonicalDocument)
		if decodeErr != nil {
			refusal := diagnostic("DSE_CANONICAL_DOCUMENT_INVALID", "composition.candidates."+candidate.ID, "candidate canonical document failed independent decoding and checking")
			realizationReceipt.Status = status.OutcomeInvalid
			realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeInvalid, []status.Diagnostic{refusal}))
			continue
		}
		checked, checkErr := algebra.Check(document)
		if checkErr != nil || !reflect.DeepEqual(checked, *candidate.Checked) {
			refusal := diagnostic("DSE_CANONICAL_CHECK_MISMATCH", "composition.candidates."+candidate.ID, "independent checker did not reproduce the source candidate's checked expression")
			realizationReceipt.Status = status.OutcomeInvalid
			realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeInvalid, []status.Diagnostic{refusal}))
			continue
		}
		if !dseTranslationExpressionSupported(document.Expression) {
			refusal := diagnostic("DSE_EXPRESSION_PROFILE_UNSUPPORTED", "composition.candidates."+candidate.ID, "this configured translation profile realizes one direct positive Atom or one Entity WH Ask over a direct Atom")
			realizationReceipt.Status = status.OutcomeUnsupported
			realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, status.OutcomeUnsupported)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeUnsupported, []status.Diagnostic{refusal}))
			continue
		}
		enriched, mappings, mapStatus, mapCode, mapMessage := enrichDSEContext(engine, document, candidate)
		if mapStatus != status.OutcomeOK {
			refusal := diagnostic(mapCode, "composition.candidates."+candidate.ID, mapMessage)
			realizationReceipt.Status = mapStatus
			realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, mapStatus)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, mapStatus, []status.Diagnostic{refusal}))
			continue
		}
		realizationReceipt.Mappings = mappings
		targetDocument, consumedConditions := consumeDSELexicalConditions(sourceResources.SourcePackage, candidate, enriched, mappings)
		realizationReceipt.ConsumedConditions = consumedConditions
		if targetDocument.RequestContext != nil {
			targetContext, cloneErr := algebra.CloneRequestContext(*targetDocument.RequestContext)
			if cloneErr != nil {
				refusal := diagnostic("DSE_TARGET_CONTEXT_INVALID", "composition.candidates."+candidate.ID+".request_context", "canonical request context could not be prepared for the target adapter")
				realizationReceipt.Status = status.OutcomeInvalid
				realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
				trace.Realizations = append(trace.Realizations, realizationReceipt)
				result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
				result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeInvalid, []status.Diagnostic{refusal}))
				continue
			}
			// These are source-analysis provenance. Typed references and their
			// exact sense-to-CORE mappings remain in the canonical trace; the
			// target Meaning only needs the already-grounded referent identity.
			targetContext.Mentions = []algebra.ContextMention{}
			targetContext.ResolutionEvidence = []algebra.ResolutionEvidence{}
			targetDocument.RequestContext = &targetContext
		}
		targetDocument.Sorts = nil
		targetDocument.Signatures = nil
		// The target checker revalidates this projection against the separately
		// pinned metadata. Inline declarations from the source candidate are not
		// allowed to become a second target-side authority.
		var realized AlgebraRealizationResult
		var realizeErr error
		if targetDocument.Expression != nil && targetDocument.Expression.Kind == algebra.TermAsk {
			// A bounded WH Ask can still produce a verified surface candidate when
			// unrelated or unsupported request-context details remain. The ordinary
			// realizer retains those details as residuals and returns an incomplete
			// result; it does not let them become a complete translation.
			realized, realizeErr = engine.realizeAlgebra(requestContext, request.To, targetDocument, targetResources.TargetPackage, profile.SelectedMetadata, false)
		} else {
			realized, realizeErr = engine.RealizeDSEAlgebra(requestContext, request.To, targetDocument, targetResources.TargetPackage, profile.SelectedMetadata)
		}
		if realizeErr != nil {
			failed := stageFailure(request, realizeErr, "DSE_TARGET_REALIZATION_FAILED")
			realizationReceipt.Status = failed.Status
			realizationReceipt.Diagnostics = append([]status.Diagnostic(nil), failed.Diagnostics...)
			trace.Realizations = append(trace.Realizations, realizationReceipt)
			result.Status = mergeEngineOutcome(result.Status, failed.Status)
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, failed.Status, failed.Diagnostics))
			continue
		}
		realizationReceipt.Status = realized.Status
		realizationReceipt.Coverage = append([]AlgebraRealizationCoverage(nil), realized.Coverage...)
		realizationReceipt.Residuals = append([]AlgebraRealizationResidual(nil), realized.Residuals...)
		realizationReceipt.Diagnostics = append([]status.Diagnostic(nil), realized.Diagnostics...)
		trace.Realizations = append(trace.Realizations, realizationReceipt)
		result.Status = mergeDSEOutcomes(result.Status, realized.Status, candidate.Status)
		if len(realized.Alternatives) == 0 {
			diagnostics := append([]status.Diagnostic(nil), realized.Diagnostics...)
			if len(diagnostics) == 0 {
				diagnostics = []status.Diagnostic{diagnostic("DSE_TARGET_NO_ALTERNATIVE", "target.alternatives", "target adapter returned no surface alternative")}
			}
			result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, mergeDSEOutcomes(realized.Status, candidate.Status), diagnostics))
			continue
		}
		for _, surface := range realized.Alternatives {
			if len(result.Candidates) >= MaxTranslationAlternatives {
				result.Status = status.OutcomeResourceLimit
				result.Diagnostics = appendDiagnostic(result.Diagnostics, diagnostic("RESOURCE_LIMIT", "target.alternatives", "DSE target alternatives exceed the translation alternative limit"))
				break
			}
			// Keep the aggregate target status on the request. A candidate keeps
			// only its own surface outcome so an explicit preview can select it
			// without erasing ambiguity among sibling alternatives.
			outcome := mergeDSEOutcomes(candidate.Status, surface.Outcome)
			if len(realized.Residuals) > 0 {
				outcome = mergeDSEOutcomes(outcome, status.OutcomeIncomplete)
			}
			diagnostics := append([]status.Diagnostic(nil), surface.Diagnostics...)
			if outcome != status.OutcomeOK && len(diagnostics) == 0 {
				diagnostics = append([]status.Diagnostic(nil), realized.Diagnostics...)
			}
			if outcome != status.OutcomeOK && len(diagnostics) == 0 {
				diagnostics = []status.Diagnostic{diagnostic("DSE_TRANSLATION_INCOMPLETE", "target", "DSE candidate did not produce a complete supported realization")}
			}
			result.Candidates = append(result.Candidates, TranslationAlternative{
				ID: translationAlternativeID(candidate.ID, surface.ID), SourceCandidateID: candidate.ID, SemanticID: candidate.ID,
				Outcome: outcome, Text: surface.Text, Evidence: []semantic.EvidenceTrace{surface.Evidence.Clone()},
				Assumptions: cloneFieldAssumptions(surface.Assumptions), Losses: append([]realizer.FieldLoss(nil), surface.Losses...), Diagnostics: diagnostics,
			})
		}
	}
	if result.Status != status.OutcomeOK && len(result.Diagnostics) == 0 {
		result.Diagnostics = []status.Diagnostic{diagnostic("DSE_TRANSLATION_NOT_OK", "composition", "DSE did not establish one complete supported translation")}
	}
	if request.PreviewID != "" {
		var selected *TranslationAlternative
		for i := range result.Candidates {
			if result.Candidates[i].ID == request.PreviewID {
				selected = &result.Candidates[i]
				break
			}
		}
		if selected == nil || !previewableAlternative(*selected) || (result.Status != status.OutcomeAmbiguous && result.Status != status.OutcomeIncomplete) {
			return invalidResultFrom(result, diagnostic("INVALID_PREVIEW_ID", "preview_id", "preview ID must name a text-bearing DSE candidate in an ambiguous or incomplete result"))
		}
		result.PreviewID, result.Text = selected.ID, selected.Text
	}
	encoded, encodeErr := json.Marshal(result)
	if encodeErr != nil || len(encoded) > MaxTranslationPayloadBytes {
		return dseOutputLimitResult(request, profile, trace.PackIdentity), nil
	}
	return finalizeResult(result)
}

func consumeDSELexicalConditions(sourcePackage *language.SemanticLanguagePackage, candidate semantic.DSECandidate, document algebra.Document, mappings []DSEReferenceConceptMapping) (algebra.Document, []DSEConditionConsumption) {
	consumed := []DSEConditionConsumption{}
	if document.RequestContext == nil {
		return document, consumed
	}
	context, err := algebra.CloneRequestContext(*document.RequestContext)
	if err != nil {
		return document, consumed
	}
	remaining := make([]algebra.ContextCondition, 0, len(context.Conditions))
	for _, condition := range context.Conditions {
		receipt, ok := dseLexicalConditionReceipt(sourcePackage, candidate, condition, mappings)
		if !ok {
			remaining = append(remaining, condition)
			continue
		}
		consumed = append(consumed, receipt)
	}
	context.Conditions = remaining
	document.RequestContext = &context
	return document, consumed
}

func dseLexicalConditionReceipt(sourcePackage *language.SemanticLanguagePackage, candidate semantic.DSECandidate, condition algebra.ContextCondition, mappings []DSEReferenceConceptMapping) (DSEConditionConsumption, bool) {
	if sourcePackage == nil || sourcePackage.LexiconIndex() == nil || condition.Kind != algebra.ContextPresupposition || condition.Code != "lexical_descriptor" || condition.Expression == nil || condition.Expression.Kind != algebra.TermAtom || len(condition.Expression.Args) != 1 {
		return DSEConditionConsumption{}, false
	}
	argument := condition.Expression.Args[0]
	if argument.Value == nil || argument.Value.Kind != algebra.TermRef || argument.Value.Type == nil || argument.Value.Type.Kind != algebra.TypeEntity || len(condition.TargetReferentIDs) != 1 || condition.TargetReferentIDs[0] != argument.Value.ID {
		return DSEConditionConsumption{}, false
	}
	var mapping *DSEReferenceConceptMapping
	for index := range mappings {
		if mappings[index].ReferenceID != argument.Value.ID {
			continue
		}
		if mapping != nil || mappings[index].MappingKind != string(lexicon.MappingEquivalent) || mappings[index].ApprovalStatus != string(lexicon.BindingApproved) {
			return DSEConditionConsumption{}, false
		}
		mapping = &mappings[index]
	}
	if mapping == nil {
		return DSEConditionConsumption{}, false
	}
	roleSources := 0
	for _, source := range candidate.RoleSources {
		if source.ReferenceID == mapping.ReferenceID && source.Source.SenseID == mapping.SourceSenseID {
			roleSources++
		}
	}
	if roleSources != 1 {
		return DSEConditionConsumption{}, false
	}
	coverage := 0
	for _, row := range candidate.ConditionCoverage {
		if row.ConditionID == condition.ID && row.Code == condition.Code && string(row.Kind) == string(condition.Kind) && len(row.TargetReferenceIDs) == 1 && row.TargetReferenceIDs[0] == mapping.ReferenceID {
			coverage++
		}
	}
	if coverage != 1 {
		return DSEConditionConsumption{}, false
	}
	sense, err := ids.ParseSenseID(mapping.SourceSenseID)
	if err != nil {
		return DSEConditionConsumption{}, false
	}
	bindings := sourcePackage.LexiconIndex().TargetBindingsForSense(sense)
	var selected *lexicon.SemanticBinding
	for index := range bindings {
		binding := &bindings[index]
		target := binding.Target
		if target == nil || target.Kind != lexicon.SemanticTargetReferencePredicate || target.SymbolRef != lexicon.SymbolRef(condition.Expression.Symbol) || target.ReferencePort == nil || target.ReferencePort.PortID != argument.Role || !binding.TargetExecutionAuthorized() {
			continue
		}
		if selected != nil {
			return DSEConditionConsumption{}, false
		}
		selected = binding
	}
	if selected == nil || !hasDSEReferenceFragment(candidate.FragmentChecks, selected.Target) {
		return DSEConditionConsumption{}, false
	}
	return DSEConditionConsumption{
		ConditionID: condition.ID, Kind: string(condition.Kind), Code: condition.Code,
		Symbol: condition.Expression.Symbol, ReferenceID: mapping.ReferenceID,
		SourceSenseID: mapping.SourceSenseID, CoreConceptID: mapping.CoreConceptID,
		SourceBindingProvenance: selected.Provenance.ID,
	}, true
}

func hasDSEReferenceFragment(checks []semantic.DSEFragmentCheck, expected *lexicon.SemanticTarget) bool {
	if expected == nil {
		return false
	}
	matches := 0
	for _, check := range checks {
		if check.Kind != "reference_condition" || check.Status != status.OutcomeOK || check.BindingKey == "" {
			continue
		}
		var target lexicon.SemanticTarget
		if json.Unmarshal([]byte(check.BindingKey), &target) != nil {
			continue
		}
		if target.Kind == expected.Kind && target.SymbolRef == expected.SymbolRef && target.ReferencePort != nil && expected.ReferencePort != nil && target.ReferencePort.PortID == expected.ReferencePort.PortID && len(target.RolePorts) == 0 {
			matches++
		}
	}
	return matches == 1
}

func dseSourceRefusal(candidate semantic.DSECandidate) (status.OutcomeKind, string, string) {
	outcome := candidate.Status
	if outcome != status.OutcomeOK {
		return outcome, "DSE_SOURCE_CANDIDATE_NOT_COMPLETE", "only a fully successful source candidate may reach target realization"
	}
	for _, check := range candidate.FragmentChecks {
		if check.Status != status.OutcomeOK {
			if !check.Status.Valid() {
				return status.OutcomeInvalid, "DSE_SOURCE_FRAGMENT_STATUS_INVALID", "a source candidate fragment check has an invalid status"
			}
			return mergeEngineOutcome(status.OutcomeOK, check.Status), "DSE_SOURCE_FRAGMENT_NOT_VERIFIED", "a source candidate fragment check is not successful"
		}
	}
	if len(candidate.Residuals) != 0 {
		return status.OutcomeIncomplete, "DSE_SOURCE_CANDIDATE_HAS_RESIDUALS", "source candidate has unresolved semantic residuals"
	}
	return status.OutcomeOK, "", ""
}

func mergeDSEOutcomes(outcomes ...status.OutcomeKind) status.OutcomeKind {
	result := status.OutcomeOK
	for _, outcome := range outcomes {
		result = mergeEngineOutcome(result, outcome)
	}
	return result
}

func dseOutputLimitResult(request Request, profile *DSEProfile, identity pack.Identity) TranslationResult {
	issue := diagnostic("DSE_RESULT_BYTE_LIMIT", "result", "complete DSE trace and realization exceed the 4 MiB public result limit; details were omitted")
	composition := semantic.DSEResult{
		Status: status.OutcomeResourceLimit, Complete: false,
		Candidates: []semantic.DSECandidate{}, Residuals: []semantic.DSEResidual{},
		Diagnostics: []status.Diagnostic{issue}, WholeVerificationScope: "details omitted because the bounded public result limit was reached",
	}
	return TranslationResult{
		Status: status.OutcomeResourceLimit, From: request.From, To: request.To,
		Candidates: []TranslationAlternative{}, Diagnostics: []status.Diagnostic{issue},
		DSETrace: &DSETranslationTrace{
			ProfileID: profile.ID, ProfileVersion: profile.Version,
			RegistryFingerprint: profile.RegistryFingerprint, Registries: append([]DSERegistryPin(nil), profile.Registries...),
			PackIdentity: identity, Composition: composition, Realizations: []DSETranslationRealization{},
		},
	}
}

func dseUnresolvedAlternative(candidate semantic.DSECandidate, outcome status.OutcomeKind, diagnostics []status.Diagnostic) TranslationAlternative {
	if len(diagnostics) == 0 {
		diagnostics = []status.Diagnostic{diagnostic("DSE_CANDIDATE_UNRESOLVED", "composition", "DSE candidate remains unresolved")}
	}
	return TranslationAlternative{
		ID: translationAlternativeID(candidate.ID, "unrealized"), SourceCandidateID: candidate.ID, SemanticID: candidate.ID,
		Outcome: outcome, Diagnostics: append([]status.Diagnostic(nil), diagnostics...),
	}
}

func dseFailure(result TranslationResult, outcome status.OutcomeKind, code, path, message string) TranslationResult {
	result.Status = outcome
	result.Candidates = []TranslationAlternative{}
	result.Diagnostics = []status.Diagnostic{diagnostic(code, path, message)}
	finalized, err := finalizeResult(result)
	if err != nil {
		return TranslationResult{Status: status.OutcomeError, From: result.From, To: result.To, Candidates: []TranslationAlternative{}, Diagnostics: []status.Diagnostic{diagnostic("DSE_RESULT_VALIDATION_FAILED", "dse_profile", "DSE refusal could not be represented")}}
	}
	return finalized
}

func dseTranslationAsk(expression *algebra.Term) (answerID, answerRole string, isAsk, supported bool) {
	if expression == nil {
		return "", "", false, false
	}
	if expression.Kind == algebra.TermAtom {
		return "", "", false, true
	}
	if expression.Kind != algebra.TermAsk || len(expression.Answers) != 1 || expression.Restrictor != nil || expression.Body == nil || expression.Body.Kind != algebra.TermAtom {
		return "", "", false, false
	}
	answer := expression.Answers[0]
	if answer.ID == "" || answer.Type.Kind != algebra.TypeEntity {
		return "", "", false, false
	}
	occurrences := 0
	for _, argument := range expression.Body.Args {
		if argument.Value != nil && argument.Value.Kind == algebra.TermVar && argument.Value.ID == answer.ID {
			occurrences++
			answerRole = argument.Role
		}
	}
	if occurrences != 1 || answerRole == "" {
		return "", "", false, false
	}
	return answer.ID, answerRole, true, true
}

func dseTranslationExpressionSupported(expression *algebra.Term) bool {
	_, _, _, supported := dseTranslationAsk(expression)
	return supported
}

func enrichDSEContext(engine *Engine, document algebra.Document, candidate semantic.DSECandidate) (algebra.Document, []DSEReferenceConceptMapping, status.OutcomeKind, string, string) {
	answerID, answerRole, isAsk, supported := dseTranslationAsk(document.Expression)
	if !supported {
		return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_EXPRESSION_PROFILE_UNSUPPORTED", "translation accepts one direct positive Atom or one Entity WH Ask without a restrictor over a direct Atom"
	}
	if document.RequestContext == nil {
		return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_REQUEST_CONTEXT_REQUIRED", "canonical document has no explicit request context"
	}
	if engine.snapshot == nil || engine.snapshot.Lexicon() == nil {
		return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_PACK_LEXICON_UNAVAILABLE", "runtime pack has no approved lexical transfer index"
	}
	type sourceReference struct{ role, referenceID, senseID string }
	sourceReferences := make([]sourceReference, 0, len(candidate.RoleSources)+1)
	if candidate.HeadEventReferenceID == "" || candidate.Source.SenseID == "" {
		return algebra.Document{}, nil, status.OutcomeIncomplete, "DSE_EVENT_SENSE_REQUIRED", "source candidate has no exact event sense and reference"
	}
	sourceReferences = append(sourceReferences, sourceReference{role: "event", referenceID: candidate.HeadEventReferenceID, senseID: candidate.Source.SenseID})
	bindingsByReference := make(map[string]string, len(candidate.RoleSources)+1)
	bindingsByReference[candidate.HeadEventReferenceID] = candidate.Source.SenseID
	whGapCount := 0
	for _, role := range candidate.RoleSources {
		if role.WHGap != nil {
			whGapCount++
			if !isAsk || role.WHGap.Role != role.Role || role.Role != answerRole {
				return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_WH_GAP_MISMATCH", "the typed WH answer must match exactly one source gap in the same semantic role"
			}
			continue
		}
		if isAsk && role.Role == answerRole {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_WH_ROLE_DUPLICATED", "the answer role cannot also be mapped to an explicit source reference"
		}
		if role.ReferenceID == "" || role.Source.SenseID == "" {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_ROLE_REFERENCE_UNRESOLVED", "every realized role requires an exact source sense and explicit reference"
		}
		if prior, exists := bindingsByReference[role.ReferenceID]; exists && prior != role.Source.SenseID {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_SHARED_REFERENCE_SENSE_CONFLICT", "one source reference has conflicting event or role senses"
		}
		bindingsByReference[role.ReferenceID] = role.Source.SenseID
		sourceReferences = append(sourceReferences, sourceReference{role: role.Role, referenceID: role.ReferenceID, senseID: role.Source.SenseID})
	}
	if (isAsk && whGapCount != 1) || (!isAsk && whGapCount != 0) {
		return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_WH_GAP_CARDINALITY", "the typed WH Ask requires exactly one matching source gap and a direct Atom permits none"
	}
	if isAsk && !dseQuestionSourcesMatch(document.Expression.Body, candidate, answerID) {
		return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_WH_ARGUMENT_MAPPING_MISMATCH", "the Ask atom arguments must match the event reference, explicit role references, and one typed source WH gap"
	}
	knownDeclarations := make(map[string]algebra.Declaration, len(document.References))
	for _, declaration := range document.References {
		knownDeclarations[declaration.ID] = declaration
	}
	context := *document.RequestContext
	context.Referents = append([]algebra.ContextReferent(nil), document.RequestContext.Referents...)
	referents := make(map[string]int, len(context.Referents))
	for index, referent := range context.Referents {
		if _, duplicate := referents[referent.ID]; duplicate {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_DUPLICATE_CONTEXT_REFERENT", "request context contains duplicate referent IDs"
		}
		referents[referent.ID] = index
	}
	resolved := make(map[string]DSEReferenceConceptMapping, len(bindingsByReference))
	for _, sourceRef := range sourceReferences {
		referenceID, senseText := sourceRef.referenceID, sourceRef.senseID
		declaration, declared := knownDeclarations[referenceID]
		index, present := referents[referenceID]
		if !declared || !present || string(declaration.Type.Kind) == "" || context.Referents[index].Type.Kind != declaration.Type.Kind {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_REFERENCE_CONTEXT_MISMATCH", "source sense reference must be declared and context-typed in the canonical document"
		}
		sense, err := ids.ParseSenseID(senseText)
		if err != nil {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_SOURCE_SENSE_INVALID", "source candidate contains a malformed exact sense ID"
		}
		bindings := engine.snapshot.Lexicon().TransferBindings(sense)
		if len(bindings) == 0 {
			return algebra.Document{}, nil, status.OutcomeUnknown, "DSE_SOURCE_SENSE_UNBOUND", "source sense has no approved equivalent CORE transfer binding"
		}
		if len(bindings) != 1 || !bindings[0].TransferAuthorized() {
			return algebra.Document{}, nil, status.OutcomeAmbiguous, "DSE_SOURCE_SENSE_AMBIGUOUS", "source sense does not have exactly one approved equivalent CORE transfer binding"
		}
		conceptID := bindings[0].CoreRef.String()
		referent := &context.Referents[index]
		if referent.ConceptID != "" && referent.ConceptID != conceptID {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_REFERENT_CONCEPT_CONFLICT", fmt.Sprintf("reference %s has an explicit concept that conflicts with its exact source sense", referenceID)
		}
		referent.ConceptID = conceptID
		resolved[referenceID] = DSEReferenceConceptMapping{
			Role: sourceRef.role, ReferenceID: referenceID, SourceSenseID: senseText, CoreConceptID: conceptID,
			MappingKind: string(bindings[0].Mapping), ApprovalStatus: string(bindings[0].Status), ProvenanceID: bindings[0].Provenance.ID,
		}
	}
	mappings := make([]DSEReferenceConceptMapping, 0, len(sourceReferences))
	for _, sourceRef := range sourceReferences {
		mapping := resolved[sourceRef.referenceID]
		mapping.Role = sourceRef.role
		mappings = append(mappings, mapping)
	}
	document.RequestContext = &context
	return document, mappings, status.OutcomeOK, "", ""
}

func dseQuestionSourcesMatch(atom *algebra.Term, candidate semantic.DSECandidate, answerID string) bool {
	if atom == nil || atom.Kind != algebra.TermAtom || answerID == "" || candidate.HeadEventReferenceID == "" {
		return false
	}
	roles := make(map[string]semantic.DSERoleSource, len(candidate.RoleSources))
	for _, role := range candidate.RoleSources {
		if role.Role == "" {
			return false
		}
		if _, duplicate := roles[role.Role]; duplicate {
			return false
		}
		roles[role.Role] = role
	}
	seenEvent := false
	seenRoles := make(map[string]struct{}, len(roles))
	for _, argument := range atom.Args {
		if argument.Value == nil || argument.Role == "" {
			return false
		}
		if argument.Role == "event" {
			if seenEvent || argument.Value.Kind != algebra.TermRef || argument.Value.ID != candidate.HeadEventReferenceID {
				return false
			}
			seenEvent = true
			continue
		}
		role, exists := roles[argument.Role]
		if !exists {
			return false
		}
		if _, duplicate := seenRoles[argument.Role]; duplicate {
			return false
		}
		seenRoles[argument.Role] = struct{}{}
		if role.WHGap != nil {
			if role.WHGap.Role != role.Role || role.Role != argument.Role || argument.Value.Kind != algebra.TermVar || argument.Value.ID != answerID {
				return false
			}
			continue
		}
		if role.ReferenceID == "" || argument.Value.Kind != algebra.TermRef || argument.Value.ID != role.ReferenceID {
			return false
		}
	}
	return seenEvent && len(seenRoles) == len(roles)
}
