package realizer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	defaultPlanMaxFeatures    = 256
	maxPlanFeatures           = 4096
	defaultPlanMaxAssumptions = 256
	maxPlanAssumptions        = 4096
	defaultPlanMaxLosses      = 256
	maxPlanLosses             = 4096
	defaultPlanMaxTextBytes   = 1 << 20
	maxPlanTextBytes          = 4 << 20
	maxPlanFieldBytes         = 4096
)

// PlanBudget bounds the annotations retained alongside one selected target
// candidate. Meaning itself has an independent node, edge, and depth budget.
type PlanBudget struct {
	MaxFeatures    int
	MaxAssumptions int
	MaxLosses      int
	MaxTextBytes   int
	Meaning        meaning.Budget
}

// DefaultPlanBudget returns finite limits for one realization plan.
func DefaultPlanBudget() PlanBudget {
	return PlanBudget{
		MaxFeatures:    defaultPlanMaxFeatures,
		MaxAssumptions: defaultPlanMaxAssumptions,
		MaxLosses:      defaultPlanMaxLosses,
		MaxTextBytes:   defaultPlanMaxTextBytes,
		Meaning:        meaning.DefaultBudget(),
	}
}

// Validate rejects disabled or above-ceiling realization plan limits.
func (budget PlanBudget) Validate() error {
	if budget.MaxFeatures <= 0 || budget.MaxFeatures > maxPlanFeatures ||
		budget.MaxAssumptions <= 0 || budget.MaxAssumptions > maxPlanAssumptions ||
		budget.MaxLosses <= 0 || budget.MaxLosses > maxPlanLosses ||
		budget.MaxTextBytes <= 0 || budget.MaxTextBytes > maxPlanTextBytes {
		return planError(status.OutcomeInvalid, "INVALID_REALIZATION_PLAN_BUDGET", "budget", "realization plan limits must be positive and within package ceilings")
	}
	if err := budget.Meaning.Validate(); err != nil {
		return err
	}
	return nil
}

// FeatureConstraint records an explicitly supplied feature value associated
// with a path in Meaning. The planner never creates feature values from CORE
// concepts, lexical choices, or defaults.
type FeatureConstraint struct {
	Path    []meaning.NodeID        `json:"path"`
	Feature features.FeatureID      `json:"feature"`
	Value   features.FeatureValueID `json:"value"`
	Reason  string                  `json:"reason,omitempty"`
}

// FieldAssumption records a caller-declared assumption for one Meaning field.
// Any assumption makes the resulting plan incomplete.
type FieldAssumption struct {
	Path   []meaning.NodeID `json:"path"`
	Field  string           `json:"field"`
	Value  string           `json:"value"`
	Reason string           `json:"reason"`
}

// FieldLoss records information that cannot be represented by this plan.
// Its path must identify a real path through the supplied Meaning graph.
type FieldLoss struct {
	Path    []meaning.NodeID   `json:"path"`
	Field   string             `json:"field"`
	Outcome status.OutcomeKind `json:"outcome"`
	Reason  string             `json:"reason"`
}

// PlanInput contains one licensed target choice. Each call builds exactly one
// plan so competing selector candidates remain separate alternatives.
type PlanInput struct {
	Meaning            meaning.Meaning
	Candidate          TargetCandidate
	Core               *core.ConceptGraph
	FeatureConstraints []FeatureConstraint
	Assumptions        []FieldAssumption
	Losses             []FieldLoss
	Budget             PlanBudget
}

// RealizationPlan is a bounded target choice over one predicate node. It
// stores stable references and the explicit ordered construction composition,
// while leaving the Meaning graph with the caller to avoid duplicating it.
type RealizationPlan struct {
	ID                   string                          `json:"id"`
	SemanticID           string                          `json:"semantic_id"`
	SourceCandidateID    string                          `json:"source_candidate_id"`
	Language             features.LanguageID             `json:"language"`
	NodeID               meaning.NodeID                  `json:"node_id"`
	Kind                 meaning.NodeKind                `json:"kind"`
	Roles                []meaning.Role                  `json:"roles,omitempty"`
	MeaningPath          []meaning.NodeID                `json:"meaning_path"`
	CoreConceptID        ids.ConceptID                   `json:"core_concept_id"`
	LexemeID             ids.LexemeID                    `json:"lexeme_id"`
	SenseID              ids.SenseID                     `json:"sense_id"`
	FrameID              valency.FrameID                 `json:"frame_id"`
	FamilyID             ids.ConstructionID              `json:"family_id"`
	FamilyPlanID         construction.ConstructionPlanID `json:"family_plan_id"`
	VariantID            construction.VariantID          `json:"variant_id"`
	Construction         construction.ConstructionPlan   `json:"construction"`
	CompiledVariant      construction.CompiledVariant    `json:"compiled_variant,omitempty"`
	TransferProvenance   provenance.SourceRef            `json:"transfer_provenance"`
	FrameProvenance      provenance.SourceRef            `json:"frame_provenance,omitempty"`
	BindingProvenance    provenance.SourceRef            `json:"binding_provenance,omitempty"`
	FamilyProvenance     provenance.SourceRef            `json:"family_provenance,omitempty"`
	VariantProvenance    provenance.SourceRef            `json:"variant_provenance,omitempty"`
	OperatorProvenance   []provenance.SourceRef          `json:"operator_provenance,omitempty"`
	SoftConstraintMisses []valency.SoftConstraintMiss    `json:"soft_constraint_misses,omitempty"`
	FeatureConstraints   []FeatureConstraint             `json:"feature_constraints,omitempty"`
	Assumptions          []FieldAssumption               `json:"assumptions,omitempty"`
	Losses               []FieldLoss                     `json:"losses,omitempty"`
	Outcome              status.OutcomeKind              `json:"outcome"`
	Diagnostics          []status.Diagnostic             `json:"diagnostics,omitempty"`
}

