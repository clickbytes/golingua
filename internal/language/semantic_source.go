package language

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	SourcePlanSchemaVersion = 1
	SourcePlanProfile       = "typed-source-fragments-v1"
	MaxSourcePlanBytes      = 1 << 20
	MaxSourcePredicateMaps  = 128
	MaxSourceFeatureMaps    = 128
	MaxSourceOperators      = 16
	MaxSourcePOSValues      = 64
	MaxSourcePlanText       = 256
	MaxSourceCodeText       = 128
)

var sourcePlanVersionRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

type SourceConditionKind string

const (
	SourceAssumption     SourceConditionKind = "assumption"
	SourcePresupposition SourceConditionKind = "presupposition"
)

// SourceCondition is an explicitly declared non-asserted placement label.
// Code is data for the consuming composition profile, never executable logic.
type SourceCondition struct {
	Kind SourceConditionKind `json:"kind"`
	Code string              `json:"code"`
}

type SourcePredicateMapping struct {
	SenseID            ids.SenseID              `json:"sense_id"`
	FrameID            valency.FrameID          `json:"frame_id"`
	HeadPortID         string                   `json:"head_port_id"`
	PredicateKind      string                   `json:"predicate_kind,omitempty"`
	AllowedFamilyIDs   []ids.ConstructionID     `json:"allowed_family_ids"`
	AllowedVariantIDs  []construction.VariantID `json:"allowed_variant_ids"`
	ReferenceCondition *SourceCondition         `json:"reference_condition,omitempty"`
}

type SourceOperatorMapping struct {
	SourceKind string           `json:"source_kind"`
	Operator   algebra.TermKind `json:"operator"`
}

type SourceFeatureCondition struct {
	FeatureID string              `json:"feature_id"`
	ValueID   string              `json:"value_id"`
	SymbolRef string              `json:"symbol_ref"`
	PortID    string              `json:"port_id"`
	Type      algebra.Type        `json:"type"`
	Condition SourceConditionKind `json:"condition_kind"`
	Code      string              `json:"code"`
}

// SourcePlanDocument is the closed declarative input. It contains no source
// text, callback, or sentence-level construction.
type SourcePlanDocument struct {
	SchemaVersion                int                       `json:"schema_version"`
	ID                           string                    `json:"id"`
	Version                      string                    `json:"version"`
	Language                     string                    `json:"language"`
	Profile                      string                    `json:"profile"`
	Provenance                   provenance.SourceRef      `json:"provenance"`
	PredicateCategory            construction.SlotCategory `json:"predicate_category"`
	ReferenceCategory            construction.SlotCategory `json:"reference_category"`
	ReferenceNonHeadPOSAllowlist []string                  `json:"reference_non_head_pos_allowlist"`
	Predicates                   []SourcePredicateMapping  `json:"predicate_mappings"`
	Operators                    []SourceOperatorMapping   `json:"operator_mappings"`
	FeatureConditions            []SourceFeatureCondition  `json:"feature_conditions"`
	RequiredFeatureIDs           []string                  `json:"required_feature_ids"`
}

// SourcePlan is an immutable validated snapshot tied to one selected semantic
// package and primitive context. Methods return detached data.
type SourcePlan struct {
	document SourcePlanDocument
}

func LoadSourcePlan(data []byte, pack *SemanticLanguagePackage, selected algebra.Document) (*SourcePlan, error) {
	doc, err := DecodeSourcePlan(data)
	if err != nil {
		return nil, err
	}
	if pack == nil || pack.lexicon == nil || pack.valency == nil {
		return nil, invalid("MISSING_SEMANTIC_PACKAGE", "semantic_package", "a built selected semantic language package is required")
	}
	if err := validateSelectedPrimitiveContext(selected); err != nil {
		return nil, err
	}
	if doc.Language != string(pack.Language()) {
		return nil, invalid("SOURCE_PLAN_LANGUAGE_MISMATCH", "language", "source plan language must exactly match the selected language package")
	}
	if err := validateSourcePlanSelection(doc, pack, selected); err != nil {
		return nil, err
	}
	return &SourcePlan{document: cloneSourcePlanDocument(doc)}, nil
}

