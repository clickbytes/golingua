package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishMovementVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_movement_verbs.go",
	Version: "2026-10-22-batch-28", License: "unknown", Author: "golingua contributors",
}

// These pending verbs name locomotion and travel events.
var englishMovementVerbEntries = []lexicalEntry{
	{"en.locomotion.v", "move", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.locomotion.event", "travel primarily by an organism using its own body", "locomotion"}}},
	{"en.walking.v", "walk", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.walking.event", "travel on foot by alternating steps while maintaining ground contact", "walking"}}},
	{"en.running.v", "run", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.running.event", "move rapidly on foot with intervals in which neither foot contacts the ground", "running"}}},
	{"en.crawling.v", "crawl", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.crawling.event", "move supported by hands, knees, or the body close to a surface", "crawling"}}},
	{"en.jumping.v", "jump", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.jumping.event", "make a propelled movement in which the body temporarily leaves its support surface", "jumping"}}},
	{"en.climbing.v", "climb", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.climbing.event", "move upward along a slope or structure by active effort", "climbing"}}},
	{"en.swimming_motion.v", "swim", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.swimming_motion.event", "travel through liquid by coordinated body movements", "swimming_motion"}}},
	{"en.flying_motion.v", "fly", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.flying_motion.event", "travel through air while supported by aerodynamic or powered lift", "flying_motion"}}},
	{"en.falling.v", "fall", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.falling.event", "move predominantly downward under gravity without sustained support", "falling"}}},
	{"en.rising_motion.v", "rise", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.rising_motion.event", "move from a lower spatial position to a higher one", "rising_motion"}}},
	{"en.entering.v", "enter", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.entering.event", "cross a boundary from outside to inside a region", "entering"}}},
	{"en.exiting.v", "exit", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.exiting.event", "cross a boundary from inside to outside a region", "exiting"}}},
	{"en.crossing.v", "cross", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.crossing.event", "move from one side of an obstacle or region to the other", "crossing"}}},
	{"en.approaching.v", "approach", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.approaching.event", "move so as to reduce distance between a mover and a target", "approaching"}}},
	{"en.moving_away.v", "move away", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.moving_away.event", "move so as to increase distance from a reference", "moving_away"}}},
	{"en.returning_to_place.v", "return", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.returning_to_place.event", "move to a location where the mover was previously present", "returning_to_place"}}},
	{"en.following_motion.v", "follow", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.following_motion.event", "move along the route of another moving participant", "following_motion"}}},
	{"en.chasing.v", "chase", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.chasing.event", "move intentionally in an attempt to catch a moving target", "chasing"}}},
	{"en.turning_direction.v", "turn", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.turning_direction.event", "change the direction of a moving entity’s path", "turning_direction"}}},
	{"en.starting_motion.v", "start moving", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.starting_motion.event", "transition from a stationary condition to spatial movement", "starting_motion"}}},
	{"en.stopping_motion.v", "stop", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.stopping_motion.event", "transition from spatial movement to a stationary condition", "stopping_motion"}}},
	{"en.traveling.v", "travel", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.traveling.event", "move between places with a destination or route", "traveling"}}},
	{"en.commuting.v", "commute", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.commuting.event", "travel regularly between a home and a work or study location", "commuting"}}},
	{"en.transporting.v", "transport", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.transporting.event", "cause a passenger or item to change location by carrying or conveyance", "transporting"}}},
	{"en.carrying.v", "carry", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.carrying.event", "move an object while physically supporting its weight during travel", "carrying"}}},
	{"en.driving_vehicle.v", "drive", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.driving_vehicle.event", "control a motor vehicle while it travels", "driving_vehicle"}}},
	{"en.riding_as_passenger.v", "ride", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.riding_as_passenger.event", "travel in a vehicle without controlling its propulsion or route", "riding_as_passenger"}}},
	{"en.boarding_vehicle.v", "board", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.boarding_vehicle.event", "enter a vehicle in preparation for travel", "boarding_vehicle"}}},
	{"en.leaving_vehicle.v", "leave", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.leaving_vehicle.event", "exit a vehicle after or during a journey", "leaving_vehicle"}}},
	{"en.parking_vehicle.v", "park", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.parking_vehicle.event", "bring a vehicle into a designated place and leave it stationary", "parking_vehicle"}}},
	{"en.taking_off.v", "take off", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.taking_off.event", "become airborne from a supporting surface, of a flying vehicle", "taking_off"}}},
	{"en.landing.v", "land", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.landing.event", "make contact with a supporting surface after flight, of a flying vehicle", "landing"}}},
}