// PlanError carries a typed outcome and bounded diagnostic for invalid or
// unsupported plan inputs.
type PlanError struct{ result status.Result }

func (err *PlanError) Error() string {
	if err == nil || len(err.result.Diagnostics) == 0 {
		return "realization plan error"
	}
	return err.result.Diagnostics[0].Message
}

// Result returns a defensive copy of the typed plan error.
func (err *PlanError) Result() status.Result {
	if err == nil {
		return status.Result{}
	}
	result := err.result
	result.Diagnostics = append([]status.Diagnostic(nil), err.result.Diagnostics...)
	return result
}

// BuildPlan validates Meaning and one selected target candidate, derives a
// deterministic source path to the predicate, and preserves only explicit
// feature constraints, assumptions, losses, and operator steps.
func BuildPlan(input PlanInput) (RealizationPlan, error) {
	budget := input.Budget
	if budget == (PlanBudget{}) {
		budget = DefaultPlanBudget()
	}
	if err := budget.Validate(); err != nil {
		return RealizationPlan{}, err
	}
	if input.Core == nil {
		return RealizationPlan{}, planError(status.OutcomeInvalid, "MISSING_CORE_GRAPH", "core", "a CORE graph is required to validate the supplied Meaning")
	}
	if err := meaning.ValidateMeaning(input.Meaning, input.Core, budget.Meaning); err != nil {
		return RealizationPlan{}, err
	}
	if !input.Candidate.Outcome.Valid() || len(input.Candidate.Diagnostics) > status.MaxDiagnostics {
		return RealizationPlan{}, planError(status.OutcomeInvalid, "INVALID_SELECTION_RESULT", "selection", "selection outcome or diagnostics are invalid")
	}
	for index, diagnostic := range input.Candidate.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return RealizationPlan{}, planError(status.OutcomeInvalid, "INVALID_SELECTION_DIAGNOSTIC", fmt.Sprintf("selection.diagnostics[%d]", index), err.Error())
		}
	}
	path, node, ok := pathToNode(input.Meaning, input.Candidate.NodeID, budget.Meaning.MaxDepth)
	if !ok {
		return RealizationPlan{}, planError(status.OutcomeUnknown, "TARGET_NODE_UNREACHABLE", "candidate.node_id", "selected target node is absent or unreachable from a Meaning root")
	}
	if node.Kind != meaning.KindEvent && node.Kind != meaning.KindState && node.Kind != meaning.KindReferent {
		return RealizationPlan{}, planError(status.OutcomeUnsupported, "TARGET_NODE_NOT_LEXICAL", "candidate.node_id", "target lexicalization requires an event, state, or referent node")
	}
	if err := validateTarget(input.Candidate, node, input.Meaning); err != nil {
		return RealizationPlan{}, err
	}
	if node.Kind == meaning.KindEvent || node.Kind == meaning.KindState {
		validation, err := valency.ValidateFrame(input.Candidate.Frame, valency.ValidationRequest{
			Meaning: input.Meaning, NodeID: node.ID,
			SemanticBinding: input.Candidate.TransferBinding, ConstructionBinding: input.Candidate.ConstructionBinding,
		}, input.Core)
		if err != nil {
			return RealizationPlan{}, err
		}
		if !sameSoftConstraintMisses(validation.SoftConstraintMisses, input.Candidate.SoftConstraintMisses) {
			return RealizationPlan{}, planError(status.OutcomeInvalid, "SOFT_CONSTRAINT_EVIDENCE_MISMATCH", "candidate.soft_constraint_misses", "candidate soft-constraint evidence does not match the validated Meaning and frame")
		}
	}
	autoLosses, err := unrepresentedWrapperLosses(input.Meaning, node.ID, input.Candidate.ConstructionPlan.Steps, budget.Meaning.MaxDepth, budget.MaxLosses-len(input.Losses))
	if err != nil {
		return RealizationPlan{}, err
	}
	losses := append(cloneLosses(input.Losses), autoLosses...)
	if err := validateAnnotations(input, budget); err != nil {
		return RealizationPlan{}, err
	}
	textBytes := 0
	account := func(values ...string) bool {
		for _, value := range values {
			textBytes += len(value)
			if textBytes > budget.MaxTextBytes {
				return false
			}
		}
		return true
	}
	candidate := input.Candidate
	if !account(candidate.SourceCandidateID, candidate.SemanticID, string(candidate.NodeID), string(candidate.Kind), candidate.Lexeme.ID.String(), candidate.Sense.ID.String(), candidate.TransferBinding.CoreRef.String(), string(candidate.Frame.ID), candidate.Family.Family.ID.String(), string(candidate.Family.PlanID), string(candidate.ConstructionPlan.ID), candidate.ConstructionPlan.OutputFamilyID.String(), string(candidate.CompiledVariant.PlanID), candidate.CompiledVariant.Variant.ID.String()) || !accountSource(account, candidate.TransferBinding.Provenance) {
		return RealizationPlan{}, planLimit("text_bytes", "selected target identity exceeds the plan text budget")
	}
	for _, id := range path {
		if !account(string(id)) {
			return RealizationPlan{}, planLimit("text_bytes", "Meaning path exceeds the plan text budget")
		}
	}
	for _, role := range candidate.Roles {
		if !account(string(role)) {
			return RealizationPlan{}, planLimit("text_bytes", "target roles exceed the plan text budget")
		}
	}
	for _, diagnostic := range candidate.Diagnostics {
		if !account(diagnostic.Code, diagnostic.Path, diagnostic.Message) {
			return RealizationPlan{}, planLimit("text_bytes", "selection diagnostics exceed the plan text budget")
		}
	}
	if node.Kind != meaning.KindReferent {
		for _, source := range []provenance.SourceRef{candidate.Frame.Provenance, candidate.ConstructionBinding.Provenance, candidate.Family.Family.Provenance, candidate.CompiledVariant.Variant.Provenance} {
			if !accountSource(account, source) {
				return RealizationPlan{}, planLimit("text_bytes", "license provenance exceeds the plan text budget")
			}
		}
	}
	for _, operator := range candidate.ConstructionPlan.Steps {
		if !account(string(operator.ID), string(operator.Kind), string(operator.Value), operator.InputFamilyID.String(), operator.OutputFamilyID.String(), string(operator.SemanticScope.TargetNodeID), string(operator.SemanticScope.ScopeNodeID), operator.Provenance.ID, operator.Provenance.Path, operator.Provenance.URI, operator.Provenance.Version, operator.Provenance.License, operator.Provenance.Author, operator.Provenance.ChecksumSHA256) {
			return RealizationPlan{}, planLimit("text_bytes", "construction operator metadata exceeds the plan text budget")
		}
		for _, dependency := range operator.Dependencies {
			if !account(string(dependency)) {
				return RealizationPlan{}, planLimit("text_bytes", "construction operator dependencies exceed the plan text budget")
			}
		}
		for _, capability := range operator.Capabilities {
			if !account(string(capability)) {
				return RealizationPlan{}, planLimit("text_bytes", "construction operator capabilities exceed the plan text budget")
			}
		}
	}
	for _, token := range candidate.CompiledVariant.Variant.Sequence {
		if !account(string(token.Kind), token.SlotID, token.Terminal) {
			return RealizationPlan{}, planLimit("text_bytes", "compiled variant exceeds the plan text budget")
		}
	}
	for _, tokens := range [][]construction.SequenceToken{candidate.CompiledVariant.FirstAllowed, candidate.CompiledVariant.LastAllowed} {
		for _, token := range tokens {
			if !account(string(token.Kind), token.SlotID, token.Terminal) {
				return RealizationPlan{}, planLimit("text_bytes", "compiled variant boundaries exceed the plan text budget")
			}
		}
	}
	for _, condition := range candidate.CompiledVariant.Variant.Conditions {
		if !account(string(condition.Kind), condition.SlotID, string(condition.Feature), string(condition.Value), string(condition.Role), string(condition.Info)) {
			return RealizationPlan{}, planLimit("text_bytes", "compiled variant conditions exceed the plan text budget")
		}
	}
	for _, miss := range candidate.SoftConstraintMisses {
		if !account(miss.ConstraintID, string(miss.Role), miss.Actual.String()) {
			return RealizationPlan{}, planLimit("text_bytes", "soft constraint evidence exceeds the plan text budget")
		}
	}
	for _, constraint := range input.FeatureConstraints {
		if !account(string(constraint.Feature), string(constraint.Value), constraint.Reason) || !accountNodePath(account, constraint.Path) {
			return RealizationPlan{}, planLimit("text_bytes", "feature constraints exceed the plan text budget")
		}
	}
	for _, assumption := range input.Assumptions {
		if !account(assumption.Field, assumption.Value, assumption.Reason) || !accountNodePath(account, assumption.Path) {
			return RealizationPlan{}, planLimit("text_bytes", "assumptions exceed the plan text budget")
		}
	}
	for _, loss := range input.Losses {
		if !account(loss.Field, loss.Reason) || !accountNodePath(account, loss.Path) {
			return RealizationPlan{}, planLimit("text_bytes", "loss records exceed the plan text budget")
		}
	}
	for _, loss := range autoLosses {
		if !account(loss.Field, loss.Reason) || !accountNodePath(account, loss.Path) {
			return RealizationPlan{}, planLimit("text_bytes", "unrepresented Meaning losses exceed the plan text budget")
		}
	}
	outcome := candidate.Outcome
	diagnostics := append([]status.Diagnostic(nil), candidate.Diagnostics...)
	if len(autoLosses) > 0 {
		outcome = mergePlanOutcome(outcome, status.OutcomeIncomplete)
		diagnostics = appendPlanDiagnostic(diagnostics, planDiagnostic("MEANING_OPERATOR_NOT_REALIZED", "Meaning ancestry contains fields without an explicit construction realization", "meaning_path"))
	}
	if len(input.Assumptions) > 0 {
		outcome = mergePlanOutcome(outcome, status.OutcomeIncomplete)
		diagnostics = appendPlanDiagnostic(diagnostics, planDiagnostic("EXPLICIT_REALIZATION_ASSUMPTION", "plan contains caller-declared assumptions", "assumptions"))
	}
	for _, loss := range input.Losses {
		outcome = mergePlanOutcome(outcome, loss.Outcome)
		diagnostics = appendPlanDiagnostic(diagnostics, planDiagnostic("REALIZATION_FIELD_LOSS", loss.Reason, loss.Field))
	}
	for _, loss := range autoLosses {
		outcome = mergePlanOutcome(outcome, loss.Outcome)
	}
	if len(diagnostics) > status.MaxDiagnostics {
		diagnostics = diagnostics[:status.MaxDiagnostics]
	}
	plan := RealizationPlan{
		SemanticID: candidate.SemanticID, SourceCandidateID: candidate.SourceCandidateID,
		Language: features.LanguageID(candidate.Lexeme.Language), NodeID: node.ID, Kind: node.Kind, Roles: sortedRoles(candidate.Roles), MeaningPath: path,
		CoreConceptID: candidate.TransferBinding.CoreRef, LexemeID: candidate.Lexeme.ID,
		SenseID:              candidate.Sense.ID,
		Construction:         cloneConstructionPlan(candidate.ConstructionPlan),
		CompiledVariant:      cloneCompiledVariant(candidate.CompiledVariant),
		TransferProvenance:   candidate.TransferBinding.Provenance,
		SoftConstraintMisses: append([]valency.SoftConstraintMiss(nil), candidate.SoftConstraintMisses...),
		FeatureConstraints:   cloneFeatureConstraints(input.FeatureConstraints),
		Assumptions:          cloneAssumptions(input.Assumptions), Losses: losses,
		Outcome: outcome,
	}
	if node.Kind != meaning.KindReferent {
		plan.FrameID = candidate.Frame.ID
		plan.FamilyID = candidate.Family.Family.ID
		plan.FamilyPlanID = candidate.Family.PlanID
		plan.VariantID = construction.VariantID(candidate.ConstructionBinding.VariantID.String())
		plan.FrameProvenance = candidate.Frame.Provenance
		plan.BindingProvenance = candidate.ConstructionBinding.Provenance
		plan.FamilyProvenance = candidate.Family.Family.Provenance
		plan.VariantProvenance = candidate.CompiledVariant.Variant.Provenance
		plan.OperatorProvenance = make([]provenance.SourceRef, len(candidate.ConstructionPlan.Steps))
		for index, operator := range candidate.ConstructionPlan.Steps {
			plan.OperatorProvenance[index] = operator.Provenance
		}
	}
	plan.Diagnostics = append([]status.Diagnostic(nil), diagnostics...)
	plan.ID = realizationPlanID(plan)
	return plan, nil
}

