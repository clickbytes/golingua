package engine

import (
	"reflect"
	"strings"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/synthesis"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
)

// validateDSERelationSourceEvidence rechecks the event relation shape and its
// typed source evidence against the independently checked candidate and the
// selected source language package. It does not establish linguistic quality
// or truth.
func validateDSERelationSourceEvidence(candidate semantic.DSECandidate, document algebra.Document, plan algebraRelationEventPathPlan, sourcePackage *language.SemanticLanguagePackage, selectedMetadata algebra.Document, sourceText string) (algebraRelationBinding, string, string) {
	fail := func(code, message string) (algebraRelationBinding, string, string) {
		return algebraRelationBinding{}, code, message
	}
	evidence := candidate.RelationSourceEvidence
	if evidence == nil || evidence.Metadata == nil || sourcePackage == nil || sourcePackage.LexiconIndex() == nil || sourcePackage.ValencyIndex() == nil {
		return fail("DSE_RELATION_SOURCE_EVIDENCE_REQUIRED", "event relation candidate requires typed source evidence and its selected source package")
	}
	if !validAlgebraEventRelationPlan(plan) || evidence.VerificationScope != "algebra.Check verifies generated function type, binder scope, declarations, and request context only; it does not establish world truth or linguistic quality" || evidence.VariantID == "" || evidence.Variant.ID != evidence.VariantID || evidence.VariantID != candidate.Source.VariantID || len(evidence.Variant.Parts) == 0 {
		return fail("DSE_RELATION_SOURCE_SHAPE_MISMATCH", "source relation variant or independently matched checked event function is incomplete")
	}
	if evidence.RelationPlanProvenance.Validate() != nil || evidence.SourcePlanProvenance.Validate() != nil {
		return fail("DSE_RELATION_SOURCE_PROVENANCE_INVALID", "source relation and source plan provenance must be valid")
	}
	if !sameScopedDeclarations(evidence.SourceContext.References, document.References) || !sameRelationRequestContext(evidence.SourceContext.RequestContext, document.RequestContext) {
		return fail("DSE_RELATION_SOURCE_CONTEXT_MISMATCH", "retained source references or request context differ from the checked canonical document")
	}
	if document.RequestContext == nil || evidence.SourceContext.ContextID == "" || evidence.SourceContext.ContextID != document.RequestContext.ID || len(document.References) != 0 || len(document.Parameters) != 0 || !document.Closed {
		return fail("DSE_RELATION_SOURCE_CONTEXT_UNSUPPORTED", "relation function target route requires the exact explicitly closed context without references or parameters")
	}
	metadata := evidence.Metadata
	if metadata.Mapping.SenseID != metadata.SenseEvidence.ID || metadata.Mapping.FrameID != metadata.FrameEvidence.ID || metadata.SenseEvidence.ID == "" || metadata.SenseEvidence.LexemeID != metadata.LexemeEvidence.ID || metadata.BindingEvidence.SenseID != metadata.SenseEvidence.ID || metadata.BindingEvidence.Target == nil || string(metadata.BindingEvidence.Target.SymbolRef) != metadata.Signature.ID || metadata.Signature.ID != plan.Symbol || metadata.Mapping.HeadPortID != metadata.HeadPort.Role || metadata.FrameEvidence.SenseID != metadata.SenseEvidence.ID || metadata.SourceProvenance.Validate() != nil || metadata.BindingProvenance.Validate() != nil || metadata.FrameProvenance.Validate() != nil || metadata.BindingProvenance != metadata.Binding.Provenance || metadata.FrameProvenance != metadata.Frame.Provenance || metadata.FrameEvidence.Provenance != metadata.FrameProvenance || metadata.SourceProvenance != evidence.SourcePlanProvenance {
		return fail("DSE_RELATION_SOURCE_METADATA_MISMATCH", "source mapping, lexical binding, signature, head port, sense, and frame identities do not agree")
	}
	allowedVariant := false
	for _, id := range metadata.Mapping.AllowedVariantIDs {
		if id == evidence.VariantID {
			allowedVariant = true
			break
		}
	}
	if !allowedVariant {
		return fail("DSE_RELATION_SOURCE_VARIANT_NOT_AUTHORIZED", "source metadata does not explicitly authorize this exact construction variant")
	}
	familySet := make(map[string]bool, len(metadata.FrameEvidence.AllowedFamilyIDs))
	for _, id := range metadata.FrameEvidence.AllowedFamilyIDs {
		familySet[id] = true
	}
	seenFamilies := map[string]bool{}
	for _, id := range metadata.Mapping.AllowedFamilyIDs {
		parsed, err := ids.ParseConstructionID(id)
		if err != nil || parsed.String() != id || !familySet[id] || seenFamilies[id] {
			return fail("DSE_RELATION_SOURCE_FAMILY_NOT_AUTHORIZED", "source mapping family IDs must be unique canonical families authorized by the selected frame")
		}
		seenFamilies[id] = true
	}
	senseID, senseErr := ids.ParseSenseID(metadata.SenseEvidence.ID)
	lexemeID, lexemeErr := ids.ParseLexemeID(metadata.LexemeEvidence.ID)
	if senseErr != nil || lexemeErr != nil {
		return fail("DSE_RELATION_SOURCE_LEXICAL_IDENTITY_MISMATCH", "stable sense and lexeme evidence IDs are invalid")
	}
	selectedSense, senseOK := sourcePackage.LexiconIndex().BySenseID(senseID)
	selectedLexeme, lexemeOK := sourcePackage.LexiconIndex().ByLexemeID(lexemeID)
	if !senseOK || !lexemeOK || selectedSense.ID != senseID || selectedSense.LexemeID.String() != metadata.SenseEvidence.LexemeID || selectedSense.LexemeID != lexemeID || selectedLexeme.ID != lexemeID || selectedLexeme.Language != lexicon.LanguageID(metadata.LexemeEvidence.Language) || selectedLexeme.Lemma != metadata.LexemeEvidence.Lemma || string(selectedLexeme.POS) != metadata.LexemeEvidence.POS || selectedLexeme.Language != sourcePackage.Language() || selectedLexeme.POS != "verb" || !sameRelationSenseEvidence(metadata.SenseEvidence, selectedSense) || !sameRelationLexemeEvidence(metadata.LexemeEvidence, selectedLexeme) {
		return fail("DSE_RELATION_SOURCE_LEXICAL_IDENTITY_MISMATCH", "source sense and lexeme identity is not selected by the configured package")
	}
	frameOK := false
	for _, frame := range sourcePackage.ValencyIndex().FramesBySenseID(senseID) {
		if string(frame.ID) == metadata.FrameEvidence.ID && relationFrameEvidenceMatches(metadata.FrameEvidence, frame) {
			frameOK = true
			break
		}
	}
	if !frameOK {
		return fail("DSE_RELATION_SOURCE_FRAME_MISMATCH", "source frame is not the exact selected frame for the source sense")
	}
	selected, found := selectedRelationSignature(document, plan.Symbol)
	if !found || !sameRelationSignature(selected, plan.Signature) || !sameRelationSignature(metadata.Signature, selected) {
		return fail("DSE_RELATION_SOURCE_SIGNATURE_MISMATCH", "source metadata signature does not match the checked and selected signature")
	}
	headMatches := 0
	for _, port := range selected.Ports {
		if port.Role == metadata.Mapping.HeadPortID && port.Type.Kind == algebra.TypeEvent && port.Type.SortID == "" && port.Type.From == nil && port.Type.To == nil && port == metadata.HeadPort {
			headMatches++
		}
	}
	if headMatches != 1 {
		return fail("DSE_RELATION_SOURCE_HEAD_PORT_MISMATCH", "source mapping head must identify exactly the selected Event port")
	}
	selection := selectAlgebraEventRelationBindingForFrame(plan, sourcePackage, selectedMetadata, semantic.DSEBudget{MaxWork: maxAlgebraRealizationBindings}, valency.FrameID(metadata.Mapping.FrameID))
	if selection.Status != status.OutcomeOK {
		return fail("DSE_RELATION_SOURCE_BINDING_NOT_SELECTED", "source symbol binding is not uniquely approved and authorized in the selected metadata")
	}
	selectedBinding := selection.Binding
	if !sameRelationBindingEvidence(metadata.BindingEvidence, selectedBinding.SemanticBinding) || !sameRelationSenseEvidence(metadata.SenseEvidence, selectedBinding.LexicalSense) || !sameRelationLexemeEvidence(metadata.LexemeEvidence, selectedBinding.Lexeme) || !relationFrameEvidenceMatches(metadata.FrameEvidence, selectedBinding.Frame) || !sameRelationTargetValidation(metadata.Validation, selectedBinding.Validation) || metadata.SubjectRole != selectedBinding.SubjectRole || metadata.ObjectRole != selectedBinding.ObjectRole || !relationSourceEvidenceRolesMatch(metadata, selectedBinding) {
		return fail("DSE_RELATION_SOURCE_BINDING_EVIDENCE_MISMATCH", "retained binding, sense, frame, validation, or role-port evidence differs from the unique selected binding")
	}
	if code, message := verifyRelationSourceTokens(sourceText, evidence, plan, selectedBinding); code != "" {
		return fail(code, message)
	}
	verbParts := 0
	verbReadings := make([]string, 0, 2)
	stepVerbs := map[string]int{}
	for _, part := range evidence.SourceParts {
		if part.Slot == "verb" {
			verbParts++
			stepVerbs[part.Step]++
			if part.Reading == nil || part.Sense == nil || part.Lexeme == nil || part.SenseID != metadata.SenseEvidence.ID || part.SenseLexemeID != metadata.SenseEvidence.LexemeID || part.LexemeID != metadata.LexemeEvidence.ID || part.ReadingLexemeID != metadata.LexemeEvidence.ID || part.Reading.LexemeID != metadata.LexemeEvidence.ID || part.Sense.ID.String() != metadata.SenseEvidence.ID || part.Sense.LexemeID.String() != metadata.SenseEvidence.LexemeID || part.Lexeme.ID.String() != metadata.LexemeEvidence.ID || string(part.Lexeme.Language) != metadata.LexemeEvidence.Language || part.Lexeme.Lemma != metadata.LexemeEvidence.Lemma || string(part.Lexeme.POS) != metadata.LexemeEvidence.POS {
				return fail("DSE_RELATION_SOURCE_VERB_EVIDENCE_MISMATCH", "both verb readings must retain the exact selected sense and lexeme")
			}
			verbReadings = append(verbReadings, part.Reading.ID)
		}
	}
	if verbParts != 2 || stepVerbs["first_atom"] != 1 || stepVerbs["second_atom"] != 1 || evidence.TemporalFrames != [2]string{plan.FirstEventTemporalFrame, plan.SecondEventTemporalFrame} || candidate.Source.SenseID != metadata.SenseEvidence.ID || candidate.Source.FrameID != metadata.FrameEvidence.ID || !reflect.DeepEqual(candidate.Source.ReadingIDs, verbReadings) {
		return fail("DSE_RELATION_SOURCE_EVENT_EVIDENCE_MISMATCH", "two repeated verb readings and both exact checked temporal frames are required")
	}
	if metadata.SubjectRole == "" || metadata.ObjectRole == "" || metadata.SubjectRole == metadata.ObjectRole || len(metadata.PortRoles) != 2 || len(metadata.RolePorts) != 2 {
		return fail("DSE_RELATION_SOURCE_ROLE_EVIDENCE_INVALID", "exactly two distinct Entity role ports and subject/object roles are required")
	}
	for _, event := range []struct{ binder, path, frame string }{{plan.FirstEventBinder, plan.FirstEventPath, plan.FirstEventTemporalFrame}, {plan.SecondEventBinder, plan.SecondEventPath, plan.SecondEventTemporalFrame}} {
		if event.binder == "" || event.path == "" || event.frame != "present" {
			return fail("DSE_RELATION_SOURCE_EVENT_TIME_UNSUPPORTED", "each event binder and path must have explicit supported temporal evidence")
		}
	}
	return selectedBinding, "", ""
}

