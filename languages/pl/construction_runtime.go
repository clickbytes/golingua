package pl

import (
	"embed"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/valency"
)

//go:embed constructions/transitive_personal.json constructions/transitive_personal_explicit.json constructions/transitive.json constructions/transitive_negative.json constructions/oblique.json constructions/with_oblique.json constructions/oblique_negative.json constructions/intransitive.json constructions/ditransitive.json constructions/ditransitive_negative.json constructions/double_np.json constructions/transitive_oblique.json constructions/transitive_instrumental.json constructions/to_oblique.json constructions/transitive_to.json constructions/buy_from.json constructions/question.json constructions/wh_question.json constructions/identification.json constructions/identification_negative.json constructions/attribution.json constructions/attribution_plural.json constructions/attribution_negative.json constructions/attributed_seeing.json constructions/attributed_seeing_object.json constructions/attributed_seeing_object_negative.json constructions/attributed_seeing_both.json constructions/attributed_seeing_plural.json constructions/attributed_seeing_object_plural.json constructions/attributed_seeing_both_plural.json constructions/coordination.json constructions/scoped_clause.json constructions/phrases.json
var sentenceConstructions embed.FS

// SentencePhraseProfile loads the shared structural phrase profile for the
// exact language feature profile used by source analysis.
func SentencePhraseProfile(profile *features.LanguageFeatureProfile) (construction.PhraseProfile, error) {
	if profile == nil || profile.Language() != "pl" {
		return construction.PhraseProfile{}, fmt.Errorf("Polish phrase profile requires a Polish feature profile")
	}
	data, err := sentenceConstructions.ReadFile("constructions/phrases.json")
	if err != nil {
		return construction.PhraseProfile{}, err
	}
	return construction.DecodePhraseProfile(data, profile)
}

// SentenceScopedClauses compiles the language-local bounded scope patterns.
func SentenceScopedClauses() (construction.ScopedClausePlan, error) {
	data, err := sentenceConstructions.ReadFile("constructions/scoped_clause.json")
	if err != nil {
		return construction.ScopedClausePlan{}, err
	}
	return construction.DecodeScopedClausePlan(data, "pl")
}

// SentenceCoordination compiles the explicit language-local clause join plan.
func SentenceCoordination() (construction.ClauseCoordinationPlan, error) {
	data, err := sentenceConstructions.ReadFile("constructions/coordination.json")
	if err != nil {
		return construction.ClauseCoordinationPlan{}, err
	}
	return construction.DecodeClauseCoordinationPlan(data, "pl")
}

// SentenceConstructions compiles the language-local templates against the
// authored valency index. Only exact approved bindings in that index license
// a sense; the presence of a family alone grants no sentence capability.
func SentenceConstructions(index *valency.Index) ([]construction.CompiledConstructionPackage, *construction.Registry, error) {
	files := []string{
		"transitive_personal.json", "transitive_personal_explicit.json", "transitive.json", "transitive_negative.json", "oblique.json", "oblique_negative.json", "intransitive.json", "ditransitive.json", "ditransitive_negative.json",
		"double_np.json", "transitive_oblique.json", "transitive_instrumental.json", "with_oblique.json", "to_oblique.json", "transitive_to.json", "buy_from.json", "question.json", "wh_question.json", "identification.json", "identification_negative.json", "attribution.json", "attribution_plural.json", "attribution_negative.json", "attributed_seeing.json", "attributed_seeing_object.json", "attributed_seeing_object_negative.json", "attributed_seeing_both.json", "attributed_seeing_plural.json", "attributed_seeing_object_plural.json", "attributed_seeing_both_plural.json",
	}
	packages := make([]construction.CompiledConstructionPackage, 0, len(files))
	families := make([]construction.ConstructionFamily, 0, len(files))
	for _, file := range files {
		data, err := sentenceConstructions.ReadFile("constructions/" + file)
		if err != nil {
			return nil, nil, fmt.Errorf("Polish construction %s: %w", file, err)
		}
		compiled, err := construction.DecodeConstructionPackage(data)
		if err != nil {
			return nil, nil, fmt.Errorf("Polish construction %s: %w", file, err)
		}
		packages = append(packages, compiled)
		families = append(families, compiled.Package.Family)
	}
	registry, err := construction.CompileRegistry(families, index)
	if err != nil {
		return nil, nil, err
	}
	return packages, registry, nil
}
