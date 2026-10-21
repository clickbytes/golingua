package dialogue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/planner"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

// ConstructionRef identifies the licensed target construction selected by
// the normal realization pipeline for one response surface.
type ConstructionRef struct {
	AlternativeID string `json:"alternative_id"`
	PlanID        string `json:"plan_id"`
	FamilyID      string `json:"family_id"`
	VariantID     string `json:"variant_id"`
}

// Result retains the content plan's evidence beside every target alternative.
// It does not select among competing surface realizations.
type Result struct {
	Goal               planner.DialogueGoal        `json:"goal"`
	Outcome            status.OutcomeKind          `json:"outcome"`
	SelectedFactIDs    []string                    `json:"selected_fact_ids"`
	ConfidenceEvidence []planner.EvidenceRef       `json:"confidence_evidence"`
	ProofRefs          []string                    `json:"proof_refs"`
	Constructions      []ConstructionRef           `json:"constructions"`
	Alternatives       []engine.SurfaceAlternative `json:"alternatives"`
	Diagnostics        []status.Diagnostic         `json:"diagnostics,omitempty"`
}

// Realize sends one supported response predicate through the same target
// selector, ConstructionPlan, morphology executor and linearizer used by the
// translation pipeline. Non-statement dialogue acts fail closed until their
// language-local constructions are verified and registered.
func Realize(ctx context.Context, plan planner.ContentPlan, ontology *core.ConceptGraph, target engine.TargetRealizer) (Result, error) {
	result := Result{
		Goal: plan.Goal, Outcome: status.OutcomeUnsupported,
		SelectedFactIDs:    append([]string(nil), plan.SelectedFactIDs...),
		ConfidenceEvidence: append([]planner.EvidenceRef(nil), plan.ConfidenceEvidence...),
		ProofRefs:          append([]string(nil), plan.ProofRefs...),
		Constructions:      []ConstructionRef{}, Alternatives: []engine.SurfaceAlternative{},
	}
	if ontology == nil {
		return Result{}, dialogueError(status.OutcomeInvalid, "CORE_UNAVAILABLE", "core", "dialogue realization requires a pinned CORE graph")
	}
	if err := meaning.ValidateMeaning(plan.ResponseMeaning, ontology, meaning.DefaultBudget()); err != nil {
		return Result{}, err
	}
	if !SupportedAct(plan) {
		result.Diagnostics = []status.Diagnostic{unsupportedAct()}
		return result, nil
	}
	if ctx == nil {
		return Result{}, dialogueError(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "dialogue realization requires a request context")
	}
	if err := ctx.Err(); err != nil {
		result.Outcome = status.OutcomeResourceLimit
		result.Diagnostics = []status.Diagnostic{diagnostic("DIALOGUE_REALIZATION_CANCELED", "context", "dialogue realization was canceled or timed out", status.SeverityError)}
		return result, nil
	}
	if target == nil {
		result.Diagnostics = []status.Diagnostic{diagnostic("TARGET_RUNTIME_UNAVAILABLE", "target", "no target construction realization pipeline is configured", status.SeverityError)}
		return result, nil
	}
	contentID, ok := lexicalContentNode(plan.ResponseMeaning)
	if !ok {
		result.Diagnostics = []status.Diagnostic{diagnostic("DIALOGUE_MEANING_UNSUPPORTED", "response.meaning", "response meaning has no single licensed lexical target or uses an unsupported scope", status.SeverityError)}
		return result, nil
	}
	contentNode, _ := findNode(plan.ResponseMeaning, contentID)
	if plan.Goal == planner.GoalAnswerYesNo && contentNode.Kind != meaning.KindEvent && contentNode.Kind != meaning.KindState ||
		plan.Goal == planner.GoalAnswerWH && contentNode.Kind != meaning.KindReferent {
		result.Diagnostics = []status.Diagnostic{diagnostic("DIALOGUE_GOAL_MEANING_MISMATCH", "response.meaning", "response goal and lexical Meaning node kind do not match", status.SeverityError)}
		return result, nil
	}
	identity, err := semanticIdentity(plan)
	if err != nil {
		return Result{}, dialogueError(status.OutcomeInvalid, "INVALID_CONTENT_PLAN", "plan", "response plan could not be assigned a stable request-local identity")
	}
	graph := cloneMeaning(plan.ResponseMeaning)
	candidate := semantic.MeaningCandidate{
		ID: identity, SemanticID: identity, Meaning: &graph, Status: status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: identity, Outcome: status.OutcomeOK, Records: []semantic.EvidenceRecord{}},
	}
	surfaces, err := target.Realize(ctx, candidate)
	if err != nil {
		return Result{}, err
	}
	if !surfaces.Outcome.Valid() || len(surfaces.Alternatives) > engine.MaxTranslationAlternatives || len(surfaces.Diagnostics) > status.MaxDiagnostics {
		return Result{}, dialogueError(status.OutcomeInvalid, "INVALID_TARGET_RESULT", "target", "target pipeline returned an invalid or unbounded realization result")
	}
	result.Outcome = surfaces.Outcome
	result.Diagnostics = append([]status.Diagnostic(nil), surfaces.Diagnostics...)
	result.Alternatives = make([]engine.SurfaceAlternative, len(surfaces.Alternatives))
	for index, surface := range surfaces.Alternatives {
		if !surface.Outcome.Valid() || len(surface.Diagnostics) > status.MaxDiagnostics {
			return Result{}, dialogueError(status.OutcomeInvalid, "INVALID_TARGET_ALTERNATIVE", fmt.Sprintf("target.alternatives[%d]", index), "target pipeline returned an invalid or unbounded surface alternative")
		}
		if err := surface.Evidence.Validate(); err != nil {
			return Result{}, dialogueError(status.OutcomeInvalid, "INVALID_TARGET_EVIDENCE", fmt.Sprintf("target.alternatives[%d].evidence", index), "target pipeline returned invalid construction evidence")
		}
		for _, diagnostic := range surface.Diagnostics {
			if err := diagnostic.Validate(); err != nil {
				return Result{}, dialogueError(status.OutcomeInvalid, "INVALID_TARGET_DIAGNOSTIC", fmt.Sprintf("target.alternatives[%d].diagnostics", index), "target pipeline returned an invalid diagnostic")
			}
		}
		result.Alternatives[index] = cloneSurface(surface)
		constructionCount := 0
		for _, evidence := range surface.Evidence.Records {
			if evidence.PlanID == "" || evidence.FamilyID.String() == "" || evidence.VariantID.String() == "" {
				continue
			}
			ref := ConstructionRef{AlternativeID: surface.ID, PlanID: string(evidence.PlanID), FamilyID: evidence.FamilyID.String(), VariantID: string(evidence.VariantID)}
			result.Constructions = appendUniqueConstruction(result.Constructions, ref)
			constructionCount++
		}
		if surface.Outcome != status.OutcomeOK || len(surface.Losses) != 0 || surface.Text == "" || constructionCount == 0 {
			result.Alternatives[index].Text = ""
			result.Alternatives[index].Outcome = status.OutcomeUnsupported
			result.Alternatives[index].Diagnostics = append(result.Alternatives[index].Diagnostics, diagnostic("DIALOGUE_CONTENT_NOT_PRESERVED", "target", "target construction did not preserve the complete response meaning and its construction evidence", status.SeverityError))
			result.Outcome = combineOutcome(result.Outcome, status.OutcomeUnsupported)
		}
	}
	if len(result.Alternatives) == 0 && result.Outcome == status.OutcomeOK {
		result.Outcome = status.OutcomeUnknown
	}
	return result, nil
}

