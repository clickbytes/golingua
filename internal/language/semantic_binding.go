package language

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/synthesis"
	"github.com/clickbytes/golingua/internal/valency"
)

// SemanticRoleReference supplies one caller-declared reference for a semantic
// role in the explicitly selected sense/frame pair.
type SemanticRoleReference struct {
	Role        string `json:"role"`
	ReferenceID string `json:"reference_id"`
}

// SemanticPortReference binds an otherwise unmapped target port directly.
type SemanticPortReference struct {
	PortID      string `json:"port_id"`
	ReferenceID string `json:"reference_id"`
}

// SemanticBindingRequest names every selection and input that may affect
// instantiation. A nil BindingIndex is valid only when exactly one tagged
// target exists for SenseID.
type SemanticBindingRequest struct {
	SenseID        ids.SenseID                      `json:"sense_id"`
	FrameID        valency.FrameID                  `json:"frame_id"`
	BindingIndex   *int                             `json:"binding_index,omitempty"`
	ContextID      string                           `json:"context_id"`
	References     []algebra.Declaration            `json:"references"`
	RequestContext *algebra.RequestContext          `json:"request_context,omitempty"`
	RoleReferences []SemanticRoleReference          `json:"role_references"`
	PortReferences []SemanticPortReference          `json:"port_references"`
	Verification   synthesis.GoalVerificationLimits `json:"verification_limits,omitempty"`
}

const MaxSemanticBindingRequestBytes = 1 << 20

// DecodeSemanticBindingRequest strictly decodes request identifiers as strings
// before converting them to typed IDs; ids structs do not define a wire codec.
func DecodeSemanticBindingRequest(data []byte) (SemanticBindingRequest, error) {
	if err := algebra.PreflightJSONLimit(data, MaxSemanticBindingRequestBytes); err != nil {
		return SemanticBindingRequest{}, err
	}
	fields, err := languageStrictObject(data, "request", []string{"sense_id", "frame_id", "binding_index", "context_id", "references", "request_context", "role_references", "port_references", "verification_limits"}, []string{"sense_id", "frame_id", "context_id", "references", "role_references", "port_references"})
	if err != nil {
		return SemanticBindingRequest{}, err
	}
	var senseText, frameID, contextID string
	for key, target := range map[string]*string{"sense_id": &senseText, "frame_id": &frameID, "context_id": &contextID} {
		if err := languageStrictValue(fields[key], target); err != nil {
			return SemanticBindingRequest{}, invalid("INVALID_BINDING_REQUEST_FIELD", key, "field must be a non-null string")
		}
	}
	senseID, err := ids.ParseSenseID(senseText)
	if err != nil {
		return SemanticBindingRequest{}, invalid("INVALID_SENSE_ID", "sense_id", "sense_id must be a typed sense identifier")
	}
	if !validSemanticPackageText(frameID, 128) || !validSemanticPackageText(contextID, algebra.MaxContextIDBytes) {
		return SemanticBindingRequest{}, invalid("INVALID_BINDING_REQUEST_ID", "request", "frame and context IDs must be bounded")
	}
	request := SemanticBindingRequest{SenseID: senseID, FrameID: valency.FrameID(frameID), ContextID: contextID, References: []algebra.Declaration{}, RoleReferences: []SemanticRoleReference{}, PortReferences: []SemanticPortReference{}}
	if raw, ok := fields["binding_index"]; ok {
		var index int
		if err := languageStrictValue(raw, &index); err != nil {
			return SemanticBindingRequest{}, invalid("INVALID_BINDING_INDEX", "binding_index", "binding_index must be a non-null integer")
		}
		request.BindingIndex = &index
	}
	if raw, ok := fields["request_context"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return SemanticBindingRequest{}, invalid("INVALID_REQUEST_CONTEXT", "request_context", "request_context cannot be null when present")
		}
		context, err := algebra.DecodeRequestContextJSON(raw)
		if err != nil {
			return SemanticBindingRequest{}, err
		}
		if context.ID != contextID {
			return SemanticBindingRequest{}, invalid("REQUEST_CONTEXT_ID_MISMATCH", "request_context.id", "request-context ID must match context_id")
		}
		request.RequestContext = &context
	}
	if request.References, err = decodeSemanticRequestReferences(fields["references"]); err != nil {
		return SemanticBindingRequest{}, err
	}
	if request.RoleReferences, err = decodeSemanticRoleReferences(fields["role_references"]); err != nil {
		return SemanticBindingRequest{}, err
	}
	if request.PortReferences, err = decodeSemanticPortReferences(fields["port_references"]); err != nil {
		return SemanticBindingRequest{}, err
	}
	if raw, ok := fields["verification_limits"]; ok {
		request.Verification, err = decodeSemanticVerificationLimits(raw)
		if err != nil {
			return SemanticBindingRequest{}, err
		}
	}
	return request, nil
}

