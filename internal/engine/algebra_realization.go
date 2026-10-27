package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/clickbytes/golingua/internal/algebra"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/language"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	maxAlgebraRealizationParts     = 128
	maxAlgebraRealizationBindings  = 4096
	maxAlgebraRealizationRefInputs = algebra.DefaultMaxPorts
)

// AlgebraRoleSubstitution records how one selected algebra port was realized.
// ReferenceID is the caller's stable typed-IR reference; MeaningNodeID is the
// request-local node passed to the existing target pipeline.
type AlgebraRoleSubstitution struct {
	PortID        string         `json:"port_id"`
	Role          string         `json:"role"`
	ReferenceID   string         `json:"reference_id"`
	MeaningNodeID meaning.NodeID `json:"meaning_node_id"`
}

// AlgebraRealizationCoverage describes the exact data binding used for one
// matched atom. It makes no claim about unmatched parts of the input.
type AlgebraRealizationCoverage struct {
	Path              string                    `json:"path"`
	Symbol            string                    `json:"symbol"`
	SenseID           string                    `json:"sense_id"`
	CoreRef           string                    `json:"core_ref"`
	EventPortID       string                    `json:"event_port_id"`
	EventReferenceID  string                    `json:"event_reference_id"`
	EventMeaningNode  meaning.NodeID            `json:"event_meaning_node_id"`
	RoleSubstitutions []AlgebraRoleSubstitution `json:"role_substitutions"`
}

// AlgebraRealizationResidual retains input meaning that this bounded route did
// not realize. The source object is included when the residual originated in
// the typed expression or request context.
type AlgebraRealizationResidual struct {
	Path               string                      `json:"path"`
	Code               string                      `json:"code"`
	Message            string                      `json:"message"`
	Term               *algebra.Term               `json:"term,omitempty"`
	ContextReferent    *algebra.ContextReferent    `json:"context_referent,omitempty"`
	ContextMention     *algebra.ContextMention     `json:"context_mention,omitempty"`
	ResolutionEvidence *algebra.ResolutionEvidence `json:"resolution_evidence,omitempty"`
	ContextCondition   *algebra.ContextCondition   `json:"context_condition,omitempty"`
	ReferenceID        string                      `json:"reference_id,omitempty"`
}

// AlgebraRealizationResult reports coverage, surface alternatives and all
// retained residuals for one independent typed-IR realization request.
type AlgebraRealizationResult struct {
	Status          status.OutcomeKind           `json:"status"`
	Language        string                       `json:"language"`
	LanguagePackage string                       `json:"language_package,omitempty"`
	PackageVersion  string                       `json:"package_version,omitempty"`
	Coverage        []AlgebraRealizationCoverage `json:"coverage"`
	Residuals       []AlgebraRealizationResidual `json:"residuals"`
	Alternatives    []SurfaceAlternative         `json:"alternatives"`
	Diagnostics     []status.Diagnostic          `json:"diagnostics"`
	Scope           string                       `json:"scope"`
}

type algebraAtomMatch struct {
	path       string
	atom       *algebra.Term
	signature  algebra.Signature
	binding    lexicon.SemanticBinding
	eventPort  string
	rolePorts  []lexicon.SemanticRolePort
	arguments  map[string]*algebra.Term
	validation language.TargetValidation
}

