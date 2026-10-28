package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

type relationSourceComposition struct {
	Status         status.OutcomeKind
	Diagnostic     status.Diagnostic
	Document       algebra.Document
	Checked        algebra.CheckedExpression
	Encoded        []byte
	Evidence       relationSourceSelection
	TemporalFrames [2]string
	Work           int
	OutputBytes    int
}

// composeRelationSource constructs the checked anonymous two-hop relation
// function while retaining selected request context and lexical provenance.
// It does not create references, assert facts, or interpret truth.
func composeRelationSource(ctx context.Context, selection relationSourceSelection, options DSEOptions) relationSourceComposition {
	out := relationSourceComposition{Status: status.OutcomeInvalid, Work: selection.Work}
	fail := func(kind status.OutcomeKind, code, path, message string) relationSourceComposition {
		out.Status = kind
		out.Diagnostic = status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
		return out
	}
	if ctx == nil {
		return fail(status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "relation source composition requires a context")
	}
	budget, err := semantic.NormalizeDSEBudget(options.Budget)
	if err != nil {
		return fail(status.OutcomeInvalid, "INVALID_BUDGET", "budget", "relation source composition requires a valid finite DSE budget")
	}
	if selection.Status != status.OutcomeOK {
		return fail(status.OutcomeUnsupported, "RELATION_SOURCE_SELECTION_NOT_READY", "selection.status", "metadata selection must be complete before relation composition")
	}
	if options.Package == nil || options.Package.LexiconIndex() == nil || options.Plan == nil {
		return fail(status.OutcomeInvalid, "RELATION_SOURCE_INPUT_REQUIRED", "options", "selected semantic package and source plan are required")
	}
	if out.Work < 0 || out.Work > budget.MaxWork {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation source selection work exceeds the composition budget")
	}
	if !relationCompositionCharge(&out, budget.MaxWork, len(selection.Mapping.AllowedVariantIDs)+len(selection.Alternative.Parts)) {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation selection evidence scan exceeded the composition work limit")
	}
	evidenceBytes, evidenceErr := json.Marshal(selection)
	if evidenceErr != nil || len(evidenceBytes) > budget.MaxOutputBytes {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.output_bytes", "selected relation evidence exceeds its finite output bound")
	}
	out.OutputBytes = len(evidenceBytes)
	out.Evidence = cloneRelationSourceSelection(selection)
	if err := ctx.Err(); err != nil {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source composition was canceled")
	}
	if options.Selected.SchemaVersion != algebra.SchemaVersion || options.Selected.Expression != nil || len(options.Selected.References) != 0 || options.Selected.RequestContext != nil {
		return fail(status.OutcomeUnsupported, "SELECTED_CONTEXT_WOULD_BE_REPLACED", "selected", "selected metadata must be declaration-only; references and request context must be supplied through source reference options")
	}
	if !relationSourceSelectionConsistent(selection, options.Selected, options) {
		return fail(status.OutcomeInvalid, "RELATION_SOURCE_SELECTION_MISMATCH", "selection", "selected binding, frame, source variant, or ordered ports differ from the selected declarations")
	}
	if !relationCompositionCharge(&out, budget.MaxWork, len(selection.Signature.Ports)+5) {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation source composition work limit exceeded")
	}
	frames := [2]string{}
	firstReading, secondReading, readingsOK := relationSourceVerbEvidence(selection.Alternative)
	if !readingsOK || !sameRelationSourceReadingEvidence(firstReading, selection.VerbReadings[0]) || !sameRelationSourceReadingEvidence(secondReading, selection.VerbReadings[1]) {
		return fail(status.OutcomeInvalid, "RELATION_SOURCE_READING_EVIDENCE_MISMATCH", "selection.verb_readings", "preserved verb readings must exactly match the two source alternative occurrences")
	}
	for index, evidence := range selection.VerbReadings {
		if err := ctx.Err(); err != nil {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source composition was canceled")
		}
		if evidence.Reading != nil && !relationCompositionCharge(&out, budget.MaxWork, len(evidence.Reading.Features.Values)) {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "event morphology scan exceeded the composition work limit")
		}
		frame, kind, message := relationSourceReadingTemporalFrame(evidence, selection.Lexeme, selection.Sense)
		if kind != status.OutcomeOK {
			return fail(kind, "RELATION_EVENT_TIME_UNAVAILABLE", fmt.Sprintf("readings[%d].features", index), message)
		}
		frames[index] = frame
	}
	out.TemporalFrames = frames
	document, err := scopedSelectedWithReferences(options.Selected, options.References)
	if err != nil {
		return fail(status.OutcomeUnsupported, "REQUEST_CONTEXT_UNSUPPORTED", "references", err.Error())
	}
	if len(document.References) != len(options.References.References) {
		return fail(status.OutcomeInvalid, "REQUEST_DECLARATION_COPY_FAILED", "references", "reference declarations were not preserved exactly")
	}
	if !relationCompositionCharge(&out, budget.MaxWork, len(document.Signatures)+len(document.References)+len(document.Parameters)+len(document.Sorts)) {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "selected declaration scan exceeded the composition work limit")
	}
	signature, found := relationCompositionSignature(document, selection.Signature.ID)
	if !found || !sameRelationSourceSignature(signature, selection.Signature) {
		return fail(status.OutcomeInvalid, "RELATION_SIGNATURE_MISMATCH", "selected.signatures", "selected document must preserve the exact ordered relation signature")
	}
	entityPorts := make([]algebra.Port, 0, 2)
	var eventPort algebra.Port
	for _, port := range signature.Ports {
		if err := ctx.Err(); err != nil {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source composition was canceled")
		}
		if !relationCompositionCharge(&out, budget.MaxWork, 1) {
			return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.work", "relation signature scan exceeded the composition work limit")
		}
		switch port.Type.Kind {
		case algebra.TypeEvent:
			if eventPort.Role != "" || !bareRelationType(port.Type, algebra.TypeEvent) || port.Role != selection.Mapping.HeadPortID || port.Role != selection.HeadPort.Role {
				return fail(status.OutcomeInvalid, "RELATION_EVENT_PORT_MISMATCH", "selected.signatures", "exactly one Event port must match the authorized source head port")
			}
			eventPort = port
		case algebra.TypeEntity:
			if !bareRelationType(port.Type, algebra.TypeEntity) {
				return fail(status.OutcomeInvalid, "RELATION_ENTITY_PORT_MISMATCH", "selected.signatures", "relation Entity ports must be unmodified Entity types")
			}
			entityPorts = append(entityPorts, port)
		default:
			return fail(status.OutcomeUnsupported, "RELATION_SIGNATURE_UNSUPPORTED", "selected.signatures", "relation signature contains an unsupported port type")
		}
	}
	if eventPort.Role == "" || len(entityPorts) != 2 || len(signature.Ports) != 3 || eventPort.Role == entityPorts[0].Role || eventPort.Role == entityPorts[1].Role || entityPorts[0].Role == entityPorts[1].Role {
		return fail(status.OutcomeInvalid, "RELATION_SIGNATURE_SHAPE", "selected.signatures", "relation signature must contain one Event head and two ordered Entity ports")
	}
	if !relationSourceSelectionRolesConsistent(selection, entityPorts, eventPort) {
		return fail(status.OutcomeInvalid, "RELATION_ROLE_MAPPING_MISMATCH", "selection", "explicit role ports do not map the selected frame subject and object to the ordered Entity ports")
	}
	firstAtom := relationSourceAtom(selection, signature, entityPorts, eventPort, "e1", "x", "z")
	secondAtom := relationSourceAtom(selection, signature, entityPorts, eventPort, "e2", "z", "y")
	if firstAtom == nil || secondAtom == nil {
		return fail(status.OutcomeInvalid, "RELATION_ATOM_COMPOSITION_FAILED", "expression", "selected ordered signature cannot be composed with its authorized ports")
	}
	firstEvent := &algebra.Term{Kind: algebra.TermExists, ID: "e1", Type: &algebra.Type{Kind: algebra.TypeEvent}, TemporalFrame: frames[0], Body: firstAtom}
	secondEvent := &algebra.Term{Kind: algebra.TermExists, ID: "e2", Type: &algebra.Type{Kind: algebra.TypeEvent}, TemporalFrame: frames[1], Body: secondAtom}
	conjunction := &algebra.Term{Kind: algebra.TermAnd, Operands: []*algebra.Term{firstEvent, secondEvent}}
	witness := &algebra.Term{Kind: algebra.TermExists, ID: "z", Type: &algebra.Type{Kind: algebra.TypeEntity}, Body: conjunction}
	inner := &algebra.Term{Kind: algebra.TermLambda, ID: "y", Type: &algebra.Type{Kind: algebra.TypeEntity}, Body: witness}
	root := &algebra.Term{Kind: algebra.TermLambda, ID: "x", Type: &algebra.Type{Kind: algebra.TypeEntity}, Body: inner}
	document.Expression, document.Closed = root, true
	if err := ctx.Err(); err != nil {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "context", "relation source composition was canceled")
	}
	checked, err := algebra.Check(document)
	if err != nil {
		return fail(status.OutcomeInvalid, "RELATION_EXPRESSION_CHECK_FAILED", "expression", err.Error())
	}
	if !isRelationFunctionType(checked.Type) || len(checked.FreeParameters) != 0 || len(checked.Holes) != 0 || len(checked.Requirements) != 0 {
		return fail(status.OutcomeInvalid, "RELATION_EXPRESSION_TYPE_MISMATCH", "expression", "checked result must be a complete Entity-to-Entity-to-Prop function")
	}
	out.Checked = checked
	encoded, err := algebra.Encode(document)
	if err != nil {
		return fail(status.OutcomeInvalid, "RELATION_EXPRESSION_ENCODE_FAILED", "document", err.Error())
	}
	if len(encoded)+out.OutputBytes > budget.MaxOutputBytes {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.output_bytes", "encoded relation document exceeds the finite output limit")
	}
	evidenceJSON, err := json.Marshal(struct {
		Selection relationSourceSelection `json:"selection"`
		Temporal  [2]string               `json:"temporal_frames"`
	}{Selection: out.Evidence, Temporal: frames})
	if err != nil {
		return fail(status.OutcomeInvalid, "RELATION_EVIDENCE_ENCODING_FAILED", "evidence", err.Error())
	}
	if len(encoded)+len(evidenceJSON) > budget.MaxOutputBytes {
		return fail(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.output_bytes", "document and preserved source evidence exceed the finite output limit")
	}
	out.Document = document
	out.Encoded = append([]byte(nil), encoded...)
	out.OutputBytes = len(encoded) + len(evidenceJSON)
	out.Status = status.OutcomeOK
	return out
}

