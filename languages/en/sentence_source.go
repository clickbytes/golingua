package en

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
	return NewSentenceSourceWithSeed(snapshot, seed)
}

// NewSentenceSourceWithSeed builds source analysis using the caller's exact
// compiled morphology seed.
func NewSentenceSourceWithSeed(snapshot *pack.Pack, seed *Seed) (*runtime.SourcePipeline, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("English sentence source requires a snapshot")
	}
	if seed == nil {
		return nil, fmt.Errorf("English sentence source requires an English morphology seed")
	}
	profile, inverse, _, _ := seed.RuntimeMorphology()
	if profile == nil || profile.Language() != "en" || inverse == nil {
		return nil, fmt.Errorf("English sentence source requires an English morphology seed")
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
	plans, variants, err := runtime.CompileSentencePlans(profile, packages, valencyIndex)
	if err != nil {
		return nil, err
	}
	predicateSenses := approvedEnglishPredicateSenses(valencyIndex, packages)
	scopedPlan, err := SentenceScopedClauses()
	if err != nil {
		return nil, err
	}
	relationPathPlan, err := SentenceRelationPathPlan()
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
	helpSense, err := ids.NewSenseID("en.helping.event")
	if err != nil {
		return nil, err
	}
	thankSense, err := ids.NewSenseID("en.thanking.event")
	if err != nil {
		return nil, err
	}
	usingSense, err := ids.NewSenseID("en.using_object.event")
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
	redSense, err := ids.NewSenseID("en.red.color")
	if err != nil {
		return nil, err
	}
	attributionSenses := map[ids.SenseID]bool{attributionSense: true, redSense: true}
	for _, spec := range englishAttributionSpecs {
		sense, senseErr := ids.NewSenseID(strings.TrimPrefix(spec.sense, "sense:"))
		if senseErr != nil {
			return nil, senseErr
		}
		attributionSenses[sense] = true
	}
	eatingSense, err := ids.NewSenseID("en.eating.event")
	if err != nil {
		return nil, err
	}
	drinkingSense, err := ids.NewSenseID("en.drinking.event")
	if err != nil {
		return nil, err
	}
	buyingSense, err := ids.NewSenseID("en.buying.event")
	if err != nil {
		return nil, err
	}
	sellingSense, err := ids.NewSenseID("en.selling.event")
	if err != nil {
		return nil, err
	}
	givingSense, err := ids.NewSenseID("en.give.transfer")
	if err != nil {
		return nil, err
	}
	manipulationSenses := map[ids.SenseID]bool{}
	for _, name := range []string{"en.opening_object.event", "en.closing_object.event", "en.carrying.event", "en.holding_object.event", "en.pulling_object.event", "en.pushing_object.event", "en.cutting_object.event", "en.cutting_object.with_instrument", "en.folding_object.event", "en.tearing_object.event"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		manipulationSenses[sense] = true
	}
	supportedSenses := map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true, readingSense: true, classificationSense: true, attributionSense: true, redSense: true, eatingSense: true, drinkingSense: true, givingSense: true, buyingSense: true, sellingSense: true, helpSense: true, thankSense: true, usingSense: true}
	for sense := range manipulationSenses {
		supportedSenses[sense] = true
	}
	for sense := range attributionSenses {
		supportedSenses[sense] = true
	}
	affectiveSenses := map[ids.SenseID]bool{}
	for _, name := range []string{"en.liking.event", "en.watching.event", "en.writing_document.event"} {
		sense, senseErr := ids.NewSenseID(name)
		if senseErr != nil {
			return nil, senseErr
		}
		affectiveSenses[sense] = true
		supportedSenses[sense] = true
	}
	writingMessageSense, err := ids.NewSenseID("en.writing_message.event")
	if err != nil {
		return nil, err
	}
	supportedSenses[writingMessageSense] = true
	writingInstrumentSense, err := ids.NewSenseID("en.writing_with_instrument.event")
	if err != nil {
		return nil, err
	}
	supportedSenses[writingInstrumentSense] = true
	conversingSense, err := ids.NewSenseID("en.conversing.event")
	if err != nil {
		return nil, err
	}
	supportedSenses[conversingSense] = true
	intransitiveSenses := approvedEnglishIntransitiveSenses(registry, valencyIndex, variants)
	for sense := range intransitiveSenses {
		supportedSenses[sense] = true
	}
	predicateAlternatives := map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID{
		givingSense: {{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"}, {"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"}},
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
			{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"},
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
		receivingSense: {{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"}},
		classificationSense: {
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "number": "singular", "verb_number": "singular"},
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "number": "plural", "verb_number": "plural"},
		},
		attributionSense: {{"adjective_degree": "positive"}},
		redSense:         {{"adjective_degree": "positive"}},
		eatingSense:      {{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"}, {"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"}},
		drinkingSense: {
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
			{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"},
		},
		// The new manipulation routes are restricted to third-person present.
		// They are filled below through a precomputed list.
	}
	for sense := range attributionSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{{"adjective_degree": "positive"}}
	}
	commerceFeatures := []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
		{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
	}
	predicateAlternatives[buyingSense] = commerceFeatures
	predicateAlternatives[sellingSense] = commerceFeatures
	predicateAlternatives[helpSense] = commerceFeatures
	predicateAlternatives[thankSense] = commerceFeatures
	predicateAlternatives[usingSense] = commerceFeatures
	for sense := range manipulationSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
		}
	}
	for sense := range affectiveSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
		}
	}
	predicateAlternatives[writingMessageSense] = []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
		{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
	}
	predicateAlternatives[writingInstrumentSense] = predicateAlternatives[writingMessageSense][:2]
	predicateAlternatives[conversingSense] = []map[features.FeatureID]features.FeatureValueID{
		{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
		{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
	}
	for sense := range intransitiveSenses {
		predicateAlternatives[sense] = []map[features.FeatureID]features.FeatureValueID{
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
			{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
		}
	}
	predicateAlternatives[eatingSense] = append(predicateAlternatives[eatingSense], map[features.FeatureID]features.FeatureValueID{"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active"})
	addApprovedEnglishTransitiveSourceRoutes(supportedSenses, predicateAlternatives, snapshot.Lexicon(), valencyIndex, registry, packages)
	resolver := englishNegationResolver{registry: registry, valency: valencyIndex, lexicon: snapshot.Lexicon(), variants: variants}
	return &runtime.SourcePipeline{
		Language: "en", PhraseProfile: phraseProfile, ClauseConnector: "and", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}}, AllowFinalPeriod: false,
		ScopedSourcePlan: scopedPlan, ScopedReadingPolicy: ScopedSourceReadingPolicy,
		RelationPathPlan: relationPathPlan, RelationSourceReadingPolicy: RelationSourceReadingPolicy,
		AllowedUnlexicalizedTerminals: map[string]bool{"at": true, "to": true, "from": true, "with": true, "do": true, "does": true, "did": true, "not": true, "this": true, "that": true, "these": true, "those": true, "it": true, "is": true, "are": true, "a": true, "an": true, "who": true, "whom": true, "and": true, "up": true, "asleep": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: englishWHGapPhraseBuilder{inner: englishIdentificationPhraseBuilder{phraseProfile: phraseProfile, featureProfile: profile, inner: englishDeterminerNPBuilder{resolver: resolver, predicateSenses: predicateSenses, inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "pronoun": {construction.CategoryNP}, "verb": {construction.CategoryVP}, "adjective": {construction.CategoryAP},
			},
		}}, lexicon: snapshot.Lexicon()}, lexicon: snapshot.Lexicon()}},
		EvidenceBuilder: englishSentenceEvidenceBuilder{lexicon: snapshot.Lexicon(), core: snapshot.CoreGraph(), constructions: registry, packages: packages, resolver: resolver, base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses:                     supportedSenses,
			PredicateFeatureAlternativesBySense: predicateAlternatives,
			AllowedUnlexicalizedTerminals:       map[string]bool{"at": true, "to": true, "from": true, "with": true, "do": true, "does": true, "did": true, "not": true, "this": true, "that": true, "these": true, "those": true, "it": true, "is": true, "are": true, "a": true, "an": true, "up": true, "asleep": true},
		}},
	}, nil
}

