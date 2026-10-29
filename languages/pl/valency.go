package pl

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

var polishValencySource = provenance.SourceRef{
	ID: "golingua-pl-valency-authored", Path: "languages/pl/valency.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

var polishConstructionReviewSource = provenance.SourceRef{
	ID:      "golingua-delegated-construction-review",
	URI:     "https://github.com/clickbytes/golingua/issues/63#issuecomment-6351925853",
	Version: "2026-10-22", License: "unknown", Author: "Codex by owner delegation",
}

var polishIssue436ConstructionSource = provenance.SourceRef{
	ID: "golingua-pl-help-thank-construction-436", URI: "https://github.com/clickbytes/golingua/issues/436",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var polishConversingConstructionSource = provenance.SourceRef{
	ID: "golingua-pl-conversing-construction-436", URI: "https://github.com/clickbytes/golingua/issues/436",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var polishNegationConstructionSource = provenance.SourceRef{
	ID: "golingua-g10-17-present-negation", URI: "https://github.com/clickbytes/golingua/issues/247#issuecomment-6358687010",
	Version: "2026-10-23-g10-17", License: "internal", Author: "Codex by owner delegation",
}

var polishIdentificationNegationSource = provenance.SourceRef{
	ID: "golingua-pl-identification-negation-322", URI: "https://github.com/clickbytes/golingua/issues/322",
	Version: "2026-10-24", License: "internal", Author: "Codex by owner delegation",
}

var polishAttributedSeeingSource = provenance.SourceRef{
	ID: "golingua-pl-attributed-seeing-327", URI: "https://github.com/clickbytes/golingua/issues/327",
	Version: "2026-10-25", License: "internal", Author: "Codex by owner delegation",
}

var polishWHQuestionSource = provenance.SourceRef{
	ID: "golingua-pl-wh-questions-395", URI: "https://github.com/clickbytes/golingua/issues/395",
	Version: "2026-10-27", License: "internal", Author: "Codex by owner delegation",
}

var polishReadingConstructionSource = provenance.SourceRef{
	ID: "golingua-g10-19-reading-route-review", URI: "https://github.com/clickbytes/golingua/issues/251#issuecomment-6359339236",
	Version: "2026-10-23-g10-19", License: "internal", Author: "Codex by owner delegation",
}

var polishPersonalConstructionSource = provenance.SourceRef{
	ID: "golingua-pl-personal-construction-358", Path: "languages/pl/constructions/transitive_personal.json",
	Version: "2026-10-25-358", License: "internal", Author: "Codex by owner delegation",
}

var polishPersonalExplicitConstructionSource = provenance.SourceRef{
	ID: "golingua-pl-personal-construction-explicit-358", Path: "languages/pl/constructions/transitive_personal_explicit.json",
	Version: "2026-10-25-358", License: "internal", Author: "Codex by owner delegation",
}

type frameDeclaration struct {
	id, sense, family string
	roles             []meaning.Role
	arguments         []valency.ArgumentSlot
}

type constructionBindingDeclaration struct {
	sense, frame, family, variant string
}

func polishAttributionConstructionBindings() []constructionBindingDeclaration {
	variants := []string{"ten", "ten.inanimate", "to", "ta", "tamten", "tamten.inanimate", "tamto", "tamta", "bare"}
	bindings := make([]constructionBindingDeclaration, 0, len(polishAttributionSpecs)*len(variants))
	for _, spec := range polishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		for _, variant := range variants {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: sense, frame: sense + ".frame", family: "pl.attribution", variant: "pl.attribution." + variant,
			})
		}
	}
	for _, spec := range polishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		for _, variant := range []string{"te", "ci", "tamte", "tamci", "bare"} {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: sense, frame: sense + ".plural.frame", family: "pl.attribution_plural", variant: "pl.attribution_plural." + variant,
			})
		}
	}
	return bindings
}

