package planner

import (
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/reason"
	"github.com/clickbytes/golingua/internal/status"
)

type PlanInput struct {
	Decision        DecisionInput
	QuestionMeaning meaning.Meaning
	Core            *core.ConceptGraph
	FactSet         *facts.Set
}

type EvidenceRef struct {
	FactID string               `json:"fact_id"`
	Source facts.SourceEvidence `json:"source"`
}

// ContentPlan is replayable semantic content and evidence. It contains no
// surface text, language-specific word choice or numeric confidence score.
type ContentPlan struct {
	Goal               DialogueGoal        `json:"goal"`
	Act                meaning.DialogueAct `json:"act"`
	ResponseMeaning    meaning.Meaning     `json:"response_meaning"`
	SelectedFactIDs    []string            `json:"selected_fact_ids"`
	ProofRefs          []string            `json:"proof_refs"`
	ConfidenceEvidence []EvidenceRef       `json:"confidence_evidence"`
	Clarification      *ClarificationGap   `json:"clarification,omitempty"`
}

// Build chooses a response act, constructs its typed MeaningGraph, and records
// every selected fact/source and proof reference from the query result.
func Build(input PlanInput) (ContentPlan, error) {
	if input.Core == nil {
		return ContentPlan{}, plannerError(status.OutcomeInvalid, "CORE_UNAVAILABLE", "core", "content planning requires a pinned CORE graph")
	}
	if input.FactSet != nil && input.FactSet.CoreGraph() != input.Core {
		return ContentPlan{}, plannerError(status.OutcomeInvalid, "CORE_SNAPSHOT_MISMATCH", "facts", "FactSet and planner must share the same pinned CORE graph")
	}
	if err := meaning.ValidateMeaning(input.QuestionMeaning, input.Core, meaning.DefaultBudget()); err != nil {
		return ContentPlan{}, err
	}
	if input.Decision.Result == nil {
		return ContentPlan{}, plannerError(status.OutcomeInvalid, "QUERY_RESULT_REQUIRED", "query_result", "content planning requires a typed query result")
	}
	if _, exists := findNode(input.QuestionMeaning, input.Decision.Query.MeaningNodeID); !exists {
		return ContentPlan{}, plannerError(status.OutcomeInvalid, "QUESTION_NODE_MISSING", "question.meaning_node_id", "planned question predicate is absent from its MeaningGraph")
	}
	if err := validateQuestionMeaning(input.QuestionMeaning, input.Decision.Query.MeaningNodeID); err != nil {
		return ContentPlan{}, err
	}
	decision, err := ChooseGoal(input.Decision)
	if err != nil {
		return ContentPlan{}, err
	}
	if err := validateDecisionInput(input.Decision); err != nil {
		return ContentPlan{}, err
	}
	if decision.Goal == GoalClarify {
		if err := validateClarification(input); err != nil {
			return ContentPlan{}, err
		}
	}
	response, err := buildResponseMeaning(input, decision)
	if err != nil {
		return ContentPlan{}, err
	}
	if err := meaning.ValidateMeaning(response, input.Core, meaning.DefaultBudget()); err != nil {
		return ContentPlan{}, plannerError(status.OutcomeInvalid, "INVALID_RESPONSE_MEANING", "response_meaning", err.Error())
	}
	selected, proofRefs, evidence, err := collectEvidence(input.Decision.Result, input.FactSet)
	if err != nil {
		return ContentPlan{}, err
	}
	plan := ContentPlan{Goal: decision.Goal, Act: decision.Act, ResponseMeaning: response, SelectedFactIDs: selected, ProofRefs: proofRefs, ConfidenceEvidence: evidence}
	if input.Decision.Clarification != nil {
		gap := *input.Decision.Clarification
		plan.Clarification = &gap
	}
	return plan, nil
}

