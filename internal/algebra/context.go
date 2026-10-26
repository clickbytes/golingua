package algebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

const (
	RequestContextSchemaVersion = 1
	MaxRequestContextBytes      = 1 << 20
	MaxContextIDBytes           = 256
	MaxContextReferents         = 512
	MaxContextMentions          = 512
	MaxContextEvidence          = 1024
	MaxContextConditions        = 128
	MaxContextConditionNodes    = 4096
	MaxContextWork              = 262144
	MaxContextHumanTextBytes    = 1024
)

// RequestContext is explicit, request-local identity and semantic-condition
// data. Its fields do not assert facts, resolve identities by themselves, or
// create conversational memory.
type RequestContext struct {
	SchemaVersion       int                  `json:"schema_version"`
	ID                  string               `json:"id"`
	Referents           []ContextReferent    `json:"referents"`
	SpeakerReferentID   string               `json:"speaker_referent_id,omitempty"`
	AddresseeReferentID string               `json:"addressee_referent_id,omitempty"`
	Mentions            []ContextMention     `json:"mentions"`
	ResolutionEvidence  []ResolutionEvidence `json:"resolution_evidence"`
	Conditions          []ContextCondition   `json:"conditions"`
}

// ContextReferent identifies one typed request-local referent. ConceptID is
// optional classification metadata; equal concepts never merge referents.
type ContextReferent struct {
	ID              string `json:"id"`
	Type            Type   `json:"type"`
	ConceptID       string `json:"concept_id,omitempty"`
	TemporalFrame   string `json:"temporal_frame,omitempty"`
	ReferenceTimeID string `json:"reference_time_id,omitempty"`
	Multiplicity    string `json:"multiplicity,omitempty"`
}

// ContextMention is a mention inside this context. ReferentID is a declared
// target claim that requires corroborating ResolutionEvidence.
type ContextMention struct {
	ID         string `json:"id"`
	ContextID  string `json:"context_id,omitempty"`
	ReferentID string `json:"referent_id,omitempty"`
}

// ResolutionEvidence is inert provenance for one explicit identity link.
// It is not a world-truth certificate or a fact-store record.
type ResolutionEvidence struct {
	ID         string               `json:"id"`
	MentionID  string               `json:"mention_id"`
	ReferentID string               `json:"referent_id"`
	Source     provenance.SourceRef `json:"source"`
}

type ContextConditionKind string

const (
	ContextAssumption     ContextConditionKind = "assumption"
	ContextPresupposition ContextConditionKind = "presupposition"
	ContextSemanticLoss   ContextConditionKind = "loss"
)

// ContextCondition records an explicit assumption, presupposition, or
// semantic loss. Code is an opaque label; the checker interprets no condition
// code or expression as an evaluator rule.
type ContextCondition struct {
	ID                string               `json:"id"`
	Kind              ContextConditionKind `json:"kind"`
	Code              string               `json:"code"`
	TargetReferentIDs []string             `json:"target_referent_ids"`
	Expression        *Term                `json:"expression,omitempty"`
	Detail            string               `json:"detail,omitempty"`
}

type MentionResolutionStatus string

const (
	MentionUnresolved MentionResolutionStatus = "unresolved"
	MentionResolved   MentionResolutionStatus = "resolved"
	MentionAmbiguous  MentionResolutionStatus = "ambiguous"
)

type MentionResolution struct {
	MentionID            string                  `json:"mention_id"`
	Status               MentionResolutionStatus `json:"status"`
	CandidateReferentIDs []string                `json:"candidate_referent_ids"`
	EvidenceIDs          []string                `json:"evidence_ids"`
}

// ContextAssessment separates derived reference status from the supplied
// context. It does not turn evidence or semantic conditions into facts.
type ContextAssessment struct {
	ContextID           string              `json:"context_id"`
	Fingerprint         string              `json:"fingerprint"`
	SpeakerReferentID   string              `json:"speaker_referent_id,omitempty"`
	AddresseeReferentID string              `json:"addressee_referent_id,omitempty"`
	Mentions            []MentionResolution `json:"mentions"`
	Conditions          []ContextCondition  `json:"conditions"`
	WorkUnits           int                 `json:"work_units"`
}

