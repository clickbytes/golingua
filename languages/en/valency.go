package en

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

var englishValencySource = provenance.SourceRef{
	ID: "golingua-en-valency-authored", Path: "languages/en/valency.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

var englishConstructionReviewSource = provenance.SourceRef{
	ID:      "golingua-delegated-construction-review",
	URI:     "https://github.com/clickbytes/golingua/issues/63#issuecomment-6351925853",
	Version: "2026-10-22", License: "unknown", Author: "Codex by owner delegation",
}

var englishLookingConstructionReviewSource = provenance.SourceRef{
	ID:      "golingua-delegated-construction-review",
	URI:     "https://github.com/clickbytes/golingua/issues/201#issuecomment-6354741947",
	Version: "2026-10-23-g05-14", License: "unknown", Author: "Codex by owner delegation",
}

var englishOwnerMassObjectConstructionSource = provenance.SourceRef{
	ID:      "golingua-owner-selected-mass-object",
	URI:     "https://github.com/clickbytes/golingua/issues/66#issuecomment-6355642886",
	Version: "2026-10-23-s077-mass-object", License: "internal", Author: "project owner",
}

var englishIndefiniteObjectConstructionSource = provenance.SourceRef{
	ID:      "golingua-g10-08-indefinite-object-alternative",
	URI:     "https://github.com/clickbytes/golingua/issues/228#issuecomment-6356966127",
	Version: "2026-10-23-g10-08", License: "internal", Author: "Codex by owner delegation",
}

var englishDeterminerAlternativeConstructionSource = provenance.SourceRef{
	ID:      "golingua-g10-10-english-determiner-alternatives",
	URI:     "https://github.com/clickbytes/golingua/issues/233#issuecomment-6357295748",
	Version: "2026-10-23-g10-10", License: "internal", Author: "Codex by owner delegation",
}

var englishIndefiniteAllomorphConstructionSource = provenance.SourceRef{
	ID:      "golingua-g10-12-english-indefinite-allomorphs",
	URI:     "https://github.com/clickbytes/golingua/issues/237#issuecomment-6357579336",
	Version: "2026-10-23-g10-12", License: "internal", Author: "Codex by owner delegation",
}

var englishNegationConstructionSource = provenance.SourceRef{
	ID: "golingua-g10-17-present-negation", URI: "https://github.com/clickbytes/golingua/issues/247#issuecomment-6358687010",
	Version: "2026-10-23-g10-17", License: "internal", Author: "Codex by owner delegation",
}

var englishIdentificationNegationSource = provenance.SourceRef{
	ID: "golingua-en-identification-negation-322", URI: "https://github.com/clickbytes/golingua/issues/322",
	Version: "2026-10-24", License: "internal", Author: "Codex by owner delegation",
}

var englishAttributedSeeingSource = provenance.SourceRef{
	ID: "golingua-en-attributed-seeing-327", URI: "https://github.com/clickbytes/golingua/issues/327",
	Version: "2026-10-25", License: "internal", Author: "Codex by owner delegation",
}

var englishNegativeArticleConstructionSource = provenance.SourceRef{
	ID: "golingua-g10-18-negative-article-alternatives", URI: "https://github.com/clickbytes/golingua/issues/249#issuecomment-6359036189",
	Version: "2026-10-23-g10-18", License: "internal", Author: "Codex by owner delegation",
}

var englishReadingPastConstructionSource = provenance.SourceRef{
	ID: "golingua-g10-20-reading-past-route", URI: "https://github.com/clickbytes/golingua/issues/253#issuecomment-6359693771",
	Version: "2026-10-23-g10-20", License: "internal", Author: "Codex by owner delegation",
}

var englishBarePluralConstructionSource = provenance.SourceRef{
	ID: "golingua-en-bare-plural-determiners-359", URI: "https://github.com/clickbytes/golingua/issues/359",
	Version: "2026-10-25-359", License: "internal", Author: "Codex by owner delegation",
}

var englishSeeingPastConstructionSource = provenance.SourceRef{
	ID: "golingua-g11-r3-past-seeing-357", URI: "https://github.com/clickbytes/golingua/issues/357",
	Version: "2026-10-25", License: "internal", Author: "Codex by owner delegation",
}

var englishWHQuestionSource = provenance.SourceRef{
	ID: "golingua-en-wh-questions-395", URI: "https://github.com/clickbytes/golingua/issues/395",
	Version: "2026-10-27", License: "internal", Author: "Codex by owner delegation",
}

type frameDeclaration struct {
	id, sense, family string
	roles             []meaning.Role
	arguments         []valency.ArgumentSlot
}

type constructionBindingDeclaration struct {
	sense, frame, family, variant string
}

var englishConstructionBindingDeclarations = append(append(append(append(append(append([]constructionBindingDeclaration{
	{"en.go.spatial_movement", "en.go.spatial_movement.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.go.spatial_movement", "en.go.spatial_movement.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.go.spatial_movement", "en.go.spatial_movement.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.go.spatial_movement", "en.go.spatial_movement.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.waking_up.event", "en.waking_up.intransitive", "en.intransitive", "en.intransitive.sv.wake_up.indefinite_subject"},
	{"en.waking_up.event", "en.waking_up.intransitive", "en.intransitive", "en.intransitive.sv.wake_up.indefinite_subject_an"},
	{"en.waking_up.event", "en.waking_up.intransitive", "en.intransitive", "en.intransitive.sv.wake_up.bare_subject_plural"},
	{"en.falling_asleep.event", "en.falling_asleep.intransitive", "en.intransitive", "en.intransitive.sv.fall_asleep.indefinite_subject"},
	{"en.falling_asleep.event", "en.falling_asleep.intransitive", "en.intransitive", "en.intransitive.sv.fall_asleep.indefinite_subject_an"},
	{"en.falling_asleep.event", "en.falling_asleep.intransitive", "en.intransitive", "en.intransitive.sv.fall_asleep.bare_subject_plural"},
	{"en.waking_up.event", "en.waking_up.intransitive", "en.intransitive", "en.intransitive.sv.wake_up"},
	{"en.falling_asleep.event", "en.falling_asleep.intransitive", "en.intransitive", "en.intransitive.sv.fall_asleep"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.sneezing.event", "en.sneezing.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.sneezing.event", "en.sneezing.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.sneezing.event", "en.sneezing.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.sneezing.event", "en.sneezing.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.coughing.event", "en.coughing.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.coughing.event", "en.coughing.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.coughing.event", "en.coughing.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.coughing.event", "en.coughing.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.walking.event", "en.walking.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.walking.event", "en.walking.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.walking.event", "en.walking.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.walking.event", "en.walking.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.living.event", "en.living.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.living.event", "en.living.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.living.event", "en.living.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.living.event", "en.living.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.flying_motion.event", "en.flying_motion.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.flying_motion.event", "en.flying_motion.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.flying_motion.event", "en.flying_motion.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.flying_motion.event", "en.flying_motion.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.working.activity", "en.working_activity.intransitive", "en.intransitive", "en.intransitive.sv"},
	{"en.working.activity", "en.working_activity.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject"},
	{"en.working.activity", "en.working_activity.intransitive", "en.intransitive", "en.intransitive.sv.indefinite_subject_an"},
	{"en.working.activity", "en.working_activity.intransitive", "en.intransitive", "en.intransitive.sv.bare_subject_plural"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.mass_object"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.indefinite_object"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.indefinite_object_an"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.indefinite_subject"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_mass_object"},
	{"en.eating.event", "en.eating.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_mass_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.mass_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.indefinite_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.indefinite_object_an"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.indefinite_subject"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_mass_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_mass_object"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.closing_object.event", "en.closing_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.carrying.event", "en.carrying.transitive", "en.transitive", "en.transitive.svo"},
	{"en.holding_object.event", "en.holding_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.pulling_object.event", "en.pulling_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.pushing_object.event", "en.pushing_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.cutting_object.event", "en.cutting_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.folding_object.event", "en.folding_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.tearing_object.event", "en.tearing_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.object.echo"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.subject.realize"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.subject.realize.plural"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.object.neutral"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.object.neutral.do"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.object.neutral.analyze"},
	{"en.see.visual", "en.see.visual.transitive", "en.question.wh", "en.question.wh.object.neutral.do.analyze"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject_a_object_an"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.this"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.these"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.those"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.this.mass"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.this.an"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.that"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.that.an"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.that.mass"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.it"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.it.mass"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification", "en.identification.it.an"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification_negative", "en.identification_negative.this"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification_negative", "en.identification_negative.that"},
	{"en.be.class_membership", "en.be.class_membership.frame", "en.identification_negative", "en.identification_negative.it"},
	{"en.black.color", "en.black.color.frame", "en.attribution", "en.attribution.this"},
	{"en.black.color", "en.black.color.frame", "en.attribution", "en.attribution.that"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", "en.transitive_negative.svo"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural"},
	{"en.liking.event", "en.liking.transitive", "en.transitive_negative", "en.transitive_negative.svo"},
	{"en.liking.event", "en.liking.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural"},
	{"en.watching.event", "en.watching.transitive", "en.transitive_negative", "en.transitive_negative.svo"},
	{"en.watching.event", "en.watching.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_personal", "en.transitive_personal.svo.first"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_personal", "en.transitive_personal.svo.second"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_personal", "en.transitive_personal.svo.first"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_personal", "en.transitive_personal.svo.second"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.first"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.second"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.reading.first"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.reading.second"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object_an"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object_an"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object_an"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.reading.first.indefinite_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.reading.first.indefinite_object_an"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.reading.second.indefinite_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_personal", "en.transitive_personal.svo.reading.second.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_a_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_a"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_mass_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.mass_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_a_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_a"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_mass_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", "en.transitive_negative.svo"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_past", "en.transitive_past.svo"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.mass_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_mass_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_object_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_a_object_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an_object_a"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an_object_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an_mass_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive", "en.transitive.svo.reading"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_target"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_target"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_target_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_a_target_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_an_target_a"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_an_target_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique_negative", "en.oblique_negative.svo"},
	{"en.look.directed", "en.look.directed.at", "en.oblique_negative", "en.oblique_negative.svo.plural"},
}, englishNegativeArticleBindingDeclarations()...), englishReadingPastBindingDeclarations()...), englishBarePluralBindingDeclarations()...), englishManipulationBindingDeclarations()...), englishAffectiveBindingDeclarations()...), englishWritingMessageBindingDeclarations()...)

func englishAffectiveBindingDeclarations() []constructionBindingDeclaration {
	suffixes := []string{
		"", ".indefinite_object", ".indefinite_object_an", ".indefinite_subject", ".indefinite_subject_an",
		".indefinite_subject_object", ".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a", ".indefinite_subject_an_object_an",
	}
	bindings := make([]constructionBindingDeclaration, 0, 27)
	for _, route := range []struct{ sense, frame string }{
		{"en.liking.event", "en.liking.transitive"},
		{"en.watching.event", "en.watching.transitive"},
		{"en.writing_document.event", "en.writing_document.transitive"},
	} {
		for _, suffix := range suffixes {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: route.sense, frame: route.frame, family: "en.transitive", variant: "en.transitive.svo" + suffix,
			})
		}
	}
	return bindings
}

func englishWritingMessageBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{
		"svo", "svo.indefinite_object", "svo.indefinite_object_an", "svo.indefinite_subject", "svo.indefinite_subject_an",
		"svo.indefinite_subject_object", "svo.indefinite_subject_a_object_an", "svo.indefinite_subject_an_object_a", "svo.indefinite_subject_an_object_an",
	} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive", "en.transitive." + variant})
	}
	for _, variant := range []string{
		"svo", "svo.indefinite_target", "svo.indefinite_target_an", "svo.indefinite_subject", "svo.indefinite_subject_an",
		"svo.indefinite_subject_target", "svo.indefinite_subject_a_target_an", "svo.indefinite_subject_an_target_a", "svo.indefinite_subject_an_target_an",
	} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.recipient_only", "en.to_oblique", "en.to_oblique." + variant})
	}
	for _, subject := range []string{"the", "a", "an"} {
		for _, recipient := range []string{"the", "a", "an"} {
			for _, message := range []string{"the", "a", "an"} {
				variant := "en.give.transfer.to_pp." + subject + "_" + message + "_" + recipient
				bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.full", "en.give.transfer", variant})
			}
		}
	}
	return bindings
}

var englishReadingPresentIndefiniteSuffixes = [...]string{
	".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
	".indefinite_object_an", ".indefinite_subject_an",
	".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a", ".indefinite_subject_an_object_an",
}

func englishBarePluralBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	add := func(sense, frame, family, variant string) {
		bindings = append(bindings, constructionBindingDeclaration{sense, frame, family, variant})
	}
	for _, suffix := range englishReadingPresentIndefiniteSuffixes {
		add("en.read.text", "en.read.text.transitive", "en.transitive", "en.transitive.svo"+suffix)
	}
	for _, route := range []struct{ sense, frame string }{
		{"en.see.visual", "en.see.visual.transitive"},
		{"en.hear.auditory", "en.hear.auditory.transitive"},
		{"en.read.text", "en.read.text.transitive"},
	} {
		for _, variant := range []string{
			"en.transitive.svo.bare_object.subject_the", "en.transitive.svo.bare_object.subject_a", "en.transitive.svo.bare_object.subject_an",
			"en.transitive.svo.bare_subject.object_the", "en.transitive.svo.bare_subject.object_a", "en.transitive.svo.bare_subject.object_an", "en.transitive.svo.bare_subject.object_bare",
			"en.transitive.svo.bare_subject.analyze",
		} {
			add(route.sense, route.frame, "en.transitive", variant)
		}
	}
	for _, route := range []struct{ sense, frame string }{
		{"en.see.visual", "en.see.visual.transitive"},
		{"en.read.text", "en.read.text.transitive"},
	} {
		for _, variant := range []string{
			"en.transitive_past.svo.bare_object.subject_the", "en.transitive_past.svo.bare_object.subject_a", "en.transitive_past.svo.bare_object.subject_an",
			"en.transitive_past.svo.bare_subject.object_the", "en.transitive_past.svo.bare_subject.object_a", "en.transitive_past.svo.bare_subject.object_an", "en.transitive_past.svo.bare_subject.object_bare",
			"en.transitive_past.svo.bare_subject.analyze",
		} {
			add(route.sense, route.frame, "en.transitive_past", variant)
		}
	}
	for _, route := range []struct{ sense, frame string }{
		{"en.see.visual", "en.see.visual.transitive"},
		{"en.hear.auditory", "en.hear.auditory.transitive"},
		{"en.read.text", "en.read.text.transitive"},
	} {
		for _, variant := range []string{
			"en.transitive_personal.svo.first.bare_object", "en.transitive_personal.svo.second.bare_object",
		} {
			add(route.sense, route.frame, "en.transitive_personal", variant)
		}
	}
	for _, variant := range []string{
		"en.transitive_personal.svo.reading.first.bare_object", "en.transitive_personal.svo.reading.second.bare_object",
	} {
		add("en.read.text", "en.read.text.transitive", "en.transitive_personal", variant)
	}
	for _, route := range []struct{ sense, frame string }{
		{"en.see.visual", "en.see.visual.transitive"},
		{"en.hear.auditory", "en.hear.auditory.transitive"},
	} {
		for _, variant := range []string{
			"en.transitive_negative.svo.plural.bare_subject.analyze",
			"en.transitive_negative.svo.bare_object.subject_the", "en.transitive_negative.svo.bare_object.subject_a", "en.transitive_negative.svo.bare_object.subject_an",
			"en.transitive_negative.svo.plural.bare_object.subject_the", "en.transitive_negative.svo.plural.bare_object.subject_bare",
			"en.transitive_negative.svo.plural.subject_bare.object_the", "en.transitive_negative.svo.plural.subject_bare.object_a", "en.transitive_negative.svo.plural.subject_bare.object_an",
		} {
			add(route.sense, route.frame, "en.transitive_negative", variant)
		}
	}
	return bindings
}

