package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishInstitutionSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_institutions.go",
	Version: "2026-10-22-batch-11", License: "unknown", Author: "golingua contributors",
}

// These English nouns identify institutions; their candidate links do not
// infer separate building, service, or government-body senses.
var englishInstitutionEntries = []lexicalEntry{
	{"en.bank_institution.n", "bank", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bank_institution.financial_services", "an institution providing regulated deposit, credit, or other financial services", "bank_institution"}}},
	{"en.government.n", "government", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.government.public_authority", "an institution exercising public administrative authority over a jurisdiction", "government"}}},
	{"en.university_institution.n", "university", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.university_institution.higher_education", "an institution providing higher education and often conducting research", "university_institution"}}},
	{"en.hospital_institution.n", "hospital", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hospital_institution.medical_services", "an organization providing inpatient or specialist medical services", "hospital_institution"}}},
	{"en.library_institution.n", "library", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.library_institution.information_resources", "an organization maintaining and providing access to curated information resources", "library_institution"}}},
	{"en.police_department.n", "police department", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.police_department.jurisdiction", "an institution responsible for organized policing within a jurisdiction", "police_department"}}},
}
