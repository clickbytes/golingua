package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

// TargetPlanBuilder supplies explicit feature constraints, assumptions and
// losses for one selector choice. Meaning, target candidate and CORE are set
// by the pipeline and cannot be replaced by this callback.
type TargetPlanBuilder interface {
	Annotations(context.Context, semantic.MeaningCandidate, realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error)
}

// TargetPlanAnnotations is one caller-declared annotation choice for an exact
// selected target candidate. The candidate, Meaning, and construction license
// are shared and cannot be replaced by this choice.
type TargetPlanAnnotations struct {
	FeatureConstraints []realizer.FeatureConstraint
	Assumptions        []realizer.FieldAssumption
	Losses             []realizer.FieldLoss
}

// TargetPlanAlternativesBuilder is an optional bounded extension for callers
// whose one selected candidate supports multiple explicit annotation choices.
// It must return at least one choice and no more than engine.MaxTranslationAlternatives.
// Builders that do not implement it retain the single-result TargetPlanBuilder
// behavior unchanged.
type TargetPlanAlternativesBuilder interface {
	TargetPlanBuilder
	AnnotationAlternatives(context.Context, semantic.MeaningCandidate, realizer.TargetCandidate) ([]TargetPlanAnnotations, error)
}

// TargetExecutionBuilder supplies target-language slot morphology inputs for
// one canonical plan. The callback decides feature values and principal parts;
// it may not choose another semantic or selector candidate.
type TargetExecutionBuilder interface {
	Build(context.Context, semantic.MeaningCandidate, []realizer.TargetCandidate, realizer.TargetCandidate, realizer.RealizationPlan) (realizer.ExecuteInput, error)
}

// TargetOperatorBuilder returns only operator evidence licensed for this exact
// Meaning candidate. The completeness bit must be false when coverage is not
// established; the pipeline never reuses request evidence across candidates.
type TargetOperatorBuilder interface {
	Build(context.Context, semantic.MeaningCandidate) ([]construction.Operator, bool, error)
}

// TargetPipeline adapts selection, planning, morphology execution and
// linearization to engine.TargetRealizer. Slot feature planning is language
// local and therefore required as an explicit compiled dependency.
type TargetPipeline struct {
	Language             features.LanguageID
	LanguageRegistry     *language.Registry
	Lexicon              *lexicon.Index
	Valency              *valency.Index
	Constructions        *construction.Registry
	Variants             map[ids.ConstructionID]*construction.VariantIndex
	Core                 *core.ConceptGraph
	OperatorBuilder      TargetOperatorBuilder
	SelectionBudget      realizer.Budget
	PlanBudget           realizer.PlanBudget
	ExecuteBudget        realizer.ExecuteBudget
	LinearizeBudget      realizer.LinearizeBudget
	PlanBuilder          TargetPlanBuilder
	ExecutionBuilder     TargetExecutionBuilder
	CandidatePolicy      realizer.TargetCandidatePolicy
	ClauseCoordination   *construction.ClauseCoordinationPlan
	ScopedRealizer       engine.ScopedTargetRealizer
	RelationPathRealizer engine.RelationPathTargetRealizer
}

var _ engine.TargetRealizer = (*TargetPipeline)(nil)
var _ engine.ScopedTargetRealizer = (*TargetPipeline)(nil)
var _ engine.ScopedTargetVerifier = (*TargetPipeline)(nil)
var _ engine.RelationPathTargetRealizer = (*TargetPipeline)(nil)
var _ engine.RelationPathTargetVerifier = (*TargetPipeline)(nil)

func (pipeline *TargetPipeline) RealizeRelationPath(ctx context.Context, request engine.RelationPathRequest) (engine.RelationPathResult, error) {
	if pipeline == nil || pipeline.RelationPathRealizer == nil {
		return engine.RelationPathResult{Status: status.OutcomeUnsupported, Diagnostics: []status.Diagnostic{{Code: "RELATION_PATH_TARGET_UNAVAILABLE", Path: "target.relation_path", Message: "target pipeline has no relation path realizer", Severity: status.SeverityError}}}, nil
	}
	return pipeline.RelationPathRealizer.RealizeRelationPath(ctx, request)
}

func (pipeline *TargetPipeline) RelationPathPlan() (engine.RelationPathPlan, error) {
	if pipeline == nil || pipeline.RelationPathRealizer == nil {
		return engine.RelationPathPlan{}, targetPipelineError(status.OutcomeUnsupported, "RELATION_PATH_TARGET_UNAVAILABLE", "target.relation_path", "target pipeline has no relation path realizer")
	}
	verifier, ok := pipeline.RelationPathRealizer.(engine.RelationPathTargetVerifier)
	if !ok {
		return engine.RelationPathPlan{}, targetPipelineError(status.OutcomeUnsupported, "RELATION_PATH_VERIFICATION_UNAVAILABLE", "target.relation_path", "target relation path realizer does not expose a verified plan")
	}
	return verifier.RelationPathPlan()
}

func (pipeline *TargetPipeline) VerifyRelationPath(ctx context.Context, request engine.RelationPathRequest, alternative engine.RelationPathAlternative) error {
	if pipeline == nil || pipeline.RelationPathRealizer == nil {
		return targetPipelineError(status.OutcomeUnsupported, "RELATION_PATH_TARGET_UNAVAILABLE", "target.relation_path", "target pipeline has no relation path realizer")
	}
	verifier, ok := pipeline.RelationPathRealizer.(engine.RelationPathTargetVerifier)
	if !ok {
		return targetPipelineError(status.OutcomeUnsupported, "RELATION_PATH_VERIFICATION_UNAVAILABLE", "target.relation_path", "target pipeline cannot independently verify a relation path alternative")
	}
	return verifier.VerifyRelationPath(ctx, request, alternative)
}

