package pl

// actionFrameForSense keeps the curated sense -> valency mapping explicit.
// A lexical declaration alone never authorizes a surface construction.
func actionFrameForSense(sense string) (string, bool) {
	switch sense {
	case "pl.dotykac_obj.touch":
		return "pl.dotykac_obj.touch.action", true
	case "pl.chwycic.grasp":
		return "pl.chwycic.grasp.action", true
	case "pl.trzymac.hold":
		return "pl.trzymac.hold.action", true
	case "pl.puszczac.release":
		return "pl.puszczac.release.action", true
	case "pl.podnosic.lift":
		return "pl.podnosic.lift.action", true
	case "pl.opuszczac.lower":
		return "pl.opuszczac.lower.action", true
	case "pl.popychac.push":
		return "pl.popychac.push.action", true
	case "pl.ciagnac.pull":
		return "pl.ciagnac.pull.action", true
	case "pl.obracac.rotate":
		return "pl.obracac.rotate.action", true
	case "pl.uderzac.hit":
		return "pl.uderzac.hit.action", true
	case "pl.rzucac.throw":
		return "pl.rzucac.throw.action", true
	case "pl.lapac.catch":
		return "pl.lapac.catch.action", true
	case "pl.upuszczac.drop":
		return "pl.upuszczac.drop.action", true
	case "pl.kopac.kick":
		return "pl.kopac.kick.action", true
	case "pl.ciac.cut":
		return "pl.ciac.cut.action", true
	case "pl.drzec.tear":
		return "pl.drzec.tear.action", true
	case "pl.skladac.fold":
		return "pl.skladac.fold.action", true
	case "pl.zaginac.bend":
		return "pl.zaginac.bend.action", true
	case "pl.rozciagac.stretch":
		return "pl.rozciagac.stretch.action", true
	case "pl.naprawiac.repair":
		return "pl.naprawiac.repair.action", true
	case "pl.rozbierac.disassemble":
		return "pl.rozbierac.disassemble.action", true
	case "pl.przecierac.wipe":
		return "pl.przecierac.wipe.action", true
	case "pl.szorowac.scrub":
		return "pl.szorowac.scrub.action", true
	case "pl.myc.wash":
		return "pl.myc.wash.action", true
	case "pl.suszyc.dry":
		return "pl.suszyc.dry.action", true
	case "pl.wlaczac.power_on":
		return "pl.wlaczac.power_on.action", true
	case "pl.wylaczac.power_off":
		return "pl.wylaczac.power_off.action", true
	case "pl.wyrzucac.discard":
		return "pl.wyrzucac.discard.action", true
	case "pl.oddychac.breathe":
		return "pl.oddychac.breathe.action", true
	case "pl.wdychac.inhale":
		return "pl.wdychac.inhale.action", true
	case "pl.wydychac.exhale":
		return "pl.wydychac.exhale.action", true
	case "pl.kaszlec.cough":
		return "pl.kaszlec.cough.action", true
	case "pl.kichac.sneeze":
		return "pl.kichac.sneeze.action", true
	case "pl.polykac.swallow":
		return "pl.polykac.swallow.action", true
	case "pl.pic.drink":
		return "pl.pic.drink.action", true
	case "pl.zuc.chew":
		return "pl.zuc.chew.action", true
	case "pl.trawic.digest":
		return "pl.trawic.digest.action", true
	case "pl.pocic_sie.sweat":
		return "pl.pocic_sie.sweat.action", true
	case "pl.krwawic.bleed":
		return "pl.krwawic.bleed.action", true
	case "pl.zasypiac.sleep":
		return "pl.zasypiac.sleep.action", true
	case "pl.budzic_sie.wake":
		return "pl.budzic_sie.wake.action", true
	case "pl.cwiczyc.exercise":
		return "pl.cwiczyc.exercise.action", true
	case "pl.starzec_sie.aging":
		return "pl.starzec_sie.aging.action", true
	case "pl.umierac.die":
		return "pl.umierac.die.action", true
	case "pl.rodzic.birth":
		return "pl.rodzic.birth.action", true
	case "pl.leczyc.treat":
		return "pl.leczyc.treat.action", true
	case "pl.diagnozowac.diagnose":
		return "pl.diagnozowac.diagnose.action", true
	case "pl.robic.make":
		return "pl.robic.make.action", true
	case "pl.produkowac.manufacture":
		return "pl.produkowac.manufacture.action", true
	case "pl.projektowac.design":
		return "pl.projektowac.design.action", true
	case "pl.fotografowac.photograph":
		return "pl.fotografowac.photograph.action", true
	case "pl.edytowac.edit":
		return "pl.edytowac.edit.action", true
	case "pl.drukowac.print":
		return "pl.drukowac.print.action", true
	case "pl.gotowac.cook":
		return "pl.gotowac.cook.action", true
	case "pl.piec.bake":
		return "pl.piec.bake.action", true
	case "pl.smazyc.fry":
		return "pl.smazyc.fry.action", true
	case "pl.sprzatac.clean":
		return "pl.sprzatac.clean.action", true
	case "pl.podlewac.water":
		return "pl.podlewac.water.action", true
	case "pl.zyc.live":
		return "pl.zyc.live.action", true
	case "pl.smiac_sie.laugh":
		return "pl.smiac_sie.laugh.action", true
	case "pl.plakac.cry":
		return "pl.plakac.cry.action", true
	case "pl.usmiechac_sie.smile":
		return "pl.usmiechac_sie.smile.action", true
	case "pl.tanczyc.dance":
		return "pl.tanczyc.dance.action", true
	case "pl.witac.greet":
		return "pl.witac.greet.action", true
	case "pl.dziekowac.thank":
		return "pl.dziekowac.thank.action", true
	case "pl.przepraszac.apologize":
		return "pl.przepraszac.apologize.action", true
	case "pl.doradzac.advise":
		return "pl.doradzac.advise.action", true
	case "pl.przekonywac.persuade":
		return "pl.przekonywac.persuade.action", true
	case "pl.chwalic.praise":
		return "pl.chwalic.praise.action", true
	case "pl.krytykowac.criticize":
		return "pl.krytykowac.criticize.action", true
	case "pl.szeptac.whisper":
		return "pl.szeptac.whisper.action", true
	case "pl.dzwonic.call":
		return "pl.dzwonic.call.action", true
	case "pl.poznawac.meet":
		return "pl.poznawac.meet.action", true
	case "pl.pomagac.help":
		return "pl.pomagac.help.action", true
	case "pl.kochac.love":
		return "pl.kochac.love.action", true
	case "pl.lubic.like":
		return "pl.lubic.like.action", true
	case "pl.bac_sie.fear":
		return "pl.bac_sie.fear.action", true
	case "pl.martwic_sie.worry":
		return "pl.martwic_sie.worry.action", true
	case "pl.potrzebowac.need":
		return "pl.potrzebowac.need.action", true
	case "pl.znajdowac.find":
		return "pl.znajdowac.find.action", true
	case "pl.gubic.lose":
		return "pl.gubic.lose.action", true
	case "pl.szukac.search":
		return "pl.szukac.search.action", true
	case "pl.rezerwowac.reserve":
		return "pl.rezerwowac.reserve.action", true
	case "pl.anulowac.cancel":
		return "pl.anulowac.cancel.action", true
	case "pl.oszczedzac.save":
		return "pl.oszczedzac.save.action", true
	case "pl.wydawac.spend":
		return "pl.wydawac.spend.action", true
	case "pl.zarabiac.earn":
		return "pl.zarabiac.earn.action", true
	default:
		return "", false
	}
}
