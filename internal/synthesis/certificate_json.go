package synthesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxSynthesisCertificateBytes = 8 << 20
	maxCertificateChoices        = 64
	maxCertificateOperations     = 64
	maxCertificateScopes         = 4096
	maxCertificateObligations    = 32
	maxCertificateAssumptions    = 3
	maxCertificateInterface      = 16
	maxCertificateListItems      = 16
	maxCertificatePathBytes      = 1024
)

// DecodeSynthesisCertificate strictly decodes a bounded certificate envelope.
// It does not verify producer claims; use VerifyGoalCertificate for that.
func DecodeSynthesisCertificate(data []byte) (SynthesisCertificate, error) {
	if err := algebra.PreflightJSONLimit(data, MaxSynthesisCertificateBytes); err != nil {
		return SynthesisCertificate{}, err
	}
	root, err := certObject(data, "certificate", certFields(
		"schema_version", "format", "producer_status", "goal_verified", "origin", "goal_fingerprint", "snapshot_id", "context_id", "profile", "concrete_goal", "choices", "candidate_fingerprint", "interface", "operations", "scopes", "obligations", "assumptions",
	), []string{"schema_version", "format", "producer_status", "goal_verified", "origin", "goal_fingerprint", "snapshot_id", "context_id", "profile", "concrete_goal", "choices", "candidate_fingerprint", "interface", "operations", "scopes", "obligations", "assumptions"})
	if err != nil {
		return SynthesisCertificate{}, err
	}
	var result SynthesisCertificate
	if result.SchemaVersion, err = certInt(root["schema_version"], "certificate.schema_version", 1, 1); err != nil {
		return result, err
	}
	if result.Format, err = certString(root["format"], "certificate.format", 256, false); err != nil {
		return result, err
	}
	if result.Format != SynthesisCertificateFormat {
		return result, certJSONInvalid("INVALID_CERTIFICATE_FORMAT", "certificate.format", "format must be finite-goal-certificate-v1")
	}
	if result.ProducerStatus, err = certString(root["producer_status"], "certificate.producer_status", 256, false); err != nil {
		return result, err
	}
	if result.GoalVerified, err = certBool(root["goal_verified"], "certificate.goal_verified"); err != nil {
		return result, err
	}
	if result.Origin, err = certString(root["origin"], "certificate.origin", 256, false); err != nil {
		return result, err
	}
	if result.Origin != CertificateOriginConstruct && result.Origin != CertificateOriginSearch {
		return result, certJSONInvalid("INVALID_CERTIFICATE_ORIGIN", "certificate.origin", "origin must be construct or search")
	}
	if result.GoalFingerprint, err = certString(root["goal_fingerprint"], "certificate.goal_fingerprint", 256, false); err != nil {
		return result, err
	}
	if result.SnapshotID, err = certString(root["snapshot_id"], "certificate.snapshot_id", 256, false); err != nil {
		return result, err
	}
	if result.ContextID, err = certString(root["context_id"], "certificate.context_id", 256, false); err != nil {
		return result, err
	}
	if result.Profile, err = certString(root["profile"], "certificate.profile", 256, false); err != nil {
		return result, err
	}
	if result.CandidateFingerprint, err = certString(root["candidate_fingerprint"], "certificate.candidate_fingerprint", 256, false); err != nil {
		return result, err
	}
	goal, err := DecodeGoal(root["concrete_goal"])
	if err != nil {
		return result, err
	}
	result.ConcreteGoal = goal
	if result.Choices, err = decodeCertificateArray(root["choices"], "certificate.choices", maxCertificateChoices, decodeCertificateChoice); err != nil {
		return result, err
	}
	if result.Interface, err = decodeCertificateArray(root["interface"], "certificate.interface", maxCertificateInterface, decodeCertificatePortBinding); err != nil {
		return result, err
	}
	if result.Operations, err = decodeCertificateArray(root["operations"], "certificate.operations", maxCertificateOperations, decodeCertifiedOperation); err != nil {
		return result, err
	}
	if result.Scopes, err = decodeCertificateArray(root["scopes"], "certificate.scopes", maxCertificateScopes, decodeCertificateScope); err != nil {
		return result, err
	}
	if result.Obligations, err = decodeCertificateArray(root["obligations"], "certificate.obligations", maxCertificateObligations, decodeCertificateObligation); err != nil {
		return result, err
	}
	if result.Assumptions, err = decodeCertificateArray(root["assumptions"], "certificate.assumptions", maxCertificateAssumptions, decodeCertificateAssumption); err != nil {
		return result, err
	}
	return result, nil
}