func validateDecisionInput(input DecisionInput) error {
	result := input.Result
	if result.Kind != input.Query.Kind {
		return plannerError(status.OutcomeInvalid, "QUESTION_RESULT_KIND_MISMATCH", "query_result.kind", "query result kind differs from the planned question")
	}
	if !result.Outcome.Valid() {
		return plannerError(status.OutcomeInvalid, "INVALID_QUERY_OUTCOME", "query_result.outcome", "query result outcome is not recognized")
	}
	if result.Outcome == status.OutcomeUnsupported {
		if result.Truth != "" && result.Truth != reason.QuestionUnknown {
			return plannerError(status.OutcomeInvalid, "INVALID_UNSUPPORTED_RESULT", "query_result.truth", "unsupported query cannot carry a positive or negative truth value")
		}
		return nil
	}
	switch result.Truth {
	case reason.QuestionTrue:
		if result.Outcome != status.OutcomeOK {
			return plannerError(status.OutcomeInvalid, "INVALID_TRUE_QUERY_RESULT", "query_result", "true query requires an OK outcome")
		}
		if result.Kind == reason.QuestionYesNo && (result.YesNo == nil || !*result.YesNo) {
			return plannerError(status.OutcomeInvalid, "INVALID_YES_NO_RESULT", "query_result.yes_no", "true yes/no answer requires true value")
		}
		if result.Kind == reason.QuestionWH && (len(result.Candidates) != 1 || result.Candidates[0].Truth != reason.QuestionTrue) {
			return plannerError(status.OutcomeInvalid, "INVALID_WH_RESULT", "query_result.candidates", "true WH answer requires exactly one true candidate")
		}
	case reason.QuestionFalse:
		if result.Outcome != status.OutcomeOK || result.Kind != reason.QuestionYesNo || result.YesNo == nil || *result.YesNo {
			return plannerError(status.OutcomeInvalid, "INVALID_FALSE_QUERY_RESULT", "query_result", "false answer requires a false yes/no value and OK outcome")
		}
	case reason.QuestionConflict:
		if result.Outcome != status.OutcomeAmbiguous {
			return plannerError(status.OutcomeInvalid, "INVALID_CONFLICT_RESULT", "query_result", "conflicting evidence requires an ambiguous outcome")
		}
	case reason.QuestionAmbiguous:
		if result.Outcome != status.OutcomeAmbiguous || result.Kind != reason.QuestionWH || len(result.Candidates) < 2 {
			return plannerError(status.OutcomeInvalid, "INVALID_AMBIGUOUS_RESULT", "query_result", "ambiguous WH answer requires multiple candidates")
		}
	case reason.QuestionUnknown:
		if result.Outcome != status.OutcomeUnknown {
			return plannerError(status.OutcomeInvalid, "INVALID_UNKNOWN_RESULT", "query_result", "unknown query requires an unknown outcome")
		}
	default:
		return plannerError(status.OutcomeInvalid, "INVALID_QUESTION_TRUTH", "query_result.truth", "question truth is not recognized")
	}
	return nil
}

func validateClarification(input PlanInput) error {
	gap := input.Decision.Clarification
	if gap == nil {
		return plannerError(status.OutcomeInvalid, "CLARIFICATION_GAP_REQUIRED", "clarification", "clarification goal requires an explicit typed gap")
	}
	predicate, exists := findNode(input.QuestionMeaning, input.Decision.Query.MeaningNodeID)
	if !exists {
		return plannerError(status.OutcomeInvalid, "CLARIFICATION_PREDICATE_MISSING", "question.meaning", "clarification predicate is missing")
	}
	conceptID, err := ids.ParseConceptID(string(predicate.Concept))
	if err != nil {
		return plannerError(status.OutcomeInvalid, "INVALID_CLARIFICATION_CONCEPT", "question.meaning", "clarification predicate concept is malformed")
	}
	concept, exists := input.Core.Lookup(conceptID)
	if !exists {
		return plannerError(status.OutcomeInvalid, "UNKNOWN_CLARIFICATION_CONCEPT", "question.meaning", "clarification predicate concept is absent from CORE")
	}
	roleExists := false
	for _, role := range concept.RoleTypes {
		if role.Name == gap.Role {
			roleExists = true
			break
		}
	}
	if !roleExists {
		return plannerError(status.OutcomeInvalid, "UNKNOWN_CLARIFICATION_ROLE", "clarification.role", "clarification role is not declared by the question predicate")
	}
	switch gap.Kind {
	case ClarifyChoice:
		if input.Decision.Query.Kind != reason.QuestionWH || gap.Role != string(input.Decision.Query.GapRole) || input.Decision.Result.Truth != reason.QuestionAmbiguous {
			return plannerError(status.OutcomeInvalid, "INVALID_CHOICE_CLARIFICATION", "clarification", "choice clarification must match an ambiguous WH role")
		}
	case ClarifyReference:
		if input.FactSet == nil {
			return plannerError(status.OutcomeInvalid, "FACT_SET_REQUIRED", "clarification", "reference clarification requires request context evidence")
		}
		nodeID := meaning.NodeID(gap.NodeID)
		participant := false
		for _, role := range predicate.Participants {
			if role.Role == meaning.Role(gap.Role) && role.Target == nodeID {
				participant = true
				break
			}
		}
		if !participant {
			return plannerError(status.OutcomeInvalid, "CLARIFICATION_NODE_MISMATCH", "clarification.node_id", "clarification node must fill the declared role in the question predicate")
		}
		unresolved, referents := false, map[string]struct{}{}
		for _, mention := range input.FactSet.Context().Mentions {
			if mention.NodeID != nodeID {
				continue
			}
			if mention.ReferentID == "" {
				unresolved = true
			} else {
				referents[string(mention.ReferentID)] = struct{}{}
			}
		}
		if !unresolved && len(referents) < 2 {
			return plannerError(status.OutcomeInvalid, "CLARIFICATION_EVIDENCE_MISSING", "clarification.node_id", "request context does not show an unresolved or ambiguous referent")
		}
	default:
		return plannerError(status.OutcomeInvalid, "INVALID_CLARIFICATION_KIND", "clarification.kind", "clarification kind is not supported")
	}
	return nil
}

