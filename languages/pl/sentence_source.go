package pl

import (
	"context"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
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
	phraseProfile, err := SentencePhraseProfile(profile)
	if err != nil {
		return nil, err
	}
	valencyIndex, err := CompileValency()
	if err != nil {
		return nil, err
	}
	packages, registry, err := SentenceConstructions(valencyIndex)
	if err != nil {
		return nil, err
	}
	negationRoutes := newPolishNegationRouteResolver(snapshot.Lexicon(), valencyIndex, registry, packages)
	plans, variants, err := runtime.CompileSentencePlans(profile, packages, valencyIndex)
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
	helpSense, err := ids.NewSenseID("pl.pomagac.help")
	if err != nil {
		return nil, err
	}
	thankSense, err := ids.NewSenseID("pl.dziekowac.thank")
	if err != nil {
		return nil, err
	}
	usingSense, err := ids.NewSenseID("pl.uzywac.using")
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
	redSense, err := ids.NewSenseID("pl.czerwony.red")
	if err != nil {
		return nil, err
	}
	attributionSenses := map[ids.SenseID]bool{attributionSense: true, redSense: true}
	for _, spec := range polishAttributionSpecs {
		sense, senseErr := ids.NewSenseID(strings.TrimPrefix(spec.sense, "sense:"))
		if senseErr != nil {
			return nil, senseErr
		}
		attributionSenses[sense] = true
	}
	eatingSense, err := ids.NewSenseID("pl.jesc.eat")
	if err != nil {
		return nil, err
	}
	drinkingSense, err := ids.NewSenseID("pl.pic.drink")
	if err != nil {
		return nil, err
	}
	buyingSense, err := ids.NewSenseID("pl.kupowac.buy")
	if err != nil {
		return nil, err
	}
	perfectiveBuyingSense, err := ids.NewSenseID("pl.kupic.buy")
	if err != nil {
		return nil, err
	}
	sellingSense, err := ids.NewSenseID("pl.sprzedawac.sell")
	if err != nil {
		return nil, err
	}
	caseValues := map[features.FeatureValueID]string{}
	for _, value := range []string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"} {
		caseValues[features.FeatureValueID(value)] = value
	}
	manipulationSenses := map[ids.SenseID]bool{}
	for _, name := range []string{"pl.otwierac_obiekt.open", "pl.zamykac_obiekt.close", "pl.niesc.carry", "pl.trzymac.hold", "pl.ciagnac.pull", "pl.pchac.push", "pl.ciac.cut", "pl.kroic.with_instrument", "pl.skladac.fold", "pl.rozdzierac.tear"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		manipulationSenses[sense] = true
	}
	givingSense, err := ids.NewSenseID("pl.dawac.transfer")
	if err != nil {
		return nil, err
	}
	perfectiveGivingSense, err := ids.NewSenseID("pl.dac.transfer")
	if err != nil {
		return nil, err
	}
	supportedSenses := map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true, readingSense: true, classificationSense: true, attributionSense: true, redSense: true, eatingSense: true, drinkingSense: true, givingSense: true, perfectiveGivingSense: true, buyingSense: true, perfectiveBuyingSense: true, sellingSense: true, helpSense: true, thankSense: true, usingSense: true}
	for sense := range manipulationSenses {
		supportedSenses[sense] = true
	}
	for sense := range attributionSenses {
		supportedSenses[sense] = true
	}
	affectiveSenses := map[ids.SenseID]bool{}
	for _, name := range []string{"pl.lubic.like", "pl.ogladac.view", "pl.pisac.document"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		affectiveSenses[sense] = true
		supportedSenses[sense] = true
	}
	writingMessageSense, err := ids.NewSenseID("pl.pisac.write")
	if err != nil {
		return nil, err
	}
	supportedSenses[writingMessageSense] = true
	writingInstrumentSense, err := ids.NewSenseID("pl.pisac.instrument")
	if err != nil {
		return nil, err
	}
	supportedSenses[writingInstrumentSense] = true
	conversingSense, err := ids.NewSenseID("pl.rozmawiac.converse")
	if err != nil {
		return nil, err
	}
	supportedSenses[conversingSense] = true
	intransitiveSenses := approvedPolishIntransitiveSenses(registry, valencyIndex, variants)
	for sense := range intransitiveSenses {
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
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		}, polishPersonalPredicateFeatures(false)...),
		drinkingSense: {
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		},
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
		redSense: {
			{"degree": "positive", "number": "singular", "gender": "feminine", "case": "nominative"},
		},
		eatingSense: {{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"}, {"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"}, {"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"}, {"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"}, {"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"}, {"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"}, {"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"}},
		givingSense: {
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		},
		perfectiveGivingSense: {
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular", "gender": "feminine"},
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular", "gender": "neuter"},
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
			{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
		},
	}
	for sense := range attributionSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"degree": "positive", "number": "singular", "gender": "masculine_animal", "case": "nominative"},
			{"degree": "positive", "number": "singular", "gender": "masculine_inanimate", "case": "nominative"},
			{"degree": "positive", "number": "singular", "gender": "neuter", "case": "nominative"},
			{"degree": "positive", "number": "singular", "gender": "feminine", "case": "nominative"},
			{"degree": "positive", "number": "plural", "gender": "masculine_personal", "case": "nominative"},
			{"degree": "positive", "number": "plural", "gender": "non_masculine_personal", "case": "nominative"},
		}
		if sense.String() == "sense:pl.otwierac_obiekt.open" || sense.String() == "sense:pl.zamykac_obiekt.close" || sense.String() == "sense:pl.trzymac.hold" || sense.String() == "sense:pl.niesc.carry" || sense.String() == "sense:pl.pchac.push" || sense.String() == "sense:pl.ciagnac.pull" || sense.String() == "sense:pl.ciac.cut" || sense.String() == "sense:pl.rozdzierac.tear" {
			predicateAlternatives[sense] = append(predicateAlternatives[sense],
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
			)
		}
	}
	commerceFeatures := []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
		{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
	}
	sellingFeatures := append(append([]map[features.FeatureID]features.FeatureValueID(nil), commerceFeatures...),
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"})
	predicateAlternatives[buyingSense] = append(append([]map[features.FeatureID]features.FeatureValueID{}, commerceFeatures...),
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
		map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"})
	predicateAlternatives[perfectiveBuyingSense] = []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
		{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular", "gender": "feminine"},
		{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular", "gender": "neuter"},
		{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
		{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
	}
	predicateAlternatives[sellingSense] = sellingFeatures
	predicateAlternatives[helpSense] = commerceFeatures
	predicateAlternatives[thankSense] = commerceFeatures
	predicateAlternatives[usingSense] = commerceFeatures
	for _, name := range []string{"pl.otwierac_obiekt.open", "pl.zamykac_obiekt.close", "pl.niesc.carry", "pl.trzymac.hold", "pl.ciagnac.pull", "pl.pchac.push", "pl.ciac.cut", "pl.kroic.with_instrument", "pl.skladac.fold", "pl.rozdzierac.tear"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		}
		if name == "pl.pchac.push" || name == "pl.ciagnac.pull" || name == "pl.ciac.cut" {
			predicateAlternatives[sense] = append(predicateAlternatives[sense],
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
				map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
			)
		}
	}
	for sense := range intransitiveSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		}
	}
	for sense := range affectiveSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
			{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
		}
	}
	predicateAlternatives[writingMessageSense] = []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "singular"},
		{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "plural"},
		{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "singular", "gender": "masculine_personal"},
		{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "singular", "gender": "feminine"},
		{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "singular", "gender": "neuter"},
		{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "plural", "gender": "masculine_personal"},
		{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
	}
	predicateAlternatives[writingInstrumentSense] = predicateAlternatives[writingMessageSense][:2]
	predicateAlternatives[conversingSense] = []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "singular"},
		{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "plural"},
	}
	addApprovedPolishTransitiveSourceRoutes(supportedSenses, predicateAlternatives, snapshot.Lexicon(), valencyIndex, registry, packages)
	return &runtime.SourcePipeline{
		Language: "pl", PhraseProfile: phraseProfile, ClauseConnector: "i", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}}, AllowFinalPeriod: false,
		ScopedSourcePlan: scopedPlan, ScopedReadingPolicy: ScopedSourceReadingPolicy,
		RelationPathPlan: relationPathPlan, RelationSourceReadingPolicy: RelationSourceReadingPolicy,
		AllowedUnlexicalizedTerminals: map[string]bool{"na": true, "od": true, "do": true, "z": true, "nie": true, "to": true, "ten": true, "ta": true, "jest": true, "są": true, "te": true, "ci": true, "tamte": true, "tamci": true, "kto": true, "kogo": true, "i": true, "się": true, "tamten": true, "tamta": true, "tamto": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: polishWHGapPhraseBuilder{inner: polishPastAgreementPhraseBuilder{inner: polishIdentificationPhraseBuilder{phraseProfile: phraseProfile, featureProfile: profile, inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "pronoun": {construction.CategoryNP}, "verb": {construction.CategoryVP}, "verb_reflexive": {construction.CategoryVP}, "adjective": {construction.CategoryAP},
			},
			CaseFeature: "case", CaseValues: caseValues,
		}}}}},
		EvidenceBuilder: polishSentenceEvidenceBuilder{lexicon: snapshot.Lexicon(), core: snapshot.CoreGraph(), valency: valencyIndex, constructions: registry, packages: packages, negationRoutes: negationRoutes, base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses:                     supportedSenses,
			PredicateFeatureAlternativesBySense: predicateAlternatives,
			AllowedUnlexicalizedTerminals:       map[string]bool{"na": true, "od": true, "do": true, "z": true, "nie": true, "to": true, "ten": true, "ta": true, "jest": true, "są": true, "te": true, "ci": true, "tamte": true, "tamci": true, "się": true, "tamten": true, "tamta": true, "tamto": true},
		}},
	}, nil
}

