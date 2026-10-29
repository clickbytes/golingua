package pl

import (
	"context"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

type polishIdentificationPhraseBuilder struct {
	inner          runtime.SimplePhraseEvidenceBuilder
	phraseProfile  construction.PhraseProfile
	featureProfile *features.LanguageFeatureProfile
}

func (builder polishIdentificationPhraseBuilder) Build(ctx context.Context, lattice analyzer.LexicalLattice) ([]parser.PhraseCandidate, parser.ParseContext, error) {
	candidates, parseContext, err := builder.inner.Build(ctx, lattice)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	allowedAttributionLexemes := make(map[string]bool, len(polishAttributionSpecs))
	for _, spec := range polishAttributionSpecs {
		allowedAttributionLexemes[spec.lexeme] = true
	}
	const copulaLexeme = "lexeme:pl.byc.v"
	// Preserve each independently licensed noun/adjective candidate. Agreement
	// is carried as a language-local feature and enforced by compiled slot
	// constraints, so homonymous noun candidates cannot license one another.
	for index := range candidates {
		candidates[index].Features = polishCandidateAgreementFeatures(candidates[index])
	}
	filtered := make([]parser.PhraseCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP {
			if !allowedAttributionLexemes[candidate.LexemeID.String()] {
				continue
			}
		}
		if candidate.Category == construction.CategoryNP && candidate.TokenStart > 0 {
			terminal := lattice.Tokens[candidate.TokenStart-1].Token.Folded
			switch terminal {
			case "te", "ci", "tamte", "tamci":
				number, numberKnown := candidate.Features["number"]
				if numberKnown && number != "plural" {
					continue
				}
				if !numberKnown {
					filtered = append(filtered, candidate)
					continue
				}
				gender, known := attestedNounGender(strings.TrimPrefix(candidate.LexemeID.String(), "lexeme:"))
				if !known { // leave the demonstrative unresolved when lexical gender is unknown
					filtered = append(filtered, candidate)
					continue
				}
				masculinePersonal := gender == "masculine_personal"
				if (terminal == "ci" || terminal == "tamci") != masculinePersonal {
					continue
				}
			}
		}
		if candidate.Category == construction.CategoryVP && candidate.LexemeID.String() == copulaLexeme && polishNominativeAttributionAt(candidates, candidate.TokenEnd, allowedAttributionLexemes) {
			continue
		}
		filtered = append(filtered, candidate)
	}
	composed, err := runtime.AddAttributiveNPEvidence(lattice, filtered, builder.phraseProfile, builder.featureProfile, polishNPPhraseDeterminer)
	if err != nil {
		return nil, parser.ParseContext{}, err
	}
	return composed, parseContext, nil
}

func polishNPPhraseDeterminer(lattice analyzer.LexicalLattice, _ parser.PhraseCandidate, modifier parser.PhraseCandidate) *parser.NPPhraseTokenEvidence {
	index := modifier.TokenStart - 1
	if index < 0 || index >= len(lattice.Tokens) {
		return nil
	}
	token := lattice.Tokens[index].Token
	surface := token.Folded
	number := ""
	switch surface {
	case "ten", "ta", "to", "tamten", "tamta", "tamto":
		number = "singular"
	case "te", "ci", "tamte", "tamci":
		number = "plural"
	default:
		return nil
	}
	values := map[features.FeatureID]features.FeatureValueID{"number": features.FeatureValueID(number)}
	var compatible []features.FeatureValueID
	switch surface {
	case "ten", "tamten":
		compatible = []features.FeatureValueID{"masculine_animal", "masculine_inanimate", "masculine_personal"}
	case "ta", "tamta":
		compatible = []features.FeatureValueID{"feminine"}
	case "to", "tamto":
		compatible = []features.FeatureValueID{"neuter"}
	case "ci", "tamci":
		compatible = []features.FeatureValueID{"masculine_personal"}
	case "te", "tamte":
		compatible = []features.FeatureValueID{"non_masculine_personal"}
	}
	return runtime.NewNPPhraseTokenEvidence("pl", index, token.Surface, token.Span, values, compatible...)
}

func polishCandidateAgreementFeatures(candidate parser.PhraseCandidate) map[features.FeatureID]features.FeatureValueID {
	values := make(map[features.FeatureID]features.FeatureValueID, len(candidate.Features)+1)
	for key, value := range candidate.Features {
		values[key] = value
	}
	number := string(candidate.Features["number"])
	gender := ""
	switch candidate.Category {
	case construction.CategoryNP:
		gender, _ = attestedNounGender(strings.TrimPrefix(candidate.LexemeID.String(), "lexeme:"))
		if gender == "" && candidate.LexemeID.String() == "lexeme:pl.drzwi.n" {
			gender = "non_masculine_personal"
		}
		if number == "plural" && gender != "" {
			gender = polishPluralAgreementGender(gender)
		}
	case construction.CategoryAP:
		gender = string(candidate.Features["gender"])
	}
	if gender != "" && (number == "singular" || number == "plural") {
		values["agreement_gender"] = features.FeatureValueID(gender)
	}
	return values
}

func polishPluralAgreementGender(gender string) string {
	if gender == "masculine_personal" {
		return "masculine_personal"
	}
	if gender == "" {
		return ""
	}
	return "non_masculine_personal"
}

func polishNominativeAttributionAt(candidates []parser.PhraseCandidate, start int, allowed map[string]bool) bool {
	for _, candidate := range candidates {
		if candidate.Category == construction.CategoryAP && candidate.TokenStart == start &&
			allowed[candidate.LexemeID.String()] && candidate.CaseKnown && candidate.Case == "nominative" {
			return true
		}
	}
	return false
}

func addPolishIdentificationReferences(evidence runtime.CompositionEvidence, lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate, forest parser.Forest) runtime.CompositionEvidence {
	byID := make(map[string]parser.PhraseCandidate, len(candidates))
	composites := make(map[string]parser.CompositeNPPhraseCandidate)
	for _, candidate := range candidates {
		byID[candidate.ID] = candidate
		for _, composite := range candidate.CompositeNPs {
			if composite.ID != "" {
				composites[composite.ID] = composite
			}
		}
	}
	seen := make(map[string]bool, len(evidence.References))
	for _, reference := range evidence.References {
		if reference.Anchor != nil {
			seen[reference.CandidateID+"\x00"+reference.Anchor.Kind+"\x00"+reference.Anchor.Value] = true
		}
	}
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alternative := range node.Alternatives {
			for _, group := range alternative.NPPhraseGroups {
				if group.DeterminerCandidateID == "" {
					continue
				}
				composite, exists := composites[group.CompositeCandidateID]
				if !exists || !polishNPGroupHeadBound(group, alternative.Slots) || !polishExactDeterminerGroup(group, composite, byID, lattice) {
					continue
				}
				anchorValue, recognized := polishDeterminerDeixis(composite.Determiner.Surface)
				if !recognized {
					continue
				}
				anchor := &meaning.DeicticAnchor{Kind: "deixis", Value: anchorValue}
				key := group.HeadCandidateID + "\x00deixis\x00" + anchorValue
				if !seen[key] {
					evidence.References = append(evidence.References, semantic.ReferenceEvidence{CandidateID: group.HeadCandidateID, Anchor: anchor})
					seen[key] = true
				}
			}
			frame := string(alternative.Permission.FrameID)
			family := alternative.FamilyID.String()
			variant := string(alternative.VariantID)
			sense := alternative.Permission.SenseID.Symbol().Name
			senseID := alternative.Permission.SenseID.String()
			var slotID, anchorValue string
			switch {
			case sense == "pl.byc.class_membership" && frame == "pl.byc.class_membership.frame" && family == "construction:pl.identification" && variant == "construction:pl.identification.to":
				slotID, anchorValue = "member", "unresolved"
			case sense == "pl.byc.class_membership" && frame == "pl.byc.class_membership.frame" && family == "construction:pl.identification_negative" && variant == "construction:pl.identification_negative.to":
				slotID, anchorValue = "member", "unresolved"
			case func() bool {
				_, allowedSense := polishAttributionForSense(senseID)
				return allowedSense && frame == sense+".frame" && family == "construction:pl.attribution" &&
					(variant == "construction:pl.attribution.ten" || variant == "construction:pl.attribution.ten.inanimate" || variant == "construction:pl.attribution.to" || variant == "construction:pl.attribution.ta" ||
						variant == "construction:pl.attribution.tamten" || variant == "construction:pl.attribution.tamten.inanimate" || variant == "construction:pl.attribution.tamto" || variant == "construction:pl.attribution.tamta" || variant == "construction:pl.attribution.bare")
			}():
				slotID, anchorValue = "bearer", "demonstrative"
				if strings.HasSuffix(variant, ".bare") {
					anchorValue = "none"
				}
				if strings.HasPrefix(variant, "construction:pl.attribution.tam") {
					anchorValue = "distal"
				}
			case func() bool {
				_, allowedSense := polishAttributionForSense(senseID)
				return allowedSense && frame == sense+".plural.frame" && family == "construction:pl.attribution_plural" &&
					(variant == "construction:pl.attribution_plural.te" || variant == "construction:pl.attribution_plural.ci" ||
						variant == "construction:pl.attribution_plural.tamte" || variant == "construction:pl.attribution_plural.tamci" || variant == "construction:pl.attribution_plural.bare")
			}():
				slotID, anchorValue = "bearer", "proximal"
				if strings.HasSuffix(variant, ".bare") {
					anchorValue = "none"
				} else if strings.HasPrefix(variant, "construction:pl.attribution_plural.tam") {
					anchorValue = "distal"
				}
			default:
				continue
			}
			for _, slot := range alternative.Slots {
				if slot.SlotID != slotID {
					continue
				}
				for _, candidateID := range slot.CandidateIDs {
					candidate, exists := byID[candidateID]
					if !exists || candidate.Category != construction.CategoryNP {
						continue
					}
					key := candidateID + "\x00deixis\x00" + anchorValue
					if seen[key] {
						continue
					}
					if anchorValue == "none" {
						evidence.References = append(evidence.References, semantic.ReferenceEvidence{CandidateID: candidateID, Unanchored: true})
						continue
					}
					anchor := &meaning.DeicticAnchor{Kind: "deixis", Value: anchorValue}
					evidence.References = append(evidence.References, semantic.ReferenceEvidence{CandidateID: candidateID, Anchor: anchor})
					seen[key] = true
				}
			}
		}
	}
	return evidence
}

