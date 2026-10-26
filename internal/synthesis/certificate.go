package synthesis

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	SynthesisCertificateSchemaVersion  = 1
	SynthesisCertificateFormat         = "finite-goal-certificate-v1"
	SynthesisCertificateProducerStatus = "constructed_unverified"

	CertificateOriginConstruct = "construct"
	CertificateOriginSearch    = "search"

	CertificateAssumptionRegistrySnapshot      = "typed_registry_snapshot"
	CertificateAssumptionRequestReferenceBound = "request_reference_boundary"
	CertificateAssumptionPositiveProfile       = "positive_structural_profile"

	DefaultGoalVerificationMaxCertificateBytes = 8 << 20
	MaxGoalVerificationMaxCertificateBytes     = 8 << 20
	DefaultGoalVerificationMaxReplaySteps      = MaxGoalSteps
	MaxGoalVerificationMaxReplaySteps          = MaxGoalSteps
	DefaultGoalVerificationMaxWorkNodes        = 262_144
	MaxGoalVerificationMaxWorkNodes            = 262_144

	GoalVerificationDispositionChecked      = "checked"
	GoalVerificationDispositionRejected     = "rejected"
	GoalVerificationDispositionUndetermined = "undetermined"
	GoalVerificationConsistencyScope        = "structural_profile_only"
)

const (
	CertificateRuleRelation = "algebra.relation.v1"
	CertificateRuleJoin     = "algebra.join.v1"
	CertificateRuleRestrict = "algebra.restrict.v1"
	CertificateRuleBind     = "algebra.bind.v1"
	CertificateRuleProject  = "algebra.project.v1"
	CertificateRulePermute  = "algebra.permute.v1"
)

// GoalVerificationLimits bound independent replay and structural checking.
// Zero fields select finite defaults; they never mean unlimited work.
type GoalVerificationLimits struct {
	MaxCertificateBytes int            `json:"max_certificate_bytes,omitempty"`
	MaxReplaySteps      int            `json:"max_replay_steps,omitempty"`
	MaxWorkNodes        int            `json:"max_work_nodes,omitempty"`
	TermLimits          algebra.Limits `json:"term_limits,omitempty"`
}

// NormalizeGoalVerificationLimits applies finite defaults before a verifier
// allocates replay state or inspects an untrusted certificate.
func NormalizeGoalVerificationLimits(requested GoalVerificationLimits) (GoalVerificationLimits, error) {
	fields := []struct {
		name     string
		value    *int
		fallback int
		maximum  int
	}{
		{"max_certificate_bytes", &requested.MaxCertificateBytes, DefaultGoalVerificationMaxCertificateBytes, MaxGoalVerificationMaxCertificateBytes},
		{"max_replay_steps", &requested.MaxReplaySteps, DefaultGoalVerificationMaxReplaySteps, MaxGoalVerificationMaxReplaySteps},
		{"max_work_nodes", &requested.MaxWorkNodes, DefaultGoalVerificationMaxWorkNodes, MaxGoalVerificationMaxWorkNodes},
	}
	for _, field := range fields {
		if *field.value == 0 {
			*field.value = field.fallback
			continue
		}
		if *field.value < 0 || *field.value > field.maximum {
			return GoalVerificationLimits{}, &algebra.Error{
				Status: status.OutcomeInvalid, Code: "INVALID_GOAL_VERIFICATION_LIMIT",
				Path:    "limits." + field.name,
				Message: fmt.Sprintf("must be zero (default) or between 1 and %d", field.maximum),
			}
		}
	}
	termLimits, err := algebra.NormalizeLimits(requested.TermLimits)
	if err != nil {
		return GoalVerificationLimits{}, err
	}
	requested.TermLimits = termLimits
	return requested, nil
}

// SynthesisCertificate contains producer claims for later independent replay.
// ProducerStatus and GoalVerified are untrusted claims; neither establishes
// that the candidate satisfies its goal until a verifier recomputes it.
type SynthesisCertificate struct {
	SchemaVersion        int                     `json:"schema_version"`
	Format               string                  `json:"format"`
	ProducerStatus       string                  `json:"producer_status"`
	GoalVerified         bool                    `json:"goal_verified"`
	Origin               string                  `json:"origin"`
	GoalFingerprint      string                  `json:"goal_fingerprint"`
	SnapshotID           string                  `json:"snapshot_id"`
	ContextID            string                  `json:"context_id"`
	Profile              string                  `json:"profile"`
	ConcreteGoal         GoalSpec                `json:"concrete_goal"`
	Choices              []ChoiceAssignment      `json:"choices"`
	CandidateFingerprint string                  `json:"candidate_fingerprint"`
	Interface            []algebra.PortBinding   `json:"interface"`
	Operations           []CertifiedOperation    `json:"operations"`
	Scopes               []CertificateScope      `json:"scopes"`
	Obligations          []GoalObligation        `json:"obligations"`
	Assumptions          []CertificateAssumption `json:"assumptions"`
	IdentityProof        *SynthesisIdentityProof `json:"identity_proof,omitempty"`
	IdentityKey          string                  `json:"identity_key,omitempty"`
}