// sameRelationRequestContext treats nil and empty collections as the same
// representation while comparing every scalar and every supplied row.
func sameRelationRequestContext(left, right *algebra.RequestContext) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	normalize := func(value *algebra.RequestContext) algebra.RequestContext {
		out := *value
		out.Referents = append([]algebra.ContextReferent{}, value.Referents...)
		out.Mentions = append([]algebra.ContextMention{}, value.Mentions...)
		out.ResolutionEvidence = append([]algebra.ResolutionEvidence{}, value.ResolutionEvidence...)
		out.Conditions = append([]algebra.ContextCondition{}, value.Conditions...)
		for index := range out.Conditions {
			out.Conditions[index].TargetReferentIDs = append([]string{}, value.Conditions[index].TargetReferentIDs...)
		}
		return out
	}
	return reflect.DeepEqual(normalize(left), normalize(right))
}

func verifyRelationSourceTokens(source string, evidence *semantic.DSERelationSourceEvidence, plan algebraRelationEventPathPlan, selected algebraRelationBinding) (string, string) {
	fail := func(code, message string) (string, string) { return code, message }
	options := text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}}
	sourceTokens, err := text.Tokenize(source, options)
	if err != nil {
		return fail("DSE_RELATION_SOURCE_TOKENIZATION_FAILED", "original source text could not be tokenized under the bounded relation-source profile")
	}
	type expectedToken struct {
		planIndex int
		part      semantic.DSERelationPlanPartEvidence
		literal   text.Token
		verb      bool
	}
	var expected []expectedToken
	for index, part := range evidence.Variant.Parts {
		if part.Slot == "verb" {
			expected = append(expected, expectedToken{planIndex: index, part: part, verb: true})
			continue
		}
		if part.Literal == "" {
			return fail("DSE_RELATION_SOURCE_PLAN_LITERAL_MISSING", "every non-verb source-plan part must retain its explicit literal")
		}
		literalTokens, tokenErr := text.Tokenize(part.Literal, options)
		if tokenErr != nil || len(literalTokens) == 0 {
			return fail("DSE_RELATION_SOURCE_PLAN_LITERAL_INVALID", "source-plan literal did not produce bounded tokens")
		}
		for _, token := range literalTokens {
			expected = append(expected, expectedToken{planIndex: index, part: part, literal: token})
		}
	}
	if len(expected) == 0 {
		return fail("DSE_RELATION_SOURCE_PLAN_EMPTY", "compiled relation variant has no token coverage")
	}
	evidenceHasFinalPeriod := len(evidence.SourceParts) > 0 && evidence.SourceParts[len(evidence.SourceParts)-1].Surface == "."
	if !evidenceHasFinalPeriod && len(expected) > 0 && expected[len(expected)-1].literal.Normalized == "." && !expected[len(expected)-1].verb {
		expected = expected[:len(expected)-1]
	}
	strippedFinalPeriod := !evidenceHasFinalPeriod && strings.HasSuffix(source, ".") && len(source) > 1 && !strings.ContainsAny(source[:len(source)-1], ".?!")
	if strippedFinalPeriod && len(sourceTokens) > 0 && sourceTokens[len(sourceTokens)-1].Surface == "." {
		sourceTokens = sourceTokens[:len(sourceTokens)-1]
	}
	if len(expected) != len(evidence.SourceParts) || len(sourceTokens) != len(evidence.SourceParts) {
		return fail("DSE_RELATION_SOURCE_PARTS_INCOMPLETE", "source evidence must cover every expanded plan token and all input tokens, allowing only one final period")
	}
	roleByPort := selected.PortRoles
	argumentCounts := map[string]int{}
	verbCounts := map[string]int{}
	for index, item := range expected {
		part := evidence.SourceParts[index]
		if part.PartIndex != item.planIndex || part.Step != item.part.Step || part.Slot != item.part.Slot || part.Literal != item.part.Literal || part.Port != item.part.Port || part.PortPosition != item.part.Port || part.GrammaticalFunction != item.part.GrammaticalFunction || part.TokenIndex != index || part.Span != sourceTokens[index].Span || part.Surface != sourceTokens[index].Surface {
			return fail("DSE_RELATION_SOURCE_PART_INVALID", "source part index, span, surface, or plan origin differs from the original token stream")
		}
		if item.verb {
			if part.Reading == nil || part.Lexeme == nil || part.Sense == nil || part.Reading.Surface == "" || part.Reading.Surface != part.Surface {
				return fail("DSE_RELATION_SOURCE_VERB_EVIDENCE_MISMATCH", "each verb token must retain its actual matching reading, sense, and lexeme")
			}
			wantTense := plan.FirstEventTemporalFrame
			if item.part.Step == "second_atom" {
				wantTense = plan.SecondEventTemporalFrame
			}
			if !relationSourceReadingHasTense(*part.Reading, string(selected.Lexeme.Language), wantTense) {
				return fail("DSE_RELATION_SOURCE_VERB_TENSE_MISMATCH", "each verb reading must independently establish the temporal frame of its corresponding checked event")
			}
			verbCounts[item.part.Step]++
			continue
		}
		if part.Reading != nil || part.Sense != nil || part.Lexeme != nil {
			return fail("DSE_RELATION_SOURCE_LITERAL_HAS_READING", "construction literals and variable labels cannot claim lexical readings")
		}
		if tokenFold(item.literal) != tokenFold(sourceTokens[index]) {
			return fail("DSE_RELATION_SOURCE_LITERAL_MISMATCH", "token surface does not match the normalized, case-folded construction literal")
		}
		if item.part.Slot == "x" || item.part.Slot == "y" || item.part.Slot == "z" {
			argumentCounts[item.part.Step+":"+item.part.Slot]++
			wantBinder := ""
			switch item.part.Step + ":" + item.part.Port {
			case "first_atom:first":
				wantBinder = plan.FirstBinder
			case "first_atom:second":
				wantBinder = plan.WitnessBinder
			case "second_atom:first":
				wantBinder = plan.WitnessBinder
			case "second_atom:second":
				wantBinder = plan.SecondBinder
			default:
				return fail("DSE_RELATION_SOURCE_ARGUMENT_SLOT_INVALID", "argument slot has no matching checked atom position")
			}
			slotBinder := map[string]string{"x": plan.FirstBinder, "y": plan.SecondBinder, "z": plan.WitnessBinder}[item.part.Slot]
			if slotBinder != wantBinder {
				return fail("DSE_RELATION_SOURCE_ARGUMENT_BINDER_MISMATCH", "argument slot name does not correspond to the checked binder at its step and port")
			}
			portID := plan.FirstPorts[0]
			if item.part.Port == "second" {
				portID = plan.FirstPorts[1]
			}
			role := roleByPort[portID]
			wantFunction := ""
			if role == selected.SubjectRole {
				wantFunction = "subject"
			} else if role == selected.ObjectRole {
				wantFunction = "object"
			}
			if wantFunction == "" || item.part.GrammaticalFunction != wantFunction {
				return fail("DSE_RELATION_SOURCE_GRAMMATICAL_ROLE_MISMATCH", "variant grammatical function does not agree with the selected frame role for this port")
			}
		}
	}
	if verbCounts["first_atom"] != 1 || verbCounts["second_atom"] != 1 {
		return fail("DSE_RELATION_SOURCE_VERB_COVERAGE_INVALID", "exactly one independently read verb token is required for each atom step")
	}
	for _, required := range []string{"first_atom:x", "first_atom:z", "second_atom:z", "second_atom:y"} {
		if argumentCounts[required] != 1 {
			return fail("DSE_RELATION_SOURCE_ARGUMENT_COVERAGE_INVALID", "each checked Entity argument must be represented exactly once by its plan slot")
		}
	}
	return "", ""
}