// CloneRequestContext returns a detached, bounded copy after validating the
// context-local structure. Document-dependent sort and expression checks are
// performed by CheckRequestContext.
func CloneRequestContext(context RequestContext) (RequestContext, error) {
	if err := validateContextShape(context); err != nil {
		return RequestContext{}, err
	}
	copyContext := RequestContext{
		SchemaVersion: context.SchemaVersion, ID: context.ID,
		SpeakerReferentID: context.SpeakerReferentID, AddresseeReferentID: context.AddresseeReferentID,
		Referents:          make([]ContextReferent, len(context.Referents)),
		Mentions:           make([]ContextMention, len(context.Mentions)),
		ResolutionEvidence: append([]ResolutionEvidence{}, context.ResolutionEvidence...),
		Conditions:         make([]ContextCondition, len(context.Conditions)),
	}
	state := newCopyState(DefaultLimits(), make(map[string]struct{}), nil, false)
	for i, referent := range context.Referents {
		typ, err := state.cloneType(&referent.Type, fmt.Sprintf("referents[%d].type", i), 1, make(map[*Type]bool))
		if err != nil {
			return RequestContext{}, err
		}
		copyContext.Referents[i] = ContextReferent{ID: referent.ID, Type: *typ, ConceptID: referent.ConceptID, TemporalFrame: referent.TemporalFrame, ReferenceTimeID: referent.ReferenceTimeID, Multiplicity: referent.Multiplicity}
	}
	copy(copyContext.Mentions, context.Mentions)
	termNodes := 0
	for i, condition := range context.Conditions {
		copyContext.Conditions[i] = ContextCondition{
			ID: condition.ID, Kind: condition.Kind, Code: condition.Code,
			TargetReferentIDs: append([]string{}, condition.TargetReferentIDs...), Detail: condition.Detail,
		}
		if condition.Expression == nil {
			continue
		}
		beforeNodes := state.nodes
		term, err := state.walk(condition.Expression, fmt.Sprintf("conditions[%d].expression", i), 1, 0, 1, "", nil, nil)
		if err != nil {
			return RequestContext{}, err
		}
		termNodes += state.nodes - beforeNodes
		if termNodes > MaxContextConditionNodes {
			return RequestContext{}, resourceLimit("conditions", "aggregate condition expression nodes exceed the request-context bound")
		}
		if err := validateTermShape(condition.Expression, fmt.Sprintf("conditions[%d].expression", i)); err != nil {
			return RequestContext{}, err
		}
		wire, err := termWire(term)
		if err != nil {
			return RequestContext{}, err
		}
		encoded, err := json.Marshal(wire)
		if err != nil {
			return RequestContext{}, invalid("INVALID_CONDITION_TERM", fmt.Sprintf("conditions[%d].expression", i), "condition term could not be encoded")
		}
		decoded, err := DecodeTermJSON(encoded, fmt.Sprintf("conditions[%d].expression", i))
		if err != nil {
			return RequestContext{}, err
		}
		copyContext.Conditions[i].Expression = decoded
	}
	wire, err := requestContextWire(copyContext, false)
	if err != nil {
		return RequestContext{}, err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return RequestContext{}, invalid("CONTEXT_ENCODE_FAILED", "context", "request context could not be encoded")
	}
	if len(encoded) > MaxRequestContextBytes {
		return RequestContext{}, resourceLimit("context", "request context exceeds one MiB")
	}
	return copyContext, nil
}