// RealizeScope delegates quantified scope realization to an optional
// language-specific implementation. Ordinary Meaning realization remains
// independent of this capability.
func (pipeline *TargetPipeline) RealizeScope(ctx context.Context, request engine.ScopedTargetRequest) (engine.ScopedTargetResult, error) {
	if pipeline == nil || pipeline.ScopedRealizer == nil {
		return engine.ScopedTargetResult{}, targetPipelineError(status.OutcomeUnsupported, "SCOPED_TARGET_UNAVAILABLE", "target.scope", "target runtime has no scope-preserving realization capability")
	}
	return pipeline.ScopedRealizer.RealizeScope(ctx, request)
}

// ScopedPlan returns an independently decoded copy of the active scope plan.
func (pipeline *TargetPipeline) ScopedPlan() (construction.ScopedClausePlan, error) {
	if pipeline == nil || pipeline.ScopedRealizer == nil {
		return construction.ScopedClausePlan{}, targetPipelineError(status.OutcomeUnsupported, "SCOPED_TARGET_UNAVAILABLE", "target.scope", "target runtime has no scope-preserving realization capability")
	}
	verifier, ok := pipeline.ScopedRealizer.(engine.ScopedTargetVerifier)
	if !ok {
		return construction.ScopedClausePlan{}, targetPipelineError(status.OutcomeUnsupported, "SCOPED_TARGET_VERIFICATION_UNAVAILABLE", "target.scope", "target runtime does not expose a verified scope plan")
	}
	return verifier.ScopedPlan()
}

// VerifyScope delegates proof of a candidate to the exact configured scope
// realizer. The pipeline does not accept unverifiable target alternatives.
func (pipeline *TargetPipeline) VerifyScope(ctx context.Context, request engine.ScopedTargetRequest, alternative engine.ScopedTargetAlternative) error {
	if pipeline == nil || pipeline.ScopedRealizer == nil {
		return targetPipelineError(status.OutcomeUnsupported, "SCOPED_TARGET_UNAVAILABLE", "target.scope", "target runtime has no scope-preserving realization capability")
	}
	verifier, ok := pipeline.ScopedRealizer.(engine.ScopedTargetVerifier)
	if !ok {
		return targetPipelineError(status.OutcomeUnsupported, "SCOPED_TARGET_VERIFICATION_UNAVAILABLE", "target.scope", "target runtime cannot independently verify a scope alternative")
	}
	return verifier.VerifyScope(ctx, request, alternative)
}

// Realize retains each licensed target candidate and returns its exact
// binding evidence with the corresponding linearized surface form.
func (pipeline *TargetPipeline) Realize(ctx context.Context, source semantic.MeaningCandidate) (engine.SurfaceResult, error) {
	if pipeline == nil || pipeline.LanguageRegistry == nil || pipeline.Lexicon == nil || pipeline.Valency == nil || pipeline.Constructions == nil || pipeline.Core == nil || pipeline.ExecutionBuilder == nil {
		return engine.SurfaceResult{}, targetPipelineError(status.OutcomeUnsupported, "TARGET_PIPELINE_UNAVAILABLE", "target_pipeline", "target pipeline requires verified lexical, valency, construction, CORE, language binding and slot feature planning dependencies")
	}
	if ctx == nil {
		return engine.SurfaceResult{}, targetPipelineError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "target realization requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return engine.SurfaceResult{}, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
	}
	if coordinated, matched, coordErr := pipeline.realizeCoordinated(ctx, source); matched {
		return coordinated, coordErr
	}
	var operators []construction.Operator
	operatorCoverageComplete := source.Meaning != nil && meaningHasNoExplicitOperators(*source.Meaning)
	var err error
	if pipeline.OperatorBuilder != nil {
		operators, operatorCoverageComplete, err = pipeline.OperatorBuilder.Build(ctx, source)
		if err != nil {
			return engine.SurfaceResult{}, err
		}
		if err := ctx.Err(); err != nil {
			return engine.SurfaceResult{}, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
		}
	}
	selectionInput := realizer.Input{
		Candidates: []semantic.MeaningCandidate{source}, Outcome: source.Status,
		Diagnostics: source.Evidence.Diagnostics, Operators: operators,
		CandidatePolicy:          pipeline.CandidatePolicy,
		OperatorCoverageComplete: operatorCoverageComplete,
		Language:                 lexicon.LanguageID(pipeline.Language), Lexicon: pipeline.Lexicon,
		Valency: pipeline.Valency, Constructions: pipeline.Constructions,
		Variants: pipeline.Variants, Core: pipeline.Core, Budget: pipeline.SelectionBudget,
	}
	selection, err := realizer.SelectTargets(selectionInput)
	if err != nil {
		return engine.SurfaceResult{}, err
	}
	result := engine.SurfaceResult{Outcome: selection.Outcome, Diagnostics: append([]status.Diagnostic(nil), selection.Diagnostics...)}
	if !operatorCoverageComplete {
		result.Outcome = status.OutcomeUnsupported
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{
			Code: "TARGET_OPERATOR_COVERAGE_UNAVAILABLE", Path: "target.operators",
			Message:  "the target runtime has no complete licensed realization for the Meaning operator scope",
			Severity: status.SeverityError,
		})
		return result, nil
	}
	if source.Meaning == nil {
		return engine.SurfaceResult{}, targetPipelineError(status.OutcomeUnsupported, "TARGET_MEANING_UNAVAILABLE", "source.meaning", "target realization requires a supported source Meaning graph")
	}
	// Referent selections supply the predicate's role slots. They are not
	// independent sentence alternatives, and must remain available to the
	// language-local execution builder when it lexicalizes those slots.
	// Surface candidates are the top-level predicates or noun phrases. Role
	// referents are passed to slot execution as supporting candidates, but are
	// never realized as independent alternatives for a sentence response.
	surfaceIDs := topLevelTargetNodeIDs(*source.Meaning, source.QuestionAnswer != nil)
	neededIDs := targetExecutionNodeIDs(*source.Meaning, surfaceIDs)
	candidates := make([]realizer.TargetCandidate, 0, len(selection.Candidates))
	surfaceCandidates := make([]realizer.TargetCandidate, 0, len(selection.Candidates))
	coverageRejected := 0
	for _, candidate := range selection.Candidates {
		if neededIDs[candidate.NodeID] && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState || candidate.Kind == meaning.KindReferent) {
			candidates = append(candidates, candidate)
		}
		if surfaceIDs[candidate.NodeID] && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState || candidate.Kind == meaning.KindReferent) {
			if realizer.CoversLexicalMeaning(*source.Meaning, candidate.NodeID, candidate.Family.Family) {
				surfaceCandidates = append(surfaceCandidates, candidate)
			} else {
				coverageRejected++
			}
		}
	}
	if len(surfaceCandidates) > engine.MaxTranslationAlternatives {
		surfaceCandidates = surfaceCandidates[:engine.MaxTranslationAlternatives]
		result.Outcome = status.OutcomeResourceLimit
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{
			Code: "RESOURCE_LIMIT", Path: "target.candidates",
			Message:  "target candidates were truncated at the engine alternative limit",
			Severity: status.SeverityWarning,
		})
	}
	if len(surfaceCandidates) == 0 && coverageRejected > 0 {
		result.Outcome = combineTargetOutcomes(result.Outcome, status.OutcomeUnsupported)
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{
			Code: "TARGET_LEXICAL_COVERAGE_UNAVAILABLE", Path: "target.candidates",
			Message:  "no selected construction covers every lexical node in the supplied Meaning",
			Severity: status.SeverityWarning,
		})
	}
	if len(surfaceCandidates) == 0 && len(result.Diagnostics) == 0 {
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{
			Code: "TARGET_NO_SURFACE_CANDIDATE", Path: "target.candidates",
			Message:  "no licensed surface candidate for the supplied Meaning, context and evidence",
			Severity: status.SeverityWarning,
		})
	}
	seenPlanIDs := make(map[string]struct{})
