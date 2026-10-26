package algebra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/provenance"
)

// MaxDocumentBytes is the hard wire-format size ceiling for one algebra
// document. It is checked before parsing.
const MaxDocumentBytes = 1 << 20

// MaxJSONNesting is a pre-decode cap on nested JSON arrays and objects.
const MaxJSONNesting = 256

const (
	MaxArtifactExportRequestBytes = MaxDocumentBytes
	MaxArtifactEnvelopeBytes      = MaxArtifactBytes
	MaxArtifactReplayRequestBytes = MaxRewriteEnvelopeBytes
)

// ArtifactExportRequest carries source text and explicit request metadata. It
// contains no trusted registry inventory; a metadata factory supplies that.
type ArtifactExportRequest struct {
	SchemaVersion       int                    `json:"schema_version"`
	Document            Document               `json:"document"`
	Profile             RewriteProfile         `json:"profile"`
	ContextID           string                 `json:"context_id"`
	ContextDependencies []ContextDependency    `json:"context_dependencies"`
	Provenance          []provenance.SourceRef `json:"provenance"`
}

// ArtifactReplayRequest separates expected metadata claims from the artifact
// being checked. Its profile, context, dependencies, provenance, and limits are
// request inputs rather than artifact assertions.
type ArtifactReplayRequest struct {
	SchemaVersion       int                    `json:"schema_version"`
	Profile             RewriteProfile         `json:"profile"`
	ContextID           string                 `json:"context_id"`
	ContextDependencies []ContextDependency    `json:"context_dependencies"`
	Provenance          []provenance.SourceRef `json:"provenance"`
	Limits              Limits                 `json:"limits,omitempty"`
	Artifact            ArtifactEnvelope       `json:"artifact"`
}

// Decode strictly decodes and structurally checks one version-1 expression
// document. It does not evaluate the expression or infer truth.
func Decode(data []byte) (Document, error) {
	document, err := DecodeDocumentJSON(data)
	if err != nil {
		return Document{}, err
	}
	if _, err := Check(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

// DecodeDocumentJSON validates the strict document wire shape and returns the
// typed document without performing declaration or expression semantics.
func DecodeDocumentJSON(data []byte) (Document, error) {
	if err := PreflightJSON(data); err != nil {
		return Document{}, err
	}
	if err := validateDocumentJSON(data); err != nil {
		return Document{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var document Document
	if err := decoder.Decode(&document); err != nil {
		return Document{}, invalid("INVALID_JSON_DOCUMENT", "document", err.Error())
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Document{}, invalid("TRAILING_DOCUMENT", "document", "expected exactly one JSON document")
	}
	return document, nil
}

// PreflightJSON applies the wire-level byte, UTF-8, surrogate, nesting,
// duplicate-key, and single-document guards without interpreting a schema.
func PreflightJSON(data []byte) error {
	return PreflightJSONLimit(data, MaxDocumentBytes)
}

// PreflightJSONLimit applies the shared strict wire guards with an explicit
// bounded envelope size. Ordinary document decoders continue to use
// PreflightJSON and its 1 MiB ceiling.
func PreflightJSONLimit(data []byte, maxBytes int) error {
	if maxBytes <= 0 || maxBytes > MaxRewriteEnvelopeBytes {
		return invalid("INVALID_JSON_BYTE_LIMIT", "max_bytes", fmt.Sprintf("requested limit must be between 1 and %d bytes", MaxRewriteEnvelopeBytes))
	}
	if len(data) > maxBytes {
		return resourceLimit("document", fmt.Sprintf("JSON document is %d bytes; maximum is %d", len(data), maxBytes))
	}
	if !utf8.Valid(data) {
		return invalid("INVALID_UTF8", "document", "JSON document is not valid UTF-8")
	}
	if err := validateUnicodeEscapes(data); err != nil {
		return err
	}
	return scanJSON(data)
}

// MarshalJSON emits the strict source-neutral wire form used by rewrite
// requests. In particular, the document term is encoded through termWire so
// zero-arity Atom arguments remain an explicit empty array.
func (request NormalizeRequest) MarshalJSON() ([]byte, error) {
	if request.SchemaVersion != 1 {
		return nil, invalid("INVALID_SCHEMA_VERSION", "request.schema_version", "schema version must be 1")
	}
	if err := ValidateRewriteProfile(request.Profile); err != nil {
		return nil, err
	}
	if _, err := NormalizeRewriteBudget(request.Budget); err != nil {
		return nil, err
	}
	document, err := rewriteDocumentWire(request.Document)
	if err != nil {
		return nil, err
	}
	wire := map[string]any{
		"schema_version": request.SchemaVersion,
		"document":       document,
		"profile":        rewriteProfileWire(request.Profile),
		"budget":         request.Budget,
	}
	return marshalBoundedWire(wire, MaxDocumentBytes, "request")
}

// MarshalJSON emits one bounded rewrite step, preserving empty root paths and
// strict term shapes without changing the legacy Term marshaler.
func (step RewriteStep) MarshalJSON() ([]byte, error) {
	wire, err := rewriteStepWire(step)
	if err != nil {
		return nil, err
	}
	return marshalBoundedWire(wire, MaxRewriteCertificateBytes, "step")
}

// MarshalJSON emits a bounded certificate using the new wire adapters.
func (certificate RewriteCertificate) MarshalJSON() ([]byte, error) {
	wire, err := rewriteCertificateWire(certificate)
	if err != nil {
		return nil, err
	}
	return marshalBoundedWire(wire, MaxRewriteCertificateBytes, "certificate")
}

// MarshalJSON bounds the complete verification envelope while nested request
// and certificate values retain their independent limits.
func (request RewriteVerificationRequest) MarshalJSON() ([]byte, error) {
	if request.SchemaVersion != 1 {
		return nil, invalid("INVALID_SCHEMA_VERSION", "verification.schema_version", "schema version must be 1")
	}
	wire := map[string]any{"schema_version": request.SchemaVersion, "request": request.Request, "certificate": request.Certificate}
	return marshalBoundedWire(wire, MaxRewriteEnvelopeBytes, "verification")
}

// MarshalJSON encodes the new normalization result shape. The ordinary
// Document and Term encodings remain unchanged for existing product routes.
func (result NormalizationResult) MarshalJSON() ([]byte, error) {
	document, err := rewriteDocumentWire(result.CanonicalDocument)
	if err != nil {
		return nil, err
	}
	expression, err := rewriteTermWire(result.Expression, "expression")
	if err != nil {
		return nil, err
	}
	certificate, err := rewriteCertificateWire(result.Certificate)
	if err != nil {
		return nil, err
	}
	wire := map[string]any{
		"status": result.Status, "normalized": result.Normalized,
		"canonical_document": document, "checked": result.Checked,
		"expression": expression, "certificate": certificate,
		"statistics": result.Statistics,
	}
	if result.Diagnostics != nil {
		if len(result.Diagnostics) > 1 {
			return nil, resourceLimit("diagnostics", "normalization diagnostics exceed the bounded result shape")
		}
		wire["diagnostics"] = result.Diagnostics
	}
	return marshalBoundedWire(wire, MaxRewriteEnvelopeBytes, "normalization_result")
}

func rewriteDocumentWire(document Document) (map[string]any, error) {
	if len(document.Sorts) > DefaultMaxSymbols || len(document.Signatures) > DefaultMaxSymbols || len(document.Parameters) > DefaultMaxSymbols || len(document.References) > DefaultMaxSymbols {
		return nil, resourceLimit("document", "rewrite document declarations exceed structural bounds")
	}
	limits, err := NormalizeLimits(document.Limits)
	if err != nil {
		return nil, err
	}
	copyDocument := document
	copyDocument.Limits = document.Limits
	copyDocument.Sorts = append([]SortDeclaration{}, document.Sorts...)
	for i := range copyDocument.Sorts {
		if !validText(copyDocument.Sorts[i].ID, 256) || !validOptionalText(copyDocument.Sorts[i].Label, 1024) {
			return nil, invalid("INVALID_SORT", fmt.Sprintf("document.sorts[%d]", i), "sort text exceeds its bound")
		}
	}
	copyDocument.Signatures = make([]Signature, len(document.Signatures))
	for i, signature := range document.Signatures {
		if !validText(signature.ID, 256) || !validOptionalText(signature.Label, 1024) || len(signature.Ports) > limits.MaxPorts {
			return nil, invalid("INVALID_SIGNATURE", fmt.Sprintf("document.signatures[%d]", i), "signature fields exceed their bounds")
		}
		copyDocument.Signatures[i] = Signature{ID: signature.ID, Label: signature.Label, Ports: make([]Port, len(signature.Ports))}
		for j, port := range signature.Ports {
			if !validText(port.Role, 256) {
				return nil, invalid("INVALID_PORT", fmt.Sprintf("document.signatures[%d].ports[%d].role", i, j), "port role exceeds its bound")
			}
			typeCopy, typeErr := cloneStandaloneType(&port.Type, limits)
			if typeErr != nil {
				return nil, typeErr
			}
			copyDocument.Signatures[i].Ports[j] = Port{Role: port.Role, Type: typeCopy}
		}
	}
	copyDeclarations := func(values []Declaration, path string) ([]Declaration, error) {
		copyValues := make([]Declaration, len(values))
		for i, declaration := range values {
			if !validText(declaration.ID, 256) {
				return nil, invalid("INVALID_DECLARATION", fmt.Sprintf("%s[%d].id", path, i), "declaration ID exceeds its bound")
			}
			typeCopy, typeErr := cloneStandaloneType(&declaration.Type, limits)
			if typeErr != nil {
				return nil, typeErr
			}
			copyValues[i] = Declaration{ID: declaration.ID, Type: typeCopy}
		}
		return copyValues, nil
	}
	copyDocument.Parameters, err = copyDeclarations(document.Parameters, "document.parameters")
	if err != nil {
		return nil, err
	}
	copyDocument.References, err = copyDeclarations(document.References, "document.references")
	if err != nil {
		return nil, err
	}
	copyDocument.Expression, err = copyTermDetached(document.Expression, limits)
	if err != nil {
		return nil, err
	}
	wire, err := documentWire(copyDocument)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("ENCODE_ERROR", "document", err.Error())
	}
	if _, err := DecodeDocumentJSON(encoded); err != nil {
		return nil, err
	}
	return wire, nil
}

func rewriteTermWire(term *Term, path string) (map[string]any, error) {
	copyTerm, err := copyTermDetached(term, DefaultLimits())
	if err != nil {
		return nil, err
	}
	wire, err := termWire(copyTerm)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("ENCODE_ERROR", path, err.Error())
	}
	if _, err := DecodeTermJSON(encoded, path); err != nil {
		return nil, err
	}
	if len(encoded) > MaxRewriteTermBytes {
		return nil, resourceLimit(path, "rewrite term exceeds its serialized byte limit")
	}
	return wire, nil
}

func rewriteProfileWire(profile RewriteProfile) map[string]any {
	rules := append([]RewriteRule{}, profile.Rules...)
	return map[string]any{"id": profile.ID, "rules": rules}
}

func rewriteStepWire(step RewriteStep) (map[string]any, error) {
	if len(step.Path) > DefaultMaxDepth || len(step.BinderMapping) > 2*DefaultMaxNodes || len(step.SideConditions) > 2*DefaultMaxNodes {
		return nil, resourceLimit("step", "rewrite step metadata exceeds structural bounds")
	}
	if len(step.ParameterMapping) != 0 {
		return nil, invalid("REWRITE_PARAMETER_MAPPING", "step.parameter_mapping", "rewrite rules cannot substitute free parameters")
	}
	if !validText(string(step.RuleID), 32) || !validText(step.SnapshotID, 256) || !validText(step.ProfileID, 256) {
		return nil, invalid("INVALID_REWRITE_STEP", "step", "rewrite step identifiers exceed their bounds")
	}
	if step.BeforeNodes < 1 || step.BeforeNodes > DefaultMaxNodes || step.AfterNodes < 1 || step.AfterNodes > DefaultMaxNodes {
		return nil, invalid("INVALID_REWRITE_STEP", "step.before_nodes", "rewrite step node counts are outside their bounds")
	}
	path := append([]int{}, step.Path...)
	for _, index := range path {
		if index < 0 || index >= DefaultMaxNodes {
			return nil, invalid("INVALID_REWRITE_STEP", "step.path", "rewrite path index is outside its bound")
		}
	}
	before, err := rewriteTermWire(step.Before, "step.before")
	if err != nil {
		return nil, err
	}
	after, err := rewriteTermWire(step.After, "step.after")
	if err != nil {
		return nil, err
	}
	binderMapping := make([]BinderChange, len(step.BinderMapping))
	for i, change := range step.BinderMapping {
		if !validOptionalText(change.Before, 256) || !validOptionalText(change.After, 256) || !validText(change.Kind, 32) {
			return nil, invalid("INVALID_BINDER_CHANGE", fmt.Sprintf("step.binder_mapping[%d]", i), "binder mapping text exceeds its bound")
		}
		switch change.Kind {
		case "retained", "renamed", "removed", "introduced":
		default:
			return nil, invalid("INVALID_BINDER_CHANGE", fmt.Sprintf("step.binder_mapping[%d].kind", i), "binder change kind is unsupported")
		}
		binderMapping[i] = change
	}
	sideConditions := append([]string{}, step.SideConditions...)
	for i, condition := range sideConditions {
		if !validOptionalText(condition, 2048) {
			return nil, invalid("INVALID_SIDE_CONDITION", fmt.Sprintf("step.side_conditions[%d]", i), "side condition exceeds its bound")
		}
	}
	wire := map[string]any{
		"rule_id": step.RuleID, "path": path, "before": before, "after": after,
		"snapshot_id": step.SnapshotID, "profile_id": step.ProfileID,
		"before_nodes": step.BeforeNodes, "after_nodes": step.AfterNodes,
	}
	if step.BinderMapping != nil {
		wire["binder_mapping"] = binderMapping
	}
	if step.ParameterMapping != nil {
		wire["parameter_mapping"] = map[string]any{}
	}
	if step.SideConditions != nil {
		wire["side_conditions"] = sideConditions
	}
	return wire, nil
}

func rewriteCertificateWire(certificate RewriteCertificate) (map[string]any, error) {
	if certificate.SchemaVersion != 1 {
		return nil, invalid("INVALID_SCHEMA_VERSION", "certificate.schema_version", "schema version must be 1")
	}
	if !validText(certificate.SnapshotID, 256) || len(certificate.Steps) > DefaultRewriteMaxSteps {
		return nil, resourceLimit("certificate", "certificate metadata exceeds structural bounds")
	}
	if err := ValidateRewriteProfile(certificate.Profile); err != nil {
		return nil, err
	}
	input, err := rewriteTermWire(certificate.Input, "certificate.input")
	if err != nil {
		return nil, err
	}
	output, err := rewriteTermWire(certificate.Output, "certificate.output")
	if err != nil {
		return nil, err
	}
	steps := make([]map[string]any, len(certificate.Steps))
	for i, step := range certificate.Steps {
		steps[i], err = rewriteStepWire(step)
		if err != nil {
			return nil, err
		}
	}
	wire := map[string]any{
		"schema_version": certificate.SchemaVersion, "snapshot_id": certificate.SnapshotID,
		"profile": rewriteProfileWire(certificate.Profile), "input": input, "output": output,
		"steps": steps, "normalized": certificate.Normalized,
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("ENCODE_ERROR", "certificate", err.Error())
	}
	if len(encoded) > MaxRewriteCertificateBytes {
		return nil, resourceLimit("certificate", "encoded certificate exceeds its byte limit")
	}
	return wire, nil
}

func marshalBoundedWire(value any, maximum int, path string) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, invalid("ENCODE_ERROR", path, err.Error())
	}
	if len(encoded) > maximum {
		return nil, resourceLimit(path, fmt.Sprintf("encoded JSON exceeds %d byte limit", maximum))
	}
	return encoded, nil
}

