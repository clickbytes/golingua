package en

import (
	"context"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/text"
)

// NewSentenceSource compiles the English analysis stages against one exact
// snapshot. Its complete declarative attestation is limited to attested
// third-person singular and plural present visual, auditory, directed-looking,
// and reading predicates, third-person past reading, bounded present negation,
// explicit articles, and past receiving.
func NewSentenceSource(snapshot *pack.Pack) (*runtime.SourcePipeline, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("English sentence source requires a snapshot")
	}
	seed, err := Load()
	if err != nil {
		return nil, err
	}
	profile, inverse, _, _ := seed.RuntimeMorphology()
	valencyIndex, err := CompileValency()
	if err != nil {
		return nil, err
	}
	packages, registry, err := SentenceConstructions(valencyIndex)
	if err != nil {
		return nil, err
	}
	plans, _, err := runtime.CompileSentencePlans(profile, packages, valencyIndex)
	if err != nil {
		return nil, err
	}
	seeingSense, err := ids.NewSenseID("en.see.visual")
	if err != nil {
		return nil, err
	}
	hearingSense, err := ids.NewSenseID("en.hear.auditory")
	if err != nil {
		return nil, err
	}
	receivingSense, err := ids.NewSenseID("en.get.receive")
	if err != nil {
		return nil, err
	}
	lookingSense, err := ids.NewSenseID("en.look.directed")
	if err != nil {
		return nil, err
	}
	readingSense, err := ids.NewSenseID("en.read.text")
	if err != nil {
		return nil, err
	}
	classificationSense, err := ids.NewSenseID("en.be.class_membership")
	if err != nil {
		return nil, err
	}
	attributionSense, err := ids.NewSenseID("en.black.color")
	if err != nil {
		return nil, err
	}
	return &runtime.SourcePipeline{
		Language: "en", ClauseConnector: "and", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}},
		AllowedUnlexicalizedTerminals: map[string]bool{"at": true, "do": true, "does": true, "not": true, "this": true, "that": true, "it": true, "is": true, "a": true, "who": true, "whom": true, "and": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: englishWHGapPhraseBuilder{inner: englishIdentificationPhraseBuilder{inner: englishDeterminerNPBuilder{inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "pronoun": {construction.CategoryNP}, "verb": {construction.CategoryVP}, "adjective": {construction.CategoryAP},
			},
		}}, lexicon: snapshot.Lexicon()}}},
		EvidenceBuilder: englishSentenceEvidenceBuilder{lexicon: snapshot.Lexicon(), base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses: map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true, readingSense: true, classificationSense: true, attributionSense: true},
			PredicateFeatureAlternativesBySense: map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID{
				seeingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "first", "verb_number": "singular", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "first", "verb_number": "plural", "number": "plural"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "second", "verb_number": "singular", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "second", "verb_number": "plural", "number": "plural"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
					{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"},
				},
				hearingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "first", "verb_number": "singular", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "first", "verb_number": "plural", "number": "plural"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "second", "verb_number": "singular", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "second", "verb_number": "plural", "number": "plural"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
				},
				lookingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
				},
				readingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "first", "verb_number": "singular", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "first", "verb_number": "plural", "number": "plural"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "second", "verb_number": "singular", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "second", "verb_number": "plural", "number": "plural"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
					{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"},
				},
				// English simple past is finite but has one form for all persons
				// and numbers, so no agreement feature is required for this cell.
				receivingSense:      {{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"}},
				classificationSense: {{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "number": "singular", "verb_number": "singular"}},
				attributionSense:    {{"adjective_degree": "positive"}},
			},
			AllowedUnlexicalizedTerminals: map[string]bool{"at": true, "do": true, "does": true, "not": true, "this": true, "that": true, "it": true, "is": true, "a": true},
		}},
	}, nil
}

// englishWHGapPhraseBuilder carries a gap only when the exact WH terminal is
// present in the analyzed lattice. The construction parser must also consume
// that terminal in an approved question variant before the gap can be used.
type englishWHGapPhraseBuilder struct{ inner runtime.PhraseEvidenceBuilder }