func addApprovedEnglishTransitiveSourceRoutes(supported map[ids.SenseID]bool, alternatives map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID, lexiconIndex *lexicon.Index, valencyIndex *valency.Index, registry *construction.Registry, packages []construction.CompiledConstructionPackage) {
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
	for _, familyName := range []string{"construction:en.transitive", "construction:en.transitive_personal", "construction:en.transitive_negative"} {
		familyID, err := ids.ParseConstructionID(familyName)
		if err != nil {
			continue
		}
		if _, found := registry.ByID(familyID); !found {
			continue
		}
		for _, binding := range registry.BindingsByFamilyID(familyID) {
			frame, frameOK := frames[string(binding.ValencyFrameID)]
			_, variantOK := variants[familyName+"\x00"+binding.VariantID.String()]
			sense, senseOK := lexiconIndex.BySenseID(binding.SenseID)
			if !binding.Approved || binding.FamilyID != familyID || !frameOK || !variantOK || !senseOK || frame.SenseID != binding.SenseID || !valencyFrameAllowsFamily(frame, familyID) {
				continue
			}
			lexeme, lexemeOK := lexiconIndex.ByLexemeID(sense.LexemeID)
			semanticBindings := lexiconIndex.BindingsForSense(binding.SenseID)
			if !lexemeOK || lexeme.Language != "en" || lexeme.POS != "verb" || len(semanticBindings) != 1 || !semanticBindings[0].TransferAuthorized() {
				continue
			}
			roles := make(map[meaning.Role]bool)
			for _, role := range frame.SemanticRoles {
				roles[role] = true
			}
			if len(roles) != 2 || len(frame.Arguments) != 2 {
				continue
			}
			required := map[features.FeatureID]features.FeatureValueID{"tense": "present", "verb_form": "finite"}
			if familyName == "construction:en.transitive_negative" {
				required["verb_form"] = "base"
			}
			supported[binding.SenseID] = true
			alternatives[binding.SenseID] = append(alternatives[binding.SenseID], required)
		}
	}
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
	base          runtime.BareDeclarativeEvidenceBuilder
	lexicon       *lexicon.Index
	core          *core.ConceptGraph
	constructions *construction.Registry
	packages      []construction.CompiledConstructionPackage
	resolver      englishNegationResolver
}

func (builder englishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	if routes := englishRelationSourceFeatureAlternatives(lattice, candidates, forest, builder.constructions, builder.packages, builder.lexicon, builder.core, builder.resolver.valency); len(routes) > 0 {
		builder.base.SupportedSenses = cloneEnglishSupportedSenses(builder.base.SupportedSenses)
		builder.base.PredicateFeatureAlternativesBySense = cloneEnglishPredicateAlternatives(builder.base.PredicateFeatureAlternativesBySense)
		for sense, alternatives := range routes {
			builder.base.SupportedSenses[sense] = true
			builder.base.PredicateFeatureAlternativesBySense[sense] = alternatives
		}
	}
	evidence, err := builder.buildSyntax(ctx, lattice, candidates, forest)
	if err != nil {
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
		forest = filterEnglishGivingArticleAlternatives(forest, lattice, candidates, builder.lexicon, builder.constructions)
		pastIntransitiveSenses, pastIntransitiveCoverage := englishIntransitiveSimplePastCapability(candidates, forest, builder.constructions)
		if !pastIntransitiveCoverage {
			pastIntransitiveSenses = nil
		}
		pastTransitiveSenses := englishTransitivePastCapability(candidates, forest, builder.constructions, builder.resolver.valency)
		commercePastSenses, commercePastCoverage := englishCommerceTransferPastCapability(candidates, forest, builder.constructions, builder.resolver.valency)
		for sense := range commercePastSenses {
			pastTransitiveSenses[sense] = true
		}
		base := builder.base
		if len(pastTransitiveSenses) > 0 {
			base.PredicateFeatureAlternativesBySense = cloneEnglishPredicateAlternatives(base.PredicateFeatureAlternativesBySense)
			for sense := range pastTransitiveSenses {
				base.PredicateFeatureAlternativesBySense[sense] = appendPastFiniteAlternative(base.PredicateFeatureAlternativesBySense[sense])
			}
		}
		if len(pastIntransitiveSenses) > 0 {
			base.PredicateFeatureAlternativesBySense = cloneEnglishPredicateAlternatives(base.PredicateFeatureAlternativesBySense)
			for sense := range pastIntransitiveSenses {
				base.PredicateFeatureAlternativesBySense[sense] = append(base.PredicateFeatureAlternativesBySense[sense], map[features.FeatureID]features.FeatureValueID{
					"verb_form": "finite", "tense": "past",
				})
			}
		}
		evidence, err := base.Build(ctx, lattice, candidates, forest)
		if err != nil {
			return evidence, err
		}
		evidence = addEnglishPersonalReferences(evidence, candidates, forest)
		if !evidence.OperatorCoverageComplete {
			if !englishWHQuestionCoverage(lattice, candidates, forest) && !pastIntransitiveCoverage &&
				!englishTransitivePastCoverage(candidates, forest, builder.constructions, builder.resolver.valency) && !commercePastCoverage {
				return evidence, nil
			}
			evidence.OperatorCoverageComplete = true
		}
		return addEnglishIdentificationReferences(evidence, candidates, forest, lattice), nil
	}
	if _, uniqueNot := englishTerminalSpan(lattice, "not"); !uniqueNot {
		return runtime.CompositionEvidence{}, nil
	}
	if englishNegativeClassificationInput(lattice) {
		return builder.buildEnglishNegativeClassification(ctx, lattice, candidates, forest)
	}
	if number, tense, notSpan, supported := englishAttributionNegationInput(lattice); supported {
		return builder.buildEnglishNegativeAttribution(ctx, lattice, candidates, forest, number, tense, notSpan)
	}
	if evidence, handled, err := builder.buildEnglishAttributedNegativeSeeing(ctx, lattice, candidates, forest); handled || err != nil {
		return evidence, err
	}
	route, person, number, tense, notSpan, supported := builder.resolver.negationInput(lattice, candidates)
	if !supported {
		return runtime.CompositionEvidence{}, nil
	}
	base := builder.base
	base.PredicateFeatureAlternativesBySense = cloneEnglishPredicateAlternatives(builder.base.PredicateFeatureAlternativesBySense)
	for _, negativeRoute := range builder.resolver.routes() {
		if tense == "past" && !englishStaticPastNegationRoute(negativeRoute) {
			continue
		}
		senseID, err := ids.NewSenseID(strings.TrimPrefix(negativeRoute.sense, "sense:"))
		if err != nil {
			return runtime.CompositionEvidence{}, err
		}
		personForRoute := "third"
		verbNumbers := []string{"singular", "plural"}
		if negativeRoute.sense == route.sense && tense == "present" {
			if person != "" {
				personForRoute = person
			}
			if number != "" {
				verbNumbers = []string{number}
			}
		}
		for _, verbNumber := range verbNumbers {
			base.PredicateFeatureAlternativesBySense[senseID] = append(
				base.PredicateFeatureAlternativesBySense[senseID],
				map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "tense": features.FeatureValueID(tense), "person": features.FeatureValueID(personForRoute),
					"number": features.FeatureValueID(verbNumber), "verb_number": features.FeatureValueID(verbNumber),
				},
			)
		}
	}
	evidence, err := base.Build(ctx, lattice, candidates, forest)
	if err != nil {
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
	seenOperatorEvidence := make(map[string]struct{})
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			attributedNegative := false
			if route.concept == "concept:seeing" && alternative.FamilyID.String() == "construction:en.attributed_seeing_object_negative" && number == "singular" {
				switch alternative.VariantID.String() {
				case "construction:en.attributed_seeing_object_negative.svo",
					"construction:en.attributed_seeing_object_negative.svo.indefinite_subject",
					"construction:en.attributed_seeing_object_negative.svo.indefinite_object",
					"construction:en.attributed_seeing_object_negative.svo.indefinite_both":
					attributedNegative = true
				}
			}
			variantID := alternative.VariantID.String()
			variantAllowed := route.variantMatchesNumber(variantID, number)
			if tense == "past" {
				variantAllowed = route.pastVariantAllowed(variantID)
			}
			sourceOnlyVariant := !attributedNegative && !variantAllowed &&
				englishApprovedSourceOnlyNegativeVariant(builder.resolver, alternative, route, person, number, tense, candidates)
			if (alternative.FamilyID != negativeFamily && !attributedNegative) || (!attributedNegative && !variantAllowed && !sourceOnlyVariant) || alternative.Permission.SenseID.String() != route.sense {
				return runtime.CompositionEvidence{}, nil
			}
			if route.family == "construction:en.transitive_negative" || route.family == "construction:en.ditransitive_negative" {
				familyRecord, familyExists := builder.resolver.registry.ByID(alternative.FamilyID)
				approvedBinding := false
				for _, binding := range builder.resolver.registry.BindingsByFamilyID(alternative.FamilyID) {
					if binding.Approved && binding.SenseID.String() == route.sense && string(binding.ValencyFrameID) == route.frame && binding.VariantID.String() == variantID {
						approvedBinding = true
						break
					}
				}
				if !familyExists || !approvedBinding || string(alternative.Permission.FrameID) != route.frame ||
					!englishAlternativeHasRequiredSlots(alternative, familyRecord.Family, byID) {
					return runtime.CompositionEvidence{}, nil
				}
				if route.family == "construction:en.ditransitive_negative" && !englishGivingNegativeFrameCoverage(alternative, familyRecord.Family, builder.resolver.valency, byID) {
					return runtime.CompositionEvidence{}, nil
				}
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					candidateRoute, candidatePerson, candidateNumber, candidateTense, negated := builder.resolver.negatedPredicateContext(lattice, candidate, candidates)
					if !exists || !negated || candidateRoute.sense != route.sense || candidateTense != tense ||
						(tense == "present" && (number != "" && candidateNumber != number || person != "" && candidatePerson != person)) ||
						candidate.Features["verb_form"] != "base" || string(candidate.Features["number"]) != candidateNumber {
						continue
					}
					operatorEvidence := semantic.OperatorEvidence{
						Operator: construction.Operator{
							ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
							InputFamilyID: alternative.FamilyID, OutputFamilyID: alternative.FamilyID,
							SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
							Capabilities:  []construction.Capability{"negation-negated"}, Provenance: englishNegationConstructionSource,
						},
						TargetCandidateID: candidateID, Span: notSpan,
					}
					key := fmt.Sprintf("%#v|%s|%#v", operatorEvidence.Operator, operatorEvidence.TargetCandidateID, operatorEvidence.Span)
					if _, duplicate := seenOperatorEvidence[key]; duplicate {
						continue
					}
					seenOperatorEvidence[key] = struct{}{}
					evidence.Operators = append(evidence.Operators, operatorEvidence)
					operatorCount++
				}
			}
		}
	}
	evidence.OperatorCoverageComplete = operatorCount > 0
	evidence = addEnglishPersonalReferences(evidence, candidates, forest)
	return evidence, nil
}

