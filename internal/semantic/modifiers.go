package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
)

type modifierComposition struct {
	states []meaning.Node
}

func familyModifiers(family construction.ConstructionFamily) []construction.SlotSpec {
	modifiers := make([]construction.SlotSpec, 0, 1)
	for _, slot := range family.SlotSpecs {
		if slot.ModifierBinding != nil {
			modifiers = append(modifiers, slot)
		}
	}
	sort.Slice(modifiers, func(i, j int) bool { return modifiers[i].Position < modifiers[j].Position })
	return modifiers
}

func modifierCompositionFor(
	input CompositionInput,
	alternative parser.ParseAlternative,
	family construction.ConstructionFamily,
	candidates map[string]parser.PhraseCandidate,
	phraseNodes map[string]parser.NodeID,
	referentNodeByCandidate map[string]meaning.NodeID,
	meaningNodes []meaning.Node,
	attributes map[string]NodeAttributesEvidence,
) (modifierComposition, status.OutcomeKind, []EvidenceRecord, []status.Diagnostic) {
	modifierSlots := familyModifiers(family)
	if len(modifierSlots) == 0 {
		if len(alternative.NPPhraseGroups) > 0 {
			return modifierComposition{}, status.OutcomeInvalid, nil, []status.Diagnostic{semanticDiagnostic("NP_GROUP_FAMILY_MISMATCH", "parse alternative contains NP group evidence for a family with no modifier slots", "modifier.np_phrase_groups")}
		}
		return modifierComposition{}, status.OutcomeOK, nil, nil
	}
	for _, group := range alternative.NPPhraseGroups {
		found := false
		for _, slot := range modifierSlots {
			if slot.ID == group.ModifierSlotID && slot.ModifierBinding.TargetSlot == group.HeadSlotID {
				found = true
				break
			}
		}
		if !found {
			return modifierComposition{}, status.OutcomeInvalid, nil, []status.Diagnostic{semanticDiagnostic("NP_GROUP_FAMILY_MISMATCH", "NP group slots do not match one declared modifier-to-NP family binding", "modifier.np_phrase_groups")}
		}
	}
	composition := modifierComposition{states: make([]meaning.Node, 0, len(modifierSlots))}
	var records []EvidenceRecord
	var diagnostics []status.Diagnostic
	outcome := status.OutcomeOK
	for index, modifierSlot := range modifierSlots {
		state, slotOutcome, slotRecords, diagnostic := modifierCompositionForSlot(
			input, alternative, modifierSlot, meaning.NodeID("modifier"+strconv.Itoa(index)), candidates,
			phraseNodes, referentNodeByCandidate, meaningNodes, attributes,
		)
		records = append(records, slotRecords...)
		if diagnostic.Code != "" {
			diagnostic.Path = "modifier.slots." + modifierSlot.ID + "." + diagnostic.Path
			diagnostics = append(diagnostics, diagnostic)
		}
		if slotOutcome != status.OutcomeOK {
			outcome = mergeSemanticOutcome(outcome, slotOutcome)
			continue
		}
		composition.states = append(composition.states, state)
	}
	if outcome != status.OutcomeOK {
		return modifierComposition{}, outcome, records, diagnostics
	}
	return composition, status.OutcomeOK, records, diagnostics
}