func tokenFold(token text.Token) string {
	if token.Folded != "" {
		return token.Folded
	}
	return strings.ToLower(token.Normalized)
}

func relationSourceReadingHasTense(reading morph.Reading, languageID string, expected string) bool {
	if reading.ID == "" || reading.POS != "verb" || reading.LexemeID == "" || string(reading.Features.Language) != languageID || reading.Features.ProfileVersion == "" {
		return false
	}
	seen := map[string]features.FeatureValue{}
	for _, value := range reading.Features.Values {
		if value.Feature != "verb_form" && value.Feature != "tense" {
			continue
		}
		key := string(value.Feature)
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = value
	}
	form, hasForm := seen["verb_form"]
	tense, hasTense := seen["tense"]
	return hasForm && hasTense && form.State == features.StateKnown && form.Value == "finite" && form.Variable == nil && tense.State == features.StateKnown && string(tense.Value) == expected && tense.Variable == nil
}

func relationSourceEvidenceRolesMatch(metadata *semantic.DSERelationMetadataEvidence, selected algebraRelationBinding) bool {
	if len(metadata.PortRoles) != len(selected.PortRoles) || len(metadata.RolePorts) != len(selected.PortRoles) {
		return false
	}
	roles := make(map[string]string, len(metadata.PortRoles))
	for _, row := range metadata.PortRoles {
		if row.PortID == "" || row.Role == "" || roles[row.PortID] != "" {
			return false
		}
		roles[row.PortID] = row.Role
	}
	for port, role := range selected.PortRoles {
		if roles[port] != role {
			return false
		}
	}
	for _, row := range metadata.RolePorts {
		if row.Type.Kind != algebra.TypeEntity || selected.PortRoles[row.PortID] != row.Role {
			return false
		}
	}
	return true
}

