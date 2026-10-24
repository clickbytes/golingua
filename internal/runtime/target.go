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
	CandidatePolicy  realizer.TargetCandidatePolicy
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
	selection, err := realizer.SelectTargets(realizer.Input{
		Candidates: []semantic.MeaningCandidate{source}, Outcome: source.Status,
		Diagnostics: source.Evidence.Diagnostics, Operators: operators,
		CandidatePolicy:          pipeline.CandidatePolicy,
		OperatorCoverageComplete: operatorCoverageComplete,
		Language:                 lexicon.LanguageID(pipeline.Language), Lexicon: pipeline.Lexicon,
		Valency: pipeline.Valency, Constructions: pipeline.Constructions,
		Variants: pipeline.Variants, Core: pipeline.Core, Budget: pipeline.SelectionBudget,
	})
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
	surfaceIDs := topLevelTargetNodeIDs(*source.Meaning)
	neededIDs := targetExecutionNodeIDs(*source.Meaning, surfaceIDs)
	candidates := make([]realizer.TargetCandidate, 0, len(selection.Candidates))
	surfaceCandidates := make([]realizer.TargetCandidate, 0, len(selection.Candidates))
	for _, candidate := range selection.Candidates {
		if neededIDs[candidate.NodeID] && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState || candidate.Kind == meaning.KindReferent) {
			candidates = append(candidates, candidate)
		}
		if surfaceIDs[candidate.NodeID] && (candidate.Kind == meaning.KindEvent || candidate.Kind == meaning.KindState || candidate.Kind == meaning.KindReferent) {
			surfaceCandidates = append(surfaceCandidates, candidate)
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
	if len(surfaceCandidates) == 0 && len(result.Diagnostics) == 0 {
		result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, status.Diagnostic{
			Code: "TARGET_NO_SURFACE_CANDIDATE", Path: "target.candidates",
			Message:  "no licensed surface candidate for the supplied Meaning, context and evidence",
			Severity: status.SeverityWarning,
		})
	}
	for _, selected := range surfaceCandidates {
		if err := ctx.Err(); err != nil {
			return engine.SurfaceResult{}, targetPipelineError(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "target realization was canceled or timed out")
		}
		planInput := realizer.PlanInput{Meaning: meaning.Meaning{}, Candidate: selected, Core: pipeline.Core, Budget: pipeline.PlanBudget}
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
		executionInput, err := pipeline.ExecutionBuilder.Build(ctx, source, candidates, selected, plan)
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
		alternative := engine.SurfaceAlternative{
			ID: plan.ID, Outcome: outcome, Text: linearized.Text,
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
	if len(result.Alternatives) == 0 && result.Outcome == status.OutcomeOK {
		result.Outcome = status.OutcomeUnknown
	}
	return result, nil
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
		case meaning.KindScope, meaning.KindModality, meaning.KindQuantifier:
			return false
		}
	}
	return true
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

func topLevelTargetNodeIDs(graph meaning.Meaning) map[meaning.NodeID]bool {
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
			if node.Act == meaning.ActStatement && len(node.Content) == 1 {
				visit(node.Content[0], depth+1)
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