func modifierCompositionForSlot(
	input CompositionInput,
	alternative parser.ParseAlternative,
	modifierSlot construction.SlotSpec,
	stateID meaning.NodeID,
	candidates map[string]parser.PhraseCandidate,
	phraseNodes map[string]parser.NodeID,
	referentNodeByCandidate map[string]meaning.NodeID,
	meaningNodes []meaning.Node,
	attributes map[string]NodeAttributesEvidence,
) (meaning.Node, status.OutcomeKind, []EvidenceRecord, status.Diagnostic) {
	modifierBinding := modifierSlot.ModifierBinding

	modifierSlotBinding, found, duplicate := alternativeSlotBinding(alternative, modifierSlot.ID)
	if duplicate {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_SLOT_AMBIGUOUS", "parse alternative binds the modifier slot more than once", "modifier.slot")
	}
	if !found || len(modifierSlotBinding.CandidateIDs) == 0 {
		return modifierCompositionFailure(status.OutcomeIncomplete, "MODIFIER_CANDIDATE_UNRESOLVED", "required modifier slot has no selected AP candidate", "modifier.slot")
	}
	if len(modifierSlotBinding.CandidateIDs) != 1 {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_CANDIDATE_AMBIGUOUS", "modifier slot has multiple AP candidates", "modifier.slot")
	}
	modifierCandidate, exists := candidates[modifierSlotBinding.CandidateIDs[0]]
	if !exists || modifierCandidate.Category != construction.CategoryAP {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_CANDIDATE_UNSUPPORTED", "modifier slot does not resolve to one AP candidate", "modifier.slot")
	}

	targetSlotBinding, found, duplicate := alternativeSlotBinding(alternative, modifierBinding.TargetSlot)
	if duplicate {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_TARGET_AMBIGUOUS", "modifier target slot is bound more than once", "modifier.target")
	}
	if !found || len(targetSlotBinding.CandidateIDs) == 0 {
		return modifierCompositionFailure(status.OutcomeIncomplete, "MODIFIER_TARGET_UNRESOLVED", "modifier target slot has no selected NP candidate", "modifier.target")
	}
	if len(targetSlotBinding.CandidateIDs) != 1 {
		return modifierCompositionFailure(status.OutcomeAmbiguous, "MODIFIER_TARGET_AMBIGUOUS", "modifier target slot has multiple NP candidates", "modifier.target")
	}
	targetCandidate, exists := candidates[targetSlotBinding.CandidateIDs[0]]
	if !exists || targetCandidate.Category != construction.CategoryNP {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_TARGET_UNSUPPORTED", "modifier target slot does not resolve to one NP candidate", "modifier.target")
	}
	groupID := ""
	var selectedGroup *parser.NPPhraseGroupBinding
	for index := range alternative.NPPhraseGroups {
		group := &alternative.NPPhraseGroups[index]
		if group.ModifierSlotID != modifierSlot.ID {
			continue
		}
		if selectedGroup != nil {
			return modifierCompositionFailure(status.OutcomeAmbiguous, "NP_GROUP_BINDING_AMBIGUOUS", "modifier slot is linked to multiple compiled NP groups", "modifier.np_phrase_group")
		}
		selectedGroup = group
	}
	if selectedGroup != nil {
		if selectedGroup.HeadSlotID != modifierBinding.TargetSlot || selectedGroup.ModifierCandidateID != modifierCandidate.ID || selectedGroup.HeadCandidateID != targetCandidate.ID {
			return modifierCompositionFailure(status.OutcomeInvalid, "NP_GROUP_BINDING_MISMATCH", "compiled NP group does not name the exact selected modifier and target candidates", "modifier.np_phrase_group")
		}
		groupID = selectedGroup.CompositeCandidateID
	}
	npPhrase, phraseOutcome, phraseDiagnostic := selectedAttributiveNPPhrase(input, targetCandidate, modifierCandidate, groupID)
	if phraseOutcome != status.OutcomeOK {
		var phraseRecords []EvidenceRecord
		if npPhrase.ID != "" {
			phraseRecords = append(phraseRecords, npPhraseEvidenceRecord(targetCandidate, modifierCandidate, npPhrase, phraseOutcome, phraseDiagnostic.Message))
		}
		return meaning.Node{}, phraseOutcome, phraseRecords, phraseDiagnostic
	}
	if selectedGroup != nil && (npPhrase.ID != selectedGroup.CompositeCandidateID || npPhrase.HeadCandidateID != selectedGroup.HeadCandidateID ||
		npPhrase.ModifierCandidateID != selectedGroup.ModifierCandidateID || npPhrase.DeterminerCandidateID != selectedGroup.DeterminerCandidateID ||
		npPhrase.TokenStart != selectedGroup.TokenStart || npPhrase.TokenEnd != selectedGroup.TokenEnd) {
		return modifierCompositionFailure(status.OutcomeInvalid, "NP_GROUP_COMPOSITE_MISMATCH", "selected composite NP differs from the parser's exact atomic group evidence", "modifier.np_phrase_group")
	}
	targetNodeID, exists := referentNodeByCandidate[targetCandidate.ID]
	if !exists {
		return modifierCompositionFailure(status.OutcomeIncomplete, "MODIFIER_TARGET_UNRESOLVED", "selected NP candidate was not materialized as a referent", "modifier.target")
	}
	targetNode, exists := meaningNodeByID(meaningNodes, targetNodeID)
	if !exists || targetNode.Kind != meaning.KindReferent {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_TARGET_UNSUPPORTED", "modifier target is not the exact materialized referent for its NP candidate", "modifier.target")
	}
	targetConceptID, err := ids.ParseConceptID(string(targetNode.Concept))
	if err != nil {
		return modifierCompositionFailure(status.OutcomeUnsupported, "MODIFIER_TARGET_UNSUPPORTED", "modifier target referent has no valid CORE concept", "modifier.target")
	}
	targetConcept, exists := input.Core.Lookup(targetConceptID)
	if !exists {
		return modifierCompositionFailure(status.OutcomeUnknown, "MODIFIER_TARGET_UNKNOWN", "modifier target concept is absent from CORE", "modifier.target")
	}

	mapping, outcome, records, reason := uniqueModifierTransfer(input.Lexicon, modifierCandidate, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID])
	if outcome != status.OutcomeOK {
		code := "MODIFIER_TRANSFER_UNAVAILABLE"
		if outcome == status.OutcomeAmbiguous {
			code = "MODIFIER_TRANSFER_AMBIGUOUS"
		} else if outcome == status.OutcomeIncomplete {
			code = "MODIFIER_TRANSFER_UNRESOLVED"
		}
		return meaning.Node{}, outcome, records, semanticDiagnostic(code, reason, "modifier.transfer")
	}
	property, exists := input.Core.Lookup(mapping.binding.CoreRef)
	if !exists {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnknown, "approved modifier binding points to a property absent from CORE")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnknown, "MODIFIER_PROPERTY_UNKNOWN", record.Reason, "modifier.transfer", []EvidenceRecord{record})
	}
	if property.Kind != core.KindProperty {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnsupported, "modifier binding does not point to a CORE property")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnsupported, "MODIFIER_PROPERTY_UNSUPPORTED", record.Reason, "modifier.transfer", []EvidenceRecord{record})
	}
	if !propertyRoleAllowsTarget(property, modifierBinding.SemanticRole, targetConcept.Kind) {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnsupported, "CORE property role does not license the selected target concept kind")
		return modifierCompositionFailureWithEvidence(status.OutcomeUnsupported, "MODIFIER_ROLE_TARGET_UNSUPPORTED", record.Reason, "modifier.target", []EvidenceRecord{record})
	}
	state, err := meaning.NewState(stateID, mapping.binding.CoreRef, []meaning.Participant{{Role: modifierBinding.SemanticRole, Target: targetNodeID}})
	if err != nil {
		record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeUnsupported, "modifier state cannot be represented by Meaning")
		_, outcome, records, diagnostic := modifierCompositionFailureWithEvidence(status.OutcomeUnsupported, "MODIFIER_STATE_UNSUPPORTED", record.Reason, "modifier", []EvidenceRecord{record})
		return meaning.Node{}, outcome, records, diagnostic
	}
	if attribute, supplied := attributes[modifierCandidate.ID]; supplied {
		state.TemporalFrame = attribute.TemporalFrame
	}
	record := modifierEvidenceRecord(modifierCandidate, mapping, modifierBinding.SemanticRole, phraseNodes[modifierCandidate.ID], status.OutcomeOK, "")
	record.NPPhraseID = npPhrase.ID
	record.NPPhrasePlanID = string(npPhrase.Plan.Plan.ID)
	return state, status.OutcomeOK, []EvidenceRecord{record}, status.Diagnostic{}
}

