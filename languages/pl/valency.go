package pl

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/valency"
)

var polishValencySource = provenance.SourceRef{
	ID: "golingua-pl-valency-authored", Path: "languages/pl/valency.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

type frameDeclaration struct {
	id, sense string
	roles     []meaning.Role
	arguments []valency.ArgumentSlot
}

var polishFrames = []frameDeclaration{
	{"pl.widziec.visual.transitive", "pl.widziec.visual", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.patrzec.directed.na", "pl.patrzec.directed", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "target", Role: "target", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.slyszec.auditory.transitive", "pl.slyszec.auditory", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dostac.receive.transitive", "pl.dostac.receive", []meaning.Role{"recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
}

// CompileValency compiles sense-specific Polish frames. No construction
// binding is activated here; the family reference only names a possible
// shape and does not license transfer or surface realization.
func CompileValency() (*valency.Index, error) {
	if err := polishValencySource.Validate(); err != nil {
		return nil, fmt.Errorf("Polish valency source: %w", err)
	}
	document, err := LexicalDocument()
	if err != nil {
		return nil, err
	}
	senses := make(map[string]struct{}, len(document.Senses))
	for _, sense := range document.Senses {
		senses[sense.ID.String()] = struct{}{}
	}
	familyID, err := ids.NewConstructionID("pl.transitive")
	if err != nil {
		return nil, err
	}
	frames := make([]valency.ValencyFrame, 0, len(polishFrames))
	for _, item := range polishFrames {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("Polish frame %s: %w", item.id, err)
		}
		if _, ok := senses[senseID.String()]; !ok {
			return nil, fmt.Errorf("Polish frame %s has no lexical sense", item.id)
		}
		frame := valency.ValencyFrame{
			ID: valency.FrameID(item.id), SenseID: senseID,
			SemanticRoles:    append([]meaning.Role(nil), item.roles...),
			Arguments:        append([]valency.ArgumentSlot(nil), item.arguments...),
			AllowedFamilyIDs: []ids.ConstructionID{familyID}, Provenance: polishValencySource,
		}
		frames = append(frames, frame)
	}
	return valency.BuildIndex(frames, nil)
}
