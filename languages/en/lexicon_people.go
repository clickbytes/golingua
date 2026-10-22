package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPeopleSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_people.go",
	Version: "2026-10-22-batch-05", License: "unknown", Author: "golingua contributors",
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
	{"en.infant.n", "infant", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.infant.early_child", "a child at the earliest stage of life who depends on others for basic care", "infant"}}},
	{"en.toddler.n", "toddler", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.toddler.early_walker", "a young child beginning independent walking and exploration", "toddler"}}},
	{"en.adolescent.n", "adolescent", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.adolescent.developmental_stage", "a person undergoing the transition from childhood to adulthood", "adolescent"}}},
	{"en.older_adult.n", "older adult", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.older_adult.later_life", "an adult in a later stage of life without a specific age threshold", "elderly_person"}}},
	{"en.stepparent.n", "stepparent", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stepparent.family_role", "a person recognized as a parent through partnership with a child's parent", "stepparent"}}},
	{"en.adoptive_parent.n", "adoptive parent", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.adoptive_parent.family_role", "a person whose parental status is established through adoption", "adoptive_parent"}}},
	{"en.host_person.n", "host", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.host_person.guest_role", "a person who receives guests or facilitates an event for participants", "host_person"}}},
	{"en.aunt.n", "aunt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.aunt.kinship", "a parent's sister or a woman recognized in an equivalent kinship role", "aunt_uncle"}}},
	{"en.uncle.n", "uncle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.uncle.kinship", "a parent's brother or a man recognized in an equivalent kinship role", "aunt_uncle"}}},
	{"en.passenger.n", "passenger", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.passenger.vehicle_rider", "a person carried by a vehicle without controlling its operation", "passenger"}}},
	{"en.driver.n", "driver", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.driver.road_vehicle", "a person controlling a road vehicle during a journey", "driver"}}},
	{"en.pedestrian.n", "pedestrian", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pedestrian.road_user", "a person traveling on foot in a space shared with traffic", "pedestrian"}}},
	{"en.worker.n", "worker", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.worker.labor", "a person engaged in labor or productive activity", "worker"}}},
	{"en.employee.n", "employee", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.employee.employment", "a person who works under an employment relationship with an employer", "employee"}}},
	{"en.employer.n", "employer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.employer.hiring_party", "a person or organization that hires another to perform work for compensation", "employer"}}},
	{"en.manager.n", "manager", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.manager.supervisor", "a person responsible for coordinating work and supervising resources or staff", "manager"}}},
	{"en.farmer.n", "farmer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.farmer.agriculture", "a person cultivating crops or raising animals for production", "farmer"}}},
	{"en.cook.n", "cook", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cook.food_preparation", "a person who prepares food as a practical or occupational activity", "cook_person"}}},
	{"en.waiter.n", "waiter", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.waiter.food_service", "a person serving food or drinks to customers at a dining establishment", "waiter"}}},
}