func relationSourceReadingTemporalFrame(evidence RelationSourcePartEvidence, selected lexicon.MorphLexeme, selectedSense lexicon.LexicalSense) (string, status.OutcomeKind, string) {
	if evidence.Reading == nil || evidence.Lexeme == nil || evidence.Sense == nil || evidence.Reading.POS != "verb" || evidence.Reading.LexemeID != selected.ID.String() || evidence.Lexeme.ID != selected.ID || evidence.Lexeme.Language != selected.Language || evidence.Lexeme.Lemma != selected.Lemma || evidence.Lexeme.POS != selected.POS || evidence.Sense.ID != selectedSense.ID || evidence.Sense.LexemeID != selectedSense.LexemeID {
		return "", status.OutcomeUnsupported, "event reading does not identify the selected verb lexeme"
	}
	if evidence.Reading.Features.Language == "" || string(evidence.Reading.Features.Language) != string(selected.Language) || evidence.Reading.Features.ProfileVersion == "" {
		return "", status.OutcomeUnsupported, "event morphology reading does not identify the selected language feature profile"
	}
	featuresByID := map[string]features.FeatureValue{}
	for _, value := range evidence.Reading.Features.Values {
		if value.Feature != "verb_form" && value.Feature != "tense" {
			continue
		}
		key := string(value.Feature)
		if _, duplicate := featuresByID[key]; duplicate {
			return "", status.OutcomeInvalid, "event reading has duplicate finite-form or tense evidence"
		}
		featuresByID[key] = value
	}
	form, hasForm := featuresByID["verb_form"]
	tense, hasTense := featuresByID["tense"]
	if !hasForm || !hasTense || form.State != features.StateKnown || form.Value != "finite" || form.Variable != nil || tense.State != features.StateKnown || tense.Variable != nil {
		return "", status.OutcomeIncomplete, "event reading lacks known finite-form and temporal-frame evidence"
	}
	if tense.Value != "present" && tense.Value != "past" {
		return "", status.OutcomeUnsupported, "event reading temporal frame is outside the supported present/past algebra values"
	}
	return string(tense.Value), status.OutcomeOK, ""
}

