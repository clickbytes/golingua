package pl

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

func polishMeaningNode(source semantic.MeaningCandidate, nodeID meaning.NodeID) (meaning.Node, bool) {
	if source.Meaning == nil {
		return meaning.Node{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}
	return meaning.Node{}, false
}

func polishReferentForRole(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role) (meaning.Node, bool) {
	predicate, exists := polishMeaningNode(source, predicateID)
	if !exists || predicate.Kind != meaning.KindEvent && predicate.Kind != meaning.KindState {
		return meaning.Node{}, false
	}
	var target meaning.NodeID
	for _, participant := range predicate.Participants {
		if participant.Role != role {
			continue
		}
		if target != "" {
			return meaning.Node{}, false
		}
		target = participant.Target
	}
	referent, exists := polishMeaningNode(source, target)
	return referent, exists && referent.Kind == meaning.KindReferent
}

func polishMeaningNumber(multiplicity meaning.Multiplicity) string {
	switch multiplicity {
	case meaning.MultiplicityOne:
		return "singular"
	case meaning.MultiplicityMany:
		return "plural"
	default:
		return ""
	}
}

func polishFrameEvidenceConsistent(source semantic.MeaningCandidate, predicateID meaning.NodeID, concept ids.ConceptID) bool {
	node, exists := polishMeaningNode(source, predicateID)
	if !exists || node.TemporalFrame == "" {
		return true
	}
	values, found := runtime.SourceFeaturesAt(source, predicateID, "", concept)
	if !found {
		return !polishSourceFeaturesPresent(source, predicateID, "", concept)
	}
	if tense := string(values["tense"]); tense != "" && tense != string(node.TemporalFrame) {
		return false
	}
	if values["verb_form"] == "past" && node.TemporalFrame != meaning.TemporalFramePast {
		return false
	}
	if values["verb_form"] == "third_person_singular" && node.TemporalFrame != meaning.TemporalFramePresent {
		return false
	}
	return true
}

func polishMultiplicityEvidenceConsistent(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, referent meaning.Node, predicateConcept ids.ConceptID, checkPredicateNumber bool) bool {
	expected := polishMeaningNumber(referent.Multiplicity)
	if expected == "" {
		return true
	}
	predicate, exists := polishMeaningNode(source, predicateID)
	if !exists {
		return false
	}
	participantFound := false
	for _, participant := range predicate.Participants {
		if participant.Role == role && participant.Target == referent.ID {
			if participantFound {
				return false
			}
			participantFound = true
		}
	}
	if !participantFound {
		return false
	}
	if concept, err := ids.ParseConceptID(string(referent.Concept)); err == nil {
		values, found := runtime.SourceFeaturesAt(source, predicateID, role, concept)
		if !found {
			if polishSourceFeaturesPresent(source, predicateID, role, concept) {
				return false
			}
		} else if !polishNumberFeaturesAgree(values, expected) {
			return false
		}
	}
	if checkPredicateNumber {
		values, found := runtime.SourceFeaturesAt(source, predicateID, "", predicateConcept)
		if !found {
			if polishSourceFeaturesPresent(source, predicateID, "", predicateConcept) {
				return false
			}
		} else if !polishNumberFeaturesAgree(values, expected) {
			return false
		}
	}
	return true
}

func polishSourceFeaturesPresent(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, concept ids.ConceptID) bool {
	return runtime.SourceFeaturesPresentAt(source, predicateID, role, concept)
}

func polishNumberFeaturesAgree(values map[features.FeatureID]features.FeatureValueID, expected string) bool {
	observed := ""
	for _, feature := range []features.FeatureID{"number", "noun_number", "verb_number"} {
		value := string(values[feature])
		if value == "" {
			continue
		}
		if value != "singular" && value != "plural" {
			return false
		}
		if observed != "" && observed != value {
			return false
		}
		observed = value
	}
	return observed == "" || observed == expected
}

func polishTargetNumber(source semantic.MeaningCandidate, referentID meaning.NodeID, properties []lexicon.Property, paradigmID string) (string, bool, bool) {
	referent, exists := polishMeaningNode(source, referentID)
	if !exists || referent.Kind != meaning.KindReferent {
		return "", false, false
	}
	behavior, grammaticalNumber := "", ""
	for _, property := range properties {
		switch property.Key {
		case "number_behavior":
			behavior = property.Value
		case "grammatical_number":
			grammaticalNumber = property.Value
		}
	}
	if grammaticalNumber == "plural_only" || paradigmID == "pl_noun_plural_only_attested" {
		lexicalPluralOnly := false
		for _, property := range properties {
			if property.Key == "multiplicity_behavior" && property.Value == "lexical_plural_not_cardinality" {
				lexicalPluralOnly = true
			}
		}
		// For pluralia tantum, the plural form is lexical and does not assert
		// cardinality; Meaning keeps cardinality independent of noun morphology.
		if referent.Multiplicity != "" && !lexicalPluralOnly {
			return "", false, false
		}
		return "plural", false, true
	}
	if behavior == "mass" {
		if referent.Multiplicity != "" {
			return "", false, false
		}
		return "singular", false, true
	}
	if number := polishMeaningNumber(referent.Multiplicity); number != "" {
		return number, true, true
	}
	return "singular", false, true
}

func polishPresentNegationNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, route polishPresentNegationRoute, candidates []realizer.TargetCandidate) (string, bool) {
	_, event, sourceRoute, found := polishNegatedPresent(source)
	if !found || event.ID != predicate.NodeID || sourceRoute.family != route.family || sourceRoute.sense != route.sense || event.TemporalFrame == meaning.TemporalFramePast {
		return "", false
	}
	referent, exists := polishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || route.concept == "concept:class_membership" && referent.Multiplicity == meaning.MultiplicityMany {
		return "", false
	}
	if !polishFrameEvidenceConsistent(source, event.ID, predicate.TransferBinding.CoreRef) ||
		!polishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, predicate.TransferBinding.CoreRef, true) {
		return "", false
	}
	return polishNumberForReferentCandidates(source, referent.ID, route.subjectRole, event.ID, predicate.TransferBinding.CoreRef, candidates)
}