// englishApprovedSourceOnlyNegativeVariant accepts an analyze-only compiled
// variant as evidence for a source operator without making it a target route.
// Every structural and feature check is derived from its exact approved
// binding, family plan, and retained parse alternative.
func englishApprovedSourceOnlyNegativeVariant(
	resolver englishNegationResolver,
	alternative parser.ParseAlternative,
	route englishPresentNegationRoute,
	person, number, tense string,
	candidates []parser.PhraseCandidate,
) bool {
	if resolver.registry == nil || resolver.valency == nil || resolver.variants == nil ||
		route.family != "construction:en.transitive_negative" || alternative.FamilyID.String() != route.family ||
		alternative.Permission.SenseID.String() != route.sense || string(alternative.Permission.FrameID) != route.frame {
		return false
	}
	familyRecord, familyExists := resolver.registry.ByID(alternative.FamilyID)
	if !familyExists || string(familyRecord.Family.Language) != "en" ||
		!familyHasOperator(familyRecord.Family, construction.OperatorID("negate")) ||
		!familyHasCapability(familyRecord.Family, construction.Capability("negation-negated")) {
		return false
	}
	if !englishAlternativeUnresolvedSlotsAreBound(alternative, familyRecord.Family) {
		return false
	}
	variantIndex := resolver.variants[alternative.FamilyID]
	if variantIndex == nil {
		return false
	}
	compiled, exists := variantIndex.ByID(alternative.VariantID)
	if !exists || compiled.Variant.DirectionPolicy != construction.DirectionAnalyze ||
		!variantHasRequiredSlots(familyRecord.Family, compiled.Variant) ||
		!sequenceHasTerminal(compiled.Variant.Sequence, "not") {
		return false
	}
	var predicate *construction.SlotSpec
	for index := range familyRecord.Family.SlotSpecs {
		slot := &familyRecord.Family.SlotSpecs[index]
		if slot.Category == construction.CategoryVP && slot.IsMainPredicate() {
			if predicate != nil {
				return false
			}
			predicate = slot
		}
	}
	if predicate == nil {
		return false
	}
	var frame valency.ValencyFrame
	frameFound := false
	for _, candidate := range resolver.valency.Frames() {
		if string(candidate.ID) == route.frame {
			frame, frameFound = candidate, true
			break
		}
	}
	if !frameFound || frame.SenseID.String() != route.sense || !valencyFrameAllowsFamily(frame, alternative.FamilyID) {
		return false
	}
	subjectSlot, _, subjectFound := approvedFamilySubjectSlot(familyRecord.Family, frame)
	if !subjectFound || !sequenceHasSlot(compiled.Variant.Sequence, subjectSlot.ID) ||
		!sequenceHasSlot(compiled.Variant.Sequence, predicate.ID) {
		return false
	}
	byID := phraseCandidatesByID(candidates)
	subjectCandidateCount := 0
	for _, slot := range alternative.Slots {
		if slot.SlotID != subjectSlot.ID || len(slot.CandidateIDs) == 0 {
			continue
		}
		for _, candidateID := range slot.CandidateIDs {
			candidate, found := byID[candidateID]
			if !found {
				return false
			}
			candidatePerson, candidateNumber := englishCandidateAgreement(candidate)
			if person == "" {
				person = candidatePerson
			} else if person != candidatePerson {
				return false
			}
			if number == "" {
				number = candidateNumber
			} else if candidateNumber != "" && number != candidateNumber {
				return false
			}
			subjectCandidateCount++
		}
	}
	if subjectCandidateCount == 0 || person == "" {
		return false
	}
	auxiliary := sequenceAuxiliary(compiled.Variant.Sequence)
	expectedAuxiliary := ""
	switch {
	case tense == "past":
		expectedAuxiliary = "did"
	case tense == "present" && englishDoAuxiliaryMatches("do", person, number):
		expectedAuxiliary = "do"
	case tense == "present" && englishDoAuxiliaryMatches("does", person, number):
		expectedAuxiliary = "does"
	default:
		return false
	}
	if auxiliary != expectedAuxiliary ||
		!englishVariantHasSubjectPersonCondition(compiled.Variant, subjectSlot.ID, person) ||
		!englishVariantHasSubjectNumberCondition(compiled.Variant, subjectSlot.ID, number) ||
		!variantHasFeatureCondition(compiled.Variant, predicate.ID, "tense", tense) {
		return false
	}
	if !englishAlternativeHasRequiredSlots(alternative, familyRecord.Family, phraseCandidatesByID(candidates)) {
		return false
	}
	approvedBinding := false
	for _, binding := range resolver.registry.BindingsByFamilyID(alternative.FamilyID) {
		if binding.Approved && binding.FamilyID == alternative.FamilyID && binding.SenseID.String() == route.sense &&
			string(binding.ValencyFrameID) == route.frame && binding.VariantID.String() == alternative.VariantID.String() {
			approvedBinding = true
			break
		}
	}
	if !approvedBinding {
		return false
	}
	for _, slot := range alternative.Slots {
		if slot.SlotID != subjectSlot.ID {
			continue
		}
		for _, candidateID := range slot.CandidateIDs {
			candidate, found := byID[candidateID]
			candidatePerson, candidateNumber := englishCandidateAgreement(candidate)
			if !found || candidatePerson != person || number != "" && candidateNumber != number {
				return false
			}
		}
	}
	return true
}

// englishAlternativeUnresolvedSlotsAreBound permits unresolved feature or
// agreement evidence only when each affected slot still has a concrete phrase
// candidate. It does not permit missing arguments or unresolved conditions.
func englishAlternativeUnresolvedSlotsAreBound(alternative parser.ParseAlternative, family construction.ConstructionFamily) bool {
	if len(alternative.Unresolved) == 0 {
		return true
	}
	slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
	for _, slot := range alternative.Slots {
		slots[slot.SlotID] = slot
	}
	specs := make(map[string]bool, len(family.SlotSpecs))
	for _, spec := range family.SlotSpecs {
		specs[spec.ID] = true
	}
	for _, unresolved := range alternative.Unresolved {
		slot, exists := slots[unresolved]
		if !specs[unresolved] || !exists || len(slot.CandidateIDs) == 0 {
			return false
		}
	}
	return true
}

func englishCandidateAgreement(candidate parser.PhraseCandidate) (string, string) {
	person := string(candidate.Features["person"])
	if person == "" {
		person = string(candidate.Features["pronoun_person"])
	}
	if person == "" {
		person = "third"
	}
	number := string(candidate.Features["number"])
	if number == "" {
		number = string(candidate.Features["noun_number"])
	}
	if number == "" {
		number = string(candidate.Features["pronoun_number"])
	}
	return person, number
}

func englishVariantHasSubjectNumberCondition(variant construction.Variant, slot, number string) bool {
	found := false
	for _, condition := range variant.Conditions {
		if condition.Kind != construction.ConditionFeatureEquals || condition.SlotID != slot {
			continue
		}
		if condition.Feature == "noun_number" || condition.Feature == "pronoun_number" || condition.Feature == "number" {
			found = true
			if number == "" || string(condition.Value) != number {
				return false
			}
		}
	}
	return !found || number != ""
}