func relationSourceSelectionConsistent(selection relationSourceSelection, selected algebra.Document, options DSEOptions) bool {
	if selection.Binding.Target == nil || selection.Binding.Target.Kind != lexicon.SemanticTargetSymbol || selection.Binding.Mapping != lexicon.MappingEquivalent || selection.Binding.Status != lexicon.BindingApproved || !selection.Binding.TargetExecutionAuthorized() || selection.Binding.Target.SymbolRef == "" || string(selection.Binding.Target.SymbolRef) != selection.Signature.ID || selection.Mapping.SenseID != selection.Binding.SenseID || selection.Mapping.FrameID != selection.Frame.ID || selection.Mapping.HeadPortID == "" || selection.Mapping.HeadPortID != selection.HeadPort.Role || selection.Frame.SenseID != selection.Binding.SenseID || selection.Sense.ID != selection.Binding.SenseID || selection.Sense.LexemeID != selection.Lexeme.ID || selection.Binding.SenseID.String() == "" {
		return false
	}
	if selection.Alternative.VariantID != selection.Variant.ID || selection.Variant.ID == "" || selection.Mapping.AllowedVariantIDs == nil {
		return false
	}
	allowed := false
	for _, id := range selection.Mapping.AllowedVariantIDs {
		if string(id) == selection.Variant.ID {
			allowed = true
		}
	}
	if !allowed || options.Plan == nil || options.Plan.Document().Language != string(options.Package.Language()) || selection.Lexeme.Language != options.Package.Language() || selection.Lexeme.POS != "verb" {
		return false
	}
	selectedSense, senseOK := options.Package.LexiconIndex().BySenseID(selection.Sense.ID)
	selectedLexeme, lexemeOK := options.Package.LexiconIndex().ByLexemeID(selection.Lexeme.ID)
	if !senseOK || !lexemeOK || selectedSense.ID != selection.Sense.ID || selectedSense.LexemeID != selection.Sense.LexemeID || selectedLexeme.ID != selection.Lexeme.ID || selectedLexeme.Language != selection.Lexeme.Language || selectedLexeme.Lemma != selection.Lexeme.Lemma || selectedLexeme.POS != selection.Lexeme.POS {
		return false
	}
	if _, found := relationCompositionSignature(selected, selection.Signature.ID); !found {
		return false
	}
	return true
}