func englishReadingPastBindingDeclarations() []constructionBindingDeclaration {
	suffixes := []string{
		"", ".mass_object", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
		".indefinite_subject_mass_object", ".indefinite_object_an", ".indefinite_subject_an",
		".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
		".indefinite_subject_an_object_an", ".indefinite_subject_an_mass_object",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(suffixes))
	for _, suffix := range suffixes {
		if suffix == "" {
			bindings = append(bindings,
				constructionBindingDeclaration{
					sense: "en.read.text", frame: "en.read.text.transitive", family: "en.transitive_past",
					variant: "en.transitive_past.svo.reading",
				},
				constructionBindingDeclaration{
					sense: "en.read.text", frame: "en.read.text.transitive", family: "en.transitive_past",
					variant: "en.transitive_past.svo.reading.realize",
				},
			)
			continue
		}
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.read.text", frame: "en.read.text.transitive", family: "en.transitive_past",
			variant: "en.transitive_past.svo" + suffix,
		})
	}
	return bindings
}

func englishGivingConstructionBindings() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, layout := range []string{"double_np", "to_pp"} {
		for _, subjectArticle := range []string{"the", "a", "an"} {
			for _, recipientArticle := range []string{"the", "a", "an"} {
				for _, themeArticle := range []string{"the", "a", "an"} {
					variant := "en.give.transfer." + layout + "." + subjectArticle + "_" + recipientArticle + "_" + themeArticle
					bindings = append(bindings, constructionBindingDeclaration{"en.give.transfer", "en.give.transfer", "en.give.transfer", variant})
				}
			}
		}
	}
	for _, article := range []string{"the", "a", "an"} {
		for _, theme := range []string{"the", "a", "an"} {
			for _, counterparty := range []string{"the", "a", "an"} {
				bindings = append(bindings,
					constructionBindingDeclaration{"en.buying.event", "en.buying.event.commerce", "en.give.transfer", "en.give.transfer.from_pp." + article + "_" + theme + "_" + counterparty},
					constructionBindingDeclaration{"en.selling.event", "en.selling.event.commerce", "en.give.transfer", "en.give.transfer.to_pp." + article + "_" + theme + "_" + counterparty},
				)
			}
		}
	}
	return bindings
}

