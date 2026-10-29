package pl

import (
	"strings"

	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type polishAttributedSeeingGraph struct {
	event, subject, object meaning.Node
	scope                  meaning.Node
	properties             map[meaning.NodeID]meaning.Node
	propertyByBearer       map[meaning.NodeID]meaning.Node
}

func polishAttributedSeeingShape(source semantic.MeaningCandidate) (polishAttributedSeeingGraph, bool) {
	if _, _, _, ok := polishAttributedPredicateBinding(source); !ok {
		return polishAttributedSeeingGraph{}, false
	}
	if source.Meaning == nil || meaning.ValidateStructure(*source.Meaning, meaning.DefaultBudget()) != nil {
		return polishAttributedSeeingGraph{}, false
	}
	graph := *source.Meaning
	shape := polishAttributedSeeingGraph{properties: map[meaning.NodeID]meaning.Node{}, propertyByBearer: map[meaning.NodeID]meaning.Node{}}
	referents := make(map[meaning.NodeID]meaning.Node, 2)
	for _, node := range graph.Nodes {
		switch node.Kind {
		case meaning.KindEvent:
			if shape.event.ID != "" {
				return polishAttributedSeeingGraph{}, false
			}
			shape.event = node
		case meaning.KindState:
			if _, duplicate := shape.properties[node.ID]; duplicate {
				return polishAttributedSeeingGraph{}, false
			}
			shape.properties[node.ID] = node
		case meaning.KindReferent:
			if _, duplicate := referents[node.ID]; duplicate {
				return polishAttributedSeeingGraph{}, false
			}
			referents[node.ID] = node
		case meaning.KindScope:
			if shape.scope.ID != "" || node.Operator != meaning.OperatorNegation {
				return polishAttributedSeeingGraph{}, false
			}
			shape.scope = node
		default:
			return polishAttributedSeeingGraph{}, false
		}
	}
	if shape.event.ID == "" || shape.event.Kind != meaning.KindEvent || shape.event.TemporalFrame != meaning.TemporalFramePresent ||
		len(referents) != 2 || len(shape.properties) < 1 || len(shape.properties) > 2 || len(shape.event.Participants) != 2 {
		return polishAttributedSeeingGraph{}, false
	}
	if shape.scope.ID != "" && (shape.scope.Target != shape.event.ID || string(shape.event.Concept) != "concept:seeing") {
		return polishAttributedSeeingGraph{}, false
	}
	var experiencer, stimulus meaning.NodeID
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
	var subjectOK, objectOK bool
	shape.subject, subjectOK = referents[experiencer]
	shape.object, objectOK = referents[stimulus]
	if !subjectOK || !objectOK || shape.subject.ID == shape.object.ID || !polishSeeingMultiplicitySupported(shape.subject.Multiplicity) ||
		!polishSeeingMultiplicitySupported(shape.object.Multiplicity) && !(shape.object.Multiplicity == "" && shape.object.Concept == "concept:door") {
		return polishAttributedSeeingGraph{}, false
	}
	for id, property := range shape.properties {
		if !polishSeeingColorConcept(string(property.Concept)) || property.Kind != meaning.KindState || property.TemporalFrame != "" || len(property.Participants) != 1 || property.Participants[0].Role != "bearer" {
			return polishAttributedSeeingGraph{}, false
		}
		bearer := property.Participants[0].Target
		if bearer != shape.subject.ID && bearer != shape.object.ID {
			return polishAttributedSeeingGraph{}, false
		}
		if _, duplicate := shape.propertyByBearer[bearer]; duplicate {
			return polishAttributedSeeingGraph{}, false
		}
		shape.propertyByBearer[bearer] = property
		if property.ID != id {
			return polishAttributedSeeingGraph{}, false
		}
	}
	rootIDs := map[meaning.NodeID]bool{}
	for _, root := range graph.Roots {
		rootIDs[root] = true
	}
	expectedRoot := shape.event.ID
	if shape.scope.ID != "" {
		expectedRoot = shape.scope.ID
	}
	if len(rootIDs) != 1+len(shape.properties) || !rootIDs[expectedRoot] {
		return polishAttributedSeeingGraph{}, false
	}
	for id := range shape.properties {
		if !rootIDs[id] {
			return polishAttributedSeeingGraph{}, false
		}
	}
	return shape, true
}

func polishAttributedPredicateBinding(source semantic.MeaningCandidate) (sense, frame, core string, ok bool) {
	if source.Meaning == nil {
		return "", "", "", false
	}
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindEvent {
			continue
		}
		switch node.Concept {
		case "concept:seeing":
			if _, event, _, negated := polishNegatedPresent(source); negated && event.ID == node.ID {
				return "sense:pl.widziec.visual", "pl.widziec.visual.attributed_negative", "concept:seeing", true
			}
			return "sense:pl.widziec.visual", "pl.widziec.visual.transitive", "concept:seeing", true
		case "concept:hearing":
			var stimulus, experiencer meaning.NodeID
			for _, participant := range node.Participants {
				switch participant.Role {
				case "stimulus":
					stimulus = participant.Target
				case "experiencer":
					experiencer = participant.Target
				}
			}
			stimulusProperty, experiencerProperty := false, false
			pluralObject, pluralExperiencer := false, false
			for _, referenced := range source.Meaning.Nodes {
				if referenced.ID == stimulus && referenced.Kind == meaning.KindReferent && referenced.Multiplicity == meaning.MultiplicityMany {
					pluralObject = true
				}
				if referenced.ID == experiencer && referenced.Kind == meaning.KindReferent && referenced.Multiplicity == meaning.MultiplicityMany {
					pluralExperiencer = true
				}
				if referenced.Kind == meaning.KindState && len(referenced.Participants) == 1 && referenced.Participants[0].Role == "bearer" {
					stimulusProperty = stimulusProperty || referenced.Participants[0].Target == stimulus
					experiencerProperty = experiencerProperty || referenced.Participants[0].Target == experiencer
				}
			}
			if stimulusProperty {
				if pluralObject {
					return "sense:pl.slyszec.auditory", "pl.slyszec.auditory.attributed_object_plural", "concept:hearing", true
				}
				return "sense:pl.slyszec.auditory", "pl.slyszec.auditory.transitive", "concept:hearing", true
			}
			if experiencerProperty {
				if pluralExperiencer {
					return "sense:pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "concept:hearing", true
				}
				return "sense:pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "concept:hearing", true
			}
			return "sense:pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "concept:hearing", true
		default:
			return "", "", "", false
		}
	}
	return "", "", "", false
}

