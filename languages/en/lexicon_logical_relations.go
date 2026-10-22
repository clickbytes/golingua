package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishLogicalRelationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_logical_relations.go",
	Version: "2026-10-22-batch-33", License: "unknown", Author: "golingua contributors",
}

// These pending entries express logical and explanatory relationships.
var englishLogicalRelationEntries = []lexicalEntry{
	{"en.logical_relation.n", "logical relation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.logical_relation.direct", "a non-spatial relation between propositions, purposes, quantities, or explanatory facts", "logical_relation"}}},
	{"en.causal_dependency.prep", "because of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.causal_dependency.relation", "one event or state contributes causally to bringing about another, not merely preceding it", "causal_dependency"}}},
	{"en.conditional_dependency.prep", "conditional on", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.conditional_dependency.relation", "a proposition or outcome holds only under a specified enabling condition", "conditional_dependency"}}},
	{"en.purpose_relation.prep", "for", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.purpose_relation.relation", "an action or artifact is directed toward achieving a particular intended goal", "purpose_relation"}}},
	{"en.evidence_support.prep", "in support of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.evidence_support.relation", "evidence provides grounds for accepting or rejecting a proposition", "evidence_support"}}},
	{"en.quantitative_order.prep", "greater than", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.quantitative_order.relation", "one measured or numerical value exceeds another on a common ordered scale", "quantitative_order"}}},
}