surfaceCandidatesLoop:
	for _, selected := range surfaceCandidates {
		if err := ctx.Err(); err != nil {
			retainTargetCandidateFailure(&result, ctx.Err())
			return result, nil
		}
		annotationSets := []TargetPlanAnnotations{{}}
		if alternativesBuilder, ok := pipeline.PlanBuilder.(TargetPlanAlternativesBuilder); ok {
			annotationSets, err = alternativesBuilder.AnnotationAlternatives(ctx, source, selected)
			if err != nil {
				if !retainTargetCandidateFailure(&result, err) {
					return engine.SurfaceResult{}, err
				}
				if result.Outcome == status.OutcomeResourceLimit {
					break surfaceCandidatesLoop
				}
				continue surfaceCandidatesLoop
			}
			if len(annotationSets) == 0 {
				return engine.SurfaceResult{}, targetPipelineError(status.OutcomeInvalid, "EMPTY_TARGET_PLAN_ALTERNATIVES", "target.plan_alternatives", "an alternatives builder must retain at least one annotation choice for each selected target candidate")
			}
			if len(annotationSets) > engine.MaxTranslationAlternatives {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{Code: "RESOURCE_LIMIT", Path: "target.plan_alternatives", Message: "annotation choices exceed the engine translation alternative limit", Severity: status.SeverityWarning})
				break surfaceCandidatesLoop
			}
		} else if pipeline.PlanBuilder != nil {
			constraints, assumptions, losses, annotationErr := pipeline.PlanBuilder.Annotations(ctx, source, selected)
			if annotationErr != nil {
				if !retainTargetCandidateFailure(&result, annotationErr) {
					return engine.SurfaceResult{}, annotationErr
				}
				if result.Outcome == status.OutcomeResourceLimit {
					break surfaceCandidatesLoop
				}
				continue surfaceCandidatesLoop
			}
			annotationSets[0] = TargetPlanAnnotations{FeatureConstraints: constraints, Assumptions: assumptions, Losses: losses}
		}
		for _, annotations := range annotationSets {
			if err := ctx.Err(); err != nil {
				retainTargetCandidateFailure(&result, ctx.Err())
				return result, nil
			}
			if len(result.Alternatives) >= engine.MaxTranslationAlternatives {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{Code: "RESOURCE_LIMIT", Path: "target.alternatives", Message: "target realization reached the engine translation alternative limit", Severity: status.SeverityWarning})
				break surfaceCandidatesLoop
			}
			planInput := realizer.PlanInput{Meaning: *source.Meaning, Candidate: selected, Core: pipeline.Core, Budget: pipeline.PlanBudget,
				FeatureConstraints: annotations.FeatureConstraints, Assumptions: annotations.Assumptions, Losses: annotations.Losses}
			// Polish pluralia tantum can realize a referent with singular or
			// plural cardinality using the same lexical plural form. Preserve that
			// loss on the referent only when the selected slot candidate actually
			// uses a lexeme declaring this behavior.
			if pipeline.Language == "pl" && source.Meaning != nil {
				for _, slot := range selected.Family.Family.SlotSpecs {
					if slot.ArgumentBinding == nil || slot.Category != construction.CategoryNP {
						continue
					}
					referentID, found := targetRoleReferent(*source.Meaning, selected.NodeID, slot.ArgumentBinding.SemanticRole)
					if !found {
						continue
					}
					referent := targetMeaningNode(*source.Meaning, referentID)
					var matching []realizer.TargetCandidate
					for _, lexicalCandidate := range candidates {
						if lexicalCandidate.NodeID != referentID || lexicalCandidate.Kind != meaning.KindReferent {
							continue
						}
						matching = append(matching, lexicalCandidate)
					}
					// A single selected lexical route is unambiguous here. If target
					// lexemes compete, leave their candidate-local number behavior to
					// slot execution rather than attributing one route's loss to all.
					if len(matching) == 1 && hasLexicalPluralWithoutCardinality(matching[0].Lexeme.Properties) {
						if path, ok := MeaningPath(*source.Meaning, referentID); ok {
							if referent.Multiplicity == "" {
								predicate := targetMeaningNode(*source.Meaning, selected.NodeID)
								if predicate.Kind == meaning.KindState && len(predicate.Participants) == 1 {
									predicatePath, _ := MeaningPath(*source.Meaning, selected.NodeID)
									planInput.Assumptions = removeNumberAssumptionsAt(planInput.Assumptions, path, predicatePath)
								}
							} else if referent.Multiplicity == meaning.MultiplicityOne || referent.Multiplicity == meaning.MultiplicityMany {
								planInput.Losses = append(planInput.Losses, realizer.FieldLoss{Path: path, Field: "number", Outcome: status.OutcomeIncomplete, Reason: "Polish pluralia tantum marks lexical grammatical number but does not distinguish source referent cardinality"})
							}
						}
					}
				}
			}
			if err := ctx.Err(); err != nil {
				retainTargetCandidateFailure(&result, ctx.Err())
				return result, nil
			}
			plan, err := realizer.BuildPlan(planInput)
			if err != nil {
				if !retainTargetCandidateFailure(&result, err) {
					return engine.SurfaceResult{}, err
				}
				if result.Outcome == status.OutcomeResourceLimit {
					break surfaceCandidatesLoop
				}
				continue
			}
			if _, duplicate := seenPlanIDs[plan.ID]; duplicate {
				return engine.SurfaceResult{}, targetPipelineError(status.OutcomeInvalid, "DUPLICATE_TARGET_PLAN_ALTERNATIVE", "target.plan_alternatives", "annotation choices produced the same canonical realization plan ID")
			}
			seenPlanIDs[plan.ID] = struct{}{}
			// A frame-licensed absent role makes the complete target candidate
			// incomplete, while its predicate lexicalization remains usable. Keep
			// that status on the plan/result, and let the morphology builder handle
			// the exact licensed candidate as a slot request.
			executionCandidate := selected
			if len(selected.MissingOmittedRoles) > 0 && plan.Outcome == status.OutcomeIncomplete {
				executionCandidate.Outcome = status.OutcomeOK
			}
			remainingAlternatives := engine.MaxTranslationAlternatives - len(result.Alternatives)
			executionChoices, limitExceeded := targetRoleExecutionChoices(*source.Meaning, selected, candidates, pipeline.Language, remainingAlternatives)
			if limitExceeded {
				result.Outcome = status.OutcomeResourceLimit
				result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{Code: "RESOURCE_LIMIT", Path: "target.role_lexicalizations", Message: "licensed role lexicalization alternatives exceed the remaining engine translation alternative limit", Severity: status.SeverityWarning})
				break surfaceCandidatesLoop
			}
			for _, executionChoice := range executionChoices {
				if len(result.Alternatives) >= engine.MaxTranslationAlternatives {
					result.Outcome = status.OutcomeResourceLimit
					result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{Code: "RESOURCE_LIMIT", Path: "target.alternatives", Message: "target realization reached the engine translation alternative limit", Severity: status.SeverityWarning})
					break surfaceCandidatesLoop
				}
				executionInput, err := pipeline.ExecutionBuilder.Build(ctx, source, executionChoice.Candidates, executionCandidate, plan)
				if err != nil {
					if !retainTargetCandidateFailure(&result, err) {
						return engine.SurfaceResult{}, err
					}
					if result.Outcome == status.OutcomeResourceLimit {
						break surfaceCandidatesLoop
					}
					continue
				}
				if err := ctx.Err(); err != nil {
					retainTargetCandidateFailure(&result, ctx.Err())
					return result, nil
				}
				// The callback supplies only language-local morphology inputs and
				// information-structure evidence; identity-bearing values come from the
				// selected candidate and canonical plan above.
				executionInput.Plan = plan
				executionInput.SourceMeaning = *source.Meaning
				executionInput.Core = pipeline.Core
				if executionInput.Budget == (realizer.ExecuteBudget{}) {
					executionInput.Budget = pipeline.ExecuteBudget
				}
				executed, err := realizer.Execute(executionInput)
				if err != nil {
					if !retainTargetCandidateFailure(&result, err) {
						return engine.SurfaceResult{}, err
					}
					if result.Outcome == status.OutcomeResourceLimit {
						break surfaceCandidatesLoop
					}
					continue
				}
				linearized, err := realizer.LinearizeWithNPProfiles(plan, executed.Family, executed.Slots, executed.NPForms, executionInput.PhraseProfile, executionInput.FeatureProfile, executionInput.InformationStructure, pipeline.LinearizeBudget)
				if err != nil {
					if !retainTargetCandidateFailure(&result, err) {
						return engine.SurfaceResult{}, err
					}
					if result.Outcome == status.OutcomeResourceLimit {
						break surfaceCandidatesLoop
					}
					continue
				}
				evidence, bindingOK := targetBindingEvidence(pipeline.LanguageRegistry, pipeline.Language, source, selected, plan)
				// Selection.Outcome summarizes all retained routes. Keep route-local
				// alternatives usable when that aggregate is ambiguous; result.Outcome
				// already preserves the overall selection status for the caller.
				outcome := combineTargetOutcomes(selected.Outcome, plan.Outcome, executed.Outcome, linearized.Outcome)
				var bindingDiagnostic status.Diagnostic
				if !bindingOK {
					outcome = status.OutcomeUnsupported
					bindingDiagnostic = status.Diagnostic{
						Code: "TARGET_BINDING_EVIDENCE_UNAVAILABLE", Path: "target.binding",
						Message:  "target candidate does not match an approved registry binding for its exact sense, frame, family, variant and CORE concept",
						Severity: status.SeverityError,
					}
					evidence = semantic.EvidenceTrace{AlternativeID: plan.ID, Outcome: outcome, Diagnostics: []status.Diagnostic{bindingDiagnostic}}
				}
				alternativeID := plan.ID
				if len(executionChoice.RoleCandidates) > 0 {
					alternativeID = targetRoleAlternativeID(plan.ID, executionChoice.RoleCandidates)
					if bindingOK {
						if !appendTargetRoleCandidateEvidence(&evidence, executionChoice.RoleCandidates, plan) {
							bindingOK = false
							outcome = status.OutcomeUnsupported
							bindingDiagnostic = status.Diagnostic{Code: "TARGET_ROLE_BINDING_EVIDENCE_UNAVAILABLE", Path: "target.role_lexicalizations", Message: "a selected role lexicalization lacks an exact approved target binding", Severity: status.SeverityError}
							evidence = semantic.EvidenceTrace{AlternativeID: alternativeID, Outcome: outcome, Diagnostics: []status.Diagnostic{bindingDiagnostic}}
						}
					}
					evidence.AlternativeID = alternativeID
				}
				alternative := engine.SurfaceAlternative{
					ID: alternativeID, Outcome: outcome, Text: linearized.Text,
					Evidence: evidence, Assumptions: cloneFieldAssumptions(plan.Assumptions),
					Losses:      append([]realizer.FieldLoss(nil), plan.Losses...),
					Diagnostics: append([]status.Diagnostic(nil), selected.Diagnostics...),
				}
				alternative.Diagnostics = appendTargetDiagnostics(alternative.Diagnostics, plan.Diagnostics...)
				alternative.Diagnostics = appendTargetDiagnostics(alternative.Diagnostics, executed.Diagnostics...)
				alternative.Diagnostics = appendTargetDiagnostics(alternative.Diagnostics, linearized.Diagnostics...)
				if !bindingOK {
					alternative.Outcome = outcome
					alternative.Text = ""
					alternative.Diagnostics = appendTargetDiagnostics(alternative.Diagnostics, bindingDiagnostic)
				}
				if alternative.Outcome != status.OutcomeOK {
					if len(alternative.Diagnostics) == 0 {
						alternative.Diagnostics = appendTargetDiagnostics(alternative.Diagnostics, status.Diagnostic{
							Code: "TARGET_REALIZATION_INCOMPLETE", Path: "target",
							Message:  "target realization did not produce a complete supported surface alternative",
							Severity: status.SeverityWarning,
						})
					}
					alternative.Evidence.Outcome = alternative.Outcome
					alternative.Evidence.Diagnostics = append([]status.Diagnostic(nil), alternative.Diagnostics...)
				}
				result.Outcome = combineTargetOutcomes(result.Outcome, alternative.Outcome)
				result.Alternatives = append(result.Alternatives, alternative)
			}
		}
	}
	if len(result.Alternatives) == 0 && result.Outcome == status.OutcomeOK {
		result.Outcome = status.OutcomeUnknown
	}
	return result, nil
}