func buildResponseMeaning(input PlanInput, decision Decision) (meaning.Meaning, error) {
	question := input.Decision.Query
	result := input.Decision.Result
	switch decision.Goal {
	case GoalClarify:
		return meaning.Build(input.QuestionMeaning.Roots, input.QuestionMeaning.Nodes, meaning.DefaultBudget())
	case GoalAnswerYesNo:
		return buildPropositionResponse(input.QuestionMeaning, question.MeaningNodeID, result.Truth == reason.QuestionFalse, input.Core)
	case GoalAnswerWH:
		if input.FactSet == nil {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "FACT_SET_REQUIRED", "facts", "WH response requires its request FactSet")
		}
		candidate := result.Candidates[0]
		referent, exists := findRequestReferent(input.FactSet.Context(), candidate.ReferentID)
		if !exists {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "WH_REFERENT_MISSING", "query_result.candidates", "selected WH referent is absent from the request context")
		}
		conceptID, err := ids.ParseConceptID(string(referent.Concept))
		if err != nil {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "INVALID_WH_REFERENT", "query_result.candidates", "selected referent concept is malformed")
		}
		nodeID := meaning.NodeID("answer_referent")
		referentNode, err := meaning.NewReferent(nodeID, conceptID)
		if err != nil {
			return meaning.Meaning{}, err
		}
		return discourseMeaning(input.Core, meaning.ActStatement, []meaning.Node{referentNode}, []meaning.NodeID{nodeID})
	case GoalExplain:
		if result.Kind == reason.QuestionYesNo && result.Truth == reason.QuestionConflict {
			return buildConflictResponse(input.QuestionMeaning, question.MeaningNodeID, input.Core)
		}
		return buildWHExplanation(input.FactSet, result, input.Core)
	case GoalRefuse:
		return refusalMeaning(input.QuestionMeaning)
	default:
		return meaning.Meaning{}, plannerError(status.OutcomeUnsupported, "UNSUPPORTED_CONTENT_GOAL", "goal", "response goal is not supported")
	}
}

func buildPropositionResponse(source meaning.Meaning, predicateID meaning.NodeID, negate bool, ontology *core.ConceptGraph) (meaning.Meaning, error) {
	predicate, exists := findNode(source, predicateID)
	if !exists || (predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState) {
		return meaning.Meaning{}, plannerError(status.OutcomeUnsupported, "UNSUPPORTED_RESPONSE_PROPOSITION", "question.meaning_node_id", "answer requires an event or state predicate")
	}
	nodes := []meaning.Node{predicate}
	for _, participant := range predicate.Participants {
		target, found := findNode(source, participant.Target)
		if !found {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "RESPONSE_PARTICIPANT_MISSING", "question.meaning", "predicate participant node is missing")
		}
		nodes = appendMeaningNode(nodes, target)
	}
	contentID := predicate.ID
	if negate {
		negationID := uniqueNodeID(nodes, "answer_negation")
		negation, err := meaning.NewNegation(negationID, predicate.ID)
		if err != nil {
			return meaning.Meaning{}, err
		}
		nodes = append(nodes, negation)
		contentID = negationID
	}
	return discourseMeaning(ontology, meaning.ActStatement, nodes, []meaning.NodeID{contentID})
}

