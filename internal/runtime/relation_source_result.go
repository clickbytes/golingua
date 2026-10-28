package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

const relationSourceVerificationScope = "algebra.Check verifies generated function type, binder scope, declarations, and request context only; it does not establish world truth or linguistic quality"

// ComposeRelationSourceMatch selects and composes every retained relation
// source alternative while enforcing one global DSE work and output budget.
func ComposeRelationSourceMatch(ctx context.Context, match RelationSourceMatch, plan engine.RelationPathPlan, options DSEOptions) (semantic.DSEResult, error) {
	budget, err := semantic.NormalizeDSEBudget(options.Budget)
	if err != nil {
		return semantic.DSEResult{}, err
	}
	result := semantic.DSEResult{
		Status: status.OutcomeIncomplete, Complete: false,
		Candidates: []semantic.DSECandidate{}, Residuals: []semantic.DSEResidual{}, Diagnostics: []status.Diagnostic{},
		WholeVerificationScope: relationSourceVerificationScope, Work: match.Work,
	}
	if ctx == nil {
		return relationDSEFailure(result, status.OutcomeInvalid, "RELATION_SOURCE_CONTEXT_REQUIRED", "context", "relation source composition requires a context"), nil
	}
	if result.Work < 0 || result.Work > budget.MaxWork {
		return relationDSELimit(result, "work", result.Work, budget.MaxWork), nil
	}
	if err := ctx.Err(); err != nil {
		return relationDSELimit(result, "canceled", 1, 0), nil
	}
	if match.Status == status.OutcomeResourceLimit {
		result.Status = status.OutcomeResourceLimit
		result.Complete = false
		appendRelationDSEMatchDiagnostics(&result, match.Diagnostics)
		result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "RELATION_SOURCE_MATCH_RESOURCE_LIMIT", Message: "relation source matching exhausted its finite budget"})
		return result, nil
	}
	if match.Work < 0 || !match.Complete || match.Status != status.OutcomeOK && match.Status != status.OutcomeAmbiguous {
		result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "RELATION_SOURCE_MATCH_INCOMPLETE", Message: "source matcher did not establish complete coverage of all licensed variants and readings"})
		appendRelationDSEMatchDiagnostics(&result, match.Diagnostics)
	}
	if len(match.Alternatives) == 0 {
		result.Status, result.Complete = status.OutcomeIncomplete, false
		if len(result.Diagnostics) == 0 {
			appendRelationDSEDiagnostic(&result, relationDSEDiagnostic("RELATION_SOURCE_NO_ALTERNATIVES", "source matcher retained no complete relation alternative"))
		}
		if len(result.Residuals) == 0 {
			result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "RELATION_SOURCE_NO_ALTERNATIVES", Message: "no complete source alternative is available for metadata selection"})
		}
		return result, nil
	}
	if len(match.Alternatives) > budget.MaxAlternatives {
		return relationDSELimit(result, "alternatives", len(match.Alternatives), budget.MaxAlternatives), nil
	}
	for index, alternative := range match.Alternatives {
		if err := ctx.Err(); err != nil {
			return relationDSELimit(result, "canceled", 1, 0), nil
		}
		baseWork := result.Work
		if baseWork >= budget.MaxWork {
			return relationDSELimit(result, "work", baseWork+1, budget.MaxWork), nil
		}
		result.Work++
		remainingWork := budget.MaxWork - result.Work
		remainingOutput := budget.MaxOutputBytes - result.OutputBytes
		if remainingWork < 1 || remainingOutput < 1 {
			return relationDSELimit(result, "work_or_output", result.Work+1, budget.MaxWork), nil
		}
		candidateOptions := options
		candidateOptions.Budget = semantic.DSEBudget{MaxAlternatives: budget.MaxAlternatives - len(result.Candidates), MaxWork: remainingWork, MaxOutputBytes: remainingOutput}
		selection := selectRelationSourceMetadata(ctx, alternative, plan, candidateOptions)
		if selection.Work < 0 || selection.Work > remainingWork {
			result.Work = budget.MaxWork + 1
			candidate := relationSourceCandidate(index, alternative, plan, selection, nil, options)
			candidate.Status = status.OutcomeResourceLimit
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: "RELATION_SOURCE_RESOURCE_LIMIT", Path: "budget.work", Message: "metadata selection exceeded the remaining global work budget"})
			if ok, appendErr := appendRelationSourceCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return relationDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			return relationDSELimit(result, "work", result.Work, budget.MaxWork), nil
		}
		result.Work += selection.Work
		if selection.Status != status.OutcomeOK {
			candidate := relationSourceCandidate(index, alternative, plan, selection, nil, options)
			candidate.Status = selection.Status
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: selection.Diagnostic.Code, Path: selection.Diagnostic.Path, Message: selection.Diagnostic.Message})
			appendRelationDSEDiagnostic(&result, selection.Diagnostic)
			result.Residuals = append(result.Residuals, candidate.Residuals...)
			if ok, appendErr := appendRelationSourceCandidate(&result, candidate, budget); appendErr != nil {
				return result, appendErr
			} else if !ok {
				return relationDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
			}
			if selection.Status == status.OutcomeResourceLimit {
				result.Status, result.Complete = status.OutcomeResourceLimit, false
				return result, nil
			}
			continue
		}
		composeOptions := options
		composeOptions.Budget = semantic.DSEBudget{MaxAlternatives: budget.MaxAlternatives - len(result.Candidates), MaxWork: remainingWork, MaxOutputBytes: remainingOutput}
		composition := composeRelationSource(ctx, selection, composeOptions)
		if composition.Work < selection.Work || composition.Work > remainingWork {
			result.Work = budget.MaxWork + 1
		} else {
			result.Work = baseWork + 1 + composition.Work
		}
		candidate := relationSourceCandidate(index, alternative, plan, selection, &composition, options)
		candidate.Status = composition.Status
		if composition.Status == status.OutcomeOK {
			candidate.Checked = &composition.Checked
			candidate.CanonicalDocument = append(json.RawMessage(nil), composition.Encoded...)
		} else if composition.Diagnostic.Code != "" {
			candidate.Residuals = append(candidate.Residuals, semantic.DSEResidual{Code: composition.Diagnostic.Code, Path: composition.Diagnostic.Path, Message: composition.Diagnostic.Message})
			appendRelationDSEDiagnostic(&result, composition.Diagnostic)
			result.Residuals = append(result.Residuals, candidate.Residuals...)
		}
		if ok, appendErr := appendRelationSourceCandidate(&result, candidate, budget); appendErr != nil {
			return result, appendErr
		} else if !ok {
			return relationDSELimit(result, "output_bytes", result.OutputBytes+1, budget.MaxOutputBytes), nil
		}
		if composition.Status == status.OutcomeResourceLimit || result.Work > budget.MaxWork {
			result.Status, result.Complete = status.OutcomeResourceLimit, false
			if result.Work > budget.MaxWork {
				return relationDSELimit(result, "work", result.Work, budget.MaxWork), nil
			}
			return result, nil
		}
	}
	allOK, hasIncomplete, hasAmbiguous := true, false, match.Status == status.OutcomeAmbiguous
	for _, candidate := range result.Candidates {
		if candidate.Status != status.OutcomeOK {
			allOK = false
			if candidate.Status == status.OutcomeAmbiguous {
				hasAmbiguous = true
			} else {
				hasIncomplete = true
			}
		}
	}
	if !match.Complete || match.Status != status.OutcomeOK && match.Status != status.OutcomeAmbiguous {
		hasIncomplete = true
	}
	result.Complete = allOK && !hasIncomplete && len(result.Candidates) > 0
	switch {
	case hasIncomplete:
		result.Status, result.Complete = status.OutcomeIncomplete, false
	case hasAmbiguous || len(result.Candidates) > 1:
		result.Status = status.OutcomeAmbiguous
	case !allOK:
		result.Status, result.Complete = status.OutcomeIncomplete, false
	default:
		result.Status = status.OutcomeOK
	}
	if len(result.Candidates) == 0 {
		result.Status, result.Complete = status.OutcomeIncomplete, false
	}
	return result, nil
}

