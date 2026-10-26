package algebra

import "fmt"

// ReferenceAttributes are explicit, request-local attributes of a declared
// reference. Empty values mean unresolved; they are never filled by inference.
type ReferenceAttributes struct {
	TemporalFrame   string `json:"temporal_frame,omitempty"`
	ReferenceTimeID string `json:"reference_time_id,omitempty"`
	Multiplicity    string `json:"multiplicity,omitempty"`
}

func declarationAttributes(value Declaration) ReferenceAttributes {
	return ReferenceAttributes{TemporalFrame: value.TemporalFrame, ReferenceTimeID: value.ReferenceTimeID, Multiplicity: value.Multiplicity}
}

func contextReferentAttributes(value ContextReferent) ReferenceAttributes {
	return ReferenceAttributes{TemporalFrame: value.TemporalFrame, ReferenceTimeID: value.ReferenceTimeID, Multiplicity: value.Multiplicity}
}

func validateReferenceAttributes(id string, typ Type, attributes ReferenceAttributes, path string) error {
	if attributes.TemporalFrame != "" {
		if attributes.TemporalFrame != "past" && attributes.TemporalFrame != "present" {
			return invalid("INVALID_TEMPORAL_FRAME", path+".temporal_frame", "temporal frame must be past or present")
		}
		if typ.Kind != TypeEvent {
			return invalid("ATTRIBUTE_TYPE_MISMATCH", path+".temporal_frame", "temporal frame is available only on Event references")
		}
	}
	if attributes.ReferenceTimeID != "" {
		if !validText(attributes.ReferenceTimeID, MaxContextIDBytes) {
			return invalid("INVALID_REFERENCE_TIME_ID", path+".reference_time_id", "reference-time ID must be non-empty bounded text")
		}
		if typ.Kind != TypeEvent {
			return invalid("ATTRIBUTE_TYPE_MISMATCH", path+".reference_time_id", "reference time is available only on Event references")
		}
	}
	if attributes.Multiplicity != "" {
		if attributes.Multiplicity != "one" && attributes.Multiplicity != "many" {
			return invalid("INVALID_MULTIPLICITY", path+".multiplicity", "multiplicity must be one or many")
		}
		if typ.Kind != TypeEntity && typ.Kind != TypeNominal {
			return invalid("ATTRIBUTE_TYPE_MISMATCH", path+".multiplicity", "multiplicity is available only on Entity or Nominal references")
		}
	}
	_ = id
	return nil
}

// EffectiveReferenceAttributes merges the explicitly supplied document and
// context attributes for one reference. Conflicting known values are invalid;
// an omitted value remains unresolved.
func EffectiveReferenceAttributes(document Document, context *RequestContext, referenceID string) (ReferenceAttributes, error) {
	if !validText(referenceID, MaxContextIDBytes) {
		return ReferenceAttributes{}, invalid("INVALID_REFERENCE_ID", "reference_id", "reference ID must be non-empty bounded text")
	}
	var typ Type
	var result ReferenceAttributes
	found := false
	for _, declaration := range document.References {
		if declaration.ID == referenceID {
			typ, result, found = declaration.Type, declarationAttributes(declaration), true
			break
		}
	}
	if context != nil {
		for _, referent := range context.Referents {
			if referent.ID != referenceID {
				continue
			}
			if found && !TypeEqual(typ, referent.Type) {
				return ReferenceAttributes{}, invalid("REFERENCE_TYPE_CONFLICT", "request_context.referents", "the same reference ID has different types")
			}
			if !found {
				typ, found = referent.Type, true
			}
			other := contextReferentAttributes(referent)
			for _, field := range []struct {
				name  string
				left  *string
				right string
			}{
				{"temporal_frame", &result.TemporalFrame, other.TemporalFrame},
				{"reference_time_id", &result.ReferenceTimeID, other.ReferenceTimeID},
				{"multiplicity", &result.Multiplicity, other.Multiplicity},
			} {
				if *field.left != "" && field.right != "" && *field.left != field.right {
					return ReferenceAttributes{}, invalid("CONFLICTING_REFERENCE_ATTRIBUTE", "request_context.referents."+referenceID+"."+field.name, "document and context provide opposing known values")
				}
				if *field.left == "" {
					*field.left = field.right
				}
			}
			break
		}
	}
	if !found {
		return ReferenceAttributes{}, invalid("UNKNOWN_REFERENCE", "reference_id", "reference is not declared in the document or context")
	}
	if err := validateReferenceAttributes(referenceID, typ, result, "references."+referenceID); err != nil {
		return ReferenceAttributes{}, err
	}
	if result.ReferenceTimeID != "" {
		var target *Type
		for i := range document.References {
			if document.References[i].ID == result.ReferenceTimeID {
				target = &document.References[i].Type
				break
			}
		}
		if context != nil {
			for i := range context.Referents {
				if context.Referents[i].ID == result.ReferenceTimeID {
					if target != nil && !TypeEqual(*target, context.Referents[i].Type) {
						return ReferenceAttributes{}, invalid("REFERENCE_TYPE_CONFLICT", "request_context.referents", "reference-time ID has conflicting types")
					}
					target = &context.Referents[i].Type
					break
				}
			}
		}
		if target == nil || target.Kind != TypeInterval {
			return ReferenceAttributes{}, invalid("INVALID_REFERENCE_TIME_TARGET", "references."+referenceID+".reference_time_id", "reference time must name a declared Interval reference")
		}
	}
	return result, nil
}

func validateDocumentReferenceAttributes(document Document) error {
	refs := make(map[string]Declaration, len(document.References))
	for i, ref := range document.References {
		if err := validateReferenceAttributes(ref.ID, ref.Type, declarationAttributes(ref), fmt.Sprintf("document.references[%d]", i)); err != nil {
			return err
		}
		refs[ref.ID] = ref
	}
	if document.RequestContext == nil {
		for i, ref := range document.References {
			if ref.ReferenceTimeID == "" {
				continue
			}
			target, ok := refs[ref.ReferenceTimeID]
			if !ok || target.Type.Kind != TypeInterval {
				return invalid("INVALID_REFERENCE_TIME_TARGET", fmt.Sprintf("document.references[%d].reference_time_id", i), "reference time must name a declared Interval reference")
			}
		}
	}
	if document.RequestContext != nil {
		for i, ref := range document.RequestContext.Referents {
			if err := validateReferenceAttributes(ref.ID, ref.Type, contextReferentAttributes(ref), fmt.Sprintf("request_context.referents[%d]", i)); err != nil {
				return err
			}
		}
		for _, ref := range document.RequestContext.Referents {
			if ref.ReferenceTimeID != "" {
				if _, err := EffectiveReferenceAttributes(document, document.RequestContext, ref.ID); err != nil {
					return err
				}
			}
		}
		for _, ref := range document.References {
			if _, err := EffectiveReferenceAttributes(document, document.RequestContext, ref.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
