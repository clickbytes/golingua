package pl

import (
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/valency"
)

var outcomeActionFrames = []frameDeclaration{
	{"pl.rozkazywac.command.action", "pl.rozkazywac.command", "pl.ditransitive", []meaning.Role{"commander", "addressee", "action"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "commander", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "addressee", Role: "addressee", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "action", Role: "action", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.udac_sie.succeed.action", "pl.udac_sie.succeed", "pl.oblique", []meaning.Role{"experiencer", "goal"}, []valency.ArgumentSlot{
		{ID: "experiencer", Role: "experiencer", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "goal", Role: "goal", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.nie_udac_sie.fail.action", "pl.nie_udac_sie.fail", "pl.oblique", []meaning.Role{"experiencer", "goal"}, []valency.ArgumentSlot{
		{ID: "experiencer", Role: "experiencer", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "goal", Role: "goal", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}