// CheckRequestContext validates a context against the actual document's
// declaration authority, then derives unresolved/resolved/ambiguous mention
// status only from explicit evidence rows.
func CheckRequestContext(context RequestContext, actualDoc Document) (ContextAssessment, error) {
	copyContext, err := CloneRequestContext(context)
	if err != nil {
		return ContextAssessment{}, err
	}
	contextDocForAttrs := actualDoc
	contextDocForAttrs.RequestContext = &copyContext
	if err := validateDocumentReferenceAttributes(contextDocForAttrs); err != nil {
		return ContextAssessment{}, err
	}
	docExpression := actualDoc.Expression
	actualDoc.Expression = nil
	actualDoc.RequestContext = nil
	work := len(copyContext.Referents) + len(copyContext.Mentions) + len(copyContext.ResolutionEvidence) + len(copyContext.Conditions)
	referents := make(map[string]ContextReferent, len(copyContext.Referents))
	for _, referent := range copyContext.Referents {
		referents[referent.ID] = referent
	}
	for i, referent := range context.Referents {
		if referent.ReferenceTimeID == "" {
			continue
		}
		target, ok := referents[referent.ReferenceTimeID]
		if !ok || target.Type.Kind != TypeInterval {
			return ContextAssessment{}, invalid("INVALID_REFERENCE_TIME_TARGET", fmt.Sprintf("referents[%d].reference_time_id", i), "reference time must name a declared Interval referent in this context")
		}
	}
	for i, declaration := range actualDoc.References {
		referent, ok := referents[declaration.ID]
		if !ok {
			return ContextAssessment{}, invalid("CONTEXT_REFERENCE_MISSING", fmt.Sprintf("document.references[%d]", i), "every document reference must exist as a typed context referent")
		}
		if !TypeEqual(declaration.Type, referent.Type) {
			return ContextAssessment{}, invalid("CONTEXT_REFERENCE_TYPE_MISMATCH", fmt.Sprintf("document.references[%d].type", i), "context referent type must match the document reference declaration")
		}
	}
	contextDoc := actualDoc
	contextDoc.References = append([]Declaration{}, actualDoc.References...)
	knownReferences := make(map[string]Type, len(contextDoc.References))
	for _, declaration := range contextDoc.References {
		knownReferences[declaration.ID] = declaration.Type
	}
	for _, referent := range copyContext.Referents {
		if prior, ok := knownReferences[referent.ID]; ok {
			if !TypeEqual(prior, referent.Type) {
				return ContextAssessment{}, invalid("CONTEXT_REFERENCE_TYPE_MISMATCH", "referents."+referent.ID+".type", "context referent type conflicts with the document declaration")
			}
			continue
		}
		contextDoc.References = append(contextDoc.References, Declaration{ID: referent.ID, Type: referent.Type, TemporalFrame: referent.TemporalFrame, ReferenceTimeID: referent.ReferenceTimeID, Multiplicity: referent.Multiplicity})
		knownReferences[referent.ID] = referent.Type
	}
	actualDoc = contextDoc
	for _, anchor := range []struct{ field, id string }{
		{"speaker_referent_id", copyContext.SpeakerReferentID},
		{"addressee_referent_id", copyContext.AddresseeReferentID},
	} {
		field, id := anchor.field, anchor.id
		if id != "" {
			if _, ok := referents[id]; !ok {
				return ContextAssessment{}, invalid("UNKNOWN_CONTEXT_ANCHOR", field, "anchor must name a referent in this request context")
			}
		}
	}
	contextDoc = actualDoc
	if err := CheckDeclarations(contextDoc); err != nil {
		return ContextAssessment{}, err
	}
	if docExpression != nil {
		docWithExpression := contextDoc
		docWithExpression.Expression = docExpression
		checked, checkErr := Check(docWithExpression)
		if checkErr != nil {
			return ContextAssessment{}, checkErr
		}
		work += checked.NodeCount
		if work > MaxContextWork {
			return ContextAssessment{}, resourceLimit("context", "request-context work exceeds the finite checker bound")
		}
	}
	for i, condition := range copyContext.Conditions {
		for _, target := range condition.TargetReferentIDs {
			if _, ok := referents[target]; !ok {
				return ContextAssessment{}, invalid("UNKNOWN_CONDITION_REFERENT", fmt.Sprintf("conditions[%d].target_referent_ids", i), "condition target must name a referent in this request context")
			}
		}
		if condition.Expression == nil {
			continue
		}
		conditionDoc := contextDoc
		conditionDoc.Expression = condition.Expression
		checked, err := Check(conditionDoc)
		if err != nil {
			return ContextAssessment{}, err
		}
		if checked.Type.Kind != TypeProp {
			return ContextAssessment{}, invalid("CONDITION_NOT_PROP", fmt.Sprintf("conditions[%d].expression", i), "condition expression must have type Prop")
		}
		work += checked.NodeCount
		if work > MaxContextWork || checked.NodeCount > MaxContextConditionNodes {
			return ContextAssessment{}, resourceLimit(fmt.Sprintf("conditions[%d].expression", i), "condition checking exceeds the context work bound")
		}
	}
	resolutions := make([]MentionResolution, 0, len(copyContext.Mentions))
	evidenceByMention := make(map[string][]ResolutionEvidence, len(copyContext.Mentions))
	for _, row := range copyContext.ResolutionEvidence {
		evidenceByMention[row.MentionID] = append(evidenceByMention[row.MentionID], row)
	}
	for _, mention := range copyContext.Mentions {
		rows := evidenceByMention[mention.ID]
		if mention.ReferentID != "" && len(rows) == 0 {
			return ContextAssessment{}, invalid("MISSING_IDENTITY_EVIDENCE", "mentions."+mention.ID+".referent_id", "declared mention target requires explicit corroborating resolution evidence")
		}
		candidateSet := make(map[string]struct{}, len(rows)+1)
		evidenceIDs := make([]string, 0, len(rows))
		for _, row := range rows {
			candidateSet[row.ReferentID] = struct{}{}
			evidenceIDs = append(evidenceIDs, row.ID)
		}
		if mention.ReferentID != "" {
			candidateSet[mention.ReferentID] = struct{}{}
		}
		candidates := make([]string, 0, len(candidateSet))
		for id := range candidateSet {
			candidates = append(candidates, id)
		}
		sort.Strings(candidates)
		sort.Strings(evidenceIDs)
		state := MentionUnresolved
		if len(candidateSet) == 1 && len(rows) > 0 {
			state = MentionResolved
		} else if len(candidateSet) > 1 {
			state = MentionAmbiguous
		}
		resolutions = append(resolutions, MentionResolution{MentionID: mention.ID, Status: state, CandidateReferentIDs: candidates, EvidenceIDs: evidenceIDs})
	}
	fingerprint, err := fingerprintRequestContext(copyContext)
	if err != nil {
		return ContextAssessment{}, err
	}
	return ContextAssessment{
		ContextID: copyContext.ID, Fingerprint: fingerprint,
		SpeakerReferentID: copyContext.SpeakerReferentID, AddresseeReferentID: copyContext.AddresseeReferentID,
		Mentions: resolutions, Conditions: copyContext.Conditions, WorkUnits: work,
	}, nil
}