func englishNegativeArticleBindingDeclarations() []constructionBindingDeclaration {
	type familyRoute struct {
		sense, frame, family string
		suffixes             []string
	}
	routes := []familyRoute{
		{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", []string{
			".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
			".indefinite_object_an", ".indefinite_subject_an",
			".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
			".indefinite_subject_an_object_an",
		}},
		{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", []string{
			".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
			".indefinite_object_an", ".indefinite_subject_an",
			".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
			".indefinite_subject_an_object_an",
		}},
		{"en.look.directed", "en.look.directed.at", "en.oblique_negative", []string{
			".indefinite_target", ".indefinite_subject", ".indefinite_subject_target",
			".indefinite_target_an", ".indefinite_subject_an",
			".indefinite_subject_a_target_an", ".indefinite_subject_an_target_a",
			".indefinite_subject_an_target_an",
		}},
		{"en.liking.event", "en.liking.transitive", "en.transitive_negative", []string{
			".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
			".indefinite_object_an", ".indefinite_subject_an",
			".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
			".indefinite_subject_an_object_an",
		}},
		{"en.watching.event", "en.watching.transitive", "en.transitive_negative", []string{
			".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
			".indefinite_object_an", ".indefinite_subject_an",
			".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
			".indefinite_subject_an_object_an",
		}},
	}
	bindings := make([]constructionBindingDeclaration, 0, 48)
	for _, route := range routes {
		for _, numberSuffix := range []string{"", ".plural"} {
			for _, suffix := range route.suffixes {
				bindings = append(bindings, constructionBindingDeclaration{
					sense: route.sense, frame: route.frame, family: route.family,
					variant: route.family + ".svo" + numberSuffix + suffix,
				})
			}
		}
	}
	return bindings
}

