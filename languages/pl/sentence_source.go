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
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/text"
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
	caseValues := map[features.FeatureValueID]string{}
	for _, value := range []string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"} {
		caseValues[features.FeatureValueID(value)] = value
	}
	return &runtime.SourcePipeline{
		Language: "pl", Tokenize: text.TokenizeOptions{Profile: &text.TokenizerProfile{CaseFold: true}},
		AllowedUnlexicalizedTerminals: map[string]bool{"na": true, "nie": true, "to": true, "ten": true, "jest": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: polishIdentificationPhraseBuilder{inner: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "verb": {construction.CategoryVP}, "adjective": {construction.CategoryAP},
			},
			CaseFeature: "case", CaseValues: caseValues,
		}}},
		EvidenceBuilder: polishSentenceEvidenceBuilder{base: runtime.BareDeclarativeEvidenceBuilder{
			SupportedSenses: map[ids.SenseID]bool{seeingSense: true, hearingSense: true, receivingSense: true, lookingSense: true, readingSense: true, classificationSense: true, attributionSense: true},
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
				readingSense: {
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "plural"},
					{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "masculine_personal"},
					{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "feminine"},
					{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "singular", "gender": "neuter"},
					{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
					{"verb_form": "finite", "tense": "past", "aspect": "imperfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
				},
				receivingSense: {
					{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular"},
					{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
					{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
				},
				classificationSense: {
					{"verb_form": "finite", "tense": "present", "aspect": "imperfective", "person": "third", "number": "singular"},
				},
				attributionSense: {
					{"degree": "positive", "number": "singular", "gender": "masculine_animal", "case": "nominative"},
				},
			},
			AllowedUnlexicalizedTerminals: map[string]bool{"na": true, "nie": true, "to": true, "ten": true, "jest": true},
		}},
	}, nil
}

type polishSentenceEvidenceBuilder struct {
	base runtime.BareDeclarativeEvidenceBuilder
}

func (builder polishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	notSpan, hasNegation := polishTerminalSpan(lattice, "nie")
	if !hasNegation {
		for _, token := range lattice.Tokens {
			if token.Token.Folded == "nie" {
				return runtime.CompositionEvidence{}, nil
			}
		}
		evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
		if err != nil || !evidence.OperatorCoverageComplete {
			return evidence, err
		}
		return addPolishIdentificationReferences(evidence, candidates, forest), nil
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
