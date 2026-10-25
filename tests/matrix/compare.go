// Package matrix contains the shared, production-backed runtime used by the
// language behavior matrix.
package matrix

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/tests/interlingua"
)

const (
	maxComparisonNodes        = 32
	maxRootAssignmentAttempts = 4096
)

// GraphMatches compares Meaning graphs while ignoring root order, local node
// IDs, and node order. Every node must have a one-to-one counterpart across
// the complete graph.
// This is an alpha-structural contract comparison, not semantic equivalence
// and not evidence of entity identity.
func GraphMatches(want, got meaning.Meaning) (bool, error) {
	return graphMatchesWithNodeConstraints(want, got, nil)
}

// graphMatchesWithNodeConstraints applies required alpha-node correspondences
// during the same bijection used to compare graph structure.
func graphMatchesWithNodeConstraints(want, got meaning.Meaning, required map[meaning.NodeID]meaning.NodeID) (bool, error) {
	budget := meaning.Budget{
		MaxNodes: maxComparisonNodes,
		MaxEdges: meaning.MaxMeaningEdges,
		MaxDepth: meaning.MaxMeaningDepth,
	}
	if len(want.Nodes) > maxComparisonNodes || len(got.Nodes) > maxComparisonNodes {
		return false, fmt.Errorf("meaning graph exceeds comparison limit of %d nodes", maxComparisonNodes)
	}
	if err := meaning.ValidateStructure(want, budget); err != nil {
		return false, fmt.Errorf("invalid expected Meaning graph: %w", err)
	}
	if err := meaning.ValidateStructure(got, budget); err != nil {
		return false, fmt.Errorf("invalid actual Meaning graph: %w", err)
	}
	if len(want.Roots) == 0 || len(got.Roots) == 0 {
		return false, fmt.Errorf("Meaning comparison requires non-empty root sets")
	}
	if len(want.Roots) != len(got.Roots) {
		return false, nil
	}
	if len(want.Nodes) != len(got.Nodes) {
		return false, nil
	}

	wantNodes := nodeIndex(want)
	gotNodes := nodeIndex(got)
	wantToGot := make(map[meaning.NodeID]meaning.NodeID, len(want.Nodes))
	gotToWant := make(map[meaning.NodeID]meaning.NodeID, len(got.Nodes))
	var match func(meaning.NodeID, meaning.NodeID) bool
	match = func(wantID, gotID meaning.NodeID) bool {
		if requiredID, constrained := required[wantID]; constrained && requiredID != gotID {
			return false
		}
		for requiredWant, requiredGot := range required {
			if requiredGot == gotID && requiredWant != wantID {
				return false
			}
		}
		if mapped, exists := wantToGot[wantID]; exists {
			return mapped == gotID
		}
		if mapped, exists := gotToWant[gotID]; exists {
			return mapped == wantID
		}
		wantNode, wantExists := wantNodes[wantID]
		gotNode, gotExists := gotNodes[gotID]
		if !wantExists || !gotExists || !sameNodeFields(wantNode, gotNode) {
			return false
		}
		wantToGot[wantID], gotToWant[gotID] = gotID, wantID
		switch wantNode.Kind {
		case meaning.KindEvent, meaning.KindState:
			wantParticipants := participantsByRole(wantNode.Participants)
			gotParticipants := participantsByRole(gotNode.Participants)
			if len(wantParticipants) != len(gotParticipants) {
				return false
			}
			for role, wantTarget := range wantParticipants {
				gotTarget, exists := gotParticipants[role]
				if !exists || !match(wantTarget, gotTarget) {
					return false
				}
			}
		case meaning.KindDiscourse:
			if len(wantNode.Content) != len(gotNode.Content) {
				return false
			}
			for index := range wantNode.Content {
				if !match(wantNode.Content[index], gotNode.Content[index]) {
					return false
				}
			}
		case meaning.KindQuestion, meaning.KindScope, meaning.KindModality, meaning.KindQuantifier:
			if !match(wantNode.Target, gotNode.Target) {
				return false
			}
		case meaning.KindReferent:
			// No edges. Distinct referents remain distinct through the bijection.
		default:
			return false
		}
		return true
	}

	cloneMap := func(source map[meaning.NodeID]meaning.NodeID) map[meaning.NodeID]meaning.NodeID {
		copy := make(map[meaning.NodeID]meaning.NodeID, len(source))
		for key, value := range source {
			copy[key] = value
		}
		return copy
	}
	usedRoots := make([]bool, len(got.Roots))
	attempts := 0
	exhausted := false
	var assignRoots func(int) bool
	assignRoots = func(index int) bool {
		if index == len(want.Roots) {
			return len(wantToGot) == len(want.Nodes) && len(gotToWant) == len(got.Nodes)
		}
		for gotIndex, gotRoot := range got.Roots {
			if usedRoots[gotIndex] {
				continue
			}
			attempts++
			if attempts > maxRootAssignmentAttempts {
				exhausted = true
				return false
			}
			wantSnapshot := cloneMap(wantToGot)
			gotSnapshot := cloneMap(gotToWant)
			if match(want.Roots[index], gotRoot) {
				usedRoots[gotIndex] = true
				if assignRoots(index + 1) {
					return true
				}
				usedRoots[gotIndex] = false
			}
			wantToGot, gotToWant = wantSnapshot, gotSnapshot
			if exhausted {
				return false
			}
		}
		return false
	}
	matched := assignRoots(0)
	if exhausted {
		return false, fmt.Errorf("Meaning root assignment exceeded comparison limit of %d attempts", maxRootAssignmentAttempts)
	}
	return matched, nil
}