func selectedAttributiveNPPhrase(input CompositionInput, head, modifier parser.PhraseCandidate, requiredID string) (parser.CompositeNPPhraseCandidate, status.OutcomeKind, status.Diagnostic) {
	var selected parser.CompositeNPPhraseCandidate
	seen := map[string]parser.CompositeNPPhraseCandidate{}
	for _, candidate := range head.CompositeNPs {
		if candidate.HeadCandidateID != head.ID || candidate.ModifierCandidateID != modifier.ID {
			continue
		}
		if requiredID != "" && candidate.ID != requiredID {
			continue
		}
		if prior, duplicate := seen[candidate.ID]; duplicate {
			if !reflect.DeepEqual(prior, candidate) {
				return parser.CompositeNPPhraseCandidate{}, status.OutcomeUnsupported, semanticDiagnostic("NP_PHRASE_ID_COLLISION", "one composite NP ID refers to conflicting evidence payloads", "modifier.np_phrase.id")
			}
			continue
		}
		seen[candidate.ID] = candidate
		if selected.ID != "" {
			return parser.CompositeNPPhraseCandidate{}, status.OutcomeAmbiguous, semanticDiagnostic("NP_PHRASE_EVIDENCE_AMBIGUOUS", "selected NP and AP resolve to multiple distinct compiled phrase candidates", "modifier.np_phrase")
		}
		selected = candidate
	}
	if selected.ID == "" {
		return parser.CompositeNPPhraseCandidate{}, status.OutcomeIncomplete, semanticDiagnostic("NP_PHRASE_EVIDENCE_UNRESOLVED", "selected NP head has no composite phrase evidence for the exact selected AP candidate", "modifier.np_phrase")
	}
	if selected.Outcome != status.OutcomeOK {
		outcome := selected.Outcome
		if !outcome.Valid() || outcome == status.OutcomeOK {
			outcome = status.OutcomeIncomplete
		}
		return selected, outcome, semanticDiagnostic("NP_PHRASE_EVIDENCE_INCOMPLETE", "compiled NP phrase evidence is not complete", "modifier.np_phrase")
	}
	if diagnostic := validateAttributiveNPPhrase(input, head, modifier, selected); diagnostic.Code != "" {
		return selected, diagnosticOutcome(diagnostic), diagnostic
	}
	return selected, status.OutcomeOK, status.Diagnostic{}
}

