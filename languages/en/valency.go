package en

import (
	"fmt"

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

type frameDeclaration struct {
	id, sense string
	roles     []meaning.Role
	arguments []valency.ArgumentSlot
}

type constructionBindingDeclaration struct {
	sense, frame, family, variant string
}

var englishConstructionBindingDeclarations = []constructionBindingDeclaration{
	{"en.see.visual", "en.see.visual.transitive", "en.transitive", "en.transitive.svo"},
	{"en.hear.auditory", "en.hear.auditory.transitive", "en.transitive", "en.transitive.svo"},
	{"en.get.receive", "en.get.receive.transitive", "en.transitive", "en.transitive.svo"},
	{"en.read.text", "en.read.text.transitive", "en.transitive", "en.transitive.svo"},
}

var englishFrames = []frameDeclaration{
	{"en.see.visual.transitive", "en.see.visual", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.look.directed.at", "en.look.directed", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject"},
		{ID: "target", Role: "target", Required: true, GrammaticalFunction: "oblique", Adposition: "at"},
	}},
	{"en.hear.auditory.transitive", "en.hear.auditory", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.get.receive.transitive", "en.get.receive", []meaning.Role{"recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object"},
	}},
	{"en.read.text.transitive", "en.read.text", []meaning.Role{"reader", "material"}, []valency.ArgumentSlot{
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
	familyID, err := ids.NewConstructionID("en.transitive")
	if err != nil {
		return nil, err
	}
	frames := make([]valency.ValencyFrame, 0, len(englishFrames))
	for _, item := range englishFrames {
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
			Provenance: englishConstructionReviewSource,
		})
	}
	return valency.BuildIndexWithCandidates(frames, bindings, nil)
}
