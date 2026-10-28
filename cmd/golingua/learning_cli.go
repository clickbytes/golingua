package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/bundled"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/learning"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/pack"
	"github.com/clickbytes/golingua/internal/provenance"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

const maxLearningRequestBytes = 64 << 10

type learningCLIOptions struct {
	operation, storePath, inputPath string
	id, actor, at                   string
	json                            bool
}

type learningAddRequest struct {
	ID                   string                  `json:"id"`
	Language             features.LanguageID     `json:"language"`
	LexemeID             string                  `json:"lexeme_id"`
	Features             []features.FeatureValue `json:"features"`
	OriginalForm         string                  `json:"original_form"`
	CorrectedForm        string                  `json:"corrected_form"`
	InitiatedBy          string                  `json:"initiated_by"`
	ReviewedBy           string                  `json:"reviewed_by"`
	RecordedAt           string                  `json:"recorded_at"`
	ReviewedAt           string                  `json:"reviewed_at"`
	PreviousCorrectionID string                  `json:"previous_correction_id,omitempty"`
	Provenance           provenance.SourceRef    `json:"provenance"`
}

type learningCLIResponse struct {
	Status      status.OutcomeKind  `json:"status"`
	Operation   string              `json:"operation,omitempty"`
	Message     string              `json:"message,omitempty"`
	Record      *learning.Record    `json:"record,omitempty"`
	Store       *learning.Store     `json:"store,omitempty"`
	Diagnostics []status.Diagnostic `json:"diagnostics,omitempty"`
}

func runLearning(args []string, stdout, stderr io.Writer) int {
	options, err := parseLearningOptions(args)
	if err != nil {
		return writeLearningResponse(stdout, stderr, learningCLIResponse{
			Status: status.OutcomeInvalid, Message: err.Error(),
			Diagnostics: []status.Diagnostic{{Code: "INVALID_LEARNING_ARGUMENTS", Path: "learning", Message: err.Error(), Severity: status.SeverityError}},
		})
	}
	var response learningCLIResponse
	switch options.operation {
	case "add":
		response = learningAdd(options)
	case "inspect":
		store, err := learning.Load(options.storePath)
		if err != nil {
			response = learningError(options.operation, status.OutcomeUnknown, "LEARNING_STORE_UNAVAILABLE", err)
		} else {
			response = learningCLIResponse{Status: status.OutcomeOK, Operation: options.operation, Store: store}
		}
	case "revoke":
		if err := validateLearningRevokeOptions(options.id, options.actor, options.at); err != nil {
			response = learningError(options.operation, status.OutcomeInvalid, "INVALID_LEARNING_REVOKE", err)
			break
		}
		if err := learning.Revoke(options.storePath, options.id, options.actor, options.at); err != nil {
			if isKnownRevokeValidationError(err) {
				response = learningError(options.operation, status.OutcomeInvalid, "LEARNING_REVOKE_FAILED", err)
			} else {
				response = learningMutationUncertain(options.operation, "LEARNING_REVOKE_STATE_UNCERTAIN", err)
			}
		} else {
			response = learningCLIResponse{Status: status.OutcomeOK, Operation: options.operation, Message: "correction revoked", Record: &learning.Record{ID: options.id}}
		}
	default:
		response = learningError(options.operation, status.OutcomeInvalid, "INVALID_LEARNING_OPERATION", errors.New("operation must be add, inspect, or revoke"))
	}
	return writeLearningResponse(stdout, stderr, response)
}

func parseLearningOptions(args []string) (learningCLIOptions, error) {
	var options learningCLIOptions
	if len(args) == 0 || len(args) > 32 {
		return options, errors.New("use learning add|inspect|revoke with bounded flags")
	}
	options.operation = args[0]
	if options.operation != "add" && options.operation != "inspect" && options.operation != "revoke" {
		return options, errors.New("operation must be add, inspect, or revoke")
	}
	seen := map[string]bool{}
	for index := 1; index < len(args); index++ {
		flag := args[index]
		if flag == "--json" {
			if seen[flag] {
				return options, errors.New("duplicate --json flag")
			}
			seen[flag], options.json = true, true
			continue
		}
		if flag != "--store" && flag != "--input" && flag != "--id" && flag != "--actor" && flag != "--at" {
			return options, fmt.Errorf("unknown learning flag %q", flag)
		}
		if seen[flag] {
			return options, fmt.Errorf("duplicate %s flag", flag)
		}
		seen[flag] = true
		if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
			return options, fmt.Errorf("missing value for %s", flag)
		}
		index++
		value := args[index]
		if value == "" || len(value) > 4096 {
			return options, fmt.Errorf("invalid value for %s", flag)
		}
		switch flag {
		case "--store":
			options.storePath = value
		case "--input":
			options.inputPath = value
		case "--id":
			options.id = value
		case "--actor":
			options.actor = value
		case "--at":
			options.at = value
		}
	}
	if !options.json || options.storePath == "" {
		return options, errors.New("learning requires --store <file> and --json")
	}
	switch options.operation {
	case "add":
		if options.inputPath == "" || options.id != "" || options.actor != "" || options.at != "" {
			return options, errors.New("learning add requires --input and accepts no revoke flags")
		}
	case "inspect":
		if options.inputPath != "" || options.id != "" || options.actor != "" || options.at != "" {
			return options, errors.New("learning inspect accepts only --store and --json")
		}
	case "revoke":
		if options.id == "" || options.actor == "" || options.at == "" || options.inputPath != "" {
			return options, errors.New("learning revoke requires --id, --actor, --at and accepts no input file")
		}
	}
	return options, nil
}

