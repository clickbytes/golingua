package pl

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/provenance"
)

var attestedNounSource = provenance.SourceRef{
	ID:      "golingua-pl-noun-forms-authored",
	Path:    "languages/pl/morphology_attested_nouns.go",
	Version: "2026-10-23-authored-v3",
	License: "unknown",
	Author:  "golingua contributors",
}

var attestedNounExternalSources = map[string]provenance.SourceRef{
	"pl.gazeta.n":   {ID: "wsjp-pl-gazeta", URI: "https://wsjp.pl/index.php/haslo/podglad/1844/gazeta/4785722/czasopismo", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.deszcz.n":   {ID: "wsjp-pl-deszcz", URI: "https://wsjp.pl/haslo/podglad/11500/deszcz/1079560/opady", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.ksiezyc.n":  {ID: "wsjp-pl-ksiezyc", URI: "https://wsjp.pl/haslo/podglad/11137/ksiezyc/4461839/na-niebie-w-nocy", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.dokument.n": {ID: "wsjp-pl-dokument", URI: "https://wsjp.pl/haslo/podglad/3667/dokument/2431628/urzedowy", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.chleb.n":    {ID: "wsjp-pl-chleb", URI: "https://wsjp.pl/haslo/podglad/20808/chleb/4804753/pieczywo", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.jajko.n":    {ID: "wsjp-pl-jajko", URI: "https://wsjp.pl/haslo/podglad/20747/jajko/4843860/do-jedzenia", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.zupa.n":     {ID: "wsjp-pl-zupa", URI: "https://wsjp.pl/haslo/podglad/2007/zupa/2126917/potrawa", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.mleko.n":    {ID: "wsjp-pl-mleko", URI: "https://wsjp.pl/haslo/podglad/1993/mleko", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.herbata.n":  {ID: "wsjp-pl-herbata", URI: "https://wsjp.pl/haslo/podglad/15163/herbata/1949148/napoj", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.kawa.n":     {ID: "wsjp-pl-kawa", URI: "https://wsjp.pl/haslo/podglad/22171/kawa/4843824/napoj", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.papier.n":   {ID: "wsjp-pl-papier", URI: "https://wsjp.pl/haslo/podglad/35381/papier/3873750/do-pisania", Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.krzeslo.n": {
		ID: "wsjp-pl-krzeslo-n", URI: "https://wsjp.pl/haslo/podglad/6705/krzeslo/4734420/mebel",
		Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.klucz.n": {
		ID: "wsjp-pl-klucz-n", URI: "https://wsjp.pl/haslo/podglad/16229/klucz/5076317/do-otwierania",
		Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.torba.n": {
		ID: "wsjp-pl-torba-n", URI: "https://wsjp.pl/haslo/podglad/3755/torba/759009/do-noszenia-ze-soba",
		Version: "2026-10-28", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.olowek.n": {ID: "wsjp-pl-olowek", URI: "https://wsjp.pl/haslo/podglad/794/olowek/180566/do-pisania", Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.noz.n":    {ID: "wsjp-pl-noz", URI: "https://wsjp.pl/haslo/podglad/19680/noz/4789950/narzedzie", Version: "2026-10-29", License: "unknown", Author: "Wielki słownik języka polskiego PAN"},
	"pl.kot.n": {
		ID: "wsjp-pl-kot-n", URI: "https://wsjp.pl/haslo/podglad/29368/kot/5136345/domowy",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.ptak.n": {
		ID: "wsjp-pl-ptak-n", URI: "https://wsjp.pl/haslo/podglad/20014/ptak/4177524/stworzenie",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.kon.n": {
		ID: "wsjp-pl-kon-n", URI: "https://wsjp.pl/haslo/podglad/16472/kon/5069035/zwierze",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.miasto.n": {
		ID: "wsjp-pl-miasto-n", URI: "https://wsjp.pl/haslo/podglad/3383/miasto/2451641/teren",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.mieszkanie.n": {
		ID: "wsjp-pl-mieszkanie-n", URI: "https://wsjp.pl/haslo/podglad/3382/mieszkanie",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.rzeka.n": {
		ID: "wsjp-pl-rzeka-n", URI: "https://wsjp.pl/haslo/podglad/1735/rzeka/667911/woda",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.siostra.n": {
		ID: "wsjp-pl-siostra-n", URI: "https://wsjp.pl/haslo/podglad/8259/siostra/1883127/w-rodzinie",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
	"pl.mewa.n": {
		ID: "wsjp-pl-mewa-n", URI: "https://wsjp.pl/haslo/podglad/37575/mewa/4059050/ptak",
		Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
	},
}

var treeNounFormSource = provenance.SourceRef{
	ID:      "golingua-pl-tree-forms-wsj",
	URI:     "https://wsjp.pl/haslo/podglad/10895/drzewo/4680852/roslina",
	Version: "2026-10-23", License: "unknown", Author: "Wielki słownik języka polskiego PAN",
}

// attestedNounForms contains the declared case/number forms. Entries assigned
// to a named partial class omit only cells licensed by that class's procedure.
type attestedNounEntry struct {
	id, lemma, gender string
	singular, plural  [7]string
}

// These assignments are lexical data. The shared profile derives only the six
// cells below; all other cells remain explicitly attested per lexeme.
var attestedNounClassByLexeme = map[string]string{
	"pl.szkola.n": "feminine_a_partial",
	"pl.ryba.n":   "feminine_a_partial",
	"pl.rzeka.n":  "feminine_a_partial",
	"pl.mewa.n":   "feminine_a_partial",
}

var feminineAPartialSuffixByCell = map[string]string{
	"sg_accusative":   "ę",
	"sg_instrumental": "ą",
	"sg_vocative":     "o",
	"pl_dative":       "om",
	"pl_instrumental": "ami",
	"pl_locative":     "ach",
}

var attestedNounForms = []attestedNounEntry{
	{"pl.gazeta.n", "gazeta", "feminine", [7]string{"gazeta", "gazety", "gazecie", "gazetę", "gazetą", "gazecie", "gazeto"}, [7]string{"gazety", "gazet", "gazetom", "gazety", "gazetami", "gazetach", "gazety"}},
	{"pl.deszcz.n", "deszcz", "masculine_inanimate", [7]string{"deszcz", "deszczu", "deszczowi", "deszcz", "deszczem", "deszczu", "deszczu"}, [7]string{"deszcze", "deszczów", "deszczom", "deszcze", "deszczami", "deszczach", "deszcze"}},
	{"pl.ksiezyc.n", "księżyc", "masculine_inanimate", [7]string{"księżyc", "księżyca", "księżycowi", "księżyc", "księżycem", "księżycu", "księżycu"}, [7]string{"księżyce", "księżyców", "księżycom", "księżyce", "księżycami", "księżycach", "księżyce"}},
	{"pl.dokument.n", "dokument", "masculine_inanimate", [7]string{"dokument", "dokumentu", "dokumentowi", "dokument", "dokumentem", "dokumencie", "dokumencie"}, [7]string{"dokumenty", "dokumentów", "dokumentom", "dokumenty", "dokumentami", "dokumentach", "dokumenty"}},
	{"pl.chleb.n", "chleb", "masculine_inanimate", [7]string{"chleb", "chleba", "chlebowi", "chleb", "chlebem", "chlebie", "chlebie"}, [7]string{"chleby", "chlebów", "chlebom", "chleby", "chlebami", "chlebach", "chleby"}},
	{"pl.jajko.n", "jajko", "neuter", [7]string{"jajko", "jajka", "jajku", "jajko", "jajkiem", "jajku", "jajko"}, [7]string{"jajka", "jajek", "jajkom", "jajka", "jajkami", "jajkach", "jajka"}},
	{"pl.zupa.n", "zupa", "feminine", [7]string{"zupa", "zupy", "zupie", "zupę", "zupą", "zupie", "zupo"}, [7]string{"zupy", "zup", "zupom", "zupy", "zupami", "zupach", "zupy"}},
	{"pl.mleko.n", "mleko", "neuter", [7]string{"mleko", "mleka", "mleku", "mleko", "mlekiem", "mleku", "mleko"}, [7]string{"mleka", "mlek", "mlekom", "mleka", "mlekami", "mlekach", "mleka"}},
	{"pl.herbata.n", "herbata", "feminine", [7]string{"herbata", "herbaty", "herbacie", "herbatę", "herbatą", "herbacie", "herbato"}, [7]string{"herbaty", "herbat", "herbatom", "herbaty", "herbatami", "herbatach", "herbaty"}},
	{"pl.kawa.n", "kawa", "feminine", [7]string{"kawa", "kawy", "kawie", "kawę", "kawą", "kawie", "kawo"}, [7]string{"kawy", "kaw", "kawom", "kawy", "kawami", "kawach", "kawy"}},
	{"pl.papier.n", "papier", "masculine_inanimate", [7]string{"papier", "papieru", "papierowi", "papier", "papierem", "papierze", "papierze"}, [7]string{"papiery", "papierów", "papierom", "papiery", "papierami", "papierach", "papiery"}},
	{"pl.olowek.n", "ołówek", "masculine_inanimate", [7]string{"ołówek", "ołówka", "ołówkowi", "ołówek", "ołówkiem", "ołówku", "ołówku"}, [7]string{"ołówki", "ołówków", "ołówkom", "ołówki", "ołówkami", "ołówkach", "ołówki"}},
	{"pl.matka.n", "matka", "feminine", [7]string{"matka", "matki", "matce", "matkę", "matką", "matce", "matko"}, [7]string{"matki", "matek", "matkom", "matki", "matkami", "matkach", "matki"}},
	{"pl.ojciec.n", "ojciec", "masculine_personal", [7]string{"ojciec", "ojca", "ojcu", "ojca", "ojcem", "ojcu", "ojcze"}, [7]string{"ojcowie", "ojców", "ojcom", "ojców", "ojcami", "ojcach", "ojcowie"}},
	{"pl.pies.n", "pies", "masculine_animal", [7]string{"pies", "psa", "psu", "psa", "psem", "psie", "psie"}, [7]string{"psy", "psów", "psom", "psy", "psami", "psach", "psy"}},
	{"pl.dom.n", "dom", "masculine_inanimate", [7]string{"dom", "domu", "domowi", "dom", "domem", "domu", "domu"}, [7]string{"domy", "domów", "domom", "domy", "domami", "domach", "domy"}},
	{"pl.woda.n", "woda", "feminine", [7]string{"woda", "wody", "wodzie", "wodę", "wodą", "wodzie", "wodo"}, [7]string{"wody", "wód", "wodom", "wody", "wodami", "wodach", "wody"}},
	{"pl.szkola.n", "szkoła", "feminine", [7]string{"szkoła", "szkoły", "szkole", "", "", "szkole", ""}, [7]string{"szkoły", "szkół", "", "szkoły", "", "", "szkoły"}},
	{"pl.ksiazka.n", "książka", "feminine", [7]string{"książka", "książki", "książce", "książkę", "książką", "książce", "książko"}, [7]string{"książki", "książek", "książkom", "książki", "książkami", "książkach", "książki"}},
	{"pl.reka.n", "ręka", "feminine", [7]string{"ręka", "ręki", "ręce", "rękę", "ręką", "ręce", "ręko"}, [7]string{"ręce", "rąk", "rękom", "ręce", "rękami", "rękach", "ręce"}},
	{"pl.okno.n", "okno", "neuter", [7]string{"okno", "okna", "oknu", "okno", "oknem", "oknie", "okno"}, [7]string{"okna", "okien", "oknom", "okna", "oknami", "oknach", "okna"}},
	{"pl.jablko.n", "jabłko", "neuter", [7]string{"jabłko", "jabłka", "jabłku", "jabłko", "jabłkiem", "jabłku", "jabłko"}, [7]string{"jabłka", "jabłek", "jabłkom", "jabłka", "jabłkami", "jabłkach", "jabłka"}},
	{"pl.list.n", "list", "masculine_inanimate", [7]string{"list", "listu", "listowi", "list", "listem", "liście", "liście"}, [7]string{"listy", "listów", "listom", "listy", "listami", "listach", "listy"}},
	{"pl.oko.n", "oko", "neuter", [7]string{"oko", "oka", "oku", "oko", "okiem", "oku", "oko"}, [7]string{"oczy", "oczu", "oczom", "oczy", "oczami", "oczach", "oczy"}},
	{"pl.dziecko.n", "dziecko", "neuter", [7]string{"dziecko", "dziecka", "dziecku", "dziecko", "dzieckiem", "dziecku", "dziecko"}, [7]string{"dzieci", "dzieci", "dzieciom", "dzieci", "dziećmi", "dzieciach", "dzieci"}},
	{"pl.stol.n", "stół", "masculine_inanimate", [7]string{"stół", "stołu", "stołowi", "stół", "stołem", "stole", "stole"}, [7]string{"stoły", "stołów", "stołom", "stoły", "stołami", "stołach", "stoły"}},
	{"pl.nauczyciel.n", "nauczyciel", "masculine_personal", [7]string{"nauczyciel", "nauczyciela", "nauczycielowi", "nauczyciela", "nauczycielem", "nauczycielu", "nauczycielu"}, [7]string{"nauczyciele", "nauczycieli", "nauczycielom", "nauczycieli", "nauczycielami", "nauczycielach", "nauczyciele"}},
	{"pl.lekarz.n", "lekarz", "masculine_personal", [7]string{"lekarz", "lekarza", "lekarzowi", "lekarza", "lekarzem", "lekarzu", "lekarzu"}, [7]string{"lekarze", "lekarzy", "lekarzom", "lekarzy", "lekarzami", "lekarzach", "lekarze"}},
	{"pl.samochod.n", "samochód", "masculine_inanimate", [7]string{"samochód", "samochodu", "samochodowi", "samochód", "samochodem", "samochodzie", "samochodzie"}, [7]string{"samochody", "samochodów", "samochodom", "samochody", "samochodami", "samochodach", "samochody"}},
	{"pl.ryba.n", "ryba", "feminine", [7]string{"ryba", "ryby", "rybie", "", "", "rybie", ""}, [7]string{"ryby", "ryb", "", "ryby", "", "", "ryby"}},
	{"pl.szpital.n", "szpital", "masculine_inanimate", [7]string{"szpital", "szpitala", "szpitalowi", "szpital", "szpitalem", "szpitalu", "szpitalu"}, [7]string{"szpitale", "szpitali", "szpitalom", "szpitale", "szpitalami", "szpitalach", "szpitale"}},
	{"pl.pizza.n", "pizza", "feminine", [7]string{"pizza", "pizzy", "pizzy", "pizzę", "pizzą", "pizzy", "pizzo"}, [7]string{"pizze", "pizz", "pizzom", "pizze", "pizzami", "pizzach", "pizze"}},
	{"pl.sasiad.n", "sąsiad", "masculine_personal", [7]string{"sąsiad", "sąsiada", "sąsiadowi", "sąsiada", "sąsiadem", "sąsiedzie", "sąsiedzie"}, [7]string{"sąsiedzi", "sąsiadów", "sąsiadom", "sąsiadów", "sąsiadami", "sąsiadach", "sąsiedzi"}},
	{"pl.klient.n", "klient", "masculine_personal", [7]string{"klient", "klienta", "klientowi", "klienta", "klientem", "kliencie", "kliencie"}, [7]string{"klienci", "klientów", "klientom", "klientów", "klientami", "klientach", "klienci"}},
	{"pl.jedzenie.n", "jedzenie", "neuter", [7]string{"jedzenie", "jedzenia", "jedzeniu", "jedzenie", "jedzeniem", "jedzeniu", "jedzenie"}, [7]string{"jedzenia", "jedzeń", "jedzeniom", "jedzenia", "jedzeniami", "jedzeniach", "jedzenia"}},
	{"pl.wiadomosc.n", "wiadomość", "feminine", [7]string{"wiadomość", "wiadomości", "wiadomości", "wiadomość", "wiadomością", "wiadomości", "wiadomości"}, [7]string{"wiadomości", "wiadomości", "wiadomościom", "wiadomości", "wiadomościami", "wiadomościach", "wiadomości"}},
	{"pl.kot.n", "kot", "masculine_animal", [7]string{"kot", "kota", "kotu", "kota", "kotem", "kocie", "kocie"}, [7]string{"koty", "kotów", "kotom", "koty", "kotami", "kotach", "koty"}},
	{"pl.ptak.n", "ptak", "masculine_animal", [7]string{"ptak", "ptaka", "ptakowi", "ptaka", "ptakiem", "ptaku", "ptaku"}, [7]string{"ptaki", "ptaków", "ptakom", "ptaki", "ptakami", "ptakach", "ptaki"}},
	{"pl.kon.n", "koń", "masculine_animal", [7]string{"koń", "konia", "koniowi", "konia", "koniem", "koniu", "koniu"}, [7]string{"konie", "koni", "koniom", "konie", "końmi", "koniach", "konie"}},
	{"pl.miasto.n", "miasto", "neuter", [7]string{"miasto", "miasta", "miastu", "miasto", "miastem", "mieście", "miasto"}, [7]string{"miasta", "miast", "miastom", "miasta", "miastami", "miastach", "miasta"}},
	{"pl.mieszkanie.n", "mieszkanie", "neuter", [7]string{"mieszkanie", "mieszkania", "mieszkaniu", "mieszkanie", "mieszkaniem", "mieszkaniu", "mieszkanie"}, [7]string{"mieszkania", "mieszkań", "mieszkaniom", "mieszkania", "mieszkaniami", "mieszkaniach", "mieszkania"}},
	{"pl.rzeka.n", "rzeka", "feminine", [7]string{"rzeka", "rzeki", "rzece", "", "", "rzece", ""}, [7]string{"rzeki", "rzek", "", "rzeki", "", "", "rzeki"}},
	{"pl.siostra.n", "siostra", "feminine", [7]string{"siostra", "siostry", "siostrze", "siostrę", "siostrą", "siostrze", "siostro"}, [7]string{"siostry", "sióstr", "siostrom", "siostry", "siostrami", "siostrach", "siostry"}},
	{"pl.mewa.n", "mewa", "feminine", [7]string{"mewa", "mewy", "mewie", "", "", "mewie", ""}, [7]string{"mewy", "mew", "", "mewy", "", "", "mewy"}},
	{"pl.krzeslo.n", "krzesło", "neuter", [7]string{"krzesło", "krzesła", "krzesłu", "krzesło", "krzesłem", "krześle", "krzesło"}, [7]string{"krzesła", "krzeseł", "krzesłom", "krzesła", "krzesłami", "krzesłach", "krzesła"}},
	{"pl.klucz.n", "klucz", "masculine_inanimate", [7]string{"klucz", "klucza", "kluczowi", "klucz", "kluczem", "kluczu", "kluczu"}, [7]string{"klucze", "kluczy", "kluczom", "klucze", "kluczami", "kluczach", "klucze"}},
	{"pl.torba.n", "torba", "feminine", [7]string{"torba", "torby", "torbie", "torbę", "torbą", "torbie", "torbo"}, [7]string{"torby", "toreb", "torbom", "torby", "torbami", "torbach", "torby"}},
	{"pl.noz.n", "nóż", "masculine_inanimate", [7]string{"nóż", "noża", "nożowi", "nóż", "nożem", "nożu", "nożu"}, [7]string{"noże", "noży", "nożom", "noże", "nożami", "nożach", "noże"}},
}

var attestedNounCases = [7]string{"nominative", "genitive", "dative", "accusative", "instrumental", "locative", "vocative"}

func attestedNounGender(lexemeID string) (string, bool) {
	switch lexemeID {
	case "pl.kot.n", "pl.ptak.n", "pl.kon.n":
		return "masculine_animal", true
	case "pl.miasto.n", "pl.mieszkanie.n", "pl.okno.n", "pl.jablko.n", "pl.krzeslo.n", "pl.jajko.n", "pl.mleko.n":
		return "neuter", true
	case "pl.rzeka.n", "pl.siostra.n", "pl.mewa.n", "pl.torba.n", "pl.zupa.n", "pl.herbata.n", "pl.kawa.n":
		return "feminine", true
	case "pl.gazeta.n":
		return "feminine", true
	case "pl.klucz.n", "pl.chleb.n", "pl.papier.n", "pl.deszcz.n", "pl.ksiezyc.n", "pl.dokument.n", "pl.olowek.n", "pl.noz.n":
		return "masculine_inanimate", true
	case "pl.matka.n":
		return "feminine", true
	case "pl.ojciec.n":
		return "masculine_personal", true
	case "pl.pies.n":
		return "masculine_animal", true
	case "pl.dom.n":
		return "masculine_inanimate", true
	case "pl.woda.n":
		return "feminine", true
	case "pl.szkola.n":
		return "feminine", true
	case "pl.ksiazka.n":
		return "feminine", true
	case "pl.reka.n":
		return "feminine", true
	case "pl.oko.n":
		return "neuter", true
	case "pl.dziecko.n":
		return "neuter", true
	case "pl.stol.n", "pl.list.n":
		return "masculine_inanimate", true
	case "pl.nauczyciel.n":
		return "masculine_personal", true
	case "pl.lekarz.n":
		return "masculine_personal", true
	case "pl.samochod.n":
		return "masculine_inanimate", true
	case "pl.ryba.n":
		return "feminine", true
	case "pl.szpital.n":
		return "masculine_inanimate", true
	case "pl.pizza.n":
		return "feminine", true
	case "pl.sasiad.n", "pl.klient.n":
		return "masculine_personal", true
	case "pl.jedzenie.n":
		return "neuter", true
	case "pl.wiadomosc.n":
		return "feminine", true
	default:
		return "", false
	}
}

func pluralOnlyNounParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion,
		ID:            "pl_noun_plural_only_attested", Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
	}
	for _, grammaticalCase := range attestedNounCases {
		partID := morph.PrincipalPartID("pl_" + grammaticalCase)
		cellID := morph.CellID(partID)
		stemID := procedure.StemID(partID)
		paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
			StemID: stemID, PartID: partID, Required: true,
		})
		paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
			ID: cellID,
			Features: []features.FeatureValue{
				{Feature: "number", State: features.StateKnown, Value: "plural"},
				{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
			},
		})
		paradigm.Rules = append(paradigm.Rules, morph.CellRule{
			CellID: cellID, ProcedureID: "procedure:pl_noun_plural_only_attested." + string(cellID), Provenance: attestedNounSource,
			Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
		})
	}
	return paradigm
}

func pluralOnlyNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	formsByLexeme := map[string][7]string{
		"lexeme:pl.pieniadze.n": {"pieniądze", "pieniędzy", "pieniądzom", "pieniądze", "pieniędzmi", "pieniądzach", "pieniądze"},
		"lexeme:pl.drzwi.n":     {"drzwi", "drzwi", "drzwiom", "drzwi", "drzwiami", "drzwiach", "drzwi"},
	}
	for _, lexeme := range lexemes {
		forms, exists := formsByLexeme[lexeme.ID.String()]
		if !exists {
			continue
		}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, len(attestedNounCases))}
		for index, grammaticalCase := range attestedNounCases {
			parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: morph.PrincipalPartID("pl_" + grammaticalCase), Form: forms[index]})
		}
		return []lexemeRecord{{Lexeme: lexeme, PrincipalParts: parts}}, nil
	}
	return nil, fmt.Errorf("Polish plural-only noun lexemes are missing")
}

func attestedNounParadigms(profile *features.LanguageFeatureProfile) []morph.Paradigm {
	genders := []string{"masculine_personal", "masculine_animal", "masculine_inanimate", "feminine", "neuter"}
	result := make([]morph.Paradigm, 0, len(genders)+len(attestedNounExternalSources)+1)
	for _, gender := range genders {
		paradigmID := morph.ParadigmID("pl_noun_attested_" + gender)
		result = append(result, attestedNounParadigm(profile, paradigmID, gender, attestedNounSource))
	}
	result = append(result, attestedNounFeminineAPartialParadigm(profile, "pl_noun_feminine_a_partial", "feminine", attestedNounSource))
	for _, entry := range attestedNounForms {
		source, external := attestedNounExternalSources[entry.id]
		if !external {
			continue
		}
		paradigmID := attestedNounParadigmID(entry.id, entry.gender)
		if attestedNounClassByLexeme[entry.id] == "feminine_a_partial" {
			result = append(result, attestedNounFeminineAPartialParadigm(profile, paradigmID, entry.gender, source))
			continue
		}
		result = append(result, attestedNounParadigm(profile, paradigmID, entry.gender, source))
	}
	return result
}