type contextMismatchError struct{ message string }

func (err contextMismatchError) Error() string { return err.message }

// ContextMatches compares a candidate with a graph-and-reference contract.
// Each anchored unresolved mention must bind to the same alpha-renamed graph
// node; anchors are not treated as a detached bag of values.
func ContextMatches(want interlingua.IdentificationMeaning, got semantic.MeaningCandidate) (bool, error) {
	if got.Meaning == nil {
		return false, fmt.Errorf("candidate has no Meaning graph")
	}
	if err := meaning.ValidateReferences(want.Graph, want.Context, meaning.DefaultBudget()); err != nil {
		return false, fmt.Errorf("invalid expected graph context: %w", err)
	}
	if len(want.Context.Referents) != 0 || len(want.Context.Mentions) == 0 {
		return false, fmt.Errorf("expected context must contain unresolved anchored mentions only")
	}
	for _, mention := range want.Context.Mentions {
		if mention.Anchor == nil || mention.ReferentID != "" || mention.EvidenceID != "" {
			return false, fmt.Errorf("expected mentions must be anchored and unresolved")
		}
	}
	if len(got.Contexts) == 0 {
		return false, contextMismatchError{"candidate has no request context for anchored mentions"}
	}
	matchedGraph, err := GraphMatches(want.Graph, *got.Meaning)
	if err != nil || !matchedGraph {
		return matchedGraph, err
	}
	for contextIndex, actual := range got.Contexts {
		if err := meaning.ValidateReferences(actual.Meaning, actual.Context, meaning.DefaultBudget()); err != nil {
			return false, fmt.Errorf("invalid candidate request context %d: %w", contextIndex, err)
		}
		if len(actual.Context.Referents) != 0 {
			return false, contextMismatchError{fmt.Sprintf("candidate request context %d has resolved identities", contextIndex)}
		}
		anchoredMentions := make([]meaning.Mention, 0, len(actual.Context.Mentions))
		for _, mention := range actual.Context.Mentions {
			if mention.ReferentID != "" || mention.EvidenceID != "" {
				return false, contextMismatchError{fmt.Sprintf("candidate request context %d has a resolved mention", contextIndex)}
			}
			if mention.Anchor != nil {
				anchoredMentions = append(anchoredMentions, mention)
			}
		}
		if len(anchoredMentions) != len(want.Context.Mentions) {
			return false, contextMismatchError{fmt.Sprintf("candidate request context %d has a different anchored mention count", contextIndex)}
		}
		if matches, err := GraphMatches(*got.Meaning, actual.Meaning); err != nil || !matches {
			if err != nil {
				return false, fmt.Errorf("candidate request context %d Meaning is invalid: %w", contextIndex, err)
			}
			return false, contextMismatchError{fmt.Sprintf("candidate request context %d Meaning differs from candidate Meaning", contextIndex)}
		}
		used := make([]bool, len(anchoredMentions))
		attempts := 0
		var assign func(int, map[meaning.NodeID]meaning.NodeID) (bool, error)
		assign = func(index int, constraints map[meaning.NodeID]meaning.NodeID) (bool, error) {
			if index == len(want.Context.Mentions) {
				return graphMatchesWithNodeConstraints(want.Graph, actual.Meaning, constraints)
			}
			wantMention := want.Context.Mentions[index]
			for actualIndex, gotMention := range anchoredMentions {
				if used[actualIndex] || gotMention.Anchor == nil || *gotMention.Anchor != *wantMention.Anchor ||
					gotMention.ReferentID != "" || gotMention.EvidenceID != "" {
					continue
				}
				attempts++
				if attempts > maxRootAssignmentAttempts {
					return false, fmt.Errorf("context mention assignment exceeded comparison limit of %d attempts", maxRootAssignmentAttempts)
				}
				if prior, exists := constraints[wantMention.NodeID]; exists && prior != gotMention.NodeID {
					continue
				}
				conflict := false
				for priorWant, priorGot := range constraints {
					if priorGot == gotMention.NodeID && priorWant != wantMention.NodeID {
						conflict = true
						break
					}
				}
				if conflict {
					continue
				}
				used[actualIndex] = true
				previous, alreadyConstrained := constraints[wantMention.NodeID]
				constraints[wantMention.NodeID] = gotMention.NodeID
				if ok, err := assign(index+1, constraints); err != nil || ok {
					return ok, err
				}
				used[actualIndex] = false
				if alreadyConstrained {
					constraints[wantMention.NodeID] = previous
				} else {
					delete(constraints, wantMention.NodeID)
				}
			}
			return false, nil
		}
		ok, err := assign(0, map[meaning.NodeID]meaning.NodeID{})
		if err != nil || !ok {
			if err != nil {
				return false, err
			}
			return false, contextMismatchError{fmt.Sprintf("candidate request context %d anchors do not match graph-bound mentions", contextIndex)}
		}
	}
	return true, nil
}

