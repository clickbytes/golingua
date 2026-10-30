package semantic

import (
	"context"
	"encoding/json"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
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
	ImplicitRoles            []ImplicitRoleEvidence
	NodeAttributes           []NodeAttributesEvidence
	Operators                []OperatorEvidence
	OperatorCoverageComplete bool
	Package                  *language.SemanticLanguagePackage
	Lexicon                  *lexicon.Index
	Valency                  *valency.Index
	Core                     *core.ConceptGraph
	Constructions            *construction.Registry
	Plan                     *language.SourcePlan
	Selected                 algebra.Document
	Rewrite                  algebra.RewriteEnvironment
	ReferenceContext         language.SourceReferenceContext
	Language                 features.LanguageID
	Budget                   DSEBudget
	// RequireCandidateHeadReference prevents an Event port mapping from being
	// reused across coordinated clauses. The source predicate candidate must
	// carry its own corroborated request-context identity.
	RequireCandidateHeadReference bool
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
	Role             string                `json:"role"`
	Source           DSESourceRecord       `json:"source"`
	ReferenceID      string                `json:"reference_id,omitempty"`
	ResolutionBasis  string                `json:"resolution_basis,omitempty"`
	WHGap            *DSEWHGapEvidence     `json:"wh_gap,omitempty"`
	ImplicitEvidence *ImplicitRoleEvidence `json:"-"`
}

type DSEWHGapEvidence struct {
	SlotID      string    `json:"slot_id"`
	Role        string    `json:"role"`
	ScopeNodeID string    `json:"scope_node_id"`
	Span        text.Span `json:"span"`
}

type DSEResidual struct {
	Code        string `json:"code"`
	CandidateID string `json:"candidate_id,omitempty"`
	Path        string `json:"path,omitempty"`
	Message     string `json:"message"`
}

type DSECandidate struct {
	ID                     string                     `json:"id"`
	Status                 status.OutcomeKind         `json:"status"`
	PredicateKind          string                     `json:"predicate_kind,omitempty"`
	Source                 DSESourceRecord            `json:"source"`
	HeadEventReferenceID   string                     `json:"head_event_reference_id,omitempty"`
	ClauseSources          []DSEClauseSource          `json:"clause_sources,omitempty"`
	RoleSources            []DSERoleSource            `json:"role_sources"`
	CanonicalDocument      json.RawMessage            `json:"canonical_document,omitempty"`
	Checked                *algebra.CheckedExpression `json:"checked,omitempty"`
	GoalSpecs              []json.RawMessage          `json:"goal_specs"`
	FragmentChecks         []DSEFragmentCheck         `json:"fragment_checks"`
	ConditionCoverage      []DSEConditionCoverage     `json:"condition_coverage"`
	Residuals              []DSEResidual              `json:"residuals"`
	Scope                  string                     `json:"scope"`
	ScopedEvidence         *DSEScopedEvidence         `json:"scoped_evidence,omitempty"`
	RelationSourceEvidence *DSERelationSourceEvidence `json:"relation_source_evidence,omitempty"`
}

// DSERelationSourceEvidence preserves the typed source relation match and
// selected metadata without depending on runtime or engine packages.
type DSERelationSourceEvidence struct {
	VariantID              string                          `json:"variant_id"`
	Variant                DSERelationPlanVariantEvidence  `json:"variant"`
	RelationPlanProvenance provenance.SourceRef            `json:"relation_plan_provenance"`
	SourcePlanProvenance   provenance.SourceRef            `json:"source_plan_provenance"`
	SourceParts            []DSERelationSourcePartEvidence `json:"source_parts"`
	SourceContext          language.SourceReferenceContext `json:"source_context"`
	Metadata               *DSERelationMetadataEvidence    `json:"metadata,omitempty"`
	TemporalFrames         [2]string                       `json:"temporal_frames,omitempty"`
	VerificationScope      string                          `json:"verification_scope"`
}