func relationSourceSelectionRolesConsistent(selection relationSourceSelection, entities []algebra.Port, event algebra.Port) bool {
	if len(entities) != 2 || len(selection.RolePorts) != 2 || len(selection.PortRoles) != 2 || selection.HeadPort.Role != event.Role || selection.SubjectRole == "" || selection.ObjectRole == "" || selection.SubjectRole == selection.ObjectRole || selection.RolePorts[selection.SubjectRole].Role == "" || selection.RolePorts[selection.ObjectRole].Role == "" {
		return false
	}
	if _, eventMapped := selection.PortRoles[event.Role]; eventMapped {
		return false
	}
	if selection.PortRoles[entities[0].Role] == "" || selection.PortRoles[entities[1].Role] == "" || selection.PortRoles[entities[0].Role] == selection.PortRoles[entities[1].Role] {
		return false
	}
	firstRolePort := selection.RolePorts[selection.PortRoles[entities[0].Role]]
	secondRolePort := selection.RolePorts[selection.PortRoles[entities[1].Role]]
	if firstRolePort.Role != entities[0].Role || secondRolePort.Role != entities[1].Role || !algebra.TypeEqual(firstRolePort.Type, entities[0].Type) || !algebra.TypeEqual(secondRolePort.Type, entities[1].Type) {
		return false
	}
	for _, evidence := range selection.Alternative.Parts {
		if evidence.Slot != "x" && evidence.Slot != "y" && evidence.Slot != "z" {
			continue
		}
		position := 0
		if evidence.PortPosition == "second" {
			position = 1
		} else if evidence.PortPosition != "first" {
			return false
		}
		if evidence.Port != evidence.PortPosition || selection.PortRoles[entities[position].Role] == "" {
			return false
		}
		role := selection.PortRoles[entities[position].Role]
		if evidence.GrammaticalFunction == "subject" && role != selection.SubjectRole || evidence.GrammaticalFunction == "object" && role != selection.ObjectRole {
			return false
		}
		if evidence.GrammaticalFunction != "subject" && evidence.GrammaticalFunction != "object" {
			return false
		}
	}
	frameRolePorts := 0
	for _, rolePort := range selection.Binding.Target.RolePorts {
		if rolePort.FrameID != string(selection.Frame.ID) {
			continue
		}
		frameRolePorts++
		selectedPort, mapped := selection.RolePorts[rolePort.Role]
		if rolePort.PortID == event.Role || rolePort.Type.Kind != algebra.TypeEntity || !mapped || selectedPort.Role != rolePort.PortID || !algebra.TypeEqual(selectedPort.Type, rolePort.Type) || selection.PortRoles[rolePort.PortID] != rolePort.Role {
			return false
		}
	}
	if frameRolePorts != 2 {
		return false
	}
	frameSubject, frameObject, frameOK := relationSourceFrameRoles(selection.Frame)
	if !frameOK || frameSubject != selection.SubjectRole || frameObject != selection.ObjectRole {
		return false
	}
	return true
}

