package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishInformationTermSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_information_terms.go",
	Version: "2026-10-22-batch-30", License: "unknown", Author: "golingua contributors",
}

// These pending nouns distinguish modality, measurement and temporal content.
var englishInformationTermEntries = []lexicalEntry{
	{"en.ability.n", "ability", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ability.direct", "a participant’s capability to perform a specified action under stated conditions", "ability"}}},
	{"en.permission.n", "permission", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.permission.direct", "a normative allowance for a participant to undertake an action", "permission"}}},
	{"en.obligation.n", "obligation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.obligation.direct", "a requirement imposed on a participant to act or ensure an outcome", "obligation"}}},
	{"en.right_entitlement.n", "right", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.right_entitlement.direct", "a recognized claim permitting access, action, or protection against interference", "right_entitlement"}}},
	{"en.responsibility.n", "responsibility", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.responsibility.direct", "accountability or assigned duty for an activity or its outcome", "responsibility"}}},
	{"en.commitment.n", "commitment", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.commitment.direct", "an undertaking to act, maintain a condition, or provide something in the future", "commitment"}}},
	{"en.agreement_content.n", "agreement", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.agreement_content.direct", "a shared set of accepted terms or propositions between participants", "agreement_content"}}},
	{"en.contract_content.n", "contract", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.contract_content.direct", "an agreement whose terms establish enforceable obligations in a given system", "contract_content"}}},
	{"en.category_type.n", "category", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.category_type.direct", "a general class used to group individual referents by shared defining conditions", "category_type"}}},
	{"en.identity_descriptor.n", "identity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.identity_descriptor.direct", "a set of identifying characteristics used to distinguish a referent from others", "identity_descriptor"}}},
	{"en.quantity.n", "quantity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.quantity.direct", "a numerical or countable magnitude of a specified kind", "quantity"}}},
	{"en.number_value.n", "number", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.number_value.direct", "an abstract mathematical number, distinct from grammatical singular or plural", "number_value"}}},
	{"en.measurement_value.n", "measurement", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.measurement_value.direct", "a quantity assigned by comparison against an explicit scale or unit", "measurement_value"}}},
	{"en.measurement_unit.n", "unit", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.measurement_unit.direct", "a conventional reference magnitude used to express measured quantities", "measurement_unit"}}},
	{"en.duration_quantity.n", "duration", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.duration_quantity.direct", "an amount of elapsed time between temporal boundaries", "duration_quantity"}}},
	{"en.time_instant.n", "instant", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.time_instant.direct", "a point on a temporal ordering without intrinsic duration", "time_instant"}}},
	{"en.time_interval.n", "interval", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.time_interval.direct", "a bounded span of time with beginning and ending conditions", "time_interval"}}},
	{"en.price_amount.n", "price", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.price_amount.direct", "monetary value requested or accepted for a specified good or service", "price_amount"}}},
	{"en.monetary_amount.n", "amount of money", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.monetary_amount.direct", "an amount of currency considered independently of a particular transaction", "monetary_amount"}}},
	{"en.probability_value.n", "probability", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.probability_value.direct", "a measure representing likelihood over a specified space of possible outcomes", "probability_value"}}},
}