// RealizeAlgebra selects one approved, exact symbol binding from an explicitly
// loaded language package, maps its typed roles and event identity to a
// request-local Meaning, and delegates surface generation to the configured
// target adapter. It supports a direct Atom, or one direct Atom within a root
// And; every other expression part is returned as a residual.
func (engine *Engine) RealizeAlgebra(
	ctx context.Context,
	languageID features.LanguageID,
	input algebra.Document,
	languagePackage *language.SemanticLanguagePackage,
	selectedMetadata algebra.Document,
) (AlgebraRealizationResult, error) {
	result := AlgebraRealizationResult{
		Status: status.OutcomeInvalid, Language: string(languageID),
		Coverage: []AlgebraRealizationCoverage{}, Residuals: []AlgebraRealizationResidual{},
		Alternatives: []SurfaceAlternative{}, Diagnostics: []status.Diagnostic{},
		Scope: "exact approved symbol target; one positive Atom; language adapter output is not a truth claim",
	}
	if ctx == nil {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "MISSING_REQUEST_CONTEXT", "context", "algebra realization requires a caller context"), nil
	}
	if err := ctx.Err(); err != nil {
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "REALIZATION_CANCELED", "context", "algebra realization was canceled or timed out"), nil
	}
	if !validLanguage(string(languageID)) {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_LANGUAGE", "language", "a canonical target language is required"), nil
	}
	if engine == nil || engine.core == nil || engine.snapshot == nil || !engine.snapshotMatchesCore {
		return algebraRealizationFailure(result, status.OutcomeUnsupported, "RUNTIME_SNAPSHOT_UNAVAILABLE", "runtime", "algebra realization requires one validated CORE and pack snapshot"), nil
	}
	if languagePackage == nil || languagePackage.LexiconIndex() == nil {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "MISSING_LANGUAGE_PACKAGE", "language_package", "an explicitly loaded semantic language package is required"), nil
	}
	if string(languagePackage.Language()) != string(languageID) {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "LANGUAGE_PACKAGE_MISMATCH", "language_package.language", "selected language package does not match the requested target language"), nil
	}
	if selectedMetadata.SchemaVersion != algebra.SchemaVersion || selectedMetadata.Expression != nil || selectedMetadata.RequestContext != nil {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SELECTED_METADATA", "metadata", "selected metadata must be a declaration-only algebra context"), nil
	}
	result.LanguagePackage = languagePackage.ID()
	result.PackageVersion = languagePackage.Version()

	merged, err := realizationCheckDocument(input, selectedMetadata)
	if err != nil {
		return algebraRealizationError(result, err, "INVALID_ALGEBRA_DOCUMENT", "input")
	}
	if checked, checkErr := algebra.Check(merged); checkErr != nil {
		return algebraRealizationError(result, checkErr, "INVALID_ALGEBRA_DOCUMENT", "input")
	} else if checked.Status != status.OutcomeOK {
		return algebraRealizationFailure(result, checked.Status, "INCOMPLETE_ALGEBRA_DOCUMENT", "input.expression", "typed expression contains unresolved holes or requirements"), nil
	}

	validations, err := languagePackage.ValidateTargets(selectedMetadata)
	if err != nil {
		return algebraRealizationError(result, err, "LANGUAGE_TARGET_VALIDATION_FAILED", "language_package")
	}
	if len(languagePackage.LexiconIndex().Document().Bindings) > maxAlgebraRealizationBindings {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "BINDING_SCAN_LIMIT", Message: "expression was retained because the language package exceeds the bounded target scan", Term: cloneAlgebraTerm(input.Expression)})
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "language_package.bindings", "semantic language package exceeds the realization binding limit"), nil
	}
	validatedSignatures := signaturesByID(selectedMetadata.Signatures)
	if input.Expression == nil {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "MISSING_EXPRESSION", "input.expression", "an expression is required for realization"), nil
	}
	parts, partPaths := realizationParts(input.Expression)
	if len(parts) > maxAlgebraRealizationParts {
		result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: "input.expression", Code: "EXPRESSION_PART_LIMIT", Message: "entire expression was retained because its top-level part count exceeds the realization limit", Term: cloneAlgebraTerm(input.Expression)})
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "input.expression", "expression has too many top-level parts for realization"), nil
	}
	matches := make([]algebraAtomMatch, 0, 1)
	unresolved := make([]AlgebraRealizationResidual, 0)
	refusalStatus := status.OutcomeUnsupported
	for index, part := range parts {
		if part == nil || part.Kind != algebra.TermAtom {
			unresolved = append(unresolved, AlgebraRealizationResidual{Path: partPaths[index], Code: "UNSUPPORTED_EXPRESSION_PART", Message: "only a directly matched positive Atom is realized by this route", Term: cloneAlgebraTerm(part)})
			continue
		}
		match, matchStatus, matchCode, matchMessage := findAlgebraAtomMatch(part, partPaths[index], languagePackage, validations, validatedSignatures)
		if match != nil {
			matches = append(matches, *match)
			continue
		}
		if matchStatus == status.OutcomeAmbiguous || matchStatus == status.OutcomeInvalid || matchStatus == status.OutcomeResourceLimit {
			refusalStatus = matchStatus
		} else if matchStatus == status.OutcomeUnknown && refusalStatus == status.OutcomeUnsupported && len(unresolved) == 0 {
			refusalStatus = status.OutcomeUnknown
		}
		unresolved = append(unresolved, AlgebraRealizationResidual{Path: partPaths[index], Code: matchCode, Message: matchMessage, Term: cloneAlgebraTerm(part)})
		if matchStatus == status.OutcomeInvalid || matchStatus == status.OutcomeResourceLimit {
			result.Residuals = append(result.Residuals, unresolved...)
			return algebraRealizationFailure(result, matchStatus, matchCode, partPaths[index], matchMessage), nil
		}
	}
	if len(matches) == 0 {
		result.Residuals = append(result.Residuals, unresolved...)
		return algebraRealizationFailure(result, refusalStatus, "NO_EXACT_REALIZATION_BINDING", "input.expression", "no exact approved symbol target covers the expression"), nil
	}
	if len(matches) != 1 {
		for _, match := range matches {
			unresolved = append(unresolved, AlgebraRealizationResidual{Path: match.path, Code: "MULTIPLE_SUPPORTED_FRAGMENTS", Message: "more than one independently realizable Atom requires composition that this route does not provide", Term: cloneAlgebraTerm(match.atom)})
		}
		result.Residuals = append(result.Residuals, unresolved...)
		return algebraRealizationFailure(result, status.OutcomeAmbiguous, "MULTIPLE_REALIZABLE_ATOMS", "input.expression", "multiple exact target atoms are present; no fragment was selected"), nil
	}
	match := matches[0]
	result.Residuals = append(result.Residuals, unresolved...)
	if len(result.Residuals) > maxAlgebraRealizationParts {
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "residuals", "expression residual count exceeds the realization limit"), nil
	}

	candidate, coverage, contextResiduals, failureStatus, failureCode, failurePath, failureMessage := buildAlgebraMeaningCandidate(engine, input, languageID, languagePackage, match)
	result.Residuals = append(result.Residuals, contextResiduals...)
	if candidate == nil {
		failureAlreadyReported := false
		for _, residual := range contextResiduals {
			if residual.Code == failureCode && residual.Path == failurePath {
				failureAlreadyReported = true
				break
			}
		}
		if !failureAlreadyReported {
			result.Residuals = append(result.Residuals, AlgebraRealizationResidual{Path: match.path, Code: failureCode, Message: failureMessage, Term: cloneAlgebraTerm(match.atom)})
		}
		if trimAlgebraResiduals(&result) {
			return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESIDUAL_REPORT_LIMIT", "residuals", "request contains more residual items than this result can report"), nil
		}
		return algebraRealizationFailure(result, failureStatus, failureCode, failurePath, failureMessage), nil
	}
	if trimAlgebraResiduals(&result) {
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESIDUAL_REPORT_LIMIT", "residuals", "request contains more residual items than this result can report"), nil
	}
	result.Coverage = append(result.Coverage, coverage)

	target, ok := engine.TargetRealizerFor(languageID)
	if !ok || target == nil {
		return algebraRealizationFailure(result, status.OutcomeUnsupported, "TARGET_RUNTIME_UNAVAILABLE", "language", "no verified target adapter is available for the requested language"), nil
	}
	targetLanguage, registered := engine.languages.Language(languageID)
	if !registered {
		return algebraRealizationFailure(result, status.OutcomeUnsupported, "LANGUAGE_NOT_REGISTERED", "language", "target language is absent from the verified runtime registry"), nil
	}
	budget := engineBudget(engine)
	requestCtx, cancel := context.WithTimeout(ctx, budget.RequestTimeout)
	defer cancel()
	surface, realizeErr := target.Realize(requestCtx, *candidate)
	if realizeErr != nil {
		return algebraRealizationError(result, realizeErr, "TARGET_REALIZATION_FAILED", "target")
	}
	if !surface.Outcome.Valid() || len(surface.Diagnostics) > status.MaxDiagnostics {
		return algebraRealizationFailure(result, status.OutcomeError, "INVALID_TARGET_OUTCOME", "target.outcome", "target adapter returned an unsupported realization status"), nil
	}
	for _, diagnostic := range surface.Diagnostics {
		if err := diagnostic.Validate(); err != nil {
			return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_TARGET_DIAGNOSTIC", "target.diagnostics", "target adapter returned an invalid diagnostic"), nil
		}
	}
	if len(surface.Alternatives) > budget.MaxSurfaceAlternatives || len(surface.Alternatives) > MaxTranslationAlternatives {
		return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "target.alternatives", "target adapter exceeded the surface alternative limit"), nil
	}
	if len(surface.Alternatives) == 0 && surface.Outcome == status.OutcomeOK {
		return algebraRealizationFailure(result, status.OutcomeIncomplete, "NO_SURFACE_ALTERNATIVE", "target.alternatives", "target adapter returned OK without a surface alternative"), nil
	}
	seenAlternatives := make(map[string]struct{}, len(surface.Alternatives))
	textBytes := 0
	usableSurface := false
	completeSurface := false
	selectedCoreRef, parseCoreErr := ids.ParseConceptID(result.Coverage[0].CoreRef)
	if parseCoreErr != nil {
		return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SELECTED_CORE_REF", "coverage.core_ref", "selected target coverage contains a malformed CORE concept ID"), nil
	}
	for index, alternative := range surface.Alternatives {
		path := fmt.Sprintf("target.alternatives[%d]", index)
		if !alternative.Outcome.Valid() || !validID(alternative.ID) || len(alternative.Diagnostics) > status.MaxDiagnostics || len(alternative.Assumptions) > MaxTranslationAssumptions || len(alternative.Losses) > MaxTranslationLosses {
			return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SURFACE_ALTERNATIVE", path, "target adapter returned malformed alternative metadata"), nil
		}
		if _, duplicate := seenAlternatives[alternative.ID]; duplicate {
			return algebraRealizationFailure(result, status.OutcomeInvalid, "DUPLICATE_SURFACE_ALTERNATIVE", path+".id", "target adapter returned a duplicate alternative ID"), nil
		}
		seenAlternatives[alternative.ID] = struct{}{}
		for _, diagnostic := range alternative.Diagnostics {
			if err := diagnostic.Validate(); err != nil {
				return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SURFACE_DIAGNOSTIC", path+".diagnostics", "target adapter returned an invalid alternative diagnostic"), nil
			}
		}
		for _, assumption := range alternative.Assumptions {
			if err := validateAssumption(assumption); err != nil {
				return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SURFACE_ASSUMPTION", path+".assumptions", "target adapter returned an invalid assumption"), nil
			}
		}
		for _, loss := range alternative.Losses {
			if err := validateLoss(loss); err != nil || alternative.Outcome == status.OutcomeOK && loss.Outcome != status.OutcomeOK {
				return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SURFACE_LOSS", path+".losses", "target adapter returned an invalid loss"), nil
			}
		}
		if err := validateText(alternative.Text, MaxTranslationTextBytes, true); err != nil {
			return algebraRealizationFailure(result, status.OutcomeResourceLimit, "INVALID_SURFACE_TEXT", path+".text", "target surface text is invalid UTF-8 or exceeds its byte limit"), nil
		}
		if alternative.Text != "" {
			textBytes += len(alternative.Text)
			if textBytes > budget.MaxCandidateTextBytes {
				return algebraRealizationFailure(result, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "target.text", "surface alternatives exceed the aggregate text byte limit"), nil
			}
			if !hasExactGeneratedAlgebraBindingEvidence(targetLanguage, alternative.Evidence, match.binding.SenseID, selectedCoreRef, candidate.Meaning) {
				return algebraRealizationFailure(result, status.OutcomeUnsupported, "TARGET_BINDING_EVIDENCE_UNAVAILABLE", path+".evidence", "text was rejected because no exact verified generated binding evidence supports the selected sense and CORE concept"), nil
			}
			if alternative.Outcome == status.OutcomeOK || alternative.Outcome == status.OutcomeIncomplete {
				usableSurface = true
			}
			if alternative.Outcome == status.OutcomeOK {
				completeSurface = true
			}
		} else if alternative.Outcome == status.OutcomeOK {
			return algebraRealizationFailure(result, status.OutcomeIncomplete, "EMPTY_COMPLETE_SURFACE", path+".text", "target adapter marked an empty surface as complete"), nil
		}
		if err := alternative.Evidence.Validate(); err != nil || alternative.Evidence.AlternativeID != alternative.ID || alternative.Evidence.Outcome != alternative.Outcome {
			return algebraRealizationFailure(result, status.OutcomeInvalid, "INVALID_SURFACE_EVIDENCE", path+".evidence", "target evidence does not validate or identify the exact alternative"), nil
		}
		if alternative.Outcome != status.OutcomeOK && len(alternative.Diagnostics) == 0 {
			return algebraRealizationFailure(result, status.OutcomeInvalid, "MISSING_SURFACE_DIAGNOSTIC", path+".diagnostics", "non-ok target alternative requires a diagnostic"), nil
		}
	}
	result.Alternatives = append([]SurfaceAlternative{}, surface.Alternatives...)
	result.Diagnostics = appendDiagnostics(result.Diagnostics, surface.Diagnostics)
	result.Status = surface.Outcome
	if completeSurface && (result.Status == status.OutcomeUnknown || result.Status == status.OutcomeUnsupported) {
		result.Status = status.OutcomeIncomplete
		result.Diagnostics = appendDiagnostic(result.Diagnostics, status.Diagnostic{Code: "PARTIAL_TARGET_REALIZATION", Path: "target.alternatives", Message: "a verified complete target alternative coexists with an unknown or unsupported aggregate target outcome", Severity: status.SeverityWarning})
	}
	if len(result.Residuals) != 0 && usableSurface && (result.Status == status.OutcomeOK || result.Status == status.OutcomeUnknown || result.Status == status.OutcomeUnsupported || result.Status == status.OutcomeIncomplete) {
		result.Status = status.OutcomeIncomplete
		result.Diagnostics = appendDiagnostic(result.Diagnostics, status.Diagnostic{Code: "REALIZATION_RESIDUALS_REMAIN", Path: "residuals", Message: "surface output covers only the reported expression fragments; residual meaning remains", Severity: status.SeverityWarning})
	}
	if len(result.Alternatives) == 0 && result.Status == status.OutcomeOK {
		return algebraRealizationFailure(result, status.OutcomeIncomplete, "NO_SURFACE_ALTERNATIVE", "target.alternatives", "target adapter returned OK without a surface alternative"), nil
	}
	return result, nil
}

