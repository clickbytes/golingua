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

var englishPastAttributionConstructionSource = provenance.SourceRef{
	ID: "golingua-en-past-attribution-constructions-437", Path: "languages/en/valency.go",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var englishNegativeAttributionConstructionSource = provenance.SourceRef{
	ID: "golingua-en-negative-attribution-437", Path: "languages/en/constructions/attribution_negative.json",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var englishIssue436ConstructionSource = provenance.SourceRef{
	ID: "golingua-en-help-thank-construction-436", URI: "https://github.com/clickbytes/golingua/issues/436",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
}

var englishConversingConstructionSource = provenance.SourceRef{
	ID: "golingua-en-conversing-construction-436", URI: "https://github.com/clickbytes/golingua/issues/436",
	Version: "2026-10-29", License: "internal", Author: "Codex by owner delegation",
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

var englishAttributedSeeingDemonstrativesSource = provenance.SourceRef{
	ID: "golingua-en-attributed-seeing-demonstratives-435", URI: "https://github.com/clickbytes/golingua/issues/435",
	Version: "2026-10-28", License: "internal", Author: "Codex by owner delegation",
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

func englishAttributionConstructionBindings() []constructionBindingDeclaration {
	bindings := make([]constructionBindingDeclaration, 0, len(englishAttributionSpecs)*2)
	for _, spec := range englishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		for _, variant := range []string{"this", "that", "bare", "the", "the_plural", "a", "an"} {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: sense, frame: sense + ".frame", family: "en.attribution", variant: "en.attribution." + variant,
			})
		}
	}
	for _, spec := range englishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		for _, variant := range []string{"these", "those"} {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: sense, frame: sense + ".frame", family: "en.attribution", variant: "en.attribution." + variant,
			})
		}
	}
	for _, spec := range englishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		for _, variant := range []string{"this", "that", "bare", "the", "the_plural", "a", "an", "these", "those"} {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: sense, frame: sense + ".frame", family: "en.attribution", variant: "en.attribution." + variant + ".past",
			})
		}
		bindings = append(bindings, constructionBindingDeclaration{
			sense: sense, frame: sense + ".frame", family: "en.attribution", variant: "en.attribution.bare_plural.past",
		})
	}
	return bindings
}

func englishAttributedHearingBindings() []constructionBindingDeclaration {
	variants := []string{
		"en.attributed_seeing.svo", "en.attributed_seeing.svo.indefinite_object", "en.attributed_seeing.svo.indefinite_object_an",
		"en.attributed_seeing.svo.plural_subject", "en.attributed_seeing.svo.plural_subject.indefinite_object", "en.attributed_seeing.svo.plural_subject.indefinite_object_an",
		"en.attributed_seeing.svo.demonstrative_this", "en.attributed_seeing.svo.demonstrative_that", "en.attributed_seeing.svo.demonstrative_these", "en.attributed_seeing.svo.demonstrative_those",
		"en.attributed_seeing.svo.indefinite_subject", "en.attributed_seeing.svo.indefinite_subject_an", "en.attributed_seeing.svo.indefinite_subject_object", "en.attributed_seeing.svo.indefinite_subject_a_object_an",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.hear.auditory", frame: "en.hear.auditory.attributed_subject", family: "en.attributed_seeing", variant: variant})
	}
	return bindings
}

func englishAttributedHearingObjectPluralBindings() []constructionBindingDeclaration {
	variants := []string{"en.attributed_seeing_object.svo.plural_object", "en.attributed_seeing_object.svo.plural_object.indefinite_subject", "en.attributed_seeing_object.svo.plural_object.indefinite_subject_an"}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.hear.auditory", frame: "en.hear.auditory.attributed_object", family: "en.attributed_seeing_object", variant: variant})
	}
	return bindings
}

func englishAttributionFrames() []frameDeclaration {
	frames := make([]frameDeclaration, 0, len(englishAttributionSpecs))
	for _, spec := range englishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		frames = append(frames, frameDeclaration{
			id: sense + ".frame", sense: sense, family: "en.attribution", roles: []meaning.Role{"bearer"},
			arguments: []valency.ArgumentSlot{{ID: "bearer", Role: "bearer", Required: true, GrammaticalFunction: "subject"}},
		})
	}
	for _, spec := range englishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		frames = append(frames, frameDeclaration{
			id: sense + ".negative.frame", sense: sense, family: "en.attribution_negative", roles: []meaning.Role{"bearer"},
			arguments: []valency.ArgumentSlot{{ID: "bearer", Role: "bearer", Required: true, GrammaticalFunction: "subject"}},
		})
	}
	return frames
}

func englishNegativeAttributionBindings() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, spec := range englishAttributionSpecs {
		sense := strings.TrimPrefix(spec.sense, "sense:")
		for _, number := range []string{"singular", "plural"} {
			for _, tense := range []string{"present", "past"} {
				determiners := []string{"the", "a", "an"}
				if number == "plural" {
					determiners = []string{"the", "bare"}
				}
				for _, determiner := range determiners {
					bindings = append(bindings, constructionBindingDeclaration{
						sense: sense, frame: sense + ".negative.frame", family: "en.attribution_negative",
						variant: "en.attribution_negative." + number + "." + tense + "." + determiner,
					})
				}
			}
		}
	}
	return bindings
}

