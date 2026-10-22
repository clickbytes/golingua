package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishTransportSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_transport.go",
	Version: "2026-10-22-batch-18", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name transport vehicles, not transport events or services.
var englishTransportEntries = []lexicalEntry{
	{"en.car.n", "car", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.car.passenger_road_vehicle", "a road motor vehicle primarily designed to carry a small number of passengers", "car"}}},
	{"en.bus.n", "bus", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bus.passenger_service_vehicle", "a road vehicle designed to carry many passengers along a service or route", "bus"}}},
	{"en.train.n", "train", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.train.rail_vehicle", "a connected rail vehicle or set of vehicles moving on a track", "train"}}},
	{"en.bicycle.n", "bicycle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bicycle.pedal_vehicle", "a human-powered two-wheeled vehicle propelled by pedals", "bicycle"}}},
	{"en.motorcycle.n", "motorcycle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.motorcycle.powered_two_wheeler", "a powered two-wheeled road vehicle ridden astride by its operator", "motorcycle"}}},
	{"en.truck.n", "truck", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.truck.cargo_vehicle", "a road motor vehicle built primarily for transporting cargo", "truck"}}},
	{"en.airplane.n", "airplane", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.airplane.fixed_wing_aircraft", "a powered fixed-wing aircraft that gains lift while traveling through air", "airplane"}}},
	{"en.boat.n", "boat", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.boat.small_watercraft", "a relatively small watercraft capable of carrying people or goods on water", "boat"}}},
	{"en.ship.n", "ship", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ship.large_watercraft", "a large watercraft equipped for sustained travel and transport over water", "ship"}}},
	{"en.taxi_vehicle.n", "taxi", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.taxi_vehicle.on_demand_car", "a passenger car operated to transport paying customers on demand", "taxi_vehicle"}}},
	{"en.tram.n", "tram", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tram.urban_rail_vehicle", "a passenger rail vehicle typically operating along urban street tracks", "tram"}}},
	{"en.subway_train.n", "subway train", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.subway_train.rapid_transit", "a train designed for rapid urban transit, commonly using segregated tracks", "subway_train"}}},
	{"en.ambulance.n", "ambulance", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ambulance.medical_transport_vehicle", "a specially equipped vehicle for transporting ill or injured people for medical care", "ambulance"}}},
}
