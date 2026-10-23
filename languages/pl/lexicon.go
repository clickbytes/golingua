package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

type lexicalMeaning struct {
	id, description, candidateCore string
}

type lexicalEntry struct {
	id, lemma string
	pos       lexicon.PartOfSpeech
	meanings  []lexicalMeaning
}

// These are locally authored lexical distinctions, not generated translations
// of CORE labels. Candidate links have no transfer authority.
var polishEntries = []lexicalEntry{
	{"pl.pies.n", "pies", "noun", []lexicalMeaning{{"pl.pies.animal", "domestic dog", "dog"}}},
	{"pl.kot.n", "kot", "noun", []lexicalMeaning{{"pl.kot.animal", "domestic cat", "cat"}}},
	{"pl.dziecko.n", "dziecko", "noun", []lexicalMeaning{{"pl.dziecko.person", "young person", "child"}, {"pl.dziecko.kinship", "a person in relation to their own parent, at any age", "child_of_person"}}},
	{"pl.nauczyciel.n", "nauczyciel", "noun", []lexicalMeaning{{"pl.nauczyciel.role", "person acting as instructor", "teacher"}}},
	{"pl.lekarz.n", "lekarz", "noun", []lexicalMeaning{{"pl.lekarz.role", "medical practitioner", "doctor"}}},
	{"pl.dom.n", "dom", "noun", []lexicalMeaning{{"pl.dom.building", "residential building", "house"}, {"pl.dom.household", "people living together as a household", "household"}}},
	{"pl.szkola.n", "szkoła", "noun", []lexicalMeaning{{"pl.szkola.facility", "building used for instruction", "school"}, {"pl.szkola.institution", "organization conducting structured teaching", "school_institution"}}},
	{"pl.ksiazka.n", "książka", "noun", []lexicalMeaning{{"pl.ksiazka.physical", "physical book for reading", "book"}, {"pl.ksiazka.work", "written work regardless of physical carrier", ""}}},
	{"pl.reka.n", "ręka", "noun", []lexicalMeaning{{"pl.reka.hand", "terminal grasping part", "hand"}, {"pl.reka.arm", "upper limb from shoulder to hand inclusive", "arm"}}},
	{"pl.oko.n", "oko", "noun", []lexicalMeaning{{"pl.oko.organ", "organ for vision", "eye"}}},
	{"pl.woda.n", "woda", "noun", []lexicalMeaning{{"pl.woda.substance", "water as a substance", "water"}}},
	{"pl.widziec.v", "widzieć", "verb", []lexicalMeaning{{"pl.widziec.visual", "perceive visually", "seeing"}}},
	{"pl.patrzec.v", "patrzeć", "verb", []lexicalMeaning{{"pl.patrzec.directed", "direct visual attention to a target", "looking"}}},
	{"pl.slyszec.v", "słyszeć", "verb", []lexicalMeaning{{"pl.slyszec.auditory", "perceive through hearing", "hearing"}}},
	{"pl.dac.v", "dać", "verb", []lexicalMeaning{{"pl.dac.transfer", "initiate transfer to a recipient", "giving"}}},
	{"pl.dostac.v", "dostać", "verb", []lexicalMeaning{{"pl.dostac.receive", "become recipient of a theme", "receiving"}, {"pl.dostac.receive_message", "Obtaining a message delivered from another participant.", "receiving_message"}}},
}

var polishLexicalSource = provenance.SourceRef{
	ID: "golingua-pl-authored", Path: "languages/pl/lexicon.go",
	Version: "2026-10-23-receiving", License: "unknown", Author: "golingua contributors",
}

var polishLexicalReviewBatch1Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/152#issuecomment-6352315888",
	Version: "2026-10-23-batch-01", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalSenses = map[string]struct{}{
	"pl.osoba.person": {}, "pl.dziecko.person": {}, "pl.dorosly.person": {},
	"pl.lekarz.role": {}, "pl.pacjent.medical": {}, "pl.zwierze.animal": {},
	"pl.pies.animal": {}, "pl.kot.animal": {}, "pl.roslina.organism": {},
	"pl.drzewo.plant": {}, "pl.ksiazka.physical": {}, "pl.drzwi.barrier": {},
	"pl.pojazd.conveyance": {}, "pl.miejsce.location": {}, "pl.dom.building": {},
	"pl.szkola.facility": {}, "pl.czesc_ciala.body_part": {}, "pl.reka.hand": {},
	"pl.oko.organ": {}, "pl.woda.substance": {}, "pl.jedzenie.food": {},
	"pl.zdarzenie.event": {}, "pl.percepcja.sense": {}, "pl.widziec.visual": {},
	"pl.patrzec.directed": {}, "pl.slyszec.auditory": {}, "pl.przybycie.arrival": {},
	"pl.dac.transfer": {}, "pl.dostac.receive": {}, "pl.komunikowanie.communication": {},
	"pl.pytac.request_info": {}, "pl.odpowiadac.answer": {}, "pl.czytac.interpret": {},
	"pl.uczyc_sie.acquire": {}, "pl.uczyc.instruct": {}, "pl.jesc.eat": {},
	"pl.spac.sleep": {}, "pl.cecha.property": {}, "pl.wielkosc.size": {},
	"pl.temperatura.thermal": {}, "pl.kolor.visual": {}, "pl.predkosc.movement": {},
	"pl.niemowle.infant": {},
}

var polishLexicalReviewBatch2Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/154#issuecomment-6352384524",
	Version: "2026-10-23-batch-02", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch2Senses = map[string]struct{}{
	"pl.nastolatek.person": {}, "pl.osoba_starsza.elderly": {}, "pl.rodzic.kinship": {},
	"pl.matka.kinship": {}, "pl.ojciec.kinship": {}, "pl.rodzic_adopcyjny.parent": {},
	"pl.dziecko.kinship": {}, "pl.brat.kinship": {}, "pl.siostra.kinship": {},
	"pl.sasiad.person": {}, "pl.znajomy.person": {}, "pl.nieznajomy.person": {},
	"pl.gosc.person": {}, "pl.gospodarz.person": {}, "pl.mieszkaniec.person": {},
	"pl.obywatel.person": {}, "pl.odwiedzajacy.person": {}, "pl.turysta.person": {},
	"pl.podroznik.person": {}, "pl.pasazer.person": {}, "pl.kierowca.person": {},
	"pl.zatrudniony.person": {}, "pl.pracodawca.person": {}, "pl.kierownik.manager": {},
	"pl.czlonek_zespolu.member": {}, "pl.klient.person": {}, "pl.nabywca.person": {},
	"pl.sprzedawca.person": {}, "pl.rolnik.person": {}, "pl.ogrodnik.person": {},
	"pl.kelner.person": {}, "pl.sprzatacz.person": {}, "pl.stolarz.craft": {},
}