var englishConstructionBindingDeclarations = []constructionBindingDeclaration{
	{"en.have.possession", "en.have.possession.transitive", "en.transitive", "en.transitive.svo"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_negative", "en.transitive_negative.svo.vp_auxiliary"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_negative", "en.transitive_negative.svo.do.indefinite_object.analyze"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_negative", "en.transitive_negative.svo.does.indefinite_object.analyze"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_personal", "en.transitive_personal.svo.first"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_personal", "en.transitive_personal.svo.first.indefinite_object_an"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_personal", "en.transitive_personal.svo.second"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object"},
	{"en.have.possession", "en.have.possession.transitive", "en.transitive_personal", "en.transitive_personal.svo.second.indefinite_object_an"},
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
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular.indefinite_subject"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular.indefinite_subject_an"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.plural"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.plural.definite"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular.indefinite_subject"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular.indefinite_subject_an"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.plural"},
	{"en.sleep.physiological", "en.sleep.physiological.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.plural.definite"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular.indefinite_subject"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular.indefinite_subject_an"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.plural"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.plural.definite"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular.indefinite_subject"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular.indefinite_subject_an"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.plural"},
	{"en.breathing.event", "en.breathing.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.plural.definite"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular.indefinite_subject"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.singular.indefinite_subject_an"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.plural"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.plural.definite"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular.indefinite_subject"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.singular.indefinite_subject_an"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.plural"},
	{"en.running.event", "en.running.intransitive", "en.intransitive", "en.intransitive.sv.negative.past.plural.definite"},
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
	{"en.drinking.event", "en.drinking.transitive", "en.transitive_negative", "en.transitive_negative.svo.mass_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive_negative", "en.transitive_negative.svo.mass_object.indefinite_subject"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural.mass_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural.mass_object.indefinite_subject"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.mass_object"},
	{"en.drinking.event", "en.drinking.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.mass_object.indefinite_subject"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.mass_object"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.mass_object.indefinite_subject"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural.mass_object"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural.mass_object.indefinite_subject"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.mass_object"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.past"},
	{"en.eating.event", "en.eating.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.mass_object.indefinite_subject"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_object"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo.opening_object_present.indefinite_subject.the_plural"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo.opening_object_present.indefinite_subject.bare_plural"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo.opening_object_present.definite_subject.the_plural"},
	{"en.opening_object.event", "en.opening_object.transitive", "en.transitive", "en.transitive.svo.opening_object_present.definite_subject.bare_plural"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.indefinite_object"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.indefinite_subject"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_object"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.bare_object.subject_the"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.bare_object.subject_a"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.opening_object_past.the_plural"},
	{"en.opening_object.event", "en.opening_object.transitive.past", "en.transitive_past", "en.transitive_past.svo.opening_object_past.indefinite_subject.the_plural"},
	{"en.closing_object.event", "en.closing_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.carrying.event", "en.carrying.transitive", "en.transitive", "en.transitive.svo"},
	{"en.holding_object.event", "en.holding_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.pulling_object.event", "en.pulling_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.pushing_object.event", "en.pushing_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.cutting_object.event", "en.cutting_object.transitive", "en.transitive", "en.transitive.svo"},
	{"en.cutting_object.with_instrument", "en.cutting_object.with_instrument", "en.transitive_with_oblique", "en.transitive_with_oblique.svo"},
	{"en.cutting_object.with_instrument", "en.cutting_object.with_instrument", "en.transitive_with_oblique", "en.transitive_with_oblique.svo.indefinite_target"},
	{"en.cutting_object.with_instrument", "en.cutting_object.with_instrument", "en.transitive_with_oblique", "en.transitive_with_oblique.svo.indefinite_subject"},
	{"en.cutting_object.with_instrument", "en.cutting_object.with_instrument", "en.transitive_with_oblique", "en.transitive_with_oblique.svo.indefinite_subject_target"},
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
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.plural_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.plural_subject.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.plural_subject.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.demonstrative_this"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.demonstrative_that"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.demonstrative_these"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.demonstrative_those"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing", "en.attributed_seeing.svo.indefinite_subject_a_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo"},
	{"en.see.visual", "en.see.visual.attributed_negative", "en.attributed_seeing_object_negative", "en.attributed_seeing_object_negative.source.present.svo"},
	{"en.see.visual", "en.see.visual.attributed_negative", "en.attributed_seeing_object_negative", "en.attributed_seeing_object_negative.source.past.svo"},
	{"en.see.visual", "en.see.visual.attributed_negative", "en.attributed_seeing_object_negative", "en.attributed_seeing_object_negative.svo.vp_auxiliary"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.this_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.that_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.these_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.those_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.plural_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.plural_object.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.plural_object.indefinite_subject_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.indefinite_subject_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.indefinite_subject_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_object", "en.attributed_seeing_object.svo.indefinite_subject_a_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_subject_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_subject_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_subject_an_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_subject_a_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.indefinite_subject_an_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.plural_both"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.plural_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.attributed_seeing_both", "en.attributed_seeing_both.svo.plural_object"},
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
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", "en.transitive_negative.svo.past"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_subject_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.plural"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.indefinite_subject"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.indefinite_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.indefinite_subject_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.past"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_subject"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_subject_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", "en.transitive_negative.svo.past"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_subject"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive_negative", "en.transitive_negative.svo.past.indefinite_subject_object"},
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
	{"en.helping.event", "en.helping.transitive", "en.transitive", "en.transitive.svo"},
	{"en.helping.event", "en.helping.transitive", "en.transitive", "en.transitive.svo.bare_subject.object_the"},
	{"en.helping.event", "en.helping.transitive", "en.transitive", "en.transitive.svo.bare_subject.object_a"},
	{"en.thanking.event", "en.thanking.transitive", "en.transitive", "en.transitive.svo"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_target"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_target_an"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_subject"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_subject_an"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_subject_target"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_subject_a_target_an"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_subject_an_target_a"},
	{"en.writing_with_instrument.event", "en.writing_with_instrument.oblique", "en.with_oblique", "en.with_oblique.svo.indefinite_subject_an_target_an"},
	{"en.using_object.event", "en.using_object.transitive", "en.transitive", "en.transitive.svo"},
}

func init() {
	for _, declarations := range [][]constructionBindingDeclaration{
		englishNegativeArticleBindingDeclarations(), englishReadingPastBindingDeclarations(),
		englishHearingPastBindingDeclarations(), englishDrinkingPastBindingDeclarations(),
		englishEatingPastBindingDeclarations(), englishBarePluralBindingDeclarations(),
		englishManipulationBindingDeclarations(), englishAffectiveBindingDeclarations(),
		englishWritingMessageBindingDeclarations(),
		englishBelongingBindingDeclarations(),
		englishWritingMessagePastBindingDeclarations(),
		englishWritingMessageNegativeBindingDeclarations(), englishAffectivePastBindingDeclarations(),
		englishOpeningNegativeBindingDeclarations(),
		englishClosingPastBindingDeclarations(), englishClosingNegativeBindingDeclarations(),
		englishHoldingPastBindingDeclarations(), englishHoldingNegativeBindingDeclarations(),
		englishCarryingPastBindingDeclarations(), englishCarryingNegativeBindingDeclarations(),
		englishPushingPastNegativeBindingDeclarations(), englishPullingPastNegativeBindingDeclarations(), englishCuttingPastNegativeBindingDeclarations(),
		englishFoldingPastNegativeBindingDeclarations(),
		englishTearingPastNegativeBindingDeclarations(),
		englishSneezeCoughNegativeBindingDeclarations(),
	} {
		englishConstructionBindingDeclarations = append(englishConstructionBindingDeclarations, declarations...)
	}
}

func englishBelongingBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{"svo", "svo.indefinite_target", "svo.indefinite_target_an", "svo.indefinite_subject", "svo.indefinite_subject_an", "svo.indefinite_subject_target", "svo.this_item"}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.belong.association", frame: "en.belong.association.to_oblique",
			family: "en.to_oblique", variant: "en.to_oblique." + variant,
		})
	}
	return bindings
}

func englishSneezeCoughNegativeBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{
		"en.intransitive.sv.negative.singular", "en.intransitive.sv.negative.singular.indefinite_subject", "en.intransitive.sv.negative.singular.indefinite_subject_an",
		"en.intransitive.sv.negative.plural", "en.intransitive.sv.negative.plural.definite",
		"en.intransitive.sv.negative.past.singular", "en.intransitive.sv.negative.past.singular.indefinite_subject", "en.intransitive.sv.negative.past.singular.indefinite_subject_an",
		"en.intransitive.sv.negative.past.plural", "en.intransitive.sv.negative.past.plural.definite",
	}
	var bindings []constructionBindingDeclaration
	for _, route := range []struct{ sense, frame string }{
		{"en.sneezing.event", "en.sneezing.intransitive"},
		{"en.coughing.event", "en.coughing.intransitive"},
	} {
		for _, variant := range variants {
			bindings = append(bindings, constructionBindingDeclaration{sense: route.sense, frame: route.frame, family: "en.intransitive", variant: variant})
		}
	}
	return bindings
}

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

func englishAffectivePastBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, item := range []struct{ sense, frame string }{
		{"en.liking.event", "en.liking.transitive"},
		{"en.watching.event", "en.watching.transitive"},
		{"en.writing_document.event", "en.writing_document.transitive"},
	} {
		variants := []string{
			"svo", "svo.indefinite_object", "svo.indefinite_object_an", "svo.indefinite_subject", "svo.indefinite_subject_an",
			"svo.indefinite_subject_object", "svo.indefinite_subject_a_object_an", "svo.indefinite_subject_an_object_a", "svo.indefinite_subject_an_object_an",
			"svo.bare_object.subject_the", "svo.bare_object.subject_a", "svo.bare_object.subject_an",
			"svo.bare_subject.object_the", "svo.bare_subject.object_a", "svo.bare_subject.object_an", "svo.bare_subject.object_bare",
		}
		for _, variant := range variants {
			bindings = append(bindings, constructionBindingDeclaration{item.sense, item.frame, "en.transitive_past", "en.transitive_past." + variant})
		}
		for _, variant := range []string{
			"svo.past", "svo.past.indefinite_object", "svo.past.indefinite_subject", "svo.past.indefinite_subject_object",
			"svo.past.bare_object.subject_the", "svo.past.bare_object.subject_a", "svo.past.plural.bare_object.subject_the",
		} {
			bindings = append(bindings, constructionBindingDeclaration{item.sense, item.frame, "en.transitive_negative", "en.transitive_negative." + variant})
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

func englishWritingMessagePastBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{
		"svo", "svo.indefinite_object", "svo.indefinite_object_an", "svo.indefinite_subject", "svo.indefinite_subject_an",
		"svo.indefinite_subject_object", "svo.indefinite_subject_a_object_an", "svo.indefinite_subject_an_object_a", "svo.indefinite_subject_an_object_an",
	} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive_past", "en.transitive_past." + variant})
	}
	bindings = append(bindings, constructionBindingDeclaration{
		"en.writing_message.event", "en.writing_message.full", "en.give.transfer.past",
		"en.give.transfer.past.to_pp.compact_articles",
	})
	bindings = append(bindings, constructionBindingDeclaration{
		"en.writing_message.event", "en.writing_message.full", "en.give.transfer.past",
		"en.give.transfer.past.to_pp.np_local_subject_theme_recipient",
	})
	return bindings
}

func englishWritingMessageNegativeBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, suffix := range []string{"", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive_negative", "en.transitive_negative.svo" + suffix})
	}
	bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive_negative", "en.transitive_negative.svo.plural"})
	for _, variant := range []string{
		"en.transitive_negative.svo.bare_object.subject_the",
		"en.transitive_negative.svo.bare_object.subject_a",
		"en.transitive_negative.svo.bare_object.subject_an",
		"en.transitive_negative.svo.plural.bare_object.subject_the",
	} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive_negative", variant})
	}
	for _, suffix := range []string{"", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive_negative", "en.transitive_negative.svo.past" + suffix})
	}
	for _, variant := range []string{
		"en.transitive_negative.svo.past.bare_object.subject_the",
		"en.transitive_negative.svo.past.bare_object.subject_a",
		"en.transitive_negative.svo.past.bare_object.subject_an",
		"en.transitive_negative.svo.past.plural.bare_object.subject_the",
	} {
		bindings = append(bindings, constructionBindingDeclaration{"en.writing_message.event", "en.writing_message.message_only", "en.transitive_negative", variant})
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
			"en.transitive_negative.svo.past.plural.bare_subject.analyze",
			"en.transitive_negative.svo.bare_object.subject_the", "en.transitive_negative.svo.bare_object.subject_a", "en.transitive_negative.svo.bare_object.subject_an",
			"en.transitive_negative.svo.plural.bare_object.subject_the", "en.transitive_negative.svo.plural.bare_object.subject_bare",
			"en.transitive_negative.svo.plural.subject_bare.object_the", "en.transitive_negative.svo.plural.subject_bare.object_a", "en.transitive_negative.svo.plural.subject_bare.object_an",
		} {
			add(route.sense, route.frame, "en.transitive_negative", variant)
		}
	}
	// Reuse the common source-only past plural pattern for every approved
	// transitive frame with exactly the required subject/object slots.
	pastPluralBareSubject := "en.transitive_negative.svo.past.plural.bare_subject.analyze"
	seenPastPluralBareSubject := make(map[string]bool)
	for _, binding := range bindings {
		if binding.family == "en.transitive_negative" && binding.variant == pastPluralBareSubject {
			seenPastPluralBareSubject[binding.sense+"\x00"+binding.frame] = true
		}
	}
	for _, frame := range englishFrames {
		if frame.family != "en.transitive" || !hasRequiredBarePluralNegativeSubjectObjectSlots(frame.arguments) {
			continue
		}
		key := frame.sense + "\x00" + frame.id
		if seenPastPluralBareSubject[key] {
			continue
		}
		add(frame.sense, frame.id, "en.transitive_negative", pastPluralBareSubject)
		seenPastPluralBareSubject[key] = true
	}
	return bindings
}