func realizationCheckDocument(input, selectedMetadata algebra.Document) (algebra.Document, error) {
	if len(input.Sorts) != 0 || len(input.Signatures) != 0 {
		return algebra.Document{}, &algebra.Error{Status: status.OutcomeInvalid, Code: "INLINE_DECLARATIONS_WITH_METADATA", Path: "input", Message: "inline sorts and signatures are forbidden when selected metadata is supplied"}
	}
	if input.SchemaVersion != algebra.SchemaVersion {
		return algebra.Document{}, &algebra.Error{Status: status.OutcomeInvalid, Code: "INVALID_SCHEMA_VERSION", Path: "input.schema_version", Message: "unsupported algebra document version"}
	}
	check := selectedMetadata
	check.Parameters = append([]algebra.Declaration{}, input.Parameters...)
	check.References = append([]algebra.Declaration{}, input.References...)
	check.Closed = input.Closed
	check.Expression = input.Expression
	check.Limits = input.Limits
	check.RequestContext = input.RequestContext
	return check, nil
}

func signaturesByID(signatures []algebra.Signature) map[string]algebra.Signature {
	byID := make(map[string]algebra.Signature, len(signatures))
	for _, signature := range signatures {
		byID[signature.ID] = signature
	}
	return byID
}

func realizationParts(expression *algebra.Term) ([]*algebra.Term, []string) {
	if expression == nil {
		return []*algebra.Term{nil}, []string{"input.expression"}
	}
	if expression.Kind == algebra.TermAnd {
		parts := append([]*algebra.Term{}, expression.Operands...)
		paths := make([]string, len(parts))
		for index := range parts {
			paths[index] = fmt.Sprintf("input.expression.operands[%d]", index)
		}
		return parts, paths
	}
	return []*algebra.Term{expression}, []string{"input.expression"}
}

