package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

const MaxSimplePhraseCandidates = 4096

// POSPhraseConfig contains explicit language-local mappings from morphology
// labels to parser categories. The builder has no built-in POS assumptions.
type POSPhraseConfig struct {
	Categories  map[lexicon.PartOfSpeech][]construction.SlotCategory
	CaseFeature features.FeatureID
	CaseValues  map[features.FeatureValueID]string
}

// AttributiveNPInput contains the exact lexical candidates and optional
// function-word token used to prepare an inspectable composite NP. It does not
// merge their senses or submit the composite to clause parsing.
type AttributiveNPInput struct {
	PhraseProfile  construction.PhraseProfile
	FeatureProfile *features.LanguageFeatureProfile
	Head           parser.PhraseCandidate
	Modifier       parser.PhraseCandidate
	Determiner     *parser.NPPhraseTokenEvidence
}

type NPDeterminerEvidenceBuilder func(analyzer.LexicalLattice, parser.PhraseCandidate, parser.PhraseCandidate) *parser.NPPhraseTokenEvidence

// AddAttributiveNPEvidence appends compiled local NP compositions to their
// exact lexical head candidates. The original candidates and IDs are retained.
func AddAttributiveNPEvidence(lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, phraseProfile construction.PhraseProfile, featureProfile *features.LanguageFeatureProfile, determiner NPDeterminerEvidenceBuilder) ([]parser.PhraseCandidate, error) {
	if featureProfile == nil || phraseProfile.Language != lattice.Language || featureProfile.Language() != lattice.Language {
		return nil, runtimeError(status.OutcomeUnsupported, "NP_PHRASE_PROFILE_LANGUAGE_MISMATCH", "phrase.np.profile", "NP evidence requires matching lattice, phrase, and feature profile languages")
	}
	result := append([]parser.PhraseCandidate(nil), candidates...)
	created := 0
	for headIndex := range result {
		head := result[headIndex]
		if head.Category != construction.CategoryNP {
			continue
		}
		for _, modifier := range result {
			if modifier.Category != construction.CategoryAP || modifier.TokenEnd != head.TokenStart {
				continue
			}
			var determinerEvidence *parser.NPPhraseTokenEvidence
			if determiner != nil {
				determinerEvidence = determiner(lattice, head, modifier)
			}
			composite, compatible, err := CompileAttributiveNP(AttributiveNPInput{
				PhraseProfile: phraseProfile, FeatureProfile: featureProfile, Head: head, Modifier: modifier,
				Determiner: determinerEvidence,
			})
			if err != nil {
				return nil, err
			}
			if !compatible {
				continue
			}
			created++
			if created > MaxSimplePhraseCandidates {
				return nil, runtimeError(status.OutcomeResourceLimit, "NP_COMPOSITE_CANDIDATE_LIMIT", "phrase.np", "attributive NP candidate limit was exhausted")
			}
			result[headIndex].CompositeNPs = append(result[headIndex].CompositeNPs, composite)
		}
	}
	return result, nil
}

// NewNPPhraseTokenEvidence creates stable typed evidence for a function-word
// determiner that is represented by a lattice token rather than a lexical NP.
func NewNPPhraseTokenEvidence(language features.LanguageID, tokenIndex int, surface string, span text.Span, values map[features.FeatureID]features.FeatureValueID, compatibleAgreementGenders ...features.FeatureValueID) *parser.NPPhraseTokenEvidence {
	return NewNPPhraseTokenEvidenceWithOnset(language, tokenIndex, surface, span, values, "", compatibleAgreementGenders...)
}

