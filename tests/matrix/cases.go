// Package matrix defines declarative cases used by the language behavior matrix.
package matrix

// AnalysisCase states what a source language must preserve while analyzing input.
type AnalysisCase struct {
	ID, MeaningID, Input                        string
	MustPreserve                                []string
	AmbiguityNote, InputOrigin, OwnerAcceptance string
}

// RealizationCase states the available proposal texts for realizing a Meaning.
type RealizationCase struct {
	ID, MeaningID                                   string
	ProposalTexts, AllowedTexts, RequiredTexts      []string
	AllowedAssumptionFields                         []string
	RequiredAssumptionFields                        []string
	RequiredTextAssumptions                         []TextAssumptionRequirement
	AllowedLossFields, RequiredLossFields           []string
	AmbiguityNote, ReferenceOrigin, OwnerAcceptance string
}

// TranslationCase links one source analysis to a target realization.
type TranslationCase struct {
	ID, From, To, AnalysisID, RealizationID string
	CoverageTags                            []string
}

// TextAssumptionRequirement ties a selected surface form to the assumption
// value that licenses that exact form.
type TextAssumptionRequirement struct {
	Text, Field, Value string
}
