package pl

// actionFrameForSense keeps the curated sense -> valency mapping explicit.
// A lexical declaration alone never authorizes a surface construction.
func actionFrameForSense(sense string) (string, bool) {
	switch sense {
	case "pl.rozkazywac.command":
		return "pl.rozkazywac.command.action", true
	case "pl.udac_sie.succeed":
		return "pl.udac_sie.succeed.action", true
	case "pl.nie_udac_sie.fail":
		return "pl.nie_udac_sie.fail.action", true
	case "pl.zamawiac_usluge.hire":
		return "pl.zamawiac_usluge.hire.action", true
	case "pl.zle_zrozumiec.misunderstand":
		return "pl.zle_zrozumiec.misunderstand.action", true
	case "pl.tlumaczyc_ustnie.interpret":
		return "pl.tlumaczyc_ustnie.interpret.action", true
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
	case "pl.pchac.push":
		return "pl.pchac.push.action", true
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
	case "pl.rozdzierac.tear":
		return "pl.rozdzierac.tear.action", true
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
	case "pl.otwierac_obiekt.open":
		return "pl.otwierac_obiekt.open.action", true
	case "pl.zamykac_obiekt.close":
		return "pl.zamykac_obiekt.close.action", true
	case "pl.lamac.break":
		return "pl.lamac.break.action", true
	case "pl.podgrzewac.heat":
		return "pl.podgrzewac.heat.action", true
	case "pl.chlodzic.cool":
		return "pl.chlodzic.cool.action", true
	case "pl.podpalac.ignite":
		return "pl.podpalac.ignite.action", true
	case "pl.gasic.extinguish":
		return "pl.gasic.extinguish.action", true
	case "pl.ladowac_baterie.charge":
		return "pl.ladowac_baterie.charge.action", true
	case "pl.zachorowac.ill":
		return "pl.zachorowac.ill.action", true
	case "pl.myc_zeby.brush":
		return "pl.myc_zeby.brush.action", true
	case "pl.wymieniac_sie.exchange":
		return "pl.wymieniac_sie.exchange.action", true
	case "pl.goic_sie.heal":
		return "pl.goic_sie.heal.action", true
	case "pl.zamykac_na_klucz.lock":
		return "pl.zamykac_na_klucz.lock.action", true
	case "pl.skladac_z_czesci.assemble":
		return "pl.skladac_z_czesci.assemble.action", true
	case "pl.targowac_sie.bargain":
		return "pl.targowac_sie.bargain.action", true
	case "pl.dawac_w_prezencie.gift":
		return "pl.dawac_w_prezencie.gift.action", true
	case "pl.dzielic_sie.share":
		return "pl.dzielic_sie.share.action", true
	case "pl.uczyc_sie.acquire_knowledge":
		return "pl.uczyc_sie.acquire_knowledge.action", true
	case "pl.pozyczac_komus.lend":
		return "pl.pozyczac_komus.lend.action", true
	case "pl.zamawiac_towary.order":
		return "pl.zamawiac_towary.order.action", true
	case "pl.pisac.type":
		return "pl.pisac.type.action", true
	case "pl.chciec.desire":
		return "pl.chciec.desire.action", true
	case "pl.ogladac_ekran.watch":
		return "pl.ogladac_ekran.watch.action", true
	case "pl.wyobrazac_sobie.imagine":
		return "pl.wyobrazac_sobie.imagine.action", true
	case "pl.miec_nadzieje.hope":
		return "pl.miec_nadzieje.hope.action", true
	case "pl.dostawac_zaplate.receive":
		return "pl.dostawac_zaplate.receive.action", true
	case "pl.zwracac_pieniadze.refund":
		return "pl.zwracac_pieniadze.refund.action", true
	case "pl.przelewac_pieniadze.transfer":
		return "pl.przelewac_pieniadze.transfer.action", true
	case "pl.nie_lubic.dislike":
		return "pl.nie_lubic.dislike.action", true
	case "pl.czerpac_przyjemnosc.enjoy":
		return "pl.czerpac_przyjemnosc.enjoy.action", true
	case "pl.dochodzic_do_siebie.recover":
		return "pl.dochodzic_do_siebie.recover.action", true
	case "pl.robic_zakupy.shop":
		return "pl.robic_zakupy.shop.action", true
	case "pl.dostac.receive_message":
		return "pl.dostac.receive_message.action", true
	case "pl.zazywac.medicine":
		return "pl.zazywac.medicine.action", true
	case "pl.pisac.document":
		return "pl.pisac.document.action", true
	case "pl.gotowac.boil":
		return "pl.gotowac.boil.action", true
	case "pl.zbierac.harvest":
		return "pl.zbierac.harvest.action", true
	case "pl.stwierdzac.state":
		return "pl.stwierdzac.state.action", true
	case "pl.nie_zgadzac_sie.disagree":
		return "pl.nie_zgadzac_sie.disagree.action", true
	case "pl.przyjmowac.accept_offer":
		return "pl.przyjmowac.accept_offer.action", true
	case "pl.wysylac.send_message":
		return "pl.wysylac.send_message.action", true
	case "pl.skarzyc_sie.complain":
		return "pl.skarzyc_sie.complain.action", true
	case "pl.zartowac.joke":
		return "pl.zartowac.joke.action", true
	case "pl.wolac.prefer":
		return "pl.wolac.prefer.action", true
	case "pl.probowac.try":
		return "pl.probowac.try.action", true
	case "pl.zgadzac_sie.agree":
		return "pl.zgadzac_sie.agree.action", true
	case "pl.odmawiac.refuse":
		return "pl.odmawiac.refuse.action", true
	case "pl.dyskutowac.discuss":
		return "pl.dyskutowac.discuss.action", true
	case "pl.uznawac.believe":
		return "pl.uznawac.believe.action", true
	case "pl.dziedziczyc.inherit":
		return "pl.dziedziczyc.inherit.action", true
	case "pl.wzrastac.increase":
		return "pl.wzrastac.increase.action", true
	case "pl.spadac.reduce":
		return "pl.spadac.reduce.action", true
	case "pl.krasc.steal":
		return "pl.krasc.steal.action", true
	case "pl.kategoryzowac.categorize":
		return "pl.kategoryzowac.categorize.action", true
	case "pl.rysowac.draw":
		return "pl.rysowac.draw.action", true
	case "pl.nagrywac.record":
		return "pl.nagrywac.record.action", true
	case "pl.wspolpracowac.cooperate":
		return "pl.wspolpracowac.cooperate.action", true
	case "pl.ubierac_sie.dress":
		return "pl.ubierac_sie.dress.action", true
	case "pl.porownywac.compare":
		return "pl.porownywac.compare.action", true
	case "pl.ranic.injure":
		return "pl.ranic.injure.action", true
	case "pl.sadzic.plant":
		return "pl.sadzic.plant.action", true
	case "pl.ostrzegac.warn":
		return "pl.ostrzegac.warn.action", true
	case "pl.polecac.recommend":
		return "pl.polecac.recommend.action", true
	case "pl.zaczynac.start":
		return "pl.zaczynac.start.action", true
	case "pl.konczyc.finish":
		return "pl.konczyc.finish.action", true
	case "pl.kontynuowac.continue":
		return "pl.kontynuowac.continue.action", true
	case "pl.informowac.inform":
		return "pl.informowac.inform.action", true
	case "pl.wyjasniac.explain":
		return "pl.wyjasniac.explain.action", true
	case "pl.opisywac.describe":
		return "pl.opisywac.describe.action", true
	case "pl.przedstawiac.introduce":
		return "pl.przedstawiac.introduce.action", true
	case "pl.prosic.request":
		return "pl.prosic.request.action", true
	case "pl.zapraszac.invite":
		return "pl.zapraszac.invite.action", true
	case "pl.oferowac.offer":
		return "pl.oferowac.offer.action", true
	case "pl.obiecywac.promise":
		return "pl.obiecywac.promise.action", true
	default:
		return "", false
	}
}