func NewNPPhraseTokenEvidenceWithOnset(language features.LanguageID, tokenIndex int, surface string, span text.Span, values map[features.FeatureID]features.FeatureValueID, onset string, compatibleAgreementGenders ...features.FeatureValueID) *parser.NPPhraseTokenEvidence {
	featureValues := make([]string, 0, len(values))
	for feature, value := range values {
		featureValues = append(featureValues, string(feature)+"\x00"+string(value))
	}
	// Include canonical feature order so candidate identity is independent of
	// Go map iteration order.
	sort.Strings(featureValues)
	compatibleGenders := append([]features.FeatureValueID(nil), compatibleAgreementGenders...)
	sort.Slice(compatibleGenders, func(i, j int) bool { return compatibleGenders[i] < compatibleGenders[j] })
	identity := fmt.Sprintf("%s\x00%d\x00%s\x00%d\x00%d", language, tokenIndex, surface, span.Start, span.End)
	if onset != "" {
		identity += "\x00determiner_onset=" + onset
	}
	for _, value := range featureValues {
		identity += "\x00" + value
	}
	for _, gender := range compatibleGenders {
		identity += "\x00agreement_gender=" + string(gender)
	}
	digest := sha256.Sum256([]byte(identity))
	featuresCopy := make(map[features.FeatureID]features.FeatureValueID, len(values))
	for feature, value := range values {
		featuresCopy[feature] = value
	}
	return &parser.NPPhraseTokenEvidence{ID: "phrase-token:" + hex.EncodeToString(digest[:]), TokenIndex: tokenIndex, Surface: surface, Span: span, Features: featuresCopy, DeterminerOnset: onset, CompatibleAgreementGenders: compatibleGenders}
}