// NormalizeRequest is a source-neutral normalization request. The caller must
// build the rewrite environment from explicitly selected trusted metadata.
type NormalizeRequest struct {
	SchemaVersion int            `json:"schema_version"`
	Document      Document       `json:"document"`
	Profile       RewriteProfile `json:"profile"`
	Budget        RewriteBudget  `json:"budget,omitempty"`
}

// RewriteVerificationRequest pairs an original normalization request with a
// claimed certificate. It carries no metadata selection or trusted snapshot.
type RewriteVerificationRequest struct {
	SchemaVersion int                `json:"schema_version"`
	Request       NormalizeRequest   `json:"request"`
	Certificate   RewriteCertificate `json:"certificate"`
}

// DecodeNormalizationRequest strictly decodes a bounded source-neutral
// normalization request. Metadata environments are never accepted on wire.
func DecodeNormalizationRequest(data []byte) (NormalizeRequest, error) {
	if err := PreflightJSON(data); err != nil {
		return NormalizeRequest{}, err
	}
	fields, err := strictObject(data, "request", []string{"schema_version", "document", "profile", "budget"}, []string{"schema_version", "document", "profile"})
	if err != nil {
		return NormalizeRequest{}, err
	}
	version, err := rewriteJSONInt(fields["schema_version"], "request.schema_version", 1, 1)
	if err != nil {
		return NormalizeRequest{}, err
	}
	document, err := DecodeDocumentJSON(fields["document"])
	if err != nil {
		return NormalizeRequest{}, err
	}
	profile, err := decodeRewriteProfile(fields["profile"], "request.profile")
	if err != nil {
		return NormalizeRequest{}, err
	}
	request := NormalizeRequest{SchemaVersion: version, Document: document, Profile: profile}
	if raw, exists := fields["budget"]; exists {
		limits, err := strictObject(raw, "request.budget", []string{"max_steps", "max_expansions", "max_trace_nodes", "max_work_nodes"}, nil)
		if err != nil {
			return NormalizeRequest{}, err
		}
		if err := strictTypedUnmarshal(raw, &request.Budget, "request.budget"); err != nil {
			return NormalizeRequest{}, err
		}
		_ = limits
		if _, err := NormalizeRewriteBudget(request.Budget); err != nil {
			return NormalizeRequest{}, err
		}
	}
	return request, nil
}