func validateTarget(candidate TargetCandidate, node meaning.Node, graph meaning.Meaning) error {
	if !validPlanField(candidate.SourceCandidateID) || !validPlanField(candidate.SemanticID) {
		return planError(status.OutcomeInvalid, "INVALID_TARGET_IDENTITY", "candidate", "selected target requires source and semantic candidate IDs")
	}
	if !candidate.Outcome.Valid() || (candidate.Outcome != status.OutcomeOK && len(candidate.Diagnostics) == 0) || len(candidate.Diagnostics) > status.MaxDiagnostics {
		return planError(status.OutcomeInvalid, "INVALID_TARGET_OUTCOME", "candidate.outcome", "target candidate must carry a valid outcome and diagnostics for non-ok results")
	}
	for index, diagnostic := range candidate.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return planError(status.OutcomeInvalid, "INVALID_TARGET_DIAGNOSTIC", fmt.Sprintf("candidate.diagnostics[%d]", index), err.Error())
		}
	}
	if candidate.Kind != node.Kind || candidate.Lexeme.ID.String() == "" || !validLanguage(candidate.Lexeme.Language) || candidate.Sense.ID.String() == "" || candidate.Sense.LexemeID != candidate.Lexeme.ID || candidate.TransferBinding.SenseID != candidate.Sense.ID || candidate.TransferBinding.CoreRef.String() != string(node.Concept) || !candidate.TransferBinding.TransferAuthorized() {
		return planError(status.OutcomeUnsupported, "TARGET_TRANSFER_MISMATCH", "candidate.transfer_binding", "selected lexeme, sense, and approved equivalent CORE transfer do not match the Meaning predicate")
	}
	if !sameRoles(candidate.Roles, node.ID, graph) {
		return planError(status.OutcomeInvalid, "TARGET_ROLE_MISMATCH", "candidate.roles", "target candidate roles do not match the selected Meaning node")
	}
	if node.Kind == meaning.KindReferent {
		if candidate.Frame.ID != "" || candidate.ConstructionBinding.SenseID.String() != "" || candidate.Family.Family.ID.String() != "" || candidate.ConstructionPlan.ID != "" || candidate.CompiledVariant.Variant.ID != "" || len(candidate.ConstructionPlan.Steps) != 0 || len(candidate.CompiledVariant.Variant.Sequence) != 0 || len(candidate.CompiledVariant.FirstAllowed) != 0 || len(candidate.CompiledVariant.LastAllowed) != 0 {
			return planError(status.OutcomeInvalid, "REFERENT_HAS_PREDICATE_LICENSE", "candidate", "referent candidates must not carry predicate frame, family, variant, or composition licenses")
		}
		return nil
	}
	if candidate.Frame.ID == "" || candidate.Frame.SenseID != candidate.Sense.ID || candidate.ConstructionBinding.SenseID != candidate.Sense.ID || candidate.ConstructionBinding.ValencyFrameID != candidate.Frame.ID || !candidate.ConstructionBinding.Approved {
		return planError(status.OutcomeUnsupported, "TARGET_FRAME_MISMATCH", "candidate.frame", "selected frame is not the approved frame for the target sense")
	}
	if len(candidate.SoftConstraintMisses) > valency.MaxConstraintsPerFrame {
		return planError(status.OutcomeResourceLimit, "SOFT_CONSTRAINT_EVIDENCE_LIMIT", "candidate.soft_constraint_misses", "soft constraint evidence exceeds the frame limit")
	}
	family := candidate.Family.Family
	if family.ID.String() == "" || family.ID != candidate.ConstructionBinding.FamilyID || family.Language != features.LanguageID(candidate.Lexeme.Language) || !candidate.Family.PlanID.Valid() || candidate.CompiledVariant.Variant.ID == "" || candidate.CompiledVariant.Variant.FamilyID != family.ID || candidate.CompiledVariant.FamilyID != family.ID || candidate.CompiledVariant.PlanID != candidate.Family.PlanID || (candidate.CompiledVariant.Variant.DirectionPolicy != construction.DirectionBoth && candidate.CompiledVariant.Variant.DirectionPolicy != construction.DirectionRealize) || candidate.ConstructionBinding.VariantID.String() != candidate.CompiledVariant.Variant.ID.String() || len(candidate.CompiledVariant.Variant.Sequence) > construction.MaxSequenceTokens || len(candidate.CompiledVariant.Variant.Conditions) > construction.MaxConditions || len(candidate.CompiledVariant.FirstAllowed) > construction.MaxSequenceTokens || len(candidate.CompiledVariant.LastAllowed) > construction.MaxSequenceTokens {
		return planError(status.OutcomeUnsupported, "TARGET_CONSTRUCTION_MISMATCH", "candidate.family", "selected family, variant, and approved construction binding do not match")
	}
	composition := candidate.ConstructionPlan
	if !composition.ID.Valid() || composition.SenseID != candidate.Sense.ID || composition.ValencyFrameID != candidate.Frame.ID || composition.InputFamilyID != family.ID || composition.OutputFamilyID == (ids.ConstructionID{}) || composition.Language != family.Language || len(composition.Steps) > construction.MaxOperatorsPerPlan {
		return planError(status.OutcomeUnsupported, "INVALID_COMPOSITION_PLAN", "candidate.construction_plan", "explicit construction composition does not match the selected target license")
	}
	currentFamily := composition.InputFamilyID
	for index, operator := range composition.Steps {
		if err := operator.Validate(); err != nil {
			return err
		}
		if operator.InputFamilyID != currentFamily || operator.SemanticScope.TargetNodeID != node.ID {
			return planError(status.OutcomeUnsupported, "INVALID_COMPOSITION_STEP", fmt.Sprintf("candidate.construction_plan.steps[%d]", index), "operator chain or explicit semantic target does not match this predicate")
		}
		if !operatorMatchesMeaning(operator, graph) {
			return planError(status.OutcomeUnsupported, "INVALID_COMPOSITION_SCOPE", fmt.Sprintf("candidate.construction_plan.steps[%d].semantic_scope", index), "operator scope does not exactly match a supported Meaning wrapper and value")
		}
		currentFamily = operator.OutputFamilyID
	}
	if composition.OutputFamilyID != currentFamily {
		return planError(status.OutcomeInvalid, "INVALID_COMPOSITION_OUTPUT", "candidate.construction_plan.output_family_id", "construction plan output does not match its ordered operator steps")
	}
	if candidate.CompiledVariant.Variant.Provenance.Validate() != nil || candidate.ConstructionBinding.Provenance.Validate() != nil || candidate.Frame.Provenance.Validate() != nil || candidate.Family.Family.Provenance.Validate() != nil {
		return planError(status.OutcomeInvalid, "INVALID_LICENSE_PROVENANCE", "candidate", "predicate license records require valid provenance")
	}
	compiled, err := construction.ValidateVariantForFrame(candidate.CompiledVariant.Variant, family, candidate.Frame)
	if err != nil {
		return err
	}
	if compiled.PlanID != candidate.CompiledVariant.PlanID || compiled.FamilyID != candidate.CompiledVariant.FamilyID || !sameSequence(compiled.FirstAllowed, candidate.CompiledVariant.FirstAllowed) || !sameSequence(compiled.LastAllowed, candidate.CompiledVariant.LastAllowed) {
		return planError(status.OutcomeInvalid, "INVALID_COMPILED_VARIANT", "candidate.compiled_variant", "compiled variant boundary metadata does not match its validated sequence")
	}
	return nil
}

