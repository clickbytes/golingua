package construction

import (
	"sort"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
)

type MWEComponentKind string

const (
	MWEFixed     MWEComponentKind = "fixed"
	MWEInflected MWEComponentKind = "inflected"
	MWEInput     MWEComponentKind = "input"
)

type MWEComponent struct {
	ID            string                 `json:"id"`
	Position      int                    `json:"position"`
	Kind          MWEComponentKind       `json:"kind"`
	LexemeID      ids.LexemeID           `json:"lexeme_id,omitempty"`
	InputPhraseID PhraseID               `json:"input_phrase_id,omitempty"`
	Features      features.FeatureBundle `json:"features,omitempty"`
}

// MWEPlan is tied to one exact lexical sense. Literal composition is modeled
// through ordinary constructions and never used as a fallback for this plan.
type MWEPlan struct {
	ID              string              `json:"id"`
	Language        features.LanguageID `json:"language"`
	LexemeID        ids.LexemeID        `json:"lexeme_id"`
	SenseID         ids.SenseID         `json:"sense_id"`
	HeadComponentID string              `json:"head_component_id"`
	Components      []MWEComponent      `json:"components"`
}

type CompiledMWEPlan struct{ Plan MWEPlan }

func CompileMWE(plan MWEPlan, dataProfile QuantityTimeMWEProfile, featureProfile *features.LanguageFeatureProfile, lexical *lexicon.Index, phrases *PhraseIndex) (CompiledMWEPlan, error) {
	if !validToken(plan.ID) || featureProfile == nil || plan.Language != dataProfile.Language || plan.Language != featureProfile.Language() {
		return CompiledMWEPlan{}, unsupported("MWE_PROFILE_MISMATCH", "language", "MWE requires matching language profiles and a canonical plan ID")
	}
	if err := ValidateQuantityTimeMWEProfile(dataProfile, featureProfile); err != nil {
		return CompiledMWEPlan{}, err
	}
	if lexical == nil {
		return CompiledMWEPlan{}, unknown("MWE_LEXICON_UNAVAILABLE", "lexicon", "MWE compilation requires an explicit lexical snapshot")
	}
	whole, ok := lexical.ByLexemeID(plan.LexemeID)
	if !ok {
		return CompiledMWEPlan{}, unknown("UNKNOWN_MWE_LEXEME", "lexeme_id", "multiword lexeme is absent from lexical snapshot")
	}
	if features.LanguageID(whole.Language) != plan.Language {
		return CompiledMWEPlan{}, unsupported("MWE_LEXEME_LANGUAGE_MISMATCH", "lexeme_id", "MWE lexeme language does not match plan")
	}
	senseFound := false
	for _, sense := range lexical.SensesForLexeme(plan.LexemeID) {
		if sense.ID == plan.SenseID {
			senseFound = true
			break
		}
	}
	if !senseFound {
		return CompiledMWEPlan{}, unknown("UNKNOWN_MWE_SENSE", "sense_id", "MWE sense is not attached to the selected lexeme")
	}
	if len(plan.Components) < 2 || len(plan.Components) > dataProfile.MaxMWEComponents {
		return CompiledMWEPlan{}, invalid("INVALID_MWE_COMPONENT_COUNT", "components", "MWE requires a bounded sequence of at least two components")
	}
	components := append([]MWEComponent(nil), plan.Components...)
	sort.Slice(components, func(i, j int) bool { return components[i].Position < components[j].Position })
	idsSeen := map[string]struct{}{}
	headFound := false
	for i, component := range components {
		if !validToken(component.ID) || component.Position != i {
			return CompiledMWEPlan{}, invalid("INVALID_MWE_COMPONENT_ORDER", "components", "component positions must be contiguous and IDs valid")
		}
		if _, exists := idsSeen[component.ID]; exists {
			return CompiledMWEPlan{}, invalid("DUPLICATE_MWE_COMPONENT_ID", "components", "component ID is duplicated")
		}
		idsSeen[component.ID] = struct{}{}
		if !containsMWEComponentKind(dataProfile.MWEComponentKinds, component.Kind) {
			return CompiledMWEPlan{}, unsupported("MWE_COMPONENT_KIND_UNSUPPORTED", "components."+component.ID, "component kind is not declared by the language profile")
		}
		switch component.Kind {
		case MWEFixed:
			if component.LexemeID.String() == "" || component.LexemeID == plan.LexemeID || component.InputPhraseID != "" || !emptyFeatureBundle(component.Features) {
				return CompiledMWEPlan{}, invalid("INVALID_FIXED_MWE_COMPONENT", "components."+component.ID, "fixed component requires a lexical reference only")
			}
			if err := validateMWEComponentLexeme(component.LexemeID, plan.Language, lexical); err != nil {
				return CompiledMWEPlan{}, err
			}
		case MWEInflected:
			if component.LexemeID.String() == "" || component.LexemeID == plan.LexemeID || component.InputPhraseID != "" || len(component.Features.Values) == 0 {
				return CompiledMWEPlan{}, invalid("INVALID_INFLECTED_MWE_COMPONENT", "components."+component.ID, "inflected component requires a lexeme and explicit feature bundle")
			}
			if err := validateMWEComponentLexeme(component.LexemeID, plan.Language, lexical); err != nil {
				return CompiledMWEPlan{}, err
			}
			if component.Features.Language != plan.Language || component.Features.ProfileVersion != featureProfile.Version() {
				return CompiledMWEPlan{}, unsupported("MWE_FEATURE_PROFILE_MISMATCH", "components."+component.ID, "inflection bundle requires the matching language profile")
			}
			if err := features.ValidateBundle(featureProfile, component.Features); err != nil {
				return CompiledMWEPlan{}, err
			}
		case MWEInput:
			if component.LexemeID.String() != "" || component.InputPhraseID == "" || !emptyFeatureBundle(component.Features) {
				return CompiledMWEPlan{}, invalid("INVALID_VARIABLE_MWE_COMPONENT", "components."+component.ID, "variable component requires a phrase-plan reference only")
			}
			phrase, found := phrases.ByID(component.InputPhraseID)
			if !found {
				return CompiledMWEPlan{}, unknown("UNKNOWN_MWE_INPUT_PHRASE", "components."+component.ID, "variable component phrase plan is absent")
			}
			if phrase.Plan.Language != plan.Language || phrase.Plan.Kind != PhraseNP {
				return CompiledMWEPlan{}, unsupported("MWE_INPUT_PHRASE_UNSUPPORTED", "components."+component.ID, "variable MWE input must reference a same-language NP plan")
			}
		default:
			return CompiledMWEPlan{}, invalid("INVALID_MWE_COMPONENT_KIND", "components."+component.ID, "component kind is unsupported")
		}
		if component.ID == plan.HeadComponentID {
			headFound = true
			if component.Kind != MWEInflected {
				return CompiledMWEPlan{}, unsupported("MWE_HEAD_NOT_INFLECTABLE", "head_component_id", "MWE head must be an explicitly inflected lexical component")
			}
		}
	}
	if !headFound {
		return CompiledMWEPlan{}, unknown("UNKNOWN_MWE_HEAD_COMPONENT", "head_component_id", "MWE head component is absent")
	}
	plan.Components = cloneMWEComponents(components)
	return CompiledMWEPlan{Plan: plan}, nil
}