func lexicalContentNode(graph meaning.Meaning) (meaning.NodeID, bool) {
	root, ok := findNode(graph, graph.Roots[0])
	if !ok || len(root.Content) != 1 {
		return "", false
	}
	current := root.Content[0]
	for depth := 0; depth < meaning.MaxMeaningDepth; depth++ {
		node, exists := findNode(graph, current)
		if !exists {
			return "", false
		}
		switch node.Kind {
		case meaning.KindEvent, meaning.KindState, meaning.KindReferent:
			return node.ID, true
		case meaning.KindScope:
			if node.Operator != meaning.OperatorNegation {
				return "", false
			}
			current = node.Target
		default:
			return "", false
		}
	}
	return "", false
}

func semanticIdentity(plan planner.ContentPlan) (string, error) {
	value := struct {
		Goal               planner.DialogueGoal  `json:"goal"`
		Meaning            meaning.Meaning       `json:"meaning"`
		SelectedFactIDs    []string              `json:"selected_fact_ids"`
		ConfidenceEvidence []planner.EvidenceRef `json:"confidence_evidence"`
		ProofRefs          []string              `json:"proof_refs"`
	}{plan.Goal, plan.ResponseMeaning, append([]string(nil), plan.SelectedFactIDs...), append([]planner.EvidenceRef(nil), plan.ConfidenceEvidence...), append([]string(nil), plan.ProofRefs...)}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return "dialogue:" + hex.EncodeToString(digest[:]), nil
}

func cloneMeaning(graph meaning.Meaning) meaning.Meaning {
	graph.Roots = append([]meaning.NodeID(nil), graph.Roots...)
	graph.Nodes = append([]meaning.Node(nil), graph.Nodes...)
	for index := range graph.Nodes {
		graph.Nodes[index].Participants = append([]meaning.Participant(nil), graph.Nodes[index].Participants...)
		graph.Nodes[index].Content = append([]meaning.NodeID(nil), graph.Nodes[index].Content...)
		if graph.Nodes[index].Quantifier != nil {
			quantifier := *graph.Nodes[index].Quantifier
			if quantifier.Count != nil {
				count := *quantifier.Count
				quantifier.Count = &count
			}
			graph.Nodes[index].Quantifier = &quantifier
		}
	}
	return graph
}

func cloneSurface(surface engine.SurfaceAlternative) engine.SurfaceAlternative {
	surface.Evidence = surface.Evidence.Clone()
	surface.Losses = append([]realizer.FieldLoss(nil), surface.Losses...)
	for index := range surface.Losses {
		surface.Losses[index].Path = append([]meaning.NodeID(nil), surface.Losses[index].Path...)
	}
	surface.Diagnostics = append([]status.Diagnostic(nil), surface.Diagnostics...)
	return surface
}

func appendUniqueConstruction(values []ConstructionRef, next ConstructionRef) []ConstructionRef {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}

func findNode(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func diagnostic(code, path, message string, severity status.Severity) status.Diagnostic {
	return status.Diagnostic{Code: code, Path: path, Message: message, Severity: severity}
}

func combineOutcome(left, right status.OutcomeKind) status.OutcomeKind {
	if left == status.OutcomeOK {
		return right
	}
	if right == status.OutcomeOK || left == right {
		return left
	}
	if left == status.OutcomeInvalid || right == status.OutcomeInvalid {
		return status.OutcomeInvalid
	}
	if left == status.OutcomeResourceLimit || right == status.OutcomeResourceLimit {
		return status.OutcomeResourceLimit
	}
	return status.OutcomeIncomplete
}

type DialogueError struct{ result status.Result }

func (failure *DialogueError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "dialogue realization failed"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *DialogueError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}

func dialogueError(kind status.OutcomeKind, code, path, message string) error {
	return &DialogueError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic(code, path, message, status.SeverityError)}}}
}