func collectContextExpressionReferences(root *Term) (map[string]struct{}, error) {
	refs := make(map[string]struct{})
	type entry struct {
		term  *Term
		depth int
	}
	stack := []entry{{term: root, depth: 1}}
	work := 0
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		if current.term == nil {
			continue
		}
		work++
		if work > DefaultMaxNodes || current.depth > DefaultMaxDepth {
			return nil, resourceLimit("document.expression", "reference collection exceeds the expression bound")
		}
		term := current.term
		if term.Kind == TermRef {
			refs[term.ID] = struct{}{}
		}
		push := func(child *Term) {
			if child != nil {
				stack = append(stack, entry{term: child, depth: current.depth + 1})
			}
		}
		switch term.Kind {
		case TermAtom:
			for i := len(term.Args) - 1; i >= 0; i-- {
				push(term.Args[i].Value)
			}
		case TermAnd:
			for i := len(term.Operands) - 1; i >= 0; i-- {
				push(term.Operands[i])
			}
		case TermExists, TermLambda, TermNot, TermModal:
			push(term.Body)
		case TermQuantifier, TermAsk:
			push(term.Body)
			push(term.Restrictor)
		case TermApply:
			push(term.Argument)
			push(term.Function)
		case TermVar, TermParam, TermRef, TermHole:
		default:
			return nil, invalid("UNKNOWN_TERM_KIND", "document.expression.kind", "term kind is unsupported")
		}
	}
	return refs, nil
}

