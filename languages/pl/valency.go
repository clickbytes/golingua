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

var polishConstructionReviewSource = provenance.SourceRef{
	ID:      "golingua-delegated-construction-review",
	URI:     "https://github.com/clickbytes/golingua/issues/63#issuecomment-6351925853",
	Version: "2026-10-22", License: "unknown", Author: "Codex by owner delegation",
}

var polishNegationConstructionSource = provenance.SourceRef{
	ID: "golingua-g10-16-negated-seeing", URI: "https://github.com/clickbytes/golingua/issues/245#issuecomment-6358351181",
	Version: "2026-10-23", License: "internal", Author: "Codex by owner delegation",
}

type frameDeclaration struct {
	id, sense, family string
	roles             []meaning.Role
	arguments         []valency.ArgumentSlot
}

type constructionBindingDeclaration struct {
	sense, frame, family, variant string
}

var polishConstructionBindingDeclarations = []constructionBindingDeclaration{
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.transitive_negative", "pl.transitive_negative.svo"},
	{"pl.widziec.visual", "pl.widziec.visual.transitive", "pl.transitive", "pl.transitive.svo"},
	{"pl.slyszec.auditory", "pl.slyszec.auditory.transitive", "pl.transitive", "pl.transitive.svo"},
	{"pl.dostac.receive", "pl.dostac.receive.transitive", "pl.transitive", "pl.transitive.svo"},
	{"pl.patrzec.directed", "pl.patrzec.directed.na", "pl.oblique", "pl.oblique.svo"},
}