func addApprovedPolishTransitiveSourceRoutes(supported map[ids.SenseID]bool, alternatives map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, lexiconIndex *lexicon.Index, valencyIndex *valency.Index, registry *construction.Registry, packages []construction.CompiledConstructionPackage) {
	if lexiconIndex == nil || valencyIndex == nil || registry == nil {
		return
	}
	frames := make(map[string]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[string(frame.ID)] = frame
	}
	variants := make(map[string]construction.Variant)
	for _, pkg := range packages {
		for _, variant := range pkg.Package.Variants {
			variants[pkg.Package.Family.ID.String()+"\x00"+variant.ID.String()] = variant
		}
	}
	for _, familyName := range []string{"construction:pl.transitive", "construction:pl.transitive_personal"} {
		familyID, err := ids.ParseConstructionID(familyName)
		if err != nil {
			continue
		}
		if _, found := registry.ByID(familyID); !found {
			continue
		}
		for _, binding := range registry.BindingsByFamilyID(familyID) {
			frame, frameOK := frames[string(binding.ValencyFrameID)]
			variant, variantOK := variants[familyName+"\x00"+binding.VariantID.String()]
			sense, senseOK := lexiconIndex.BySenseID(binding.SenseID)
			if !binding.Approved || binding.FamilyID != familyID || !frameOK || !variantOK || !senseOK || frame.SenseID != binding.SenseID || !polishFrameAllowsFamily(frame, familyID) {
				continue
			}
			lexeme, lexemeOK := lexiconIndex.ByLexemeID(sense.LexemeID)
			semanticBindings := lexiconIndex.BindingsForSense(binding.SenseID)
			if !lexemeOK || lexeme.Language != "pl" || lexeme.POS != "verb" || len(semanticBindings) != 1 || !semanticBindings[0].TransferAuthorized() || len(frame.SemanticRoles) != 2 || len(frame.Arguments) != 2 {
				continue
			}
			required := map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "aspect": "imperfective"}
			for _, condition := range variant.Conditions {
				if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == "verb" {
					required[condition.Feature] = condition.Value
				}
			}
			supported[binding.SenseID] = true
			alternatives[binding.SenseID] = append(alternatives[binding.SenseID], required)
		}
	}
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

