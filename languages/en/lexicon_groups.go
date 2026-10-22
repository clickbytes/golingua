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
	}
}
