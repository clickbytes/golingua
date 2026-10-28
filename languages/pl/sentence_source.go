package pl

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
	"github.com/clickbytes/golingua/internal/valency"
)

// NewSentenceSource builds the trusted Polish source stages against one exact
// CORE/lexicon snapshot. The declarative completion attestation covers
// singular and plural present visual/auditory/looking, present negation for
// those routes, third-person past receiving predicates, and reading in the
// third-person singular/plural present and past.
func NewSentenceSource(snapshot *pack.Pack) (*runtime.SourcePipeline, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("Polish sentence source requires a snapshot")
	}
	seed, err := Load()
	if err != nil {
		return nil, err
	}
	return NewSentenceSourceWithSeed(snapshot, seed)
}

// NewSentenceSourceWithSeed builds source analysis using the caller's exact
// compiled morphology seed.
func NewSentenceSourceWithSeed(snapshot *pack.Pack, seed *Seed) (*runtime.SourcePipeline, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("Polish sentence source requires a snapshot")
	}
	if seed == nil {
		return nil, fmt.Errorf("Polish sentence source requires a Polish morphology seed")
	}
	profile, inverse, _, _ := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "pl" || inverse == nil {
		return nil, fmt.Errorf("Polish sentence source requires a Polish morphology seed")
	}
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
	scopedPlan, err := SentenceScopedClauses()
	if err != nil {
		return nil, err
	}
	relationPathPlan, err := SentenceRelationPathPlan()
	if err != nil {
		return nil, err
	}
	seeingSense, err := ids.NewSenseID("pl.widziec.visual")
	if err != nil {
		return nil, err
	}
	hearingSense, err := ids.NewSenseID("pl.slyszec.auditory")
	if err != nil {
		return nil, err
	}
	receivingSense, err := ids.NewSenseID("pl.dostac.receive")
	if err != nil {
		return nil, err
	}
	lookingSense, err := ids.NewSenseID("pl.patrzec.directed")
	if err != nil {
		return nil, err
	}
	readingSense, err := ids.NewSenseID("pl.czytac.interpret")
	if err != nil {
		return nil, err
	}
	classificationSense, err := ids.NewSenseID("pl.byc.class_membership")
	if err != nil {
		return nil, err
	}
	attributionSense, err := ids.NewSenseID("pl.czarny.black")
	if err != nil {
		return nil, err
	}
	eatingSense, err := ids.NewSenseID("pl.jesc.eat")
	if err != nil {
		return nil, err
	}
	drinkingSense, err := ids.NewSenseID("pl.pic.drink")
	if err != nil {
		return nil, err
	}
	caseValues := map[features.FeatureValueID]string{}
	for _, value := range []string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"} {
		caseValues[features.FeatureValueID(value)] = value
	}
	manipulationSenses := map[ids.SenseID]bool{}
	for _, name := range []string{"pl.otwierac_obiekt.open", "pl.zamykac_obiekt.close", "pl.niesc.carry", "pl.trzymac.hold", "pl.ciagnac.pull", "pl.pchac.push", "pl.ciac.cut", "pl.skladac.fold", "pl.rozdzierac.tear"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		manipulationSenses[sense] = true
	}
	supportedSenses := map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true, readingSense: true, classificationSense: true, attributionSense: true, eatingSense: true, drinkingSense: true}
	for sense := range manipulationSenses {
		supportedSenses[sense] = true
	}
	predicateAlternatives := map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID{
		seeingSense: append([]map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		}, polishPersonalPredicateFeatures(false)...),
		hearingSense: append([]map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		}, polishPersonalPredicateFeatures(false)...),
		lookingSense: {
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		},
		readingSense: append([]map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		}, polishPersonalPredicateFeatures(true)...),
		receivingSense: {
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		},
		classificationSense: {
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		},
		attributionSense: {
			{"degree": "positive", "number": "singular", "gender": "masculine_animal", "case": "nominative"},
			{"degree": "positive", "number": "singular", "gender": "neuter", "case": "nominative"},
		},
		eatingSense:   {{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"}, {"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"}},
		drinkingSense: {{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"}, {"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"}},
	}
	for _, name := range []string{"pl.otwierac_obiekt.open", "pl.zamykac_obiekt.close", "pl.niesc.carry", "pl.trzymac.hold", "pl.ciagnac.pull", "pl.pchac.push", "pl.ciac.cut", "pl.skladac.fold", "pl.rozdzierac.tear"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		}
	}
	return &runtime.SourcePipeline{
		Language: "pl", ClauseConnector: "i", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}}, AllowFinalPeriod: false,
		ScopedSourcePlan: scopedPlan, ScopedReadingPolicy: ScopedSourceReadingPolicy,
		RelationPathPlan: relationPathPlan, RelationSourceReadingPolicy: RelationSourceReadingPolicy,
		AllowedUnlexicalizedTerminals: map[string]bool{"na": true, "nie": true, "to": true, "ten": true, "jest": true, "są": true, "kto": true, "kogo": true, "i": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: polishWHGapPhraseBuilder{inner: polishPastSeeingPhraseBuilder{inner: polishIdentificationPhraseBuilder{inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "pronoun": {construction.CategoryNP}, "verb": {construction.CategoryVP}, "adjective": {construction.CategoryAP},
			},
			CaseFeature: "case", CaseValues: caseValues,
		}}}}},
		EvidenceBuilder: polishSentenceEvidenceBuilder{lexicon: snapshot.Lexicon(), valency: valencyIndex, constructions: registry, base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses:                     supportedSenses,
			PredicateFeatureAlternativesBySense: predicateAlternatives,
			AllowedUnlexicalizedTerminals:       map[string]bool{"na": true, "nie": true, "to": true, "ten": true, "jest": true},
		}},
	}, nil
}