// MatchSourceCandidateMeaning binds a translation alternative to the exact
// authored MeaningID already validated for its source candidate ID.
func MatchSourceCandidateMeaning(sourceCandidateID string, candidate meaning.Meaning, sourceMeaningByCandidateID map[string]string, fixtures map[string]meaning.Meaning) (string, bool, error) {
	if sourceCandidateID == "" {
		return "", false, fmt.Errorf("translation candidate has an empty source candidate ID")
	}
	meaningID, ok := sourceMeaningByCandidateID[sourceCandidateID]
	if !ok || meaningID == "" {
		return "", false, fmt.Errorf("source candidate ID %q has no validated MeaningID binding", sourceCandidateID)
	}
	want, ok := fixtures[meaningID]
	if !ok {
		return "", false, fmt.Errorf("source candidate ID %q is bound to missing fixture %q", sourceCandidateID, meaningID)
	}
	matches, err := GraphMatches(want, candidate)
	return meaningID, matches, err
}

// MatchAllowedCandidate matches both the Meaning graph and any authored
// request-context contract, returning the unique licensed MeaningID.
func MatchAllowedCandidate(ids []string, fixtures map[string]meaning.Meaning, contexts map[string]*interlingua.IdentificationMeaning, candidate semantic.MeaningCandidate) (string, bool, error) {
	if candidate.Meaning == nil {
		return "", false, fmt.Errorf("candidate has no Meaning graph")
	}
	for _, id := range ids {
		if contexts[id] == nil {
			continue
		}
		if err := meaning.ValidateReferences(contexts[id].Graph, contexts[id].Context, meaning.DefaultBudget()); err != nil {
			return "", false, fmt.Errorf("invalid expected context fixture %q: %w", id, err)
		}
		if matches, err := GraphMatches(fixtures[id], contexts[id].Graph); err != nil {
			return "", false, fmt.Errorf("compare expected context fixture %q: %w", id, err)
		} else if !matches {
			return "", false, fmt.Errorf("expected context fixture %q differs from its allowed Meaning", id)
		}
	}
	matchedID := ""
	var contextMismatch error
	for _, id := range ids {
		fixture, ok := fixtures[id]
		if !ok {
			return "", false, fmt.Errorf("missing allowed fixture %q", id)
		}
		matches, err := GraphMatches(fixture, *candidate.Meaning)
		if err != nil {
			return "", false, err
		}
		if !matches {
			continue
		}
		if expected := contexts[id]; expected != nil {
			matches, err = ContextMatches(*expected, candidate)
			if err != nil {
				if mismatch, ok := err.(contextMismatchError); ok {
					contextMismatch = mismatch
					continue
				}
				return "", false, err
			}
			if !matches {
				continue
			}
		}
		if matchedID != "" {
			return "", false, fmt.Errorf("candidate matches multiple allowed meanings: %q and %q", matchedID, id)
		}
		matchedID = id
	}
	if matchedID == "" && contextMismatch != nil {
		return "", false, contextMismatch
	}
	return matchedID, matchedID != "", nil
}