type targetRoleCandidateSelection struct {
	SlotID    string
	Role      meaning.Role
	NodeID    meaning.NodeID
	Candidate realizer.TargetCandidate
}

type targetRoleExecutionChoice struct {
	Candidates     []realizer.TargetCandidate
	RoleCandidates []targetRoleCandidateSelection
}

type targetRoleCandidateGroup struct {
	NodeID  meaning.NodeID
	Choices []realizer.TargetCandidate
	Slots   []struct {
		ID   string
		Role meaning.Role
	}
}

// targetRoleExecutionChoices expands only exact, already-selected lexical
// candidates for NP argument slots. It preserves the predicate plan and limits
// the Cartesian product before building any execution inputs.
func targetRoleExecutionChoices(
	graph meaning.Meaning,
	selected realizer.TargetCandidate,
	candidates []realizer.TargetCandidate,
	languageID features.LanguageID,
	limit int,
) ([]targetRoleExecutionChoice, bool) {
	if limit <= 0 {
		return nil, true
	}
	predicate := targetMeaningNode(graph, selected.NodeID)
	if predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return []targetRoleExecutionChoice{{Candidates: candidates}}, false
	}
	groups := make([]targetRoleCandidateGroup, 0)
	groupByNode := make(map[meaning.NodeID]int)
	for _, slot := range selected.Family.Family.SlotSpecs {
		if slot.Category != construction.CategoryNP || slot.ArgumentBinding == nil {
			continue
		}
		var referentID meaning.NodeID
		matches := 0
		for _, participant := range predicate.Participants {
			if participant.Role == slot.ArgumentBinding.SemanticRole && targetMeaningNode(graph, participant.Target).Kind == meaning.KindReferent {
				referentID = participant.Target
				matches++
			}
		}
		if matches != 1 {
			// Keep existing behavior for missing or genuinely ambiguous semantic
			// role links; candidate fan-out cannot repair those cases.
			return []targetRoleExecutionChoice{{Candidates: candidates}}, false
		}
		if groupIndex, exists := groupByNode[referentID]; exists {
			groups[groupIndex].Slots = append(groups[groupIndex].Slots, struct {
				ID   string
				Role meaning.Role
			}{slot.ID, slot.ArgumentBinding.SemanticRole})
			continue
		}
		groupByNode[referentID] = len(groups)
		groups = append(groups, targetRoleCandidateGroup{NodeID: referentID, Slots: []struct {
			ID   string
			Role meaning.Role
		}{{slot.ID, slot.ArgumentBinding.SemanticRole}}})
	}
	if len(groups) == 0 {
		return []targetRoleExecutionChoice{{Candidates: candidates}}, false
	}
	product := 1
	for groupIndex := range groups {
		group := &groups[groupIndex]
		concept := string(targetMeaningNode(graph, group.NodeID).Concept)
		unique := make(map[string]realizer.TargetCandidate)
		for _, candidate := range candidates {
			if !targetRoleCandidateAuthorized(candidate, selected.SourceCandidateID, selected.SemanticID, group.NodeID, concept, languageID) {
				continue
			}
			key := targetRoleCandidateKey(candidate)
			unique[key] = candidate
		}
		for _, candidate := range unique {
			group.Choices = append(group.Choices, candidate)
		}
		sort.Slice(group.Choices, func(i, j int) bool {
			return targetRoleCandidateKey(group.Choices[i]) < targetRoleCandidateKey(group.Choices[j])
		})
		if len(group.Choices) == 0 {
			return []targetRoleExecutionChoice{{Candidates: candidates}}, false
		}
		behavior := targetRoleNPBehavior(group.Choices[0])
		for _, candidate := range group.Choices[1:] {
			if targetRoleNPBehavior(candidate) != behavior {
				// The plan annotations are shared across lexical fan-out. If
				// candidates differ in any declared NP realization behavior, retain
				// the existing ambiguity rather than copying one candidate's article
				// or number assumptions to another.
				return []targetRoleExecutionChoice{{Candidates: candidates}}, false
			}
		}
		if product > limit/len(group.Choices) {
			return nil, true
		}
		product *= len(group.Choices)
	}
	if product == 1 {
		return []targetRoleExecutionChoice{{Candidates: candidates}}, false
	}
	choices := make([]targetRoleExecutionChoice, 0, product)
	selectedChoices := make([]realizer.TargetCandidate, len(groups))
	var expand func(int)
	expand = func(index int) {
		if index < len(groups) {
			for _, candidate := range groups[index].Choices {
				selectedChoices[index] = candidate
				expand(index + 1)
			}
			return
		}
		choiceCandidates := make([]realizer.TargetCandidate, 0, len(candidates))
		for _, candidate := range candidates {
			include := true
			for groupIndex, group := range groups {
				if candidate.NodeID == group.NodeID && candidate.Kind == meaning.KindReferent && targetRoleCandidateKey(candidate) != targetRoleCandidateKey(selectedChoices[groupIndex]) {
					include = false
					break
				}
			}
			if include {
				choiceCandidates = append(choiceCandidates, candidate)
			}
		}
		roleSelections := make([]targetRoleCandidateSelection, 0)
		for groupIndex, group := range groups {
			for _, slot := range group.Slots {
				roleSelections = append(roleSelections, targetRoleCandidateSelection{SlotID: slot.ID, Role: slot.Role, NodeID: group.NodeID, Candidate: selectedChoices[groupIndex]})
			}
		}
		choices = append(choices, targetRoleExecutionChoice{Candidates: choiceCandidates, RoleCandidates: roleSelections})
	}
	expand(0)
	return choices, false
}