func validateAnnotations(input PlanInput, budget PlanBudget) error {
	if len(input.FeatureConstraints) > budget.MaxFeatures || len(input.Assumptions) > budget.MaxAssumptions || len(input.Losses) > budget.MaxLosses {
		return planLimit("annotations", "feature constraints, assumptions, or losses exceed their configured limits")
	}
	for index, constraint := range input.FeatureConstraints {
		if !validPlanField(string(constraint.Feature)) || !validPlanField(string(constraint.Value)) || !validPlanText(constraint.Reason, maxPlanFieldBytes) {
			return planError(status.OutcomeInvalid, "INVALID_FEATURE_CONSTRAINT", fmt.Sprintf("feature_constraints[%d]", index), "feature constraint requires bounded explicit feature/value identities")
		}
		if err := validateMeaningPath(input.Meaning, constraint.Path, budget.Meaning.MaxDepth); err != nil {
			return planError(status.OutcomeInvalid, "INVALID_FEATURE_CONSTRAINT_PATH", fmt.Sprintf("feature_constraints[%d].path", index), err.Error())
		}
	}
	for index, assumption := range input.Assumptions {
		if !validPlanField(assumption.Field) || !validPlanField(assumption.Value) || assumption.Reason == "" || !validPlanText(assumption.Reason, maxPlanFieldBytes) {
			return planError(status.OutcomeInvalid, "INVALID_FIELD_ASSUMPTION", fmt.Sprintf("assumptions[%d]", index), "assumption requires bounded field, explicit value, and reason")
		}
		if err := validateMeaningPath(input.Meaning, assumption.Path, budget.Meaning.MaxDepth); err != nil {
			return planError(status.OutcomeInvalid, "INVALID_FIELD_ASSUMPTION_PATH", fmt.Sprintf("assumptions[%d].path", index), err.Error())
		}
	}
	for index, loss := range input.Losses {
		if !validPlanField(loss.Field) || loss.Reason == "" || !validPlanText(loss.Reason, maxPlanFieldBytes) || !loss.Outcome.Valid() || loss.Outcome == status.OutcomeOK {
			return planError(status.OutcomeInvalid, "INVALID_FIELD_LOSS", fmt.Sprintf("losses[%d]", index), "loss requires a bounded field, non-ok outcome, and reason")
		}
		if err := validateMeaningPath(input.Meaning, loss.Path, budget.Meaning.MaxDepth); err != nil {
			return planError(status.OutcomeInvalid, "INVALID_FIELD_LOSS_PATH", fmt.Sprintf("losses[%d].path", index), err.Error())
		}
	}
	return nil
}

