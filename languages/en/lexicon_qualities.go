package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishQualitySource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_qualities.go",
	Version: "2026-10-22-batch-24", License: "unknown", Author: "golingua contributors",
}

// These pending candidates distinguish measurable dimensions and material qualities.
var englishQualityEntries = []lexicalEntry{
	{"en.length.n", "length", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.length.direct", "extent measured along an object’s principal longitudinal direction", "length"}}},
	{"en.width.n", "width", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.width.direct", "extent measured across an object’s lateral dimension", "width"}}},
	{"en.height.n", "height", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.height.direct", "vertical extent of an object or position above a reference level", "height"}}},
	{"en.depth.n", "depth", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.depth.direct", "extent inward from a reference surface or downward from a reference level", "depth"}}},
	{"en.thickness.n", "thickness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.thickness.direct", "distance between two opposing surfaces of a material or object", "thickness"}}},
	{"en.area.n", "area", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.area.direct", "two-dimensional extent of a surface or region", "area"}}},
	{"en.volume.n", "volume", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.volume.direct", "three-dimensional extent of a bounded region or object", "volume"}}},
	{"en.mass.n", "mass", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mass.direct", "amount of matter in a body as characterized by its inertia", "mass"}}},
	{"en.weight.n", "weight", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.weight.direct", "force exerted on a body by gravity in a given environment", "weight"}}},
	{"en.shape.n", "shape", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shape.direct", "geometric organization of an object’s boundary, independent of its material", "shape"}}},
	{"en.curvature.n", "curvature", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.curvature.direct", "degree to which a line or surface departs locally from straightness or flatness", "curvature"}}},
	{"en.roundness.n", "roundness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.roundness.direct", "degree to which a boundary approximates a circle or sphere", "roundness"}}},
	{"en.straightness.n", "straightness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.straightness.direct", "degree to which a path or edge avoids bending or curvature", "straightness"}}},
	{"en.symmetry.n", "symmetry", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.symmetry.direct", "degree of invariance of a form under a specified geometric transformation", "symmetry"}}},
	{"en.density.n", "density", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.density.direct", "mass per unit volume of a material or bounded body", "density"}}},
	{"en.texture.n", "texture", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.texture.direct", "character of small-scale surface structure as felt or seen", "texture"}}},
	{"en.hardness.n", "hardness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hardness.direct", "resistance of material to local deformation, scratching, or indentation", "hardness"}}},
	{"en.flexibility.n", "flexibility", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.flexibility.direct", "capacity of an object to bend without losing required integrity", "flexibility"}}},
	{"en.transparency.n", "transparency", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.transparency.direct", "degree to which an object permits light to pass without major scattering", "transparency"}}},
	{"en.brightness.n", "brightness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.brightness.direct", "perceived or measured intensity of light from a surface or source", "brightness"}}},
	{"en.sound_loudness.n", "loudness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sound_loudness.direct", "perceptual magnitude of a sound’s intensity", "sound_loudness"}}},
	{"en.sound_pitch.n", "pitch", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sound_pitch.direct", "perceived highness or lowness of a sound chiefly associated with frequency", "sound_pitch"}}},
	{"en.sound_timbre.n", "timbre", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sound_timbre.direct", "qualitative tone character distinguishing sounds of similar pitch and loudness", "sound_timbre"}}},
	{"en.smell_character.n", "smell", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.smell_character.direct", "qualitative sensory character of an odor, distinct from its intensity", "smell_character"}}},
	{"en.taste_character.n", "taste", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.taste_character.direct", "qualitative gustatory character of a substance", "taste_character"}}},
	{"en.sweetness.n", "sweetness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sweetness.direct", "perceived intensity of a sweet taste component", "sweetness"}}},
	{"en.sourness.n", "sourness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sourness.direct", "perceived intensity of an acidic or sour taste component", "sourness"}}},
	{"en.bitterness.n", "bitterness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bitterness.direct", "perceived intensity of a bitter taste component", "bitterness"}}},
	{"en.saltiness.n", "saltiness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.saltiness.direct", "perceived intensity of a salty taste component", "saltiness"}}},
	{"en.pungency.n", "pungency", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pungency.direct", "degree of burning or irritating oral or nasal sensation caused by substances", "pungency"}}},
	{"en.moisture_content.n", "moisture", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.moisture_content.direct", "amount of water held in or on a material", "moisture_content"}}},
	{"en.air_humidity.n", "humidity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.air_humidity.direct", "concentration of water vapor in surrounding air", "air_humidity"}}},
	{"en.cleanliness.n", "cleanliness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cleanliness.direct", "degree to which a surface or object is free of unwanted dirt or contaminants", "cleanliness"}}},
	{"en.sharpness.n", "sharpness", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sharpness.direct", "ability of an edge or point to concentrate force for cutting or piercing", "sharpness"}}},
	{"en.flammability.n", "flammability", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.flammability.direct", "ease with which a material can ignite and sustain combustion", "flammability"}}},
	{"en.acidity.n", "acidity", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.acidity.direct", "acid strength or hydrogen-ion-related condition of a solution", "acidity"}}},
	{"en.pressure.n", "pressure", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pressure.direct", "mechanical force distributed per unit area of a surface", "pressure"}}},
	{"en.energy_content.n", "energy", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.energy_content.direct", "amount of energy stored in or available from a system", "energy_content"}}},
	{"en.durability.n", "durability", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.durability.direct", "capacity of an object to retain function despite use or environmental exposure", "durability"}}},
}