func emptyFeatureBundle(bundle features.FeatureBundle) bool {
	return bundle.Language == "" && bundle.ProfileVersion == "" && len(bundle.Values) == 0
}
func cloneMWEComponents(values []MWEComponent) []MWEComponent {
	result := append([]MWEComponent(nil), values...)
	for i := range result {
		result[i].Features.Values = append([]features.FeatureValue(nil), values[i].Features.Values...)
		for j := range result[i].Features.Values {
			if values[i].Features.Values[j].Variable != nil {
				variable := *values[i].Features.Values[j].Variable
				result[i].Features.Values[j].Variable = &variable
			}
		}
	}
	return result
}

func validateMWEComponentLexeme(id ids.LexemeID, language features.LanguageID, lexical *lexicon.Index) error {
	lexeme, ok := lexical.ByLexemeID(id)
	if !ok {
		return unknown("UNKNOWN_MWE_COMPONENT_LEXEME", "components.lexeme_id", "component lexeme is absent")
	}
	if features.LanguageID(lexeme.Language) != language {
		return unsupported("MWE_COMPONENT_LANGUAGE_MISMATCH", "components.lexeme_id", "component lexeme language does not match MWE")
	}
	return nil
}
func containsMWEComponentKind(values []MWEComponentKind, value MWEComponentKind) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
