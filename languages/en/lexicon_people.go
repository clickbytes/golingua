package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPeopleSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_people.go",
	Version: "2026-10-22-batch-03", License: "unknown", Author: "golingua contributors",
}

// These common person and relationship nouns have distinct English senses.
// Candidate links remain pending and carry no inferred morphology or frames.
var englishPeopleEntries = []lexicalEntry{
	{"en.spouse.n", "spouse", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.spouse.married_partner", "a person joined to another by marriage", "spouse"}}},
	{"en.romantic_partner.n", "romantic partner", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.romantic_partner.relationship", "a person in a continuing romantic relationship with another", "romantic_partner"}}},
	{"en.friend.n", "friend", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.friend.social", "a person connected to another by mutual personal affection or goodwill", "friend"}}},
	{"en.acquaintance.n", "acquaintance", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.acquaintance.social", "a person known to someone but not necessarily a close friend", "acquaintance"}}},
	{"en.neighbor.n", "neighbor", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.neighbor.proximity", "a person who lives near another person or place", "neighbor"}}},
	{"en.stranger.n", "stranger", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stranger.unknown_person", "a person not known personally to the relevant people", "stranger"}}},
	{"en.guest.n", "guest", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.guest.received_person", "a person invited or received by a host", "guest"}}},
	{"en.resident.n", "resident", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.resident.inhabitant", "a person who lives in a particular place", "resident"}}},
	{"en.citizen.n", "citizen", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.citizen.state_member", "a person recognized as a member of a state", "citizen"}}},
	{"en.visitor.n", "visitor", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.visitor.temporary_presence", "a person who visits a place or another person", "visitor"}}},
	{"en.tourist.n", "tourist", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tourist.leisure_travel", "a person traveling to places for interest or pleasure", "tourist"}}},
	{"en.traveler.n", "traveler", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.traveler.journey", "a person undertaking a journey", "traveler"}}},
	{"en.parent.n", "parent", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.parent.kinship", "a person in a parental relationship to a child", "parent"}}},
	{"en.mother.n", "mother", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mother.parent", "a parent identified as a mother in a kinship or caregiving relationship", "mother"}}},
	{"en.father.n", "father", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.father.parent", "a parent identified as a father in a kinship or caregiving relationship", "father"}}},
	{"en.sibling.n", "sibling", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sibling.kinship", "a person sharing at least one parent or a recognized sibling relationship", "sibling"}}},
	{"en.brother.n", "brother", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.brother.sibling", "a sibling identified as a brother", "brother"}}},
	{"en.sister.n", "sister", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sister.sibling", "a sibling identified as a sister", "sister"}}},
	{"en.grandparent.n", "grandparent", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.grandparent.kinship", "a parent of a person's parent", "grandparent"}}},
	{"en.grandchild.n", "grandchild", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.grandchild.kinship", "a child of a person's child", "grandchild"}}},
	{"en.cousin.n", "cousin", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cousin.kinship", "a person related through collateral kinship rather than direct descent or siblinghood", "cousin"}}},
	{"en.offspring.n", "offspring", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.offspring.child", "a person's child, considered in relation to that parent", "child_of_person"}}},
}
