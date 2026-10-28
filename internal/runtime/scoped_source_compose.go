package runtime

import (
	"encoding/json"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
)

// scopedSourceComposition keeps the selected lexical evidence beside the
// generated expression. A successful algebra check is structural evidence,
// not a judgment of linguistic quality.
type scopedSourceComposition struct {
	Status     status.OutcomeKind
	Diagnostic status.Diagnostic
	Document   algebra.Document
	Evidence   scopedSourceSelection
}

// composeScopedSource builds one closed x/y/e interpretation without creating
// reference terms or discarding any selected declarations or request context.
func composeScopedSource(selection scopedSourceSelection, selected algebra.Document) scopedSourceComposition {
	out := scopedSourceComposition{Status: status.OutcomeInvalid, Evidence: cloneScopedSourceSelection(selection)}
	fail := func(kind status.OutcomeKind, code, path, message string) scopedSourceComposition {
		out.Status = kind
		out.Diagnostic = status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
		return out
	}
	if selection.Status != status.OutcomeOK {
		return fail(status.OutcomeUnsupported, "SOURCE_SELECTION_NOT_READY", "selection.status", "source metadata selection must be complete before composition")
	}
	if selected.Expression != nil {
		return fail(status.OutcomeUnsupported, "SELECTED_EXPRESSION_WOULD_BE_REPLACED", "selected.expression", "selected document already has an expression and cannot be copied without loss")
	}
	if selection.Nouns["x"].Binding.Target == nil || selection.Nouns["y"].Binding.Target == nil || selection.Event.Binding.Target == nil || selection.Event.SubjectBinder != "x" && selection.Event.SubjectBinder != "y" || selection.Event.ObjectBinder != "x" && selection.Event.ObjectBinder != "y" || selection.Event.SubjectBinder == selection.Event.ObjectBinder {
		return fail(status.OutcomeIncomplete, "SOURCE_BINDINGS_INCOMPLETE", "selection", "selected noun and event targets plus distinct subject/object binders are required")
	}
	// A JSON round trip provides a bounded, detached copy of every existing
	// declaration and all request-local context fields. No fields are elided.
	encoded, err := json.Marshal(selected)
	if err != nil || len(encoded) > algebra.MaxDocumentBytes {
		return fail(status.OutcomeResourceLimit, "SELECTED_CONTEXT_LIMIT", "selected", "selected context cannot be copied within the algebra document limit")
	}
	var document algebra.Document
	if err := json.Unmarshal(encoded, &document); err != nil {
		return fail(status.OutcomeInvalid, "SELECTED_CONTEXT_COPY_FAILED", "selected", "selected context could not be copied without loss")
	}
	if document.SchemaVersion != algebra.SchemaVersion {
		return fail(status.OutcomeInvalid, "INVALID_SELECTED_SCHEMA", "selected.schema_version", "selected context has an unsupported algebra schema")
	}
	xAtom, err := scopedReferenceAtom(document, selection.Nouns["x"], "x")
	if err != nil {
		return fail(status.OutcomeUnsupported, "REFERENCE_ATOM_UNAVAILABLE", "selection.nouns.x", err.Error())
	}
	yAtom, err := scopedReferenceAtom(document, selection.Nouns["y"], "y")
	if err != nil {
		return fail(status.OutcomeUnsupported, "REFERENCE_ATOM_UNAVAILABLE", "selection.nouns.y", err.Error())
	}
	eventAtom, err := scopedEventAtom(document, selection.Event)
	if err != nil {
		return fail(status.OutcomeUnsupported, "EVENT_ATOM_UNAVAILABLE", "selection.event", err.Error())
	}
	eventExists := &algebra.Term{Kind: algebra.TermExists, ID: "e", Type: &algebra.Type{Kind: algebra.TypeEvent}, Body: eventAtom}
	if temporal := scopedFiniteTemporalFrame(selection); temporal != "" {
		eventExists.TemporalFrame = temporal
	}
	body := &algebra.Term{Kind: algebra.TermAnd, Operands: []*algebra.Term{yAtom, eventExists}}
	body = &algebra.Term{Kind: algebra.TermExists, ID: "y", Type: &algebra.Type{Kind: algebra.TypeEntity}, Body: body}
	if selection.Variant.Negation == "inner" {
		body = &algebra.Term{Kind: algebra.TermNot, Body: body}
	} else if selection.Variant.Negation != "outer" {
		return fail(status.OutcomeInvalid, "INVALID_SCOPE", "selection.variant.negation", "scope must be explicitly outer or inner")
	}
	quantifier := &algebra.Term{Kind: algebra.TermQuantifier, ID: "x", Type: &algebra.Type{Kind: algebra.TypeEntity}, Quantifier: "all", Restrictor: xAtom, Body: body}
	if selection.Variant.Negation == "outer" {
		quantifier = &algebra.Term{Kind: algebra.TermNot, Body: quantifier}
	}
	document.Closed = true
	document.Expression = quantifier
	if _, err := algebra.Check(document); err != nil {
		return fail(status.OutcomeInvalid, "SCOPED_EXPRESSION_CHECK_FAILED", "expression", err.Error())
	}
	out.Status = status.OutcomeOK
	out.Document = document
	return out
}

