package runtime

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	sentenceSeeingCapabilityEvidenceID       = "sentence.seeing.present.singular.v1"
	sentenceSeeingNegationEvidenceID         = "sentence.seeing.negation.present.singular.v1"
	sentencePresentNegationEvidenceID        = "sentence.present.negation.transitive_oblique.singular_plural.v1"
	sentenceHearingCapabilityEvidenceID      = "sentence.hearing.present.singular.v1"
	sentenceReceivingCapabilityEvidenceID    = "sentence.receiving.past.singular.v1"
	sentenceLookingCapabilityEvidenceID      = "sentence.looking.present.singular.v1"
	sentenceReadingCapabilityEvidenceID      = "sentence.reading.present.singular_plural.v1"
	sentenceReadingPastCapabilityEvidenceID  = "sentence.reading.past.third_person.v1"
	sentenceIdentificationClassEvidenceID    = "A1-ID-002-PL-EN"
	sentenceIdentificationBlackEvidenceID    = "A1-ID-004-PL-EN"
	englishIdentificationClassEvidenceID     = "A1-ID-001-EN-PL"
	englishIdentificationBlackEvidenceID     = "A1-ID-003-EN-PL"
	sentencePolishNegativeClassEvidenceID    = "A1-NEG-002-PL-EN"
	sentencePolishNegativeSeeingEvidenceID   = "A1-NEG-004-PL-EN"
	sentenceEnglishNegativeClassEvidenceID   = "A1-NEG-001-EN-PL"
	sentenceEnglishNegativeSeeingEvidenceID  = "A1-NEG-003-EN-PL"
	sentencePolishBuyingOmissionEvidenceID   = "A1.4-AC-472-PL-EN:analyze"
	sentencePolishSellingOmissionEvidenceID  = "A1.4-AC-476-PL-EN:analyze"
	sentenceEnglishBuyingOmissionEvidenceID  = "A1.4-AC-471-EN-PL:analyze"
	sentenceEnglishSellingOmissionEvidenceID = "A1.4-AC-475-EN-PL:analyze"
)

// BuildSentenceLanguagePackage exposes only requested exact approved sentence
// routes backed by the compiled valency and lexical indexes.
func BuildSentenceLanguagePackage(languageID features.LanguageID, version string, valencyIndex *valency.Index, lexicalIndex *lexicon.Index, senseIDs ...ids.SenseID) (language.Package, error) {
	if (languageID != "en" && languageID != "pl") || version == "" || len(senseIDs) == 0 || valencyIndex == nil || lexicalIndex == nil {
		return language.Package{}, fmt.Errorf("sentence package requires language, version, exact senses and compiled indexes")
	}
	allowedSenses := make(map[ids.SenseID]bool, len(senseIDs))
	for _, senseID := range senseIDs {
		if senseID.String() == "" {
			return language.Package{}, fmt.Errorf("sentence package requires non-empty exact senses")
		}
		if allowedSenses[senseID] {
			return language.Package{}, fmt.Errorf("sentence package repeats exact sense %s", senseID)
		}
		allowedSenses[senseID] = true
	}
	var bindings []language.Binding
	bindingIndexes := make(map[string]int)
	boundSenses := make(map[ids.SenseID]bool, len(allowedSenses))
	for _, route := range valencyIndex.Bindings() {
		if !route.Approved || !allowedSenses[route.SenseID] {
			continue
		}
		for _, transfer := range lexicalIndex.TransferBindings(route.SenseID) {
			if !transfer.TransferAuthorized() {
				continue
			}
			key := route.SenseID.String() + "\x00" + string(route.ValencyFrameID) + "\x00" + route.FamilyID.String()
			if index, exists := bindingIndexes[key]; exists {
				binding := &bindings[index]
				if binding.CoreRef != transfer.CoreRef || binding.Provenance != transfer.Provenance {
					return language.Package{}, fmt.Errorf("sentence route variants for %s disagree on their exact lexical transfer", route.SenseID)
				}
				found := false
				for _, variantID := range binding.VariantIDs {
					if variantID == route.VariantID {
						found = true
						break
					}
				}
				if !found {
					binding.VariantIDs = append(binding.VariantIDs, route.VariantID)
				}
			} else {
				bindingIndexes[key] = len(bindings)
				bindings = append(bindings, language.Binding{
					SenseID: route.SenseID, ValencyFrameID: string(route.ValencyFrameID), FamilyID: route.FamilyID,
					VariantIDs: []ids.ConstructionID{route.VariantID}, CoreRef: transfer.CoreRef,
					Mapping: lexicon.MappingEquivalent, Approved: true, Provenance: transfer.Provenance,
					Analyze: true, Realize: true,
				})
			}
			boundSenses[route.SenseID] = true
		}
	}
	for _, senseID := range senseIDs {
		if !boundSenses[senseID] {
			return language.Package{}, fmt.Errorf("no approved exact sentence route for %s", senseID)
		}
	}
	capabilities := []language.Capability{
		{ID: language.CapabilityAnalyze, Implemented: true, Analyze: true, EvidenceCaseIDs: sentenceCapabilityEvidenceIDs(languageID)},
		{ID: language.CapabilityRealize, Implemented: true, Realize: true, EvidenceCaseIDs: sentenceCapabilityEvidenceIDs(languageID)},
	}
	return language.Package{Manifest: language.Manifest{
		SchemaVersion: language.SchemaVersion, Language: languageID, Version: version, Capabilities: capabilities,
	}, Bindings: bindings}, nil
}

// SentenceCapabilityEvidence returns the exact case claimed by the limited
// PL/EN providers.
func SentenceCapabilityEvidence(languageID features.LanguageID) []language.CapabilityEvidence {
	return []language.CapabilityEvidence{
		{ID: language.CapabilityAnalyze, Analyze: true, CaseIDs: sentenceCapabilityEvidenceIDs(languageID)},
		{ID: language.CapabilityRealize, Realize: true, CaseIDs: sentenceCapabilityEvidenceIDs(languageID)},
	}
}

func sentenceCapabilityEvidenceIDs(languageID features.LanguageID) []string {
	caseIDs := []string{sentenceSeeingCapabilityEvidenceID, sentenceSeeingNegationEvidenceID, sentencePresentNegationEvidenceID, sentenceHearingCapabilityEvidenceID, sentenceReceivingCapabilityEvidenceID, sentenceLookingCapabilityEvidenceID, sentenceReadingCapabilityEvidenceID, sentenceReadingPastCapabilityEvidenceID}
	if languageID == "pl" {
		caseIDs = append(caseIDs, sentenceIdentificationClassEvidenceID, sentenceIdentificationBlackEvidenceID, sentencePolishNegativeClassEvidenceID, sentencePolishNegativeSeeingEvidenceID, sentencePolishBuyingOmissionEvidenceID, sentencePolishSellingOmissionEvidenceID)
	} else if languageID == "en" {
		caseIDs = append(caseIDs, englishIdentificationClassEvidenceID, englishIdentificationBlackEvidenceID, sentenceEnglishNegativeClassEvidenceID, sentenceEnglishNegativeSeeingEvidenceID, sentenceEnglishBuyingOmissionEvidenceID, sentenceEnglishSellingOmissionEvidenceID)
	}
	return caseIDs
}

// SentenceRouteProvenance returns internal provenance for assumptions emitted
// by the limited low-level Meaning realization API.
func SentenceRouteProvenance(languageID features.LanguageID) provenance.SourceRef {
	return provenance.SourceRef{ID: "golingua-sentence-provider", Path: "internal/runtime/compiled_plan_realizer.go", Version: "2026-10-23", License: "internal", Author: string(languageID)}
}