func validateAttributiveNPPhrase(input CompositionInput, head, modifier parser.PhraseCandidate, phrase parser.CompositeNPPhraseCandidate) status.Diagnostic {
	plan := phrase.Plan.Plan
	if input.PhraseProfile.Version == "" || input.PhraseProfile.Language != input.Lattice.Language {
		return semanticDiagnostic("NP_PHRASE_PROFILE_INCOMPLETE", "source NP evidence requires the exact matching language phrase profile", "modifier.np_phrase.profile")
	}
	var template *construction.PhraseTemplate
	for index := range input.PhraseProfile.Templates {
		candidate := &input.PhraseProfile.Templates[index]
		if candidate.ID == "np-attributive" && candidate.Kind == construction.PhraseNP {
			template = candidate
			break
		}
	}
	if template == nil {
		return semanticDiagnostic("NP_PHRASE_PROFILE_INCOMPLETE", "source NP evidence requires its declared np-attributive template", "modifier.np_phrase.profile")
	}
	if phrase.Outcome != status.OutcomeOK || phrase.Plan.Completeness != construction.PhrasePlanComplete || len(phrase.Diagnostics) != 0 || plan.Kind != construction.PhraseNP || plan.TemplateID != "np-attributive" ||
		plan.Language != input.Lattice.Language || phrase.TokenStart < 0 || phrase.TokenEnd > len(input.Lattice.Tokens) ||
		phrase.TokenStart >= phrase.TokenEnd || phrase.Span.End <= phrase.Span.Start || phrase.HeadCandidateID != head.ID || phrase.ModifierCandidateID != modifier.ID {
		return semanticDiagnostic("NP_PHRASE_EVIDENCE_INVALID", "composite outcome, diagnostics, language, or exact child IDs do not match this parse selection", "modifier.np_phrase")
	}
	hasDeterminer := phrase.Determiner != nil
	if hasDeterminer != (phrase.DeterminerCandidateID != "") {
		return semanticDiagnostic("NP_DETERMINER_EVIDENCE_INVALID", "determiner pointer and exact candidate ID disagree", "modifier.np_phrase.determiner")
	}
	expectedEvidenceIDs := []string{head.ID, modifier.ID}
	if hasDeterminer {
		expectedEvidenceIDs = append(expectedEvidenceIDs, phrase.Determiner.ID)
	}
	if !equalStringSlices(phrase.EvidenceCandidateIDs, expectedEvidenceIDs) {
		return semanticDiagnostic("NP_PHRASE_EVIDENCE_IDS_INVALID", "composite evidence candidate IDs do not exactly name the selected children", "modifier.np_phrase.evidence_candidate_ids")
	}
	determinerID := ""
	if hasDeterminer {
		determinerID = phrase.Determiner.ID
	}
	expectedPlanID := "np-plan-" + digestStrings(string(input.Lattice.Language), head.ID, modifier.ID, determinerID)
	if string(plan.ID) != expectedPlanID || phrase.ID != "np-candidate-"+digestStrings(expectedPlanID) {
		return semanticDiagnostic("NP_PHRASE_ID_MISMATCH", "phrase plan and composite IDs do not identify the exact source child evidence", "modifier.np_phrase.id")
	}
	if len(plan.Children) != 0 || len(plan.Features.Values) != 0 || plan.Features.Language != "" || plan.Features.ProfileVersion != "" || plan.FrameID != "" ||
		plan.GovernedCase.Feature != "" || plan.GovernedCase.State != "" || plan.GovernedCase.Value != "" || plan.GovernedCase.Variable != nil {
		return semanticDiagnostic("NP_PHRASE_PLAN_FIELDS_INVALID", "compiled NP plan contains unexpected children, features, frame, or governed case", "modifier.np_phrase.plan")
	}
	inputs := plan.Inputs
	expectedInputCount := 2
	if hasDeterminer {
		expectedInputCount++
	}
	if len(inputs) != expectedInputCount {
		return semanticDiagnostic("NP_PHRASE_INPUTS_INVALID", "compiled NP plan must contain exactly its selected lexical inputs", "modifier.np_phrase.plan.inputs")
	}
	inputSpecs := []struct {
		id        string
		role      string
		kind      construction.PhraseInputKind
		required  bool
		candidate parser.PhraseCandidate
		features  []features.FeatureID
	}{
		{"head-input", "head", construction.PhraseInputLexical, true, head, npTemplateInputFeatures(*template, "head")},
		{"modifier-input", "modifier", construction.PhraseInputAP, false, modifier, npTemplateInputFeatures(*template, "modifier")},
	}
	if hasDeterminer {
		inputSpecs = append(inputSpecs, struct {
			id        string
			role      string
			kind      construction.PhraseInputKind
			required  bool
			candidate parser.PhraseCandidate
			features  []features.FeatureID
		}{id: "determiner-input", role: "determiner", kind: construction.PhraseInputLexical, required: false, features: npTemplateInputFeatures(*template, "determiner")})
	}
	profileVersion := ""
	for index, spec := range inputSpecs {
		actual := inputs[index]
		if actual.ID != spec.id || actual.Role != spec.role || actual.Kind != spec.kind || actual.Required != spec.required {
			return semanticDiagnostic("NP_PHRASE_INPUT_IDENTITY_INVALID", "compiled phrase input identity differs from its exact source role", "modifier.np_phrase.plan.inputs")
		}
		expected := map[features.FeatureID]features.FeatureValueID{}
		if spec.role == "determiner" {
			for feature, value := range phrase.Determiner.Features {
				expected[feature] = value
			}
		} else {
			expected = projectNPFeatures(spec.candidate.Features, spec.features)
		}
		version, ok := matchNPFeatureBundle(actual.Features, input.Lattice.Language, expected)
		if !ok {
			return semanticDiagnostic("NP_PHRASE_FEATURE_EVIDENCE_MISMATCH", "compiled input feature bundle does not equal the exact selected lexical/token features", "modifier.np_phrase.plan.inputs."+spec.role+".features")
		}
		if version != "" {
			if profileVersion != "" && profileVersion != version {
				return semanticDiagnostic("NP_PHRASE_PROFILE_MISMATCH", "compiled input bundles use different language feature profile versions", "modifier.np_phrase.plan.inputs")
			}
			profileVersion = version
		}
	}
	if profileVersion == "" {
		return semanticDiagnostic("NP_PHRASE_PROFILE_UNKNOWN", "compiled NP plan has no typed language feature profile identity", "modifier.np_phrase.plan.inputs")
	}
	presentRoles := map[string]bool{"head": true, "modifier": true, "determiner": hasDeterminer}
	expectedOrder := make([]string, 0, len(plan.Inputs))
	for _, role := range template.Order {
		if presentRoles[role] {
			expectedOrder = append(expectedOrder, role)
		}
	}
	if !equalStringSlices(plan.Order, expectedOrder) {
		return semanticDiagnostic("NP_PHRASE_ORDER_INVALID", "compiled NP input order does not match the language template order", "modifier.np_phrase.plan.order")
	}
	expectedAgreements := make([]construction.FeatureAgreement, 0, len(template.Agreements))
	roleFeatures := map[string]map[features.FeatureID]features.FeatureValueID{
		"head": head.Features, "modifier": modifier.Features,
	}
	if hasDeterminer {
		roleFeatures["determiner"] = phrase.Determiner.Features
	}
	for _, declaration := range template.Agreements {
		if !presentRoles[declaration.ControllerRole] || !presentRoles[declaration.TargetRole] {
			continue
		}
		left, leftKnown := roleFeatures[declaration.ControllerRole][declaration.Feature]
		right, rightKnown := roleFeatures[declaration.TargetRole][declaration.Feature]
		if !leftKnown || !rightKnown {
			return semanticDiagnostic("NP_AGREEMENT_UNKNOWN", "declared NP agreement feature is not known on both selected inputs", "modifier.np_phrase.agreement."+string(declaration.Feature))
		}
		if left != right {
			return semanticDiagnostic("NP_AGREEMENT_MISMATCH", "selected NP inputs conflict on a declared agreement feature", "modifier.np_phrase.agreement."+string(declaration.Feature))
		}
		controller := "head-input"
		target := "modifier-input"
		if declaration.ControllerRole == "determiner" {
			controller = "determiner-input"
		}
		if declaration.TargetRole == "determiner" {
			target = "determiner-input"
		}
		if declaration.TargetRole == "head" {
			target = "head-input"
		}
		if declaration.ControllerRole == "modifier" {
			controller = "modifier-input"
		}
		expectedAgreements = append(expectedAgreements, construction.FeatureAgreement{ControllerInput: controller, TargetInput: target, Feature: declaration.Feature})
	}
	if !equalFeatureAgreements(plan.Agreements, expectedAgreements) {
		return semanticDiagnostic("NP_PHRASE_AGREEMENTS_INVALID", "compiled agreement constraints do not match the selected language profile and child evidence", "modifier.np_phrase.plan.agreements")
	}
	expectedStart, expectedByteStart := modifier.TokenStart, modifier.Span.Start
	if hasDeterminer {
		determiner := phrase.Determiner
		if determiner.TokenIndex < 0 || determiner.TokenIndex+1 != modifier.TokenStart || determiner.TokenIndex >= len(input.Lattice.Tokens) ||
			input.Lattice.Tokens[determiner.TokenIndex].Token.Span != determiner.Span || input.Lattice.Tokens[determiner.TokenIndex].Token.Surface != determiner.Surface || determiner.Span.End > modifier.Span.Start {
			return semanticDiagnostic("NP_DETERMINER_EVIDENCE_INVALID", "determiner child does not occupy the exact local position before the selected AP", "modifier.np_phrase.determiner")
		}
		if phrase.DeterminerCandidateID != determiner.ID || determiner.ID != npTokenEvidenceID(input.Lattice.Language, determiner) {
			return semanticDiagnostic("NP_DETERMINER_ID_MISMATCH", "determiner token ID does not identify its exact token and features", "modifier.np_phrase.determiner.id")
		}
		matchStatus := determinerFeaturesLicensed(input.PhraseProfile, inputs[0].Features, inputs[len(inputs)-1].Features, determiner.Surface, determiner)
		switch matchStatus {
		case construction.DeterminerIncomplete:
			return semanticDiagnostic("NP_DETERMINER_EVIDENCE_INCOMPLETE", "determiner profile match requires unresolved lexical or feature evidence", "modifier.np_phrase.determiner.features")
		case construction.DeterminerAmbiguous:
			return semanticDiagnostic("NP_DETERMINER_EVIDENCE_AMBIGUOUS", "determiner profile match has multiple non-identical exact rules", "modifier.np_phrase.determiner.features")
		case construction.DeterminerUnsupported:
			return semanticDiagnostic("NP_DETERMINER_FEATURES_INVALID", "determiner features and compatible agreement classes do not match the observed surface", "modifier.np_phrase.determiner.features")
		}
		expectedStart, expectedByteStart = determiner.TokenIndex, determiner.Span.Start
	} else if len(input.PhraseProfile.DeterminerForms) > 0 {
		if determinerClass, state, known := headDeterminerFeature(head, inputs[0].Features); known {
			switch state {
			case features.StateAbsent:
				// The exact compiled phrase plan has no determiner input; explicit
				// absence is compatible with the source token evidence.
			case features.StateUnknown:
				return semanticDiagnostic("NP_DETERMINER_EVIDENCE_INCOMPLETE", "NP determiner state is unresolved", "modifier.np_phrase.determiner")
			case features.StateKnown:
				if determinerClass == "bare" {
					// Explicit bare evidence is valid for profiles that provide it.
				} else {
					return semanticDiagnostic("NP_DETERMINER_EVIDENCE_MISSING", "NP determiner feature requires an observed determiner token", "modifier.np_phrase.determiner")
				}
			}
		}
	}
	if phrase.TokenStart != expectedStart || phrase.TokenEnd != head.TokenEnd || phrase.Span.Start != expectedByteStart || phrase.Span.End != head.Span.End || modifier.TokenEnd != head.TokenStart {
		return semanticDiagnostic("NP_PHRASE_SPAN_MISMATCH", "compiled NP span does not exactly cover the selected determiner/AP/head sequence", "modifier.np_phrase.span")
	}
	return status.Diagnostic{}
}