func DecodeSourcePlan(data []byte) (SourcePlanDocument, error) {
	if err := algebra.PreflightJSONLimit(data, MaxSourcePlanBytes); err != nil {
		return SourcePlanDocument{}, err
	}
	fields, err := languageStrictObject(data, "source_plan", []string{"schema_version", "id", "version", "language", "profile", "provenance", "predicate_category", "reference_category", "reference_non_head_pos_allowlist", "predicate_mappings", "operator_mappings", "feature_conditions", "required_feature_ids"}, []string{"schema_version", "id", "version", "language", "profile", "provenance", "predicate_category", "reference_category", "predicate_mappings", "operator_mappings", "feature_conditions"})
	if err != nil {
		return SourcePlanDocument{}, err
	}
	var out SourcePlanDocument
	for key, dst := range map[string]*int{"schema_version": &out.SchemaVersion} {
		if err := languageStrictValue(fields[key], dst); err != nil {
			return out, invalid("INVALID_SOURCE_PLAN_FIELD", key, "field must be a non-null integer")
		}
	}
	for key, dst := range map[string]*string{"id": &out.ID, "version": &out.Version, "language": &out.Language, "profile": &out.Profile} {
		if err := languageStrictValue(fields[key], dst); err != nil {
			return out, invalid("INVALID_SOURCE_PLAN_FIELD", key, "field must be a non-null string")
		}
	}
	for key, dst := range map[string]*string{"predicate_category": (*string)(&out.PredicateCategory), "reference_category": (*string)(&out.ReferenceCategory)} {
		if err := languageStrictValue(fields[key], dst); err != nil {
			return out, invalid("INVALID_SOURCE_PLAN_FIELD", key, "category must be a non-null string")
		}
	}
	if err := validateSourcePlanHeader(out); err != nil {
		return out, err
	}
	out.Provenance, err = decodePackageSource(fields["provenance"], "provenance")
	if err != nil {
		return out, err
	}
	if err := out.Provenance.Validate(); err != nil {
		return out, invalid("INVALID_SOURCE_PLAN_PROVENANCE", "provenance", "source plan provenance is malformed")
	}
	if raw, ok := fields["reference_non_head_pos_allowlist"]; ok {
		if err := languageStrictArrayInto(raw, &out.ReferenceNonHeadPOSAllowlist, "reference_non_head_pos_allowlist", MaxSourcePOSValues); err != nil {
			return out, err
		}
	} else {
		out.ReferenceNonHeadPOSAllowlist = []string{}
	}
	if err := decodeSourcePredicateMappings(fields["predicate_mappings"], &out.Predicates); err != nil {
		return out, err
	}
	if err := decodeSourceOperatorMappings(fields["operator_mappings"], &out.Operators); err != nil {
		return out, err
	}
	if err := decodeSourceFeatureConditions(fields["feature_conditions"], &out.FeatureConditions); err != nil {
		return out, err
	}
	if raw, ok := fields["required_feature_ids"]; ok {
		if err := languageStrictArrayInto(raw, &out.RequiredFeatureIDs, "required_feature_ids", MaxSourceFeatureMaps); err != nil {
			return out, err
		}
	} else {
		out.RequiredFeatureIDs = []string{}
	}
	if err := validateSourcePlanData(out); err != nil {
		return out, err
	}
	return cloneSourcePlanDocument(out), nil
}

func (plan *SourcePlan) Document() SourcePlanDocument {
	if plan == nil {
		return SourcePlanDocument{}
	}
	return cloneSourcePlanDocument(plan.document)
}

func (plan *SourcePlan) ID() string {
	if plan == nil {
		return ""
	}
	return plan.document.ID
}
func (plan *SourcePlan) Version() string {
	if plan == nil {
		return ""
	}
	return plan.document.Version
}
func (plan *SourcePlan) Language() string {
	if plan == nil {
		return ""
	}
	return plan.document.Language
}

func DecodeSourceReferenceContext(data []byte, selected algebra.Document) (SourceReferenceContext, error) {
	if err := algebra.PreflightJSONLimit(data, MaxSemanticBindingRequestBytes); err != nil {
		return SourceReferenceContext{}, err
	}
	fields, err := languageStrictObject(data, "source_reference_context", []string{"schema_version", "context_id", "references", "request_context", "role_references", "port_references"}, []string{"schema_version", "context_id", "references", "role_references", "port_references"})
	if err != nil {
		return SourceReferenceContext{}, err
	}
	var out SourceReferenceContext
	if err := languageStrictValue(fields["schema_version"], &out.SchemaVersion); err != nil || out.SchemaVersion != 1 {
		return out, invalid("UNSUPPORTED_SOURCE_CONTEXT_VERSION", "schema_version", "source reference context schema_version must be 1")
	}
	if err := languageStrictValue(fields["context_id"], &out.ContextID); err != nil || !validSemanticPackageText(out.ContextID, algebra.MaxContextIDBytes) {
		return out, invalid("INVALID_SOURCE_CONTEXT_ID", "context_id", "context ID must be explicit and bounded")
	}
	if out.References, err = decodeSemanticRequestReferences(fields["references"]); err != nil {
		return out, err
	}
	if out.RoleReferences, err = decodeSemanticRoleReferences(fields["role_references"]); err != nil {
		return out, err
	}
	if out.PortReferences, err = decodeSemanticPortReferences(fields["port_references"]); err != nil {
		return out, err
	}
	if raw, ok := fields["request_context"]; ok {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return out, invalid("INVALID_REQUEST_CONTEXT", "request_context", "request_context cannot be null when present")
		}
		request, e := algebra.DecodeRequestContextJSON(raw)
		if e != nil {
			return out, e
		}
		if request.ID != out.ContextID {
			return out, invalid("REQUEST_CONTEXT_ID_MISMATCH", "request_context.id", "request-context ID must match context_id")
		}
		out.RequestContext = &request
	}
	if len(out.References) > algebra.DefaultMaxSymbols {
		return out, resource("source_context_references", len(out.References), algebra.DefaultMaxSymbols)
	}
	if _, err := mergeReferences(nil, out.References); err != nil {
		return out, err
	}
	if err := validateSourceContextDeclarations(out, selected); err != nil {
		return out, err
	}
	return cloneSourceReferenceContext(out), nil
}

