// Package catalog contains project-authored, language-neutral CORE declarations.
// It has no dependency on Polish or English lexical packages.
package catalog

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

// TargetConcepts is an editorial size goal, not a runtime or quality gate.
const TargetConcepts = 1000

// Entry retains the source of one authored definition beside its typed CORE
// data. A source reference records provenance; it does not approve a lexical
// binding or claim that a sentence can be translated.
type Entry struct {
	Definition core.ConceptDefinition
	Source     provenance.SourceRef
}

type declaration struct {
	id, parent, description string
	kind                    core.ConceptKind
	roles                   []core.RoleType
}

var authoredSource = provenance.SourceRef{
	ID: "golingua-core-authored", Path: "internal/catalog/concepts.go",
	Version: "2026-10-21-initial", License: "unknown", Author: "golingua contributors",
}

// The labels in issue #59 are candidates. Only the distinct meanings below
// have been written as CORE definitions. Grammatical operators live elsewhere.
var declarations = []declaration{
	{"entity", "", "A continuant that can be identified independently of a particular event.", core.KindEntity, nil},
	{"person", "entity", "A human individual, irrespective of age, social role, or name.", core.KindEntity, nil},
	{"child", "person", "A person in an early developmental stage; parenthood is not implied.", core.KindEntity, nil},
	{"adult", "person", "A person past the child stage; no legal age threshold is asserted.", core.KindEntity, nil},
	{"teacher", "person", "A person acting in a role of guiding another person's learning.", core.KindEntity, nil},
	{"student", "person", "A person acting in a role of learning from instruction or study.", core.KindEntity, nil},
	{"doctor", "person", "A person acting in a professional role of providing medical care.", core.KindEntity, nil},
	{"patient", "person", "A person receiving or seeking medical care.", core.KindEntity, nil},
	{"animal", "entity", "A living nonhuman organism capable of animal life processes.", core.KindEntity, nil},
	{"dog", "animal", "An individual domestic dog, regardless of breed or role.", core.KindEntity, nil},
	{"cat", "animal", "An individual domestic cat, regardless of breed or role.", core.KindEntity, nil},
	{"plant", "entity", "A living organism of the plant kingdom.", core.KindEntity, nil},
	{"tree", "plant", "A woody perennial plant with a persistent main stem.", core.KindEntity, nil},
	{"artifact", "entity", "A physical object intentionally made or modified by an agent.", core.KindEntity, nil},
	{"book", "artifact", "A bound or otherwise ordered work intended for reading, in physical form.", core.KindEntity, nil},
	{"door", "artifact", "A movable barrier used to open or close an entrance.", core.KindEntity, nil},
	{"vehicle", "artifact", "An artifact designed to carry people or goods between places.", core.KindEntity, nil},
	{"location", "entity", "A spatial region that can serve as the site of an event or entity.", core.KindEntity, nil},
	{"house", "artifact", "A building designed as a dwelling; occupancy is not asserted.", core.KindEntity, nil},
	{"school", "artifact", "A facility designed for organized instruction; a school organization is a distinct meaning.", core.KindEntity, nil},
	{"body_part", "entity", "A physical component of an organism's body.", core.KindEntity, nil},
	{"hand", "body_part", "The terminal grasping part of a human arm.", core.KindEntity, nil},
	{"eye", "body_part", "An organ used for visual perception.", core.KindEntity, nil},
	{"water", "entity", "The substance water, without specifying a quantity or container.", core.KindEntity, nil},
	{"food", "entity", "A substance or prepared item intended for nourishment.", core.KindEntity, nil},

	{"event", "", "An occurrence with participants or a change that can be located in time.", core.KindEvent, nil},
	{"perception", "event", "An event in which a perceiver acquires sensory information about a stimulus.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}}},
	{"seeing", "perception", "Visual perception of a stimulus; directed attention is not required.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}}},
	{"looking", "perception", "Intentional direction of visual attention toward a target; successful seeing is not implied.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "target", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"hearing", "perception", "Auditory perception of a stimulus without requiring deliberate attention.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}}},
	{"movement", "event", "A change in an entity's spatial position over time.", core.KindEvent, []core.RoleType{{Name: "mover", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"arrival", "movement", "Movement ending at a destination.", core.KindEvent, []core.RoleType{{Name: "mover", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "destination", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"departure", "movement", "Movement beginning at a source location.", core.KindEvent, []core.RoleType{{Name: "mover", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "source", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"transfer", "event", "A change in control or possession of a theme between participants.", core.KindEvent, []core.RoleType{{Name: "source", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "recipient", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"giving", "transfer", "A transfer initiated by a source toward a recipient.", core.KindEvent, []core.RoleType{{Name: "giver", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "recipient", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"receiving", "transfer", "A transfer considered from the participant gaining the theme.", core.KindEvent, []core.RoleType{{Name: "recipient", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"communication", "event", "An event in which a sender conveys communicative content to an addressee.", core.KindEvent, []core.RoleType{{Name: "sender", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"asking", "communication", "A communicative act requesting information from an addressee.", core.KindEvent, []core.RoleType{{Name: "sender", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"answering", "communication", "A communicative act supplying information in response to a prior request.", core.KindEvent, []core.RoleType{{Name: "sender", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"reading", "event", "An event of interpreting written or otherwise recorded symbolic content.", core.KindEvent, []core.RoleType{{Name: "reader", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "material", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"learning", "event", "An event in which a participant acquires knowledge or skill.", core.KindEvent, []core.RoleType{{Name: "learner", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"teaching", "event", "An event in which an instructor intentionally supports another's learning.", core.KindEvent, []core.RoleType{{Name: "instructor", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "learner", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"eating", "event", "An event of taking food into a body for nourishment.", core.KindEvent, []core.RoleType{{Name: "eater", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "food", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"sleeping", "event", "A physiological period of sleep undergone by a living participant.", core.KindEvent, []core.RoleType{{Name: "sleeper", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},

	{"property", "", "A quality that can be attributed to a bearer without identifying a particular bearer.", core.KindProperty, nil},
	{"size", "property", "The extent of a bearer along one or more spatial dimensions.", core.KindProperty, nil},
	{"temperature", "property", "The thermal state of a bearer on a temperature scale.", core.KindProperty, nil},
	{"color", "property", "A visual quality of a bearer associated with perceived light.", core.KindProperty, nil},
	{"speed", "property", "The rate of a participant's movement relative to time.", core.KindProperty, nil},
}

// Entries returns freshly constructed declarations. Callers cannot mutate the
// catalog's package-level source table through a returned value.
func Entries() ([]Entry, error) {
	if err := authoredSource.Validate(); err != nil {
		return nil, fmt.Errorf("catalog source: %w", err)
	}
	entries := make([]Entry, 0, len(declarations))
	for _, item := range declarations {
		id, err := ids.NewConceptID(item.id)
		if err != nil {
			return nil, fmt.Errorf("catalog %s: %w", item.id, err)
		}
		definition := core.ConceptDefinition{ID: id, Kind: item.kind, Definition: item.description}
		if item.parent != "" {
			parent, err := ids.NewConceptID(item.parent)
			if err != nil {
				return nil, fmt.Errorf("catalog %s parent: %w", item.id, err)
			}
			definition.Parent = &parent
		}
		definition.RoleTypes = make([]core.RoleType, len(item.roles))
		for index, role := range item.roles {
			definition.RoleTypes[index] = core.RoleType{Name: role.Name, AllowedKinds: append([]core.ConceptKind(nil), role.AllowedKinds...)}
		}
		entries = append(entries, Entry{Definition: definition, Source: authoredSource})
	}
	return entries, nil
}

// Compile validates every declaration and returns a private, immutable CORE
// snapshot. Language-specific bindings are intentionally separate.
func Compile() (*core.ConceptGraph, error) {
	entries, err := Entries()
	if err != nil {
		return nil, err
	}
	definitions := make([]core.ConceptDefinition, len(entries))
	for index, entry := range entries {
		definitions[index] = entry.Definition
	}
	return core.CompileConceptGraph(definitions, nil)
}