func decodeSemanticRequestReferences(raw []byte) ([]algebra.Declaration, error) {
	items, err := languageStrictArray(raw, "references", algebra.DefaultMaxSymbols)
	if err != nil {
		return nil, err
	}
	result := make([]algebra.Declaration, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("references[%d]", i)
		fields, err := languageStrictObject(item, path, []string{"id", "type", "temporal_frame", "reference_time_id", "multiplicity"}, []string{"id", "type"})
		if err != nil {
			return nil, err
		}
		var row algebra.Declaration
		if err := languageStrictValue(fields["id"], &row.ID); err != nil || !validSemanticPackageText(row.ID, 256) {
			return nil, invalid("INVALID_REFERENCE_ID", path+".id", "reference ID must be a bounded non-null string")
		}
		row.Type, err = algebra.DecodeTypeJSON(fields["type"], path+".type")
		if err != nil {
			return nil, err
		}
		for key, target := range map[string]*string{"temporal_frame": &row.TemporalFrame, "reference_time_id": &row.ReferenceTimeID, "multiplicity": &row.Multiplicity} {
			if rawValue, ok := fields[key]; ok {
				if err := languageStrictValue(rawValue, target); err != nil {
					return nil, invalid("INVALID_REFERENCE_ATTRIBUTE", path+"."+key, "attribute must be a non-null string")
				}
			}
		}
		result = append(result, row)
	}
	return result, nil
}

func decodeSemanticRoleReferences(raw []byte) ([]SemanticRoleReference, error) {
	items, err := languageStrictArray(raw, "role_references", algebra.DefaultMaxPorts)
	if err != nil {
		return nil, err
	}
	out := make([]SemanticRoleReference, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("role_references[%d]", i)
		f, err := languageStrictObject(item, path, []string{"role", "reference_id"}, []string{"role", "reference_id"})
		if err != nil {
			return nil, err
		}
		var row SemanticRoleReference
		if err := languageStrictValue(f["role"], &row.Role); err != nil {
			return nil, invalid("INVALID_ROLE_REFERENCE", path+".role", "role must be a non-null string")
		}
		if err := languageStrictValue(f["reference_id"], &row.ReferenceID); err != nil {
			return nil, invalid("INVALID_ROLE_REFERENCE", path+".reference_id", "reference_id must be a non-null string")
		}
		out = append(out, row)
	}
	return out, nil
}

func decodeSemanticPortReferences(raw []byte) ([]SemanticPortReference, error) {
	items, err := languageStrictArray(raw, "port_references", algebra.DefaultMaxPorts)
	if err != nil {
		return nil, err
	}
	out := make([]SemanticPortReference, 0, len(items))
	for i, item := range items {
		path := fmt.Sprintf("port_references[%d]", i)
		f, err := languageStrictObject(item, path, []string{"port_id", "reference_id"}, []string{"port_id", "reference_id"})
		if err != nil {
			return nil, err
		}
		var row SemanticPortReference
		if err := languageStrictValue(f["port_id"], &row.PortID); err != nil {
			return nil, invalid("INVALID_PORT_REFERENCE", path+".port_id", "port_id must be a non-null string")
		}
		if err := languageStrictValue(f["reference_id"], &row.ReferenceID); err != nil {
			return nil, invalid("INVALID_PORT_REFERENCE", path+".reference_id", "reference_id must be a non-null string")
		}
		out = append(out, row)
	}
	return out, nil
}

func decodeSemanticVerificationLimits(raw []byte) (synthesis.GoalVerificationLimits, error) {
	f, err := languageStrictObject(raw, "verification_limits", []string{"max_certificate_bytes", "max_replay_steps", "max_work_nodes", "term_limits"}, nil)
	if err != nil {
		return synthesis.GoalVerificationLimits{}, err
	}
	var out synthesis.GoalVerificationLimits
	for key, target := range map[string]*int{"max_certificate_bytes": &out.MaxCertificateBytes, "max_replay_steps": &out.MaxReplaySteps, "max_work_nodes": &out.MaxWorkNodes} {
		if value, ok := f[key]; ok {
			if err := languageStrictValue(value, target); err != nil {
				return out, invalid("INVALID_VERIFICATION_LIMIT", "verification_limits."+key, "limit must be a non-null integer")
			}
		}
	}
	if value, ok := f["term_limits"]; ok {
		nested, err := languageStrictObject(value, "verification_limits.term_limits", []string{"max_nodes", "max_depth", "max_type_depth", "max_symbols", "max_ports"}, nil)
		if err != nil {
			return out, err
		}
		for key, target := range map[string]*int{"max_nodes": &out.TermLimits.MaxNodes, "max_depth": &out.TermLimits.MaxDepth, "max_type_depth": &out.TermLimits.MaxTypeDepth, "max_symbols": &out.TermLimits.MaxSymbols, "max_ports": &out.TermLimits.MaxPorts} {
			if rawValue, ok := nested[key]; ok {
				if err := languageStrictValue(rawValue, target); err != nil {
					return out, invalid("INVALID_VERIFICATION_LIMIT", "verification_limits.term_limits."+key, "limit must be a non-null integer")
				}
			}
		}
	}
	if _, err := synthesis.NormalizeGoalVerificationLimits(out); err != nil {
		return synthesis.GoalVerificationLimits{}, err
	}
	return out, nil
}

