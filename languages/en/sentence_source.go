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
// third-person singular present visual and auditory predicates and explicit
// definite articles in the transitive template.
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
		}}},
		EvidenceBuilder: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses: map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true},
			PredicateFeatureAlternativesBySense: map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID{
				seeingSense:  {{"verb_form": "third_person_singular"}},
				hearingSense: {{"verb_form": "third_person_singular"}},
				lookingSense: {{"verb_form": "third_person_singular"}},
				// The English principal-parts paradigm exposes its past cell as
				// verb_form=past. Do not require tense here: that value is not
				// currently declared by the English analysis cell.
				receivingSense: {{"verb_form": "past"}},
			},
			AllowedUnlexicalizedTerminals: map[string]bool{"at": true},
		},
	}, nil
}

type englishDeterminerNPBuilder struct {
	inner runtime.SimplePhraseEvidenceBuilder
}

func (builder englishDeterminerNPBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	for index := range candidates {
		candidate := &candidates[index]
		if candidate.Category != construction.CategoryNP || candidate.TokenStart == 0 {
			continue
		}
		previous := lattice.Tokens[candidate.TokenStart-1]
		for _, reading := range previous.Readings {
			if reading.HasLexeme && (reading.Lexeme.ID.String() == "lexeme:en.the.det" || reading.Lexeme.ID.String() == "lexeme:en.a.det") {
				if candidate.Features == nil {
					candidate.Features = make(map[features.FeatureID]features.FeatureValueID)
				}
				if reading.Lexeme.ID.String() == "lexeme:en.a.det" {
					candidate.Features["determiner"] = "indefinite"
				} else {
					candidate.Features["determiner"] = "definite"
				}
				break
			}
		}
	}
	return candidates, parseContext, nil
}
