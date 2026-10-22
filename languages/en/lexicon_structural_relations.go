package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishStructuralRelationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_structural_relations.go",
	Version: "2026-10-22-batch-31", License: "unknown", Author: "golingua contributors",
}

// These pending nouns distinguish structural relations and their content.
var englishStructuralRelationEntries = []lexicalEntry{
	{"en.structural_relation.n", "structural relation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.structural_relation.direct", "a relation specifying constitution, organization, control, or representational connection", "structural_relation"}}},
	{"en.part_of.n", "part", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.part_of.direct", "a relation in which one object or component is a constituent of a larger whole", "part_of"}}},
	{"en.material_composition.n", "composition", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.material_composition.direct", "a relation in which a physical item has a particular substance among its constituent materials", "material_composition"}}},
	{"en.attached_to.n", "attachment", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.attached_to.direct", "a relation in which an object is mechanically joined to another without necessarily being a permanent component", "attached_to"}}},
	{"en.component_function.n", "function", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.component_function.direct", "a relation in which a component has a designated practical role within an artifact or system", "component_function"}}},
	{"en.representation_of.n", "representation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.representation_of.direct", "a relation in which an artifact or information structure depicts or encodes another referent", "representation_of"}}},
	{"en.legal_ownership.n", "ownership", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.legal_ownership.direct", "a relation in which a participant holds recognized ownership rights over an identifiable asset", "legal_ownership"}}},
	{"en.temporary_possession.n", "possession", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.temporary_possession.direct", "a relation in which a participant controls or holds an object without implying legal title", "temporary_possession"}}},
	{"en.origin_from.n", "origin", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.origin_from.direct", "a relation in which an entity has a historical source from which it was made or came", "origin_from"}}},
}
