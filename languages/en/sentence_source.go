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
	"github.com/clickbytes/golingua/internal/text"
)

// NewSentenceSource compiles the English analysis stages against one exact
// snapshot. Its complete declarative attestation is limited to attested
// third-person singular and plural present visual and auditory predicates, explicit
// definite and indefinite articles, and the registered past receiving route.
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
		AllowedUnlexicalizedTerminals: map[string]bool{"at": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: englishDeterminerNPBuilder{inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "verb": {construction.CategoryVP},
			},
		}}, lexicon: snapshot.Lexicon()},
		EvidenceBuilder: runtime.BareDeclarativeEvidenceBuilder{
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
			AllowedUnlexicalizedTerminals: map[string]bool{"at": true},
		},
	}, nil
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
	for _, candidate := range candidates {
		if candidate.Features == nil {
			candidate.Features = make(map[features.FeatureID]features.FeatureValueID)
		}
		if candidate.Category == construction.CategoryVP {
			if number := candidate.Features["verb_number"]; number != "" {
				candidate.Features["number"] = number
			}
			if !supportedEnglishPredicateRoute(candidate) {
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
			if sense.String() == "sense:en.get.receive" {
				return true
			}
		}
	}
	return false
}