func scopedReferenceAtom(document algebra.Document, target scopedNounTarget, binder string) (*algebra.Term, error) {
	if target.Symbol == "" || target.Binding.Target == nil || target.Binding.Target.ReferencePort == nil {
		return nil, errScopedComposeMetadata
	}
	ports := signaturePorts(document, target.Symbol)
	if len(ports) != 1 || ports[0].Role != target.Port.PortID || ports[0].Type.Kind != algebra.TypeEntity {
		return nil, errScopedComposeMetadata
	}
	return &algebra.Term{Kind: algebra.TermAtom, Symbol: string(target.Symbol), Args: []algebra.Argument{{Role: ports[0].Role, Value: &algebra.Term{Kind: algebra.TermVar, ID: binder}}}}, nil
}

func scopedEventAtom(document algebra.Document, event scopedEventTarget) (*algebra.Term, error) {
	if event.Symbol == "" || event.Head.Type.Kind != algebra.TypeEvent || event.Mapping.HeadPortID != event.Head.Role || event.SubjectRole == "" || event.ObjectRole == "" {
		return nil, errScopedComposeMetadata
	}
	ports := signaturePorts(document, event.Symbol)
	if len(ports) != 3 {
		return nil, errScopedComposeMetadata
	}
	args := make([]algebra.Argument, 0, len(ports))
	for _, port := range ports {
		var binder string
		switch {
		case port.Role == event.Head.Role && port.Type.Kind == algebra.TypeEvent:
			binder = "e"
		case port.Role == event.RolePorts[event.SubjectRole].Role && port.Type.Kind == algebra.TypeEntity:
			binder = event.SubjectBinder
		case port.Role == event.RolePorts[event.ObjectRole].Role && port.Type.Kind == algebra.TypeEntity:
			binder = event.ObjectBinder
		default:
			return nil, errScopedComposeMetadata
		}
		args = append(args, algebra.Argument{Role: port.Role, Value: &algebra.Term{Kind: algebra.TermVar, ID: binder}})
	}
	return &algebra.Term{Kind: algebra.TermAtom, Symbol: string(event.Symbol), Args: args}, nil
}

func scopedFiniteTemporalFrame(selection scopedSourceSelection) string {
	for _, binding := range selection.Bindings {
		if binding.Binder != "event" {
			continue
		}
		if binding.Reading.POS != "verb" {
			return ""
		}
		finite := false
		for _, value := range binding.Reading.Features.Values {
			if value.Feature == features.FeatureID("verb_form") && value.State == features.StateKnown && value.Value == features.FeatureValueID("finite") {
				finite = true
			}
		}
		if !finite {
			return ""
		}
		for _, value := range binding.Reading.Features.Values {
			if value.Feature == features.FeatureID("tense") && value.State == features.StateKnown && (value.Value == features.FeatureValueID("past") || value.Value == features.FeatureValueID("present")) {
				return string(value.Value)
			}
		}
	}
	return ""
}

func cloneScopedSourceSelection(source scopedSourceSelection) scopedSourceSelection {
	result := source
	result.Bindings = append([]ScopedSourceBinding(nil), source.Bindings...)
	result.Provenance = append([]provenance.SourceRef(nil), source.Provenance...)
	result.Nouns = make(map[string]scopedNounTarget, len(source.Nouns))
	for key, value := range source.Nouns {
		result.Nouns[key] = value
	}
	result.Event.RolePorts = make(map[string]algebra.Port, len(source.Event.RolePorts))
	for key, value := range source.Event.RolePorts {
		result.Event.RolePorts[key] = value
	}
	return result
}

var errScopedComposeMetadata = scopedComposeError("selected target metadata does not match the selected signature")

type scopedComposeError string

func (err scopedComposeError) Error() string { return string(err) }