func sameRelationSenseEvidence(evidence semantic.DSERelationSenseEvidence, selected lexicon.LexicalSense) bool {
	return sameRelationSenseEvidenceValues(evidence, relationSenseEvidenceFromSelected(selected))
}

func sameRelationLexemeEvidence(evidence semantic.DSERelationLexemeEvidence, selected lexicon.MorphLexeme) bool {
	return sameRelationLexemeEvidenceValues(evidence, relationLexemeEvidenceFromSelected(selected))
}

func sameRelationSenseEvidenceValues(left, right semantic.DSERelationSenseEvidence) bool {
	return left.ID == right.ID && left.LexemeID == right.LexemeID && left.Register == right.Register && reflect.DeepEqual(append([]string{}, left.Usage...), append([]string{}, right.Usage...)) && reflect.DeepEqual(append([]string{}, left.ValencyIDs...), append([]string{}, right.ValencyIDs...))
}

func sameRelationLexemeEvidenceValues(left, right semantic.DSERelationLexemeEvidence) bool {
	return left.ID == right.ID && left.Language == right.Language && left.Lemma == right.Lemma && left.POS == right.POS && left.ParadigmID == right.ParadigmID && reflect.DeepEqual(append([]lexicon.Property{}, left.Properties...), append([]lexicon.Property{}, right.Properties...))
}