// CertifiedOperation records claimed public algebra operator inputs and
// outputs. Replay must independently recompute each operation and its scope.
type CertifiedOperation struct {
	ID            string                  `json:"id"`
	RuleID        string                  `json:"rule_id"`
	Op            string                  `json:"op"`
	Inputs        []string                `json:"inputs"`
	Symbol        string                  `json:"symbol,omitempty"`
	Mapping       []algebra.JoinPair      `json:"mapping"`
	Outputs       []algebra.JoinOutput    `json:"outputs"`
	Port          string                  `json:"port,omitempty"`
	Ports         []string                `json:"ports"`
	ReferenceID   string                  `json:"reference_id,omitempty"`
	GoalPortID    string                  `json:"goal_port_id,omitempty"`
	ClaimedStatus status.OutcomeKind      `json:"claimed_status"`
	CheckedType   *algebra.Type           `json:"checked_type,omitempty"`
	Interface     []algebra.PortBinding   `json:"interface"`
	Substitutions []CertifiedSubstitution `json:"substitutions"`
}

// CertifiedSubstitution binds one claimed interface binder to an explicitly
// declared goal reference. Replay checks the reference and type itself.
type CertifiedSubstitution struct {
	BinderID    string       `json:"binder_id"`
	ReferenceID string       `json:"reference_id"`
	Type        algebra.Type `json:"type"`
}

// CertificateScope records claimed lambda and existential binder ancestry.
// It is descriptive input to verification, not authority over AST scope.
type CertificateScope struct {
	BinderID       string           `json:"binder_id"`
	Kind           algebra.TermKind `json:"kind"`
	Type           algebra.Type     `json:"type"`
	Path           string           `json:"path"`
	ParentBinderID string           `json:"parent_binder_id,omitempty"`
}

// CertificateAssumption is restricted to explicit request and profile bounds;
// it cannot declare domain axioms or world facts.
type CertificateAssumption struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// SynthesisIdentityProof carries the alpha-only identity certificate emitted
// by the shared rewrite kernel. It does not prove goal satisfaction.
type SynthesisIdentityProof struct {
	Profile     algebra.RewriteProfile     `json:"profile"`
	Certificate algebra.RewriteCertificate `json:"certificate"`
}

// GoalVerificationStatistics reports bounded structural work, not elapsed
// compute, energy, theorem search, or world evaluation.
type GoalVerificationStatistics struct {
	ReplaySteps      int `json:"replay_steps"`
	VisitedNodes     int `json:"visited_nodes"`
	WorkNodes        int `json:"work_nodes"`
	CertificateBytes int `json:"certificate_bytes"`
}

// GoalVerificationResult separates replay, type correctness and goal coverage.
// A single candidate cannot establish search uniqueness or global completeness.
type GoalVerificationResult struct {
	Status                     status.OutcomeKind         `json:"status"`
	Disposition                string                     `json:"disposition"`
	TypeCorrect                bool                       `json:"type_correct"`
	RefinementChecked          bool                       `json:"refinement_checked"`
	ReplayChecked              bool                       `json:"replay_checked"`
	CertificateValid           bool                       `json:"certificate_valid"`
	GoalSatisfied              bool                       `json:"goal_satisfied"`
	Complete                   bool                       `json:"complete"`
	GoalFingerprint            string                     `json:"goal_fingerprint,omitempty"`
	SnapshotID                 string                     `json:"snapshot_id,omitempty"`
	ContextID                  string                     `json:"context_id,omitempty"`
	ConstraintCheck            *GoalConstraintCheck       `json:"constraint_check,omitempty"`
	Assumptions                []string                   `json:"assumptions"`
	ConsistencyScope           string                     `json:"consistency_scope"`
	UniquenessVerified         bool                       `json:"uniqueness_verified"`
	SearchCompletenessVerified bool                       `json:"search_completeness_verified"`
	Diagnostics                []algebra.Error            `json:"diagnostics"`
	Statistics                 GoalVerificationStatistics `json:"statistics"`
}
