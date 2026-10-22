package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishPlantSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_plants.go",
	Version: "2026-10-22-batch-17", License: "unknown", Author: "golingua contributors",
}

// These pending noun candidates distinguish whole plants from plant structures.
var englishPlantEntries = []lexicalEntry{
	{"en.shrub.n", "shrub", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shrub.woody_plant", "a woody perennial plant with multiple stems near ground level", "shrub"}}},
	{"en.grass_plant.n", "grass", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.grass_plant.grass_family_plant", "a narrow-leaved plant of the grass family, not ground cover as material", "grass_plant"}}},
	{"en.flowering_plant.n", "flowering plant", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.flowering_plant.reproductive_plant", "a plant that produces flowers as reproductive structures", "flowering_plant"}}},
	{"en.apple_tree.n", "apple tree", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.apple_tree.fruit_tree", "a tree cultivated or growing to produce apples", "apple_tree"}}},
	{"en.root_plant.n", "root", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.root_plant.plant_organ", "a plant organ usually anchoring the plant and absorbing water or minerals", "root_plant"}}},
	{"en.stem_plant.n", "stem", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stem_plant.plant_axis", "the primary supporting and conducting axis of a plant", "stem_plant"}}},
	{"en.branch_plant.n", "branch", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.branch_plant.plant_extension", "a woody or herbaceous extension growing from a plant's main axis", "branch_plant"}}},
	{"en.leaf_plant.n", "leaf", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.leaf_plant.photosynthetic_organ", "a plant organ specialized primarily for photosynthesis and gas exchange", "leaf_plant"}}},
	{"en.flower_organ.n", "flower", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.flower_organ.reproductive_structure", "the reproductive structure of a flowering plant, distinct from a whole plant", "flower_organ"}}},
	{"en.seed.n", "seed", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.seed.plant_reproductive_unit", "a plant reproductive unit containing an embryo and stored or protective material", "seed"}}},
	{"en.fruit_botanical.n", "fruit", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fruit_botanical.seed_bearing_structure", "a seed-bearing structure developing from a flowering plant's ovary", "fruit_botanical"}}},
	{"en.bark_tree.n", "bark", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bark_tree.outer_tissue", "the protective outer tissue layers of a woody stem", "bark_tree"}}},
	{"en.trunk_tree.n", "trunk", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.trunk_tree.main_woody_stem", "a tree's main upright woody stem from which branches grow", "trunk_tree"}}},
}