// EncodeRequestContextJSON validates and serializes a bounded context against
// the actual document declarations. Empty collections are emitted as arrays.
func EncodeRequestContextJSON(context RequestContext, actualDoc Document) ([]byte, error) {
	if _, err := CheckRequestContext(context, actualDoc); err != nil {
		return nil, err
	}
	copyContext, err := CloneRequestContext(context)
	if err != nil {
		return nil, err
	}
	wire, err := requestContextWire(copyContext, false)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("CONTEXT_ENCODE_FAILED", "context", "request context could not be encoded")
	}
	if len(encoded) > MaxRequestContextBytes {
		return nil, resourceLimit("context", "encoded request context exceeds one MiB")
	}
	return encoded, nil
}

// RequestContextFingerprint hashes a canonical bounded representation. The
// context is checked against the actual declaration authority first.
func RequestContextFingerprint(context RequestContext, actualDoc Document) (string, error) {
	if _, err := CheckRequestContext(context, actualDoc); err != nil {
		return "", err
	}
	copyContext, err := CloneRequestContext(context)
	if err != nil {
		return "", err
	}
	wire, err := requestContextWire(copyContext, true)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return "", invalid("CONTEXT_ENCODE_FAILED", "context", "request context could not be canonicalized")
	}
	if len(encoded) > MaxRequestContextBytes {
		return "", resourceLimit("context", "canonical request context exceeds one MiB")
	}
	digest := sha256.Sum256(encoded)
	return "algebra-request-context-v1:sha256:" + hex.EncodeToString(digest[:]), nil
}

// CanonicalRequestContext returns a detached context with ID-ordered record
// collections and its independently checked request-local fingerprint.
func CanonicalRequestContext(context RequestContext, actualDoc Document) (RequestContext, string, error) {
	fingerprint, err := RequestContextFingerprint(context, actualDoc)
	if err != nil {
		return RequestContext{}, "", err
	}
	copyContext, err := CloneRequestContext(context)
	if err != nil {
		return RequestContext{}, "", err
	}
	sort.Slice(copyContext.Referents, func(i, j int) bool { return copyContext.Referents[i].ID < copyContext.Referents[j].ID })
	sort.Slice(copyContext.Mentions, func(i, j int) bool { return copyContext.Mentions[i].ID < copyContext.Mentions[j].ID })
	sort.Slice(copyContext.ResolutionEvidence, func(i, j int) bool {
		return copyContext.ResolutionEvidence[i].ID < copyContext.ResolutionEvidence[j].ID
	})
	sort.Slice(copyContext.Conditions, func(i, j int) bool { return copyContext.Conditions[i].ID < copyContext.Conditions[j].ID })
	for i := range copyContext.Conditions {
		sort.Strings(copyContext.Conditions[i].TargetReferentIDs)
	}
	return copyContext, fingerprint, nil
}

func requestContextWireBytes(context RequestContext, canonical bool) ([]byte, error) {
	copyContext, err := CloneRequestContext(context)
	if err != nil {
		return nil, err
	}
	wire, err := requestContextWire(copyContext, canonical)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, invalid("CONTEXT_ENCODE_FAILED", "context", "request context could not be encoded")
	}
	if len(encoded) > MaxRequestContextBytes {
		return nil, resourceLimit("context", "request context exceeds one MiB")
	}
	return encoded, nil
}

func requestContextRaw(context RequestContext, canonical bool) (json.RawMessage, error) {
	encoded, err := requestContextWireBytes(context, canonical)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(encoded), nil
}

