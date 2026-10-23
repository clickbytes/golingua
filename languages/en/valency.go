package en

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

var englishValencySource = provenance.SourceRef{
	ID: "golingua-en-valency-authored", Path: "languages/en/valency.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

var englishConstructionReviewSource = provenance.SourceRef{
	ID:      "golingua-delegated-construction-review",
	URI:     "https://github.com/clickbytes/golingua/issues/63#issuecomment-6351925853",
	Version: "2026-10-22", License: "unknown", Author: "Codex by owner delegation",
}

var englishLookingConstructionReviewSource = provenance.SourceRef{
	ID:      "golingua-delegated-construction-review",
	URI:     "https://github.com/clickbytes/golingua/issues/201#issuecomment-6354741947",
	Version: "2026-10-23-g05-14", License: "unknown", Author: "Codex by owner delegation",
}

var englishOwnerMassObjectConstructionSource = provenance.SourceRef{
	ID:      "golingua-owner-selected-mass-object",
	URI:     "https://github.com/clickbytes/golingua/issues/66#issuecomment-6355642886",
	Version: "2026-10-23-s077-mass-object", License: "internal", Author: "project owner",
}

var englishIndefiniteObjectConstructionSource = provenance.SourceRef{
	ID:      "golingua-g10-08-indefinite-object-alternative",
	URI:     "https://github.com/clickbytes/golingua/issues/228#issuecomment-6356966127",
	Version: "2026-10-23-g10-08", License: "internal", Author: "Codex by owner delegation",
}

var englishDeterminerAlternativeConstructionSource = provenance.SourceRef{
	ID:      "golingua-g10-10-english-determiner-alternatives",
	URI:     "https://github.com/clickbytes/golingua/issues/233#issuecomment-6357295748",
	Version: "2026-10-23-g10-10", License: "internal", Author: "Codex by owner delegation",
}

var englishIndefiniteAllomorphConstructionSource = provenance.SourceRef{
	ID:      "golingua-g10-12-english-indefinite-allomorphs",
	URI:     "https://github.com/clickbytes/golingua/issues/237#issuecomment-6357579336",
	Version: "2026-10-23-g10-12", License: "internal", Author: "Codex by owner delegation",
}

type frameDeclaration struct {
	id, sense, family string
	roles             []meaning.Role
	arguments         []valency.ArgumentSlot
}

type constructionBindingDeclaration struct {
	sense, frame, family, variant string
}

var englishConstructionBindingDeclarations = []constructionBindingDeclaration{
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_object"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_a_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_a"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_an"},
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_mass_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_object"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_a_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_a"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_object_an"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo.indefinite_subject_an_mass_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.mass_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_mass_object"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_object_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_a_object_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an_object_a"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an_object_an"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive_past", "en.transitive_past.svo.indefinite_subject_an_mass_object"},
	{"en.read.text", "en.read.text.transitive", "en.transitive", "en.transitive.svo"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_target"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_target"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_target_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_a_target_an"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_an_target_a"},
	{"en.look.directed", "en.look.directed.at", "en.oblique", "en.oblique.svo.indefinite_subject_an_target_an"},
}

var englishFrames = []frameDeclaration{
	{"en.see.visual.transitive", "en.see.visual", "en.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.look.directed.at", "en.look.directed", "en.oblique", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "target", Required: true, GrammaticalFunction: "oblique", Adposition: "at"},
	}},
	{"en.hear.auditory.transitive", "en.hear.auditory", "en.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.get.receive.transitive", "en.get.receive", "en.transitive_past", []meaning.Role{"recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.read.text.transitive", "en.read.text", "en.transitive", []meaning.Role{"reader", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "reader", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object"},
	}},
}

