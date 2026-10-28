package pl

// mediaEntries is the batch-7 Polish documents/media/money-paper group.
// Every entry is an authored editorial decision recorded in
// data/pl/1000-core-review.tsv; senses link to CORE as pending candidates
// without transfer authority. No paradigm is asserted for new lexemes yet.
var mediaEntries = []lexicalEntry{
	{"pl.gazeta.n", "gazeta", "noun", []lexicalMeaning{{"pl.gazeta.publication", "A periodically issued printed publication containing news, commentary, or other articles.", "newspaper"}}},
	{"pl.dokument.n", "dokument", "noun", []lexicalMeaning{{"pl.dokument.artifact", "An intentionally recorded information artifact with an identifiable continuing content and form", "document"}}},
	{"pl.wiadomosc.n", "wiadomość", "noun", []lexicalMeaning{{"pl.wiadomosc.artifact", "A discrete communicated or recorded unit intended to convey information to a recipient", "message"}}},
	{"pl.list.n", "list", "noun", []lexicalMeaning{{"pl.list.artifact", "A written document composed as correspondence to a specific recipient", "letter_document"}}},
	{"pl.e_mail.n", "e-mail", "noun", []lexicalMeaning{{"pl.e_mail.artifact", "An electronically transmitted message addressed through an email system", "email_message"}}},
	{"pl.sms.n", "SMS", "noun", []lexicalMeaning{{"pl.sms.artifact", "A short written message exchanged through a messaging network or service", "text_message"}}},
	{"pl.notatnik.n", "notatnik", "noun", []lexicalMeaning{{"pl.notatnik.artifact", "A bound or organized set of blank or used pages intended for notes", "notebook"}}},
	{"pl.kartka_papieru.n", "kartka papieru", "noun", []lexicalMeaning{{"pl.kartka_papieru.artifact", "A single thin piece of paper made for writing, printing, or wrapping", "paper_sheet"}}},
	{"pl.pisak.n", "pisak", "noun", []lexicalMeaning{{"pl.pisak.artifact", "A handheld writing instrument delivering ink to a surface", "pen"}}},
	{"pl.olowek.n", "ołówek", "noun", []lexicalMeaning{{"pl.olowek.artifact", "A handheld writing instrument depositing graphite or similar material", "pencil"}}},
	{"pl.mapa.n", "mapa", "noun", []lexicalMeaning{{"pl.mapa.artifact", "A spatial representation showing locations, routes, or boundaries", "map_document"}}},
	{"pl.kalendarz.n", "kalendarz", "noun", []lexicalMeaning{{"pl.kalendarz.artifact", "A physical or digital display organizing dates and scheduled occasions", "calendar_object"}}},
	{"pl.fotografia.n", "fotografia", "noun", []lexicalMeaning{{"pl.fotografia.artifact", "A still visual image produced by recording light from a scene", "photograph"}}},
	{"pl.rysunek.n", "rysunek", "noun", []lexicalMeaning{{"pl.rysunek.artifact", "A visual representation produced mainly through lines or marks", "drawing_artifact"}}},
	{"pl.nagranie.n", "nagranie", "noun", []lexicalMeaning{{"pl.nagranie.artifact", "A durable representation of audible musical performance or composition", "music_recording"}}},
	{"pl.plik.n", "plik", "noun", []lexicalMeaning{{"pl.plik.artifact", "An addressable stored unit of digital data distinct from the information it represents", "digital_file"}}},
	{"pl.strona_internetowa.n", "strona internetowa", "noun", []lexicalMeaning{{"pl.strona_internetowa.artifact", "A persistent collection of related web resources identified by a site address", "website"}}},
	{"pl.aplikacja.n", "aplikacja", "noun", []lexicalMeaning{{"pl.aplikacja.artifact", "A program product intended to support an end user's tasks", "application_software"}}},
	{"pl.pieniadze.n", "pieniądze", "noun", []lexicalMeaning{{"pl.pieniadze.medium", "A socially accepted medium used to pay, measure value, or store purchasing power", "money"}}},
	{"pl.moneta.n", "moneta", "noun", []lexicalMeaning{{"pl.moneta.medium", "A minted piece of money, ordinarily metal, bearing a recognized denomination", "coin"}}},
	{"pl.banknot.n", "banknot", "noun", []lexicalMeaning{{"pl.banknot.medium", "A printed or polymer note issued as a unit of currency", "banknote"}}},
	{"pl.konto_bankowe.n", "konto bankowe", "noun", []lexicalMeaning{{"pl.konto_bankowe.record", "An institutional record identifying monetary funds held for an account owner", "bank_account"}}},
	{"pl.karta_kredytowa.n", "karta kredytowa", "noun", []lexicalMeaning{{"pl.karta_kredytowa.medium", "A payment card enabling purchases using an authorized credit facility", "credit_card"}}},
	{"pl.karta_debetowa.n", "karta debetowa", "noun", []lexicalMeaning{{"pl.karta_debetowa.medium", "A payment card used to pay from a linked bank account's available funds", "debit_card"}}},
	{"pl.paragon.n", "paragon", "noun", []lexicalMeaning{{"pl.paragon.record", "A record acknowledging that goods, money, or a service were received", "receipt"}}},
	{"pl.bilet.n", "bilet", "noun", []lexicalMeaning{{"pl.bilet.credential", "A credential granting access to a journey, venue, or specified service", "ticket"}}},
	{"pl.paszport.n", "paszport", "noun", []lexicalMeaning{{"pl.paszport.document", "An official travel identity document issued by a national authority", "passport"}}},
	{"pl.legitymacja.n", "legitymacja", "noun", []lexicalMeaning{{"pl.legitymacja.document", "An official or institutional card identifying its bearer", "identity_card"}}},
	{"pl.metka_z_cena.n", "metka z ceną", "noun", []lexicalMeaning{{"pl.metka_z_cena.label", "A label displaying the proposed price of a particular article", "price_tag"}}},
	{"pl.kasa.n", "kasa", "noun", []lexicalMeaning{{"pl.kasa.machine", "A machine recording retail transactions and often storing cash", "cash_register"}}},
}
