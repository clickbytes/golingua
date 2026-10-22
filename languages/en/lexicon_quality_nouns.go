package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishQualityNounSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_quality_nouns.go",
	Version: "2026-10-22-batch-30", License: "unknown", Author: "golingua contributors",
}

// These pending nouns name biological, social and practical qualities.
var englishQualityNounEntries = []lexicalEntry{
	{"en.age.n", "age", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.age.direct", "elapsed time since a person’s, organism’s, or object’s origin", "age"}}},
	{"en.health_condition.n", "health", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.health_condition.direct", "the overall state of bodily functioning and freedom from impairment", "health_condition"}}},
	{"en.physical_strength.n", "strength", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.physical_strength.direct", "capacity to exert force through bodily muscles", "physical_strength"}}},
	{"en.stamina.n", "stamina", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stamina.direct", "capacity to sustain exertion without excessive fatigue", "stamina"}}},
	{"en.coordination.n", "coordination", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.coordination.direct", "capacity to organize body movements accurately in time and space", "coordination"}}},
	{"en.skill_level.n", "skill", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.skill_level.direct", "degree of learned proficiency at a specified activity", "skill_level"}}},
	{"en.knowledge_level.n", "knowledge", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.knowledge_level.direct", "extent of information and understanding held about a domain", "knowledge_level"}}},
	{"en.attention_level.n", "attention", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.attention_level.direct", "degree of mental focus allocated to a task or stimulus", "attention_level"}}},
	{"en.confidence_level.n", "confidence", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.confidence_level.direct", "strength of a person’s expectation that they can achieve or judge correctly", "confidence_level"}}},
	{"en.emotional_valence.n", "valence", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.emotional_valence.direct", "positive or negative affective tone of a person’s current emotional state", "emotional_valence"}}},
	{"en.satisfaction_level.n", "satisfaction", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.satisfaction_level.direct", "degree to which an experience meets a participant’s wants or expectations", "satisfaction_level"}}},
	{"en.importance.n", "importance", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.importance.direct", "relative significance of an item to a purpose or decision", "importance"}}},
	{"en.usefulness.n", "usefulness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.usefulness.direct", "degree to which something serves a stated practical purpose", "usefulness"}}},
	{"en.difficulty.n", "difficulty", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.difficulty.direct", "degree of effort or capability required to perform a task", "difficulty"}}},
	{"en.accessibility.n", "accessibility", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.accessibility.direct", "degree to which a resource or place can be reached or used by relevant people", "accessibility"}}},
	{"en.safety_level.n", "safety", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.safety_level.direct", "degree to which a situation avoids or controls risk of harm", "safety_level"}}},
	{"en.availability.n", "availability", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.availability.direct", "extent to which a resource is obtainable at a specified time or context", "availability"}}},
	{"en.reliability.n", "reliability", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.reliability.direct", "likelihood that a system or source performs consistently as expected", "reliability"}}},
	{"en.accuracy.n", "accuracy", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.accuracy.direct", "degree of agreement between a representation or measurement and its reference", "accuracy"}}},
	{"en.completeness.n", "completeness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.completeness.direct", "extent to which required components or information are present", "completeness"}}},
}
