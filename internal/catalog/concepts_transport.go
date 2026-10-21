package catalog

import "github.com/clickbytes/golingua/internal/core"

// transportDeclarations returns the reviewed, typed concepts in editorial order.
func transportDeclarations() []declaration {
	return []declaration{
		{"car", "vehicle", "A road motor vehicle primarily designed to carry a small number of passengers.", core.KindEntity, nil},
		{"bus", "vehicle", "A road vehicle designed to carry many passengers along a service or route.", core.KindEntity, nil},
		{"train", "vehicle", "A connected rail vehicle or set of vehicles moving on a track.", core.KindEntity, nil},
		{"bicycle", "vehicle", "A human-powered two-wheeled vehicle propelled by pedals.", core.KindEntity, nil},
		{"motorcycle", "vehicle", "A powered two-wheeled road vehicle ridden astride by its operator.", core.KindEntity, nil},
		{"truck", "vehicle", "A road motor vehicle built primarily for transporting cargo.", core.KindEntity, nil},
		{"airplane", "vehicle", "A powered fixed-wing aircraft that gains lift while traveling through air.", core.KindEntity, nil},
		{"boat", "vehicle", "A relatively small watercraft capable of carrying people or goods on water.", core.KindEntity, nil},
		{"ship", "vehicle", "A large watercraft equipped for sustained travel and transport over water.", core.KindEntity, nil},
		{"taxi_vehicle", "car", "A passenger car operated to transport paying customers on demand.", core.KindEntity, nil},
		{"tram", "vehicle", "A passenger rail vehicle typically operating along urban street tracks.", core.KindEntity, nil},
		{"subway_train", "train", "A train designed for rapid urban transit, commonly using segregated tracks.", core.KindEntity, nil},
		{"ambulance", "vehicle", "A specially equipped vehicle for transporting ill or injured people for medical care.", core.KindEntity, nil},
	}
}