func polishNPGroupHeadBound(group parser.NPPhraseGroupBinding, slots []parser.SlotBinding) bool {
	for _, slot := range slots {
		if slot.SlotID == group.HeadSlotID {
			return len(slot.CandidateIDs) == 1 && slot.CandidateIDs[0] == group.HeadCandidateID
		}
	}
	return false
}

// polishExactDeterminerGroup accepts only the exact compiled NP group bound by
// the parser. In particular, the reference attaches to its lexical head, never
// the AP modifier or a different candidate occupying another semantic role.
func polishExactDeterminerGroup(group parser.NPPhraseGroupBinding, composite parser.CompositeNPPhraseCandidate, candidates map[string]parser.PhraseCandidate, lattice analyzer.LexicalLattice) bool {
	if composite.ID != group.CompositeCandidateID || composite.HeadCandidateID != group.HeadCandidateID ||
		composite.ModifierCandidateID != group.ModifierCandidateID || composite.DeterminerCandidateID != group.DeterminerCandidateID ||
		composite.TokenStart != group.TokenStart || composite.TokenEnd != group.TokenEnd || composite.Determiner == nil ||
		composite.Determiner.ID != group.DeterminerCandidateID || composite.Determiner.TokenIndex != group.TokenStart ||
		composite.TokenStart < 0 || composite.TokenEnd > len(lattice.Tokens) || composite.TokenEnd <= composite.TokenStart ||
		composite.Plan.Plan.Language != "pl" || composite.Plan.Plan.Kind != construction.PhraseNP || composite.Plan.Plan.TemplateID != "np-attributive" {
		return false
	}
	head, headExists := candidates[group.HeadCandidateID]
	modifier, modifierExists := candidates[group.ModifierCandidateID]
	if !headExists || !modifierExists || head.Category != construction.CategoryNP || modifier.Category != construction.CategoryAP ||
		composite.Determiner.TokenIndex < 0 || composite.Determiner.TokenIndex >= len(lattice.Tokens) || len(composite.EvidenceCandidateIDs) != 3 ||
		composite.EvidenceCandidateIDs[0] != group.HeadCandidateID || composite.EvidenceCandidateIDs[1] != group.ModifierCandidateID ||
		composite.EvidenceCandidateIDs[2] != group.DeterminerCandidateID {
		return false
	}
	token := lattice.Tokens[composite.Determiner.TokenIndex].Token
	if token.Surface != composite.Determiner.Surface || token.Span != composite.Determiner.Span ||
		composite.TokenStart != composite.Determiner.TokenIndex || composite.Span.Start != composite.Determiner.Span.Start {
		return false
	}
	heads, modifiers, determiners := 0, 0, 0
	for _, input := range composite.Plan.Plan.Inputs {
		switch input.Role {
		case "head":
			heads++
		case "modifier":
			modifiers++
		case "determiner":
			determiners++
		}
	}
	if heads != 1 || modifiers != 1 || determiners != 1 {
		return false
	}
	return true
}

func polishDeterminerDeixis(surface string) (string, bool) {
	switch strings.ToLower(surface) {
	case "tamten", "tamta", "tamto", "tamte", "tamci":
		return "distal", true
	case "ten", "ta", "to", "te", "ci":
		return "demonstrative", true
	default:
		return "", false
	}
}