func targetRoleCandidateAuthorized(candidate realizer.TargetCandidate, sourceID, semanticID string, nodeID meaning.NodeID, concept string, languageID features.LanguageID) bool {
	if candidate.SourceCandidateID != sourceID || candidate.SemanticID != semanticID || candidate.NodeID != nodeID ||
		candidate.Kind != meaning.KindReferent || candidate.Outcome != status.OutcomeOK || candidate.TransferBinding.CoreRef.String() != concept ||
		!candidate.TransferBinding.TransferAuthorized() || string(candidate.Lexeme.Language) != string(languageID) ||
		candidate.Lexeme.POS != lexicon.PartOfSpeech("noun") || candidate.Sense.LexemeID != candidate.Lexeme.ID ||
		candidate.Sense.ID != candidate.TransferBinding.SenseID {
		return false
	}
	return true
}

func targetRoleCandidateKey(candidate realizer.TargetCandidate) string {
	return candidate.Lexeme.ID.String() + "\x00" + candidate.Sense.ID.String() + "\x00" + candidate.TransferBinding.CoreRef.String()
}

func targetRoleNPBehavior(candidate realizer.TargetCandidate) string {
	properties := make([]string, 0, len(candidate.Lexeme.Properties))
	for _, property := range candidate.Lexeme.Properties {
		properties = append(properties, property.Key+"="+property.Value)
	}
	sort.Strings(properties)
	return string(candidate.Lexeme.ParadigmID) + "\x00" + fmt.Sprint(properties)
}

