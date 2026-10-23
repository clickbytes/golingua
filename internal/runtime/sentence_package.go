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
	sentenceSeeingCapabilityEvidenceID    = "sentence.seeing.present.singular.v1"
	sentenceHearingCapabilityEvidenceID   = "sentence.hearing.present.singular.v1"
	sentenceReceivingCapabilityEvidenceID = "sentence.receiving.past.singular.v1"
	sentenceLookingCapabilityEvidenceID   = "sentence.looking.present.singular.v1"
)

// BuildSentenceLanguagePackage exposes only requested exact approved sentence
// routes exercised by the current source and target providers.
func BuildSentenceLanguagePackage(languageID features.LanguageID, version string, valencyIndex *valency.Index, lexicalIndex *lexicon.Index, senseIDs ...ids.SenseID) (language.Package, error) {
	if languageID == "" || version == "" || len(senseIDs) == 0 || valencyIndex == nil || lexicalIndex == nil {
		return language.Package{}, fmt.Errorf("sentence package requires language, version, exact senses and compiled indexes")
	}
	allowedSenses := make(map[ids.SenseID]bool, len(senseIDs))
	for _, senseID := range senseIDs {
		if senseID.String() == "" {
			return language.Package{}, fmt.Errorf("sentence package requires non-empty exact senses")
		}
		allowedSenses[senseID] = true
	}
	expectedSenseNames, ok := currentSentenceSenseNames(languageID)
	if !ok || len(allowedSenses) != len(expectedSenseNames) {
		return language.Package{}, fmt.Errorf("sentence package currently requires the exact seeing, hearing, receiving, and looking routes for Polish or English")
	}
	for _, name := range expectedSenseNames {
		expectedSense, err := ids.NewSenseID(name)
		if err != nil {
			return language.Package{}, err
		}
		if !allowedSenses[expectedSense] {
			return language.Package{}, fmt.Errorf("sentence package requires exact route %s", expectedSense)
		}
	}
	var bindings []language.Binding
	boundSenses := make(map[ids.SenseID]bool, len(allowedSenses))
	for _, route := range valencyIndex.Bindings() {
		if !route.Approved || !allowedSenses[route.SenseID] {
			continue
		}
		for _, transfer := range lexicalIndex.TransferBindings(route.SenseID) {
			if !transfer.TransferAuthorized() {
				continue
			}
			bindings = append(bindings, language.Binding{
				SenseID: route.SenseID, ValencyFrameID: string(route.ValencyFrameID), FamilyID: route.FamilyID,
				VariantIDs: []ids.ConstructionID{route.VariantID}, CoreRef: transfer.CoreRef,
				Mapping: lexicon.MappingEquivalent, Approved: true, Provenance: transfer.Provenance,
				Analyze: true, Realize: true,
			})
			boundSenses[route.SenseID] = true
		}
	}
	for _, senseID := range senseIDs {
		if !boundSenses[senseID] {
			return language.Package{}, fmt.Errorf("no approved exact sentence route for %s", senseID)
		}
	}
	capabilities := []language.Capability{
		{ID: language.CapabilityAnalyze, Implemented: true, Analyze: true, EvidenceCaseIDs: sentenceCapabilityEvidenceIDs()},
		{ID: language.CapabilityRealize, Implemented: true, Realize: true, EvidenceCaseIDs: sentenceCapabilityEvidenceIDs()},
	}
	return language.Package{Manifest: language.Manifest{
		SchemaVersion: language.SchemaVersion, Language: languageID, Version: version, Capabilities: capabilities,
	}, Bindings: bindings}, nil
}

// SentenceCapabilityEvidence returns the exact case claimed by the limited
// PL/EN providers.
func SentenceCapabilityEvidence() []language.CapabilityEvidence {
	return []language.CapabilityEvidence{
		{ID: language.CapabilityAnalyze, Analyze: true, CaseIDs: sentenceCapabilityEvidenceIDs()},
		{ID: language.CapabilityRealize, Realize: true, CaseIDs: sentenceCapabilityEvidenceIDs()},
	}
}

func sentenceCapabilityEvidenceIDs() []string {
	return []string{sentenceSeeingCapabilityEvidenceID, sentenceHearingCapabilityEvidenceID, sentenceReceivingCapabilityEvidenceID, sentenceLookingCapabilityEvidenceID}
}

func currentSentenceSenseNames(languageID features.LanguageID) ([]string, bool) {
	switch languageID {
	case "en":
		return []string{"en.see.visual", "en.hear.auditory", "en.get.receive", "en.look.directed"}, true
	case "pl":
		return []string{"pl.widziec.visual", "pl.slyszec.auditory", "pl.dostac.receive", "pl.patrzec.directed"}, true
	default:
		return nil, false
	}
}

// SentenceRouteProvenance returns internal provenance for assumptions emitted
// by the limited low-level Meaning realization API.
func SentenceRouteProvenance(languageID features.LanguageID) provenance.SourceRef {
	return provenance.SourceRef{ID: "golingua-sentence-provider", Path: "internal/runtime/compiled_plan_realizer.go", Version: "2026-10-23", License: "internal", Author: string(languageID)}
}