var polishLexicalReviewBatch3Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/156#issuecomment-6352443719",
	Version: "2026-10-23-batch-03", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch3Senses = map[string]struct{}{
	"pl.elektryk.role": {}, "pl.inzynier.role": {}, "pl.programista.role": {},
	"pl.artysta.role": {}, "pl.muzyk.role": {}, "pl.autor.work": {},
	"pl.dziennikarz.role": {}, "pl.bibliotekarz.role": {}, "pl.policjant.role": {},
	"pl.strazak.role": {}, "pl.zolnierz.role": {}, "pl.sportowiec.role": {},
	"pl.opiekun.role": {}, "pl.dentysta.role": {}, "pl.tlumacz.interpreter": {},
	"pl.tlumacz.written": {}, "pl.technik.role": {}, "pl.kurier.role": {},
	"pl.grupa.people": {}, "pl.rodzina.group": {}, "pl.dom.household": {},
	"pl.spolecznosc.group": {}, "pl.tlum.people": {}, "pl.publicznosc.people": {},
	"pl.zespol.group": {}, "pl.klasa.group": {}, "pl.komisja.group": {},
	"pl.organizacja.group": {}, "pl.firma.company": {}, "pl.restauracja.business": {},
	"pl.bank.institution": {}, "pl.rzad.government": {}, "pl.szkola.institution": {},
	"pl.uniwersytet.institution": {}, "pl.szpital.institution": {},
	"pl.stowarzyszenie.group": {}, "pl.klub.group": {}, "pl.biblioteka.institution": {},
	"pl.policja.department": {}, "pl.ssak.animal": {}, "pl.ptak.animal": {},
	"pl.ryba.animal": {}, "pl.gad.animal": {}, "pl.plaz.animal": {},
}

var polishLexicalReviewBatch4Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/158#issuecomment-6352494254",
	Version: "2026-10-23-batch-04", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch4Senses = map[string]struct{}{
	"pl.owad.animal": {}, "pl.pajak.animal": {}, "pl.kon.animal": {},
	"pl.krowa.animal": {}, "pl.swinia.animal": {}, "pl.owca.animal": {},
	"pl.koza.animal": {}, "pl.krolik.animal": {}, "pl.mysz.animal": {},
	"pl.szczur.animal": {}, "pl.niedzwiedz.animal": {}, "pl.kaczka.animal": {},
	"pl.ges.animal": {}, "pl.zaba.animal": {}, "pl.waz.animal": {},
	"pl.jaszczurka.animal": {}, "pl.zolw.animal": {}, "pl.motyl.animal": {},
	"pl.pszczola.animal": {}, "pl.mrowka.animal": {}, "pl.mucha.animal": {},
	"pl.komar.animal": {}, "pl.krzew.plant": {}, "pl.trawa.plant": {},
	"pl.roslina_kwitnaca.flowering": {}, "pl.jablon.plant": {},
	"pl.korzen.organ": {}, "pl.lodyga.plant": {}, "pl.lisc.plant": {},
	"pl.kwiat.organ": {}, "pl.nasienie.plant": {}, "pl.owoc.organ": {},
	"pl.kora.plant": {}, "pl.pien.plant": {}, "pl.cialo.body": {},
	"pl.glowa.part": {}, "pl.twarz.part": {}, "pl.wlosy.part": {},
	"pl.ucho.part": {}, "pl.nos.part": {}, "pl.usta.part": {},
	"pl.warga.part": {}, "pl.jezyk.part": {}, "pl.zab.part": {},
	"pl.szyja.part": {},
}

var polishLexicalReviewBatch5Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/160#issuecomment-6352533739",
	Version: "2026-10-23-batch-05", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch5Senses = map[string]struct{}{
	"pl.ramie.part": {}, "pl.reka.arm": {}, "pl.lokiec.part": {}, "pl.palec.part": {},
	"pl.kciuk.part": {}, "pl.klatka_piersiowa.part": {}, "pl.plecy.part": {},
	"pl.brzuch.part": {}, "pl.biodro.part": {}, "pl.noga.part": {},
	"pl.kolano.part": {}, "pl.kostka.part": {}, "pl.stopa.part": {},
	"pl.palec_u_nogi.part": {}, "pl.skora.part": {}, "pl.miesien.part": {},
	"pl.kosc.part": {}, "pl.staw.part": {}, "pl.serce.organ": {},
	"pl.pluco.organ": {}, "pl.mozg.organ": {}, "pl.zoladek.organ": {},
	"pl.jelito.organ": {}, "pl.watroba.organ": {}, "pl.nerka.organ": {},
	"pl.pecherz.organ": {}, "pl.krew.substance": {}, "pl.zyla.part": {},
	"pl.arteria.part": {}, "pl.nerw.part": {}, "pl.powietrze.substance": {},
	"pl.tlen.element": {}, "pl.para.steam": {}, "pl.lod.state": {},
	"pl.woda_slodka.fresh": {}, "pl.woda_slona.saline": {}, "pl.gleba.soil": {},
	"pl.piasek.material": {}, "pl.bloto.mixture": {}, "pl.skala.material": {},
	"pl.kamien.jednostka": {}, "pl.drewno.material": {}, "pl.metal.material": {},
	"pl.zelazo.element": {}, "pl.miedz.element": {}, "pl.zloto.material": {},
	"pl.srebro.material": {}, "pl.szklo.material": {}, "pl.plastik.material": {},
}

