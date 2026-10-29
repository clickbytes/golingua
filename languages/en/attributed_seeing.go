package en

import (
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type englishAttributedSeeingGraph struct {
	event, subject, object meaning.Node
	properties             map[meaning.NodeID]meaning.Node
	propertyByBearer       map[meaning.NodeID]meaning.Node
}

func englishAttributedSeeingShape(source semantic.MeaningCandidate) (englishAttributedSeeingGraph, bool) {
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil {
		return englishAttributedSeeingGraph{}, false
	}
	graph := *source.Meaning
	shape := englishAttributedSeeingGraph{properties: map[meaning.NodeID]meaning.Node{}, propertyByBearer: map[meaning.NodeID]meaning.Node{}}
	referents := make(map[meaning.NodeID]meaning.Node, 2)
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindEvent:
			if shape.event.ID != "" {
				return englishAttributedSeeingGraph{}, false
			}
			shape.event = node
		case meaning.KindState:
			if _, duplicate := shape.properties[node.ID]; duplicate {
				return englishAttributedSeeingGraph{}, false
			}
			shape.properties[node.ID] = node
		case meaning.KindReferent:
			if _, duplicate := referents[node.ID]; duplicate {
				return englishAttributedSeeingGraph{}, false
			}
			referents[node.ID] = node
		default:
			return englishAttributedSeeingGraph{}, false
		}
	}
	if shape.event.ID == "" || shape.event.Kind != meaning.KindEvent || shape.event.Concept != "concept:seeing" || shape.event.TemporalFrame != meaning.TemporalFramePresent ||
		len(referents) != 2 || len(shape.properties) < 1 || len(shape.properties) > 2 || len(shape.event.Participants) != 2 {
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
	var subjectOK, objectOK bool
	shape.subject, subjectOK = referents[experiencer]
	shape.object, objectOK = referents[stimulus]
	if !subjectOK || !objectOK || shape.subject.ID == shape.object.ID || !englishSeeingMultiplicitySupported(shape.subject.Multiplicity) ||
		!englishSeeingMultiplicitySupported(shape.object.Multiplicity) && !(shape.object.Multiplicity == "" && shape.object.Concept == "concept:door") {
		return englishAttributedSeeingGraph{}, false
	}
	for id, property := range shape.properties {
		if !englishSeeingColorConcept(string(property.Concept)) || property.Kind != meaning.KindState || property.TemporalFrame != "" || len(property.Participants) != 1 || property.Participants[0].Role != "bearer" {
			return englishAttributedSeeingGraph{}, false
		}
		bearer := property.Participants[0].Target
		if bearer != shape.subject.ID && bearer != shape.object.ID {
			return englishAttributedSeeingGraph{}, false
		}
		if _, duplicate := shape.propertyByBearer[bearer]; duplicate {
			return englishAttributedSeeingGraph{}, false
		}
		shape.propertyByBearer[bearer] = property
		if property.ID != id {
			return englishAttributedSeeingGraph{}, false
		}
	}
	rootIDs := map[meaning.NodeID]bool{}
	for _, root := range graph.Roots {
		rootIDs[root] = true
	}
	if len(rootIDs) != 1+len(shape.properties) || !rootIDs[shape.event.ID] {
		return englishAttributedSeeingGraph{}, false
	}
	for id := range shape.properties {
		if !rootIDs[id] {
			return englishAttributedSeeingGraph{}, false
		}
	}
	return shape, true
}

func englishSeeingColorConcept(concept string) bool {
	switch concept {
	case "concept:black", "concept:red", "concept:white", "concept:blue", "concept:green", "concept:open_state", "concept:sad":
		return true
	default:
		return false
	}
}

func englishSeeingMultiplicitySupported(multiplicity meaning.Multiplicity) bool {
	return multiplicity == meaning.MultiplicityOne || multiplicity == meaning.MultiplicityMany
}

func englishAttributedSeeingFamilyID(id string) bool {
	return id == "construction:en.attributed_seeing" || id == "construction:en.attributed_seeing_object" || id == "construction:en.attributed_seeing_both"
}

