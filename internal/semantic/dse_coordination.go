package semantic

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// ComposeDSEConjunction composes exactly two already analyzed source clauses.
// Both clauses must independently produce one complete proposition. The
// result preserves their order and shares request-local entity identities only
// where the checked input documents explicitly use the same reference ID.
func ComposeDSEConjunction(inputs [2]DSEInput) (DSEResult, error) {
	leftBudget, err := NormalizeDSEBudget(inputs[0].Budget)
	if err != nil {
		return DSEResult{}, err
	}
	rightBudget, err := NormalizeDSEBudget(inputs[1].Budget)
	if err != nil {
		return DSEResult{}, err
	}
	if leftBudget.MaxWork != rightBudget.MaxWork || leftBudget.MaxOutputBytes != rightBudget.MaxOutputBytes || leftBudget.MaxAlternatives != rightBudget.MaxAlternatives {
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_BUDGET_MISMATCH", "budget", "both clause compositions must use one common request budget")
	}
	budget := leftBudget
	if inputs[0].Context == nil || inputs[1].Context == nil {
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_REQUEST_CONTEXT_MISMATCH", "context", "both clauses must be composed within the same request context")
	}
	if inputs[0].Context.Err() != nil || inputs[1].Context.Err() != nil {
		return dseCoordinationResource("DSE_CANCELED", "context", "the coordinated request was canceled before the clauses could be composed"), nil
	}
	for index := range inputs {
		if len(inputs[index].ReferenceContext.RoleReferences) != 0 || len(inputs[index].ReferenceContext.PortReferences) != 0 {
			return dseCoordinationUnsupported("COORDINATION_GLOBAL_REFERENCE_MAPPING_UNSUPPORTED", fmt.Sprintf("clauses[%d].reference_context", index), "coordinated clauses require candidate-linked identity evidence; global role or port mappings cannot be reused safely"), nil
		}
	}
	if budget.MaxWork < 2 || budget.MaxOutputBytes < 2 {
		return dseCoordinationResource("DSE_COORDINATION_BUDGET", "budget", "the shared request budget cannot cover two clause compositions"), nil
	}
	leftPart := budget.MaxWork / 2
	rightPart := budget.MaxWork - leftPart
	leftBytes := budget.MaxOutputBytes / 2
	rightBytes := budget.MaxOutputBytes - leftBytes
	if leftPart < 1 || rightPart < 1 || leftBytes < 1 || rightBytes < 1 {
		return dseCoordinationResource("DSE_COORDINATION_BUDGET", "budget", "the shared request budget cannot cover both clauses"), nil
	}
	inputs[0].Budget = DSEBudget{MaxAlternatives: 1, MaxWork: leftPart, MaxOutputBytes: leftBytes}
	inputs[1].Budget = DSEBudget{MaxAlternatives: 1, MaxWork: rightPart, MaxOutputBytes: rightBytes}
	inputs[0].RequireCandidateHeadReference = true
	inputs[1].RequireCandidateHeadReference = true

	left, err := ComposeDSE(inputs[0])
	if err != nil {
		return DSEResult{}, err
	}
	right, err := ComposeDSE(inputs[1])
	if err != nil {
		return DSEResult{}, err
	}
	work := left.Work + right.Work
	if inputs[0].Context.Err() != nil || inputs[1].Context.Err() != nil {
		return dseCoordinationResource("DSE_CANCELED", "context", "the coordinated request was canceled before the clauses could be combined"), nil
	}
	if left.Status != status.OutcomeOK || !left.Complete || len(left.Candidates) != 1 ||
		right.Status != status.OutcomeOK || !right.Complete || len(right.Candidates) != 1 {
		return dseCoordinationIncomplete(left, right, work, budget.MaxOutputBytes), nil
	}
	first, second := left.Candidates[0], right.Candidates[0]
	if first.Status != status.OutcomeOK || second.Status != status.OutcomeOK ||
		first.CanonicalDocument == nil || second.CanonicalDocument == nil ||
		first.HeadEventReferenceID == "" || second.HeadEventReferenceID == "" {
		return dseCoordinationIncomplete(left, right, work, budget.MaxOutputBytes), nil
	}
	if first.HeadEventReferenceID == second.HeadEventReferenceID {
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_EVENT_IDENTITY_COLLAPSED", "clauses", "the two source clauses must retain distinct explicit Event references")
	}
	firstDoc, err := algebra.DecodeDocumentJSON(first.CanonicalDocument)
	if err != nil {
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_LEFT_DOCUMENT_INVALID", "clauses[0].canonical_document", err.Error())
	}
	secondDoc, err := algebra.DecodeDocumentJSON(second.CanonicalDocument)
	if err != nil {
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_RIGHT_DOCUMENT_INVALID", "clauses[1].canonical_document", err.Error())
	}
	if firstDoc.Expression == nil || firstDoc.Expression.Kind != algebra.TermAtom || secondDoc.Expression == nil || secondDoc.Expression.Kind != algebra.TermAtom {
		return dseCoordinationIncomplete(left, right, work, budget.MaxOutputBytes), nil
	}
	if !reflect.DeepEqual(firstDoc.Sorts, secondDoc.Sorts) || !reflect.DeepEqual(firstDoc.Signatures, secondDoc.Signatures) ||
		!reflect.DeepEqual(firstDoc.Parameters, secondDoc.Parameters) || firstDoc.Limits != secondDoc.Limits || firstDoc.Closed != secondDoc.Closed {
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_DECLARATION_CONFLICT", "clauses", "the clause documents disagree on their selected declarations or checker limits")
	}
	sharedEntity := dseSharedEntityReference(firstDoc.Expression, secondDoc.Expression)
	if !sharedEntity {
		return dseCoordinationIncomplete(left, right, work, budget.MaxOutputBytes), nil
	}
	references, err := mergeDSEDeclarations(firstDoc.References, secondDoc.References)
	if err != nil {
		return DSEResult{}, err
	}
	requestContext, err := mergeDSERequestContexts(firstDoc.RequestContext, secondDoc.RequestContext)
	if err != nil {
		return DSEResult{}, err
	}
	expression := &algebra.Term{Kind: algebra.TermAnd, Operands: []*algebra.Term{firstDoc.Expression, secondDoc.Expression}}
	combinedDoc := algebra.Document{
		SchemaVersion:  firstDoc.SchemaVersion,
		Sorts:          append([]algebra.SortDeclaration{}, firstDoc.Sorts...),
		Signatures:     cloneAlgebraSignatures(firstDoc.Signatures),
		Parameters:     cloneAlgebraDeclarations(firstDoc.Parameters),
		References:     references,
		Closed:         firstDoc.Closed,
		Expression:     expression,
		Limits:         firstDoc.Limits,
		RequestContext: requestContext,
	}
	checked, err := algebra.Check(combinedDoc)
	if err != nil || checked.Status != status.OutcomeOK || checked.Type.Kind != algebra.TypeProp {
		message := "conjoined source clauses did not form one complete typed proposition"
		if err != nil {
			message = err.Error()
		}
		return DSEResult{}, dseError(status.OutcomeInvalid, "COORDINATION_EXPRESSION_INVALID", "clauses", message)
	}
	encoded, err := algebra.Encode(combinedDoc)
	if err != nil {
		return DSEResult{}, err
	}
	if inputs[0].Context.Err() != nil || inputs[1].Context.Err() != nil {
		return dseCoordinationResource("DSE_CANCELED", "context", "the coordinated request was canceled before the clauses could be combined"), nil
	}
	combined := DSECandidate{
		ID:     dseStableID("coordination", first.ID, second.ID),
		Status: status.OutcomeOK,
		Source: DSESourceRecord{ReadingIDs: []string{}, Span: combinedForestSourceSpan(inputs[0].Forest, inputs[1].Forest, first.Source.Span, second.Source.Span)},
		ClauseSources: []DSEClauseSource{
			{Source: first.Source, HeadEventReferenceID: first.HeadEventReferenceID, RoleSources: cloneDSERoleSources(first.RoleSources)},
			{Source: second.Source, HeadEventReferenceID: second.HeadEventReferenceID, RoleSources: cloneDSERoleSources(second.RoleSources)},
		},
		RoleSources:       append(cloneDSERoleSources(first.RoleSources), cloneDSERoleSources(second.RoleSources)...),
		CanonicalDocument: append(json.RawMessage{}, encoded...),
		Checked:           &checked,
		GoalSpecs:         append(append([]json.RawMessage{}, first.GoalSpecs...), second.GoalSpecs...),
		FragmentChecks:    append(append([]DSEFragmentCheck{}, first.FragmentChecks...), second.FragmentChecks...),
		ConditionCoverage: append(append([]DSEConditionCoverage{}, first.ConditionCoverage...), second.ConditionCoverage...),
		Residuals:         []DSEResidual{},
		Scope:             "two explicitly coordinated source clauses; typed proposition conjunction and request-local identity only; no world-fact claim",
	}
	encodedCandidate, err := json.Marshal(combined)
	if err != nil {
		return DSEResult{}, err
	}
	combinedWork := work + checked.NodeCount
	if combinedWork > budget.MaxWork || len(encoded) > budget.MaxOutputBytes || len(encodedCandidate) > budget.MaxOutputBytes {
		return dseCoordinationResource("DSE_COORDINATION_OUTPUT_LIMIT", combined.ID, "the combined candidate exceeds the shared request work or output budget"), nil
	}
	return DSEResult{
		Status: status.OutcomeOK, Complete: true,
		Candidates: []DSECandidate{combined}, Residuals: []DSEResidual{}, Diagnostics: []status.Diagnostic{},
		Work: combinedWork, OutputBytes: len(encodedCandidate),
		WholeVerificationScope: "type_scope_context_only",
	}, nil
}

