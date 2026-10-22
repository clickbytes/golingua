package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPlaceSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_places.go",
	Version: "2026-10-22-batch-22", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name rooms, buildings, routes and public facilities.
var englishPlaceEntries = []lexicalEntry{
	{"en.bedroom.n", "bedroom", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bedroom.direct", "a room intended primarily for sleeping", "bedroom"}}},
	{"en.living_room.n", "living room", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.living_room.direct", "a domestic room intended for shared sitting and relaxation", "living_room"}}},
	{"en.kitchen.n", "kitchen", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.kitchen.direct", "a room or dedicated area equipped for preparing food", "kitchen"}}},
	{"en.bathroom.n", "bathroom", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bathroom.direct", "a room equipped for washing or bathing, sometimes including a toilet", "bathroom"}}},
	{"en.toilet_room.n", "toilet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.toilet_room.direct", "a room or enclosed space fitted with a sanitation toilet", "toilet_room"}}},
	{"en.dining_room.n", "dining room", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.dining_room.direct", "a room intended primarily for eating meals", "dining_room"}}},
	{"en.office_room.n", "office", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.office_room.direct", "a room for desk-based administrative or professional work", "office_room"}}},
	{"en.classroom.n", "classroom", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.classroom.direct", "a room arranged for organized teaching and learning", "classroom"}}},
	{"en.waiting_room.n", "waiting room", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.waiting_room.direct", "a room provided for people awaiting a service or appointment", "waiting_room"}}},
	{"en.hallway.n", "hallway", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hallway.direct", "an interior passage connecting rooms or building spaces", "hallway"}}},
	{"en.entrance.n", "entrance", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.entrance.direct", "a passage point through which a person enters a place or structure", "entrance"}}},
	{"en.exit_place.n", "exit", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.exit_place.direct", "a passage point through which a person leaves a place or structure", "exit_place"}}},
	{"en.staircase.n", "staircase", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.staircase.direct", "a built series of steps linking levels of a structure", "staircase"}}},
	{"en.elevator.n", "elevator", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.elevator.direct", "a powered enclosed lift moving people or goods between floors", "elevator"}}},
	{"en.floor_level.n", "floor", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.floor_level.direct", "a designated horizontal level of a building", "floor_level"}}},
	{"en.roof.n", "roof", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.roof.direct", "the upper external covering of a building that protects its interior", "roof"}}},
	{"en.wall.n", "wall", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.wall.direct", "a vertical structural partition or enclosure that divides or supports spaces", "wall"}}},
	{"en.ceiling.n", "ceiling", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ceiling.direct", "the upper interior boundary of a room", "ceiling"}}},
	{"en.floor_surface.n", "floor", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.floor_surface.direct", "the load-bearing lower interior surface of a room", "floor_surface"}}},
	{"en.window.n", "window", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.window.direct", "a glazed or openable opening admitting light or air through a wall or roof", "window"}}},
	{"en.garage.n", "garage", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.garage.direct", "a structure or enclosed space designed to store vehicles", "garage"}}},
	{"en.shop_building.n", "shop", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shop_building.direct", "premises equipped for retail display and sale of goods", "shop_building"}}},
	{"en.hospital_building.n", "hospital", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hospital_building.direct", "a building equipped for medical care and treatment", "hospital_building"}}},
	{"en.library_building.n", "library", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.library_building.direct", "a building or dedicated space housing library collections and users", "library_building"}}},
	{"en.restaurant_building.n", "restaurant", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.restaurant_building.direct", "premises used for preparing and serving meals to customers", "restaurant_building"}}},
	{"en.hotel_building.n", "hotel", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hotel_building.direct", "a building providing temporary lodging rooms to paying guests", "hotel_building"}}},
	{"en.museum_building.n", "museum", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.museum_building.direct", "a facility exhibiting curated cultural, historical, or scientific objects", "museum_building"}}},
	{"en.theater_building.n", "theater", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.theater_building.direct", "a facility equipped for staged live performances", "theater_building"}}},
	{"en.cinema_building.n", "cinema", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cinema_building.direct", "a facility equipped for screening films to an audience", "cinema_building"}}},
	{"en.church_building.n", "church", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.church_building.direct", "a building designed for Christian congregational worship, distinct from an institution", "church_building"}}},
	{"en.airport.n", "airport", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.airport.direct", "an organized facility where aircraft take off, land, and passengers or cargo are handled", "airport"}}},
	{"en.train_station.n", "train station", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.train_station.direct", "a transport facility where passengers board or leave trains", "train_station"}}},
	{"en.bus_stop.n", "bus stop", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bus_stop.direct", "a designated location where a bus collects or sets down passengers", "bus_stop"}}},
	{"en.parking_area.n", "parking area", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.parking_area.direct", "a designated area in which vehicles may be left stationary", "parking_area"}}},
	{"en.playground.n", "playground", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.playground.direct", "an area equipped or designated for children’s outdoor play", "playground"}}},
	{"en.sports_field.n", "sports field", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sports_field.direct", "a prepared open area used for organized sporting activity", "sports_field"}}},
	{"en.swimming_pool.n", "swimming pool", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.swimming_pool.direct", "an artificial water-filled basin designed for swimming", "swimming_pool"}}},
	{"en.road.n", "road", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.road.direct", "a prepared route for vehicles or travelers moving between places", "road"}}},
	{"en.street.n", "street", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.street.direct", "a road through a settled area, often bordered by buildings", "street"}}},
	{"en.sidewalk.n", "sidewalk", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sidewalk.direct", "a dedicated pedestrian path alongside a road", "sidewalk"}}},
	{"en.path_walkway.n", "path", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.path_walkway.direct", "a narrow route intended for travel on foot", "path_walkway"}}},
	{"en.bridge.n", "bridge", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bridge.direct", "a constructed structure spanning an obstacle to carry traffic or people", "bridge"}}},
	{"en.tunnel.n", "tunnel", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tunnel.direct", "a constructed underground or enclosed passage through an obstacle", "tunnel"}}},
	{"en.crosswalk.n", "crosswalk", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.crosswalk.direct", "a designated part of a roadway intended for pedestrian crossing", "crosswalk"}}},
	{"en.intersection.n", "intersection", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.intersection.direct", "a location where two or more roads meet or cross", "intersection"}}},
}
