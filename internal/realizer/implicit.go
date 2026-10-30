package realizer

import (
	"github.com/clickbytes/golingua/internal/meaning"
)

// ImplicitParticipantAnchor validates a person referent and its participant
// anchor across retained request contexts, whether mentions are unresolved
// anchors or resolved mentions with explicit evidence.
func ImplicitParticipantAnchor(graph meaning.Meaning, contexts []meaning.RequestContext, nodeID meaning.NodeID) (meaning.DeicticAnchor, bool) {
	var participant meaning.Node
	foundNode := false
	for _, node := range graph.Nodes {
		if node.ID == nodeID {
			participant = node
			foundNode = true
			break
		}
	}
	if !foundNode || participant.Kind != meaning.KindReferent || participant.Concept != "concept:person" || len(contexts) == 0 || len(contexts) > meaning.MaxMentions {
		return meaning.DeicticAnchor{}, false
	}
	var selected meaning.DeicticAnchor
	for _, context := range contexts {
		if err := meaning.ValidateReferences(graph, context, meaning.DefaultBudget()); err != nil {
			return meaning.DeicticAnchor{}, false
		}
		foundMention := false
		for _, mention := range context.Mentions {
			if mention.NodeID != nodeID {
				continue
			}
			foundMention = true
			resolved := mention.ReferentID != "" && mention.EvidenceID != ""
			unresolved := mention.ReferentID == "" && mention.EvidenceID == ""
			if (!resolved && !unresolved) || mention.Anchor == nil || mention.Anchor.Kind != "participant" ||
				(mention.Anchor.Value != "speaker" && mention.Anchor.Value != "speaker_inclusive_group" && mention.Anchor.Value != "addressee") {
				return meaning.DeicticAnchor{}, false
			}
			anchor := *mention.Anchor
			if selected.Kind != "" && selected != anchor {
				return meaning.DeicticAnchor{}, false
			}
			selected = anchor
		}
		if !foundMention {
			return meaning.DeicticAnchor{}, false
		}
	}
	if selected.Kind == "" || selected.Value == "speaker" && participant.Multiplicity != meaning.MultiplicityOne ||
		selected.Value == "speaker_inclusive_group" && participant.Multiplicity != meaning.MultiplicityMany {
		return meaning.DeicticAnchor{}, false
	}
	return selected, true
}
