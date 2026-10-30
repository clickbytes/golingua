package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/provenance"
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
		scopePlan, scoped := matchAlgebraScope(document.Expression, signaturesByID(document.Signatures))
		if scoped {
			if code, message := validateDSEScopedSourceEvidence(candidate, document, scopePlan, sourceResources.SourcePackage); code != "" {
				refusal := diagnostic(code, "composition.candidates."+candidate.ID+".scoped_evidence", message)
				realizationReceipt.Status = status.OutcomeInvalid
				realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
				trace.Realizations = append(trace.Realizations, realizationReceipt)
				result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
				result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeInvalid, []status.Diagnostic{refusal}))
				continue
			}
			// Keep all canonical source declarations, references and request-local
			// context. The target verifier owns scope matching; this bridge removes
			// only inline sorts/signatures so the separately pinned metadata is the
			// sole declaration authority.
			targetDocument := document
			targetDocument.Sorts = nil
			targetDocument.Signatures = nil
			realized, realizeErr := engine.RealizeDSEAlgebra(requestContext, request.To, targetDocument, targetResources.TargetPackage, profile.SelectedMetadata)
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
			realizationReceipt.ScopeCoverage = realized.ScopeCoverage
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
			continue
		}
		if eventPlan, isEventRelation := matchAlgebraEventRelationPath(document.Expression, signaturesByID(document.Signatures)); isEventRelation {
			_, code, message := validateDSERelationSourceEvidence(candidate, document, eventPlan, sourceResources.SourcePackage, profile.SelectedMetadata, request.Text)
			if code != "" {
				refusal := diagnostic(code, "composition.candidates."+candidate.ID+".relation_source_evidence", message)
				realizationReceipt.Status = status.OutcomeInvalid
				realizationReceipt.Diagnostics = []status.Diagnostic{refusal}
				trace.Realizations = append(trace.Realizations, realizationReceipt)
				result.Status = mergeEngineOutcome(result.Status, status.OutcomeInvalid)
				result.Candidates = append(result.Candidates, dseUnresolvedAlternative(candidate, status.OutcomeInvalid, []status.Diagnostic{refusal}))
				continue
			}
			targetDocument := document
			targetDocument.Sorts = nil
			targetDocument.Signatures = nil
			realized, realizeErr := engine.RealizeDSEAlgebra(requestContext, request.To, targetDocument, targetResources.TargetPackage, profile.SelectedMetadata)
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
			realizationReceipt.RelationPathCoverage = realized.RelationPathCoverage
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
				outcome := mergeDSEOutcomes(candidate.Status, surface.Outcome)
				if len(realized.Residuals) > 0 {
					outcome = mergeDSEOutcomes(outcome, status.OutcomeIncomplete)
				}
				diagnostics := append([]status.Diagnostic(nil), surface.Diagnostics...)
				if outcome != status.OutcomeOK && len(diagnostics) == 0 {
					diagnostics = append([]status.Diagnostic(nil), realized.Diagnostics...)
				}
				result.Candidates = append(result.Candidates, TranslationAlternative{
					ID: translationAlternativeID(candidate.ID, surface.ID), SourceCandidateID: candidate.ID, SemanticID: candidate.ID,
					Outcome: outcome, Text: surface.Text, Evidence: []semantic.EvidenceTrace{surface.Evidence.Clone()},
					Assumptions: cloneFieldAssumptions(surface.Assumptions), Losses: append([]realizer.FieldLoss(nil), surface.Losses...), Diagnostics: diagnostics,
				})
			}
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
		var enriched algebra.Document
		var mappings []DSEReferenceConceptMapping
		var mapStatus status.OutcomeKind
		var mapCode, mapMessage string
		if document.Expression != nil && document.Expression.Kind == algebra.TermAnd {
			enriched, mappings, mapStatus, mapCode, mapMessage = enrichDSECoordination(engine, sourceResources.SourcePackage, document, candidate)
		} else {
			enriched, mappings, mapStatus, mapCode, mapMessage = enrichDSEContext(engine, document, candidate)
		}
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
		} else if dseCoordinationDefinitePreview(targetDocument, candidate) || dsePredicateConditionPreview(targetDocument, candidate) {
			// A checked source definite presupposition may remain unexpressed in
			// the target language. Permit text as an explicitly incomplete preview
			// only for this narrow coordinated shape; the target residual remains
			// attached and no condition is consumed here.
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
	if result.Status == status.OutcomeOK && request.PreviewID == "" && len(result.Candidates) == 1 && result.Candidates[0].Outcome == status.OutcomeOK {
		result.Text = result.Candidates[0].Text
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

func dseCoordinationDefinitePreview(document algebra.Document, candidate semantic.DSECandidate) bool {
	root := document.Expression
	if root == nil || root.Kind != algebra.TermAnd || len(root.Operands) != 2 || root.Operands[0] == nil || root.Operands[0].Kind != algebra.TermAtom || root.Operands[1] == nil || root.Operands[1].Kind != algebra.TermAtom || document.RequestContext == nil || len(document.RequestContext.Conditions) == 0 {
		return false
	}
	requestContext := document.RequestContext
	if requestContext.SpeakerReferentID != "" || requestContext.AddresseeReferentID != "" || len(requestContext.Mentions) != 0 || len(requestContext.ResolutionEvidence) != 0 {
		return false
	}
	roleReferences := map[string]map[string]bool{}
	eventReferences := map[string]bool{}
	for _, operand := range root.Operands {
		for _, argument := range operand.Args {
			if argument.Role == "" || argument.Value == nil || argument.Value.Kind != algebra.TermRef || argument.Value.Type == nil {
				return false
			}
			if argument.Value.Type.Kind == algebra.TypeEvent {
				eventReferences[argument.Value.ID] = true
				continue
			}
			if argument.Value.Type.Kind != algebra.TypeEntity {
				return false
			}
			if roleReferences[argument.Value.ID] == nil {
				roleReferences[argument.Value.ID] = map[string]bool{}
			}
			roleReferences[argument.Value.ID][argument.Role] = true
		}
	}
	for _, referent := range requestContext.Referents {
		switch referent.Type.Kind {
		case algebra.TypeEvent:
			if referent.ReferenceTimeID != "" || referent.Multiplicity != "" || !eventReferences[referent.ID] {
				return false
			}
		case algebra.TypeEntity:
			if referent.TemporalFrame != "" || referent.ReferenceTimeID != "" || len(roleReferences[referent.ID]) == 0 {
				return false
			}
		default:
			return false
		}
	}
	for _, condition := range document.RequestContext.Conditions {
		if condition.Kind != algebra.ContextPresupposition || condition.Code != "definite_reference" || condition.Expression == nil || condition.Expression.Kind != algebra.TermAtom || len(condition.Expression.Args) != 1 || condition.Expression.Args[0].Role == "" {
			return false
		}
		argument := condition.Expression.Args[0]
		if argument.Value == nil || argument.Value.Kind != algebra.TermRef || argument.Value.Type == nil || argument.Value.Type.Kind != algebra.TypeEntity || len(condition.TargetReferentIDs) != 1 || condition.TargetReferentIDs[0] != argument.Value.ID {
			return false
		}
		roles, referenced := roleReferences[argument.Value.ID]
		if !referenced || len(roles) == 0 {
			return false
		}
		covered := false
		for _, roleSource := range candidate.RoleSources {
			if roleSource.ReferenceID == argument.Value.ID && roleSource.WHGap == nil && roles[roleSource.Role] {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
		coverageCount := 0
		for _, coverage := range candidate.ConditionCoverage {
			if coverage.ConditionID != condition.ID {
				continue
			}
			if coverage.Kind != "presupposition" || coverage.Code != condition.Code || len(coverage.TargetReferenceIDs) != 1 || coverage.TargetReferenceIDs[0] != argument.Value.ID {
				return false
			}
			coverageCount++
		}
		if coverageCount == 0 {
			return false
		}
	}
	return true
}

// dsePredicateConditionPreview permits an incomplete target preview when all
// remaining reference conditions are explicitly linked to predicate roles.
// It does not claim the target language expressed those conditions.
func dsePredicateConditionPreview(document algebra.Document, candidate semantic.DSECandidate) bool {
	atom := document.Expression
	if atom != nil && atom.Kind == algebra.TermNot {
		atom = atom.Body
	}
	if atom == nil || atom.Kind != algebra.TermAtom || document.RequestContext == nil || len(document.RequestContext.Conditions) == 0 {
		return false
	}
	roleReferences := make(map[string]string, len(candidate.RoleSources))
	for _, source := range candidate.RoleSources {
		if source.WHGap != nil || source.Role == "" || source.ReferenceID == "" {
			return false
		}
		if _, duplicate := roleReferences[source.Role]; duplicate {
			return false
		}
		roleReferences[source.Role] = source.ReferenceID
	}
	if len(roleReferences) == 0 {
		return false
	}
	for _, argument := range atom.Args {
		if argument.Role == "event" {
			if candidate.HeadEventReferenceID == "" || argument.Value == nil || argument.Value.Kind != algebra.TermRef || argument.Value.ID != candidate.HeadEventReferenceID {
				return false
			}
			continue
		}
		if argument.Value == nil || argument.Value.Kind != algebra.TermRef || roleReferences[argument.Role] != argument.Value.ID {
			return false
		}
	}
	seenConditions := 0
	for _, condition := range document.RequestContext.Conditions {
		if condition.Kind == algebra.ContextPresupposition && condition.Code == "deictic_anchor" {
			if !dsePreviewParticipantAnchor(document, candidate, condition, roleReferences) {
				return false
			}
			seenConditions++
			continue
		}
		if condition.Kind != algebra.ContextPresupposition || (condition.Code != "definite_reference" && condition.Code != "indefinite_reference") || condition.Expression == nil || condition.Expression.Kind != algebra.TermAtom || len(condition.Expression.Args) != 1 || len(condition.TargetReferentIDs) != 1 {
			return false
		}
		argument := condition.Expression.Args[0]
		if argument.Role != "entity" || argument.Value == nil || argument.Value.Kind != algebra.TermRef || argument.Value.Type == nil || argument.Value.Type.Kind != algebra.TypeEntity || argument.Value.ID != condition.TargetReferentIDs[0] {
			return false
		}
		roleLinked := false
		for _, referenceID := range roleReferences {
			if referenceID == argument.Value.ID {
				roleLinked = true
			}
		}
		if !roleLinked {
			return false
		}
		covered := false
		for _, coverage := range candidate.ConditionCoverage {
			if coverage.ConditionID == condition.ID && string(coverage.Kind) == string(condition.Kind) && coverage.Code == condition.Code && len(coverage.TargetReferenceIDs) == 1 && coverage.TargetReferenceIDs[0] == argument.Value.ID {
				covered = true
			}
		}
		if !covered {
			return false
		}
		seenConditions++
	}
	return seenConditions > 0
}

func dsePreviewParticipantAnchor(document algebra.Document, candidate semantic.DSECandidate, condition algebra.ContextCondition, roleReferences map[string]string) bool {
	if document.RequestContext == nil || condition.Expression != nil || len(condition.TargetReferentIDs) != 1 {
		return false
	}
	referenceID, anchor := condition.TargetReferentIDs[0], condition.Detail
	request := document.RequestContext
	if anchor == "speaker" && request.SpeakerReferentID != referenceID ||
		anchor == "speaker_inclusive_group" && request.SpeakerReferentID != referenceID ||
		anchor == "addressee" && request.AddresseeReferentID != referenceID {
		return false
	}
	if anchor != "speaker" && anchor != "speaker_inclusive_group" && anchor != "addressee" {
		return false
	}
	roleLinked := false
	for _, id := range roleReferences {
		if id == referenceID {
			roleLinked = true
		}
	}
	if !roleLinked {
		return false
	}
	multiplicity := ""
	for _, referent := range request.Referents {
		if referent.ID == referenceID && referent.Type.Kind == algebra.TypeEntity && referent.ConceptID == "concept:person" {
			multiplicity = referent.Multiplicity
		}
	}
	if anchor == "speaker" && multiplicity != "one" || anchor == "speaker_inclusive_group" && multiplicity != "many" || multiplicity == "" {
		return false
	}
	for _, reference := range document.References {
		if reference.ID == referenceID && reference.Type.Kind == algebra.TypeEntity && (reference.Multiplicity == "" || reference.Multiplicity == multiplicity) {
			for _, coverage := range candidate.ConditionCoverage {
				if coverage.ConditionID == condition.ID && string(coverage.Kind) == string(condition.Kind) && coverage.Code == condition.Code && len(coverage.TargetReferenceIDs) == 1 && coverage.TargetReferenceIDs[0] == referenceID {
					return true
				}
			}
		}
	}
	return false
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
	coordinated := len(candidate.ClauseSources) == 2
	for index := range mappings {
		if mappings[index].ReferenceID != argument.Value.ID {
			continue
		}
		if mappings[index].MappingKind != string(lexicon.MappingEquivalent) || mappings[index].ApprovalStatus != string(lexicon.BindingApproved) {
			return DSEConditionConsumption{}, false
		}
		if mapping != nil {
			if !coordinated || mapping.SourceSenseID != mappings[index].SourceSenseID || mapping.CoreConceptID != mappings[index].CoreConceptID || mapping.MappingKind != mappings[index].MappingKind || mapping.ApprovalStatus != mappings[index].ApprovalStatus || mapping.ProvenanceID != mappings[index].ProvenanceID {
				return DSEConditionConsumption{}, false
			}
			continue
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
	if roleSources < 1 || (roleSources != 1 && !coordinated) {
		return DSEConditionConsumption{}, false
	}
	coverage := 0
	for _, row := range candidate.ConditionCoverage {
		if row.ConditionID != condition.ID {
			continue
		}
		if row.Code != condition.Code || string(row.Kind) != string(condition.Kind) || len(row.TargetReferenceIDs) != 1 || row.TargetReferenceIDs[0] != mapping.ReferenceID {
			return DSEConditionConsumption{}, false
		}
		if !coordinated && coverage != 0 {
			return DSEConditionConsumption{}, false
		}
		coverage++
	}
	if coverage < 1 || (coverage != 1 && !coordinated) {
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
	if selected == nil || !hasDSEReferenceFragment(candidate.FragmentChecks, selected.Target, coordinated) {
		return DSEConditionConsumption{}, false
	}
	return DSEConditionConsumption{
		ConditionID: condition.ID, Kind: string(condition.Kind), Code: condition.Code,
		Symbol: condition.Expression.Symbol, ReferenceID: mapping.ReferenceID,
		SourceSenseID: mapping.SourceSenseID, CoreConceptID: mapping.CoreConceptID,
		SourceBindingProvenance: selected.Provenance.ID,
	}, true
}

func hasDSEReferenceFragment(checks []semantic.DSEFragmentCheck, expected *lexicon.SemanticTarget, allowRepeated bool) bool {
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
	return matches == 1 || (allowRepeated && matches > 1)
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
	if expression.Kind == algebra.TermNot && expression.Body != nil && expression.Body.Kind == algebra.TermAtom {
		return "", "", false, true
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
	if expression != nil && expression.Kind == algebra.TermAnd && len(expression.Operands) == 2 && expression.Operands[0] != nil && expression.Operands[0].Kind == algebra.TermAtom && expression.Operands[1] != nil && expression.Operands[1].Kind == algebra.TermAtom {
		return true
	}
	_, _, _, supported := dseTranslationAsk(expression)
	return supported
}

func validateDSEScopedSourceEvidence(candidate semantic.DSECandidate, document algebra.Document, plan algebraScopePlan, sourcePackage *language.SemanticLanguagePackage) (string, string) {
	evidence := candidate.ScopedEvidence
	if evidence == nil || sourcePackage == nil || sourcePackage.LexiconIndex() == nil || sourcePackage.ValencyIndex() == nil {
		return "DSE_SCOPE_EVIDENCE_REQUIRED", "checked scoped expression requires retained scoped lexical evidence and its selected source package"
	}
	if plan.OuterBinderID != "x" || plan.InnerBinderID != "y" || plan.EventBinderID != "e" || evidence.Variant.ID == "" || evidence.Variant.ID != candidate.Source.VariantID || evidence.SubjectBinder != evidence.Variant.SubjectBinder || evidence.ObjectBinder == evidence.SubjectBinder || evidence.SubjectBinder != "x" && evidence.SubjectBinder != "y" || evidence.ObjectBinder != "x" && evidence.ObjectBinder != "y" {
		return "DSE_SCOPE_EVIDENCE_MISMATCH", "source variant and binder evidence do not agree with the independently matched x/y/e scope"
	}
	if evidence.Variant.Negation == "outer" && (!plan.OuterNegated || plan.InnerNegated) || evidence.Variant.Negation == "inner" && (!plan.InnerNegated || plan.OuterNegated) || evidence.Variant.Negation != "outer" && evidence.Variant.Negation != "inner" {
		return "DSE_SCOPE_NEGATION_MISMATCH", "source variant negation placement differs from the checked algebra expression"
	}
	if candidate.Source.FrameID != string(evidence.Frame.ID) || evidence.Frame.ID == "" || evidence.Frame.SenseID.String() == "" {
		return "DSE_SCOPE_FRAME_MISMATCH", "source frame evidence does not identify the event frame in the candidate"
	}
	if plan.EventFrame != scopedReadingTemporal(evidence.Bindings) {
		return "DSE_SCOPE_TEMPORAL_MISMATCH", "event temporal_frame differs from the retained finite morphology reading"
	}
	if !sameScopedDeclarations(evidence.SourceReferences.References, document.References) || !reflect.DeepEqual(evidence.SourceReferences.RequestContext, document.RequestContext) {
		return "DSE_SCOPE_CONTEXT_EVIDENCE_MISMATCH", "canonical references or request context differ from retained scoped source context evidence"
	}
	if document.RequestContext != nil && (evidence.SourceReferences.ContextID == "" || evidence.SourceReferences.ContextID != document.RequestContext.ID) {
		return "DSE_SCOPE_CONTEXT_ID_MISMATCH", "retained source context ID does not match the canonical request context"
	}

	byBinder := map[string]semantic.DSEScopedBindingEvidence{}
	eventCount := 0
	for _, binding := range evidence.Bindings {
		if binding.Binder != "x" && binding.Binder != "y" && binding.Binder != "event" || binding.TokenIndex < 0 || binding.Span.Start < 0 || binding.Span.End <= binding.Span.Start || binding.Reading.ID == "" || binding.Lexeme.ID.String() == "" || binding.Sense.ID.String() == "" || binding.Sense.LexemeID != binding.Lexeme.ID || binding.Reading.LexemeID != binding.Lexeme.ID.String() {
			return "DSE_SCOPE_BINDING_EVIDENCE_INVALID", "scoped binder evidence has malformed token, span, reading, lexeme or sense identity"
		}
		if binding.Binder == "event" {
			eventCount++
			if eventCount > 1 {
				return "DSE_SCOPE_EVENT_EVIDENCE_AMBIGUOUS", "scoped evidence must contain exactly one event reading"
			}
		}
		if previous, exists := byBinder[binding.Binder]; exists {
			if previous.Sense.ID != binding.Sense.ID || previous.Lexeme.ID != binding.Lexeme.ID {
				return "DSE_SCOPE_REPEATED_BINDER_MISMATCH", "repeated binder readings do not retain one exact lexical sense and lexeme"
			}
		} else {
			byBinder[binding.Binder] = binding
		}
	}
	if byBinder["x"].Sense.ID.String() == "" || byBinder["y"].Sense.ID.String() == "" || byBinder["event"].Sense.ID.String() == "" || eventCount != 1 || byBinder["event"].Sense.ID.String() != candidate.Source.SenseID || byBinder["event"].Reading.ID != candidate.Source.ReadingID || byBinder["event"].Sense.ID != evidence.Frame.SenseID {
		return "DSE_SCOPE_BINDER_EVIDENCE_INCOMPLETE", "x, y and event evidence must match the source candidate and selected frame"
	}
	frameFound := false
	for _, frame := range sourcePackage.ValencyIndex().FramesBySenseID(evidence.Frame.SenseID) {
		if frame.ID == evidence.Frame.ID && reflect.DeepEqual(frame, evidence.Frame) {
			frameFound = true
			break
		}
	}
	if !frameFound {
		return "DSE_SCOPE_FRAME_NOT_SELECTED", "retained frame is not the exact frame in the selected source package"
	}

	subjectRoles, objectRoles := 0, 0
	for _, argument := range evidence.Frame.Arguments {
		if string(argument.Role) == evidence.SubjectRole && argument.GrammaticalFunction == "subject" {
			subjectRoles++
		}
		if string(argument.Role) == evidence.ObjectRole && argument.GrammaticalFunction == "object" && (evidence.Variant.FrameObjectCase == "" || argument.Case == evidence.Variant.FrameObjectCase) {
			objectRoles++
		}
	}
	if subjectRoles != 1 || objectRoles != 1 || evidence.SubjectRole == "" || evidence.ObjectRole == "" || evidence.SubjectRole == evidence.ObjectRole {
		return "DSE_SCOPE_ROLE_EVIDENCE_MISMATCH", "retained subject/object roles do not match the selected valency frame and variant case"
	}
	rolePorts := map[string]semantic.DSEScopedRoleEvidence{}
	for _, role := range evidence.RolePorts {
		if role.FrameID != string(evidence.Frame.ID) || role.Role == "" || role.PortID == "" || role.Type.Kind != algebra.TypeEntity {
			return "DSE_SCOPE_ROLE_PORT_INVALID", "retained role port is malformed or belongs to another frame"
		}
		if _, duplicate := rolePorts[role.Role]; duplicate {
			return "DSE_SCOPE_ROLE_PORT_AMBIGUOUS", "retained frame role has multiple target ports"
		}
		rolePorts[role.Role] = role
	}
	if len(rolePorts) != 2 || rolePorts[evidence.SubjectRole].PortID == "" || rolePorts[evidence.ObjectRole].PortID == "" || rolePorts[evidence.SubjectRole].PortID == rolePorts[evidence.ObjectRole].PortID {
		return "DSE_SCOPE_ROLE_PORT_MISMATCH", "exactly two distinct selected Entity role ports must map subject and object roles"
	}
	roleForBinder := map[string]string{evidence.SubjectBinder: evidence.SubjectRole, evidence.ObjectBinder: evidence.ObjectRole}
	if rolePorts[roleForBinder["x"]].PortID != plan.EventOuterPort || rolePorts[roleForBinder["y"]].PortID != plan.EventInnerPort {
		return "DSE_SCOPE_IR_ROLE_MISMATCH", "event atom binder ports do not match retained source role and binder evidence"
	}
	for _, binding := range byBinder {
		selectedSense, senseOK := sourcePackage.LexiconIndex().BySenseID(binding.Sense.ID)
		selectedLexeme, lexemeOK := sourcePackage.LexiconIndex().ByLexemeID(binding.Lexeme.ID)
		if !senseOK || !lexemeOK || selectedSense.ID != binding.Sense.ID || selectedSense.LexemeID != binding.Sense.LexemeID || selectedSense.LexemeID != binding.Lexeme.ID || selectedLexeme.ID != binding.Lexeme.ID || selectedLexeme.Language != binding.Lexeme.Language || selectedLexeme.Lemma != binding.Lexeme.Lemma || selectedLexeme.POS != binding.Lexeme.POS || binding.Reading.LexemeID != selectedLexeme.ID.String() {
			return "DSE_SCOPE_LEXICAL_IDENTITY_NOT_SELECTED", "retained sense ID, sense-to-lexeme link, or lexeme language/lemma/POS differs from the selected source package"
		}
		if binding.Reading.Provenance != nil && !containsScopedProvenance(evidence.Provenance, *binding.Reading.Provenance) {
			return "DSE_SCOPE_READING_PROVENANCE_MISSING", "retained reading provenance is absent from scoped evidence"
		}
	}
	if !containsScopedProvenance(evidence.Provenance, evidence.Frame.Provenance) {
		return "DSE_SCOPE_FRAME_PROVENANCE_MISSING", "selected frame provenance is absent from scoped evidence"
	}
	if !hasScopedReferenceTarget(sourcePackage, byBinder["x"], plan.Restrictor, evidence.Provenance) || !hasScopedReferenceTarget(sourcePackage, byBinder["y"], plan.Predicate, evidence.Provenance) {
		return "DSE_SCOPE_REFERENCE_TARGET_MISMATCH", "scoped noun atoms do not match their exact approved source reference targets"
	}
	if !hasScopedEventTarget(sourcePackage, byBinder["event"], evidence, plan) {
		return "DSE_SCOPE_EVENT_TARGET_MISMATCH", "scoped event atom does not match its exact approved source symbol target and frame role ports"
	}
	for _, source := range evidence.Provenance {
		if source.Validate() != nil {
			return "DSE_SCOPE_PROVENANCE_INVALID", "scoped source evidence contains invalid provenance"
		}
	}
	return "", ""
}

func hasScopedReferenceTarget(sourcePackage *language.SemanticLanguagePackage, binding semantic.DSEScopedBindingEvidence, atom *algebra.Term, proof []provenance.SourceRef) bool {
	if atom == nil || atom.Kind != algebra.TermAtom || len(atom.Args) != 1 || atom.Args[0].Value == nil || atom.Args[0].Value.Kind != algebra.TermVar || atom.Args[0].Value.ID != binding.Binder {
		return false
	}
	matches := 0
	for _, target := range sourcePackage.LexiconIndex().TargetBindingsForSense(binding.Sense.ID) {
		if target.Target != nil && target.Target.Kind == lexicon.SemanticTargetReferencePredicate && target.Target.SymbolRef == lexicon.SymbolRef(atom.Symbol) && target.Target.ReferencePort != nil && target.Target.ReferencePort.PortID == atom.Args[0].Role && target.Mapping == lexicon.MappingEquivalent && target.Status == lexicon.BindingApproved && target.TargetExecutionAuthorized() && containsScopedProvenance(proof, target.Provenance) {
			matches++
		}
	}
	return matches == 1
}

func hasScopedEventTarget(sourcePackage *language.SemanticLanguagePackage, binding semantic.DSEScopedBindingEvidence, evidence *semantic.DSEScopedEvidence, plan algebraScopePlan) bool {
	if plan.EventAtom == nil || plan.EventAtom.Symbol == "" {
		return false
	}
	roles := make(map[string]string, len(evidence.RolePorts))
	for _, role := range evidence.RolePorts {
		roles[role.Role] = role.PortID
	}
	matching := 0
	for _, target := range sourcePackage.LexiconIndex().TargetBindingsForSense(binding.Sense.ID) {
		if target.Target == nil || target.Target.Kind != lexicon.SemanticTargetSymbol || target.Target.SymbolRef != lexicon.SymbolRef(plan.EventAtom.Symbol) || target.Mapping != lexicon.MappingEquivalent || target.Status != lexicon.BindingApproved || !target.TargetExecutionAuthorized() {
			continue
		}
		mapped := map[string]string{}
		for _, role := range target.Target.RolePorts {
			if role.FrameID == string(evidence.Frame.ID) {
				if _, duplicate := mapped[role.Role]; duplicate {
					mapped[role.Role] = "!duplicate!"
				} else {
					mapped[role.Role] = role.PortID
				}
			}
		}
		if len(mapped) != len(roles) {
			continue
		}
		equal := true
		for role, port := range roles {
			if mapped[role] != port {
				equal = false
				break
			}
		}
		if equal && containsScopedProvenance(evidence.Provenance, target.Provenance) {
			matching++
		}
	}
	return matching == 1
}

func containsScopedProvenance(rows []provenance.SourceRef, expected provenance.SourceRef) bool {
	for _, row := range rows {
		if reflect.DeepEqual(row, expected) {
			return true
		}
	}
	return false
}

func sameScopedDeclarations(left, right []algebra.Declaration) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !reflect.DeepEqual(left[i], right[i]) {
			return false
		}
	}
	return true
}

func scopedReadingTemporal(bindings []semantic.DSEScopedBindingEvidence) string {
	for _, binding := range bindings {
		if binding.Binder != "event" {
			continue
		}
		finite, tense := false, ""
		for _, feature := range binding.Reading.Features.Values {
			if string(feature.Feature) == "verb_form" && string(feature.State) == "known" && string(feature.Value) == "finite" {
				finite = true
			}
			if string(feature.Feature) == "tense" && string(feature.State) == "known" {
				tense = string(feature.Value)
			}
		}
		if finite && (tense == "past" || tense == "present") {
			return tense
		}
		return ""
	}
	return ""
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
	bindingsByReference := make(map[string]string, len(candidate.RoleSources)+1)
	participantAnchors := make(map[string]string)
	if candidate.PredicateKind == "relation" {
		if candidate.HeadEventReferenceID != "" || candidate.Source.SenseID == "" {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_RELATION_EVENT_FORBIDDEN", "relation predicate must carry its exact sense without an Event identity"
		}
		predicateSense, parseErr := ids.ParseSenseID(candidate.Source.SenseID)
		if parseErr != nil {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_RELATION_SENSE_INVALID", "relation predicate carries a malformed source sense"
		}
		predicateBindings := engine.snapshot.Lexicon().TransferBindings(predicateSense)
		if len(predicateBindings) != 1 || !predicateBindings[0].TransferAuthorized() {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_RELATION_METADATA_MISMATCH", "selected relation source sense must resolve to one approved relation concept"
		}
		concept, found := engine.core.Lookup(predicateBindings[0].CoreRef)
		if !found || concept.Kind != core.KindRelation || predicateBindings[0].CoreRef.String() != "concept:belonging_association" && predicateBindings[0].CoreRef.String() != "concept:legal_ownership" || len(concept.RoleTypes) != 2 {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_RELATION_METADATA_MISMATCH", "selected relation source sense must resolve to a two-role CORE relation concept"
		}
	} else if candidate.PredicateKind == "property" {
		if candidate.HeadEventReferenceID != "" || candidate.Source.SenseID == "" || candidate.Source.FrameID == "" || candidate.Source.FamilyID == "" || candidate.Source.VariantID == "" {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_EVIDENCE_INVALID", "property candidate must carry its exact source sense, frame, family, variant and no Event identity"
		}
		predicateSense, parseErr := ids.ParseSenseID(candidate.Source.SenseID)
		if parseErr != nil {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_SENSE_INVALID", "property candidate carries a malformed source sense"
		}
		predicateBindings := engine.snapshot.Lexicon().TransferBindings(predicateSense)
		if len(predicateBindings) != 1 || !predicateBindings[0].TransferAuthorized() {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_PROPERTY_METADATA_MISMATCH", "selected source sense must resolve to one approved property transfer"
		}
		concept, found := engine.core.Lookup(predicateBindings[0].CoreRef)
		if !found || concept.Kind != core.KindProperty || len(concept.RoleTypes) == 0 {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_PROPERTY_ROLE_SHAPE_MISMATCH", "selected CORE concept must be a property with declared roles"
		}
		propertyRoles := make(map[string]bool, len(concept.RoleTypes))
		for _, role := range concept.RoleTypes {
			if role.Name == "" || propertyRoles[role.Name] || len(role.AllowedKinds) != 1 || role.AllowedKinds[0] != core.KindEntity {
				return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_PROPERTY_ROLE_SHAPE_MISMATCH", "selected property requires unique Entity-valued CORE roles"
			}
			propertyRoles[role.Name] = true
		}
		atom := document.Expression
		if atom.Kind == algebra.TermNot {
			atom = atom.Body
		}
		if atom == nil || atom.Kind != algebra.TermAtom || atom.Symbol != predicateBindings[0].CoreRef.String() || len(candidate.RoleSources) != len(propertyRoles) || len(atom.Args) != len(propertyRoles) {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_ROLE_SHAPE_MISMATCH", "property expression must cover every selected CORE role exactly once"
		}
		seen := make(map[string]bool, len(propertyRoles))
		for _, role := range candidate.RoleSources {
			if role.WHGap != nil || !propertyRoles[role.Role] || seen[role.Role] || role.ReferenceID == "" {
				return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_ROLE_EVIDENCE_MISMATCH", "each property role must retain exact source frame, family and variant provenance"
			}
			if role.ImplicitEvidence != nil && (role.Source.FrameID != candidate.Source.FrameID || role.Source.FamilyID != candidate.Source.FamilyID || role.Source.VariantID != candidate.Source.VariantID || role.Source.DerivationID != candidate.Source.DerivationID) {
				return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_ROLE_EVIDENCE_MISMATCH", "implicit property roles must retain exact source frame, family and variant provenance"
			}
			seen[role.Role] = true
		}
		for _, argument := range atom.Args {
			if !propertyRoles[argument.Role] || argument.Value == nil || argument.Value.Kind != algebra.TermRef {
				return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_ROLE_EVIDENCE_MISMATCH", "each CORE property role must bind one explicit reference"
			}
			matched := false
			for _, role := range candidate.RoleSources {
				if role.Role == argument.Role && role.ReferenceID == argument.Value.ID {
					matched = true
				}
			}
			if !matched {
				return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PROPERTY_ROLE_EVIDENCE_MISMATCH", "property argument references must match the retained source role evidence"
			}
		}
		if document.Expression.Kind == algebra.TermNot && document.Expression.Body != nil && document.Expression.Body.Kind == algebra.TermAtom {
			// Negation is licensed only for this explicitly approved property route.
		} else if document.Expression.Kind != algebra.TermAtom {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_PROPERTY_SCOPE_UNSUPPORTED", "property realization supports only a direct Atom or one Not-scoped Atom"
		}
	} else {
		if candidate.HeadEventReferenceID == "" || candidate.Source.SenseID == "" {
			return algebra.Document{}, nil, status.OutcomeIncomplete, "DSE_EVENT_SENSE_REQUIRED", "source candidate has no exact event sense and reference"
		}
		sourceReferences = append(sourceReferences, sourceReference{role: "event", referenceID: candidate.HeadEventReferenceID, senseID: candidate.Source.SenseID})
		bindingsByReference[candidate.HeadEventReferenceID] = candidate.Source.SenseID
	}
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
		if role.ReferenceID == "" {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_ROLE_REFERENCE_UNRESOLVED", "every realized role requires an exact source sense and explicit reference"
		}
		if role.Source.SenseID == "" {
			anchor, ok := dseLicensedImplicitParticipantAnchor(candidate, role, document)
			if !ok {
				return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_ROLE_REFERENCE_UNRESOLVED", "a role without an overt source NP requires exact validated implicit-role evidence and matching participant context"
			}
			participantAnchors[role.ReferenceID] = anchor
			continue
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
	if isAsk && (candidate.PredicateKind == "relation" || candidate.PredicateKind == "property" || !dseQuestionSourcesMatch(document.Expression.Body, candidate, answerID)) {
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
	for referenceID, anchor := range participantAnchors {
		index, found := referents[referenceID]
		if !found || context.Referents[index].Type.Kind != algebra.TypeEntity || context.Referents[index].ConceptID != "" && context.Referents[index].ConceptID != "concept:person" {
			return algebra.Document{}, nil, status.OutcomeInvalid, "DSE_PARTICIPANT_REFERENT_INVALID", "licensed implicit participant must be a compatible typed person referent"
		}
		context.Referents[index].ConceptID = "concept:person"
		anchored := false
		for _, condition := range context.Conditions {
			if condition.Kind == algebra.ContextPresupposition && condition.Code == "deictic_anchor" && len(condition.TargetReferentIDs) == 1 && condition.TargetReferentIDs[0] == referenceID && condition.Detail == anchor {
				anchored = true
			}
		}
		if !anchored {
			return algebra.Document{}, nil, status.OutcomeUnsupported, "DSE_PARTICIPANT_ANCHOR_MISSING", "implicit participant realization requires the matching validated participant anchor condition"
		}
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

func dseLicensedImplicitParticipantAnchor(candidate semantic.DSECandidate, role semantic.DSERoleSource, document algebra.Document) (string, bool) {
	evidence := role.ImplicitEvidence
	badProvenance := evidence == nil || role.ResolutionBasis != "licensed_implicit_role_evidence" || document.RequestContext == nil ||
		role.ReferenceID == "" || role.Source.CandidateID == "" || role.Source.CandidateID != candidate.Source.CandidateID ||
		role.Source.DerivationID != candidate.Source.DerivationID || role.Source.FrameID != candidate.Source.FrameID ||
		role.Source.FamilyID != candidate.Source.FamilyID || role.Source.VariantID != candidate.Source.VariantID ||
		evidence.PredicateCandidateID != candidate.Source.CandidateID || evidence.DerivationID != candidate.Source.DerivationID ||
		string(evidence.Role) != role.Role || evidence.CoreConceptID.String() != "concept:person" || evidence.SourceSpan != candidate.Source.Span ||
		evidence.Multiplicity == "" || evidence.ParticipantAnchor.Kind != "participant"
	if badProvenance {
		return "", false
	}
	anchor := evidence.ParticipantAnchor.Value
	request := document.RequestContext
	if anchor == "speaker" && request.SpeakerReferentID != role.ReferenceID ||
		anchor == "speaker_inclusive_group" && request.SpeakerReferentID != role.ReferenceID ||
		anchor == "addressee" && request.AddresseeReferentID != role.ReferenceID {
		return "", false
	}
	declared, contextual := false, false
	for _, reference := range document.References {
		if reference.ID == role.ReferenceID && reference.Type.Kind == algebra.TypeEntity &&
			(reference.Multiplicity == "" || reference.Multiplicity == string(evidence.Multiplicity)) {
			declared = true
		}
	}
	for _, referent := range request.Referents {
		if referent.ID == role.ReferenceID && referent.Type.Kind == algebra.TypeEntity &&
			(referent.ConceptID == "" || referent.ConceptID == "concept:person") && referent.Multiplicity == string(evidence.Multiplicity) {
			contextual = true
		}
	}
	if !declared || !contextual {
		return "", false
	}
	atom := document.Expression
	if atom != nil && atom.Kind == algebra.TermNot {
		atom = atom.Body
	}
	if atom == nil || atom.Kind != algebra.TermAtom {
		return "", false
	}
	matchedArgument := false
	for _, argument := range atom.Args {
		if argument.Role == role.Role && argument.Value != nil && argument.Value.Kind == algebra.TermRef && argument.Value.ID == role.ReferenceID {
			matchedArgument = true
		}
	}
	if !matchedArgument {
		return "", false
	}
	coveredAnchor := false
	for _, condition := range request.Conditions {
		if condition.Kind == algebra.ContextPresupposition && condition.Code == "deictic_anchor" && condition.Expression == nil &&
			condition.Detail == anchor && len(condition.TargetReferentIDs) == 1 && condition.TargetReferentIDs[0] == role.ReferenceID {
			for _, coverage := range candidate.ConditionCoverage {
				if coverage.ConditionID == condition.ID && string(coverage.Kind) == string(condition.Kind) && coverage.Code == condition.Code && len(coverage.TargetReferenceIDs) == 1 && coverage.TargetReferenceIDs[0] == role.ReferenceID {
					coveredAnchor = true
				}
			}
		}
	}
	return anchor, coveredAnchor
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