func englishVariantHasSubjectPersonCondition(variant construction.Variant, slot, person string) bool {
	found := false
	for _, condition := range variant.Conditions {
		if condition.Kind != construction.ConditionFeatureEquals || condition.SlotID != slot ||
			condition.Feature != "person" && condition.Feature != "pronoun_person" {
			continue
		}
		found = true
		if person == "" || string(condition.Value) != person {
			return false
		}
	}
	return !found || person != ""
}

func variantHasFeatureCondition(variant construction.Variant, slot, feature, value string) bool {
	for _, condition := range variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == slot &&
			string(condition.Feature) == feature && string(condition.Value) == value {
			return true
		}
	}
	return false
}

func phraseCandidatesByID(candidates []parser.PhraseCandidate) map[string]parser.PhraseCandidate {
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	return byID
}

// buildEnglishAttributedNegativeSeeing binds the dedicated object-attribute
// negative family directly from its approved compiled alternative. This keeps
// its frame distinct from ordinary transitive seeing while both use the lemma
// "see".
func (builder englishSentenceEvidenceBuilder) buildEnglishAttributedNegativeSeeing(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, bool, error) {
	const familyText = "construction:en.attributed_seeing_object_negative"
	const senseText = "sense:en.see.visual"
	const frameText = "en.see.visual.attributed_negative"
	notSpan, hasNot := englishTerminalSpan(lattice, "not")
	if !hasNot || builder.resolver.registry == nil || builder.resolver.valency == nil || builder.resolver.lexicon == nil || builder.resolver.variants == nil {
		return runtime.CompositionEvidence{}, false, nil
	}
	familyID, err := ids.NewConstructionID(strings.TrimPrefix(familyText, "construction:"))
	if err != nil {
		return runtime.CompositionEvidence{}, false, err
	}
	senseID, err := ids.NewSenseID(strings.TrimPrefix(senseText, "sense:"))
	if err != nil {
		return runtime.CompositionEvidence{}, false, err
	}
	familyRecord, exists := builder.resolver.registry.ByID(familyID)
	variantIndex := builder.resolver.variants[familyID]
	if !exists || variantIndex == nil {
		return runtime.CompositionEvidence{}, false, nil
	}
	frameFound := false
	for _, frame := range builder.resolver.valency.FramesBySenseID(senseID) {
		if string(frame.ID) == frameText && frame.SenseID == senseID && valencyFrameAllowsFamily(frame, familyID) {
			frameFound = true
			break
		}
	}
	if !frameFound {
		return runtime.CompositionEvidence{}, false, nil
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	approved := make(map[string]bool)
	for _, binding := range builder.resolver.registry.BindingsByFamilyID(familyID) {
		if binding.Approved && binding.SenseID == senseID && string(binding.ValencyFrameID) == frameText {
			approved[binding.VariantID.String()] = true
		}
	}
	if len(approved) == 0 {
		return runtime.CompositionEvidence{}, false, nil
	}
	hasAttributedAlternative := false
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID == familyID && alternative.Permission.SenseID == senseID && string(alternative.Permission.FrameID) == frameText {
				hasAttributedAlternative = true
				break
			}
		}
	}
	if !hasAttributedAlternative {
		return runtime.CompositionEvidence{}, false, nil
	}
	base := builder.base
	evidence, err := base.Build(ctx, lattice, candidates, forest)
	if err != nil {
		return evidence, true, err
	}
	operatorCount := 0
	seen := make(map[string]bool)
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			variantID := alternative.VariantID.String()
			if alternative.FamilyID != familyID || alternative.Permission.SenseID != senseID || string(alternative.Permission.FrameID) != frameText ||
				!approved[variantID] || len(alternative.Unresolved) != 0 {
				continue
			}
			parsedID, parseErr := construction.ParseVariantID(variantID)
			if parseErr != nil {
				continue
			}
			compiled, exists := variantIndex.ByID(parsedID)
			if !exists || compiled.Variant.DirectionPolicy == construction.DirectionAnalyze || !variantHasRequiredSlots(familyRecord.Family, compiled.Variant) ||
				(!sequenceHasTerminal(compiled.Variant.Sequence, "does") && !sequenceHasTerminal(compiled.Variant.Sequence, "did")) || !sequenceHasTerminal(compiled.Variant.Sequence, "not") {
				continue
			}
			conditionNumber, numberConsistent := variantSubjectNumberIfConsistent(compiled.Variant, "subject")
			if !numberConsistent || conditionNumber != "" && conditionNumber != "singular" || !variantSubjectPersonThird(compiled.Variant, "subject") {
				continue
			}
			if !englishAlternativeHasRequiredSlots(alternative, familyRecord.Family, byID) {
				continue
			}
			var predicateCandidate *parser.PhraseCandidate
			for _, slot := range alternative.Slots {
				if slot.SlotID != "verb" || slot.Role != "" || len(slot.CandidateIDs) != 1 {
					continue
				}
				candidate := byID[slot.CandidateIDs[0]]
				if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == "lexeme:en.see.v" &&
					containsEnglishSenseName(candidate.SenseIDs, strings.TrimPrefix(senseText, "sense:")) && candidate.Features["verb_form"] == "base" &&
					candidate.Features["person"] == "third" && candidate.Features["number"] == "singular" {
					copy := candidate
					predicateCandidate = &copy
				}
			}
			if predicateCandidate == nil {
				continue
			}
			key := predicateCandidate.ID + "|" + fmt.Sprint(notSpan.Start) + "|" + fmt.Sprint(notSpan.End)
			if seen[key] {
				continue
			}
			seen[key] = true
			evidence.Operators = append(evidence.Operators, semantic.OperatorEvidence{
				Operator: construction.Operator{ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
					InputFamilyID: familyID, OutputFamilyID: familyID, SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
					Capabilities: []construction.Capability{"negation-negated"}, Provenance: englishNegationConstructionSource},
				TargetCandidateID: predicateCandidate.ID, Span: notSpan,
			})
			operatorCount++
		}
	}
	evidence.OperatorCoverageComplete = operatorCount > 0
	return evidence, true, nil
}

func englishAttributionNegationInput(lattice analyzer.LexicalLattice) (string, string, text.Span, bool) {
	notSpan, found := englishTerminalSpan(lattice, "not")
	if !found {
		return "", "", text.Span{}, false
	}
	for index, token := range lattice.Tokens {
		if token.Token.Span != notSpan || index == 0 {
			continue
		}
		number, tense := "", "present"
		switch lattice.Tokens[index-1].Token.Folded {
		case "is":
			number = "singular"
		case "are":
			number = "plural"
		case "was":
			number, tense = "singular", "past"
		case "were":
			number, tense = "plural", "past"
		default:
			return "", "", text.Span{}, false
		}
		if index+1 >= len(lattice.Tokens) {
			return "", "", text.Span{}, false
		}
		return number, tense, notSpan, true
	}
	return "", "", text.Span{}, false
}

func (builder englishSentenceEvidenceBuilder) buildEnglishNegativeAttribution(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest, number, tense string, notSpan text.Span) (runtime.CompositionEvidence, error) {
	evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
	if err != nil {
		return evidence, err
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	operatorCount := 0
	seen := make(map[string]bool)
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID.String() != "construction:en.attribution_negative" || alternative.Permission.SenseID.String() == "" {
				continue
			}
			route, ok := builder.resolver.routeForSense(alternative.Permission.SenseID)
			variantPrefix := "construction:en.attribution_negative." + number + "." + tense + "."
			if !ok || !route.statePredicate || !strings.HasPrefix(alternative.VariantID.String(), variantPrefix) {
				continue
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" || slot.SlotID != "predicate" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || candidate.Category != construction.CategoryAP || candidate.LexemeID.String() != route.lexeme || candidate.Features["adjective_degree"] != "positive" {
						continue
					}
					key := candidateID + "|" + fmt.Sprint(notSpan.Start) + "|" + fmt.Sprint(notSpan.End)
					if seen[key] {
						continue
					}
					seen[key] = true
					evidence.Operators = append(evidence.Operators, semantic.OperatorEvidence{
						Operator: construction.Operator{ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
							InputFamilyID: alternative.FamilyID, OutputFamilyID: alternative.FamilyID,
							SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
							Capabilities:  []construction.Capability{"negation-negated"}, Provenance: englishNegationConstructionSource},
						TargetCandidateID: candidateID, Span: notSpan,
					})
					operatorCount++
				}
			}
		}
	}
	evidence.OperatorCoverageComplete = operatorCount > 0
	if evidence.OperatorCoverageComplete {
		evidence = addEnglishIdentificationReferences(evidence, candidates, forest, lattice)
	}
	return evidence, nil
}

