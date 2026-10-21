package catalog

import "github.com/clickbytes/golingua/internal/core"

// everydayDeclarations returns the reviewed, typed concepts in editorial order.
func everydayDeclarations() []declaration {
	return []declaration{
		{"pizza_food", "food", "A baked flat dough base topped with sauce, cheese, or other ingredients.", core.KindEntity, nil},
		{"ice_cream_food", "food", "A frozen sweet dessert typically made from milk or comparable ingredients.", core.KindEntity, nil},
		{"chocolate_food", "food", "A sweet or bitter food preparation derived mainly from cocoa solids or butter.", core.KindEntity, nil},
		{"honey_food", "food", "A viscous sweet substance produced by bees from plant nectar or secretions.", core.KindEntity, nil},
		{"grape_fruit_food", "fruit_food", "A small edible grape berry, distinct from the citrus fruit grapefruit.", core.KindEntity, nil},
		{"gas_station", "location", "A roadside service facility where motor vehicles can obtain fuel.", core.KindEntity, nil},
		{"bank_building", "building", "Premises used by a banking institution to serve customers.", core.KindEntity, nil},
		{"pharmacy_building", "building", "Premises where medicines are dispensed or sold under pharmacy practice.", core.KindEntity, nil},
	}
}
