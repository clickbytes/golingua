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
	Version: "2026-10-21-growth-01", License: "unknown", Author: "golingua contributors",
}

// The labels in issue #59 are candidates. Only the distinct meanings below
// have been written as CORE definitions. Grammatical operators live elsewhere.
var declarations = []declaration{
	{"entity", "", "A continuant that can be identified independently of a particular event.", core.KindEntity, nil},
	{"abstract", "", "A nonphysical meaning, state, or content considered independently of a particular bearer.", core.KindAbstract, nil},
	{"relation", "", "A typed association that can hold between two or more entities, events, states, or contents.", core.KindRelation, nil},
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
	{"school_organization", "organization", "An organization whose primary purpose is to provide organized instruction.", core.KindEntity, nil},
	{"body_part", "entity", "A physical component of an organism's body.", core.KindEntity, nil},
	{"hand", "body_part", "The terminal grasping part of a human arm.", core.KindEntity, nil},
	{"eye", "body_part", "An organ used for visual perception.", core.KindEntity, nil},
	{"water", "entity", "The substance water, without specifying a quantity or container.", core.KindEntity, nil},
	{"food", "entity", "A substance or prepared item intended for nourishment.", core.KindEntity, nil},
	{"physical_object", "entity", "A discrete material object with spatial boundaries.", core.KindEntity, nil},
	{"substance", "entity", "Material considered by its composition rather than as a bounded object.", core.KindEntity, nil},
	{"parent", "person", "A person who is a mother or father of another person.", core.KindEntity, nil},
	{"caregiver", "person", "A person who provides ongoing care to another person.", core.KindEntity, nil},
	{"worker", "person", "A person who performs work in an employment or service role.", core.KindEntity, nil},
	{"customer", "person", "A person who obtains or seeks goods or services from a provider.", core.KindEntity, nil},
	{"friend", "person", "A person connected to another by a voluntary relationship of mutual personal regard.", core.KindEntity, nil},
	{"people_group", "entity", "A plurality of people considered together as a group.", core.KindEntity, nil},
	{"family", "people_group", "A group of people connected by kinship, partnership, or recognized family ties.", core.KindEntity, nil},
	{"organization", "entity", "A continuing organized group formed to pursue shared purposes.", core.KindEntity, nil},
	{"mammal", "animal", "A vertebrate animal whose females produce milk to nourish their young.", core.KindEntity, nil},
	{"bird", "animal", "A feathered, beaked vertebrate animal.", core.KindEntity, nil},
	{"fish", "animal", "An aquatic vertebrate animal that breathes through gills during at least one life stage.", core.KindEntity, nil},
	{"insect", "animal", "An arthropod with six legs and a body organized into three main regions.", core.KindEntity, nil},
	{"horse", "animal", "An individual domesticated equine of the species commonly used for riding or work.", core.KindEntity, nil},
	{"cow", "animal", "An adult female domestic bovine, especially one kept for milk or breeding.", core.KindEntity, nil},
	{"flower", "plant", "The reproductive structure of a flowering plant.", core.KindEntity, nil},
	{"grass", "plant", "A plant of the grass family, typically with narrow leaves and jointed stems.", core.KindEntity, nil},
	{"leaf", "body_part", "A flattened plant organ that commonly performs photosynthesis.", core.KindEntity, nil},
	{"seed", "entity", "A plant reproductive unit capable of developing into a new plant.", core.KindEntity, nil},
	{"fruit", "entity", "The mature ovary of a flowering plant, often containing seeds.", core.KindEntity, nil},
	{"body", "entity", "The physical structure of a living organism considered as a whole.", core.KindEntity, nil},
	{"head", "body_part", "The body part that contains the brain and the main organs of sight, hearing, smell, and taste.", core.KindEntity, nil},
	{"face", "body_part", "The front part of a head containing the eyes, nose, and mouth.", core.KindEntity, nil},
	{"ear", "body_part", "An organ of hearing and balance, or its externally visible part.", core.KindEntity, nil},
	{"nose", "body_part", "The facial organ used for smelling and for passage of air in breathing.", core.KindEntity, nil},
	{"mouth", "body_part", "The body opening and associated structures used for taking in food and producing speech.", core.KindEntity, nil},
	{"leg", "body_part", "A limb used primarily for standing or locomotion.", core.KindEntity, nil},
	{"heart", "body_part", "A muscular organ that circulates blood through the body.", core.KindEntity, nil},
	{"blood", "entity", "The circulating bodily fluid that transports substances through many animals.", core.KindEntity, nil},
	{"illness", "abstract", "A condition in which normal bodily or mental functioning is impaired.", core.KindAbstract, nil},
	{"injury", "abstract", "Physical damage to an organism's body caused by an external event or force.", core.KindAbstract, nil},
	{"health", "abstract", "The condition of bodily or mental functioning considered in relation to well-being.", core.KindAbstract, nil},
	{"pain", "event", "An unpleasant sensory or emotional experience associated with actual or potential harm.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "source", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"air", "substance", "The mixture of gases surrounding Earth and commonly breathed by organisms.", core.KindEntity, nil},
	{"fire", "event", "A process of combustion that produces heat and often light.", core.KindEvent, []core.RoleType{{Name: "fuel", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "cause", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}}},
	{"earth_material", "substance", "Soil or ground material, distinguished from the planet Earth.", core.KindEntity, nil},
	{"stone", "substance", "A naturally occurring solid piece of rock.", core.KindEntity, nil},
	{"wood", "substance", "The hard fibrous material forming the trunk and branches of trees.", core.KindEntity, nil},
	{"metal", "substance", "A class of material with characteristic metallic properties such as conductivity and malleability.", core.KindEntity, nil},
	{"glass", "substance", "A hard, brittle, usually transparent material made by cooling a fused mixture.", core.KindEntity, nil},
	{"sand", "entity", "Loose granular material composed mainly of small mineral or rock particles.", core.KindEntity, nil},
	{"rain", "event", "Water falling from clouds as liquid precipitation.", core.KindEvent, []core.RoleType{{Name: "precipitation", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"snow", "entity", "Frozen precipitation consisting of ice crystals or their aggregates.", core.KindEntity, nil},
	{"wind", "event", "Air moving relative to the surface of a place.", core.KindEvent, nil},
	{"cloud", "entity", "A visible mass of condensed water droplets or ice crystals suspended in the atmosphere.", core.KindEntity, nil},
	{"river", "location", "A natural flowing body of water occupying a channel.", core.KindEntity, nil},
	{"lake", "location", "A sizable body of inland water surrounded by land.", core.KindEntity, nil},
	{"sea", "location", "A large body of salt water, usually part of an ocean or connected to it.", core.KindEntity, nil},
	{"mountain", "location", "A prominent natural elevation of the Earth's surface.", core.KindEntity, nil},
	{"forest", "location", "A tract of land dominated by trees and associated vegetation.", core.KindEntity, nil},
	{"room", "location", "An enclosed or partly enclosed space within a building.", core.KindEntity, nil},
	{"apartment", "artifact", "A self-contained set of rooms intended as a dwelling within a larger building.", core.KindEntity, nil},
	{"hospital", "artifact", "A facility organized to provide medical diagnosis and treatment.", core.KindEntity, nil},
	{"shop", "artifact", "A place where goods are offered for sale.", core.KindEntity, nil},
	{"road", "location", "A prepared route intended for travel between places.", core.KindEntity, nil},
	{"street", "location", "A public road within a settlement, commonly bordered by buildings.", core.KindEntity, nil},
	{"city", "location", "A large, densely settled human community with organized civic status or functions.", core.KindEntity, nil},
	{"country", "location", "A territory organized as a political unit under a government.", core.KindEntity, nil},
	{"table", "artifact", "A piece of furniture with a raised flat surface supported by a base.", core.KindEntity, nil},
	{"chair", "artifact", "A piece of furniture designed to support one seated person.", core.KindEntity, nil},
	{"window", "artifact", "An opening in a wall or vehicle fitted to admit light or view, commonly with a transparent covering.", core.KindEntity, nil},
	{"bed", "artifact", "A piece of furniture intended for a person to lie on while resting or sleeping.", core.KindEntity, nil},
	{"clothing", "entity", "Items worn on the body for covering, protection, or social presentation.", core.KindEntity, nil},
	{"document", "artifact", "A recorded item of information intended to be read or consulted.", core.KindEntity, nil},
	{"telephone", "artifact", "A device for transmitting speech or other signals over a distance.", core.KindEntity, nil},
	{"computer", "artifact", "A programmable electronic device that processes and stores information.", core.KindEntity, nil},
	{"drink", "entity", "A liquid prepared or regarded for consumption.", core.KindEntity, nil},
	{"bread", "entity", "A food made by baking dough, commonly based on flour and water.", core.KindEntity, nil},
	{"milk", "entity", "A nutrient liquid produced by mammals to feed their young, or a commonly consumed equivalent.", core.KindEntity, nil},
	{"egg", "entity", "An animal reproductive structure containing an embryo or its food supply; commonly used as food.", core.KindEntity, nil},
	{"meat", "entity", "Animal flesh used as food.", core.KindEntity, nil},
	{"vegetable", "entity", "An edible plant or plant part used as food, excluding fruit in the culinary sense.", core.KindEntity, nil},
	{"salt", "substance", "A crystalline mineral substance, especially sodium chloride, used in food and other processes.", core.KindEntity, nil},
	{"money", "entity", "A medium accepted in exchange and used to measure or store economic value.", core.KindEntity, nil},
	{"price", "property", "The amount of money or other value required to obtain a good or service.", core.KindProperty, nil},
	{"tool", "artifact", "An artifact used to perform work or bring about a practical result.", core.KindEntity, nil},
	{"container", "artifact", "An artifact designed to hold or enclose other things.", core.KindEntity, nil},
	{"key", "artifact", "A shaped instrument used to operate a lock.", core.KindEntity, nil},
	{"bag", "artifact", "A flexible container used to carry or store items.", core.KindEntity, nil},
	{"knowledge", "abstract", "Information or understanding available to a person or group through learning or experience.", core.KindAbstract, nil},
	{"belief", "abstract", "A proposition accepted as true by a person, whether or not it is true.", core.KindAbstract, nil},
	{"thought", "abstract", "A mental content or process considered as something entertained by a mind.", core.KindAbstract, nil},
	{"thinking", "event", "A cognitive process in which a person considers, connects, or evaluates mental content.", core.KindEvent, []core.RoleType{{Name: "thinker", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"memory", "abstract", "Retained information or capacity to recall prior experience.", core.KindAbstract, nil},
	{"decision", "event", "An act of selecting a course of action or outcome after consideration.", core.KindEvent, []core.RoleType{{Name: "decider", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "selected_option", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"choice", "event", "An act or result of selecting one option from alternatives.", core.KindEvent, []core.RoleType{{Name: "chooser", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "selected_option", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"position", "property", "The spatial placement of an entity relative to a reference or other entities.", core.KindProperty, nil},
	{"staying", "event", "An entity's continued presence at a location without departure during a relevant interval.", core.KindEvent, []core.RoleType{{Name: "entity", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "location", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"entering", "event", "Movement of an entity from outside a bounded place to its inside.", core.KindEvent, []core.RoleType{{Name: "mover", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "destination", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"exiting", "event", "Movement of an entity from inside a bounded place to its outside.", core.KindEvent, []core.RoleType{{Name: "mover", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "source", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"falling", "event", "Downward movement of an entity under gravity or loss of support.", core.KindEvent, []core.RoleType{{Name: "mover", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"lifting", "event", "Raising an entity to a higher position by applying force.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"carrying", "event", "Transporting an entity while supporting or holding it.", core.KindEvent, []core.RoleType{{Name: "carrier", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"placing", "event", "Causing an entity to occupy a selected position or location.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "destination", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"removing", "event", "Causing an entity to cease being in or on a location or container.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "source", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"turning", "event", "Changing the orientation or direction of an entity or mover.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "affected", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"stopping", "event", "An event that brings an ongoing movement or activity to an end.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "activity", AllowedKinds: []core.ConceptKind{core.KindEvent}}}},
	{"possession", "relation", "A relation in which an entity has control or recognized ownership of another entity.", core.KindRelation, nil},
	{"taking", "event", "An event in which an agent obtains control of a theme, often from another possessor.", core.KindEvent, []core.RoleType{{Name: "recipient", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "source", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"borrowing", "event", "Temporary acquisition of use or possession with an expectation of return.", core.KindEvent, []core.RoleType{{Name: "borrower", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "lender", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"returning", "event", "Transfer or movement of a theme back to a prior possessor or location.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "destination", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"buying", "event", "Acquiring a good or service in exchange for payment.", core.KindEvent, []core.RoleType{{Name: "buyer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "seller", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "good", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "payment", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"selling", "event", "Transferring a good or service in exchange for payment.", core.KindEvent, []core.RoleType{{Name: "seller", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "buyer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "good", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "payment", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"payment", "event", "Transfer of money or another accepted value in settlement of an obligation or exchange.", core.KindEvent, []core.RoleType{{Name: "payer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "payee", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "amount", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"exchange", "event", "A reciprocal transfer in which participants give and receive items or values.", core.KindEvent, []core.RoleType{{Name: "participant", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "given", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"loss", "event", "An event in which an entity ceases to possess or have access to a valued theme.", core.KindEvent, []core.RoleType{{Name: "former_possessor", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"finding", "event", "An event in which an agent discovers or locates something previously unknown or sought.", core.KindEvent, []core.RoleType{{Name: "finder", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "theme", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"touching", "event", "Physical contact between an entity and a surface or other entity.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "target", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"tasting", "event", "Perception of flavor through contact with substances in the mouth.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"odor", "property", "A perceptible quality of a substance or source associated with the olfactory sense.", core.KindProperty, nil},
	{"listening", "event", "Intentional attention to sounds in order to perceive or interpret them.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}}},
	{"remembering", "event", "Retrieval or continued accessibility of information from prior experience.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"forgetting", "event", "Loss of ability to retrieve previously learned or experienced information.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"understanding", "event", "A cognitive state or process in which information or a situation is comprehended.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"speaking", "event", "Producing spoken language to communicate content to another participant.", core.KindEvent, []core.RoleType{{Name: "speaker", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"informing", "event", "Communicating information with the intention that an addressee learn it.", core.KindEvent, []core.RoleType{{Name: "sender", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"helping", "event", "Providing assistance that supports another participant or their activity.", core.KindEvent, []core.RoleType{{Name: "helper", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "beneficiary", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "supported_activity", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"care", "event", "Providing attention or support for the well-being of another entity.", core.KindEvent, []core.RoleType{{Name: "caregiver", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "recipient", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"acquaintance", "relation", "A social relation in which people know one another without implying close friendship.", core.KindRelation, nil},
	{"action", "event", "An intentional event performed by an agent.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"performing", "event", "Carrying out a specified task, instruction, or planned activity.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "task", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"beginning", "event", "The transition into the initial phase of an event or state.", core.KindEvent, []core.RoleType{{Name: "initiated", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"ending", "event", "The transition into the final phase or cessation of an event or state.", core.KindEvent, []core.RoleType{{Name: "terminated", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"succession", "relation", "A temporal relation in which one event or interval follows another.", core.KindRelation, nil},
	{"simultaneity", "relation", "A temporal relation in which events or intervals overlap in time.", core.KindRelation, nil},
	{"satisfaction", "event", "A positive state of contentment resulting from fulfilled needs, expectations, or aims.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "cause", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"request", "event", "A communicative act asking another participant to do or provide something.", core.KindEvent, []core.RoleType{{Name: "requester", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "requested_content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"command", "event", "A directive communicative act presented as an instruction or order.", core.KindEvent, []core.RoleType{{Name: "commander", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "addressee", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "directed_action", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"promise", "event", "A communicative commitment by a speaker to a future action or state.", core.KindEvent, []core.RoleType{{Name: "promiser", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "content", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"agreement", "abstract", "A state of shared acceptance or commitment between participants.", core.KindAbstract, nil},
	{"refusal", "event", "A communicative act rejecting an offer, request, or proposal.", core.KindEvent, []core.RoleType{{Name: "refuser", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "rejected_content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"cooperation", "event", "Activity in which participants coordinate actions toward a shared result.", core.KindEvent, []core.RoleType{{Name: "participant", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "shared_goal", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"meeting", "event", "An event in which people come together for interaction or a shared purpose.", core.KindEvent, []core.RoleType{{Name: "participant", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "purpose", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"conflict", "event", "A situation or event involving incompatible goals, claims, or actions between participants.", core.KindEvent, []core.RoleType{{Name: "participant", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "contested_content", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"attempt", "event", "An intentional effort to perform an action without implying success.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "intended_action", AllowedKinds: []core.ConceptKind{core.KindEvent, core.KindAbstract}}}},
	{"creation", "event", "An event that brings an entity or artifact into existence.", core.KindEvent, []core.RoleType{{Name: "creator", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "created", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"destruction", "event", "An event that causes an entity or artifact to cease to exist or lose its organized form.", core.KindEvent, []core.RoleType{{Name: "cause", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}, {Name: "affected", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"repair", "event", "An action that restores or improves the functioning of a damaged artifact.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "artifact", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"opening", "event", "An event that changes something from closed to open.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "affected", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"closing", "event", "An event that changes something from open to closed.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "affected", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"cutting", "event", "An action that separates material using a sharp implement or force.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "material", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "instrument", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"joining", "event", "An event that connects previously separate entities or parts.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "part", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"separation", "event", "An event that makes connected or grouped entities distinct or apart.", core.KindEvent, []core.RoleType{{Name: "agent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "affected", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"growth", "event", "An increase in size, amount, or development of an entity over time.", core.KindEvent, []core.RoleType{{Name: "growing_entity", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"decrease", "event", "A reduction in size, amount, or degree over time.", core.KindEvent, []core.RoleType{{Name: "decreasing_entity", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	{"change", "event", "An event in which an entity or situation becomes different in some respect.", core.KindEvent, []core.RoleType{{Name: "affected", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"existence", "abstract", "The state or fact that an entity or situation is real or present.", core.KindAbstract, nil},
	{"absence", "abstract", "The state of not being present at a relevant place, time, or set.", core.KindAbstract, nil},
	{"life", "abstract", "The condition distinguishing living organisms from nonliving things.", core.KindAbstract, nil},
	{"death", "event", "The irreversible ending of the life of an organism.", core.KindEvent, nil},
	{"length", "property", "The linear extent of an entity from one end to another.", core.KindProperty, nil},
	{"height", "property", "The vertical extent of an entity relative to a reference point.", core.KindProperty, nil},
	{"weight", "property", "The force exerted on an entity by gravity, or its measured value in common usage.", core.KindProperty, nil},
	{"shape", "property", "The geometric form of an entity's boundary or arrangement.", core.KindProperty, nil},
	{"cleanliness", "property", "The degree to which a thing is free from dirt or unwanted material.", core.KindProperty, nil},
	{"fullness", "property", "The degree to which a container or space holds its capacity.", core.KindProperty, nil},
	{"emptiness", "property", "The state of a container or space holding none or little of what it can contain.", core.KindProperty, nil},
	{"proximity", "property", "The degree to which two entities are near one another in space or another ordered domain.", core.KindProperty, nil},
	{"distance", "property", "The amount of spatial separation between two locations or entities.", core.KindProperty, nil},
	{"similarity", "property", "The degree to which entities share relevant characteristics.", core.KindProperty, nil},
	{"difference", "property", "A way in which entities or situations are not the same.", core.KindProperty, nil},
	{"time_point", "abstract", "A location on a temporal ordering, without inherent duration.", core.KindAbstract, nil},
	{"time_period", "abstract", "An interval with a beginning and an end in a temporal ordering.", core.KindAbstract, nil},
	{"duration", "property", "The amount of time occupied by an event or state.", core.KindProperty, nil},
	{"part_whole", "relation", "A relation in which one entity is a constituent part of another.", core.KindRelation, nil},
	{"membership", "relation", "A relation in which an entity belongs to a group or organized whole.", core.KindRelation, nil},
	{"causation", "relation", "A relation in which one event or state contributes to bringing about another.", core.KindRelation, nil},
	{"effect", "abstract", "An outcome or change that is brought about by a cause.", core.KindAbstract, nil},
	{"purpose", "abstract", "An intended outcome that motivates or guides an action.", core.KindAbstract, nil},
	{"condition", "abstract", "A circumstance whose holding is relevant to another event or state.", core.KindAbstract, nil},
	{"joy", "event", "A positive emotional experience associated with satisfaction or well-being.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"sadness", "event", "An unpleasant emotional experience associated with loss or unmet expectations.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"fear", "event", "An emotional response to a perceived threat or danger.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "threat", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"anger", "event", "An emotional response involving displeasure at a perceived offense or obstruction.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "cause", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"love", "abstract", "A strong enduring affective attachment or care directed toward someone or something.", core.KindAbstract, nil},
	{"liking", "abstract", "A positive personal attitude toward someone or something.", core.KindAbstract, nil},
	{"desire", "event", "A mental state in which an experiencer wants a possible outcome or entity.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "desired", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"hope", "event", "A mental state involving desire for an outcome regarded as possible.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "hoped_for", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"surprise", "event", "An emotional response to an unexpected event or information.", core.KindEvent, []core.RoleType{{Name: "experiencer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "stimulus", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent, core.KindAbstract}}}},
	{"difficulty", "property", "The degree of effort or challenge involved in performing or understanding something.", core.KindProperty, nil},
	{"ease", "property", "The degree to which performing or understanding something requires little effort.", core.KindProperty, nil},
	{"importance", "property", "The degree to which something matters for a purpose, decision, or outcome.", core.KindProperty, nil},
	{"usefulness", "property", "The degree to which something can serve a practical purpose.", core.KindProperty, nil},

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