func (builder englishWHGapPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	for _, token := range lattice.Tokens {
		slot, role := "", ""
		switch token.Token.Folded {
		case "who":
			slot, role = "subject", "experiencer"
		case "whom":
			slot, role = "object", "stimulus"
		}
		if slot != "" {
			parseContext.WHGaps = append(parseContext.WHGaps, parser.WHGapEvidence{SlotID: slot, Role: meaning.Role(role), ScopeNodeID: meaning.NodeID(fmt.Sprintf("wh_en_%s_%d_%d", slot, token.Token.Span.Start, token.Token.Span.End)), Span: token.Token.Span})
		}
	}
	return candidates, parseContext, nil
}

type englishSentenceEvidenceBuilder struct {
	base    runtime.BareDeclarativeEvidenceBuilder
	lexicon *lexicon.Index
}

func (builder englishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	evidence, err := builder.buildSyntax(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	attributes, err := englishNodeAttributes(builder.base.SupportedSenses, builder.lexicon, candidates, forest, evidence)
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	evidence.NodeAttributes = attributes
	return evidence, nil
}

func (builder englishSentenceEvidenceBuilder) buildSyntax(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	hasNot := false
	for _, token := range lattice.Tokens {
		if token.Token.Folded == "not" {
			hasNot = true
			break
		}
	}
	if !hasNot {
		evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
		if err != nil {
			return evidence, err
		}
		evidence = addEnglishPersonalReferences(evidence, candidates, forest)
		if !evidence.OperatorCoverageComplete {
			if !englishWHQuestionCoverage(lattice, forest) {
				return evidence, nil
			}
			evidence.OperatorCoverageComplete = true
		}
		return addEnglishIdentificationReferences(evidence, candidates, forest), nil
	}
	if _, uniqueNot := englishTerminalSpan(lattice, "not"); !uniqueNot {
		return runtime.CompositionEvidence{}, nil
	}
	if englishNegativeClassificationInput(lattice) {
		return builder.buildEnglishNegativeClassification(ctx, lattice, candidates, forest)
	}
	route, number, notSpan, supported := englishPresentNegationInput(lattice)
	if !supported {
		return runtime.CompositionEvidence{}, nil
	}
	base := builder.base
	base.PredicateFeatureAlternativesBySense = cloneEnglishPredicateAlternatives(builder.base.PredicateFeatureAlternativesBySense)
	for _, negativeRoute := range englishPresentNegationRoutes {
		senseID, err := ids.NewSenseID(strings.TrimPrefix(negativeRoute.sense, "sense:"))
		if err != nil {
			return runtime.CompositionEvidence{}, err
		}
		for _, verbNumber := range []string{"singular", "plural"} {
			base.PredicateFeatureAlternativesBySense[senseID] = append(
				base.PredicateFeatureAlternativesBySense[senseID],
				map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "tense": "present", "person": "third",
					"number": features.FeatureValueID(verbNumber), "verb_number": features.FeatureValueID(verbNumber),
				},
			)
		}
	}
	evidence, err := base.Build(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	negativeFamily, err := ids.NewConstructionID(strings.TrimPrefix(route.family, "construction:"))
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	operatorCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != negativeFamily || !route.variantMatchesNumber(alternative.VariantID.String(), number) || alternative.Permission.SenseID.String() != route.sense {
				return runtime.CompositionEvidence{}, nil
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					candidateRoute, candidateNumber, negated := englishNegatedPredicateContext(lattice, candidate)
					if !exists || !negated || candidateRoute.sense != route.sense || candidateNumber != number ||
						candidate.Features["verb_form"] != "base" || string(candidate.Features["number"]) != number {
						continue
					}
					evidence.Operators = append(evidence.Operators, semantic.OperatorEvidence{
						Operator: construction.Operator{
							ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
							InputFamilyID: negativeFamily, OutputFamilyID: negativeFamily,
							SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
							Capabilities:  []construction.Capability{"negation-negated"}, Provenance: englishNegationConstructionSource,
						},
						TargetCandidateID: candidateID, Span: notSpan,
					})
					operatorCount++
				}
			}
		}
	}
	evidence.OperatorCoverageComplete = operatorCount > 0
	return evidence, nil
}