func buildConflictResponse(source meaning.Meaning, predicateID meaning.NodeID, ontology *core.ConceptGraph) (meaning.Meaning, error) {
	predicate, exists := findNode(source, predicateID)
	if !exists {
		return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "CONFLICT_PREDICATE_MISSING", "question.meaning", "conflict response has no source predicate")
	}
	nodes := []meaning.Node{predicate}
	for _, participant := range predicate.Participants {
		target, found := findNode(source, participant.Target)
		if !found {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "RESPONSE_PARTICIPANT_MISSING", "question.meaning", "predicate participant node is missing")
		}
		nodes = appendMeaningNode(nodes, target)
	}
	negationID := uniqueNodeID(nodes, "conflict_negation")
	negation, err := meaning.NewNegation(negationID, predicate.ID)
	if err != nil {
		return meaning.Meaning{}, err
	}
	nodes = append(nodes, negation)
	return discourseMeaning(ontology, meaning.ActExplanation, nodes, []meaning.NodeID{predicate.ID, negationID})
}

func buildWHExplanation(set *facts.Set, result *reason.QuestionResult, ontology *core.ConceptGraph) (meaning.Meaning, error) {
	if set == nil || len(result.Candidates) == 0 {
		return meaning.Meaning{}, plannerError(status.OutcomeUnsupported, "EXPLANATION_EVIDENCE_UNAVAILABLE", "query_result", "ambiguous explanation requires explicit referent evidence")
	}
	context := set.Context()
	nodes, content := []meaning.Node{}, []meaning.NodeID{}
	for index, candidate := range result.Candidates {
		referent, exists := findRequestReferent(context, candidate.ReferentID)
		if !exists {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "EXPLANATION_REFERENT_MISSING", "query_result.candidates", "candidate referent is absent from its request context")
		}
		conceptID, err := ids.ParseConceptID(string(referent.Concept))
		if err != nil {
			return meaning.Meaning{}, plannerError(status.OutcomeInvalid, "INVALID_EXPLANATION_REFERENT", "query_result.candidates", "candidate concept ID is malformed")
		}
		nodeID := meaning.NodeID(fmt.Sprintf("explanation_ref_%d", index))
		node, err := meaning.NewReferent(nodeID, conceptID)
		if err != nil {
			return meaning.Meaning{}, err
		}
		nodes = append(nodes, node)
		content = append(content, nodeID)
	}
	return discourseMeaning(ontology, meaning.ActExplanation, nodes, content)
}

func validateQuestionMeaning(graph meaning.Meaning, predicateID meaning.NodeID) error {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	if len(graph.Roots) != 1 {
		return plannerError(status.OutcomeInvalid, "INVALID_PLANNED_QUESTION", "question_meaning.roots", "question plan requires one root")
	}
	root := nodes[graph.Roots[0]]
	if root.Kind != meaning.KindDiscourse || root.Act != meaning.ActQuestion || len(root.Content) != 1 {
		return plannerError(status.OutcomeInvalid, "INVALID_PLANNED_QUESTION", "question_meaning", "content planner requires an explicit question discourse act")
	}
	current := nodes[root.Content[0]]
	seen := 0
	for depth := 0; depth < meaning.MaxMeaningDepth; depth++ {
		if current.ID == predicateID {
			if current.Kind != meaning.KindEvent && current.Kind != meaning.KindState {
				return plannerError(status.OutcomeInvalid, "INVALID_PLANNED_PREDICATE", "question_meaning", "query predicate must be an event or state")
			}
			if seen != 1 {
				return plannerError(status.OutcomeInvalid, "INVALID_PLANNED_QUESTION", "question_meaning", "question meaning must contain exactly one typed question scope")
			}
			return nil
		}
		switch current.Kind {
		case meaning.KindQuestion:
			seen++
			current = nodes[current.Target]
		case meaning.KindScope:
			if current.Operator != meaning.OperatorQuestion {
				return plannerError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_SCOPE", "question_meaning", "planner cannot construct a response for this operator scope")
			}
			seen++
			current = nodes[current.Target]
		default:
			return plannerError(status.OutcomeInvalid, "INVALID_PLANNED_QUESTION", "question_meaning", "query predicate is not the target of a question scope")
		}
		if seen > 1 {
			return plannerError(status.OutcomeInvalid, "NESTED_QUESTION_SCOPE", "question_meaning", "nested question scopes cannot be planned")
		}
	}
	return plannerError(status.OutcomeResourceLimit, "QUESTION_SCOPE_DEPTH", "question_meaning", "question scope exceeds the traversal bound")
}

