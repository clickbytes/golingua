// Package morph defines bounded language-specific paradigms and explicit
// principal parts. It delegates transformations to the language-neutral
// procedure interpreter.
package morph

import (
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/procedure"
)

type ParadigmID string
type CellID string
type PrincipalPartID string

type PrincipalPartRequirement struct {
	StemID   procedure.StemID `json:"stem_id"`
	PartID   PrincipalPartID  `json:"part_id"`
	Required bool             `json:"required"`
}

// ParadigmCell binds a stable named form slot to an explicit known feature
// bundle. A slot's rule and its principal parts are compiled separately.
type ParadigmCell struct {
	ID       CellID                  `json:"id"`
	Features []features.FeatureValue `json:"features"`
}

type CellRule struct {
	CellID CellID                `json:"cell_id"`
	Steps  []procedure.Operation `json:"steps"`
}

type Paradigm struct {
	SchemaVersion             int                        `json:"schema_version"`
	ID                        ParadigmID                 `json:"id"`
	Language                  features.LanguageID        `json:"language"`
	ProfileVersion            features.ProfileVersion    `json:"profile_version"`
	POS                       lexicon.PartOfSpeech       `json:"pos"`
	Inherits                  []ParadigmID               `json:"inherits,omitempty"`
	PrincipalPartRequirements []PrincipalPartRequirement `json:"principal_part_requirements"`
	Cells                     []ParadigmCell             `json:"cells"`
	Rules                     []CellRule                 `json:"rules"`
}