func englishManipulationBindingDeclarations() []constructionBindingDeclaration {
	routes := []struct{ sense, frame string }{
		{"en.opening_object.event", "en.opening_object.transitive"},
		{"en.closing_object.event", "en.closing_object.transitive"},
		{"en.carrying.event", "en.carrying.transitive"},
		{"en.holding_object.event", "en.holding_object.transitive"},
		{"en.pulling_object.event", "en.pulling_object.transitive"},
		{"en.pushing_object.event", "en.pushing_object.transitive"},
		{"en.cutting_object.event", "en.cutting_object.transitive"},
		{"en.folding_object.event", "en.folding_object.transitive"},
		{"en.tearing_object.event", "en.tearing_object.transitive"},
	}
	suffixes := []string{".mass_object", ".indefinite_object", ".indefinite_object_an", ".indefinite_subject", ".indefinite_subject_an", ".indefinite_subject_mass_object", ".indefinite_subject_an_mass_object"}
	bindings := make([]constructionBindingDeclaration, 0, len(routes)*len(suffixes))
	for _, route := range routes {
		for _, suffix := range suffixes {
			bindings = append(bindings, constructionBindingDeclaration{sense: route.sense, frame: route.frame, family: "en.transitive", variant: "en.transitive.svo" + suffix})
		}
	}
	return bindings
}