// Families are explicit per frame, never silently assumed transitive.
// They are structural candidates, not approved construction bindings.
var polishFrames = []frameDeclaration{
	{"pl.widziec.visual.transitive", "pl.widziec.visual", "pl.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.patrzec.directed.na", "pl.patrzec.directed", "pl.oblique", []meaning.Role{"agent", "target"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "agent", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "oblique", Role: "target", Required: true, GrammaticalFunction: "oblique", Case: "accusative", Adposition: "na"},
	}},
	{"pl.slyszec.auditory.transitive", "pl.slyszec.auditory", "pl.transitive", []meaning.Role{"experiencer", "stimulus"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "experiencer", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "stimulus", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dostac.receive.transitive", "pl.dostac.receive", "pl.transitive", []meaning.Role{"recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "recipient", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.dac.transfer.ditransitive", "pl.dac.transfer", "pl.ditransitive", []meaning.Role{"giver", "recipient", "theme"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "giver", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "recipient", Role: "recipient", Required: true, GrammaticalFunction: "indirect_object", Case: "dative"},
		{ID: "theme", Role: "theme", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.pytac.request_info.frame", "pl.pytac.request_info", "pl.transitive", []meaning.Role{"sender", "addressee"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "addressee", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.odpowiadac.answer.frame", "pl.odpowiadac.answer", "pl.transitive", []meaning.Role{"sender", "addressee"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sender", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "addressee", Required: true, GrammaticalFunction: "object", Case: "dative"},
	}},
	{"pl.czytac.interpret.frame", "pl.czytac.interpret", "pl.transitive", []meaning.Role{"reader", "material"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "reader", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "material", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.uczyc.instruct.frame", "pl.uczyc.instruct", "pl.transitive", []meaning.Role{"instructor", "learner"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "instructor", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "learner", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.uczyc_sie.acquire.frame", "pl.uczyc_sie.acquire", "pl.intransitive", []meaning.Role{"learner"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "learner", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
	{"pl.jesc.eat.frame", "pl.jesc.eat", "pl.transitive", []meaning.Role{"eater", "food"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "eater", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
		{ID: "object", Role: "food", Required: true, GrammaticalFunction: "object", Case: "accusative"},
	}},
	{"pl.spac.sleep.frame", "pl.spac.sleep", "pl.intransitive", []meaning.Role{"sleeper"}, []valency.ArgumentSlot{
		{ID: "subject", Role: "sleeper", Required: true, GrammaticalFunction: "subject", Case: "nominative"},
	}},
}

// CompileValency compiles sense-specific Polish frames and the construction
// bindings approved in the delegated review. These routes alone do not license
// sentence transfer or surface realization.
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
	declarations := append(append([]frameDeclaration(nil), polishFrames...), polishMotionFrames...)
	declarations = append(declarations, physicalActionFrames...)
	declarations = append(declarations, activityActionFrames...)
	declarations = append(declarations, healthRoutineActionFrames...)
	declarations = append(declarations, productionActionFrames...)
	declarations = append(declarations, lifeActionFrames...)
	declarations = append(declarations, socialActionFrames...)
	declarations = append(declarations, communicationActionFrames...)
	declarations = append(declarations, communicationMoreFrames...)
	declarations = append(declarations, abstractActionFrames...)
	declarations = append(declarations, remainingActionFrames...)
	declarations = append(declarations, phraseLevelActionFrames...)
	declarations = append(declarations, transferCognitionFrames...)
	declarations = append(declarations, remainingActionFrames2...)
	declarations = append(declarations, exchangeActionFrames...)
	declarations = append(declarations, healthRemainingFrames...)
	declarations = append(declarations, cognitionActionFrames...)
	declarations = append(declarations, multiRoleActionFrames...)
	declarations = append(declarations, communicationLastFrames...)
	declarations = append(declarations, outcomeActionFrames...)
	frames := make([]valency.ValencyFrame, 0, len(declarations))
	for _, item := range declarations {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("Polish frame %s: %w", item.id, err)
		}
		if _, ok := senses[senseID.String()]; !ok {
			return nil, fmt.Errorf("Polish frame %s has no lexical sense", item.id)
		}
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, fmt.Errorf("Polish frame %s family: %w", item.id, err)
		}
		frame := valency.ValencyFrame{
			ID: valency.FrameID(item.id), SenseID: senseID,
			SemanticRoles:    append([]meaning.Role(nil), item.roles...),
			Arguments:        append([]valency.ArgumentSlot(nil), item.arguments...),
			AllowedFamilyIDs: []ids.ConstructionID{familyID}, Provenance: polishValencySource,
		}
		if item.id == "pl.widziec.visual.transitive" {
			negativeFamilyID, err := ids.NewConstructionID("pl.transitive_negative")
			if err != nil {
				return nil, fmt.Errorf("Polish negative frame family: %w", err)
			}
			frame.AllowedFamilyIDs = append(frame.AllowedFamilyIDs, negativeFamilyID)
		}
		frames = append(frames, frame)
	}
	bindings := make([]valency.ConstructionBinding, 0, len(polishConstructionBindingDeclarations))
	for _, item := range polishConstructionBindingDeclarations {
		senseID, err := ids.NewSenseID(item.sense)
		if err != nil {
			return nil, fmt.Errorf("Polish construction binding sense: %w", err)
		}
		familyID, err := ids.NewConstructionID(item.family)
		if err != nil {
			return nil, fmt.Errorf("Polish construction binding family: %w", err)
		}
		variantID, err := ids.NewConstructionID(item.variant)
		if err != nil {
			return nil, fmt.Errorf("Polish construction binding variant: %w", err)
		}
		bindings = append(bindings, valency.ConstructionBinding{
			SenseID: senseID, ValencyFrameID: valency.FrameID(item.frame),
			FamilyID: familyID, VariantID: variantID, Approved: true,
			Provenance: polishConstructionBindingProvenance(item),
		})
	}
	return valency.BuildIndexWithCandidates(frames, bindings, nil)
}

func polishConstructionBindingProvenance(item constructionBindingDeclaration) provenance.SourceRef {
	if item.family == "pl.transitive_negative" {
		return polishNegationConstructionSource
	}
	return polishConstructionReviewSource
}
