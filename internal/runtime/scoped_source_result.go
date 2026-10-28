package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

const scopedVerificationScope = "algebra.Check validates generated type, binder scope, declarations, and request context only; it does not establish truth or linguistic quality"

// ComposeScopedSourceMatch converts each complete scoped alternative into one
// DSE candidate while retaining its lexical and construction evidence.
func ComposeScopedSourceMatch(ctx context.Context, match ScopedSourceMatch, plan construction.ScopedClausePlan, options DSEOptions) (semantic.DSEResult, error) {
	budget, err := semantic.NormalizeDSEBudget(options.Budget)
	if err != nil {
		return semantic.DSEResult{}, err
	}
	result := semantic.DSEResult{Status: status.OutcomeIncomplete, Complete: false, Candidates: []semantic.DSECandidate{}, Residuals: []semantic.DSEResidual{}, Diagnostics: []status.Diagnostic{}, WholeVerificationScope: scopedVerificationScope, Work: match.Work}
	if ctx == nil || options.Package == nil || options.Plan == nil {
		result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("SCOPED_SOURCE_INPUT_REQUIRED", "context, semantic package, and source plan are required"))
		result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "SCOPED_SOURCE_INPUT_REQUIRED", Message: "context, semantic package, and source plan are required"})
		return result, nil
	}
	if result.Work < 0 || result.Work > budget.MaxWork {
		return scopedDSELimit(result, "work", result.Work, budget.MaxWork), nil
	}
	if err := ctx.Err(); err != nil {
		return scopedDSELimit(result, "canceled", 1, 0), nil
	}
	if match.Status == status.OutcomeResourceLimit {
		result.Status = status.OutcomeResourceLimit
		result.Complete = false
		result.Diagnostics = append(result.Diagnostics, match.Diagnostics...)
		result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "SCOPED_SOURCE_RESOURCE_LIMIT", Message: "source matching exhausted its finite budget"})
		return result, nil
	}
	if options.Selected.Expression != nil || len(options.Selected.References) != 0 || options.Selected.RequestContext != nil {
		result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("UNSUPPORTED_SELECTED_CONTEXT", "selected primitive context must be declaration-only; request declarations and context belong in options.References"))
		result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "UNSUPPORTED_SELECTED_CONTEXT", Message: "selected primitive context contains an expression, references, or request context that cannot be merged losslessly"})
		return result, nil
	}
	if match.Work < 0 || match.Status != status.OutcomeOK && match.Status != status.OutcomeAmbiguous {
		result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("SCOPED_SOURCE_MATCH_INCOMPLETE", "source matcher did not provide a complete set of alternatives"))
		result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "SCOPED_SOURCE_MATCH_INCOMPLETE", Message: "source match status is not complete"})
		for _, diagnostic := range match.Diagnostics {
			result.Diagnostics = append(result.Diagnostics, diagnostic)
			result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: diagnostic.Code, Path: diagnostic.Path, Message: diagnostic.Message})
		}
	}
	if len(match.Alternatives) > budget.MaxAlternatives {
		return scopedDSELimit(result, "alternatives", len(match.Alternatives), budget.MaxAlternatives), nil
	}
	for i, alternative := range match.Alternatives {
		if err := ctx.Err(); err != nil {
			return scopedDSELimit(result, "canceled", 1, 0), nil
		}
		if result.Work >= budget.MaxWork {
			return scopedDSELimit(result, "work", result.Work+1, budget.MaxWork), nil
		}
		result.Work++
		remaining := semantic.DSEBudget{MaxAlternatives: budget.MaxAlternatives, MaxWork: budget.MaxWork - result.Work, MaxOutputBytes: budget.MaxOutputBytes - result.OutputBytes}
		if remaining.MaxWork < 1 || remaining.MaxOutputBytes < 1 {
			return scopedDSELimit(result, "work_or_output", result.Work+1, budget.MaxWork), nil
		}
		selection := selectScopedSourceMetadata(alternative, plan, options.Plan, options.Package, options.Selected, remaining)
		result.Work += selection.Work
		candidate := scopedSourceCandidate(i, alternative, selection, plan, options.References)
		if selection.Status != status.OutcomeOK {
			candidate.Status = selection.Status
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: selection.Diagnostic.Code, Path: selection.Diagnostic.Path, Message: selection.Diagnostic.Message})
			result.Diagnostics = append(result.Diagnostics, selection.Diagnostic)
			result.Residuals = append(result.Residuals, candidate.Residuals...)
			if ok, appendErr := appendScopedCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return scopedDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			if selection.Status == status.OutcomeResourceLimit {
				result.Status = status.OutcomeResourceLimit
				result.Complete = false
				return result, nil
			}
			continue
		}
		selected, mergeErr := scopedSelectedWithReferences(options.Selected, options.References)
		if mergeErr != nil {
			candidate.Status = status.OutcomeUnsupported
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: "REQUEST_CONTEXT_UNSUPPORTED", Message: mergeErr.Error()})
			result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("REQUEST_CONTEXT_UNSUPPORTED", mergeErr.Error()))
			result.Residuals = append(result.Residuals, candidate.Residuals...)
			if ok, appendErr := appendScopedCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return scopedDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			continue
		}
		composed := composeScopedSource(selection, selected)
		if composed.Status != status.OutcomeOK {
			candidate.Status = composed.Status
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: composed.Diagnostic.Code, Path: composed.Diagnostic.Path, Message: composed.Diagnostic.Message})
			result.Diagnostics = append(result.Diagnostics, composed.Diagnostic)
			result.Residuals = append(result.Residuals, candidate.Residuals...)
			if ok, appendErr := appendScopedCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return scopedDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			continue
		}
		checked, checkErr := algebra.Check(composed.Document)
		if checkErr != nil {
			candidate.Status = status.OutcomeInvalid
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: "SCOPED_EXPRESSION_CHECK_FAILED", Path: "expression", Message: checkErr.Error()})
			result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("SCOPED_EXPRESSION_CHECK_FAILED", checkErr.Error()))
			result.Residuals = append(result.Residuals, candidate.Residuals...)
			if ok, appendErr := appendScopedCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return scopedDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			continue
		}
		encoded, encodeErr := algebra.Encode(composed.Document)
		if encodeErr != nil {
			candidate.Status = status.OutcomeInvalid
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: "SCOPED_EXPRESSION_ENCODE_FAILED", Path: "document", Message: encodeErr.Error()})
			result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("SCOPED_EXPRESSION_ENCODE_FAILED", encodeErr.Error()))
			result.Residuals = append(result.Residuals, candidate.Residuals...)
			if ok, appendErr := appendScopedCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return scopedDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			continue
		}
		candidate.Status = status.OutcomeOK
		candidate.Checked = &checked
		candidate.CanonicalDocument = append(json.RawMessage(nil), encoded...)
		if ok, appendErr := appendScopedCandidate(&result, candidate, budget); appendErr != nil {
			return result, appendErr
		} else if !ok {
			return scopedDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
		}
	}
	if len(result.Candidates) == 0 {
		result.Status = status.OutcomeIncomplete
		result.Complete = false
		return result, nil
	}
	allOK := len(result.Residuals) == 0
	for _, candidate := range result.Candidates {
		if candidate.Status != status.OutcomeOK {
			allOK = false
		}
	}
	result.Complete = match.Complete && allOK
	switch {
	case result.Work >= budget.MaxWork && !result.Complete:
		result.Status = status.OutcomeResourceLimit
	case !allOK || !match.Complete:
		result.Status = status.OutcomeIncomplete
	case len(result.Candidates) > 1 || match.Status == status.OutcomeAmbiguous:
		result.Status = status.OutcomeAmbiguous
	default:
		result.Status = status.OutcomeOK
	}
	return result, nil
}

