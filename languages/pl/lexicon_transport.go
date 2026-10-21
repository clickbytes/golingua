package pl

// transportEntries is the batch-7 Polish everyday vehicles group.
// Every entry is an authored editorial decision recorded in
// data/pl/1000-core-review.tsv; senses link to CORE as pending candidates
// without transfer authority. No paradigm is asserted for new lexemes yet.
var transportEntries = []lexicalEntry{
	{"pl.samochod.n", "samochód", "noun", []lexicalMeaning{{"pl.samochod.vehicle", "A road motor vehicle primarily designed to carry a small number of passengers", "car"}}},
	{"pl.autobus.n", "autobus", "noun", []lexicalMeaning{{"pl.autobus.vehicle", "A road vehicle designed to carry many passengers along a service or route", "bus"}}},
	{"pl.pociag.n", "pociąg", "noun", []lexicalMeaning{{"pl.pociag.vehicle", "A connected rail vehicle or set of vehicles moving on a track", "train"}}},
	{"pl.rower.n", "rower", "noun", []lexicalMeaning{{"pl.rower.vehicle", "A human-powered two-wheeled vehicle propelled by pedals", "bicycle"}}},
	{"pl.motocykl.n", "motocykl", "noun", []lexicalMeaning{{"pl.motocykl.vehicle", "A powered two-wheeled road vehicle ridden astride by its operator", "motorcycle"}}},
	{"pl.ciezarowka.n", "ciężarówka", "noun", []lexicalMeaning{{"pl.ciezarowka.vehicle", "A road motor vehicle built primarily for transporting cargo", "truck"}}},
	{"pl.samolot.n", "samolot", "noun", []lexicalMeaning{{"pl.samolot.vehicle", "A powered fixed-wing aircraft that gains lift while traveling through air", "airplane"}}},
	{"pl.lodz.n", "łódź", "noun", []lexicalMeaning{{"pl.lodz.vehicle", "A relatively small watercraft capable of carrying people or goods on water", "boat"}}},
	{"pl.statek.n", "statek", "noun", []lexicalMeaning{{"pl.statek.vehicle", "A large watercraft equipped for sustained travel and transport over water", "ship"}}},
	{"pl.taksowka.n", "taksówka", "noun", []lexicalMeaning{{"pl.taksowka.vehicle", "A passenger car operated to transport paying customers on demand", "taxi_vehicle"}}},
	{"pl.tramwaj.n", "tramwaj", "noun", []lexicalMeaning{{"pl.tramwaj.vehicle", "A passenger rail vehicle typically operating along urban street tracks", "tram"}}},
	{"pl.metro.n", "metro", "noun", []lexicalMeaning{{"pl.metro.vehicle", "A train designed for rapid urban transit, commonly using segregated tracks", "subway_train"}}},
	{"pl.karetka.n", "karetka", "noun", []lexicalMeaning{{"pl.karetka.vehicle", "A specially equipped vehicle for transporting ill or injured people for medical care", "ambulance"}}},
}