func learningAdd(options learningCLIOptions) learningCLIResponse {
	request, err := readLearningAddRequest(options.inputPath)
	if err != nil {
		return learningError(options.operation, status.OutcomeInvalid, "INVALID_LEARNING_INPUT", err)
	}
	candidate, err := bundled.Candidate()
	if err != nil {
		return learningError(options.operation, status.OutcomeUnknown, "LEARNING_BASELINE_UNAVAILABLE", err)
	}
	baseline, err := pack.Compile(candidate)
	if err != nil {
		return learningError(options.operation, status.OutcomeUnknown, "LEARNING_BASELINE_INVALID", err)
	}
	identity := baseline.Identity()
	basePolish, err := pl.Load()
	if err != nil {
		return learningError(options.operation, status.OutcomeUnknown, "POLISH_BASE_SEED_UNAVAILABLE", err)
	}
	baseEnglish, err := en.Load()
	if err != nil {
		return learningError(options.operation, status.OutcomeUnknown, "ENGLISH_BASE_SEED_UNAVAILABLE", err)
	}
	generate := func(languageID features.LanguageID, lexemeID string, values []features.FeatureValue) (morph.GenerationResult, error) {
		switch languageID {
		case "pl":
			return basePolish.Generate(lexemeID, values)
		case "en":
			return baseEnglish.Generate(lexemeID, values)
		default:
			return morph.GenerationResult{}, fmt.Errorf("unsupported language %q", languageID)
		}
	}
	baseSeed, err := baseSeedFor(request.Language, basePolish, baseEnglish)
	if err != nil {
		return learningError(options.operation, status.OutcomeUnsupported, "UNSUPPORTED_LEARNING_LANGUAGE", err)
	}
	generated, err := baseSeed.Generate(request.LexemeID, request.Features)
	if err != nil {
		return learningError(options.operation, status.OutcomeInvalid, "BASELINE_GENERATION_FAILED", err)
	}
	if generated.State != morph.GenerationGenerated || len(generated.Forms) != 1 {
		return learningError(options.operation, status.OutcomeIncomplete, "BASELINE_NOT_UNIQUE", fmt.Errorf("base seed returned state %q and %d forms; exactly one generated form is required", generated.State, len(generated.Forms)))
	}
	baseForm := generated.Forms[0]
	if baseForm.Surface != request.OriginalForm {
		return learningError(options.operation, status.OutcomeInvalid, "BASELINE_SURFACE_MISMATCH", fmt.Errorf("generated baseline surface does not match original_form"))
	}
	if baseForm.Source != morph.ExceptionSourceRegular || baseForm.ExceptionID != "" {
		return learningError(options.operation, status.OutcomeUnsupported, "BASELINE_USES_EXCEPTION", fmt.Errorf("corrections cannot replace an existing exception-based form"))
	}
	record, err := learning.NewRecord(
		request.ID, identity.DatasetID, identity.DatasetVersion, identity.Digest,
		request.CorrectedForm, request.InitiatedBy, request.RecordedAt, request.ReviewedBy, request.ReviewedAt, request.PreviousCorrectionID,
		request.Provenance, baseForm,
	)
	if err != nil {
		return learningError(options.operation, status.OutcomeInvalid, "INVALID_LEARNING_RECORD", err)
	}

	store, loadErr := learning.Load(options.storePath)
	var activeRules map[features.LanguageID][]morph.ExceptionRule
	if loadErr != nil && !errors.Is(loadErr, os.ErrNotExist) {
		return learningError(options.operation, status.OutcomeUnknown, "LEARNING_STORE_UNAVAILABLE", loadErr)
	}
	if loadErr == nil {
		activeRules, err = learning.ActiveRules(options.storePath, identity.DatasetID, identity.DatasetVersion, identity.Digest, generate)
		if err != nil {
			return learningError(options.operation, status.OutcomeInvalid, "ACTIVE_LEARNING_RULES_INVALID", err)
		}
	} else {
		activeRules = map[features.LanguageID][]morph.ExceptionRule{}
	}
	proposedRules := append([]morph.ExceptionRule(nil), activeRules[request.Language]...)
	alreadyActive := false
	if store != nil {
		for _, existing := range store.Records {
			if existing.ID == record.ID && reflect.DeepEqual(existing, record) {
				alreadyActive = true
				break
			}
		}
	}
	if !alreadyActive {
		proposedRules = append(proposedRules, record.Rule)
	}
	switch request.Language {
	case "pl":
		if _, err := pl.LoadWithExceptions(proposedRules); err != nil {
			return learningError(options.operation, status.OutcomeInvalid, "LEARNING_RULES_CONFLICT", err)
		}
	case "en":
		if _, err := en.LoadWithExceptions(proposedRules); err != nil {
			return learningError(options.operation, status.OutcomeInvalid, "LEARNING_RULES_CONFLICT", err)
		}
	}
	if err := learning.Add(options.storePath, record); err != nil {
		if isLearningAddLimitError(err) {
			return learningError(options.operation, status.OutcomeResourceLimit, "LEARNING_STORE_LIMIT", err)
		}
		if isKnownAddValidationError(err) {
			return learningError(options.operation, status.OutcomeInvalid, "LEARNING_ADD_FAILED", err)
		}
		return learningMutationUncertain(options.operation, "LEARNING_ADD_STATE_UNCERTAIN", err)
	}
	return learningCLIResponse{Status: status.OutcomeOK, Operation: options.operation, Record: &record}
}