func hasRequiredBarePluralNegativeSubjectObjectSlots(arguments []valency.ArgumentSlot) bool {
	if len(arguments) != 2 {
		return false
	}
	var subject, object bool
	for _, argument := range arguments {
		if !argument.Required {
			return false
		}
		switch argument.ID {
		case "subject":
			if argument.GrammaticalFunction != "subject" {
				return false
			}
			subject = true
		case "object":
			if argument.GrammaticalFunction != "object" {
				return false
			}
			object = true
		default:
			return false
		}
	}
	return subject && object
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

func englishHearingPastBindingDeclarations() []constructionBindingDeclaration {
	suffixes := []string{
		"", ".mass_object", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
		".indefinite_subject_mass_object", ".indefinite_object_an", ".indefinite_subject_an",
		".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
		".indefinite_subject_an_object_an", ".indefinite_subject_an_mass_object",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(suffixes))
	for _, suffix := range suffixes {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.hear.auditory", frame: "en.hear.auditory.transitive", family: "en.transitive_past",
			variant: "en.transitive_past.svo" + suffix,
		})
	}
	return bindings
}

func englishDrinkingPastBindingDeclarations() []constructionBindingDeclaration {
	suffixes := []string{
		"", ".mass_object", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
		".indefinite_subject_mass_object", ".indefinite_object_an", ".indefinite_subject_an",
		".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
		".indefinite_subject_an_object_an", ".indefinite_subject_an_mass_object",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(suffixes))
	for _, suffix := range suffixes {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.drinking.event", frame: "en.drinking.transitive", family: "en.transitive_past",
			variant: "en.transitive_past.svo" + suffix,
		})
	}
	return bindings
}

func englishEatingPastBindingDeclarations() []constructionBindingDeclaration {
	suffixes := []string{
		"", ".mass_object", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object",
		".indefinite_subject_mass_object", ".indefinite_object_an", ".indefinite_subject_an",
		".indefinite_subject_a_object_an", ".indefinite_subject_an_object_a",
		".indefinite_subject_an_object_an", ".indefinite_subject_an_mass_object",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(suffixes))
	for _, suffix := range suffixes {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.eating.event", frame: "en.eating.transitive", family: "en.transitive_past",
			variant: "en.transitive_past.svo" + suffix,
		})
	}
	return bindings
}

func englishGivingConstructionBindings() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	bindings = append(bindings, constructionBindingDeclaration{
		sense: "en.give.transfer", frame: "en.give.transfer.past", family: "en.give.transfer.past",
		variant: "en.give.transfer.past.double_np.np_local_subject_recipient_theme",
	})
	bindings = append(bindings, constructionBindingDeclaration{
		sense: "en.give.transfer", frame: "en.give.transfer.past", family: "en.give.transfer.past",
		variant: "en.give.transfer.past.double_np.compact_articles",
	})
	bindings = append(bindings, constructionBindingDeclaration{
		sense: "en.give.transfer", frame: "en.give.transfer.past", family: "en.give.transfer.past",
		variant: "en.give.transfer.past.to_pp.compact_articles",
	})
	bindings = append(bindings, constructionBindingDeclaration{
		sense: "en.give.transfer", frame: "en.give.transfer.past", family: "en.give.transfer.past",
		variant: "en.give.transfer.past.to_pp.np_local_subject_theme_recipient",
	})
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

func englishGivingNegativeConstructionBindings() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, order := range []string{"double_np", "to_pp"} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.give.transfer", frame: "en.give.transfer.negative", family: "en.ditransitive_negative",
			variant: "en.ditransitive_negative." + order + ".vp_auxiliary",
		})
		for _, tense := range []string{"present", "past"} {
			bindings = append(bindings, constructionBindingDeclaration{
				sense: "en.give.transfer", frame: "en.give.transfer.negative", family: "en.ditransitive_negative",
				variant: "en.ditransitive_negative.source." + tense + "." + order,
			})
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
		{"en.buying.event", "en.buying.event.omit_seller.negative", "en.transitive_negative", []string{
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
	for _, variant := range []string{"en.transitive_negative.svo", "en.transitive_negative.svo.plural"} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.buying.event", frame: "en.buying.event.omit_seller.negative", family: "en.transitive_negative",
			variant: variant,
		})
	}
	for _, suffix := range []string{"", ".indefinite_object", ".indefinite_subject", ".indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.buying.event", frame: "en.buying.event.omit_seller.negative", family: "en.transitive_negative",
			variant: "en.transitive_negative.svo.past" + suffix,
		})
	}
	for _, variant := range []string{
		"en.transitive_negative.svo.bare_object.subject_the",
		"en.transitive_negative.svo.bare_object.subject_a",
		"en.transitive_negative.svo.plural.bare_object.subject_the",
		"en.transitive_negative.svo.past.bare_object.subject_the",
		"en.transitive_negative.svo.past.bare_object.subject_a",
		"en.transitive_negative.svo.past.plural.bare_object.subject_the",
	} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.buying.event", frame: "en.buying.event.omit_seller.negative", family: "en.transitive_negative",
			variant: variant,
		})
	}
	return bindings
}

func englishOpeningNegativeBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{
		"en.transitive_negative.svo", "en.transitive_negative.svo.indefinite_object",
		"en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object",
		"en.transitive_negative.svo.bare_object.subject_the", "en.transitive_negative.svo.bare_object.subject_a",
		"en.transitive_negative.svo.opening_object.present.subject_the.object_the_plural",
		"en.transitive_negative.svo.opening_object.present.subject_a.object_the_plural",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.indefinite_object",
		"en.transitive_negative.svo.past.indefinite_subject", "en.transitive_negative.svo.past.indefinite_subject_object",
		"en.transitive_negative.svo.past.bare_object.subject_the", "en.transitive_negative.svo.past.bare_object.subject_a",
		"en.transitive_negative.svo.past.opening_object.past.subject_the.object_the_plural",
		"en.transitive_negative.svo.past.opening_object.past.subject_a.object_the_plural",
	} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.opening_object.event", frame: "en.opening_object.transitive.negative",
			family: "en.transitive_negative", variant: variant,
		})
	}
	return bindings
}

func englishClosingPastBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{"en.transitive_past.svo", "en.transitive_past.svo.indefinite_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_object", "en.transitive_past.svo.closing_object_past.the_plural", "en.transitive_past.svo.closing_object_past.indefinite_subject.the_plural"} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.closing_object.event", frame: "en.closing_object.transitive.past",
			family: "en.transitive_past", variant: variant,
		})
	}
	return bindings
}

func englishClosingNegativeBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{
		"en.transitive_negative.svo", "en.transitive_negative.svo.plural",
		"en.transitive_negative.svo.plural.bare_object.subject_the",
		"en.transitive_negative.svo.indefinite_object", "en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.indefinite_object",
		"en.transitive_negative.svo.past.indefinite_subject", "en.transitive_negative.svo.past.indefinite_subject_object",
		"en.transitive_negative.svo.closing_object.present.subject_the.object_the_plural", "en.transitive_negative.svo.closing_object.present.subject_a.object_the_plural",
		"en.transitive_negative.svo.closing_object.present.plural.subject_the.object_the_plural",
		"en.transitive_negative.svo.past.closing_object.subject_the.object_the_plural", "en.transitive_negative.svo.past.closing_object.subject_a.object_the_plural",
		"en.transitive_negative.svo.past.closing_object.plural.subject_the.object_the_plural",
	} {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.closing_object.event", frame: "en.closing_object.transitive.negative",
			family: "en.transitive_negative", variant: variant,
		})
	}
	return bindings
}

func englishHoldingPastBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{"en.transitive_past.svo", "en.transitive_past.svo.indefinite_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.holding_object.event", frame: "en.holding_object.transitive.past", family: "en.transitive_past", variant: variant})
	}
	return bindings
}

func englishHoldingNegativeBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{"en.transitive_negative.svo", "en.transitive_negative.svo.indefinite_object", "en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object", "en.transitive_negative.svo.past", "en.transitive_negative.svo.past.indefinite_object", "en.transitive_negative.svo.past.indefinite_subject", "en.transitive_negative.svo.past.indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.holding_object.event", frame: "en.holding_object.transitive.negative", family: "en.transitive_negative", variant: variant})
	}
	return bindings
}

func englishCarryingPastBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{"en.transitive_past.svo", "en.transitive_past.svo.indefinite_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.carrying.event", frame: "en.carrying.transitive.past", family: "en.transitive_past", variant: variant})
	}
	return bindings
}

func englishCarryingNegativeBindingDeclarations() []constructionBindingDeclaration {
	var bindings []constructionBindingDeclaration
	for _, variant := range []string{"en.transitive_negative.svo", "en.transitive_negative.svo.plural", "en.transitive_negative.svo.indefinite_object", "en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object", "en.transitive_negative.svo.past", "en.transitive_negative.svo.past.indefinite_object", "en.transitive_negative.svo.past.indefinite_subject", "en.transitive_negative.svo.past.indefinite_subject_object"} {
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.carrying.event", frame: "en.carrying.transitive.negative", family: "en.transitive_negative", variant: variant})
	}
	return bindings
}

func englishPushingPastNegativeBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{
		"en.transitive_past.svo", "en.transitive_past.svo.indefinite_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_object",
		"en.transitive_past.svo.bare_object.subject_the", "en.transitive_past.svo.bare_object.subject_a", "en.transitive_past.svo.bare_subject.object_the", "en.transitive_past.svo.bare_subject.object_a",
		"en.transitive_negative.svo", "en.transitive_negative.svo.plural", "en.transitive_negative.svo.indefinite_object", "en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object",
		"en.transitive_negative.svo.plural.indefinite_object", "en.transitive_negative.svo.plural.indefinite_subject", "en.transitive_negative.svo.plural.indefinite_subject_object",
		"en.transitive_negative.svo.bare_object.subject_the", "en.transitive_negative.svo.bare_object.subject_a", "en.transitive_negative.svo.plural.bare_object.subject_the",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.indefinite_object", "en.transitive_negative.svo.past.indefinite_subject", "en.transitive_negative.svo.past.indefinite_subject_object",
		"en.transitive_negative.svo.past.bare_object.subject_the", "en.transitive_negative.svo.past.bare_object.subject_a", "en.transitive_negative.svo.past.plural.bare_object.subject_the",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		family := "en.transitive_past"
		frame := "en.pushing_object.transitive.past"
		if strings.HasPrefix(variant, "en.transitive_negative.") {
			family = "en.transitive_negative"
			frame = "en.pushing_object.transitive.negative"
		}
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.pushing_object.event", frame: frame, family: family, variant: variant})
	}
	return bindings
}

func englishPullingPastNegativeBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{
		"en.transitive_past.svo", "en.transitive_past.svo.indefinite_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_object",
		"en.transitive_past.svo.bare_object.subject_the", "en.transitive_past.svo.bare_object.subject_a", "en.transitive_past.svo.bare_subject.object_the", "en.transitive_past.svo.bare_subject.object_a",
		"en.transitive_negative.svo", "en.transitive_negative.svo.plural", "en.transitive_negative.svo.indefinite_object", "en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object",
		"en.transitive_negative.svo.plural.indefinite_object", "en.transitive_negative.svo.plural.indefinite_subject", "en.transitive_negative.svo.plural.indefinite_subject_object",
		"en.transitive_negative.svo.bare_object.subject_the", "en.transitive_negative.svo.bare_object.subject_a", "en.transitive_negative.svo.plural.bare_object.subject_the",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.indefinite_object", "en.transitive_negative.svo.past.indefinite_subject", "en.transitive_negative.svo.past.indefinite_subject_object",
		"en.transitive_negative.svo.past.bare_object.subject_the", "en.transitive_negative.svo.past.bare_object.subject_a", "en.transitive_negative.svo.past.plural.bare_object.subject_the",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		family, frame := "en.transitive_past", "en.pulling_object.transitive.past"
		if strings.HasPrefix(variant, "en.transitive_negative.") {
			family, frame = "en.transitive_negative", "en.pulling_object.transitive.negative"
		}
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.pulling_object.event", frame: frame, family: family, variant: variant})
	}
	return bindings
}

func englishCuttingPastNegativeBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{
		"en.transitive_past.svo", "en.transitive_past.svo.mass_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_mass_object",
		"en.transitive_negative.svo", "en.transitive_negative.svo.plural", "en.transitive_negative.svo.mass_object", "en.transitive_negative.svo.mass_object.indefinite_subject",
		"en.transitive_negative.svo.plural.mass_object", "en.transitive_negative.svo.plural.mass_object.indefinite_subject",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.mass_object", "en.transitive_negative.svo.past.mass_object.indefinite_subject",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		family, frame := "en.transitive_past", "en.cutting_object.transitive.past"
		if strings.HasPrefix(variant, "en.transitive_negative.") {
			family, frame = "en.transitive_negative", "en.cutting_object.transitive.negative"
		}
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.cutting_object.event", frame: frame, family: family, variant: variant})
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
	{"en.own.legal_ownership.transitive", "en.own.legal_ownership", "en.transitive", []meaning.Role{"owner", "asset"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "owner", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "asset", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.have.possession.transitive", "en.have.possession", "en.transitive", []meaning.Role{"holder", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.hear.auditory.attributed_object", "en.hear.auditory", "en.attributed_seeing_object", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.hear.auditory.attributed_subject", "en.hear.auditory", "en.attributed_seeing", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.see.visual.attributed_negative", "en.see.visual", "en.attributed_seeing_object_negative", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.conversing.with", "en.conversing.event", "en.with_oblique", []meaning.Role{"participant_a", "participant_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "participant_a", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "participant_b", Required: true, GrammaticalFunction: "oblique", Adposition: "with"},
	}},
	{"en.using_object.transitive", "en.using_object.event", "en.transitive", []meaning.Role{"user", "tool"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "user", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "tool", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.helping.transitive", "en.helping.event", "en.transitive", []meaning.Role{"helper", "beneficiary"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "helper", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "beneficiary", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.thanking.transitive", "en.thanking.event", "en.transitive", []meaning.Role{"thanker", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "thanker", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "recipient", Required: true, GrammaticalFunction: "object"},
	}},
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
	{"en.belong.association.to_oblique", "en.belong.association", "en.to_oblique", []meaning.Role{"item", "associate"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "item", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "associate", Required: true, GrammaticalFunction: "oblique", Adposition: "to"},
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
	{"en.writing_with_instrument.oblique", "en.writing_with_instrument.event", "en.with_oblique", []meaning.Role{"writer", "instrument"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "writer", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "instrument", Required: true, GrammaticalFunction: "oblique", Adposition: "with"},
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
	{"en.opening_object.transitive.past", "en.opening_object.event", "en.transitive_past", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.closing_object.transitive.past", "en.closing_object.event", "en.transitive_past", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.opening_object.transitive.negative", "en.opening_object.event", "en.transitive_negative", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.closing_object.transitive.negative", "en.closing_object.event", "en.transitive_negative", []meaning.Role{"agent", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.holding_object.transitive.past", "en.holding_object.event", "en.transitive_past", []meaning.Role{"holder", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.holding_object.transitive.negative", "en.holding_object.event", "en.transitive_negative", []meaning.Role{"holder", "object"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "holder", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "object", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.carrying.transitive.past", "en.carrying.event", "en.transitive_past", []meaning.Role{"carrier", "load"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "carrier", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "load", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.carrying.transitive.negative", "en.carrying.event", "en.transitive_negative", []meaning.Role{"carrier", "load"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "carrier", Required: true, GrammaticalFunction: "subject"}, {ID: "object", Role: "load", Required: true, GrammaticalFunction: "object"},
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
	{"en.cutting_object.with_instrument", "en.cutting_object.with_instrument", "en.transitive_with_oblique", []meaning.Role{"agent", "material", "instrument"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
		{ID: "target", Role: "instrument", Required: true, GrammaticalFunction: "oblique", Adposition: "with"},
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
	{"en.give.transfer.past", "en.give.transfer", "en.give.transfer.past", []meaning.Role{"giver", "recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "giver", Required: true, GrammaticalFunction: "subject"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object"},
		{ID: "theme", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.give.transfer.negative", "en.give.transfer", "en.ditransitive_negative", []meaning.Role{"giver", "recipient", "theme"}, []valency.ArgumentSlot{
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

var englishPushingPastNegativeFrames = []frameDeclaration{
	{"en.pushing_object.transitive.past", "en.pushing_object.event", "en.transitive_past", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.pushing_object.transitive.negative", "en.pushing_object.event", "en.transitive_negative", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
}

var englishPullingPastNegativeFrames = []frameDeclaration{
	{"en.pulling_object.transitive.past", "en.pulling_object.event", "en.transitive_past", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.pulling_object.transitive.negative", "en.pulling_object.event", "en.transitive_negative", []meaning.Role{"agent", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
}

var englishCuttingPastNegativeFrames = []frameDeclaration{
	{"en.cutting_object.transitive.past", "en.cutting_object.event", "en.transitive_past", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.cutting_object.transitive.negative", "en.cutting_object.event", "en.transitive_negative", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
}

var englishFoldingPastNegativeFrames = []frameDeclaration{
	{"en.folding_object.transitive.past", "en.folding_object.event", "en.transitive_past", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.folding_object.transitive.negative", "en.folding_object.event", "en.transitive_negative", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
}

var englishTearingPastNegativeFrames = []frameDeclaration{
	{"en.tearing_object.transitive.past", "en.tearing_object.event", "en.transitive_past", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.tearing_object.transitive.negative", "en.tearing_object.event", "en.transitive_negative", []meaning.Role{"agent", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
}

func englishTearingPastNegativeBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{
		"en.transitive_past.svo", "en.transitive_past.svo.mass_object", "en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_mass_object",
		"en.transitive_negative.svo", "en.transitive_negative.svo.plural", "en.transitive_negative.svo.mass_object", "en.transitive_negative.svo.mass_object.indefinite_subject",
		"en.transitive_negative.svo.plural.mass_object", "en.transitive_negative.svo.plural.mass_object.indefinite_subject",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.mass_object", "en.transitive_negative.svo.past.mass_object.indefinite_subject",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		family, frame := "en.transitive_past", "en.tearing_object.transitive.past"
		if strings.HasPrefix(variant, "en.transitive_negative.") {
			family, frame = "en.transitive_negative", "en.tearing_object.transitive.negative"
		}
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.tearing_object.event", frame: frame, family: family, variant: variant})
	}
	return bindings
}

func englishFoldingPastNegativeBindingDeclarations() []constructionBindingDeclaration {
	variants := []string{
		"en.transitive_past.svo", "en.transitive_past.svo.mass_object", "en.transitive_past.svo.indefinite_object",
		"en.transitive_past.svo.indefinite_subject", "en.transitive_past.svo.indefinite_subject_object",
		"en.transitive_past.svo.indefinite_subject_mass_object", "en.transitive_past.svo.indefinite_subject_an_mass_object",
		"en.transitive_past.svo.indefinite_object_an", "en.transitive_past.svo.indefinite_subject_an",
		"en.transitive_past.svo.indefinite_subject_a_object_an", "en.transitive_past.svo.indefinite_subject_an_object_a", "en.transitive_past.svo.indefinite_subject_an_object_an",
		"en.transitive_negative.svo", "en.transitive_negative.svo.plural", "en.transitive_negative.svo.mass_object",
		"en.transitive_negative.svo.mass_object.indefinite_subject", "en.transitive_negative.svo.plural.mass_object",
		"en.transitive_negative.svo.plural.mass_object.indefinite_subject", "en.transitive_negative.svo.indefinite_object",
		"en.transitive_negative.svo.indefinite_subject", "en.transitive_negative.svo.indefinite_subject_object",
		"en.transitive_negative.svo.past", "en.transitive_negative.svo.past.mass_object",
		"en.transitive_negative.svo.past.mass_object.indefinite_subject",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		family, frame := "en.transitive_past", "en.folding_object.transitive.past"
		if strings.HasPrefix(variant, "en.transitive_negative.") {
			family, frame = "en.transitive_negative", "en.folding_object.transitive.negative"
		}
		bindings = append(bindings, constructionBindingDeclaration{sense: "en.folding_object.event", frame: frame, family: family, variant: variant})
	}
	return bindings
}

var englishCommercePastOmissionFrames = []frameDeclaration{
	{"en.buying.event.omit_seller.past", "en.buying.event", "en.transitive_past", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique"},
	}},
	{"en.selling.event.omit_buyer.past", "en.selling.event", "en.transitive_past", []meaning.Role{"seller", "goods", "buyer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "seller", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object"},
		{ID: "counterparty", Role: "buyer", Required: false, GrammaticalFunction: "oblique"},
	}},
}

var englishCommerceNegativeFrames = []frameDeclaration{
	{"en.buying.event.omit_seller.negative", "en.buying.event", "en.transitive_negative", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object"},
		{ID: "counterparty", Role: "seller", Required: false, GrammaticalFunction: "oblique"},
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
	"en.buying.event.omit_seller":          {"seller"},
	"en.buying.event.omit_seller.past":     {"seller"},
	"en.buying.event.omit_seller.negative": {"seller"},
	"en.selling.event.omit_buyer":          {"buyer"},
	"en.selling.event.omit_buyer.past":     {"buyer"},
	"en.writing_message.message_only":      {"recipient"},
	"en.writing_message.recipient_only":    {"message"},
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
	for _, variant := range []string{"svo", "svo.mass_object", "svo.indefinite_object", "svo.indefinite_subject", "svo.indefinite_subject_object", "svo.indefinite_subject_mass_object", "svo.indefinite_object_an", "svo.indefinite_subject_an", "svo.indefinite_subject_a_object_an", "svo.indefinite_subject_an_object_a", "svo.indefinite_subject_an_object_an", "svo.indefinite_subject_an_mass_object"} {
		bindings = append(bindings, constructionBindingDeclaration{"en.buying.event", "en.buying.event.omit_seller.past", "en.transitive_past", "en.transitive_past." + variant})
		bindings = append(bindings, constructionBindingDeclaration{"en.selling.event", "en.selling.event.omit_buyer.past", "en.transitive_past", "en.transitive_past." + variant})
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
	declarations := append(append([]frameDeclaration(nil), englishFrames...), englishAttributionFrames()...)
	declarations = append(declarations, englishCommerceOmissionFrames...)
	declarations = append(declarations, englishCommercePastOmissionFrames...)
	declarations = append(declarations, englishCommerceNegativeFrames...)
	declarations = append(declarations, englishPushingPastNegativeFrames...)
	declarations = append(declarations, englishPullingPastNegativeFrames...)
	declarations = append(declarations, englishCuttingPastNegativeFrames...)
	declarations = append(declarations, englishFoldingPastNegativeFrames...)
	declarations = append(declarations, englishTearingPastNegativeFrames...)
	declarations = append(declarations, englishCommerceFullFrames...)
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
		if item.sense == "en.open.state" {
			person, err := ids.NewConceptID("person")
			if err != nil {
				return nil, fmt.Errorf("English physical-open person exclusion: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "physical_open_excludes_person", Role: "bearer", Strength: valency.ConstraintHard,
				ExcludedConceptIDs: []ids.ConceptID{person},
			}}
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
		if item.id == "en.writing_with_instrument.oblique" {
			pencil, err := ids.NewConceptID("pencil")
			if err != nil {
				return nil, fmt.Errorf("English writing-instrument selection concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{
				ID: "writing_instrument_pencil", Role: "instrument", Strength: valency.ConstraintHard,
				AllowedConceptIDs: []ids.ConceptID{pencil},
			}}
		}
		if item.id == "en.living.intransitive" {
			meat, err := ids.NewConceptID("meat_food")
			if err != nil {
				return nil, fmt.Errorf("English living exclusion concept: %w", err)
			}
			frame.SelectionalConstraints = []valency.SelectionalConstraint{{ID: "living_organism_excludes_meat", Role: "organism", Strength: valency.ConstraintHard, ExcludedConceptIDs: []ids.ConceptID{meat}}}
		}
		if item.id == "en.be.class_membership.frame" || func() bool {
			_, attributed := englishAttributionForSense("sense:" + strings.TrimSuffix(item.id, ".frame"))
			return attributed
		}() {
			frame.Provenance = englishIdentificationSource
		}
		switch item.id {
		case "en.have.possession.transitive":
			for _, family := range []string{"en.transitive_personal", "en.transitive_negative"} {
				familyID, err := ids.NewConstructionID(family)
				if err != nil {
					return nil, fmt.Errorf("English have frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, familyID)
			}
		case "en.writing_message.message_only":
			pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
			if err != nil {
				return nil, fmt.Errorf("English past writing-message frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative writing-message frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "en.writing_document.transitive":
			for _, family := range []string{"en.transitive_past", "en.transitive_negative"} {
				familyID, err := ids.NewConstructionID(family)
				if err != nil {
					return nil, fmt.Errorf("English writing-document frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, familyID)
			}
		case "en.writing_message.full":
			pastFamilyID, err := ids.NewConstructionID("en.give.transfer.past")
			if err != nil {
				return nil, fmt.Errorf("English past full writing-message frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
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
			if item.id == "en.hear.auditory.transitive" {
				familyID, err := ids.NewConstructionID("en.attributed_seeing_object")
				if err != nil {
					return nil, fmt.Errorf("English hearing object-modifier construction family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, familyID)
				pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
				if err != nil {
					return nil, fmt.Errorf("English past hearing frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
			}
			if item.id == "en.see.visual.transitive" {
				pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
				if err != nil {
					return nil, fmt.Errorf("English past seeing frame family: %w", err)
				}
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
				for _, family := range []string{"en.attributed_seeing", "en.attributed_seeing_object", "en.attributed_seeing_both"} {
					attributedFamilyID, err := ids.NewConstructionID(family)
					if err != nil {
						return nil, fmt.Errorf("English attributed seeing frame family: %w", err)
					}
					frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, attributedFamilyID)
				}
			}
		case "en.drinking.transitive", "en.eating.transitive":
			pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
			if err != nil {
				return nil, fmt.Errorf("English past drinking frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative drinking frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "en.liking.transitive":
			pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
			if err != nil {
				return nil, fmt.Errorf("English past liking frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative liking frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "en.watching.transitive":
			pastFamilyID, err := ids.NewConstructionID("en.transitive_past")
			if err != nil {
				return nil, fmt.Errorf("English past watching frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, pastFamilyID)
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative affective frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		case "en.read.text.transitive":
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English negative reading frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
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
		if item.family == "en.transitive" && hasRequiredBarePluralNegativeSubjectObjectSlots(item.arguments) {
			negativeFamilyID, err := ids.NewConstructionID("en.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("English transitive negative family: %w", err)
			}
			alreadyAllowed := false
			for _, allowedFamilyID := range frame.AllowedFamilyIDs {
				if allowedFamilyID == negativeFamilyID {
					alreadyAllowed = true
					break
				}
			}
			if !alreadyAllowed {
				frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
			}
		}
		frames = append(frames, frame)
	}
	bindingDeclarations := append([]constructionBindingDeclaration(nil), englishConstructionBindingDeclarations...)
	bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
		sense: "en.own.legal_ownership", frame: "en.own.legal_ownership.transitive", family: "en.transitive", variant: "en.transitive.svo",
	})
	for _, item := range englishConstructionBindingDeclarations {
		if item.sense != "en.see.visual" || item.frame != "en.see.visual.transitive" || item.family != "en.attributed_seeing_object" {
			continue
		}
		bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
			sense: "en.hear.auditory", frame: "en.hear.auditory.transitive", family: item.family, variant: item.variant,
		})
	}
	for _, entry := range []struct{ article, numberSuffix string }{
		{"this", ""}, {"that", ""},
		{"these", ".plural_subject"}, {"those", ".plural_subject"},
		{"these", ".plural_both"}, {"those", ".plural_both"},
	} {
		bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
			sense: "en.see.visual", frame: "en.see.visual.transitive", family: "en.attributed_seeing_both",
			variant: "en.attributed_seeing_both.svo.demonstrative_" + entry.article + entry.numberSuffix,
		})
	}
	bindingDeclarations = append(bindingDeclarations, englishAttributionConstructionBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishNegativeAttributionBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishAttributedHearingBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishAttributedHearingObjectPluralBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishGivingConstructionBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishGivingNegativeConstructionBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishCommerceOmissionBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishUsingTargetBindings()...)
	bindingDeclarations = append(bindingDeclarations, englishConversingBindings()...)
	seenNegativeRoutes := make(map[string]bool)
	for _, declaration := range bindingDeclarations {
		if declaration.family != "en.transitive_negative" {
			continue
		}
		key := declaration.sense + "\x00" + declaration.frame
		if seenNegativeRoutes[key] {
			continue
		}
		seenNegativeRoutes[key] = true
		if declaration.variant == "en.transitive_negative.svo.vp_auxiliary" {
			continue
		}
		bindingDeclarations = append(bindingDeclarations, constructionBindingDeclaration{
			sense: declaration.sense, frame: declaration.frame,
			family: "en.transitive_negative", variant: "en.transitive_negative.svo.vp_auxiliary",
		})
	}
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

func englishConversingBindings() []constructionBindingDeclaration {
	variants := []string{
		"svo", "svo.indefinite_target", "svo.indefinite_target_an", "svo.indefinite_subject", "svo.indefinite_subject_an",
		"svo.indefinite_subject_target", "svo.indefinite_subject_a_target_an", "svo.indefinite_subject_an_target_a", "svo.indefinite_subject_an_target_an",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.conversing.event", frame: "en.conversing.with", family: "en.with_oblique", variant: "en.with_oblique." + variant,
		})
	}
	return bindings
}

func englishUsingTargetBindings() []constructionBindingDeclaration {
	variants := []string{
		"en.transitive.svo.indefinite_object",
		"en.transitive.svo.indefinite_subject",
		"en.transitive.svo.indefinite_subject_object",
	}
	bindings := make([]constructionBindingDeclaration, 0, len(variants))
	for _, variant := range variants {
		bindings = append(bindings, constructionBindingDeclaration{
			sense: "en.using_object.event", frame: "en.using_object.transitive", family: "en.transitive", variant: variant,
		})
	}
	return bindings
}

func englishConstructionBindingProvenance(item constructionBindingDeclaration) provenance.SourceRef {
	if item.sense == "en.belong.association" {
		return englishBelongingLexicalSource
	}
	if item.family == "en.attribution_negative" {
		return englishNegativeAttributionConstructionSource
	}
	if item.family == "en.attribution" && strings.HasSuffix(item.variant, ".past") {
		return englishPastAttributionConstructionSource
	}
	if item.sense == "en.conversing.event" {
		return englishConversingConstructionSource
	}
	if item.sense == "en.helping.event" || item.sense == "en.thanking.event" || item.sense == "en.using_object.event" || item.sense == "en.cutting_object.with_instrument" {
		return englishIssue436ConstructionSource
	}
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
	if item.family == "en.attributed_seeing" || item.family == "en.attributed_seeing_object" || item.family == "en.attributed_seeing_both" {
		if item.family == "en.attributed_seeing" && strings.Contains(item.variant, ".demonstrative_") {
			return englishAttributedSeeingDemonstrativesSource
		}
		return englishAttributedSeeingSource
	}
	if item.family == "en.identification_negative" {
		return englishIdentificationNegationSource
	}
	if item.sense == "en.be.class_membership" {
		return englishIdentificationSource
	}
	if _, attributed := englishAttributionForSense("sense:" + item.sense); attributed {
		return englishAdjectiveDataSource
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