type SourceReferenceContext struct {
	SchemaVersion  int                     `json:"schema_version"`
	ContextID      string                  `json:"context_id"`
	References     []algebra.Declaration   `json:"references"`
	RequestContext *algebra.RequestContext `json:"request_context,omitempty"`
	RoleReferences []SemanticRoleReference `json:"role_references"`
	PortReferences []SemanticPortReference `json:"port_references"`
}

func (context SourceReferenceContext) Clone() SourceReferenceContext {
	return cloneSourceReferenceContext(context)
}

func validateSourcePlanHeader(doc SourcePlanDocument) error {
	if doc.SchemaVersion != SourcePlanSchemaVersion {
		return invalid("UNSUPPORTED_SOURCE_PLAN_VERSION", "schema_version", "source plan schema version must be 1")
	}
	if !validSemanticPackageText(doc.ID, MaxSourcePlanText) || !sourcePlanVersionRE.MatchString(doc.Version) {
		return invalid("INVALID_SOURCE_PLAN_IDENTITY", "source_plan", "plan ID and version must be bounded")
	}
	if !validSemanticPackageText(doc.Language, 35) {
		return invalid("INVALID_SOURCE_PLAN_LANGUAGE", "language", "language tag must be explicit and bounded")
	}
	if doc.Profile != SourcePlanProfile {
		return unsupported("UNSUPPORTED_SOURCE_PLAN_PROFILE", "profile", "source plan profile is not supported")
	}
	if doc.PredicateCategory != construction.CategoryVP || doc.ReferenceCategory != construction.CategoryNP {
		return invalid("INVALID_SOURCE_PLAN_CATEGORIES", "source_plan", "predicate and reference categories must be VP and NP")
	}
	return nil
}

func decodeSourcePredicateMappings(raw []byte, out *[]SourcePredicateMapping) error {
	rows, err := languageStrictArray(raw, "predicate_mappings", MaxSourcePredicateMaps)
	if err != nil {
		return err
	}
	result := make([]SourcePredicateMapping, 0, len(rows))
	for i, item := range rows {
		path := fmt.Sprintf("predicate_mappings[%d]", i)
		f, e := languageStrictObject(item, path, []string{"sense_id", "frame_id", "head_port_id", "predicate_kind", "allowed_family_ids", "allowed_variant_ids", "reference_condition"}, []string{"sense_id", "frame_id", "head_port_id", "allowed_family_ids", "allowed_variant_ids"})
		if e != nil {
			return e
		}
		var row SourcePredicateMapping
		var senseText, frameText string
		if e = languageStrictValue(f["sense_id"], &senseText); e != nil {
			return invalid("INVALID_SOURCE_PREDICATE", path+".sense_id", "sense_id must be a non-null typed identifier")
		}
		row.SenseID, e = ids.ParseSenseID(senseText)
		if e != nil {
			return invalid("INVALID_SOURCE_PREDICATE", path+".sense_id", "sense_id must be a typed sense identifier")
		}
		if e = languageStrictValue(f["frame_id"], &frameText); e != nil || !validSemanticPackageText(frameText, 128) {
			return invalid("INVALID_SOURCE_PREDICATE", path+".frame_id", "frame_id must be a bounded non-null ID")
		}
		row.FrameID = valency.FrameID(frameText)
		if e = languageStrictValue(f["head_port_id"], &row.HeadPortID); e != nil {
			return invalid("INVALID_SOURCE_PREDICATE", path+".head_port_id", "head_port_id must be a non-null string")
		}
		row.PredicateKind = "event"
		if rawKind, present := f["predicate_kind"]; present {
			if e = languageStrictValue(rawKind, &row.PredicateKind); e != nil || (row.PredicateKind != "event" && row.PredicateKind != "relation" && row.PredicateKind != "property") {
				return invalid("INVALID_SOURCE_PREDICATE_KIND", path+".predicate_kind", "predicate_kind must be event, relation, or an explicitly validated property")
			}
		}
		if row.PredicateKind == "relation" || row.PredicateKind == "property" {
			if row.HeadPortID != "" {
				return invalid("INVALID_SOURCE_PREDICATE", path+".head_port_id", "non-event predicates must not declare an Event head port")
			}
		} else if !validSemanticPackageText(row.HeadPortID, 256) {
			return invalid("INVALID_SOURCE_PREDICATE", path+".head_port_id", "event predicates require a bounded non-null Event head port")
		}
		var familyTexts, variantTexts []string
		if e = languageStrictArrayInto(f["allowed_family_ids"], &familyTexts, path+".allowed_family_ids", construction.MaxFamilyItems); e != nil {
			return e
		}
		if len(familyTexts) == 0 {
			return invalid("EMPTY_SOURCE_FAMILIES", path+".allowed_family_ids", "at least one allowed family ID is required")
		}
		seen := map[string]bool{}
		for j, text := range familyTexts {
			id, parseErr := ids.ParseConstructionID(text)
			if parseErr != nil {
				return invalid("INVALID_SOURCE_FAMILY_ID", fmt.Sprintf("%s.allowed_family_ids[%d]", path, j), "family ID must be a typed construction ID")
			}
			if seen[id.String()] {
				return invalid("DUPLICATE_SOURCE_FAMILY_ID", path+".allowed_family_ids", "family IDs must be unique")
			}
			seen[id.String()] = true
			row.AllowedFamilyIDs = append(row.AllowedFamilyIDs, id)
		}
		seen = map[string]bool{}
		if e = languageStrictArrayInto(f["allowed_variant_ids"], &variantTexts, path+".allowed_variant_ids", construction.MaxFamilyItems); e != nil {
			return e
		}
		for j, text := range variantTexts {
			id, parseErr := construction.ParseVariantID(text)
			if parseErr != nil {
				return invalid("INVALID_SOURCE_VARIANT_ID", fmt.Sprintf("%s.allowed_variant_ids[%d]", path, j), "variant ID must be a typed construction ID")
			}
			if seen[id.String()] {
				return invalid("DUPLICATE_SOURCE_VARIANT_ID", path+".allowed_variant_ids", "variant IDs must be unique")
			}
			seen[id.String()] = true
			row.AllowedVariantIDs = append(row.AllowedVariantIDs, id)
		}
		if rawCondition, ok := f["reference_condition"]; ok {
			condition, e := decodeSourceCondition(rawCondition, path+".reference_condition")
			if e != nil {
				return e
			}
			row.ReferenceCondition = &condition
		}
		result = append(result, row)
	}
	*out = result
	return nil
}