func englishAttributedSeeingSupportVariantAllowed(source semantic.MeaningCandidate, property meaning.Node, variant string) bool {
	if len(property.Participants) != 1 {
		return false
	}
	bearer, exists := englishMeaningNode(source, property.Participants[0].Target)
	if !exists || bearer.Kind != meaning.KindReferent {
		return false
	}
	if bearer.Multiplicity == meaning.MultiplicityMany {
		return variant == "construction:en.attribution.these" || variant == "construction:en.attribution.those"
	}
	if bearer.Multiplicity == "" && bearer.Concept == "concept:door" {
		return variant == "construction:en.attribution.this" || variant == "construction:en.attribution.that" ||
			variant == "construction:en.attribution.these" || variant == "construction:en.attribution.those"
	}
	return bearer.Multiplicity == meaning.MultiplicityOne && (variant == "construction:en.attribution.this" || variant == "construction:en.attribution.that" || variant == "construction:en.attribution.bare" || variant == "construction:en.attribution.the" || variant == "construction:en.attribution.a" || variant == "construction:en.attribution.an")
}

func englishAttributedSeeingDoorNumberChoice(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, slotID, role string) (string, bool) {
	shape, valid := englishAttributedSeeingShape(source)
	if !valid || predicate.NodeID != shape.event.ID || predicate.Family.Family.ID.String() != "construction:en.attributed_seeing_object" ||
		slotID != "object" || role != "stimulus" || shape.object.Concept != "concept:door" || shape.object.Multiplicity != "" {
		return "", false
	}
	if englishVariantRequiresNPNumber(predicate, slotID, "plural") {
		return "plural", true
	}
	if englishVariantRequiresNPNumber(predicate, slotID, "singular") {
		return "singular", true
	}
	return "", false
}

func englishAttributedSeeingDeixis(source semantic.MeaningCandidate) string {
	if source.Meaning == nil {
		return ""
	}
	shape, ok := englishAttributedSeeingShape(source)
	if !ok {
		return ""
	}
	values := map[string]bool{}
	for _, paired := range source.Contexts {
		for _, mention := range paired.Context.Mentions {
			if mention.NodeID == shape.object.ID && mention.Anchor != nil && mention.Anchor.Kind == "deixis" {
				values[mention.Anchor.Value] = true
			}
		}
	}
	if len(values) == 0 {
		return ""
	}
	if len(values) > 1 {
		return "ambiguous"
	}
	for value := range values {
		return value
	}
	return ""
}

// englishAttributedSeeingDeixisAllows reports whether a target demonstrative
// remains possible under every explicit source-context alternative. An
// unresolved demonstrative permits both distance classes; conflicting anchors
// are therefore preserved as candidates instead of being resolved by order.
func englishAttributedSeeingDeixisAllows(source semantic.MeaningCandidate, article string) bool {
	if source.Meaning == nil {
		return false
	}
	shape, ok := englishAttributedSeeingShape(source)
	if !ok {
		return false
	}
	values := map[string]bool{}
	for _, paired := range source.Contexts {
		for _, mention := range paired.Context.Mentions {
			if mention.NodeID == shape.object.ID && mention.Anchor != nil && mention.Anchor.Kind == "deixis" {
				values[mention.Anchor.Value] = true
			}
		}
	}
	if len(values) == 0 {
		return false
	}
	proximal := article == "this" || article == "these"
	distal := article == "that" || article == "those"
	if !proximal && !distal {
		return false
	}
	// Contexts are alternatives for the same meaning. A target candidate is
	// retained when at least one explicit context licenses it; retaining both
	// proximal and distal candidates preserves the unresolved source ambiguity.
	return values["demonstrative"] || (proximal && values["proximal"]) || (distal && values["distal"])
}

func englishAttributedSeeingFamily(candidate realizer.TargetCandidate) bool {
	return englishAttributedSeeingFamilyID(candidate.Family.Family.ID.String())
}

func englishAttributedSeeingFamilyMatches(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, shape englishAttributedSeeingGraph) bool {
	family := candidate.Family.Family
	familyID := family.ID.String()
	if !englishAttributedSeeingFamilyID(familyID) || candidate.NodeID != shape.event.ID {
		return false
	}
	modifierSlots := 0
	for _, slot := range family.SlotSpecs {
		if slot.ModifierBinding == nil {
			continue
		}
		modifierSlots++
		var targetSlot construction.SlotSpec
		foundTarget := false
		for _, possible := range family.SlotSpecs {
			if possible.ID == slot.ModifierBinding.TargetSlot {
				targetSlot, foundTarget = possible, true
				break
			}
		}
		if !foundTarget || targetSlot.ArgumentBinding == nil {
			return false
		}
		var expected meaning.Node
		switch targetSlot.ArgumentBinding.SemanticRole {
		case "experiencer":
			expected = shape.propertyByBearer[shape.subject.ID]
		case "stimulus":
			expected = shape.propertyByBearer[shape.object.ID]
		default:
			return false
		}
		if expected.ID == "" {
			return false
		}
		modifier, found := realizer.FindModifier(*source.Meaning, candidate.NodeID, family, slot.ID)
		if !found || modifier.ID != expected.ID {
			return false
		}
	}
	return modifierSlots == len(shape.properties)
}

