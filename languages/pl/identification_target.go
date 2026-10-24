package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func polishIdentificationPredicateConcept(source semantic.MeaningCandidate, predicateID meaning.NodeID) string {
	if source.Meaning == nil {
		return ""
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID == predicateID && node.Kind == meaning.KindState {
			switch string(node.Concept) {
			case "concept:class_membership", "concept:black":
				return string(node.Concept)
			}
		}
	}
	return ""
}

func polishIdentificationState(source semantic.MeaningCandidate, predicateID meaning.NodeID) (meaning.Role, meaning.NodeID, string, bool) {
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil {
		return "", "", "", false
	}
	graph := *source.Meaning
	var predicate meaning.Node
	for _, node := range graph.Nodes {
		if node.ID == predicateID {
			predicate = node
			break
		}
	}
	if predicate.ID == "" || predicate.Kind != meaning.KindState || len(graph.Nodes) != 2 || len(graph.Roots) != 1 || graph.Roots[0] != predicate.ID || len(predicate.Participants) != 1 {
		return "", "", "", false
	}
	var role meaning.Role
	switch string(predicate.Concept) {
	case "concept:class_membership":
		role = "member"
	case "concept:black":
		role = "bearer"
	default:
		return "", "", "", false
	}
	participant := predicate.Participants[0]
	if participant.Role != role {
		return "", "", "", false
	}
	var referent meaning.Node
	for _, node := range graph.Nodes {
		if node.ID == participant.Target {
			referent = node
			break
		}
	}
	if referent.ID == "" || referent.Kind != meaning.KindReferent || string(referent.Concept) != "concept:cat" {
		return "", "", "", false
	}
	return role, referent.ID, string(predicate.Concept), true
}

func polishNegatedClassificationState(source semantic.MeaningCandidate, predicateID meaning.NodeID) (meaning.NodeID, meaning.NodeID, bool) {
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil {
		return "", "", false
	}
	graph := *source.Meaning
	if len(graph.Roots) != 1 || len(graph.Nodes) != 3 {
		return "", "", false
	}
	var scope, predicate, referent meaning.Node
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindScope:
			if scope.ID != "" {
				return "", "", false
			}
			scope = node
		case meaning.KindState:
			if predicate.ID != "" {
				return "", "", false
			}
			predicate = node
		case meaning.KindReferent:
			if referent.ID != "" {
				return "", "", false
			}
			referent = node
		default:
			return "", "", false
		}
	}
	if scope.ID == "" || scope.ID != graph.Roots[0] || scope.Operator != meaning.OperatorNegation || scope.Target != predicateID ||
		predicate.ID != predicateID || predicate.Concept != "concept:class_membership" || len(predicate.Participants) != 1 ||
		predicate.Participants[0].Role != "member" || predicate.Participants[0].Target != referent.ID ||
		referent.ID == "" || referent.Concept != "concept:cat" {
		return "", "", false
	}
	return scope.ID, referent.ID, true
}

func polishNegatedClassificationReferenceMode(source semantic.MeaningCandidate, predicateID meaning.NodeID) (string, bool) {
	_, sourceReferentID, ok := polishNegatedClassificationState(source, predicateID)
	if !ok || source.Meaning == nil || len(source.Contexts) == 0 {
		return "", false
	}
	sourcePredicate, _ := polishMeaningNode(source, predicateID)
	sourceReferent, _ := polishMeaningNode(source, sourceReferentID)
	var commonMode string
	for _, candidateContext := range source.Contexts {
		graph, context := candidateContext.Meaning, candidateContext.Context
		if meaning.ValidateReferences(graph, context, meaning.DefaultBudget()) != nil || len(graph.Roots) != 1 || len(graph.Nodes) != 3 ||
			len(context.Referents) != 0 || len(context.Mentions) != 1 {
			return "", false
		}
		var scope, predicate, referent meaning.Node
		for _, node := range graph.Nodes {
			switch node.Kind {
			case meaning.KindScope:
				if scope.ID != "" {
					return "", false
				}
				scope = node
			case meaning.KindState:
				if predicate.ID != "" {
					return "", false
				}
				predicate = node
			case meaning.KindReferent:
				if referent.ID != "" {
					return "", false
				}
				referent = node
			default:
				return "", false
			}
		}
		if scope.ID == "" || graph.Roots[0] != scope.ID || scope.Operator != meaning.OperatorNegation || scope.Target != predicate.ID ||
			predicate.ID == "" || predicate.Concept != "concept:class_membership" || len(predicate.Participants) != 1 ||
			predicate.Participants[0].Role != "member" || predicate.Participants[0].Target != referent.ID ||
			referent.ID == "" || referent.Concept != "concept:cat" {
			return "", false
		}
		if sourceReferent := sourceNodeConcept(*source.Meaning, sourceReferentID); sourceReferent != referent.Concept {
			return "", false
		}
		if predicate.TemporalFrame != sourcePredicate.TemporalFrame || referent.Multiplicity != sourceReferent.Multiplicity {
			return "", false
		}
		mention := context.Mentions[0]
		if mention.NodeID != referent.ID || mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil || mention.Anchor.Kind != "deixis" {
			return "", false
		}
		mode := mention.Anchor.Value
		if mode != "proximal" && mode != "unresolved" {
			return "", false
		}
		if commonMode == "" {
			commonMode = mode
		} else if commonMode != mode {
			return "", false
		}
	}
	return commonMode, commonMode != ""
}