// englishTransitivePastCapability recognizes only retained, complete forest
// alternatives backed by an approved binding in the compiled transitive-past
// family. It derives eligible senses from the exact VP candidate features.
func englishTransitivePastCapability(candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry, valencyIndex *valency.Index) map[ids.SenseID]bool {
	capabilities := make(map[ids.SenseID]bool)
	if registry == nil || valencyIndex == nil {
		return capabilities
	}
	familyID, err := ids.NewConstructionID("en.transitive_past")
	if err != nil {
		return capabilities
	}
	record, exists := registry.ByID(familyID)
	if !exists {
		return capabilities
	}
	frames := make(map[valency.FrameID]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[frame.ID] = frame
	}
	approved := make(map[string]bool)
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID {
			continue
		}
		approved[binding.SenseID.String()+"\x00"+string(binding.ValencyFrameID)+"\x00"+binding.VariantID.String()] = true
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != familyID || len(alternative.Unresolved) != 0 {
				continue
			}
			sense := alternative.Permission.SenseID
			bindingKey := sense.String() + "\x00" + string(alternative.Permission.FrameID) + "\x00" + alternative.VariantID.String()
			if !approved[bindingKey] {
				continue
			}
			frame, frameExists := frames[valency.FrameID(alternative.Permission.FrameID)]
			if !frameExists || frame.SenseID != sense || !valencyFrameAllowsFamily(frame, familyID) {
				continue
			}
			subjectRole, objectRole, frameOK := englishTransitiveFrameRoles(record.Family, frame)
			if !frameOK {
				continue
			}
			slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
			for _, slot := range alternative.Slots {
				if _, duplicate := slots[slot.SlotID]; duplicate {
					slots = nil
					break
				}
				slots[slot.SlotID] = slot
			}
			if slots == nil {
				continue
			}
			if slots["subject"].Role != subjectRole || slots["object"].Role != objectRole {
				continue
			}
			validAlternative := true
			for _, spec := range record.Family.SlotSpecs {
				if spec.Optionality != construction.SlotRequired {
					continue
				}
				slot, found := slots[spec.ID]
				if !found || slot.WHGap != nil || len(slot.CandidateIDs) == 0 {
					validAlternative = false
					break
				}
				validSlot := false
				for _, candidateID := range slot.CandidateIDs {
					candidate, found := byID[candidateID]
					if !found || candidate.Category != spec.Category {
						continue
					}
					if spec.Category == construction.CategoryVP &&
						(candidate.Features["verb_form"] != "finite" || candidate.Features["tense"] != "past" || !containsEnglishSenseName(candidate.SenseIDs, strings.TrimPrefix(sense.String(), "sense:"))) {
						continue
					}
					validSlot = true
					break
				}
				if !validSlot {
					validAlternative = false
					break
				}
			}
			if validAlternative {
				capabilities[sense] = true
			}
		}
	}
	return capabilities
}

func englishTransitivePastCoverage(candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry, valencyIndex *valency.Index) bool {
	return len(englishTransitivePastCapability(candidates, forest, registry, valencyIndex)) > 0
}

