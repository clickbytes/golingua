package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// physicalActionFrames declares only role signatures actually present in CORE and
// independently reviewed Polish case/adposition choices for these candidates.
var physicalActionFrames = []frameDeclaration{
	{"pl.dotykac_obj.touch.action", "pl.dotykac_obj.touch", "pl.transitive", []meaning.Role{"actor", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "actor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.chwycic.grasp.action", "pl.chwycic.grasp", "pl.transitive", []meaning.Role{"holder", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.trzymac.hold.action", "pl.trzymac.hold", "pl.transitive", []meaning.Role{"holder", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.puszczac.release.action", "pl.puszczac.release", "pl.transitive", []meaning.Role{"releaser", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "releaser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.podnosic.lift.action", "pl.podnosic.lift", "pl.transitive", []meaning.Role{"actor", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "actor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.opuszczac.lower.action", "pl.opuszczac.lower", "pl.transitive", []meaning.Role{"actor", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "actor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.popychac.push.action", "pl.popychac.push", "pl.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.ciagnac.pull.action", "pl.ciagnac.pull", "pl.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.obracac.rotate.action", "pl.obracac.rotate", "pl.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.uderzac.hit.action", "pl.uderzac.hit", "pl.transitive", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.rzucac.throw.action", "pl.rzucac.throw", "pl.transitive", []meaning.Role{"thrower", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "thrower", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.lapac.catch.action", "pl.lapac.catch", "pl.transitive", []meaning.Role{"catcher", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "catcher", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.upuszczac.drop.action", "pl.upuszczac.drop", "pl.transitive", []meaning.Role{"dropper", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "dropper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.kopac.kick.action", "pl.kopac.kick", "pl.transitive", []meaning.Role{"actor", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "actor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.ciac.cut.action", "pl.ciac.cut", "pl.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.drzec.tear.action", "pl.drzec.tear", "pl.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.skladac.fold.action", "pl.skladac.fold", "pl.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zaginac.bend.action", "pl.zaginac.bend", "pl.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.rozciagac.stretch.action", "pl.rozciagac.stretch", "pl.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.naprawiac.repair.action", "pl.naprawiac.repair", "pl.transitive", []meaning.Role{"repairer", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "repairer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.rozbierac.disassemble.action", "pl.rozbierac.disassemble", "pl.transitive", []meaning.Role{"agent", "artifact"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "artifact", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.przecierac.wipe.action", "pl.przecierac.wipe", "pl.transitive", []meaning.Role{"agent", "surface"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "surface", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.szorowac.scrub.action", "pl.szorowac.scrub", "pl.transitive", []meaning.Role{"agent", "surface"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "surface", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.myc.wash.action", "pl.myc.wash", "pl.transitive", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.suszyc.dry.action", "pl.suszyc.dry", "pl.transitive", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wlaczac.power_on.action", "pl.wlaczac.power_on", "pl.transitive", []meaning.Role{"agent", "device"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "device", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wylaczac.power_off.action", "pl.wylaczac.power_off", "pl.transitive", []meaning.Role{"agent", "device"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "device", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wyrzucac.discard.action", "pl.wyrzucac.discard", "pl.transitive", []meaning.Role{"disposer", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "disposer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}
