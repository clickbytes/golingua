package pl

// additionalFrameForSense only accepts the explicit reviewed sense list.
func additionalFrameForSense(sense string) (string, bool) {
	switch sense {
	case "pl.plynac.swim":
		return "pl.plynac.swim.reviewed", true
	case "pl.leiec.fly":
		return "pl.leiec.fly.reviewed", true
	case "pl.podnosic_sie.rise":
		return "pl.podnosic_sie.rise.reviewed", true
	case "pl.zblizac_sie.approach":
		return "pl.zblizac_sie.approach.reviewed", true
	case "pl.oddalac_sie.recede":
		return "pl.oddalac_sie.recede.reviewed", true
	case "pl.dojezdzac.commute":
		return "pl.dojezdzac.commute.reviewed", true
	case "pl.jechac.ride":
		return "pl.jechac.ride.reviewed", true
	case "pl.startowac.takeoff":
		return "pl.startowac.takeoff.reviewed", true
	case "pl.ladowac.land":
		return "pl.ladowac.land.reviewed", true
	case "pl.manipulowac.manipulate":
		return "pl.manipulowac.manipulate.reviewed", true
	case "pl.odblokowywac.unlock":
		return "pl.odblokowywac.unlock.reviewed", true
	case "pl.zamarzac.freeze":
		return "pl.zamarzac.freeze.reviewed", true
	case "pl.topniec.melt":
		return "pl.topniec.melt.reviewed", true
	case "pl.palic_sie.burn":
		return "pl.palic_sie.burn.reviewed", true
	case "pl.sluchac.listen":
		return "pl.sluchac.listen.reviewed", true
	case "pl.wachac.smell":
		return "pl.wachac.smell.reviewed", true
	case "pl.skosztowac.taste":
		return "pl.skosztowac.taste.reviewed", true
	case "pl.dotykac.touch":
		return "pl.dotykac.touch.reviewed", true
	case "pl.dostrzegac.notice":
		return "pl.dostrzegac.notice.reviewed", true
	case "pl.rozpoznawac.recognize":
		return "pl.rozpoznawac.recognize.reviewed", true
	case "pl.obserwowac.observe":
		return "pl.obserwowac.observe.reviewed", true
	case "pl.monitorowac.monitor":
		return "pl.monitorowac.monitor.reviewed", true
	case "pl.myslec.think":
		return "pl.myslec.think.reviewed", true
	case "pl.pamietac.remember":
		return "pl.pamietac.remember.reviewed", true
	case "pl.zapominac.forget":
		return "pl.zapominac.forget.reviewed", true
	case "pl.snic.dream":
		return "pl.snic.dream.reviewed", true
	case "pl.podejrzewac.suspect":
		return "pl.podejrzewac.suspect.reviewed", true
	case "pl.rozumiec.understand":
		return "pl.rozumiec.understand.reviewed", true
	case "pl.wnioskowac.reason":
		return "pl.wnioskowac.reason.reviewed", true
	case "pl.szacowac.estimate":
		return "pl.szacowac.estimate.reviewed", true
	case "pl.obliczac.calculate":
		return "pl.obliczac.calculate.reviewed", true
	case "pl.decydowac.decide":
		return "pl.decydowac.decide.reviewed", true
	case "pl.wybierac.choose":
		return "pl.wybierac.choose.reviewed", true
	case "pl.planowac.plan":
		return "pl.planowac.plan.reviewed", true
	case "pl.przewidywac.anticipate":
		return "pl.przewidywac.anticipate.reviewed", true
	case "pl.oczekiwac.expect":
		return "pl.oczekiwac.expect.reviewed", true
	case "pl.mowic.speak":
		return "pl.mowic.speak.reviewed", true
	case "pl.opowiadac.narrate":
		return "pl.opowiadac.narrate.reviewed", true
	case "pl.oglaszac.announce":
		return "pl.oglaszac.announce.reviewed", true
	case "pl.nazywac.name":
		return "pl.nazywac.name.reviewed", true
	case "pl.sprzeczac_sie.argue":
		return "pl.sprzeczac_sie.argue.reviewed", true
	case "pl.krzyczec.shout":
		return "pl.krzyczec.shout.reviewed", true
	case "pl.pisac.write":
		return "pl.pisac.write.reviewed", true
	case "pl.tlumaczyc.translate":
		return "pl.tlumaczyc.translate.reviewed", true
	case "pl.kapac_sie.bathe":
		return "pl.kapac_sie.bathe.reviewed", true
	case "pl.zmieniac_sie.change":
		return "pl.zmieniac_sie.change.reviewed", true
	case "pl.grac.play_game":
		return "pl.grac.play_game.reviewed", true
	case "pl.pracowac.work":
		return "pl.pracowac.work.reviewed", true
	case "pl.klac.place":
		return "pl.klac.place.reviewed", true
	case "pl.wkladac.insert":
		return "pl.wkladac.insert.reviewed", true
	case "pl.wyjmowac.extract":
		return "pl.wyjmowac.extract.reviewed", true
	case "pl.laczyc.connect":
		return "pl.laczyc.connect.reviewed", true
	case "pl.rozlaczac.disconnect":
		return "pl.rozlaczac.disconnect.reviewed", true
	case "pl.napelniac.fill":
		return "pl.napelniac.fill.reviewed", true
	case "pl.oprozniac.empty":
		return "pl.oprozniac.empty.reviewed", true
	case "pl.lac.pour":
		return "pl.lac.pour.reviewed", true
	case "pl.mieszac.mix":
		return "pl.mieszac.mix.reviewed", true
	case "pl.montowac.install":
		return "pl.montowac.install.reviewed", true
	case "pl.wymieniac.replace":
		return "pl.wymieniac.replace.reviewed", true
	case "pl.parkowac.park":
		return "pl.parkowac.park.reviewed", true
	case "pl.przykrywac.cover":
		return "pl.przykrywac.cover.reviewed", true
	case "pl.kupowac.buy":
		return "pl.kupowac.buy.reviewed", true
	case "pl.kupic.buy":
		return "pl.kupic.buy.reviewed", true
	case "pl.sprzedawac.sell":
		return "pl.sprzedawac.sell.reviewed", true
	case "pl.placic.pay":
		return "pl.placic.pay.reviewed", true
	case "pl.pozyczac.borrow":
		return "pl.pozyczac.borrow.reviewed", true
	case "pl.zwracac.return":
		return "pl.zwracac.return.reviewed", true
	case "pl.wynajmowac.rent":
		return "pl.wynajmowac.rent.reviewed", true
	case "pl.darowac.donate":
		return "pl.darowac.donate.reviewed", true
	case "pl.dostarczac.deliver":
		return "pl.dostarczac.deliver.reviewed", true
	case "pl.wysylac.ship":
		return "pl.wysylac.ship.reviewed", true
	default:
		return "", false
	}
}
