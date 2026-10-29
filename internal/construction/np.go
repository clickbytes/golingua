package construction

import "github.com/clickbytes/golingua/internal/features"

// CompileNP checks the shared NP structure and any explicitly supplied
// feature/agreement constraints. It does not choose determiners or inflect.
func CompileNP(plan PhrasePlan, phraseProfile PhraseProfile, profile *features.LanguageFeatureProfile) (CompiledPhrasePlan, error) {
	if plan.Kind != PhraseNP {
		return CompiledPhrasePlan{}, invalid("PHRASE_KIND_MISMATCH", "kind", "CompileNP requires an NP plan")
	}
	if err := ValidatePhraseProfile(phraseProfile, profile); err != nil {
		return CompiledPhrasePlan{}, err
	}
	return compilePhraseNode(plan, phraseProfile, profile, nil)
}

// CompileNPPartial checks the complete NP structure and every agreement whose
// inputs are known. Unknown or absent agreement evidence is retained as a
// typed partial plan; known conflicting values still return an error.
func CompileNPPartial(plan PhrasePlan, phraseProfile PhraseProfile, profile *features.LanguageFeatureProfile) (CompiledPhrasePlan, error) {
	if plan.Kind != PhraseNP {
		return CompiledPhrasePlan{}, invalid("PHRASE_KIND_MISMATCH", "kind", "CompileNPPartial requires an NP plan")
	}
	if err := ValidatePhraseProfile(phraseProfile, profile); err != nil {
		return CompiledPhrasePlan{}, err
	}
	return compilePhraseNodeWithPartial(plan, phraseProfile, profile, nil, true)
}