func relationSenseEvidenceFromSelected(selected lexicon.LexicalSense) semantic.DSERelationSenseEvidence {
	out := semantic.DSERelationSenseEvidence{ID: selected.ID.String(), LexemeID: selected.LexemeID.String(), Register: selected.Register, Usage: append([]string{}, selected.Usage...), ValencyIDs: make([]string, len(selected.ValencyIDs))}
	for index, id := range selected.ValencyIDs {
		out.ValencyIDs[index] = string(id)
	}
	return out
}

func relationLexemeEvidenceFromSelected(selected lexicon.MorphLexeme) semantic.DSERelationLexemeEvidence {
	return semantic.DSERelationLexemeEvidence{ID: selected.ID.String(), Language: string(selected.Language), Lemma: selected.Lemma, POS: string(selected.POS), Properties: append([]lexicon.Property{}, selected.Properties...), ParadigmID: selected.ParadigmID}
}

func sameRelationBindingEvidence(evidence semantic.DSERelationBindingEvidence, selected lexicon.SemanticBinding) bool {
	var target *lexicon.SemanticTarget
	if selected.Target != nil {
		copy := *selected.Target
		copy.RolePorts = append([]lexicon.SemanticRolePort{}, selected.Target.RolePorts...)
		target = &copy
	}
	evidenceTarget := evidence.Target
	if evidenceTarget != nil {
		copy := *evidenceTarget
		copy.RolePorts = append([]lexicon.SemanticRolePort{}, evidenceTarget.RolePorts...)
		evidenceTarget = &copy
	}
	return evidence.SenseID == selected.SenseID.String() && evidence.CoreRef == selected.CoreRef.String() && evidence.Mapping == string(selected.Mapping) && evidence.Status == string(selected.Status) && reflect.DeepEqual(evidenceTarget, target) && evidence.Provenance == selected.Provenance
}