func englishWHQuestionCoverage(lattice analyzer.LexicalLattice, forest parser.Forest) bool {
	seenClause := false
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		seenClause = true
		if len(node.Alternatives) == 0 {
			return false
		}
		var gapAlternative *parser.ParseAlternative
		var omissionAlternatives []parser.ParseAlternative
		for altIndex := range node.Alternatives {
			alternative := node.Alternatives[altIndex]
			if alternative.FamilyID.String() != "construction:en.question.wh" {
				return false
			}
			terminal, auxiliary, slotID, role := "", "", "", ""
			switch alternative.VariantID.String() {
			case "construction:en.question.wh.subject":
				terminal, slotID, role = "who", "subject", "experiencer"
			case "construction:en.question.wh.object.echo":
				terminal, slotID, role = "whom", "object", "stimulus"
			case "construction:en.question.wh.object.neutral.analyze":
				terminal, auxiliary, slotID, role = "whom", "does", "object", "stimulus"
			default:
				return false
			}
			terminalCount, auxiliaryCount, gapCount := 0, 0, 0
			for _, token := range lattice.Tokens {
				if token.Token.Folded == terminal {
					terminalCount++
				}
				if auxiliary != "" && token.Token.Folded == auxiliary {
					auxiliaryCount++
				}
			}
			for _, slot := range alternative.Slots {
				if slot.WHGap != nil {
					if slot.SlotID != slotID || string(slot.WHGap.Role) != role || string(slot.WHGap.SlotID) != slotID || slot.WHGap.ScopeNodeID == "" || len(slot.CandidateIDs) != 0 {
						return false
					}
					gapCount++
				}
			}
			if terminalCount != 1 || auxiliary != "" && auxiliaryCount != 1 || gapCount > 1 {
				return false
			}
			if gapCount == 1 {
				for _, requiredSlotID := range []string{"subject", "verb", "object"} {
					if requiredSlotID == slotID {
						continue
					}
					found, populated := false, false
					for _, slot := range alternative.Slots {
						if slot.SlotID == requiredSlotID {
							found = true
							populated = len(slot.CandidateIDs) > 0 && slot.WHGap == nil
							break
						}
					}
					if !found || !populated {
						return false
					}
				}
				copy := alternative
				gapAlternative = &copy
			} else {
				omissionAlternatives = append(omissionAlternatives, alternative)
			}
		}
		if gapAlternative == nil {
			return false
		}
		for _, omission := range omissionAlternatives {
			if !englishSameParseExceptGap(*gapAlternative, omission) {
				return false
			}
		}
	}
	return seenClause
}

func englishSameParseExceptGap(withGap, withoutGap parser.ParseAlternative) bool {
	if withGap.PlanID != withoutGap.PlanID || withGap.FamilyID != withoutGap.FamilyID || withGap.VariantID != withoutGap.VariantID || withGap.Permission != withoutGap.Permission || len(withGap.Slots) != len(withoutGap.Slots) || len(withGap.Unresolved) != len(withoutGap.Unresolved) {
		return false
	}
	for i := range withGap.Unresolved {
		if withGap.Unresolved[i] != withoutGap.Unresolved[i] {
			return false
		}
	}
	gapCount := 0
	for i, a := range withGap.Slots {
		b := withoutGap.Slots[i]
		if a.SlotID != b.SlotID || a.Role != b.Role || len(a.CandidateIDs) != len(b.CandidateIDs) || len(a.NodeIDs) != len(b.NodeIDs) || b.WHGap != nil {
			return false
		}
		for j := range a.CandidateIDs {
			if a.CandidateIDs[j] != b.CandidateIDs[j] {
				return false
			}
		}
		for j := range a.NodeIDs {
			if a.NodeIDs[j] != b.NodeIDs[j] {
				return false
			}
		}
		if a.WHGap != nil {
			gapCount++
		}
	}
	return gapCount == 1
}

