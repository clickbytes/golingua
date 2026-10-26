package algebra

import (
	"encoding/json"
	"fmt"

	"github.com/clickbytes/golingua/internal/provenance"
)

// DecodeRequestContextJSON validates the strict wire shape and returns a
// detached context. Document-dependent sort and proposition checks are done
// by CheckRequestContext after declarations have been selected.
func DecodeRequestContextJSON(data []byte) (RequestContext, error) {
	if err := PreflightJSONLimit(data, MaxRequestContextBytes); err != nil {
		return RequestContext{}, err
	}
	context, err := decodeRequestContext(data, "context")
	if err != nil {
		return RequestContext{}, err
	}
	return CloneRequestContext(context)
}

func decodeRequestContext(data []byte, path string) (RequestContext, error) {
	fields, err := strictObject(data, path,
		[]string{"schema_version", "id", "referents", "speaker_referent_id", "addressee_referent_id", "mentions", "resolution_evidence", "conditions"},
		[]string{"schema_version", "id", "referents", "mentions", "resolution_evidence", "conditions"})
	if err != nil {
		return RequestContext{}, err
	}
	var out RequestContext
	if err := strictTypedUnmarshal(fields["schema_version"], &out.SchemaVersion, path+".schema_version"); err != nil {
		return out, err
	}
	if err := strictTypedUnmarshal(fields["id"], &out.ID, path+".id"); err != nil {
		return out, err
	}
	if raw, ok := fields["speaker_referent_id"]; ok {
		if err := strictTypedUnmarshal(raw, &out.SpeakerReferentID, path+".speaker_referent_id"); err != nil {
			return out, err
		}
		if !validText(out.SpeakerReferentID, MaxContextIDBytes) {
			return out, invalid("INVALID_CONTEXT_ANCHOR", path+".speaker_referent_id", "provided anchor must be a non-empty bounded ID")
		}
	}
	if raw, ok := fields["addressee_referent_id"]; ok {
		if err := strictTypedUnmarshal(raw, &out.AddresseeReferentID, path+".addressee_referent_id"); err != nil {
			return out, err
		}
		if !validText(out.AddresseeReferentID, MaxContextIDBytes) {
			return out, invalid("INVALID_CONTEXT_ANCHOR", path+".addressee_referent_id", "provided anchor must be a non-empty bounded ID")
		}
	}
	if out.Referents, err = decodeContextReferents(fields["referents"], path+".referents"); err != nil {
		return out, err
	}
	if out.Mentions, err = decodeContextMentions(fields["mentions"], path+".mentions"); err != nil {
		return out, err
	}
	if out.ResolutionEvidence, err = decodeContextEvidence(fields["resolution_evidence"], path+".resolution_evidence"); err != nil {
		return out, err
	}
	if out.Conditions, err = decodeContextConditions(fields["conditions"], path+".conditions"); err != nil {
		return out, err
	}
	if err := validateContextShape(out); err != nil {
		return RequestContext{}, err
	}
	return out, nil
}

