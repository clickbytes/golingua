package pl

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

// NewSentenceSource builds the trusted Polish source stages against one exact
// CORE/lexicon snapshot. The declarative completion attestation covers
// singular and plural present visual/auditory/looking, singular visual negation,
// and third-person past receiving predicates.
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
		AllowedUnlexicalizedTerminals: map[string]bool{"na": true, "nie": true},
		Morphology:                    inverse, MWEPlans: []construction.CompiledMWEPlan{},
		Lexicon: snapshot.Lexicon(), ParserIndex: plans, Valency: valencyIndex,
		Core: snapshot.CoreGraph(), Constructions: registry,
		PhraseBuilder: runtime.SimplePhraseEvidenceBuilder{Config: runtime.POSPhraseConfig{
			Categories: map[lexicon.PartOfSpeech][]construction.SlotCategory{
				"noun": {construction.CategoryNP}, "verb": {construction.CategoryVP},
			},
			CaseFeature: "case", CaseValues: caseValues,
		}},
		EvidenceBuilder: polishSentenceEvidenceBuilder{base: runtime.BareDeclarativeEvidenceBuilder{
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
				receivingSense: {
					{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "singular"},
					{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "masculine_personal"},
					{"verb_form": "finite", "tense": "past", "aspect": "perfective", "person": "third", "number": "plural", "gender": "non_masculine_personal"},
				},
			},
			AllowedUnlexicalizedTerminals: map[string]bool{"na": true, "nie": true},
		}},
	}, nil
}

type polishSentenceEvidenceBuilder struct {
	base runtime.BareDeclarativeEvidenceBuilder
}

func (builder polishSentenceEvidenceBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) (runtime.CompositionEvidence, error) {
	notSpan, hasNegation := polishTerminalSpan(lattice, "nie")
	if !hasNegation {
		return builder.base.Build(ctx, lattice, candidates, forest)
	}
	seeingSense, err := ids.NewSenseID("pl.widziec.visual")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	evidence, err := builder.base.Build(ctx, lattice, candidates, forest)
	if err != nil || !evidence.OperatorCoverageComplete {
		return evidence, err
	}
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
	}
	negativeFamily, err := ids.NewConstructionID("pl.transitive_negative")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	negativeVariant, err := ids.NewConstructionID("pl.transitive_negative.svo")
	if err != nil {
		return runtime.CompositionEvidence{}, err
	}
	operatorCount := 0
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			if alternative.FamilyID != negativeFamily || alternative.VariantID.String() != negativeVariant.String() || alternative.Permission.SenseID != seeingSense {
				return runtime.CompositionEvidence{}, nil
			}
			for _, slot := range alternative.Slots {
				if slot.Role != "" {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || !polishNegatedSeeingCandidate(candidate) {
						continue
					}
					evidence.Operators = append(evidence.Operators, semantic.OperatorEvidence{
						Operator: construction.Operator{
							ID: "negate", Kind: construction.OperatorNegation, Value: construction.ValueNegated,
							InputFamilyID: negativeFamily, OutputFamilyID: negativeFamily,
							SemanticScope: construction.SemanticScope{TargetNodeID: "event"},
							Capabilities:  []construction.Capability{"negation-negated"}, Provenance: polishNegationConstructionSource,
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

func polishTerminalSpan(lattice analyzer.LexicalLattice, terminal string) (text.Span, bool) {
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

func polishNegatedSeeingCandidate(candidate parser.PhraseCandidate) bool {
	return candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == "lexeme:pl.widziec.v" &&
		candidate.Features["verb_form"] == "finite" && candidate.Features["tense"] == "present" &&
		candidate.Features["person"] == "third" && candidate.Features["number"] == "singular"
}
