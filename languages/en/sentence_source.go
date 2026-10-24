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
		Language: "en", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}},
		AllowedUnlexicalizedTerminals: map[string]bool{"at": true, "do": true, "does": true, "not": true, "this": true, "that": true, "it": true, "is": true, "a": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: englishIdentificationPhraseBuilder{inner: englishDeterminerNPBuilder{inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "verb": {construction.CategoryVP}, "adjective": {construction.CategoryAP},
			},
		}}, lexicon: snapshot.Lexicon()}},
		EvidenceBuilder: englishSentenceEvidenceBuilder{lexicon: snapshot.Lexicon(), base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses: map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true, readingSense: true, classificationSense: true, attributionSense: true},
			PredicateFeatureAlternativesBySense: map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID{
				seeingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
				},
				hearingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
				},
				lookingSense: {
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
					{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
				},
				readingSense: {
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
		if err != nil || !evidence.OperatorCoverageComplete {
			return evidence, err
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
	filtered := candidates[:0]
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
			if !supportedEnglishPredicateRoute(candidate) && (!negated || candidate.LexemeID.String() != route.lexeme) {
				continue
			}
			filtered = append(filtered, candidate)
			continue
		}
		if candidate.Category != construction.CategoryNP || candidate.TokenStart == 0 {
			if candidate.Category == construction.CategoryNP {
				if number := candidate.Features["noun_number"]; number != "" {
					candidate.Features["number"] = number
				}
			}
			filtered = append(filtered, candidate)
			continue
		}
		previous := lattice.Tokens[candidate.TokenStart-1]
		keep := true
		for _, reading := range previous.Readings {
			if reading.HasLexeme && (reading.Lexeme.ID.String() == "lexeme:en.the.det" || reading.Lexeme.ID.String() == "lexeme:en.a.det" || reading.Lexeme.ID.String() == "lexeme:en.an.det") {
				indefinite := reading.Lexeme.ID.String() == "lexeme:en.a.det" || reading.Lexeme.ID.String() == "lexeme:en.an.det"
				if indefinite {
					if number := candidate.Features["noun_number"]; number != "" && number != "singular" {
						keep = false
						break
					}
					if candidate.LexemeID.String() != "" && builder.lexicon != nil {
						if noun, exists := builder.lexicon.ByLexemeID(candidate.LexemeID); exists {
							article := "a"
							if reading.Lexeme.ID.String() == "lexeme:en.an.det" {
								article = "an"
							}
							if expectedArticle, hasLexicalArticle := englishIndefiniteArticle(noun.Properties); hasLexicalArticle && expectedArticle != article {
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
			if number := candidate.Features["noun_number"]; number != "" {
				candidate.Features["number"] = number
			}
			filtered = append(filtered, candidate)
		}
	}
	return filtered, parseContext, nil
}

func supportedEnglishPredicateRoute(candidate parser.PhraseCandidate) bool {
	if candidate.Features["verb_form"] != "finite" {
		return false
	}
	switch candidate.Features["tense"] {
	case "present":
		return candidate.Features["person"] == "third" &&
			(candidate.Features["verb_number"] == "singular" || candidate.Features["verb_number"] == "plural")
	case "past":
		for _, sense := range candidate.SenseIDs {
			if sense.String() == "sense:en.get.receive" || sense.String() == "sense:en.read.text" {
				return true
			}
		}
	}
	return false
}
