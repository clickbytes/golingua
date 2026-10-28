package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/valency"
)

const (
	MaxRelationTargetAlternatives = 16
	MaxRelationPlanBytes          = 64 << 10
	MaxRelationPlanVariants       = 16
	MaxRelationPlanParts          = 128
)

type RelationPathSelection struct {
	Binding     lexicon.SemanticBinding
	Sense       lexicon.LexicalSense
	Lexeme      lexicon.MorphLexeme
	Frame       valency.ValencyFrame
	Validation  language.TargetValidation
	SubjectRole string
	ObjectRole  string
	PortRoles   map[string]string
	PortIDs     map[string]string
}

// RelationPathRequest retains the checked function document and exact tree
// coordinates for λx.λy.∃z.And(r(x,z),r(z,y)). Selected records are evidence;
// this request does not denote a fact or introduce an individual reference.
type RelationPathRequest struct {
	Checked          algebra.Document
	FirstLambdaPath  string
	SecondLambdaPath string
	WitnessPath      string
	ConjunctionPath  string
	FirstAtomPath    string
	SecondAtomPath   string
	BinderX          string
	BinderY          string
	BinderZ          string
	Symbol           string
	OrderedPorts     [2]string
	EventPair        *RelationPathEventPairEvidence
	Selection        RelationPathSelection
	Package          *language.SemanticLanguagePackage
	Selected         algebra.Document
	Budget           semantic.DSEBudget
}

// RelationPathEventPairEvidence preserves the two explicit event witnesses
// surrounding the ordered Entity arguments. TemporalFrame is copied from the
// checked expression and is never inferred by the target realizer.
type RelationPathEventPairEvidence struct {
	EventPort           string
	FirstEventBinder    string
	FirstEventPath      string
	FirstTemporalFrame  string
	SecondEventBinder   string
	SecondEventPath     string
	SecondTemporalFrame string
}

type RelationPathPlan struct {
	SchemaVersion int                   `json:"schema_version"`
	Language      string                `json:"language"`
	Provenance    provenance.SourceRef  `json:"provenance"`
	Variants      []RelationPathVariant `json:"variants"`
}

type RelationPathVariant struct {
	ID    string             `json:"id"`
	Parts []RelationPathPart `json:"parts"`
}

// Slot names are construction-plan variables, not binder identifiers. The
// request binds x/y/z to the checked expression's concrete IDs.
type RelationPathPart struct {
	Step                string `json:"step"`
	Literal             string `json:"literal,omitempty"`
	Slot                string `json:"slot,omitempty"`
	Port                string `json:"port,omitempty"`
	GrammaticalFunction string `json:"grammatical_function,omitempty"`
}

type RelationPathSurfacePart struct {
	Step                string               `json:"step"`
	Text                string               `json:"text"`
	BinderID            string               `json:"binder_id,omitempty"`
	Port                string               `json:"port,omitempty"`
	GrammaticalFunction string               `json:"grammatical_function,omitempty"`
	Generated           *morph.GeneratedForm `json:"generated,omitempty"`
}

type RelationPathAlternative struct {
	ID        string                    `json:"id"`
	VariantID string                    `json:"variant_id"`
	Parts     []RelationPathSurfacePart `json:"parts"`
}

func RelationPathAlternativeID(variantID string, parts []RelationPathSurfacePart) string {
	canonical, err := json.Marshal(struct {
		VariantID string                    `json:"variant_id"`
		Parts     []RelationPathSurfacePart `json:"parts"`
	}{VariantID: variantID, Parts: parts})
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(canonical)
	return "relation-path:" + hex.EncodeToString(digest[:])
}

type RelationPathResult struct {
	Status       status.OutcomeKind        `json:"status"`
	Alternatives []RelationPathAlternative `json:"alternatives,omitempty"`
	Diagnostics  []status.Diagnostic       `json:"diagnostics,omitempty"`
}

type RelationPathTargetRealizer interface {
	RealizeRelationPath(context.Context, RelationPathRequest) (RelationPathResult, error)
}

// RelationPathTargetVerifier must return its immutable, compiled language
// plan. Verify implementations should compare an alternative against this
// plan and independently realize the same request before accepting it.
type RelationPathTargetVerifier interface {
	RelationPathPlan() (RelationPathPlan, error)
	VerifyRelationPath(context.Context, RelationPathRequest, RelationPathAlternative) error
}