var polishLexicalReviewBatch6Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/162#issuecomment-6352575592",
	Version: "2026-10-23-batch-06", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch6Senses = map[string]struct{}{
	"pl.guma.material": {}, "pl.papier.material": {}, "pl.tkanina.material": {},
	"pl.skora.material": {}, "pl.olej.substance": {}, "pl.paliwo.substance": {},
	"pl.wegiel.fuel": {}, "pl.benzyna.fuel": {}, "pl.swiatlo.physical": {},
	"pl.cien.region": {}, "pl.plomien.region": {}, "pl.dym.substance": {},
	"pl.popiol.substance": {}, "pl.kurz.substance": {}, "pl.powierzchnia.region": {},
	"pl.grunt.surface": {}, "pl.wzgorze.place": {}, "pl.gora.place": {},
	"pl.dolina.place": {}, "pl.wyspa.place": {}, "pl.kontynent.place": {},
	"pl.wybrzeze.place": {}, "pl.plaza.place": {}, "pl.pustynia.place": {},
	"pl.las.place": {}, "pl.laka.place": {}, "pl.pole.land": {},
	"pl.grunt_rolny.farmland": {}, "pl.ogrod.place": {}, "pl.park.place": {},
	"pl.rzeka.place": {}, "pl.strumien.place": {}, "pl.jezioro.place": {},
	"pl.staw_zbiornik.place": {}, "pl.morze.place": {}, "pl.ocean.place": {},
	"pl.wodospad.place": {}, "pl.studnia.artifact": {}, "pl.brzeg.place": {},
	"pl.niebo.region": {}, "pl.chmura.region": {}, "pl.kropla_deszczu.entity": {},
	"pl.platek_sniegu.entity": {}, "pl.gradzina.entity": {}, "pl.mgla.region": {},
	"pl.slonce.body": {}, "pl.ksiezyc.body": {}, "pl.gwiazda.body": {},
	"pl.planeta.body": {},
}

var polishLexicalReviewBatch7Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/164#issuecomment-6352606332",
	Version: "2026-10-23-batch-07", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch7Senses = map[string]struct{}{
	"pl.atmosfera.region": {}, "pl.ziemia_planeta.body": {}, "pl.wies.place": {},
	"pl.miasto.place": {}, "pl.kraj.territory": {}, "pl.region.area": {},
	"pl.dzielnica.area": {}, "pl.adres.location": {}, "pl.budynek.artifact": {},
	"pl.mieszkanie.place": {}, "pl.pokoj.place": {}, "pl.sypialnia.place": {},
	"pl.salon.place": {}, "pl.kuchnia.place": {}, "pl.lazienka.place": {},
	"pl.toaleta.place": {}, "pl.jadalnia.place": {}, "pl.sala_lekcyjna.place": {},
	"pl.poczekalnia.place": {}, "pl.korytarz.place": {}, "pl.wejscie.place": {},
	"pl.wyjscie.place": {}, "pl.schody.artifact": {}, "pl.winda.artifact": {},
	"pl.dach.artifact": {}, "pl.sciana.artifact": {}, "pl.sufit.artifact": {},
	"pl.podloga.artifact": {}, "pl.okno.artifact": {}, "pl.garaz.building": {},
	"pl.sklep.building": {}, "pl.szpital.building": {}, "pl.biblioteka.building": {},
	"pl.restauracja.building": {}, "pl.hotel.building": {}, "pl.muzeum.building": {},
	"pl.teatr.building": {}, "pl.kino.building": {}, "pl.kosciol.building": {},
	"pl.lotnisko.place": {}, "pl.dworzec.place": {}, "pl.parking.place": {},
	"pl.plac_zabaw.place": {}, "pl.boisko.place": {}, "pl.basen.artifact": {},
}

var polishLexicalReviewBatch8Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/166#issuecomment-6352633509",
	Version: "2026-10-23-batch-08", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch8Senses = map[string]struct{}{
	"pl.droga.route": {}, "pl.ulica.route": {}, "pl.chodnik.route": {},
	"pl.sciezka.route": {}, "pl.most.artifact": {}, "pl.tunel.artifact": {},
	"pl.przejscie_dla_pieszych.place": {}, "pl.skrzyzowanie.place": {}, "pl.mebel.artifact": {},
	"pl.stol.artifact": {}, "pl.biurko.artifact": {}, "pl.krzeslo.artifact": {},
	"pl.kanapa.artifact": {}, "pl.lozko.artifact": {}, "pl.szafa.artifact": {},
	"pl.polka.artifact": {}, "pl.szuflada.artifact": {}, "pl.lustro.artifact": {},
	"pl.lampa.artifact": {}, "pl.zarowka.artifact": {}, "pl.zegar.artifact": {},
	"pl.zegarek.artifact": {}, "pl.klucz.tool": {}, "pl.zamek.device": {},
	"pl.uchwyt.artifact": {}, "pl.pojemnik.artifact": {}, "pl.pudelko.artifact": {},
	"pl.torba.artifact": {}, "pl.plecak.artifact": {}, "pl.walizka.artifact": {},
	"pl.butelka.artifact": {}, "pl.sloik.artifact": {}, "pl.kubek.artifact": {},
	"pl.szklanka.vessel": {}, "pl.talerz.artifact": {}, "pl.miska.vessel": {},
	"pl.garnek.vessel": {}, "pl.patelnia.vessel": {}, "pl.czajnik.vessel": {},
	"pl.lyzka.artifact": {}, "pl.widelec.artifact": {}, "pl.noz.artifact": {},
	"pl.nozyczki.tool": {}, "pl.mlotek.tool": {}, "pl.srubokret.tool": {},
	"pl.pila.tool": {}, "pl.lina.artifact": {},
}