func relationSourceCandidate(index int, alternative RelationSourceAlternative, plan engine.RelationPathPlan, selection relationSourceSelection, composition *relationSourceComposition, options DSEOptions) semantic.DSECandidate {
	selection = cloneRelationSourceSelection(selection)
	variant := selection.Variant
	if variant.ID == "" {
		for _, candidate := range plan.Variants {
			if candidate.ID == alternative.VariantID {
				variant = candidate
				break
			}
		}
	}
	variantEvidence := semantic.DSERelationPlanVariantEvidence{ID: variant.ID, Parts: make([]semantic.DSERelationPlanPartEvidence, len(variant.Parts))}
	for partIndex, part := range variant.Parts {
		variantEvidence.Parts[partIndex] = semantic.DSERelationPlanPartEvidence{Step: part.Step, Literal: part.Literal, Slot: part.Slot, Port: part.Port, GrammaticalFunction: part.GrammaticalFunction}
	}
	parts := make([]semantic.DSERelationSourcePartEvidence, len(alternative.Parts))
	readingIDs := make([]string, 0, 2)
	var senseID string
	span := text.Span{}
	for partIndex, part := range alternative.Parts {
		part = cloneRelationSourcePart(part)
		parts[partIndex] = semantic.DSERelationSourcePartEvidence{
			PartIndex: part.PartIndex, Step: part.Step, Slot: part.Slot, Literal: part.Literal,
			Port: part.Port, PortPosition: part.PortPosition, GrammaticalFunction: part.GrammaticalFunction,
			TokenIndex: part.TokenIndex, Span: part.Span, Surface: part.Surface,
			Reading: part.Reading, Lexeme: part.Lexeme, Sense: part.Sense,
			ReadingLexemeID: "", LexemeID: "", SenseID: "", SenseLexemeID: "",
		}
		if part.Reading != nil {
			parts[partIndex].ReadingLexemeID = part.Reading.LexemeID
		}
		if part.Lexeme != nil {
			parts[partIndex].LexemeID = part.Lexeme.ID.String()
		}
		if part.Sense != nil {
			parts[partIndex].SenseID = part.Sense.ID.String()
			parts[partIndex].SenseLexemeID = part.Sense.LexemeID.String()
		}
		if part.Slot == "verb" && part.Reading != nil && part.Sense != nil {
			readingIDs = append(readingIDs, part.Reading.ID)
			if senseID == "" {
				senseID = part.Sense.ID.String()
				span = part.Span
			}
		}
	}
	if readingIDs == nil {
		readingIDs = []string{}
	}
	source := semantic.DSESourceRecord{CandidateID: fmt.Sprintf("relation-source:%d", index), VariantID: alternative.VariantID, ReadingIDs: readingIDs, SenseID: senseID, FrameID: string(selection.Frame.ID), Span: span}
	if len(readingIDs) > 0 {
		source.ReadingID = readingIDs[0]
	}
	evidence := &semantic.DSERelationSourceEvidence{
		VariantID: alternative.VariantID, Variant: variantEvidence,
		RelationPlanProvenance: plan.Provenance, SourceParts: parts,
		SourceContext: options.References.Clone(), VerificationScope: relationSourceVerificationScope,
	}
	if options.Plan != nil {
		evidence.SourcePlanProvenance = options.Plan.Document().Provenance
	}
	if composition != nil {
		evidence.TemporalFrames = composition.TemporalFrames
		for eventIndex := range evidence.TemporalFrames {
			if evidence.TemporalFrames[eventIndex] != "" {
				continue
			}
			if frame, outcome, _ := relationSourceReadingTemporalFrame(selection.VerbReadings[eventIndex], selection.Lexeme, selection.Sense); outcome == status.OutcomeOK {
				evidence.TemporalFrames[eventIndex] = frame
			}
		}
	}
	if selection.Binding.Target != nil || selection.Sense.ID.String() != "" || selection.Frame.ID != "" {
		rolePorts := make([]semantic.DSERelationRolePortEvidence, 0, len(selection.RolePorts))
		for role, port := range selection.RolePorts {
			rolePorts = append(rolePorts, semantic.DSERelationRolePortEvidence{Role: role, PortID: port.Role, Type: port.Type})
		}
		sort.Slice(rolePorts, func(i, j int) bool { return rolePorts[i].Role < rolePorts[j].Role })
		portRoles := make([]semantic.DSERelationPortRoleEvidence, 0, len(selection.PortRoles))
		for portID, role := range selection.PortRoles {
			portRoles = append(portRoles, semantic.DSERelationPortRoleEvidence{PortID: portID, Role: role})
		}
		sort.Slice(portRoles, func(i, j int) bool { return portRoles[i].PortID < portRoles[j].PortID })
		metadata := &semantic.DSERelationMetadataEvidence{
			Mapping: relationSourceMappingEvidence(selection.Mapping), Binding: selection.Binding,
			BindingEvidence: semantic.DSERelationBindingEvidence{SenseID: selection.Binding.SenseID.String(), CoreRef: selection.Binding.CoreRef.String(), Target: selection.Binding.Target, Mapping: string(selection.Binding.Mapping), Status: string(selection.Binding.Status), Provenance: selection.Binding.Provenance},
			Sense:           selection.Sense, SenseEvidence: relationSourceSenseEvidence(selection.Sense),
			Lexeme: selection.Lexeme, LexemeEvidence: relationSourceLexemeEvidence(selection.Lexeme),
			Frame: selection.Frame, FrameEvidence: relationSourceFrameEvidence(selection.Frame), Validation: selection.Validation, Signature: selection.Signature, HeadPort: selection.HeadPort,
			RolePorts: rolePorts, PortRoles: portRoles, SubjectRole: selection.SubjectRole, ObjectRole: selection.ObjectRole,
			SourceProvenance: selection.SourceProvenance, BindingProvenance: selection.BindingProvenance, FrameProvenance: selection.FrameProvenance,
		}
		evidence.Metadata = metadata
	}
	return semantic.DSECandidate{
		ID: fmt.Sprintf("relation-source:%d:%s", index, alternative.VariantID), Status: status.OutcomeIncomplete, Source: source,
		RoleSources: []semantic.DSERoleSource{}, GoalSpecs: []json.RawMessage{}, FragmentChecks: []semantic.DSEFragmentCheck{},
		ConditionCoverage: []semantic.DSEConditionCoverage{}, Residuals: []semantic.DSEResidual{}, Scope: relationSourceVerificationScope,
		RelationSourceEvidence: evidence,
	}
}