func polishIdentificationReferenceMode(source semantic.MeaningCandidate, predicateID meaning.NodeID) (string, bool) {
	role, sourceReferentID, concept, ok := polishIdentificationState(source, predicateID)
	if !ok || len(source.Contexts) == 0 {
		return "", false
	}
	sourcePredicate, _ := polishMeaningNode(source, predicateID)
	sourceReferent, _ := polishMeaningNode(source, sourceReferentID)
	var commonMode string
	for _, candidateContext := range source.Contexts {
		graph := candidateContext.Meaning
		context := candidateContext.Context
		if meaning.ValidateReferences(graph, context, meaning.DefaultBudget()) != nil || len(context.Referents) != 0 || len(context.Mentions) != 1 || len(graph.Nodes) != 2 || len(graph.Roots) != 1 {
			return "", false
		}
		mention := context.Mentions[0]
		if mention.ReferentID != "" || mention.EvidenceID != "" || mention.Anchor == nil || mention.Anchor.Kind != "deixis" {
			return "", false
		}
		switch mention.Anchor.Value {
		case "proximal", "unresolved", "demonstrative":
		default:
			return "", false
		}
		if commonMode == "" {
			commonMode = mention.Anchor.Value
		} else if commonMode != mention.Anchor.Value {
			return "", false
		}
		var contextPredicate, contextReferent meaning.Node
		for _, node := range graph.Nodes {
			if node.Kind == meaning.KindState && string(node.Concept) == concept {
				if contextPredicate.ID != "" || len(node.Participants) != 1 || node.Participants[0].Role != role {
					return "", false
				}
				contextPredicate = node
				for _, target := range graph.Nodes {
					if target.ID == node.Participants[0].Target {
						contextReferent = target
						break
					}
				}
			}
		}
		if contextPredicate.ID == "" || contextReferent.ID == "" || contextReferent.Kind != meaning.KindReferent ||
			string(contextReferent.Concept) != string(sourceNodeConcept(*source.Meaning, sourceReferentID)) || mention.NodeID != contextReferent.ID ||
			contextPredicate.TemporalFrame != sourcePredicate.TemporalFrame || contextReferent.Multiplicity != sourceReferent.Multiplicity ||
			len(graph.Roots) != 1 || graph.Roots[0] != contextPredicate.ID {
			return "", false
		}
	}
	return commonMode, commonMode != ""
}

func sourceNodeConcept(graph meaning.Meaning, nodeID meaning.NodeID) meaning.ConceptRef {
	for _, node := range graph.Nodes {
		if node.ID == nodeID {
			return node.Concept
		}
	}
	return ""
}

func polishIdentificationReferentFeatures(source semantic.MeaningCandidate, referentID meaning.NodeID, candidates []realizer.TargetCandidate) (string, string, bool) {
	gender, number := "", ""
	found := false
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != referentID || candidate.Outcome != status.OutcomeOK || candidate.TransferBinding.CoreRef.String() != "concept:cat" {
			continue
		}
		candidateGender := ""
		candidateNumber, _, numberOK := polishTargetNumber(source, referentID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if !numberOK {
			return "", "", false
		}
		for _, property := range candidate.Lexeme.Properties {
			switch property.Key {
			case "gender":
				candidateGender = property.Value
			}
		}
		if candidateGender == "" || found && (gender != candidateGender || number != candidateNumber) {
			return "", "", false
		}
		gender, number, found = candidateGender, candidateNumber, true
	}
	return gender, number, found
}