func sameRelationSourceReadingEvidence(left, right RelationSourcePartEvidence) bool {
	return reflect.DeepEqual(left, right)
}

func relationSourceAtom(selection relationSourceSelection, signature algebra.Signature, entities []algebra.Port, event algebra.Port, eventBinder, firstEntityBinder, secondEntityBinder string) *algebra.Term {
	if len(entities) != 2 || eventBinder == "" || firstEntityBinder == "" || secondEntityBinder == "" {
		return nil
	}
	args := make([]algebra.Argument, 0, len(signature.Ports))
	for _, port := range signature.Ports {
		var binder string
		switch port.Role {
		case event.Role:
			if !algebra.TypeEqual(port.Type, event.Type) {
				return nil
			}
			binder = eventBinder
		case entities[0].Role:
			binder = firstEntityBinder
		case entities[1].Role:
			binder = secondEntityBinder
		default:
			return nil
		}
		args = append(args, algebra.Argument{Role: port.Role, Value: &algebra.Term{Kind: algebra.TermVar, ID: binder}})
	}
	return &algebra.Term{Kind: algebra.TermAtom, Symbol: string(selection.Binding.Target.SymbolRef), Args: args}
}

func relationCompositionSignature(document algebra.Document, id string) (algebra.Signature, bool) {
	var found algebra.Signature
	count := 0
	for _, signature := range document.Signatures {
		if signature.ID == id {
			found, count = signature, count+1
		}
	}
	return found, count == 1
}

func sameRelationSourceSignature(left, right algebra.Signature) bool {
	if left.ID != right.ID || len(left.Ports) != len(right.Ports) {
		return false
	}
	for index := range left.Ports {
		if left.Ports[index].Role != right.Ports[index].Role || !algebra.TypeEqual(left.Ports[index].Type, right.Ports[index].Type) {
			return false
		}
	}
	return true
}

func isRelationFunctionType(value algebra.Type) bool {
	if value.Kind != algebra.TypeArrow || value.From == nil || value.To == nil || !bareRelationType(*value.From, algebra.TypeEntity) {
		return false
	}
	inner := value.To
	return inner.Kind == algebra.TypeArrow && inner.From != nil && inner.To != nil && bareRelationType(*inner.From, algebra.TypeEntity) && bareRelationType(*inner.To, algebra.TypeProp)
}

func relationCompositionCharge(result *relationSourceComposition, maxWork, amount int) bool {
	if amount < 0 {
		return false
	}
	result.Work += amount
	return result.Work <= maxWork
}

func cloneRelationSourceSelection(selection relationSourceSelection) relationSourceSelection {
	copy := selection
	copy.Alternative = cloneRelationSourceAlternative(selection.Alternative)
	copy.Variant.Parts = append(selection.Variant.Parts[:0:0], selection.Variant.Parts...)
	copy.Mapping = cloneSourcePredicateMapping(selection.Mapping)
	copy.Binding = cloneRelationSourceBinding(selection.Binding)
	copy.Sense = cloneRelationSourceSense(selection.Sense)
	copy.Lexeme = cloneRelationSourceLexeme(selection.Lexeme)
	copy.Frame = cloneRelationSourceFrame(selection.Frame)
	copy.Validation.Unresolved = append(selection.Validation.Unresolved[:0:0], selection.Validation.Unresolved...)
	copy.Signature = cloneRelationSourceSignature(selection.Signature)
	copy.Signature.Ports = append(selection.Signature.Ports[:0:0], selection.Signature.Ports...)
	copy.RolePorts = make(map[string]algebra.Port, len(selection.RolePorts))
	for role, port := range selection.RolePorts {
		port.Type = cloneRelationSourceType(port.Type)
		copy.RolePorts[role] = port
	}
	copy.PortRoles = cloneRelationSourcePortRoles(selection.PortRoles)
	copy.VerbReadings = [2]RelationSourcePartEvidence{cloneRelationSourcePart(selection.VerbReadings[0]), cloneRelationSourcePart(selection.VerbReadings[1])}
	return copy
}

func cloneRelationSourcePortRoles(source map[string]string) map[string]string {
	copy := make(map[string]string, len(source))
	for port, role := range source {
		copy[port] = role
	}
	return copy
}