func dseCoordinationIncomplete(left, right DSEResult, work, maxOutputBytes int) DSEResult {
	resultStatus := status.OutcomeIncomplete
	statuses := []status.OutcomeKind{left.Status, right.Status}
	for _, result := range []DSEResult{left, right} {
		for _, candidate := range result.Candidates {
			statuses = append(statuses, candidate.Status)
		}
	}
	for _, candidateStatus := range statuses {
		switch candidateStatus {
		case status.OutcomeResourceLimit:
			resultStatus = status.OutcomeResourceLimit
		case status.OutcomeAmbiguous:
			if resultStatus != status.OutcomeResourceLimit {
				resultStatus = status.OutcomeAmbiguous
			}
		case status.OutcomeUnsupported:
			if resultStatus != status.OutcomeResourceLimit && resultStatus != status.OutcomeAmbiguous {
				resultStatus = status.OutcomeUnsupported
			}
		case status.OutcomeUnknown:
			if resultStatus == status.OutcomeIncomplete {
				resultStatus = status.OutcomeUnknown
			}
		}
	}
	out := DSEResult{
		Status: resultStatus, Complete: false,
		Candidates: []DSECandidate{}, Residuals: append([]DSEResidual{}, left.Residuals...),
		Diagnostics: append([]status.Diagnostic{}, left.Diagnostics...), Work: work,
		WholeVerificationScope: "two source clauses not yet composed; partial clause candidates are not returned as a complete conjunction",
	}
	for _, candidate := range left.Candidates {
		out.Residuals = append(out.Residuals, candidate.Residuals...)
	}
	for _, candidate := range right.Candidates {
		out.Residuals = append(out.Residuals, candidate.Residuals...)
	}
	out.Residuals = append(out.Residuals, right.Residuals...)
	out.Diagnostics = append(out.Diagnostics, right.Diagnostics...)
	if len(out.Residuals) == 0 && len(out.Diagnostics) == 0 {
		out.Residuals = append(out.Residuals, DSEResidual{Code: "COORDINATION_CLAUSE_INCOMPLETE", Message: "both source clauses must independently yield one complete typed Atom"})
	}
	encoded, err := json.Marshal(out)
	if err != nil || len(encoded) > maxOutputBytes {
		return dseCoordinationResource("DSE_COORDINATION_OUTPUT_LIMIT", "clauses", "the retained clause diagnostics and residuals exceed the shared output budget")
	}
	out.OutputBytes = len(encoded)
	return out
}

