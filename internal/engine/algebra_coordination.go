package engine

import (
	"context"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// realizeTwoAtomAnd builds one statement graph for exactly two directly
// supported positive atoms. A non-nil result means the bounded route matched.
func (engine *Engine) realizeTwoAtomAnd(ctx context.Context, input algebra.Document, languageID features.LanguageID, pkg *language.SemanticLanguagePackage, matches []algebraAtomMatch, residuals []AlgebraRealizationResidual, refuseContextResiduals bool) (*AlgebraRealizationResult, error) {
	if input.Expression == nil || input.Expression.Kind != algebra.TermAnd || len(input.Expression.Operands) != 2 {
		return nil, nil
	}
	result := AlgebraRealizationResult{Status: status.OutcomeInvalid, Language: string(languageID), LanguagePackage: pkg.ID(), PackageVersion: pkg.Version(), Coverage: []AlgebraRealizationCoverage{}, Residuals: append([]AlgebraRealizationResidual{}, residuals...), Alternatives: []SurfaceAlternative{}, Diagnostics: []status.Diagnostic{}, Scope: "exact two ordered positive direct Atoms under And; each clause must have a verified target realization, then the language-local coordination plan joins the clauses; residuals and target assumptions remain explicit"}
	if len(residuals) != 0 {
		result.Status = status.OutcomeUnsupported
		return &result, nil
	}
	if len(matches) != 2 {
		result.Status = status.OutcomeUnsupported
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "COORDINATION_REQUIRES_TWO_EXACT_ATOMS", Message: "both direct positive clauses must have exact approved target bindings", Term: cloneAlgebraTerm(input.Expression)})
		return &result, nil
	}
	for i := range matches {
		if matches[i].atom == nil || matches[i].atom.Kind != algebra.TermAtom || matches[i].answerVariableID != "" {
			return &result, nil
		}
	}
	if matches[0].arguments[matches[0].eventPort].ID == matches[1].arguments[matches[1].eventPort].ID {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "REPEATED_EVENT_IDENTITY", Message: "coordinated clauses require distinct Event references", Term: cloneAlgebraTerm(input.Expression)})
		result.Status = status.OutcomeAmbiguous
		return &result, nil
	}
	parts := make([]*semantic.MeaningCandidate, 2)
	cov := make([]AlgebraRealizationCoverage, 2)
	for i, match := range matches {
		candidate, coverage, ctxResiduals, failure, code, path, message := buildAlgebraMeaningCandidate(engine, input, languageID, pkg, match)
		if candidate == nil {
			result.Residuals = append(result.Residuals, ctxResiduals...)
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: path, Code: code, Message: message, Term: cloneAlgebraTerm(match.atom)})
			result.Status = failure
			return &result, nil
		}
		parts[i], cov[i] = candidate, coverage
		for _, r := range ctxResiduals {
			if r.Code == "UNCONSUMED_CONTEXT_REFERENT" && coveredByEither(matches, r.ReferenceID) {
				continue
			}
			if !containsAlgebraResidual(result.Residuals, r) {
				result.Residuals = append(result.Residuals, r)
			}
		}
	}
	if refuseContextResiduals && len(result.Residuals) > 0 {
		result.Status = status.OutcomeUnsupported
		return &result, nil
	}
	graph, requestContext, remapped, entityIDs, err := mergeTwoAlgebraCandidates(parts, cov, input.RequestContext.ID)
	if err != nil {
		result.Status = status.OutcomeInvalid
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "meaning", Code: "COORDINATION_MEANING_BUILD_FAILED", Message: err.Error()})
		return &result, nil
	}
	for i := range cov {
		cov[i].EventMeaningNode = remapped[i]
		for j := range cov[i].RoleSubstitutions {
			if cov[i].RoleSubstitutions[j].Role == "event_identity" {
				cov[i].RoleSubstitutions[j].MeaningNodeID = remapped[i]
			} else {
				cov[i].RoleSubstitutions[j].MeaningNodeID = entityIDs[cov[i].RoleSubstitutions[j].ReferenceID]
			}
		}
	}
	result.Coverage = append(result.Coverage, cov...)
	if err := meaning.ValidateMeaning(graph, engine.core, meaning.DefaultBudget()); err != nil {
		result.Status = status.OutcomeInvalid
		return &result, nil
	}
	if err := meaning.ValidateReferences(graph, requestContext, meaning.DefaultBudget()); err != nil {
		result.Status = status.OutcomeInvalid
		return &result, nil
	}
	combinedID := algebraCandidateID(input, pkg.ID(), cov[0].SenseID+":"+cov[1].SenseID)
	candidate := &semantic.MeaningCandidate{ID: combinedID, SemanticID: combinedID, Meaning: &graph,
		Contexts: []semantic.MeaningContext{{Meaning: graph, Context: requestContext}}, Status: status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: combinedID, Outcome: status.OutcomeOK, Records: []semantic.EvidenceRecord{}}}
	target, ok := engine.TargetRealizerFor(languageID)
	if !ok || target == nil {
		result.Status = status.OutcomeUnsupported
		result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: "TARGET_RUNTIME_UNAVAILABLE", Path: "language", Message: "no verified target adapter is available for the requested language", Severity: status.SeverityError})
		return &result, nil
	}
	targetLanguage, registered := engine.languages.Language(languageID)
	if !registered {
		result.Status = status.OutcomeUnsupported
		return &result, nil
	}
	budget := engineBudget(engine)
	requestCtx, cancel := context.WithTimeout(ctx, budget.RequestTimeout)
	defer cancel()
	surface, realizeErr := target.Realize(requestCtx, *candidate)
	if realizeErr != nil {
		return nil, realizeErr
	}
	returnResult, validationErr := validateAlgebraTargetSurface(result, surface, matches, candidate.Meaning, targetLanguage, budget)
	return &returnResult, validationErr
}