// CompileAttributiveNP compiles one exact head+AP composition with optional
// determiner evidence, retaining unresolved feature agreements as incomplete.
func CompileAttributiveNP(input AttributiveNPInput) (parser.CompositeNPPhraseCandidate, bool, error) {
	if input.FeatureProfile == nil || input.PhraseProfile.Language != input.FeatureProfile.Language() ||
		input.Head.Category != construction.CategoryNP || input.Modifier.Category != construction.CategoryAP ||
		input.Modifier.TokenEnd != input.Head.TokenStart || input.Head.TokenStart >= input.Head.TokenEnd {
		return parser.CompositeNPPhraseCandidate{}, false, nil
	}
	if input.Head.ID == "" || input.Modifier.ID == "" {
		return parser.CompositeNPPhraseCandidate{}, false, runtimeError(status.OutcomeInvalid, "INVALID_NP_CHILD_EVIDENCE", "phrase.np", "composite NP requires exact head and modifier candidate IDs")
	}
	templateID := "np-attributive"
	var template *construction.PhraseTemplate
	for index := range input.PhraseProfile.Templates {
		if input.PhraseProfile.Templates[index].ID == templateID && input.PhraseProfile.Templates[index].Kind == construction.PhraseNP {
			template = &input.PhraseProfile.Templates[index]
			break
		}
	}
	if template == nil {
		return parser.CompositeNPPhraseCandidate{}, false, runtimeError(status.OutcomeUnknown, "ATTRIBUTIVE_NP_TEMPLATE_UNAVAILABLE", "phrase.np.template", "language phrase profile has no np-attributive template")
	}
	headInputFeatures, err := phraseInputFeatureBundle(input.FeatureProfile, template, "head", input.Head.Features)
	if err != nil {
		return parser.CompositeNPPhraseCandidate{}, false, err
	}
	modifierInputFeatures, err := phraseInputFeatureBundle(input.FeatureProfile, template, "modifier", input.Modifier.Features)
	if err != nil {
		return parser.CompositeNPPhraseCandidate{}, false, err
	}
	inputs := []construction.PhraseInput{
		{ID: "head-input", Role: "head", Kind: construction.PhraseInputLexical, Required: true, Features: headInputFeatures},
		{ID: "modifier-input", Role: "modifier", Kind: construction.PhraseInputAP, Required: false, Features: modifierInputFeatures},
	}
	childIDs := []string{input.Head.ID, input.Modifier.ID}
	determinerAgreementUnknown := false
	determinerRuleAmbiguous := false
	var determinerValues map[features.FeatureID]features.FeatureValueID
	determinerEvidence := cloneNPPhraseToken(input.Determiner)
	if input.Determiner != nil {
		if input.Determiner.ID == "" || input.Determiner.TokenIndex < 0 || input.Determiner.Surface == "" {
			return parser.CompositeNPPhraseCandidate{}, false, runtimeError(status.OutcomeInvalid, "INVALID_NP_DETERMINER_EVIDENCE", "phrase.np.determiner", "determiner token evidence is incomplete")
		}
		determinerValues = cloneNPPhraseFeatureValues(input.Determiner.Features)
		resolved, matchStatus, resolveErr := resolveSourceDeterminerFeatures(input.PhraseProfile, input.FeatureProfile, template, input.Head, input.Determiner)
		if resolveErr != nil {
			return parser.CompositeNPPhraseCandidate{}, false, resolveErr
		}
		switch matchStatus {
		case construction.DeterminerUnsupported:
			return parser.CompositeNPPhraseCandidate{}, false, nil
		case construction.DeterminerIncomplete:
			determinerAgreementUnknown = true
		case construction.DeterminerAmbiguous:
			determinerRuleAmbiguous = true
		case construction.DeterminerOK:
			for feature, value := range resolved {
				determinerValues[feature] = value
			}
			determinerEvidence = NewNPPhraseTokenEvidenceWithOnset(input.FeatureProfile.Language(), input.Determiner.TokenIndex, input.Determiner.Surface, input.Determiner.Span, determinerValues, input.Determiner.DeterminerOnset, input.Determiner.CompatibleAgreementGenders...)
		}
		for _, declaration := range template.Agreements {
			if declaration.ControllerRole != "head" || declaration.TargetRole != "determiner" {
				continue
			}
			feature := declaration.Feature
			headValue, headKnown := input.Head.Features[feature]
			targetValue, targetKnown := determinerValues[feature]
			if !headKnown || !targetKnown {
				determinerAgreementUnknown = true
			} else if headValue != targetValue {
				return parser.CompositeNPPhraseCandidate{}, false, nil
			}
		}
		determinerFeatures, featureErr := phraseInputFeatureBundle(input.FeatureProfile, template, "determiner", determinerValues)
		if featureErr != nil {
			return parser.CompositeNPPhraseCandidate{}, false, featureErr
		}
		inputs = append(inputs, construction.PhraseInput{ID: "determiner-input", Role: "determiner", Kind: construction.PhraseInputLexical, Required: false, Features: determinerFeatures})
		childIDs = append(childIDs, determinerEvidence.ID)
	}
	agreements := make([]construction.FeatureAgreement, 0, len(template.Agreements))
	outcome := status.OutcomeOK
	diagnostics := make([]status.Diagnostic, 0, 1)
	if determinerRuleAmbiguous {
		outcome = status.OutcomeAmbiguous
		diagnostics = append(diagnostics, status.Diagnostic{Code: "NP_DETERMINER_RULE_AMBIGUOUS", Message: "multiple non-identical profile determiner rules match the exact source evidence", Path: "phrase.np.determiner", Severity: status.SeverityWarning})
	}
	if determinerAgreementUnknown {
		if outcome != status.OutcomeAmbiguous {
			outcome = status.OutcomeIncomplete
		}
		diagnostics = append(diagnostics, status.Diagnostic{Code: "NP_DETERMINER_AGREEMENT_UNKNOWN", Message: "determiner agreement with the selected NP head is incomplete", Path: "phrase.np.determiner.agreement", Severity: status.SeverityWarning})
	}
	presentRoles := map[string]bool{"head": true, "modifier": true}
	if input.Determiner != nil {
		presentRoles["determiner"] = true
	}
	for _, declaration := range template.Agreements {
		if !presentRoles[declaration.ControllerRole] || !presentRoles[declaration.TargetRole] {
			continue
		}
		feature := declaration.Feature
		if declaration.ControllerRole != "head" || declaration.TargetRole != "modifier" {
			continue
		}
		headValue, headKnown := input.Head.Features[feature]
		modifierValue, modifierKnown := input.Modifier.Features[feature]
		if headKnown && modifierKnown && headValue != modifierValue {
			return parser.CompositeNPPhraseCandidate{}, false, nil
		}
		if !headKnown || !modifierKnown {
			outcome = status.OutcomeIncomplete
			diagnostics = append(diagnostics, status.Diagnostic{Code: "NP_AGREEMENT_UNKNOWN", Message: "head and adjective agreement evidence is incomplete", Path: "phrase.np.agreements." + string(feature), Severity: status.SeverityWarning})
		}
		agreements = append(agreements, construction.FeatureAgreement{ControllerInput: "head-input", TargetInput: "modifier-input", Feature: feature})
	}
	if input.Determiner != nil {
		for _, declaration := range template.Agreements {
			if declaration.ControllerRole == "head" && declaration.TargetRole == "determiner" {
				agreements = append(agreements, construction.FeatureAgreement{ControllerInput: "head-input", TargetInput: "determiner-input", Feature: declaration.Feature})
			}
		}
	}
	order := make([]string, 0, len(inputs))
	present := make(map[string]bool, len(inputs))
	for _, item := range inputs {
		present[item.Role] = true
	}
	for _, role := range template.Order {
		if present[role] {
			order = append(order, role)
		}
	}
	if len(order) != len(inputs) {
		return parser.CompositeNPPhraseCandidate{}, false, runtimeError(status.OutcomeInvalid, "INCOMPLETE_NP_TEMPLATE_ORDER", "phrase.np.order", "NP template order does not cover the exact input roles")
	}
	planID := compositeNPPlanID(input.FeatureProfile.Language(), input.Head.ID, input.Modifier.ID, tokenEvidenceID(determinerEvidence))
	plan := construction.PhrasePlan{
		ID: construction.PhraseID(planID), TemplateID: templateID, Language: input.FeatureProfile.Language(), Kind: construction.PhraseNP,
		Inputs: inputs, Order: order, Agreements: agreements,
	}
	compiled, err := construction.CompileNPPartial(plan, input.PhraseProfile, input.FeatureProfile)
	if err != nil {
		return parser.CompositeNPPhraseCandidate{}, false, err
	}
	if compiled.Completeness == construction.PhrasePlanPartial {
		if outcome == status.OutcomeOK {
			outcome = status.OutcomeIncomplete
		}
		diagnostics = append(diagnostics, status.Diagnostic{Code: "NP_PLAN_AGREEMENT_INCOMPLETE", Message: "NP structure is verified but one or more agreement constraints remain unresolved", Path: "phrase.np.agreements", Severity: status.SeverityWarning})
	}
	start, end := input.Modifier.TokenStart, input.Head.TokenEnd
	span := text.Span{Start: input.Modifier.Span.Start, End: input.Head.Span.End}
	var determinerID string
	if input.Determiner != nil {
		determinerID = determinerEvidence.ID
		start = input.Determiner.TokenIndex
		span.Start = input.Determiner.Span.Start
	}
	return parser.CompositeNPPhraseCandidate{
		ID: compositeNPEvidenceID(planID), TokenStart: start, TokenEnd: end, Span: span,
		HeadCandidateID: input.Head.ID, ModifierCandidateID: input.Modifier.ID, DeterminerCandidateID: determinerID,
		Determiner: determinerEvidence, EvidenceCandidateIDs: childIDs, Plan: compiled, Outcome: outcome, Diagnostics: diagnostics,
	}, true, nil
}