// englishRelationSourceFeatureAlternatives derives the source-side relation
// capability from the exact selected parse routes. Returning alternatives
// keyed by sense lets BareDeclarativeEvidenceBuilder apply its ordinary
// complete-parse and finite-feature checks instead of bypassing them.
func englishRelationSourceFeatureAlternatives(lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry, packages []construction.CompiledConstructionPackage, lexiconIndex *lexicon.Index, ontology *core.ConceptGraph, valencyIndex *valency.Index) map[ids.SenseID][]map[features.FeatureID]features.FeatureValueID {
	if registry == nil || lexiconIndex == nil || ontology == nil || valencyIndex == nil || len(forest.Roots) != 1 || len(forest.Unmatched) != 0 ||
		(forest.Outcome != status.OutcomeOK && forest.Outcome != status.OutcomeAmbiguous) {
		return nil
	}
	for _, token := range lattice.Tokens {
		if token.Token.Folded == "not" || token.Token.Folded == "never" || strings.Contains(token.Token.Folded, "n't") {
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
				len(frame.Arguments) != 2 || len(frame.SurfaceOmittedRoles) != 0 || !valencyFrameAllowsFamily(frame, alternative.FamilyID) {
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
			if !exists || coreConcept.Kind != core.KindRelation {
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
				if spec.ArgumentBinding != nil || spec.Category != construction.CategoryVP || spec.ModifierBinding != nil || phrase.Category != construction.CategoryVP || !englishPhraseHasSense(phrase.SenseIDs, senseID) {
					return nil
				}
				predicateCount++
				predicate = phrase
			}
			if predicateCount != 1 || len(roleSlotCount) != 2 || !englishRelationPresentFinite(predicate.Features) {
				return nil
			}
			for _, role := range frame.SemanticRoles {
				if roleSlotCount[role] != 1 {
					return nil
				}
			}
			result[senseID] = []map[features.FeatureID]features.FeatureValueID{
				{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "singular"},
				{"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third", "verb_number": "plural"},
			}
		}
	}
	if clauseCount != 1 || len(result) == 0 {
		return nil
	}
	return result
}

func englishRelationPresentFinite(actual map[features.FeatureID]features.FeatureValueID) bool {
	for key, expected := range map[features.FeatureID]features.FeatureValueID{
		"verb_form": "finite", "tense": "present", "mood": "indicative", "voice": "active", "person": "third",
	} {
		if actual[key] != expected {
			return false
		}
	}
	return actual["verb_number"] == "singular" || actual["verb_number"] == "plural"
}

func englishPhraseHasSense(senses []ids.SenseID, wanted ids.SenseID) bool {
	for _, sense := range senses {
		if sense == wanted {
			return true
		}
	}
	return false
}

func cloneEnglishSupportedSenses(source map[ids.SenseID]bool) map[ids.SenseID]bool {
	cloned := make(map[ids.SenseID]bool, len(source))
	for sense, supported := range source {
		cloned[sense] = supported
	}
	return cloned
}

func englishTransitiveFrameRoles(family construction.ConstructionFamily, frame valency.ValencyFrame) (meaning.Role, meaning.Role, bool) {
	specs := make(map[string]construction.SlotSpec, len(family.SlotSpecs))
	for _, spec := range family.SlotSpecs {
		if _, duplicate := specs[spec.ID]; duplicate {
			return "", "", false
		}
		specs[spec.ID] = spec
	}
	for _, id := range []string{"subject", "verb", "object"} {
		spec, found := specs[id]
		if !found || spec.Optionality != construction.SlotRequired || spec.Cardinality.Minimum != 1 || spec.Cardinality.Maximum != 1 {
			return "", "", false
		}
		category := construction.CategoryNP
		if id == "verb" {
			category = construction.CategoryVP
		}
		if spec.Category != category {
			return "", "", false
		}
	}
	for _, spec := range family.SlotSpecs {
		if spec.Optionality == construction.SlotRequired && spec.Cardinality.Minimum > 0 && spec.ID != "subject" && spec.ID != "verb" && spec.ID != "object" {
			return "", "", false
		}
	}
	var subjectRole, objectRole meaning.Role
	for _, argument := range frame.Arguments {
		if !argument.Required {
			continue
		}
		switch argument.ID {
		case "subject":
			if argument.GrammaticalFunction != "subject" || argument.Role == "" || subjectRole != "" {
				return "", "", false
			}
			subjectRole = argument.Role
		case "object":
			if argument.GrammaticalFunction != "object" || argument.Role == "" || objectRole != "" {
				return "", "", false
			}
			objectRole = argument.Role
		default:
			return "", "", false
		}
	}
	if subjectRole == "" || objectRole == "" || subjectRole == objectRole {
		return "", "", false
	}
	if !englishFrameDeclaresRole(frame, subjectRole) || !englishFrameDeclaresRole(frame, objectRole) {
		return "", "", false
	}
	return subjectRole, objectRole, true
}

func englishFrameDeclaresRole(frame valency.ValencyFrame, role meaning.Role) bool {
	for _, declared := range frame.SemanticRoles {
		if declared == role {
			return true
		}
	}
	return false
}

func appendPastFiniteAlternative(alternatives []map[features.FeatureID]features.FeatureValueID) []map[features.FeatureID]features.FeatureValueID {
	for _, alternative := range alternatives {
		if alternative["verb_form"] == "finite" && alternative["tense"] == "past" && alternative["mood"] == "indicative" && alternative["voice"] == "active" {
			return alternatives
		}
	}
	return append(alternatives, map[features.FeatureID]features.FeatureValueID{
		"verb_form": "finite", "tense": "past", "mood": "indicative", "voice": "active",
	})
}

// englishCommerceTransferPastCapability derives simple-past coverage from
// retained plans, approved bindings, finite-past VP evidence, and the exact
// frame argument/omission contract. Transfer and writing-message routes need
// every required role; commerce routes may omit only roles expressly licensed
// by their frame.
func englishCommerceTransferPastCapability(candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry, valencyIndex *valency.Index) (map[ids.SenseID]bool, bool) {
	capabilities := make(map[ids.SenseID]bool)
	if registry == nil || valencyIndex == nil {
		return capabilities, false
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
		clauseCovered := false
		for _, alternative := range node.Alternatives {
			if len(alternative.Unresolved) != 0 {
				continue
			}
			record, exists := registry.ByID(alternative.FamilyID)
			if !exists {
				continue
			}
			approved := false
			for _, binding := range registry.BindingsByFamilyID(alternative.FamilyID) {
				if binding.Approved && binding.SenseID == alternative.Permission.SenseID && binding.ValencyFrameID == alternative.Permission.FrameID && binding.VariantID.String() == alternative.VariantID.String() {
					approved = true
					break
				}
			}
			if !approved {
				continue
			}
			var frame *valency.ValencyFrame
			for _, candidateFrame := range valencyIndex.FramesBySenseID(alternative.Permission.SenseID) {
				if candidateFrame.ID == alternative.Permission.FrameID {
					copy := candidateFrame
					frame = &copy
					break
				}
			}
			if frame == nil || !commerceTransferFrameRoles(*frame) {
				continue
			}
			slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
			duplicate := false
			for _, slot := range alternative.Slots {
				if _, exists := slots[slot.SlotID]; exists {
					duplicate = true
					break
				}
				slots[slot.SlotID] = slot
			}
			if duplicate || !commerceTransferRequiredSlotsPresent(record.Family.SlotSpecs, *frame, slots, byID, alternative.Permission.SenseID) {
				continue
			}
			capabilities[alternative.Permission.SenseID] = true
			clauseCovered = true
		}
		if !clauseCovered {
			return capabilities, false
		}
	}
	return capabilities, seenClause
}

func commerceTransferFrameRoles(frame valency.ValencyFrame) bool {
	roles := make(map[meaning.Role]bool, len(frame.SemanticRoles))
	for _, role := range frame.SemanticRoles {
		roles[role] = true
	}
	transfer := roles["giver"] && roles["recipient"] && roles["theme"]
	commerce := roles["buyer"] && roles["goods"] && roles["seller"]
	writingMessage := roles["writer"] && roles["message"] && roles["recipient"]
	return transfer || commerce || writingMessage
}

func commerceTransferRequiredSlotsPresent(specs []construction.SlotSpec, frame valency.ValencyFrame, slots map[string]parser.SlotBinding, candidates map[string]parser.PhraseCandidate, sense ids.SenseID) bool {
	specByID := make(map[string]construction.SlotSpec, len(specs))
	for _, spec := range specs {
		if _, duplicate := specByID[spec.ID]; duplicate {
			return false
		}
		specByID[spec.ID] = spec
	}
	argumentByID := make(map[string]valency.ArgumentSlot, len(frame.Arguments))
	for _, argument := range frame.Arguments {
		if _, duplicate := argumentByID[argument.ID]; duplicate {
			return false
		}
		argumentByID[argument.ID] = argument
		spec, hasSpec := specByID[argument.ID]
		slot, hasSlot := slots[argument.ID]
		if argument.Required {
			if !hasSpec || spec.Category != construction.CategoryNP || spec.Optionality != construction.SlotRequired || !hasSlot || slot.Role != argument.Role || slot.WHGap != nil || len(slot.CandidateIDs) == 0 || !allEnglishNPs(slot.CandidateIDs, candidates) {
				return false
			}
			continue
		}
		omissionLicensed := containsMeaningRole(frame.SurfaceOmittedRoles, argument.Role)
		if !hasSpec {
			if !omissionLicensed || hasSlot {
				return false
			}
			continue
		}
		if spec.Category != construction.CategoryNP || hasSlot && (slot.Role != argument.Role || slot.WHGap != nil || len(slot.CandidateIDs) == 0 || !allEnglishNPs(slot.CandidateIDs, candidates)) {
			return false
		}
		if !hasSlot && (spec.Optionality != construction.SlotOptional || !omissionLicensed) {
			return false
		}
	}
	for _, spec := range specs {
		if spec.Category == construction.CategoryVP && spec.IsMainPredicate() {
			continue
		}
		if spec.Category != construction.CategoryNP {
			return false
		}
		if _, backed := argumentByID[spec.ID]; !backed {
			return false
		}
	}
	var verbSpec construction.SlotSpec
	verbCount := 0
	for _, spec := range specs {
		if spec.Category == construction.CategoryVP && spec.IsMainPredicate() {
			verbSpec = spec
			verbCount++
		}
	}
	if verbCount != 1 {
		return false
	}
	verbSlot, exists := slots[verbSpec.ID]
	if !exists || verbSlot.Role != "" || verbSlot.WHGap != nil || len(verbSlot.CandidateIDs) != 1 {
		return false
	}
	verb, exists := candidates[verbSlot.CandidateIDs[0]]
	if !exists || verb.Category != construction.CategoryVP || verb.Features["verb_form"] != "finite" || verb.Features["tense"] != "past" || !containsEnglishSenseName(verb.SenseIDs, strings.TrimPrefix(sense.String(), "sense:")) {
		return false
	}
	return true
}

func containsMeaningRole(roles []meaning.Role, target meaning.Role) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func allEnglishNPs(candidateIDs []string, candidates map[string]parser.PhraseCandidate) bool {
	for _, candidateID := range candidateIDs {
		candidate, exists := candidates[candidateID]
		if !exists || candidate.Category != construction.CategoryNP {
			return false
		}
	}
	return len(candidateIDs) > 0
}

// englishIntransitiveSimplePastCapability accepts a past predicate only when a
// retained compiled parse alternative is licensed by the approved intransitive
// binding and contains exactly one finite past VP candidate for that sense.
// The parse forest supplies the selected plan/variant; the registry supplies
// the approved sense/frame/family/variant relation.
func englishIntransitiveSimplePastCapability(candidates []parser.PhraseCandidate, forest parser.Forest, registry *construction.Registry) (map[ids.SenseID]bool, bool) {
	allowed := make(map[ids.SenseID]bool)
	if registry == nil {
		return allowed, false
	}
	familyID, err := ids.ParseConstructionID("construction:en.intransitive")
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
			verbCount := 0
			for _, slot := range alternative.Slots {
				if slot.SlotID != "verb" {
					continue
				}
				for _, id := range slot.CandidateIDs {
					candidate, exists := byID[id]
					if !exists || candidate.Category != construction.CategoryVP || candidate.Features["verb_form"] != "finite" || candidate.Features["tense"] != "past" {
						continue
					}
					hasSense := false
					for _, sense := range candidate.SenseIDs {
						if sense == alternative.Permission.SenseID {
							hasSense = true
							break
						}
					}
					if hasSense {
						verbCount++
					}
				}
			}
			if verbCount != 1 {
				return allowed, false
			}
			allowed[alternative.Permission.SenseID] = true
		}
	}
	return allowed, seenClause
}

// approvedEnglishIntransitiveSenses derives the finite-predicate inventory
// from approved bindings whose frame and family each contain exactly one
// required subject argument and one main VP slot.
func approvedEnglishIntransitiveSenses(registry *construction.Registry, valencyIndex *valency.Index, variants map[ids.ConstructionID]*construction.VariantIndex) map[ids.SenseID]bool {
	allowed := make(map[ids.SenseID]bool)
	if registry == nil || valencyIndex == nil {
		return allowed
	}
	familyID, err := ids.ParseConstructionID("construction:en.intransitive")
	if err != nil {
		return allowed
	}
	record, exists := registry.ByID(familyID)
	variantIndex := variants[familyID]
	if !exists || variantIndex == nil || !englishIntransitiveFamilyShape(record.Family) {
		return allowed
	}
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID || !englishIntransitiveFrameShape(binding.SenseID, binding.ValencyFrameID, familyID, valencyIndex) {
			continue
		}
		if _, exists := variantIndex.ByID(construction.VariantID(binding.VariantID.String())); !exists {
			continue
		}
		allowed[binding.SenseID] = true
	}
	return allowed
}