func englishAttributedSeeingContextsValid(source semantic.MeaningCandidate) bool {
	if source.Meaning == nil {
		return false
	}
	var anyAnchored, anyUnanchored bool
	for _, candidateContext := range source.Contexts {
		if meaning.ValidateReferences(candidateContext.Meaning, candidateContext.Context, meaning.DefaultBudget()) != nil {
			return false
		}
		equal, err := meaning.CanonicalEqual(candidateContext.Meaning, *source.Meaning, meaning.DefaultBudget())
		if err != nil || !equal {
			return false
		}
		deixisCount := 0
		for _, mention := range candidateContext.Context.Mentions {
			if mention.Anchor == nil {
				if mention.ReferentID != "" || mention.EvidenceID != "" {
					return false
				}
				continue
			}
			shape, valid := englishAttributedSeeingShape(source)
			if mention.Anchor.Kind != "deixis" ||
				(mention.Anchor.Value != "demonstrative" && mention.Anchor.Value != "proximal" && mention.Anchor.Value != "distal") ||
				!valid || (mention.NodeID != shape.subject.ID && mention.NodeID != shape.object.ID) || mention.ReferentID != "" || mention.EvidenceID != "" {
				return false
			}
			deixisCount++
		}
		if len(candidateContext.Context.Referents) != 0 || deixisCount > 2 {
			return false
		}
		if deixisCount == 1 {
			anyAnchored = true
		} else {
			anyUnanchored = true
		}
	}
	if anyAnchored && anyUnanchored {
		return false
	}
	return true
}

func englishAttributedSeeingVariant(candidate realizer.TargetCandidate) bool {
	variant := candidate.CompiledVariant.Variant.ID.String()
	switch candidate.Family.Family.ID.String() {
	case "construction:en.attributed_seeing":
		return strings.HasPrefix(variant, "construction:en.attributed_seeing.svo")
	case "construction:en.attributed_seeing_object":
		return strings.HasPrefix(variant, "construction:en.attributed_seeing_object.svo")
	case "construction:en.attributed_seeing_both":
		return strings.HasPrefix(variant, "construction:en.attributed_seeing_both.svo")
	default:
		return false
	}
}

func englishAttributedSeeingSubjectAnchors(source semantic.MeaningCandidate, subjectID meaning.NodeID) map[string]bool {
	values := map[string]bool{}
	if source.Meaning == nil || len(source.Contexts) == 0 {
		return values
	}
	for _, paired := range source.Contexts {
		value := ""
		for _, mention := range paired.Context.Mentions {
			if mention.NodeID != subjectID || mention.Anchor == nil {
				continue
			}
			if mention.Anchor.Kind != "deixis" ||
				(mention.Anchor.Value != "proximal" && mention.Anchor.Value != "distal" && mention.Anchor.Value != "demonstrative") ||
				mention.ReferentID != "" || mention.EvidenceID != "" || value != "" {
				return map[string]bool{}
			}
			value = mention.Anchor.Value
		}
		if value == "" {
			return map[string]bool{}
		}
		values[value] = true
	}
	return values
}

