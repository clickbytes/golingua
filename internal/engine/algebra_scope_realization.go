package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func (engine *Engine) realizeAlgebraScope(ctx context.Context, result AlgebraRealizationResult, input algebra.Document, languageID features.LanguageID, pack *language.SemanticLanguagePackage, validations []language.TargetValidation, signatures map[string]algebra.Signature, plan algebraScopePlan) (AlgebraRealizationResult, error) {
	retain := func(code, message string, outcome status.OutcomeKind) (AlgebraRealizationResult, error) {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: code, Message: message, Term: cloneAlgebraTerm(input.Expression)})
		return algebraRealizationFailure(result, outcome, code, "input.expression", message), nil
	}
	if len(input.Parameters) != 0 || len(input.References) != 0 || !emptyScopeContext(input.RequestContext) {
		return retain("SCOPED_REQUEST_NOT_CLOSED", "scope target realization requires a closed document with no parameters, references, or request context", status.OutcomeUnsupported)
	}
	if input.Closed != true {
		return retain("SCOPED_REQUEST_NOT_CLOSED", "scope target realization requires an explicitly closed algebra document", status.OutcomeUnsupported)
	}
	outer, inner, event := selectAlgebraScopeBindings(plan, pack, validations, signatures)
	for _, selection := range []algebraScopeSelection{outer, inner, event} {
		if selection.Status != status.OutcomeOK || selection.Match == nil {
			code := selection.Code
			if code == "" {
				code = "SCOPE_BINDING_UNAVAILABLE"
			}
			message := "one or more exact scope lexical selections are unavailable or ambiguous"
			return retain(code, message, selection.Status)
		}
	}
	target, ok := engine.TargetRealizerFor(languageID)
	scoped, supportsScope := target.(ScopedTargetRealizer)
	verifier, supportsVerification := target.(ScopedTargetVerifier)
	if !ok || target == nil || !supportsScope || !supportsVerification {
		return retain("SCOPED_TARGET_UNAVAILABLE", "target adapter does not provide the scoped realization contract", status.OutcomeUnsupported)
	}
	eventMatch := event.Match
	outerMatch, innerMatch := outer.Match, inner.Match
	negation := ScopedNegationOuter
	if plan.InnerNegated {
		negation = ScopedNegationInner
	}
	request := ScopedTargetRequest{
		CheckedExpression: cloneAlgebraTerm(input.Expression),
		Paths:             ScopedTargetPaths{OuterQuantifier: plan.OuterPath, Restrictor: plan.RestrictorPath, InnerExistential: plan.InnerPath, InnerPredicate: plan.PredicatePath, EventBinder: plan.EventPath, EventAtom: plan.EventAtomPath, Negation: plan.OuterNegPath},
		BinderX:           plan.OuterBinderID, BinderY: plan.InnerBinderID, BinderEvent: plan.EventBinderID,
		NegationLocation: negation, EventTemporalFrame: plan.EventFrame,
		Roles:          ScopedTargetRoles{XRole: eventMatch.OuterRole, XPort: plan.EventOuterPort, YRole: eventMatch.InnerRole, YPort: plan.EventInnerPort, EventPort: plan.EventPort},
		XSelection:     scopeTargetSelection(*outerMatch, "", plan.OuterPort),
		YSelection:     scopeTargetSelection(*innerMatch, "", plan.InnerPort),
		EventSelection: scopeTargetSelection(*eventMatch, "event_identity", plan.EventPort),
	}
	request.XSelection.Role, request.XSelection.Port = eventMatch.OuterRole, plan.EventOuterPort
	request.YSelection.Role, request.YSelection.Port = eventMatch.InnerRole, plan.EventInnerPort
	if plan.OuterNegated {
		request.Paths.Negation = plan.OuterNegPath
	} else {
		request.Paths.Negation = plan.InnerNegPath
	}
	budget := engineBudget(engine)
	requestCtx, cancel := context.WithTimeout(ctx, budget.RequestTimeout)
	defer cancel()
	surface, err := scoped.RealizeScope(requestCtx, request)
	if err != nil {
		return algebraRealizationError(result, err, "SCOPED_TARGET_FAILED", "target.scope")
	}
	return validateScopedAlgebraSurface(requestCtx, result, surface, request, input.RequestContext, budget, verifier, pack)
}

func scopeTargetSelection(match algebraScopeLexicalMatch, role, port string) ScopedTargetLexicalSelection {
	return ScopedTargetLexicalSelection{Binding: match.Binding, Sense: match.Sense, Lexeme: match.Lexeme, Validation: match.Validation, Role: role, Port: port}
}

func emptyScopeContext(context *algebra.RequestContext) bool {
	return context == nil || (len(context.Referents) == 0 && context.SpeakerReferentID == "" && context.AddresseeReferentID == "" && len(context.Mentions) == 0 && len(context.ResolutionEvidence) == 0 && len(context.Conditions) == 0)
}

