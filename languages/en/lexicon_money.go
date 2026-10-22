package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishMoneySource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_money.go",
	Version: "2026-10-22-batch-31", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name money, payment records and travel credentials.
var englishMoneyEntries = []lexicalEntry{
	{"en.money.n", "money", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.money.direct", "a socially accepted medium used to pay, measure value, or store purchasing power", "money"}}},
	{"en.coin.n", "coin", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.coin.direct", "a minted piece of money, ordinarily metal, bearing a recognized denomination", "coin"}}},
	{"en.banknote.n", "banknote", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.banknote.direct", "a printed or polymer note issued as a unit of currency", "banknote"}}},
	{"en.bank_account.n", "bank account", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bank_account.direct", "an institutional record identifying monetary funds held for an account owner", "bank_account"}}},
	{"en.credit_card.n", "credit card", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.credit_card.direct", "a payment card enabling purchases using an authorized credit facility", "credit_card"}}},
	{"en.debit_card.n", "debit card", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.debit_card.direct", "a payment card used to pay from a linked bank account’s available funds", "debit_card"}}},
	{"en.receipt.n", "receipt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.receipt.direct", "a record acknowledging that goods, money, or a service were received", "receipt"}}},
	{"en.ticket.n", "ticket", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ticket.direct", "a credential granting access to a journey, venue, or specified service", "ticket"}}},
	{"en.passport.n", "passport", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.passport.direct", "an official travel identity document issued by a national authority", "passport"}}},
	{"en.identity_card.n", "identity card", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.identity_card.direct", "an official or institutional card identifying its bearer", "identity_card"}}},
	{"en.invoice.n", "invoice", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.invoice.direct", "a document requesting payment for specified supplied goods or services", "invoice"}}},
	{"en.price_tag.n", "price tag", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.price_tag.direct", "a label displaying the proposed price of a particular article", "price_tag"}}},
	{"en.cash_register.n", "cash register", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cash_register.direct", "a machine recording retail transactions and often storing cash", "cash_register"}}},
}