func targetRoleAlternativeID(planID string, selections []targetRoleCandidateSelection) string {
	hash := sha256.New()
	_, _ = fmt.Fprintf(hash, "%s", planID)
	for _, selection := range selections {
		_, _ = fmt.Fprintf(hash, "\x00%s\x00%s\x00%s\x00%s", selection.SlotID, selection.Role, selection.NodeID, targetRoleCandidateKey(selection.Candidate))
	}
	return planID + "-role-" + hex.EncodeToString(hash.Sum(nil)[:8])
}

func appendTargetRoleCandidateEvidence(trace *semantic.EvidenceTrace, selections []targetRoleCandidateSelection, plan realizer.RealizationPlan) bool {
	if trace == nil {
		return false
	}
	seen := make(map[string]struct{})
	for _, selection := range selections {
		candidate := selection.Candidate
		key := targetRoleCandidateKey(candidate)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		if !candidate.TransferBinding.TransferAuthorized() || candidate.TransferBinding.SenseID != candidate.Sense.ID ||
			candidate.TransferBinding.CoreRef.String() == "" {
			return false
		}
		trace.Records = append(trace.Records, semantic.EvidenceRecord{
			Origin: semantic.EvidenceOriginGenerated, SenseID: candidate.TransferBinding.SenseID, CoreConceptID: candidate.TransferBinding.CoreRef,
			BindingKey: semantic.BindingKeyFor(candidate.TransferBinding.SenseID, candidate.TransferBinding.CoreRef), PlanID: plan.FamilyPlanID,
			FamilyID: plan.FamilyID, VariantID: plan.VariantID,
			FrameID: plan.FrameID, Outcome: status.OutcomeOK, Provenance: candidate.TransferBinding.Provenance,
		})
	}
	return trace.Validate() == nil
}

