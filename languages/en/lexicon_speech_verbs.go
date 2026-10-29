package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishSpeechVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_speech_verbs.go",
	Version: "2026-10-22-batch-25", License: "unknown", Author: "golingua contributors",
}

// These pending verb candidates distinguish communication and interaction acts.
var englishSpeechVerbEntries = []lexicalEntry{
	{"en.speaking.v", "speak", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.speaking.event", "produce articulated spoken language directed or available to listeners", "speaking"}}},
	{"en.stating.v", "state", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.stating.event", "express a proposition as an assertion of fact or opinion", "stating"}}},
	{"en.narrating.v", "narrate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.narrating.event", "relate a sequence of events as an organized account", "narrating"}}},
	{"en.announcing.v", "announce", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.announcing.event", "make information publicly or formally known to an audience", "announcing"}}},
	{"en.informing.v", "inform", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.informing.event", "convey information that the sender expects the recipient to learn", "informing"}}},
	{"en.explaining.v", "explain", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.explaining.event", "make a subject intelligible by presenting reasons, structure, or instructions", "explaining"}}},
	{"en.describing.v", "describe", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.describing.event", "communicate distinguishing features or conditions of a referent", "describing"}}},
	{"en.naming.v", "name", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.naming.event", "assign or state a linguistic name for an entity or category", "naming"}}},
	{"en.introducing_people.v", "introduce", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.introducing_people.event", "present people to each other to establish social familiarity", "introducing_people"}}},
	{"en.greeting.v", "greet", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.greeting.event", "acknowledge another participant at the start of an encounter", "greeting"}}},
	{"en.thanking.v", "thank", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.thanking.event", "express gratitude to another participant for a perceived benefit", "thanking"}}},
	{"en.apologizing.v", "apologize", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.apologizing.event", "express regret or take responsibility for an action perceived as wrong", "apologizing"}}},
	{"en.requesting_action.v", "request", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.requesting_action.event", "ask another participant to perform an action or grant a resource", "requesting_action"}}},
	{"en.commanding.v", "command", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.commanding.event", "direct another participant to act with asserted authority", "commanding"}}},
	{"en.inviting.v", "invite", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.inviting.event", "offer another participant the opportunity to attend or participate", "inviting"}}},
	{"en.making_offer.v", "offer", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.making_offer.event", "propose to provide a resource or perform an action for another", "making_offer"}}},
	{"en.promising.v", "promise", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.promising.event", "commit oneself to a future action or result in relation to another", "promising"}}},
	{"en.agreeing.v", "agree", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.agreeing.event", "express acceptance of a proposal, assertion, or position", "agreeing"}}},
	{"en.disagreeing.v", "disagree", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.disagreeing.event", "express rejection of an assertion or position without necessarily refusing action", "disagreeing"}}},
	{"en.refusing.v", "refuse", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.refusing.event", "decline a request, offer, or proposed course of action", "refusing"}}},
	{"en.accepting_offer.v", "accept", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.accepting_offer.event", "express consent to terms proposed by another participant", "accepting_offer"}}},
	{"en.arguing.v", "argue", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.arguing.event", "engage in a verbal dispute in which positions conflict", "arguing"}}},
	{"en.discussing.v", "discuss", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.discussing.event", "exchange views or information about a shared topic", "discussing"}}},
	{"en.conversing.v", "talk", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.conversing.event", "engage in spoken exchange with another participant without specifying a topic or disagreement", "conversing"}}},
	{"en.persuading.v", "persuade", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.persuading.event", "attempt to change another person’s beliefs or intended action through communication", "persuading"}}},
	{"en.warning.v", "warn", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.warning.event", "alert a recipient to a risk or unfavorable possible occurrence", "warning"}}},
	{"en.advising.v", "advise", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.advising.event", "offer an assessment or recommended course of action for a recipient’s benefit", "advising"}}},
	{"en.recommending.v", "recommend", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.recommending.event", "express positive preference for an option as suitable for a goal", "recommending"}}},
	{"en.complaining.v", "complain", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.complaining.event", "express dissatisfaction about a perceived problem or treatment", "complaining"}}},
	{"en.praising.v", "praise", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.praising.event", "express approval of a participant or their achievement", "praising"}}},
	{"en.criticizing.v", "criticize", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.criticizing.event", "express an unfavorable evaluation of conduct, work, or an outcome", "criticizing"}}},
	{"en.joking.v", "joke", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.joking.event", "produce speech intended primarily to amuse rather than assert literally", "joking"}}},
	{"en.whispering.v", "whisper", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.whispering.event", "speak with very low vocal intensity, typically to limit who hears", "whispering"}}},
	{"en.shouting.v", "shout", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.shouting.event", "produce speech at unusually high vocal intensity", "shouting"}}},
	{"en.writing_message.v", "write", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.writing_message.event", "compose linguistic content in a written medium for a recipient", "writing_message"}}},
	{"en.sending_message.v", "send", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.sending_message.event", "dispatch a message through a physical or electronic channel", "sending_message"}}},
	{"en.receiving_message.v", "receive", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.receiving_message.event", "obtain a message delivered from another participant", "receiving_message"}}},
	{"en.calling_by_phone.v", "call", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.calling_by_phone.event", "initiate or participate in spoken communication through a telephone connection", "calling_by_phone"}}},
	{"en.translating_written.v", "translate", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.translating_written.event", "produce written content in a different language while preserving intended meaning", "translating_written"}}},
	{"en.interpreting_spoken.v", "interpret", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.interpreting_spoken.event", "relay live spoken or signed content into another language for recipients", "interpreting_spoken"}}},
	{"en.meeting_people.v", "meet", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.meeting_people.event", "come together with others for direct social or organized interaction", "meeting_people"}}},
}