// SemanticSymbolInstantiation contains a checked anonymous expression. An
// open interface is an explicit function boundary, not inferred entities.
type SemanticSymbolInstantiation struct {
	Document  algebra.Document          `json:"document"`
	Checked   algebra.CheckedExpression `json:"checked"`
	Interface []algebra.PortBinding     `json:"interface"`
}

// MarshalJSON uses the strict algebra document codec for the embedded
// canonical document so nil declaration slices retain the required wire
// arrays and the document remains directly readable by algebra check.
func (result SemanticSymbolInstantiation) MarshalJSON() ([]byte, error) {
	document, err := algebra.Encode(result.Document)
	if err != nil {
		return nil, err
	}
	interfacePorts := append([]algebra.PortBinding{}, result.Interface...)
	return json.Marshal(struct {
		Document  json.RawMessage           `json:"document"`
		Checked   algebra.CheckedExpression `json:"checked"`
		Interface []algebra.PortBinding     `json:"interface"`
	}{Document: document, Checked: result.Checked, Interface: interfacePorts})
}

// SemanticInstantiationResult reports the exact selected record and the
// bounded result. A checked expression or verified goal is not a world fact.
type SemanticInstantiationResult struct {
	Status           status.OutcomeKind                `json:"status"`
	PackageID        string                            `json:"package_id"`
	PackageVersion   string                            `json:"package_version"`
	Language         string                            `json:"language"`
	SenseID          string                            `json:"sense_id"`
	FrameID          string                            `json:"frame_id"`
	BindingIndex     int                               `json:"binding_index"`
	BindingKey       string                            `json:"binding_key"`
	TargetKind       lexicon.SemanticTargetKind        `json:"target_kind"`
	TargetAuthorized bool                              `json:"target_authorized"`
	ContextID        string                            `json:"context_id"`
	Scope            string                            `json:"scope"`
	Symbol           *SemanticSymbolInstantiation      `json:"symbol,omitempty"`
	Template         *synthesis.GoalSpec               `json:"template,omitempty"`
	Goal             *synthesis.GoalSpec               `json:"goal,omitempty"`
	Construction     *synthesis.Result                 `json:"construction,omitempty"`
	Certificate      *synthesis.SynthesisCertificate   `json:"certificate,omitempty"`
	Verification     *synthesis.GoalVerificationResult `json:"verification,omitempty"`
}