func coveredByEither(matches []algebraAtomMatch, ref string) bool {
	for _, m := range matches {
		for _, a := range m.arguments {
			if a != nil && a.Kind == algebra.TermRef && a.ID == ref {
				return true
			}
		}
	}
	return false
}
func containsAlgebraResidual(values []AlgebraRealizationResidual, value AlgebraRealizationResidual) bool {
	for _, v := range values {
		if v.Code == value.Code && v.ReferenceID == value.ReferenceID && v.Path == value.Path {
			return true
		}
	}
	return false
}

func mergeTwoAlgebraCandidates(parts []*semantic.MeaningCandidate, coverage []AlgebraRealizationCoverage, originalContextID string) (meaning.Meaning, meaning.RequestContext, [2]meaning.NodeID, map[string]meaning.NodeID, error) {
	var events [2]meaning.NodeID
	entityIDs := map[string]meaning.NodeID{}
	referentIDs := map[string]meaning.ReferentID{}
	keysSet := map[string]bool{}
	for _, c := range coverage {
		for _, s := range c.RoleSubstitutions {
			if s.Role != "event_identity" {
				keysSet[s.ReferenceID] = true
			}
		}
	}
	keys := make([]string, 0, len(keysSet))
	for k := range keysSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		entityIDs[k] = meaning.NodeID(fmt.Sprintf("dse_ref_%d", i))
		referentIDs[k] = meaning.ReferentID(fmt.Sprintf("ref_%d", i))
	}
	concepts := map[string]meaning.ConceptRef{}
	nodes := []meaning.Node{}
	seen := map[meaning.NodeID]bool{}
	for i, p := range parts {
		if p == nil || p.Meaning == nil || len(p.Contexts) != 1 {
			return meaning.Meaning{}, meaning.RequestContext{}, events, entityIDs, fmt.Errorf("clause candidate lacks one context")
		}
		events[i] = meaning.NodeID(fmt.Sprintf("dse_event_%d", i))
		local := map[meaning.NodeID]meaning.NodeID{}
		for _, n := range p.Meaning.Nodes {
			if n.Kind == meaning.KindReferent {
				ref := ""
				for _, s := range coverage[i].RoleSubstitutions {
					if s.MeaningNodeID == n.ID && s.Role != "event_identity" {
						ref = s.ReferenceID
						break
					}
				}
				if ref == "" || entityIDs[ref] == "" {
					return meaning.Meaning{}, meaning.RequestContext{}, events, entityIDs, fmt.Errorf("referent lacks bounded source identity mapping")
				}
				local[n.ID] = entityIDs[ref]
			}
		}
		for _, n := range p.Meaning.Nodes {
			if n.Kind == meaning.KindDiscourse {
				continue
			}
			if n.Kind == meaning.KindEvent {
				n.Participants = append([]meaning.Participant(nil), n.Participants...)
				n.ID = events[i]
				for j := range n.Participants {
					if id, ok := local[n.Participants[j].Target]; ok {
						n.Participants[j].Target = id
					}
				}
			}
			if n.Kind == meaning.KindReferent {
				old := n.ID
				ref := ""
				for _, s := range coverage[i].RoleSubstitutions {
					if s.MeaningNodeID == old && s.Role != "event_identity" {
						ref = s.ReferenceID
						break
					}
				}
				n.ID = local[old]
				if prior, ok := concepts[ref]; ok && prior != n.Concept {
					return meaning.Meaning{}, meaning.RequestContext{}, events, entityIDs, fmt.Errorf("shared reference has conflicting concepts")
				}
				concepts[ref] = n.Concept
			}
			if !seen[n.ID] {
				seen[n.ID] = true
				nodes = append(nodes, n)
			}
		}
	}
	d, err := meaning.NewDiscourse("dse_discourse_0", meaning.ActStatement, []meaning.NodeID{events[0], events[1]})
	if err != nil {
		return meaning.Meaning{}, meaning.RequestContext{}, events, entityIDs, err
	}
	nodes = append(nodes, d)
	graph, err := meaning.Build([]meaning.NodeID{d.ID}, nodes, meaning.DefaultBudget())
	if err != nil {
		return meaning.Meaning{}, meaning.RequestContext{}, events, entityIDs, err
	}
	ctx := meaning.RequestContext{ID: realizationContextID(originalContextID), Referents: []meaning.Referent{}, Mentions: []meaning.Mention{}}
	for i, k := range keys {
		concept := concepts[k]
		ctx.Referents = append(ctx.Referents, meaning.Referent{ID: referentIDs[k], Concept: concept})
		ctx.Mentions = append(ctx.Mentions, meaning.Mention{ID: meaning.MentionID(fmt.Sprintf("ir_ref_%d", i)), NodeID: entityIDs[k], ReferentID: referentIDs[k], EvidenceID: algebraReferenceEvidenceID(k)})
	}
	return graph, ctx, events, entityIDs, nil
}