// EncodeSynthesisCertificate emits the strict bounded wire form. Encoding
// checks representation and resource bounds only, not certificate validity.
func EncodeSynthesisCertificate(certificate SynthesisCertificate) ([]byte, error) {
	if err := validateCertificateValue(certificate); err != nil {
		return nil, err
	}
	copy := normalizeCertificate(certificate)
	data, err := json.Marshal(copy)
	if err != nil {
		return nil, certJSONInvalid("CERTIFICATE_ENCODING", "certificate", err.Error())
	}
	if len(data) > MaxSynthesisCertificateBytes {
		return nil, certJSONResource("certificate", "encoded certificate exceeds the 8 MiB limit")
	}
	// A final strict decode keeps the encoder and decoder wire contracts aligned.
	if _, err := DecodeSynthesisCertificate(data); err != nil {
		return nil, err
	}
	return data, nil
}

func validateCertificateValue(c SynthesisCertificate) error {
	if c.SchemaVersion != SynthesisCertificateSchemaVersion {
		return certJSONInvalid("INVALID_SCHEMA_VERSION", "certificate.schema_version", "schema version must be 1")
	}
	if c.Format == "" || c.ProducerStatus == "" || c.Origin == "" || c.GoalFingerprint == "" || c.SnapshotID == "" || c.ContextID == "" || c.Profile == "" || c.CandidateFingerprint == "" {
		return certJSONInvalid("MISSING_CERTIFICATE_FIELD", "certificate", "required certificate identifiers must be nonempty")
	}
	for name, value := range map[string]string{
		"format": c.Format, "producer_status": c.ProducerStatus, "origin": c.Origin,
		"goal_fingerprint": c.GoalFingerprint, "snapshot_id": c.SnapshotID,
		"context_id": c.ContextID, "profile": c.Profile, "candidate_fingerprint": c.CandidateFingerprint,
	} {
		if err := certText(value, "certificate."+name, 256, false); err != nil {
			return err
		}
	}
	if c.IdentityProof != nil || c.IdentityKey != "" {
		return certJSONUnsupported("UNSUPPORTED_IDENTITY_PROOF", "certificate.identity_proof", "identity proof fields are not part of the finite-goal certificate format")
	}
	if len(c.Choices) > maxCertificateChoices || len(c.Interface) > maxCertificateInterface || len(c.Operations) > maxCertificateOperations || len(c.Scopes) > maxCertificateScopes || len(c.Obligations) > maxCertificateObligations || len(c.Assumptions) > maxCertificateAssumptions {
		return certJSONResource("certificate", "certificate collection exceeds its item limit")
	}
	if err := validateGoalWireValue(c.ConcreteGoal); err != nil {
		return err
	}
	for i, ch := range c.Choices {
		p := fmt.Sprintf("certificate.choices[%d]", i)
		if err := certText(ch.Kind, p+".kind", 256, false); err != nil {
			return err
		}
		if err := certText(ch.AxisID, p+".axis_id", 256, false); err != nil {
			return err
		}
		if err := boundedStrings(p, map[string]string{"kind": ch.Kind, "axis_id": ch.AxisID, "constraint_id": ch.ConstraintID, "symbol": ch.Symbol, "from_role": ch.FromRole, "to_role": ch.ToRole, "port_id": ch.PortID, "unary_role": ch.UnaryRole, "hole_id": ch.HoleID, "reference_id": ch.ReferenceID}); err != nil {
			return err
		}
		if !oneOf(ch.Kind, "path_signature", "path_length", "unary", "hole_binding") {
			return certJSONInvalid("INVALID_CHOICE_KIND", p+".kind", "choice kind is unknown")
		}
		if ch.AtomCount < 0 || ch.AtomCount > MaxGoalAtoms {
			return certJSONInvalid("INVALID_INTEGER", p+".atom_count", "atom count is outside the bounded range")
		}
	}
	for i, b := range c.Interface {
		if err := certText(b.Port, fmt.Sprintf("certificate.interface[%d].port", i), 256, false); err != nil {
			return err
		}
		if err := certText(b.Binder, fmt.Sprintf("certificate.interface[%d].binder", i), 256, false); err != nil {
			return err
		}
		if err := boundedStrings(fmt.Sprintf("certificate.interface[%d]", i), map[string]string{"port": b.Port, "binder": b.Binder}); err != nil {
			return err
		}
	}
	for i, o := range c.Operations {
		if err := validateCertifiedOperation(o, fmt.Sprintf("certificate.operations[%d]", i)); err != nil {
			return err
		}
	}
	for i, s := range c.Scopes {
		p := fmt.Sprintf("certificate.scopes[%d]", i)
		if err := certText(s.BinderID, p+".binder_id", 256, false); err != nil {
			return err
		}
		if err := certText(s.Path, p+".path", maxCertificatePathBytes, false); err != nil {
			return err
		}
		if s.ParentBinderID != "" {
			if err := certText(s.ParentBinderID, p+".parent_binder_id", 256, false); err != nil {
				return err
			}
		}
		if s.Kind != algebra.TermLambda && s.Kind != algebra.TermExists {
			return certJSONInvalid("INVALID_SCOPE_KIND", p+".kind", "scope kind must be Lambda or Exists")
		}
		if err := boundedStrings(p, map[string]string{"binder_id": s.BinderID, "parent_binder_id": s.ParentBinderID}); err != nil {
			return err
		}
		if len(s.Path) > maxCertificatePathBytes {
			return certJSONResource(p+".path", "scope path exceeds byte limit")
		}
		if err := validateCertType(s.Type, p+".type", 0, map[*algebra.Type]bool{}); err != nil {
			return err
		}
	}
	for i, o := range c.Obligations {
		p := fmt.Sprintf("certificate.obligations[%d]", i)
		if err := certText(o.ConstraintID, p+".constraint_id", 256, false); err != nil {
			return err
		}
		if err := certText(string(o.Kind), p+".kind", 32, false); err != nil {
			return err
		}
		if o.Kind != ConstraintPath && o.Kind != ConstraintUnary {
			return certJSONInvalid("INVALID_CONSTRAINT_KIND", p+".kind", "constraint kind is unknown")
		}
		if err := certText(o.Status, p+".status", 256, false); err != nil {
			return err
		}
		if err := boundedStrings(p, map[string]string{"constraint_id": o.ConstraintID, "status": o.Status}); err != nil {
			return err
		}
		if len(o.StepIDs) > maxCertificateOperations || len(o.Fragments) > 16 {
			return certJSONResource(p, "obligation list exceeds limit")
		}
		for j, v := range o.StepIDs {
			if err := certText(v, p+fmt.Sprintf(".step_ids[%d]", j), 256, false); err != nil {
				return err
			}
		}
		for j, v := range o.Fragments {
			if err := certText(v, p+fmt.Sprintf(".fragments[%d]", j), maxCertificatePathBytes, false); err != nil {
				return err
			}
		}
	}
	for i, a := range c.Assumptions {
		p := fmt.Sprintf("certificate.assumptions[%d]", i)
		if err := certText(a.Kind, p+".kind", 256, false); err != nil {
			return err
		}
		if !oneOf(a.Kind, CertificateAssumptionRegistrySnapshot, CertificateAssumptionRequestReferenceBound, CertificateAssumptionPositiveProfile) {
			return certJSONInvalid("INVALID_ASSUMPTION_KIND", p+".kind", "assumption kind is unknown")
		}
		if err := certText(a.Value, p+".value", 256, false); err != nil {
			return err
		}
		if err := boundedStrings(p, map[string]string{"kind": a.Kind, "value": a.Value}); err != nil {
			return err
		}
	}
	return nil
}