type DSERelationPlanVariantEvidence struct {
	ID    string                        `json:"id"`
	Parts []DSERelationPlanPartEvidence `json:"parts"`
}

type DSERelationPlanPartEvidence struct {
	Step                string `json:"step"`
	Literal             string `json:"literal,omitempty"`
	Slot                string `json:"slot,omitempty"`
	Port                string `json:"port,omitempty"`
	GrammaticalFunction string `json:"grammatical_function,omitempty"`
}

type DSERelationSourcePartEvidence struct {
	PartIndex           int                   `json:"part_index"`
	Step                string                `json:"step"`
	Slot                string                `json:"slot,omitempty"`
	Literal             string                `json:"literal,omitempty"`
	Port                string                `json:"port,omitempty"`
	PortPosition        string                `json:"port_position,omitempty"`
	GrammaticalFunction string                `json:"grammatical_function,omitempty"`
	TokenIndex          int                   `json:"token_index"`
	Span                text.Span             `json:"span"`
	Surface             string                `json:"surface"`
	Reading             *morph.Reading        `json:"reading,omitempty"`
	Lexeme              *lexicon.MorphLexeme  `json:"lexeme,omitempty"`
	Sense               *lexicon.LexicalSense `json:"sense,omitempty"`
	ReadingLexemeID     string                `json:"reading_lexeme_id,omitempty"`
	LexemeID            string                `json:"lexeme_id,omitempty"`
	SenseID             string                `json:"sense_id,omitempty"`
	SenseLexemeID       string                `json:"sense_lexeme_id,omitempty"`
}

type DSERelationMetadataEvidence struct {
	Mapping           DSERelationPredicateMappingEvidence `json:"mapping"`
	Binding           lexicon.SemanticBinding             `json:"binding"`
	BindingEvidence   DSERelationBindingEvidence          `json:"binding_evidence"`
	Sense             lexicon.LexicalSense                `json:"sense"`
	SenseEvidence     DSERelationSenseEvidence            `json:"sense_evidence"`
	Lexeme            lexicon.MorphLexeme                 `json:"lexeme"`
	LexemeEvidence    DSERelationLexemeEvidence           `json:"lexeme_evidence"`
	Frame             valency.ValencyFrame                `json:"frame"`
	FrameEvidence     DSERelationFrameEvidence            `json:"frame_evidence"`
	Validation        language.TargetValidation           `json:"target_validation"`
	Signature         algebra.Signature                   `json:"signature"`
	HeadPort          algebra.Port                        `json:"head_port"`
	RolePorts         []DSERelationRolePortEvidence       `json:"role_ports"`
	PortRoles         []DSERelationPortRoleEvidence       `json:"port_roles"`
	SubjectRole       string                              `json:"subject_role"`
	ObjectRole        string                              `json:"object_role"`
	SourceProvenance  provenance.SourceRef                `json:"source_provenance"`
	BindingProvenance provenance.SourceRef                `json:"binding_provenance"`
	FrameProvenance   provenance.SourceRef                `json:"frame_provenance"`
}

// These DTOs expose stable string identities for records whose ID wrappers
// intentionally have unexported representations and therefore encode as {}.
type DSERelationPredicateMappingEvidence struct {
	SenseID            string                    `json:"sense_id"`
	FrameID            string                    `json:"frame_id"`
	HeadPortID         string                    `json:"head_port_id"`
	AllowedFamilyIDs   []string                  `json:"allowed_family_ids"`
	AllowedVariantIDs  []string                  `json:"allowed_variant_ids"`
	ReferenceCondition *language.SourceCondition `json:"reference_condition,omitempty"`
}

type DSERelationBindingEvidence struct {
	SenseID    string                  `json:"sense_id"`
	CoreRef    string                  `json:"core_ref"`
	Target     *lexicon.SemanticTarget `json:"target,omitempty"`
	Mapping    string                  `json:"mapping"`
	Status     string                  `json:"status"`
	Provenance provenance.SourceRef    `json:"provenance"`
}

