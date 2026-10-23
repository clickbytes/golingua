package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/text"
)

// NewSentenceSource builds the trusted Polish source stages against one exact
// CORE/lexicon snapshot. The bare declarative completion attestation covers
// singular and plural present visual/auditory/looking and singular past receiving predicates.
func NewSentenceSource(snapshot *pack.Pack) (*runtime.SourcePipeline, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("Polish sentence source requires a snapshot")
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
	caseValues := map[features.FeatureValueID]string{}
	for _, value := range []string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"} {
		caseValues[features.FeatureValueID(value)] = value
	}
	return &runtime.SourcePipeline{
		Language: "pl", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}},
		AllowedUnlexicalizedTerminals: map[string]bool{"na": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "verb": {construction.CategoryVP},
			},
			CaseFeature: "case", CaseValues: caseValues,
		}},
		EvidenceBuilder: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses: map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true},
			PredicateFeatureAlternativesBySense: map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID{
				seeingSense: {
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
				},
				hearingSense: {
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
				},
				lookingSense: {
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
				},
				receivingSense: {{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular"}},
			},
			AllowedUnlexicalizedTerminals: map[string]bool{"na": true},
		},
	}, nil
}
