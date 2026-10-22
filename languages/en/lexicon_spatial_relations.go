package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishSpatialRelationSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_spatial_relations.go",
	Version: "2026-10-22-batch-32", License: "unknown", Author: "golingua contributors",
}

// These pending candidates encode spatial relations with English relational forms.
var englishSpatialRelationEntries = []lexicalEntry{
	{"en.relation.n", "relation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.relation.direct", "a recurring semantic link connecting two or more referents without itself being an occurrence in time", "relation"}}},
	{"en.spatial_relation.n", "spatial relation", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.spatial_relation.direct", "a relationship determined by position, direction, extent, or contact in space", "spatial_relation"}}},
	{"en.located_in.prep", "in", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.located_in.relation", "an entity occupies a location, without asserting that the location fully encloses it", "located_in"}}},
	{"en.spatial_containment.prep", "inside", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.spatial_containment.relation", "one bounded region or container encloses another object or region", "spatial_containment"}}},
	{"en.spatial_overlap.n", "overlap", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.spatial_overlap.direct", "two regions have a nonempty portion of space in common", "spatial_overlap"}}},
	{"en.spatial_adjacency.prep", "next to", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.spatial_adjacency.relation", "two entities or regions are immediately adjacent with no relevant intervening region", "spatial_adjacency"}}},
	{"en.spatial_contact.prep", "in contact with", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.spatial_contact.relation", "two physical bodies have points of their surfaces in direct contact", "spatial_contact"}}},
	{"en.vertical_order.prep", "above", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.vertical_order.relation", "one referent is vertically higher than another relative to gravity", "vertical_order"}}},
	{"en.lateral_order.prep", "to the left of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.lateral_order.relation", "one referent lies to the left of another from a selected viewpoint", "lateral_order"}}},
	{"en.front_back_order.prep", "in front of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.front_back_order.relation", "one referent lies in front of another relative to a specified orientation", "front_back_order"}}},
	{"en.between_position.prep", "between", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.between_position.relation", "a referent lies along or within the spatial interval between two anchors", "between_position"}}},
	{"en.distance_between.n", "distance", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.distance_between.direct", "a measurable length separating two spatial referents", "distance_between"}}},
	{"en.compass_north_south.prep", "north of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.compass_north_south.relation", "one geographic location lies north of another under a terrestrial compass frame", "compass_north_south"}}},
	{"en.compass_east_west.prep", "east of", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.compass_east_west.relation", "one geographic location lies east of another under a terrestrial compass frame", "compass_east_west"}}},
	{"en.proximity_relation.prep", "near", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.proximity_relation.relation", "two referents are near one another under a context-dependent distance threshold", "proximity_relation"}}},
	{"en.spatial_separation.prep", "apart from", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.spatial_separation.relation", "a spatial gap with no direct overlap lies between two referents", "spatial_separation"}}},
	{"en.supports_physical.n", "support", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.supports_physical.direct", "one physical entity bears another’s load against gravity or another applied force", "supports_physical"}}},
	{"en.connected_physically.prep", "connected to", lexicon.PartOfSpeech("preposition"), []lexicalMeaning{{"en.connected_physically.relation", "a continuous physical pathway or attachment links two objects or places", "connected_physically"}}},
}
