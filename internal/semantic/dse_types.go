package semantic

import (
	"context"
	"encoding/json"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

const (
	DefaultDSEAlternatives = 64
	MaxDSEAlternatives     = 128
	DefaultDSEWork         = 4096
	MaxDSEWork             = 16384
	MaxDSEOutputBytes      = 8 << 20
)

// DSEBudget bounds interpretation expansion and retained canonical documents.
// Zero fields select finite defaults; negative and above-cap values are errors.
type DSEBudget struct {
	MaxAlternatives int `json:"max_alternatives,omitempty"`
	MaxWork         int `json:"max_work,omitempty"`
	MaxOutputBytes  int `json:"max_output_bytes,omitempty"`
}

func NormalizeDSEBudget(requested DSEBudget) (DSEBudget, error) {
	if requested.MaxAlternatives == 0 {
		requested.MaxAlternatives = DefaultDSEAlternatives
	}
	if requested.MaxWork == 0 {
		requested.MaxWork = DefaultDSEWork
	}
	if requested.MaxOutputBytes == 0 {
		requested.MaxOutputBytes = MaxDSEOutputBytes
	}
	if requested.MaxAlternatives < 1 || requested.MaxAlternatives > MaxDSEAlternatives || requested.MaxWork < 1 || requested.MaxWork > MaxDSEWork || requested.MaxOutputBytes < 1 || requested.MaxOutputBytes > MaxDSEOutputBytes {
		return DSEBudget{}, dseError(status.OutcomeInvalid, "INVALID_DSE_BUDGET", "budget", "DSE limits must be positive and within the finite profile caps")
	}
	return requested, nil
}

// DSEInput contains already analyzed structures and explicitly selected,
// validated metadata. ComposeDSE never reads files or resolves a language.
type DSEInput struct {
	Context                  context.Context
	Forest                   parser.Forest
	Candidates               []parser.PhraseCandidate
	Lattice                  analyzer.LexicalLattice
	SourceBytes              int
	References               []ReferenceEvidence
	NodeAttributes           []NodeAttributesEvidence
	Operators                []OperatorEvidence
	OperatorCoverageComplete bool
	Package                  *language.SemanticLanguagePackage
	Plan                     *language.SourcePlan
	Selected                 algebra.Document
	Rewrite                  algebra.RewriteEnvironment
	ReferenceContext         language.SourceReferenceContext
	Language                 features.LanguageID
	Budget                   DSEBudget
}

// DSESourceRecord is provenance carried beside, never substituted for, the
// constructed algebra. Typed IDs are represented by their stable string form.
type DSESourceRecord struct {
	CandidateID  string    `json:"candidate_id,omitempty"`
	ReadingID    string    `json:"reading_id,omitempty"`
	ReadingIDs   []string  `json:"reading_ids"`
	SenseID      string    `json:"sense_id,omitempty"`
	FrameID      string    `json:"frame_id,omitempty"`
	FamilyID     string    `json:"family_id,omitempty"`
	VariantID    string    `json:"variant_id,omitempty"`
	DerivationID string    `json:"derivation_id,omitempty"`
	Span         text.Span `json:"span"`
}

type DSEConditionCoverage struct {
	Kind               language.SourceConditionKind `json:"kind"`
	Code               string                       `json:"code"`
	ConditionID        string                       `json:"condition_id"`
	TargetReferenceIDs []string                     `json:"target_reference_ids"`
}

type DSERoleSource struct {
	Role            string          `json:"role"`
	Source          DSESourceRecord `json:"source"`
	ReferenceID     string          `json:"reference_id,omitempty"`
	ResolutionBasis string          `json:"resolution_basis,omitempty"`
}

type DSEResidual struct {
	Code        string `json:"code"`
	CandidateID string `json:"candidate_id,omitempty"`
	Path        string `json:"path,omitempty"`
	Message     string `json:"message"`
}

type DSECandidate struct {
	ID                   string                     `json:"id"`
	Status               status.OutcomeKind         `json:"status"`
	Source               DSESourceRecord            `json:"source"`
	HeadEventReferenceID string                     `json:"head_event_reference_id,omitempty"`
	RoleSources          []DSERoleSource            `json:"role_sources"`
	CanonicalDocument    json.RawMessage            `json:"canonical_document,omitempty"`
	Checked              *algebra.CheckedExpression `json:"checked,omitempty"`
	GoalSpecs            []json.RawMessage          `json:"goal_specs"`
	FragmentChecks       []DSEFragmentCheck         `json:"fragment_checks"`
	ConditionCoverage    []DSEConditionCoverage     `json:"condition_coverage"`
	Residuals            []DSEResidual              `json:"residuals"`
	Scope                string                     `json:"scope"`
}

// DSEFragmentCheck is only a structural check of an instantiated target.
// GoalSatisfied is never inferred from a Symbol or reference fragment.
type DSEFragmentCheck struct {
	Kind         string                `json:"kind"`
	BindingKey   string                `json:"binding_key,omitempty"`
	Status       status.OutcomeKind    `json:"status"`
	CheckedType  *algebra.Type         `json:"checked_type,omitempty"`
	Interface    []algebra.PortBinding `json:"interface"`
	GoalEvidence json.RawMessage       `json:"goal_evidence,omitempty"`
}

type DSEResult struct {
	Status                 status.OutcomeKind  `json:"status"`
	Complete               bool                `json:"complete"`
	Candidates             []DSECandidate      `json:"candidates"`
	Residuals              []DSEResidual       `json:"residuals"`
	Diagnostics            []status.Diagnostic `json:"diagnostics"`
	Work                   int                 `json:"work"`
	OutputBytes            int                 `json:"output_bytes"`
	WholeVerificationScope string              `json:"whole_verification_scope"`
}

type DSEError struct{ result status.Result }

func (e *DSEError) Error() string {
	if e == nil || len(e.result.Diagnostics) == 0 {
		return "DSE composition error"
	}
	return e.result.Diagnostics[0].Message
}
func (e *DSEError) Result() status.Result {
	if e == nil {
		return status.Result{}
	}
	out := e.result
	out.Diagnostics = append([]status.Diagnostic(nil), e.result.Diagnostics...)
	return out
}