func englishNegativeClassificationInput(lattice analyzer.LexicalLattice) bool {
	words := make([]string, len(lattice.Tokens))
	for i, token := range lattice.Tokens {
		words[i] = token.Token.Folded
	}
	if len(words) < 5 || words[0] != "this" || words[1] != "is" || words[2] != "not" || words[3] != "a" {
		return false
	}
	for _, word := range words[4:] {
		if word == "not" {
			return false
		}
	}
	return true
}

func (builder englishSentenceEvidenceBuilder) buildEnglishNegativeClassification(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	familyID, err := ids.NewConstructionID("en.identification_negative")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	notSpan, _ := englishTerminalSpan(lattice, "not")
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	count := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != familyID || alternative.VariantID.String() != "construction:en.identification_negative.this" || alternative.Permission.SenseID.String() != "sense:en.be.class_membership" {
				continue
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, ok := byID[candidateID]
					if !ok || candidate.Category != construction.CategoryVP || candidate.LexemeID.String() != "lexeme:en.be.v" || candidate.Features["verb_form"] != "finite" || candidate.Features["tense"] != "present" || candidate.Features["person"] != "third" || candidate.Features["verb_number"] != "singular" {
						continue
					}
					evidence.Operators = append(evidence.Operators, semantic.OperatorEvidence{
						Operator: construction.Operator{ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
							InputFamilyID: familyID, OutputFamilyID: familyID,
							SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
							Capabilities:  []construction.Capability{"negation-negated"}, Provenance: englishIdentificationNegationSource},
						TargetCandidateID: candidateID, Span: notSpan,
					})
					count++
				}
			}
		}
	}
	evidence.OperatorCoverageComplete = count > 0
	if evidence.OperatorCoverageComplete {
		evidence = addEnglishIdentificationReferences(evidence, candidates, forest)
	}
	return evidence, nil
}

func cloneEnglishPredicateAlternatives(source map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID) map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID {
	result := make(map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, len(source))
	for sense, alternatives := range source {
		result[sense] = append([]map[features.FeatureID]features.FeatureValueID(nil), alternatives...)
	}
	return result
}

type englishDeterminerNPBuilder struct {
	inner   runtime.SimplePhraseEvidenceBuilder
	lexicon *lexicon.Index
}

