package en

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type englishParticipantContextError struct{}

func (englishParticipantContextError) Error() string {
	return "participant anchors must identify consistent supported subjects in their paired Meaning graphs"
}

func (failure englishParticipantContextError) Result() status.Result {
	return status.Result{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{
		Code: "INVALID_PARTICIPANT_CONTEXT", Path: "source.contexts", Message: failure.Error(), Severity: status.SeverityError,
	}}}
}

func englishParticipantAnchor(source semantic.MeaningCandidate, nodeID meaning.NodeID) (meaning.DeicticAnchor, bool) {
	if source.Meaning == nil {
		return meaning.DeicticAnchor{}, false
	}
	contexts := make([]meaning.RequestContext, 0, len(source.Contexts))
	for _, c := range source.Contexts {
		same, err := meaning.StructuralEquivalence(*source.Meaning, c.Meaning, meaning.DefaultBudget())
		if err != nil || !same || meaning.ValidateReferences(c.Meaning, c.Context, meaning.DefaultBudget()) != nil {
			return meaning.DeicticAnchor{}, false
		}
		contexts = append(contexts, c.Context)
	}
	return realizer.ImplicitParticipantAnchor(*source.Meaning, contexts, nodeID)
}

func englishParticipantPronoun(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, nodeID meaning.NodeID, lexemeID string) bool {
	if englishSubjectNode(source, predicate) != nodeID {
		return false
	}
	anchor, ok := englishParticipantAnchor(source, nodeID)
	if !ok || anchor.Kind != "participant" {
		return false
	}
	switch anchor.Value {
	case "speaker":
		return lexemeID == "lexeme:en.i.pron"
	case "speaker_inclusive_group":
		return lexemeID == "lexeme:en.we.pron"
	case "addressee":
		return lexemeID == "lexeme:en.you.pron"
	}
	return false
}

func englishParticipantAnchorIntent(source semantic.MeaningCandidate, nodeID meaning.NodeID) bool {
	for _, paired := range source.Contexts {
		for _, mention := range paired.Context.Mentions {
			if mention.NodeID == nodeID && mention.Anchor != nil && mention.Anchor.Kind == "participant" {
				return true
			}
		}
	}
	return false
}

func englishParticipantAttached(graph meaning.Meaning, nodeID meaning.NodeID) bool {
	count := 0
	for _, node := range graph.Nodes {
		for _, participant := range node.Participants {
			if participant.Target == nodeID {
				count++
			}
		}
	}
	return count == 1
}

func englishParticipantContextIntent(source semantic.MeaningCandidate) (bool, bool) {
	if source.Meaning == nil {
		return false, true
	}
	hasIntent := false
	for _, paired := range source.Contexts {
		for _, mention := range paired.Context.Mentions {
			if mention.Anchor != nil && mention.Anchor.Kind == "participant" {
				hasIntent = true
			}
		}
	}
	if !hasIntent {
		return false, true
	}
	for _, paired := range source.Contexts {
		same, err := meaning.StructuralEquivalence(*source.Meaning, paired.Meaning, meaning.DefaultBudget())
		if err != nil || !same || meaning.ValidateReferences(paired.Meaning, paired.Context, meaning.DefaultBudget()) != nil {
			return true, false
		}
	}
	for _, paired := range source.Contexts {
		for _, mention := range paired.Context.Mentions {
			if mention.Anchor == nil || mention.Anchor.Kind != "participant" {
				continue
			}
			if !englishParticipantAttached(*source.Meaning, mention.NodeID) {
				return true, false
			}
			if _, valid := englishParticipantAnchor(source, mention.NodeID); !valid {
				return true, false
			}
		}
	}
	return hasIntent, true
}
