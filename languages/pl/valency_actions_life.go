package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// lifeActionFrames declares only role signatures actually present in CORE and
// independently reviewed Polish case/adposition choices for these candidates.
var lifeActionFrames = []frameDeclaration{
	{"pl.oddychac.breathe.action", "pl.oddychac.breathe", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.wdychac.inhale.action", "pl.wdychac.inhale", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.wydychac.exhale.action", "pl.wydychac.exhale", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.kaszlec.cough.action", "pl.kaszlec.cough", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.kichac.sneeze.action", "pl.kichac.sneeze", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.polykac.swallow.action", "pl.polykac.swallow", "pl.transitive", []meaning.Role{"swallower", "substance"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "swallower", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "substance", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pic.drink.action", "pl.pic.drink", "pl.transitive", []meaning.Role{"drinker", "liquid"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "drinker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "liquid", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zuc.chew.action", "pl.zuc.chew", "pl.transitive", []meaning.Role{"chewer", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "chewer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.trawic.digest.action", "pl.trawic.digest", "pl.transitive", []meaning.Role{"organism", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pocic_sie.sweat.action", "pl.pocic_sie.sweat", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.krwawic.bleed.action", "pl.krwawic.bleed", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.zasypiac.sleep.action", "pl.zasypiac.sleep", "pl.intransitive", []meaning.Role{"sleeper"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sleeper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.budzic_sie.wake.action", "pl.budzic_sie.wake", "pl.intransitive", []meaning.Role{"person"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "person", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.cwiczyc.exercise.action", "pl.cwiczyc.exercise", "pl.intransitive", []meaning.Role{"participant"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.starzec_sie.aging.action", "pl.starzec_sie.aging", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.umierac.die.action", "pl.umierac.die", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.rodzic.birth.action", "pl.rodzic.birth", "pl.transitive", []meaning.Role{"parent", "offspring"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "parent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "offspring", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.leczyc.treat.action", "pl.leczyc.treat", "pl.transitive", []meaning.Role{"care_provider", "patient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "care_provider", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "patient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.diagnozowac.diagnose.action", "pl.diagnozowac.diagnose", "pl.transitive", []meaning.Role{"clinician", "patient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "clinician", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "patient", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.robic.make.action", "pl.robic.make", "pl.transitive", []meaning.Role{"maker", "product"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "maker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "product", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.produkowac.manufacture.action", "pl.produkowac.manufacture", "pl.transitive", []meaning.Role{"producer", "product"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "producer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "product", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.projektowac.design.action", "pl.projektowac.design", "pl.transitive", []meaning.Role{"designer", "product"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "designer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "product", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.fotografowac.photograph.action", "pl.fotografowac.photograph", "pl.transitive", []meaning.Role{"photographer", "subject"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "photographer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "subject", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.edytowac.edit.action", "pl.edytowac.edit", "pl.transitive", []meaning.Role{"editor", "document"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "editor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "document", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.drukowac.print.action", "pl.drukowac.print", "pl.transitive", []meaning.Role{"operator", "document"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "operator", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "document", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.gotowac.cook.action", "pl.gotowac.cook", "pl.transitive", []meaning.Role{"cook", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "cook", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.piec.bake.action", "pl.piec.bake", "pl.transitive", []meaning.Role{"cook", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "cook", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.smazyc.fry.action", "pl.smazyc.fry", "pl.transitive", []meaning.Role{"cook", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "cook", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.sprzatac.clean.action", "pl.sprzatac.clean", "pl.transitive", []meaning.Role{"cleaner", "home"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "cleaner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "home", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.podlewac.water.action", "pl.podlewac.water", "pl.transitive", []meaning.Role{"gardener", "plant"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "gardener", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "plant", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zyc.live.action", "pl.zyc.live", "pl.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.smiac_sie.laugh.action", "pl.smiac_sie.laugh", "pl.intransitive", []meaning.Role{"person"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "person", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.plakac.cry.action", "pl.plakac.cry", "pl.intransitive", []meaning.Role{"person"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "person", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.usmiechac_sie.smile.action", "pl.usmiechac_sie.smile", "pl.intransitive", []meaning.Role{"person"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "person", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.tanczyc.dance.action", "pl.tanczyc.dance", "pl.intransitive", []meaning.Role{"dancer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "dancer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}