func cloneFieldAssumptions(values []realizer.FieldAssumption) []realizer.FieldAssumption {
	result := make([]realizer.FieldAssumption, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Path = append([]meaning.NodeID(nil), value.Path...)
	}
	return result
}

func meaningHasNoExplicitOperators(graph meaning.Meaning) bool {
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindScope, meaning.KindModality, meaning.KindQuantifier, meaning.KindQuestion:
			return false
		}
	}
	return true
}

func hasLexicalPluralWithoutCardinality(properties []lexicon.Property) bool {
	for _, property := range properties {
		if property.Key == "multiplicity_behavior" && property.Value == "lexical_plural_not_cardinality" {
			return true
		}
	}
	return false
}

func removeNumberAssumptionsAt(values []realizer.FieldAssumption, paths ...[]meaning.NodeID) []realizer.FieldAssumption {
	result := make([]realizer.FieldAssumption, 0, len(values))
	for _, value := range values {
		remove := false
		for _, path := range paths {
			if value.Field != "number" || len(path) == 0 || len(value.Path) != len(path) {
				continue
			}
			match := true
			for index := range path {
				if path[index] != value.Path[index] {
					match = false
					break
				}
			}
			remove = remove || match
		}
		if !remove {
			result = append(result, value)
		}
	}
	return result
}

func targetMeaningNode(graph meaning.Meaning, id meaning.NodeID) meaning.Node {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	return meaning.Node{}
}

func targetRoleReferent(graph meaning.Meaning, predicateID meaning.NodeID, role meaning.Role) (meaning.NodeID, bool) {
	for _, node := range graph.Nodes {
		if node.ID != predicateID || node.Kind != meaning.KindEvent && node.Kind != meaning.KindState {
			continue
		}
		for _, participant := range node.Participants {
			if participant.Role == role && targetMeaningNode(graph, participant.Target).Kind == meaning.KindReferent {
				return participant.Target, true
			}
		}
	}
	return "", false
}

func targetExecutionNodeIDs(graph meaning.Meaning, surfaceIDs map[meaning.NodeID]bool) map[meaning.NodeID]bool {
	result := make(map[meaning.NodeID]bool, len(surfaceIDs))
	for id := range surfaceIDs {
		result[id] = true
	}
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	for id := range surfaceIDs {
		predicate, exists := nodes[id]
		if !exists || predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
			continue
		}
		for _, participant := range predicate.Participants {
			if target, exists := nodes[participant.Target]; exists && target.Kind == meaning.KindReferent {
				result[target.ID] = true
			}
		}
	}
	return result
}

func topLevelTargetNodeIDs(graph meaning.Meaning, allowTypedQuestion bool) map[meaning.NodeID]bool {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	result := make(map[meaning.NodeID]bool)
	var visit func(meaning.NodeID, int)
	visit = func(id meaning.NodeID, depth int) {
		if depth >= meaning.MaxMeaningDepth {
			return
		}
		node, exists := nodes[id]
		if !exists {
			return
		}
		switch node.Kind {
		case meaning.KindDiscourse:
			if (node.Act == meaning.ActStatement || allowTypedQuestion && node.Act == meaning.ActQuestion) && len(node.Content) == 1 {
				visit(node.Content[0], depth+1)
			}
		case meaning.KindQuestion:
			if allowTypedQuestion {
				visit(node.Target, depth+1)
			}
		case meaning.KindScope:
			if node.Operator == meaning.OperatorNegation {
				visit(node.Target, depth+1)
			}
		case meaning.KindEvent, meaning.KindState, meaning.KindReferent:
			result[node.ID] = true
		}
	}
	for _, root := range graph.Roots {
		visit(root, 0)
	}
	return result
}

