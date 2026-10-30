package en

import (
	"strings"

	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
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
		default:
			attribution, allowed := englishAttributionForConcept(string(node.Concept))
			if !allowed {
				continue
			}
			id = englishIdentification{stateID: node.ID, role: "bearer", senseID: attribution.sense, family: "construction:en.attribution"}
		}
		id.valid = true
		if err := meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()); err != nil || len(source.Meaning.Roots) != 1 || source.Meaning.Roots[0] != node.ID || len(source.Meaning.Nodes) != 2 || len(node.Participants) != 1 || node.Participants[0].Role != meaning.Role(id.role) {
			id.valid = false
			return id, true
		}
		referent := findMeaningNode(*source.Meaning, node.Participants[0].Target)
		if referent.Kind != meaning.KindReferent || referent.Concept == "" {
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
		if id.mode != "proximal" && id.mode != "distal" && id.mode != "unresolved" && id.mode != "demonstrative" && id.mode != "none" {
			id.valid = false
		}
		return id, true
	}
	return englishIdentification{}, false
}

func englishNegativeAttributionState(source semantic.MeaningCandidate) (englishIdentification, bool) {
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil || len(source.Meaning.Nodes) != 3 || len(source.Meaning.Roots) != 1 {
		return englishIdentification{}, false
	}
	var scope, state, referent meaning.Node
	for _, node := range source.Meaning.Nodes {
		switch node.Kind {
		case meaning.KindScope:
			scope = node
		case meaning.KindState:
			state = node
		case meaning.KindReferent:
			referent = node
		}
	}
	if scope.ID == "" || state.ID == "" || referent.ID == "" || source.Meaning.Roots[0] != scope.ID || scope.Operator != meaning.OperatorNegation || scope.Target != state.ID ||
		len(state.Participants) != 1 || state.Participants[0].Role != "bearer" || state.Participants[0].Target != referent.ID || string(referent.Concept) == "" {
		return englishIdentification{}, false
	}
	attribution, allowed := englishAttributionForConcept(string(state.Concept))
	if !allowed {
		return englishIdentification{}, false
	}
	id := englishIdentification{stateID: state.ID, referent: referent.ID, role: "bearer", concept: referent.Concept, senseID: attribution.sense,
		family: "construction:en.attribution_negative", temporalFrame: state.TemporalFrame, multiplicity: referent.Multiplicity, mode: "none", valid: true}
	if len(source.Contexts) > 0 {
		for _, item := range source.Contexts {
			if len(item.Context.Referents) != 0 || len(item.Context.Mentions) > 1 {
				return englishIdentification{}, false
			}
			if len(item.Context.Mentions) == 1 {
				mention := item.Context.Mentions[0]
				if mention.NodeID != referent.ID || mention.Anchor != nil && mention.Anchor.Kind != "deixis" {
					return englishIdentification{}, false
				}
				if mention.Anchor != nil {
					id.mode = mention.Anchor.Value
				}
			}
		}
	}
	return id, true
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
	if scope.ID == "" || state.ID == "" || ref.ID == "" || source.Meaning.Roots[0] != scope.ID || scope.Operator != meaning.OperatorNegation || scope.Target != state.ID || string(state.Concept) != "concept:class_membership" || ref.Concept == "" || len(state.Participants) != 1 || state.Participants[0].Role != "member" || state.Participants[0].Target != ref.ID {
		if state.ID != "" && string(state.Concept) != "concept:class_membership" {
			return englishIdentification{}, false
		}
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

// englishIdentificationMemberCountSupported admits class-membership routes
// only when the exact Meaning referent has an approved count-noun target.
func englishIdentificationMemberCountSupported(source semantic.MeaningCandidate, expected englishIdentification, candidates []realizer.TargetCandidate) bool {
	if expected.role != "member" {
		return true
	}
	if source.Meaning == nil || expected.referent == "" || expected.concept == "" {
		return false
	}
	var referent meaning.Node
	for _, node := range source.Meaning.Nodes {
		if node.ID == expected.referent {
			referent = node
			break
		}
	}
	if referent.ID == "" || referent.Kind != meaning.KindReferent || referent.Concept != expected.concept {
		return false
	}
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != expected.referent || candidate.Outcome != status.OutcomeOK || candidate.TransferBinding.CoreRef.String() != string(referent.Concept) {
			continue
		}
		mass, pluralOnly := englishNumberClass(candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		explicitCount := false
		for _, property := range candidate.Lexeme.Properties {
			if property.Key == "number_behavior" && property.Value == "count" {
				explicitCount = true
				break
			}
		}
		_, articleKnown := englishIndefiniteArticle(candidate.Lexeme.Properties)
		if mass && !pluralOnly && !explicitCount {
			return true
		}
		if explicitCount && !mass && !pluralOnly && articleKnown {
			return true
		}
	}
	return false
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
	if mention.NodeID == referent.ID && mention.ReferentID == "" && mention.EvidenceID == "" && mention.Anchor == nil {
		return "none", true
	}
	if mention.NodeID != referent.ID || mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil || mention.Anchor.Kind != "deixis" {
		return "", false
	}
	mode := mention.Anchor.Value
	if mode != "proximal" && mode != "distal" && mode != "unresolved" && mode != "demonstrative" {
		return "", false
	}
	return mode, true
}

func englishIdentificationConcept(senseID string) string {
	if senseID == "sense:en.be.class_membership" {
		return "concept:class_membership"
	}
	if attribution, allowed := englishAttributionForSense(senseID); allowed {
		return attribution.concept
	}
	return ""
}

func englishAttributionDoorNumberChoice(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slotID, role string) (string, bool) {
	identification, recognized := englishTargetIdentificationState(source)
	if !recognized || !identification.valid || identification.role != "bearer" || identification.concept != "concept:door" ||
		identification.mode != "none" || predicate.NodeID != identification.stateID || predicate.Family.Family.ID.String() != "construction:en.attribution" {
		return "", false
	}
	if role != "" && (role != "bearer" || slotID != "bearer") {
		return "", false
	}
	variant := predicate.CompiledVariant.Variant.ID.String()
	switch variant {
	case "construction:en.attribution.the_plural":
		if identification.multiplicity == meaning.MultiplicityOne {
			return "", false
		}
		return "plural", true
	case "construction:en.attribution.the":
		if identification.multiplicity == meaning.MultiplicityMany {
			return "", false
		}
		return "singular", true
	default:
		return "", false
	}
}

func englishIdentificationReferenceChoice(variant, family string) (string, bool) {
	variant = strings.TrimSuffix(variant, ".past")
	switch variant {
	case family + ".these":
		return "these", true
	case family + ".those":
		return "those", true
	case family + ".this", family + ".this.an", family + ".this.mass":
		return "this", true
	case family + ".that", family + ".that.an", family + ".that.mass":
		return "that", true
	case "construction:en.identification.it", "construction:en.identification.it.an", "construction:en.identification.it.mass", "construction:en.identification_negative.it", "construction:en.identification_negative.it.an", "construction:en.identification_negative.it.mass":
		if family == "construction:en.identification" || family == "construction:en.identification_negative" {
			return "it", true
		}
	}
	return "", false
}

var _ realizer.TargetCandidatePolicy = englishDeterminerCandidatePolicy{}
var _ runtime.TargetPlanBuilder = englishSentencePlanAnnotations{}
