package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPerceptionVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_perception_verbs.go",
	Version: "2026-10-22-batch-29", License: "unknown", Author: "golingua contributors",
}

// These pending verbs distinguish sensory attention, cognition and decision acts.
var englishPerceptionVerbEntries = []lexicalEntry{
	{"en.listening.v", "listen", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.listening.event", "direct auditory attention toward sounds, whether or not their information is understood", "listening"}}},
	{"en.smelling.v", "smell", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.smelling.event", "perceive a volatile substance or source through the sense of smell", "smelling"}}},
	{"en.tasting.v", "taste", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.tasting.event", "perceive a substance’s flavor by contact with taste receptors", "tasting"}}},
	{"en.tactile_perception.v", "feel", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.tactile_perception.event", "obtain sensory information by bodily contact or pressure", "tactile_perception"}}},
	{"en.noticing.v", "notice", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.noticing.event", "become consciously aware of a feature or occurrence not previously attended to", "noticing"}}},
	{"en.recognizing.v", "recognize", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.recognizing.event", "identify a perceived entity as familiar or belonging to a known category", "recognizing"}}},
	{"en.observing.v", "observe", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.observing.event", "attend to an occurrence or object to collect information about it", "observing"}}},
	{"en.watching_screen.v", "watch", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.watching_screen.event", "direct sustained visual attention toward content on a display", "watching_screen"}}},
	{"en.monitoring.v", "monitor", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.monitoring.event", "repeatedly check an ongoing condition for changes requiring attention", "monitoring"}}},
	{"en.thinking.v", "think", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.thinking.event", "engage in conscious mental processing without a specified conclusion", "thinking"}}},
	{"en.remembering.v", "remember", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.remembering.event", "retrieve or maintain a mental representation of a past item or occurrence", "remembering"}}},
	{"en.forgetting.v", "forget", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.forgetting.event", "lose access to previously available knowledge or recollection", "forgetting"}}},
	{"en.imagining.v", "imagine", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.imagining.event", "construct a mental representation not presently perceived as an external fact", "imagining"}}},
	{"en.dreaming_during_sleep.v", "dream", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.dreaming_during_sleep.event", "experience internally generated imagery or narratives while asleep", "dreaming_during_sleep"}}},
	{"en.acquiring_knowledge.v", "learn", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.acquiring_knowledge.event", "come to know a fact or relationship through evidence or instruction", "acquiring_knowledge"}}},
	{"en.forming_belief.v", "believe", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.forming_belief.event", "adopt a proposition as likely or true without establishing its actual truth", "forming_belief"}}},
	{"en.suspecting.v", "suspect", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.suspecting.event", "form a tentative belief that a possibility may be the case", "suspecting"}}},
	{"en.understanding.v", "understand", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.understanding.event", "arrive at a coherent interpretation of information or how something works", "understanding"}}},
	{"en.misunderstanding.v", "misunderstand", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.misunderstanding.event", "construct an interpretation that differs from the intended or actual meaning", "misunderstanding"}}},
	{"en.reasoning.v", "reason", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.reasoning.event", "derive or evaluate conclusions from premises or evidence", "reasoning"}}},
	{"en.comparing.v", "compare", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.comparing.event", "examine at least two items for similarities or differences", "comparing"}}},
	{"en.categorizing.v", "categorize", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.categorizing.event", "assign an item to a class based on chosen distinguishing properties", "categorizing"}}},
	{"en.estimating.v", "estimate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.estimating.event", "produce an approximate value or judgment from incomplete information", "estimating"}}},
	{"en.calculating.v", "calculate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.calculating.event", "apply numerical operations to derive a quantitative result", "calculating"}}},
	{"en.deciding.v", "decide", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.deciding.event", "commit to a course or conclusion after considering alternatives", "deciding"}}},
	{"en.choosing.v", "choose", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.choosing.event", "select one option from available alternatives", "choosing"}}},
	{"en.planning.v", "plan", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.planning.event", "form a prospective sequence of actions intended to achieve a goal", "planning"}}},
	{"en.anticipating.v", "anticipate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.anticipating.event", "form expectations or preparation for a possible future occurrence", "anticipating"}}},
	{"en.expecting.v", "expect", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.expecting.event", "regard an occurrence as likely to happen or an outcome as due", "expecting"}}},
	{"en.hoping.v", "hope", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.hoping.event", "favor a possible future state while recognizing uncertainty", "hoping"}}},
}
