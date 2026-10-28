package pl

// transferCognitionEntries adds selected pending perception, cognition and
// monetary transfer candidates with explicit lexical senses.
var transferCognitionEntries = []lexicalEntry{
	{"pl.ogladac.v", "oglądać", "verb", []lexicalMeaning{{"pl.ogladac.view", "Directing visual attention toward a person, object, or scene without specifying a display.", "watching"}}},
	{"pl.ogladac_ekran.v", "oglądać ekran", "verb", []lexicalMeaning{{"pl.ogladac_ekran.watch", "Directing sustained visual attention toward content on a display.", "watching_screen"}}},
	{"pl.wyobrazac_sobie.v", "wyobrażać sobie", "verb_reflexive", []lexicalMeaning{{"pl.wyobrazac_sobie.imagine", "Constructing a mental representation not presently perceived as an external fact.", "imagining"}}},
	{"pl.miec_nadzieje.v", "mieć nadzieję", "verb", []lexicalMeaning{{"pl.miec_nadzieje.hope", "Mentally favoring a possible future state while recognizing uncertainty.", "hoping"}}},
	{"pl.dostawac_zaplate.v", "dostawać zapłatę", "verb", []lexicalMeaning{{"pl.dostawac_zaplate.receive", "Acquiring a monetary payment from another participant.", "receiving_payment"}}},
	{"pl.zwracac_pieniadze.v", "zwracać pieniądze", "verb", []lexicalMeaning{{"pl.zwracac_pieniadze.refund", "Returning money paid earlier because of cancellation, return, or correction.", "refunding_payment"}}},
	{"pl.przelewac_pieniadze.v", "przelewać pieniądze", "verb", []lexicalMeaning{{"pl.przelewac_pieniadze.transfer", "Moving funds from one holder or account to another.", "transferring_money"}}},
}