func polishSeeingColorConcept(concept string) bool {
	switch concept {
	case "concept:black", "concept:red", "concept:white", "concept:blue", "concept:green", "concept:open_state", "concept:sad":
		return true
	default:
		return false
	}
}

func polishSeeingMultiplicitySupported(multiplicity meaning.Multiplicity) bool {
	return multiplicity == meaning.MultiplicityOne || multiplicity == meaning.MultiplicityMany
}

func polishAttributedSeeingFamilyID(id string) bool {
	return id == "construction:pl.attributed_seeing" || id == "construction:pl.attributed_seeing_object" || id == "construction:pl.attributed_seeing_both" ||
		id == "construction:pl.attributed_seeing_plural" || id == "construction:pl.attributed_seeing_object_plural" || id == "construction:pl.attributed_seeing_both_plural" ||
		id == "construction:pl.attributed_seeing_object_negative"
}

func polishAttributedSeeingFamily(candidate realizer.TargetCandidate) bool {
	return polishAttributedSeeingFamilyID(candidate.Family.Family.ID.String())
}

func polishAttributedSeeingFamilyMatches(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, shape polishAttributedSeeingGraph) bool {
	family := candidate.Family.Family
	familyID := family.ID.String()
	if !polishAttributedSeeingFamilyID(familyID) || candidate.NodeID != shape.event.ID {
		return false
	}
	subjectMany := shape.subject.Multiplicity == meaning.MultiplicityMany
	objectMany := shape.object.Multiplicity == meaning.MultiplicityMany
	switch familyID {
	case "construction:pl.attributed_seeing", "construction:pl.attributed_seeing_object", "construction:pl.attributed_seeing_object_negative":
		if subjectMany || objectMany {
			return false
		}
	case "construction:pl.attributed_seeing_both":
		if subjectMany != objectMany {
			return false
		}
	case "construction:pl.attributed_seeing_both_plural":
		if !subjectMany && !objectMany {
			return false
		}
	case "construction:pl.attributed_seeing_plural":
		if !subjectMany || objectMany {
			return false
		}
	case "construction:pl.attributed_seeing_object_plural":
		if subjectMany || !objectMany {
			return false
		}
	default:
		return false
	}
	modifierSlots := 0
	for _, slot := range family.SlotSpecs {
		if slot.ModifierBinding == nil {
			continue
		}
		modifierSlots++
		targetSlot := family.SlotSpecs[0]
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

func polishAttributedSeeingContextsValid(source semantic.MeaningCandidate) bool {
	if source.Meaning == nil {
		return false
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
			if mention.ReferentID != "" || mention.EvidenceID != "" {
				return false
			}
			if mention.Anchor != nil && (mention.Anchor.Kind != "deixis" || mention.Anchor.Value != "proximal" && mention.Anchor.Value != "distal") {
				return false
			}
		}
	}
	return true
}

// polishAttributedSeeingDeixis returns one consistent explicit anchor for a
// role. Conflicting request contexts remain unresolved instead of choosing one.
func polishAttributedSeeingDeixis(source semantic.MeaningCandidate, referent meaning.NodeID) (string, bool) {
	if source.Meaning == nil || !polishAttributedSeeingContextsValid(source) {
		return "", false
	}
	mode := ""
	for _, candidateContext := range source.Contexts {
		for _, mention := range candidateContext.Context.Mentions {
			if mention.NodeID != referent {
				continue
			}
			value := "unresolved"
			if mention.Anchor != nil {
				value = mention.Anchor.Value
			}
			if mode != "" && mode != value {
				return "", false
			}
			mode = value
		}
	}
	if mode == "" {
		// No mention context leaves reference unresolved. Preserve the bare
		// lexical route; do not select proximal or distal deixis by default.
		return "unresolved", true
	}
	return mode, true
}

func polishAttributedSeeingGenderAvailable(source semantic.MeaningCandidate, referent meaning.Node, feature, gender, requiredNumber string, candidates []realizer.TargetCandidate) bool {
	if referent.Kind != meaning.KindReferent || referent.Concept == "" {
		return false
	}
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != referent.ID || candidate.Outcome != status.OutcomeOK ||
			candidate.TransferBinding.CoreRef.String() != string(referent.Concept) || candidate.Lexeme.POS != "noun" {
			continue
		}
		rawGender, agreementGender := "", ""
		for _, property := range candidate.Lexeme.Properties {
			switch property.Key {
			case "gender":
				rawGender = property.Value
			case "agreement_gender":
				agreementGender = property.Value
			}
		}
		number, _, supported := polishTargetNumber(source, referent.ID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if agreementGender == "" && rawGender != "" {
			agreementGender = rawGender
			if number == "plural" && rawGender != "masculine_personal" {
				agreementGender = "non_masculine_personal"
			}
		}
		candidateGender := rawGender
		if feature == "agreement_gender" {
			candidateGender = agreementGender
		}
		if supported && number == requiredNumber && candidateGender == gender {
			return true
		}
	}
	return false
}

func polishAttributedSeeingLexicalNumberAlternatives(source semantic.MeaningCandidate, referent meaning.Node, candidates []realizer.TargetCandidate) (pluralOnly, nonPlural bool) {
	if referent.Kind != meaning.KindReferent || referent.Concept == "" {
		return false, false
	}
	for _, candidate := range candidates {
		if candidate.Kind != meaning.KindReferent || candidate.NodeID != referent.ID || candidate.Outcome != status.OutcomeOK ||
			candidate.TransferBinding.CoreRef.String() != string(referent.Concept) || candidate.Lexeme.POS != "noun" {
			continue
		}
		lexicalPlural := candidate.Lexeme.ParadigmID == "pl_noun_plural_only_attested"
		for _, property := range candidate.Lexeme.Properties {
			if property.Key == "grammatical_number" && property.Value == "plural_only" {
				lexicalPlural = true
			}
		}
		number, _, supported := polishTargetNumber(source, referent.ID, candidate.Lexeme.Properties, candidate.Lexeme.ParadigmID)
		if !supported {
			continue
		}
		if lexicalPlural && number == "plural" {
			pluralOnly = true
		} else if !lexicalPlural && number == "singular" {
			nonPlural = true
		}
	}
	return pluralOnly, nonPlural
}

func polishAttributedSeeingDeterminerVariant(source semantic.MeaningCandidate, variant, base, deixis string, referent meaning.Node, slot, number string, candidates []realizer.TargetCandidate) bool {
	if deixis != "proximal" && deixis != "distal" {
		return false
	}
	feature := "gender"
	if number == "plural" {
		feature = "agreement_gender"
	}
	parts := strings.Split(strings.TrimPrefix(variant, base+"."), ".")
	terminalIndex := len(parts) - 1
	if len(parts) == 2 && parts[0] == "masculine_personal" || len(parts) == 2 && parts[0] == "non_masculine_personal" {
		terminalIndex = 1
	}
	if len(parts) != 1 && terminalIndex != 1 {
		return false
	}
	for _, gender := range []string{"masculine_animal", "masculine_inanimate", "masculine_personal", "feminine", "neuter", "non_masculine_personal"} {
		if number == "plural" && gender != "masculine_personal" && gender != "non_masculine_personal" {
			continue
		}
		if !polishAttributedSeeingGenderAvailable(source, referent, feature, gender, number, candidates) {
			continue
		}
		terminal, supported := polishAttributedSeeingDeterminer(deixis, gender, number)
		variantGenderMatches := len(parts) == 1 || parts[0] == gender
		if supported && variantGenderMatches && parts[terminalIndex] == terminal && (number != "plural" || gender == "masculine_personal" && (terminal == "ci" || terminal == "tamci") || gender == "non_masculine_personal" && (terminal == "te" || terminal == "tamte")) {
			return true
		}
	}
	return false
}

func polishAttributedSeeingDeterminer(deixis, gender, number string) (string, bool) {
	prefix := "ten"
	if deixis == "distal" {
		prefix = "tamten"
	}
	if number == "plural" {
		if gender == "masculine_personal" {
			if deixis == "distal" {
				return "tamci", true
			}
			return "ci", true
		}
		if gender == "non_masculine_personal" {
			if deixis == "distal" {
				return "tamte", true
			}
			return "te", true
		}
		return "", false
	}
	switch gender {
	case "masculine_animal", "masculine_inanimate", "masculine_personal":
		return prefix, true
	case "feminine":
		if deixis == "distal" {
			return "tamta", true
		}
		return "ta", true
	case "neuter":
		if deixis == "distal" {
			return "tamto", true
		}
		return "to", true
	default:
		return "", false
	}
}

func polishAttributedSeeingVariant(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, shape polishAttributedSeeingGraph, candidates []realizer.TargetCandidate) bool {
	variant := candidate.CompiledVariant.Variant.ID.String()
	subjectDeixis, subjectDeixisOK := polishAttributedSeeingDeixis(source, shape.subject.ID)
	objectDeixis, objectDeixisOK := polishAttributedSeeingDeixis(source, shape.object.ID)
	if !subjectDeixisOK || !objectDeixisOK {
		return false
	}
	switch candidate.Family.Family.ID.String() {
	case "construction:pl.attributed_seeing":
		if subjectDeixis != "unresolved" {
			return polishAttributedSeeingDeterminerVariant(source, variant, "construction:pl.attributed_seeing.svo", subjectDeixis, shape.subject, "subject", "singular", candidates)
		}
		return variant == "construction:pl.attributed_seeing.svo"
	case "construction:pl.attributed_seeing_object":
		if variant == "construction:pl.attributed_seeing_object.ovs" {
			if subjectDeixis != "unresolved" || objectDeixis != "unresolved" || shape.subject.Multiplicity != meaning.MultiplicityOne || shape.object.Multiplicity != meaning.MultiplicityOne {
				return false
			}
			pluralOnly, nonPlural := polishAttributedSeeingLexicalNumberAlternatives(source, shape.object, candidates)
			return !pluralOnly || nonPlural
		}
		if objectDeixis == "proximal" || objectDeixis == "distal" {
			pluralOnly, _ := polishAttributedSeeingLexicalNumberAlternatives(source, shape.object, candidates)
			if pluralOnly {
				return polishAttributedSeeingObjectDeterminerVariant(source, variant, "construction:pl.attributed_seeing_object.svo.plural_lexical", objectDeixis, shape.object, "plural", candidates)
			}
			return polishAttributedSeeingObjectDeterminerVariant(source, variant, "construction:pl.attributed_seeing_object.svo", objectDeixis, shape.object, "singular", candidates)
		}
		if variant == "construction:pl.attributed_seeing_object.svo" {
			pluralOnly, nonPlural := polishAttributedSeeingLexicalNumberAlternatives(source, shape.object, candidates)
			return !pluralOnly || nonPlural
		}
		pluralOnly, _ := polishAttributedSeeingLexicalNumberAlternatives(source, shape.object, candidates)
		return variant == "construction:pl.attributed_seeing_object.svo.plural_lexical" && pluralOnly &&
			polishAttributedSeeingGenderAvailable(source, shape.object, "agreement_gender", "non_masculine_personal", "plural", candidates)
	case "construction:pl.attributed_seeing_object_negative":
		return shape.scope.ID != "" && shape.object.Multiplicity == meaning.MultiplicityOne && shape.subject.Multiplicity == meaning.MultiplicityOne &&
			(subjectDeixis == "unresolved" && objectDeixis == "unresolved") &&
			(variant == "construction:pl.attributed_seeing_object_negative.svo" || variant == "construction:pl.attributed_seeing_object_negative.ovs")
	case "construction:pl.attributed_seeing_both":
		if objectDeixis == "proximal" || objectDeixis == "distal" {
			return false
		}
		if shape.subject.Multiplicity == meaning.MultiplicityOne && shape.object.Multiplicity == meaning.MultiplicityOne {
			pluralOnly, ordinary := polishAttributedSeeingLexicalNumberAlternatives(source, shape.object, candidates)
			if subjectDeixis == "proximal" || subjectDeixis == "distal" {
				base := "construction:pl.attributed_seeing_both.svo"
				if pluralOnly {
					base += ".object_plural_lexical"
				}
				return polishAttributedSeeingDeterminerVariant(source, variant, base, subjectDeixis, shape.subject, "subject", "singular", candidates)
			}
			switch variant {
			case "construction:pl.attributed_seeing_both.svo":
				return ordinary
			case "construction:pl.attributed_seeing_both.svo.object_plural_lexical":
				return pluralOnly && polishAttributedSeeingGenderAvailable(source, shape.object, "agreement_gender", "non_masculine_personal", "plural", candidates)
			default:
				return false
			}
		}
		prefix := "construction:pl.attributed_seeing_both.svo.plural."
		if !strings.HasPrefix(variant, prefix) || shape.subject.Multiplicity != meaning.MultiplicityMany || shape.object.Multiplicity != meaning.MultiplicityMany {
			return false
		}
		genders := strings.Split(strings.TrimPrefix(variant, prefix), ".")
		return len(genders) == 2 &&
			polishAttributedSeeingGenderAvailable(source, shape.subject, "gender", genders[0], "plural", candidates) &&
			polishAttributedSeeingGenderAvailable(source, shape.object, "gender", genders[1], "plural", candidates)
	case "construction:pl.attributed_seeing_plural":
		prefix := "construction:pl.attributed_seeing_plural.svo."
		if !strings.HasPrefix(variant, prefix) {
			return false
		}
		parts := strings.Split(strings.TrimPrefix(variant, prefix), ".")
		if subjectDeixis == "unresolved" {
			// Bare plural variants remain available only when the source carries
			// no explicit deixis anchor. These variants are split by the noun's
			// lexical gender so genuine target ambiguity is preserved.
			return len(parts) == 1 && polishAttributedSeeingGenderAvailable(source, shape.subject, "gender", parts[0], "plural", candidates)
		}
		if subjectDeixis != "proximal" && subjectDeixis != "distal" || len(parts) != 2 {
			return false
		}
		class, terminal := parts[0], parts[1]
		if class != "masculine_personal" && class != "non_masculine_personal" {
			return false
		}
		if !polishAttributedSeeingGenderAvailable(source, shape.subject, "agreement_gender", class, "plural", candidates) {
			return false
		}
		expected, supported := polishAttributedSeeingDeterminer(subjectDeixis, class, "plural")
		return supported && terminal == expected
	case "construction:pl.attributed_seeing_object_plural":
		if strings.HasPrefix(variant, "construction:pl.attributed_seeing_object_plural.ovs.") {
			genderClass := strings.TrimPrefix(variant, "construction:pl.attributed_seeing_object_plural.ovs.")
			return subjectDeixis == "unresolved" && objectDeixis == "unresolved" &&
				shape.subject.Multiplicity == meaning.MultiplicityOne && shape.object.Multiplicity == meaning.MultiplicityMany &&
				(genderClass == "masculine_personal" || genderClass == "non_masculine_personal") &&
				polishAttributedSeeingGenderAvailable(source, shape.object, "agreement_gender", genderClass, "plural", candidates)
		}
		prefix := "construction:pl.attributed_seeing_object_plural.svo."
		if !strings.HasPrefix(variant, prefix) {
			return false
		}
		parts := strings.Split(strings.TrimPrefix(variant, prefix), ".")
		genderClass := parts[0]
		if genderClass != "masculine_personal" && genderClass != "non_masculine_personal" {
			return false
		}
		if !polishAttributedSeeingGenderAvailable(source, shape.object, "agreement_gender", genderClass, "plural", candidates) {
			return false
		}
		if objectDeixis != "unresolved" {
			return polishAttributedSeeingObjectDeterminerVariant(source, variant, "construction:pl.attributed_seeing_object_plural.svo", objectDeixis, shape.object, "plural", candidates)
		}
		return len(parts) == 1
	case "construction:pl.attributed_seeing_both_plural":
		if objectDeixis == "proximal" || objectDeixis == "distal" {
			return false
		}
		prefix := "construction:pl.attributed_seeing_both_plural.svo."
		if !strings.HasPrefix(variant, prefix) {
			return false
		}
		parts := strings.Split(strings.TrimPrefix(variant, prefix), ".")
		if len(parts) != 4 && len(parts) != 5 {
			return false
		}
		subjectNumber, objectNumber := "singular", "singular"
		if parts[0] == "many" {
			subjectNumber = "plural"
		} else if parts[0] != "one" {
			return false
		}
		if parts[1] == "many" {
			objectNumber = "plural"
		} else if parts[1] != "one" {
			return false
		}
		if subjectNumber == "plural" != (shape.subject.Multiplicity == meaning.MultiplicityMany) ||
			objectNumber == "plural" != (shape.object.Multiplicity == meaning.MultiplicityMany) {
			return false
		}
		if subjectDeixis == "proximal" || subjectDeixis == "distal" {
			if len(parts) != 5 {
				return false
			}
			form, supported := polishAttributedSeeingDeterminer(subjectDeixis, parts[2], "plural")
			if !supported || parts[4] != form {
				return false
			}
		} else if len(parts) != 4 {
			return false
		}
		subjectGenderFeature, objectGenderFeature := "gender", "gender"
		if subjectNumber == "plural" {
			subjectGenderFeature = "agreement_gender"
		}
		if objectNumber == "plural" {
			objectGenderFeature = "agreement_gender"
		}
		return polishAttributedSeeingGenderAvailable(source, shape.subject, subjectGenderFeature, parts[2], subjectNumber, candidates) &&
			polishAttributedSeeingGenderAvailable(source, shape.object, objectGenderFeature, parts[3], objectNumber, candidates)
	default:
		return false
	}
}

func polishAttributedSeeingObjectDeterminerVariant(source semantic.MeaningCandidate, variant, base, deixis string, referent meaning.Node, number string, candidates []realizer.TargetCandidate) bool {
	if deixis != "proximal" && deixis != "distal" {
		return false
	}
	features := "gender"
	if number == "plural" {
		features = "agreement_gender"
	}
	for _, gender := range []string{"masculine_animal", "masculine_inanimate", "masculine_personal", "feminine", "neuter", "non_masculine_personal"} {
		if number == "plural" && gender != "masculine_personal" && gender != "non_masculine_personal" {
			continue
		}
		if !polishAttributedSeeingGenderAvailable(source, referent, features, gender, number, candidates) {
			continue
		}
		form, ok := polishAttributedSeeingObjectDeterminer(deixis, gender, number)
		if number == "plural" && base == "construction:pl.attributed_seeing_object.svo.plural_lexical" {
			if ok && gender == "non_masculine_personal" && variant == base+"."+form {
				return true
			}
		} else {
			suffix := form
			if form == "tę" {
				suffix = "te_acc"
			} else if form == "tamtą" {
				suffix = "tamta_acc"
			}
			if ok && variant == base+"."+gender+"."+suffix {
				return true
			}
		}
	}
	return false
}

func polishAttributedSeeingObjectDeterminer(deixis, gender, number string) (string, bool) {
	if number == "plural" {
		if gender == "masculine_personal" {
			if deixis == "distal" {
				return "tamtych", true
			}
			return "tych", true
		}
		if gender == "non_masculine_personal" {
			if deixis == "distal" {
				return "tamte", true
			}
			return "te", true
		}
		return "", false
	}
	switch gender {
	case "feminine":
		if deixis == "distal" {
			return "tamtą", true
		}
		return "tę", true
	case "neuter":
		if deixis == "distal" {
			return "tamto", true
		}
		return "to", true
	case "masculine_animal", "masculine_personal":
		if deixis == "distal" {
			return "tamtego", true
		}
		return "tego", true
	case "masculine_inanimate":
		if deixis == "distal" {
			return "tamten", true
		}
		return "ten", true
	default:
		return "", false
	}
}

func polishAttributedSeeingMainCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, candidates []realizer.TargetCandidate) bool {
	shape, ok := polishAttributedSeeingShape(source)
	if !ok || !polishAttributedSeeingContextsValid(source) || candidate.NodeID != shape.event.ID || candidate.Kind != meaning.KindEvent || candidate.Outcome != status.OutcomeOK {
		return false
	}
	sense, frame, core, supportedPredicate := polishAttributedPredicateBinding(source)
	if !supportedPredicate || candidate.Sense.ID.String() != sense || string(candidate.Frame.ID) != frame ||
		candidate.TransferBinding.CoreRef.String() != core || !polishAttributedSeeingFamily(candidate) || !polishAttributedSeeingVariant(source, candidate, shape, candidates) ||
		!realizer.CoversLexicalMeaning(*source.Meaning, candidate.NodeID, candidate.Family.Family) {
		return false
	}
	return polishAttributedSeeingFamilyMatches(source, candidate, shape)
}