func headDeterminerFeature(head parser.PhraseCandidate, bundle features.FeatureBundle) (features.FeatureValueID, features.ValueState, bool) {
	if value, ok := head.Features["determiner"]; ok {
		return value, features.StateKnown, true
	}
	for _, value := range bundle.Values {
		if value.Feature == "determiner" {
			return value.Value, value.State, true
		}
	}
	return "", "", false
}

func npTemplateInputFeatures(template construction.PhraseTemplate, role string) []features.FeatureID {
	for _, input := range template.Inputs {
		if input.Role == role {
			return append([]features.FeatureID(nil), input.FeatureIDs...)
		}
	}
	return nil
}

func projectNPFeatures(source map[features.FeatureID]features.FeatureValueID, allowed []features.FeatureID) map[features.FeatureID]features.FeatureValueID {
	result := make(map[features.FeatureID]features.FeatureValueID)
	for _, feature := range allowed {
		if value, exists := source[feature]; exists {
			result[feature] = value
		}
	}
	return result
}

func matchNPFeatureBundle(bundle features.FeatureBundle, language features.LanguageID, expected map[features.FeatureID]features.FeatureValueID) (string, bool) {
	if len(expected) == 0 {
		return "", len(bundle.Values) == 0 && bundle.Language == "" && bundle.ProfileVersion == ""
	}
	if bundle.Language != language || strings.TrimSpace(string(bundle.ProfileVersion)) == "" || len(bundle.Values) != len(expected) {
		return "", false
	}
	seen := make(map[features.FeatureID]struct{}, len(bundle.Values))
	for _, value := range bundle.Values {
		if value.State != features.StateKnown || value.Variable != nil || value.Value != expected[value.Feature] {
			return "", false
		}
		if _, duplicate := seen[value.Feature]; duplicate {
			return "", false
		}
		if _, exists := expected[value.Feature]; !exists {
			return "", false
		}
		seen[value.Feature] = struct{}{}
	}
	return string(bundle.ProfileVersion), true
}