func validateContextShape(context RequestContext) error {
	if context.SchemaVersion != RequestContextSchemaVersion {
		return invalid("INVALID_SCHEMA_VERSION", "schema_version", "request-context schema version must be 1")
	}
	if !validText(context.ID, MaxContextIDBytes) {
		return invalid("INVALID_CONTEXT_ID", "id", "context ID must be non-empty valid UTF-8 and at most 256 bytes")
	}
	if len(context.Referents) > MaxContextReferents || len(context.Mentions) > MaxContextMentions || len(context.ResolutionEvidence) > MaxContextEvidence || len(context.Conditions) > MaxContextConditions {
		return resourceLimit("context", "request-context collection exceeds its finite item bound")
	}
	work := len(context.Referents) + len(context.Mentions) + len(context.ResolutionEvidence) + len(context.Conditions)
	if work > MaxContextWork {
		return resourceLimit("context", "request-context work exceeds the finite checker bound")
	}
	referents := make(map[string]ContextReferent, len(context.Referents))
	for i, referent := range context.Referents {
		path := fmt.Sprintf("referents[%d]", i)
		if !validText(referent.ID, MaxContextIDBytes) {
			return invalid("INVALID_REFERENT_ID", path+".id", "referent ID is malformed or exceeds its bound")
		}
		if _, exists := referents[referent.ID]; exists {
			return invalid("DUPLICATE_REFERENT_ID", path+".id", "referent ID is duplicated")
		}
		if !validContextDataType(referent.Type) {
			return invalid("INVALID_REFERENT_TYPE", path+".type", "referent type must be Entity, Event, Interval, or a declared Nominal sort")
		}
		if err := validateReferenceAttributes(referent.ID, referent.Type, contextReferentAttributes(referent), path); err != nil {
			return err
		}
		if referent.Type.Kind == TypeNominal && !validText(referent.Type.SortID, 256) {
			return invalid("INVALID_REFERENT_TYPE", path+".type.sort_id", "nominal referent sort ID is malformed")
		}
		if referent.ConceptID != "" {
			if !validText(referent.ConceptID, MaxContextIDBytes) {
				return invalid("INVALID_CONCEPT_ID", path+".concept_id", "optional concept classification must be valid bounded text")
			}
			if _, err := ids.ParseConceptID(referent.ConceptID); err != nil {
				return invalid("INVALID_CONCEPT_ID", path+".concept_id", "optional concept classification must be a valid ConceptID")
			}
		}
		referents[referent.ID] = referent
		localDoc := Document{SchemaVersion: 1, References: []Declaration{{ID: referent.ID, Type: referent.Type}}}
		if referent.Type.Kind == TypeNominal {
			localDoc.Sorts = []SortDeclaration{{ID: referent.Type.SortID}}
		}
		if err := CheckDeclarations(localDoc); err != nil {
			return err
		}
	}
	for _, anchor := range []struct{ label, id string }{{"speaker_referent_id", context.SpeakerReferentID}, {"addressee_referent_id", context.AddresseeReferentID}} {
		label, id := anchor.label, anchor.id
		if id != "" {
			if !validText(id, MaxContextIDBytes) {
				return invalid("INVALID_CONTEXT_ANCHOR", label, "anchor ID is malformed or exceeds its bound")
			}
			if _, exists := referents[id]; !exists {
				return invalid("UNKNOWN_CONTEXT_ANCHOR", label, "anchor must name a referent in this request context")
			}
		}
	}
	mentions := make(map[string]ContextMention, len(context.Mentions))
	for i, mention := range context.Mentions {
		path := fmt.Sprintf("mentions[%d]", i)
		if !validText(mention.ID, MaxContextIDBytes) {
			return invalid("INVALID_MENTION_ID", path+".id", "mention ID is malformed or exceeds its bound")
		}
		if _, exists := mentions[mention.ID]; exists {
			return invalid("DUPLICATE_MENTION_ID", path+".id", "mention ID is duplicated")
		}
		if mention.ContextID != "" && mention.ContextID != context.ID {
			return invalid("CROSS_CONTEXT_MENTION", path+".context_id", "mention context ID must match the containing request context")
		}
		if mention.ContextID != "" && !validText(mention.ContextID, MaxContextIDBytes) {
			return invalid("INVALID_MENTION_CONTEXT_ID", path+".context_id", "mention context ID is malformed")
		}
		if mention.ReferentID != "" {
			if !validText(mention.ReferentID, MaxContextIDBytes) {
				return invalid("INVALID_MENTION_REFERENT_ID", path+".referent_id", "declared referent ID is malformed")
			}
			if _, exists := referents[mention.ReferentID]; !exists {
				return invalid("UNKNOWN_MENTION_REFERENT", path+".referent_id", "declared mention target is missing from this request context")
			}
		}
		mentions[mention.ID] = mention
	}
	evidenceIDs := make(map[string]struct{}, len(context.ResolutionEvidence))
	for i, row := range context.ResolutionEvidence {
		path := fmt.Sprintf("resolution_evidence[%d]", i)
		if !validText(row.ID, MaxContextIDBytes) || !validText(row.MentionID, MaxContextIDBytes) || !validText(row.ReferentID, MaxContextIDBytes) {
			return invalid("INVALID_RESOLUTION_EVIDENCE", path, "evidence and linked IDs must be non-empty bounded text")
		}
		if _, exists := evidenceIDs[row.ID]; exists {
			return invalid("DUPLICATE_EVIDENCE_ID", path+".id", "resolution evidence ID is duplicated")
		}
		evidenceIDs[row.ID] = struct{}{}
		if _, exists := mentions[row.MentionID]; !exists {
			return invalid("UNKNOWN_EVIDENCE_MENTION", path+".mention_id", "evidence must name a mention in this request context")
		}
		if _, exists := referents[row.ReferentID]; !exists {
			return invalid("UNKNOWN_EVIDENCE_REFERENT", path+".referent_id", "evidence must name a referent in this request context")
		}
		if err := row.Source.Validate(); err != nil {
			return invalid("INVALID_EVIDENCE_SOURCE", path+".source", err.Error())
		}
	}
	conditionIDs := make(map[string]struct{}, len(context.Conditions))
	termNodes := 0
	for i, condition := range context.Conditions {
		path := fmt.Sprintf("conditions[%d]", i)
		if !validText(condition.ID, MaxContextIDBytes) || !validText(string(condition.Kind), 32) || !validText(condition.Code, MaxContextIDBytes) {
			return invalid("INVALID_CONTEXT_CONDITION", path, "condition ID, kind, and code must be bounded non-empty text")
		}
		switch condition.Kind {
		case ContextAssumption, ContextPresupposition, ContextSemanticLoss:
		default:
			return invalid("INVALID_CONTEXT_CONDITION_KIND", path+".kind", "condition kind must be assumption, presupposition, or loss")
		}
		if _, exists := conditionIDs[condition.ID]; exists {
			return invalid("DUPLICATE_CONDITION_ID", path+".id", "condition ID is duplicated")
		}
		conditionIDs[condition.ID] = struct{}{}
		if !validOptionalText(condition.Detail, MaxContextHumanTextBytes) {
			return invalid("INVALID_CONDITION_DETAIL", path+".detail", "condition detail must be valid UTF-8 and at most 1024 bytes")
		}
		if len(condition.TargetReferentIDs) > MaxContextReferents {
			return resourceLimit(path+".target_referent_ids", "condition target list exceeds the referent bound")
		}
		seenTargets := make(map[string]struct{}, len(condition.TargetReferentIDs))
		for j, target := range condition.TargetReferentIDs {
			if !validText(target, MaxContextIDBytes) {
				return invalid("INVALID_CONDITION_TARGET", fmt.Sprintf("%s.target_referent_ids[%d]", path, j), "condition target ID is malformed")
			}
			if _, exists := referents[target]; !exists {
				return invalid("UNKNOWN_CONDITION_REFERENT", fmt.Sprintf("%s.target_referent_ids[%d]", path, j), "condition target must name a referent in this context")
			}
			if _, exists := seenTargets[target]; exists {
				return invalid("DUPLICATE_CONDITION_TARGET", fmt.Sprintf("%s.target_referent_ids[%d]", path, j), "condition target IDs are unique")
			}
			seenTargets[target] = struct{}{}
		}
		if condition.Expression != nil {
			state := newCopyState(DefaultLimits(), make(map[string]struct{}), nil, false)
			if _, err := state.walk(condition.Expression, path+".expression", 1, 0, 1, "", nil, nil); err != nil {
				return err
			}
			termNodes += state.nodes
			if termNodes > MaxContextConditionNodes {
				return resourceLimit("conditions", "aggregate condition expression nodes exceed 4096")
			}
		}
	}
	wire, err := requestContextWire(context, false)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return invalid("CONTEXT_ENCODE_FAILED", "context", "request context could not be encoded")
	}
	if len(encoded) > MaxRequestContextBytes {
		return resourceLimit("context", "request context exceeds one MiB")
	}
	return nil
}