func decodeSourceOperatorMappings(raw []byte, out *[]SourceOperatorMapping) error {
	rows, err := languageStrictArray(raw, "operator_mappings", MaxSourceOperators)
	if err != nil {
		return err
	}
	result := make([]SourceOperatorMapping, 0, len(rows))
	seen := map[string]bool{}
	for i, item := range rows {
		path := fmt.Sprintf("operator_mappings[%d]", i)
		f, e := languageStrictObject(item, path, []string{"source_kind", "operator"}, []string{"source_kind", "operator"})
		if e != nil {
			return e
		}
		var row SourceOperatorMapping
		if e = languageStrictValue(f["source_kind"], &row.SourceKind); e != nil || row.SourceKind != "negation" {
			return unsupported("UNSUPPORTED_SOURCE_OPERATOR", path+".source_kind", "only the declared negation source operator is supported")
		}
		if e = languageStrictValue(f["operator"], &row.Operator); e != nil || row.Operator != algebra.TermNot {
			return unsupported("UNSUPPORTED_SOURCE_OPERATOR", path+".operator", "negation maps only to algebra Not")
		}
		if seen[row.SourceKind] {
			return invalid("DUPLICATE_SOURCE_OPERATOR", path, "source operator mapping is duplicated")
		}
		seen[row.SourceKind] = true
		result = append(result, row)
	}
	*out = result
	return nil
}

func decodeSourceFeatureConditions(raw []byte, out *[]SourceFeatureCondition) error {
	rows, err := languageStrictArray(raw, "feature_conditions", MaxSourceFeatureMaps)
	if err != nil {
		return err
	}
	result := make([]SourceFeatureCondition, 0, len(rows))
	seen := map[string]bool{}
	for i, item := range rows {
		path := fmt.Sprintf("feature_conditions[%d]", i)
		f, e := languageStrictObject(item, path, []string{"feature_id", "value_id", "symbol_ref", "port_id", "type", "condition_kind", "code"}, []string{"feature_id", "value_id", "symbol_ref", "port_id", "type", "condition_kind", "code"})
		if e != nil {
			return e
		}
		var row SourceFeatureCondition
		for key, dst := range map[string]*string{"feature_id": &row.FeatureID, "value_id": &row.ValueID, "symbol_ref": &row.SymbolRef, "port_id": &row.PortID, "condition_kind": (*string)(&row.Condition), "code": &row.Code} {
			if e = languageStrictValue(f[key], dst); e != nil {
				return invalid("INVALID_SOURCE_FEATURE_CONDITION", path+"."+key, "field must be a non-null string")
			}
		}
		row.Type, e = algebra.DecodeTypeJSON(f["type"], path+".type")
		if e != nil {
			return e
		}
		key := row.FeatureID + "\x00" + row.ValueID
		if seen[key] {
			return invalid("DUPLICATE_SOURCE_FEATURE_CONDITION", path, "feature and value mapping is duplicated")
		}
		seen[key] = true
		result = append(result, row)
	}
	*out = result
	return nil
}