// polishWHGapPhraseBuilder carries exact token-span evidence for the two
// authored WH terminals; only a matching parser construction consumes it.
type polishWHGapPhraseBuilder struct{ inner runtime.PhraseEvidenceBuilder }

func (builder polishWHGapPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	for _, token := range lattice.Tokens {
		slot, role := "", ""
		switch token.Token.Folded {
		case "kto":
			slot, role = "subject", "experiencer"
		case "kogo":
			slot, role = "object", "stimulus"
		}
		if slot != "" {
			parseContext.WHGaps = append(parseContext.WHGaps, parser.WHGapEvidence{SlotID: slot, Role: meaning.Role(role), ScopeNodeID: meaning.NodeID(fmt.Sprintf("wh_pl_%s_%d_%d", slot, token.Token.Span.Start, token.Token.Span.End)), Span: token.Token.Span})
		}
	}
	return candidates, parseContext, nil
}

type polishPastSeeingPhraseBuilder struct{ inner runtime.PhraseEvidenceBuilder }

func (builder polishPastSeeingPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	for index := range candidates {
		candidate := &candidates[index]
		if candidate.Category == construction.CategoryNP && !isPolishPersonalPronoun(candidate.LexemeID) {
			if candidate.Features == nil {
				candidate.Features = map[features.FeatureID]features.FeatureValueID{}
			}
			candidate.Features["person"] = "third"
		}
	}
	filtered := make([]parser.PhraseCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryVP && !polishPersonalVerbHasCompatibleSubject(candidate, candidates) {
			continue
		}
		if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == "lexeme:pl.widziec.v" &&
			candidate.Features["tense"] == "past" && !polishPastSeeingVerbHasCompatibleSubject(candidate, candidates) {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered, parseContext, nil
}

func polishPastSeeingVerbHasCompatibleSubject(verb parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	hasKnownMismatch, hasCompatible, hasUnknown := false, false, false
	for _, subject := range candidates {
		if subject.Category != construction.CategoryNP || subject.TokenEnd != verb.TokenStart || subject.Features["case"] != "nominative" {
			continue
		}
		subjectGender, subjectNumber := string(subject.Features["gender"]), string(subject.Features["number"])
		verbGender, verbNumber := string(verb.Features["gender"]), string(verb.Features["number"])
		if subjectGender == "" || subjectNumber == "" || verbGender == "" || verbNumber == "" {
			hasUnknown = true
			continue
		}
		expectedGender := subjectGender
		if subjectNumber == "singular" && (subjectGender == "masculine_animal" || subjectGender == "masculine_inanimate") {
			expectedGender = "masculine_personal"
		}
		if subjectNumber == "plural" && subjectGender != "masculine_personal" {
			expectedGender = "non_masculine_personal"
		}
		if verbNumber == subjectNumber && verbGender == expectedGender {
			hasCompatible = true
		} else {
			hasKnownMismatch = true
		}
	}
	return !hasKnownMismatch || hasCompatible || hasUnknown
}

func polishPersonalVerbHasCompatibleSubject(verb parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	person, number := string(verb.Features["person"]), string(verb.Features["number"])
	if person == "" || number == "" {
		return true
	}
	knownMismatch, compatible, unknown := false, false, false
	for _, subject := range candidates {
		if subject.Category != construction.CategoryNP || subject.TokenEnd != verb.TokenStart || subject.Features["case"] != "nominative" {
			continue
		}
		subjectPerson, subjectNumber := string(subject.Features["person"]), string(subject.Features["number"])
		if subjectPerson == "" || subjectNumber == "" {
			unknown = true
			continue
		}
		if subjectPerson == person && subjectNumber == number {
			compatible = true
		} else {
			knownMismatch = true
		}
	}
	return !knownMismatch || compatible || unknown
}

type polishSentenceEvidenceBuilder struct {
	base          runtime.BareDeclarativeEvidenceBuilder
	lexicon       *lexicon.Index
	valency       *valency.Index
	constructions *construction.Registry
}

func (builder polishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	evidence, err := builder.buildSyntax(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	attributes, err := polishNodeAttributes(builder.base.SupportedSenses, builder.lexicon, candidates, forest)
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	evidence.NodeAttributes = attributes
	return evidence, nil
}

func (builder polishSentenceEvidenceBuilder) buildSyntax(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	notSpan, hasNegation := polishTerminalSpan(lattice, "nie")
	if !hasNegation {
		for _, token := range lattice.Tokens {
			if token.Token.Folded == "nie" {
				return runtime.CompositionEvidence{}, nil
			}
		}
		evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
		if err != nil {
			return evidence, err
		}
		if !evidence.OperatorCoverageComplete {
			if !polishWHQuestionCoverage(lattice, forest) {
				return evidence, nil
			}
			evidence.OperatorCoverageComplete = true
		}
		evidence = addPolishIdentificationReferences(evidence, candidates, forest)
		return addPolishPersonalEvidence(evidence, candidates, forest, builder.valency, builder.constructions), nil
	}
	evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	operatorCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			route, supported := polishPresentNegationRouteForSense(alternative.Permission.SenseID)
			if !supported || alternative.FamilyID.String() != route.family || alternative.VariantID.String() != route.variant {
				return runtime.CompositionEvidence{}, nil
			}
			negativeFamily, err := ids.NewConstructionID(strings.TrimPrefix(route.family, "construction:"))
			if err != nil {
				return runtime.CompositionEvidence{}, err
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || !polishNegatedPresentCandidate(candidate, route) {
						continue
					}
					operatorProvenance := polishNegationConstructionSource
					if route.family == "construction:pl.identification_negative" {
						operatorProvenance = polishIdentificationNegationSource
					}
					evidence.Operators = append(evidence.Operators, semantic.OperatorEvidence{
						Operator: construction.Operator{
							ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
							InputFamilyID: negativeFamily, OutputFamilyID: negativeFamily,
							SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
							Capabilities:  []construction.Capability{"negation-negated"}, Provenance: operatorProvenance,
						},
						TargetCandidateID: candidateID, Span: notSpan,
					})
					operatorCount++
				}
			}
		}
	}
	evidence.OperatorCoverageComplete = operatorCount > 0
	if !evidence.OperatorCoverageComplete {
		return evidence, nil
	}
	return addPolishIdentificationReferences(evidence, candidates, forest), nil
}

func polishWHQuestionCoverage(lattice analyzer.LexicalLattice, forest parser.Forest) bool {
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
			if alternative.FamilyID.String() != "construction:pl.question.wh" {
				return false
			}
			terminal, slotID, role := "", "", ""
			switch alternative.VariantID.String() {
			case "construction:pl.question.wh.subject":
				terminal, slotID, role = "kto", "subject", "experiencer"
			case "construction:pl.question.wh.object.fronted":
				terminal, slotID, role = "kogo", "object", "stimulus"
			default:
				return false
			}
			terminalCount, gapCount := 0, 0
			for _, token := range lattice.Tokens {
				if token.Token.Folded == terminal {
					terminalCount++
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
			if terminalCount != 1 || gapCount > 1 {
				return false
			}
			if gapCount == 1 {
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
			if !polishSameParseExceptGap(*gapAlternative, omission) {
				return false
			}
		}
	}
	return seenClause
}

func polishSameParseExceptGap(withGap, withoutGap parser.ParseAlternative) bool {
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
