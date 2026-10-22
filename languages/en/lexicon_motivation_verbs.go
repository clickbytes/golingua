package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishMotivationVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_motivation_verbs.go",
	Version: "2026-10-22-batch-27", License: "unknown", Author: "golingua contributors",
}

// These pending candidates distinguish preferences, intentions and activity phases.
var englishMotivationVerbEntries = []lexicalEntry{
	{"en.desiring.v", "desire", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.desiring.event", "experience a directed want for an outcome, object, or state", "desiring"}}},
	{"en.liking.v", "like", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.liking.event", "experience positive preference or enjoyment toward an item or activity", "liking"}}},
	{"en.disliking.v", "dislike", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.disliking.event", "experience negative preference toward an item or activity", "disliking"}}},
	{"en.needing.v", "need", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.needing.event", "be in a condition where a resource or action is required for a goal", "needing"}}},
	{"en.preferring.v", "prefer", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.preferring.event", "favor one alternative over at least one competing option", "preferring"}}},
	{"en.loving_person.v", "love", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.loving_person.event", "experience enduring strong affection or attachment toward another person", "loving_person"}}},
	{"en.fearing.v", "fear", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.fearing.event", "experience fear in response to an anticipated or perceived threat", "fearing"}}},
	{"en.worrying.v", "worry", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.worrying.event", "repeatedly anticipate an unfavorable possibility with emotional concern", "worrying"}}},
	{"en.enjoying.v", "enjoy", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.enjoying.event", "experience pleasure during or from an activity or situation", "enjoying"}}},
	{"en.helping.v", "help", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.helping.event", "act to make another participant’s task easier or outcome more achievable", "helping"}}},
	{"en.cooperating.v", "cooperate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.cooperating.event", "coordinate action with another participant toward a shared objective", "cooperating"}}},
	{"en.trying.v", "try", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.trying.event", "initiate or sustain an attempt to bring about an outcome without entailing success", "trying"}}},
	{"en.succeeding.v", "succeed", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.succeeding.event", "achieve the specified objective of an attempted action", "succeeding"}}},
	{"en.failing.v", "fail", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.failing.event", "end an attempted action without achieving its specified objective", "failing"}}},
	{"en.starting_activity.v", "start", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.starting_activity.event", "cause an activity to pass from not occurring into an occurring state", "starting_activity"}}},
	{"en.finishing_activity.v", "finish", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.finishing_activity.event", "bring an activity to its intended or defined terminal boundary", "finishing_activity"}}},
	{"en.continuing_activity.v", "continue", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.continuing_activity.event", "keep an activity in progress across a later time segment", "continuing_activity"}}},
	{"en.changing_state.v", "change", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.changing_state.event", "cause or undergo a condition different from a preceding condition", "changing_state"}}},
	{"en.increasing_quantity.v", "increase", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.increasing_quantity.event", "make a measurable magnitude larger between observations", "increasing_quantity"}}},
	{"en.reducing_quantity.v", "reduce", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.reducing_quantity.event", "make a measurable magnitude smaller between observations", "reducing_quantity"}}},
	{"en.laughing.v", "laugh", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.laughing.event", "produce characteristic voiced or breathy sounds in amusement or other strong feeling", "laughing"}}},
	{"en.crying_emotionally.v", "cry", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.crying_emotionally.event", "express distress or emotion with tears or associated vocalizations", "crying_emotionally"}}},
	{"en.smiling.v", "smile", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.smiling.event", "raise the corners of the mouth in a facial expression often linked to positive affect", "smiling"}}},
	{"en.living.v", "live", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.living.event", "maintain integrated biological processes over a span of time", "living"}}},
	{"en.dancing.v", "dance", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.dancing.event", "perform patterned expressive body movements, often coordinated with music", "dancing"}}},
}