// resolveSourceDeterminerFeatures selects determiner features only from an
// exact surface rule whose head and observed determiner evidence match. The
// rule may supply otherwise unobserved features such as syncretic case.
func resolveSourceDeterminerFeatures(phraseProfile construction.PhraseProfile, profile *features.LanguageFeatureProfile, template *construction.PhraseTemplate, head parser.PhraseCandidate, evidence *parser.NPPhraseTokenEvidence) (map[features.FeatureID]features.FeatureValueID, construction.DeterminerMatchStatus, error) {
	headBundle, err := phraseInputFeatureBundle(profile, template, "head", head.Features)
	if err != nil {
		return nil, construction.DeterminerUnsupported, err
	}
	var selected []map[features.FeatureID]features.FeatureValueID
	incomplete := false
	for _, rule := range phraseProfile.DeterminerForms {
		if !strings.EqualFold(rule.Surface, evidence.Surface) {
			continue
		}
		ruleValues := make(map[features.FeatureID]features.FeatureValueID, len(rule.DeterminerFeatures))
		for _, value := range rule.DeterminerFeatures {
			if value.State == features.StateKnown {
				ruleValues[value.Feature] = value.Value
			}
		}
		compatible := true
		for feature, observed := range evidence.Features {
			if licensed, known := ruleValues[feature]; !known || licensed != observed {
				compatible = false
				break
			}
		}
		if !compatible {
			continue
		}
		if gender, known := ruleValues["agreement_gender"]; known && len(evidence.CompatibleAgreementGenders) > 0 && !containsAgreementGender(evidence.CompatibleAgreementGenders, gender) {
			continue
		}
		ruleBundle, bundleErr := features.NewBundle(profile, rule.DeterminerFeatures)
		if bundleErr != nil {
			return nil, construction.DeterminerUnsupported, runtimeError(status.OutcomeInvalid, "NP_DETERMINER_RULE_FEATURE_INVALID", "phrase.np.determiner", bundleErr.Error())
		}
		match := construction.EvaluateDeterminerForms(phraseProfile, evidence.Surface, headBundle, ruleBundle, evidence.DeterminerOnset)
		switch match.Status {
		case construction.DeterminerOK:
			for _, matched := range match.Candidates {
				values := make(map[features.FeatureID]features.FeatureValueID, len(matched.DeterminerFeatures))
				for _, value := range matched.DeterminerFeatures {
					if value.State == features.StateKnown {
						values[value.Feature] = value.Value
					}
				}
				selected = append(selected, values)
			}
		case construction.DeterminerIncomplete, construction.DeterminerAmbiguous:
			incomplete = true
		}
	}
	unique := make(map[string]map[features.FeatureID]features.FeatureValueID, len(selected))
	for _, values := range selected {
		keys := make([]string, 0, len(values))
		for feature, value := range values {
			keys = append(keys, string(feature)+"\x00"+string(value))
		}
		sort.Strings(keys)
		unique[strings.Join(keys, "\x01")] = values
	}
	if len(unique) == 1 {
		for _, values := range unique {
			return values, construction.DeterminerOK, nil
		}
	}
	if len(unique) > 1 {
		return nil, construction.DeterminerAmbiguous, nil
	}
	if incomplete {
		return nil, construction.DeterminerIncomplete, nil
	}
	return nil, construction.DeterminerUnsupported, nil
}