var englishFrames = []frameDeclaration{
	{"en.writing_message.message_only", "en.writing_message.event", "en.transitive", []meaning.Role{"writer", "message", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "message", Required: true, GrammaticalFunction: "object"},
		{ID: "recipient", Role: "recipient", Required: false, GrammaticalFunction: "oblique", Adposition: "to"},
	}},
	{"en.writing_message.recipient_only", "en.writing_message.event", "en.to_oblique", []meaning.Role{"writer", "message", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "recipient", Required: true, GrammaticalFunction: "oblique", Adposition: "to"},
		{ID: "object", Role: "message", Required: false, GrammaticalFunction: "object"},
	}},
	{"en.writing_message.full", "en.writing_message.event", "en.give.transfer", []meaning.Role{"writer", "message", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject"},
		{ID: "theme", Role: "message", Required: true, GrammaticalFunction: "object"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "oblique", Adposition: "to"},
	}},
	{"en.liking.transitive", "en.liking.event", "en.transitive", []meaning.Role{"experiencer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.watching.transitive", "en.watching.event", "en.transitive", []meaning.Role{"viewer", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "viewer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.writing_document.transitive", "en.writing_document.event", "en.transitive", []meaning.Role{"author", "document"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "author", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "document", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.go.spatial_movement.intransitive", "en.go.spatial_movement", "en.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject"}}},
	{"en.working_activity.intransitive", "en.working.activity", "en.intransitive", []meaning.Role{"worker"}, []valency.ArgumentSlot{{ID: "subject", Role: "worker", Required: true, GrammaticalFunction: "subject"}}},
	{"en.waking_up.intransitive", "en.waking_up.event", "en.intransitive", []meaning.Role{"person"}, []valency.ArgumentSlot{{ID: "subject", Role: "person", Required: true, GrammaticalFunction: "subject"}}},
	{"en.falling_asleep.intransitive", "en.falling_asleep.event", "en.intransitive", []meaning.Role{"sleeper"}, []valency.ArgumentSlot{{ID: "subject", Role: "sleeper", Required: true, GrammaticalFunction: "subject"}}},
	{"en.sleep.physiological.intransitive", "en.sleep.physiological", "en.intransitive", []meaning.Role{"sleeper"}, []valency.ArgumentSlot{{ID: "subject", Role: "sleeper", Required: true, GrammaticalFunction: "subject"}}},
	{"en.breathing.intransitive", "en.breathing.event", "en.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject"}}},
	{"en.sneezing.intransitive", "en.sneezing.event", "en.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject"}}},
	{"en.running.intransitive", "en.running.event", "en.intransitive", []meaning.Role{"runner"}, []valency.ArgumentSlot{{ID: "subject", Role: "runner", Required: true, GrammaticalFunction: "subject"}}},
	{"en.coughing.intransitive", "en.coughing.event", "en.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject"}}},
	{"en.walking.intransitive", "en.walking.event", "en.intransitive", []meaning.Role{"walker"}, []valency.ArgumentSlot{{ID: "subject", Role: "walker", Required: true, GrammaticalFunction: "subject"}}},
	{"en.living.intransitive", "en.living.event", "en.intransitive", []meaning.Role{"organism"}, []valency.ArgumentSlot{{ID: "subject", Role: "organism", Required: true, GrammaticalFunction: "subject"}}},
	{"en.flying_motion.intransitive", "en.flying_motion.event", "en.intransitive", []meaning.Role{"mover"}, []valency.ArgumentSlot{{ID: "subject", Role: "mover", Required: true, GrammaticalFunction: "subject"}}},
	{"en.opening_object.transitive", "en.opening_object.event", "en.transitive", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.closing_object.transitive", "en.closing_object.event", "en.transitive", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.carrying.transitive", "en.carrying.event", "en.transitive", []meaning.Role{"carrier", "load"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "carrier", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "load", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.holding_object.transitive", "en.holding_object.event", "en.transitive", []meaning.Role{"holder", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.pulling_object.transitive", "en.pulling_object.event", "en.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.pushing_object.transitive", "en.pushing_object.event", "en.transitive", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.cutting_object.transitive", "en.cutting_object.event", "en.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.folding_object.transitive", "en.folding_object.event", "en.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.tearing_object.transitive", "en.tearing_object.event", "en.transitive", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.eating.transitive", "en.eating.event", "en.transitive", []meaning.Role{"eater", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "eater", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.drinking.transitive", "en.drinking.event", "en.transitive", []meaning.Role{"drinker", "liquid"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "drinker", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "liquid", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.be.class_membership.frame", "en.be.class_membership", "en.identification", []meaning.Role{"member"}, []valency.ArgumentSlot{
		{ID: "member", Role: "member", Required: true, GrammaticalFunction: "complement"},
	}},
	{"en.black.color.frame", "en.black.color", "en.attribution", []meaning.Role{"bearer"}, []valency.ArgumentSlot{
		{ID: "bearer", Role: "bearer", Required: true, GrammaticalFunction: "subject"},
	}},
	{"en.see.visual.transitive", "en.see.visual", "en.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.look.directed.at", "en.look.directed", "en.oblique", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "target", Required: true, GrammaticalFunction: "oblique", Adposition: "at"},
	}},
	{"en.hear.auditory.transitive", "en.hear.auditory", "en.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.get.receive.transitive", "en.get.receive", "en.transitive_past", []meaning.Role{"recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.give.transfer", "en.give.transfer", "en.give.transfer", []meaning.Role{"giver", "recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "giver", Required: true, GrammaticalFunction: "subject"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object"},
		{ID: "theme", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.read.text.transitive", "en.read.text", "en.transitive", []meaning.Role{"reader", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "reader", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
}

var englishCommerceOmissionFrames = []frameDeclaration{
	{"en.buying.event.omit_seller", "en.buying.event", "en.transitive", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique"},
	}},
	{"en.selling.event.omit_buyer", "en.selling.event", "en.transitive", []meaning.Role{"seller", "goods", "buyer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "seller", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object"},
		{ID: "counterparty", Role: "buyer", Required: false, GrammaticalFunction: "oblique"},
	}},
}

