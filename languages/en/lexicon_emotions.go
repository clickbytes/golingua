package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishEmotionSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_emotions.go",
	Version: "2026-10-22-batch-24", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name affective conditions, separate from bodily needs.
var englishEmotionEntries = []lexicalEntry{
	{"en.emotion_state.n", "emotion", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.emotion_state.direct", "a person’s temporally situated affective condition, distinguished by felt quality and appraisal", "emotion_state"}}},
	{"en.joy.n", "joy", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.joy.direct", "a positive affective condition characterized by delight or pleasure", "joy"}}},
	{"en.sadness.n", "sadness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sadness.direct", "an affective response characterized by sorrow or perceived loss", "sadness"}}},
	{"en.fear_emotion.n", "fear", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fear_emotion.direct", "an affective response to a perceived or anticipated danger", "fear_emotion"}}},
	{"en.anger.n", "anger", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.anger.direct", "an affective response to frustration, injustice, or perceived obstruction", "anger"}}},
	{"en.love_affection.n", "love", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.love_affection.direct", "a sustained affective attachment involving care or strong positive regard", "love_affection"}}},
	{"en.surprise_emotion.n", "surprise", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.surprise_emotion.direct", "an affective response to an occurrence that violates an expectation", "surprise_emotion"}}},
	{"en.disgust_emotion.n", "disgust", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.disgust_emotion.direct", "an aversive affective response to something perceived as repellent", "disgust_emotion"}}},
	{"en.anxiety.n", "anxiety", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.anxiety.direct", "persistent apprehension concerning uncertain future threats or outcomes", "anxiety"}}},
	{"en.boredom.n", "boredom", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.boredom.direct", "an unpleasant state of low interest or stimulation during an activity", "boredom"}}},
	{"en.curiosity.n", "curiosity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.curiosity.direct", "an affective motivation to explore or acquire unfamiliar information", "curiosity"}}},
}