// CompileValency constructs independently authored English argument frames
// and the construction bindings approved in the delegated review. The routes
// alone do not activate sentence realization.
func CompileValency() (*valency.Index, error) {
	if err := englishValencySource.Validate(); err != nil {
		return nil, fmt.Errorf("English valency source: %w", err)
	}
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	senses := make(map[string]struct{}, len(document.Senses))
	for _, sense := range document.Senses {
		senses[sense.ID.String()] = struct{}{}
	}
	frames := make([]valency.ValencyFrame, 0, len(englishFrames))
	for _, item := range englishFrames {
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, err
		}
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("English frame %s: %w", item.id, err)
		}
		if _, ok := senses[senseID.String()]; !ok {
			return nil, fmt.Errorf("English frame %s has no lexical sense", item.id)
		}
		frame := valency.ValencyFrame{
			ID: valency.FrameID(item.id), SenseID: senseID,
			SemanticRoles:    append([]meaning.Role(nil), item.roles...),
			Arguments:        append([]valency.ArgumentSlot(nil), item.arguments...),
			AllowedFamilyIDs: []ids.ConstructionID{familyID}, Provenance: englishValencySource,
		}
		frames = append(frames, frame)
	}
	bindings := make([]valency.ConstructionBinding, 0, len(englishConstructionBindingDeclarations))
	for _, item := range englishConstructionBindingDeclarations {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("English construction binding sense: %w", err)
		}
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, fmt.Errorf("English construction binding family: %w", err)
		}
		variantID, err := ids.NewConstructionID(item.variant)
		if err != nil {
			return nil, fmt.Errorf("English construction binding variant: %w", err)
		}
		bindings = append(bindings, valency.ConstructionBinding{
			SenseID: senseID, ValencyFrameID: valency.FrameID(item.frame),
			FamilyID: familyID, VariantID: variantID, Approved: true,
			Provenance: englishConstructionBindingProvenance(item),
		})
	}
	return valency.BuildIndexWithCandidates(frames, bindings, nil)
}

func englishConstructionBindingProvenance(item constructionBindingDeclaration) provenance.SourceRef {
	variant := item.variant
	if suffix, isPastFamily := strings.CutPrefix(variant, "en.transitive_past."); isPastFamily {
		variant = "en.transitive." + suffix
	}
	if variant == "en.transitive.svo.mass_object" {
		return englishOwnerMassObjectConstructionSource
	}
	if variant == "en.transitive.svo.indefinite_object" && item.sense == "en.get.receive" {
		return englishIndefiniteObjectConstructionSource
	}
	if englishIndefiniteAllomorphVariant(variant) {
		return englishIndefiniteAllomorphConstructionSource
	}
	if englishDeterminerAlternativeVariant(variant) {
		return englishDeterminerAlternativeConstructionSource
	}
	return englishBindingProvenance(item.sense)
}

func englishIndefiniteAllomorphVariant(variant string) bool {
	switch variant {
	case "en.transitive.svo.indefinite_object_an",
		"en.transitive.svo.indefinite_subject_an",
		"en.transitive.svo.indefinite_subject_a_object_an",
		"en.transitive.svo.indefinite_subject_an_object_a",
		"en.transitive.svo.indefinite_subject_an_object_an",
		"en.transitive.svo.indefinite_subject_an_mass_object",
		"en.oblique.svo.indefinite_target_an",
		"en.oblique.svo.indefinite_subject_an",
		"en.oblique.svo.indefinite_subject_a_target_an",
		"en.oblique.svo.indefinite_subject_an_target_a",
		"en.oblique.svo.indefinite_subject_an_target_an":
		return true
	default:
		return false
	}
}

func englishDeterminerAlternativeVariant(variant string) bool {
	switch variant {
	case "en.transitive.svo.indefinite_object",
		"en.transitive.svo.indefinite_subject",
		"en.transitive.svo.indefinite_subject_object",
		"en.transitive.svo.indefinite_subject_mass_object",
		"en.oblique.svo.indefinite_target",
		"en.oblique.svo.indefinite_subject",
		"en.oblique.svo.indefinite_subject_target":
		return true
	default:
		return false
	}
}

func englishBindingProvenance(sense string) provenance.SourceRef {
	if sense == "en.look.directed" {
		return englishLookingConstructionReviewSource
	}
	return englishConstructionReviewSource
}