func findAlgebraAtomMatch(
	atom *algebra.Term,
	path string,
	languagePackage *language.SemanticLanguagePackage,
	validations []language.TargetValidation,
	signatures map[string]algebra.Signature,
) (*algebraAtomMatch, status.OutcomeKind, string, string) {
	var signature algebra.Signature
	var hasSignature bool
	for _, candidate := range signatures {
		if candidate.ID == atom.Symbol {
			signature, hasSignature = candidate, true
			break
		}
	}
	if !hasSignature {
		return nil, status.OutcomeUnknown, "UNDECLARED_ALGEBRA_SYMBOL", "expression symbol has no declaration in the selected metadata"
	}
	bindings := languagePackage.LexiconIndex().Document().Bindings
	matchedSymbol := false
	valid := make([]lexicon.SemanticBinding, 0, 1)
	for _, binding := range bindings {
		if binding.Target == nil || binding.Target.Kind != lexicon.SemanticTargetSymbol || string(binding.Target.SymbolRef) != atom.Symbol {
			continue
		}
		matchedSymbol = true
		if !binding.TargetExecutionAuthorized() || !targetValidationAuthorized(validations, binding) {
			continue
		}
		valid = append(valid, binding)
	}
	if !matchedSymbol {
		return nil, status.OutcomeUnsupported, "NO_DIRECT_SYMBOL_TARGET", "this route requires a direct symbol target; GoalSpec and compound matching are reported as residuals"
	}
	if len(valid) == 0 {
		return nil, status.OutcomeUnsupported, "UNAPPROVED_OR_UNRESOLVED_BINDING", "matching language target is not an approved, metadata-validated equivalent binding"
	}
	if len(valid) > 1 {
		return nil, status.OutcomeAmbiguous, "AMBIGUOUS_LANGUAGE_BINDING", "multiple approved language bindings name this exact symbol"
	}
	binding := valid[0]
	if binding.Mapping != lexicon.MappingEquivalent || binding.Status != lexicon.BindingApproved || binding.Target == nil {
		return nil, status.OutcomeUnsupported, "UNAPPROVED_OR_UNRESOLVED_BINDING", "matching language target is not an approved equivalent mapping"
	}
	if binding.Target.Kind != lexicon.SemanticTargetSymbol || string(binding.Target.SymbolRef) != atom.Symbol {
		return nil, status.OutcomeUnsupported, "TARGET_KIND_UNSUPPORTED", "only a direct symbol target can match this route"
	}
	if len(binding.Target.RolePorts) == 0 || len(binding.Target.RolePorts) >= len(signature.Ports) {
		return nil, status.OutcomeUnsupported, "TARGET_PORT_SHAPE_UNSUPPORTED", "target must map participant ports and leave exactly one typed event port"
	}
	portMappings := make(map[string]lexicon.SemanticRolePort, len(binding.Target.RolePorts))
	roleNames := make(map[string]bool, len(binding.Target.RolePorts))
	for _, rolePort := range binding.Target.RolePorts {
		if rolePort.PortID == "" || rolePort.Role == "" || portMappings[rolePort.PortID].PortID != "" || roleNames[rolePort.Role] {
			return nil, status.OutcomeUnsupported, "INVALID_TARGET_ROLE_PORTS", "target role/port mapping is empty or duplicated"
		}
		portMappings[rolePort.PortID] = rolePort
		roleNames[rolePort.Role] = true
	}
	args := make(map[string]*algebra.Term, len(atom.Args))
	for _, argument := range atom.Args {
		if _, duplicate := args[argument.Role]; duplicate {
			return nil, status.OutcomeInvalid, "DUPLICATE_ATOM_ARGUMENT", "atom argument role is duplicated"
		}
		args[argument.Role] = argument.Value
	}
	eventPort := ""
	for _, port := range signature.Ports {
		rolePort, mapped := portMappings[port.Role]
		argument, hasArgument := args[port.Role]
		if !hasArgument || argument == nil {
			return nil, status.OutcomeInvalid, "MISSING_ATOM_ARGUMENT", "atom argument does not cover every declared port"
		}
		if mapped {
			if !algebra.TypeEqual(port.Type, rolePort.Type) || port.Type.Kind != algebra.TypeEntity {
				return nil, status.OutcomeUnsupported, "TARGET_PORT_TYPE_MISMATCH", "mapped role port does not exactly match an Entity-valued metadata port"
			}
			continue
		}
		if port.Type.Kind != algebra.TypeEvent || eventPort != "" {
			return nil, status.OutcomeUnsupported, "UNMAPPED_TARGET_PORT", "every non-role port must be the unique Event identity port"
		}
		eventPort = port.Role
	}
	if eventPort == "" || len(args) != len(signature.Ports) {
		return nil, status.OutcomeUnsupported, "TARGET_PORT_SHAPE_UNSUPPORTED", "target does not cover the signature with participant roles and one Event identity"
	}
	validation, found := findTargetValidation(validations, binding)
	if !found || !validation.Authorized || len(validation.Unresolved) != 0 {
		return nil, status.OutcomeUnsupported, "UNRESOLVED_TARGET_METADATA", "matching target ports did not validate against the selected metadata"
	}
	return &algebraAtomMatch{path: path, atom: atom, signature: signature, binding: binding, eventPort: eventPort, rolePorts: append([]lexicon.SemanticRolePort{}, binding.Target.RolePorts...), arguments: args, validation: validation}, status.OutcomeOK, "", ""
}