func decodeSourceCondition(raw []byte, path string) (SourceCondition, error) {
	f, err := languageStrictObject(raw, path, []string{"kind", "code"}, []string{"kind", "code"})
	if err != nil {
		return SourceCondition{}, err
	}
	var out SourceCondition
	if err := languageStrictValue(f["kind"], &out.Kind); err != nil {
		return out, invalid("INVALID_SOURCE_CONDITION", path+".kind", "condition kind must be a non-null string")
	}
	if err := languageStrictValue(f["code"], &out.Code); err != nil {
		return out, invalid("INVALID_SOURCE_CONDITION", path+".code", "condition code must be a non-null string")
	}
	if err := validateSourceCondition(out, path); err != nil {
		return out, err
	}
	return out, nil
}

func validateSourceCondition(c SourceCondition, path string) error {
	if (c.Kind != SourceAssumption && c.Kind != SourcePresupposition) || !validSemanticPackageText(c.Code, MaxSourceCodeText) {
		return invalid("INVALID_SOURCE_CONDITION", path, "condition requires assumption/presupposition kind and bounded non-empty code")
	}
	return nil
}

func validateSourcePlanData(doc SourcePlanDocument) error {
	if len(doc.Predicates) > MaxSourcePredicateMaps || len(doc.FeatureConditions) > MaxSourceFeatureMaps || len(doc.Operators) > MaxSourceOperators {
		return resource("source_plan_items", maxIntLanguage(len(doc.Predicates), len(doc.FeatureConditions), len(doc.Operators)), MaxSourcePredicateMaps)
	}
	seenPred := map[string]bool{}
	for i, p := range doc.Predicates {
		key := p.SenseID.String() + "\x00" + string(p.FrameID)
		if seenPred[key] {
			return invalid("DUPLICATE_SOURCE_PREDICATE", fmt.Sprintf("predicate_mappings[%d]", i), "sense/frame selection is duplicated")
		}
		seenPred[key] = true
		if p.ReferenceCondition != nil {
			if err := validateSourceCondition(*p.ReferenceCondition, fmt.Sprintf("predicate_mappings[%d].reference_condition", i)); err != nil {
				return err
			}
		}
	}
	featureKeys := map[string]bool{}
	for i, c := range doc.FeatureConditions {
		if !validSemanticPackageText(c.FeatureID, 128) || !validSemanticPackageText(c.ValueID, 128) || !validSemanticPackageText(c.SymbolRef, 256) || !validSemanticPackageText(c.PortID, 256) {
			return invalid("INVALID_SOURCE_FEATURE_CONDITION", fmt.Sprintf("feature_conditions[%d]", i), "feature/value/symbol/port IDs must be bounded")
		}
		if c.Type.Kind != algebra.TypeEntity && c.Type.Kind != algebra.TypeNominal {
			return invalid("INVALID_SOURCE_FEATURE_TYPE", fmt.Sprintf("feature_conditions[%d].type", i), "feature conditions require Entity or Nominal input")
		}
		if c.Type.Kind == algebra.TypeEntity && (c.Type.SortID != "" || c.Type.From != nil || c.Type.To != nil) || c.Type.Kind == algebra.TypeNominal && (!validSemanticPackageText(c.Type.SortID, 256) || c.Type.From != nil || c.Type.To != nil) {
			return invalid("INVALID_SOURCE_FEATURE_TYPE", fmt.Sprintf("feature_conditions[%d].type", i), "feature type has malformed data-sort fields")
		}
		if c.Condition != SourceAssumption && c.Condition != SourcePresupposition || !validSemanticPackageText(c.Code, MaxSourceCodeText) {
			return invalid("INVALID_SOURCE_CONDITION", fmt.Sprintf("feature_conditions[%d]", i), "feature condition requires explicit non-asserted placement")
		}
		key := c.FeatureID + "\x00" + c.ValueID
		if featureKeys[key] {
			return invalid("DUPLICATE_SOURCE_FEATURE_CONDITION", fmt.Sprintf("feature_conditions[%d]", i), "feature and value mapping is duplicated")
		}
		featureKeys[key] = true
	}
	required := map[string]bool{}
	for i, id := range doc.RequiredFeatureIDs {
		if !validSemanticPackageText(id, 128) || required[id] {
			return invalid("INVALID_REQUIRED_SOURCE_FEATURE", fmt.Sprintf("required_feature_ids[%d]", i), "required feature IDs must be bounded and unique")
		}
		required[id] = true
		found := false
		for _, c := range doc.FeatureConditions {
			if c.FeatureID == id {
				found = true
				break
			}
		}
		if !found {
			return invalid("UNMAPPED_REQUIRED_SOURCE_FEATURE", fmt.Sprintf("required_feature_ids[%d]", i), "required feature must have an explicit condition mapping")
		}
	}
	pos := map[string]bool{}
	for i, value := range doc.ReferenceNonHeadPOSAllowlist {
		if !validSemanticPackageText(value, 64) || pos[value] {
			return invalid("INVALID_REFERENCE_POS", fmt.Sprintf("reference_non_head_pos_allowlist[%d]", i), "POS allowlist values must be bounded and unique")
		}
		pos[value] = true
	}
	return nil
}