func polishAttributedSeeingModifierCandidate(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, all []realizer.TargetCandidate) bool {
	shape, ok := polishAttributedSeeingShape(source)
	if !ok || !polishAttributedSeeingContextsValid(source) || candidate.Kind != meaning.KindState || candidate.Outcome != status.OutcomeOK {
		return false
	}
	property, exists := shape.properties[candidate.NodeID]
	if !exists {
		return false
	}
	spec, licensed := polishAttributionForConcept(string(property.Concept))
	if !licensed || candidate.Sense.ID.String() != spec.sense || candidate.Lexeme.ID.String() != spec.lexeme || candidate.TransferBinding.CoreRef.String() != spec.concept {
		return false
	}
	variant := candidate.CompiledVariant.Variant.ID.String()
	frameName := strings.TrimPrefix(spec.sense, "sense:") + ".frame"
	familyID := "construction:pl.attribution"
	if len(property.Participants) != 1 {
		return false
	}
	bearer, bearerExists := polishMeaningNode(source, property.Participants[0].Target)
	if !bearerExists || bearer.Kind != meaning.KindReferent {
		return false
	}
	if bearer.Multiplicity == meaning.MultiplicityMany {
		gender, number, known := polishIdentificationReferentFeatures(source, bearer.ID, all)
		if !known || number != "plural" {
			return false
		}
		familyID = "construction:pl.attribution_plural"
		frameName = strings.TrimPrefix(spec.sense, "sense:") + ".plural.frame"
		if gender == "masculine_personal" {
			if variant != "construction:pl.attribution_plural.ci" && variant != "construction:pl.attribution_plural.tamci" {
				return false
			}
		} else {
			if variant != "construction:pl.attribution_plural.te" && variant != "construction:pl.attribution_plural.tamte" {
				return false
			}
		}
	} else if bearer.Multiplicity != meaning.MultiplicityOne || candidate.Family.Family.ID.String() != familyID || string(candidate.Frame.ID) != frameName || !strings.HasPrefix(variant, "construction:pl.attribution.") {
		return false
	}
	if candidate.Family.Family.ID.String() != familyID || string(candidate.Frame.ID) != frameName {
		return false
	}
	for _, main := range all {
		if main.SourceCandidateID == candidate.SourceCandidateID && main.SemanticID == candidate.SemanticID && polishAttributedSeeingMainCandidate(source, main, all) {
			return true
		}
	}
	return false
}
