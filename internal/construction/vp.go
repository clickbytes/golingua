package construction

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/valency"
)

// CompileVP requires the caller to supply the sense-specific frame set. It
// rejects an omitted required argument and never derives roles from word order.
func CompileVP(plan PhrasePlan, phraseProfile PhraseProfile, profile *features.LanguageFeatureProfile, frames map[valency.FrameID]valency.ValencyFrame) (CompiledPhrasePlan, error) {
	if plan.Kind != PhraseVP {
		return CompiledPhrasePlan{}, invalid("PHRASE_KIND_MISMATCH", "kind", "CompileVP requires a VP plan")
	}
	if err := ValidatePhraseProfile(phraseProfile, profile); err != nil {
		return CompiledPhrasePlan{}, err
	}
	return compilePhraseNode(plan, phraseProfile, profile, frames)
}