func scopedSelectedWithReferences(selected algebra.Document, refs language.SourceReferenceContext) (algebra.Document, error) {
	if len(refs.RoleReferences) != 0 || len(refs.PortReferences) != 0 {
		return algebra.Document{}, fmt.Errorf("scoped binders cannot consume explicit role or port reference mappings")
	}
	if refs.ContextID != "" && refs.RequestContext == nil {
		return algebra.Document{}, fmt.Errorf("context ID requires a corresponding request context")
	}
	if refs.RequestContext != nil && (refs.ContextID == "" || refs.RequestContext.ID != refs.ContextID) {
		return algebra.Document{}, fmt.Errorf("source reference context ID must match request_context.id")
	}
	copy, err := json.Marshal(selected)
	if err != nil || len(copy) > algebra.MaxDocumentBytes {
		return algebra.Document{}, fmt.Errorf("selected declarations exceed the algebra document limit")
	}
	var out algebra.Document
	if err := json.Unmarshal(copy, &out); err != nil {
		return algebra.Document{}, fmt.Errorf("selected declarations cannot be copied")
	}
	out.References = append([]algebra.Declaration(nil), refs.References...)
	if refs.RequestContext != nil {
		contextCopy, err := algebra.CloneRequestContext(*refs.RequestContext)
		if err != nil {
			return algebra.Document{}, fmt.Errorf("request context is invalid: %w", err)
		}
		out.RequestContext = &contextCopy
	}
	return out, nil
}