func validateSourcePlanSelection(doc SourcePlanDocument, pack *SemanticLanguagePackage, selected algebra.Document) error {
	lexDoc := pack.lexicon.Document()
	senseByID := map[string]lexicon.LexicalSense{}
	for _, s := range lexDoc.Senses {
		senseByID[s.ID.String()] = s
	}
	frames := map[string]valency.ValencyFrame{}
	for _, f := range pack.Frames() {
		frames[string(f.ID)] = f
	}
	for i, m := range doc.Predicates {
		path := fmt.Sprintf("predicate_mappings[%d]", i)
		sense, ok := senseByID[m.SenseID.String()]
		if !ok {
			return unknown("UNKNOWN_SOURCE_SENSE", path+".sense_id", "sense is absent from selected package")
		}
		linked := false
		for _, fid := range sense.ValencyIDs {
			if string(fid) == string(m.FrameID) {
				linked = true
				break
			}
		}
		if !linked {
			return invalid("SOURCE_FRAME_NOT_LINKED", path+".frame_id", "frame must be linked from exact selected sense")
		}
		frame, ok := frames[string(m.FrameID)]
		if !ok || frame.SenseID != m.SenseID {
			return unknown("UNKNOWN_SOURCE_FRAME", path+".frame_id", "frame is absent from exact selected sense")
		}
		var chosen *lexicon.SemanticBinding
		for _, b := range pack.lexicon.TargetBindingsForSense(m.SenseID) {
			if b.Target == nil {
				continue
			}
			matches := false
			for _, rp := range b.Target.RolePorts {
				if rp.FrameID == string(m.FrameID) {
					matches = true
					break
				}
			}
			if matches {
				if chosen != nil {
					return languageError(status.OutcomeAmbiguous, "AMBIGUOUS_SOURCE_TARGET", path, "multiple selected targets map this exact frame")
				}
				copyB := b
				chosen = &copyB
			}
		}
		if chosen == nil {
			return unknown("MISSING_SOURCE_TARGET", path, "no explicit semantic target maps the selected frame")
		}
		if !chosen.TargetExecutionAuthorized() {
			return unsupported("SOURCE_TARGET_NOT_AUTHORIZED", path, "source plan requires an approved equivalent target with valid provenance")
		}
		if _, err := validateSemanticTargetEnvironment(selected, chosen.Target, frame); err != nil {
			return err
		}
		mapped := map[string]bool{}
		for _, rp := range chosen.Target.RolePorts {
			if rp.FrameID == string(m.FrameID) {
				mapped[rp.Role] = true
			}
		}
		if len(mapped) != len(frame.SemanticRoles) {
			return invalid("SOURCE_ROLE_MAP_INCOMPLETE", path, "selected target must preserve every semantic role in the source frame")
		}
		for _, role := range frame.SemanticRoles {
			if !mapped[string(role)] {
				return invalid("SOURCE_ROLE_MAP_INCOMPLETE", path, "selected target omits an authored semantic role")
			}
		}
		if !containsSourceFamily(frame.AllowedFamilyIDs, m.AllowedFamilyIDs) {
			return invalid("SOURCE_FAMILY_NOT_LINKED", path+".allowed_family_ids", "source plan families must be a subset of selected frame permissions")
		}
		if m.PredicateKind == "relation" || m.PredicateKind == "property" {
			frameRoleSet := map[string]bool{}
			for _, role := range frame.SemanticRoles {
				frameRoleSet[string(role)] = true
			}
			expectedSymbol := "concept:belonging_association"
			if m.PredicateKind == "property" {
				expectedSymbol = "concept:possession"
			}
			if m.HeadPortID != "" || chosen.Target.Kind != lexicon.SemanticTargetSymbol || (m.PredicateKind == "relation" && string(chosen.Target.SymbolRef) != "concept:belonging_association" && string(chosen.Target.SymbolRef) != "concept:legal_ownership") || (m.PredicateKind == "property" && string(chosen.Target.SymbolRef) != expectedSymbol) || len(frame.SemanticRoles) != 2 {
				return invalid("INVALID_SOURCE_RELATION_MAPPING", path, "non-event predicate requires the exact approved two-role symbol and no Event head")
			}
			rolePortIDs := map[string]string{}
			selectedPorts := 0
			for _, rolePort := range chosen.Target.RolePorts {
				if rolePort.FrameID != string(m.FrameID) {
					continue
				}
				selectedPorts++
				if !frameRoleSet[rolePort.Role] || rolePort.PortID != rolePort.Role || rolePortIDs[rolePort.Role] != "" {
					return invalid("INVALID_SOURCE_RELATION_MAPPING", path, "target role ports must map the exact source roles one-to-one to identically named ports")
				}
				rolePortIDs[rolePort.Role] = rolePort.PortID
			}
			if len(rolePortIDs) != 2 || selectedPorts != 2 {
				return invalid("INVALID_SOURCE_RELATION_MAPPING", path, "target must map both exact property/relation roles")
			}
			var signature *algebra.Signature
			for signatureIndex := range selected.Signatures {
				if selected.Signatures[signatureIndex].ID == string(chosen.Target.SymbolRef) {
					signature = &selected.Signatures[signatureIndex]
					break
				}
			}
			if signature == nil || len(signature.Ports) != 2 {
				return invalid("INVALID_SOURCE_RELATION_MAPPING", path, "property/relation symbol must declare exactly two target ports")
			}
			for _, port := range signature.Ports {
				if !frameRoleSet[port.Role] || port.Type.Kind != algebra.TypeEntity || rolePortIDs[port.Role] != port.Role {
					return invalid("INVALID_SOURCE_RELATION_MAPPING", path, "property/relation signature must declare only the exact two Entity roles")
				}
			}
		} else {
			if m.PredicateKind != "" && m.PredicateKind != "event" {
				return invalid("INVALID_SOURCE_PREDICATE_KIND", path+".predicate_kind", "predicate kind must be event, relation, or the exact validated possession property")
			}
			typ, roleMapped := sourceTargetPortType(chosen.Target, selected, m.HeadPortID, string(m.FrameID))
			if !typ.ok || roleMapped || typ.typ.Kind != algebra.TypeEvent {
				return invalid("INVALID_SOURCE_HEAD_PORT", path+".head_port_id", "head port must be an unmapped typed Event port on the selected target")
			}
		}
	}
	for i, c := range doc.FeatureConditions {
		path := fmt.Sprintf("feature_conditions[%d]", i)
		var sig *algebra.Signature
		for j := range selected.Signatures {
			if selected.Signatures[j].ID == c.SymbolRef {
				sig = &selected.Signatures[j]
				break
			}
		}
		if sig == nil {
			return unknown("UNKNOWN_FEATURE_SYMBOL", path+".symbol_ref", "feature predicate must be an approved selected primitive")
		}
		if len(sig.Ports) != 1 || sig.Ports[0].Role != c.PortID {
			return invalid("FEATURE_SYMBOL_NOT_UNARY", path+".symbol_ref", "feature condition must select exactly its declared unary port")
		}
		if !algebra.TypeEqual(sig.Ports[0].Type, c.Type) {
			return invalid("FEATURE_PORT_TYPE_MISMATCH", path+".type", "feature condition type must equal selected unary port type")
		}
	}
	return nil
}

