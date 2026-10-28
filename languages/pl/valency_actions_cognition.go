package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// cognitionActionFrames conservatively licenses named CORE participants and Polish
// syntactic arguments. Template existence is NOT construction approval.
var cognitionActionFrames = []frameDeclaration{
	{"pl.pisac.write.message_only", "pl.pisac.write", "pl.transitive", []meaning.Role{"writer", "message", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "message", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "recipient", Required: false, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.pisac.write.recipient_only", "pl.pisac.write", "pl.to_oblique", []meaning.Role{"writer", "message", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "recipient", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
		{ID: "object", Role: "message", Required: false, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pisac.write.full", "pl.pisac.write", "pl.transitive_to", []meaning.Role{"writer", "message", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "message", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "recipient", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.plynac.swim.reviewed", "pl.plynac.swim", "pl.intransitive", []meaning.Role{"swimmer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "swimmer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.leiec.fly.reviewed", "pl.leiec.fly", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.podnosic_sie.rise.reviewed", "pl.podnosic_sie.rise", "pl.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.zblizac_sie.approach.reviewed", "pl.zblizac_sie.approach", "pl.oblique", []meaning.Role{"mover", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "target", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.oddalac_sie.recede.reviewed", "pl.oddalac_sie.recede", "pl.oblique", []meaning.Role{"mover", "reference"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "reference", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.dojezdzac.commute.reviewed", "pl.dojezdzac.commute", "pl.oblique", []meaning.Role{"traveler", "workplace"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "traveler", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "workplace", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.jechac.ride.reviewed", "pl.jechac.ride", "pl.transitive", []meaning.Role{"passenger", "vehicle"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "passenger", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "vehicle", Required: true, GrammaticalFunction: "object", Case: "instrumental"},
	}},
	{"pl.startowac.takeoff.reviewed", "pl.startowac.takeoff", "pl.oblique", []meaning.Role{"vehicle", "origin"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "vehicle", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "origin", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.ladowac.land.reviewed", "pl.ladowac.land", "pl.oblique", []meaning.Role{"vehicle", "destination"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "vehicle", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "destination", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "na"},
	}},
	{"pl.manipulowac.manipulate.reviewed", "pl.manipulowac.manipulate", "pl.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "instrumental"},
	}},
	{"pl.odblokowywac.unlock.reviewed", "pl.odblokowywac.unlock", "pl.transitive", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zamarzac.freeze.reviewed", "pl.zamarzac.freeze", "pl.intransitive", []meaning.Role{"substance"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "substance", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.topniec.melt.reviewed", "pl.topniec.melt", "pl.intransitive", []meaning.Role{"substance"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "substance", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.palic_sie.burn.reviewed", "pl.palic_sie.burn", "pl.intransitive", []meaning.Role{"material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "material", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.sluchac.listen.reviewed", "pl.sluchac.listen", "pl.transitive", []meaning.Role{"listener", "source"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "listener", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "source", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.wachac.smell.reviewed", "pl.wachac.smell", "pl.transitive", []meaning.Role{"experiencer", "source"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "source", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.skosztowac.taste.reviewed", "pl.skosztowac.taste", "pl.transitive", []meaning.Role{"experiencer", "substance"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "substance", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.dotykac.touch.reviewed", "pl.dotykac.touch", "pl.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.dostrzegac.notice.reviewed", "pl.dostrzegac.notice", "pl.transitive", []meaning.Role{"perceiver", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "perceiver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.rozpoznawac.recognize.reviewed", "pl.rozpoznawac.recognize", "pl.transitive", []meaning.Role{"recognizer", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recognizer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.obserwowac.observe.reviewed", "pl.obserwowac.observe", "pl.transitive", []meaning.Role{"observer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "observer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.monitorowac.monitor.reviewed", "pl.monitorowac.monitor", "pl.transitive", []meaning.Role{"observer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "observer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.myslec.think.reviewed", "pl.myslec.think", "pl.oblique", []meaning.Role{"thinker", "subject"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "thinker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "subject", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "o"},
	}},
	{"pl.pamietac.remember.reviewed", "pl.pamietac.remember", "pl.transitive", []meaning.Role{"rememberer", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "rememberer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "content", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.zapominac.forget.reviewed", "pl.zapominac.forget", "pl.oblique", []meaning.Role{"experiencer", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "content", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "o"},
	}},
	{"pl.snic.dream.reviewed", "pl.snic.dream", "pl.intransitive", []meaning.Role{"dreamer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "dreamer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.podejrzewac.suspect.reviewed", "pl.podejrzewac.suspect", "pl.transitive", []meaning.Role{"thinker", "proposition"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "thinker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "proposition", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.rozumiec.understand.reviewed", "pl.rozumiec.understand", "pl.transitive", []meaning.Role{"understander", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "understander", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "content", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.wnioskowac.reason.reviewed", "pl.wnioskowac.reason", "pl.oblique", []meaning.Role{"reasoner", "premise"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "reasoner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "premise", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.szacowac.estimate.reviewed", "pl.szacowac.estimate", "pl.transitive", []meaning.Role{"estimator", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "estimator", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.obliczac.calculate.reviewed", "pl.obliczac.calculate", "pl.transitive", []meaning.Role{"calculator", "values"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "calculator", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "values", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.decydowac.decide.reviewed", "pl.decydowac.decide", "pl.oblique", []meaning.Role{"decider", "choice"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "decider", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "choice", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "o"},
	}},
	{"pl.wybierac.choose.reviewed", "pl.wybierac.choose", "pl.transitive", []meaning.Role{"chooser", "option"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "chooser", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "option", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.planowac.plan.reviewed", "pl.planowac.plan", "pl.transitive", []meaning.Role{"planner", "goal"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "planner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goal", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.przewidywac.anticipate.reviewed", "pl.przewidywac.anticipate", "pl.transitive", []meaning.Role{"participant", "future_event"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "future_event", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.oczekiwac.expect.reviewed", "pl.oczekiwac.expect", "pl.transitive", []meaning.Role{"experiencer", "outcome"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "outcome", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.mowic.speak.reviewed", "pl.mowic.speak", "pl.oblique", []meaning.Role{"speaker", "listener"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "listener", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.opowiadac.narrate.reviewed", "pl.opowiadac.narrate", "pl.transitive", []meaning.Role{"narrator", "audience"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "narrator", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "audience", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.oglaszac.announce.reviewed", "pl.oglaszac.announce", "pl.transitive", []meaning.Role{"announcer", "audience"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "announcer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "audience", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.nazywac.name.reviewed", "pl.nazywac.name", "pl.transitive", []meaning.Role{"namer", "referent"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "namer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "referent", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.sprzeczac_sie.argue.reviewed", "pl.sprzeczac_sie.argue", "pl.oblique", []meaning.Role{"participant_a", "participant_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant_a", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "participant_b", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
	}},
	{"pl.krzyczec.shout.reviewed", "pl.krzyczec.shout", "pl.oblique", []meaning.Role{"speaker", "audience"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "speaker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "audience", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.pisac.write.reviewed", "pl.pisac.write", "pl.oblique", []meaning.Role{"writer", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "recipient", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.tlumaczyc.translate.reviewed", "pl.tlumaczyc.translate", "pl.transitive", []meaning.Role{"translator", "source_document"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "translator", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "source_document", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.kapac_sie.bathe.reviewed", "pl.kapac_sie.bathe", "pl.intransitive", []meaning.Role{"bather"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "bather", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.zmieniac_sie.change.reviewed", "pl.zmieniac_sie.change", "pl.intransitive", []meaning.Role{"entity"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "entity", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.grac.play_game.reviewed", "pl.grac.play_game", "pl.oblique", []meaning.Role{"player", "game"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "player", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "game", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "w"},
	}},
	{"pl.pracowac.work.reviewed", "pl.pracowac.work", "pl.oblique", []meaning.Role{"worker", "task"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "worker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "task", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "nad"},
	}},
}
