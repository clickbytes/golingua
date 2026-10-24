package en

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

type englishIdentification struct {
	stateID       meaning.NodeID
	referent      meaning.NodeID
	role          string
	concept       meaning.ConceptRef
	senseID       string
	family        string
	mode          string
	temporalFrame meaning.TemporalFrame
	multiplicity  meaning.Multiplicity
	valid         bool
}

func englishIdentificationState(source semantic.MeaningCandidate) (englishIdentification, bool) {
	if source.Meaning == nil {
		return englishIdentification{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindState {
			continue
		}
		var id englishIdentification
		switch string(node.Concept) {
		case "concept:class_membership":
			id = englishIdentification{stateID: node.ID, role: "member", senseID: "sense:en.be.class_membership", family: "construction:en.identification"}
		case "concept:black":
			id = englishIdentification{stateID: node.ID, role: "bearer", senseID: "sense:en.black.color", family: "construction:en.attribution"}
		default:
			continue
		}
		id.valid = true
		if err := meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()); err != nil || len(source.Meaning.Roots) != 1 || source.Meaning.Roots[0] != node.ID || len(source.Meaning.Nodes) != 2 || len(node.Participants) != 1 || node.Participants[0].Role != meaning.Role(id.role) {
			id.valid = false
			return id, true
		}
		referent := findMeaningNode(*source.Meaning, node.Participants[0].Target)
		if referent.Kind != meaning.KindReferent || string(referent.Concept) != "concept:cat" {
			id.valid = false
			return id, true
		}
		id.referent, id.concept = referent.ID, referent.Concept
		id.temporalFrame, id.multiplicity = node.TemporalFrame, referent.Multiplicity
		if len(source.Contexts) == 0 {
			id.valid = false
			return id, true
		}
		for i, context := range source.Contexts {
			mode, ok := englishIdentificationContext(context, id)
			if !ok || i > 0 && mode != id.mode {
				id.valid = false
				return id, true
			}
			if i == 0 {
				id.mode = mode
			}
		}
		if id.mode != "proximal" && id.mode != "unresolved" && id.mode != "demonstrative" {
			id.valid = false
		}
		return id, true
	}
	return englishIdentification{}, false
}

func englishNegativeIdentificationState(source semantic.MeaningCandidate) (englishIdentification, bool) {
	if source.Meaning == nil {
		return englishIdentification{}, false
	}
	negative := false
	for _, node := range source.Meaning.Nodes {
		if node.Kind == meaning.KindScope && node.Operator == meaning.OperatorNegation {
			negative = true
		}
	}
	if !negative {
		return englishIdentification{}, false
	}
	id := englishIdentification{role: "member", senseID: "sense:en.be.class_membership", family: "construction:en.identification_negative", valid: false}
	if meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil || len(source.Meaning.Nodes) != 3 || len(source.Meaning.Roots) != 1 {
		return id, true
	}
	var scope, state, ref meaning.Node
	for _, node := range source.Meaning.Nodes {
		switch node.Kind {
		case meaning.KindScope:
			if scope.ID != "" {
				return id, true
			}
			scope = node
		case meaning.KindState:
			if state.ID != "" {
				return id, true
			}
			state = node
		case meaning.KindReferent:
			if ref.ID != "" {
				return id, true
			}
			ref = node
		default:
			return id, true
		}
	}
	if scope.ID == "" || state.ID == "" || ref.ID == "" || source.Meaning.Roots[0] != scope.ID || scope.Operator != meaning.OperatorNegation || scope.Target != state.ID || string(state.Concept) != "concept:class_membership" || string(ref.Concept) != "concept:cat" || len(state.Participants) != 1 || state.Participants[0].Role != "member" || state.Participants[0].Target != ref.ID {
		return id, true
	}
	id.stateID, id.referent, id.concept = state.ID, ref.ID, ref.Concept
	id.temporalFrame, id.multiplicity = state.TemporalFrame, ref.Multiplicity
	if len(source.Contexts) == 0 {
		return id, true
	}
	for i, context := range source.Contexts {
		mode, ok := englishNegativeIdentificationContext(context, id)
		if !ok || i > 0 && mode != id.mode {
			return id, true
		}
		if i == 0 {
			id.mode = mode
		}
	}
	if id.mode != "proximal" && id.mode != "unresolved" {
		return id, true
	}
	id.valid = true
	return id, true
}