func scopedSourceCandidate(index int, alternative ScopedSourceAlternative, selection scopedSourceSelection, plan construction.ScopedClausePlan, refs language.SourceReferenceContext) semantic.DSECandidate {
	source := semantic.DSESourceRecord{VariantID: alternative.VariantID, ReadingIDs: []string{}}
	var event *ScopedSourceBinding
	bindings := make([]semantic.DSEScopedBindingEvidence, 0, len(alternative.Bindings))
	for i := range alternative.Bindings {
		b := alternative.Bindings[i]
		bindings = append(bindings, semantic.DSEScopedBindingEvidence{Binder: b.Binder, Case: b.Case, TokenIndex: b.TokenIndex, Span: b.Span, Reading: b.Reading, Lexeme: b.Lexeme, Sense: b.Sense})
		if b.Binder == "event" {
			copy := b
			event = &copy
		}
	}
	if event != nil {
		source.SenseID = event.Sense.ID.String()
		source.ReadingID = event.Reading.ID
		source.ReadingIDs = []string{event.Reading.ID}
		source.Span = event.Span
	}
	source.FrameID = string(selection.Event.Frame.ID)
	roleEvidence := make([]semantic.DSEScopedRoleEvidence, 0, len(selection.Event.RolePorts))
	for role, port := range selection.Event.RolePorts {
		roleEvidence = append(roleEvidence, semantic.DSEScopedRoleEvidence{FrameID: string(selection.Event.Frame.ID), Role: role, PortID: port.Role, Type: port.Type})
	}
	sort.Slice(roleEvidence, func(i, j int) bool { return roleEvidence[i].Role < roleEvidence[j].Role })
	variant := selection.Variant
	if variant.ID == "" {
		for _, candidate := range plan.Variants {
			if candidate.ID == alternative.VariantID {
				variant = candidate
				break
			}
		}
	}
	proof := append([]provenance.SourceRef(nil), selection.Provenance...)
	proof = append(proof, plan.Provenance)
	for _, binding := range alternative.Bindings {
		if binding.Reading.Provenance != nil {
			proof = append(proof, *binding.Reading.Provenance)
		}
	}
	evidence := &semantic.DSEScopedEvidence{Variant: variant, Bindings: bindings, Frame: selection.Event.Frame, SourceReferences: refs.Clone(), SubjectBinder: selection.Event.SubjectBinder, ObjectBinder: selection.Event.ObjectBinder, SubjectRole: selection.Event.SubjectRole, ObjectRole: selection.Event.ObjectRole, RolePorts: roleEvidence, Provenance: proof}
	return semantic.DSECandidate{ID: fmt.Sprintf("scoped:%d", index), Status: status.OutcomeIncomplete, Source: source, RoleSources: []semantic.DSERoleSource{}, GoalSpecs: []json.RawMessage{}, FragmentChecks: []semantic.DSEFragmentCheck{}, ConditionCoverage: []semantic.DSEConditionCoverage{}, Residuals: []semantic.DSEResidual{}, Scope: scopedVerificationScope, ScopedEvidence: evidence}
}

func scopedDiagnostic(code, message string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Severity: status.SeverityError}
}
func appendScopedCandidate(result *semantic.DSEResult, candidate semantic.DSECandidate, budget semantic.DSEBudget) (bool, error) {
	encoded, err := json.Marshal(candidate)
	if err != nil {
		return false, err
	}
	if len(encoded) > budget.MaxOutputBytes-result.OutputBytes {
		return false, nil
	}
	result.OutputBytes += len(encoded)
	result.Candidates = append(result.Candidates, candidate)
	return true, nil
}
func scopedDSELimit(result semantic.DSEResult, resource string, actual, limit int) semantic.DSEResult {
	result.Status = status.OutcomeResourceLimit
	result.Complete = false
	message := fmt.Sprintf("scoped source %s limit exceeded: %d > %d", resource, actual, limit)
	result.Diagnostics = append(result.Diagnostics, scopedDiagnostic("RESOURCE_LIMIT", message))
	result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "RESOURCE_LIMIT", Path: "budget." + resource, Message: message})
	return result
}
