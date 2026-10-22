package pl

// lexicalGroups lists the domain-declared Polish lexical group tables in
// stable editorial order. Each table is authored data copied from the review
// ledger in data/pl/1000-core-review.tsv; the group boundaries follow the
// ledger domains. Nothing here is generated at runtime.
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
	}
}