func attestedNounParadigmID(lexemeID, gender string) morph.ParadigmID {
	if _, external := attestedNounExternalSources[lexemeID]; external {
		stableSuffix := strings.NewReplacer(".", "_", "-", "_").Replace(lexemeID)
		return morph.ParadigmID("pl_noun_attested_" + stableSuffix)
	}
	if attestedNounClassByLexeme[lexemeID] == "feminine_a_partial" {
		return "pl_noun_feminine_a_partial"
	}
	return morph.ParadigmID("pl_noun_attested_" + gender)
}

func attestedNounParadigm(profile *features.LanguageFeatureProfile, id morph.ParadigmID, gender string, source provenance.SourceRef) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: id, Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
	}
	for _, number := range []string{"singular", "plural"} {
		prefix := "sg"
		if number == "plural" {
			prefix = "pl"
		}
		for _, grammaticalCase := range attestedNounCases {
			partID := morph.PrincipalPartID(prefix + "_" + grammaticalCase)
			stemID := procedure.StemID(partID)
			cellID := morph.CellID(partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: stemID, PartID: partID, Required: true,
			})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
				ID: cellID,
				Features: []features.FeatureValue{
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
					{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
					{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(gender)},
				},
			})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{
				CellID: cellID, ProcedureID: "procedure:" + string(paradigm.ID) + "." + string(cellID),
				Provenance: source, Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
			})
		}
	}
	return paradigm
}