func validateScopedAlgebraSurface(ctx context.Context, result AlgebraRealizationResult, surface ScopedTargetResult, request ScopedTargetRequest, requestContext *algebra.RequestContext, budget Budget, verifier ScopedTargetVerifier, pack *language.SemanticLanguagePackage) (AlgebraRealizationResult, error) {
	fail := func(outcome status.OutcomeKind, code, path, message string) (AlgebraRealizationResult, error) {
		result.Alternatives = nil
		result.ScopeCoverage = nil
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: code, Message: message, Term: cloneAlgebraTerm(request.CheckedExpression)})
		return algebraRealizationFailure(result, outcome, code, path, message), nil
	}
	if !surface.Status.Valid() || len(surface.Diagnostics) > status.MaxDiagnostics {
		return fail(status.OutcomeInvalid, "INVALID_SCOPED_TARGET_RESULT", "target.scope", "scoped target returned invalid status or excessive diagnostics")
	}
	for _, d := range surface.Diagnostics {
		if d.Validate() != nil {
			return fail(status.OutcomeInvalid, "INVALID_SCOPED_TARGET_DIAGNOSTIC", "target.scope.diagnostics", "scoped target returned an invalid diagnostic")
		}
	}
	if surface.Status != status.OutcomeOK && surface.Status != status.OutcomeIncomplete && surface.Status != status.OutcomeAmbiguous {
		result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
		return fail(surface.Status, "SCOPED_TARGET_DID_NOT_REALIZE", "target.scope", "scoped target status does not authorize emitting surface text")
	}
	if len(surface.Alternatives) > MaxScopedTargetAlternatives || len(surface.Alternatives) > budget.MaxSurfaceAlternatives {
		return fail(status.OutcomeResourceLimit, "SCOPED_ALTERNATIVE_LIMIT", "target.scope.alternatives", "scoped target exceeded the alternative limit")
	}
	if len(surface.Alternatives) == 0 {
		result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
		outcome := surface.Status
		if outcome == status.OutcomeOK {
			outcome = status.OutcomeIncomplete
		}
		return fail(outcome, "NO_SCOPED_ALTERNATIVE", "target.scope.alternatives", "scoped target returned no construction alternative")
	}
	if verifier == nil {
		return fail(status.OutcomeUnsupported, "SCOPED_VERIFIER_UNAVAILABLE", "target.scope", "target does not expose independent compiled-plan verification")
	}
	compiled, err := verifier.ScopedPlan()
	if err != nil {
		return fail(status.OutcomeInvalid, "SCOPED_PLAN_UNAVAILABLE", "target.scope.plan", "target scoped plan could not be retrieved")
	}
	planBytes, err := json.Marshal(compiled)
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_SCOPED_PLAN", "target.scope.plan", "target scoped plan could not be encoded")
	}
	validatedPlan, err := construction.DecodeScopedClausePlan(planBytes, features.LanguageID(pack.Language()))
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_SCOPED_PLAN", "target.scope.plan", "target scoped plan failed strict language and provenance validation")
	}
	subjectBinder, err := validateScopeFrame(request, pack)
	if err != nil {
		return fail(status.OutcomeInvalid, "SCOPED_FRAME_MISMATCH", "target.scope.frame", err.Error())
	}
	coverage := &AlgebraScopeCoverage{Paths: request.Paths, BinderX: request.BinderX, BinderY: request.BinderY, BinderEvent: request.BinderEvent, Negation: request.NegationLocation, TemporalFrame: request.EventTemporalFrame, Selections: []AlgebraScopeSelectionEvidence{
		scopeEvidence(request.BinderX, request.XSelection, request.CheckedExpression), scopeEvidence(request.BinderY, request.YSelection, request.CheckedExpression), scopeEvidence(request.BinderEvent, request.EventSelection, request.CheckedExpression),
	}, Alternatives: []AlgebraScopeAlternativeEvidence{}}
	coverage.Verified = true
	coverage.VerificationScope = "compiled-plan matching and fresh regeneration by the same adapter; not an independent linguistic proof"
	coverage.PlanProvenance = validatedPlan.Provenance.ID
	if requestContext != nil {
		coverage.RequestContext = &AlgebraScopeContextEvidence{SchemaVersion: requestContext.SchemaVersion, ID: requestContext.ID}
	}
	seen := map[string]bool{}
	var total int
	for i, alt := range surface.Alternatives {
		path := fmt.Sprintf("target.scope.alternatives[%d]", i)
		if !validID(alt.ID) || seen[alt.ID] || alt.ConstructionID.String() == "" || len(alt.Parts) == 0 || len(alt.Parts) > maxAlgebraRealizationParts {
			return fail(status.OutcomeInvalid, "INVALID_SCOPED_ALTERNATIVE", path, "scoped target returned invalid identity, construction, or parts")
		}
		seen[alt.ID] = true
		variant, found := scopedVariant(validatedPlan, alt, request, subjectBinder)
		if !found {
			return fail(status.OutcomeInvalid, "SCOPED_VARIANT_MISMATCH", path, "alternative does not identify exactly one matching scope construction variant")
		}
		text, err := validateScopedParts(alt.Parts, request)
		if err != nil {
			return fail(status.OutcomeInvalid, "INVALID_SCOPED_PARTS", path+".parts", err.Error())
		}
		if err := validateScopeVariantParts(variant, alt.Parts, request); err != nil {
			return fail(status.OutcomeInvalid, "SCOPED_PLAN_PART_MISMATCH", path+".parts", err.Error())
		}
		if err := verifier.VerifyScope(ctx, request, alt); err != nil {
			if ctx.Err() != nil {
				return fail(status.OutcomeResourceLimit, "REALIZATION_CANCELED", path, "scope verification was canceled or timed out")
			}
			result.Alternatives = nil
			result.ScopeCoverage = nil
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "SCOPED_TARGET_VERIFICATION_FAILED", Message: "scope was retained because compiled-plan checking and fresh morphology regeneration failed", Term: cloneAlgebraTerm(request.CheckedExpression)})
			return algebraRealizationError(result, err, "SCOPED_TARGET_VERIFICATION_FAILED", path)
		}
		if text == "" || validateText(text, MaxTranslationTextBytes, false) != nil {
			return fail(status.OutcomeInvalid, "INVALID_SCOPED_TEXT", path+".text", "scoped construction produced empty or invalid bounded text")
		}
		total += len(text)
		if total > budget.MaxCandidateTextBytes {
			return fail(status.OutcomeResourceLimit, "SCOPED_TEXT_LIMIT", "target.scope.alternatives", "scoped alternatives exceed the aggregate text limit")
		}
		id := alt.ID
		trace := semantic.EvidenceTrace{AlternativeID: id, Outcome: status.OutcomeOK, Records: []semantic.EvidenceRecord{}, Diagnostics: []status.Diagnostic{}}
		result.Alternatives = append(result.Alternatives, SurfaceAlternative{ID: id, Outcome: status.OutcomeOK, Text: text, Evidence: trace, Losses: []realizer.FieldLoss{}, Diagnostics: []status.Diagnostic{}})
		coverage.Alternatives = append(coverage.Alternatives, AlgebraScopeAlternativeEvidence{ID: id, ConstructionID: alt.ConstructionID.String(), Parts: append([]ScopedSurfacePart{}, alt.Parts...)})
	}
	result.ScopeCoverage = coverage
	result.Status = status.OutcomeOK
	if surface.Status == status.OutcomeIncomplete {
		result.Status = status.OutcomeIncomplete
	}
	if surface.Status == status.OutcomeAmbiguous || len(surface.Alternatives) > 1 {
		result.Status = status.OutcomeAmbiguous
	}
	result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
	return result, nil
}