func sameRelationTargetValidation(left, right language.TargetValidation) bool {
	left.Unresolved = append([]synthesis.Unresolved{}, left.Unresolved...)
	right.Unresolved = append([]synthesis.Unresolved{}, right.Unresolved...)
	return reflect.DeepEqual(left, right)
}

func relationFrameEvidenceMatches(evidence semantic.DSERelationFrameEvidence, selected valency.ValencyFrame) bool {
	expected := relationFrameEvidenceFrom(selected)
	normalize := func(frame semantic.DSERelationFrameEvidence) semantic.DSERelationFrameEvidence {
		frame.SemanticRoles = append([]string{}, frame.SemanticRoles...)
		frame.Arguments = append([]semantic.DSERelationFrameArgumentEvidence{}, frame.Arguments...)
		frame.SelectionalConstraints = append([]semantic.DSERelationFrameConstraintEvidence{}, frame.SelectionalConstraints...)
		frame.AllowedFamilyIDs = append([]string{}, frame.AllowedFamilyIDs...)
		for index := range frame.SelectionalConstraints {
			frame.SelectionalConstraints[index].AllowedConceptIDs = append([]string{}, frame.SelectionalConstraints[index].AllowedConceptIDs...)
			frame.SelectionalConstraints[index].AllowedKinds = append([]string{}, frame.SelectionalConstraints[index].AllowedKinds...)
		}
		return frame
	}
	return reflect.DeepEqual(normalize(evidence), normalize(expected))
}