func targetBindingEvidence(registry *language.Registry, languageID features.LanguageID, source semantic.MeaningCandidate, candidate realizer.TargetCandidate, plan realizer.RealizationPlan) (semantic.EvidenceTrace, bool) {
	registered, ok := registry.Language(languageID)
	if !ok || source.Meaning == nil {
		return semantic.EvidenceTrace{}, false
	}
	var binding *language.Binding
	for index := range registered.Bindings {
		item := &registered.Bindings[index]
		if item.SenseID == candidate.Sense.ID && item.ValencyFrameID == string(candidate.Frame.ID) && item.FamilyID == candidate.Family.Family.ID && item.CoreRef == candidate.TransferBinding.CoreRef && item.Mapping == lexicon.MappingEquivalent && item.Approved && item.Realize && hasVariantString(item.VariantIDs, candidate.CompiledVariant.Variant.ID.String()) {
			binding = item
			break
		}
	}
	if binding == nil || plan.Language != languageID || plan.SenseID != binding.SenseID || plan.FrameID != valency.FrameID(binding.ValencyFrameID) || plan.FamilyID != binding.FamilyID || plan.VariantID.String() == "" || !hasVariantString(binding.VariantIDs, plan.VariantID.String()) || plan.CoreConceptID != binding.CoreRef {
		return semantic.EvidenceTrace{}, false
	}
	var sourceSpan = struct{ Start, End int }{}
	foundSpan := false
	for _, record := range source.Evidence.Records {
		if record.Outcome == status.OutcomeOK && record.CoreConceptID == binding.CoreRef && record.SenseID.String() != "" && record.SourceSpan.End > record.SourceSpan.Start {
			sourceSpan.Start, sourceSpan.End = record.SourceSpan.Start, record.SourceSpan.End
			foundSpan = true
			break
		}
	}
	generated := !foundSpan && len(source.Evidence.Records) == 0
	if !foundSpan && !generated {
		return semantic.EvidenceTrace{}, false
	}
	evidenceRecord := semantic.EvidenceRecord{
		SenseID: binding.SenseID, CoreConceptID: binding.CoreRef,
		BindingKey: semantic.BindingKeyFor(binding.SenseID, binding.CoreRef),
		PlanID:     plan.FamilyPlanID, FamilyID: plan.FamilyID, VariantID: plan.VariantID,
		FrameID: plan.FrameID, Outcome: status.OutcomeOK, Provenance: binding.Provenance,
	}
	if generated {
		evidenceRecord.Origin = semantic.EvidenceOriginGenerated
	} else {
		evidenceRecord.SourceSpan.Start, evidenceRecord.SourceSpan.End = sourceSpan.Start, sourceSpan.End
	}
	trace := semantic.EvidenceTrace{
		AlternativeID: plan.ID, Outcome: status.OutcomeOK,
		Records: []semantic.EvidenceRecord{evidenceRecord},
	}
	return trace, trace.Validate() == nil
}

func hasVariantString(variants []ids.ConstructionID, selected string) bool {
	for _, variant := range variants {
		if variant.String() == selected {
			return true
		}
	}
	return false
}

func appendTargetDiagnostics(target []status.Diagnostic, values ...status.Diagnostic) []status.Diagnostic {
	result := make([]status.Diagnostic, 0, len(target)+len(values))
	seen := make(map[status.Diagnostic]struct{}, len(target)+len(values))
	appendUnique := func(items []status.Diagnostic) {
		for _, value := range items {
			if _, exists := seen[value]; exists {
				continue
			}
			if len(result) >= status.MaxDiagnostics {
				return
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	appendUnique(target)
	appendUnique(values)
	return result
}

// retainTargetCandidateFailure keeps a bounded, typed coverage refusal local
// to a candidate. Resource exhaustion preserves already completed alternatives
// but must stop enumeration. Invalid, internal and malformed errors stay fatal;
// an arbitrary Go error is never reclassified as an unsupported translation.
func retainTargetCandidateFailure(result *engine.SurfaceResult, err error) bool {
	var failure status.Result
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		failure = status.Result{Status: status.OutcomeResourceLimit, Diagnostics: []status.Diagnostic{{
			Code: "TARGET_REALIZATION_CANCELED", Path: "context", Message: "target realization was canceled or timed out", Severity: status.SeverityWarning,
		}}}
	} else {
		var typed interface{ Result() status.Result }
		if !errors.As(err, &typed) {
			return false
		}
		failure = typed.Result()
		if failure.Validate() != nil || failure.Text != "" {
			return false
		}
	}
	switch failure.Status {
	case status.OutcomeUnsupported, status.OutcomeUnknown, status.OutcomeIncomplete, status.OutcomeAmbiguous, status.OutcomeResourceLimit:
		result.Outcome = combineTargetOutcomes(result.Outcome, failure.Status)
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, failure.Diagnostics...)
		return true
	default:
		return false
	}
}

func combineTargetOutcomes(outcomes ...status.OutcomeKind) status.OutcomeKind {
	result := status.OutcomeOK
	for _, outcome := range outcomes {
		if outcome == "" || outcome == status.OutcomeOK {
			continue
		}
		if targetOutcomeRank(outcome) > targetOutcomeRank(result) {
			result = outcome
		}
	}
	return result
}

func targetOutcomeRank(outcome status.OutcomeKind) int {
	switch outcome {
	case status.OutcomeInvalid:
		return 100
	case status.OutcomeError:
		return 90
	case status.OutcomeResourceLimit:
		return 80
	case status.OutcomeUnsupported:
		return 70
	case status.OutcomeUnknown:
		return 60
	case status.OutcomeAmbiguous:
		return 50
	case status.OutcomeIncomplete:
		return 40
	case status.OutcomeOK:
		return 0
	default:
		return 100
	}
}

type targetPipelineFailure struct{ result status.Result }

func (failure *targetPipelineFailure) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "target pipeline unavailable"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *targetPipelineFailure) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func targetPipelineError(kind status.OutcomeKind, code, path, message string) error {
	result, err := status.NewResult(kind, "", []status.Diagnostic{{
		Code: code, Path: path, Message: message, Severity: status.SeverityError,
	}}, nil)
	if err != nil {
		return fmt.Errorf("%s: %s", code, message)
	}
	return &targetPipelineFailure{result: result}
}