func relationSourceMappingEvidence(mapping language.SourcePredicateMapping) semantic.DSERelationPredicateMappingEvidence {
	out := semantic.DSERelationPredicateMappingEvidence{
		SenseID: mapping.SenseID.String(), FrameID: string(mapping.FrameID), HeadPortID: mapping.HeadPortID,
		AllowedFamilyIDs: make([]string, len(mapping.AllowedFamilyIDs)), AllowedVariantIDs: make([]string, len(mapping.AllowedVariantIDs)),
	}
	for index, id := range mapping.AllowedFamilyIDs {
		out.AllowedFamilyIDs[index] = id.String()
	}
	for index, id := range mapping.AllowedVariantIDs {
		out.AllowedVariantIDs[index] = string(id)
	}
	if mapping.ReferenceCondition != nil {
		condition := *mapping.ReferenceCondition
		out.ReferenceCondition = &condition
	}
	return out
}

func relationSourceSenseEvidence(sense lexicon.LexicalSense) semantic.DSERelationSenseEvidence {
	out := semantic.DSERelationSenseEvidence{ID: sense.ID.String(), LexemeID: sense.LexemeID.String(), Register: sense.Register, Usage: append([]string{}, sense.Usage...), ValencyIDs: make([]string, len(sense.ValencyIDs))}
	for index, id := range sense.ValencyIDs {
		out.ValencyIDs[index] = string(id)
	}
	return out
}

