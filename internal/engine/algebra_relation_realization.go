package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func (engine *Engine) realizeAlgebraRelationPath(ctx context.Context, result AlgebraRealizationResult, checked algebra.Document, languageID features.LanguageID, pack *language.SemanticLanguagePackage, selectedMetadata algebra.Document, plan algebraRelationPathPlan) (AlgebraRealizationResult, error) {
	return engine.realizeAlgebraRelationPathCore(ctx, result, checked, languageID, pack, selectedMetadata, plan, nil)
}

func (engine *Engine) realizeAlgebraEventRelationPath(ctx context.Context, result AlgebraRealizationResult, checked algebra.Document, languageID features.LanguageID, pack *language.SemanticLanguagePackage, selectedMetadata algebra.Document, eventPlan algebraRelationEventPathPlan) (AlgebraRealizationResult, error) {
	return engine.realizeAlgebraRelationPathCore(ctx, result, checked, languageID, pack, selectedMetadata, eventPlan.algebraRelationPathPlan, &eventPlan)
}

func (engine *Engine) realizeAlgebraRelationPathCore(ctx context.Context, result AlgebraRealizationResult, checked algebra.Document, languageID features.LanguageID, pack *language.SemanticLanguagePackage, selectedMetadata algebra.Document, plan algebraRelationPathPlan, eventPlan *algebraRelationEventPathPlan) (AlgebraRealizationResult, error) {
	result.Scope = "anonymous relation function path; structural IR, selected lexical metadata, compiled plan, generated forms and fresh target regeneration are checked; truth, lexical quality and event tense are not established"
	if !checked.Closed || len(checked.Parameters) != 0 || len(checked.References) != 0 || !relationPathContextEmpty(checked.RequestContext) {
		result.Residuals = appendRelationRequestResiduals(result.Residuals, checked)
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "RELATION_PATH_CONTEXT_NOT_CLOSED", Message: "relation path realization requires an explicitly closed document without parameters, references, or request-context conditions", Term: cloneAlgebraTerm(checked.Expression)})
		if trimAlgebraResiduals(&result) {
			return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESIDUAL_REPORT_LIMIT", "residuals", "request contains more residual items than this bounded result can report"), nil
		}
		return algebraRealizationFailure(result, status.OutcomeUnsupported, "RELATION_PATH_CONTEXT_NOT_CLOSED", "input", "request context was retained as residual and target generation was refused"), nil
	}
	var selection algebraRelationBindingResult
	if eventPlan == nil {
		selection = selectAlgebraRelationBinding(plan, pack, selectedMetadata, semantic.DSEBudget{MaxWork: maxAlgebraRealizationBindings})
	} else {
		selection = selectAlgebraEventRelationBinding(*eventPlan, pack, selectedMetadata, semantic.DSEBudget{MaxWork: maxAlgebraRealizationBindings})
	}
	if selection.Status != status.OutcomeOK {
		diagnostic := selection.Diagnostic
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: diagnostic.Path, Code: diagnostic.Code, Message: diagnostic.Message, Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, selection.Status, diagnostic.Code, diagnostic.Path, diagnostic.Message), nil
	}
	coverage := AlgebraRelationPathCoverage{
		Verified:                    false,
		VerificationScope:           "checked function shape, exact selected symbol/sense/lexeme/frame and role ports, strict relation plan, complete ordered parts and fresh target regeneration; no truth or linguistic-owner judgment",
		GrammaticalTenseInterpreted: false,
		Paths:                       RelationPathPaths{FirstLambda: plan.FirstPath, SecondLambda: plan.SecondPath, Witness: plan.WitnessPath, Conjunction: plan.ConjunctionPath, FirstAtom: plan.FirstAtomPath, SecondAtom: plan.SecondAtomPath},
		BinderX:                     plan.FirstBinder, BinderY: plan.SecondBinder, BinderZ: plan.WitnessBinder, Symbol: string(selection.Binding.Symbol),
		SenseID: selection.Binding.SenseID, LexemeID: selection.Binding.LexemeID, Lemma: selection.Binding.Lemma, POS: selection.Binding.POS,
		FrameID: string(selection.Binding.Frame.ID), FrameProvenance: selection.Binding.Frame.Provenance, BindingProvenance: selection.Binding.SemanticBinding.Provenance,
		TargetValidation: cloneRelationTargetValidation(selection.Binding.Validation), SubjectRole: selection.Binding.SubjectRole, ObjectRole: selection.Binding.ObjectRole,
		PortRoles: cloneRelationStringMap(selection.Binding.PortRoles), Alternatives: []AlgebraRelationPathAlternativeEvidence{},
	}
	if checked.RequestContext != nil {
		coverage.RequestContext = &AlgebraScopeContextEvidence{SchemaVersion: checked.RequestContext.SchemaVersion, ID: checked.RequestContext.ID}
	}
	if eventPlan != nil {
		coverage.Events = []AlgebraRelationEventCoverage{
			{BinderID: eventPlan.FirstEventBinder, EventPortID: eventPlan.EventPort, Path: eventPlan.FirstEventPath, AtomPath: eventPlan.FirstAtomPath, TemporalFrame: eventPlan.FirstEventTemporalFrame},
			{BinderID: eventPlan.SecondEventBinder, EventPortID: eventPlan.EventPort, Path: eventPlan.SecondEventPath, AtomPath: eventPlan.SecondAtomPath, TemporalFrame: eventPlan.SecondEventTemporalFrame},
		}
		result.RelationPathCoverage = &coverage
		missingTime := false
		unsupportedTime := false
		for _, event := range coverage.Events {
			switch event.TemporalFrame {
			case "present":
			case "":
				missingTime = true
				result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: event.Path + ".temporal_frame", Code: "EVENT_TEMPORAL_FRAME_UNSPECIFIED", Message: "event has no explicit temporal frame; present-tense target generation cannot establish event time", Term: cloneAlgebraTerm(eventTermAtPath(*eventPlan, event.BinderID))})
			default:
				unsupportedTime = true
				result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: event.Path + ".temporal_frame", Code: "EVENT_TEMPORAL_FRAME_UNSUPPORTED", Message: "explicit event temporal frame is not supported by the compiled present-tense plan", Term: cloneAlgebraTerm(eventTermAtPath(*eventPlan, event.BinderID))})
			}
		}
		if unsupportedTime {
			return algebraRealizationFailure(result, status.OutcomeUnsupported, "EVENT_TEMPORAL_FRAME_UNSUPPORTED", "input.expression", "explicit event time is unsupported; no target text was generated"), nil
		}
		if missingTime {
			return algebraRealizationFailure(result, status.OutcomeIncomplete, "EVENT_TEMPORAL_FRAME_UNSPECIFIED", "input.expression", "event time is unspecified; no target text was generated"), nil
		}
	} else {
		result.RelationPathCoverage = &coverage
	}
	target, ok := engine.TargetRealizerFor(languageID)
	relationTarget, supportsRelation := target.(RelationPathTargetRealizer)
	verifier, supportsVerification := target.(RelationPathTargetVerifier)
	if !ok || target == nil || !supportsRelation || !supportsVerification {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "RELATION_PATH_TARGET_UNAVAILABLE", Message: "target adapter does not provide relation path realization and fresh verification", Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, status.OutcomeUnsupported, "RELATION_PATH_TARGET_UNAVAILABLE", "target", "relation path target contract is unavailable"), nil
	}
	request := RelationPathRequest{
		Checked:         checked,
		FirstLambdaPath: plan.FirstPath, SecondLambdaPath: plan.SecondPath,
		WitnessPath: plan.WitnessPath, ConjunctionPath: plan.ConjunctionPath,
		FirstAtomPath: plan.FirstAtomPath, SecondAtomPath: plan.SecondAtomPath,
		BinderX: plan.FirstBinder, BinderY: plan.SecondBinder, BinderZ: plan.WitnessBinder,
		Symbol: plan.Symbol, OrderedPorts: plan.FirstPorts,
		Package: pack, Selected: selectedMetadata,
		Budget: semantic.DSEBudget{MaxWork: maxAlgebraRealizationBindings},
		Selection: RelationPathSelection{
			Binding: selection.Binding.SemanticBinding, Sense: selection.Binding.LexicalSense,
			Lexeme: selection.Binding.Lexeme, Frame: selection.Binding.Frame,
			Validation: selection.Binding.Validation, SubjectRole: selection.Binding.SubjectRole,
			ObjectRole: selection.Binding.ObjectRole, PortRoles: cloneRelationStringMap(selection.Binding.PortRoles),
			PortIDs: cloneRelationStringMap(selection.Binding.PortIDs),
		},
	}
	if eventPlan != nil {
		request.EventPair = &RelationPathEventPairEvidence{
			EventPort:        eventPlan.EventPort,
			FirstEventBinder: eventPlan.FirstEventBinder, FirstEventPath: eventPlan.FirstEventPath, FirstTemporalFrame: eventPlan.FirstEventTemporalFrame,
			SecondEventBinder: eventPlan.SecondEventBinder, SecondEventPath: eventPlan.SecondEventPath, SecondTemporalFrame: eventPlan.SecondEventTemporalFrame,
		}
	}
	planCopy, err := verifier.RelationPathPlan()
	if err != nil {
		return algebraRealizationError(result, err, "RELATION_PATH_PLAN_UNAVAILABLE", "target.relation_path.plan")
	}
	planJSON, err := json.Marshal(planCopy)
	if err != nil || len(planJSON) > MaxRelationPlanBytes {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path.plan", Code: "RELATION_PATH_PLAN_INVALID", Message: "relation path plan exceeds its bounded representation", Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "target.relation_path.plan", "relation path plan is invalid or oversized"), nil
	}
	validatedPlan, err := DecodeRelationPathPlan(planJSON, string(languageID))
	if err != nil {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path.plan", Code: "RELATION_PATH_PLAN_INVALID", Message: err.Error(), Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "target.relation_path.plan", "relation path plan failed strict decoding"), nil
	}
	budget := engineBudget(engine)
	requestCtx, cancel := context.WithTimeout(ctx, budget.RequestTimeout)
	defer cancel()
	surface, err := relationTarget.RealizeRelationPath(requestCtx, request)
	if err != nil {
		return algebraRealizationError(result, err, "RELATION_PATH_TARGET_FAILED", "target.relation_path")
	}
	if err := ValidateRelationPathResult(request, validatedPlan, surface); err != nil {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path", Code: "INVALID_RELATION_PATH_RESULT", Message: err.Error(), Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_RELATION_PATH_RESULT", "target.relation_path", "target result failed bounded structural validation"), nil
	}
	for _, diagnostic := range surface.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.diagnostics", Code: "INVALID_RELATION_PATH_DIAGNOSTIC", Message: err.Error(), Term: cloneAlgebraTerm(checked.Expression)})
			return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_RELATION_PATH_DIAGNOSTIC", "target.diagnostics", "target returned malformed diagnostics"), nil
		}
	}
	if surface.Status != status.OutcomeOK && surface.Status != status.OutcomeAmbiguous && surface.Status != status.OutcomeIncomplete {
		result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path", Code: "RELATION_PATH_TARGET_REFUSED", Message: "target did not establish a complete licensed alternative", Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, surface.Status, "RELATION_PATH_TARGET_REFUSED", "target.relation_path", "target did not establish a complete licensed alternative"), nil
	}
	if len(surface.Alternatives) == 0 {
		result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path", Code: "RELATION_PATH_COVERAGE_INCOMPLETE", Message: "no complete surface alternative covers the checked relation path", Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, status.OutcomeIncomplete, "RELATION_PATH_COVERAGE_INCOMPLETE", "target.relation_path", "no complete surface alternative was produced"), nil
	}
	if len(surface.Alternatives) > budget.MaxSurfaceAlternatives || len(surface.Alternatives) > MaxTranslationAlternatives {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path", Code: "RELATION_PATH_ALTERNATIVE_LIMIT", Message: "all alternatives were retained as residual input because the target exceeded engine limits", Term: cloneAlgebraTerm(checked.Expression)})
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RELATION_PATH_ALTERNATIVE_LIMIT", "target.alternatives", "relation path result exceeds the engine alternative limit"), nil
	}
	coverage.PlanProvenance = validatedPlan.Provenance
	textBytes := 0
	for _, alternative := range surface.Alternatives {
		if err := requestCtx.Err(); err != nil {
			return algebraRealizationFailure(result, status.OutcomeResourceLimit, "REALIZATION_CANCELED", "context", "relation path verification was canceled"), nil
		}
		if err := ValidateRelationPathAlternative(request, validatedPlan, alternative); err != nil {
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.alternatives", Code: "INVALID_RELATION_PATH_ALTERNATIVE", Message: err.Error(), Term: cloneAlgebraTerm(checked.Expression)})
			return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_RELATION_PATH_ALTERNATIVE", "target.alternatives", "target alternative failed plan, binder, role, port, or morphology validation"), nil
		}
		if err := verifier.VerifyRelationPath(requestCtx, request, alternative); err != nil {
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.alternatives", Code: "RELATION_PATH_FRESH_VERIFICATION_FAILED", Message: err.Error(), Term: cloneAlgebraTerm(checked.Expression)})
			return algebraRealizationError(result, err, "RELATION_PATH_FRESH_VERIFICATION_FAILED", "target.alternatives")
		}
		text, err := relationPathSurfaceText(alternative.Parts)
		if err != nil || validateText(text, MaxTranslationTextBytes, false) != nil {
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.alternatives", Code: "INVALID_RELATION_PATH_SURFACE", Message: "ordered target parts do not form bounded nonempty surface text", Term: cloneAlgebraTerm(checked.Expression)})
			return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_RELATION_PATH_SURFACE", "target.alternatives", "target surface text is invalid or oversized"), nil
		}
		textBytes += len(text)
		if textBytes > budget.MaxCandidateTextBytes {
			return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "target.text", "relation path alternatives exceed the candidate text limit"), nil
		}
		result.Alternatives = append(result.Alternatives, SurfaceAlternative{ID: alternative.ID, Outcome: status.OutcomeOK, Text: text, Evidence: semantic.EvidenceTrace{AlternativeID: alternative.ID, Outcome: status.OutcomeOK}})
		coverage.Alternatives = append(coverage.Alternatives, AlgebraRelationPathAlternativeEvidence{ID: alternative.ID, VariantID: alternative.VariantID, Parts: append([]RelationPathSurfacePart(nil), alternative.Parts...)})
	}
	coverage.Verified = true
	result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
	result.Status = surface.Status
	if surface.Status == status.OutcomeIncomplete {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "target.relation_path", Code: "RELATION_PATH_PARTIAL_COVERAGE", Message: "some licensed plan variants or forms remain unresolved; emitted alternatives are individually complete", Term: cloneAlgebraTerm(checked.Expression)})
	}
	return result, nil
}