func baseSeedFor(languageID features.LanguageID, polish *pl.Seed, english *en.Seed) (interface {
	Generate(string, []features.FeatureValue) (morph.GenerationResult, error)
}, error) {
	switch languageID {
	case "pl":
		return polish, nil
	case "en":
		return english, nil
	default:
		return nil, fmt.Errorf("language must be pl or en")
	}
}

func readLearningAddRequest(path string) (learningAddRequest, error) {
	var request learningAddRequest
	if path == "" {
		return request, errors.New("input path is required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return request, fmt.Errorf("stat learning input: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxLearningRequestBytes {
		return request, fmt.Errorf("learning input must be a nonempty regular file no larger than %d bytes", maxLearningRequestBytes)
	}
	file, err := os.Open(path)
	if err != nil {
		return request, fmt.Errorf("open learning input: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxLearningRequestBytes+1))
	if err != nil {
		return request, fmt.Errorf("read learning input: %w", err)
	}
	if len(data) == 0 || len(data) > maxLearningRequestBytes {
		return request, fmt.Errorf("learning input exceeds %d bytes or is empty", maxLearningRequestBytes)
	}
	if err := validateLearningJSON(data); err != nil {
		return request, err
	}
	if err := validateLearningInputShape(data); err != nil {
		return request, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, fmt.Errorf("decode learning input: %w", err)
	}
	if err := requireLearningEOF(decoder); err != nil {
		return request, err
	}
	if request.Language != "pl" && request.Language != "en" {
		return request, errors.New("language must be pl or en")
	}
	return request, nil
}

func validateLearningInputShape(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return errors.New("learning input must be a JSON object")
	}
	allowed := map[string]bool{
		"id": true, "language": true, "lexeme_id": true, "features": true,
		"original_form": true, "corrected_form": true, "initiated_by": true,
		"reviewed_by": true, "recorded_at": true, "reviewed_at": true,
		"previous_correction_id": true, "provenance": true,
	}
	required := []string{"id", "language", "lexeme_id", "features", "original_form", "corrected_form", "initiated_by", "reviewed_by", "recorded_at", "reviewed_at", "provenance"}
	for key := range fields {
		if !allowed[key] {
			return fmt.Errorf("unknown learning input field %q", key)
		}
	}
	for _, key := range required {
		raw, exists := fields[key]
		if !exists || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("required learning input field %q is missing or null", key)
		}
	}
	if raw, exists := fields["previous_correction_id"]; exists && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return errors.New("previous_correction_id cannot be null")
	}
	var featureValues []json.RawMessage
	if err := json.Unmarshal(fields["features"], &featureValues); err != nil || featureValues == nil || len(featureValues) == 0 || len(featureValues) > features.MaxBundleValues {
		return errors.New("features must be a nonempty bounded array")
	}
	for index, raw := range featureValues {
		if _, err := strictLearningObject(raw, []string{"feature", "state", "value", "variable"}, []string{"feature", "state"}); err != nil {
			return fmt.Errorf("features[%d]: %w", index, err)
		}
	}
	provenanceFields, err := strictLearningObject(fields["provenance"],
		[]string{"id", "path", "uri", "version", "license", "author", "checksum_sha256"},
		[]string{"id", "version", "license"})
	if err != nil {
		return fmt.Errorf("provenance: %w", err)
	}
	if _, pathExists := provenanceFields["path"]; !pathExists {
		if _, uriExists := provenanceFields["uri"]; !uriExists {
			return errors.New("provenance requires path or uri")
		}
	}
	return nil
}

func strictLearningObject(data []byte, allowed, required []string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return nil, errors.New("expected a JSON object")
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = true
	}
	for key, value := range fields {
		if !allowedSet[key] {
			return nil, fmt.Errorf("unknown field %q", key)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("field %q cannot be null", key)
		}
	}
	for _, key := range required {
		if _, exists := fields[key]; !exists {
			return nil, fmt.Errorf("required field %q is missing", key)
		}
	}
	return fields, nil
}