func englishAttributedSeeingSubjectVariantAllowed(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, shape englishAttributedSeeingGraph) bool {
	variant := candidate.CompiledVariant.Variant.ID.String()
	if candidate.Family.Family.ID.String() == "construction:en.attributed_seeing_both" && englishAttributedSeeingDeixis(source) != "" {
		// The bounded two-AP route carries the subject anchor only. Do not let
		// an object or dual anchor pass through a target variant that only says
		// "the" for the object NP.
		return false
	}
	article, found := englishDeterminersForCandidate(candidate)
	if !found {
		return false
	}
	singular := englishVariantRequiresNPNumber(candidate, "subject", "singular")
	plural := englishVariantRequiresNPNumber(candidate, "subject", "plural")
	if singular == plural {
		return false
	}
	if shape.subject.Multiplicity == meaning.MultiplicityOne && !singular || shape.subject.Multiplicity == meaning.MultiplicityMany && !plural {
		return false
	}
	if shape.subject.Multiplicity == "" {
		return false
	}
	switch article.subject {
	case "this", "that", "these", "those":
		anchors := englishAttributedSeeingSubjectAnchors(source, shape.subject.ID)
		if len(anchors) == 0 {
			return false
		}
		proximal := article.subject == "this" || article.subject == "these"
		distal := article.subject == "that" || article.subject == "those"
		if proximal && !anchors["proximal"] && !anchors["demonstrative"] || distal && !anchors["distal"] && !anchors["demonstrative"] {
			return false
		}
		if singular && article.subject != "this" && article.subject != "that" || plural && article.subject != "these" && article.subject != "those" {
			return false
		}
		prefix := "construction:en.attributed_seeing.svo.demonstrative_"
		if candidate.Family.Family.ID.String() == "construction:en.attributed_seeing_both" {
			prefix = "construction:en.attributed_seeing_both.svo.demonstrative_"
		}
		return strings.HasPrefix(variant, prefix)
	default:
		if len(englishAttributedSeeingSubjectAnchors(source, shape.subject.ID)) != 0 {
			return false
		}
		return !strings.HasPrefix(variant, "construction:en.attributed_seeing.svo.demonstrative_") &&
			!strings.HasPrefix(variant, "construction:en.attributed_seeing_both.svo.demonstrative_")
	}
}

func englishAttributedSeeingMainCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) bool {
	shape, ok := englishAttributedSeeingShape(source)
	if !ok || !englishAttributedSeeingContextsValid(source) || candidate.NodeID != shape.event.ID || candidate.Kind != meaning.KindEvent || candidate.Outcome != status.OutcomeOK {
		return false
	}
	if candidate.Sense.ID.String() != "sense:en.see.visual" || string(candidate.Frame.ID) != "en.see.visual.transitive" ||
		candidate.TransferBinding.CoreRef.String() != "concept:seeing" || !englishAttributedSeeingFamily(candidate) || !englishAttributedSeeingVariant(candidate) ||
		!realizer.CoversLexicalMeaning(*source.Meaning, candidate.NodeID, candidate.Family.Family) {
		return false
	}
	return englishAttributedSeeingFamilyMatches(source, candidate, shape) && englishAttributedSeeingSubjectVariantAllowed(source, candidate, shape)
}

func englishAttributedSeeingModifierCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, all []realizer.TargetCandidate) bool {
	shape, ok := englishAttributedSeeingShape(source)
	if !ok || !englishAttributedSeeingContextsValid(source) || candidate.Kind != meaning.KindState || candidate.Outcome != status.OutcomeOK {
		return false
	}
	property, exists := shape.properties[candidate.NodeID]
	if !exists {
		return false
	}
	spec, licensed := englishAttributionForConcept(string(property.Concept))
	if !licensed || candidate.Sense.ID.String() != spec.sense ||
		string(candidate.Frame.ID) != strings.TrimPrefix(spec.sense, "sense:")+".frame" ||
		candidate.Lexeme.ID.String() != spec.lexeme ||
		candidate.TransferBinding.CoreRef.String() != spec.concept || candidate.Family.Family.ID.String() != "construction:en.attribution" {
		return false
	}
	variant := candidate.CompiledVariant.Variant.ID.String()
	if !englishAttributedSeeingSupportVariantAllowed(source, property, variant) {
		return false
	}
	bestVariant := ""
	for _, support := range all {
		if support.SourceCandidateID != candidate.SourceCandidateID || support.SemanticID != candidate.SemanticID || support.NodeID != candidate.NodeID ||
			support.Kind != meaning.KindState || support.Outcome != status.OutcomeOK || support.Sense.ID.String() != candidate.Sense.ID.String() ||
			support.Lexeme.ID.String() != candidate.Lexeme.ID.String() || support.TransferBinding.CoreRef.String() != candidate.TransferBinding.CoreRef.String() || support.Family.Family.ID.String() != "construction:en.attribution" {
			continue
		}
		supportVariant := support.CompiledVariant.Variant.ID.String()
		if !englishAttributedSeeingSupportVariantAllowed(source, property, supportVariant) {
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