// InstantiateSemanticBinding selects one exact target and applies only
// caller-provided typed reference substitutions. Metadata and package must
// already be selected and validated by the caller.
func InstantiateSemanticBinding(pack *SemanticLanguagePackage, selected algebra.Document, rewrite algebra.RewriteEnvironment, request SemanticBindingRequest) (SemanticInstantiationResult, error) {
	result := SemanticInstantiationResult{Status: status.OutcomeInvalid, Scope: "typed semantic target instantiation; no world-fact or language-quality claim"}
	if pack == nil || pack.lexicon == nil || pack.valency == nil {
		return result, invalid("MISSING_SEMANTIC_PACKAGE", "package", "a built semantic language package is required")
	}
	if err := validateSelectedPrimitiveContext(selected); err != nil {
		return result, err
	}
	if err := validateRewriteSelection(selected, rewrite); err != nil {
		return result, err
	}
	if !validSemanticPackageText(request.ContextID, algebra.MaxContextIDBytes) {
		return result, invalid("INVALID_CONTEXT_ID", "context_id", "request context ID must be explicit, bounded, and non-empty")
	}
	if request.RequestContext != nil && request.RequestContext.ID != request.ContextID {
		return result, invalid("REQUEST_CONTEXT_ID_MISMATCH", "request_context.id", "request-context ID must match context_id")
	}
	if len(request.RoleReferences) > algebra.DefaultMaxPorts || len(request.PortReferences) > algebra.DefaultMaxPorts || len(request.References) > algebra.DefaultMaxSymbols {
		return result, resource("binding_request_items", maxIntLanguage(len(request.RoleReferences), len(request.PortReferences), len(request.References)), algebra.DefaultMaxSymbols)
	}
	_, ok := pack.lexicon.BySenseID(request.SenseID)
	if !ok {
		return result, unknown("UNKNOWN_SENSE", "sense_id", "sense ID is not present in the selected language package")
	}
	frame, ok := pack.frameByIDForSense(request.SenseID, request.FrameID)
	if !ok {
		return result, unknown("UNKNOWN_VALENCY_FRAME", "frame_id", "frame ID is not linked to the exact selected sense")
	}
	bindings := pack.lexicon.TargetBindingsForSense(request.SenseID)
	if len(bindings) == 0 {
		return result, unsupported("LEGACY_OR_UNMAPPED_BINDING", "sense_id", "selected sense has no explicit semantic target; legacy CORE bindings are not used as fallback")
	}
	selectedIndex := 0
	if request.BindingIndex != nil {
		selectedIndex = *request.BindingIndex
	} else if len(bindings) != 1 {
		return result, languageError(status.OutcomeAmbiguous, "AMBIGUOUS_SEMANTIC_TARGET", "binding_index", "multiple semantic targets exist; select one explicitly")
	}
	if selectedIndex < 0 || selectedIndex >= len(bindings) {
		return result, invalid("INVALID_BINDING_INDEX", "binding_index", "binding index is outside the exact sense target list")
	}
	binding := bindings[selectedIndex]
	if binding.Target == nil {
		return result, invalid("MISSING_SEMANTIC_TARGET", "binding.target", "selected binding has no tagged target")
	}
	if err := validateTargetForFrame(binding.Target, frame); err != nil {
		return result, err
	}
	if _, err := validateSemanticTargetEnvironment(selected, binding.Target, frame); err != nil {
		return result, err
	}
	authorized := binding.TargetExecutionAuthorized()
	if !authorized {
		return result, unsupported("SEMANTIC_TARGET_NOT_AUTHORIZED", "binding", "only approved equivalent targets with valid provenance may be instantiated")
	}
	result = SemanticInstantiationResult{
		Status:    status.OutcomeInvalid,
		PackageID: pack.id, PackageVersion: pack.version, Language: string(pack.language),
		SenseID: request.SenseID.String(), FrameID: string(frame.ID), BindingIndex: selectedIndex,
		BindingKey: semanticBindingKey(binding), TargetKind: binding.Target.Kind, TargetAuthorized: true,
		ContextID: request.ContextID, Scope: "typed semantic target instantiation; no world-fact or language-quality claim",
	}
	switch binding.Target.Kind {
	case lexicon.SemanticTargetSymbol:
		refs, err := mergeReferences(nil, request.References)
		if err != nil {
			return result, err
		}
		contextDoc, err := buildInstantiationContext(selected, refs, request.RequestContext)
		if err != nil {
			return result, err
		}
		roleAssignments, portAssignments, err := resolveSemanticAssignments(binding.Target, frame, request, refs, selected)
		if err != nil {
			return result, err
		}
		function, err := algebra.Relation(selected, string(binding.Target.SymbolRef))
		if err != nil {
			return result, err
		}
		for _, item := range roleAssignments {
			arg := referenceTerm(refs, item.referenceID)
			function, err = algebra.Bind(contextDoc, function, item.portID, arg)
			if err != nil {
				return result, err
			}
		}
		for _, item := range portAssignments {
			arg := referenceTerm(refs, item.referenceID)
			function, err = algebra.Bind(contextDoc, function, item.portID, arg)
			if err != nil {
				return result, err
			}
		}
		checkedDoc := contextDoc
		checkedDoc.Expression = function.Expression
		checked, err := algebra.Check(checkedDoc)
		if err != nil {
			return result, err
		}
		result.Symbol = &SemanticSymbolInstantiation{Document: checkedDoc, Checked: checked, Interface: append([]algebra.PortBinding{}, function.Interface...)}
		result.Status = checked.Status
		return result, nil
	case lexicon.SemanticTargetGoal:
		refs, err := mergeReferences(binding.Target.GoalSpec.References, request.References)
		if err != nil {
			return result, err
		}
		_, err = buildInstantiationContext(selected, refs, request.RequestContext)
		if err != nil {
			return result, err
		}
		roleAssignments, portAssignments, err := resolveSemanticAssignments(binding.Target, frame, request, refs, selected)
		if err != nil {
			return result, err
		}
		return instantiateGoalTarget(result, selected, rewrite, binding.Target, frame, request, refs, roleAssignments, portAssignments)
	default:
		return result, unsupported("UNSUPPORTED_TARGET_KIND", "binding.target.kind", "semantic target kind is not supported")
	}
}

func buildInstantiationContext(selected algebra.Document, refs []algebra.Declaration, supplied *algebra.RequestContext) (algebra.Document, error) {
	doc := selected
	doc.References = cloneLanguageDeclarations(refs)
	doc.RequestContext = nil
	if supplied != nil {
		contextCopy, err := algebra.CloneRequestContext(*supplied)
		if err != nil {
			return algebra.Document{}, err
		}
		doc.RequestContext = &contextCopy
		if _, err := algebra.CheckRequestContext(contextCopy, doc); err != nil {
			return algebra.Document{}, err
		}
	}
	if err := algebra.CheckDeclarations(doc); err != nil {
		return algebra.Document{}, err
	}
	return doc, nil
}

type resolvedSemanticAssignment struct{ portID, referenceID string }

func (pack *SemanticLanguagePackage) frameByIDForSense(senseID ids.SenseID, frameID valency.FrameID) (valency.ValencyFrame, bool) {
	frame, ok := pack.frameByID[string(frameID)]
	return frame, ok && frame.SenseID == senseID
}

