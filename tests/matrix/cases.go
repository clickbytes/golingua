// Package matrix defines declarative cases used by the language behavior matrix.
package matrix

// AnalysisCase states what a source language must preserve while analyzing input.
type AnalysisCase struct {
	ID, MeaningID, Input                        string
	AllowedMeaningIDs, RequiredMeaningIDs       []string
	MustPreserve                                []string
	SemanticRequirement                         SemanticRequirement
	AmbiguityNote, InputOrigin, OwnerAcceptance string
}

// RealizationCase states the available proposal texts for realizing a Meaning.
type RealizationCase struct {
	ID, MeaningID                                   string
	ProposalTexts, AllowedTexts, RequiredTexts      []string
	MeaningVariants                                 []MeaningRealization
	AllowedAssumptionFields                         []string
	RequiredAssumptionFields                        []string
	RequiredTextAssumptions                         []TextAssumptionRequirement
	SemanticRequirement                             SemanticRequirement
	AllowedLossFields, RequiredLossFields           []string
	AmbiguityNote, ReferenceOrigin, OwnerAcceptance string
}

// MeaningRealization binds text expectations to one allowed authored meaning.
// It inherits the parent case's assumptions and loss rules.
type MeaningRealization struct {
	MeaningID                   string
	AllowedTexts, RequiredTexts []string
	RequiredTextAssumptions     []TextAssumptionRequirement
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

// SemanticRequirement records explicit meaning that a case requires. A
// RepresentationGap names the exact requirement the current Meaning model
// cannot encode; it keeps the case evaluable without inventing a graph.
type SemanticRequirement struct {
	RequiredPredicate, RequiredAssertion string
	RequiredRoles                        []string
	RequiredPerson, RequiredQuantity     string
	RequiredReference, RequiredAct       string
	RepresentationGap                    string
}
