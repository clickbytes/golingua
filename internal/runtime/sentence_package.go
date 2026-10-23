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

const sentenceCapabilityEvidenceID = "sentence.seeing.present.singular.v1"

// BuildSentenceLanguagePackage exposes only the approved exact seeing route
// exercised by the current source and target providers.
func BuildSentenceLanguagePackage(languageID features.LanguageID, version string, senseID ids.SenseID, valencyIndex *valency.Index, lexicalIndex *lexicon.Index) (language.Package, error) {
	if languageID == "" || version == "" || senseID.String() == "" || valencyIndex == nil || lexicalIndex == nil {
		return language.Package{}, fmt.Errorf("sentence package requires language, version, exact sense and compiled indexes")
	}
	var bindings []language.Binding
	for _, route := range valencyIndex.Bindings() {
		if !route.Approved || route.SenseID != senseID {
			continue
		}
		for _, transfer := range lexicalIndex.TransferBindings(senseID) {
			if !transfer.TransferAuthorized() {
				continue
			}
			bindings = append(bindings, language.Binding{
				SenseID: senseID, ValencyFrameID: string(route.ValencyFrameID), FamilyID: route.FamilyID,
				VariantIDs: []ids.ConstructionID{route.VariantID}, CoreRef: transfer.CoreRef,
				Mapping: lexicon.MappingEquivalent, Approved: true, Provenance: transfer.Provenance,
				Analyze: true, Realize: true,
			})
		}
	}
	if len(bindings) == 0 {
		return language.Package{}, fmt.Errorf("no approved exact sentence route for %s", senseID)
	}
	capabilities := []language.Capability{
		{ID: language.CapabilityAnalyze, Implemented: true, Analyze: true, EvidenceCaseIDs: []string{sentenceCapabilityEvidenceID}},
		{ID: language.CapabilityRealize, Implemented: true, Realize: true, EvidenceCaseIDs: []string{sentenceCapabilityEvidenceID}},
	}
	return language.Package{Manifest: language.Manifest{
		SchemaVersion: language.SchemaVersion, Language: languageID, Version: version, Capabilities: capabilities,
	}, Bindings: bindings}, nil
}

// SentenceCapabilityEvidence returns the exact case claimed by the limited
// PL/EN providers.
func SentenceCapabilityEvidence() []language.CapabilityEvidence {
	return []language.CapabilityEvidence{
		{ID: language.CapabilityAnalyze, Analyze: true, CaseIDs: []string{sentenceCapabilityEvidenceID}},
		{ID: language.CapabilityRealize, Realize: true, CaseIDs: []string{sentenceCapabilityEvidenceID}},
	}
}

// SentenceRouteProvenance returns internal provenance for assumptions emitted
// by the limited low-level Meaning realization API.
func SentenceRouteProvenance(languageID features.LanguageID) provenance.SourceRef {
	return provenance.SourceRef{ID: "golingua-sentence-provider", Path: "internal/runtime/compiled_plan_realizer.go", Version: "2026-10-23", License: "internal", Author: string(languageID)}
}