var polishLexicalReviewBatch9Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/168#issuecomment-6352662804",
	Version: "2026-10-23-batch-09", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch9Senses = map[string]struct{}{
	"pl.klej.substance": {}, "pl.szczotka.tool": {}, "pl.szczoteczka_do_zebow.tool": {},
	"pl.grzebien.tool": {}, "pl.recznik.artifact": {}, "pl.mydlo.substance": {},
	"pl.miotla.tool": {}, "pl.odkurzacz.artifact": {}, "pl.lodowka.artifact": {},
	"pl.zamrazarka.artifact": {}, "pl.kuchenka.artifact": {}, "pl.piekarnik.artifact": {},
	"pl.kuchenka_mikrofalowa.artifact": {}, "pl.pralka.artifact": {}, "pl.zmywarka.artifact": {},
	"pl.klimatyzator.artifact": {}, "pl.wentylator.artifact": {}, "pl.komputer.artifact": {},
	"pl.laptop.artifact": {}, "pl.tablet.artifact": {}, "pl.smartfon.artifact": {},
	"pl.telefon.device": {}, "pl.klawiatura.artifact": {}, "pl.myszka.device": {},
	"pl.monitor.device": {}, "pl.drukarka.artifact": {}, "pl.aparat.device": {},
	"pl.sluchawki.artifact": {}, "pl.mikrofon.artifact": {}, "pl.glosnik.device": {},
	"pl.telewizor.artifact": {}, "pl.radio.device": {}, "pl.pilot.remote": {},
	"pl.bateria.device": {}, "pl.kabel.artifact": {}, "pl.wtyczka.artifact": {},
	"pl.napoj.food": {}, "pl.posilek.food": {}, "pl.sniadanie.food": {},
	"pl.obiad.food": {}, "pl.kolacja.food": {}, "pl.chleb.food": {},
	"pl.ryz.food": {}, "pl.makaron.food": {}, "pl.ziemniak.food": {},
	"pl.warzywo.food": {}, "pl.owoc.food": {}, "pl.jablko.food": {},
	"pl.banan.food": {},
}

var polishLexicalReviewBatch10Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/170#issuecomment-6352689075",
	Version: "2026-10-23-batch-10", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch10Senses = map[string]struct{}{
	"pl.pomarancza.food": {}, "pl.cytryna.food": {}, "pl.pomidor.food": {},
	"pl.marchew.food": {}, "pl.cebula.food": {}, "pl.salata.food": {},
	"pl.mieso.food": {}, "pl.wolowina.food": {}, "pl.wieprzowina.food": {},
	"pl.kurczak.food": {}, "pl.ryba.food": {}, "pl.jajko.food": {},
	"pl.mleko.food": {}, "pl.ser.food": {}, "pl.jogurt.food": {},
	"pl.maslo.food": {}, "pl.zupa.food": {}, "pl.salatka.food": {},
	"pl.kanapka.food": {}, "pl.ciasto.food": {}, "pl.ciastko.food": {},
	"pl.cukier.food": {}, "pl.sol.food": {}, "pl.pieprz.food": {},
	"pl.kawa.food": {}, "pl.herbata.food": {}, "pl.sok.food": {},
	"pl.ubranie.artifact": {}, "pl.koszula.artifact": {}, "pl.t_shirt.artifact": {},
	"pl.sweter.artifact": {}, "pl.kurtka.artifact": {}, "pl.plaszcz.artifact": {},
	"pl.sukienka.artifact": {}, "pl.spodnica.artifact": {}, "pl.spodnie.artifact": {},
	"pl.szorty.artifact": {}, "pl.skarpetka.artifact": {}, "pl.but.artifact": {},
	"pl.rekawica.artifact": {}, "pl.szalik.artifact": {}, "pl.pasek.artifact": {},
	"pl.parasol.artifact": {}, "pl.okulary.artifact": {}, "pl.portfel.artifact": {},
	"pl.przemieszczanie_sie.motion": {}, "pl.biec.run": {},
}

var polishLexicalReviewBatch11Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/172#issuecomment-6352712039",
	Version: "2026-10-23-batch-11", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch11Senses = map[string]struct{}{
	"pl.czolgac_sie.crawl": {}, "pl.skakac.jump": {}, "pl.wspinac_sie.climb": {},
	"pl.leiec.fly": {}, "pl.spadac.fall": {}, "pl.podnosic_sie.rise": {},
	"pl.wchodzic.enter": {}, "pl.wychodzic.exit": {}, "pl.przechodzic.cross": {},
	"pl.zblizac_sie.approach": {}, "pl.oddalac_sie.recede": {}, "pl.wracac.return": {},
	"pl.podazac.follow": {}, "pl.gonic.chase": {}, "pl.skrecac.turn": {},
	"pl.ruszac.start": {}, "pl.zatrzymywac_sie.stop": {}, "pl.podrozowac.travel": {},
	"pl.przewozic.transport": {}, "pl.niesc.carry": {}, "pl.prowadzic.drive": {},
	"pl.wsiadac.board": {}, "pl.wysiadac.alight": {}, "pl.parkowac.park": {},
	"pl.startowac.takeoff": {}, "pl.ladowac.land": {}, "pl.czekac.wait": {},
	"pl.odpoczywac.rest": {}, "pl.celowe_dzialanie.intentional_action": {}, "pl.manipulowac.manipulate": {},
	"pl.dotykac_obj.touch": {}, "pl.chwycic.grasp": {}, "pl.trzymac.hold": {},
	"pl.puszczac.release": {}, "pl.podnosic.lift": {}, "pl.opuszczac.lower": {},
	"pl.klac.place": {}, "pl.wkladac.insert": {}, "pl.wyjmowac.extract": {},
	"pl.popychac.push": {}, "pl.ciagnac.pull": {}, "pl.obracac.rotate": {},
	"pl.uderzac.hit": {}, "pl.rzucac.throw": {}, "pl.lapac.catch": {},
}

var polishLexicalReviewBatch12Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/174#issuecomment-6352902802",
	Version: "2026-10-23-batch-12", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch12Senses = map[string]struct{}{
	"pl.upuszczac.drop": {}, "pl.kopac.kick": {}, "pl.otwierac_obiekt.open": {},
	"pl.zamykac_obiekt.close": {}, "pl.zamykac_na_klucz.lock": {}, "pl.odblokowywac.unlock": {},
	"pl.ciac.cut": {}, "pl.drzec.tear": {}, "pl.skladac.fold": {},
	"pl.rozciagac.stretch": {}, "pl.naprawiac.repair": {}, "pl.skladac_z_czesci.assemble": {},
	"pl.rozbierac.disassemble": {}, "pl.laczyc.connect": {}, "pl.rozlaczac.disconnect": {},
	"pl.przykrywac.cover": {}, "pl.napelniac.fill": {}, "pl.oprozniac.empty": {},
	"pl.lac.pour": {}, "pl.mieszac.mix": {}, "pl.szorowac.scrub": {},
	"pl.myc.wash": {}, "pl.suszyc.dry": {}, "pl.chlodzic.cool": {},
	"pl.topniec.melt": {}, "pl.palic_sie.burn": {}, "pl.podpalac.ignite": {},
	"pl.gasic.extinguish": {}, "pl.wlaczac.power_on": {}, "pl.wylaczac.power_off": {},
	"pl.ladowac_baterie.charge": {}, "pl.montowac.install": {}, "pl.wymieniac.replace": {},
	"pl.wyrzucac.discard": {}, "pl.kupowac.buy": {}, "pl.sprzedawac.sell": {},
	"pl.placic.pay": {}, "pl.dostawac_zaplate.receive": {}, "pl.targowac_sie.bargain": {},
	"pl.pozyczac.borrow": {}, "pl.pozyczac_komus.lend": {}, "pl.zwracac.return": {},
	"pl.wynajmowac.rent": {}, "pl.zamawiac_usluge.hire": {},
}