func validateRewriteSelection(selected algebra.Document, rewrite algebra.RewriteEnvironment) error {
	if err := algebra.ValidateRewriteEnvironment(rewrite); err != nil {
		return err
	}
	if rewrite.SnapshotID == "" {
		return invalid("MISSING_SELECTED_SNAPSHOT", "rewrite.snapshot_id", "caller must supply an explicitly selected rewrite snapshot")
	}
	bySort := make(map[string]bool, len(selected.Sorts))
	for _, row := range selected.Sorts {
		bySort[row.ID] = true
	}
	for _, row := range rewrite.Document.Sorts {
		if !bySort[row.ID] {
			return invalid("REWRITE_CONTEXT_MISMATCH", "rewrite.document.sorts", "rewrite context contains a sort absent from the selected primitive context")
		}
	}
	for _, row := range selected.Sorts {
		found := false
		for _, candidate := range rewrite.Document.Sorts {
			if candidate.ID == row.ID {
				found = true
				break
			}
		}
		if !found {
			return invalid("REWRITE_CONTEXT_MISMATCH", "rewrite.document.sorts", "rewrite context omits a selected primitive sort")
		}
	}
	selectedSigs := make(map[string]algebra.Signature, len(selected.Signatures))
	for _, sig := range selected.Signatures {
		selectedSigs[sig.ID] = sig
	}
	aliasIDs := make(map[string]bool, len(rewrite.Aliases))
	for _, alias := range rewrite.Aliases {
		aliasIDs[alias.ID] = true
	}
	for _, sig := range rewrite.Document.Signatures {
		base, ok := selectedSigs[sig.ID]
		if !ok {
			if !aliasIDs[sig.ID] {
				return invalid("REWRITE_CONTEXT_MISMATCH", "rewrite.document.signatures", "rewrite context contains a non-selected non-alias signature")
			}
			continue
		}
		if !sameAlgebraPorts(base.Ports, sig.Ports) {
			return invalid("REWRITE_CONTEXT_MISMATCH", "rewrite.document.signatures", "rewrite context changed a selected primitive signature")
		}
	}
	for id := range selectedSigs {
		found := false
		for _, sig := range rewrite.Document.Signatures {
			if sig.ID == id && sameAlgebraPorts(selectedSigs[id].Ports, sig.Ports) {
				found = true
				break
			}
		}
		if !found {
			return invalid("REWRITE_CONTEXT_MISMATCH", "rewrite.document.signatures", "rewrite context omits a selected primitive signature")
		}
	}
	return nil
}

func sameAlgebraPorts(left, right []algebra.Port) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].Role != right[i].Role || !algebra.TypeEqual(left[i].Type, right[i].Type) {
			return false
		}
	}
	return true
}

func resolveSemanticAssignments(target *lexicon.SemanticTarget, frame valency.ValencyFrame, request SemanticBindingRequest, refs []algebra.Declaration, selected algebra.Document) ([]resolvedSemanticAssignment, []resolvedSemanticAssignment, error) {
	portTypes := make(map[string]algebra.Type)
	switch target.Kind {
	case lexicon.SemanticTargetSymbol:
		for _, signature := range selected.Signatures {
			if signature.ID == string(target.SymbolRef) {
				for _, port := range signature.Ports {
					portTypes[port.Role] = port.Type
				}
				break
			}
		}
	case lexicon.SemanticTargetGoal:
		for _, port := range target.GoalSpec.Ports {
			portTypes[port.ID] = port.Type
		}
	}
	mapped := make(map[string]lexicon.SemanticRolePort)
	for _, mapping := range target.RolePorts {
		if mapping.FrameID == string(frame.ID) {
			mapped[mapping.Role] = mapping
			portTypes[mapping.PortID] = mapping.Type
		}
	}
	refMap := make(map[string]algebra.Declaration, len(refs))
	for _, ref := range refs {
		refMap[ref.ID] = ref
	}
	roleSeen, portSeen := map[string]bool{}, map[string]bool{}
	roleOut := make([]resolvedSemanticAssignment, 0, len(request.RoleReferences))
	for i, item := range request.RoleReferences {
		if !validSemanticPackageText(item.Role, 256) || !validSemanticPackageText(item.ReferenceID, 256) {
			return nil, nil, invalid("INVALID_ROLE_REFERENCE", fmt.Sprintf("role_references[%d]", i), "role and reference IDs must be explicit bounded strings")
		}
		if roleSeen[item.Role] {
			return nil, nil, invalid("DUPLICATE_ROLE_REFERENCE", fmt.Sprintf("role_references[%d].role", i), "a frame role may be bound only once")
		}
		roleSeen[item.Role] = true
		mapping, ok := mapped[item.Role]
		if !ok {
			return nil, nil, unknown("UNMAPPED_FRAME_ROLE", fmt.Sprintf("role_references[%d].role", i), "role is not explicitly mapped to a target port for this frame")
		}
		ref, ok := refMap[item.ReferenceID]
		if !ok {
			return nil, nil, unknown("UNDECLARED_REFERENCE", fmt.Sprintf("role_references[%d].reference_id", i), "reference must be declared by this request or goal template")
		}
		if !algebra.TypeEqual(ref.Type, mapping.Type) {
			return nil, nil, invalid("ROLE_REFERENCE_TYPE_MISMATCH", fmt.Sprintf("role_references[%d].reference_id", i), "reference type differs from the explicitly mapped role port")
		}
		if portSeen[mapping.PortID] {
			return nil, nil, invalid("DUPLICATE_TARGET_PORT_BINDING", fmt.Sprintf("role_references[%d]", i), "a target port may be bound only once")
		}
		portSeen[mapping.PortID] = true
		roleOut = append(roleOut, resolvedSemanticAssignment{mapping.PortID, item.ReferenceID})
	}
	portOut := make([]resolvedSemanticAssignment, 0, len(request.PortReferences))
	for i, item := range request.PortReferences {
		if !validSemanticPackageText(item.PortID, 256) || !validSemanticPackageText(item.ReferenceID, 256) {
			return nil, nil, invalid("INVALID_PORT_REFERENCE", fmt.Sprintf("port_references[%d]", i), "port and reference IDs must be explicit bounded strings")
		}
		if portSeen[item.PortID] {
			return nil, nil, invalid("DUPLICATE_TARGET_PORT_BINDING", fmt.Sprintf("port_references[%d].port_id", i), "a target port may be bound only once")
		}
		if _, isMapped := portMappingForFrame(target, string(frame.ID), item.PortID); isMapped {
			return nil, nil, invalid("ROLE_MAPPED_PORT_REQUIRES_ROLE", fmt.Sprintf("port_references[%d].port_id", i), "a role-mapped port must be bound through its explicit frame role")
		}
		typ, ok := portTypes[item.PortID]
		if !ok {
			return nil, nil, unknown("UNKNOWN_TARGET_PORT", fmt.Sprintf("port_references[%d].port_id", i), "port is not declared by the selected target")
		}
		ref, ok := refMap[item.ReferenceID]
		if !ok {
			return nil, nil, unknown("UNDECLARED_REFERENCE", fmt.Sprintf("port_references[%d].reference_id", i), "reference must be declared by this request or goal template")
		}
		if !algebra.TypeEqual(ref.Type, typ) {
			return nil, nil, invalid("PORT_REFERENCE_TYPE_MISMATCH", fmt.Sprintf("port_references[%d].reference_id", i), "reference type differs from target port")
		}
		portSeen[item.PortID] = true
		portOut = append(portOut, resolvedSemanticAssignment{item.PortID, item.ReferenceID})
	}
	sort.Slice(roleOut, func(i, j int) bool { return roleOut[i].portID < roleOut[j].portID })
	sort.Slice(portOut, func(i, j int) bool { return portOut[i].portID < portOut[j].portID })
	return roleOut, portOut, nil
}