func (builder englishDeterminerNPBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	// Do not reuse candidates' backing array while scanning it: the modifier
	// lookup below still needs the original AP entries after earlier NPs have
	// been appended to filtered.
	filtered := make([]parser.PhraseCandidate, 0, len(candidates))
	negatedBaseSeen := make(map[int]bool)
	for _, candidate := range candidates {
		if candidate.Features == nil {
			candidate.Features = make(map[features.FeatureID]features.FeatureValueID)
		}
		if candidate.Category == construction.CategoryVP {
			if number := candidate.Features["verb_number"]; number != "" {
				candidate.Features["number"] = number
			}
			if englishNegationAuxiliaryCandidate(lattice, candidate) {
				continue
			}
			route, number, negated := englishNegatedPredicateContext(lattice, candidate)
			whDoSupport := englishWHDoSupportContext(lattice, candidate)
			if whDoSupport && !englishWHDoSupportAgreement(candidate, candidates) {
				continue
			}
			if negated {
				if candidate.Features["mood"] != "imperative" || negatedBaseSeen[candidate.TokenStart] {
					continue
				}
				// In this exact auxiliary context, the imperative cell supplies
				// the English base surface; do-support carries person and number.
				candidate.Features = map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "tense": "present", "person": "third",
					"number": features.FeatureValueID(number), "verb_number": features.FeatureValueID(number),
					"mood": "indicative", "voice": "active",
				}
				negatedBaseSeen[candidate.TokenStart] = true
			}
			if whDoSupport {
				// In the licensed neutral object-WH pattern, does supplies the
				// finite features and the main verb keeps its base cell.
				candidate.Features = map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "tense": "present", "person": "third",
					"number": "singular", "verb_number": "singular",
					"mood": "indicative", "voice": "active",
				}
			}
			if !supportedEnglishPredicateRoute(candidate) && (!negated || candidate.LexemeID.String() != route.lexeme) && !whDoSupport {
				continue
			}
			if !englishPersonalAgreementCandidate(candidate, candidates) {
				continue
			}
			filtered = append(filtered, candidate)
			continue
		}
		if candidate.Category != construction.CategoryNP {
			filtered = append(filtered, candidate)
			continue
		}
		lexeme, lexemeFound := builder.lexicon.ByLexemeID(candidate.LexemeID)
		if lexemeFound && lexeme.POS == "pronoun" {
			if person := candidate.Features["pronoun_person"]; person != "" {
				candidate.Features["person"] = person
			}
			filtered = append(filtered, candidate)
			continue
		}
		if lexemeFound && lexeme.POS == "noun" {
			candidate.Features["person"] = "third"
		}
		determinerToken := candidate.TokenStart - 1
		_, hasDeterminer := englishArticleAt(lattice, determinerToken)
		modifiedByBlack := false
		if !hasDeterminer {
			for _, phrase := range candidates {
				if phrase.Category != construction.CategoryAP || phrase.LexemeID.String() != "lexeme:en.black.adj" ||
					!containsEnglishSenseName(phrase.SenseIDs, "en.black.color") || phrase.TokenEnd != candidate.TokenStart || phrase.TokenStart == 0 {
					continue
				}
				articleToken := phrase.TokenStart - 1
				_, found := englishArticleAt(lattice, articleToken)
				if found {
					determinerToken, hasDeterminer, modifiedByBlack = articleToken, true, true
					break
				}
			}
		}
		if !hasDeterminer {
			if hasEnglishThisNPPrefix(lattice, candidate.TokenStart) {
				if number := knownEnglishNounNumber(candidate.Features); number != "" {
					candidate.Features["number"] = features.FeatureValueID(number)
				}
				filtered = append(filtered, candidate)
				continue
			}
			if !lexemeFound || lexeme.POS != "noun" {
				continue
			}
			number := knownEnglishNounNumber(candidate.Features)
			knownMass := hasNumberBehavior(builder.lexicon, candidate.LexemeID, "mass")
			knownCount := hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count")
			if knownMass || knownCount && number == "plural" {
				candidate.Features["determiner"] = "bare"
				if number != "" {
					candidate.Features["number"] = features.FeatureValueID(number)
				}
				filtered = append(filtered, candidate)
			}
			continue
		}
		previous := lattice.Tokens[determinerToken]
		keep := true
		for _, reading := range previous.Readings {
			if reading.HasLexeme && (reading.Lexeme.ID.String() == "lexeme:en.the.det" || reading.Lexeme.ID.String() == "lexeme:en.a.det" || reading.Lexeme.ID.String() == "lexeme:en.an.det") {
				indefinite := reading.Lexeme.ID.String() == "lexeme:en.a.det" || reading.Lexeme.ID.String() == "lexeme:en.an.det"
				if indefinite {
					if number := knownEnglishNounNumber(candidate.Features); number != "" && number != "singular" {
						keep = false
						break
					}
					if candidate.LexemeID.String() != "" && builder.lexicon != nil {
						if noun, exists := builder.lexicon.ByLexemeID(candidate.LexemeID); exists {
							article := "a"
							if reading.Lexeme.ID.String() == "lexeme:en.an.det" {
								article = "an"
							}
							expectedArticle, hasLexicalArticle := englishIndefiniteArticle(noun.Properties)
							if modifiedByBlack {
								expectedArticle, hasLexicalArticle = "a", true
							}
							if hasLexicalArticle && expectedArticle != article {
								keep = false
								break
							}
							if !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count") {
								keep = false
								break
							}
						}
					}
				}
				if candidate.Features == nil {
					candidate.Features = make(map[features.FeatureID]features.FeatureValueID)
				}
				if indefinite {
					candidate.Features["determiner"] = "indefinite"
					if candidate.Features["noun_number"] == "" {
						candidate.Features["noun_number"] = "singular"
					}
				} else {
					candidate.Features["determiner"] = "definite"
				}
				break
			}
		}
		if keep {
			if number := knownEnglishNounNumber(candidate.Features); number != "" {
				candidate.Features["number"] = features.FeatureValueID(number)
			}
			filtered = append(filtered, candidate)
		}
	}
	return filtered, parseContext, nil
}