var polishLexicalReviewBatch13Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/176#issuecomment-6352936410",
	Version: "2026-10-23-batch-13", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch13Senses = map[string]struct{}{
	"pl.dawac_w_prezencie.gift": {}, "pl.krasc.steal": {}, "pl.gubic.lose": {},
	"pl.znajdowac.find": {}, "pl.szukac.search": {}, "pl.dostarczac.deliver": {},
	"pl.wysylac.ship": {}, "pl.wymieniac_sie.exchange": {}, "pl.dzielic_sie.share": {},
	"pl.rezerwowac.reserve": {}, "pl.anulowac.cancel": {}, "pl.zwracac_pieniadze.refund": {},
	"pl.oszczedzac.save": {}, "pl.wydawac.spend": {}, "pl.zarabiac.earn": {},
	"pl.przelewac_pieniadze.transfer": {}, "pl.dziedziczyc.inherit": {}, "pl.sluchac.listen": {},
	"pl.dostrzegac.notice": {}, "pl.rozpoznawac.recognize": {}, "pl.obserwowac.observe": {},
	"pl.ogladac_ekran.watch": {}, "pl.monitorowac.monitor": {}, "pl.myslec.think": {},
	"pl.pamietac.remember": {}, "pl.zapominac.forget": {}, "pl.wyobrazac_sobie.imagine": {},
	"pl.snic.dream": {}, "pl.podejrzewac.suspect": {}, "pl.rozumiec.understand": {},
	"pl.zle_zrozumiec.misunderstand": {}, "pl.porownywac.compare": {}, "pl.kategoryzowac.categorize": {},
	"pl.szacowac.estimate": {}, "pl.obliczac.calculate": {}, "pl.decydowac.decide": {},
	"pl.wybierac.choose": {}, "pl.planowac.plan": {}, "pl.oczekiwac.expect": {},
	"pl.miec_nadzieje.hope": {}, "pl.mowic.speak": {}, "pl.stwierdzac.state": {},
}

var polishLexicalReviewBatch14Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/178#issuecomment-6352966221",
	Version: "2026-10-23-batch-14", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch14Senses = map[string]struct{}{
	"pl.opowiadac.narrate": {}, "pl.oglaszac.announce": {}, "pl.informowac.inform": {},
	"pl.wyjasniac.explain": {}, "pl.opisywac.describe": {}, "pl.nazywac.name": {},
	"pl.przedstawiac.introduce": {}, "pl.witac.greet": {}, "pl.dziekowac.thank": {},
	"pl.przepraszac.apologize": {}, "pl.prosic.request": {}, "pl.rozkazywac.command": {},
	"pl.zapraszac.invite": {}, "pl.oferowac.offer": {}, "pl.obiecywac.promise": {},
	"pl.zgadzac_sie.agree": {}, "pl.nie_zgadzac_sie.disagree": {}, "pl.odmawiac.refuse": {},
	"pl.przyjmowac.accept_offer": {}, "pl.sprzeczac_sie.argue": {}, "pl.dyskutowac.discuss": {},
	"pl.przekonywac.persuade": {}, "pl.ostrzegac.warn": {}, "pl.doradzac.advise": {},
	"pl.polecac.recommend": {}, "pl.skarzyc_sie.complain": {}, "pl.chwalic.praise": {},
	"pl.krytykowac.criticize": {}, "pl.zartowac.joke": {}, "pl.szeptac.whisper": {},
	"pl.krzyczec.shout": {}, "pl.pisac.write": {}, "pl.wysylac.send_message": {},
	"pl.dostac.receive_message": {}, "pl.dzwonic.call": {}, "pl.tlumaczyc.translate": {},
	"pl.tlumaczyc_ustnie.interpret": {}, "pl.oddychac.breathe": {}, "pl.wdychac.inhale": {},
	"pl.wydychac.exhale": {}, "pl.kaszlec.cough": {}, "pl.kichac.sneeze": {},
	"pl.polykac.swallow": {}, "pl.pic.drink": {}, "pl.zuc.chew": {},
	"pl.trawic.digest": {}, "pl.pocic_sie.sweat": {}, "pl.krwawic.bleed": {},
	"pl.goic_sie.heal": {},
}

var polishLexicalReviewBatch15Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/180#issuecomment-6353080488",
	Version: "2026-10-23-batch-15", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch15Senses = map[string]struct{}{
	"pl.zasypiac.sleep": {}, "pl.budzic_sie.wake": {}, "pl.starzec_sie.aging": {},
	"pl.umierac.die": {}, "pl.ranic.injure": {}, "pl.leczyc.treat": {},
	"pl.diagnozowac.diagnose": {}, "pl.dochodzic_do_siebie.recover": {},
	"pl.pracowac.work": {}, "pl.produkowac.manufacture": {},
	"pl.projektowac.design": {}, "pl.rysowac.draw": {},
	"pl.fotografowac.photograph": {}, "pl.edytowac.edit": {},
	"pl.smazyc.fry": {}, "pl.sprzatac.clean": {}, "pl.ubierac_sie.dress": {},
	"pl.kapac_sie.bathe": {}, "pl.sadzic.plant": {},
	"pl.podlewac.water": {}, "pl.zbierac.harvest": {},
	"pl.dlugosc.measure": {}, "pl.szerokosc.measure": {},
	"pl.wysokosc.measure": {}, "pl.glebokosc.measure": {},
	"pl.grubosc.measure": {}, "pl.powierzchnia.area": {},
	"pl.objetosc.measure": {}, "pl.masa.measure": {}, "pl.ciezar.force": {},
	"pl.ksztalt.form": {}, "pl.krzywizna.form": {},
	"pl.okraglosc.form": {}, "pl.symetria.form": {},
}

