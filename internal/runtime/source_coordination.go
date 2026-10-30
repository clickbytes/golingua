package runtime

import (
	"context"

	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func (pipeline *SourcePipeline) analyzeDSECoordination(ctx context.Context, syntax SourceSyntax, options DSEOptions) (DSEAnalysis, error) {
	coordination := syntax.Coordination
	if coordination == nil || len(coordination.Clauses) != 2 || len(syntax.Forest.Roots) != 0 {
		return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationAnalysis("INVALID_COORDINATION_FOREST", "the source did not produce exactly two independently retained clause forests")}, nil
	}
	if ctx == nil {
		return DSEAnalysis{}, pipelineError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "DSE coordination requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return DSEAnalysis{}, pipelineError(status.OutcomeResourceLimit, "DSE_CANCELED", "context", "DSE coordination was canceled or timed out")
	}
	if len(options.References.RoleReferences) != 0 || len(options.References.PortReferences) != 0 {
		return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationAnalysis("COORDINATION_GLOBAL_REFERENCE_MAPPING_UNSUPPORTED", "candidate-linked identity evidence is required for each coordinated clause")}, nil
	}
	if coordination.Connector != pipeline.ClauseConnector || coordination.Clauses[0].TokenStart >= coordination.Clauses[0].TokenEnd ||
		coordination.Clauses[1].TokenStart >= coordination.Clauses[1].TokenEnd || coordination.Clauses[0].TokenEnd > coordination.Clauses[1].TokenStart {
		return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationAnalysis("INVALID_COORDINATION_BOUNDARY", "two source clause ranges and their language connector must be explicit and ordered")}, nil
	}
	inputs := [2]semantic.DSEInput{}
	for index, clause := range coordination.Clauses {
		if clause.Forest.Outcome != status.OutcomeOK {
			return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationWithStatus(clause.Forest.Outcome, clause.Forest.Diagnostics, "COORDINATION_CLAUSE_STATUS", "a coordinated source clause did not produce an OK parse")}, nil
		}
		if len(clause.Forest.Roots) != 1 || len(clause.Forest.Unmatched) != 0 || !clause.Evidence.OperatorCoverageComplete || len(clause.Evidence.Operators) != 0 {
			return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationAnalysis("COORDINATION_CLAUSE_NOT_POSITIVE_DECLARATIVE", "both coordinated clauses must be complete positive declaratives with complete operator coverage")}, nil
		}
		root, found := sourceClauseRoot(clause.Forest)
		if !found || root.Start != clause.TokenStart || root.End != clause.TokenEnd || len(root.Alternatives) != 1 {
			return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationWithStatus(status.OutcomeAmbiguous, clause.Forest.Diagnostics, "COORDINATION_CLAUSE_AMBIGUOUS", "each clause must retain exactly one construction alternative over its declared range")}, nil
		}
		candidates, evidence, sourceContext, ok := coordinationClauseInputs(syntax, clause, root, options.References)
		if !ok {
			return DSEAnalysis{Syntax: syntax, Composition: incompleteCoordinationAnalysis("COORDINATION_CLAUSE_IDENTITY_UNRESOLVED", "clause roles require candidate-linked explicit reference evidence")}, nil
		}
		inputs[index] = semantic.DSEInput{
			Context: ctx, Forest: clause.Forest, Candidates: candidates, Lattice: syntax.Lattice,
			SourceBytes: syntax.OriginalSourceBytes, References: evidence.references, ImplicitRoles: evidence.implicitRoles,
			NodeAttributes: evidence.attributes, Operators: clause.Evidence.Operators,
			OperatorCoverageComplete: clause.Evidence.OperatorCoverageComplete,
			Package:                  options.Package, Lexicon: pipeline.Lexicon, Valency: pipeline.Valency,
			Core: pipeline.Core, Constructions: pipeline.Constructions, Plan: options.Plan, Selected: options.Selected,
			Rewrite: options.Rewrite, ReferenceContext: sourceContext, Language: pipeline.Language,
			Budget: options.Budget,
		}
	}
	composition, err := semantic.ComposeDSEConjunction(inputs)
	if err != nil {
		return DSEAnalysis{}, err
	}
	return DSEAnalysis{Syntax: syntax, Composition: composition}, nil
}

type clauseInputEvidence struct {
	references    []semantic.ReferenceEvidence
	implicitRoles []semantic.ImplicitRoleEvidence
	attributes    []semantic.NodeAttributesEvidence
}

func sourceClauseRoot(forest parser.Forest) (parser.PackedNode, bool) {
	if len(forest.Roots) != 1 {
		return parser.PackedNode{}, false
	}
	for _, node := range forest.Nodes {
		if node.ID == forest.Roots[0] && string(node.Category) == "clause" {
			return node, true
		}
	}
	return parser.PackedNode{}, false
}