type polishPastAgreementPhraseBuilder struct{ inner runtime.PhraseEvidenceBuilder }

func (builder polishPastAgreementPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
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
		if candidate.Category == construction.CategoryVP && !polishPersonalVerbHasCompatibleSubject(lattice, candidate, candidates) {
			continue
		}
		if candidate.Category == construction.CategoryVP && (candidate.LexemeID.String() == "lexeme:pl.czytac.v" || candidate.LexemeID.String() == "lexeme:pl.widziec.v" || candidate.LexemeID.String() == "lexeme:pl.slyszec.v" || candidate.LexemeID.String() == "lexeme:pl.pic.v" || candidate.LexemeID.String() == "lexeme:pl.spac.v" || candidate.LexemeID.String() == "lexeme:pl.oddychac.v" || candidate.LexemeID.String() == "lexeme:pl.kichac.v" || candidate.LexemeID.String() == "lexeme:pl.kaszlec.v" || candidate.LexemeID.String() == "lexeme:pl.jesc.v" || candidate.LexemeID.String() == "lexeme:pl.pisac.v" || candidate.LexemeID.String() == "lexeme:pl.dac.v" || candidate.LexemeID.String() == "lexeme:pl.dawac.v" || candidate.LexemeID.String() == "lexeme:pl.kupowac.v" || candidate.LexemeID.String() == "lexeme:pl.kupic.v" || candidate.LexemeID.String() == "lexeme:pl.otwierac_obiekt.v" || candidate.LexemeID.String() == "lexeme:pl.zamykac_obiekt.v" || candidate.LexemeID.String() == "lexeme:pl.trzymac.v" || candidate.LexemeID.String() == "lexeme:pl.niesc.v" || candidate.LexemeID.String() == "lexeme:pl.pchac.v" || candidate.LexemeID.String() == "lexeme:pl.ciagnac.v" || candidate.LexemeID.String() == "lexeme:pl.ciac.v" || candidate.LexemeID.String() == "lexeme:pl.skladac.v" || candidate.LexemeID.String() == "lexeme:pl.biec.v" || candidate.LexemeID.String() == "lexeme:pl.rozdzierac.v" || candidate.LexemeID.String() == "lexeme:pl.lubic.v") &&
			candidate.Features["tense"] == "past" && !polishPastAgreementVerbHasCompatibleSubject(lattice, candidate, candidates) {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return filtered, parseContext, nil
}

func polishPastAgreementVerbHasCompatibleSubject(lattice analyzer.LexicalLattice, verb parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	if verb.LexemeID.String() == "lexeme:pl.czytac.v" {
		return polishReadingPastAgreementVerbHasCompatibleSubject(lattice, verb, candidates)
	}
	if verb.LexemeID.String() == "lexeme:pl.pisac.v" {
		return polishWritingPastAgreementVerbHasSVOCompatibleSubject(lattice, verb, candidates)
	}
	hasKnownMismatch, hasCompatible, hasUnknown := false, false, false
	for _, subject := range candidates {
		adjacent := subject.TokenEnd == verb.TokenStart || subject.TokenStart == verb.TokenEnd
		if !adjacent && subject.TokenEnd+1 == verb.TokenStart && subject.TokenEnd < len(lattice.Tokens) && lattice.Tokens[subject.TokenEnd].Token.Folded == "nie" {
			adjacent = true
		}
		if subject.Category != construction.CategoryNP || !adjacent || subject.Features["case"] != "nominative" {
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
	if hasCompatible {
		return true
	}
	if hasKnownMismatch {
		return false
	}
	return !hasKnownMismatch || hasUnknown
}

// Writing-message is licensed in SVO only. Its object can be nominative-
// syncretic, so past agreement must use a nominative NP before the verb (with
// the optional intervening negation particle), never a postverbal NP.
func polishWritingPastAgreementVerbHasSVOCompatibleSubject(lattice analyzer.LexicalLattice, verb parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	hasKnownMismatch, hasCompatible, hasUnknown := false, false, false
	for _, subject := range candidates {
		beforeVerb := subject.TokenEnd == verb.TokenStart ||
			(subject.TokenEnd+1 == verb.TokenStart && subject.TokenEnd < len(lattice.Tokens) && lattice.Tokens[subject.TokenEnd].Token.Folded == "nie")
		if subject.Category != construction.CategoryNP || !beforeVerb || subject.Features["case"] != "nominative" {
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
	if hasCompatible {
		return true
	}
	if hasKnownMismatch {
		return false
	}
	return hasUnknown
}

func polishReadingPastAgreementVerbHasCompatibleSubject(lattice analyzer.LexicalLattice, verb parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	var beforeVerb, afterVerb []parser.PhraseCandidate
	for _, candidate := range candidates {
		if candidate.Category != construction.CategoryNP || candidate.Features["case"] != "nominative" {
			continue
		}
		if candidate.TokenEnd == verb.TokenStart || candidate.TokenEnd+1 == verb.TokenStart && candidate.TokenEnd < len(lattice.Tokens) && lattice.Tokens[candidate.TokenEnd].Token.Folded == "nie" {
			beforeVerb = append(beforeVerb, candidate)
		} else if candidate.TokenStart == verb.TokenEnd {
			afterVerb = append(afterVerb, candidate)
		}
	}
	subjects := beforeVerb
	if len(subjects) == 0 {
		subjects = afterVerb
	}
	if len(subjects) == 0 {
		return true
	}
	hasKnownMismatch, hasCompatible, hasUnknown := false, false, false
	for _, subject := range subjects {
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
	if hasCompatible {
		return true
	}
	if hasKnownMismatch {
		return false
	}
	return hasUnknown
}

func polishPersonalVerbHasCompatibleSubject(lattice analyzer.LexicalLattice, verb parser.PhraseCandidate, candidates []parser.PhraseCandidate) bool {
	person, number := string(verb.Features["person"]), string(verb.Features["number"])
	if person == "" || number == "" {
		return true
	}
	knownMismatch, compatible, unknown := false, false, false
	hasAdjacentNonNominative := false
	for _, subject := range candidates {
		adjacent := subject.TokenEnd == verb.TokenStart || subject.TokenStart == verb.TokenEnd
		if !adjacent && subject.TokenEnd+1 == verb.TokenStart && subject.TokenEnd < len(lattice.Tokens) && lattice.Tokens[subject.TokenEnd].Token.Folded == "nie" {
			adjacent = true
		}
		if subject.Category != construction.CategoryNP || !adjacent {
			continue
		}
		if subject.Features["case"] != "nominative" {
			hasAdjacentNonNominative = true
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
	return !knownMismatch || compatible || unknown || hasAdjacentNonNominative
}

type polishSentenceEvidenceBuilder struct {
	base           runtime.BareDeclarativeEvidenceBuilder
	lexicon        *lexicon.Index
	core           *core.ConceptGraph
	valency        *valency.Index
	constructions  *construction.Registry
	packages       []construction.CompiledConstructionPackage
	negationRoutes polishNegationRouteResolver
}

func (builder polishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	if routes := polishRelationSourceFeatureAlternatives(lattice, candidates, forest, builder.constructions, builder.packages, builder.lexicon, builder.core, builder.valency); len(routes) > 0 {
		builder.base.SupportedSenses = clonePolishSupportedSenses(builder.base.SupportedSenses)
		builder.base.PredicateFeatureAlternativesBySense = clonePolishPredicateAlternatives(builder.base.PredicateFeatureAlternativesBySense)
		for sense, alternatives := range routes {
			builder.base.SupportedSenses[sense] = true
			builder.base.PredicateFeatureAlternativesBySense[sense] = alternatives
		}
	}
	evidence, err := builder.buildSyntax(ctx, lattice, candidates, forest)
	if err != nil {
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
		pastTransitiveFeatures, pastTransitiveCoverage := polishTransitiveSimplePastCapability(candidates, forest, builder.valency, builder.constructions)
		pastIntransitiveFeatures, pastIntransitiveCoverage := polishIntransitiveSimplePastCapability(candidates, forest, builder.constructions)
		if !pastIntransitiveCoverage {
			pastIntransitiveFeatures = nil
		}
		base := builder.base
		if len(pastIntransitiveFeatures) > 0 {
			base.PredicateFeatureAlternativesBySense = clonePolishPredicateAlternatives(base.PredicateFeatureAlternativesBySense)
			for sense, alternatives := range pastIntransitiveFeatures {
				base.PredicateFeatureAlternativesBySense[sense] = append(base.PredicateFeatureAlternativesBySense[sense], alternatives...)
			}
		}
		if len(pastTransitiveFeatures) > 0 {
			base.PredicateFeatureAlternativesBySense = clonePolishPredicateAlternatives(base.PredicateFeatureAlternativesBySense)
			for sense, alternatives := range pastTransitiveFeatures {
				base.PredicateFeatureAlternativesBySense[sense] = append(base.PredicateFeatureAlternativesBySense[sense], alternatives...)
			}
		}
		evidence, err := base.Build(ctx, lattice, candidates, forest)
		if err != nil {
			return evidence, err
		}
		if !evidence.OperatorCoverageComplete {
			if !polishWHQuestionCoverage(lattice, forest) && !pastIntransitiveCoverage && !pastTransitiveCoverage {
				return evidence, nil
			}
			evidence.OperatorCoverageComplete = true
		}
		evidence = addPolishIdentificationReferences(evidence, lattice, candidates, forest)
		return addPolishPersonalEvidence(evidence, candidates, forest, builder.valency, builder.constructions), nil
	}
	evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
	if err != nil {
		return evidence, err
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	operatorCount := 0
	seenOperatorEvidence := make(map[string]struct{})
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			route, supported := builder.negationRoutes.routeForSense(alternative.Permission.SenseID)
			if !supported || !polishPresentNegationVariantAllowed(route, alternative.FamilyID.String(), alternative.VariantID.String()) {
				return runtime.CompositionEvidence{}, nil
			}
			if route.derived && !polishAlternativeHasRequiredNegationSlots(alternative, route, byID) {
				return runtime.CompositionEvidence{}, nil
			}
			negativeFamily, err := ids.NewConstructionID(strings.TrimPrefix(alternative.FamilyID.String(), "construction:"))
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
					operatorKey := fmt.Sprintf("%s|%d|%d", candidateID, notSpan.Start, notSpan.End)
					if _, duplicate := seenOperatorEvidence[operatorKey]; duplicate {
						continue
					}
					seenOperatorEvidence[operatorKey] = struct{}{}
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
	evidence = addPolishIdentificationReferences(evidence, lattice, candidates, forest)
	return addPolishPersonalEvidence(evidence, candidates, forest, builder.valency, builder.constructions), nil
}

func polishAlternativeHasRequiredNegationSlots(alternative parser.ParseAlternative, route polishPresentNegationRoute, candidates map[string]parser.PhraseCandidate) bool {
	specs := route.slotsByFamily[alternative.FamilyID.String()]
	if len(alternative.Unresolved) > 0 || len(specs) == 0 {
		return false
	}
	slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
	for _, slot := range alternative.Slots {
		slots[slot.SlotID] = slot
	}
	for slotID, category := range specs {
		slot, exists := slots[slotID]
		if !exists || len(slot.CandidateIDs) == 0 {
			return false
		}
		for _, candidateID := range slot.CandidateIDs {
			candidate, found := candidates[candidateID]
			if !found || candidate.Category != category {
				return false
			}
		}
	}
	return true
}

// polishTransitiveSimplePastCapability derives past predicate alternatives and
// operator coverage from retained, fully bound transitive plans. It only
// authorizes a clause when its approved binding, required slots, finite past VP,
// and nominative subject agreement all remain available in the parse forest.
func polishTransitiveSimplePastCapability(candidates []parser.PhraseCandidate, forest parser.Forest, valencyIndex *valency.Index, registry *construction.Registry) (map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, bool) {
	allowed := make(map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID)
	if valencyIndex == nil || registry == nil {
		return allowed, false
	}
	familyID, err := ids.ParseConstructionID("construction:pl.transitive")
	if err != nil {
		return allowed, false
	}
	familyRecord, exists := registry.ByID(familyID)
	if !exists {
		return allowed, false
	}
	bindings := registry.BindingsByFamilyID(familyID)
	frames := make(map[string]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[string(frame.ID)] = frame
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	seenClause := false
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		seenClause = true
		if len(node.Alternatives) == 0 {
			return allowed, false
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != familyID || len(alternative.Unresolved) != 0 {
				return allowed, false
			}
			var approvedBinding *valency.ConstructionBinding
			for index := range bindings {
				binding := &bindings[index]
				if binding.Approved && binding.SenseID == alternative.Permission.SenseID && string(binding.ValencyFrameID) == string(alternative.Permission.FrameID) && binding.FamilyID == alternative.FamilyID && binding.VariantID.String() == alternative.VariantID.String() {
					approvedBinding = binding
					break
				}
			}
			if approvedBinding == nil {
				return allowed, false
			}
			frame, frameExists := frames[string(approvedBinding.ValencyFrameID)]
			if !frameExists || frame.SenseID != approvedBinding.SenseID || !polishFrameAllowsFamily(frame, familyID) {
				return allowed, false
			}
			subjectRole, objectRole, subjectCase, objectCase, frameOK := polishTransitivePastFrameRoles(familyRecord.Family, frame)
			if !frameOK {
				return allowed, false
			}
			slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
			for _, slot := range alternative.Slots {
				if _, duplicate := slots[slot.SlotID]; duplicate {
					return allowed, false
				}
				slots[slot.SlotID] = slot
			}
			for _, spec := range familyRecord.Family.SlotSpecs {
				slot, present := slots[spec.ID]
				if !present || (spec.Optionality == construction.SlotRequired && len(slot.CandidateIDs) == 0) || len(slot.CandidateIDs) < spec.Cardinality.Minimum {
					return allowed, false
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, found := byID[candidateID]
					if !found || candidate.Category != spec.Category {
						return allowed, false
					}
				}
			}
			subjectSlot, subjectPresent := slots["subject"]
			verbSlot, verbPresent := slots["verb"]
			if !subjectPresent || !verbPresent || len(subjectSlot.CandidateIDs) == 0 || len(verbSlot.CandidateIDs) == 0 {
				return allowed, false
			}
			objectSlot, objectPresent := slots["object"]
			if !objectPresent || len(objectSlot.CandidateIDs) == 0 || subjectSlot.Role != subjectRole || objectSlot.Role != objectRole {
				return allowed, false
			}
			for _, subjectID := range subjectSlot.CandidateIDs {
				subject := byID[subjectID]
				if subject.Category != construction.CategoryNP || subject.Features["case"] != features.FeatureValueID(subjectCase) {
					return allowed, false
				}
			}
			for _, objectID := range objectSlot.CandidateIDs {
				object := byID[objectID]
				if object.Category != construction.CategoryNP || object.Features["case"] != features.FeatureValueID(objectCase) {
					return allowed, false
				}
			}
			var matched map[features.FeatureID]features.FeatureValueID
			matchCount := 0
			for _, verbID := range verbSlot.CandidateIDs {
				verb := byID[verbID]
				if verb.Category != construction.CategoryVP || !candidateHasPolishSense(verb, alternative.Permission.SenseID) || verb.Features["verb_form"] != "finite" || verb.Features["tense"] != "past" || verb.Features["person"] != "third" {
					continue
				}
				if !polishPastIntransitiveVerbNumberGenderValid(verb.Features) {
					continue
				}
				agrees := false
				for _, subjectID := range subjectSlot.CandidateIDs {
					subject := byID[subjectID]
					if subject.Category != construction.CategoryNP || subject.Features["case"] != "nominative" || !polishPastIntransitiveAgreementMatches(subject.Features, verb.Features) {
						continue
					}
					agrees = true
					break
				}
				if agrees {
					matchCount++
					matched = make(map[features.FeatureID]features.FeatureValueID, len(verb.Features))
					for feature, value := range verb.Features {
						matched[feature] = value
					}
				}
			}
			if matchCount != 1 || matched == nil {
				return allowed, false
			}
			allowed[alternative.Permission.SenseID] = append(allowed[alternative.Permission.SenseID], matched)
		}
	}
	return allowed, seenClause
}

// polishRelationSourceFeatureAlternatives derives relation capability from
// the exact approved parse routes, then leaves ordinary completeness and
// predicate-feature validation to BareDeclarativeEvidenceBuilder.
func polishRelationSourceFeatureAlternatives(lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry, packages []construction.CompiledConstructionPackage, lexiconIndex *lexicon.Index, ontology *core.ConceptGraph, valencyIndex *valency.Index) map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID {
	if registry == nil || lexiconIndex == nil || ontology == nil || valencyIndex == nil || len(forest.Roots) != 1 || len(forest.Unmatched) != 0 ||
		(forest.Outcome != status.OutcomeOK && forest.Outcome != status.OutcomeAmbiguous) {
		return nil
	}
	for _, token := range lattice.Tokens {
		if token.Token.Folded == "nie" || token.Token.Folded == "nigdy" {
			return nil
		}
	}
	byCandidate := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byCandidate[candidate.ID] = candidate
	}
	frames := make(map[valency.FrameID]valency.ValencyFrame, len(valencyIndex.Frames()))
	for _, frame := range valencyIndex.Frames() {
		frames[frame.ID] = frame
	}
	packagesByFamily := make(map[string]construction.CompiledConstructionPackage, len(packages))
	for _, compiledPackage := range packages {
		packagesByFamily[compiledPackage.Package.Family.ID.String()] = compiledPackage
	}
	result := make(map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID)
	clauseCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		clauseCount++
		if len(node.Alternatives) == 0 {
			return nil
		}
		for _, alternative := range node.Alternatives {
			if len(alternative.Unresolved) != 0 || !alternative.PlanID.Valid() || alternative.Permission.SenseID.String() == "" || alternative.FamilyID.String() == "" {
				return nil
			}
			senseID := alternative.Permission.SenseID
			frameID := valency.FrameID(alternative.Permission.FrameID)
			frame, exists := frames[frameID]
			if !exists || frame.SenseID != senseID || frame.Validate() != nil || len(frame.SemanticRoles) != 2 ||
				len(frame.Arguments) != 2 || len(frame.SurfaceOmittedRoles) != 0 || !polishFrameAllowsFamily(frame, alternative.FamilyID) {
				return nil
			}
			roleSet := make(map[meaning.Role]bool, len(frame.SemanticRoles))
			argumentsByRole := make(map[meaning.Role]valency.ArgumentSlot, len(frame.Arguments))
			for _, role := range frame.SemanticRoles {
				if role == "" || roleSet[role] {
					return nil
				}
				roleSet[role] = true
			}
			for _, argument := range frame.Arguments {
				if !argument.Required || !roleSet[argument.Role] {
					return nil
				}
				if _, duplicate := argumentsByRole[argument.Role]; duplicate {
					return nil
				}
				argumentsByRole[argument.Role] = argument
			}
			if len(argumentsByRole) != len(roleSet) {
				return nil
			}
			familyRecord, exists := registry.ByID(alternative.FamilyID)
			if !exists || familyRecord.Family.Validate() != nil || len(familyRecord.Family.SlotSpecs) != 3 {
				return nil
			}
			boundFamily, err := construction.BindFrame(familyRecord, frame)
			if err != nil {
				return nil
			}
			compiledPackage, exists := packagesByFamily[alternative.FamilyID.String()]
			if !exists || compiledPackage.Package.Family.ID != familyRecord.Family.ID || compiledPackage.Index == nil {
				return nil
			}
			compiledVariant, exists := compiledPackage.Index.ByID(alternative.VariantID)
			if !exists || compiledVariant.FamilyID != alternative.FamilyID || compiledVariant.Variant.DirectionPolicy == construction.DirectionRealize {
				return nil
			}
			if _, err := construction.ValidateVariantForFrame(compiledVariant.Variant, boundFamily.Family, frame); err != nil {
				return nil
			}
			bindingCount := 0
			for _, binding := range registry.BindingsByFamilyID(alternative.FamilyID) {
				if binding.Approved && binding.SenseID == senseID && binding.ValencyFrameID == frameID && binding.FamilyID == alternative.FamilyID && binding.VariantID.String() == alternative.VariantID.String() {
					bindingCount++
				}
			}
			if bindingCount != 1 {
				return nil
			}
			transferBindings := lexiconIndex.TransferBindings(senseID)
			if len(transferBindings) != 1 {
				return nil
			}
			coreConcept, exists := ontology.Lookup(transferBindings[0].CoreRef)
			if !exists || coreConcept.Kind != core.KindRelation || !polishOwnershipRelationRolesMatch(coreConcept, frame.SemanticRoles) {
				return nil
			}
			if len(alternative.Slots) != len(boundFamily.Family.SlotSpecs) {
				return nil
			}
			slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
			for _, slot := range alternative.Slots {
				if _, duplicate := slots[slot.SlotID]; duplicate {
					return nil
				}
				slots[slot.SlotID] = slot
			}
			predicateCount := 0
			var predicate parser.PhraseCandidate
			nominalPredicate := false
			roleSlotCount := make(map[meaning.Role]int, len(frame.SemanticRoles))
			for _, spec := range boundFamily.Family.SlotSpecs {
				slot, exists := slots[spec.ID]
				if !exists || spec.Optionality != construction.SlotRequired || spec.Cardinality.Minimum != 1 || spec.Cardinality.Maximum != 1 ||
					len(slot.CandidateIDs) != 1 || slot.WHGap != nil || len(slot.CandidateIDs[0]) == 0 {
					return nil
				}
				phrase, exists := byCandidate[slot.CandidateIDs[0]]
				if !exists || phrase.Category != spec.Category {
					return nil
				}
				if slot.Role != "" {
					role := slot.Role
					argument, exists := argumentsByRole[role]
					if !exists || spec.Category != construction.CategoryNP || spec.ModifierBinding != nil ||
						(spec.ArgumentBinding != nil && spec.ArgumentBinding.SemanticRole != role) ||
						spec.Case != argument.Case || spec.Adposition != argument.Adposition {
						return nil
					}
					roleSlotCount[role]++
					continue
				}
				if !spec.IsMainPredicate() || spec.ModifierBinding != nil || !polishPhraseHasSense(phrase.SenseIDs, senseID) {
					return nil
				}
				if spec.NominalPredicate {
					lexeme, found := lexiconIndex.ByLexemeID(phrase.LexemeID)
					if spec.Category != construction.CategoryNP || phrase.Category != construction.CategoryNP || !phrase.CaseKnown || phrase.Case != "instrumental" || !found || lexeme.POS != lexicon.PartOfSpeech("noun") {
						return nil
					}
					nominalPredicate = true
				} else if spec.Category != construction.CategoryVP || phrase.Category != construction.CategoryVP {
					return nil
				}
				predicateCount++
				predicate = phrase
			}
			if predicateCount != 1 || len(roleSlotCount) != 2 || (nominalPredicate && predicate.Features["case"] != "instrumental") || (!nominalPredicate && !polishRelationPresentFinite(predicate.Features)) {
				return nil
			}
			for _, role := range frame.SemanticRoles {
				if roleSlotCount[role] != 1 {
					return nil
				}
			}
			if nominalPredicate {
				result[senseID] = []map[features.FeatureID]features.FeatureValueID{{"case": "instrumental", "number": "singular"}}
				continue
			}
			result[senseID] = []map[features.FeatureID]features.FeatureValueID{
				{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "singular"},
				{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third", "number": "plural"},
			}
		}
	}
	if clauseCount != 1 || len(result) == 0 {
		return nil
	}
	return result
}

func polishOwnershipRelationRolesMatch(concept core.Concept, roles []meaning.Role) bool {
	expected := map[meaning.Role]bool{}
	switch concept.ID.String() {
	case "concept:belonging_association":
		expected["associate"], expected["item"] = true, true
	case "concept:legal_ownership":
		expected["owner"], expected["asset"] = true, true
	default:
		return false
	}
	if len(roles) != len(expected) {
		return false
	}
	for _, role := range roles {
		if !expected[role] {
			return false
		}
		delete(expected, role)
	}
	return len(expected) == 0
}

func polishRelationPresentFinite(actual map[features.FeatureID]features.FeatureValueID) bool {
	for key, expected := range map[features.FeatureID]features.FeatureValueID{
		"verb_form": "finite", "tense": "present", "aspect": "imperfective", "mood": "indicative", "voice": "active", "person": "third",
	} {
		if actual[key] != expected {
			return false
		}
	}
	return actual["number"] == "singular" || actual["number"] == "plural"
}

func polishPhraseHasSense(senses []ids.SenseID, wanted ids.SenseID) bool {
	for _, sense := range senses {
		if sense == wanted {
			return true
		}
	}
	return false
}

func clonePolishSupportedSenses(source map[ids.SenseID]bool) map[ids.SenseID]bool {
	cloned := make(map[ids.SenseID]bool, len(source))
	for sense, supported := range source {
		cloned[sense] = supported
	}
	return cloned
}

func polishTransitivePastFrameRoles(family construction.ConstructionFamily, frame valency.ValencyFrame) (meaning.Role, meaning.Role, string, string, bool) {
	specs := make(map[string]construction.SlotSpec, len(family.SlotSpecs))
	for _, spec := range family.SlotSpecs {
		if _, duplicate := specs[spec.ID]; duplicate {
			return "", "", "", "", false
		}
		specs[spec.ID] = spec
	}
	for _, id := range []string{"subject", "verb", "object"} {
		spec, found := specs[id]
		if !found || spec.Optionality != construction.SlotRequired || spec.Cardinality.Minimum != 1 || spec.Cardinality.Maximum != 1 {
			return "", "", "", "", false
		}
		category := construction.CategoryNP
		if id == "verb" {
			category = construction.CategoryVP
		}
		if spec.Category != category {
			return "", "", "", "", false
		}
	}
	for _, spec := range family.SlotSpecs {
		if spec.Optionality == construction.SlotRequired && spec.Cardinality.Minimum > 0 && spec.ID != "subject" && spec.ID != "verb" && spec.ID != "object" {
			return "", "", "", "", false
		}
	}
	var subjectRole, objectRole meaning.Role
	var subjectCase, objectCase string
	for _, argument := range frame.Arguments {
		if !argument.Required {
			continue
		}
		switch argument.ID {
		case "subject":
			if argument.GrammaticalFunction != "subject" || argument.Case != "nominative" || argument.Role == "" || subjectRole != "" {
				return "", "", "", "", false
			}
			subjectRole, subjectCase = argument.Role, argument.Case
		case "object":
			if argument.GrammaticalFunction != "object" || argument.Case == "" || argument.Role == "" || objectRole != "" {
				return "", "", "", "", false
			}
			objectRole, objectCase = argument.Role, argument.Case
		default:
			return "", "", "", "", false
		}
	}
	if subjectRole == "" || objectRole == "" || subjectRole == objectRole ||
		!polishFrameDeclaresRole(frame, subjectRole) || !polishFrameDeclaresRole(frame, objectRole) {
		return "", "", "", "", false
	}
	return subjectRole, objectRole, subjectCase, objectCase, true
}

func polishFrameDeclaresRole(frame valency.ValencyFrame, role meaning.Role) bool {
	for _, declared := range frame.SemanticRoles {
		if declared == role {
			return true
		}
	}
	return false
}

func clonePolishPredicateAlternatives(source map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID) map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID {
	result := make(map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, len(source))
	for sense, alternatives := range source {
		result[sense] = append([]map[features.FeatureID]features.FeatureValueID(nil), alternatives...)
	}
	return result
}

// approvedPolishIntransitiveSenses derives the finite-predicate inventory
// from approved bindings whose frame and family each contain exactly one
// required subject argument and one main VP slot.
func approvedPolishIntransitiveSenses(registry *construction.Registry, valencyIndex *valency.Index, variants map[ids.ConstructionID]*construction.VariantIndex) map[ids.SenseID]bool {
	allowed := make(map[ids.SenseID]bool)
	if registry == nil || valencyIndex == nil {
		return allowed
	}
	familyID, err := ids.ParseConstructionID("construction:pl.intransitive")
	if err != nil {
		return allowed
	}
	record, exists := registry.ByID(familyID)
	variantIndex := variants[familyID]
	if !exists || variantIndex == nil || !polishIntransitiveFamilyShape(record.Family) {
		return allowed
	}
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID || !polishIntransitiveFrameShape(binding.SenseID, binding.ValencyFrameID, familyID, valencyIndex) {
			continue
		}
		if _, exists := variantIndex.ByID(construction.VariantID(binding.VariantID.String())); !exists {
			continue
		}
		allowed[binding.SenseID] = true
	}
	return allowed
}

func polishIntransitiveFamilyShape(family construction.ConstructionFamily) bool {
	if len(family.SlotSpecs) != 2 {
		return false
	}
	subject, verb := false, false
	for _, slot := range family.SlotSpecs {
		if slot.Optionality != construction.SlotRequired || slot.Cardinality.Minimum != 1 || slot.Cardinality.Maximum != 1 {
			return false
		}
		switch {
		case slot.ID == "subject" && slot.Category == construction.CategoryNP:
			subject = true
		case slot.ID == "verb" && slot.Category == construction.CategoryVP:
			verb = true
		default:
			return false
		}
	}
	return subject && verb
}

func polishIntransitiveFrameShape(senseID ids.SenseID, frameID valency.FrameID, familyID ids.ConstructionID, valencyIndex *valency.Index) bool {
	for _, frame := range valencyIndex.FramesBySenseID(senseID) {
		if frame.ID != frameID || frame.SenseID != senseID || !polishFrameAllowsFamily(frame, familyID) || len(frame.SemanticRoles) != 1 || len(frame.Arguments) != 1 {
			continue
		}
		argument := frame.Arguments[0]
		if argument.ID == "subject" && argument.Required && argument.GrammaticalFunction == "subject" && argument.Role == frame.SemanticRoles[0] {
			return true
		}
	}
	return false
}

// polishIntransitiveSimplePastCapability derives the past feature alternatives
// and coverage from retained compiled plans, approved valency bindings, and
// finite past VP features. The bound VP must agree with a nominative subject in
// Polish gender and number before it can attest coverage.
func polishIntransitiveSimplePastCapability(candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry) (map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, bool) {
	allowed := make(map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID)
	if registry == nil {
		return allowed, false
	}
	familyID, err := ids.ParseConstructionID("construction:pl.intransitive")
	if err != nil {
		return allowed, false
	}
	bindings := registry.BindingsByFamilyID(familyID)
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	seenClause := false
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		seenClause = true
		if len(node.Alternatives) == 0 {
			return allowed, false
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != familyID || len(alternative.Unresolved) != 0 {
				return allowed, false
			}
			bound := false
			for _, binding := range bindings {
				if binding.Approved && binding.SenseID == alternative.Permission.SenseID && string(binding.ValencyFrameID) == string(alternative.Permission.FrameID) && binding.FamilyID == alternative.FamilyID && binding.VariantID.String() == alternative.VariantID.String() {
					bound = true
					break
				}
			}
			if !bound {
				return allowed, false
			}
			var subjectIDs, verbIDs []string
			for _, slot := range alternative.Slots {
				switch slot.SlotID {
				case "subject":
					subjectIDs = slot.CandidateIDs
				case "verb":
					verbIDs = slot.CandidateIDs
				}
			}
			if len(subjectIDs) == 0 || len(verbIDs) == 0 {
				return allowed, false
			}
			matchedFeatures := map[features.FeatureID]features.FeatureValueID(nil)
			verbMatches := 0
			for _, id := range verbIDs {
				verb, exists := byID[id]
				if !exists || verb.Category != construction.CategoryVP || !candidateHasPolishSense(verb, alternative.Permission.SenseID) || verb.Features["verb_form"] != "finite" || verb.Features["tense"] != "past" || verb.Features["aspect"] != "imperfective" || verb.Features["person"] != "third" {
					continue
				}
				if !polishPastIntransitiveVerbNumberGenderValid(verb.Features) {
					continue
				}
				agreement := false
				for _, subjectID := range subjectIDs {
					subject, found := byID[subjectID]
					if found && subject.Category == construction.CategoryNP && subject.Features["case"] == "nominative" && polishPastIntransitiveAgreementMatches(subject.Features, verb.Features) {
						agreement = true
						break
					}
				}
				if agreement {
					verbMatches++
					matchedFeatures = map[features.FeatureID]features.FeatureValueID{
						"verb_form": "finite", "tense": "past", "aspect": "imperfective",
						"person": verb.Features["person"], "number": verb.Features["number"], "gender": verb.Features["gender"],
					}
				}
			}
			if verbMatches != 1 || matchedFeatures == nil {
				return allowed, false
			}
			allowed[alternative.Permission.SenseID] = append(allowed[alternative.Permission.SenseID], matchedFeatures)
		}
	}
	return allowed, seenClause
}

func candidateHasPolishSense(candidate parser.PhraseCandidate, sense ids.SenseID) bool {
	for _, candidateSense := range candidate.SenseIDs {
		if candidateSense == sense {
			return true
		}
	}
	return false
}

func polishPastIntransitiveVerbNumberGenderValid(verb map[features.FeatureID]features.FeatureValueID) bool {
	number, gender := verb["number"], verb["gender"]
	if number == "singular" {
		return gender == "masculine_personal" || gender == "feminine" || gender == "neuter"
	}
	return number == "plural" && (gender == "masculine_personal" || gender == "non_masculine_personal")
}

func polishPastIntransitiveAgreementMatches(subject, verb map[features.FeatureID]features.FeatureValueID) bool {
	subjectNumber, subjectGender := subject["number"], subject["gender"]
	verbNumber, verbGender := verb["number"], verb["gender"]
	if subjectNumber == "" || subjectGender == "" || verbNumber != subjectNumber || verbGender == "" {
		return false
	}
	expectedGender := subjectGender
	if subjectNumber == "singular" && (subjectGender == "masculine_animal" || subjectGender == "masculine_inanimate") {
		expectedGender = "masculine_personal"
	} else if subjectNumber == "plural" && subjectGender != "masculine_personal" {
		expectedGender = "non_masculine_personal"
	}
	return verbGender == expectedGender
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