var polishLexicalReviewBatch16Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/182#issuecomment-6353127583",
	Version: "2026-10-23-batch-16", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch16Senses = map[string]struct{}{
	"pl.gestosc.measure": {}, "pl.faktura.surface": {}, "pl.twardosc.hardness": {},
	"pl.elastycznosc.form": {}, "pl.przezroczystosc.form": {}, "pl.jasnosc.measure": {},
	"pl.glosnosc.measure": {}, "pl.wysokosc_dzwieku.measure": {}, "pl.gorzkosc.quality": {},
	"pl.slonosc.quality": {}, "pl.ostry_smak.quality": {}, "pl.wilgotnosc.measure": {},
	"pl.wilgotnosc_powietrza.measure": {}, "pl.czystosc.quality": {}, "pl.ostrosc.quality": {},
	"pl.palnosc.quality": {}, "pl.kwasowosc.quality": {}, "pl.cisnienie.measure": {},
	"pl.zawartosc_energii.measure": {}, "pl.trwalosc.quality": {}, "pl.wiek.quality": {},
	"pl.stan_zdrowia.quality": {}, "pl.sila.body": {}, "pl.koordynacja.quality": {},
	"pl.poziom_umiejetnosci.quality": {}, "pl.poziom_wiedzy.quality": {}, "pl.poziom_uwagi.quality": {},
	"pl.pewnosc_siebie.quality": {}, "pl.poziom_satysfakcji.quality": {}, "pl.znaczenie.quality": {},
	"pl.przydatnosc.quality": {}, "pl.trudnosc.quality": {}, "pl.dostepnosc.quality": {},
	"pl.poziom_bezpieczenstwa.quality": {}, "pl.dostepnosc_zasobu.quality": {}, "pl.niezawodnosc.quality": {},
	"pl.dokladnosc.quality": {}, "pl.kompletnosc.quality": {}, "pl.relacja_przestrzenna.link": {},
	"pl.zawieranie_przestrzenne.link": {},
}

var polishLexicalReviewBatch17Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/184#issuecomment-6353162926",
	Version: "2026-10-23-batch-17", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch17Senses = map[string]struct{}{
	"pl.sasiedztwo.link": {}, "pl.stycznosc.link": {}, "pl.polozenie_pomiedzy.link": {},
	"pl.odleglosc.link": {}, "pl.bliskosc.link": {}, "pl.podpieranie.link": {},
	"pl.relacja_strukturalna.link": {}, "pl.przynaleznosc_czesci.link": {}, "pl.sklad_materialowy.link": {},
	"pl.rola_komponentu.link": {}, "pl.prawo_wlasnosci.link": {}, "pl.posiadanie_tymczasowe.link": {},
	"pl.pochodzenie.link": {}, "pl.relacja_spoleczna.link": {}, "pl.rodzicielstwo_biologiczne.link": {},
	"pl.rodzicielstwo_uznane.link": {}, "pl.rodzenstwo.link": {}, "pl.malzenstwo.link": {},
	"pl.przyjazn.link": {}, "pl.czlonkostwo.link": {}, "pl.zatrudnienie.link": {},
	"pl.relacja_czasowa.link": {}, "pl.rownoczesnosc.link": {}, "pl.nakladanie_czasowe.link": {},
	"pl.relacja_logiczna.link": {}, "pl.zaleznosc_przyczynowa.link": {}, "pl.zaleznosc_warunkowa.link": {},
	"pl.relacja_celu.link": {}, "pl.tresc_informacji.content": {}, "pl.tresc_twierdzenia.content": {},
	"pl.tresc_pytania.content": {},
}

var polishLexicalReviewBatch18Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/186#issuecomment-6353193500",
	Version: "2026-10-23-batch-18", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch18Senses = map[string]struct{}{
	"pl.tresc_odpowiedzi.content": {}, "pl.hipoteza.content": {}, "pl.tresc_przekonania.content": {},
	"pl.tresc_opinii.content": {}, "pl.tresc_instrukcji.content": {}, "pl.tresc_prosby.content": {},
	"pl.tresc_zasady.content": {}, "pl.tresc_planu.content": {}, "pl.cel_do_osiagniecia.content": {},
	"pl.tresc_dowodu.content": {}, "pl.tresc_wyjasnienia.content": {}, "pl.tresc_wiedzy.content": {},
	"pl.tresc_pamieci.content": {}, "pl.mozliwosc.modality": {}, "pl.koniecznosc.modality": {},
	"pl.zdolnosc.modality": {}, "pl.pozwolenie.modality": {}, "pl.obowiazek.modality": {},
	"pl.prawo.modality": {}, "pl.odpowiedzialnosc.modality": {}, "pl.zobowiazanie.modality": {},
	"pl.tresc_porozumienia.content": {}, "pl.tresc_umowy.content": {}, "pl.kategoria.class": {},
	"pl.ilosc.quantity": {}, "pl.liczba.value": {}, "pl.wartosc_pomiaru.value": {},
	"pl.jednostka_miary.unit": {}, "pl.czas_trwania.quantity": {}, "pl.moment.value": {},
	"pl.przedzial_czasowy.value": {}, "pl.cena.value": {}, "pl.kwota_pieniedzy.value": {},
	"pl.prawdopodobienstwo.value": {}, "pl.samochod.vehicle": {}, "pl.autobus.vehicle": {},
	"pl.pociag.vehicle": {}, "pl.rower.vehicle": {}, "pl.motocykl.vehicle": {},
	"pl.ciezarowka.vehicle": {}, "pl.samolot.vehicle": {}, "pl.lodz.vehicle": {},
	"pl.statek.vehicle": {}, "pl.taksowka.vehicle": {}, "pl.tramwaj.vehicle": {},
	"pl.metro.vehicle": {}, "pl.karetka.vehicle": {}, "pl.dokument.artifact": {},
	"pl.wiadomosc.artifact": {},
}

