package en

import "github.com/clickbytes/golingua/internal/provenance"

type lexicalGroup struct {
	source  provenance.SourceRef
	entries []lexicalEntry
}

// englishLexiconGroups gathers language-local domains in a stable order.
func englishLexiconGroups() []lexicalGroup {
	return []lexicalGroup{
		{source: englishLexicalSource, entries: englishEntries},
		{source: englishNounSource, entries: englishNounEntries},
		{source: englishPeopleSource, entries: englishPeopleEntries},
		{source: englishSocialGroupsSource, entries: englishSocialGroupEntries},
		{source: englishOrganizationSource, entries: englishOrganizationEntries},
		{source: englishInstitutionSource, entries: englishInstitutionEntries},
		{source: englishCoreTermsSource, entries: englishCoreTermEntries},
		{source: englishAnimalSource, entries: englishAnimalEntries},
		{source: englishPlantSource, entries: englishPlantEntries},
		{source: englishBodySource, entries: englishBodyEntries},
		{source: englishMaterialSource, entries: englishMaterialEntries},
		{source: englishTransportSource, entries: englishTransportEntries},
		{source: englishObjectSource, entries: englishObjectEntries},
		{source: englishDeviceSource, entries: englishDeviceEntries},
		{source: englishFoodSource, entries: englishFoodEntries},
		{source: englishFoodMoreSource, entries: englishFoodMoreEntries},
		{source: englishClothingSource, entries: englishClothingEntries},
		{source: englishSettlementSource, entries: englishSettlementEntries},
		{source: englishPlaceSource, entries: englishPlaceEntries},
		{source: englishLandformSource, entries: englishLandformEntries},
		{source: englishLandformMoreSource, entries: englishLandformMoreEntries},
		{source: englishDocumentSource, entries: englishDocumentEntries},
		{source: englishQualitySource, entries: englishQualityEntries},
		{source: englishEmotionSource, entries: englishEmotionEntries},
		{source: englishSpeechVerbSource, entries: englishSpeechVerbEntries},
		{source: englishManipulationVerbSource, entries: englishManipulationVerbEntries},
		{source: englishCommerceVerbSource, entries: englishCommerceVerbEntries},
		{source: englishExchangeVerbSource, entries: englishExchangeVerbEntries},
		{source: englishMotivationVerbSource, entries: englishMotivationVerbEntries},
		{source: englishPhysiologyVerbSource, entries: englishPhysiologyVerbEntries},
	}
}