func attestedNounFeminineAPartialParadigm(profile *features.LanguageFeatureProfile, id morph.ParadigmID, gender string, source provenance.SourceRef) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion: morph.ParadigmSchemaVersion, ID: id, Language: "pl", ProfileVersion: profile.Version(), POS: "noun",
	}
	for _, number := range []string{"singular", "plural"} {
		prefix := "sg"
		if number == "plural" {
			prefix = "pl"
		}
		for _, grammaticalCase := range attestedNounCases {
			partID := morph.PrincipalPartID(prefix + "_" + grammaticalCase)
			cellID := morph.CellID(partID)
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
				ID: cellID,
				Features: []features.FeatureValue{
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
					{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
					{Feature: "gender", State: features.StateKnown, Value: features.FeatureValueID(gender)},
				},
			})
			if _, derived := feminineAPartialSuffixByCell[string(cellID)]; derived {
				paradigm.Rules = append(paradigm.Rules, morph.CellRule{
					CellID: cellID, ProcedureID: "procedure:" + string(paradigm.ID) + "." + string(cellID), Provenance: source,
					Steps: []procedure.Operation{
						{Kind: procedure.OperationSelectStem, StemID: "sg_nominative"},
						{Kind: procedure.OperationStripSuffix, Suffix: "a"},
						{Kind: procedure.OperationAppend, Text: feminineAPartialSuffixByCell[string(cellID)]},
					},
				})
				continue
			}
			stemID := procedure.StemID(partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: stemID, PartID: partID, Required: true,
			})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{
				CellID: cellID, ProcedureID: "procedure:" + string(paradigm.ID) + "." + string(cellID), Provenance: source,
				Steps: []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
			})
		}
	}
	return paradigm
}

func attestedNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	indexed := make(map[string]lexicon.MorphLexeme, len(lexemes))
	for _, lexeme := range lexemes {
		indexed[lexeme.ID.String()] = lexeme
	}
	result := make([]lexemeRecord, 0, len(attestedNounForms))
	for _, entry := range attestedNounForms {
		lexeme, ok := indexed["lexeme:"+entry.id]
		if !ok || lexeme.Lemma != entry.lemma || lexeme.ParadigmID != string(attestedNounParadigmID(entry.id, entry.gender)) {
			return nil, fmt.Errorf("Polish attested noun %s has no compatible lexical declaration", entry.id)
		}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, 14)}
		for caseIndex, grammaticalCase := range attestedNounCases {
			for _, item := range []struct {
				id   string
				form string
			}{{"sg_" + grammaticalCase, entry.singular[caseIndex]}, {"pl_" + grammaticalCase, entry.plural[caseIndex]}} {
				_, derived := feminineAPartialSuffixByCell[item.id]
				if attestedNounClassByLexeme[entry.id] == "feminine_a_partial" && derived {
					if item.form != "" {
						return nil, fmt.Errorf("Polish attested noun %s duplicates procedural form for %s", entry.id, item.id)
					}
					continue
				}
				if item.form == "" {
					return nil, fmt.Errorf("Polish attested noun %s is missing explicit form for %s", entry.id, item.id)
				}
				parts.Parts = append(parts.Parts, morph.PrincipalPart{ID: morph.PrincipalPartID(item.id), Form: item.form})
			}
		}
		result = append(result, lexemeRecord{Lexeme: lexeme, PrincipalParts: parts})
	}
	return result, nil
}

