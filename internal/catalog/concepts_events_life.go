package catalog

import "github.com/clickbytes/golingua/internal/core"

// eventsLifeDeclarations returns the reviewed, typed concepts in editorial order.
func eventsLifeDeclarations() []declaration {
	return []declaration{
		{"breathing", "event", "Repeatedly moving respiratory gases into and out of a living organism.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"inhaling", "breathing", "Drawing air or another respiratory gas into the respiratory system.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"exhaling", "breathing", "Expelling respiratory gas from the respiratory system.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"coughing", "event", "Forcefully expelling air to clear or respond to irritation in the airways.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"sneezing", "event", "Involuntarily expelling air through the nose and mouth in a sneeze reflex.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"swallowing", "event", "Moving a substance from the mouth through the throat toward the stomach.", core.KindEvent, []core.RoleType{{Name: "swallower", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "substance", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"drinking", "event", "Taking liquid into the mouth and swallowing it.", core.KindEvent, []core.RoleType{{Name: "drinker", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "liquid", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"chewing", "event", "Mechanically crushing food with teeth before swallowing.", core.KindEvent, []core.RoleType{{Name: "chewer", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "food", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"digesting", "event", "Breaking down consumed material and absorbing nutrients through bodily processes.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "food", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"sweating", "event", "Releasing fluid through sweat glands in the skin.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"bleeding", "event", "Blood escaping from a damaged blood vessel or body surface.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"healing", "event", "Biological recovery of damaged tissue toward its ordinary function.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "tissue", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"becoming_ill", "event", "Transition from an ordinary healthy condition to a state of disease.", core.KindEvent, []core.RoleType{{Name: "patient", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"falling_asleep", "event", "Transition from wakefulness into physiological sleep.", core.KindEvent, []core.RoleType{{Name: "sleeper", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"waking_up", "event", "Transition from physiological sleep into wakefulness.", core.KindEvent, []core.RoleType{{Name: "person", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"exercising_body", "intentional_action", "Performing physical activity for fitness, training, or health.", core.KindEvent, []core.RoleType{{Name: "participant", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"aging", "event", "Undergoing biological changes associated with increasing age over time.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"dying", "event", "The irreversible end of a living organism's integrated life processes.", core.KindEvent, []core.RoleType{{Name: "organism", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"giving_birth", "event", "Delivering offspring from a pregnant organism's body.", core.KindEvent, []core.RoleType{{Name: "parent", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "offspring", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"injuring_person", "event", "Causing bodily tissue damage to a living participant.", core.KindEvent, []core.RoleType{{Name: "cause", AllowedKinds: []core.ConceptKind{core.KindEntity, core.KindEvent}}, {Name: "patient", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"treating_patient", "intentional_action", "Providing a therapeutic intervention intended to improve a patient's condition.", core.KindEvent, []core.RoleType{{Name: "care_provider", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "patient", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"diagnosing_condition", "intentional_action", "Determining the likely identity or nature of a medical condition from evidence.", core.KindEvent, []core.RoleType{{Name: "clinician", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "patient", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"taking_medicine", "intentional_action", "Administering a medicinal substance to oneself or receiving its administration.", core.KindEvent, []core.RoleType{{Name: "patient", AllowedKinds: []core.ConceptKind{core.KindEntity}}, {Name: "medicine", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
		{"recovering_from_illness", "healing", "Returning toward an ordinary state of health following illness.", core.KindEvent, []core.RoleType{{Name: "patient", AllowedKinds: []core.ConceptKind{core.KindEntity}}}},
	}
}