func polishAttributionFrames() []frameDeclaration {
	frames := make([]frameDeclaration, 0, len(polishAttributionSpecs))
	for _, spec := range polishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		frames = append(frames, frameDeclaration{
			id: sense + ".frame", sense: sense, family: "pl.attribution", roles: []meaning.Role{"bearer"},
			arguments: []valency.ArgumentSlot{{ID: "bearer", Role: "bearer", Required: true, GrammaticalFunction: "subject", Case: "nominative"}},
		})
	}
	for _, spec := range polishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		frames = append(frames, frameDeclaration{
			id: sense + ".plural.frame", sense: sense, family: "pl.attribution_plural", roles: []meaning.Role{"bearer"},
			arguments: []valency.ArgumentSlot{{ID: "bearer", Role: "bearer", Required: true, GrammaticalFunction: "subject", Case: "nominative"}},
		})
	}
	return frames
}

var polishConstructionBindingDeclarations = []constructionBindingDeclaration{
	{"pl.rozmawiac.converse", "pl.rozmawiac.converse.action", "pl.with_oblique", "pl.with_oblique.svo"},
	{"pl.uzywac.using", "pl.uzywac.using.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.pomagac.help", "pl.pomagac.help.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.pomagac.help", "pl.pomagac.help.action", "pl.transitive", "pl.transitive.ovs"},
	{"pl.dziekowac.thank", "pl.dziekowac.thank.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.dziekowac.thank", "pl.dziekowac.thank.action", "pl.transitive", "pl.transitive.ovs"},
	{"pl.zasypiac.sleep", "pl.zasypiac.sleep.action", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.budzic_sie.wake", "pl.budzic_sie.wake.action", "pl.intransitive", "pl.intransitive.sv.reflexive.period"},
	{"pl.spac.sleep", "pl.spac.sleep.frame", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.spac.sleep", "pl.spac.sleep.frame", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.oddychac.breathe", "pl.oddychac.breathe.action", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.oddychac.breathe", "pl.oddychac.breathe.action", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.kichac.sneeze", "pl.kichac.sneeze.action", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.kichac.sneeze", "pl.kichac.sneeze.action", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.biec.run", "pl.biec.run.motion", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.biec.run", "pl.biec.run.motion", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.kaszlec.cough", "pl.kaszlec.cough.action", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.kaszlec.cough", "pl.kaszlec.cough.action", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.chodzic.walk", "pl.chodzic.walk.motion", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.chodzic.walk", "pl.chodzic.walk.motion", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.isc.move", "pl.isc.move.motion", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.zyc.live", "pl.zyc.live.action", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.zyc.live", "pl.zyc.live.action", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.latac.fly", "pl.latac.fly.motion", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.latac.fly", "pl.latac.fly.motion", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.pracowac.activity", "pl.pracowac.activity.frame", "pl.intransitive", "pl.intransitive.sv"},
	{"pl.pracowac.activity", "pl.pracowac.activity.frame", "pl.intransitive", "pl.intransitive.sv.period"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.question.wh", "pl.question.wh.subject"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.question.wh", "pl.question.wh.object.fronted"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.question.wh", "pl.question.wh.subject.realize"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.question.wh", "pl.question.wh.object.realize"},
	{"pl.byc.class_membership", "pl.byc.class_membership.frame", "pl.identification", "pl.identification.to"},
	{"pl.byc.class_membership", "pl.byc.class_membership.frame", "pl.identification_negative", "pl.identification_negative.to"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive_negative", "pl.transitive_negative", "pl.transitive_negative.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive_negative", "pl.transitive_negative", "pl.transitive_negative.ovs"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.transitive_negative", "pl.transitive_negative", "pl.transitive_negative.svo"},
	{"pl.patrzec.directed", "pl.patrzec.directed.na", "pl.oblique_negative", "pl.oblique_negative.svo"},
	{"pl.lubic.like", "pl.lubic.like.negative", "pl.transitive_negative", "pl.transitive_negative.svo"},
	{"pl.ogladac.view", "pl.ogladac.view.negative", "pl.transitive_negative", "pl.transitive_negative.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.transitive", "pl.transitive.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.transitive", "pl.transitive.ovs"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo.ten"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo.tamten"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo.ta"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo.tamta"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo.to"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject", "pl.attributed_seeing", "pl.attributed_seeing.svo.tamto"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo.ten"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo.tamten"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo.ta"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo.tamta"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo.to"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing", "pl.attributed_seeing.svo.tamto"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.attributed_negative", "pl.attributed_seeing_object_negative", "pl.attributed_seeing_object_negative.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.attributed_negative", "pl.attributed_seeing_object_negative", "pl.attributed_seeing_object_negative.ovs"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.ovs"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.masculine_animal.tego"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.masculine_animal.tamtego"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.masculine_inanimate.ten"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.masculine_inanimate.tamten"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.masculine_personal.tego"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.masculine_personal.tamtego"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.feminine.te_acc"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.feminine.tamta_acc"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.neuter.to"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.neuter.tamto"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.plural_lexical"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.plural_lexical.te"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object", "pl.attributed_seeing_object.svo.plural_lexical.tamte"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both", "pl.attributed_seeing_both.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both", "pl.attributed_seeing_both.svo.object_plural_lexical"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_animal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_inanimate"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.feminine"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.neuter"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_personal.ci"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_personal.tamci"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.non_masculine_personal.te"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.non_masculine_personal.tamte"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_personal"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_animal"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_inanimate"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.feminine"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.neuter"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_personal.ci"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.masculine_personal.tamci"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.non_masculine_personal.te"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_subject_plural", "pl.attributed_seeing_plural", "pl.attributed_seeing_plural.svo.non_masculine_personal.tamte"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.ovs.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.ovs.non_masculine_personal"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_object_plural", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.ovs.masculine_personal"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.attributed_object_plural", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.ovs.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.masculine_personal.ci"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.masculine_personal.tamci"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.non_masculine_personal.te"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.non_masculine_personal.tamte"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.masculine_personal.tych"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_object_plural.svo.masculine_personal.tamtych"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.many.masculine_personal.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.many.masculine_personal.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.many.non_masculine_personal.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.many.non_masculine_personal.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.masculine_personal.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.masculine_personal.masculine_animal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.masculine_personal.masculine_inanimate"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.masculine_personal.feminine"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.masculine_personal.neuter"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.non_masculine_personal.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.non_masculine_personal.masculine_animal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.non_masculine_personal.masculine_inanimate"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.non_masculine_personal.feminine"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.many.one.non_masculine_personal.neuter"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.masculine_personal.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.masculine_personal.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.masculine_animal.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.masculine_animal.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.masculine_inanimate.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.masculine_inanimate.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.feminine.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.feminine.non_masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.neuter.masculine_personal"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.attributed_seeing_both_plural", "pl.attributed_seeing_both_plural.svo.one.many.neuter.non_masculine_personal"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.transitive", "pl.transitive", "pl.transitive.svo"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.transitive", "pl.transitive", "pl.transitive.ovs"},
	{"pl.dostac.receive", "pl.dostac.receive.transitive", "pl.transitive", "pl.transitive.svo"},
	{"pl.czytac.interpret", "pl.czytac.interpret.frame", "pl.transitive", "pl.transitive.svo"},
	{"pl.jesc.eat", "pl.jesc.eat.frame", "pl.transitive", "pl.transitive.svo"},
	{"pl.lubic.like", "pl.lubic.like.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.ogladac.view", "pl.ogladac.view.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.pisac.document", "pl.pisac.document.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.pisac.instrument", "pl.pisac.instrument.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.pisac.write", "pl.pisac.write.message_only", "pl.transitive", "pl.transitive.svo"},
	{"pl.pisac.write", "pl.pisac.write.recipient_only", "pl.to_oblique", "pl.to_oblique.svo"},
	{"pl.pisac.write", "pl.pisac.write.full", "pl.transitive_to", "pl.transitive_to.svoc"},
	{"pl.pic.drink", "pl.pic.drink.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.patrzec.directed", "pl.patrzec.directed.na", "pl.oblique", "pl.oblique.svo"},
	{"pl.otwierac_obiekt.open", "pl.otwierac_obiekt.open.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.zamykac_obiekt.close", "pl.zamykac_obiekt.close.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.niesc.carry", "pl.niesc.carry.motion", "pl.transitive", "pl.transitive.svo"},
	{"pl.trzymac.hold", "pl.trzymac.hold.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.ciagnac.pull", "pl.ciagnac.pull.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.pchac.push", "pl.pchac.push.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.ciac.cut", "pl.ciac.cut.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.kroic.with_instrument", "pl.kroic.with_instrument.action", "pl.transitive_instrumental", "pl.transitive_instrumental.svo"},
	{"pl.skladac.fold", "pl.skladac.fold.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.rozdzierac.tear", "pl.rozdzierac.tear.action", "pl.transitive", "pl.transitive.svo"},
	{"pl.dawac.transfer", "pl.dawac.transfer.ditransitive", "pl.ditransitive", "pl.ditransitive.svo"},
	{"pl.dawac.transfer", "pl.dawac.transfer.ditransitive", "pl.ditransitive", "pl.ditransitive.svo.period"},
	{"pl.dawac.transfer", "pl.dawac.transfer.ditransitive", "pl.ditransitive", "pl.ditransitive.recipient_front"},
	{"pl.dawac.transfer", "pl.dawac.transfer.ditransitive", "pl.ditransitive", "pl.ditransitive.recipient_front.period"},
	{"pl.dawac.transfer", "pl.dawac.transfer.ditransitive", "pl.ditransitive", "pl.ditransitive.object_front"},
	{"pl.dawac.transfer", "pl.dawac.transfer.ditransitive", "pl.ditransitive", "pl.ditransitive.object_front.period"},
	{"pl.sprzedawac.sell", "pl.sprzedawac.sell.reviewed", "pl.ditransitive", "pl.ditransitive.object_before_recipient"},
	{"pl.sprzedawac.sell", "pl.sprzedawac.sell.reviewed", "pl.ditransitive", "pl.ditransitive.object_before_recipient.period"},
}

// Families are explicit per frame, never silently assumed transitive.
// They are structural candidates, not approved construction bindings.
var polishFrames = []frameDeclaration{
	{"pl.uzywac.using.action", "pl.uzywac.using", "pl.transitive", []meaning.Role{"user", "tool"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "user", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "tool", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.byc.class_membership.frame", "pl.byc.class_membership", "pl.identification", []meaning.Role{"member"}, []valency.ArgumentSlot{
		{ID: "member", Role: "member", Required: true, GrammaticalFunction: "complement", Case: "nominative"},
	}},
	{"pl.widziec.visual.transitive", "pl.widziec.visual", "pl.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.widziec.visual.transitive_negative", "pl.widziec.visual", "pl.transitive_negative", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.widziec.visual.attributed_negative", "pl.widziec.visual", "pl.attributed_seeing_object_negative", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.patrzec.directed.na", "pl.patrzec.directed", "pl.oblique", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "target", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.slyszec.auditory.attributed_object_plural", "pl.slyszec.auditory", "pl.attributed_seeing_object_plural", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.slyszec.auditory.transitive", "pl.slyszec.auditory", "pl.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.slyszec.auditory.attributed_subject", "pl.slyszec.auditory", "pl.attributed_seeing", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.slyszec.auditory.attributed_subject_plural", "pl.slyszec.auditory", "pl.attributed_seeing_plural", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.slyszec.auditory.transitive_negative", "pl.slyszec.auditory", "pl.transitive_negative", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.lubic.like.negative", "pl.lubic.like", "pl.transitive_negative", []meaning.Role{"experiencer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.ogladac.view.negative", "pl.ogladac.view", "pl.transitive_negative", []meaning.Role{"viewer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "viewer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "genitive"},
	}},
	{"pl.dostac.receive.transitive", "pl.dostac.receive", "pl.transitive", []meaning.Role{"recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dac.transfer.ditransitive", "pl.dac.transfer", "pl.ditransitive", []meaning.Role{"giver", "recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "giver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "theme", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dawac.transfer.ditransitive", "pl.dawac.transfer", "pl.ditransitive", []meaning.Role{"giver", "recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "giver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pytac.request_info.frame", "pl.pytac.request_info", "pl.transitive", []meaning.Role{"sender", "addressee"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "addressee", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.odpowiadac.answer.frame", "pl.odpowiadac.answer", "pl.transitive", []meaning.Role{"sender", "addressee"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "addressee", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.czytac.interpret.frame", "pl.czytac.interpret", "pl.transitive", []meaning.Role{"reader", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "reader", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.uczyc.instruct.frame", "pl.uczyc.instruct", "pl.transitive", []meaning.Role{"instructor", "learner"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "instructor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "learner", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.uczyc_sie.acquire.frame", "pl.uczyc_sie.acquire", "pl.intransitive", []meaning.Role{"learner"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "learner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.jesc.eat.frame", "pl.jesc.eat", "pl.transitive", []meaning.Role{"eater", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "eater", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.spac.sleep.frame", "pl.spac.sleep", "pl.intransitive", []meaning.Role{"sleeper"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sleeper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.pracowac.activity.frame", "pl.pracowac.activity", "pl.intransitive", []meaning.Role{"worker"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "worker", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}

// CompileValency compiles sense-specific Polish frames and the construction
// bindings approved in the delegated review. These routes alone do not license
// sentence transfer or surface realization.
func CompileValency() (*valency.Index, error) {
	if err := polishValencySource.Validate(); err != nil {
		return nil, fmt.Errorf("Polish valency source: %w", err)
	}
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	senses := make(map[string]struct{}, len(document.Senses))
	for _, sense := range document.Senses {
		senses[sense.ID.String()] = struct{}{}
	}
	declarations := append(append([]frameDeclaration(nil), polishFrames...), polishAttributionFrames()...)
	declarations = append(declarations, polishMotionFrames...)
	declarations = append(declarations, physicalActionFrames...)
	declarations = append(declarations, activityActionFrames...)
	declarations = append(declarations, healthRoutineActionFrames...)
	declarations = append(declarations, productionActionFrames...)
	declarations = append(declarations, lifeActionFrames...)
	declarations = append(declarations, socialActionFrames...)
	declarations = append(declarations, communicationActionFrames...)
	declarations = append(declarations, communicationMoreFrames...)
	declarations = append(declarations, abstractActionFrames...)
	declarations = append(declarations, remainingActionFrames...)
	declarations = append(declarations, phraseLevelActionFrames...)
	declarations = append(declarations, transferCognitionFrames...)
	declarations = append(declarations, remainingActionFrames2...)
	declarations = append(declarations, exchangeActionFrames...)
	declarations = append(declarations, healthRemainingFrames...)
	declarations = append(declarations, cognitionActionFrames...)
	declarations = append(declarations, multiRoleActionFrames...)
	declarations = append(declarations, communicationLastFrames...)
	declarations = append(declarations, outcomeActionFrames...)
	declarations = append(declarations, polishCommerceOmissionFrames...)
	declarations = append(declarations, polishCommerceFullPurchaseFrames...)
	frames := make([]valency.ValencyFrame, 0, len(declarations))
	for _, item := range declarations {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("Polish frame %s: %w", item.id, err)
		}
		if _, ok := senses[senseID.String()]; !ok {
			return nil, fmt.Errorf("Polish frame %s has no lexical sense", item.id)
		}
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, fmt.Errorf("Polish frame %s family: %w", item.id, err)
		}
		frame := valency.ValencyFrame{
			ID: valency.FrameID(item.id), SenseID: senseID,
			SemanticRoles:    append([]meaning.Role(nil), item.roles...),
			Arguments:        append([]valency.ArgumentSlot(nil), item.arguments...),
			AllowedFamilyIDs: []ids.ConstructionID{familyID}, Provenance: polishValencySource,
		}
		if item.sense == "pl.otwarty.open" {
			person, err := ids.NewConceptID("person")
			if err != nil {
				return nil, fmt.Errorf("Polish physical-open person exclusion: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "physical_open_excludes_person", Role: "bearer", Strength: valency.ConstraintHard,
				ExcludedConceptIDs: []ids.ConceptID{person},
			}}
		}
		if item.id == "pl.pisac.write.message_only" || item.id == "pl.pisac.write.recipient_only" || item.id == "pl.pisac.write.full" {
			message, err := ids.NewConceptID("message")
			if err != nil {
				return nil, fmt.Errorf("Polish writing-message selection concept: %w", err)
			}
			letter, err := ids.NewConceptID("letter_document")
			if err != nil {
				return nil, fmt.Errorf("Polish writing-message letter concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "writing_message_content", Role: "message", Strength: valency.ConstraintHard,
				AllowedConceptIDs: []ids.ConceptID{message, letter},
			}}
		}
		if item.id == "pl.pisac.document.action" {
			document, err := ids.NewConceptID("document")
			if err != nil {
				return nil, fmt.Errorf("Polish writing-document selection concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "writing_document_content", Role: "document", Strength: valency.ConstraintHard,
				AllowedConceptIDs: []ids.ConceptID{document},
			}}
		}
		if item.id == "pl.pisac.instrument.action" {
			pencil, err := ids.NewConceptID("pencil")
			if err != nil {
				return nil, fmt.Errorf("Polish writing-instrument selection concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "writing_instrument_pencil", Role: "instrument", Strength: valency.ConstraintHard,
				AllowedConceptIDs: []ids.ConceptID{pencil},
			}}
		}
		frame.SurfaceOmittedRoles = append([]meaning.Role(nil), polishCommerceOmittedRoles[item.id]...)
		if item.id == "pl.zyc.live.action" {
			meat, err := ids.NewConceptID("meat_food")
			if err != nil {
				return nil, fmt.Errorf("Polish living exclusion concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{ID: "living_organism_excludes_meat", Role: "organism", Strength: valency.ConstraintHard, ExcludedConceptIDs: []ids.ConceptID{meat}}}
		}
		if item.id == "pl.widziec.visual.transitive" || item.id == "pl.slyszec.auditory.transitive" || item.id == "pl.czytac.interpret.frame" {
			for _, family := range []string{"pl.transitive_personal", "pl.transitive_personal_explicit"} {
				personalFamilyID, err := ids.NewConstructionID(family)
				if err != nil {
					return nil, fmt.Errorf("Polish personal construction family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, personalFamilyID)
			}
		}
		if item.id == "pl.byc.class_membership.frame" || strings.HasPrefix(item.id, "pl.") && strings.HasSuffix(item.id, ".frame") && func() bool {
			_, attributed := polishAttributionForSense("sense:" + strings.TrimSuffix(item.id, ".frame"))
			return attributed
		}() {
			frame.Provenance = polishIdentificationSource
		}
		if strings.HasSuffix(item.id, ".plural.frame") {
			frame.Provenance = polishIdentificationSource
		}
		switch item.id {
		case "pl.byc.class_membership.frame":
			negativeFamilyID, err := ids.NewConstructionID("pl.identification_negative")
			if err != nil {
				return nil, fmt.Errorf("Polish identification negative frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "pl.widziec.visual.transitive":
			for _, family := range []string{"pl.attributed_seeing", "pl.attributed_seeing_object", "pl.attributed_seeing_both", "pl.attributed_seeing_plural", "pl.attributed_seeing_object_plural", "pl.attributed_seeing_both_plural"} {
				attributedFamilyID, err := ids.NewConstructionID(family)
				if err != nil {
					return nil, fmt.Errorf("Polish attributed seeing frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, attributedFamilyID)
			}
			questionFamilyID, err := ids.NewConstructionID("pl.question.wh")
			if err != nil {
				return nil, fmt.Errorf("Polish WH question family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, questionFamilyID)
		case "pl.lubic.like.action", "pl.ogladac.view.action":
			negativeFamilyID, err := ids.NewConstructionID("pl.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("Polish negative affective frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "pl.patrzec.directed.na":
			negativeFamilyID, err := ids.NewConstructionID("pl.oblique_negative")
			if err != nil {
				return nil, fmt.Errorf("Polish negative oblique frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		}
		if item.id == "pl.slyszec.auditory.transitive" {
			familyID, err := ids.NewConstructionID("pl.attributed_seeing_object")
			if err != nil {
				return nil, fmt.Errorf("Polish hearing object-modifier construction family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, familyID)
		}
		frames = append(frames, frame)
	}
	bindingDeclarations := append([]constructionBindingDeclaration(nil), polishConstructionBindingDeclarations...)
	for _, item := range polishConstructionBindingDeclarations {
		if item.sense != "pl.widziec.visual" || item.frame != "pl.widziec.visual.transitive" || item.family != "pl.attributed_seeing_object" {
			continue
		}
		bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
			sense: "pl.slyszec.auditory", frame: "pl.slyszec.auditory.transitive", family: item.family, variant: item.variant,
		})
	}
	for _, base := range []string{"pl.attributed_seeing_both.svo", "pl.attributed_seeing_both.svo.object_plural_lexical"} {
		for _, determiner := range []struct {
			gender string
			forms  [2]string
		}{
			{"masculine_animal", [2]string{"ten", "tamten"}},
			{"masculine_inanimate", [2]string{"ten", "tamten"}},
			{"masculine_personal", [2]string{"ten", "tamten"}},
			{"feminine", [2]string{"ta", "tamta"}},
			{"neuter", [2]string{"to", "tamto"}},
		} {
			for _, form := range determiner.forms {
				bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
					sense: "pl.widziec.visual", frame: "pl.widziec.visual.transitive", family: "pl.attributed_seeing_both",
					variant: base + "." + determiner.gender + "." + form,
				})
			}
		}
	}
	for _, item := range polishConstructionBindingDeclarations {
		if item.family != "pl.attributed_seeing_both_plural" || !strings.Contains(item.variant, ".svo.many.") {
			continue
		}
		parts := strings.Split(strings.TrimPrefix(item.variant, "pl.attributed_seeing_both_plural.svo."), ".")
		if len(parts) != 4 {
			continue
		}
		forms := []string{"te", "tamte"}
		if parts[2] == "masculine_personal" {
			forms = []string{"ci", "tamci"}
		}
		for _, form := range forms {
			bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
				sense: "pl.widziec.visual", frame: "pl.widziec.visual.transitive", family: "pl.attributed_seeing_both_plural",
				variant: item.variant + "." + form,
			})
		}
	}
	bindingDeclarations = append(bindingDeclarations, polishAttributionConstructionBindings()...)
	bindingDeclarations = append(bindingDeclarations, polishPersonalConstructionBindings()...)
	bindingDeclarations = append(bindingDeclarations, polishCommerceOmissionBindings()...)
	bindingDeclarations = append(bindingDeclarations, polishCommerceFullPurchaseBindings()...)
	bindings := make([]valency.ConstructionBinding, 0, len(bindingDeclarations))
	for _, item := range bindingDeclarations {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("Polish construction binding sense: %w", err)
		}
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, fmt.Errorf("Polish construction binding family: %w", err)
		}
		variantID, err := ids.NewConstructionID(item.variant)
		if err != nil {
			return nil, fmt.Errorf("Polish construction binding variant: %w", err)
		}
		bindings = append(bindings, valency.ConstructionBinding{
			SenseID: senseID, ValencyFrameID: valency.FrameID(item.frame),
			FamilyID: familyID, VariantID: variantID, Approved: true,
			Provenance: polishConstructionBindingProvenance(item),
		})
	}
	return valency.BuildIndexWithCandidates(frames, bindings, nil)
}

func polishConstructionBindingProvenance(item constructionBindingDeclaration) provenance.SourceRef {
	if item.sense == "pl.rozmawiac.converse" {
		return polishConversingConstructionSource
	}
	if item.sense == "pl.pomagac.help" || item.sense == "pl.dziekowac.thank" || item.sense == "pl.uzywac.using" || item.sense == "pl.kroic.with_instrument" {
		return polishIssue436ConstructionSource
	}
	if item.family == "pl.question.wh" {
		return polishWHQuestionSource
	}
	if item.family == "pl.transitive_personal" {
		return polishPersonalConstructionSource
	}
	if item.family == "pl.transitive_personal_explicit" {
		return polishPersonalExplicitConstructionSource
	}
	if item.family == "pl.attributed_seeing" || item.family == "pl.attributed_seeing_object" || item.family == "pl.attributed_seeing_both" {
		return polishAttributedSeeingSource
	}
	if item.family == "pl.identification_negative" {
		return polishIdentificationNegationSource
	}
	if item.sense == "pl.byc.class_membership" {
		return polishIdentificationSource
	}
	if _, attributed := polishAttributionForSense("sense:" + item.sense); attributed {
		return polishPositiveAdjectiveSource
	}
	if item.sense == "pl.czytac.interpret" {
		return polishReadingConstructionSource
	}
	if item.family == "pl.transitive_negative" || item.family == "pl.oblique_negative" {
		return polishNegationConstructionSource
	}
	return polishConstructionReviewSource
}
