package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishDeviceVerbSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_device_verbs.go",
	Version: "2026-10-22-batch-33", License: "unknown", Author: "golingua contributors",
}

// These pending verbs cover common device operation and maintenance.
var englishDeviceVerbEntries = []lexicalEntry{
	{"en.switching_on_device.v", "turn on", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.switching_on_device.event", "move an apparatus from an off state into an operating or enabled state", "switching_on_device"}}},
	{"en.switching_off_device.v", "turn off", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.switching_off_device.event", "move an apparatus from an operating or enabled state into an off state", "switching_off_device"}}},
	{"en.charging_battery.v", "charge", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.charging_battery.event", "transfer electrical energy into a battery for later use", "charging_battery"}}},
	{"en.installing_object.v", "install", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.installing_object.event", "put a component or device in place so that it can perform its intended role", "installing_object"}}},
	{"en.replacing_component.v", "replace", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.replacing_component.event", "remove one component and install another in its functional position", "replacing_component"}}},
	{"en.throwing_away.v", "discard", lexicon.PartOfSpeech("verb"), []lexicalMeaning{{"en.throwing_away.event", "dispose of an object because the disposer no longer intends to retain it", "throwing_away"}}},
}
