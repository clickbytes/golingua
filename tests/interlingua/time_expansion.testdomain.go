// Package interlingua holds independent A1.10 time and multiplicity fixtures.
package interlingua

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type timeExpansionSpec struct {
	index                      int
	predicate, roles, temporal string
}

// These descriptions contain only semantic predicates, roles, counts, and
// time. Authored bilingual surface forms remain in the selected source data.
var timeExpansionSpecs = []timeExpansionSpec{
	{1, "seeing", "experiencer=person:one:speaker,stimulus=dog:one", "present"},
	{2, "seeing", "experiencer=person:one:addressee,stimulus=cat:one", "present"},
	{3, "seeing", "experiencer=person:many:speaker_group,stimulus=bird:one", "present"},
	{4, "seeing", "experiencer=person:many:addressee_group,stimulus=book:one", "present"},
	{5, "hearing", "experiencer=person:one:speaker,stimulus=dog:one", "present"},
	{6, "hearing", "experiencer=person:one:addressee,stimulus=bird:one", "present"},
	{7, "hearing", "experiencer=person:many:speaker_group,stimulus=cat:one", "present"},
	{8, "hearing", "experiencer=person:many:addressee_group,stimulus=child:one", "present"},
	{9, "reading", "reader=person:one:speaker,material=book:one", "present"},
	{10, "reading", "reader=person:one:addressee,material=message:one", "present"},
	{11, "reading", "reader=person:many:speaker_group,material=letter_document:one", "present"},
	{12, "reading", "reader=person:many:addressee_group,material=book:one", "present"},
	{13, "drinking", "drinker=person:one:speaker,liquid=water:mass", "present"},
	{14, "drinking", "drinker=person:one:addressee,liquid=milk_food:mass", "present"},
	{15, "drinking", "drinker=person:many:speaker_group,liquid=coffee_beverage:mass", "present"},
	{16, "drinking", "drinker=person:many:addressee_group,liquid=tea_beverage:mass", "present"},
	{17, "seeing", "experiencer=teacher:one,stimulus=dog:one", "past"},
	{18, "seeing", "experiencer=teacher:many,stimulus=dog:one", "past"},
	{19, "hearing", "experiencer=doctor:one,stimulus=bird:one", "past"},
	{20, "hearing", "experiencer=doctor:many,stimulus=bird:one", "past"},
	{21, "reading", "reader=mother:one,material=letter_document:one", "past"},
	{22, "reading", "reader=mother:many,material=letter_document:one", "past"},
	{23, "drinking", "drinker=father:one,liquid=water:mass", "past"},
	{24, "drinking", "drinker=father:many,liquid=water:mass", "past"},
	{25, "eating", "eater=child:one,food=bread:mass", "past"},
	{26, "eating", "eater=child:many,food=bread:mass", "past"},
	{27, "sleeping", "sleeper=cat:one", "past"},
	{28, "sleeping", "sleeper=cat:many", "past"},
	{29, "running", "runner=dog:one", "past"},
	{30, "running", "runner=dog:many", "past"},
	{31, "flying_motion", "mover=bird:one", "past"},
	{32, "flying_motion", "mover=bird:many", "past"},
	{33, "reading", "reader=doctor:one,material=letter_document:one", "present"},
	{34, "reading", "reader=doctor:many,material=letter_document:one", "present"},
	{35, "drinking", "drinker=child:one,liquid=milk_food:mass", "present"},
	{36, "drinking", "drinker=child:many,liquid=milk_food:mass", "present"},
	{37, "eating", "eater=cat:one,food=bread:mass", "present"},
	{38, "eating", "eater=cat:many,food=bread:mass", "present"},
	{39, "sleeping", "sleeper=dog:one", "present"},
	{40, "sleeping", "sleeper=dog:many", "present"},
	{41, "black", "bearer=book:one", "past"},
	{42, "black", "bearer=book:many", "past"},
	{43, "black", "bearer=window:one", "past"},
	{44, "black", "bearer=window:many", "past"},
	{45, "black", "bearer=door:one", "past"},
	{46, "black", "bearer=dog:one", "past"},
	{47, "black", "bearer=dog:many", "past"},
	{48, "black", "bearer=cat:one", "past"},
}

// TimeExpansion returns one independent ordinary A1.10 interpretation. IDs
// are a1.10.time.NNN.en|pl with optional .past, .present, or .offspring
// suffixes. The base child concept is the age-child reading; .offspring selects
// child_of_person. Row 22's English base is past and .present selects its
// alternate reading.
// Rows 49-56 are intentionally unavailable: they require future, progressive,
// frequency, or linked-event time beyond the Meaning v1 temporal frames.
func TimeExpansion(id string) (IdentificationMeaning, bool, error) {
	parts := strings.Split(id, ".")
	if len(parts) < 5 || len(parts) > 7 || parts[0] != "a1" || parts[1] != "10" || parts[2] != "time" {
		return IdentificationMeaning{}, false, nil
	}
	index, err := strconv.Atoi(parts[3])
	if err != nil || index < 1 || index > 56 || (parts[4] != "en" && parts[4] != "pl") {
		return IdentificationMeaning{}, false, nil
	}
	variant, frameVariant := "", ""
	for _, part := range parts[5:] {
		switch part {
		case "past", "present":
			if frameVariant != "" {
				return IdentificationMeaning{}, false, nil
			}
			frameVariant = part
		case "offspring":
			if variant != "" {
				return IdentificationMeaning{}, false, nil
			}
			variant = part
		default:
			return IdentificationMeaning{}, false, nil
		}
	}
	if index > 48 {
		return IdentificationMeaning{}, false, nil
	}
	var spec timeExpansionSpec
	for _, candidate := range timeExpansionSpecs {
		if candidate.index == index {
			spec = candidate
			break
		}
	}
	if spec.index == 0 {
		return IdentificationMeaning{}, false, nil
	}
	frame := spec.temporal
	if frameVariant == "past" {
		if parts[4] != "en" || !englishReadPastAllowed(index) {
			return IdentificationMeaning{}, false, nil
		}
		frame = "past"
	}
	if frameVariant == "present" {
		if parts[4] != "en" || index != 22 {
			return IdentificationMeaning{}, false, nil
		}
		frame = "present"
	}
	if variant != "" && !strings.Contains(spec.roles, "child:") {
		return IdentificationMeaning{}, false, nil
	}
	if index == 45 && parts[4] == "pl" && variant != "" {
		return IdentificationMeaning{}, false, nil
	}
	return buildTimeExpansion(spec, parts[4], frame, variant)
}