var englishCommerceFullFrames = []frameDeclaration{
	{"en.buying.event.commerce", "en.buying.event", "en.give.transfer", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject"},
		{ID: "recipient", Role: "seller", Required: true, GrammaticalFunction: "oblique", Adposition: "from"},
		{ID: "theme", Role: "goods", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.selling.event.commerce", "en.selling.event", "en.give.transfer", []meaning.Role{"seller", "goods", "buyer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "seller", Required: true, GrammaticalFunction: "subject"},
		{ID: "recipient", Role: "buyer", Required: true, GrammaticalFunction: "oblique", Adposition: "to"},
		{ID: "theme", Role: "goods", Required: true, GrammaticalFunction: "object"},
	}},
}

var englishCommerceOmittedRoles = map[string][]meaning.Role{
	"en.buying.event.omit_seller":       {"seller"},
	"en.selling.event.omit_buyer":       {"buyer"},
	"en.writing_message.message_only":   {"recipient"},
	"en.writing_message.recipient_only": {"message"},
}

func englishCommerceOmissionBindings() []constructionBindingDeclaration {
	routes := []struct{ sense, frame string }{
		{"en.buying.event", "en.buying.event.omit_seller"},
		{"en.selling.event", "en.selling.event.omit_buyer"},
	}
	variants := []string{
		"svo", "svo.mass_object", "svo.indefinite_object", "svo.indefinite_subject", "svo.indefinite_subject_object", "svo.indefinite_subject_mass_object",
		"svo.indefinite_object_an", "svo.indefinite_subject_an", "svo.indefinite_subject_a_object_an", "svo.indefinite_subject_an_object_a", "svo.indefinite_subject_an_object_an", "svo.indefinite_subject_an_mass_object",
		"svo.bare_object.subject_the", "svo.bare_object.subject_a", "svo.bare_object.subject_an", "svo.bare_subject.object_the", "svo.bare_subject.object_a", "svo.bare_subject.object_an", "svo.bare_subject.object_bare",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(routes)*len(variants))
	for _, route := range routes {
		for _, variant := range variants {
			bindings = append(bindings, constructionBindingDeclaration{route.sense, route.frame, "en.transitive", "en.transitive." + variant})
		}
	}
	return bindings
}

