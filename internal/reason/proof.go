package reason

import (
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/meaning"
)

// Claim is an explicitly asserted or derived positive binary proposition.
type Claim struct {
	Predicate string                `json:"predicate"`
	Args      [2]meaning.ReferentID `json:"args"`
}

// ProofNode records one derivation. A direct node has an empty RuleID and one
// source fact. Derived nodes name their immediate premise proof IDs and every
// originating fact ID; provenance can be retrieved from the returned facts.
type ProofNode struct {
	ID              string                 `json:"id"`
	RuleID          string                 `json:"rule_id,omitempty"`
	PremiseProofIDs []string               `json:"premise_proof_ids,omitempty"`
	PremiseFactIDs  []string               `json:"premise_fact_ids"`
	Sources         []facts.SourceEvidence `json:"sources"`
	Conclusion      Claim                  `json:"conclusion"`
	Depth           int                    `json:"depth"`
}

// Truth distinguishes lack of evidence from an explicit negative assertion.
type Truth string

const (
	TruthUnknown  Truth = "unknown"
	TruthTrue     Truth = "true"
	TruthFalse    Truth = "false"
	TruthConflict Truth = "conflict"
)

// Result contains all proof nodes needed for the selected positive claim and
// direct negative evidence. No answer is inferred from missing facts.
type Result struct {
	Truth            Truth        `json:"truth"`
	Proof            []ProofNode  `json:"proof,omitempty"`
	NegativeEvidence []facts.Fact `json:"negative_evidence,omitempty"`
}
