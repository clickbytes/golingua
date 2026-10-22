package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishSocialRelationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_social_relations.go",
	Version: "2026-10-22-batch-31", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name social affiliations and roles between participants.
var englishSocialRelationEntries = []lexicalEntry{
	{"en.social_relation.n", "social relationship", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.social_relation.direct", "a sustained role or affiliation connecting people or organizations", "social_relation"}}},
	{"en.biological_parenthood.n", "biological parenthood", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.biological_parenthood.direct", "a relation in which one organism is a biological parent of another through reproduction", "biological_parenthood"}}},
	{"en.recognized_parenthood.n", "parenthood", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.recognized_parenthood.direct", "a relation in which a person holds recognized parental status toward another", "recognized_parenthood"}}},
	{"en.siblinghood.n", "siblinghood", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.siblinghood.direct", "a relation in which two people share recognized sibling kinship or at least one parent", "siblinghood"}}},
	{"en.marriage_relationship.n", "marriage", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.marriage_relationship.direct", "a relation in which two persons are partners in a recognized marriage institution", "marriage_relationship"}}},
	{"en.friendship_relationship.n", "friendship", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.friendship_relationship.direct", "a mutually familiar, voluntary relationship of goodwill between participants", "friendship_relationship"}}},
	{"en.organization_membership.n", "membership", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.organization_membership.direct", "a relation in which a person or organization belongs to an organized collective", "organization_membership"}}},
	{"en.employment_relationship.n", "employment", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.employment_relationship.direct", "a continuing work arrangement in which a person supplies labor to an employer", "employment_relationship"}}},
	{"en.instruction_relationship.n", "instruction relationship", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.instruction_relationship.direct", "a relation in which one person guides another’s learning in an educator-learner role", "instruction_relationship"}}},
	{"en.customer_provider_relationship.n", "customer relationship", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.customer_provider_relationship.direct", "a commercial relation in which one participant obtains goods or services from another", "customer_provider_relationship"}}},
}