func scopeEvidence(binder string, selection ScopedTargetLexicalSelection, expression *algebra.Term) AlgebraScopeSelectionEvidence {
	symbol := ""
	referencePort := ""
	if expression != nil { // the selected atom is identified by its binding target.
		if selection.Binding.Target != nil {
			symbol = string(selection.Binding.Target.SymbolRef)
			if selection.Binding.Target.ReferencePort != nil {
				referencePort = selection.Binding.Target.ReferencePort.PortID
			}
		}
	}
	return AlgebraScopeSelectionEvidence{BinderID: binder, Role: selection.Role, Port: selection.Port, ReferencePort: referencePort, SenseID: selection.Sense.ID.String(), LexemeID: selection.Lexeme.ID.String(), Symbol: symbol, ProvenanceID: selection.Binding.Provenance.ID}
}

func validateScopedParts(parts []ScopedSurfacePart, request ScopedTargetRequest) (string, error) {
	var joined strings.Builder
	seen := map[string]bool{}
	for i, part := range parts {
		if !utf8.ValidString(part.Text) || strings.TrimSpace(part.Text) != part.Text || len(part.Text) == 0 || len(part.Text) > MaxTranslationTextBytes {
			return "", fmt.Errorf("part %d has empty, untrimmed, invalid, or oversized text", i)
		}
		if part.Generated == nil {
			if part.BinderID != "" || part.Port != "" {
				return "", fmt.Errorf("literal part %d claims a binder or port", i)
			}
		} else {
			var expected ScopedTargetLexicalSelection
			switch part.BinderID {
			case request.BinderX:
				expected = request.XSelection
			case request.BinderY:
				expected = request.YSelection
			case request.BinderEvent:
				expected = request.EventSelection
			default:
				return "", fmt.Errorf("generated part %d names an unknown binder", i)
			}
			if part.Port != expected.Port || part.Generated.LexemeID != expected.Lexeme.ID.String() || string(part.Generated.Language) != string(expected.Lexeme.Language) || part.Generated.Surface != part.Text {
				return "", fmt.Errorf("generated part %d does not match its selected lexical identity", i)
			}
			seen[part.BinderID] = true
		}
		if joined.Len() != 0 && !isScopedPunctuation(part.Text) {
			joined.WriteByte(' ')
		}
		joined.WriteString(part.Text)
	}
	if !seen[request.BinderX] || !seen[request.BinderY] || !seen[request.BinderEvent] {
		return "", fmt.Errorf("construction omits one or more selected binder forms")
	}
	return joined.String(), nil
}

func isScopedPunctuation(text string) bool {
	return text == "." || text == "," || text == ";" || text == ":" || text == "?" || text == "!" || text == ")" || text == "]" || text == "}"
}
