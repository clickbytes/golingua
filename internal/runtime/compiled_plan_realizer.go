package runtime

import (
	"context"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
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

// ConstructionPlanRealizer adapts the low-level language.Realizer API to the
// same selector, plan, morphology and linearization stages as sentence
// translation. Because Meaning v1 has no tense or number, this API records
// those target defaults as explicit assumptions and returns incomplete.
type ConstructionPlanRealizer struct {
	Language         features.LanguageID
	Lexicon          *lexicon.Index
	Valency          *valency.Index
	Constructions    *construction.Registry
	Variants         map[ids.ConstructionID]*construction.VariantIndex
	Core             *core.ConceptGraph
	PlanBuilder      TargetPlanBuilder
	ExecutionBuilder TargetExecutionBuilder
	SelectionBudget  realizer.Budget
	PlanBudget       realizer.PlanBudget
	ExecuteBudget    realizer.ExecuteBudget
	LinearizeBudget  realizer.LinearizeBudget
	CandidatePolicy  realizer.TargetCandidatePolicy
}

var _ language.Realizer = (*ConstructionPlanRealizer)(nil)

func (adapter *ConstructionPlanRealizer) Realize(ctx context.Context, graph meaning.Meaning, nodeID meaning.NodeID, compiled construction.CompiledConstructionPlan) (language.RealizationResult, error) {
	if adapter == nil || adapter.Lexicon == nil || adapter.Valency == nil || adapter.Constructions == nil || adapter.Core == nil || adapter.ExecutionBuilder == nil {
		return failedLanguageRealization(status.OutcomeUnsupported, "LANGUAGE_REALIZER_UNAVAILABLE", "compiled_plan", "compiled plan realizer requires verified language indexes and slot execution")
	}
	if ctx == nil {
		return failedLanguageRealization(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "compiled plan realization requires a request context")
	}
	if err := ctx.Err(); err != nil {
		return failedLanguageRealization(status.OutcomeResourceLimit, "TARGET_REALIZATION_CANCELED", "context", "compiled plan realization was canceled or timed out")
	}
	if !compiled.ValidateIdentity() || features.LanguageID(compiled.Language) != adapter.Language || compiled.Language == "" || len(compiled.Permissions) != 1 {
		return failedLanguageRealization(status.OutcomeInvalid, "INVALID_COMPILED_PLAN", "compiled_plan", "compiled construction plan identity or language does not match this provider")
	}
	if len(graph.Roots) != 1 || graph.Roots[0] != nodeID {
		return failedLanguageRealization(status.OutcomeUnsupported, "ROOT_PREDICATE_REQUIRED", "meaning.roots", "this low-level provider currently realizes one root predicate and its explicit arguments")
	}
	predicate, exists := meaningNode(graph, nodeID)
	if !exists || predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return failedLanguageRealization(status.OutcomeUnsupported, "PREDICATE_NODE_REQUIRED", "meaning.node", "compiled plan realization requires one event or state node")
	}
	source, err := assumedProviderCandidate(graph, predicate, adapter.Core)
	if err != nil {
		return failedLanguageRealization(status.OutcomeUnsupported, "SOURCE_FEATURES_UNAVAILABLE", "meaning", err.Error())
	}
	selection, err := realizer.SelectTargets(realizer.Input{
		Candidates: []semantic.MeaningCandidate{source}, Outcome: status.OutcomeOK,
		CandidatePolicy:          adapter.CandidatePolicy,
		OperatorCoverageComplete: true, Language: lexicon.LanguageID(adapter.Language),
		Lexicon: adapter.Lexicon, Valency: adapter.Valency, Constructions: adapter.Constructions,
		Variants: adapter.Variants, Core: adapter.Core, Budget: adapter.SelectionBudget,
	})
	if err != nil {
		return language.RealizationResult{}, err
	}
	var selected realizer.TargetCandidate
	count := 0
	for _, candidate := range selection.Candidates {
		if candidate.Kind == predicate.Kind && candidate.NodeID == nodeID && candidate.Sense.ID == compiled.Permissions[0].SenseID &&
			candidate.Frame.ID == compiled.Permissions[0].FrameID && candidate.Family.Family.ID == compiled.FamilyID &&
			candidate.CompiledVariant.Variant.ID == compiled.VariantID {
			selected = candidate
			count++
		}
	}
	if count == 0 {
		return failedLanguageRealization(status.OutcomeUnsupported, "COMPILED_ROUTE_UNAVAILABLE", "compiled_plan", "compiled plan does not match an approved target sense, frame, family and variant route")
	}
	if count > 1 {
		return failedLanguageRealization(status.OutcomeAmbiguous, "COMPILED_ROUTE_AMBIGUOUS", "compiled_plan", "multiple target lexical candidates match the compiled plan route")
	}
	assumptions := providerFeatureAssumptions(predicate)
	planInput := realizer.PlanInput{Meaning: graph, Candidate: selected, Core: adapter.Core, Budget: adapter.PlanBudget, Assumptions: assumptions}
	if adapter.PlanBuilder != nil {
		var planAssumptions []realizer.FieldAssumption
		var planLosses []realizer.FieldLoss
		planInput.FeatureConstraints, planAssumptions, planLosses, err = adapter.PlanBuilder.Annotations(ctx, source, selected)
		if err != nil {
			return language.RealizationResult{}, err
		}
		planInput.Assumptions = append(planInput.Assumptions, planAssumptions...)
		planInput.Losses = append(planInput.Losses, planLosses...)
	}
	plan, err := realizer.BuildPlan(planInput)
	if err != nil {
		return language.RealizationResult{}, err
	}
	execution, err := adapter.ExecutionBuilder.Build(ctx, source, selection.Candidates, selected, plan)
	if err != nil {
		return language.RealizationResult{}, err
	}
	execution.Plan = plan
	execution.SourceMeaning = graph
	execution.Core = adapter.Core
	if execution.Budget == (realizer.ExecuteBudget{}) {
		execution.Budget = adapter.ExecuteBudget
	}
	executed, err := realizer.Execute(execution)
	if err != nil {
		return language.RealizationResult{}, err
	}
	linearized, err := realizer.Linearize(plan, executed.Family, executed.Slots, execution.InformationStructure, adapter.LinearizeBudget)
	if err != nil {
		return language.RealizationResult{}, err
	}
	result := language.RealizationResult{Text: linearized.Text, Outcome: combineTargetOutcomes(selection.Outcome, selected.Outcome, plan.Outcome, executed.Outcome, linearized.Outcome)}
	result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, plan.Diagnostics...)
	result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, executed.Diagnostics...)
	result.Diagnostics = appendTargetDiagnostics(result.Diagnostics, linearized.Diagnostics...)
	if result.Outcome != status.OutcomeOK && len(result.Diagnostics) == 0 {
		result.Diagnostics = []status.Diagnostic{{Code: "LANGUAGE_REALIZATION_INCOMPLETE", Path: "compiled_plan", Message: "realization used explicit defaults because Meaning does not carry all grammatical features", Severity: status.SeverityWarning}}
	}
	return result, nil
}

