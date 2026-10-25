package pl

// lexicalGroups lists the domain-declared Polish lexical group tables in
// stable editorial order. Review-ledger groups are authored data copied from
// data/pl/1000-core-review.tsv; the final morphology-only group has no senses
// or semantic bindings and only makes sourced form data addressable by lexeme.
func lexicalGroups() [][]lexicalEntry {
	return [][]lexicalEntry{
		peopleEntries,
		natureEntries,
		bodyEntries,
		placesEntries,
		objectsEntries,
		foodEntries,
		coreVerbsEntries,
		basicsEntries,
		qualitiesEntries,
		abstractsEntries,
		relationsEntries,
		transportEntries,
		mediaEntries,
		suppliesEntries,
		institutionFoodEntries,
		motionEntries,
		physicalActionEntries,
		activityActionEntries,
		productionActionEntries,
		lifeActionEntries,
		socialActionEntries,
		communicationActionEntries,
		communicationMoreEntries,
		abstractActionEntries,
		remainingActionEntries,
		phraseLevelActionEntries,
		transferCognitionEntries,
		remainingActionEntries2,
		exchangeActionEntries,
		healthRemainingEntries,
		eventNounEntries,
		cognitionActionEntries,
		multiRoleActionEntries,
		communicationLastEntries,
		outcomeActionEntries,
		peoplePhraseEntries,
		kinshipVariantEntries,
		toddlerEntries,
		coreCategoryEntries,
		polishPersonalEntries,
		polishMorphologyOnlyEntries,
	}
}

var polishMorphologyOnlyEntries = []lexicalEntry{
	{"pl.mewa.n", "mewa", "noun", nil},
}