func polishNumberForReferentCandidates(source semantic.MeaningCandidate, referentID meaning.NodeID, role meaning.Role, predicateID meaning.NodeID, predicateConcept ids.ConceptID, candidates []realizer.TargetCandidate) (string, bool) {
	number := ""
	found := false
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != referentID || candidate.Outcome != status.OutcomeOK {
			continue
		}
		candidateNumber, semanticNumber, ok := polishTargetNumber(source, referentID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if !ok {
			return "", false
		}
		if semanticNumber && !polishMultiplicityEvidenceConsistent(source, predicateID, role, mustPolishReferent(source, referentID), predicateConcept, true) {
			return "", false
		}
		if found && number != candidateNumber {
			return "", false
		}
		number, found = candidateNumber, true
	}
	return number, found
}

func mustPolishReferent(source semantic.MeaningCandidate, referentID meaning.NodeID) meaning.Node {
	referent, _ := polishMeaningNode(source, referentID)
	return referent
}

func polishIdentificationAttributesAllowed(source semantic.MeaningCandidate, predicateID meaning.NodeID, role meaning.Role, referentID meaning.NodeID, coreID ids.ConceptID) bool {
	predicate, predicateExists := polishMeaningNode(source, predicateID)
	referent, referentExists := polishMeaningNode(source, referentID)
	if !predicateExists || !referentExists || referent.Kind != meaning.KindReferent {
		return false
	}
	if predicate.TemporalFrame == meaning.TemporalFramePast {
		if role != "bearer" {
			return false
		}
		if _, attribution := polishAttributionForConcept(string(predicate.Concept)); !attribution {
			return false
		}
	}
	return polishFrameEvidenceConsistent(source, predicateID, coreID) &&
		polishMultiplicityEvidenceConsistent(source, predicateID, role, referent, coreID, true)
}

func polishPredicateEvidenceConsistent(source semantic.MeaningCandidate, predicate realizer.TargetCandidate) bool {
	if !polishFrameEvidenceConsistent(source, predicate.NodeID, predicate.TransferBinding.CoreRef) {
		return false
	}
	meaningPredicate, exists := polishMeaningNode(source, predicate.NodeID)
	if !exists {
		return false
	}
	if meaningPredicate.TemporalFrame == meaning.TemporalFramePast && predicate.TransferBinding.CoreRef.String() == "concept:giving" {
		targetAspect := ""
		switch predicate.TransferBinding.SenseID.String() {
		case "sense:pl.dac.transfer":
			targetAspect = "perfective"
		case "sense:pl.dawac.transfer":
			targetAspect = "imperfective"
		}
		if targetAspect != "" {
			values, known := runtime.SourceFeaturesAt(source, predicate.NodeID, "", predicate.TransferBinding.CoreRef)
			if known && string(values["aspect"]) != "" && string(values["aspect"]) != targetAspect {
				return false
			}
		}
	}
	identificationRole := meaning.Role("")
	switch predicate.TransferBinding.CoreRef.String() {
	case "concept:class_membership":
		identificationRole = "member"
	default:
		if _, attributable := polishAttributionForConcept(predicate.TransferBinding.CoreRef.String()); attributable {
			identificationRole = "bearer"
		}
	}
	for _, participant := range meaningPredicate.Participants {
		referent, exists := polishMeaningNode(source, participant.Target)
		if !exists || referent.Kind != meaning.KindReferent || referent.Multiplicity == "" {
			continue
		}
		checkPredicateNumber := participant.Role == identificationRole && identificationRole != ""
		for _, argument := range predicate.Frame.Arguments {
			if argument.Role == participant.Role && argument.GrammaticalFunction == "subject" {
				checkPredicateNumber = true
			}
		}
		if !polishMultiplicityEvidenceConsistent(source, predicate.NodeID, participant.Role, referent, predicate.TransferBinding.CoreRef, checkPredicateNumber) {
			return false
		}
	}
	return true
}

func polishPredicateSubjectRole(predicate realizer.TargetCandidate, role meaning.Role) bool {
	if predicate.TransferBinding.CoreRef.String() == "concept:class_membership" && role == "member" ||
		role == "bearer" && func() bool {
			_, attributable := polishAttributionForConcept(predicate.TransferBinding.CoreRef.String())
			return attributable
		}() {
		return true
	}
	for _, argument := range predicate.Frame.Arguments {
		if argument.Role == role && argument.GrammaticalFunction == "subject" {
			return true
		}
	}
	return false
}