func relationSourceLexemeEvidence(lexeme lexicon.MorphLexeme) semantic.DSERelationLexemeEvidence {
	return semantic.DSERelationLexemeEvidence{ID: lexeme.ID.String(), Language: string(lexeme.Language), Lemma: lexeme.Lemma, POS: string(lexeme.POS), Properties: append([]lexicon.Property{}, lexeme.Properties...), ParadigmID: lexeme.ParadigmID}
}

func relationSourceFrameEvidence(frame valency.ValencyFrame) semantic.DSERelationFrameEvidence {
	out := semantic.DSERelationFrameEvidence{
		ID: string(frame.ID), SenseID: frame.SenseID.String(), SemanticRoles: make([]string, len(frame.SemanticRoles)),
		Arguments:              make([]semantic.DSERelationFrameArgumentEvidence, len(frame.Arguments)),
		SelectionalConstraints: make([]semantic.DSERelationFrameConstraintEvidence, len(frame.SelectionalConstraints)),
		AllowedFamilyIDs:       make([]string, len(frame.AllowedFamilyIDs)), Provenance: frame.Provenance,
	}
	for index, role := range frame.SemanticRoles {
		out.SemanticRoles[index] = string(role)
	}
	for index, argument := range frame.Arguments {
		out.Arguments[index] = semantic.DSERelationFrameArgumentEvidence{ID: argument.ID, Role: string(argument.Role), Required: argument.Required, GrammaticalFunction: argument.GrammaticalFunction, Case: argument.Case, Adposition: argument.Adposition}
	}
	for index, constraint := range frame.SelectionalConstraints {
		row := semantic.DSERelationFrameConstraintEvidence{ID: constraint.ID, Role: string(constraint.Role), Strength: string(constraint.Strength), AllowedConceptIDs: make([]string, len(constraint.AllowedConceptIDs)), AllowedKinds: make([]string, len(constraint.AllowedKinds))}
		for conceptIndex, id := range constraint.AllowedConceptIDs {
			row.AllowedConceptIDs[conceptIndex] = id.String()
		}
		for kindIndex, kind := range constraint.AllowedKinds {
			row.AllowedKinds[kindIndex] = string(kind)
		}
		out.SelectionalConstraints[index] = row
	}
	for index, id := range frame.AllowedFamilyIDs {
		out.AllowedFamilyIDs[index] = id.String()
	}
	return out
}

