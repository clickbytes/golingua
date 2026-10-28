package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishLandformMoreSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_landforms_more.go",
	Version: "2026-10-22-batch-23", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name common terrain, water and celestial entities.
var englishLandformMoreEntries = []lexicalEntry{
	{"en.island.n", "island", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.island.direct", "a landmass surrounded by water and smaller than a continent", "island"}}},
	{"en.continent.n", "continent", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.continent.direct", "one of Earth’s major continuous landmass regions", "continent"}}},
	{"en.coast.n", "coast", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.coast.direct", "the land region adjacent to the edge of a sea or ocean", "coast"}}},
	{"en.beach.n", "beach", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.beach.direct", "a shore area commonly covered with sand, pebbles, or similar sediment", "beach"}}},
	{"en.desert.n", "desert", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.desert.direct", "a region with very low precipitation and sparse moisture-dependent vegetation", "desert"}}},
	{"en.forest.n", "forest", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.forest.direct", "an area of land dominated by relatively dense tree cover", "forest"}}},
	{"en.meadow.n", "meadow", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.meadow.direct", "an open area of land dominated by grasses and other low vegetation", "meadow"}}},
	{"en.field_land.n", "field", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.field_land.direct", "an open expanse of land, often distinguished by use or vegetation", "field_land"}}},
	{"en.farmland.n", "farmland", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.farmland.direct", "land managed to grow crops or raise agricultural animals", "farmland"}}},
	{"en.garden.n", "garden", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.garden.direct", "a cultivated space for growing plants or outdoor enjoyment", "garden"}}},
	{"en.park.n", "park", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.park.direct", "a managed open space for recreation, conservation, or public use", "park"}}},
	{"en.river.n", "river", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.river.direct", "a naturally flowing channel of water leading toward another body of water", "river"}}},
	{"en.stream.n", "stream", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stream.direct", "a relatively small natural flowing watercourse", "stream"}}},
	{"en.lake.n", "lake", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lake.direct", "a substantial inland body of standing water surrounded by land", "lake"}}},
	{"en.pond.n", "pond", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pond.direct", "a relatively small, still inland body of water", "pond"}}},
	{"en.sea.n", "sea", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sea.direct", "a large connected body of saline water, usually part of an ocean system", "sea"}}},
	{"en.ocean.n", "ocean", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ocean.direct", "one of Earth’s enormous connected bodies of saline water", "ocean"}}},
	{"en.waterfall.n", "waterfall", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.waterfall.direct", "a place where flowing water drops steeply over an edge", "waterfall"}}},
	{"en.well_water.n", "well", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.well_water.direct", "a constructed shaft or bore providing access to underground water", "well_water"}}},
	{"en.shore.n", "shore", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shore.direct", "the land boundary directly next to a body of water", "shore"}}},
	{"en.sky_region.n", "sky", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sky_region.direct", "the apparent overhead region containing atmospheric and celestial phenomena", "sky_region"}}},
	{"en.cloud.n", "cloud", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cloud.direct", "a visible mass of suspended water droplets or ice crystals in the atmosphere", "cloud"}}},
	{"en.rain_drop.n", "raindrop", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rain_drop.direct", "a bounded droplet of liquid water falling from atmospheric condensation", "rain_drop"}}},
	{"en.rain.n", "rain", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rain.direct", "liquid water falling from clouds as precipitation", "rain"}}},
	{"en.snowflake.n", "snowflake", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.snowflake.direct", "a small aggregate or crystal of atmospheric ice falling as snow", "snowflake"}}},
	{"en.hailstone.n", "hailstone", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hailstone.direct", "a pellet of layered ice formed inside a convective storm", "hailstone"}}},
	{"en.fog.n", "fog", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fog.direct", "a near-surface suspension of water droplets that reduces visibility", "fog"}}},
	{"en.sun.n", "sun", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sun.direct", "the star around which Earth orbits, regarded as a celestial body", "sun"}}},
	{"en.moon_earth.n", "moon", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.moon_earth.direct", "Earth’s natural satellite considered as a celestial body", "moon_earth"}}},
	{"en.star.n", "star", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.star.direct", "a self-luminous celestial body powered by internal physical processes", "star"}}},
	{"en.planet.n", "planet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.planet.direct", "a major body orbiting a star and not itself undergoing stellar fusion", "planet"}}},
	{"en.atmosphere.n", "atmosphere", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.atmosphere.direct", "the envelope of gases surrounding a planet", "atmosphere"}}},
	{"en.earth_planet.n", "Earth", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.earth_planet.direct", "the specific planet Earth, distinct from soil or a plot of land", "earth_planet"}}},
}
