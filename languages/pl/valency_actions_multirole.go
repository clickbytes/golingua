package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

// multiRoleActionFrames conservatively licenses named CORE participants and Polish
// syntactic arguments. Template existence is NOT construction approval.
var multiRoleActionFrames = []frameDeclaration{
	{"pl.klac.place.reviewed", "pl.klac.place", "pl.transitive_oblique", []meaning.Role{"agent", "theme", "destination"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "destination", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.wkladac.insert.reviewed", "pl.wkladac.insert", "pl.transitive_oblique", []meaning.Role{"agent", "theme", "container"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "container", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.wyjmowac.extract.reviewed", "pl.wyjmowac.extract", "pl.transitive_oblique", []meaning.Role{"agent", "theme", "container"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "container", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.laczyc.connect.reviewed", "pl.laczyc.connect", "pl.transitive_oblique", []meaning.Role{"agent", "part_a", "part_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "part_a", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "part_b", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
	}},
	{"pl.rozlaczac.disconnect.reviewed", "pl.rozlaczac.disconnect", "pl.transitive_oblique", []meaning.Role{"agent", "part_a", "part_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "part_a", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "part_b", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.napelniac.fill.reviewed", "pl.napelniac.fill", "pl.double_np", []meaning.Role{"agent", "container", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "container", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "complement", Role: "content", Required: true, GrammaticalFunction: "complement", Case: "instrumental"},
	}},
	{"pl.oprozniac.empty.reviewed", "pl.oprozniac.empty", "pl.transitive_oblique", []meaning.Role{"agent", "container", "content"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "container", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "content", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "z"},
	}},
	{"pl.lac.pour.reviewed", "pl.lac.pour", "pl.transitive_oblique", []meaning.Role{"agent", "liquid", "destination"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "liquid", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "destination", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
	{"pl.mieszac.mix.reviewed", "pl.mieszac.mix", "pl.transitive_oblique", []meaning.Role{"agent", "material_a", "material_b"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material_a", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "material_b", Required: true, GrammaticalFunction: "oblique", Case: "instrumental", Adposition: "z"},
	}},
	{"pl.montowac.install.reviewed", "pl.montowac.install", "pl.transitive_oblique", []meaning.Role{"installer", "object", "site"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "installer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "object", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "site", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "w"},
	}},
	{"pl.wymieniac.replace.reviewed", "pl.wymieniac.replace", "pl.transitive_oblique", []meaning.Role{"agent", "old", "new"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "old", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "new", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.parkowac.park.reviewed", "pl.parkowac.park", "pl.transitive_oblique", []meaning.Role{"driver", "vehicle", "location"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "driver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "vehicle", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "location", Required: true, GrammaticalFunction: "oblique", Case: "locative", Adposition: "na"},
	}},
	{"pl.przykrywac.cover.reviewed", "pl.przykrywac.cover", "pl.double_np", []meaning.Role{"agent", "target", "cover"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "target", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "complement", Role: "cover", Required: true, GrammaticalFunction: "complement", Case: "instrumental"},
	}},
	{"pl.kupowac.buy.reviewed", "pl.kupowac.buy", "pl.transitive_oblique", []meaning.Role{"buyer", "goods", "seller"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "buyer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "seller", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.sprzedawac.sell.reviewed", "pl.sprzedawac.sell", "pl.ditransitive", []meaning.Role{"seller", "goods", "buyer"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "seller", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "buyer", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.placic.pay.reviewed", "pl.placic.pay", "pl.ditransitive", []meaning.Role{"payer", "value", "payee"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "payer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "value", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "payee", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.pozyczac.borrow.reviewed", "pl.pozyczac.borrow", "pl.transitive_oblique", []meaning.Role{"borrower", "theme", "lender"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "borrower", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "lender", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.zwracac.return.reviewed", "pl.zwracac.return", "pl.ditransitive", []meaning.Role{"returner", "theme", "owner"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "returner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "owner", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.wynajmowac.rent.reviewed", "pl.wynajmowac.rent", "pl.transitive_oblique", []meaning.Role{"tenant", "property", "lessor"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "tenant", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "property", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "lessor", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "od"},
	}},
	{"pl.darowac.donate.reviewed", "pl.darowac.donate", "pl.ditransitive", []meaning.Role{"donor", "resource", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "donor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "resource", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.dostarczac.deliver.reviewed", "pl.dostarczac.deliver", "pl.ditransitive", []meaning.Role{"courier", "goods", "recipient"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "courier", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
	}},
	{"pl.wysylac.ship.reviewed", "pl.wysylac.ship", "pl.transitive_oblique", []meaning.Role{"sender", "goods", "destination"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "goods", Required: true, GrammaticalFunction: "object", Case: "accusative"},
		{ID: "oblique", Role: "destination", Required: true, GrammaticalFunction: "oblique", Case: "genitive", Adposition: "do"},
	}},
}