var polishLexicalReviewBatch19Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/188#issuecomment-6353224888",
	Version: "2026-10-23-batch-19", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch19Senses = map[string]struct{}{
	"pl.list.artifact": {}, "pl.e_mail.artifact": {}, "pl.notatnik.artifact": {},
	"pl.kartka_papieru.artifact": {}, "pl.olowek.artifact": {}, "pl.mapa.artifact": {},
	"pl.kalendarz.artifact": {}, "pl.fotografia.artifact": {}, "pl.rysunek.artifact": {},
	"pl.plik.artifact": {}, "pl.aplikacja.artifact": {}, "pl.pieniadze.medium": {},
	"pl.moneta.medium": {}, "pl.banknot.medium": {}, "pl.konto_bankowe.record": {},
	"pl.karta_kredytowa.medium": {}, "pl.karta_debetowa.medium": {}, "pl.bilet.credential": {},
	"pl.paszport.document": {}, "pl.legitymacja.document": {}, "pl.faktura.document": {},
	"pl.kasa.machine": {}, "pl.muszla_klozetowa.fixture": {}, "pl.prysznic.fixture": {},
	"pl.wanna.fixture": {}, "pl.zlew.fixture": {}, "pl.kran.fixture": {},
	"pl.pasta_do_zebow.paste": {}, "pl.szampon.preparation": {}, "pl.kosz_na_smieci.container": {},
	"pl.gabka.cleaner": {}, "pl.pizza.food": {}, "pl.lody.food": {},
	"pl.czekolada.food": {}, "pl.miod.food": {}, "pl.winogrono.food": {},
	"pl.stacja_benzynowa.facility": {}, "pl.bank.facility": {}, "pl.apteka.facility": {},
	"pl.chciec.desire": {}, "pl.lubic.like": {}, "pl.nie_lubic.dislike": {},
	"pl.potrzebowac.need": {}, "pl.wolac.prefer": {},
}