func targetValidationAuthorized(validations []language.TargetValidation, binding lexicon.SemanticBinding) bool {
	validation, found := findTargetValidation(validations, binding)
	return found && validation.Authorized && validation.Status == lexicon.BindingApproved && validation.Mapping == lexicon.MappingEquivalent && len(validation.Unresolved) == 0
}

func findTargetValidation(validations []language.TargetValidation, binding lexicon.SemanticBinding) (language.TargetValidation, bool) {
	var found language.TargetValidation
	count := 0
	for _, validation := range validations {
		if validation.SenseID == binding.SenseID.String() && validation.Kind == binding.Target.Kind {
			found = validation
			count++
		}
	}
	return found, count == 1
}

func buildAlgebraMeaningCandidate(
	engine *Engine,
	input algebra.Document,
	languageID features.LanguageID,
	languagePackage *language.SemanticLanguagePackage,
	match algebraAtomMatch,
) (*semantic.MeaningCandidate, AlgebraRealizationCoverage, []AlgebraRealizationResidual, status.OutcomeKind, string, string, string) {
	emptyCoverage := AlgebraRealizationCoverage{}
	noResiduals := []AlgebraRealizationResidual{}
	snapshotLexicon := engine.snapshot.Lexicon()
	if snapshotLexicon == nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeUnsupported, "RUNTIME_LEXICON_UNAVAILABLE", "runtime.lexicon", "validated runtime snapshot has no lexical index"
	}
	sense, hasSense := languagePackage.LexiconIndex().BySenseID(match.binding.SenseID)
	if !hasSense {
		return nil, emptyCoverage, noResiduals, status.OutcomeUnknown, "UNKNOWN_LANGUAGE_SENSE", "language_package.bindings", "selected target sense is absent from its language package"
	}
	lexeme, hasLexeme := languagePackage.LexiconIndex().ByLexemeID(sense.LexemeID)
	if !hasLexeme || string(lexeme.Language) != string(languageID) {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "LANGUAGE_SENSE_MISMATCH", "language_package.lexicon", "selected sense does not belong to the requested target language"
	}
	coreBindings := snapshotLexicon.TransferBindings(match.binding.SenseID)
	if len(coreBindings) == 0 {
		return nil, emptyCoverage, noResiduals, status.OutcomeUnknown, "SENSE_CORE_BINDING_MISSING", "runtime.lexicon.bindings", "selected exact sense has no approved equivalent CORE binding in the runtime snapshot"
	}
	if len(coreBindings) != 1 || !coreBindings[0].TransferAuthorized() {
		return nil, emptyCoverage, noResiduals, status.OutcomeAmbiguous, "AMBIGUOUS_SENSE_CORE_BINDING", "runtime.lexicon.bindings", "selected exact sense does not resolve to one unique approved equivalent CORE concept"
	}
	coreRef := coreBindings[0].CoreRef
	if _, exists := engine.core.Lookup(coreRef); !exists {
		return nil, emptyCoverage, noResiduals, status.OutcomeUnknown, "UNKNOWN_CORE_CONCEPT", "runtime.lexicon.bindings.core_ref", "selected sense CORE concept is absent from the active snapshot"
	}
	if input.RequestContext == nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeIncomplete, "MISSING_REFERENCE_CONTEXT", "input.request_context", "each algebra reference requires an explicit typed request context"
	}
	if match.atom == nil || len(match.rolePorts) == 0 {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "INVALID_ATOM_MATCH", match.path, "matched atom or its role mappings are missing"
	}
	if len(match.rolePorts) > maxAlgebraRealizationRefInputs {
		return nil, emptyCoverage, noResiduals, status.OutcomeResourceLimit, "RESOURCE_LIMIT", match.path + ".role_ports", "target has too many participant roles for realization"
	}
	args := match.arguments
	eventArg := args[match.eventPort]
	if eventArg == nil || eventArg.Kind != algebra.TermRef {
		return nil, emptyCoverage, noResiduals, status.OutcomeUnsupported, "EVENT_IDENTITY_NOT_REFERENCE", match.path + ".args." + match.eventPort, "the unique Event port must be bound to an explicit reference"
	}
	declarations := make(map[string]algebra.Declaration, len(input.References))
	for _, declaration := range input.References {
		declarations[declaration.ID] = declaration
	}
	contextReferents := make(map[string]algebra.ContextReferent, len(input.RequestContext.Referents))
	for _, referent := range input.RequestContext.Referents {
		contextReferents[referent.ID] = referent
	}
	eventReferenceID := eventArg.ID
	eventDeclaration, declared := declarations[eventReferenceID]
	eventContext, hasEventContext := contextReferents[eventReferenceID]
	if !declared || eventDeclaration.Type.Kind != algebra.TypeEvent || !hasEventContext || eventContext.Type.Kind != algebra.TypeEvent {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "INVALID_EVENT_REFERENCE", match.path + ".args." + match.eventPort, "Event identity must be declared and context-typed as Event"
	}
	if eventContext.ConceptID != "" && eventContext.ConceptID != coreRef.String() {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "EVENT_CONCEPT_CONFLICT", "input.request_context.referents." + eventReferenceID + ".concept_id", "event context classification conflicts with the selected exact lexical binding"
	}
	if eventDeclaration.Multiplicity != "" || eventContext.Multiplicity != "" {
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.referents." + eventReferenceID + ".multiplicity", Code: "UNSUPPORTED_EVENT_MULTIPLICITY", Message: "event multiplicity is not represented by this positive sentence route", ContextReferent: ptrContextReferent(eventContext), ReferenceID: eventReferenceID})
	}
	if eventDeclaration.ReferenceTimeID != "" || eventContext.ReferenceTimeID != "" {
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.referents." + eventReferenceID + ".reference_time_id", Code: "UNSUPPORTED_EVENT_REFERENCE_TIME", Message: "event reference-time link is retained but not realized by this route", ContextReferent: ptrContextReferent(eventContext), ReferenceID: eventReferenceID})
	}
	eventFrame := eventDeclaration.TemporalFrame
	if eventContext.TemporalFrame != "" {
		if eventFrame != "" && eventFrame != eventContext.TemporalFrame {
			return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "EVENT_TIME_CONFLICT", "input.request_context.referents." + eventReferenceID + ".temporal_frame", "event declaration and request context temporal frames differ"
		}
		eventFrame = eventContext.TemporalFrame
	}
	eventNodeID := meaning.NodeID("dse_event_0")
	participantNodes := make(map[string]meaning.NodeID)
	participantConcepts := make(map[string]ids.ConceptID)
	participantMultiplicity := make(map[string]meaning.Multiplicity)
	roleSubstitutions := make([]AlgebraRoleSubstitution, 0, len(match.rolePorts)+1)
	rolePorts := append([]lexicon.SemanticRolePort{}, match.rolePorts...)
	sort.Slice(rolePorts, func(i, j int) bool { return rolePorts[i].PortID < rolePorts[j].PortID })
	for _, rolePort := range rolePorts {
		argument := args[rolePort.PortID]
		if argument == nil || argument.Kind != algebra.TermRef {
			return nil, emptyCoverage, noResiduals, status.OutcomeUnsupported, "ROLE_ARGUMENT_NOT_REFERENCE", match.path + ".args." + rolePort.PortID, "every mapped Entity role must be bound to an explicit reference"
		}
		referenceID := argument.ID
		declaration, ok := declarations[referenceID]
		contextReferent, contextOK := contextReferents[referenceID]
		if !ok || declaration.Type.Kind != algebra.TypeEntity || !contextOK || contextReferent.Type.Kind != algebra.TypeEntity {
			return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "INVALID_ENTITY_REFERENCE", match.path + ".args." + rolePort.PortID, "participant reference and request context must both be typed Entity"
		}
		if contextReferent.ConceptID == "" {
			return nil, emptyCoverage, append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.referents." + referenceID + ".concept_id", Code: "MISSING_REFERENT_CONCEPT", Message: "participant noun cannot be selected without an explicit entity concept", ContextReferent: ptrContextReferent(contextReferent), ReferenceID: referenceID}), status.OutcomeIncomplete, "MISSING_REFERENT_CONCEPT", "input.request_context.referents." + referenceID + ".concept_id", "participant noun cannot be selected without an explicit entity concept"
		}
		conceptID, parseErr := ids.ParseConceptID(contextReferent.ConceptID)
		if parseErr != nil {
			return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "INVALID_REFERENT_CONCEPT", "input.request_context.referents." + referenceID + ".concept_id", "referent concept ID is malformed"
		}
		if existing, exists := participantConcepts[referenceID]; exists && existing != conceptID {
			return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "SHARED_REFERENT_CONCEPT_CONFLICT", "input.request_context.referents." + referenceID + ".concept_id", "one shared reference cannot have conflicting concept classifications"
		}
		participantConcepts[referenceID] = conceptID
		multiplicity := declaration.Multiplicity
		if contextReferent.Multiplicity != "" {
			if multiplicity != "" && multiplicity != contextReferent.Multiplicity {
				return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "REFERENT_MULTIPLICITY_CONFLICT", "input.request_context.referents." + referenceID + ".multiplicity", "reference declaration and context multiplicities differ"
			}
			multiplicity = contextReferent.Multiplicity
		}
		if multiplicity != "" {
			value := meaning.Multiplicity(multiplicity)
			if !value.Valid() {
				return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "INVALID_REFERENT_MULTIPLICITY", "input.request_context.referents." + referenceID + ".multiplicity", "referent multiplicity is unsupported"
			}
			participantMultiplicity[referenceID] = value
		}
		if declaration.TemporalFrame != "" || contextReferent.TemporalFrame != "" {
			noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.referents." + referenceID + ".temporal_frame", Code: "UNSUPPORTED_ENTITY_TEMPORAL_FRAME", Message: "entity temporal-frame metadata is retained but not realized by this route", ContextReferent: ptrContextReferent(contextReferent), ReferenceID: referenceID})
		}
		if declaration.ReferenceTimeID != "" || contextReferent.ReferenceTimeID != "" {
			noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.referents." + referenceID + ".reference_time_id", Code: "UNSUPPORTED_ENTITY_REFERENCE_TIME", Message: "entity reference-time metadata is retained but not realized by this route", ContextReferent: ptrContextReferent(contextReferent), ReferenceID: referenceID})
		}
		if _, exists := participantNodes[referenceID]; !exists {
			participantNodes[referenceID] = meaning.NodeID(fmt.Sprintf("dse_ref_%d", len(participantNodes)))
		}
		roleSubstitutions = append(roleSubstitutions, AlgebraRoleSubstitution{PortID: rolePort.PortID, Role: rolePort.Role, ReferenceID: referenceID, MeaningNodeID: participantNodes[referenceID]})
	}
	consumedContextReferences := map[string]struct{}{eventReferenceID: {}}
	for referenceID := range participantNodes {
		consumedContextReferences[referenceID] = struct{}{}
	}
	for index, referent := range input.RequestContext.Referents {
		if _, consumed := consumedContextReferences[referent.ID]; consumed {
			continue
		}
		noResiduals = append(noResiduals, AlgebraRealizationResidual{
			Path: fmt.Sprintf("input.request_context.referents[%d]", index), Code: "UNCONSUMED_CONTEXT_REFERENT",
			Message:         "context referent is retained but is neither the event identity nor a participant of the selected atom",
			ContextReferent: ptrContextReferent(referent), ReferenceID: referent.ID,
		})
	}
	if input.RequestContext.SpeakerReferentID != "" {
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.speaker_referent_id", Code: "UNSUPPORTED_SPEAKER_CONTEXT", Message: "speaker anchoring is retained but not used to choose a surface form in this route", ReferenceID: input.RequestContext.SpeakerReferentID})
	}
	if input.RequestContext.AddresseeReferentID != "" {
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: "input.request_context.addressee_referent_id", Code: "UNSUPPORTED_ADDRESSEE_CONTEXT", Message: "addressee anchoring is retained but not used to choose a surface form in this route", ReferenceID: input.RequestContext.AddresseeReferentID})
	}
	for index := range input.RequestContext.Mentions {
		mention := input.RequestContext.Mentions[index]
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.mentions[%d]", index), Code: "UNSUPPORTED_CONTEXT_MENTION", Message: "mention resolution is retained but not consumed by the direct symbol route", ContextMention: &mention})
	}
	for index := range input.RequestContext.ResolutionEvidence {
		evidence := input.RequestContext.ResolutionEvidence[index]
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.resolution_evidence[%d]", index), Code: "UNSUPPORTED_RESOLUTION_EVIDENCE", Message: "identity evidence is retained but not consumed by the direct symbol route", ResolutionEvidence: &evidence})
	}
	for index := range input.RequestContext.Conditions {
		condition := cloneAlgebraCondition(input.RequestContext.Conditions[index])
		noResiduals = append(noResiduals, AlgebraRealizationResidual{Path: fmt.Sprintf("input.request_context.conditions[%d]", index), Code: "UNSUPPORTED_CONTEXT_CONDITION", Message: "assumption, presupposition, or loss condition is retained as an explicit residual", ContextCondition: &condition})
	}
	if len(noResiduals) > maxAlgebraRealizationParts {
		return nil, emptyCoverage, noResiduals, status.OutcomeResourceLimit, "RESOURCE_LIMIT", "input.request_context", "request context residual count exceeds the realization limit"
	}

	participants := make([]meaning.Participant, 0, len(roleSubstitutions))
	for _, substitution := range roleSubstitutions {
		participants = append(participants, meaning.Participant{Role: meaning.Role(substitution.Role), Target: substitution.MeaningNodeID})
	}
	event, err := meaning.NewEvent(eventNodeID, coreRef, participants)
	if err != nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "MEANING_BUILD_FAILED", "meaning.event", err.Error()
	}
	event.TemporalFrame = meaning.TemporalFrame(eventFrame)
	if eventFrame != "" && !event.TemporalFrame.Valid() {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "INVALID_EVENT_TEMPORAL_FRAME", "input.request_context.referents." + eventReferenceID + ".temporal_frame", "event temporal frame is unsupported"
	}
	nodes := make([]meaning.Node, 0, len(participantNodes)+2)
	nodes = append(nodes, event)
	referenceIDs := make([]string, 0, len(participantNodes))
	for referenceID := range participantNodes {
		referenceIDs = append(referenceIDs, referenceID)
	}
	sort.Strings(referenceIDs)
	meaningContext := meaning.RequestContext{ID: realizationContextID(input.RequestContext.ID), Referents: []meaning.Referent{}, Mentions: []meaning.Mention{}}
	for _, referenceID := range referenceIDs {
		referent, err := meaning.NewReferent(participantNodes[referenceID], participantConcepts[referenceID])
		if err != nil {
			return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "MEANING_BUILD_FAILED", "meaning.referents", err.Error()
		}
		referent.Multiplicity = participantMultiplicity[referenceID]
		nodes = append(nodes, referent)
		contextReferentID := meaning.ReferentID(fmt.Sprintf("ref_%d", len(meaningContext.Referents)))
		meaningContext.Referents = append(meaningContext.Referents, meaning.Referent{ID: contextReferentID, Concept: meaning.ConceptRefFromID(participantConcepts[referenceID])})
		// This bridge preserves typed algebra Ref identity only. Its explicit
		// evidence ID does not claim a source mention, surface span, or observed
		// linguistic form.
		meaningContext.Mentions = append(meaningContext.Mentions, meaning.Mention{
			ID:     meaning.MentionID(fmt.Sprintf("ir_ref_%d", len(meaningContext.Mentions))),
			NodeID: participantNodes[referenceID], ReferentID: contextReferentID,
			EvidenceID: algebraReferenceEvidenceID(referenceID),
		})
	}
	discourseID := meaning.NodeID("dse_discourse_0")
	discourse, err := meaning.NewDiscourse(discourseID, meaning.ActStatement, []meaning.NodeID{eventNodeID})
	if err != nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "MEANING_BUILD_FAILED", "meaning.discourse", err.Error()
	}
	nodes = append(nodes, discourse)
	graph, err := meaning.Build([]meaning.NodeID{discourseID}, nodes, meaning.DefaultBudget())
	if err != nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "MEANING_BUILD_FAILED", "meaning", err.Error()
	}
	if err := meaning.ValidateMeaning(graph, engine.core, meaning.DefaultBudget()); err != nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "MEANING_CORE_VALIDATION_FAILED", "meaning", err.Error()
	}
	if err := meaning.ValidateReferences(graph, meaningContext, meaning.DefaultBudget()); err != nil {
		return nil, emptyCoverage, noResiduals, status.OutcomeInvalid, "MEANING_CONTEXT_VALIDATION_FAILED", "meaning.context", err.Error()
	}
	candidateID := algebraCandidateID(input, languagePackage.ID(), match.binding.SenseID.String())
	candidate := &semantic.MeaningCandidate{
		ID: candidateID, SemanticID: candidateID, Meaning: &graph,
		Contexts: []semantic.MeaningContext{{Meaning: graph, Context: meaningContext}}, Status: status.OutcomeOK,
		Evidence: semantic.EvidenceTrace{AlternativeID: candidateID, Outcome: status.OutcomeOK, Records: []semantic.EvidenceRecord{}},
	}
	roleSubstitutions = append(roleSubstitutions, AlgebraRoleSubstitution{PortID: match.eventPort, Role: "event_identity", ReferenceID: eventReferenceID, MeaningNodeID: eventNodeID})
	coverage := AlgebraRealizationCoverage{
		Path: match.path, Symbol: match.atom.Symbol, SenseID: match.binding.SenseID.String(), CoreRef: coreRef.String(),
		EventPortID: match.eventPort, EventReferenceID: eventReferenceID, EventMeaningNode: eventNodeID,
		RoleSubstitutions: roleSubstitutions,
	}
	return candidate, coverage, noResiduals, status.OutcomeOK, "", "", ""
}

