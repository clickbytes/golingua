package pl

// motionEntries is a narrow, hand-checked valency slice of the event editorial
// worklist. These 24 meanings have no asserted morphology, approved bindings,
// or sentence generation capability.
var motionEntries = []lexicalEntry{
	{"pl.chodzic.v", "chodzić", "verb", []lexicalMeaning{{"pl.chodzic.walk", "Locomotion on foot by alternating steps while maintaining ground contact.", "walking"}}},
	{"pl.biec.v", "biec", "verb", []lexicalMeaning{{"pl.biec.run", "Rapid locomotion on foot involving intervals without ground contact.", "running"}}},
	{"pl.czolgac_sie.v", "czołgać się", "verb_reflexive", []lexicalMeaning{{"pl.czolgac_sie.crawl", "Locomotion supported by hands, knees, or the body close to a surface.", "crawling"}}},
	{"pl.skakac.v", "skakać", "verb", []lexicalMeaning{{"pl.skakac.jump", "Propelled movement in which a body temporarily leaves its support surface.", "jumping"}}},
	{"pl.spadac.v", "spadać", "verb", []lexicalMeaning{{"pl.spadac.fall", "Predominantly downward movement under gravity without sustained support.", "falling"}}},
	{"pl.skrecac.v", "skręcać", "verb", []lexicalMeaning{{"pl.skrecac.turn", "Change in the direction of a moving entity's path.", "turning_direction"}}},
	{"pl.ruszac.v", "ruszać", "verb", []lexicalMeaning{{"pl.ruszac.start", "Transition from stationary condition to spatial movement.", "starting_motion"}}},
	{"pl.zatrzymywac_sie.v", "zatrzymywać się", "verb_reflexive", []lexicalMeaning{{"pl.zatrzymywac_sie.stop", "Transition from spatial movement to a stationary condition.", "stopping_motion"}}},
	{"pl.podrozowac.v", "podróżować", "verb", []lexicalMeaning{{"pl.podrozowac.travel", "Sustained movement between places with a destination or route.", "traveling"}}},
	{"pl.odpoczywac.v", "odpoczywać", "verb", []lexicalMeaning{{"pl.odpoczywac.rest", "Temporarily reducing activity to recover energy or comfort.", "resting"}}},
	{"pl.wspinac_sie.v", "wspinać się", "verb_reflexive", []lexicalMeaning{{"pl.wspinac_sie.climb", "Movement gaining height along a slope or structure by active effort.", "climbing"}}},
	{"pl.wchodzic.v", "wchodzić", "verb", []lexicalMeaning{{"pl.wchodzic.enter", "Movement crossing a boundary from outside to inside a region.", "entering"}}},
	{"pl.wychodzic.v", "wychodzić", "verb", []lexicalMeaning{{"pl.wychodzic.exit", "Movement crossing a boundary from inside to outside a region.", "exiting"}}},
	{"pl.przechodzic.v", "przechodzić", "verb", []lexicalMeaning{{"pl.przechodzic.cross", "Movement from one side of an obstacle or region to the other.", "crossing"}}},
	{"pl.wracac.v", "wracać", "verb", []lexicalMeaning{{"pl.wracac.return", "Movement to a location where the mover was previously present.", "returning_to_place"}}},
	{"pl.podazac.v", "podążać", "verb", []lexicalMeaning{{"pl.podazac.follow", "Movement along the route of another moving participant.", "following_motion"}}},
	{"pl.gonic.v", "gonić", "verb", []lexicalMeaning{{"pl.gonic.chase", "Intentional movement attempting to catch a moving target.", "chasing"}}},
	{"pl.niesc.v", "nieść", "verb", []lexicalMeaning{{"pl.niesc.carry", "Moving an object while physically supporting its weight during travel.", "carrying"}}},
	{"pl.wsiadac.v", "wsiadać", "verb", []lexicalMeaning{{"pl.wsiadac.board", "Entering a vehicle in preparation for travel.", "boarding_vehicle"}}},
	{"pl.wysiadac.v", "wysiadać", "verb", []lexicalMeaning{{"pl.wysiadac.alight", "Exiting a vehicle after or during a journey.", "leaving_vehicle"}}},
	{"pl.czekac.v", "czekać", "verb", []lexicalMeaning{{"pl.czekac.wait", "Remaining ready or inactive until an anticipated condition or event occurs.", "waiting"}}},
	{"pl.pozostawac.v", "pozostawać", "verb", []lexicalMeaning{{"pl.pozostawac.stay", "Maintaining presence at a location without an implied change of position.", "staying_at_place"}}},
	{"pl.prowadzic.v", "prowadzić", "verb", []lexicalMeaning{{"pl.prowadzic.drive", "Controlling a motor vehicle while it travels.", "driving_vehicle"}}},
	{"pl.przewozic.v", "przewozić", "verb", []lexicalMeaning{{"pl.przewozic.transport", "Causing a passenger or item to change location by carrying or conveyance.", "transporting"}}},
}

// motionFrameForSense does not infer a frame from a concept or POS. Every
// entry corresponds to an explicit, manually selected motion frame below.
func motionFrameForSense(senseID string) (string, bool) {
	switch senseID {
	case "pl.chodzic.walk":
		return "pl.chodzic.walk.motion", true
	case "pl.biec.run":
		return "pl.biec.run.motion", true
	case "pl.czolgac_sie.crawl":
		return "pl.czolgac_sie.crawl.motion", true
	case "pl.skakac.jump":
		return "pl.skakac.jump.motion", true
	case "pl.spadac.fall":
		return "pl.spadac.fall.motion", true
	case "pl.skrecac.turn":
		return "pl.skrecac.turn.motion", true
	case "pl.ruszac.start":
		return "pl.ruszac.start.motion", true
	case "pl.zatrzymywac_sie.stop":
		return "pl.zatrzymywac_sie.stop.motion", true
	case "pl.podrozowac.travel":
		return "pl.podrozowac.travel.motion", true
	case "pl.odpoczywac.rest":
		return "pl.odpoczywac.rest.motion", true
	case "pl.wspinac_sie.climb":
		return "pl.wspinac_sie.climb.motion", true
	case "pl.wchodzic.enter":
		return "pl.wchodzic.enter.motion", true
	case "pl.wychodzic.exit":
		return "pl.wychodzic.exit.motion", true
	case "pl.przechodzic.cross":
		return "pl.przechodzic.cross.motion", true
	case "pl.wracac.return":
		return "pl.wracac.return.motion", true
	case "pl.podazac.follow":
		return "pl.podazac.follow.motion", true
	case "pl.gonic.chase":
		return "pl.gonic.chase.motion", true
	case "pl.niesc.carry":
		return "pl.niesc.carry.motion", true
	case "pl.wsiadac.board":
		return "pl.wsiadac.board.motion", true
	case "pl.wysiadac.alight":
		return "pl.wysiadac.alight.motion", true
	case "pl.czekac.wait":
		return "pl.czekac.wait.motion", true
	case "pl.pozostawac.stay":
		return "pl.pozostawac.stay.motion", true
	case "pl.prowadzic.drive":
		return "pl.prowadzic.drive.motion", true
	case "pl.przewozic.transport":
		return "pl.przewozic.transport.motion", true
	default:
		return "", false
	}
}