func pathToNode(graph meaning.Meaning, target meaning.NodeID, maxDepth int) ([]meaning.NodeID, meaning.Node, bool) {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	if _, exists := nodes[target]; !exists {
		return nil, meaning.Node{}, false
	}
	roots := append([]meaning.NodeID(nil), graph.Roots...)
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	parents := make(map[meaning.NodeID]meaning.NodeID, len(nodes))
	depths := make(map[meaning.NodeID]int, len(nodes))
	queue := make([]meaning.NodeID, 0, len(roots))
	for _, root := range roots {
		if _, exists := depths[root]; exists {
			continue
		}
		depths[root] = 1
		queue = append(queue, root)
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if current == target {
			path := []meaning.NodeID{target}
			for {
				parent, exists := parents[current]
				if !exists {
					break
				}
				path = append(path, parent)
				current = parent
			}
			for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
				path[left], path[right] = path[right], path[left]
			}
			if len(path) <= maxDepth {
				return path, nodes[target], true
			}
			return nil, meaning.Node{}, false
		}
		if depths[current] >= maxDepth {
			continue
		}
		children := meaningTargets(nodes[current])
		sort.Slice(children, func(i, j int) bool { return children[i] < children[j] })
		for _, child := range children {
			if _, exists := depths[child]; exists {
				continue
			}
			parents[child] = current
			depths[child] = depths[current] + 1
			queue = append(queue, child)
		}
	}
	return nil, meaning.Node{}, false
}