func coordinationClauseInputs(syntax SourceSyntax, clause SourceClauseSyntax, root parser.PackedNode, references language.SourceReferenceContext) ([]parser.PhraseCandidate, clauseInputEvidence, language.SourceReferenceContext, bool) {
	alternative := root.Alternatives[0]
	selectedIDs := make(map[string]bool)
	candidatesByID := make(map[string]parser.PhraseCandidate)
	for _, candidate := range syntax.PhraseCandidates {
		if candidate.TokenStart >= clause.TokenStart && candidate.TokenEnd <= clause.TokenEnd {
			candidatesByID[candidate.ID] = candidate
		}
	}
	for _, slot := range alternative.Slots {
		for _, candidateID := range slot.CandidateIDs {
			if _, exists := candidatesByID[candidateID]; !exists {
				return nil, clauseInputEvidence{}, language.SourceReferenceContext{}, false
			}
			selectedIDs[candidateID] = true
		}
	}
	if len(selectedIDs) < 3 { // one predicate plus the two explicit frame roles
		return nil, clauseInputEvidence{}, language.SourceReferenceContext{}, false
	}
	selectedCandidates := make([]parser.PhraseCandidate, 0, len(candidatesByID))
	for _, candidate := range syntax.PhraseCandidates {
		if _, found := candidatesByID[candidate.ID]; found {
			selectedCandidates = append(selectedCandidates, candidate)
		}
	}
	_, linked := candidateLinkedReferenceIDs(references, selectedIDs, clause.Evidence.References)
	if !linked {
		return nil, clauseInputEvidence{}, language.SourceReferenceContext{}, false
	}
	// Keep every original declaration and request-context record. ComposeDSE
	// validates the caller's full records while considering this clause.
	local := references.Clone()
	clauseEvidence := clauseInputEvidence{}
	for _, row := range clause.Evidence.References {
		if selectedIDs[row.CandidateID] {
			clauseEvidence.references = append(clauseEvidence.references, row)
		}
	}
	for _, row := range clause.Evidence.ImplicitRoles {
		if row.DerivationID == alternative.DerivationID && selectedIDs[row.PredicateCandidateID] {
			clauseEvidence.implicitRoles = append(clauseEvidence.implicitRoles, row)
		}
	}
	for _, row := range clause.Evidence.NodeAttributes {
		if selectedIDs[row.CandidateID] {
			clauseEvidence.attributes = append(clauseEvidence.attributes, row)
		}
	}
	return selectedCandidates, clauseEvidence, local, true
}

func candidateLinkedReferenceIDs(references language.SourceReferenceContext, selected map[string]bool, evidence []semantic.ReferenceEvidence) (map[string]bool, bool) {
	ids := make(map[string]bool)
	declarations := make(map[string]bool, len(references.References))
	for _, declaration := range references.References {
		declarations[declaration.ID] = true
	}
	mentionByCandidate := make(map[string]string)
	resolutionByMention := make(map[string]map[string]bool)
	if references.RequestContext != nil {
		for _, mention := range references.RequestContext.Mentions {
			mentionByCandidate[mention.ID] = mention.ReferentID
		}
		for _, resolution := range references.RequestContext.ResolutionEvidence {
			if resolutionByMention[resolution.MentionID] == nil {
				resolutionByMention[resolution.MentionID] = map[string]bool{}
			}
			resolutionByMention[resolution.MentionID][resolution.ReferentID] = true
		}
	}
	evidenceByCandidate := make(map[string]string)
	for _, row := range evidence {
		if row.ReferentID == "" || row.Anchor != nil {
			continue
		}
		evidenceByCandidate[row.CandidateID] = string(row.ReferentID)
	}
	for candidateID := range selected {
		resolved := map[string]bool{}
		if referenceID := evidenceByCandidate[candidateID]; referenceID != "" {
			resolved[referenceID] = true
		}
		mentionID := candidateID
		if referenceID := mentionByCandidate[mentionID]; referenceID != "" {
			resolved[referenceID] = true
		}
		for referenceID := range resolutionByMention[mentionID] {
			resolved[referenceID] = true
		}
		if len(resolved) != 1 {
			return nil, false
		}
		var referenceID string
		for candidateReferenceID := range resolved {
			referenceID = candidateReferenceID
		}
		if !declarations[referenceID] {
			return nil, false
		}
		ids[referenceID] = true
	}
	return ids, true
}

func incompleteCoordinationAnalysis(code, message string) semantic.DSEResult {
	return incompleteCoordinationWithStatus(status.OutcomeIncomplete, nil, code, message)
}

func incompleteCoordinationWithStatus(outcome status.OutcomeKind, diagnostics []status.Diagnostic, code, message string) semantic.DSEResult {
	if outcome == status.OutcomeOK || outcome == "" {
		outcome = status.OutcomeIncomplete
	}
	return semantic.DSEResult{
		Status: outcome, Complete: false, Candidates: []semantic.DSECandidate{},
		Residuals:   []semantic.DSEResidual{{Code: code, Path: "coordination", Message: message}},
		Diagnostics: append([]status.Diagnostic{}, diagnostics...), WholeVerificationScope: "source coordination was retained without selecting or asserting one complete meaning",
	}
}