func validateCertifiedOperation(o CertifiedOperation, p string) error {
	if o.CheckedType == nil {
		return certJSONInvalid("MISSING_TYPE", p+".checked_type", "checked_type is required")
	}
	for name, value := range map[string]string{"id": o.ID, "rule_id": o.RuleID, "op": o.Op} {
		if err := certText(value, p+"."+name, 256, false); err != nil {
			return err
		}
	}
	if !validCertificateOperation(o.Op, o.RuleID) {
		return certJSONInvalid("INVALID_OPERATION_RULE", p, "operation and versioned rule_id do not match the closed certificate vocabulary")
	}
	if err := boundedStrings(p, map[string]string{"id": o.ID, "rule_id": o.RuleID, "op": o.Op, "symbol": o.Symbol, "port": o.Port, "reference_id": o.ReferenceID, "goal_port_id": o.GoalPortID}); err != nil {
		return err
	}
	if len(o.Inputs) > maxCertificateOperations || len(o.Mapping) > maxCertificateListItems || len(o.Outputs) > maxCertificateListItems || len(o.Ports) > maxCertificateListItems || len(o.Interface) > maxCertificateInterface || len(o.Substitutions) > maxCertificateListItems {
		return certJSONResource(p, "operation collection exceeds item limit")
	}
	if !o.ClaimedStatus.Valid() {
		return certJSONInvalid("INVALID_STATUS", p+".claimed_status", "unknown status value")
	}
	for i, v := range o.Inputs {
		if err := certText(v, fmt.Sprintf("%s.inputs[%d]", p, i), 256, false); err != nil {
			return err
		}
	}
	for i, v := range o.Ports {
		if err := certText(v, fmt.Sprintf("%s.ports[%d]", p, i), 256, false); err != nil {
			return err
		}
	}
	for i, v := range o.Mapping {
		if err := boundedStrings(fmt.Sprintf("%s.mapping[%d]", p, i), map[string]string{"left_port": v.LeftPort, "right_port": v.RightPort}); err != nil {
			return err
		}
	}
	for i, v := range o.Outputs {
		if err := boundedStrings(fmt.Sprintf("%s.outputs[%d]", p, i), map[string]string{"side": v.Side, "port": v.Port, "as": v.As}); err != nil {
			return err
		}
	}
	for i, v := range o.Interface {
		if err := boundedStrings(fmt.Sprintf("%s.interface[%d]", p, i), map[string]string{"port": v.Port, "binder": v.Binder}); err != nil {
			return err
		}
	}
	for i, v := range o.Substitutions {
		q := fmt.Sprintf("%s.substitutions[%d]", p, i)
		if err := boundedStrings(q, map[string]string{"binder_id": v.BinderID, "reference_id": v.ReferenceID}); err != nil {
			return err
		}
		if err := validateCertType(v.Type, q+".type", 0, map[*algebra.Type]bool{}); err != nil {
			return err
		}
	}
	if o.CheckedType != nil {
		if err := validateCertType(*o.CheckedType, p+".checked_type", 0, map[*algebra.Type]bool{}); err != nil {
			return err
		}
	}
	return nil
}

