package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishExchangeVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_exchange_verbs.go",
	Version: "2026-10-22-batch-27", License: "unknown", Author: "golingua contributors",
}

// These pending candidates distinguish exchange, search, payment and ownership events.
var englishExchangeVerbEntries = []lexicalEntry{
	{"en.making_donation.v", "donate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.making_donation.event", "give resources voluntarily to support a cause or recipient without equivalent compensation", "making_donation"}}},
	{"en.gifting.v", "gift", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.gifting.event", "transfer something to a recipient as a gift without expected payment", "gifting"}}},
	{"en.stealing.v", "steal", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.stealing.event", "take control of an item without the entitled possessor’s consent", "stealing"}}},
	{"en.losing_possession.v", "lose", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.losing_possession.event", "cease to know or control where a possessed item is, without intentional transfer", "losing_possession"}}},
	{"en.finding_item.v", "find", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.finding_item.event", "discover the location of an item previously unknown or missing", "finding_item"}}},
	{"en.searching_for_item.v", "search", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.searching_for_item.event", "actively examine places to locate a desired physical object", "searching_for_item"}}},
	{"en.ordering_goods.v", "order", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.ordering_goods.event", "submit a request for specified goods or services to a supplier", "ordering_goods"}}},
	{"en.delivering_goods.v", "deliver", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.delivering_goods.event", "transport an item to an intended recipient and make it available to them", "delivering_goods"}}},
	{"en.shipping_goods.v", "ship", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.shipping_goods.event", "dispatch goods through a transport system for delivery elsewhere", "shipping_goods"}}},
	{"en.exchanging_goods.v", "exchange", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.exchanging_goods.event", "transfer goods or value in both directions as one transaction", "exchanging_goods"}}},
	{"en.sharing_resource.v", "share", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.sharing_resource.event", "allow multiple participants to access or use a resource under an arrangement", "sharing_resource"}}},
	{"en.reserving_resource.v", "reserve", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.reserving_resource.event", "arrange in advance for a resource or service to be available on an occasion", "reserving_resource"}}},
	{"en.canceling_reservation.v", "cancel", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.canceling_reservation.event", "revoke a previous arrangement securing future access to a resource or service", "canceling_reservation"}}},
	{"en.refunding_payment.v", "refund", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.refunding_payment.event", "return money paid earlier because of cancellation, return, or correction", "refunding_payment"}}},
	{"en.saving_money.v", "save", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.saving_money.event", "retain monetary resources instead of using them for immediate spending", "saving_money"}}},
	{"en.spending_money.v", "spend", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.spending_money.event", "use monetary value to obtain goods or services or fulfill obligations", "spending_money"}}},
	{"en.earning_money.v", "earn", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.earning_money.event", "receive monetary compensation for labor, investment, or another entitled activity", "earning_money"}}},
	{"en.transferring_money.v", "transfer", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.transferring_money.event", "move funds from one holder or account to another", "transferring_money"}}},
	{"en.inheriting_property.v", "inherit", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.inheriting_property.event", "receive an asset under inheritance rules because of a predecessor’s death", "inheriting_property"}}},
}