func relationPathContextEmpty(request *algebra.RequestContext) bool {
	return request == nil || len(request.Referents) == 0 && request.SpeakerReferentID == "" && request.AddresseeReferentID == "" && len(request.Mentions) == 0 && len(request.ResolutionEvidence) == 0 && len(request.Conditions) == 0
}

func eventTermAtPath(plan algebraRelationEventPathPlan, binder string) *algebra.Term {
	if binder == plan.FirstEventBinder {
		return plan.FirstEvent
	}
	if binder == plan.SecondEventBinder {
		return plan.SecondEvent
	}
	return nil
}

func relationPathSurfaceText(parts []RelationPathSurfacePart) (string, error) {
	if len(parts) == 0 || len(parts) > maxAlgebraRealizationParts {
		return "", fmt.Errorf("invalid relation path part count")
	}
	var joined strings.Builder
	for _, part := range parts {
		if part.Text == "" {
			return "", fmt.Errorf("empty relation path part")
		}
		if joined.Len() != 0 && !isScopedPunctuation(part.Text) {
			joined.WriteByte(' ')
		}
		joined.WriteString(part.Text)
	}
	return joined.String(), nil
}

func appendRelationRequestResiduals(residuals []AlgebraRealizationResidual, document algebra.Document) []AlgebraRealizationResidual {
	for i, declaration := range document.Parameters {
		copyDeclaration := declaration
		copyDeclaration.Type = cloneAlgebraType(declaration.Type)
		residuals = append(residuals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.parameters[%d]", i), Code: "UNSUPPORTED_PARAMETER", Message: "parameter is retained and cannot be discarded by this closed relation route", Declaration: &copyDeclaration})
	}
	for i, declaration := range document.References {
		copyDeclaration := declaration
		copyDeclaration.Type = cloneAlgebraType(declaration.Type)
		residuals = append(residuals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.references[%d]", i), Code: "UNSUPPORTED_REFERENCE", Message: "reference is retained and cannot be discarded by this closed relation route", ReferenceID: declaration.ID, Declaration: &copyDeclaration})
	}
	if request := document.RequestContext; request != nil {
		if request.ID != "" || request.SchemaVersion != 0 {
			residuals = append(residuals, AlgebraRealizationResidual{Path: "input.request_context", Code: "UNSUPPORTED_REQUEST_CONTEXT", Message: "request context metadata is retained as a residual", ContextID: request.ID})
		}
		if request.SpeakerReferentID != "" {
			residuals = append(residuals, AlgebraRealizationResidual{Path: "input.request_context.speaker_referent_id", Code: "UNSUPPORTED_SPEAKER_REFERENT", Message: "speaker identity is not consumed", ContextID: request.ID, ReferenceID: request.SpeakerReferentID})
		}
		if request.AddresseeReferentID != "" {
			residuals = append(residuals, AlgebraRealizationResidual{Path: "input.request_context.addressee_referent_id", Code: "UNSUPPORTED_ADDRESSEE_REFERENT", Message: "addressee identity is not consumed", ContextID: request.ID, ReferenceID: request.AddresseeReferentID})
		}
		for i, referent := range request.Referents {
			copyReferent := referent
			copyReferent.Type = cloneAlgebraType(referent.Type)
			residuals = append(residuals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.referents[%d]", i), Code: "UNSUPPORTED_CONTEXT_REFERENT", Message: "context referent is retained and not mapped to this function", ContextID: request.ID, ContextReferent: &copyReferent})
		}
		for i, mention := range request.Mentions {
			copyMention := mention
			residuals = append(residuals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.mentions[%d]", i), Code: "UNSUPPORTED_CONTEXT_MENTION", Message: "context mention is retained and not resolved by this route", ContextID: request.ID, ContextMention: &copyMention})
		}
		for i, evidence := range request.ResolutionEvidence {
			copyEvidence := evidence
			residuals = append(residuals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.resolution_evidence[%d]", i), Code: "UNSUPPORTED_RESOLUTION_EVIDENCE", Message: "resolution evidence is retained but not consumed", ContextID: request.ID, ResolutionEvidence: &copyEvidence})
		}
		for i, condition := range request.Conditions {
			copyCondition := condition
			copyCondition.TargetReferentIDs = append([]string(nil), condition.TargetReferentIDs...)
			copyCondition.Expression = cloneAlgebraTerm(condition.Expression)
			residuals = append(residuals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.conditions[%d]", i), Code: "UNSUPPORTED_CONTEXT_CONDITION", Message: "context condition is retained and not interpreted", ContextID: request.ID, ContextCondition: &copyCondition})
		}
	}
	return residuals
}

func cloneRelationStringMap(source map[string]string) map[string]string {
	copy := make(map[string]string, len(source))
	for key, value := range source {
		copy[key] = value
	}
	return copy
}

func cloneRelationTargetValidation(source language.TargetValidation) language.TargetValidation {
	copy := source
	copy.Unresolved = append(source.Unresolved[:0:0], source.Unresolved...)
	return copy
}