func validateGoalWireValue(g GoalSpec) error {
	if len(g.Ports) > MaxGoalPorts || len(g.References) > 512 || len(g.ProjectPorts) > MaxGoalPorts || len(g.AllowedOperators) > 6 || len(g.Constraints) > MaxGoalConstraints || len(g.Provenance) > MaxGoalSources {
		return certJSONResource("certificate.concrete_goal", "goal collection exceeds wire limit")
	}
	for name, value := range map[string]string{"profile": g.Profile, "context_id": g.ContextID} {
		if err := certText(value, "certificate.concrete_goal."+name, 256, false); err != nil {
			return err
		}
	}
	if err := validateCertType(g.ExpectedType, "certificate.concrete_goal.expected_type", 0, map[*algebra.Type]bool{}); err != nil {
		return err
	}
	for i, p := range g.Ports {
		q := fmt.Sprintf("certificate.concrete_goal.ports[%d]", i)
		if err := certText(p.ID, q+".id", 256, false); err != nil {
			return err
		}
		if err := validateCertType(p.Type, q+".type", 0, map[*algebra.Type]bool{}); err != nil {
			return err
		}
		if p.Binding != nil {
			if err := certText(p.Binding.ID, q+".binding.id", 256, false); err != nil {
				return err
			}
			if p.Binding.Type == nil {
				return certJSONInvalid("MISSING_TYPE", q+".binding.type", "binding type is required")
			}
			if err := validateCertType(*p.Binding.Type, q+".binding.type", 0, map[*algebra.Type]bool{}); err != nil {
				return err
			}
		}
	}
	for i, d := range g.References {
		q := fmt.Sprintf("certificate.concrete_goal.references[%d]", i)
		if err := certText(d.ID, q+".id", 256, false); err != nil {
			return err
		}
		if err := validateCertType(d.Type, q+".type", 0, map[*algebra.Type]bool{}); err != nil {
			return err
		}
	}
	for i, s := range g.AllowedOperators {
		if err := certText(s, fmt.Sprintf("certificate.concrete_goal.allowed_operators[%d]", i), 256, false); err != nil {
			return err
		}
	}
	for i, c := range g.Constraints {
		q := fmt.Sprintf("certificate.concrete_goal.constraints[%d]", i)
		if err := boundedStrings(q, map[string]string{"id": c.ID, "symbol": c.Symbol, "from_port": c.FromPort, "to_port": c.ToPort, "from_role": c.FromRole, "to_role": c.ToRole, "port": c.Port}); err != nil {
			return err
		}
		if len(c.AllowedAtomCounts) > 16 {
			return certJSONResource(q+".allowed_atom_counts", "count list exceeds limit")
		}
	}
	for i, s := range g.Provenance {
		if err := s.Validate(); err != nil {
			return certJSONInvalid("INVALID_SOURCE_REF", fmt.Sprintf("certificate.concrete_goal.provenance[%d]", i), err.Error())
		}
	}
	return nil
}

