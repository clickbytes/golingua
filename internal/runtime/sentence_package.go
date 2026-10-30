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
		return language.Package{}, fmt.Errorf("sentence package requires the exact declared routes for language %s", languageID)
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

func currentSentenceSenseNames(languageID features.LanguageID) ([]string, bool) {
	switch languageID {
	case "en":
		return []string{"en.see.visual", "en.hear.auditory", "en.get.receive", "en.look.directed", "en.read.text", "en.be.class_membership", "en.black.color", "en.red.color", "en.white.color", "en.blue.color", "en.green.color", "en.open.state", "en.sad.state", "en.eating.event", "en.drinking.event", "en.buying.event", "en.selling.event", "en.opening_object.event", "en.closing_object.event", "en.carrying.event", "en.holding_object.event", "en.pulling_object.event", "en.pushing_object.event", "en.cutting_object.event", "en.cutting_object.with_instrument", "en.folding_object.event", "en.tearing_object.event", "en.sleep.physiological", "en.breathing.event", "en.sneezing.event", "en.running.event", "en.coughing.event", "en.walking.event", "en.living.event", "en.flying_motion.event", "en.waking_up.event", "en.falling_asleep.event", "en.working.activity", "en.go.spatial_movement", "en.give.transfer", "en.helping.event", "en.thanking.event", "en.using_object.event", "en.liking.event", "en.watching.event", "en.writing_document.event", "en.writing_message.event", "en.writing_with_instrument.event", "en.conversing.event"}, true
	case "pl":
		return []string{"pl.widziec.visual", "pl.slyszec.auditory", "pl.dostac.receive", "pl.patrzec.directed", "pl.czytac.interpret", "pl.byc.class_membership", "pl.czarny.black", "pl.czerwony.red", "pl.bialy.white", "pl.niebieski.blue", "pl.zielony.green", "pl.otwarty.open", "pl.smutny.sad", "pl.jesc.eat", "pl.pic.drink", "pl.kupowac.buy", "pl.kupic.buy", "pl.sprzedawac.sell", "pl.otwierac_obiekt.open", "pl.zamykac_obiekt.close", "pl.niesc.carry", "pl.trzymac.hold", "pl.ciagnac.pull", "pl.pchac.push", "pl.ciac.cut", "pl.kroic.with_instrument", "pl.skladac.fold", "pl.rozdzierac.tear", "pl.spac.sleep", "pl.oddychac.breathe", "pl.kichac.sneeze", "pl.biec.run", "pl.kaszlec.cough", "pl.chodzic.walk", "pl.zyc.live", "pl.latac.fly", "pl.zasypiac.sleep", "pl.budzic_sie.wake", "pl.pracowac.activity", "pl.isc.move", "pl.dawac.transfer", "pl.dac.transfer", "pl.pomagac.help", "pl.dziekowac.thank", "pl.uzywac.using", "pl.lubic.like", "pl.ogladac.view", "pl.pisac.document", "pl.pisac.write", "pl.pisac.instrument", "pl.rozmawiac.converse"}, true
	default:
		return nil, false
	}
}

// SentenceRouteProvenance returns internal provenance for assumptions emitted
// by the limited low-level Meaning realization API.
func SentenceRouteProvenance(languageID features.LanguageID) provenance.SourceRef {
	return provenance.SourceRef{ID: "golingua-sentence-provider", Path: "internal/runtime/compiled_plan_realizer.go", Version: "2026-10-23", License: "internal", Author: string(languageID)}
}