func validateMeaningPath(graph meaning.Meaning, path []meaning.NodeID, maxDepth int) error {
	if len(path) == 0 || len(path) > maxDepth {
		return fmt.Errorf("path must contain between 1 and %d nodes", maxDepth)
	}
	known := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		known[node.ID] = node
	}
	root := false
	for _, candidate := range graph.Roots {
		if candidate == path[0] {
			root = true
			break
		}
	}
	if !root {
		return fmt.Errorf("path must start at a Meaning root")
	}
	for index, id := range path {
		if _, exists := known[id]; !exists {
			return fmt.Errorf("path node %q does not exist", id)
		}
		if index == 0 {
			continue
		}
		if !containsNodeID(meaningTargets(known[path[index-1]]), id) {
			return fmt.Errorf("path edge %q to %q does not exist", path[index-1], id)
		}
	}
	return nil
}

func meaningTargets(node meaning.Node) []meaning.NodeID {
	result := []meaning.NodeID{}
	switch node.Kind {
	case meaning.KindEvent, meaning.KindState:
		for _, participant := range node.Participants {
			result = append(result, participant.Target)
		}
	case meaning.KindQuestion, meaning.KindScope, meaning.KindModality, meaning.KindQuantifier:
		if node.Target != "" {
			result = append(result, node.Target)
		}
	case meaning.KindDiscourse:
		result = append(result, node.Content...)
	}
	return result
}