func portMappingForFrame(target *lexicon.SemanticTarget, frameID, portID string) (lexicon.SemanticRolePort, bool) {
	for _, mapping := range target.RolePorts {
		if mapping.FrameID == frameID && mapping.PortID == portID {
			return mapping, true
		}
	}
	return lexicon.SemanticRolePort{}, false
}

func referenceTerm(refs []algebra.Declaration, id string) *algebra.Term {
	for _, ref := range refs {
		if ref.ID == id {
			typ := cloneLanguageType(ref.Type)
			return &algebra.Term{Kind: algebra.TermRef, ID: id, Type: &typ}
		}
	}
	return nil
}

func instantiateGoalTarget(result SemanticInstantiationResult, selected algebra.Document, rewrite algebra.RewriteEnvironment, target *lexicon.SemanticTarget, frame valency.ValencyFrame, request SemanticBindingRequest, requestRefs []algebra.Declaration, roleAssignments, portAssignments []resolvedSemanticAssignment) (SemanticInstantiationResult, error) {
	template, err := cloneLanguageGoal(*target.GoalSpec)
	if err != nil {
		return result, err
	}
	result.Template = &template
	contextFree := len(template.References) == 0 && template.RequestContext == nil
	instantiationTemplate, err := cloneLanguageGoal(template)
	if err != nil {
		return result, err
	}
	if contextFree {
		instantiationTemplate.ContextID = request.ContextID
	} else if template.ContextID != request.ContextID {
		return result, invalid("GOAL_CONTEXT_ID_MISMATCH", "goal.context_id", "request context ID must exactly match the target goal context ID")
	}
	refs, err := mergeReferences(instantiationTemplate.References, requestRefs)
	if err != nil {
		return result, err
	}
	original, err := cloneLanguageGoal(instantiationTemplate)
	if err != nil {
		return result, err
	}
	original.References = cloneLanguageDeclarations(refs)
	if request.RequestContext != nil {
		provided, err := algebra.CloneRequestContext(*request.RequestContext)
		if err != nil {
			return result, err
		}
		if original.RequestContext != nil {
			a := selected
			a.References = cloneLanguageDeclarations(refs)
			a.RequestContext = original.RequestContext
			b := selected
			b.References = cloneLanguageDeclarations(refs)
			b.RequestContext = &provided
			left, e := algebra.RequestContextFingerprint(*a.RequestContext, a)
			if e != nil {
				return result, e
			}
			right, e := algebra.RequestContextFingerprint(*b.RequestContext, b)
			if e != nil {
				return result, e
			}
			if left != right {
				return result, invalid("REQUEST_CONTEXT_CONFLICT", "request_context", "caller context conflicts with the target template context")
			}
		} else {
			original.RequestContext = &provided
		}
	}
	if original.RequestContext != nil && original.RequestContext.ID != request.ContextID {
		return result, invalid("REQUEST_CONTEXT_ID_MISMATCH", "goal.request_context.id", "goal request-context ID must match the instantiation context_id")
	}
	concrete, err := cloneLanguageGoal(original)
	if err != nil {
		return result, err
	}
	concrete.References = cloneLanguageDeclarations(refs)
	concrete.RequestContext, err = cloneLanguageRequestContext(original.RequestContext)
	if err != nil {
		return result, err
	}
	for _, assignment := range append(append([]resolvedSemanticAssignment{}, roleAssignments...), portAssignments...) {
		port := findGoalPort(&concrete, assignment.portID)
		if port == nil {
			return result, unknown("UNKNOWN_GOAL_PORT", "goal.ports", "assignment port is absent from goal")
		}
		ref := findLanguageDeclaration(refs, assignment.referenceID)
		if ref == nil {
			return result, unknown("UNDECLARED_REFERENCE", "references", "assignment reference is absent from goal declarations")
		}
		if port.Binding != nil {
			if port.Binding.Kind == algebra.TermRef && port.Binding.ID == ref.ID && port.Binding.Type != nil && algebra.TypeEqual(*port.Binding.Type, ref.Type) {
				continue
			}
			return result, invalid("GOAL_PORT_ALREADY_BOUND", "goal.ports", "request cannot replace an authored goal-port binding")
		}
		typ := cloneLanguageType(ref.Type)
		port.Binding = &synthesis.GoalValue{Kind: algebra.TermRef, ID: ref.ID, Type: &typ}
	}
	concrete.ExpectedType = expectedGoalType(concrete.Ports, concrete.ProjectPorts)
	// For the exact instantiated request the concrete Goal is the certificate's
	// original authority; the separately returned Template preserves authorship.
	primitiveContext := primitiveContextFromRewrite(rewrite)
	primitiveContext.Limits = selected.Limits
	primitiveContext.References = []algebra.Declaration{}
	primitiveContext.RequestContext = nil
	if err := validateSelectedPrimitiveContext(primitiveContext); err != nil {
		return result, err
	}
	if err := algebra.CheckDeclarations(algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: primitiveContext.Sorts, Parameters: []algebra.Declaration{{ID: "goal_expected_type", Type: concrete.ExpectedType}}, Limits: concrete.Limits.TermLimits}); err != nil {
		return result, err
	}
	if concrete.RequestContext != nil {
		contextDoc := primitiveContext
		contextDoc.References = cloneLanguageDeclarations(concrete.References)
		contextDoc.RequestContext = concrete.RequestContext
		if _, err := algebra.CheckRequestContext(*concrete.RequestContext, contextDoc); err != nil {
			return result, err
		}
	}
	construction, err := synthesis.Construct(primitiveContext, concrete)
	if err != nil {
		return result, err
	}
	result.Goal = &concrete
	result.Construction = &construction
	if construction.Status != status.OutcomeOK || construction.CanonicalDocument == nil {
		result.Status = construction.Status
		return result, nil
	}
	fullRewrite := rewrite
	fullRewrite.Document.References = cloneLanguageDeclarations(concrete.References)
	fullRewrite.Document.RequestContext, err = cloneLanguageRequestContext(concrete.RequestContext)
	if err != nil {
		return result, err
	}
	certificate, _, err := synthesis.CreateSynthesisCertificate(primitiveContext, fullRewrite, concrete, concrete, nil, synthesis.CertificateOriginConstruct, construction, request.Verification)
	if err != nil {
		return result, err
	}
	verified, err := synthesis.VerifyGoalCertificate(primitiveContext, fullRewrite, concrete, *construction.CanonicalDocument, certificate, request.Verification)
	if err != nil {
		return result, err
	}
	result.Certificate = &certificate
	result.Verification = &verified
	construction.Certificate = &certificate
	construction.Verification = &verified
	construction.FullyValidated = verified.CertificateValid && verified.GoalSatisfied && verified.Complete
	result.Construction = &construction
	result.Status = verified.Status
	if !verified.CertificateValid || !verified.GoalSatisfied || !verified.Complete {
		result.Scope = "goal-target construction checked by independent structural certificate replay; no world-fact or language-quality claim"
	}
	return result, nil
}