func validateCertType(t algebra.Type, path string, depth int, active map[*algebra.Type]bool) error {
	if depth > algebra.DefaultMaxTypeDepth {
		return certJSONResource(path, "type nesting exceeds the maximum depth")
	}
	// Type values contain pointers only on arrows. Track active pointers to
	// reject programmatic cycles before json.Marshal can recurse indefinitely.
	var visit func(*algebra.Type, int) error
	visit = func(v *algebra.Type, d int) error {
		if v == nil {
			return nil
		}
		if d > algebra.DefaultMaxTypeDepth {
			return certJSONResource(path, "type nesting exceeds the maximum depth")
		}
		if active[v] {
			return certJSONInvalid("CYCLIC_TYPE", path, "type contains a pointer cycle")
		}
		active[v] = true
		defer delete(active, v)
		switch v.Kind {
		case algebra.TypeEntity, algebra.TypeEvent, algebra.TypeInterval, algebra.TypeProp, algebra.TypeQuestion:
			if v.SortID != "" || v.From != nil || v.To != nil {
				return certJSONInvalid("INVALID_TYPE_SHAPE", path, "built-in sort has irrelevant fields")
			}
		case algebra.TypeNominal:
			if v.SortID == "" || len(v.SortID) > 256 || !utf8.ValidString(v.SortID) || v.From != nil || v.To != nil {
				return certJSONInvalid("INVALID_TYPE_SHAPE", path, "nominal sort requires sort_id only")
			}
		case algebra.TypeArrow:
			if v.SortID != "" || v.From == nil || v.To == nil {
				return certJSONInvalid("INVALID_TYPE_SHAPE", path, "arrow requires from and to types only")
			}
			if err := visit(v.From, d+1); err != nil {
				return err
			}
			if err := visit(v.To, d+1); err != nil {
				return err
			}
		default:
			return certJSONInvalid("UNKNOWN_TYPE_KIND", path, "type kind is not supported")
		}
		return nil
	}
	return visit(&t, depth)
}

func normalizeCertificateGoal(g GoalSpec) GoalSpec {
	g.Ports = append([]GoalPort{}, g.Ports...)
	g.References = append([]algebra.Declaration{}, g.References...)
	g.ProjectPorts = append([]string{}, g.ProjectPorts...)
	g.AllowedOperators = append([]string{}, g.AllowedOperators...)
	g.Constraints = append([]Constraint{}, g.Constraints...)
	g.Provenance = append([]provenance.SourceRef{}, g.Provenance...)
	return g
}

func normalizeCertificate(c SynthesisCertificate) SynthesisCertificate {
	c.Choices = append([]ChoiceAssignment{}, c.Choices...)
	c.Interface = append([]algebra.PortBinding{}, c.Interface...)
	c.Operations = append([]CertifiedOperation{}, c.Operations...)
	c.Scopes = append([]CertificateScope{}, c.Scopes...)
	c.Obligations = append([]GoalObligation{}, c.Obligations...)
	c.Assumptions = append([]CertificateAssumption{}, c.Assumptions...)
	c.ConcreteGoal = normalizeCertificateGoal(c.ConcreteGoal)
	for i := range c.Operations {
		o := &c.Operations[i]
		o.Inputs = append([]string{}, o.Inputs...)
		o.Mapping = append([]algebra.JoinPair{}, o.Mapping...)
		o.Outputs = append([]algebra.JoinOutput{}, o.Outputs...)
		o.Ports = append([]string{}, o.Ports...)
		o.Interface = append([]algebra.PortBinding{}, o.Interface...)
		o.Substitutions = append([]CertifiedSubstitution{}, o.Substitutions...)
	}
	for i := range c.Obligations {
		c.Obligations[i].StepIDs = append([]string{}, c.Obligations[i].StepIDs...)
		c.Obligations[i].Fragments = append([]string{}, c.Obligations[i].Fragments...)
	}
	return c
}