// DecodeRewriteCertificate strictly decodes a version-1 certificate using
// the explicit 8 MiB proof cap. Semantic claims are checked by the independent
// verifier against a separately constructed trusted environment.
func DecodeRewriteCertificate(data []byte) (RewriteCertificate, error) {
	if err := PreflightJSONLimit(data, MaxRewriteCertificateBytes); err != nil {
		return RewriteCertificate{}, err
	}
	fields, err := strictObject(data, "certificate", []string{"schema_version", "snapshot_id", "profile", "input", "output", "steps", "normalized"}, []string{"schema_version", "snapshot_id", "profile", "input", "output", "steps", "normalized"})
	if err != nil {
		return RewriteCertificate{}, err
	}
	version, err := rewriteJSONInt(fields["schema_version"], "certificate.schema_version", 1, 1)
	if err != nil {
		return RewriteCertificate{}, err
	}
	snapshotID, err := rewriteJSONString(fields["snapshot_id"], "certificate.snapshot_id", 256, false)
	if err != nil {
		return RewriteCertificate{}, err
	}
	profile, err := decodeRewriteProfile(fields["profile"], "certificate.profile")
	if err != nil {
		return RewriteCertificate{}, err
	}
	input, err := DecodeTermJSON(fields["input"], "certificate.input")
	if err != nil {
		return RewriteCertificate{}, err
	}
	output, err := DecodeTermJSON(fields["output"], "certificate.output")
	if err != nil {
		return RewriteCertificate{}, err
	}
	stepValues, err := strictArray(fields["steps"], "certificate.steps", DefaultRewriteMaxSteps)
	if err != nil {
		return RewriteCertificate{}, err
	}
	steps := make([]RewriteStep, len(stepValues))
	for i, raw := range stepValues {
		step, stepErr := decodeRewriteStep(raw, fmt.Sprintf("certificate.steps[%d]", i))
		if stepErr != nil {
			return RewriteCertificate{}, stepErr
		}
		steps[i] = step
	}
	normalized, err := rewriteJSONBool(fields["normalized"], "certificate.normalized")
	if err != nil {
		return RewriteCertificate{}, err
	}
	return RewriteCertificate{SchemaVersion: version, SnapshotID: snapshotID, Profile: profile, Input: input, Output: output, Steps: steps, Normalized: normalized}, nil
}

// DecodeRewriteVerificationRequest decodes a finite envelope. The nested
// request and certificate retain their smaller independent byte ceilings.
func DecodeRewriteVerificationRequest(data []byte) (RewriteVerificationRequest, error) {
	if err := PreflightJSONLimit(data, MaxRewriteEnvelopeBytes); err != nil {
		return RewriteVerificationRequest{}, err
	}
	fields, err := strictObject(data, "verification", []string{"schema_version", "request", "certificate"}, []string{"schema_version", "request", "certificate"})
	if err != nil {
		return RewriteVerificationRequest{}, err
	}
	version, err := rewriteJSONInt(fields["schema_version"], "verification.schema_version", 1, 1)
	if err != nil {
		return RewriteVerificationRequest{}, err
	}
	request, err := DecodeNormalizationRequest(fields["request"])
	if err != nil {
		return RewriteVerificationRequest{}, err
	}
	certificate, err := DecodeRewriteCertificate(fields["certificate"])
	if err != nil {
		return RewriteVerificationRequest{}, err
	}
	return RewriteVerificationRequest{SchemaVersion: version, Request: request, Certificate: certificate}, nil
}

func decodeRewriteProfile(data []byte, path string) (RewriteProfile, error) {
	fields, err := strictObject(data, path, []string{"id", "rules"}, []string{"id", "rules"})
	if err != nil {
		return RewriteProfile{}, err
	}
	id, err := rewriteJSONString(fields["id"], path+".id", 256, false)
	if err != nil {
		return RewriteProfile{}, err
	}
	rawRules, err := strictArray(fields["rules"], path+".rules", 6)
	if err != nil {
		return RewriteProfile{}, err
	}
	rules := make([]RewriteRule, len(rawRules))
	for i, raw := range rawRules {
		value, stringErr := rewriteJSONString(raw, fmt.Sprintf("%s.rules[%d]", path, i), 32, false)
		if stringErr != nil {
			return RewriteProfile{}, stringErr
		}
		rules[i] = RewriteRule(value)
	}
	profile := RewriteProfile{ID: id, Rules: rules}
	if err := ValidateRewriteProfile(profile); err != nil {
		return RewriteProfile{}, err
	}
	return profile, nil
}

func decodeRewriteStep(data []byte, path string) (RewriteStep, error) {
	allowed := []string{"rule_id", "path", "before", "after", "binder_mapping", "parameter_mapping", "side_conditions", "snapshot_id", "profile_id", "before_nodes", "after_nodes"}
	required := []string{"rule_id", "path", "before", "after", "snapshot_id", "profile_id", "before_nodes", "after_nodes"}
	fields, err := strictObject(data, path, allowed, required)
	if err != nil {
		return RewriteStep{}, err
	}
	ruleText, err := rewriteJSONString(fields["rule_id"], path+".rule_id", 32, false)
	if err != nil {
		return RewriteStep{}, err
	}
	rule := RewriteRule(ruleText)
	if err := ValidateRewriteProfile(RewriteProfile{ID: "wire-step", Rules: []RewriteRule{rule}}); err != nil {
		return RewriteStep{}, err
	}
	pathValues, err := strictArray(fields["path"], path+".path", DefaultMaxDepth)
	if err != nil {
		return RewriteStep{}, err
	}
	pathIndexes := make([]int, len(pathValues))
	for i, raw := range pathValues {
		index, intErr := rewriteJSONInt(raw, fmt.Sprintf("%s.path[%d]", path, i), 0, DefaultMaxNodes-1)
		if intErr != nil {
			return RewriteStep{}, intErr
		}
		pathIndexes[i] = index
	}
	before, err := DecodeTermJSON(fields["before"], path+".before")
	if err != nil {
		return RewriteStep{}, err
	}
	after, err := DecodeTermJSON(fields["after"], path+".after")
	if err != nil {
		return RewriteStep{}, err
	}
	snapshotID, err := rewriteJSONString(fields["snapshot_id"], path+".snapshot_id", 256, false)
	if err != nil {
		return RewriteStep{}, err
	}
	profileID, err := rewriteJSONString(fields["profile_id"], path+".profile_id", 256, false)
	if err != nil {
		return RewriteStep{}, err
	}
	beforeNodes, err := rewriteJSONInt(fields["before_nodes"], path+".before_nodes", 1, DefaultMaxNodes)
	if err != nil {
		return RewriteStep{}, err
	}
	afterNodes, err := rewriteJSONInt(fields["after_nodes"], path+".after_nodes", 1, DefaultMaxNodes)
	if err != nil {
		return RewriteStep{}, err
	}
	step := RewriteStep{RuleID: rule, Path: pathIndexes, Before: before, After: after, SnapshotID: snapshotID, ProfileID: profileID, BeforeNodes: beforeNodes, AfterNodes: afterNodes}
	if raw, exists := fields["binder_mapping"]; exists {
		items, arrayErr := strictArray(raw, path+".binder_mapping", 2*DefaultMaxNodes)
		if arrayErr != nil {
			return RewriteStep{}, arrayErr
		}
		step.BinderMapping = make([]BinderChange, len(items))
		for i, item := range items {
			itemPath := fmt.Sprintf("%s.binder_mapping[%d]", path, i)
			mapping, objectErr := strictObject(item, itemPath, []string{"before", "after", "kind"}, []string{"before", "after", "kind"})
			if objectErr != nil {
				return RewriteStep{}, objectErr
			}
			beforeID, e := rewriteJSONString(mapping["before"], itemPath+".before", 256, true)
			if e != nil {
				return RewriteStep{}, e
			}
			afterID, e := rewriteJSONString(mapping["after"], itemPath+".after", 256, true)
			if e != nil {
				return RewriteStep{}, e
			}
			kind, e := rewriteJSONString(mapping["kind"], itemPath+".kind", 32, false)
			if e != nil {
				return RewriteStep{}, e
			}
			switch kind {
			case "retained", "renamed", "removed", "introduced":
			default:
				return RewriteStep{}, invalid("INVALID_BINDER_CHANGE", itemPath+".kind", "binder change kind is unsupported")
			}
			step.BinderMapping[i] = BinderChange{Before: beforeID, After: afterID, Kind: kind}
		}
	}
	if raw, exists := fields["parameter_mapping"]; exists {
		mapping, objectErr := strictObject(raw, path+".parameter_mapping", []string{}, []string{})
		if objectErr != nil {
			return RewriteStep{}, objectErr
		}
		if len(mapping) != 0 {
			return RewriteStep{}, invalid("REWRITE_PARAMETER_MAPPING", path+".parameter_mapping", "rewrite rules cannot substitute free parameters")
		}
		step.ParameterMapping = make(Substitution)
	}
	if raw, exists := fields["side_conditions"]; exists {
		items, arrayErr := strictArray(raw, path+".side_conditions", 2*DefaultMaxNodes)
		if arrayErr != nil {
			return RewriteStep{}, arrayErr
		}
		step.SideConditions = make([]string, len(items))
		for i, item := range items {
			value, stringErr := rewriteJSONString(item, fmt.Sprintf("%s.side_conditions[%d]", path, i), 2048, true)
			if stringErr != nil {
				return RewriteStep{}, stringErr
			}
			step.SideConditions[i] = value
		}
	}
	return step, nil
}