func containsNodeID(values []meaning.NodeID, wanted meaning.NodeID) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func sameRoles(roles []meaning.Role, target meaning.NodeID, graph meaning.Meaning) bool {
	want := []meaning.Role{}
	for _, node := range graph.Nodes {
		if node.Kind != meaning.KindEvent && node.Kind != meaning.KindState {
			continue
		}
		for _, participant := range node.Participants {
			if participant.Target == target {
				want = append(want, participant.Role)
			}
		}
	}
	sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
	got := append([]meaning.Role(nil), roles...)
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func sortedRoles(roles []meaning.Role) []meaning.Role {
	copy := append([]meaning.Role(nil), roles...)
	sort.Slice(copy, func(i, j int) bool { return copy[i] < copy[j] })
	return copy
}

func sameSequence(left, right []construction.SequenceToken) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameSoftConstraintMisses(left, right []valency.SoftConstraintMiss) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func accountNodePath(account func(...string) bool, path []meaning.NodeID) bool {
	for _, id := range path {
		if !account(string(id)) {
			return false
		}
	}
	return true
}

func unrepresentedWrapperLosses(graph meaning.Meaning, target meaning.NodeID, operators []construction.Operator, maxDepth, maxLosses int) ([]FieldLoss, error) {
	fields := make(map[meaning.NodeID]string)
	for _, node := range meaningAncestors(graph.Nodes, target) {
		field := ""
		switch node.Kind {
		case meaning.KindQuestion, meaning.KindScope, meaning.KindModality:
			if !exactWrapperOperator(node, target, operators) {
				field = "meaning." + string(node.Kind)
			}
		case meaning.KindQuantifier:
			field = "meaning.quantifier"
		case meaning.KindDiscourse:
			if node.Act != meaning.ActStatement {
				field = "meaning.discourse_act"
			}
		}
		if field != "" {
			fields[node.ID] = field
			if len(fields) > maxLosses {
				return nil, planLimit("losses", "unrepresented Meaning operators exceed the configured loss limit")
			}
		}
	}
	paths := pathsToNodes(graph, fields, maxDepth)
	losses := make([]FieldLoss, 0, len(fields))
	for nodeID, field := range fields {
		path, exists := paths[nodeID]
		if !exists {
			return nil, planError(status.OutcomeInvalid, "MEANING_LOSS_PATH_UNAVAILABLE", "meaning", "validated Meaning loss node has no bounded path from a root")
		}
		losses = append(losses, FieldLoss{Path: path, Field: field, Outcome: status.OutcomeIncomplete, Reason: "Meaning field has no explicit licensed target construction mapping"})
	}
	sort.Slice(losses, func(i, j int) bool {
		if losses[i].Path[len(losses[i].Path)-1] != losses[j].Path[len(losses[j].Path)-1] {
			return losses[i].Path[len(losses[i].Path)-1] < losses[j].Path[len(losses[j].Path)-1]
		}
		return losses[i].Field < losses[j].Field
	})
	return losses, nil
}

func pathsToNodes(graph meaning.Meaning, targets map[meaning.NodeID]string, maxDepth int) map[meaning.NodeID][]meaning.NodeID {
	nodes := make(map[meaning.NodeID]meaning.Node, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	roots := append([]meaning.NodeID(nil), graph.Roots...)
	sort.Slice(roots, func(i, j int) bool { return roots[i] < roots[j] })
	parents := make(map[meaning.NodeID]meaning.NodeID, len(nodes))
	depths := make(map[meaning.NodeID]int, len(nodes))
	queue := make([]meaning.NodeID, 0, len(roots))
	for _, root := range roots {
		if _, exists := depths[root]; exists {
			continue
		}
		depths[root] = 1
		queue = append(queue, root)
	}
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if depths[current] >= maxDepth {
			continue
		}
		children := meaningTargets(nodes[current])
		sort.Slice(children, func(i, j int) bool { return children[i] < children[j] })
		for _, child := range children {
			if _, exists := depths[child]; exists {
				continue
			}
			parents[child] = current
			depths[child] = depths[current] + 1
			queue = append(queue, child)
		}
	}
	paths := make(map[meaning.NodeID][]meaning.NodeID, len(targets))
	for target := range targets {
		if depths[target] == 0 || depths[target] > maxDepth {
			continue
		}
		path := []meaning.NodeID{target}
		for current := target; ; {
			parent, exists := parents[current]
			if !exists {
				break
			}
			path = append(path, parent)
			current = parent
		}
		for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
			path[left], path[right] = path[right], path[left]
		}
		paths[target] = path
	}
	return paths
}

func exactWrapperOperator(node meaning.Node, target meaning.NodeID, operators []construction.Operator) bool {
	// Compose's scope contract links each mapped wrapper directly to its
	// event/state target. Nested outer wrappers therefore remain explicit loss.
	if node.Target != target {
		return false
	}
	matches := 0
	for _, operator := range operators {
		if operatorMatchesWrapper(operator, node, target) {
			matches++
		}
	}
	return matches == 1
}

