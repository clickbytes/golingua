package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishCommerceVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_commerce_verbs.go",
	Version: "2026-10-22-batch-26", License: "unknown", Author: "golingua contributors",
}

// These pending verb candidates distinguish common exchange and service events.
var englishCommerceVerbEntries = []lexicalEntry{
	{"en.buying.v", "buy", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.buying.event", "acquire goods or services from a seller in return for agreed payment", "buying"}}},
	{"en.selling.v", "sell", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.selling.event", "supply goods or services to a buyer under an agreement for payment", "selling"}}},
	{"en.paying.v", "pay", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.paying.event", "transfer money or another accepted value to settle an obligation or purchase", "paying"}}},
	{"en.receiving_payment.v", "receive", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.receiving_payment.event", "acquire a monetary payment from another participant", "receiving_payment"}}},
	{"en.bargaining.v", "bargain", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.bargaining.event", "negotiate transaction terms with the aim of changing the proposed price", "bargaining"}}},
	{"en.borrowing_object.v", "borrow", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.borrowing_object.event", "temporarily receive use or possession of an item with an obligation to return it", "borrowing_object"}}},
	{"en.lending_object.v", "lend", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.lending_object.event", "temporarily allow another participant to use an item with an expectation of return", "lending_object"}}},
	{"en.returning_borrowed.v", "return", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.returning_borrowed.event", "restore an item to the participant from whom it was borrowed", "returning_borrowed"}}},
	{"en.renting_property.v", "rent", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.renting_property.event", "obtain temporary use of property in exchange for payment", "renting_property"}}},
	{"en.hiring_service.v", "hire", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.hiring_service.event", "engage a provider to perform work or services under an agreement", "hiring_service"}}},
}