func dseCoordinationResource(code, path, message string) DSEResult {
	return DSEResult{
		Status: status.OutcomeResourceLimit, Complete: false,
		Candidates: []DSECandidate{}, Residuals: []DSEResidual{{Code: code, Path: path, Message: message}},
		Diagnostics: []status.Diagnostic{}, WholeVerificationScope: "coordination was not completed within the shared finite request budget",
	}
}

func dseCoordinationUnsupported(code, path, message string) DSEResult {
	return DSEResult{
		Status: status.OutcomeUnsupported, Complete: false,
		Candidates: []DSECandidate{}, Residuals: []DSEResidual{{Code: code, Path: path, Message: message}},
		Diagnostics: []status.Diagnostic{}, WholeVerificationScope: "coordination was refused because this reference mapping is not supported",
	}
}

func dseSharedEntityReference(left, right *algebra.Term) bool {
	if left == nil || right == nil || left.Kind != algebra.TermAtom || right.Kind != algebra.TermAtom {
		return false
	}
	leftRefs := map[string]bool{}
	for _, arg := range left.Args {
		if arg.Value != nil && arg.Value.Kind == algebra.TermRef && arg.Value.Type != nil && arg.Value.Type.Kind == algebra.TypeEntity {
			leftRefs[arg.Value.ID] = true
		}
	}
	for _, arg := range right.Args {
		if arg.Value != nil && arg.Value.Kind == algebra.TermRef && arg.Value.Type != nil && arg.Value.Type.Kind == algebra.TypeEntity && leftRefs[arg.Value.ID] {
			return true
		}
	}
	return false
}