func DecodeRelationPathPlan(data []byte, languageID string) (RelationPathPlan, error) {
	var plan RelationPathPlan
	if len(data) == 0 || len(data) > MaxRelationPlanBytes || languageID == "" {
		return plan, errors.New("invalid relation path plan size or language")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return RelationPathPlan{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return RelationPathPlan{}, errors.New("relation path plan must contain one JSON value")
	}
	if plan.SchemaVersion != 1 || plan.Language != languageID || len(plan.Variants) == 0 || len(plan.Variants) > MaxRelationPlanVariants {
		return RelationPathPlan{}, errors.New("invalid relation path plan header")
	}
	if err := plan.Provenance.Validate(); err != nil {
		return RelationPathPlan{}, fmt.Errorf("invalid relation path plan provenance: %w", err)
	}
	seen := map[string]bool{}
	for i, variant := range plan.Variants {
		constructionID, err := ids.ParseConstructionID(variant.ID)
		if err != nil || constructionID.String() != variant.ID || seen[variant.ID] || len(variant.Parts) == 0 || len(variant.Parts) > MaxRelationPlanParts {
			return RelationPathPlan{}, fmt.Errorf("invalid relation path variant %d", i)
		}
		seen[variant.ID] = true
		for j, part := range variant.Parts {
			if part.Step != "first_atom" && part.Step != "second_atom" && part.Step != "connective" {
				return RelationPathPlan{}, fmt.Errorf("variant %d part %d has invalid step", i, j)
			}
			if part.Slot == "" && part.Literal == "" {
				return RelationPathPlan{}, fmt.Errorf("variant %d part %d is empty", i, j)
			}
			if part.Slot == "" && part.Port != "" {
				return RelationPathPlan{}, fmt.Errorf("variant %d literal cannot name a port", i)
			}
			if part.Slot != "" && part.Step == "connective" {
				return RelationPathPlan{}, fmt.Errorf("variant %d connective cannot contain a slot", i)
			}
			if part.Literal != "" && (strings.TrimSpace(part.Literal) == "" || strings.TrimSpace(part.Literal) != part.Literal) {
				return RelationPathPlan{}, fmt.Errorf("variant %d literal is empty or padded", i)
			}
			if (part.Slot == "x" || part.Slot == "y" || part.Slot == "z") && ((part.Port != "first" && part.Port != "second") || part.Literal == "" || (part.Step != "first_atom" && part.Step != "second_atom")) {
				return RelationPathPlan{}, fmt.Errorf("variant %d argument slot requires an atom step, port, and explicit literal", i)
			}
			if part.Slot == "x" || part.Slot == "y" || part.Slot == "z" {
				if part.GrammaticalFunction != "subject" && part.GrammaticalFunction != "object" {
					return RelationPathPlan{}, fmt.Errorf("variant %d argument slot requires subject/object grammatical function", i)
				}
			} else if part.GrammaticalFunction != "" {
				return RelationPathPlan{}, fmt.Errorf("variant %d grammatical function is allowed only on x/y/z slots", i)
			}
			if part.Slot == "verb" && (part.Port != "" || part.Literal != "") {
				return RelationPathPlan{}, fmt.Errorf("variant %d verb slot cannot name a port or literal", i)
			}
			if part.Slot == "verb" && part.GrammaticalFunction != "" {
				return RelationPathPlan{}, fmt.Errorf("variant %d verb slot cannot name a grammatical function", i)
			}
			if part.Slot == "" && part.GrammaticalFunction != "" {
				return RelationPathPlan{}, fmt.Errorf("variant %d literal cannot name a grammatical function", i)
			}
			if part.Slot != "" && part.Slot != "x" && part.Slot != "y" && part.Slot != "z" && part.Slot != "verb" {
				return RelationPathPlan{}, fmt.Errorf("variant %d has unknown slot", i)
			}
		}
		if err := validateRelationPlanCoverage(variant.Parts); err != nil {
			return RelationPathPlan{}, fmt.Errorf("variant %d: %w", i, err)
		}
	}
	return plan, nil
}

// ValidateRelationPathAlternative checks that a generated alternative is
// exactly licensed by one compiled plan variant and that all four argument
// occurrences in the two relation atoms have explicit surface coverage.
func ValidateRelationPathAlternative(request RelationPathRequest, plan RelationPathPlan, alternative RelationPathAlternative) error {
	if err := validateRelationPathRequest(request); err != nil {
		return err
	}
	if request.Package == nil {
		return errors.New("selected relation package is absent")
	}
	encodedPlan, err := json.Marshal(plan)
	if err != nil || len(encodedPlan) > MaxRelationPlanBytes {
		return errors.New("compiled relation plan cannot be encoded within its bound")
	}
	decodedPlan, err := DecodeRelationPathPlan(encodedPlan, string(request.Package.Language()))
	if err != nil {
		return fmt.Errorf("compiled relation plan failed strict decoding: %w", err)
	}
	plan = decodedPlan
	if len(alternative.Parts) == 0 || len(alternative.Parts) > MaxRelationPlanParts {
		return errors.New("relation alternative has invalid part count")
	}
	var variant *RelationPathVariant
	for i := range plan.Variants {
		if plan.Variants[i].ID == alternative.VariantID {
			if variant != nil {
				return errors.New("duplicate matching relation plan variant")
			}
			variant = &plan.Variants[i]
		}
	}
	if variant == nil || len(variant.Parts) != len(alternative.Parts) || alternative.ID == "" || alternative.ID != RelationPathAlternativeID(alternative.VariantID, alternative.Parts) {
		return errors.New("alternative is not an exact compiled plan variant")
	}
	covered := map[string]int{}
	verbs := map[string]int{}
	for i, want := range variant.Parts {
		got := alternative.Parts[i]
		if got.Step != want.Step || got.Text == "" {
			return fmt.Errorf("surface part %d has wrong step or empty text", i)
		}
		if want.Slot == "" {
			if got.Text != want.Literal || got.Generated != nil || got.BinderID != "" || got.Port != "" || got.GrammaticalFunction != "" {
				return fmt.Errorf("literal part %d differs from plan", i)
			}
			continue
		}
		switch want.Slot {
		case "x", "y", "z":
			binder := map[string]string{"x": request.BinderX, "y": request.BinderY, "z": request.BinderZ}[want.Slot]
			expectedPosition := relationExpectedPosition(got.Step, want.Slot)
			expectedPort := request.OrderedPorts[0]
			if expectedPosition == "second" {
				expectedPort = request.OrderedPorts[1]
			}
			if got.Text != want.Literal || got.BinderID != binder || want.Port != expectedPosition || got.Port != expectedPort || got.GrammaticalFunction != want.GrammaticalFunction || got.Generated != nil || !relationOccurrenceAllowed(got.Step, want.Slot, want.Port) || !relationFunctionMatches(request.Selection, expectedPort, want.GrammaticalFunction) {
				return fmt.Errorf("surface part %d has incorrect checked binder/port assignment", i)
			}
			covered[got.Step+"\x00"+got.BinderID+"\x00"+got.Port]++
		case "verb":
			if got.BinderID != "" || got.Port != "" || got.GrammaticalFunction != "" || got.Generated == nil || got.Generated.Surface != got.Text || got.Generated.LexemeID != request.Selection.Lexeme.ID.String() || string(got.Generated.Language) != string(request.Package.Language()) {
				return fmt.Errorf("verb part %d lacks the selected generated lexical form", i)
			}
			verbs[got.Step]++
		default:
			return fmt.Errorf("plan part %d has unknown slot", i)
		}
	}
	for _, requirement := range [][3]string{{"first_atom", request.BinderX, request.OrderedPorts[0]}, {"first_atom", request.BinderZ, request.OrderedPorts[1]}, {"second_atom", request.BinderZ, request.OrderedPorts[0]}, {"second_atom", request.BinderY, request.OrderedPorts[1]}} {
		if covered[requirement[0]+"\x00"+requirement[1]+"\x00"+requirement[2]] != 1 {
			return errors.New("alternative does not surface every relation argument occurrence")
		}
	}
	if verbs["first_atom"] != 1 || verbs["second_atom"] != 1 {
		return errors.New("alternative must contain exactly one generated selected verb for each relation atom")
	}
	return nil
}

func ValidateRelationPathResult(request RelationPathRequest, plan RelationPathPlan, result RelationPathResult) error {
	if !result.Status.Valid() {
		return errors.New("relation target returned an invalid status")
	}
	if len(result.Alternatives) > MaxRelationTargetAlternatives {
		return errors.New("relation target exceeded the alternative limit")
	}
	if len(result.Diagnostics) > status.MaxDiagnostics {
		return errors.New("relation target exceeded the diagnostic limit")
	}
	if result.Status == status.OutcomeOK && len(result.Alternatives) == 0 {
		return errors.New("relation target cannot report ok without a covered alternative")
	}
	if result.Status == status.OutcomeOK && len(result.Alternatives) > 1 {
		return errors.New("multiple relation alternatives cannot have ok status")
	}
	if len(result.Alternatives) > 1 && result.Status != status.OutcomeAmbiguous && result.Status != status.OutcomeIncomplete {
		return errors.New("multiple relation alternatives require ambiguous or incomplete status")
	}
	if result.Status != status.OutcomeOK && len(result.Diagnostics) == 0 {
		return errors.New("non-ok relation target result requires diagnostics")
	}
	seen := map[string]bool{}
	for _, alternative := range result.Alternatives {
		if seen[alternative.ID] {
			return errors.New("relation target returned duplicate alternatives")
		}
		seen[alternative.ID] = true
		if err := ValidateRelationPathAlternative(request, plan, alternative); err != nil {
			return err
		}
	}
	return nil
}

func relationOccurrenceAllowed(step, slot, position string) bool {
	if step == "first_atom" {
		return (slot == "x" && position == "first") || (slot == "z" && position == "second")
	}
	if step == "second_atom" {
		return (slot == "z" && position == "first") || (slot == "y" && position == "second")
	}
	return false
}

func relationExpectedPosition(step, slot string) string {
	if (step == "first_atom" && slot == "x") || (step == "second_atom" && slot == "z") {
		return "first"
	}
	if (step == "first_atom" && slot == "z") || (step == "second_atom" && slot == "y") {
		return "second"
	}
	return ""
}

func relationFunctionMatches(selection RelationPathSelection, port, grammaticalFunction string) bool {
	role, ok := selection.PortRoles[port]
	if !ok {
		return false
	}
	switch grammaticalFunction {
	case "subject":
		return role == selection.SubjectRole
	case "object":
		return role == selection.ObjectRole
	default:
		return false
	}
}

func validateRelationPlanCoverage(parts []RelationPathPart) error {
	want := map[string]bool{}
	counts := map[string]int{}
	verbCounts := map[string]int{}
	functionCounts := map[string]int{}
	portCounts := map[string]int{}
	for _, part := range parts {
		if part.Slot == "verb" {
			verbCounts[part.Step]++
			continue
		}
		if part.Slot == "" {
			continue
		}
		key := part.Step + "\x00" + part.Slot
		counts[key]++
		want[key] = true
		functionCounts[part.Step+"\x00"+part.GrammaticalFunction]++
		portCounts[part.Step+"\x00"+part.Port]++
	}
	for _, key := range []string{"first_atom\x00x", "first_atom\x00z", "second_atom\x00z", "second_atom\x00y"} {
		if !want[key] || counts[key] != 1 {
			return errors.New("plan must map each of the four relation argument occurrences exactly once")
		}
	}
	if verbCounts["first_atom"] != 1 || verbCounts["second_atom"] != 1 {
		return errors.New("plan must contain one generated verb slot per relation atom")
	}
	for _, step := range []string{"first_atom", "second_atom"} {
		if functionCounts[step+"\x00subject"] != 1 || functionCounts[step+"\x00object"] != 1 || portCounts[step+"\x00first"] != 1 || portCounts[step+"\x00second"] != 1 {
			return errors.New("each atom requires exactly one subject and object across first and second ports")
		}
	}
	return nil
}

func validateRelationPathRequest(request RelationPathRequest) error {
	if request.Checked.Expression == nil || request.BinderX == "" || request.BinderY == "" || request.BinderZ == "" || request.BinderX == request.BinderY || request.BinderX == request.BinderZ || request.BinderY == request.BinderZ {
		return errors.New("relation path request has missing or duplicate binders")
	}
	if _, err := algebra.Check(request.Checked); err != nil {
		return fmt.Errorf("checked relation expression is invalid: %w", err)
	}
	root := request.Checked.Expression
	signatures := make(map[string]algebra.Signature, len(request.Checked.Signatures))
	for _, signature := range request.Checked.Signatures {
		signatures[signature.ID] = signature
	}
	match, binaryOK := matchAlgebraRelationPath(root, signatures)
	var selected algebraRelationBindingResult
	if binaryOK {
		if request.EventPair != nil {
			return errors.New("binary relation expression cannot carry event-pair evidence")
		}
		if match.FirstLambda.ID != request.BinderX || match.SecondLambda.ID != request.BinderY || match.WitnessBinder != request.BinderZ || match.Symbol != request.Symbol || match.FirstPorts != request.OrderedPorts {
			return errors.New("request binders, symbol, or ordered ports do not match the checked relation path")
		}
		if match.FirstPath != request.FirstLambdaPath || match.SecondPath != request.SecondLambdaPath || match.WitnessPath != request.WitnessPath || match.ConjunctionPath != request.ConjunctionPath || match.FirstAtomPath != request.FirstAtomPath || match.SecondAtomPath != request.SecondAtomPath {
			return errors.New("request tree paths do not match the checked expression")
		}
		selected = selectAlgebraRelationBinding(match, request.Package, request.Selected, request.Budget)
	} else {
		eventMatch, eventOK := matchAlgebraEventRelationPath(root, signatures)
		if !eventOK || request.EventPair == nil {
			return errors.New("checked expression must match exactly one supported binary or event relation path and carry matching evidence")
		}
		evidence := request.EventPair
		if eventMatch.FirstLambda.ID != request.BinderX || eventMatch.SecondLambda.ID != request.BinderY || eventMatch.WitnessBinder != request.BinderZ || eventMatch.Symbol != request.Symbol || eventMatch.FirstPorts != request.OrderedPorts {
			return errors.New("request binders, symbol, or ordered ports do not match the checked event relation path")
		}
		if eventMatch.FirstPath != request.FirstLambdaPath || eventMatch.SecondPath != request.SecondLambdaPath || eventMatch.WitnessPath != request.WitnessPath || eventMatch.ConjunctionPath != request.ConjunctionPath || eventMatch.FirstAtomPath != request.FirstAtomPath || eventMatch.SecondAtomPath != request.SecondAtomPath {
			return errors.New("request tree paths do not match the checked event expression")
		}
		if evidence.EventPort != eventMatch.EventPort || evidence.FirstEventBinder != eventMatch.FirstEventBinder || evidence.FirstEventPath != eventMatch.FirstEventPath || evidence.FirstTemporalFrame != eventMatch.FirstEventTemporalFrame || evidence.SecondEventBinder != eventMatch.SecondEventBinder || evidence.SecondEventPath != eventMatch.SecondEventPath || evidence.SecondTemporalFrame != eventMatch.SecondEventTemporalFrame {
			return errors.New("explicit event-pair evidence differs from the checked event binders, paths, port, or temporal frames")
		}
		selected = selectAlgebraEventRelationBinding(eventMatch, request.Package, request.Selected, request.Budget)
	}
	if selected.Status != status.OutcomeOK {
		return fmt.Errorf("selected relation metadata status is %s: %s", selected.Status, selected.Diagnostic.Message)
	}
	want := request.Selection
	got := selected.Binding
	if !reflect.DeepEqual(want.Binding, got.SemanticBinding) || !reflect.DeepEqual(want.Sense, got.LexicalSense) || !reflect.DeepEqual(want.Lexeme, got.Lexeme) || !reflect.DeepEqual(want.Validation, got.Validation) || !reflect.DeepEqual(want.Frame, got.Frame) || want.SubjectRole != got.SubjectRole || want.ObjectRole != got.ObjectRole || !reflect.DeepEqual(want.PortRoles, got.PortRoles) || !reflect.DeepEqual(want.PortIDs, got.PortIDs) {
		return errors.New("explicit relation lexical selection differs from the unique selected-package binding")
	}
	return nil
}
