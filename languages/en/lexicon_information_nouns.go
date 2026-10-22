package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishInformationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_information_nouns.go",
	Version: "2026-10-22-batch-29", License: "unknown", Author: "golingua contributors",
}

// These pending nouns distinguish information content from its carrier.
var englishInformationEntries = []lexicalEntry{
	{"en.abstract.n", "abstraction", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.abstract.content", "an identifiable meaning, value, or conceptual object that is not itself a physical continuant or occurrence", "abstract"}}},
	{"en.information_content.n", "information", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.information_content.content", "meaning encoded or conveyable by a representation independently of its physical medium", "information_content"}}},
	{"en.proposition.n", "proposition", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.proposition.content", "a content unit that can be evaluated as true or false under an interpretation", "proposition"}}},
	{"en.assertion_content.n", "assertion", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.assertion_content.content", "a proposition presented by a speaker or document as holding", "assertion_content"}}},
	{"en.question_content.n", "question", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.question_content.content", "a specified gap in knowledge and the information sought to resolve it", "question_content"}}},
	{"en.answer_content.n", "answer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.answer_content.content", "content offered to resolve a particular information-seeking question", "answer_content"}}},
	{"en.hypothesis.n", "hypothesis", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hypothesis.content", "a provisional explanatory proposition advanced for investigation or testing", "hypothesis"}}},
	{"en.belief_content.n", "belief", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.belief_content.content", "a proposition held with subjective acceptance by a cognizer, without guaranteed truth", "belief_content"}}},
	{"en.opinion_content.n", "opinion", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.opinion_content.content", "an evaluative judgment or personal position whose acceptance may vary between people", "opinion_content"}}},
	{"en.instruction_content.n", "instruction", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.instruction_content.content", "represented directions telling an agent how or when to perform an action", "instruction_content"}}},
	{"en.request_content.n", "request", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.request_content.content", "represented desire that another participant act or supply something", "request_content"}}},
	{"en.rule_content.n", "rule", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rule_content.content", "a general constraint or prescription applicable to defined situations", "rule_content"}}},
	{"en.plan_content.n", "plan", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.plan_content.content", "a representation of intended future steps arranged toward a goal", "plan_content"}}},
	{"en.goal_content.n", "goal", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.goal_content.content", "a desired outcome or state that can guide purposeful action", "goal_content"}}},
	{"en.evidence_content.n", "evidence", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.evidence_content.content", "information offered as grounds bearing on acceptance of a claim", "evidence_content"}}},
	{"en.explanation_content.n", "explanation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.explanation_content.content", "an account identifying reasons, mechanisms, or circumstances for a result", "explanation_content"}}},
	{"en.knowledge_content.n", "knowledge", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.knowledge_content.content", "information treated as established within a specified knowledge context", "knowledge_content"}}},
	{"en.memory_content.n", "memory", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.memory_content.content", "a mental representation retained or retrievable from earlier experience", "memory_content"}}},
	{"en.possibility.n", "possibility", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.possibility.content", "a state of affairs compatible with specified circumstances or constraints", "possibility"}}},
	{"en.necessity.n", "necessity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.necessity.content", "a condition in which alternatives are excluded by specified facts or rules", "necessity"}}},
}