func appendRelationSourceCandidate(result *semantic.DSEResult, candidate semantic.DSECandidate, budget semantic.DSEBudget) (bool, error) {
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

func appendRelationDSEDiagnostic(result *semantic.DSEResult, diagnostic status.Diagnostic) {
	if len(result.Diagnostics) < status.MaxDiagnostics {
		result.Diagnostics = append(result.Diagnostics, diagnostic)
	}
}

func appendRelationDSEMatchDiagnostics(result *semantic.DSEResult, diagnostics []status.Diagnostic) {
	for _, diagnostic := range diagnostics {
		appendRelationDSEDiagnostic(result, diagnostic)
		if len(result.Diagnostics) >= status.MaxDiagnostics {
			return
		}
	}
}

func relationDSEDiagnostic(code, message string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Severity: status.SeverityError}
}

func relationDSELimit(result semantic.DSEResult, kind string, actual, maximum int) semantic.DSEResult {
	result.Status, result.Complete = status.OutcomeResourceLimit, false
	message := fmt.Sprintf("relation source %s budget exceeded: %d > %d", kind, actual, maximum)
	result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: "RELATION_SOURCE_RESOURCE_LIMIT", Path: "budget." + kind, Message: message})
	appendRelationDSEDiagnostic(&result, relationDSEDiagnostic("RESOURCE_LIMIT", message))
	return result
}

func relationDSEFailure(result semantic.DSEResult, kind status.OutcomeKind, code, path, message string) semantic.DSEResult {
	result.Status, result.Complete = kind, false
	appendRelationDSEDiagnostic(&result, status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError})
	result.Residuals = append(result.Residuals, semantic.DSEResidual{Code: code, Path: path, Message: message})
	return result
}