func treeNounParadigm(profile *features.LanguageFeatureProfile) morph.Paradigm {
	paradigm := morph.Paradigm{
		SchemaVersion:  morph.ParadigmSchemaVersion,
		ID:             "pl_noun_drzewo_attested",
		Language:       "pl",
		ProfileVersion: profile.Version(),
		POS:            "noun",
	}
	for _, number := range []string{"singular", "plural"} {
		prefix := "sg"
		if number == "plural" {
			prefix = "pl"
		}
		for _, grammaticalCase := range attestedNounCases {
			partID := morph.PrincipalPartID(prefix + "_" + grammaticalCase)
			stemID := procedure.StemID(partID)
			cellID := morph.CellID(partID)
			paradigm.PrincipalPartRequirements = append(paradigm.PrincipalPartRequirements, morph.PrincipalPartRequirement{
				StemID: stemID, PartID: partID, Required: true,
			})
			paradigm.Cells = append(paradigm.Cells, morph.ParadigmCell{
				ID: cellID,
				Features: []features.FeatureValue{
					{Feature: "number", State: features.StateKnown, Value: features.FeatureValueID(number)},
					{Feature: "case", State: features.StateKnown, Value: features.FeatureValueID(grammaticalCase)},
					{Feature: "gender", State: features.StateKnown, Value: "neuter"},
				},
			})
			paradigm.Rules = append(paradigm.Rules, morph.CellRule{
				CellID: cellID, ProcedureID: "procedure:pl_noun_drzewo_attested." + string(cellID),
				Provenance: treeNounFormSource,
				Steps:      []procedure.Operation{{Kind: procedure.OperationSelectStem, StemID: stemID}},
			})
		}
	}
	return paradigm
}