func englishArticleAt(lattice analyzer.LexicalLattice, index int) (string, bool) {
	if index < 0 || index >= len(lattice.Tokens) {
		return "", false
	}
	for _, reading := range lattice.Tokens[index].Readings {
		if !reading.HasLexeme {
			continue
		}
		switch reading.Lexeme.ID.String() {
		case "lexeme:en.a.det":
			return "a", true
		case "lexeme:en.an.det":
			return "an", true
		case "lexeme:en.the.det":
			return "the", true
		}
	}
	return "", false
}

func containsEnglishSenseName(senses []ids.SenseID, expected string) bool {
	for _, sense := range senses {
		if sense.String() == "sense:"+expected {
			return true
		}
	}
	return false
}

func supportedEnglishPredicateRoute(candidate parser.PhraseCandidate) bool {
	if candidate.Features["verb_form"] != "finite" {
		return false
	}
	switch candidate.Features["tense"] {
	case "present":
		person := string(candidate.Features["person"])
		if person != "third" && person != "first" && person != "second" {
			return false
		}
		if candidate.Features["verb_number"] != "singular" && candidate.Features["verb_number"] != "plural" {
			return false
		}
		for _, sense := range candidate.SenseIDs {
			if sense.String() == "sense:en.see.visual" || sense.String() == "sense:en.hear.auditory" || sense.String() == "sense:en.read.text" {
				return true
			}
		}
		return person == "third"
	case "past":
		for _, sense := range candidate.SenseIDs {
			if sense.String() == "sense:en.get.receive" || sense.String() == "sense:en.read.text" || sense.String() == "sense:en.see.visual" {
				return true
			}
		}
	}
	return false
}

func englishWHDoSupportContext(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	if candidate.Category != construction.CategoryVP || candidate.LexemeID.String() != "lexeme:en.see.v" ||
		candidate.Features["verb_form"] != "base" || candidate.TokenStart < 2 || candidate.TokenStart >= len(lattice.Tokens) {
		return false
	}
	if lattice.Tokens[0].Token.Folded != "whom" || lattice.Tokens[1].Token.Folded != "does" || candidate.TokenStart <= 1 {
		return false
	}
	whomCount, doesCount := 0, 0
	for _, token := range lattice.Tokens {
		if token.Token.Folded == "whom" {
			whomCount++
		}
		if token.Token.Folded == "does" {
			doesCount++
		}
	}
	return whomCount == 1 && doesCount == 1
}

func englishWHDoSupportAgreement(vp parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	for _, np := range candidates {
		if np.Category == construction.CategoryNP && np.TokenEnd == vp.TokenStart &&
			np.Features["person"] == "third" && np.Features["number"] == "singular" {
			return true
		}
	}
	return false
}

func englishPersonalAgreementCandidate(vp parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	if vp.Features["tense"] != "present" || (vp.LexemeID.String() != "lexeme:en.see.v" && vp.LexemeID.String() != "lexeme:en.hear.v" && vp.LexemeID.String() != "lexeme:en.read.v") {
		return true
	}
	for _, np := range candidates {
		if np.Category != construction.CategoryNP || np.TokenEnd != vp.TokenStart {
			continue
		}
		expectedPerson, expectedNumber := "third", ""
		switch np.LexemeID.String() {
		case "lexeme:en.i.pron":
			expectedPerson, expectedNumber = "first", "singular"
		case "lexeme:en.we.pron":
			expectedPerson, expectedNumber = "first", "plural"
		case "lexeme:en.you.pron":
			// English you takes base/plural agreement without resolving how
			// many addressees the utterance refers to.
			expectedPerson, expectedNumber = "second", "plural"
		}
		if string(vp.Features["person"]) != expectedPerson {
			return false
		}
		return expectedNumber == "" || string(vp.Features["verb_number"]) == expectedNumber
	}
	return true
}