type sourcePortType struct {
	typ algebra.Type
	ok  bool
}

func sourceTargetPortType(target *lexicon.SemanticTarget, selected algebra.Document, id, frameID string) (sourcePortType, bool) {
	if target.Kind == lexicon.SemanticTargetGoal && target.GoalSpec != nil {
		for _, p := range target.GoalSpec.Ports {
			if p.ID == id {
				return sourcePortType{p.Type, true}, targetPortMapped(target, id, frameID)
			}
		}
		return sourcePortType{}, false
	}
	for _, s := range selected.Signatures {
		if s.ID == string(target.SymbolRef) {
			for _, p := range s.Ports {
				if p.Role == id {
					return sourcePortType{p.Type, true}, targetPortMapped(target, id, frameID)
				}
			}
			break
		}
	}
	return sourcePortType{}, false
}
func targetPortMapped(target *lexicon.SemanticTarget, id, frameID string) bool {
	for _, p := range target.RolePorts {
		if p.FrameID == frameID && p.PortID == id {
			return true
		}
	}
	return false
}
func containsSourceFamily(frame []ids.ConstructionID, requested []ids.ConstructionID) bool {
	allowed := map[string]bool{}
	for _, id := range frame {
		allowed[id.String()] = true
	}
	for _, id := range requested {
		if !allowed[id.String()] {
			return false
		}
	}
	return len(requested) > 0
}

