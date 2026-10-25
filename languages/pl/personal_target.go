package pl

import (
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type polishParticipantContextError struct{}

func (polishParticipantContextError) Error() string {
	return "participant anchors must identify consistent supported subjects in their paired Meaning graphs"
}

func (failure polishParticipantContextError) Result() status.Result {
	return status.Result{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{
		Code: "INVALID_PARTICIPANT_CONTEXT", Path: "source.contexts", Message: failure.Error(), Severity: status.SeverityError,
	}}}
}

const polishPersonalTargetFamily = "construction:pl.transitive_personal"

type polishPersonalTarget struct {
	role    meaning.Role
	subject meaning.Node
	anchor  meaning.DeicticAnchor
}

// polishPersonalContext validates every paired request graph whenever any
// participant anchor is present. Invalid anchor evidence cannot fall through
// to the ordinary nominal realization route.
func polishPersonalContext(source semantic.MeaningCandidate) (bool, bool) {
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
	contexts := make([]meaning.RequestContext, 0, len(source.Contexts))
	for _, paired := range source.Contexts {
		same, err := meaning.StructuralEquivalence(*source.Meaning, paired.Meaning, meaning.DefaultBudget())
		if err != nil || !same || meaning.ValidateReferences(paired.Meaning, paired.Context, meaning.DefaultBudget()) != nil {
			return true, false
		}
		contexts = append(contexts, paired.Context)
	}
	for _, context := range contexts {
		for _, mention := range context.Mentions {
			if mention.Anchor == nil || mention.Anchor.Kind != "participant" {
				continue
			}
			matches := 0
			for _, node := range source.Meaning.Nodes {
				if node.Kind != meaning.KindEvent && node.Kind != meaning.KindState {
					continue
				}
				if target, ok := polishPersonalTargetForPredicate(source, node.ID); ok && target.subject.ID == mention.NodeID {
					matches++
				}
			}
			if matches != 1 {
				return true, false
			}
		}
	}
	return true, len(contexts) > 0
}

func polishPersonalTargetForPredicate(source semantic.MeaningCandidate, predicateID meaning.NodeID) (polishPersonalTarget, bool) {
	if source.Meaning == nil {
		return polishPersonalTarget{}, false
	}
	predicate := findSentencePredicate(*source.Meaning, predicateID)
	if predicate.ID == "" {
		return polishPersonalTarget{}, false
	}
	role := meaning.Role("")
	switch predicate.Concept {
	case "concept:seeing", "concept:hearing":
		role = "experiencer"
	case "concept:reading":
		role = "reader"
	default:
		return polishPersonalTarget{}, false
	}
	var subject meaning.Node
	for _, participant := range predicate.Participants {
		if participant.Role != role {
			continue
		}
		if subject.ID != "" {
			return polishPersonalTarget{}, false
		}
		var found bool
		subject, found = polishMeaningNode(source, participant.Target)
		if !found || subject.Kind != meaning.KindReferent || subject.Concept != "concept:person" {
			return polishPersonalTarget{}, false
		}
	}
	if subject.ID == "" {
		return polishPersonalTarget{}, false
	}
	contexts := make([]meaning.RequestContext, 0, len(source.Contexts))
	for _, paired := range source.Contexts {
		same, err := meaning.StructuralEquivalence(*source.Meaning, paired.Meaning, meaning.DefaultBudget())
		if err != nil || !same || meaning.ValidateReferences(paired.Meaning, paired.Context, meaning.DefaultBudget()) != nil {
			return polishPersonalTarget{}, false
		}
		contexts = append(contexts, paired.Context)
	}
	anchor, ok := realizer.ImplicitParticipantAnchor(*source.Meaning, contexts, subject.ID)
	if !ok || anchor.Kind != "participant" || anchor.Value != "speaker" && anchor.Value != "speaker_inclusive_group" && anchor.Value != "addressee" {
		return polishPersonalTarget{}, false
	}
	return polishPersonalTarget{role: role, subject: subject, anchor: anchor}, true
}

func polishPersonalVariantFeatures(variant construction.Variant) map[string]string {
	values := make(map[string]string, len(variant.Conditions))
	for _, condition := range variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == "verb" {
			values[string(condition.Feature)] = string(condition.Value)
		}
	}
	return values
}

