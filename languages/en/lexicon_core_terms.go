package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishCoreTermsSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_core_terms.go",
	Version: "2026-10-22-batch-13", License: "unknown", Author: "golingua contributors",
}

// These broad English nouns are pending lexical candidates for CORE meanings.
// Event nouns do not imply a matching verb, frame, or sentence construction.
var englishCoreTermEntries = []lexicalEntry{
	{"en.entity.n", "entity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.entity.continuant", "a continuant that can be identified independently of a particular event", "entity"}}},
	{"en.student.n", "student", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.student.learner", "a person acting in a role of learning from instruction or study", "student"}}},
	{"en.patient.n", "patient", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.patient.care_recipient", "a person receiving or seeking medical care", "patient"}}},
	{"en.food.n", "food", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.food.nourishment", "a substance or prepared item intended for nourishment", "food"}}},
	{"en.perception.n", "perception", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.perception.sensory_event", "an event in which a perceiver acquires sensory information about a stimulus", "perception"}}},
	{"en.movement.n", "movement", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.movement.spatial_change", "a change in an entity's spatial position over time", "movement"}}},
	{"en.arrival.n", "arrival", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.arrival.destination_event", "movement ending at a destination", "arrival"}}},
	{"en.departure.n", "departure", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.departure.source_event", "movement beginning at a source location", "departure"}}},
	{"en.transfer.n", "transfer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.transfer.possession_event", "a change in control or possession of a theme between participants", "transfer"}}},
	{"en.communication.n", "communication", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.communication.message_event", "an event in which a sender conveys communicative content to an addressee", "communication"}}},
	{"en.question.n", "question", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.question.information_request", "a communicative act requesting information from an addressee", "asking"}}},
	{"en.answer.n", "answer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.answer.response", "a communicative act supplying information in response to a prior request", "answering"}}},
	{"en.teaching.n", "teaching", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.teaching.instruction_event", "an event in which an instructor intentionally supports another's learning", "teaching"}}},
	{"en.eating.n", "eating", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.eating.ingestion_event", "an event of taking food into a body for nourishment", "eating"}}},
	{"en.attribute.n", "attribute", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.attribute.quality", "a quality that can be attributed to a bearer without identifying a particular bearer", "property"}}},
	{"en.size.n", "size", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.size.extent", "the extent of a bearer along one or more spatial dimensions", "size"}}},
	{"en.temperature.n", "temperature", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.temperature.thermal_state", "the thermal state of a bearer on a temperature scale", "temperature"}}},
	{"en.color.n", "color", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.color.visual_quality", "a visual quality of a bearer associated with perceived light", "color"}}},
	{"en.speed.n", "speed", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.speed.movement_rate", "the rate of a participant's movement relative to time", "speed"}}},
}