func rewriteJSONString(data []byte, path string, maximum int, allowEmpty bool) (string, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return "", invalid("NULL_FIELD", path, "string cannot be null")
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return "", invalid("INVALID_STRING", path, "expected a JSON string")
	}
	if (allowEmpty && !validOptionalText(value, maximum)) || (!allowEmpty && !validText(value, maximum)) {
		return "", invalid("INVALID_STRING", path, "string is empty, invalid UTF-8, or exceeds its byte limit")
	}
	return value, nil
}

func rewriteJSONInt(data []byte, path string, minimum, maximum int) (int, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return 0, invalid("INVALID_INTEGER", path, "integer cannot be null")
	}
	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return 0, invalid("INVALID_INTEGER", path, "expected a JSON integer")
	}
	if value < minimum || value > maximum {
		return 0, invalid("INVALID_INTEGER", path, fmt.Sprintf("integer must be between %d and %d", minimum, maximum))
	}
	return value, nil
}

func rewriteJSONBool(data []byte, path string) (bool, error) {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return false, invalid("INVALID_BOOLEAN", path, "boolean cannot be null")
	}
	var value bool
	if err := json.Unmarshal(data, &value); err != nil {
		return false, invalid("INVALID_BOOLEAN", path, "expected a JSON boolean")
	}
	return value, nil
}

// DecodeTypeJSON strictly decodes one type value after applying the shared
// wire guards. It does not resolve nominal sorts; use CheckDeclarations for
// declaration-context validation.
func DecodeTypeJSON(data []byte, path string) (Type, error) {
	if err := PreflightJSON(data); err != nil {
		return Type{}, err
	}
	if err := validateTypeJSON(json.RawMessage(data), path); err != nil {
		return Type{}, err
	}
	var value Type
	if err := strictTypedUnmarshal(data, &value, path); err != nil {
		return Type{}, err
	}
	return value, nil
}

// DecodeTermJSON strictly decodes one term value after applying shared wire
// guards. Declaration, scope, and type checking remain the caller's job.
func DecodeTermJSON(data []byte, path string) (*Term, error) {
	if err := PreflightJSON(data); err != nil {
		return nil, err
	}
	if err := validateTermJSON(json.RawMessage(data), path); err != nil {
		return nil, err
	}
	var value Term
	if err := strictTypedUnmarshal(data, &value, path); err != nil {
		return nil, err
	}
	return &value, nil
}

func strictTypedUnmarshal(data []byte, target any, path string) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return invalid("INVALID_JSON_VALUE", path, err.Error())
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return invalid("TRAILING_DOCUMENT", path, "expected exactly one JSON value")
	}
	return nil
}

// Encode validates and encodes a document without changing it. Incomplete
// expressions containing typed holes remain encodable; no truth is evaluated.
func Encode(document Document) ([]byte, error) {
	if _, err := Check(document); err != nil {
		return nil, err
	}
	wire, err := documentWire(document)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("ENCODE_ERROR", "document", err.Error())
	}
	if len(data) > MaxDocumentBytes {
		return nil, resourceLimit("document", fmt.Sprintf("encoded JSON document is %d bytes; maximum is %d", len(data), MaxDocumentBytes))
	}
	if err := validateDocumentJSON(data); err != nil {
		return nil, err
	}
	return data, nil
}

func documentWire(document Document) (map[string]any, error) {
	wire := map[string]any{"schema_version": document.SchemaVersion}
	if len(document.Sorts) != 0 {
		wire["sorts"] = document.Sorts
	}
	if len(document.Signatures) != 0 {
		signatures := make([]map[string]any, 0, len(document.Signatures))
		for _, signature := range document.Signatures {
			ports := make([]Port, len(signature.Ports))
			copy(ports, signature.Ports)
			entry := map[string]any{"id": signature.ID, "ports": ports}
			if signature.Label != "" {
				entry["label"] = signature.Label
			}
			signatures = append(signatures, entry)
		}
		wire["signatures"] = signatures
	}
	if len(document.Parameters) != 0 {
		wire["parameters"] = document.Parameters
	}
	if len(document.References) != 0 {
		wire["references"] = document.References
	}
	if document.Closed {
		wire["closed"] = true
	}
	if document.Limits != (Limits{}) {
		wire["limits"] = document.Limits
	}
	expression, err := termWire(document.Expression)
	if err != nil {
		return nil, err
	}
	wire["expression"] = expression
	return wire, nil
}

func termWire(term *Term) (map[string]any, error) {
	if term == nil {
		return nil, invalid("MISSING_TERM", "expression", "term is required")
	}
	wire := map[string]any{"kind": term.Kind}
	if term.Label != "" {
		wire["label"] = term.Label
	}
	switch term.Kind {
	case TermVar:
		wire["id"] = term.ID
	case TermParam, TermRef, TermHole:
		wire["id"], wire["type"] = term.ID, term.Type
	case TermAtom:
		wire["symbol"] = term.Symbol
		args := make([]map[string]any, 0, len(term.Args))
		for _, arg := range term.Args {
			value, err := termWire(arg.Value)
			if err != nil {
				return nil, err
			}
			args = append(args, map[string]any{"role": arg.Role, "value": value})
		}
		wire["args"] = args
	case TermAnd:
		operands := make([]map[string]any, 0, len(term.Operands))
		for _, operand := range term.Operands {
			value, err := termWire(operand)
			if err != nil {
				return nil, err
			}
			operands = append(operands, value)
		}
		wire["operands"] = operands
	case TermExists, TermLambda:
		body, err := termWire(term.Body)
		if err != nil {
			return nil, err
		}
		wire["id"], wire["type"], wire["body"] = term.ID, term.Type, body
	case TermApply:
		function, err := termWire(term.Function)
		if err != nil {
			return nil, err
		}
		argument, err := termWire(term.Argument)
		if err != nil {
			return nil, err
		}
		wire["function"], wire["argument"] = function, argument
	default:
		return nil, invalid("UNKNOWN_TERM_KIND", "expression.kind", "term kind is unsupported")
	}
	return wire, nil
}

func scanJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder, "document", 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return invalid("INVALID_JSON", "document", err.Error())
		}
		return invalid("TRAILING_DOCUMENT", "document", "expected exactly one JSON value")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder, path string, depth int) error {
	token, err := decoder.Token()
	if err != nil {
		return invalid("INVALID_JSON", path, err.Error())
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	if depth >= MaxJSONNesting {
		return resourceLimit(path, fmt.Sprintf("JSON nesting exceeds limit %d", MaxJSONNesting))
	}
	switch delimiter {
	case '{':
		seen := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return invalid("INVALID_JSON", path, err.Error())
			}
			key, ok := keyToken.(string)
			if !ok {
				return invalid("INVALID_JSON", path, "object key must be a string")
			}
			if _, exists := seen[key]; exists {
				return invalid("DUPLICATE_JSON_KEY", appendJSONPath(path, key), "object key is duplicated")
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder, appendJSONPath(path, key), depth+1); err != nil {
				return err
			}
		}
		closeToken, err := decoder.Token()
		if err != nil || closeToken != json.Delim('}') {
			return invalid("INVALID_JSON", path, "object is not properly closed")
		}
	case '[':
		index := 0
		for decoder.More() {
			if err := scanJSONValue(decoder, fmt.Sprintf("%s[%d]", path, index), depth+1); err != nil {
				return err
			}
			index++
		}
		closeToken, err := decoder.Token()
		if err != nil || closeToken != json.Delim(']') {
			return invalid("INVALID_JSON", path, "array is not properly closed")
		}
	default:
		return invalid("INVALID_JSON", path, "unexpected JSON delimiter")
	}
	return nil
}

func validateUnicodeEscapes(data []byte) error {
	inString := false
	for i := 0; i < len(data); i++ {
		if !inString {
			if data[i] == '"' {
				inString = true
			}
			continue
		}
		if data[i] == '"' {
			inString = false
			continue
		}
		if data[i] != '\\' {
			continue
		}
		if i+1 >= len(data) {
			return invalid("INVALID_JSON", "document", "trailing escape in JSON string")
		}
		if data[i+1] != 'u' {
			i++
			continue
		}
		code, ok := parseHexRune(data, i+2)
		if !ok {
			return invalid("INVALID_JSON", "document", "malformed Unicode escape")
		}
		switch {
		case code >= 0xD800 && code <= 0xDBFF:
			if i+12 > len(data) || data[i+6] != '\\' || data[i+7] != 'u' {
				return invalid("INVALID_UTF8", "document", "high surrogate is not followed by a low surrogate")
			}
			low, valid := parseHexRune(data, i+8)
			if !valid || low < 0xDC00 || low > 0xDFFF {
				return invalid("INVALID_UTF8", "document", "high surrogate is not followed by a low surrogate")
			}
			i += 11
		case code >= 0xDC00 && code <= 0xDFFF:
			return invalid("INVALID_UTF8", "document", "unpaired low surrogate")
		default:
			i += 5
		}
	}
	return nil
}