func refusalMeaning(source meaning.Meaning) (meaning.Meaning, error) {
	rootID := uniqueNodeID(source.Nodes, "refusal")
	refusal, err := meaning.NewDiscourse(rootID, meaning.ActRefusal, []meaning.NodeID{source.Roots[0]})
	if err != nil {
		return meaning.Meaning{}, err
	}
	nodes := append(append([]meaning.Node(nil), source.Nodes...), refusal)
	return meaning.Build([]meaning.NodeID{rootID}, nodes, meaning.DefaultBudget())
}

func discourseMeaning(_ *core.ConceptGraph, act meaning.DialogueAct, nodes []meaning.Node, content []meaning.NodeID) (meaning.Meaning, error) {
	rootID := uniqueNodeID(nodes, "response")
	root, err := meaning.NewDiscourse(rootID, act, content)
	if err != nil {
		return meaning.Meaning{}, err
	}
	nodes = append(nodes, root)
	return meaning.Build([]meaning.NodeID{rootID}, nodes, meaning.DefaultBudget())
}

func collectEvidence(result *reason.QuestionResult, factSet *facts.Set) ([]string, []string, []EvidenceRef, error) {
	factIDs, proofIDs := map[string]struct{}{}, map[string]struct{}{}
	var collectProof func([]reason.ProofNode)
	collectProof = func(proofs []reason.ProofNode) {
		for _, proof := range proofs {
			proofIDs[proof.ID] = struct{}{}
			for _, id := range proof.PremiseFactIDs {
				factIDs[id] = struct{}{}
			}
		}
	}
	collectProof(result.Proof)
	for _, candidate := range result.Candidates {
		collectProof(candidate.Proof)
		for _, fact := range candidate.NegativeEvidence {
			factIDs[fact.ID] = struct{}{}
		}
	}
	for _, fact := range result.NegativeEvidence {
		factIDs[fact.ID] = struct{}{}
	}
	selected, proofRefs := sortedKeys(factIDs), sortedKeys(proofIDs)
	evidence := make([]EvidenceRef, 0, len(selected))
	for _, id := range selected {
		if factSet == nil {
			return nil, nil, nil, plannerError(status.OutcomeInvalid, "FACT_SET_REQUIRED", "facts", "evidence references require their request FactSet")
		}
		fact, exists := factSet.ByID(id)
		if !exists {
			return nil, nil, nil, plannerError(status.OutcomeInvalid, "UNKNOWN_EVIDENCE_FACT", "query_result", "query references a fact outside its request FactSet")
		}
		evidence = append(evidence, EvidenceRef{FactID: id, Source: fact.Source})
	}
	return selected, proofRefs, evidence, nil
}

func findNode(graph meaning.Meaning, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}
func findRequestReferent(context meaning.RequestContext, id meaning.ReferentID) (meaning.Referent, bool) {
	for _, ref := range context.Referents {
		if ref.ID == id {
			return ref, true
		}
	}
	return meaning.Referent{}, false
}

func appendMeaningNode(nodes []meaning.Node, node meaning.Node) []meaning.Node {
	for _, existing := range nodes {
		if existing.ID == node.ID {
			return nodes
		}
	}
	return append(nodes, node)
}

func uniqueNodeID(nodes []meaning.Node, prefix string) meaning.NodeID {
	used := map[meaning.NodeID]struct{}{}
	for _, node := range nodes {
		used[node.ID] = struct{}{}
	}
	for index := 0; ; index++ {
		candidate := meaning.NodeID(prefix)
		if index > 0 {
			candidate = meaning.NodeID(fmt.Sprintf("%s_%d", prefix, index))
		}
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func sortedKeys(set map[string]struct{}) []string {
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func plannerError(kind status.OutcomeKind, code, path, message string) error {
	return &PlannerError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: strings.TrimSpace(message), Severity: status.SeverityError}}}}
}

type PlannerError struct{ result status.Result }

func (failure *PlannerError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "content planning failed"
	}
	return failure.result.Diagnostics[0].Message
}
func (failure *PlannerError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}