func validContextDataType(value Type) bool {
	switch value.Kind {
	case TypeEntity, TypeEvent, TypeInterval:
		return value.SortID == "" && value.From == nil && value.To == nil
	case TypeNominal:
		return validText(value.SortID, 256) && value.From == nil && value.To == nil
	default:
		return false
	}
}

func fingerprintRequestContext(context RequestContext) (string, error) {
	wire, err := requestContextWire(context, true)
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return "", invalid("CONTEXT_ENCODE_FAILED", "context", "request context could not be canonicalized")
	}
	if len(encoded) > MaxRequestContextBytes {
		return "", resourceLimit("context", "canonical request context exceeds one MiB")
	}
	digest := sha256.Sum256(encoded)
	return "algebra-request-context-v1:sha256:" + hex.EncodeToString(digest[:]), nil
}

func requestContextWire(context RequestContext, canonical bool) (map[string]any, error) {
	referents := append([]ContextReferent{}, context.Referents...)
	mentions := append([]ContextMention{}, context.Mentions...)
	evidence := append([]ResolutionEvidence{}, context.ResolutionEvidence...)
	conditions := append([]ContextCondition{}, context.Conditions...)
	if canonical {
		sort.Slice(referents, func(i, j int) bool { return referents[i].ID < referents[j].ID })
		sort.Slice(mentions, func(i, j int) bool { return mentions[i].ID < mentions[j].ID })
		sort.Slice(evidence, func(i, j int) bool { return evidence[i].ID < evidence[j].ID })
		sort.Slice(conditions, func(i, j int) bool { return conditions[i].ID < conditions[j].ID })
	}
	referentWires := make([]map[string]any, 0, len(referents))
	for _, referent := range referents {
		item := map[string]any{"id": referent.ID, "type": referent.Type}
		if referent.ConceptID != "" {
			item["concept_id"] = referent.ConceptID
		}
		if referent.TemporalFrame != "" {
			item["temporal_frame"] = referent.TemporalFrame
		}
		if referent.ReferenceTimeID != "" {
			item["reference_time_id"] = referent.ReferenceTimeID
		}
		if referent.Multiplicity != "" {
			item["multiplicity"] = referent.Multiplicity
		}
		referentWires = append(referentWires, item)
	}
	mentionWires := make([]map[string]any, 0, len(mentions))
	for _, mention := range mentions {
		item := map[string]any{"id": mention.ID}
		if mention.ContextID != "" {
			item["context_id"] = mention.ContextID
		}
		if mention.ReferentID != "" {
			item["referent_id"] = mention.ReferentID
		}
		mentionWires = append(mentionWires, item)
	}
	evidenceWires := make([]map[string]any, 0, len(evidence))
	for _, row := range evidence {
		evidenceWires = append(evidenceWires, map[string]any{"id": row.ID, "mention_id": row.MentionID, "referent_id": row.ReferentID, "source": row.Source})
	}
	conditionWires := make([]map[string]any, 0, len(conditions))
	for i, condition := range conditions {
		targets := append([]string{}, condition.TargetReferentIDs...)
		if canonical {
			sort.Strings(targets)
		}
		item := map[string]any{
			"id": condition.ID, "kind": condition.Kind, "code": condition.Code,
			"target_referent_ids": targets,
		}
		if condition.Detail != "" {
			item["detail"] = condition.Detail
		}
		if condition.Expression != nil {
			term, err := termWire(condition.Expression)
			if err != nil {
				return nil, invalid("INVALID_CONDITION_TERM", fmt.Sprintf("conditions[%d].expression", i), err.Error())
			}
			item["expression"] = term
		}
		conditionWires = append(conditionWires, item)
	}
	wire := map[string]any{
		"schema_version": context.SchemaVersion, "id": context.ID,
		"referents": referentWires, "mentions": mentionWires,
		"resolution_evidence": evidenceWires, "conditions": conditionWires,
	}
	if context.SpeakerReferentID != "" {
		wire["speaker_referent_id"] = context.SpeakerReferentID
	}
	if context.AddresseeReferentID != "" {
		wire["addressee_referent_id"] = context.AddresseeReferentID
	}
	return wire, nil
}