func englishIntransitiveFamilyShape(family construction.ConstructionFamily) bool {
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

func englishIntransitiveFrameShape(senseID ids.SenseID, frameID valency.FrameID, familyID ids.ConstructionID, valencyIndex *valency.Index) bool {
	for _, frame := range valencyIndex.FramesBySenseID(senseID) {
		if frame.ID != frameID || frame.SenseID != senseID || !valencyFrameAllowsFamily(frame, familyID) || len(frame.SemanticRoles) != 1 || len(frame.Arguments) != 1 {
			continue
		}
		argument := frame.Arguments[0]
		if argument.ID == "subject" && argument.Required && argument.GrammaticalFunction == "subject" && argument.Role == frame.SemanticRoles[0] {
			return true
		}
	}
	return false
}

func filterEnglishGivingArticleAlternatives(forest parser.Forest, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, lexical *lexicon.Index, registry *construction.Registry) parser.Forest {
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	for nodeIndex := range forest.Nodes {
		node := &forest.Nodes[nodeIndex]
		if node.Category != "clause" {
			continue
		}
		kept := node.Alternatives[:0]
		for _, alternative := range node.Alternatives {
			if (alternative.FamilyID.String() == "construction:en.give.transfer" || alternative.FamilyID.String() == "construction:en.give.transfer.past") && !englishGivingSourceAlternativeAllowed(alternative, lattice, byID, lexical, registry) {
				continue
			}
			kept = append(kept, alternative)
		}
		node.Alternatives = kept
	}
	return forest
}

func englishGivingSourceAlternativeAllowed(alternative parser.ParseAlternative, lattice analyzer.LexicalLattice, candidates map[string]parser.PhraseCandidate, lexical *lexicon.Index, registry *construction.Registry) bool {
	if alternative.VariantID.String() == "construction:en.give.transfer.past.double_np.compact_articles" ||
		alternative.VariantID.String() == "construction:en.give.transfer.past.to_pp.compact_articles" {
		if alternative.FamilyID.String() != "construction:en.give.transfer.past" || len(alternative.Unresolved) != 0 || registry == nil || lexical == nil {
			return false
		}
		allowAmbiguousNPs := false
		var wantRoles map[string]string
		switch {
		case alternative.Permission.SenseID.String() == "sense:en.give.transfer" && string(alternative.Permission.FrameID) == "en.give.transfer.past":
			wantRoles = map[string]string{"subject": "giver", "recipient": "recipient", "theme": "theme"}
		case alternative.VariantID.String() == "construction:en.give.transfer.past.to_pp.compact_articles" &&
			alternative.Permission.SenseID.String() == "sense:en.writing_message.event" && string(alternative.Permission.FrameID) == "en.writing_message.full":
			allowAmbiguousNPs = true
			wantRoles = map[string]string{"subject": "writer", "recipient": "recipient", "theme": "message"}
		default:
			return false
		}
		approved := false
		for _, binding := range registry.BindingsByFamilyID(alternative.FamilyID) {
			if binding.Approved && binding.FamilyID == alternative.FamilyID && binding.SenseID == alternative.Permission.SenseID &&
				string(binding.ValencyFrameID) == string(alternative.Permission.FrameID) && binding.VariantID.String() == alternative.VariantID.String() {
				approved = true
				break
			}
		}
		if !approved {
			return false
		}
		if len(alternative.Slots) != len(wantRoles)+1 {
			return false
		}
		seenSlots := make(map[string]bool, len(alternative.Slots))
		for _, slot := range alternative.Slots {
			if seenSlots[slot.SlotID] {
				return false
			}
			seenSlots[slot.SlotID] = true
			if slot.SlotID == "verb" {
				if len(slot.CandidateIDs) == 0 {
					return false
				}
				foundPastWritingVerb := false
				for _, candidateID := range slot.CandidateIDs {
					phrase, exists := candidates[candidateID]
					if !exists || phrase.Category != construction.CategoryVP {
						return false
					}
					if phrase.Features["verb_form"] == "finite" && phrase.Features["tense"] == "past" &&
						containsEnglishSenseName(phrase.SenseIDs, strings.TrimPrefix(alternative.Permission.SenseID.String(), "sense:")) {
						foundPastWritingVerb = true
					}
				}
				if !foundPastWritingVerb {
					return false
				}
				continue
			}
			if _, requiredNP := wantRoles[slot.SlotID]; !requiredNP {
				return false
			}
		}
		if !seenSlots["verb"] {
			return false
		}
		for slotID, wantRole := range wantRoles {
			var slot *parser.SlotBinding
			for index := range alternative.Slots {
				if alternative.Slots[index].SlotID == slotID {
					if slot != nil {
						return false
					}
					slot = &alternative.Slots[index]
				}
			}
			if slot == nil || string(slot.Role) != wantRole || len(slot.CandidateIDs) == 0 || !allowAmbiguousNPs && len(slot.CandidateIDs) != 1 {
				return false
			}
			for _, candidateID := range slot.CandidateIDs {
				phrase, exists := candidates[candidateID]
				if !exists || phrase.Category != construction.CategoryNP || phrase.LexemeID.String() == "" ||
					phrase.Features["determiner"] != "definite" && phrase.Features["determiner"] != "indefinite" || phrase.TokenStart <= 0 {
					return false
				}
				article, hasArticle := englishArticleAt(lattice, phrase.TokenStart-1)
				if !hasArticle || string(phrase.Features["determiner"]) != englishDeterminerFeature(article) {
					return false
				}
				lexeme, hasLexeme := lexical.ByLexemeID(phrase.LexemeID)
				if !hasLexeme {
					return false
				}
				if article == "the" {
					continue
				}
				articleForm, knownArticle := englishIndefiniteArticle(lexeme.Properties)
				mass, pluralOnly := englishNumberClass(lexeme.Properties, lexeme.ParadigmID)
				number := knownEnglishNounNumber(phrase.Features)
				if !knownArticle || articleForm != article || mass || pluralOnly || number != "singular" || !hasNumberBehavior(lexical, phrase.LexemeID, "count") {
					return false
				}
			}
		}
		return true
	}
	for _, slotID := range []string{"subject", "recipient", "theme"} {
		article, found := englishGivingSequenceArticle(alternative.VariantID.String(), slotID)
		if !found {
			return false
		}
		var binding *parser.SlotBinding
		for index := range alternative.Slots {
			if alternative.Slots[index].SlotID == slotID {
				binding = &alternative.Slots[index]
				break
			}
		}
		if binding == nil || len(binding.CandidateIDs) == 0 {
			return false
		}
		for _, candidateID := range binding.CandidateIDs {
			phrase, exists := candidates[candidateID]
			if !exists || phrase.LexemeID.String() == "" {
				return false
			}
			lexeme, exists := lexical.ByLexemeID(phrase.LexemeID)
			if !exists {
				return false
			}
			if article == "a" || article == "an" {
				articleForm, knownArticle := englishIndefiniteArticle(lexeme.Properties)
				mass, pluralOnly := englishNumberClass(lexeme.Properties, lexeme.ParadigmID)
				number := string(phrase.Features["noun_number"])
				if !knownArticle || articleForm != article || mass || pluralOnly || number != "singular" {
					return false
				}
			}
		}
	}
	return true
}

func englishGivingSequenceArticle(variantID, slotID string) (string, bool) {
	parts := strings.Split(variantID, ".")
	if len(parts) == 0 {
		return "", false
	}
	articles := strings.Split(parts[len(parts)-1], "_")
	// Variant IDs encode the articles in subject_recipient_theme order for
	// both layouts. The to_pp surface sequence places the theme before the
	// recipient, but that does not change the suffix order.
	positions := map[string]int{"subject": 0, "recipient": 1, "theme": 2}
	position := positions[slotID]
	if len(articles) != 3 || position >= len(articles) {
		return "", false
	}
	article := articles[position]
	return article, article == "the" || article == "a" || article == "an"
}

func englishWHQuestionCoverage(lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) bool {
	seenClause := false
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
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
			case "construction:en.question.wh.object.neutral.do.analyze":
				terminal, auxiliary, slotID, role = "whom", "do", "object", "stimulus"
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
			if auxiliary != "" {
				if len(lattice.Tokens) < 4 || lattice.Tokens[0].Token.Folded != "whom" ||
					lattice.Tokens[1].Token.Folded != auxiliary || lattice.Tokens[2].Token.Folded != "the" ||
					lattice.Tokens[len(lattice.Tokens)-1].Token.Folded != "?" {
					return false
				}
				var subject, verb *parser.SlotBinding
				for i := range alternative.Slots {
					switch alternative.Slots[i].SlotID {
					case "subject":
						subject = &alternative.Slots[i]
					case "verb":
						verb = &alternative.Slots[i]
					}
				}
				if subject == nil || verb == nil || subject.WHGap != nil || verb.WHGap != nil ||
					len(subject.CandidateIDs) == 0 || len(verb.CandidateIDs) == 0 {
					return false
				}
				for _, verbID := range verb.CandidateIDs {
					vp, exists := byID[verbID]
					if !exists || vp.Category != construction.CategoryVP || vp.TokenStart <= 3 || vp.TokenEnd != len(lattice.Tokens)-1 {
						return false
					}
					for _, subjectID := range subject.CandidateIDs {
						np, exists := byID[subjectID]
						if !exists || np.Category != construction.CategoryNP || np.TokenStart != 3 || np.TokenEnd != vp.TokenStart {
							return false
						}
					}
				}
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
		evidence = addEnglishIdentificationReferences(evidence, candidates, forest, lattice)
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
	inner           runtime.SimplePhraseEvidenceBuilder
	lexicon         *lexicon.Index
	resolver        englishNegationResolver
	predicateSenses map[ids.SenseID]bool
}

func englishAttributionArticleForPhrase(candidate parser.PhraseCandidate, index *lexicon.Index) (string, bool) {
	licensed := false
	for _, spec := range englishAttributionSpecs {
		if candidate.LexemeID.String() == spec.lexeme && containsEnglishSenseName(candidate.SenseIDs, strings.TrimPrefix(spec.sense, "sense:")) {
			licensed = true
			break
		}
	}
	if !licensed {
		return "", false
	}
	if index == nil {
		return "", false
	}
	lexeme, exists := index.LookupLexeme(candidate.LexemeID)
	if !exists {
		return "", false
	}
	return englishIndefiniteArticle(lexeme.Properties)
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
			if candidate.LexemeID.String() == "lexeme:en.falling.v" && candidate.TokenEnd < len(lattice.Tokens) && lattice.Tokens[candidate.TokenEnd].Token.Folded == "asleep" {
				continue
			}
			if number := candidate.Features["verb_number"]; number != "" {
				candidate.Features["number"] = number
			}
			if englishNegationAuxiliaryCandidate(lattice, candidate) {
				continue
			}
			route, person, number, tense, negated := builder.resolver.negatedPredicateContext(lattice, candidate, candidates)
			whDoSupport := englishWHDoSupportContext(lattice, candidate)
			whDoPerson, whDoNumber := "", ""
			if whDoSupport {
				var agreed bool
				whDoPerson, whDoNumber, agreed = englishWHDoSupportAgreement(candidate, candidates, lattice)
				if !agreed {
					continue
				}
			}
			if negated {
				if candidate.Features["mood"] != "imperative" || negatedBaseSeen[candidate.TokenStart] {
					continue
				}
				// In this exact auxiliary context, the imperative cell supplies
				// the English base surface; do-support carries available agreement.
				candidate.Features = map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "tense": features.FeatureValueID(tense), "person": features.FeatureValueID(person),
					"mood": "indicative", "voice": "active",
				}
				if number != "" {
					candidate.Features["number"] = features.FeatureValueID(number)
					candidate.Features["verb_number"] = features.FeatureValueID(number)
				}
				negatedBaseSeen[candidate.TokenStart] = true
			}
			if whDoSupport {
				// The auxiliary supplies finite agreement; retain the explicit
				// person and number of its parsed subject on the base verb evidence.
				candidate.Features = map[features.FeatureID]features.FeatureValueID{
					"verb_form": "base", "tense": "present", "person": features.FeatureValueID(whDoPerson),
					"mood": "indicative", "voice": "active",
				}
				if whDoNumber != "" {
					candidate.Features["number"] = features.FeatureValueID(whDoNumber)
					candidate.Features["verb_number"] = features.FeatureValueID(whDoNumber)
				}
			}
			if !supportedEnglishPredicateRoute(candidate, builder.predicateSenses) && (!negated || candidate.LexemeID.String() != route.lexeme) && !whDoSupport {
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
		demonstrative, hasDemonstrative := englishDemonstrativeAt(lattice, determinerToken)
		modifiedByAttribution := ""
		if !hasDeterminer {
			for _, phrase := range candidates {
				if phrase.Category != construction.CategoryAP || phrase.TokenEnd != candidate.TokenStart || phrase.TokenStart == 0 {
					continue
				}
				attribution, licensed := englishAttributionArticleForPhrase(phrase, builder.lexicon)
				if !licensed {
					continue
				}
				articleToken := phrase.TokenStart - 1
				_, found := englishArticleAt(lattice, articleToken)
				if found {
					determinerToken, hasDeterminer, modifiedByAttribution = articleToken, true, attribution
					break
				}
				if deixis, ok := englishDemonstrativeAt(lattice, articleToken); ok {
					determinerToken, demonstrative, hasDemonstrative = articleToken, deixis, true
					break
				}
			}
		}
		if hasDemonstrative {
			if !lexemeFound || lexeme.POS != "noun" || knownEnglishNounNumber(candidate.Features) != demonstrative.number ||
				!hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count") && !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "mass") {
				continue
			}
			candidate.Features["determiner"] = features.FeatureValueID("demonstrative_" + demonstrative.distance)
			candidate.Features["number"] = features.FeatureValueID(demonstrative.number)
			filtered = append(filtered, candidate)
			continue
		}
		if !hasDeterminer {
			if hasEnglishThisNPPrefix(lattice, candidate.TokenStart) {
				if !lexemeFound || lexeme.POS != "noun" ||
					!hasNumberBehavior(builder.lexicon, candidate.LexemeID, "mass") && !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count") {
					continue
				}
				if number := knownEnglishNounNumber(candidate.Features); number != "" {
					candidate.Features["number"] = features.FeatureValueID(number)
				}
				candidate.Features["determiner"] = "bare"
				filtered = append(filtered, candidate)
				continue
			}
			if article, hasIdentificationArticle := englishIdentificationPrefixArticle(lattice, candidate.TokenStart); hasIdentificationArticle {
				if !lexemeFound || lexeme.POS != "noun" || knownEnglishNounNumber(candidate.Features) != "singular" {
					continue
				}
				expected, known := englishIndefiniteArticle(lexeme.Properties)
				if !known || expected != article || !hasNumberBehavior(builder.lexicon, candidate.LexemeID, "count") {
					continue
				}
				candidate.Features["determiner"] = "indefinite"
				candidate.Features["noun_number"] = "singular"
				candidate.Features["number"] = "singular"
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
							if modifiedByAttribution != "" {
								expectedArticle, hasLexicalArticle = modifiedByAttribution, true
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

func approvedEnglishPredicateSenses(index *valency.Index, packages []construction.CompiledConstructionPackage) map[ids.SenseID]bool {
	result := make(map[ids.SenseID]bool)
	if index == nil {
		return result
	}
	parseVariants := make(map[string]bool)
	for _, pkg := range packages {
		for _, variant := range pkg.Package.Variants {
			if variant.DirectionPolicy == construction.DirectionBoth || variant.DirectionPolicy == construction.DirectionAnalyze {
				parseVariants[pkg.Package.Family.ID.String()+"\x00"+variant.ID.String()] = true
			}
		}
	}
	for _, binding := range index.Bindings() {
		if binding.Approved && parseVariants[binding.FamilyID.String()+"\x00"+binding.VariantID.String()] {
			result[binding.SenseID] = true
		}
	}
	return result
}

func supportedEnglishPredicateRoute(candidate parser.PhraseCandidate, approvedSenses map[ids.SenseID]bool) bool {
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
	case "past":
	default:
		return false
	}
	for _, sense := range candidate.SenseIDs {
		if approvedSenses[sense] {
			return true
		}
	}
	return false
}

func englishWHDoSupportContext(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	if candidate.Category != construction.CategoryVP || candidate.LexemeID.String() != "lexeme:en.see.v" ||
		candidate.Features["verb_form"] != "base" || candidate.TokenStart <= 3 || candidate.TokenStart >= len(lattice.Tokens) || len(lattice.Tokens) < 4 {
		return false
	}
	auxiliary := lattice.Tokens[1].Token.Folded
	if lattice.Tokens[0].Token.Folded != "whom" || (auxiliary != "does" && auxiliary != "do") || lattice.Tokens[2].Token.Folded != "the" {
		return false
	}
	whomCount, auxiliaryCount := 0, 0
	for _, token := range lattice.Tokens {
		if token.Token.Folded == "whom" {
			whomCount++
		}
		if token.Token.Folded == auxiliary {
			auxiliaryCount++
		}
	}
	return whomCount == 1 && auxiliaryCount == 1
}

func englishWHDoSupportAgreement(vp parser.PhraseCandidate, candidates []parser.PhraseCandidate, lattice analyzer.LexicalLattice) (string, string, bool) {
	if len(lattice.Tokens) < 4 || lattice.Tokens[0].Token.Folded != "whom" ||
		lattice.Tokens[2].Token.Folded != "the" || vp.TokenStart <= 3 {
		return "", "", false
	}
	auxiliary := lattice.Tokens[1].Token.Folded
	if auxiliary != "does" && auxiliary != "do" {
		return "", "", false
	}
	agreedPerson, agreedNumber, found := "", "", false
	for _, np := range candidates {
		if np.Category != construction.CategoryNP || np.TokenStart != 3 || np.TokenEnd != vp.TokenStart {
			continue
		}
		person, number := string(np.Features["person"]), string(np.Features["number"])
		if number != "singular" && number != "plural" {
			return "", "", false
		}
		compatible := auxiliary == "does" && person == "third" && number == "singular"
		compatible = compatible || auxiliary == "do" && ((person == "third" && number == "plural") || ((person == "first" || person == "second") && (number == "singular" || number == "plural")))
		if !compatible || found && (person != agreedPerson || number != agreedNumber) {
			return "", "", false
		}
		agreedPerson, agreedNumber, found = person, number, true
	}
	return agreedPerson, agreedNumber, found
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