func polishPersonalVariantMatches(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, target polishPersonalTarget) bool {
	if candidate.Family.Family.ID.String() != polishPersonalTargetFamily || candidate.CompiledVariant.Variant.DirectionPolicy == construction.DirectionAnalyze {
		return false
	}
	values := polishPersonalVariantFeatures(candidate.CompiledVariant.Variant)
	person := values["person"]
	number := values["number"]
	tense := values["tense"]
	wantPerson := "second"
	if target.anchor.Value == "speaker" || target.anchor.Value == "speaker_inclusive_group" {
		wantPerson = "first"
	}
	if person != wantPerson || number != "singular" && number != "plural" || tense != "present" && tense != "past" {
		return false
	}
	frameRole := meaning.Role("")
	for _, argument := range candidate.Frame.Arguments {
		if argument.GrammaticalFunction == "subject" {
			if frameRole != "" {
				return false
			}
			frameRole = argument.Role
		}
	}
	if frameRole != target.role {
		return false
	}
	predicate := findSentencePredicate(*source.Meaning, candidate.NodeID)
	if predicate.TemporalFrame == meaning.TemporalFramePast {
		if candidate.Sense.ID.String() != "sense:pl.czytac.interpret" || tense != "past" {
			return false
		}
	} else if tense != "present" {
		return false
	}
	if target.anchor.Value == "speaker" && number != "singular" || target.anchor.Value == "speaker_inclusive_group" && number != "plural" {
		return false
	}
	semanticNumber := polishMeaningNumber(target.subject.Multiplicity)
	if semanticNumber != "" && semanticNumber != number {
		return false
	}
	if tense == "past" {
		gender := values["gender"]
		if number == "singular" && gender != "masculine_personal" && gender != "feminine" || number == "plural" && gender != "masculine_personal" && gender != "non_masculine_personal" {
			return false
		}
		knownGender, known, evidenceValid := polishPersonalPastAgreementGender(source, candidate, target, number)
		if !evidenceValid || known && knownGender != gender {
			return false
		}
	}
	return true
}

func polishPersonalVariantAssumptions(source semantic.MeaningCandidate, candidate realizer.TargetCandidate) []realizer.FieldAssumption {
	if source.Meaning == nil {
		return nil
	}
	values := polishPersonalVariantFeatures(candidate.CompiledVariant.Variant)
	role := meaning.Role("experiencer")
	if candidate.Sense.ID.String() == "sense:pl.czytac.interpret" {
		role = "reader"
	}
	subject, ok := polishReferentForRole(source, candidate.NodeID, role)
	if !ok {
		return nil
	}
	path, ok := runtime.MeaningPath(*source.Meaning, candidate.NodeID)
	if !ok {
		return nil
	}
	var result []realizer.FieldAssumption
	if predicate, ok := polishMeaningNode(source, candidate.NodeID); ok && predicate.TemporalFrame == "" && values["tense"] == "present" {
		result = append(result, realizer.FieldAssumption{Path: path, Field: "tense", Value: "present", Reason: "Meaning carries no temporal frame; this response uses the approved present-tense route"})
	}
	if subject.Multiplicity == "" && values["number"] != "" {
		result = append(result, realizer.FieldAssumption{Path: path, Field: "number", Value: values["number"], Reason: "English you does not encode addressee number; the Polish personal variant selects this inflection"})
	}
	if values["tense"] == "past" && values["gender"] != "" {
		_, known, valid := polishPersonalPastAgreementGender(source, candidate, polishPersonalTarget{role: role, subject: subject}, values["number"])
		if valid && !known {
			gender := values["gender"]
			if gender == "masculine_personal" && values["number"] == "singular" {
				gender = "masculine"
			}
			result = append(result, realizer.FieldAssumption{Path: path, Field: "gender", Value: gender, Reason: "Meaning evidence and participant context do not specify the grammatical agreement class required by Polish past reading inflection"})
		}
	}
	return result
}

func polishPersonalPastAgreementGender(source semantic.MeaningCandidate, candidate realizer.TargetCandidate, target polishPersonalTarget, number string) (string, bool, bool) {
	personConcept, err := ids.NewConceptID("person")
	if err != nil {
		return "", false, false
	}
	knownGender := ""
	for _, evidence := range []struct {
		role    meaning.Role
		concept ids.ConceptID
	}{{target.role, personConcept}, {"", candidate.TransferBinding.CoreRef}} {
		values, found := runtime.SourceFeatures(source, evidence.role, evidence.concept)
		if !found {
			if polishSourceFeaturesPresent(source, evidence.role, evidence.concept) {
				return "", false, false
			}
			continue
		}
		gender := string(values["gender"])
		if gender == "" {
			continue
		}
		agreementGender, ok := polishPersonalPastAgreementClass(gender, number)
		if !ok || knownGender != "" && knownGender != agreementGender {
			return "", false, false
		}
		knownGender = agreementGender
	}
	return knownGender, knownGender != "", true
}

func polishPersonalPastAgreementClass(sourceGender, number string) (string, bool) {
	if number == "singular" {
		switch sourceGender {
		case "masculine", "masculine_personal", "masculine_animal", "masculine_inanimate":
			return "masculine_personal", true
		case "feminine":
			return "feminine", true
		default:
			return "", false
		}
	}
	if number == "plural" {
		switch sourceGender {
		case "masculine_personal":
			return "masculine_personal", true
		case "feminine", "neuter", "masculine", "masculine_animal", "masculine_inanimate", "non_masculine_personal":
			return "non_masculine_personal", true
		default:
			return "", false
		}
	}
	return "", false
}