func operatorMatchesMeaning(operator construction.Operator, graph meaning.Meaning) bool {
	if operator.SemanticScope.ScopeNodeID == "" {
		if operator.Kind != construction.OperatorTense && operator.Kind != construction.OperatorVoice {
			return false
		}
		for _, node := range graph.Nodes {
			if node.ID == operator.SemanticScope.TargetNodeID {
				return node.Kind == meaning.KindEvent
			}
		}
		return false
	}
	for _, node := range graph.Nodes {
		if node.ID == operator.SemanticScope.ScopeNodeID {
			return operatorMatchesWrapper(operator, node, operator.SemanticScope.TargetNodeID)
		}
	}
	return false
}

func operatorMatchesWrapper(operator construction.Operator, node meaning.Node, target meaning.NodeID) bool {
	if operator.SemanticScope.ScopeNodeID != node.ID || operator.SemanticScope.TargetNodeID != target || node.Target != target {
		return false
	}
	switch node.Kind {
	case meaning.KindScope:
		switch node.Operator {
		case meaning.OperatorNegation:
			return operator.Kind == construction.OperatorNegation && operator.Value == construction.ValueNegated
		case meaning.OperatorQuestion:
			return operator.Kind == construction.OperatorQuestion && (operator.Value == construction.ValueYesNo || operator.Value == construction.ValueWH)
		}
	case meaning.KindQuestion:
		return operator.Kind == construction.OperatorQuestion && (operator.Value == construction.ValueYesNo || operator.Value == construction.ValueWH)
	case meaning.KindModality:
		return operator.Kind == construction.OperatorModality && string(operator.Value) == string(node.Modality)
	}
	return false
}

func cloneConstructionPlan(plan construction.ConstructionPlan) construction.ConstructionPlan {
	copy := plan
	copy.Steps = make([]construction.Operator, len(plan.Steps))
	for index, operator := range plan.Steps {
		copy.Steps[index] = operator
		copy.Steps[index].Dependencies = append([]construction.OperatorID(nil), operator.Dependencies...)
		copy.Steps[index].Capabilities = append([]construction.Capability(nil), operator.Capabilities...)
	}
	return copy
}

func cloneCompiledVariant(variant construction.CompiledVariant) construction.CompiledVariant {
	copy := variant
	copy.Variant.Sequence = append([]construction.SequenceToken(nil), variant.Variant.Sequence...)
	copy.Variant.Conditions = append([]construction.Condition(nil), variant.Variant.Conditions...)
	copy.FirstAllowed = append([]construction.SequenceToken(nil), variant.FirstAllowed...)
	copy.LastAllowed = append([]construction.SequenceToken(nil), variant.LastAllowed...)
	return copy
}

func cloneFeatureConstraints(values []FeatureConstraint) []FeatureConstraint {
	result := make([]FeatureConstraint, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Path = append([]meaning.NodeID(nil), value.Path...)
	}
	return result
}

func cloneAssumptions(values []FieldAssumption) []FieldAssumption {
	result := make([]FieldAssumption, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Path = append([]meaning.NodeID(nil), value.Path...)
	}
	return result
}

func cloneLosses(values []FieldLoss) []FieldLoss {
	result := make([]FieldLoss, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Path = append([]meaning.NodeID(nil), value.Path...)
	}
	return result
}

func realizationPlanID(plan RealizationPlan) string {
	plan.ID = ""
	canonical, _ := json.Marshal(plan)
	sum := sha256.Sum256(canonical)
	return "realization:" + hex.EncodeToString(sum[:])
}

func accountSource(account func(...string) bool, source provenance.SourceRef) bool {
	return account(source.ID, source.Path, source.URI, source.Version, source.License, source.Author, source.ChecksumSHA256)
}

func validPlanField(value string) bool {
	if value == "" || len(value) > maxPlanFieldBytes || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func validPlanText(value string, maximum int) bool {
	if value == "" {
		return true
	}
	if len(value) > maximum || !utf8.ValidString(value) || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func planError(kind status.OutcomeKind, code, path, message string) error {
	return &PlanError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}}
}

func planLimit(path, message string) error {
	return planError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget."+path, message)
}

func planDiagnostic(code, message, path string) status.Diagnostic {
	return status.Diagnostic{Code: code, Message: message, Path: path, Severity: status.SeverityWarning}
}

func appendPlanDiagnostic(values []status.Diagnostic, value status.Diagnostic) []status.Diagnostic {
	if len(values) >= status.MaxDiagnostics {
		return values
	}
	return append(values, value)
}

func mergePlanOutcome(left, right status.OutcomeKind) status.OutcomeKind {
	if left == right {
		return left
	}
	priority := map[status.OutcomeKind]int{
		status.OutcomeOK: 0, status.OutcomeUnsupported: 1, status.OutcomeUnknown: 2,
		status.OutcomeAmbiguous: 3, status.OutcomeIncomplete: 4,
		status.OutcomeResourceLimit: 7, status.OutcomeError: 8, status.OutcomeInvalid: 9,
	}
	if priority[right] > priority[left] {
		return right
	}
	return left
}