func primitiveContextFromRewrite(rewrite algebra.RewriteEnvironment) algebra.Document {
	doc := rewrite.Document
	alias := map[string]bool{}
	for _, definition := range rewrite.Aliases {
		alias[definition.ID] = true
	}
	sigs := make([]algebra.Signature, 0, len(doc.Signatures))
	for _, sig := range doc.Signatures {
		if !alias[sig.ID] {
			sig.Ports = append([]algebra.Port{}, sig.Ports...)
			sigs = append(sigs, sig)
		}
	}
	doc.Signatures = sigs
	doc.Expression = nil
	doc.Parameters = []algebra.Declaration{}
	doc.References = []algebra.Declaration{}
	doc.RequestContext = nil
	return doc
}

func expectedGoalType(ports []synthesis.GoalPort, projected []string) algebra.Type {
	closed := map[string]bool{}
	for _, id := range projected {
		closed[id] = true
	}
	typ := algebra.Type{Kind: algebra.TypeProp}
	for i := len(ports) - 1; i >= 0; i-- {
		port := ports[i]
		if port.Binding != nil || closed[port.ID] {
			continue
		}
		from := cloneLanguageType(port.Type)
		to := typ
		typ = algebra.Type{Kind: algebra.TypeArrow, From: &from, To: &to}
	}
	return typ
}