func algebraCandidateID(input algebra.Document, packageID, senseID string) string {
	encoded, err := json.Marshal(input)
	if err != nil {
		encoded = []byte(fmt.Sprintf("%s:%s", packageID, senseID))
	}
	hash := sha256.Sum256(append(append(encoded, 0), []byte(packageID+":"+senseID)...))
	return "dse-algebra-" + hex.EncodeToString(hash[:8])
}

func realizationContextID(contextID string) string {
	hash := sha256.Sum256([]byte(contextID))
	return "dse_ctx_" + hex.EncodeToString(hash[:8])
}

func algebraReferenceEvidenceID(referenceID string) string {
	hash := sha256.Sum256([]byte(referenceID))
	return "algebra_ref_" + hex.EncodeToString(hash[:8])
}

func hasExactGeneratedAlgebraBindingEvidence(pack language.Language, trace semantic.EvidenceTrace, senseID ids.SenseID, coreRef ids.ConceptID, graph *meaning.Meaning) bool {
	if graph == nil || (trace.Outcome != status.OutcomeOK && trace.Outcome != status.OutcomeIncomplete) || trace.Validate() != nil {
		return false
	}
	coreReferenced := false
	for _, node := range graph.Nodes {
		if (node.Kind == meaning.KindEvent || node.Kind == meaning.KindState || node.Kind == meaning.KindReferent) &&
			node.Concept == meaning.ConceptRefFromID(coreRef) {
			coreReferenced = true
			break
		}
	}
	if !coreReferenced {
		return false
	}
	for _, record := range trace.Records {
		if record.Origin != semantic.EvidenceOriginGenerated || record.Outcome != status.OutcomeOK || record.SenseID != senseID ||
			record.CoreConceptID != coreRef || record.BindingKey != semantic.BindingKeyFor(senseID, coreRef) ||
			record.FrameID == "" || record.FamilyID.String() == "" || record.VariantID.String() == "" {
			continue
		}
		for _, binding := range pack.Bindings {
			if binding.SenseID == senseID && binding.CoreRef == coreRef && binding.ValencyFrameID == string(record.FrameID) &&
				binding.FamilyID == record.FamilyID && binding.Mapping == "equivalent" && binding.Approved && binding.Realize &&
				containsVariantID(binding.VariantIDs, string(record.VariantID)) {
				return true
			}
		}
	}
	return false
}