func relationFrameEvidenceFrom(frame valency.ValencyFrame) semantic.DSERelationFrameEvidence {
	out := semantic.DSERelationFrameEvidence{
		ID: string(frame.ID), SenseID: frame.SenseID.String(), SemanticRoles: make([]string, len(frame.SemanticRoles)),
		Arguments:              make([]semantic.DSERelationFrameArgumentEvidence, len(frame.Arguments)),
		SelectionalConstraints: make([]semantic.DSERelationFrameConstraintEvidence, len(frame.SelectionalConstraints)),
		AllowedFamilyIDs:       make([]string, len(frame.AllowedFamilyIDs)), Provenance: frame.Provenance,
	}
	for index, role := range frame.SemanticRoles {
		out.SemanticRoles[index] = string(role)
	}
	for index, argument := range frame.Arguments {
		out.Arguments[index] = semantic.DSERelationFrameArgumentEvidence{ID: argument.ID, Role: string(argument.Role), Required: argument.Required, GrammaticalFunction: argument.GrammaticalFunction, Case: argument.Case, Adposition: argument.Adposition}
	}
	for index, constraint := range frame.SelectionalConstraints {
		row := semantic.DSERelationFrameConstraintEvidence{ID: constraint.ID, Role: string(constraint.Role), Strength: string(constraint.Strength), AllowedConceptIDs: make([]string, len(constraint.AllowedConceptIDs)), AllowedKinds: make([]string, len(constraint.AllowedKinds))}
		for conceptIndex, id := range constraint.AllowedConceptIDs {
			row.AllowedConceptIDs[conceptIndex] = id.String()
		}
		for kindIndex, kind := range constraint.AllowedKinds {
			row.AllowedKinds[kindIndex] = string(kind)
		}
		out.SelectionalConstraints[index] = row
	}
	for index, id := range frame.AllowedFamilyIDs {
		out.AllowedFamilyIDs[index] = id.String()
	}
	return out
}
