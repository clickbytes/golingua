package reason

import (
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

type QuestionKind string

const (
	QuestionYesNo QuestionKind = "yes_no"
	QuestionWH    QuestionKind = "wh"
)

// RelationMapping is explicit semantic evidence supplied by a language or
// domain provider. Role names map to relation argument positions 0 and 1.
type RelationMapping struct {
	Predicate      string               `json:"predicate"`
	RoleToArgument map[meaning.Role]int `json:"role_to_argument"`
}

// QuestionInput combines a validated typed MeaningGraph with an explicit
// relation binding. No natural language text or label matching is performed.
type QuestionInput struct {
	Meaning meaning.Meaning        `json:"meaning"`
	Context meaning.RequestContext `json:"context"`
	Kind    QuestionKind           `json:"kind"`
	Mapping RelationMapping        `json:"mapping"`
	GapRole meaning.Role           `json:"gap_role,omitempty"`
	At      string                 `json:"at,omitempty"`
}

// Question is the grounded query planned from a question MeaningGraph.
// A WH gap has an empty argument at GapIndex; yes/no uses GapIndex -1.
type Question struct {
	Kind               QuestionKind          `json:"kind"`
	Predicate          string                `json:"predicate"`
	Args               [2]meaning.ReferentID `json:"args"`
	GapIndex           int                   `json:"gap_index"`
	GapRole            meaning.Role          `json:"gap_role,omitempty"`
	PredicateConceptID string                `json:"predicate_concept_id,omitempty"`
	At                 string                `json:"at,omitempty"`
	MeaningNodeID      meaning.NodeID        `json:"meaning_node_id"`
}

type QuestionTruth string

const (
	QuestionUnknown   QuestionTruth = "unknown"
	QuestionTrue      QuestionTruth = "true"
	QuestionFalse     QuestionTruth = "false"
	QuestionConflict  QuestionTruth = "conflict"
	QuestionAmbiguous QuestionTruth = "ambiguous"
)

type CandidateAnswer struct {
	ReferentID       meaning.ReferentID `json:"referent_id"`
	Truth            QuestionTruth      `json:"truth"`
	Proof            []ProofNode        `json:"proof,omitempty"`
	NegativeEvidence []facts.Fact       `json:"negative_evidence,omitempty"`
}

// QuestionResult contains typed referent IDs and evidence. It carries no
// guessed surface wording; a later response planner handles realization.
type QuestionResult struct {
	Kind             QuestionKind       `json:"kind"`
	Truth            QuestionTruth      `json:"truth"`
	Outcome          status.OutcomeKind `json:"outcome"`
	YesNo            *bool              `json:"yes_no,omitempty"`
	Candidates       []CandidateAnswer  `json:"candidates,omitempty"`
	Proof            []ProofNode        `json:"proof,omitempty"`
	NegativeEvidence []facts.Fact       `json:"negative_evidence,omitempty"`
}

// QueryBudget bounds WH referent enumeration in addition to each reasoner
// evaluation's own work and derivation budgets.
type QueryBudget struct {
	MaxWHCandidates int
	Reasoner        Budget
}

func DefaultQueryBudget() QueryBudget {
	return QueryBudget{MaxWHCandidates: 64, Reasoner: DefaultBudget()}
}