func decodeCertificateArray[T any](raw json.RawMessage, path string, max int, decode func(json.RawMessage, string) (T, error)) ([]T, error) {
	items, err := certArray(raw, path, max)
	if err != nil {
		return nil, err
	}
	out := make([]T, len(items))
	for i, item := range items {
		v, e := decode(item, fmt.Sprintf("%s[%d]", path, i))
		if e != nil {
			return nil, e
		}
		out[i] = v
	}
	return out, nil
}
func decodeCertificateChoice(raw json.RawMessage, p string) (ChoiceAssignment, error) {
	m, e := certObject(raw, p, certFields("kind", "axis_id", "constraint_id", "symbol", "from_role", "to_role", "atom_count", "port_id", "unary_role", "hole_id", "reference_id", "outside_path_bound"), []string{"kind", "axis_id"})
	if e != nil {
		return ChoiceAssignment{}, e
	}
	var v ChoiceAssignment
	if v.Kind, e = rawString(m, "kind", p+".kind", 256, false); e != nil {
		return v, e
	}
	if v.AxisID, e = rawString(m, "axis_id", p+".axis_id", 256, false); e != nil {
		return v, e
	}
	for k, dst := range map[string]*string{"constraint_id": &v.ConstraintID, "symbol": &v.Symbol, "from_role": &v.FromRole, "to_role": &v.ToRole, "port_id": &v.PortID, "unary_role": &v.UnaryRole, "hole_id": &v.HoleID, "reference_id": &v.ReferenceID} {
		if r, ok := m[k]; ok {
			if *dst, e = certString(r, p+"."+k, 256, false); e != nil {
				return v, e
			}
		}
	}
	if r, ok := m["atom_count"]; ok {
		if v.AtomCount, e = certInt(r, p+".atom_count", 0, MaxGoalAtoms); e != nil {
			return v, e
		}
	}
	if r, ok := m["outside_path_bound"]; ok {
		if v.OutsidePathBound, e = certBool(r, p+".outside_path_bound"); e != nil {
			return v, e
		}
	}
	return v, nil
}
func decodeCertificatePortBinding(raw json.RawMessage, p string) (algebra.PortBinding, error) {
	m, e := certObject(raw, p, certFields("port", "binder"), []string{"port", "binder"})
	if e != nil {
		return algebra.PortBinding{}, e
	}
	a, e := rawString(m, "port", p+".port", 256, false)
	if e != nil {
		return algebra.PortBinding{}, e
	}
	b, e := rawString(m, "binder", p+".binder", 256, false)
	return algebra.PortBinding{Port: a, Binder: b}, e
}
func decodeCertifiedOperation(raw json.RawMessage, p string) (CertifiedOperation, error) {
	m, e := certObject(raw, p, certFields("id", "rule_id", "op", "inputs", "symbol", "mapping", "outputs", "port", "ports", "reference_id", "goal_port_id", "claimed_status", "checked_type", "interface", "substitutions"), []string{"id", "rule_id", "op", "inputs", "mapping", "outputs", "ports", "claimed_status", "checked_type", "interface", "substitutions"})
	if e != nil {
		return CertifiedOperation{}, e
	}
	var v CertifiedOperation
	for k, d := range map[string]*string{"id": &v.ID, "rule_id": &v.RuleID, "op": &v.Op, "symbol": &v.Symbol, "port": &v.Port, "reference_id": &v.ReferenceID, "goal_port_id": &v.GoalPortID} {
		if r, ok := m[k]; ok {
			if *d, e = certString(r, p+"."+k, 256, false); e != nil {
				return v, e
			}
		} else if k == "id" || k == "rule_id" || k == "op" {
			return v, certJSONInvalid("MISSING_FIELD", p+"."+k, "required field is missing")
		}
	}
	if v.Inputs, e = decodeStringArray(m["inputs"], p+".inputs", maxCertificateOperations, false); e != nil {
		return v, e
	}
	if v.Mapping, e = decodeCertificateArray(m["mapping"], p+".mapping", maxCertificateListItems, decodeJoinPair); e != nil {
		return v, e
	}
	if v.Outputs, e = decodeCertificateArray(m["outputs"], p+".outputs", maxCertificateListItems, decodeJoinOutput); e != nil {
		return v, e
	}
	if v.Ports, e = decodeStringArray(m["ports"], p+".ports", maxCertificateListItems, false); e != nil {
		return v, e
	}
	if v.ClaimedStatus, e = decodeOutcome(m["claimed_status"], p+".claimed_status"); e != nil {
		return v, e
	}
	if v.Interface, e = decodeCertificateArray(m["interface"], p+".interface", maxCertificateInterface, decodeCertificatePortBinding); e != nil {
		return v, e
	}
	if v.Substitutions, e = decodeCertificateArray(m["substitutions"], p+".substitutions", maxCertificateListItems, decodeCertificateSubstitution); e != nil {
		return v, e
	}
	t, x := algebra.DecodeTypeJSON(m["checked_type"], p+".checked_type")
	if x != nil {
		return v, x
	}
	v.CheckedType = &t
	if !validCertificateOperation(v.Op, v.RuleID) {
		return v, certJSONInvalid("INVALID_OPERATION_RULE", p, "operation and versioned rule_id do not match the closed certificate vocabulary")
	}
	return v, nil
}

func validCertificateOperation(op, rule string) bool {
	want := map[string]string{
		"relation": CertificateRuleRelation, "join": CertificateRuleJoin,
		"restrict": CertificateRuleRestrict, "bind": CertificateRuleBind,
		"project": CertificateRuleProject, "permute": CertificateRulePermute,
	}
	return want[op] != "" && want[op] == rule
}

