package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishMaterialSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_materials.go",
	Version: "2026-10-22-batch-18", License: "unknown", Author: "golingua contributors",
}

// These pending nouns distinguish material substances from bounded objects.
var englishMaterialEntries = []lexicalEntry{
	{"en.physical_substance.n", "substance", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.physical_substance.material", "matter regarded as a material rather than as an individually bounded object", "physical_substance"}}},
	{"en.air.n", "air", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.air.ambient_gas_mixture", "the mixture of gases forming the ambient atmosphere near Earth's surface", "air"}}},
	{"en.oxygen.n", "oxygen", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.oxygen.element_material", "the chemical element oxygen in a material form such as gas", "oxygen"}}},
	{"en.steam.n", "steam", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.steam.gaseous_water", "water in gaseous form, distinguished from liquid water and suspended droplets", "steam"}}},
	{"en.ice.n", "ice", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ice.solid_water", "water in its solid crystalline state", "ice"}}},
	{"en.fresh_water.n", "fresh water", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fresh_water.low_salinity", "water with low dissolved salt content, as opposed to saline water", "fresh_water"}}},
	{"en.salt_water.n", "salt water", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.salt_water.saline", "water containing substantial dissolved salts", "salt_water"}}},
	{"en.soil.n", "soil", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.soil.terrestrial_growth_medium", "the mixed mineral and organic material supporting terrestrial plant growth", "soil"}}},
	{"en.sand.n", "sand", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sand.granular_mineral", "loose granular mineral particles smaller than gravel", "sand"}}},
	{"en.mud.n", "mud", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mud.wet_soil_mixture", "a soft mixture of wet soil or sediment and water", "mud"}}},
	{"en.rock_material.n", "rock", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rock_material.consolidated_mineral", "naturally consolidated mineral matter, distinguished from one stone object", "rock_material"}}},
	{"en.stone_piece.n", "stone", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stone_piece.bounded_rock", "a bounded piece of natural rock of portable or identifiable extent", "stone_piece"}}},
	{"en.wood_material.n", "wood", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.wood_material.tree_fiber", "the fibrous structural material of trees used as a substance", "wood_material"}}},
	{"en.metal_material.n", "metal", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.metal_material.material_class", "a class of typically conductive, malleable elemental or alloyed materials", "metal_material"}}},
	{"en.iron.n", "iron", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.iron.metal_material", "a material consisting mainly of the chemical element iron", "iron"}}},
	{"en.copper.n", "copper", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.copper.metal_material", "a material made predominantly of the element copper", "copper"}}},
	{"en.gold_material.n", "gold", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.gold_material.precious_metal", "the precious metallic element gold considered as material", "gold_material"}}},
	{"en.silver_material.n", "silver", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.silver_material.metal", "the metallic element silver considered as material", "silver_material"}}},
	{"en.glass_material.n", "glass", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.glass_material.amorphous_material", "a hard usually transparent amorphous material made by cooling a melt", "glass_material"}}},
	{"en.plastic_material.n", "plastic", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.plastic_material.polymer", "a moldable synthetic or semi-synthetic polymer material", "plastic_material"}}},
	{"en.rubber_material.n", "rubber", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rubber_material.elastic_polymer", "an elastic polymer material, natural or synthetic", "rubber_material"}}},
	{"en.paper_material.n", "paper", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.paper_material.sheet_material", "a thin sheet-form material made from pressed fibers", "paper_material"}}},
	{"en.cloth_material.n", "cloth", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cloth_material.textile", "a flexible textile material formed from interlaced or otherwise joined fibers", "cloth_material"}}},
	{"en.leather_material.n", "leather", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.leather_material.processed_hide", "animal hide processed into a durable flexible material", "leather_material"}}},
	{"en.oil_substance.n", "oil", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.oil_substance.immiscible_liquid", "a liquid oily substance that resists mixing uniformly with water", "oil_substance"}}},
	{"en.fuel.n", "fuel", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fuel.energy_release_material", "a material consumed to release usable energy by reaction", "fuel"}}},
	{"en.coal.n", "coal", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.coal.fossil_material", "a carbon-rich fossil material formed from ancient plant matter", "coal"}}},
	{"en.gasoline.n", "gasoline", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.gasoline.engine_fuel", "a refined petroleum fuel intended primarily for spark-ignition engines", "gasoline"}}},
	{"en.electricity.n", "electricity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.electricity.electrical_resource", "electric charge or energy in motion or potential, considered as a physical resource", "electricity"}}},
	{"en.light_physical.n", "light", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.light_physical.visible_radiation", "visible electromagnetic radiation, not a lighting device", "light_physical"}}},
	{"en.shadow.n", "shadow", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shadow.occluded_region", "a region receiving less illumination because an object obstructs light", "shadow"}}},
	{"en.flame.n", "flame", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.flame.combustion_region", "a visible region of combustion with hot reacting gases", "flame"}}},
	{"en.smoke.n", "smoke", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.smoke.combustion_emission", "airborne particles and gases produced by combustion or heating", "smoke"}}},
	{"en.ash.n", "ash", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ash.burn_residue", "solid residue remaining after material has burned", "ash"}}},
}
