package en

import (
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
)

var englishBodySource = provenance.SourceRef{
	ID: "golingua-en-authored", Path: "languages/en/lexicon_body.go",
	Version: "2026-10-22-batch-17", License: "unknown", Author: "golingua contributors",
}

// These pending noun candidates name body regions and structures without morphology.
var englishBodyEntries = []lexicalEntry{
	{"en.body.n", "body", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.body.organism_structure", "the integrated physical structure of an organism, distinct from its individual parts", "body"}}},
	{"en.head.n", "head", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.head.body_region", "the upper or front body region containing the brain and principal sensory organs", "head"}}},
	{"en.face.n", "face", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.face.front_of_head", "the front surface of a human head bearing the eyes, nose, and mouth", "face"}}},
	{"en.hair_body.n", "hair", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hair_body.body_fiber", "filamentous keratin structures growing from the skin of a mammal", "hair_body"}}},
	{"en.ear.n", "ear", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ear.hearing_organ", "an organ system or external structure associated with hearing and balance", "ear"}}},
	{"en.nose.n", "nose", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.nose.facial_structure", "the projecting facial structure housing external passages of breathing and smell", "nose"}}},
	{"en.mouth.n", "mouth", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.mouth.oral_cavity", "the opening and cavity used for eating and speech production", "mouth"}}},
	{"en.lip.n", "lip", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lip.mouth_boundary", "a soft movable boundary of the mouth opening", "lip"}}},
	{"en.tongue.n", "tongue", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tongue.oral_muscle", "a muscular organ in the mouth used in taste, swallowing, and articulation", "tongue"}}},
	{"en.tooth.n", "tooth", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.tooth.jaw_structure", "a hard structure in the jaws used to bite or grind food", "tooth"}}},
	{"en.neck.n", "neck", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.neck.connecting_region", "the body region connecting the head to the torso", "neck"}}},
	{"en.shoulder.n", "shoulder", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.shoulder.limb_joint_region", "the body region joining an upper limb to the torso", "shoulder"}}},
	{"en.arm.n", "arm", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.arm.upper_limb", "a human upper limb between shoulder and hand, including upper and lower segments", "arm"}}},
	{"en.elbow.n", "elbow", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.elbow.arm_joint", "the joint region between upper arm and forearm", "elbow"}}},
	{"en.finger.n", "finger", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.finger.hand_digit", "one of the articulated digits of a hand", "finger"}}},
	{"en.thumb.n", "thumb", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.thumb.opposable_digit", "the first opposable digit of the human hand", "thumb"}}},
	{"en.chest.n", "chest", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.chest.upper_torso", "the front and sides of the torso enclosing heart and lungs", "chest"}}},
	{"en.back_body.n", "back", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.back_body.rear_torso", "the rear region of the torso, not a direction in space", "back_body"}}},
	{"en.abdomen.n", "abdomen", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.abdomen.torso_region", "the torso region between the chest and pelvis containing digestive organs", "abdomen"}}},
	{"en.hip.n", "hip", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.hip.pelvic_region", "the lateral pelvic region joining the torso and upper leg", "hip"}}},
	{"en.leg.n", "leg", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.leg.lower_limb", "a lower limb used for standing and locomotion", "leg"}}},
	{"en.knee.n", "knee", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.knee.leg_joint", "the joint region linking thigh and lower leg", "knee"}}},
	{"en.ankle.n", "ankle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.ankle.foot_joint", "the joint region connecting lower leg to foot", "ankle"}}},
	{"en.foot.n", "foot", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.foot.weight_bearing_terminal_limb", "the terminal weight-bearing part of a leg", "foot"}}},
	{"en.toe.n", "toe", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.toe.foot_digit", "one articulated digit of a foot", "toe"}}},
	{"en.skin.n", "skin", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.skin.outer_body_tissue", "the outer protective tissue covering the body", "skin"}}},
	{"en.muscle.n", "muscle", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.muscle.contractile_tissue", "contractile body tissue capable of generating force and movement", "muscle"}}},
	{"en.bone.n", "bone", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bone.skeletal_organ", "a rigid mineralized organ forming part of the skeleton", "bone"}}},
	{"en.joint_body.n", "joint", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.joint_body.bone_connection", "a structural connection between bones allowing support or movement", "joint_body"}}},
	{"en.heart.n", "heart", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.heart.circulatory_organ", "a muscular organ pumping blood through the circulatory system", "heart"}}},
	{"en.lung.n", "lung", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.lung.respiratory_organ", "an organ responsible for exchanging respiratory gases with the environment", "lung"}}},
	{"en.brain.n", "brain", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.brain.central_nervous_organ", "the central nervous organ coordinating cognition and body functions", "brain"}}},
	{"en.stomach_organ.n", "stomach", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.stomach_organ.digestive_organ", "an organ receiving and mechanically or chemically processing swallowed food", "stomach_organ"}}},
	{"en.intestine.n", "intestine", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.intestine.digestive_tube", "a tubular digestive organ absorbing nutrients or processing waste", "intestine"}}},
	{"en.liver.n", "liver", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.liver.metabolic_organ", "a large abdominal organ with metabolic and detoxification functions", "liver"}}},
	{"en.kidney.n", "kidney", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.kidney.filtering_organ", "an organ filtering blood and regulating fluid and solute balance", "kidney"}}},
	{"en.bladder.n", "bladder", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.bladder.urinary_storage_organ", "an organ temporarily storing urine before excretion", "bladder"}}},
	{"en.blood.n", "blood", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.blood.circulating_fluid", "the circulating body fluid transporting gases, nutrients, and other substances", "blood"}}},
	{"en.vein.n", "vein", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.vein.vessel_to_heart", "a blood vessel carrying blood toward the heart", "vein"}}},
	{"en.artery.n", "artery", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.artery.vessel_from_heart", "a blood vessel carrying blood away from the heart", "artery"}}},
	{"en.nerve.n", "nerve", lexicon.PartOfSpeech("noun"), []lexicalMeaning{{"en.nerve.peripheral_signal_bundle", "a bundle of fibers transmitting signals in the peripheral nervous system", "nerve"}}},
}