// TimeExpansionIDs enumerates all licensed fixture IDs, including explicit
// English bare-read time readings and offspring readings of child.
func TimeExpansionIDs() []string {
	var out []string
	for _, spec := range timeExpansionSpecs {
		for _, lang := range []string{"en", "pl"} {
			base := fmt.Sprintf("a1.10.time.%03d.%s", spec.index, lang)
			frames := []string{""}
			if lang == "en" && englishReadPastAllowed(spec.index) {
				frames = append(frames, ".past")
			}
			if lang == "en" && spec.index == 22 {
				frames = append(frames, ".present")
			}
			variants := []string{""}
			if strings.Contains(spec.roles, "child:") {
				variants = append(variants, ".offspring")
			}
			for _, frame := range frames {
				for _, variant := range variants {
					out = append(out, base+frame+variant)
				}
			}
		}
	}
	return out
}

func englishReadPastAllowed(index int) bool { return index >= 9 && index <= 12 || index == 34 }

func buildTimeExpansion(spec timeExpansionSpec, language, frame, variant string) (IdentificationMeaning, bool, error) {
	predicate, err := ids.NewConceptID(spec.predicate)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("time predicate: %w", err)
	}
	participants := make([]meaning.Participant, 0, 2)
	referents := make([]meaning.Node, 0, 2)
	mentions := make([]meaning.Mention, 0, 1)
	for i, roleText := range strings.Split(spec.roles, ",") {
		fields := strings.Split(strings.TrimSpace(roleText), "=")
		if len(fields) != 2 {
			return IdentificationMeaning{}, false, fmt.Errorf("invalid time role %q", roleText)
		}
		details := strings.Split(fields[1], ":")
		concept, multiplicity := details[0], ""
		if len(details) > 1 && details[1] != "mass" {
			multiplicity = details[1]
		}
		anchor := ""
		if len(details) > 2 {
			anchor = details[2]
		}
		if concept == "child" {
			if variant == "offspring" {
				concept = "child_of_person"
			}
		}
		conceptID, e := ids.NewConceptID(concept)
		if e != nil {
			return IdentificationMeaning{}, false, fmt.Errorf("time referent concept: %w", e)
		}
		nodeID := meaning.NodeID(fmt.Sprintf("r%d", i+1))
		referent, e := meaning.NewReferent(nodeID, conceptID)
		if e != nil {
			return IdentificationMeaning{}, false, fmt.Errorf("build time referent: %w", e)
		}
		if multiplicity == "one" {
			referent.Multiplicity = meaning.MultiplicityOne
		}
		if multiplicity == "many" {
			referent.Multiplicity = meaning.MultiplicityMany
		}
		// English you leaves participant count unresolved; the PL source keeps
		// overt finite number. Both retain the addressee anchor.
		if concept == "person" && (anchor == "addressee" || anchor == "addressee_group") && language == "en" {
			referent.Multiplicity = ""
		}
		if spec.index == 45 && language == "pl" && concept == "door" {
			referent.Multiplicity = ""
		}
		referents = append(referents, referent)
		participants = append(participants, meaning.Participant{Role: meaning.Role(fields[0]), Target: nodeID})
		if anchor != "" {
			anchorValue := anchor
			switch anchor {
			case "speaker_group":
				anchorValue = "speaker_inclusive_group"
			case "addressee_group":
				anchorValue = "addressee"
			}
			mentions = append(mentions, meaning.Mention{ID: meaning.MentionID("participant"), NodeID: nodeID, Anchor: &meaning.DeicticAnchor{Kind: "participant", Value: anchorValue}})
		}
	}
	var proposition meaning.Node
	if spec.predicate == "black" {
		proposition, err = meaning.NewState("pred", predicate, participants)
	} else {
		proposition, err = meaning.NewEvent("pred", predicate, participants)
	}
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build time proposition: %w", err)
	}
	if frame == "past" {
		proposition.TemporalFrame = meaning.TemporalFramePast
	} else {
		proposition.TemporalFrame = meaning.TemporalFramePresent
	}
	nodes := append([]meaning.Node{proposition}, referents...)
	graph, err := meaning.Build([]meaning.NodeID{"pred"}, nodes, meaning.DefaultBudget())
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build time Meaning: %w", err)
	}
	context := meaning.RequestContext{ID: "a1-10-time", Mentions: mentions}
	if err := meaning.ValidateReferences(graph, context, meaning.DefaultBudget()); err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("validate time references: %w", err)
	}
	return IdentificationMeaning{Graph: graph, Context: context}, true, nil
}