func findGoalPort(goal *synthesis.GoalSpec, id string) *synthesis.GoalPort {
	for i := range goal.Ports {
		if goal.Ports[i].ID == id {
			return &goal.Ports[i]
		}
	}
	return nil
}
func findLanguageDeclaration(refs []algebra.Declaration, id string) *algebra.Declaration {
	for i := range refs {
		if refs[i].ID == id {
			return &refs[i]
		}
	}
	return nil
}

func mergeReferences(left, right []algebra.Declaration) ([]algebra.Declaration, error) {
	if len(left)+len(right) > algebra.DefaultMaxSymbols {
		return nil, resource("semantic_references", len(left)+len(right), algebra.DefaultMaxSymbols)
	}
	byID := make(map[string]algebra.Declaration, len(left)+len(right))
	for groupIndex, group := range [][]algebra.Declaration{left, right} {
		incoming := make(map[string]bool, len(group))
		for _, ref := range group {
			if !validSemanticPackageText(ref.ID, 256) {
				return nil, invalid("INVALID_REFERENCE_ID", "references", "reference ID must be explicit and bounded")
			}
			if groupIndex == 1 && incoming[ref.ID] {
				return nil, invalid("DUPLICATE_REFERENCE_ID", "references", "request reference declarations cannot repeat an ID")
			}
			incoming[ref.ID] = true
			if prior, ok := byID[ref.ID]; ok {
				if !sameLanguageDeclaration(prior, ref) {
					return nil, invalid("REFERENCE_DECLARATION_CONFLICT", "references", "same reference ID has conflicting type or attributes")
				}
				continue
			}
			copyRef := ref
			copyRef.Type = cloneLanguageType(ref.Type)
			byID[ref.ID] = copyRef
		}
	}
	result := make([]algebra.Declaration, 0, len(byID))
	for _, ref := range byID {
		result = append(result, ref)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
func sameLanguageDeclaration(a, b algebra.Declaration) bool {
	return a.ID == b.ID && algebra.TypeEqual(a.Type, b.Type) && a.TemporalFrame == b.TemporalFrame && a.ReferenceTimeID == b.ReferenceTimeID && a.Multiplicity == b.Multiplicity
}
func cloneLanguageDeclarations(src []algebra.Declaration) []algebra.Declaration {
	out := make([]algebra.Declaration, len(src))
	for i, row := range src {
		out[i] = row
		out[i].Type = cloneLanguageType(row.Type)
	}
	return out
}
func cloneLanguageType(src algebra.Type) algebra.Type {
	out := src
	if src.From != nil {
		v := cloneLanguageType(*src.From)
		out.From = &v
	}
	if src.To != nil {
		v := cloneLanguageType(*src.To)
		out.To = &v
	}
	return out
}
func cloneLanguageRequestContext(src *algebra.RequestContext) (*algebra.RequestContext, error) {
	if src == nil {
		return nil, nil
	}
	out, err := algebra.CloneRequestContext(*src)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func cloneLanguageGoal(source synthesis.GoalSpec) (synthesis.GoalSpec, error) {
	result := source
	result.ExpectedType = cloneLanguageType(source.ExpectedType)
	result.Ports = make([]synthesis.GoalPort, len(source.Ports))
	for i, port := range source.Ports {
		result.Ports[i] = port
		result.Ports[i].Type = cloneLanguageType(port.Type)
		if port.Binding != nil {
			binding := *port.Binding
			if binding.Type != nil {
				typ := cloneLanguageType(*binding.Type)
				binding.Type = &typ
			}
			result.Ports[i].Binding = &binding
		}
	}
	result.References = cloneLanguageDeclarations(source.References)
	result.ProjectPorts = append([]string{}, source.ProjectPorts...)
	result.AllowedOperators = append([]string{}, source.AllowedOperators...)
	result.Constraints = make([]synthesis.Constraint, len(source.Constraints))
	for i, constraint := range source.Constraints {
		result.Constraints[i] = constraint
		result.Constraints[i].AllowedAtomCounts = append([]int{}, constraint.AllowedAtomCounts...)
	}
	result.Provenance = append([]provenance.SourceRef{}, source.Provenance...)
	contextCopy, err := cloneLanguageRequestContext(source.RequestContext)
	if err != nil {
		return synthesis.GoalSpec{}, err
	}
	result.RequestContext = contextCopy
	return result, nil
}

func semanticBindingKey(binding lexicon.SemanticBinding) string {
	if binding.Target == nil {
		return ""
	}
	b, _ := json.Marshal(binding.Target)
	return string(b)
}
func maxIntLanguage(values ...int) int {
	max := 0
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}
