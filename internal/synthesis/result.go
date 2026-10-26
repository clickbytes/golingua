package synthesis

import (
	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

// GoalObligation reports which bounded construction steps were associated
// with one input constraint. Its status records construction provenance only;
// it is not an independent proof or a fact about the world.
type GoalObligation struct {
	ConstraintID string         `json:"constraint_id"`
	Kind         ConstraintKind `json:"kind"`
	Status       string         `json:"status"`
	StepIDs      []string       `json:"step_ids"`
	Fragments    []string       `json:"fragments"`
}

// Result contains a checked representation only when the goal was fully
// specified and construction completed. An incomplete result has no
// expression, checked document, or construction trace.
type Result struct {
	Status            status.OutcomeKind         `json:"status"`
	Expression        *algebra.Term              `json:"expression,omitempty"`
	Checked           *algebra.CheckedExpression `json:"checked,omitempty"`
	Interface         []algebra.PortBinding      `json:"interface,omitempty"`
	CanonicalDocument *algebra.Document          `json:"canonical_document,omitempty"`
	Certificate       *SynthesisCertificate      `json:"certificate,omitempty"`
	Verification      *GoalVerificationResult    `json:"verification,omitempty"`
	FullyValidated    bool                       `json:"fully_validated"`
	Trace             []algebra.BuildTrace       `json:"trace"`
	Unresolved        []Unresolved               `json:"unresolved"`
	Obligations       []GoalObligation           `json:"obligations"`
}

const (
	SearchOutcomeCandidates             = "candidates"
	SearchOutcomeNoSolutionWithinBounds = "no_solution_within_profile_bounds"
	SearchOutcomeIncomplete             = "incomplete"
	SearchOutcomeUnresolvedGoal         = "unresolved_goal"
	DraftGoalCertificateStatus          = "draft_goal_unverified"
	SearchValidationScope               = "finite positive-facts path profile; construction and alpha identity only"
)

// SearchResult reports bounded candidate construction within a declared
// profile. Complete and UniqueWithinDeclaredSpace are scoped to that finite
// space and never establish world truth, speaker intent, or language quality.
type SearchResult struct {
	SchemaVersion             int                   `json:"schema_version"`
	Status                    status.OutcomeKind    `json:"status"`
	Outcome                   string                `json:"outcome"`
	Profile                   string                `json:"profile"`
	Goal                      GoalSpec              `json:"goal"`
	GoalFingerprint           string                `json:"goal_fingerprint"`
	RegistrySnapshotID        string                `json:"registry_snapshot_id"`
	ContextID                 string                `json:"context_id"`
	Limits                    SearchLimits          `json:"limits"`
	Space                     SearchSpaceDescriptor `json:"space"`
	Candidates                []SearchCandidate     `json:"candidates"`
	Complete                  bool                  `json:"complete"`
	UniqueWithinDeclaredSpace bool                  `json:"unique_within_declared_space"`
	Usage                     SearchUsage           `json:"usage"`
	Trace                     []SearchTraceEntry    `json:"trace"`
	Rejected                  int                   `json:"rejected"`
	EquivalentDuplicates      int                   `json:"equivalent_duplicates"`
	ExhaustedBounds           []SearchStopReason    `json:"exhausted_bounds,omitempty"`
	StopReason                SearchStopReason      `json:"stop_reason,omitempty"`
	ValidationScope           string                `json:"validation_scope"`
}

// SearchCandidate is one constructed representative and its alpha-only key.
// Its draft certificate binds construction provenance but explicitly does not
// claim independent verification of the original goal.
type SearchCandidate struct {
	ID               string                         `json:"id"`
	Key              string                         `json:"alpha_key"`
	ConcreteGoal     GoalSpec                       `json:"concrete_goal"`
	Choices          []ChoiceAssignment             `json:"choices"`
	Construction     Result                         `json:"construction"`
	DraftCertificate DraftSynthesisCertificate      `json:"draft_certificate"`
	ExportRequest    *algebra.ArtifactExportRequest `json:"export_request,omitempty"`
}

// DraftSynthesisCertificate is a candidate receipt for a later independent
// verifier. The embedded rewrite verification covers alpha normalization
// only; GoalVerified intentionally remains false until a goal verifier exists.
type DraftSynthesisCertificate struct {
	SchemaVersion        int                        `json:"schema_version"`
	Status               string                     `json:"status"`
	GoalVerified         bool                       `json:"goal_verified"`
	GoalFingerprint      string                     `json:"goal_fingerprint"`
	RegistrySnapshotID   string                     `json:"registry_snapshot_id"`
	ContextID            string                     `json:"context_id"`
	CandidateKey         string                     `json:"candidate_key"`
	Choices              []ChoiceAssignment         `json:"choices"`
	OperatorTrace        []algebra.BuildTrace       `json:"operator_trace"`
	Obligations          []GoalObligation           `json:"obligations"`
	IdentityProfile      algebra.RewriteProfile     `json:"identity_profile"`
	IdentityCertificate  algebra.RewriteCertificate `json:"identity_certificate"`
	IdentityVerification algebra.VerificationResult `json:"identity_verification"`
}

// SearchTraceEntry is a compact operational record. It deliberately omits
// complete goal and expression copies so a bounded trace cannot multiply
// candidate-sized payloads.
type SearchTraceEntry struct {
	Attempt      int                `json:"attempt"`
	Stage        string             `json:"stage"`
	Code         string             `json:"code,omitempty"`
	ConstraintID string             `json:"constraint_id,omitempty"`
	InputPath    string             `json:"input_path,omitempty"`
	Message      string             `json:"message,omitempty"`
	CandidateID  string             `json:"candidate_id,omitempty"`
	ChoiceKey    string             `json:"choice_key,omitempty"`
	Status       status.OutcomeKind `json:"status,omitempty"`
}
