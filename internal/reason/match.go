package reason

import (
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

// Plan validates a typed question MeaningGraph and applies the caller's
// explicit semantic relation/role mapping. It never interprets surface text.
func Plan(input QuestionInput, ontology *core.ConceptGraph, budget meaning.Budget) (Question, error) {
	if ontology == nil {
		return Question{}, queryError(status.OutcomeInvalid, "CORE_UNAVAILABLE", "core", "query planning requires a pinned CORE graph")
	}
	if err := meaning.ValidateMeaning(input.Meaning, ontology, budget); err != nil {
		return Question{}, err
	}
	if err := input.Context.Validate(); err != nil {
		return Question{}, err
	}
	if input.Kind != QuestionYesNo && input.Kind != QuestionWH {
		return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_KIND", "question.kind", "question kind is not handled")
	}
	predicate, err := core.ParseRelationID(input.Mapping.Predicate)
	if err != nil || predicate.String() != input.Mapping.Predicate {
		return Question{}, queryError(status.OutcomeInvalid, "INVALID_QUERY_PREDICATE", "mapping.predicate", "predicate must be a canonical CORE relation ID")
	}
	relation, exists := ontology.Relation(predicate)
	if !exists {
		return Question{}, queryError(status.OutcomeUnsupported, "UNKNOWN_QUERY_PREDICATE", "mapping.predicate", "predicate relation is not declared in CORE")
	}
	if len(input.Mapping.RoleToArgument) != 2 {
		return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_RELATION_MAPPING", "mapping.role_to_argument", "a binary relation requires exactly two explicit semantic role bindings")
	}
	roleForArgument := [2]meaning.Role{}
	for role, argument := range input.Mapping.RoleToArgument {
		if role == "" || argument < 0 || argument > 1 || roleForArgument[argument] != "" {
			return Question{}, queryError(status.OutcomeInvalid, "INVALID_RELATION_MAPPING", "mapping.role_to_argument", "role mapping must bind two distinct roles to argument positions 0 and 1")
		}
		roleForArgument[argument] = role
	}
	if roleForArgument[0] == "" || roleForArgument[1] == "" {
		return Question{}, queryError(status.OutcomeInvalid, "INVALID_RELATION_MAPPING", "mapping.role_to_argument", "both relation argument positions must be bound")
	}
	nodes := make(map[meaning.NodeID]meaning.Node, len(input.Meaning.Nodes))
	for _, node := range input.Meaning.Nodes {
		nodes[node.ID] = node
	}
	if len(input.Meaning.Roots) != 1 {
		return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_ROOTS", "meaning.roots", "question planning requires one discourse root")
	}
	root := nodes[input.Meaning.Roots[0]]
	if root.Kind != meaning.KindDiscourse || root.Act != meaning.ActQuestion || len(root.Content) != 1 {
		return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_ACT", "meaning.roots", "question requires one explicit question discourse act")
	}
	current := nodes[root.Content[0]]
	questionMarkers := 0
	for depth := 0; depth < meaning.MaxMeaningDepth; depth++ {
		switch current.Kind {
		case meaning.KindQuestion:
			questionMarkers++
			current = nodes[current.Target]
		case meaning.KindScope:
			if current.Operator != meaning.OperatorQuestion {
				return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_OPERATOR_SCOPE", "meaning.nodes."+string(current.ID), "negated or hypothetical questions require an explicitly licensed query policy")
			}
			questionMarkers++
			current = nodes[current.Target]
		case meaning.KindEvent, meaning.KindState:
			if questionMarkers != 1 {
				return Question{}, queryError(status.OutcomeUnsupported, "MISSING_QUESTION_SCOPE", "meaning", "question requires exactly one typed question scope")
			}
			goto predicateFound
		default:
			return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_SHAPE", "meaning.nodes."+string(current.ID), "question target shape is not handled")
		}
		if questionMarkers > 1 {
			return Question{}, queryError(status.OutcomeUnsupported, "NESTED_QUESTION_SCOPE", "meaning", "nested question scopes are not handled")
		}
	}
	return Question{}, queryError(status.OutcomeResourceLimit, "QUERY_SCOPE_DEPTH", "meaning", "question scope exceeds the traversal bound")

predicateFound:
	conceptID, err := ids.ParseConceptID(string(current.Concept))
	if err != nil {
		return Question{}, queryError(status.OutcomeInvalid, "INVALID_QUESTION_CONCEPT", "meaning", "question predicate concept is malformed")
	}
	concept, exists := ontology.Lookup(conceptID)
	if !exists {
		return Question{}, queryError(status.OutcomeUnsupported, "UNKNOWN_QUESTION_CONCEPT", "meaning", "question predicate concept is not in CORE")
	}
	if relation.EventConceptID.String() != "" {
		if relation.EventConceptID != conceptID {
			return Question{}, queryError(status.OutcomeInvalid, "QUERY_EVENT_RELATION_MISMATCH", "mapping.predicate", "relation schema is bound to a different CORE event concept")
		}
		if string(roleForArgument[0]) != relation.ArgumentRoles[0] || string(roleForArgument[1]) != relation.ArgumentRoles[1] {
			return Question{}, queryError(status.OutcomeInvalid, "QUERY_EVENT_ROLE_ORDER_MISMATCH", "mapping.role_to_argument", "role mapping must preserve the event relation's declared argument order")
		}
	}
	declaredRoles := map[meaning.Role]struct{}{}
	for _, role := range concept.RoleTypes {
		roleID := meaning.Role(role.Name)
		declaredRoles[roleID] = struct{}{}
	}
	for _, role := range roleForArgument {
		if _, exists := declaredRoles[role]; !exists {
			return Question{}, queryError(status.OutcomeInvalid, "UNKNOWN_MAPPED_ROLE", "mapping.role_to_argument", "mapped role is not declared by the question concept")
		}
	}
	if input.Kind == QuestionWH {
		if input.GapRole == "" || (input.GapRole != roleForArgument[0] && input.GapRole != roleForArgument[1]) {
			return Question{}, queryError(status.OutcomeInvalid, "INVALID_WH_GAP", "question.gap_role", "WH question requires one mapped semantic role gap")
		}
	} else if input.GapRole != "" {
		return Question{}, queryError(status.OutcomeInvalid, "UNEXPECTED_QUESTION_GAP", "question.gap_role", "yes/no question cannot declare a WH role gap")
	}
	participantByRole := make(map[meaning.Role]meaning.NodeID, len(current.Participants))
	for _, participant := range current.Participants {
		if participant.Role != roleForArgument[0] && participant.Role != roleForArgument[1] {
			return Question{}, queryError(status.OutcomeUnsupported, "UNMAPPED_QUESTION_ROLE", "meaning.nodes."+string(current.ID)+".participants", "question contains a semantic role without an explicit relation mapping")
		}
		participantByRole[participant.Role] = participant.Target
	}
	query := Question{Kind: input.Kind, Predicate: predicate.String(), GapIndex: -1, At: input.At, MeaningNodeID: current.ID}
	for argument, role := range roleForArgument {
		target, present := participantByRole[role]
		if input.Kind == QuestionWH && role == input.GapRole {
			if present {
				return Question{}, queryError(status.OutcomeInvalid, "WH_GAP_HAS_VALUE", "meaning.nodes."+string(current.ID)+".participants", "WH gap role must be absent from grounded participants")
			}
			query.GapIndex = argument
			query.GapRole = role
			query.PredicateConceptID = string(current.Concept)
			continue
		}
		if !present {
			return Question{}, queryError(status.OutcomeUnsupported, "UNBOUND_QUESTION_ROLE", "meaning.nodes."+string(current.ID)+".participants", "question role is not grounded by a referent")
		}
		targetNode := nodes[target]
		if targetNode.Kind != meaning.KindReferent {
			return Question{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_ARGUMENT", "meaning.nodes."+string(target)+".kind", "binary fact questions require grounded referent arguments")
		}
		referent, err := resolveQuestionReferent(input.Context, target)
		if err != nil {
			return Question{}, err
		}
		requestReferent, found := findReferent(input.Context, referent)
		if !found || string(targetNode.Concept) != string(requestReferent.Concept) {
			return Question{}, queryError(status.OutcomeInvalid, "QUESTION_REFERENT_CONCEPT_MISMATCH", "question.context", "MeaningGraph referent and request referent concepts differ")
		}
		query.Args[argument] = referent
	}
	if input.Kind == QuestionWH && query.GapIndex < 0 {
		return Question{}, queryError(status.OutcomeInvalid, "MISSING_WH_GAP", "question.gap_role", "WH question did not produce a relation argument gap")
	}
	for argument, refID := range query.Args {
		if refID == "" {
			continue
		}
		referent, found := findReferent(input.Context, refID)
		if !found {
			return Question{}, queryError(status.OutcomeInvalid, "UNKNOWN_QUERY_REFERENT", "question.args", "grounded question referent is not in the request context")
		}
		conceptID, _ := ids.ParseConceptID(string(referent.Concept))
		targetConcept, _ := ontology.Lookup(conceptID)
		allowed := relation.SubjectKind
		if argument == 1 {
			allowed = relation.ObjectKind
		}
		if !allowsKind(allowed, targetConcept.Kind) {
			return Question{}, queryError(status.OutcomeInvalid, "QUERY_RELATION_KIND_MISMATCH", "question.args", "grounded referent violates the relation signature")
		}
	}
	return query, nil
}

func resolveQuestionReferent(context meaning.RequestContext, nodeID meaning.NodeID) (meaning.ReferentID, error) {
	var result meaning.ReferentID
	for _, mention := range context.Mentions {
		if mention.NodeID != nodeID {
			continue
		}
		if mention.ReferentID == "" {
			return "", queryError(status.OutcomeUnknown, "UNRESOLVED_QUESTION_REFERENT", "question.context", "question participant has no resolved request referent")
		}
		if result != "" && result != mention.ReferentID {
			return "", queryError(status.OutcomeAmbiguous, "AMBIGUOUS_QUESTION_REFERENT", "question.context", "question participant resolves to multiple referents")
		}
		result = mention.ReferentID
	}
	if result == "" {
		return "", queryError(status.OutcomeUnknown, "MISSING_QUESTION_REFERENT", "question.context", "question participant has no request-scoped mention evidence")
	}
	return result, nil
}

func findReferent(context meaning.RequestContext, id meaning.ReferentID) (meaning.Referent, bool) {
	for _, referent := range context.Referents {
		if referent.ID == id {
			return referent, true
		}
	}
	return meaning.Referent{}, false
}

// Execute resolves a grounded yes/no or WH query through the reasoner.
// WH enumeration is bounded and retains every positive/conflicting referent.
func Execute(set *facts.Set, rules []Rule, question Question, budget QueryBudget) (QuestionResult, error) {
	if set == nil {
		return QuestionResult{}, queryError(status.OutcomeInvalid, "FACT_SET_UNAVAILABLE", "facts", "question execution requires a request FactSet")
	}
	if budget == (QueryBudget{}) {
		budget = DefaultQueryBudget()
	}
	if budget.MaxWHCandidates <= 0 || budget.MaxWHCandidates > 256 {
		return QuestionResult{}, queryError(status.OutcomeInvalid, "INVALID_QUERY_BUDGET", "budget", "WH candidate limit must be positive and at most 256")
	}
	if question.Kind != QuestionYesNo && question.Kind != QuestionWH {
		return QuestionResult{}, queryError(status.OutcomeUnsupported, "UNSUPPORTED_QUESTION_KIND", "question.kind", "question kind is not handled")
	}
	if question.Kind == QuestionYesNo {
		if question.GapIndex != -1 || question.Args[0] == "" || question.Args[1] == "" {
			return QuestionResult{}, queryError(status.OutcomeInvalid, "INVALID_YES_NO_QUERY", "question", "yes/no query requires two grounded referents and no gap")
		}
		result, err := Evaluate(set, rules, Query{Predicate: question.Predicate, Args: question.Args, At: question.At}, budget.Reasoner)
		if err != nil {
			return QuestionResult{}, err
		}
		return yesNoResult(result), nil
	}
	if question.GapIndex < 0 || question.GapIndex > 1 || question.Args[question.GapIndex] != "" {
		return QuestionResult{}, queryError(status.OutcomeInvalid, "INVALID_WH_QUERY", "question", "WH query requires exactly one ungrounded relation argument")
	}
	context := set.Context()
	predicate, err := core.ParseRelationID(question.Predicate)
	if err != nil {
		return QuestionResult{}, queryError(status.OutcomeInvalid, "INVALID_QUERY_PREDICATE", "question.predicate", "predicate must be a canonical CORE relation ID")
	}
	relation, exists := set.CoreGraph().Relation(predicate)
	if !exists {
		return QuestionResult{}, queryError(status.OutcomeUnsupported, "UNKNOWN_QUERY_PREDICATE", "question.predicate", "predicate relation is not declared in CORE")
	}
	gapKinds, err := queryGapKinds(set.CoreGraph(), question)
	if err != nil {
		return QuestionResult{}, err
	}
	var candidates []meaning.ReferentID
	for _, referent := range context.Referents {
		conceptID, err := ids.ParseConceptID(string(referent.Concept))
		if err != nil {
			continue
		}
		concept, found := set.CoreGraph().Lookup(conceptID)
		if !found {
			continue
		}
		allowed := relation.SubjectKind
		if question.GapIndex == 1 {
			allowed = relation.ObjectKind
		}
		if allowsKind(allowed, concept.Kind) && allowsKind(gapKinds, concept.Kind) {
			candidates = append(candidates, referent.ID)
		}
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i] < candidates[j] })
	if len(candidates) > budget.MaxWHCandidates {
		return QuestionResult{}, queryLimit("wh_candidates", len(candidates))
	}
	answer := QuestionResult{Kind: QuestionWH, Truth: QuestionUnknown, Outcome: status.OutcomeUnknown, Candidates: []CandidateAnswer{}}
	for _, candidate := range candidates {
		args := question.Args
		args[question.GapIndex] = candidate
		result, err := Evaluate(set, rules, Query{Predicate: question.Predicate, Args: args, At: question.At}, budget.Reasoner)
		if err != nil {
			return QuestionResult{}, err
		}
		if result.Truth == TruthUnknown || result.Truth == TruthFalse {
			continue
		}
		truth := QuestionTrue
		if result.Truth == TruthConflict {
			truth = QuestionConflict
		}
		answer.Candidates = append(answer.Candidates, CandidateAnswer{ReferentID: candidate, Truth: truth, Proof: result.Proof, NegativeEvidence: result.NegativeEvidence})
	}
	if len(answer.Candidates) == 0 {
		return answer, nil
	}
	if len(answer.Candidates) > 1 {
		answer.Truth, answer.Outcome = QuestionAmbiguous, status.OutcomeAmbiguous
		return answer, nil
	}
	if answer.Candidates[0].Truth == QuestionConflict {
		answer.Truth, answer.Outcome = QuestionConflict, status.OutcomeAmbiguous
		return answer, nil
	}
	answer.Truth, answer.Outcome = QuestionTrue, status.OutcomeOK
	return answer, nil
}

