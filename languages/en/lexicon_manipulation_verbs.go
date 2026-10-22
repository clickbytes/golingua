package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishManipulationVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_manipulation_verbs.go",
	Version: "2026-10-22-batch-26", License: "unknown", Author: "golingua contributors",
}

// These pending verb candidates name deliberate physical actions.
var englishManipulationVerbEntries = []lexicalEntry{
	{"en.intentional_action.v", "act", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.intentional_action.event", "undertake an event deliberately to produce an effect", "intentional_action"}}},
	{"en.manipulating_object.v", "manipulate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.manipulating_object.event", "intentionally control an object’s position, shape, or state by physical action", "manipulating_object"}}},
	{"en.touching_object.v", "touch", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.touching_object.event", "make physical surface contact with an object without requiring it to move", "touching_object"}}},
	{"en.grasping.v", "grasp", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.grasping.event", "close a hand or implement around an object to hold it securely", "grasping"}}},
	{"en.holding_object.v", "hold", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.holding_object.event", "maintain physical support or control of an object without releasing it", "holding_object"}}},
	{"en.releasing_object.v", "release", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.releasing_object.event", "end a maintained physical hold or restraint on an object", "releasing_object"}}},
	{"en.lifting_object.v", "lift", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.lifting_object.event", "move a supported object to a higher position", "lifting_object"}}},
	{"en.lowering_object.v", "lower", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.lowering_object.event", "deliberately move a supported object to a lower position", "lowering_object"}}},
	{"en.placing_object.v", "place", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.placing_object.event", "move an object to a target position and leave it there", "placing_object"}}},
	{"en.inserting_object.v", "insert", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.inserting_object.event", "move an object through an opening into a containing region", "inserting_object"}}},
	{"en.extracting_object.v", "extract", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.extracting_object.event", "remove an object from within a container or enclosed place", "extracting_object"}}},
	{"en.pushing_object.v", "push", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.pushing_object.event", "apply force directed away from the acting participant to move or press an object", "pushing_object"}}},
	{"en.pulling_object.v", "pull", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.pulling_object.event", "apply force directed toward the acting participant to move or draw an object", "pulling_object"}}},
	{"en.rotating_object.v", "rotate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.rotating_object.event", "turn an object around an axis without requiring it to change location", "rotating_object"}}},
	{"en.hitting_object.v", "hit", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.hitting_object.event", "deliver a discrete physical impact to an object or person", "hitting_object"}}},
	{"en.throwing_object.v", "throw", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.throwing_object.event", "propel an object into free motion by releasing it after exerting force", "throwing_object"}}},
	{"en.catching_object.v", "catch", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.catching_object.event", "intercept and secure an object while it is moving freely", "catching_object"}}},
	{"en.dropping_object.v", "drop", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.dropping_object.event", "cause a held object to fall by letting it go", "dropping_object"}}},
	{"en.kicking_object.v", "kick", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.kicking_object.event", "strike an object or participant with a foot", "kicking_object"}}},
	{"en.opening_object.v", "open", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.opening_object.event", "move an obstruction or mechanism so access to an interior becomes possible", "opening_object"}}},
	{"en.closing_object.v", "close", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.closing_object.event", "move an obstruction or mechanism to prevent ordinary access through an opening", "closing_object"}}},
	{"en.locking_object.v", "lock", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.locking_object.event", "engage a mechanism that prevents opening or unauthorized access", "locking_object"}}},
	{"en.unlocking_object.v", "unlock", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.unlocking_object.event", "disengage a mechanism to permit an object or entrance to open", "unlocking_object"}}},
	{"en.cutting_object.v", "cut", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.cutting_object.event", "separate material using a sharp edge or equivalent concentrated force", "cutting_object"}}},
	{"en.tearing_object.v", "tear", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.tearing_object.event", "separate flexible material by pulling it until it ruptures", "tearing_object"}}},
	{"en.folding_object.v", "fold", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.folding_object.event", "bend flexible material to bring parts of it over one another", "folding_object"}}},
	{"en.bending_object.v", "bend", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.bending_object.event", "change an object’s shape into a curve or angle without necessarily separating it", "bending_object"}}},
	{"en.stretching_material.v", "stretch", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.stretching_material.event", "increase a material’s length or area by tension", "stretching_material"}}},
	{"en.breaking_object.v", "break", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.breaking_object.event", "cause a physical object to lose structural integrity or function", "breaking_object"}}},
	{"en.repairing_object.v", "repair", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.repairing_object.event", "restore a damaged or malfunctioning object toward its intended function", "repairing_object"}}},
	{"en.assembling_object.v", "assemble", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.assembling_object.event", "join separate components to produce a larger structured artifact", "assembling_object"}}},
	{"en.disassembling_object.v", "disassemble", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.disassembling_object.event", "separate a structured artifact into its components", "disassembling_object"}}},
	{"en.connecting_objects.v", "connect", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.connecting_objects.event", "establish physical attachment or a working link between two objects", "connecting_objects"}}},
	{"en.disconnecting_objects.v", "disconnect", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.disconnecting_objects.event", "remove an attachment or working link between previously joined objects", "disconnecting_objects"}}},
	{"en.covering_object.v", "cover", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.covering_object.event", "place a layer or object over a target to conceal or protect its surface", "covering_object"}}},
	{"en.filling_container.v", "fill", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.filling_container.event", "add material until a container holds more contents than before", "filling_container"}}},
	{"en.emptying_container.v", "empty", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.emptying_container.event", "remove contents from a container until it contains less or none", "emptying_container"}}},
	{"en.pouring_liquid.v", "pour", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.pouring_liquid.event", "transfer liquid by allowing it to flow from a tilted or opened source", "pouring_liquid"}}},
	{"en.mixing_materials.v", "mix", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.mixing_materials.event", "combine distinct substances so their components become interspersed", "mixing_materials"}}},
	{"en.wiping_surface.v", "wipe", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.wiping_surface.event", "pass a cloth or implement over a surface to remove material or moisture", "wiping_surface"}}},
	{"en.scrubbing_surface.v", "scrub", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.scrubbing_surface.event", "clean a surface by repeated forceful rubbing", "scrubbing_surface"}}},
	{"en.washing_object.v", "wash", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.washing_object.event", "apply water or cleaning liquid to remove unwanted matter from an object", "washing_object"}}},
	{"en.drying_object.v", "dry", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.drying_object.event", "reduce moisture in or on an object through heat, air, or absorbent material", "drying_object"}}},
	{"en.heating_object.v", "heat", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.heating_object.event", "cause the temperature of an object or substance to increase", "heating_object"}}},
	{"en.cooling_object.v", "cool", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.cooling_object.event", "cause the temperature of an object or substance to decrease", "cooling_object"}}},
	{"en.freezing_substance.v", "freeze", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.freezing_substance.event", "cause a liquid substance to become solid through loss of heat", "freezing_substance"}}},
	{"en.melting_substance.v", "melt", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.melting_substance.event", "change a solid substance into a liquid through addition of energy", "melting_substance"}}},
	{"en.burning_material.v", "burn", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.burning_material.event", "undergo sustained combustion with an oxidizer", "burning_material"}}},
	{"en.igniting_material.v", "ignite", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.igniting_material.event", "cause combustion to begin in a combustible material", "igniting_material"}}},
	{"en.extinguishing_fire.v", "extinguish", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.extinguishing_fire.event", "end an existing combustion process by removing conditions needed for burning", "extinguishing_fire"}}},
}