func decodeContextReferents(raw []byte, path string) ([]ContextReferent, error) {
	items, err := strictArray(raw, path, MaxContextReferents)
	if err != nil {
		return nil, err
	}
	out := make([]ContextReferent, 0, len(items))
	for i, item := range items {
		p := fmt.Sprintf("%s[%d]", path, i)
		f, err := strictObject(item, p, []string{"id", "type", "concept_id"}, []string{"id", "type"})
		if err != nil {
			return nil, err
		}
		var row ContextReferent
		if err := strictTypedUnmarshal(f["id"], &row.ID, p+".id"); err != nil {
			return nil, err
		}
		row.Type, err = DecodeTypeJSON(f["type"], p+".type")
		if err != nil {
			return nil, err
		}
		if v, ok := f["concept_id"]; ok {
			if err := strictTypedUnmarshal(v, &row.ConceptID, p+".concept_id"); err != nil {
				return nil, err
			}
			if !validText(row.ConceptID, MaxContextIDBytes) {
				return nil, invalid("INVALID_CONCEPT_ID", p+".concept_id", "provided concept classification must be non-empty bounded text")
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func decodeContextMentions(raw []byte, path string) ([]ContextMention, error) {
	items, err := strictArray(raw, path, MaxContextMentions)
	if err != nil {
		return nil, err
	}
	out := make([]ContextMention, 0, len(items))
	for i, item := range items {
		p := fmt.Sprintf("%s[%d]", path, i)
		f, err := strictObject(item, p, []string{"id", "context_id", "referent_id"}, []string{"id"})
		if err != nil {
			return nil, err
		}
		var row ContextMention
		if err := strictTypedUnmarshal(f["id"], &row.ID, p+".id"); err != nil {
			return nil, err
		}
		if v, ok := f["context_id"]; ok {
			if err := strictTypedUnmarshal(v, &row.ContextID, p+".context_id"); err != nil {
				return nil, err
			}
			if !validText(row.ContextID, MaxContextIDBytes) {
				return nil, invalid("INVALID_MENTION_CONTEXT_ID", p+".context_id", "provided context ID must be non-empty bounded text")
			}
		}
		if v, ok := f["referent_id"]; ok {
			if err := strictTypedUnmarshal(v, &row.ReferentID, p+".referent_id"); err != nil {
				return nil, err
			}
			if !validText(row.ReferentID, MaxContextIDBytes) {
				return nil, invalid("INVALID_MENTION_REFERENT_ID", p+".referent_id", "provided referent ID must be non-empty bounded text")
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func decodeContextEvidence(raw []byte, path string) ([]ResolutionEvidence, error) {
	items, err := strictArray(raw, path, MaxContextEvidence)
	if err != nil {
		return nil, err
	}
	out := make([]ResolutionEvidence, 0, len(items))
	for i, item := range items {
		p := fmt.Sprintf("%s[%d]", path, i)
		f, err := strictObject(item, p, []string{"id", "mention_id", "referent_id", "source"}, []string{"id", "mention_id", "referent_id", "source"})
		if err != nil {
			return nil, err
		}
		var row ResolutionEvidence
		for _, field := range []struct {
			key  string
			dest *string
		}{{"id", &row.ID}, {"mention_id", &row.MentionID}, {"referent_id", &row.ReferentID}} {
			key, dest := field.key, field.dest
			if err := strictTypedUnmarshal(f[key], dest, p+"."+key); err != nil {
				return nil, err
			}
		}
		row.Source, err = decodeContextSource(f["source"], p+".source")
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, nil
}

func decodeContextSource(raw []byte, path string) (source provenance.SourceRef, err error) {
	f, err := strictObject(raw, path, []string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"}, []string{"id", "version", "license"})
	if err != nil {
		return source, err
	}
	if err := strictTypedUnmarshal(f["id"], &source.ID, path+".id"); err != nil {
		return source, err
	}
	if err := strictTypedUnmarshal(f["version"], &source.Version, path+".version"); err != nil {
		return source, err
	}
	if err := strictTypedUnmarshal(f["license"], &source.License, path+".license"); err != nil {
		return source, err
	}
	for _, field := range []struct {
		key  string
		dest *string
	}{{"path", &source.Path}, {"uri", &source.URI}, {"author", &source.Author}, {"checksum_sha256", &source.ChecksumSHA256}} {
		key, dest := field.key, field.dest
		if value, ok := f[key]; ok {
			if err := strictTypedUnmarshal(value, dest, path+"."+key); err != nil {
				return source, err
			}
		}
	}
	return source, nil
}

func decodeContextConditions(raw []byte, path string) ([]ContextCondition, error) {
	items, err := strictArray(raw, path, MaxContextConditions)
	if err != nil {
		return nil, err
	}
	out := make([]ContextCondition, 0, len(items))
	for i, item := range items {
		p := fmt.Sprintf("%s[%d]", path, i)
		f, err := strictObject(item, p, []string{"id", "kind", "code", "target_referent_ids", "expression", "detail"}, []string{"id", "kind", "code", "target_referent_ids"})
		if err != nil {
			return nil, err
		}
		var row ContextCondition
		if err := strictTypedUnmarshal(f["id"], &row.ID, p+".id"); err != nil {
			return nil, err
		}
		if err := strictTypedUnmarshal(f["kind"], &row.Kind, p+".kind"); err != nil {
			return nil, err
		}
		if err := strictTypedUnmarshal(f["code"], &row.Code, p+".code"); err != nil {
			return nil, err
		}
		if row.TargetReferentIDs, err = decodeContextStringArray(f["target_referent_ids"], p+".target_referent_ids", MaxContextReferents); err != nil {
			return nil, err
		}
		if v, ok := f["detail"]; ok {
			if err := strictTypedUnmarshal(v, &row.Detail, p+".detail"); err != nil {
				return nil, err
			}
		}
		if v, ok := f["expression"]; ok {
			row.Expression, err = DecodeTermJSON(v, p+".expression")
			if err != nil {
				return nil, err
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func decodeContextStringArray(raw []byte, path string, maximum int) ([]string, error) {
	items, err := strictArray(raw, path, maximum)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(items))
	for i, item := range items {
		var value string
		if err := strictTypedUnmarshal(item, &value, fmt.Sprintf("%s[%d]", path, i)); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

// The custom marshal boundary is intentionally scoped to RequestContext. It
// validates and preserves typed terms without changing global term encoding.
func (context RequestContext) MarshalJSON() ([]byte, error) {
	copyContext, err := CloneRequestContext(context)
	if err != nil {
		return nil, err
	}
	wire, err := requestContextWire(copyContext, false)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wire)
}

// UnmarshalJSON supports only the strict bounded context codec.
func (context *RequestContext) UnmarshalJSON(data []byte) error {
	if context == nil {
		return invalid("NULL_CONTEXT_TARGET", "context", "context target is nil")
	}
	decoded, err := DecodeRequestContextJSON(data)
	if err != nil {
		return err
	}
	*context = decoded
	return nil
}