func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func decodeJoinPair(raw json.RawMessage, p string) (algebra.JoinPair, error) {
	m, e := certObject(raw, p, certFields("left_port", "right_port"), []string{"left_port", "right_port"})
	if e != nil {
		return algebra.JoinPair{}, e
	}
	a, e := rawString(m, "left_port", p+".left_port", 256, false)
	if e != nil {
		return algebra.JoinPair{}, e
	}
	b, e := rawString(m, "right_port", p+".right_port", 256, false)
	return algebra.JoinPair{LeftPort: a, RightPort: b}, e
}
func decodeJoinOutput(raw json.RawMessage, p string) (algebra.JoinOutput, error) {
	m, e := certObject(raw, p, certFields("side", "port", "as"), []string{"side", "port", "as"})
	if e != nil {
		return algebra.JoinOutput{}, e
	}
	a, e := rawString(m, "side", p+".side", 256, false)
	if e != nil {
		return algebra.JoinOutput{}, e
	}
	b, e := rawString(m, "port", p+".port", 256, false)
	if e != nil {
		return algebra.JoinOutput{}, e
	}
	c, e := rawString(m, "as", p+".as", 256, false)
	return algebra.JoinOutput{Side: a, Port: b, As: c}, e
}
func decodeCertificateSubstitution(raw json.RawMessage, p string) (CertifiedSubstitution, error) {
	m, e := certObject(raw, p, certFields("binder_id", "reference_id", "type"), []string{"binder_id", "reference_id", "type"})
	if e != nil {
		return CertifiedSubstitution{}, e
	}
	a, e := rawString(m, "binder_id", p+".binder_id", 256, false)
	if e != nil {
		return CertifiedSubstitution{}, e
	}
	b, e := rawString(m, "reference_id", p+".reference_id", 256, false)
	if e != nil {
		return CertifiedSubstitution{}, e
	}
	t, e := algebra.DecodeTypeJSON(m["type"], p+".type")
	return CertifiedSubstitution{BinderID: a, ReferenceID: b, Type: t}, e
}
func decodeCertificateScope(raw json.RawMessage, p string) (CertificateScope, error) {
	m, e := certObject(raw, p, certFields("binder_id", "kind", "type", "path", "parent_binder_id"), []string{"binder_id", "kind", "type", "path"})
	if e != nil {
		return CertificateScope{}, e
	}
	var v CertificateScope
	if v.BinderID, e = rawString(m, "binder_id", p+".binder_id", 256, false); e != nil {
		return v, e
	}
	k, e := rawString(m, "kind", p+".kind", 32, false)
	if e != nil {
		return v, e
	}
	v.Kind = algebra.TermKind(k)
	if v.Kind != algebra.TermLambda && v.Kind != algebra.TermExists {
		return v, certJSONInvalid("INVALID_SCOPE_KIND", p+".kind", "scope kind must be Lambda or Exists")
	}
	v.Type, e = algebra.DecodeTypeJSON(m["type"], p+".type")
	if e != nil {
		return v, e
	}
	if v.Path, e = certString(m["path"], p+".path", maxCertificatePathBytes, false); e != nil {
		return v, e
	}
	if r, ok := m["parent_binder_id"]; ok {
		v.ParentBinderID, e = certString(r, p+".parent_binder_id", 256, false)
	}
	return v, e
}
func decodeCertificateObligation(raw json.RawMessage, p string) (GoalObligation, error) {
	m, e := certObject(raw, p, certFields("constraint_id", "kind", "status", "step_ids", "fragments"), []string{"constraint_id", "kind", "status", "step_ids", "fragments"})
	if e != nil {
		return GoalObligation{}, e
	}
	var v GoalObligation
	if v.ConstraintID, e = rawString(m, "constraint_id", p+".constraint_id", 256, false); e != nil {
		return v, e
	}
	k, e := rawString(m, "kind", p+".kind", 32, false)
	if e != nil {
		return v, e
	}
	v.Kind = ConstraintKind(k)
	if v.Kind != ConstraintPath && v.Kind != ConstraintUnary {
		return v, certJSONInvalid("INVALID_CONSTRAINT_KIND", p+".kind", "constraint kind is unknown")
	}
	if v.Status, e = rawString(m, "status", p+".status", 256, false); e != nil {
		return v, e
	}
	if v.StepIDs, e = decodeStringArray(m["step_ids"], p+".step_ids", maxCertificateOperations, false); e != nil {
		return v, e
	}
	if v.Fragments, e = decodeStringArray(m["fragments"], p+".fragments", 16, false); e != nil {
		return v, e
	}
	for i, x := range v.Fragments {
		if len(x) > maxCertificatePathBytes {
			return v, certJSONResource(fmt.Sprintf("%s.fragments[%d]", p, i), "fragment path exceeds byte limit")
		}
	}
	return v, nil
}
func decodeCertificateAssumption(raw json.RawMessage, p string) (CertificateAssumption, error) {
	m, e := certObject(raw, p, certFields("kind", "value"), []string{"kind", "value"})
	if e != nil {
		return CertificateAssumption{}, e
	}
	a, e := rawString(m, "kind", p+".kind", 256, false)
	if e != nil {
		return CertificateAssumption{}, e
	}
	if !oneOf(a, CertificateAssumptionRegistrySnapshot, CertificateAssumptionRequestReferenceBound, CertificateAssumptionPositiveProfile) {
		return CertificateAssumption{}, certJSONInvalid("INVALID_ASSUMPTION_KIND", p+".kind", "assumption kind is unknown")
	}
	b, e := rawString(m, "value", p+".value", 256, false)
	return CertificateAssumption{Kind: a, Value: b}, e
}

