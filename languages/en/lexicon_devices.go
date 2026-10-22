package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishDeviceSource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_devices.go",
	Version: "2026-10-22-batch-20", License: "unknown", Author: "golingua contributors",
}

// These pending candidates name common home and information technology devices.
var englishDeviceEntries = []lexicalEntry{
	{"en.refrigerator.n", "refrigerator", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.refrigerator.direct", "an appliance with a cooled compartment for preserving contents", "refrigerator"}}},
	{"en.freezer.n", "freezer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.freezer.direct", "an appliance or compartment that keeps stored items frozen", "freezer"}}},
	{"en.stove_appliance.n", "stove", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stove_appliance.direct", "a fixed or portable appliance that provides heat for cooking", "stove_appliance"}}},
	{"en.oven.n", "oven", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.oven.direct", "an enclosed appliance that heats food for baking or roasting", "oven"}}},
	{"en.microwave_oven.n", "microwave oven", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.microwave_oven.direct", "an appliance that heats food using microwave energy", "microwave_oven"}}},
	{"en.washing_machine.n", "washing machine", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.washing_machine.direct", "an appliance that mechanically washes clothes or textiles with water", "washing_machine"}}},
	{"en.dishwasher.n", "dishwasher", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.dishwasher.direct", "an appliance that automatically cleans dishes and utensils", "dishwasher"}}},
	{"en.air_conditioner.n", "air conditioner", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.air_conditioner.direct", "an appliance that cools or conditions indoor air", "air_conditioner"}}},
	{"en.heater_device.n", "heater", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.heater_device.direct", "a device intended to raise the temperature of a space or material", "heater_device"}}},
	{"en.fan_device.n", "fan", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.fan_device.direct", "a rotating-blade device that moves air", "fan_device"}}},
	{"en.computer.n", "computer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.computer.direct", "an electronic programmable machine that processes digital information", "computer"}}},
	{"en.laptop_computer.n", "laptop", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.laptop_computer.direct", "a portable computer with an integrated screen and keyboard", "laptop_computer"}}},
	{"en.tablet_computer.n", "tablet", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tablet_computer.direct", "a portable flat computer primarily operated through a touch display", "tablet_computer"}}},
	{"en.smartphone.n", "smartphone", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.smartphone.direct", "a portable telephone with general computing and network functions", "smartphone"}}},
	{"en.telephone_device.n", "telephone", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.telephone_device.direct", "a device for transmitting and receiving speech over a communications network", "telephone_device"}}},
	{"en.keyboard_device.n", "keyboard", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.keyboard_device.direct", "a set of keys used to enter characters or commands into a machine", "keyboard_device"}}},
	{"en.mouse_device.n", "computer mouse", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mouse_device.direct", "a handheld pointing device used to control a computer cursor", "mouse_device"}}},
	{"en.computer_monitor.n", "monitor", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.computer_monitor.direct", "a display screen used to present computer-generated information", "computer_monitor"}}},
	{"en.printer_device.n", "printer", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.printer_device.direct", "a device that produces physical text or graphics from digital input", "printer_device"}}},
	{"en.camera_device.n", "camera", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.camera_device.direct", "an apparatus that captures still images or moving visual recordings", "camera_device"}}},
	{"en.headphones.n", "headphones", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.headphones.direct", "sound transducers worn at or over the ears for listening", "headphones"}}},
	{"en.microphone.n", "microphone", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.microphone.direct", "a transducer that converts sound into an electrical or digital signal", "microphone"}}},
	{"en.speaker_device.n", "speaker", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.speaker_device.direct", "a transducer that produces audible sound from an input signal", "speaker_device"}}},
	{"en.television_device.n", "television", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.television_device.direct", "a receiver and display system for broadcast or streamed video", "television_device"}}},
	{"en.radio_device.n", "radio", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.radio_device.direct", "a receiver or transceiver handling information carried by radio waves", "radio_device"}}},
	{"en.remote_control.n", "remote control", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.remote_control.direct", "a handheld device that sends commands to another apparatus from a distance", "remote_control"}}},
	{"en.battery.n", "battery", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.battery.direct", "a device that stores chemical energy for delivery as electrical energy", "battery"}}},
	{"en.electric_cable.n", "electric cable", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.electric_cable.direct", "a flexible assembly of conductors transmitting electrical power or signals", "electric_cable"}}},
	{"en.plug_electric.n", "plug", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.plug_electric.direct", "a connector inserted into a socket to make an electrical connection", "plug_electric"}}},
}