func determinerFeaturesLicensed(profile construction.PhraseProfile, head, determiner features.FeatureBundle, surface string, evidence *parser.NPPhraseTokenEvidence) construction.DeterminerMatchStatus {
	if evidence == nil {
		return construction.DeterminerUnsupported
	}
	match := construction.EvaluateDeterminerForms(profile, surface, head, determiner, evidence.DeterminerOnset)
	if match.Status != construction.DeterminerOK {
		return match.Status
	}
	for _, rule := range match.Candidates {
		if rule.Onset != "" && rule.Onset != evidence.DeterminerOnset {
			continue
		}
		ruleValues := make(map[features.FeatureID]features.FeatureValueID, len(rule.DeterminerFeatures))
		for _, value := range rule.DeterminerFeatures {
			if value.State == features.StateKnown {
				ruleValues[value.Feature] = value.Value
			}
		}
		if equalFeatureValueMaps(ruleValues, evidence.Features) && ruleCompatibleAgreement(rule, evidence.CompatibleAgreementGenders) {
			return construction.DeterminerOK
		}
	}
	return construction.DeterminerUnsupported
}

func ruleCompatibleAgreement(rule construction.DeterminerFormRule, supplied []features.FeatureValueID) bool {
	var expected []features.FeatureValueID
	for _, value := range rule.HeadFeatures {
		if value.Feature == "agreement_gender" && value.State == features.StateKnown {
			expected = append(expected, value.Value)
		}
	}
	if len(expected) != len(supplied) {
		return false
	}
	for i := range expected {
		if expected[i] != supplied[i] {
			return false
		}
	}
	return true
}