func decodeStringArray(raw json.RawMessage, p string, max int, allowEmpty bool) ([]string, error) {
	items, e := certArray(raw, p, max)
	if e != nil {
		return nil, e
	}
	out := make([]string, len(items))
	for i, v := range items {
		out[i], e = certString(v, fmt.Sprintf("%s[%d]", p, i), 256, allowEmpty)
		if e != nil {
			return nil, e
		}
	}
	return out, nil
}
func certObject(raw []byte, p string, allowed map[string]bool, required []string) (map[string]json.RawMessage, error) {
	if len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, certJSONInvalid("INVALID_OBJECT", p, "expected a non-null JSON object")
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return nil, certJSONInvalid("INVALID_OBJECT", p, "expected a JSON object")
	}
	for k := range m {
		if !allowed[k] {
			return nil, certJSONInvalid("UNKNOWN_FIELD", p+"."+k, "field is not permitted")
		}
	}
	for _, k := range required {
		if _, ok := m[k]; !ok {
			return nil, certJSONInvalid("MISSING_FIELD", p+"."+k, "required field is missing")
		}
	}
	return m, nil
}
func certFields(values ...string) map[string]bool {
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}
func certArray(raw json.RawMessage, p string, max int) ([]json.RawMessage, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, certJSONInvalid("INVALID_ARRAY", p, "expected a non-null JSON array")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return nil, certJSONInvalid("INVALID_ARRAY", p, "expected a JSON array")
	}
	if len(values) > max {
		return nil, certJSONResource(p, "array exceeds item limit")
	}
	return values, nil
}
func rawString(m map[string]json.RawMessage, k, p string, max int, allowEmpty bool) (string, error) {
	r, ok := m[k]
	if !ok {
		return "", certJSONInvalid("MISSING_FIELD", p, "required field is missing")
	}
	return certString(r, p, max, allowEmpty)
}
func certString(raw []byte, p string, max int, allowEmpty bool) (string, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", certJSONInvalid("NULL_FIELD", p, "string cannot be null")
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", certJSONInvalid("INVALID_STRING", p, "expected a string")
	}
	if !utf8.ValidString(s) || len(s) > max || (!allowEmpty && strings.TrimSpace(s) == "") {
		return "", certJSONInvalid("INVALID_STRING", p, "string is empty or exceeds its UTF-8 byte limit")
	}
	return s, nil
}
func certText(s, p string, max int, allowEmpty bool) error {
	if !utf8.ValidString(s) || len(s) > max || (!allowEmpty && strings.TrimSpace(s) == "") {
		return certJSONInvalid("INVALID_STRING", p, "string is empty or exceeds its UTF-8 byte limit")
	}
	return nil
}
func certInt(raw []byte, p string, min, max int) (int, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return 0, certJSONInvalid("INVALID_INTEGER", p, "integer cannot be null")
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0, certJSONInvalid("INVALID_INTEGER", p, "expected integer")
	}
	if n < min || n > max {
		return 0, certJSONInvalid("INVALID_INTEGER", p, fmt.Sprintf("integer must be between %d and %d", min, max))
	}
	return n, nil
}
func certBool(raw []byte, p string) (bool, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return false, certJSONInvalid("INVALID_BOOLEAN", p, "boolean cannot be null")
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err != nil {
		return false, certJSONInvalid("INVALID_BOOLEAN", p, "expected boolean")
	}
	return v, nil
}
func decodeOutcome(raw []byte, p string) (status.OutcomeKind, error) {
	s, e := certString(raw, p, 64, false)
	if e != nil {
		return "", e
	}
	v := status.OutcomeKind(s)
	if !v.Valid() {
		return "", certJSONInvalid("INVALID_STATUS", p, "unknown status value")
	}
	return v, nil
}
func boundedStrings(path string, values map[string]string) error {
	for name, value := range values {
		if value != "" {
			if err := certText(value, path+"."+name, 256, false); err != nil {
				return err
			}
		}
	}
	return nil
}
func certJSONInvalid(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeInvalid, Code: code, Path: path, Message: message}
}
func certJSONUnsupported(code, path, message string) error {
	return &algebra.Error{Status: status.OutcomeUnsupported, Code: code, Path: path, Message: message}
}
func certJSONResource(path, message string) error {
	return &algebra.Error{Status: status.OutcomeResourceLimit, Code: "RESOURCE_LIMIT", Path: path, Message: message}
}
