package runtime

import (
	"context"
	"fmt"

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
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

// TargetPlanBuilder supplies explicit feature constraints, assumptions and
// losses for one selector choice. Meaning, target candidate and CORE are set
// by the pipeline and cannot be replaced by this callback.
type TargetPlanBuilder interface {
	Annotations(context.Context, semantic.MeaningCandidate, realizer.TargetCandidate) ([]realizer.FeatureConstraint, []realizer.FieldAssumption, []realizer.FieldLoss, error)
}

// TargetExecutionBuilder supplies target-language slot morphology inputs for
// one canonical plan. The callback decides feature values and principal parts;
// it may not choose another semantic or selector candidate.
type TargetExecutionBuilder interface {
	Build(context.Context, semantic.MeaningCandidate, realizer.TargetCandidate, realizer.RealizationPlan) (realizer.ExecuteInput, error)
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
	Language         features.LanguageID
	LanguageRegistry *language.Registry
	Lexicon          *lexicon.Index
	Valency          *valency.Index
	Constructions    *construction.Registry
	Variants         map[ids.ConstructionID]*construction.VariantIndex
	Core             *core.ConceptGraph
	OperatorBuilder  TargetOperatorBuilder
	SelectionBudget  realizer.Budget
	PlanBudget       realizer.PlanBudget
	ExecuteBudget    realizer.ExecuteBudget
	LinearizeBudget  realizer.LinearizeBudget
	PlanBuilder      TargetPlanBuilder
	ExecutionBuilder TargetExecutionBuilder
}

var _ engine.TargetRealizer = (*TargetPipeline)(nil)

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
	var operators []construction.Operator
	operatorCoverageComplete := false
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
	selection, err := realizer.SelectTargets(realizer.Input{
		Candidates: []semantic.MeaningCandidate{source}, Outcome: source.Status,
		Diagnostics: source.Evidence.Diagnostics, Operators: operators,
		OperatorCoverageComplete: operatorCoverageComplete,
		Language:                 lexicon.LanguageID(pipeline.Language), Lexicon: pipeline.Lexicon,
		Valency: pipeline.Valency, Constructions: pipeline.Constructions,
		Variants: pipeline.Variants, Core: pipeline.Core, Budget: pipeline.SelectionBudget,
	})
	if err != nil {
		return engine.SurfaceResult{}, err
	}
	result := engine.SurfaceResult{Outcome: selection.Outcome, Diagnostics: append([]status.Diagnostic(nil), selection.Diagnostics...)}
	candidates := selection.Candidates
	if len(candidates) > engine.MaxTranslationAlternatives {
		candidates = candidates[:engine.MaxTranslationAlternatives]
		result.Outcome = status.OutcomeResourceLimit
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{
			Code: "RESOURCE_LIMIT", Path: "target.candidates",
			Message:  "target candidates were truncated at the engine alternative limit",
			Severity: status.SeverityWarning,
		})
	}
	for _, selected := range candidates {
		if err := ctx.Err(); err != nil {
			return engine.SurfaceResult{}, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
		}
		planInput := realizer.PlanInput{Meaning: meaning.Meaning{}, Candidate: selected, Core: pipeline.Core, Budget: pipeline.PlanBudget}
		if source.Meaning == nil {
			return engine.SurfaceResult{}, targetPipelineError(status.OutcomeUnsupported, "TARGET_MEANING_UNAVAILABLE", "source.meaning", "target realization requires a supported source Meaning graph")
		}
		planInput.Meaning = *source.Meaning
		if pipeline.PlanBuilder != nil {
			planInput.FeatureConstraints, planInput.Assumptions, planInput.Losses, err = pipeline.PlanBuilder.Annotations(ctx, source, selected)
			if err != nil {
				return engine.SurfaceResult{}, err
			}
		}
		if err := ctx.Err(); err != nil {
			return engine.SurfaceResult{}, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
		}
		plan, err := realizer.BuildPlan(planInput)
		if err != nil {
			return engine.SurfaceResult{}, err
		}
		executionInput, err := pipeline.ExecutionBuilder.Build(ctx, source, selected, plan)
		if err != nil {
			return engine.SurfaceResult{}, err
		}
		if err := ctx.Err(); err != nil {
			return engine.SurfaceResult{}, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
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
			return engine.SurfaceResult{}, err
		}
		linearized, err := realizer.Linearize(plan, executed.Family, executed.Slots, executionInput.InformationStructure, pipeline.LinearizeBudget)
		if err != nil {
			return engine.SurfaceResult{}, err
		}
		evidence, bindingOK := targetBindingEvidence(pipeline.LanguageRegistry, pipeline.Language, source, selected, plan)
		outcome := combineTargetOutcomes(selection.Outcome, selected.Outcome, plan.Outcome, executed.Outcome, linearized.Outcome)
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
		alternative := engine.SurfaceAlternative{
			ID: plan.ID, Outcome: outcome, Text: linearized.Text,
			Evidence: evidence, Losses: append([]realizer.FieldLoss(nil), plan.Losses...),
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
	if len(result.Alternatives) == 0 && result.Outcome == status.OutcomeOK {
		result.Outcome = status.OutcomeUnknown
	}
	return result, nil
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
	if !foundSpan {
		return semantic.EvidenceTrace{}, false
	}
	trace := semantic.EvidenceTrace{
		AlternativeID: plan.ID, Outcome: status.OutcomeOK,
		Records: []semantic.EvidenceRecord{{
			SourceSpan: semanticSourceSpan(sourceSpan.Start, sourceSpan.End),
			SenseID:    binding.SenseID, CoreConceptID: binding.CoreRef,
			BindingKey: semantic.BindingKeyFor(binding.SenseID, binding.CoreRef),
			PlanID:     plan.FamilyPlanID, FamilyID: plan.FamilyID, VariantID: plan.VariantID,
			FrameID: plan.FrameID, Outcome: status.OutcomeOK, Provenance: binding.Provenance,
		}},
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
	for _, value := range values {
		if len(target) >= status.MaxDiagnostics {
			break
		}
		target = append(target, value)
	}
	return target
}

func semanticSourceSpan(start, end int) text.Span {
	return text.Span{Start: start, End: end}
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
