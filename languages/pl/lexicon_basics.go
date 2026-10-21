package pl

// basicsEntries collects the resolved original-49 nouns and leftover person/property rows that were still worklist-only.
// Every entry is an authored editorial decision recorded in
// data/pl/1000-core-review.tsv; senses link to CORE as pending candidates
// without transfer authority. No paradigm is asserted for new lexemes yet.
var basicsEntries = []lexicalEntry{
	{"pl.osoba.n", "osoba", "noun", []lexicalMeaning{{"pl.osoba.person", "A human individual, irrespective of age, social role, or name", "person"}}},
	{"pl.dorosly.adj", "dorosły", "adjective", []lexicalMeaning{{"pl.dorosly.person", "A person past the child stage; no legal age threshold is asserted", "adult"}}},
	{"pl.uczen.n", "uczeń", "noun", []lexicalMeaning{{"pl.uczen.learner", "A person acting in a role of learning from instruction or study", "student"}}},
	{"pl.pacjent.n", "pacjent", "noun", []lexicalMeaning{{"pl.pacjent.medical", "A person receiving or seeking medical care", "patient"}}},
	{"pl.zwierze.n", "zwierzę", "noun", []lexicalMeaning{{"pl.zwierze.animal", "A living nonhuman organism capable of animal life processes", "animal"}}},
	{"pl.roslina.n", "roślina", "noun", []lexicalMeaning{{"pl.roslina.organism", "A living organism of the plant kingdom", "plant"}}},
	{"pl.drzewo.n", "drzewo", "noun", []lexicalMeaning{{"pl.drzewo.plant", "A woody perennial plant with a persistent main stem", "tree"}}},
	{"pl.rzecz.n", "rzecz", "noun", []lexicalMeaning{{"pl.rzecz.object", "A physical object intentionally made or modified by an agent", "artifact"}}},
	{"pl.drzwi.n", "drzwi", "noun", []lexicalMeaning{{"pl.drzwi.barrier", "A movable barrier used to open or close an entrance", "door"}}},
	{"pl.pojazd.n", "pojazd", "noun", []lexicalMeaning{{"pl.pojazd.conveyance", "An artifact designed to carry people or goods between places", "vehicle"}}},
	{"pl.miejsce.n", "miejsce", "noun", []lexicalMeaning{{"pl.miejsce.location", "A spatial region that can serve as the site of an event or entity", "location"}}},
	{"pl.jedzenie.n", "jedzenie", "noun", []lexicalMeaning{{"pl.jedzenie.food", "A substance or prepared item intended for nourishment", "food"}}},
	{"pl.percepcja.n", "percepcja", "noun", []lexicalMeaning{{"pl.percepcja.sense", "An event in which a perceiver acquires sensory information about a stimulus", "perception"}}},
	{"pl.ruch.n", "ruch", "noun", []lexicalMeaning{{"pl.ruch.motion", "A change in an entity's spatial position over time", "movement"}}},
	{"pl.przybycie.n", "przybycie", "noun", []lexicalMeaning{{"pl.przybycie.arrival", "Movement ending at a destination", "arrival"}}},
	{"pl.wyjazd.n", "wyjazd", "noun", []lexicalMeaning{{"pl.wyjazd.departure", "Movement beginning at a source location", "departure"}}},
	{"pl.cecha.n", "cecha", "noun", []lexicalMeaning{{"pl.cecha.property", "A quality that can be attributed to a bearer without identifying a particular bearer", "property"}}},
	{"pl.wielkosc.n", "wielkość", "noun", []lexicalMeaning{{"pl.wielkosc.size", "The extent of a bearer along one or more spatial dimensions", "size"}}},
	{"pl.temperatura.n", "temperatura", "noun", []lexicalMeaning{{"pl.temperatura.thermal", "The thermal state of a bearer on a temperature scale", "temperature"}}},
	{"pl.kolor.n", "kolor", "noun", []lexicalMeaning{{"pl.kolor.visual", "A visual quality of a bearer associated with perceived light", "color"}}},
	{"pl.predkosc.n", "prędkość", "noun", []lexicalMeaning{{"pl.predkosc.movement", "The rate of a participant's movement relative to time", "speed"}}},
	{"pl.kurier.n", "kurier", "noun", []lexicalMeaning{{"pl.kurier.role", "A person transporting items for delivery to a recipient", "courier"}}},
	{"pl.wolontariusz.n", "wolontariusz", "noun", []lexicalMeaning{{"pl.wolontariusz.role", "A person contributing labor without an ordinary employment obligation for that contribution", "volunteer_person"}}},
}