func queryGapKinds(ontology *core.ConceptGraph, question Question) ([]core.ConceptKind, error) {
	if question.Kind != QuestionWH {
		return nil, nil
	}
	if question.GapRole == "" || question.PredicateConceptID == "" {
		return nil, queryError(status.OutcomeInvalid, "MISSING_WH_ROLE_SIGNATURE", "question.gap_role", "WH question requires its predicate concept and gap role")
	}
	conceptID, err := ids.ParseConceptID(question.PredicateConceptID)
	if err != nil || conceptID.String() != question.PredicateConceptID {
		return nil, queryError(status.OutcomeInvalid, "INVALID_WH_PREDICATE_CONCEPT", "question.predicate_concept_id", "predicate concept ID is not canonical")
	}
	concept, exists := ontology.Lookup(conceptID)
	if !exists {
		return nil, queryError(status.OutcomeUnsupported, "UNKNOWN_WH_PREDICATE_CONCEPT", "question.predicate_concept_id", "predicate concept is not in CORE")
	}
	for _, role := range concept.RoleTypes {
		if meaning.Role(role.Name) == question.GapRole {
			return append([]core.ConceptKind(nil), role.AllowedKinds...), nil
		}
	}
	return nil, queryError(status.OutcomeInvalid, "UNKNOWN_WH_GAP_ROLE", "question.gap_role", "gap role is not declared by the predicate concept")
}

func yesNoResult(result Result) QuestionResult {
	answer := QuestionResult{Kind: QuestionYesNo, Outcome: status.OutcomeOK}
	switch result.Truth {
	case TruthTrue:
		answer.Truth = QuestionTrue
		value := true
		answer.YesNo = &value
	case TruthFalse:
		answer.Truth = QuestionFalse
		value := false
		answer.YesNo = &value
	case TruthConflict:
		answer.Truth, answer.Outcome = QuestionConflict, status.OutcomeAmbiguous
	case TruthUnknown:
		answer.Truth, answer.Outcome = QuestionUnknown, status.OutcomeUnknown
	}
	answer.Proof = result.Proof
	answer.NegativeEvidence = result.NegativeEvidence
	return answer
}

func queryError(kind status.OutcomeKind, code, path, message string) error {
	return &StatusError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}}
}

func queryLimit(path string, count int) error {
	return queryError(status.OutcomeResourceLimit, "QUERY_RESOURCE_LIMIT", path, fmt.Sprintf("query %s limit exceeded at %d", path, count))
}
