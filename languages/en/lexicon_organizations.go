package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishOrganizationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_organizations.go",
	Version: "2026-10-22-batch-10", License: "unknown", Author: "golingua contributors",
}

// These English nouns identify coordinated groups and commercial organizations.
// Candidate links remain pending and have no inferred morphology or frames.
var englishOrganizationEntries = []lexicalEntry{
	{"en.team_group.n", "team", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.team_group.coordinated_unit", "people cooperating as a coordinated unit toward a goal", "team_group"}}},
	{"en.class_group.n", "class", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.class_group.learners", "learners organized into one instructional group", "class_group"}}},
	{"en.committee.n", "committee", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.committee.deliberative", "a formally selected group assigned a deliberative or administrative task", "committee"}}},
	{"en.organization.n", "organization", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.organization.structured_collective", "a continuing structured collective with members or governing procedures and a purpose", "organization"}}},
	{"en.company.n", "company", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.company.commercial", "an organization established to carry on commercial activity", "company"}}},
	{"en.retail_business.n", "retail business", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.retail_business.goods_sales", "a commercial organization whose activity includes selling goods to end customers", "retail_business"}}},
	{"en.restaurant_business.n", "restaurant business", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.restaurant_business.meals", "a business primarily preparing and serving meals to customers", "restaurant_business"}}},
	{"en.transport_company.n", "transport company", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.transport_company.service", "a business providing movement of passengers or goods as a service", "transport_business"}}},
	{"en.association.n", "association", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.association.shared_purpose", "an organization formed by members for a shared purpose", "association"}}},
	{"en.club_organization.n", "club", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.club_organization.shared_activity", "an association of participants who meet around a shared activity or interest", "club_organization"}}},
}
