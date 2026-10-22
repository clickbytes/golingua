package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishClothingSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_clothing.go",
	Version: "2026-10-22-batch-21", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name common garments and clothing.
var englishClothingEntries = []lexicalEntry{
	{"en.clothing.n", "clothing", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.clothing.direct", "wearable textile or comparable articles intended to cover or protect the body", "clothing"}}},
	{"en.shirt.n", "shirt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shirt.direct", "a garment for the upper torso, generally with sleeves and a neck opening", "shirt"}}},
	{"en.t_shirt.n", "T-shirt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.t_shirt.direct", "a casual soft-knit upper-body shirt commonly without a collar", "t_shirt"}}},
	{"en.sweater.n", "sweater", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sweater.direct", "a knitted warm garment worn on the upper body", "sweater"}}},
	{"en.jacket.n", "jacket", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.jacket.direct", "an outer garment covering the upper body, usually shorter than a coat", "jacket"}}},
	{"en.coat.n", "coat", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.coat.direct", "an outer garment generally extending below the hip for warmth or protection", "coat"}}},
	{"en.dress_garment.n", "dress", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.dress_garment.direct", "a one-piece garment covering the torso and extending as a skirt", "dress_garment"}}},
	{"en.skirt_garment.n", "skirt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.skirt_garment.direct", "a garment hanging from the waist and covering part of the legs without separate leg tubes", "skirt_garment"}}},
	{"en.trousers.n", "trousers", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.trousers.direct", "a bifurcated garment covering each leg separately from the waist downward", "trousers"}}},
	{"en.shorts_garment.n", "shorts", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shorts_garment.direct", "short trousers ending above or near the knee", "shorts_garment"}}},
	{"en.sock.n", "sock", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sock.direct", "a knitted or woven covering worn over a foot, sometimes extending up the leg", "sock"}}},
	{"en.shoe.n", "shoe", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shoe.direct", "footwear with a sole and upper designed to protect a foot during movement", "shoe"}}},
	{"en.boot.n", "boot", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.boot.direct", "footwear extending above the ankle for greater coverage or protection", "boot"}}},
	{"en.hat.n", "hat", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hat.direct", "a head covering worn for protection, warmth, or style", "hat"}}},
	{"en.glove.n", "glove", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.glove.direct", "a hand covering with enclosed fingers or finger sections", "glove"}}},
	{"en.scarf.n", "scarf", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.scarf.direct", "a long fabric strip worn around the neck or head", "scarf"}}},
	{"en.belt_garment.n", "belt", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.belt_garment.direct", "a flexible band worn around the waist to support clothing or as an accessory", "belt_garment"}}},
	{"en.umbrella.n", "umbrella", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.umbrella.direct", "a collapsible canopy carried for protection from rain or sunlight", "umbrella"}}},
	{"en.eyeglasses.n", "eyeglasses", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.eyeglasses.direct", "a wearable optical frame with lenses to correct vision or protect eyes", "eyeglasses"}}},
	{"en.wallet.n", "wallet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.wallet.direct", "a small portable holder used for money, cards, or documents", "wallet"}}},
}