func validateLearningJSON(data []byte) error {
	if !utf8.Valid(data) {
		return errors.New("learning input is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanLearningJSON(decoder, 0); err != nil {
		return err
	}
	return requireLearningEOF(decoder)
}

func scanLearningJSON(decoder *json.Decoder, depth int) error {
	if depth > 32 {
		return errors.New("learning input JSON nesting exceeds 32 levels")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("learning input object key must be a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON key %q", key)
			}
			seen[key] = true
			if err := scanLearningJSON(decoder, depth+1); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		for decoder.More() {
			if err := scanLearningJSON(decoder, depth+1); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func requireLearningEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("learning input must contain exactly one JSON value")
		}
		return err
	}
	return nil
}

func learningError(operation string, outcome status.OutcomeKind, code string, err error) learningCLIResponse {
	return learningCLIResponse{
		Status: outcome, Operation: operation, Message: err.Error(),
		Diagnostics: []status.Diagnostic{{Code: code, Path: "learning." + operation, Message: err.Error(), Severity: status.SeverityError}},
	}
}

func learningMutationUncertain(operation, code string, err error) learningCLIResponse {
	message := fmt.Sprintf("mutation failed; write durability may be uncertain. Inspect the learning store before retrying. Original error: %v", err)
	return learningError(operation, status.OutcomeUnknown, code, errors.New(message))
}

func isLearningAddLimitError(err error) bool {
	message := err.Error()
	return strings.HasPrefix(message, "learning record count exceeds ") ||
		strings.HasPrefix(message, "active learning record count exceeds ")
}

func isKnownAddValidationError(err error) bool {
	message := err.Error()
	for _, prefix := range []string{
		"invalid learning record:",
		"new learning records must be active",
		"learning record ID ",
		"active correction already exists",
		"new correction must identify revoked predecessor",
		"previous correction ID was supplied",
		"target already has an active correction",
	} {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

func isKnownRevokeValidationError(err error) bool {
	message := err.Error()
	if !strings.HasPrefix(message, "learning record \"") {
		return false
	}
	return strings.HasSuffix(message, " is already revoked") || strings.HasSuffix(message, " does not exist")
}

func validateLearningRevokeOptions(id, actor, at string) error {
	if !validLearningID(id) {
		return errors.New("revocation ID must be a bounded lowercase identifier")
	}
	if actor == "" || len(actor) > 256 || !utf8.ValidString(actor) || strings.TrimSpace(actor) != actor || utf8.RuneCountInString(actor) > morph.MaxPrincipalPartRunes {
		return errors.New("revocation actor is empty, invalid, or too long")
	}
	for _, character := range actor {
		if unicode.IsControl(character) {
			return errors.New("revocation actor contains a control character")
		}
	}
	if at == "" || len(at) > 64 {
		return errors.New("revocation time must be RFC3339")
	}
	if _, err := time.Parse(time.RFC3339Nano, at); err != nil {
		return errors.New("revocation time must be RFC3339")
	}
	return nil
}

func validLearningID(value string) bool {
	if len(value) == 0 || len(value) > morph.MaxExceptionIDBytes {
		return false
	}
	for index, character := range value {
		if index == 0 && (character < 'a' || character > 'z') {
			return false
		}
		if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

func writeLearningResponse(stdout, stderr io.Writer, response learningCLIResponse) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(true)
	if err := encoder.Encode(response); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize learning response: %v\n", err)
		return 1
	}
	switch response.Status {
	case status.OutcomeOK:
		return 0
	case status.OutcomeInvalid:
		return 2
	case status.OutcomeUnsupported:
		return 3
	case status.OutcomeResourceLimit:
		return 4
	case status.OutcomeUnknown:
		return 5
	case status.OutcomeAmbiguous:
		return 6
	case status.OutcomeIncomplete:
		return 7
	default:
		return 1
	}
}