type DSERelationSenseEvidence struct {
	ID         string   `json:"id"`
	LexemeID   string   `json:"lexeme_id"`
	Register   string   `json:"register,omitempty"`
	Usage      []string `json:"usage"`
	ValencyIDs []string `json:"valency_ids"`
}

type DSERelationLexemeEvidence struct {
	ID         string             `json:"id"`
	Language   string             `json:"language"`
	Lemma      string             `json:"lemma"`
	POS        string             `json:"pos"`
	Properties []lexicon.Property `json:"properties"`
	ParadigmID string             `json:"paradigm_id,omitempty"`
}

type DSERelationFrameEvidence struct {
	ID                     string                               `json:"id"`
	SenseID                string                               `json:"sense_id"`
	SemanticRoles          []string                             `json:"semantic_roles"`
	Arguments              []DSERelationFrameArgumentEvidence   `json:"arguments"`
	SelectionalConstraints []DSERelationFrameConstraintEvidence `json:"selectional_constraints"`
	AllowedFamilyIDs       []string                             `json:"allowed_family_ids"`
	Provenance             provenance.SourceRef                 `json:"provenance"`
}

type DSERelationFrameArgumentEvidence struct {
	ID                  string `json:"id"`
	Role                string `json:"role"`
	Required            bool   `json:"required"`
	GrammaticalFunction string `json:"grammatical_function,omitempty"`
	Case                string `json:"case,omitempty"`
	Adposition          string `json:"adposition,omitempty"`
}

type DSERelationFrameConstraintEvidence struct {
	ID                string   `json:"id"`
	Role              string   `json:"role"`
	Strength          string   `json:"strength"`
	AllowedConceptIDs []string `json:"allowed_concept_ids"`
	AllowedKinds      []string `json:"allowed_kinds"`
}

type DSERelationRolePortEvidence struct {
	Role   string       `json:"role"`
	PortID string       `json:"port_id"`
	Type   algebra.Type `json:"type"`
}

type DSERelationPortRoleEvidence struct {
	PortID string `json:"port_id"`
	Role   string `json:"role"`
}

// DSEScopedEvidence preserves the source-plan and lexical decisions used by
// scoped composition. It is provenance, not an owner judgment of language.
type DSEScopedEvidence struct {
	Variant          construction.ScopedClauseVariant `json:"variant"`
	Bindings         []DSEScopedBindingEvidence       `json:"bindings"`
	Frame            valency.ValencyFrame             `json:"frame"`
	SourceReferences language.SourceReferenceContext  `json:"source_references"`
	SubjectBinder    string                           `json:"subject_binder"`
	ObjectBinder     string                           `json:"object_binder"`
	SubjectRole      string                           `json:"subject_role"`
	ObjectRole       string                           `json:"object_role"`
	RolePorts        []DSEScopedRoleEvidence          `json:"role_ports"`
	Provenance       []provenance.SourceRef           `json:"provenance"`
}

type DSEScopedBindingEvidence struct {
	Binder     string               `json:"binder"`
	Case       string               `json:"case,omitempty"`
	TokenIndex int                  `json:"token_index"`
	Span       text.Span            `json:"span"`
	Reading    morph.Reading        `json:"reading"`
	Lexeme     lexicon.MorphLexeme  `json:"lexeme"`
	Sense      lexicon.LexicalSense `json:"sense"`
}

type DSEScopedRoleEvidence struct {
	FrameID string       `json:"frame_id"`
	Role    string       `json:"role"`
	PortID  string       `json:"port_id"`
	Type    algebra.Type `json:"type"`
}

// DSEClauseSource preserves the independently analyzed source provenance for
// one clause of a composed source coordination.
type DSEClauseSource struct {
	Source               DSESourceRecord `json:"source"`
	HeadEventReferenceID string          `json:"head_event_reference_id"`
	RoleSources          []DSERoleSource `json:"role_sources"`
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