var polishLexicalReviewBatch20Source = provenance.SourceRef{
	ID: "golingua-pl-delegated-lexical-review", URI: "https://github.com/clickbytes/golingua/issues/190#issuecomment-6353253949",
	Version: "2026-10-23-batch-20", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedLexicalBatch20Senses = map[string]struct{}{
	"pl.kochac.love": {}, "pl.bac_sie.fear": {}, "pl.martwic_sie.worry": {},
	"pl.czerpac_przyjemnosc.enjoy": {}, "pl.pomagac.help": {}, "pl.wspolpracowac.cooperate": {},
	"pl.probowac.try": {}, "pl.udac_sie.succeed": {}, "pl.nie_udac_sie.fail": {},
	"pl.zaczynac.start": {}, "pl.konczyc.finish": {}, "pl.kontynuowac.continue": {},
	"pl.zmieniac_sie.change": {}, "pl.wzrastac.increase": {}, "pl.spadac.reduce": {},
	"pl.smiac_sie.laugh": {}, "pl.plakac.cry": {}, "pl.usmiechac_sie.smile": {},
	"pl.zyc.live": {}, "pl.tanczyc.dance": {}, "pl.stan_emocjonalny.state": {},
	"pl.radosc.state": {}, "pl.smutek.state": {}, "pl.strach.state": {},
	"pl.gniew.state": {}, "pl.milosc.state": {}, "pl.zaskoczenie.state": {},
	"pl.wstret.state": {}, "pl.nuda.state": {}, "pl.ciekawosc.state": {},
	"pl.glod.state": {}, "pl.pragnienie.state": {}, "pl.zmeczenie.state": {},
	"pl.bol.state": {}, "pl.istnienie.state": {}, "pl.opis_stanu.content": {},
	"pl.opis_zmiany.content": {}, "pl.dzien.interval": {}, "pl.tydzien.interval": {},
	"pl.miesiac.interval": {}, "pl.rok.interval": {}, "pl.poranek.interval": {},
	"pl.wieczor.interval": {}, "pl.noc.interval": {}, "pl.data.value": {},
	"pl.pora_roku.interval": {}, "pl.jednostka_waluty.value": {},
}

var polishPendingAuditSource = provenance.SourceRef{
	ID: "golingua-pl-delegated-pending-audit", URI: "https://github.com/clickbytes/golingua/issues/192#issuecomment-6353300680",
	Version: "2026-10-23-pending-audit", License: "unknown", Author: "codex-owner-delegated",
}

var polishApprovedPendingAuditSenses = map[string]struct{}{
	"pl.nauczyciel.role": {}, "pl.ruch.motion": {}, "pl.malzonek.spouse": {},
	"pl.partner.romantic": {}, "pl.pieszy.person": {}, "pl.slodycz.quality": {},
}

var polishFrameBySense = map[string]lexicon.ValencyID{
	"pl.widziec.visual":     "pl.widziec.visual.transitive",
	"pl.patrzec.directed":   "pl.patrzec.directed.na",
	"pl.slyszec.auditory":   "pl.slyszec.auditory.transitive",
	"pl.dostac.receive":     "pl.dostac.receive.transitive",
	"pl.dac.transfer":       "pl.dac.transfer.ditransitive",
	"pl.pytac.request_info": "pl.pytac.request_info.frame",
	"pl.odpowiadac.answer":  "pl.odpowiadac.answer.frame",
	"pl.czytac.interpret":   "pl.czytac.interpret.frame",
	"pl.uczyc.instruct":     "pl.uczyc.instruct.frame",
	"pl.uczyc_sie.acquire":  "pl.uczyc_sie.acquire.frame",
	"pl.jesc.eat":           "pl.jesc.eat.frame",
	"pl.spac.sleep":         "pl.spac.sleep.frame",
}

var polishParadigmByLexeme = map[string]string{
	"pl.kot.n":     "pl_masc_animal_regular",
	"pl.ptak.n":    "pl_masc_animal_regular",
	"pl.sasiad.n":  "pl_noun_attested_masculine_personal",
	"pl.klient.n":  "pl_noun_attested_masculine_personal",
	"pl.widziec.v": "pl_widziec_present_third_singular",
	"pl.slyszec.v": "pl_slyszec_present_third_singular",
	"pl.patrzec.v": "pl_patrzec_present_third_singular",
	"pl.drzewo.n":  "pl_noun_drzewo_attested",
	"pl.dostac.v":  "pl_verb_past_third_singular",
}

// LexicalDocument constructs the Polish lexical data from Go declarations.
// Pending candidate links preserve possible CORE relationships without
// licensing translation. This function does not invent inflection or valency.
func LexicalDocument() (lexicon.Document, error) {
	if err := polishLexicalSource.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical source: %w", err)
	}
	if err := polishLexicalReviewBatch1Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch2Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch3Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch4Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch5Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch6Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch7Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch8Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch9Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch10Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch11Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch12Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch13Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch14Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch15Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch16Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch17Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch18Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch19Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishLexicalReviewBatch20Source.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish lexical review source: %w", err)
	}
	if err := polishPendingAuditSource.Validate(); err != nil {
		return lexicon.Document{}, fmt.Errorf("Polish pending audit source: %w", err)
	}
	document := lexicon.Document{SchemaVersion: lexicon.SchemaVersion}
	groups := append([][]lexicalEntry{polishEntries}, lexicalGroups()...)
	for _, group := range groups {
		for _, entry := range group {
			lexemeID, err := ids.NewLexemeID(entry.id)
			if err != nil {
				return lexicon.Document{}, fmt.Errorf("Polish lexeme %s: %w", entry.id, err)
			}
			lexeme := lexicon.MorphLexeme{ID: lexemeID, Language: "pl", Lemma: entry.lemma, POS: entry.pos, ParadigmID: polishParadigmByLexeme[entry.id]}
			if gender, ok := attestedNounGender(entry.id); ok {
				lexeme.ParadigmID = "pl_noun_attested_" + gender
				lexeme.Properties = []lexicon.Property{{Key: "gender", Value: gender}}
			} else if entry.id == "pl.pieniadze.n" {
				lexeme.ParadigmID = "pl_noun_plural_only_attested"
				lexeme.Properties = []lexicon.Property{{Key: "grammatical_number", Value: "plural_only"}}
			} else if entry.id == "pl.drzewo.n" {
				lexeme.Properties = []lexicon.Property{{Key: "gender", Value: "neuter"}}
			} else if lexeme.ParadigmID == "pl_masc_animal_regular" {
				lexeme.Properties = []lexicon.Property{{Key: "gender", Value: "masculine_animal"}}
			}
			document.Lexemes = append(document.Lexemes, lexeme)
			for _, meaning := range entry.meanings {
				senseID, err := ids.NewSenseID(meaning.id)
				if err != nil {
					return lexicon.Document{}, fmt.Errorf("Polish sense %s: %w", meaning.id, err)
				}
				sense := lexicon.LexicalSense{ID: senseID, LexemeID: lexemeID, Usage: []string{meaning.description}}
				if frameID, ok := polishFrameBySense[meaning.id]; ok {
					sense.ValencyIDs = []lexicon.ValencyID{frameID}
				} else if frameID, ok := motionFrameForSense(meaning.id); ok {
					sense.ValencyIDs = []lexicon.ValencyID{lexicon.ValencyID(frameID)}
				} else if frameID, ok := actionFrameForSense(meaning.id); ok {
					sense.ValencyIDs = []lexicon.ValencyID{lexicon.ValencyID(frameID)}
				} else if frameID, ok := additionalFrameForSense(meaning.id); ok {
					sense.ValencyIDs = []lexicon.ValencyID{lexicon.ValencyID(frameID)}
				}
				document.Senses = append(document.Senses, sense)
				if meaning.candidateCore == "" {
					continue
				}
				conceptID, err := ids.NewConceptID(meaning.candidateCore)
				if err != nil {
					return lexicon.Document{}, fmt.Errorf("Polish candidate %s: %w", meaning.candidateCore, err)
				}
				binding := lexicon.SemanticBinding{
					SenseID: senseID, CoreRef: conceptID, Mapping: lexicon.MappingCandidate,
					Status: lexicon.BindingPending, Provenance: polishLexicalSource,
				}
				if _, approved := polishApprovedLexicalSenses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch1Source
				}
				if _, approved := polishApprovedLexicalBatch2Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch2Source
				}
				if _, approved := polishApprovedLexicalBatch3Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch3Source
				}
				if _, approved := polishApprovedLexicalBatch4Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch4Source
				}
				if _, approved := polishApprovedLexicalBatch5Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch5Source
				}
				if _, approved := polishApprovedLexicalBatch6Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch6Source
				}
				if _, approved := polishApprovedLexicalBatch7Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch7Source
				}
				if _, approved := polishApprovedLexicalBatch8Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch8Source
				}
				if _, approved := polishApprovedLexicalBatch9Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch9Source
				}
				if _, approved := polishApprovedLexicalBatch10Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch10Source
				}
				if _, approved := polishApprovedLexicalBatch11Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch11Source
				}
				if _, approved := polishApprovedLexicalBatch12Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch12Source
				}
				if _, approved := polishApprovedLexicalBatch13Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch13Source
				}
				if _, approved := polishApprovedLexicalBatch14Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch14Source
				}
				if _, approved := polishApprovedLexicalBatch15Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch15Source
				}
				if _, approved := polishApprovedLexicalBatch16Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch16Source
				}
				if _, approved := polishApprovedLexicalBatch17Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch17Source
				}
				if _, approved := polishApprovedLexicalBatch18Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch18Source
				}
				if _, approved := polishApprovedLexicalBatch19Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch19Source
				}
				if _, approved := polishApprovedLexicalBatch20Senses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishLexicalReviewBatch20Source
				}
				if _, approved := polishApprovedPendingAuditSenses[senseID.Symbol().Name]; approved {
					binding.Mapping = lexicon.MappingEquivalent
					binding.Status = lexicon.BindingApproved
					binding.Provenance = polishPendingAuditSource
				}
				document.Bindings = append(document.Bindings, binding)
			}
		}
	}
	return document, nil
}

// CompileLexicon validates the in-code Polish data against an exact CORE
// snapshot. The resulting index includes pending mappings for review; its
// TransferBindings view remains empty until separately approved data exists.
func CompileLexicon(ontology *core.ConceptGraph) (*lexicon.Index, error) {
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	return lexicon.Build(document, ontology, lexicon.DefaultBudget())
}
