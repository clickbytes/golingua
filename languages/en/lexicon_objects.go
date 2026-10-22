package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishObjectSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_objects.go",
	Version: "2026-10-22-batch-19", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name common furniture, containers, utensils and tools.
var englishObjectEntries = []lexicalEntry{
	{"en.furniture.n", "furniture", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.furniture.object", "movable articles used to support domestic or workplace activities", "furniture"}}},
	{"en.table.n", "table", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.table.object", "an item of furniture with an elevated horizontal surface for placing things", "table"}}},
	{"en.desk.n", "desk", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.desk.object", "a table-like piece of furniture mainly used for writing or office work", "desk"}}},
	{"en.chair.n", "chair", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.chair.object", "a seat for one person, usually with a back", "chair"}}},
	{"en.sofa.n", "sofa", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.sofa.object", "a padded seat designed for several people", "sofa"}}},
	{"en.bed.n", "bed", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bed.object", "a piece of furniture for sleeping or resting", "bed"}}},
	{"en.wardrobe.n", "wardrobe", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.wardrobe.object", "a tall storage cabinet chiefly used for clothing", "wardrobe"}}},
	{"en.cabinet.n", "cabinet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cabinet.object", "an enclosed storage unit with doors or drawers", "cabinet"}}},
	{"en.shelf.n", "shelf", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shelf.object", "a horizontal support for holding stored or displayed things", "shelf"}}},
	{"en.bookcase.n", "bookcase", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bookcase.object", "a piece of furniture with shelves for storing books", "bookcase"}}},
	{"en.drawer.n", "drawer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.drawer.object", "a box-like storage compartment that slides out of furniture", "drawer"}}},
	{"en.mirror.n", "mirror", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mirror.object", "a surface that reflects an image of objects before it", "mirror"}}},
	{"en.lamp.n", "lamp", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lamp.object", "a device that produces controlled illumination", "lamp"}}},
	{"en.light_bulb.n", "light bulb", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.light_bulb.object", "a replaceable electric source of light in a luminaire", "light_bulb"}}},
	{"en.clock_device.n", "clock", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.clock_device.object", "a device that indicates the time of day", "clock_device"}}},
	{"en.watch_device.n", "watch", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.watch_device.object", "a small wearable device for telling the time", "watch_device"}}},
	{"en.key_tool.n", "key", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.key_tool.object", "a shaped device that operates a corresponding lock", "key_tool"}}},
	{"en.lock_device.n", "lock", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lock_device.object", "a mechanism that restricts access or movement until released", "lock_device"}}},
	{"en.handle_object.n", "handle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.handle_object.object", "a part designed to be grasped by hand when manipulating an object", "handle_object"}}},
	{"en.container.n", "container", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.container.object", "an object made to hold, transport, or protect contents", "container"}}},
	{"en.box.n", "box", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.box.object", "a rigid container with a bounded base and sides, often rectangular", "box"}}},
	{"en.bag.n", "bag", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bag.object", "a flexible container used for carrying or storing things", "bag"}}},
	{"en.backpack.n", "backpack", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.backpack.object", "a bag carried on the back with shoulder straps", "backpack"}}},
	{"en.suitcase.n", "suitcase", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.suitcase.object", "a portable closable case for carrying luggage during travel", "suitcase"}}},
	{"en.bottle.n", "bottle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bottle.object", "a narrow-necked rigid container chiefly used to store liquids", "bottle"}}},
	{"en.jar.n", "jar", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.jar.object", "a wide-mouthed rigid container commonly closed with a removable lid", "jar"}}},
	{"en.cup.n", "cup", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.cup.object", "a small open vessel used for drinking, often with a handle", "cup"}}},
	{"en.glass_drinking.n", "drinking glass", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.glass_drinking.object", "a handleless vessel used for drinking, commonly made of glass", "glass_drinking"}}},
	{"en.mug.n", "mug", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mug.object", "a relatively large handled cup used for drinking", "mug"}}},
	{"en.plate.n", "plate", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.plate.object", "a shallow dish for holding food while it is served or eaten", "plate"}}},
	{"en.bowl.n", "bowl", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bowl.object", "a deep open dish for holding food, liquids, or small objects", "bowl"}}},
	{"en.pot_cooking.n", "pot", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pot_cooking.object", "a deep heat-resistant vessel used for cooking food", "pot_cooking"}}},
	{"en.pan_cooking.n", "pan", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.pan_cooking.object", "a shallow metal vessel used to cook food by direct heat", "pan_cooking"}}},
	{"en.kettle.n", "kettle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.kettle.object", "a vessel or appliance designed to heat water", "kettle"}}},
	{"en.spoon.n", "spoon", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.spoon.object", "a utensil with a small concave end for scooping or stirring", "spoon"}}},
	{"en.fork_utensil.n", "fork", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fork_utensil.object", "a pronged utensil used to lift or hold food", "fork_utensil"}}},
	{"en.knife_utensil.n", "knife", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.knife_utensil.object", "a cutting utensil with a blade and handle, especially for food", "knife_utensil"}}},
	{"en.scissors.n", "scissors", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.scissors.object", "a cutting tool with two pivoted blades and handles", "scissors"}}},
	{"en.hammer.n", "hammer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hammer.object", "a hand tool with a weighted head used to deliver impacts", "hammer"}}},
	{"en.screwdriver.n", "screwdriver", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.screwdriver.object", "a tool used to turn screws by engaging their heads", "screwdriver"}}},
	{"en.saw_tool.n", "saw", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.saw_tool.object", "a toothed blade tool used to cut through material", "saw_tool"}}},
	{"en.rope.n", "rope", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.rope.object", "a strong flexible length made by twisting or braiding fibers or strands", "rope"}}},
	{"en.glue.n", "glue", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.glue.object", "an adhesive material applied to bond surfaces together", "glue"}}},
	{"en.brush_tool.n", "brush", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.brush_tool.object", "a tool with bristles used to apply, clean, or distribute substances", "brush_tool"}}},
	{"en.toothbrush.n", "toothbrush", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.toothbrush.object", "a small bristled implement used for cleaning teeth", "toothbrush"}}},
	{"en.comb.n", "comb", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.comb.object", "a toothed implement used to arrange or detangle hair", "comb"}}},
	{"en.towel.n", "towel", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.towel.object", "a textile article used to absorb water and dry surfaces or bodies", "towel"}}},
	{"en.soap.n", "soap", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.soap.object", "a cleansing preparation that helps detach oils and dirt in water", "soap"}}},
	{"en.broom.n", "broom", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.broom.object", "a cleaning implement with a brush head on a handle, used for sweeping", "broom"}}},
	{"en.vacuum_cleaner.n", "vacuum cleaner", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.vacuum_cleaner.object", "a powered cleaning machine that draws in dust using airflow", "vacuum_cleaner"}}},
}