func assumedProviderCandidate(graph meaning.Meaning, predicate meaning.Node, ontology *core.ConceptGraph) (semantic.MeaningCandidate, error) {
	allowedNodes := map[meaning.NodeID]bool{predicate.ID: true}
	for _, participant := range predicate.Participants {
		allowedNodes[participant.Target] = true
	}
	if len(allowedNodes) != len(graph.Nodes) {
		return semantic.MeaningCandidate{}, fmt.Errorf("only one bare event or state with its explicit referents is supported")
	}
	for _, node := range graph.Nodes {
		if !allowedNodes[node.ID] || node.ID == predicate.ID && node.Kind != meaning.KindEvent && node.Kind != meaning.KindState ||
			node.ID != predicate.ID && node.Kind != meaning.KindReferent {
			return semantic.MeaningCandidate{}, fmt.Errorf("questions, scope, modality, quantifiers, discourse and additional predicates are outside this compiled plan route")
		}
	}
	concept, err := ids.ParseConceptID(string(predicate.Concept))
	if err != nil {
		return semantic.MeaningCandidate{}, err
	}
	records := []semantic.EvidenceRecord{{CoreConceptID: concept, Features: []features.FeatureValue{
		{Feature: "verb_form", State: features.StateKnown, Value: "third_person_singular"},
		{Feature: "tense", State: features.StateKnown, Value: "present"},
		{Feature: "number", State: features.StateKnown, Value: "singular"},
		{Feature: "person", State: features.StateKnown, Value: "third"},
		{Feature: "mood", State: features.StateKnown, Value: "indicative"},
		{Feature: "voice", State: features.StateKnown, Value: "active"},
	}, Outcome: status.OutcomeOK}}
	for _, participant := range predicate.Participants {
		node, found := meaningNode(graph, participant.Target)
		if !found || node.Kind != meaning.KindReferent {
			return semantic.MeaningCandidate{}, fmt.Errorf("predicate participant must identify an explicit referent")
		}
		participantConcept, parseErr := ids.ParseConceptID(string(node.Concept))
		if parseErr != nil {
			return semantic.MeaningCandidate{}, parseErr
		}
		records = append(records, semantic.EvidenceRecord{CoreConceptID: participantConcept, Role: participant.Role, Outcome: status.OutcomeOK, Features: []features.FeatureValue{
			{Feature: "number", State: features.StateKnown, Value: "singular"},
			{Feature: "noun_number", State: features.StateKnown, Value: "singular"},
		}})
	}
	if _, err := meaning.Build(graph.Roots, graph.Nodes, meaning.DefaultBudget()); err != nil {
		return semantic.MeaningCandidate{}, err
	}
	if err := meaning.ValidateMeaning(graph, ontology, meaning.DefaultBudget()); err != nil {
		return semantic.MeaningCandidate{}, err
	}
	return semantic.MeaningCandidate{
		ID: "compiled-plan-provider", SemanticID: "compiled-plan-provider", Meaning: &graph,
		Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: "compiled-plan-provider", Outcome: status.OutcomeOK, Records: records},
	}, nil
}

func providerFeatureAssumptions(predicate meaning.Node) []realizer.FieldAssumption {
	assumptions := []realizer.FieldAssumption{
		{Path: []meaning.NodeID{predicate.ID}, Field: "tense", Value: "present", Reason: "Meaning v1 has no tense field; this low-level provider uses the present form"},
		{Path: []meaning.NodeID{predicate.ID}, Field: "number", Value: "singular", Reason: "Meaning v1 has no grammatical number field; this provider realizes third-person singular agreement"},
	}
	for _, participant := range predicate.Participants {
		assumptions = append(assumptions, realizer.FieldAssumption{
			Path: []meaning.NodeID{predicate.ID, participant.Target}, Field: "number", Value: "singular",
			Reason: "Meaning v1 has no grammatical number field; this provider realizes singular noun forms",
		})
	}
	return assumptions
}

func failedLanguageRealization(outcome status.OutcomeKind, code, path, message string) (language.RealizationResult, error) {
	return language.RealizationResult{Outcome: outcome, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityWarning}}}, nil
}