func englishNegativeIdentificationContext(context semantic.MeaningContext, expected englishIdentification) (string, bool) {
	m := context.Meaning
	if meaning.ValidateReferences(m, context.Context, meaning.DefaultBudget()) != nil || len(m.Nodes) != 3 || len(m.Roots) != 1 || len(context.Context.Referents) != 0 || len(context.Context.Mentions) != 1 {
		return "", false
	}
	var scope, state, ref meaning.Node
	for _, node := range m.Nodes {
		switch node.Kind {
		case meaning.KindScope:
			if scope.ID != "" {
				return "", false
			}
			scope = node
		case meaning.KindState:
			if state.ID != "" {
				return "", false
			}
			state = node
		case meaning.KindReferent:
			if ref.ID != "" {
				return "", false
			}
			ref = node
		default:
			return "", false
		}
	}
	if scope.ID == "" || state.ID == "" || ref.ID == "" || m.Roots[0] != scope.ID || scope.Operator != meaning.OperatorNegation || scope.Target != state.ID || string(state.Concept) != "concept:class_membership" || string(ref.Concept) != string(expected.concept) || state.TemporalFrame != expected.temporalFrame || ref.Multiplicity != expected.multiplicity || len(state.Participants) != 1 || state.Participants[0].Role != "member" || state.Participants[0].Target != ref.ID {
		return "", false
	}
	mention := context.Context.Mentions[0]
	if mention.NodeID != ref.ID || mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil || mention.Anchor.Kind != "deixis" || (mention.Anchor.Value != "proximal" && mention.Anchor.Value != "unresolved") {
		return "", false
	}
	return mention.Anchor.Value, true
}

func englishIdentificationContextForSource(source semantic.MeaningCandidate) (englishIdentification, bool) {
	id, recognized := englishIdentificationState(source)
	return id, recognized && id.valid
}

func englishIdentificationContext(context semantic.MeaningContext, expected englishIdentification) (string, bool) {
	if meaning.ValidateReferences(context.Meaning, context.Context, meaning.DefaultBudget()) != nil || len(context.Meaning.Nodes) != 2 || len(context.Meaning.Roots) != 1 || len(context.Context.Referents) != 0 || len(context.Context.Mentions) != 1 {
		return "", false
	}
	var state, referent meaning.Node
	for _, node := range context.Meaning.Nodes {
		switch node.Kind {
		case meaning.KindState:
			if state.ID != "" {
				return "", false
			}
			state = node
		case meaning.KindReferent:
			if referent.ID != "" {
				return "", false
			}
			referent = node
		default:
			return "", false
		}
	}
	if state.ID == "" || referent.ID == "" || context.Meaning.Roots[0] != state.ID || string(state.Concept) != string(findMeaningNode(context.Meaning, state.ID).Concept) || string(state.Concept) != englishIdentificationConcept(expected.senseID) || string(referent.Concept) != string(expected.concept) || state.TemporalFrame != expected.temporalFrame || referent.Multiplicity != expected.multiplicity || len(state.Participants) != 1 || state.Participants[0].Role != meaning.Role(expected.role) || state.Participants[0].Target != referent.ID {
		return "", false
	}
	mention := context.Context.Mentions[0]
	if mention.NodeID != referent.ID || mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil || mention.Anchor.Kind != "deixis" {
		return "", false
	}
	mode := mention.Anchor.Value
	if mode != "proximal" && mode != "unresolved" && mode != "demonstrative" {
		return "", false
	}
	return mode, true
}

func englishIdentificationConcept(senseID string) string {
	if senseID == "sense:en.be.class_membership" {
		return "concept:class_membership"
	}
	if senseID == "sense:en.black.color" {
		return "concept:black"
	}
	return ""
}

func englishIdentificationReferenceChoice(variant, family string) (string, bool) {
	switch variant {
	case family + ".this":
		return "this", true
	case family + ".that":
		return "that", true
	case "construction:en.identification.it", "construction:en.identification_negative.it":
		if family == "construction:en.identification" || family == "construction:en.identification_negative" {
			return "it", true
		}
	}
	return "", false
}

var _ realizer.TargetCandidatePolicy = englishDeterminerCandidatePolicy{}
var _ runtime.TargetPlanBuilder = englishSentencePlanAnnotations{}
