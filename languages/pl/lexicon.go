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
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
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
	"pl.kot.n": "pl_masc_animal_regular",
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
			} else if entry.id == "pl.kot.n" {
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
