package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// polishMotionFrames records a conservative one- or two-participant analysis
// for each selected motion sense. A missing destination/adjunct has NOT been
// silently inserted as a CORE semantic role.
var polishMotionFrames = []frameDeclaration{
	{"pl.chodzic.walk.motion", "pl.chodzic.walk", "pl.intransitive", []meaning.Role{"walker"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "walker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.biec.run.motion", "pl.biec.run", "pl.intransitive", []meaning.Role{"runner"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "runner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.czolgac_sie.crawl.motion", "pl.czolgac_sie.crawl", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.skakac.jump.motion", "pl.skakac.jump", "pl.intransitive", []meaning.Role{"jumper"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "jumper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.spadac.fall.motion", "pl.spadac.fall", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.skrecac.turn.motion", "pl.skrecac.turn", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.ruszac.start.motion", "pl.ruszac.start", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.zatrzymywac_sie.stop.motion", "pl.zatrzymywac_sie.stop", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.podrozowac.travel.motion", "pl.podrozowac.travel", "pl.intransitive", []meaning.Role{"traveler"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "traveler", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.odpoczywac.rest.motion", "pl.odpoczywac.rest", "pl.intransitive", []meaning.Role{"participant"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.wspinac_sie.climb.motion", "pl.wspinac_sie.climb", "pl.oblique", []meaning.Role{"climber", "surface"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "climber", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "surface", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "po"},
	}},
	{"pl.wchodzic.enter.motion", "pl.wchodzic.enter", "pl.oblique", []meaning.Role{"mover", "destination"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "destination", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.wychodzic.exit.motion", "pl.wychodzic.exit", "pl.oblique", []meaning.Role{"mover", "source"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "source", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.przechodzic.cross.motion", "pl.przechodzic.cross", "pl.oblique", []meaning.Role{"mover", "region"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "region", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "przez"},
	}},
	{"pl.wracac.return.motion", "pl.wracac.return", "pl.oblique", []meaning.Role{"mover", "destination"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "destination", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.podazac.follow.motion", "pl.podazac.follow", "pl.oblique", []meaning.Role{"follower", "leader"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "follower", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "leader", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "za"},
	}},
	{"pl.gonic.chase.motion", "pl.gonic.chase", "pl.transitive", []meaning.Role{"pursuer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "pursuer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.niesc.carry.motion", "pl.niesc.carry", "pl.transitive", []meaning.Role{"carrier", "load"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "carrier", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "load", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wsiadac.board.motion", "pl.wsiadac.board", "pl.oblique", []meaning.Role{"passenger", "vehicle"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "passenger", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "vehicle", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.wysiadac.alight.motion", "pl.wysiadac.alight", "pl.oblique", []meaning.Role{"passenger", "vehicle"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "passenger", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "vehicle", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.czekac.wait.motion", "pl.czekac.wait", "pl.oblique", []meaning.Role{"waiter", "anticipated"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "waiter", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "anticipated", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.pozostawac.stay.motion", "pl.pozostawac.stay", "pl.oblique", []meaning.Role{"occupant", "location"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "occupant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "location", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "w"},
	}},
	{"pl.prowadzic.drive.motion", "pl.prowadzic.drive", "pl.transitive", []meaning.Role{"driver", "vehicle"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "driver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "vehicle", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.przewozic.transport.motion", "pl.przewozic.transport", "pl.transitive", []meaning.Role{"carrier", "passenger_or_load"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "carrier", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "passenger_or_load", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}
