package en

import (
	"embed"
	"fmt"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/valency"
)

//go:embed constructions/intransitive.json constructions/transitive_personal.json constructions/transitive.json constructions/transitive_past.json constructions/transitive_negative.json constructions/oblique.json constructions/to_oblique.json constructions/oblique_negative.json constructions/question.json constructions/wh_question.json constructions/identification.json constructions/identification_negative.json constructions/attribution.json constructions/attributed_seeing.json constructions/give_transfer.json constructions/coordination.json constructions/scoped_clause.json
var sentenceConstructions embed.FS

// SentenceScopedClauses compiles the language-local bounded scope patterns.
func SentenceScopedClauses() (construction.ScopedClausePlan, error) {
	data, err := sentenceConstructions.ReadFile("constructions/scoped_clause.json")
	if err != nil {
		return construction.ScopedClausePlan{}, err
	}
	return construction.DecodeScopedClausePlan(data, "en")
}

// SentenceCoordination compiles the explicit language-local clause join plan.
func SentenceCoordination() (construction.ClauseCoordinationPlan, error) {
	data, err := sentenceConstructions.ReadFile("constructions/coordination.json")
	if err != nil {
		return construction.ClauseCoordinationPlan{}, err
	}
	return construction.DecodeClauseCoordinationPlan(data, "en")
}

// SentenceConstructions compiles the language-local templates against the
// authored valency index. Only exact approved bindings in that index license
// a sense; the presence of a family alone grants no sentence capability.
func SentenceConstructions(index *valency.Index) ([]construction.CompiledConstructionPackage, *construction.Registry, error) {
	files := []string{"intransitive.json", "transitive_personal.json", "transitive.json", "transitive_past.json", "transitive_negative.json", "oblique.json", "to_oblique.json", "oblique_negative.json", "question.json", "wh_question.json", "identification.json", "identification_negative.json", "attribution.json", "attributed_seeing.json", "give_transfer.json"}
	packages := make([]construction.CompiledConstructionPackage, 0, len(files))
	families := make([]construction.ConstructionFamily, 0, len(files))
	for _, file := range files {
		data, err := sentenceConstructions.ReadFile("constructions/" + file)
		if err != nil {
			return nil, nil, fmt.Errorf("English construction %s: %w", file, err)
		}
		compiled, err := construction.DecodeConstructionPackage(data)
		if err != nil {
			return nil, nil, fmt.Errorf("English construction %s: %w", file, err)
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