func parseHexRune(data []byte, start int) (uint16, bool) {
	if start+4 > len(data) {
		return 0, false
	}
	var value uint16
	for _, char := range data[start : start+4] {
		value <<= 4
		switch {
		case char >= '0' && char <= '9':
			value |= uint16(char - '0')
		case char >= 'a' && char <= 'f':
			value |= uint16(char-'a') + 10
		case char >= 'A' && char <= 'F':
			value |= uint16(char-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}

func validateDocumentJSON(data []byte) error {
	root, err := strictObject(data, "document", []string{"schema_version", "sorts", "signatures", "parameters", "references", "closed", "expression", "limits"}, []string{"schema_version", "expression"})
	if err != nil {
		return err
	}
	var version int
	if err := json.Unmarshal(root["schema_version"], &version); err != nil {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "schema_version must be an integer")
	}
	if version != SchemaVersion {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "unsupported algebra document version")
	}
	if raw, ok := root["sorts"]; ok {
		items, err := strictArray(raw, "sorts", 512)
		if err != nil {
			return err
		}
		for i, item := range items {
			path := fmt.Sprintf("sorts[%d]", i)
			if _, err := strictObject(item, path, []string{"id", "label"}, []string{"id"}); err != nil {
				return err
			}
		}
	}
	if raw, ok := root["signatures"]; ok {
		items, err := strictArray(raw, "signatures", 512)
		if err != nil {
			return err
		}
		for i, item := range items {
			path := fmt.Sprintf("signatures[%d]", i)
			fields, err := strictObject(item, path, []string{"id", "label", "ports"}, []string{"id", "ports"})
			if err != nil {
				return err
			}
			ports, err := strictArray(fields["ports"], path+".ports", DefaultMaxPorts)
			if err != nil {
				return err
			}
			for j, port := range ports {
				portPath := fmt.Sprintf("%s.ports[%d]", path, j)
				portFields, err := strictObject(port, portPath, []string{"role", "type"}, []string{"role", "type"})
				if err != nil {
					return err
				}
				if err := validateSortTypeJSON(portFields["type"], portPath+".type"); err != nil {
					return err
				}
			}
		}
	}
	for _, name := range []string{"parameters", "references"} {
		if raw, ok := root[name]; ok {
			items, err := strictArray(raw, name, 512)
			if err != nil {
				return err
			}
			for i, item := range items {
				path := fmt.Sprintf("%s[%d]", name, i)
				fields, err := strictObject(item, path, []string{"id", "type"}, []string{"id", "type"})
				if err != nil {
					return err
				}
				if err := validateTypeJSON(fields["type"], path+".type"); err != nil {
					return err
				}
			}
		}
	}
	if raw, ok := root["limits"]; ok {
		if _, err := strictObject(raw, "limits", []string{"max_nodes", "max_depth", "max_type_depth", "max_symbols", "max_ports"}, nil); err != nil {
			return err
		}
	}
	return validateTermJSON(root["expression"], "expression")
}

func validateTypeJSON(data json.RawMessage, path string) error {
	base, err := strictObject(data, path, []string{"kind", "sort_id", "from", "to"}, []string{"kind"})
	if err != nil {
		return err
	}
	var kind TypeKind
	if err := json.Unmarshal(base["kind"], &kind); err != nil {
		return invalid("UNKNOWN_TYPE_KIND", path+".kind", "type kind must be a string")
	}
	var allowed, required []string
	switch kind {
	case TypeEntity, TypeEvent, TypeInterval, TypeProp:
		allowed, required = []string{"kind"}, []string{"kind"}
	case TypeNominal:
		allowed, required = []string{"kind", "sort_id"}, []string{"kind", "sort_id"}
	case TypeArrow:
		allowed, required = []string{"kind", "from", "to"}, []string{"kind", "from", "to"}
	default:
		return invalid("UNKNOWN_TYPE_KIND", path+".kind", "type kind is unsupported")
	}
	fields, err := strictObject(data, path, allowed, required)
	if err != nil {
		return err
	}
	if kind == TypeArrow {
		if err := validateTypeJSON(fields["from"], path+".from"); err != nil {
			return err
		}
		return validateTypeJSON(fields["to"], path+".to")
	}
	return nil
}

func validateSortTypeJSON(data json.RawMessage, path string) error {
	if err := validateTypeJSON(data, path); err != nil {
		return err
	}
	var shape struct {
		Kind TypeKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &shape); err != nil {
		return invalid("MALFORMED_TYPE", path, "port type must be an object")
	}
	if shape.Kind == TypeArrow {
		return invalid("INVALID_RELATION_PORT_TYPE", path, "relation port must have a base sort type")
	}
	return nil
}

func validateTermJSON(data json.RawMessage, path string) error {
	base, err := strictObject(data, path, []string{"kind", "id", "label", "type", "symbol", "args", "operands", "body", "function", "argument"}, []string{"kind"})
	if err != nil {
		return err
	}
	var kind TermKind
	if err := json.Unmarshal(base["kind"], &kind); err != nil {
		return invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind must be a string")
	}
	var allowed, required []string
	switch kind {
	case TermVar:
		allowed, required = []string{"kind", "id", "label"}, []string{"kind", "id"}
	case TermParam, TermRef, TermHole:
		allowed, required = []string{"kind", "id", "label", "type"}, []string{"kind", "id", "type"}
	case TermAtom:
		allowed, required = []string{"kind", "label", "symbol", "args"}, []string{"kind", "symbol", "args"}
	case TermAnd:
		allowed, required = []string{"kind", "label", "operands"}, []string{"kind", "operands"}
	case TermExists, TermLambda:
		allowed, required = []string{"kind", "id", "label", "type", "body"}, []string{"kind", "id", "type", "body"}
	case TermApply:
		allowed, required = []string{"kind", "label", "function", "argument"}, []string{"kind", "function", "argument"}
	default:
		return invalid("UNKNOWN_TERM_KIND", path+".kind", "term kind is unsupported")
	}
	fields, err := strictObject(data, path, allowed, required)
	if err != nil {
		return err
	}
	switch kind {
	case TermParam, TermRef, TermHole, TermExists, TermLambda:
		if err := validateTypeJSON(fields["type"], path+".type"); err != nil {
			return err
		}
	}
	switch kind {
	case TermAtom:
		args, err := strictArray(fields["args"], path+".args", DefaultMaxPorts)
		if err != nil {
			return err
		}
		for i, item := range args {
			argPath := fmt.Sprintf("%s.args[%d]", path, i)
			argument, err := strictObject(item, argPath, []string{"role", "value"}, []string{"role", "value"})
			if err != nil {
				return err
			}
			if err := validateTermJSON(argument["value"], argPath+".value"); err != nil {
				return err
			}
		}
	case TermAnd:
		operands, err := strictArray(fields["operands"], path+".operands", DefaultMaxNodes)
		if err != nil {
			return err
		}
		for i, operand := range operands {
			if err := validateTermJSON(operand, fmt.Sprintf("%s.operands[%d]", path, i)); err != nil {
				return err
			}
		}
	case TermExists, TermLambda:
		return validateTermJSON(fields["body"], path+".body")
	case TermApply:
		if err := validateTermJSON(fields["function"], path+".function"); err != nil {
			return err
		}
		return validateTermJSON(fields["argument"], path+".argument")
	}
	return nil
}

func strictObject(data []byte, path string, allowed, required []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, invalid("EXPECTED_OBJECT", path, "expected a non-null JSON object")
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := fields[key]
		if _, ok := allowedSet[key]; !ok {
			return nil, invalid("UNKNOWN_FIELD", appendJSONPath(path, key), "field is not allowed in this object variant")
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, invalid("NULL_FIELD", appendJSONPath(path, key), "field cannot be null")
		}
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return nil, invalid("MISSING_FIELD", appendJSONPath(path, key), "required field is missing")
		}
	}
	return fields, nil
}

func strictArray(data []byte, path string, maxItems int) ([]json.RawMessage, error) {
	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil || values == nil {
		return nil, invalid("EXPECTED_ARRAY", path, "expected a non-null JSON array")
	}
	if len(values) > maxItems {
		return nil, resourceLimit(path, fmt.Sprintf("array has %d items; maximum is %d", len(values), maxItems))
	}
	return values, nil
}

func appendJSONPath(path, key string) string {
	return path + "." + key
}

// DecodeArtifactExportRequest strictly decodes one bounded artifact export
// request. Metadata declarations are supplied separately by a trusted caller.
func DecodeArtifactExportRequest(data []byte) (ArtifactExportRequest, error) {
	if err := PreflightJSONLimit(data, MaxArtifactExportRequestBytes); err != nil {
		return ArtifactExportRequest{}, err
	}
	fields, err := strictObject(data, "request", []string{"schema_version", "document", "profile", "context_id", "context_dependencies", "provenance"}, []string{"schema_version", "document", "profile", "context_id", "context_dependencies", "provenance"})
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	version, err := rewriteJSONInt(fields["schema_version"], "request.schema_version", 1, 1)
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	document, err := DecodeDocumentJSON(fields["document"])
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	profile, err := decodeRewriteProfile(fields["profile"], "request.profile")
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	contextID, err := rewriteJSONString(fields["context_id"], "request.context_id", 256, false)
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	dependencies, err := decodeContextDependencies(fields["context_dependencies"], "request.context_dependencies")
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	provenanceRefs, err := decodeArtifactProvenance(fields["provenance"], "request.provenance")
	if err != nil {
		return ArtifactExportRequest{}, err
	}
	return ArtifactExportRequest{SchemaVersion: version, Document: document, Profile: profile, ContextID: contextID, ContextDependencies: dependencies, Provenance: provenanceRefs}, nil
}