func phraseInputFeatureBundle(profile *features.LanguageFeatureProfile, template *construction.PhraseTemplate, role string, source map[features.FeatureID]features.FeatureValueID) (features.FeatureBundle, error) {
	if len(source) == 0 {
		return features.FeatureBundle{}, nil
	}
	var allowed []features.FeatureID
	for _, input := range template.Inputs {
		if input.Role == role {
			allowed = input.FeatureIDs
			break
		}
	}
	values := make([]features.FeatureValue, 0, len(source))
	for feature, value := range source {
		if !containsFeatureID(allowed, feature) {
			continue
		}
		values = append(values, features.FeatureValue{Feature: feature, State: features.StateKnown, Value: value})
	}
	if len(values) == 0 {
		return features.FeatureBundle{}, nil
	}
	bundle, err := features.NewBundle(profile, values)
	if err != nil {
		return features.FeatureBundle{}, runtimeError(status.OutcomeInvalid, "INVALID_NP_INPUT_FEATURES", "phrase.np.inputs."+role, err.Error())
	}
	return bundle, nil
}

func containsFeatureID(values []features.FeatureID, candidate features.FeatureID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func compositeNPPlanID(language features.LanguageID, childIDs ...string) string {
	values := append([]string{string(language)}, childIDs...)
	return "np-plan-" + compositeDigest(values...)
}

func compositeNPEvidenceID(planID string) string { return "np-candidate-" + compositeDigest(planID) }

func tokenEvidenceID(evidence *parser.NPPhraseTokenEvidence) string {
	if evidence == nil {
		return ""
	}
	return evidence.ID
}

func compositeDigest(values ...string) string {
	identity := strings.Join(values, "\x00")
	digest := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(digest[:])
}

func cloneNPPhraseToken(evidence *parser.NPPhraseTokenEvidence) *parser.NPPhraseTokenEvidence {
	if evidence == nil {
		return nil
	}
	result := *evidence
	result.Features = make(map[features.FeatureID]features.FeatureValueID, len(evidence.Features))
	for feature, value := range evidence.Features {
		result.Features[feature] = value
	}
	result.CompatibleAgreementGenders = append([]features.FeatureValueID(nil), evidence.CompatibleAgreementGenders...)
	return &result
}

func cloneNPPhraseFeatureValues(source map[features.FeatureID]features.FeatureValueID) map[features.FeatureID]features.FeatureValueID {
	result := make(map[features.FeatureID]features.FeatureValueID, len(source))
	for feature, value := range source {
		result[feature] = value
	}
	return result
}

func containsAgreementGender(values []features.FeatureValueID, wanted features.FeatureValueID) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// SimplePhraseEvidenceBuilder creates one-token phrase candidates from every
// licensed morphology reading and sense. It preserves ambiguity and only
// copies known feature values. Multiword expressions require a dedicated
// language-specific PhraseEvidenceBuilder.
type SimplePhraseEvidenceBuilder struct {
	Config POSPhraseConfig
}

func (builder SimplePhraseEvidenceBuilder) Build(_ context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	if len(lattice.MWEs) != 0 {
		return nil, parser.ParseContext{}, runtimeError(status.OutcomeUnsupported, "MWE_PHRASE_BUILDER_REQUIRED", "lattice.mwes", "multiword expression edges require a language-specific phrase evidence builder")
	}
	candidates := make([]parser.PhraseCandidate, 0)
	for _, token := range lattice.Tokens {
		for _, reading := range token.Readings {
			if !reading.HasLexeme {
				continue
			}
			categories := builder.Config.Categories[reading.Lexeme.POS]
			if len(categories) == 0 {
				continue
			}
			features := knownReadingFeatures(reading)
			senseIDs := make([]ids.SenseID, 0, len(reading.Senses))
			for _, sense := range reading.Senses {
				senseIDs = append(senseIDs, sense.Sense.ID)
			}
			caseValue, caseKnown := knownCase(reading, builder.Config.CaseFeature, builder.Config.CaseValues)
			if _, exists := features["case"]; exists && !caseKnown {
				// Parser case evidence must have a matching dedicated label.
				// The original morphology reading still retains its exact case.
				delete(features, "case")
			}
			seenCategories := make(map[construction.SlotCategory]struct{}, len(categories))
			for _, category := range categories {
				if !validPhraseCategory(category) {
					return nil, parser.ParseContext{}, runtimeError(status.OutcomeInvalid, "INVALID_POS_PHRASE_CATEGORY", "phrase_categories."+string(reading.Lexeme.POS), "POS category mapping contains an unsupported phrase category")
				}
				if _, duplicate := seenCategories[category]; duplicate {
					return nil, parser.ParseContext{}, runtimeError(status.OutcomeInvalid, "DUPLICATE_POS_PHRASE_CATEGORY", "phrase_categories."+string(reading.Lexeme.POS), "POS category mapping contains a duplicate category")
				}
				seenCategories[category] = struct{}{}
				if len(candidates) >= MaxSimplePhraseCandidates {
					return nil, parser.ParseContext{}, runtimeError(status.OutcomeResourceLimit, "PHRASE_CANDIDATE_LIMIT", "phrase_candidates", "simple phrase candidate limit was exhausted")
				}
				headTokenIndex := 0
				if category.IsPredicate() || category == construction.CategoryNP {
					headTokenIndex = token.Index
				}
				candidates = append(candidates, parser.PhraseCandidate{
					ID:       simplePhraseID(token.Index, reading.Reading.ID, category),
					Category: category, TokenStart: token.Index, TokenEnd: token.Index + 1,
					Span: token.Token.Span, HeadTokenIndex: headTokenIndex,
					ReadingIDs: []string{reading.Reading.ID}, LexemeID: reading.Lexeme.ID,
					SenseIDs: append([]ids.SenseID(nil), senseIDs...), Features: clonePhraseFeatures(features), Case: caseValue, CaseKnown: caseKnown,
				})
			}
		}
	}
	return candidates, parser.ParseContext{}, nil
}

func clonePhraseFeatures(source map[features.FeatureID]features.FeatureValueID) map[features.FeatureID]features.FeatureValueID {
	result := make(map[features.FeatureID]features.FeatureValueID, len(source))
	for feature, value := range source {
		result[feature] = value
	}
	return result
}

func knownReadingFeatures(candidate analyzer.ReadingCandidate) map[features.FeatureID]features.FeatureValueID {
	values := make(map[features.FeatureID]features.FeatureValueID)
	for _, value := range candidate.Reading.Features.Values {
		if value.State == features.StateKnown {
			values[value.Feature] = value.Value
		}
	}
	return values
}

func knownCase(candidate analyzer.ReadingCandidate, feature features.FeatureID, labels map[features.FeatureValueID]string) (string, bool) {
	if feature == "" {
		return "", false
	}
	for _, value := range candidate.Reading.Features.Values {
		if value.Feature == feature && value.State == features.StateKnown {
			label, exists := labels[value.Value]
			return label, exists && label != ""
		}
	}
	return "", false
}

func validPhraseCategory(category construction.SlotCategory) bool {
	return category.Valid()
}

func simplePhraseID(tokenIndex int, readingID string, category construction.SlotCategory) string {
	identity := fmt.Sprintf("%d\x00%s\x00%s", tokenIndex, readingID, category)
	digest := sha256.Sum256([]byte(identity))
	return "phrase:" + hex.EncodeToString(digest[:])
}
