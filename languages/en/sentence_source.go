package en

import (
	"context"
	"fmt"

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
// third-person singular and plural present visual and auditory predicates,
// singular visual negation, explicit articles, and the registered past receiving route.
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
	return &runtime.SourcePipeline{
		Language: "en", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}},
		AllowedUnlexicalizedTerminals: map[string]bool{"at": true, "does": true, "not": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: englishDeterminerNPBuilder{inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "verb": {construction.CategoryVP},
			},
		}}, lexicon: snapshot.Lexicon()},
		EvidenceBuilder: englishSentenceEvidenceBuilder{base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses: map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true},
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
				// English simple past is finite but has one form for all persons
				// and numbers, so no agreement feature is required for this cell.
				receivingSense: {{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"}},
			},
			AllowedUnlexicalizedTerminals: map[string]bool{"at": true, "does": true, "not": true},
		}},
	}, nil
}

type englishSentenceEvidenceBuilder struct {
	base runtime.BareDeclarativeEvidenceBuilder
}

func (builder englishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	notSpan, hasNot := englishTerminalSpan(lattice, "not")
	if !hasNot {
		return builder.base.Build(ctx, lattice, candidates, forest)
	}
	if !englishDoesNotSeeInput(lattice) {
		return runtime.CompositionEvidence{}, nil
	}
	seeingSense, err := ids.NewSenseID("en.see.visual")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	base := builder.base
	base.PredicateFeatureAlternativesBySense = cloneEnglishPredicateAlternatives(builder.base.PredicateFeatureAlternativesBySense)
	base.PredicateFeatureAlternativesBySense[seeingSense] = append(base.PredicateFeatureAlternativesBySense[seeingSense], map[features.FeatureID]features.FeatureValueID{"verb_form": "base"})
	evidence, err := base.Build(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	negativeFamily, err := ids.NewConstructionID("en.transitive_negative")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	negativeVariant, err := ids.NewConstructionID("en.transitive_negative.svo")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	operatorCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != negativeFamily || alternative.VariantID.String() != negativeVariant.String() {
				return runtime.CompositionEvidence{}, nil
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || !englishNegatedSeeCandidate(lattice, candidate) {
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

func cloneEnglishPredicateAlternatives(source map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID) map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID {
	result := make(map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, len(source))
	for sense, alternatives := range source {
		result[sense] = append([]map[features.FeatureID]features.FeatureValueID(nil), alternatives...)
	}
	return result
}

func englishTerminalSpan(lattice analyzer.LexicalLattice, terminal string) (text.Span, bool) {
	var span text.Span
	found := false
	for _, token := range lattice.Tokens {
		if token.Token.Folded != terminal {
			continue
		}
		if found {
			return text.Span{}, false
		}
		span, found = token.Token.Span, true
	}
	return span, found
}

func englishDoesNotSeeInput(lattice analyzer.LexicalLattice) bool {
	for index := 0; index+2 < len(lattice.Tokens); index++ {
		if lattice.Tokens[index].Token.Folded == "does" && lattice.Tokens[index+1].Token.Folded == "not" && lattice.Tokens[index+2].Token.Folded == "see" {
			return true
		}
	}
	return false
}

func englishNegatedSeeCandidate(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	return candidate.Features["verb_form"] == "base" && englishNegatedSeeToken(lattice, candidate)
}

func englishNegatedSeeToken(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	if candidate.Category != construction.CategoryVP || candidate.LexemeID.String() != "lexeme:en.see.v" || candidate.TokenStart < 2 || candidate.TokenStart >= len(lattice.Tokens) {
		return false
	}
	return lattice.Tokens[candidate.TokenStart-2].Token.Folded == "does" && lattice.Tokens[candidate.TokenStart-1].Token.Folded == "not"
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
			if englishNegatedSeeToken(lattice, candidate) {
				if candidate.Features["mood"] != "imperative" || negatedBaseSeen[candidate.TokenStart] {
					continue
				}
				// In this exact auxiliary context, the imperative cell supplies
				// the English base surface; finite readings are not licensed here.
				candidate.Features = map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "number": "singular", "verb_number": "singular",
				}
				negatedBaseSeen[candidate.TokenStart] = true
			}
			if !supportedEnglishPredicateRoute(candidate) && !englishNegatedSeeCandidate(lattice, candidate) {
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

func englishNegationAuxiliaryCandidate(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	start := candidate.TokenStart
	return candidate.Category == construction.CategoryVP && start >= 0 && start+1 < len(lattice.Tokens) &&
		lattice.Tokens[start].Token.Folded == "does" && lattice.Tokens[start+1].Token.Folded == "not"
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
			if sense.String() == "sense:en.get.receive" {
				return true
			}
		}
	}
	return false
}
