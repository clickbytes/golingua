package construction

import "github.com/clickbytes/golingua/internal/features"

// CompilePP requires an explicit nested NP complement. Governed case, when
// present, is validated against the language profile rather than guessed.
func CompilePP(plan PhrasePlan, phraseProfile PhraseProfile, profile *features.LanguageFeatureProfile) (CompiledPhrasePlan, error) {
	if plan.Kind != PhrasePP {
		return CompiledPhrasePlan{}, invalid("PHRASE_KIND_MISMATCH", "kind", "CompilePP requires a PP plan")
	}
	if err := ValidatePhraseProfile(phraseProfile, profile); err != nil {
		return CompiledPhrasePlan{}, err
	}
	return compilePhraseNode(plan, phraseProfile, profile, nil)
}
