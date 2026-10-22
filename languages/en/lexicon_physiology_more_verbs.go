package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPhysiologyMoreVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_physiology_more_verbs.go",
	Version: "2026-10-22-batch-28", License: "unknown", Author: "golingua contributors",
}

// These pending verbs name further bodily and health processes.
var englishPhysiologyMoreVerbEntries = []lexicalEntry{
	{"en.drinking.v", "drink", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.drinking.event", "take liquid into the mouth and swallow it", "drinking"}}},
	{"en.chewing.v", "chew", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.chewing.event", "mechanically crush food with teeth before swallowing", "chewing"}}},
	{"en.digesting.v", "digest", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.digesting.event", "break down consumed material and absorb nutrients through bodily processes", "digesting"}}},
	{"en.sweating.v", "sweat", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.sweating.event", "release fluid through sweat glands in the skin", "sweating"}}},
	{"en.bleeding.v", "bleed", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.bleeding.event", "lose blood through a damaged blood vessel or body surface", "bleeding"}}},
	{"en.healing.v", "heal", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.healing.event", "undergo biological recovery of damaged tissue toward its ordinary function", "healing"}}},
	{"en.becoming_ill.v", "become ill", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.becoming_ill.event", "transition from an ordinary healthy condition to a state of disease", "becoming_ill"}}},
	{"en.falling_asleep.v", "fall asleep", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.falling_asleep.event", "transition from wakefulness into physiological sleep", "falling_asleep"}}},
	{"en.waking_up.v", "wake up", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.waking_up.event", "transition from physiological sleep into wakefulness", "waking_up"}}},
	{"en.exercising_body.v", "exercise", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.exercising_body.event", "perform physical activity for fitness, training, or health", "exercising_body"}}},
	{"en.aging.v", "age", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.aging.event", "undergo biological changes associated with increasing age over time", "aging"}}},
	{"en.dying.v", "die", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.dying.event", "reach the irreversible end of an organism’s integrated life processes", "dying"}}},
	{"en.giving_birth.v", "give birth", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.giving_birth.event", "deliver offspring from a pregnant organism’s body", "giving_birth"}}},
	{"en.injuring_person.v", "injure", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.injuring_person.event", "cause bodily tissue damage to a living participant", "injuring_person"}}},
	{"en.treating_patient.v", "treat", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.treating_patient.event", "provide a therapeutic intervention intended to improve a patient’s condition", "treating_patient"}}},
	{"en.diagnosing_condition.v", "diagnose", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.diagnosing_condition.event", "determine the likely identity or nature of a medical condition from evidence", "diagnosing_condition"}}},
	{"en.taking_medicine.v", "take medicine", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.taking_medicine.event", "administer a medicinal substance to oneself or receive its administration", "taking_medicine"}}},
	{"en.recovering_from_illness.v", "recover", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.recovering_from_illness.event", "return toward an ordinary state of health following illness", "recovering_from_illness"}}},
}
