package en

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type englishAttributedSeeingGraph struct {
	event, property, subject, object meaning.Node
}

func englishAttributedSeeingShape(source semantic.MeaningCandidate) (englishAttributedSeeingGraph, bool) {
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil {
		return englishAttributedSeeingGraph{}, false
	}
	graph := *source.Meaning
	if len(graph.Roots) != 2 || len(graph.Nodes) != 4 {
		return englishAttributedSeeingGraph{}, false
	}
	shape := englishAttributedSeeingGraph{}
	referents := make(map[meaning.NodeID]meaning.Node, 2)
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindEvent:
			if shape.event.ID != "" {
				return englishAttributedSeeingGraph{}, false
			}
			shape.event = node
		case meaning.KindState:
			if shape.property.ID != "" {
				return englishAttributedSeeingGraph{}, false
			}
			shape.property = node
		case meaning.KindReferent:
			if _, duplicate := referents[node.ID]; duplicate {
				return englishAttributedSeeingGraph{}, false
			}
			referents[node.ID] = node
		default:
			return englishAttributedSeeingGraph{}, false
		}
	}
	if shape.event.ID == "" || shape.property.ID == "" || len(referents) != 2 ||
		shape.event.Kind != meaning.KindEvent || shape.event.Concept != "concept:seeing" || shape.event.TemporalFrame != meaning.TemporalFramePresent ||
		shape.property.Kind != meaning.KindState || shape.property.Concept != "concept:black" || shape.property.TemporalFrame != "" ||
		len(shape.event.Participants) != 2 || len(shape.property.Participants) != 1 {
		return englishAttributedSeeingGraph{}, false
	}
	var experiencer, stimulus meaning.NodeID
	for _, participant := range shape.event.Participants {
		switch participant.Role {
		case "experiencer":
			if experiencer != "" {
				return englishAttributedSeeingGraph{}, false
			}
			experiencer = participant.Target
		case "stimulus":
			if stimulus != "" {
				return englishAttributedSeeingGraph{}, false
			}
			stimulus = participant.Target
		default:
			return englishAttributedSeeingGraph{}, false
		}
	}
	if shape.property.Participants[0].Role != "bearer" {
		return englishAttributedSeeingGraph{}, false
	}
	bearer := shape.property.Participants[0].Target
	var subjectOK, objectOK bool
	shape.subject, subjectOK = referents[experiencer]
	shape.object, objectOK = referents[stimulus]
	if !subjectOK || !objectOK || shape.subject.ID == shape.object.ID || shape.subject.Kind != meaning.KindReferent || shape.object.Kind != meaning.KindReferent ||
		shape.subject.Multiplicity != meaning.MultiplicityOne || shape.object.Multiplicity != meaning.MultiplicityOne || bearer != shape.subject.ID {
		return englishAttributedSeeingGraph{}, false
	}
	rootIDs := map[meaning.NodeID]bool{}
	for _, root := range graph.Roots {
		rootIDs[root] = true
	}
	if len(rootIDs) != 2 || !rootIDs[shape.event.ID] || !rootIDs[shape.property.ID] {
		return englishAttributedSeeingGraph{}, false
	}
	return shape, true
}

func englishAttributedSeeingContextsValid(source semantic.MeaningCandidate) bool {
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

func englishAttributedSeeingFamily(candidate realizer.TargetCandidate) bool {
	return candidate.Family.Family.ID.String() == "construction:en.attributed_seeing"
}

func englishAttributedSeeingVariant(variant string) bool {
	switch variant {
	case "construction:en.attributed_seeing.svo",
		"construction:en.attributed_seeing.svo.indefinite_object",
		"construction:en.attributed_seeing.svo.indefinite_object_an",
		"construction:en.attributed_seeing.svo.indefinite_subject",
		"construction:en.attributed_seeing.svo.indefinite_subject_object",
		"construction:en.attributed_seeing.svo.indefinite_subject_a_object_an":
		return true
	default:
		return false
	}
}

func englishAttributedSeeingMainCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	shape, ok := englishAttributedSeeingShape(source)
	if !ok || !englishAttributedSeeingContextsValid(source) || candidate.NodeID != shape.event.ID || candidate.Kind != meaning.KindEvent || candidate.Outcome != status.OutcomeOK {
		return false
	}
	variant := candidate.CompiledVariant.Variant.ID.String()
	if candidate.Sense.ID.String() != "sense:en.see.visual" || string(candidate.Frame.ID) != "en.see.visual.transitive" ||
		candidate.TransferBinding.CoreRef.String() != "concept:seeing" || !englishAttributedSeeingFamily(candidate) || !englishAttributedSeeingVariant(variant) {
		return false
	}
	if !realizer.CoversLexicalMeaning(*source.Meaning, candidate.NodeID, candidate.Family.Family) {
		return false
	}
	modifier, found := realizer.FindModifier(*source.Meaning, candidate.NodeID, candidate.Family.Family, "modifier")
	return found && modifier.ID == shape.property.ID
}

func englishAttributedSeeingModifierCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, all []realizer.TargetCandidate) bool {
	shape, ok := englishAttributedSeeingShape(source)
	if !ok || !englishAttributedSeeingContextsValid(source) || candidate.NodeID != shape.property.ID || candidate.Kind != meaning.KindState || candidate.Outcome != status.OutcomeOK ||
		candidate.Sense.ID.String() != "sense:en.black.color" || string(candidate.Frame.ID) != "en.black.color.frame" ||
		candidate.Lexeme.ID.String() != "lexeme:en.black.adj" || candidate.TransferBinding.CoreRef.String() != "concept:black" || candidate.Family.Family.ID.String() != "construction:en.attribution" {
		return false
	}
	variant := candidate.CompiledVariant.Variant.ID.String()
	if variant != "construction:en.attribution.this" && variant != "construction:en.attribution.that" {
		return false
	}
	bestVariant := ""
	for _, support := range all {
		if support.SourceCandidateID != candidate.SourceCandidateID || support.SemanticID != candidate.SemanticID || support.NodeID != shape.property.ID ||
			support.Kind != meaning.KindState || support.Outcome != status.OutcomeOK || support.Sense.ID.String() != "sense:en.black.color" ||
			support.Lexeme.ID.String() != "lexeme:en.black.adj" || support.TransferBinding.CoreRef.String() != "concept:black" || support.Family.Family.ID.String() != "construction:en.attribution" {
			continue
		}
		supportVariant := support.CompiledVariant.Variant.ID.String()
		if supportVariant != "construction:en.attribution.this" && supportVariant != "construction:en.attribution.that" {
			continue
		}
		mainAvailable := false
		for _, main := range all {
			if main.SourceCandidateID == support.SourceCandidateID && main.SemanticID == support.SemanticID && englishAttributedSeeingMainCandidate(source, main) {
				mainAvailable = true
				break
			}
		}
		if mainAvailable && (bestVariant == "" || supportVariant < bestVariant) {
			bestVariant = supportVariant
		}
	}
	return variant == bestVariant
}