// IdentificationMatches compares an analyzed candidate with an independent
// identification contract and checks its unresolved request-scoped mention.
func IdentificationMatches(want interlingua.IdentificationMeaning, got semantic.MeaningCandidate) (bool, error) {
	if got.Meaning == nil {
		return false, fmt.Errorf("candidate has no Meaning graph")
	}
	if err := meaning.ValidateReferences(want.Graph, want.Context, meaning.DefaultBudget()); err != nil {
		return false, fmt.Errorf("invalid expected identification fixture: %w", err)
	}
	if !hasOneReferent(want.Graph) || !hasOneExpectedMention(want.Context) {
		return false, fmt.Errorf("expected identification fixture must contain one referent and one unresolved anchored mention")
	}
	if matches, err := GraphMatches(want.Graph, *got.Meaning); err != nil || !matches {
		return matches, err
	}
	if !hasOneReferent(*got.Meaning) {
		return false, fmt.Errorf("candidate Meaning must contain exactly one referent")
	}
	if len(got.Contexts) == 0 {
		return false, fmt.Errorf("candidate has no request context for the identification mention")
	}
	for index, context := range got.Contexts {
		if err := meaning.ValidateReferences(context.Meaning, context.Context, meaning.DefaultBudget()); err != nil {
			return false, fmt.Errorf("invalid candidate request context %d: %w", index, err)
		}
		if matches, err := GraphMatches(*got.Meaning, context.Meaning); err != nil || !matches {
			if err != nil {
				return false, fmt.Errorf("candidate request context %d has invalid Meaning: %w", index, err)
			}
			return false, fmt.Errorf("candidate request context %d Meaning differs from candidate Meaning", index)
		}
		if len(context.Context.Referents) != 0 {
			return false, fmt.Errorf("candidate request context %d resolves an entity identity", index)
		}
		referentID, ok := onlyReferent(context.Meaning)
		if !ok || len(context.Context.Mentions) != 1 {
			return false, fmt.Errorf("candidate request context %d must contain exactly one mention of its only referent", index)
		}
		mention := context.Context.Mentions[0]
		if mention.NodeID != referentID || mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil {
			return false, fmt.Errorf("candidate request context %d mention must be unresolved and anchored to the sole referent", index)
		}
		wantAnchor := want.Context.Mentions[0].Anchor
		if wantAnchor == nil || *mention.Anchor != *wantAnchor {
			return false, fmt.Errorf("candidate request context %d anchor differs from the identification fixture", index)
		}
	}
	return true, nil
}

func hasOneReferent(graph meaning.Meaning) bool {
	_, ok := onlyReferent(graph)
	if !ok {
		return false
	}
	count := 0
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindReferent {
			count++
		}
	}
	return count == 1
}

func onlyReferent(graph meaning.Meaning) (meaning.NodeID, bool) {
	var id meaning.NodeID
	count := 0
	for _, node := range graph.Nodes {
		if node.Kind == meaning.KindReferent {
			id = node.ID
			count++
		}
	}
	return id, count == 1
}

func hasOneExpectedMention(context meaning.RequestContext) bool {
	if len(context.Referents) != 0 || len(context.Mentions) != 1 {
		return false
	}
	mention := context.Mentions[0]
	return mention.ReferentID == "" && mention.EvidenceID == "" && mention.Anchor != nil
}

func nodeIndex(graph meaning.Meaning) map[meaning.NodeID]meaning.Node {
	result := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		result[node.ID] = node
	}
	return result
}

func participantsByRole(participants []meaning.Participant) map[meaning.Role]meaning.NodeID {
	result := make(map[meaning.Role]meaning.NodeID, len(participants))
	for _, participant := range participants {
		result[participant.Role] = participant.Target
	}
	return result
}

func sameNodeFields(want, got meaning.Node) bool {
	if want.Kind != got.Kind || want.Concept != got.Concept || want.Operator != got.Operator ||
		want.Modality != got.Modality || want.Act != got.Act || len(want.Participants) != len(got.Participants) ||
		len(want.Content) != len(got.Content) || want.TemporalFrame != got.TemporalFrame || want.Multiplicity != got.Multiplicity {
		return false
	}
	if (want.Quantifier == nil) != (got.Quantifier == nil) {
		return false
	}
	if want.Quantifier != nil {
		if want.Quantifier.Kind != got.Quantifier.Kind || (want.Quantifier.Count == nil) != (got.Quantifier.Count == nil) {
			return false
		}
		if want.Quantifier.Count != nil && *want.Quantifier.Count != *got.Quantifier.Count {
			return false
		}
	}
	return true
}
