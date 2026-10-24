package en

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

type englishIdentification struct {
	stateID  meaning.NodeID
	referent meaning.NodeID
	role     string
	concept  meaning.ConceptRef
	senseID  string
	family   string
	mode     string
	valid    bool
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
	if state.ID == "" || referent.ID == "" || context.Meaning.Roots[0] != state.ID || string(state.Concept) != string(findMeaningNode(context.Meaning, state.ID).Concept) || string(state.Concept) != englishIdentificationConcept(expected.senseID) || string(referent.Concept) != string(expected.concept) || len(state.Participants) != 1 || state.Participants[0].Role != meaning.Role(expected.role) || state.Participants[0].Target != referent.ID {
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
	case "construction:en.identification.it":
		if family == "construction:en.identification" {
			return "it", true
		}
	}
	return "", false
}

var _ realizer.TargetCandidatePolicy = englishDeterminerCandidatePolicy{}
var _ runtime.TargetPlanBuilder = englishSentencePlanAnnotations{}