func npTokenEvidenceID(language features.LanguageID, evidence *parser.NPPhraseTokenEvidence) string {
	featureValues := make([]string, 0, len(evidence.Features))
	for feature, value := range evidence.Features {
		featureValues = append(featureValues, string(feature)+"\x00"+string(value))
	}
	sort.Strings(featureValues)
	identity := string(language) + "\x00" + strconv.Itoa(evidence.TokenIndex) + "\x00" + evidence.Surface + "\x00" + strconv.Itoa(evidence.Span.Start) + "\x00" + strconv.Itoa(evidence.Span.End)
	if evidence.DeterminerOnset != "" {
		identity += "\x00determiner_onset=" + evidence.DeterminerOnset
	}
	for _, value := range featureValues {
		identity += "\x00" + value
	}
	for _, gender := range evidence.CompatibleAgreementGenders {
		identity += "\x00agreement_gender=" + string(gender)
	}
	digest := sha256.Sum256([]byte(identity))
	return "phrase-token:" + hex.EncodeToString(digest[:])
}

func digestStrings(values ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return hex.EncodeToString(digest[:])
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalFeatureAgreements(left, right []construction.FeatureAgreement) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalFeatureValueMaps(left, right map[features.FeatureID]features.FeatureValueID) bool {
	if len(left) != len(right) {
		return false
	}
	for feature, value := range left {
		if right[feature] != value {
			return false
		}
	}
	return true
}

func containsFeatureValueID(values []features.FeatureValueID, value features.FeatureValueID) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func diagnosticOutcome(diagnostic status.Diagnostic) status.OutcomeKind {
	if diagnostic.Severity == status.SeverityError {
		return status.OutcomeUnsupported
	}
	return status.OutcomeIncomplete
}

func phraseOrderContains(order []string, role string) bool {
	for _, item := range order {
		if item == role {
			return true
		}
	}
	return false
}

func npPhraseEvidenceRecord(head, modifier parser.PhraseCandidate, phrase parser.CompositeNPPhraseCandidate, outcome status.OutcomeKind, reason string) EvidenceRecord {
	return EvidenceRecord{SourceSpan: phrase.Span, CandidateID: modifier.ID, NPPhraseID: phrase.ID, NPPhrasePlanID: string(phrase.Plan.Plan.ID), Outcome: outcome, Reason: reason}
}

func uniqueModifierTransfer(index *lexicon.Index, candidate parser.PhraseCandidate, role meaning.Role, phraseNode parser.NodeID) (transfer, status.OutcomeKind, []EvidenceRecord, string) {
	if index == nil || len(candidate.SenseIDs) == 0 {
		return transfer{}, status.OutcomeIncomplete, nil, "AP candidate has no lexical senses to bind as a modifier"
	}
	seenSense := make(map[ids.SenseID]struct{}, len(candidate.SenseIDs))
	selected := []transfer{}
	records := []EvidenceRecord{}
	problem := status.OutcomeOK
	reason := ""
	for _, sense := range candidate.SenseIDs {
		if _, seen := seenSense[sense]; seen {
			continue
		}
		seenSense[sense] = struct{}{}
		all := index.BindingsForSense(sense)
		authorized := index.TransferBindings(sense)
		if len(all) == 0 {
			problem = mergeSemanticOutcome(problem, status.OutcomeUnsupported)
			reason = "an AP candidate sense has no semantic transfer binding"
			records = append(records, modifierEvidenceRecord(candidate, transfer{sense: sense}, role, phraseNode, status.OutcomeUnsupported, reason))
			continue
		}
		if len(authorized) != 1 || len(all) != 1 {
			if len(authorized) > 1 || len(all) > 1 && len(authorized) > 0 {
				problem = mergeSemanticOutcome(problem, status.OutcomeAmbiguous)
				reason = "AP sense has competing approved or unapproved CORE mappings"
			} else if problem == status.OutcomeOK || problem == status.OutcomeUnsupported {
				problem = mergeSemanticOutcome(problem, status.OutcomeIncomplete)
				reason = "AP sense has no approved equivalent CORE binding"
			}
			for _, binding := range all {
				outcome := status.OutcomeIncomplete
				bindingReason := "AP sense mapping is not an approved equivalent transfer"
				if binding.TransferAuthorized() {
					outcome = status.OutcomeAmbiguous
					bindingReason = "AP sense has competing modifier mappings"
				}
				records = append(records, modifierEvidenceRecord(candidate, transfer{binding: binding, sense: sense}, role, phraseNode, outcome, bindingReason))
			}
			continue
		}
		mapping := transfer{binding: authorized[0], sense: sense}
		selected = append(selected, mapping)
		records = append(records, modifierEvidenceRecord(candidate, mapping, role, phraseNode, status.OutcomeOK, ""))
	}
	if problem != status.OutcomeOK {
		if problem == status.OutcomeAmbiguous || len(selected) > 0 && problem != status.OutcomeOK {
			problem = status.OutcomeAmbiguous
		}
		for index := range records {
			if records[index].Outcome == status.OutcomeOK {
				records[index].Outcome = problem
				records[index].Reason = "modifier sense competes with an unresolved or unsupported candidate sense"
			}
		}
		return transfer{}, problem, records, reason
	}
	if len(selected) != 1 {
		if len(selected) == 0 {
			return transfer{}, status.OutcomeUnsupported, records, "AP candidate has no unique approved equivalent CORE mapping"
		}
		for index := range records {
			records[index].Outcome = status.OutcomeAmbiguous
			records[index].Reason = "AP candidate has multiple lexical senses with modifier mappings"
		}
		return transfer{}, status.OutcomeAmbiguous, records, "AP candidate has multiple lexical senses with modifier mappings"
	}
	return selected[0], status.OutcomeOK, records, ""
}

func modifierEvidenceRecord(candidate parser.PhraseCandidate, mapping transfer, role meaning.Role, phraseNode parser.NodeID, outcome status.OutcomeKind, reason string) EvidenceRecord {
	record := EvidenceRecord{
		SourceSpan: candidate.Span, CandidateID: candidate.ID, PhraseNodeID: phraseNode,
		ReadingID: firstString(candidate.ReadingIDs), SenseID: mapping.sense,
		Role: role, Outcome: outcome, Reason: reason,
		Features: evidenceFeatures(candidate.Features),
	}
	if mapping.binding.CoreRef.String() != "" {
		record.CoreConceptID = mapping.binding.CoreRef
		record.BindingKey = BindingKeyFor(mapping.sense, mapping.binding.CoreRef)
		record.Provenance = mapping.binding.Provenance
	}
	return record
}

func modifierCompositionFailure(outcome status.OutcomeKind, code, message, path string) (meaning.Node, status.OutcomeKind, []EvidenceRecord, status.Diagnostic) {
	return meaning.Node{}, outcome, nil, semanticDiagnostic(code, message, path)
}

func modifierCompositionFailureWithEvidence(outcome status.OutcomeKind, code, message, path string, records []EvidenceRecord) (meaning.Node, status.OutcomeKind, []EvidenceRecord, status.Diagnostic) {
	return meaning.Node{}, outcome, records, semanticDiagnostic(code, message, path)
}

func alternativeSlotBinding(alternative parser.ParseAlternative, slotID string) (parser.SlotBinding, bool, bool) {
	var selected parser.SlotBinding
	found := false
	for _, slot := range alternative.Slots {
		if slot.SlotID != slotID {
			continue
		}
		if found {
			return parser.SlotBinding{}, false, true
		}
		selected, found = slot, true
	}
	return selected, found, false
}

func propertyRoleAllowsTarget(property core.Concept, role meaning.Role, targetKind core.ConceptKind) bool {
	for _, propertyRole := range property.RoleTypes {
		if propertyRole.Name != string(role) {
			continue
		}
		if len(propertyRole.AllowedKinds) == 0 {
			return true
		}
		for _, allowed := range propertyRole.AllowedKinds {
			if allowed == targetKind {
				return true
			}
		}
		return false
	}
	return false
}

func meaningNodeByID(nodes []meaning.Node, id meaning.NodeID) (meaning.Node, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
	}
	return meaning.Node{}, false
}