func ptrContextReferent(value algebra.ContextReferent) *algebra.ContextReferent {
	copyValue := value
	return &copyValue
}

func cloneAlgebraCondition(value algebra.ContextCondition) algebra.ContextCondition {
	copyValue := value
	copyValue.TargetReferentIDs = append([]string{}, value.TargetReferentIDs...)
	copyValue.Expression = cloneAlgebraTerm(value.Expression)
	return copyValue
}

func cloneAlgebraTerm(term *algebra.Term) *algebra.Term {
	if term == nil {
		return nil
	}
	copyValue := *term
	if term.Type != nil {
		typeCopy := cloneAlgebraType(*term.Type)
		copyValue.Type = &typeCopy
	}
	copyValue.Args = make([]algebra.Argument, len(term.Args))
	for index, argument := range term.Args {
		copyValue.Args[index] = algebra.Argument{Role: argument.Role, Value: cloneAlgebraTerm(argument.Value)}
	}
	copyValue.Operands = make([]*algebra.Term, len(term.Operands))
	for index, operand := range term.Operands {
		copyValue.Operands[index] = cloneAlgebraTerm(operand)
	}
	copyValue.Body = cloneAlgebraTerm(term.Body)
	copyValue.Function = cloneAlgebraTerm(term.Function)
	copyValue.Argument = cloneAlgebraTerm(term.Argument)
	copyValue.Restrictor = cloneAlgebraTerm(term.Restrictor)
	copyValue.Answers = append([]algebra.Declaration{}, term.Answers...)
	if term.Count != nil {
		count := *term.Count
		copyValue.Count = &count
	}
	return &copyValue
}

