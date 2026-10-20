package morph

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/procedure"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	MaxPrincipalParts     = 64
	MaxPrincipalPartBytes = 1_024
	MaxPrincipalPartRunes = 512
)

type PrincipalPart struct {
	ID   PrincipalPartID `json:"id"`
	Form string          `json:"form"`
}

// PrincipalParts are explicitly supplied diagnostic forms for one lexeme.
type PrincipalParts struct {
	LexemeID ids.LexemeID    `json:"lexeme_id"`
	Parts    []PrincipalPart `json:"parts"`
}

func (parts PrincipalParts) MarshalJSON() ([]byte, error) {
	if !validLexemeID(parts.LexemeID) {
		return nil, fmt.Errorf("principal parts require a typed lexeme ID")
	}
	partValues := parts.Parts
	if partValues == nil {
		partValues = []PrincipalPart{}
	}
	return json.Marshal(struct {
		LexemeID string          `json:"lexeme_id"`
		Parts    []PrincipalPart `json:"parts"`
	}{LexemeID: parts.LexemeID.String(), Parts: partValues})
}

func (parts *PrincipalParts) UnmarshalJSON(data []byte) error {
	type wirePrincipalParts struct {
		LexemeID string          `json:"lexeme_id"`
		Parts    []PrincipalPart `json:"parts"`
	}
	var decoded wirePrincipalParts
	if err := decodeMorphObject(data, map[string]bool{"lexeme_id": true, "parts": true}, []string{"lexeme_id", "parts"}, []string{"parts"}, &decoded); err != nil {
		return err
	}
	lexemeID, err := ids.ParseLexemeID(decoded.LexemeID)
	if err != nil {
		return fmt.Errorf("invalid principal-parts lexeme_id: %w", err)
	}
	*parts = PrincipalParts{LexemeID: lexemeID, Parts: decoded.Parts}
	return nil
}

type MorphError struct{ result status.Result }

func (failure *MorphError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "morphology operation failed"
	}
	return failure.result.Diagnostics[0].Message
}

func (failure *MorphError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), failure.result.Diagnostics...)
	return result
}

func morphError(kind status.OutcomeKind, code, path, message string) error {
	diagnostic := status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError}
	return &MorphError{result: status.Result{Status: kind, Diagnostics: []status.Diagnostic{diagnostic}}}
}

// ResolveStem returns a requested diagnostic form only when the caller
// supplied it explicitly. It never derives an irregular stem from a lemma.
func ResolveStem(parts PrincipalParts, requirement PrincipalPartRequirement, budget procedure.Budget) (string, error) {
	if err := budget.Validate(); err != nil {
		return "", morphError(status.OutcomeInvalid, "INVALID_PROCEDURE_BUDGET", "budget", err.Error())
	}
	if !validPrincipalPartID(requirement.PartID) || !validStemID(requirement.StemID) {
		return "", morphError(status.OutcomeInvalid, "INVALID_PRINCIPAL_PART_REQUIREMENT", "requirement", "principal-part and stem IDs must be bounded identifiers")
	}
	if !validLexemeID(parts.LexemeID) {
		return "", morphError(status.OutcomeInvalid, "INVALID_LEXEME_ID", "principal_parts.lexeme_id", "principal parts require a typed lexeme ID")
	}
	if len(parts.Parts) > MaxPrincipalParts {
		return "", morphError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.principal_parts", "principal-part count exceeds its limit")
	}
	found := false
	form := ""
	seen := make(map[PrincipalPartID]struct{}, len(parts.Parts))
	for index, part := range parts.Parts {
		path := fmt.Sprintf("principal_parts.parts[%d]", index)
		if !validPrincipalPartID(part.ID) {
			return "", morphError(status.OutcomeInvalid, "INVALID_PRINCIPAL_PART_ID", path+".id", "principal-part ID is malformed")
		}
		if _, exists := seen[part.ID]; exists {
			return "", morphError(status.OutcomeInvalid, "DUPLICATE_PRINCIPAL_PART_ID", path+".id", "principal-part ID is duplicated")
		}
		seen[part.ID] = struct{}{}
		if !validPrincipalForm(part.Form, budget.MaxInputCodepoints) {
			return "", morphError(status.OutcomeInvalid, "INVALID_PRINCIPAL_PART_FORM", path+".form", "principal-part form is empty, invalid UTF-8, contains controls or exceeds its bound")
		}
		if len(part.Form) > MaxPrincipalPartBytes || utf8.RuneCountInString(part.Form) > MaxPrincipalPartRunes {
			return "", morphError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", path+".form", "principal-part form exceeds its length limit")
		}
		if part.ID == requirement.PartID {
			found, form = true, part.Form
		}
	}
	if !found {
		return "", morphError(status.OutcomeUnknown, "MISSING_PRINCIPAL_PART", "principal_parts."+string(requirement.PartID), "required diagnostic form was not supplied")
	}
	return form, nil
}