// DecodeArtifactEnvelope strictly decodes one artifact's explicit versioned
// wire representation. It validates shape, not registry approval or trust.
func DecodeArtifactEnvelope(data []byte) (ArtifactEnvelope, error) {
	if err := PreflightJSONLimit(data, MaxArtifactEnvelopeBytes); err != nil {
		return ArtifactEnvelope{}, err
	}
	allowed := []string{"kind", "schema_version", "normalizer_version", "registry_snapshot_id", "registries", "profile", "context_id", "context_dependencies", "sorts", "signatures", "parameters", "references", "closed", "limits", "root", "nodes", "provenance", "resource_fingerprint", "meaning_fingerprint", "cache_key"}
	fields, err := strictObject(data, "artifact", allowed, allowed)
	if err != nil {
		return ArtifactEnvelope{}, err
	}
	var envelope ArtifactEnvelope
	if envelope.Kind, err = rewriteJSONString(fields["kind"], "artifact.kind", 64, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.SchemaVersion, err = rewriteJSONInt(fields["schema_version"], "artifact.schema_version", 1, ArtifactFormatVersion); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.NormalizerVersion, err = rewriteJSONString(fields["normalizer_version"], "artifact.normalizer_version", 256, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.RegistrySnapshotID, err = rewriteJSONString(fields["registry_snapshot_id"], "artifact.registry_snapshot_id", 256, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Registries, err = decodeRegistryPins(fields["registries"], "artifact.registries"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Profile, err = decodeRewriteProfile(fields["profile"], "artifact.profile"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.ContextID, err = rewriteJSONString(fields["context_id"], "artifact.context_id", 256, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.ContextDependencies, err = decodeContextDependencies(fields["context_dependencies"], "artifact.context_dependencies"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Sorts, err = decodeArtifactSorts(fields["sorts"], "artifact.sorts"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Signatures, err = decodeArtifactSignatures(fields["signatures"], "artifact.signatures"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Parameters, err = decodeDeclarations(fields["parameters"], "artifact.parameters"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.References, err = decodeDeclarations(fields["references"], "artifact.references"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Closed, err = rewriteJSONBool(fields["closed"], "artifact.closed"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Limits, err = decodeArtifactLimits(fields["limits"], "artifact.limits"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Root, err = decodeNodeID(fields["root"], "artifact.root"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Nodes, err = decodeArtifactNodes(fields["nodes"]); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.Provenance, err = decodeArtifactProvenance(fields["provenance"], "artifact.provenance"); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.ResourceFingerprint, err = rewriteJSONString(fields["resource_fingerprint"], "artifact.resource_fingerprint", 256, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.MeaningFingerprint, err = rewriteJSONString(fields["meaning_fingerprint"], "artifact.meaning_fingerprint", 256, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if envelope.CacheKey, err = rewriteJSONString(fields["cache_key"], "artifact.cache_key", 256, false); err != nil {
		return ArtifactEnvelope{}, err
	}
	if _, err := NormalizeLimits(envelope.Limits); err != nil {
		return ArtifactEnvelope{}, err
	}
	if err := validateReplayHeader(envelope); err != nil {
		return ArtifactEnvelope{}, err
	}
	return envelope, nil
}

// DecodeArtifactReplayRequest strictly decodes a replay request. Expected
// profile/context/provenance fields remain independent of the embedded claims.
func DecodeArtifactReplayRequest(data []byte) (ArtifactReplayRequest, error) {
	if err := PreflightJSONLimit(data, MaxArtifactReplayRequestBytes); err != nil {
		return ArtifactReplayRequest{}, err
	}
	fields, err := strictObject(data, "request", []string{"schema_version", "profile", "context_id", "context_dependencies", "provenance", "limits", "artifact"}, []string{"schema_version", "profile", "context_id", "context_dependencies", "provenance", "artifact"})
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	version, err := rewriteJSONInt(fields["schema_version"], "request.schema_version", 1, 1)
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	profile, err := decodeRewriteProfile(fields["profile"], "request.profile")
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	contextID, err := rewriteJSONString(fields["context_id"], "request.context_id", 256, false)
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	dependencies, err := decodeContextDependencies(fields["context_dependencies"], "request.context_dependencies")
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	provenanceRefs, err := decodeArtifactProvenance(fields["provenance"], "request.provenance")
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	var limits Limits
	if raw, exists := fields["limits"]; exists {
		limits, err = decodeArtifactLimits(raw, "request.limits")
		if err != nil {
			return ArtifactReplayRequest{}, err
		}
		if _, err := NormalizeLimits(limits); err != nil {
			return ArtifactReplayRequest{}, err
		}
	}
	envelope, err := DecodeArtifactEnvelope(fields["artifact"])
	if err != nil {
		return ArtifactReplayRequest{}, err
	}
	return ArtifactReplayRequest{SchemaVersion: version, Profile: profile, ContextID: contextID, ContextDependencies: dependencies, Provenance: provenanceRefs, Limits: limits, Artifact: envelope}, nil
}

func decodeRegistryPins(data []byte, path string) ([]RegistryPin, error) {
	items, err := strictArray(data, path, 16)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, invalid("MISSING_REGISTRY_PINS", path, "at least one registry pin is required")
	}
	result := make([]RegistryPin, len(items))
	seen := make(map[string]struct{}, len(items))
	for i, raw := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		fields, err := strictObject(raw, itemPath, []string{"id", "version", "resource_fingerprint", "formal_denotation_id"}, []string{"id", "version", "resource_fingerprint", "formal_denotation_id"})
		if err != nil {
			return nil, err
		}
		values := []*string{&result[i].ID, &result[i].Version, &result[i].ResourceFingerprint, &result[i].FormalDenotationID}
		keys := []string{"id", "version", "resource_fingerprint", "formal_denotation_id"}
		for j, key := range keys {
			value, e := rewriteJSONString(fields[key], itemPath+"."+key, 256, false)
			if e != nil {
				return nil, e
			}
			*values[j] = value
		}
		if _, exists := seen[result[i].ID]; exists {
			return nil, invalid("DUPLICATE_REGISTRY_PIN", itemPath+".id", "registry pin ID is duplicated")
		}
		seen[result[i].ID] = struct{}{}
	}
	return result, nil
}

func decodeContextDependencies(data []byte, path string) ([]ContextDependency, error) {
	items, err := strictArray(data, path, DefaultMaxSymbols)
	if err != nil {
		return nil, err
	}
	result := make([]ContextDependency, len(items))
	seen := make(map[string]struct{}, len(items))
	for i, raw := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		fields, err := strictObject(raw, itemPath, []string{"id", "fingerprint"}, []string{"id", "fingerprint"})
		if err != nil {
			return nil, err
		}
		if result[i].ID, err = rewriteJSONString(fields["id"], itemPath+".id", 256, false); err != nil {
			return nil, err
		}
		if result[i].Fingerprint, err = rewriteJSONString(fields["fingerprint"], itemPath+".fingerprint", 256, false); err != nil {
			return nil, err
		}
		if _, exists := seen[result[i].ID]; exists {
			return nil, invalid("DUPLICATE_CONTEXT_DEPENDENCY", itemPath+".id", "context dependency ID is duplicated")
		}
		seen[result[i].ID] = struct{}{}
	}
	return result, nil
}

func decodeArtifactSorts(data []byte, path string) ([]SortDeclaration, error) {
	items, err := strictArray(data, path, DefaultMaxSymbols)
	if err != nil {
		return nil, err
	}
	result := make([]SortDeclaration, len(items))
	for i, raw := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		fields, err := strictObject(raw, itemPath, []string{"id", "label"}, []string{"id"})
		if err != nil {
			return nil, err
		}
		if result[i].ID, err = rewriteJSONString(fields["id"], itemPath+".id", 256, false); err != nil {
			return nil, err
		}
		if rawLabel, ok := fields["label"]; ok {
			if result[i].Label, err = rewriteJSONString(rawLabel, itemPath+".label", 1024, true); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func decodeArtifactSignatures(data []byte, path string) ([]Signature, error) {
	items, err := strictArray(data, path, DefaultMaxSymbols)
	if err != nil {
		return nil, err
	}
	result := make([]Signature, len(items))
	for i, raw := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		fields, err := strictObject(raw, itemPath, []string{"id", "label", "ports"}, []string{"id", "ports"})
		if err != nil {
			return nil, err
		}
		if result[i].ID, err = rewriteJSONString(fields["id"], itemPath+".id", 256, false); err != nil {
			return nil, err
		}
		if rawLabel, ok := fields["label"]; ok {
			if result[i].Label, err = rewriteJSONString(rawLabel, itemPath+".label", 1024, true); err != nil {
				return nil, err
			}
		}
		ports, err := strictArray(fields["ports"], itemPath+".ports", DefaultMaxPorts)
		if err != nil {
			return nil, err
		}
		result[i].Ports = make([]Port, len(ports))
		for j, portRaw := range ports {
			portPath := fmt.Sprintf("%s.ports[%d]", itemPath, j)
			portFields, err := strictObject(portRaw, portPath, []string{"role", "type"}, []string{"role", "type"})
			if err != nil {
				return nil, err
			}
			role, err := rewriteJSONString(portFields["role"], portPath+".role", 256, false)
			if err != nil {
				return nil, err
			}
			if err := validateSortTypeJSON(portFields["type"], portPath+".type"); err != nil {
				return nil, err
			}
			typeValue, err := DecodeTypeJSON(portFields["type"], portPath+".type")
			if err != nil {
				return nil, err
			}
			result[i].Ports[j] = Port{Role: role, Type: typeValue}
		}
	}
	return result, nil
}

func decodeDeclarations(data []byte, path string) ([]Declaration, error) {
	items, err := strictArray(data, path, DefaultMaxSymbols)
	if err != nil {
		return nil, err
	}
	result := make([]Declaration, len(items))
	for i, raw := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		fields, err := strictObject(raw, itemPath, []string{"id", "type"}, []string{"id", "type"})
		if err != nil {
			return nil, err
		}
		if result[i].ID, err = rewriteJSONString(fields["id"], itemPath+".id", 256, false); err != nil {
			return nil, err
		}
		if result[i].Type, err = DecodeTypeJSON(fields["type"], itemPath+".type"); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func decodeArtifactLimits(data []byte, path string) (Limits, error) {
	fields, err := strictObject(data, path, []string{"max_nodes", "max_depth", "max_type_depth", "max_symbols", "max_ports"}, nil)
	if err != nil {
		return Limits{}, err
	}
	limits := Limits{}
	values := []struct {
		name string
		cap  int
		dest *int
	}{
		{"max_nodes", DefaultMaxNodes, &limits.MaxNodes}, {"max_depth", DefaultMaxDepth, &limits.MaxDepth},
		{"max_type_depth", DefaultMaxTypeDepth, &limits.MaxTypeDepth}, {"max_symbols", DefaultMaxSymbols, &limits.MaxSymbols},
		{"max_ports", DefaultMaxPorts, &limits.MaxPorts},
	}
	for _, item := range values {
		if raw, exists := fields[item.name]; exists {
			value, err := rewriteJSONInt(raw, path+"."+item.name, 0, item.cap)
			if err != nil {
				return Limits{}, err
			}
			*item.dest = value
		}
	}
	return limits, nil
}

func decodeNodeID(data []byte, path string) (NodeID, error) {
	value, err := rewriteJSONInt(data, path, 1, DefaultMaxNodes)
	return NodeID(value), err
}

func decodeArtifactNodes(data []byte) ([]Node, error) {
	items, err := strictArray(data, "artifact.nodes", DefaultMaxNodes)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, invalid("EMPTY_ARTIFACT_NODES", "artifact.nodes", "artifact requires at least one node")
	}
	result := make([]Node, len(items))
	for i, raw := range items {
		path := fmt.Sprintf("artifact.nodes[%d]", i)
		base, err := strictObject(raw, path, []string{"node_id", "kind", "label", "identity", "type", "symbol", "args", "operands", "body", "function", "argument"}, []string{"node_id", "kind"})
		if err != nil {
			return nil, err
		}
		id, err := decodeNodeID(base["node_id"], path+".node_id")
		if err != nil {
			return nil, err
		}
		kindText, err := rewriteJSONString(base["kind"], path+".kind", 32, false)
		if err != nil {
			return nil, err
		}
		node := Node{NodeID: id, Kind: TermKind(kindText)}
		if rawLabel, ok := base["label"]; ok {
			if node.Label, err = rewriteJSONString(rawLabel, path+".label", 1024, true); err != nil {
				return nil, err
			}
		}
		var allowed, required []string
		switch node.Kind {
		case TermVar:
			allowed, required = []string{"node_id", "kind", "label", "identity"}, []string{"node_id", "kind", "identity"}
		case TermParam, TermRef, TermHole:
			allowed, required = []string{"node_id", "kind", "label", "identity", "type"}, []string{"node_id", "kind", "identity", "type"}
		case TermAtom:
			allowed, required = []string{"node_id", "kind", "label", "symbol", "args"}, []string{"node_id", "kind", "symbol", "args"}
		case TermAnd:
			allowed, required = []string{"node_id", "kind", "label", "operands"}, []string{"node_id", "kind", "operands"}
		case TermExists, TermLambda:
			allowed, required = []string{"node_id", "kind", "label", "identity", "type", "body"}, []string{"node_id", "kind", "identity", "type", "body"}
		case TermApply:
			allowed, required = []string{"node_id", "kind", "label", "function", "argument"}, []string{"node_id", "kind", "function", "argument"}
		default:
			return nil, invalid("UNKNOWN_TERM_KIND", path+".kind", "artifact node kind is unsupported")
		}
		fields, err := strictObject(raw, path, allowed, required)
		if err != nil {
			return nil, err
		}
		if value, ok := fields["identity"]; ok {
			if node.Identity, err = rewriteJSONString(value, path+".identity", 256, false); err != nil {
				return nil, err
			}
		}
		if value, ok := fields["type"]; ok {
			typeValue, typeErr := DecodeTypeJSON(value, path+".type")
			if typeErr != nil {
				return nil, typeErr
			}
			node.Type = &typeValue
		}
		switch node.Kind {
		case TermAtom:
			if node.Symbol, err = rewriteJSONString(fields["symbol"], path+".symbol", 256, false); err != nil {
				return nil, err
			}
			args, err := strictArray(fields["args"], path+".args", DefaultMaxPorts)
			if err != nil {
				return nil, err
			}
			node.Args = make([]NodeArgument, len(args))
			for j, argRaw := range args {
				argPath := fmt.Sprintf("%s.args[%d]", path, j)
				argFields, err := strictObject(argRaw, argPath, []string{"role", "value"}, []string{"role", "value"})
				if err != nil {
					return nil, err
				}
				role, err := rewriteJSONString(argFields["role"], argPath+".role", 256, false)
				if err != nil {
					return nil, err
				}
				value, err := decodeNodeID(argFields["value"], argPath+".value")
				if err != nil {
					return nil, err
				}
				node.Args[j] = NodeArgument{Role: role, Value: value}
			}
		case TermAnd:
			operands, err := strictArray(fields["operands"], path+".operands", DefaultMaxNodes)
			if err != nil {
				return nil, err
			}
			if len(operands) == 0 {
				return nil, invalid("EMPTY_AND", path+".operands", "And requires at least one operand")
			}
			node.Operands = make([]NodeID, len(operands))
			for j, rawID := range operands {
				if node.Operands[j], err = decodeNodeID(rawID, fmt.Sprintf("%s.operands[%d]", path, j)); err != nil {
					return nil, err
				}
			}
		case TermExists, TermLambda:
			if node.Body, err = decodeNodeID(fields["body"], path+".body"); err != nil {
				return nil, err
			}
		case TermApply:
			if node.Function, err = decodeNodeID(fields["function"], path+".function"); err != nil {
				return nil, err
			}
			if node.Argument, err = decodeNodeID(fields["argument"], path+".argument"); err != nil {
				return nil, err
			}
		}
		result[i] = node
	}
	return result, nil
}

func decodeArtifactProvenance(data []byte, path string) ([]provenance.SourceRef, error) {
	items, err := strictArray(data, path, DefaultMaxSymbols)
	if err != nil {
		return nil, err
	}
	result := make([]provenance.SourceRef, len(items))
	seen := make(map[string]struct{}, len(items))
	for i, raw := range items {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		fields, err := strictObject(raw, itemPath, []string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}, []string{"id", "version", "license"})
		if err != nil {
			return nil, err
		}
		limits := map[string]int{"id": provenance.MaxSourceIDBytes, "path": provenance.MaxSourceLocationBytes, "uri": provenance.MaxSourceLocationBytes, "version": provenance.MaxSourceVersionBytes, "license": provenance.MaxSourceLicenseBytes, "author": provenance.MaxSourceAuthorBytes, "checksum_sha256": provenance.SHA256HexBytes}
		values := map[string]*string{"id": &result[i].ID, "path": &result[i].Path, "uri": &result[i].URI, "version": &result[i].Version, "license": &result[i].License, "author": &result[i].Author, "checksum_sha256": &result[i].ChecksumSHA256}
		for key, rawValue := range fields {
			value, err := rewriteJSONString(rawValue, itemPath+"."+key, limits[key], key == "path" || key == "uri" || key == "author" || key == "checksum_sha256")
			if err != nil {
				return nil, err
			}
			*values[key] = value
		}
		if err := result[i].Validate(); err != nil {
			return nil, invalid("INVALID_PROVENANCE", itemPath, err.Error())
		}
		if _, exists := seen[result[i].ID]; exists {
			return nil, invalid("DUPLICATE_PROVENANCE_ID", itemPath+".id", "provenance ID is duplicated")
		}
		seen[result[i].ID] = struct{}{}
	}
	return result, nil
}

func (request ArtifactExportRequest) MarshalJSON() ([]byte, error) {
	if request.SchemaVersion != 1 {
		return nil, invalid("INVALID_SCHEMA_VERSION", "request.schema_version", "schema version must be 1")
	}
	if err := ValidateRewriteProfile(request.Profile); err != nil {
		return nil, err
	}
	if !validText(request.ContextID, 256) || len(request.ContextDependencies) > DefaultMaxSymbols || len(request.Provenance) > DefaultMaxSymbols {
		return nil, invalid("INVALID_ARTIFACT_REQUEST", "request", "context or request metadata exceeds bounds")
	}
	documentWireValue, err := rewriteDocumentWire(request.Document)
	if err != nil {
		return nil, err
	}
	dependencies, err := encodeContextDependencies(request.ContextDependencies)
	if err != nil {
		return nil, err
	}
	provenanceValues, err := encodeArtifactProvenance(request.Provenance)
	if err != nil {
		return nil, err
	}
	wire := map[string]any{"schema_version": request.SchemaVersion, "document": documentWireValue, "profile": rewriteProfileWire(request.Profile), "context_id": request.ContextID, "context_dependencies": dependencies, "provenance": provenanceValues}
	return marshalBoundedWire(wire, MaxArtifactExportRequestBytes, "request")
}

func (request ArtifactReplayRequest) MarshalJSON() ([]byte, error) {
	if request.SchemaVersion != 1 {
		return nil, invalid("INVALID_SCHEMA_VERSION", "request.schema_version", "schema version must be 1")
	}
	if err := ValidateRewriteProfile(request.Profile); err != nil {
		return nil, err
	}
	if !validText(request.ContextID, 256) {
		return nil, invalid("INVALID_CONTEXT_ID", "request.context_id", "context ID must be bounded nonempty text")
	}
	dependencies, err := encodeContextDependencies(request.ContextDependencies)
	if err != nil {
		return nil, err
	}
	provenanceValues, err := encodeArtifactProvenance(request.Provenance)
	if err != nil {
		return nil, err
	}
	wire := map[string]any{"schema_version": request.SchemaVersion, "profile": rewriteProfileWire(request.Profile), "context_id": request.ContextID, "context_dependencies": dependencies, "provenance": provenanceValues, "artifact": request.Artifact}
	if request.Limits != (Limits{}) {
		if _, err := NormalizeLimits(request.Limits); err != nil {
			return nil, err
		}
		wire["limits"] = request.Limits
	}
	return marshalBoundedWire(wire, MaxArtifactReplayRequestBytes, "request")
}

func encodeContextDependencies(values []ContextDependency) ([]ContextDependency, error) {
	if len(values) > DefaultMaxSymbols {
		return nil, resourceLimit("context_dependencies", "context dependency count exceeds 512")
	}
	result := append([]ContextDependency{}, values...)
	seen := make(map[string]struct{}, len(result))
	for i, value := range result {
		if !validText(value.ID, 256) || !validText(value.Fingerprint, 256) {
			return nil, invalid("INVALID_CONTEXT_DEPENDENCY", fmt.Sprintf("context_dependencies[%d]", i), "dependency ID and fingerprint must be bounded nonempty text")
		}
		if _, exists := seen[value.ID]; exists {
			return nil, invalid("DUPLICATE_CONTEXT_DEPENDENCY", fmt.Sprintf("context_dependencies[%d].id", i), "dependency ID is duplicated")
		}
		seen[value.ID] = struct{}{}
	}
	return result, nil
}

func encodeArtifactProvenance(values []provenance.SourceRef) ([]provenance.SourceRef, error) {
	if len(values) > DefaultMaxSymbols {
		return nil, resourceLimit("provenance", "provenance count exceeds 512")
	}
	result := append([]provenance.SourceRef{}, values...)
	seen := make(map[string]struct{}, len(result))
	for i, source := range result {
		if err := source.Validate(); err != nil {
			return nil, invalid("INVALID_PROVENANCE", fmt.Sprintf("provenance[%d]", i), err.Error())
		}
		if _, exists := seen[source.ID]; exists {
			return nil, invalid("DUPLICATE_PROVENANCE_ID", fmt.Sprintf("provenance[%d].id", i), "provenance ID is duplicated")
		}
		seen[source.ID] = struct{}{}
	}
	return result, nil
}

func (envelope ArtifactEnvelope) MarshalJSON() ([]byte, error) {
	if err := validateArtifactEnvelopeLocal(envelope); err != nil {
		return nil, err
	}
	return artifactEnvelopeBytes(envelope)
}

// validateArtifactEnvelopeLocal checks wire integrity against declarations
// carried by the envelope itself. It is structural only and is not trusted
// metadata admission; ReplayArtifact still requires an independent environment.
func validateArtifactEnvelopeLocal(envelope ArtifactEnvelope) error {
	if err := validateReplayHeader(envelope); err != nil {
		return err
	}
	if err := validateCanonicalArtifactOrder(envelope); err != nil {
		return err
	}
	for i, pin := range envelope.Registries {
		if !validText(pin.ID, 256) || !validText(pin.Version, 256) || !validText(pin.ResourceFingerprint, 256) || !validText(pin.FormalDenotationID, 256) {
			return invalid("INVALID_REGISTRY_PIN", fmt.Sprintf("registries[%d]", i), "registry pin fields must be bounded nonempty text")
		}
	}
	for i, dependency := range envelope.ContextDependencies {
		if !validText(dependency.ID, 256) || !validText(dependency.Fingerprint, 256) {
			return invalid("INVALID_CONTEXT_DEPENDENCY", fmt.Sprintf("context_dependencies[%d]", i), "context dependency fields must be bounded nonempty text")
		}
	}
	seenProvenance := make(map[string]struct{}, len(envelope.Provenance))
	for i, source := range envelope.Provenance {
		if err := source.Validate(); err != nil {
			return invalid("INVALID_PROVENANCE", fmt.Sprintf("provenance[%d]", i), err.Error())
		}
		if _, exists := seenProvenance[source.ID]; exists {
			return invalid("DUPLICATE_PROVENANCE_ID", fmt.Sprintf("provenance[%d].id", i), "provenance ID is duplicated")
		}
		seenProvenance[source.ID] = struct{}{}
	}
	limits, err := NormalizeLimits(envelope.Limits)
	if err != nil {
		return err
	}
	if len(envelope.Nodes) > limits.MaxNodes {
		return resourceLimit("nodes", "artifact node count exceeds its declared limits")
	}
	if err := preflightReplayNodes(envelope.Nodes, limits); err != nil {
		return err
	}
	if err := validateReplayDeclarations(envelope); err != nil {
		return err
	}
	context := Document{SchemaVersion: SchemaVersion, Sorts: envelope.Sorts, Signatures: envelope.Signatures, Parameters: envelope.Parameters, References: envelope.References, Closed: envelope.Closed, Limits: limits}
	arena, err := NewArena(context, envelope.Root, envelope.Nodes)
	if err != nil {
		return err
	}
	resourceBytes, err := artifactResourceBytes(envelope)
	if err != nil {
		return err
	}
	if len(resourceBytes) > MaxArtifactBytes {
		return resourceLimit("artifact", "canonical artifact resource descriptor exceeds 8 MiB")
	}
	document, err := arena.Document()
	if err != nil {
		return err
	}
	meaningBytes, err := artifactMeaningBytes(envelope, document.Expression)
	if err != nil {
		return err
	}
	if artifactDigest("golingua-artifact-resource-v1\x00", resourceBytes) != envelope.ResourceFingerprint ||
		artifactDigest("golingua-artifact-meaning-v1\x00", meaningBytes) != envelope.MeaningFingerprint ||
		artifactDigest("golingua-artifact-cache-v1\x00", meaningBytes) != envelope.CacheKey {
		return invalid("ARTIFACT_RECEIPT_MISMATCH", "artifact", "artifact receipts do not match the structurally checked envelope")
	}
	payload, err := artifactEnvelopeBytes(envelope)
	if err != nil {
		return err
	}
	if len(payload) > MaxArtifactBytes {
		return resourceLimit("artifact", "complete artifact envelope exceeds 8 MiB")
	}
	return nil
}

func validateCanonicalArtifactOrder(envelope ArtifactEnvelope) error {
	check := func(ids []string, path string) error {
		for i := 1; i < len(ids); i++ {
			if ids[i-1] >= ids[i] {
				return invalid("NONCANONICAL_ARTIFACT_ORDER", fmt.Sprintf("%s[%d]", path, i), "artifact records must use strict ID order")
			}
		}
		return nil
	}
	registryIDs := make([]string, len(envelope.Registries))
	for i, value := range envelope.Registries {
		registryIDs[i] = value.ID
	}
	if err := check(registryIDs, "registries"); err != nil {
		return err
	}
	dependencyIDs := make([]string, len(envelope.ContextDependencies))
	for i, value := range envelope.ContextDependencies {
		dependencyIDs[i] = value.ID
	}
	if err := check(dependencyIDs, "context_dependencies"); err != nil {
		return err
	}
	provenanceIDs := make([]string, len(envelope.Provenance))
	for i, value := range envelope.Provenance {
		provenanceIDs[i] = value.ID
	}
	if err := check(provenanceIDs, "provenance"); err != nil {
		return err
	}
	sortIDs := make([]string, len(envelope.Sorts))
	for i, value := range envelope.Sorts {
		sortIDs[i] = value.ID
	}
	if err := check(sortIDs, "sorts"); err != nil {
		return err
	}
	signatureIDs := make([]string, len(envelope.Signatures))
	for i, value := range envelope.Signatures {
		signatureIDs[i] = value.ID
	}
	if err := check(signatureIDs, "signatures"); err != nil {
		return err
	}
	if !strictlySortedDeclarations(envelope.Parameters) || !strictlySortedDeclarations(envelope.References) {
		return invalid("NONCANONICAL_ARTIFACT_ORDER", "declarations", "parameter and reference declarations must use strict ID order")
	}
	return nil
}