func cloneAlgebraType(value algebra.Type) algebra.Type {
	copyValue := value
	if value.From != nil {
		from := cloneAlgebraType(*value.From)
		copyValue.From = &from
	}
	if value.To != nil {
		to := cloneAlgebraType(*value.To)
		copyValue.To = &to
	}
	return copyValue
}

func trimAlgebraResiduals(result *AlgebraRealizationResult) bool {
	if result == nil || len(result.Residuals) <= maxAlgebraRealizationParts {
		return false
	}
	omitted := len(result.Residuals) - maxAlgebraRealizationParts + 1
	result.Residuals = append([]AlgebraRealizationResidual{}, result.Residuals[:maxAlgebraRealizationParts-1]...)
	result.Residuals = append(result.Residuals, AlgebraRealizationResidual{
		Path: "input", Code: "RESIDUAL_REPORT_LIMIT",
		Message: fmt.Sprintf("%d additional residual items remain in the caller-supplied input and were not expanded in this bounded report", omitted),
	})
	return true
}

func algebraRealizationFailure(result AlgebraRealizationResult, kind status.OutcomeKind, code, path, message string) AlgebraRealizationResult {
	result.Status = kind
	result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: code, Path: path, Message: message, Severity: status.SeverityError})
	return result
}

func algebraRealizationError(result AlgebraRealizationResult, err error, fallbackCode, path string) (AlgebraRealizationResult, error) {
	if err == nil {
		return result, nil
	}
	var algebraFailure *algebra.Error
	if errors.As(err, &algebraFailure) {
		return algebraRealizationFailure(result, algebraFailure.Status, algebraFailure.Code, algebraFailure.Path, algebraFailure.Message), nil
	}
	var resultFailure interface{ Result() status.Result }
	if errors.As(err, &resultFailure) {
		failure := resultFailure.Result()
		result.Status = failure.Status
		result.Diagnostics = append(result.Diagnostics, failure.Diagnostics...)
		return result, nil
	}
	return algebraRealizationFailure(result, status.OutcomeError, fallbackCode, path, "realization failed unexpectedly: "+strings.TrimSpace(err.Error())), nil
}