// CompileValency constructs independently authored English argument frames
// and the construction bindings approved in the delegated review. The routes
// alone do not activate sentence realization.
func CompileValency() (*valency.Index, error) {
	if err := englishValencySource.Validate(); err != nil {
		return nil, fmt.Errorf("English valency source: %w", err)
	}
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	senses := make(map[string]struct{}, len(document.Senses))
	for _, sense := range document.Senses {
		senses[sense.ID.String()] = struct{}{}
	}
	declarations := append(append(append([]frameDeclaration(nil), englishFrames...), englishCommerceOmissionFrames...), englishCommerceFullFrames...)
	frames := make([]valency.ValencyFrame, 0, len(declarations))
	for _, item := range declarations {
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, err
		}
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("English frame %s: %w", item.id, err)
		}
		if _, ok := senses[senseID.String()]; !ok {
			return nil, fmt.Errorf("English frame %s has no lexical sense", item.id)
		}
		frame := valency.ValencyFrame{
			ID: valency.FrameID(item.id), SenseID: senseID,
			SemanticRoles:    append([]meaning.Role(nil), item.roles...),
			Arguments:        append([]valency.ArgumentSlot(nil), item.arguments...),
			AllowedFamilyIDs: []ids.ConstructionID{familyID}, Provenance: englishValencySource,
		}
		frame.SurfaceOmittedRoles = append([]meaning.Role(nil), englishCommerceOmittedRoles[item.id]...)
		if item.id == "en.writing_message.message_only" || item.id == "en.writing_message.recipient_only" || item.id == "en.writing_message.full" {
			message, err := ids.NewConceptID("message")
			if err != nil {
				return nil, fmt.Errorf("English writing-message selection concept: %w", err)
			}
			letter, err := ids.NewConceptID("letter_document")
			if err != nil {
				return nil, fmt.Errorf("English writing-message letter concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "writing_message_content", Role: "message", Strength: valency.ConstraintHard,
				AllowedConceptIDs: []ids.ConceptID{message, letter},
			}}
		}
		if item.id == "en.writing_document.transitive" {
			document, err := ids.NewConceptID("document")
			if err != nil {
				return nil, fmt.Errorf("English writing-document selection concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "writing_document_content", Role: "document", Strength: valency.ConstraintHard,
				AllowedConceptIDs: []ids.ConceptID{document},
			}}
		}
		if item.id == "en.living.intransitive" {
			meat, err := ids.NewConceptID("meat_food")
			if err != nil {
				return nil, fmt.Errorf("English living exclusion concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{ID: "living_organism_excludes_meat", Role: "organism", Strength: valency.ConstraintHard, ExcludedConceptIDs: []ids.ConceptID{meat}}}
		}
		if item.id == "en.be.class_membership.frame" || item.id == "en.black.color.frame" {
			frame.Provenance = englishIdentificationSource
		}
		switch item.id {
		case "en.be.class_membership.frame":
			negativeFamilyID, err := ids.NewConstructionID("en.identification_negative")
			if err != nil {
				return nil, fmt.Errorf("English identification negative frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "en.see.visual.transitive", "en.hear.auditory.transitive":
			personalFamilyID, err := ids.NewConstructionID("en.transitive_personal")
			if err != nil {
				return nil, fmt.Errorf("English personal family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, personalFamilyID)
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
			if item.id == "en.see.visual.transitive" {
				questionFamilyID, err := ids.NewConstructionID("en.question.wh")
				if err != nil {
					return nil, fmt.Errorf("English WH question family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, questionFamilyID)
			}
			if item.id == "en.see.visual.transitive" {
				pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
				if err != nil {
					return nil, fmt.Errorf("English past seeing frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
				attributedFamilyID, err := ids.NewConstructionID("en.attributed_seeing")
				if err != nil {
					return nil, fmt.Errorf("English attributed seeing frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, attributedFamilyID)
			}
		case "en.liking.transitive", "en.watching.transitive":
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative affective frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "en.read.text.transitive":
			personalFamilyID, err := ids.NewConstructionID("en.transitive_personal")
			if err != nil {
				return nil, fmt.Errorf("English personal family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, personalFamilyID)
			pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
			if err != nil {
				return nil, fmt.Errorf("English past reading frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
		case "en.look.directed.at":
			negativeFamilyID, err := ids.NewConstructionID("en.oblique_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative oblique frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		}
		frames = append(frames, frame)
	}
	bindingDeclarations := append(append(append([]constructionBindingDeclaration(nil), englishConstructionBindingDeclarations...), englishGivingConstructionBindings()...), englishCommerceOmissionBindings()...)
	bindings := make([]valency.ConstructionBinding, 0, len(bindingDeclarations))
	for _, item := range bindingDeclarations {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("English construction binding sense: %w", err)
		}
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, fmt.Errorf("English construction binding family: %w", err)
		}
		variantID, err := ids.NewConstructionID(item.variant)
		if err != nil {
			return nil, fmt.Errorf("English construction binding variant: %w", err)
		}
		bindings = append(bindings, valency.ConstructionBinding{
			SenseID: senseID, ValencyFrameID: valency.FrameID(item.frame),
			FamilyID: familyID, VariantID: variantID, Approved: true,
			Provenance: englishConstructionBindingProvenance(item),
		})
	}
	return valency.BuildIndexWithCandidates(frames, bindings, nil)
}

func englishConstructionBindingProvenance(item constructionBindingDeclaration) provenance.SourceRef {
	if item.family == "en.question.wh" {
		return englishWHQuestionSource
	}
	if item.sense == "en.read.text" && item.family == "en.transitive" {
		for _, suffix := range englishReadingPresentIndefiniteSuffixes {
			if item.variant == "en.transitive.svo"+suffix {
				return englishBarePluralConstructionSource
			}
		}
	}
	if strings.Contains(item.variant, "bare") {
		return englishBarePluralConstructionSource
	}
	if item.sense == "en.see.visual" && item.family == "en.transitive_past" {
		return englishSeeingPastConstructionSource
	}
	if item.family == "en.attributed_seeing" {
		return englishAttributedSeeingSource
	}
	if item.family == "en.identification_negative" {
		return englishIdentificationNegationSource
	}
	if item.sense == "en.be.class_membership" || item.sense == "en.black.color" {
		return englishIdentificationSource
	}
	if item.sense == "en.read.text" && item.family == "en.transitive_past" {
		return englishReadingPastConstructionSource
	}
	if strings.HasSuffix(item.family, "_negative") {
		if strings.Contains(item.variant, ".indefinite_") {
			return englishNegativeArticleConstructionSource
		}
		return englishNegationConstructionSource
	}
	variant := item.variant
	if suffix, isPastFamily := strings.CutPrefix(variant, "en.transitive_past."); isPastFamily {
		variant = "en.transitive." + suffix
	}
	if variant == "en.transitive.svo.mass_object" {
		return englishOwnerMassObjectConstructionSource
	}
	if variant == "en.transitive.svo.indefinite_object" && item.sense == "en.get.receive" {
		return englishIndefiniteObjectConstructionSource
	}
	if englishIndefiniteAllomorphVariant(variant) {
		return englishIndefiniteAllomorphConstructionSource
	}
	if englishDeterminerAlternativeVariant(variant) {
		return englishDeterminerAlternativeConstructionSource
	}
	return englishBindingProvenance(item.sense)
}

func englishIndefiniteAllomorphVariant(variant string) bool {
	switch variant {
	case "en.intransitive.sv.indefinite_subject_an",
		"en.transitive.svo.indefinite_object_an",
		"en.transitive.svo.indefinite_subject_an",
		"en.transitive.svo.indefinite_subject_a_object_an",
		"en.transitive.svo.indefinite_subject_an_object_a",
		"en.transitive.svo.indefinite_subject_an_object_an",
		"en.transitive.svo.indefinite_subject_an_mass_object",
		"en.oblique.svo.indefinite_target_an",
		"en.oblique.svo.indefinite_subject_an",
		"en.oblique.svo.indefinite_subject_a_target_an",
		"en.oblique.svo.indefinite_subject_an_target_a",
		"en.oblique.svo.indefinite_subject_an_target_an":
		return true
	default:
		return false
	}
}

func englishDeterminerAlternativeVariant(variant string) bool {
	switch variant {
	case "en.intransitive.sv.indefinite_subject",
		"en.transitive.svo.indefinite_object",
		"en.transitive.svo.indefinite_subject",
		"en.transitive.svo.indefinite_subject_object",
		"en.transitive.svo.indefinite_subject_mass_object",
		"en.oblique.svo.indefinite_target",
		"en.oblique.svo.indefinite_subject",
		"en.oblique.svo.indefinite_subject_target":
		return true
	default:
		return false
	}
}

func englishBindingProvenance(sense string) provenance.SourceRef {
	if sense == "en.look.directed" {
		return englishLookingConstructionReviewSource
	}
	return englishConstructionReviewSource
}