func treeNounRecords(lexemes []lexicon.MorphLexeme) ([]lexemeRecord, error) {
	for _, lexeme := range lexemes {
		if lexeme.ID.String() != "lexeme:pl.drzewo.n" {
			continue
		}
		if lexeme.Lemma != "drzewo" || lexeme.ParadigmID != "pl_noun_drzewo_attested" {
			return nil, fmt.Errorf("Polish tree noun has an incompatible lexical declaration")
		}
		singular := [7]string{"drzewo", "drzewa", "drzewu", "drzewo", "drzewem", "drzewie", "drzewo"}
		plural := [7]string{"drzewa", "drzew", "drzewom", "drzewa", "drzewami", "drzewach", "drzewa"}
		parts := morph.PrincipalParts{LexemeID: lexeme.ID, Parts: make([]morph.PrincipalPart, 0, 14)}
		for index, grammaticalCase := range attestedNounCases {
			parts.Parts = append(parts.Parts,
				morph.PrincipalPart{ID: morph.PrincipalPartID("sg_" + grammaticalCase), Form: singular[index]},
				morph.PrincipalPart{ID: morph.PrincipalPartID("pl_" + grammaticalCase), Form: plural[index]},
			)
		}
		return []lexemeRecord{{Lexeme: lexeme, PrincipalParts: parts}}, nil
	}
	return nil, fmt.Errorf("Polish tree noun lexeme is missing")
}
