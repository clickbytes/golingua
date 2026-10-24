package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type polishAttributedSeeingGraph struct {
	event, property, subject, object meaning.Node
}

func polishAttributedSeeingShape(source semantic.MeaningCandidate) (polishAttributedSeeingGraph, bool) {
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil {
		return polishAttributedSeeingGraph{}, false
	}
	graph := *source.Meaning
	if len(graph.Roots) != 2 || len(graph.Nodes) != 4 {
		return polishAttributedSeeingGraph{}, false
	}
	shape := polishAttributedSeeingGraph{}
	referents := make(map[meaning.NodeID]meaning.Node, 2)
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindEvent:
			if shape.event.ID != "" {
				return polishAttributedSeeingGraph{}, false
			}
			shape.event = node
		case meaning.KindState:
			if shape.property.ID != "" {
				return polishAttributedSeeingGraph{}, false
			}
			shape.property = node
		case meaning.KindReferent:
			if _, duplicate := referents[node.ID]; duplicate {
				return polishAttributedSeeingGraph{}, false
			}
			referents[node.ID] = node
		default:
			return polishAttributedSeeingGraph{}, false
		}
	}
	if shape.event.ID == "" || shape.property.ID == "" || len(referents) != 2 ||
		shape.event.Kind != meaning.KindEvent || shape.event.Concept != "concept:seeing" || shape.event.TemporalFrame != meaning.TemporalFramePresent ||
		shape.property.Kind != meaning.KindState || shape.property.Concept != "concept:black" || shape.property.TemporalFrame != "" ||
		len(shape.event.Participants) != 2 || len(shape.property.Participants) != 1 {
		return polishAttributedSeeingGraph{}, false
	}
	var experiencer, stimulus, bearer meaning.NodeID
	for _, participant := range shape.event.Participants {
		switch participant.Role {
		case "experiencer":
			if experiencer != "" {
				return polishAttributedSeeingGraph{}, false
			}
			experiencer = participant.Target
		case "stimulus":
			if stimulus != "" {
				return polishAttributedSeeingGraph{}, false
			}
			stimulus = participant.Target
		default:
			return polishAttributedSeeingGraph{}, false
		}
	}
	if shape.property.Participants[0].Role != "bearer" {
		return polishAttributedSeeingGraph{}, false
	}
	bearer = shape.property.Participants[0].Target
	var subjectOK, objectOK bool
	shape.subject, subjectOK = referents[experiencer]
	shape.object, objectOK = referents[stimulus]
	if !subjectOK || !objectOK || shape.subject.ID == shape.object.ID || shape.subject.Kind != meaning.KindReferent || shape.object.Kind != meaning.KindReferent ||
		shape.subject.Multiplicity != meaning.MultiplicityOne || shape.object.Multiplicity != meaning.MultiplicityOne || bearer != shape.subject.ID {
		return polishAttributedSeeingGraph{}, false
	}
	rootIDs := map[meaning.NodeID]bool{}
	for _, root := range graph.Roots {
		rootIDs[root] = true
	}
	if len(rootIDs) != 2 || !rootIDs[shape.event.ID] || !rootIDs[shape.property.ID] {
		return polishAttributedSeeingGraph{}, false
	}
	return shape, true
}

func polishAttributedSeeingContextsValid(source semantic.MeaningCandidate) bool {
	if source.Meaning == nil {
		return false
	}
	if len(source.Contexts) == 0 {
		return true
	}
	for _, candidateContext := range source.Contexts {
		if meaning.ValidateReferences(candidateContext.Meaning, candidateContext.Context, meaning.DefaultBudget()) != nil {
			return false
		}
		equal, err := meaning.CanonicalEqual(candidateContext.Meaning, *source.Meaning, meaning.DefaultBudget())
		if err != nil || !equal || len(candidateContext.Context.Referents) != 0 {
			return false
		}
		for _, mention := range candidateContext.Context.Mentions {
			if mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor != nil {
				return false
			}
		}
	}
	return true
}

func polishAttributedSeeingMainCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	shape, ok := polishAttributedSeeingShape(source)
	if !ok || !polishAttributedSeeingContextsValid(source) || candidate.NodeID != shape.event.ID || candidate.Kind != meaning.KindEvent || candidate.Outcome != status.OutcomeOK {
		return false
	}
	if candidate.Sense.ID.String() != "sense:pl.widziec.visual" || string(candidate.Frame.ID) != "pl.widziec.visual.transitive" ||
		candidate.TransferBinding.CoreRef.String() != "concept:seeing" || candidate.Family.Family.ID.String() != "construction:pl.attributed_seeing" ||
		candidate.CompiledVariant.Variant.ID.String() != "construction:pl.attributed_seeing.svo" {
		return false
	}
	if !realizer.CoversLexicalMeaning(*source.Meaning, candidate.NodeID, candidate.Family.Family) {
		return false
	}
	modifier, found := realizer.FindModifier(*source.Meaning, candidate.NodeID, candidate.Family.Family, "modifier")
	return found && modifier.ID == shape.property.ID
}

func polishAttributedSeeingModifierCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, all []realizer.TargetCandidate) bool {
	shape, ok := polishAttributedSeeingShape(source)
	if !ok || !polishAttributedSeeingContextsValid(source) || candidate.NodeID != shape.property.ID || candidate.Kind != meaning.KindState || candidate.Outcome != status.OutcomeOK ||
		candidate.Sense.ID.String() != "sense:pl.czarny.black" || string(candidate.Frame.ID) != "pl.czarny.black.frame" ||
		candidate.TransferBinding.CoreRef.String() != "concept:black" || candidate.Family.Family.ID.String() != "construction:pl.attribution" ||
		candidate.CompiledVariant.Variant.ID.String() != "construction:pl.attribution.ten" {
		return false
	}
	for _, main := range all {
		if main.SourceCandidateID == candidate.SourceCandidateID && main.SemanticID == candidate.SemanticID && polishAttributedSeeingMainCandidate(source, main) {
			return true
		}
	}
	return false
}

func polishAttributedSeeingFamily(candidate realizer.TargetCandidate) bool {
	return candidate.Family.Family.ID.String() == "construction:pl.attributed_seeing"
}