// ResolveStems resolves only supplied parts required by a paradigm. Optional
// missing parts are omitted; required missing parts return unknown.
func ResolveStems(parts PrincipalParts, requirements []PrincipalPartRequirement, budget procedure.Budget) (map[procedure.StemID]string, error) {
	if err := budget.Validate(); err != nil {
		return nil, morphError(status.OutcomeInvalid, "INVALID_PROCEDURE_BUDGET", "budget", err.Error())
	}
	if !validLexemeID(parts.LexemeID) {
		return nil, morphError(status.OutcomeInvalid, "INVALID_LEXEME_ID", "principal_parts.lexeme_id", "principal parts require a typed lexeme ID")
	}
	if len(parts.Parts) > MaxPrincipalParts {
		return nil, morphError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.principal_parts", "principal-part count exceeds its limit")
	}
	if len(requirements) > MaxPrincipalParts {
		return nil, morphError(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "budget.principal_part_requirements", "principal-part requirement count exceeds its limit")
	}
	result := make(map[procedure.StemID]string, len(requirements))
	seenStems := make(map[procedure.StemID]struct{}, len(requirements))
	seenParts := make(map[PrincipalPartID]struct{}, len(requirements))
	for _, requirement := range requirements {
		if !validPrincipalPartID(requirement.PartID) || !validStemID(requirement.StemID) {
			return nil, morphError(status.OutcomeInvalid, "INVALID_PRINCIPAL_PART_REQUIREMENT", "requirements", "stem and diagnostic part IDs must be bounded identifiers")
		}
		if _, exists := seenStems[requirement.StemID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_STEM_REQUIREMENT", "requirements", "stem requirement is duplicated")
		}
		if _, exists := seenParts[requirement.PartID]; exists {
			return nil, morphError(status.OutcomeInvalid, "DUPLICATE_PRINCIPAL_PART_REQUIREMENT", "requirements", "diagnostic part is mapped more than once")
		}
		seenStems[requirement.StemID] = struct{}{}
		seenParts[requirement.PartID] = struct{}{}
		form, err := ResolveStem(parts, requirement, budget)
		if err == nil {
			result[requirement.StemID] = form
			continue
		}
		failure, ok := err.(*MorphError)
		if ok && failure.Result().Status == status.OutcomeUnknown && !requirement.Required {
			continue
		}
		return nil, err
	}
	return result, nil
}

func validPrincipalForm(value string, maximumRunes int) bool {
	if len(value) == 0 || len(value) > MaxPrincipalPartBytes || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maximumRunes || strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validPrincipalPartID(id PrincipalPartID) bool {
	return validMorphToken(string(id))
}

func validStemID(id procedure.StemID) bool {
	return validMorphToken(string(id))
}

func validLexemeID(id ids.LexemeID) bool {
	_, err := ids.ParseLexemeID(id.String())
	return err == nil
}

func validMorphToken(value string) bool {
	if len(value) == 0 || len(value) > 64 || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') && character != '_' && character != '-' && character != '.' {
			return false
		}
	}
	return true
}