func mergeDSEDeclarations(left, right []algebra.Declaration) ([]algebra.Declaration, error) {
	merged := cloneAlgebraDeclarations(left)
	indexes := make(map[string]int, len(merged))
	for i, row := range merged {
		if _, duplicate := indexes[row.ID]; duplicate {
			return nil, dseError(status.OutcomeInvalid, "COORDINATION_DUPLICATE_REFERENCE", "references", "a clause document contains duplicate reference declarations")
		}
		indexes[row.ID] = i
	}
	for _, row := range right {
		index, exists := indexes[row.ID]
		if !exists {
			indexes[row.ID] = len(merged)
			merged = append(merged, row)
			continue
		}
		current := &merged[index]
		if !reflect.DeepEqual(current.Type, row.Type) {
			return nil, dseError(status.OutcomeInvalid, "COORDINATION_REFERENCE_TYPE_CONFLICT", "references."+row.ID, "the shared reference has incompatible types")
		}
		if err := mergeOptionalDSEField(&current.TemporalFrame, row.TemporalFrame, "COORDINATION_TEMPORAL_FRAME_CONFLICT", row.ID); err != nil {
			return nil, err
		}
		if err := mergeOptionalDSEField(&current.ReferenceTimeID, row.ReferenceTimeID, "COORDINATION_REFERENCE_TIME_CONFLICT", row.ID); err != nil {
			return nil, err
		}
		if err := mergeOptionalDSEField(&current.Multiplicity, row.Multiplicity, "COORDINATION_MULTIPLICITY_CONFLICT", row.ID); err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func mergeOptionalDSEField(target *string, incoming, code, id string) error {
	if *target != "" && incoming != "" && *target != incoming {
		return dseError(status.OutcomeInvalid, code, "references."+id, "the shared reference has conflicting explicit attributes")
	}
	if *target == "" {
		*target = incoming
	}
	return nil
}

func mergeDSERequestContexts(left, right *algebra.RequestContext) (*algebra.RequestContext, error) {
	if left == nil && right == nil {
		return nil, nil
	}
	if left == nil || right == nil || left.SchemaVersion != right.SchemaVersion || left.ID != right.ID ||
		left.SpeakerReferentID != right.SpeakerReferentID || left.AddresseeReferentID != right.AddresseeReferentID {
		return nil, dseError(status.OutcomeInvalid, "COORDINATION_CONTEXT_CONFLICT", "request_context", "the clauses do not share one compatible explicit request context")
	}
	merged := *left
	merged.Referents = append([]algebra.ContextReferent{}, left.Referents...)
	merged.Mentions = append([]algebra.ContextMention{}, left.Mentions...)
	merged.ResolutionEvidence = append([]algebra.ResolutionEvidence{}, left.ResolutionEvidence...)
	merged.Conditions = append([]algebra.ContextCondition{}, left.Conditions...)
	referentIndexes := make(map[string]int, len(merged.Referents))
	for i, row := range merged.Referents {
		referentIndexes[row.ID] = i
	}
	for _, row := range right.Referents {
		index, exists := referentIndexes[row.ID]
		if !exists {
			referentIndexes[row.ID] = len(merged.Referents)
			merged.Referents = append(merged.Referents, row)
			continue
		}
		current := &merged.Referents[index]
		if !reflect.DeepEqual(current.Type, row.Type) {
			return nil, dseError(status.OutcomeInvalid, "COORDINATION_CONTEXT_TYPE_CONFLICT", "request_context.referents."+row.ID, "the shared context referent has incompatible types")
		}
		if err := mergeOptionalDSEField(&current.ConceptID, row.ConceptID, "COORDINATION_CONTEXT_CONCEPT_CONFLICT", row.ID); err != nil {
			return nil, err
		}
		if err := mergeOptionalDSEField(&current.TemporalFrame, row.TemporalFrame, "COORDINATION_CONTEXT_TIME_CONFLICT", row.ID); err != nil {
			return nil, err
		}
		if err := mergeOptionalDSEField(&current.ReferenceTimeID, row.ReferenceTimeID, "COORDINATION_CONTEXT_REFERENCE_TIME_CONFLICT", row.ID); err != nil {
			return nil, err
		}
		if err := mergeOptionalDSEField(&current.Multiplicity, row.Multiplicity, "COORDINATION_CONTEXT_MULTIPLICITY_CONFLICT", row.ID); err != nil {
			return nil, err
		}
	}
	for _, row := range right.Mentions {
		if !appendUniqueDSEContext(&merged.Mentions, row.ID, row, func(existing, incoming any) bool { return reflect.DeepEqual(existing, incoming) }) {
			return nil, dseError(status.OutcomeInvalid, "COORDINATION_MENTION_CONFLICT", "request_context.mentions."+row.ID, "a repeated context mention has conflicting explicit resolution")
		}
	}
	for _, row := range right.ResolutionEvidence {
		if !appendUniqueDSEContext(&merged.ResolutionEvidence, row.ID, row, func(existing, incoming any) bool { return reflect.DeepEqual(existing, incoming) }) {
			return nil, dseError(status.OutcomeInvalid, "COORDINATION_EVIDENCE_CONFLICT", "request_context.resolution_evidence."+row.ID, "repeated identity evidence has conflicting provenance")
		}
	}
	for _, row := range right.Conditions {
		if !appendUniqueDSEContext(&merged.Conditions, row.ID, row, func(existing, incoming any) bool {
			oldBytes, _ := json.Marshal(existing)
			newBytes, _ := json.Marshal(incoming)
			var oldCondition, newCondition algebra.ContextCondition
			_ = json.Unmarshal(oldBytes, &oldCondition)
			_ = json.Unmarshal(newBytes, &newCondition)
			oldCondition.ID, newCondition.ID = "", ""
			return reflect.DeepEqual(oldCondition, newCondition)
		}) {
			return nil, dseError(status.OutcomeInvalid, "COORDINATION_CONDITION_CONFLICT", "request_context.conditions."+row.ID, "same condition ID has different explicit content")
		}
	}
	return &merged, nil
}

func appendUniqueDSEContext[T any](rows *[]T, id string, incoming T, equal func(any, any) bool) bool {
	for _, current := range *rows {
		data, _ := json.Marshal(current)
		var object map[string]any
		_ = json.Unmarshal(data, &object)
		if object["id"] == id {
			return equal(current, incoming)
		}
	}
	*rows = append(*rows, incoming)
	return true
}

func cloneDSERoleSources(source []DSERoleSource) []DSERoleSource {
	return append([]DSERoleSource{}, source...)
}

func combinedSourceSpan(left, right text.Span) text.Span {
	return text.Span{Start: left.Start, End: right.End}
}

func combinedForestSourceSpan(leftForest, rightForest parser.Forest, leftFallback, rightFallback text.Span) text.Span {
	left, leftOK := forestRootSpan(leftForest)
	right, rightOK := forestRootSpan(rightForest)
	if !leftOK {
		left = leftFallback
	}
	if !rightOK {
		right = rightFallback
	}
	return combinedSourceSpan(left, right)
}

func forestRootSpan(forest parser.Forest) (text.Span, bool) {
	if len(forest.Roots) != 1 {
		return text.Span{}, false
	}
	for _, node := range forest.Nodes {
		if node.ID == forest.Roots[0] {
			return node.Span, true
		}
	}
	return text.Span{}, false
}