func validateSourceContextDeclarations(context SourceReferenceContext, selected algebra.Document) error {
	if selected.SchemaVersion != algebra.SchemaVersion || selected.Expression != nil || selected.RequestContext != nil || len(selected.Parameters) != 0 || len(selected.References) > algebra.DefaultMaxSymbols {
		return invalid("INVALID_SELECTED_DECLARATIONS", "selected", "selected context must be a bounded declaration-only algebra document")
	}
	if err := algebra.CheckDeclarations(selected); err != nil {
		return err
	}
	selectedSorts := make(map[string]bool, len(selected.Sorts))
	for _, declaration := range selected.Sorts {
		selectedSorts[declaration.ID] = true
	}
	usedSorts := map[string]bool{}
	active := map[*algebra.Type]bool{}
	for i := range context.References {
		if err := collectSourceTypeSorts(context.References[i].Type, 1, active, usedSorts); err != nil {
			return invalid("INVALID_SOURCE_REFERENCE_TYPE", fmt.Sprintf("references[%d].type", i), err.Error())
		}
	}
	if context.RequestContext != nil {
		for i := range context.RequestContext.Referents {
			if err := collectSourceTypeSorts(context.RequestContext.Referents[i].Type, 1, active, usedSorts); err != nil {
				return invalid("INVALID_SOURCE_CONTEXT_REFERENT_TYPE", fmt.Sprintf("request_context.referents[%d].type", i), err.Error())
			}
		}
	}
	for id := range usedSorts {
		if !selectedSorts[id] {
			return unknown("UNKNOWN_SOURCE_CONTEXT_SORT", "request_context", "nominal source reference type is absent from the selected algebra metadata")
		}
	}
	rows := append([]algebra.SortDeclaration{}, selected.Sorts...)
	refs, err := mergeReferences(selected.References, context.References)
	if err != nil {
		return err
	}
	doc := algebra.Document{SchemaVersion: algebra.SchemaVersion, Sorts: rows, Signatures: selected.Signatures, References: refs, Limits: selected.Limits, RequestContext: context.RequestContext}
	if err := algebra.CheckDeclarations(doc); err != nil {
		return err
	}
	if context.RequestContext != nil {
		if _, err := algebra.CheckRequestContext(*context.RequestContext, doc); err != nil {
			return err
		}
	}
	return nil
}
func collectSourceTypeSorts(t algebra.Type, depth int, active map[*algebra.Type]bool, sorts map[string]bool) error {
	if depth > algebra.DefaultMaxTypeDepth {
		return fmt.Errorf("type depth exceeds %d", algebra.DefaultMaxTypeDepth)
	}
	switch t.Kind {
	case algebra.TypeEntity, algebra.TypeEvent, algebra.TypeInterval, algebra.TypeProp, algebra.TypeQuestion:
		if t.SortID != "" || t.From != nil || t.To != nil {
			return fmt.Errorf("built-in type has irrelevant fields")
		}
	case algebra.TypeNominal:
		if !validSemanticPackageText(t.SortID, 256) || t.From != nil || t.To != nil {
			return fmt.Errorf("nominal type requires only a bounded sort_id")
		}
		sorts[t.SortID] = true
	case algebra.TypeArrow:
		if t.SortID != "" || t.From == nil || t.To == nil {
			return fmt.Errorf("arrow type requires from and to only")
		}
		for _, child := range []*algebra.Type{t.From, t.To} {
			if active[child] {
				return fmt.Errorf("cyclic type graph")
			}
			active[child] = true
			if err := collectSourceTypeSorts(*child, depth+1, active, sorts); err != nil {
				return err
			}
			delete(active, child)
		}
	default:
		return fmt.Errorf("unknown type kind")
	}
	return nil
}

func cloneSourcePlanDocument(src SourcePlanDocument) SourcePlanDocument {
	out := src
	out.ReferenceNonHeadPOSAllowlist = append([]string{}, src.ReferenceNonHeadPOSAllowlist...)
	out.Predicates = append([]SourcePredicateMapping{}, src.Predicates...)
	for i := range out.Predicates {
		out.Predicates[i].AllowedFamilyIDs = append([]ids.ConstructionID{}, src.Predicates[i].AllowedFamilyIDs...)
		out.Predicates[i].AllowedVariantIDs = append([]construction.VariantID{}, src.Predicates[i].AllowedVariantIDs...)
		if src.Predicates[i].ReferenceCondition != nil {
			c := *src.Predicates[i].ReferenceCondition
			out.Predicates[i].ReferenceCondition = &c
		}
	}
	out.Operators = append([]SourceOperatorMapping{}, src.Operators...)
	out.FeatureConditions = append([]SourceFeatureCondition{}, src.FeatureConditions...)
	for i := range out.FeatureConditions {
		out.FeatureConditions[i].Type = cloneLanguageType(src.FeatureConditions[i].Type)
	}
	out.RequiredFeatureIDs = append([]string{}, src.RequiredFeatureIDs...)
	return out
}

func cloneSourceReferenceContext(src SourceReferenceContext) SourceReferenceContext {
	out := src
	out.References = cloneLanguageDeclarations(src.References)
	out.RoleReferences = append([]SemanticRoleReference{}, src.RoleReferences...)
	out.PortReferences = append([]SemanticPortReference{}, src.PortReferences...)
	if src.RequestContext != nil {
		if c, err := algebra.CloneRequestContext(*src.RequestContext); err == nil {
			out.RequestContext = &c
		} else {
			out.RequestContext = nil
		}
	}
	return out
}
