// Package matrix contains the shared, production-backed runtime used by the
// language behavior matrix.
package matrix

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/tests/interlingua"
)

const maxComparisonNodes = 32

// GraphMatches compares two single-root Meaning graphs while ignoring local
// node IDs and node order. Every node must have a one-to-one counterpart.
// This is an alpha-structural contract comparison, not semantic equivalence
// and not evidence of entity identity.
func GraphMatches(want, got meaning.Meaning) (bool, error) {
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
	if len(want.Roots) != 1 || len(got.Roots) != 1 {
		return false, fmt.Errorf("Meaning comparison requires exactly one root")
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

	if !match(want.Roots[0], got.Roots[0]) {
		return false, nil
	}
	if len(wantToGot) != len(want.Nodes) || len(gotToWant) != len(got.Nodes) {
		return false, nil
	}
	return true, nil
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
