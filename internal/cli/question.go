package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/planner"
	"github.com/clickbytes/golingua/internal/reason"
	"github.com/clickbytes/golingua/internal/status"
)

const (
	QuestionRequestSchemaVersion = 1
	MaxQuestionRequestBytes      = 4 << 20
	maxQuestionOutputBytes       = 8 << 20
	maxQuestionJSONDepth         = 64
)

// CoreLoader resolves the caller-selected immutable CORE snapshot.
type CoreLoader func(string) (*core.ConceptGraph, error)

type questionRequest struct {
	SchemaVersion int                  `json:"schema_version"`
	Question      reason.QuestionInput `json:"question"`
	Facts         []facts.Fact         `json:"facts"`
	Rules         []reason.Rule        `json:"rules,omitempty"`
}

type questionOutput struct {
	SchemaVersion int                    `json:"schema_version"`
	Status        status.OutcomeKind     `json:"status"`
	Goal          planner.DialogueGoal   `json:"goal,omitempty"`
	Question      *reason.Question       `json:"question,omitempty"`
	Answer        *reason.QuestionResult `json:"answer,omitempty"`
	Evidence      []facts.Fact           `json:"evidence"`
	Diagnostics   []status.Diagnostic    `json:"diagnostics,omitempty"`
}

func runQuestion(args []string, input io.Reader, stdout, stderr io.Writer, loadCore CoreLoader) int {
	configPath, parseError := parseQuestionArgs(args)
	if parseError != "" {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeInvalid, "INVALID_QUESTION_ARGUMENTS", parseError, "question")))
	}
	if input == nil {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeError, "INPUT_UNAVAILABLE", "stdin reader is unavailable", "question")))
	}
	data, err := io.ReadAll(io.LimitReader(input, MaxQuestionRequestBytes+1))
	if err != nil {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeError, "QUESTION_INPUT_UNAVAILABLE", "question request could not be read", "question")))
	}
	if len(data) == 0 || len(data) > MaxQuestionRequestBytes {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeResourceLimit, "QUESTION_REQUEST_SIZE", "question request must be non-empty and within the input limit", "question")))
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeInvalid, "INVALID_QUESTION_JSON", "question request has invalid or duplicate JSON fields", "question")))
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var request questionRequest
	if err := decoder.Decode(&request); err != nil {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeInvalid, "INVALID_QUESTION_JSON", "question request does not match the supported schema", "question")))
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeInvalid, "TRAILING_QUESTION_JSON", "question request must contain exactly one JSON value", "question")))
	}
	if request.SchemaVersion != QuestionRequestSchemaVersion {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeInvalid, "QUESTION_SCHEMA_VERSION", "unsupported question request schema version", "schema_version")))
	}
	if loadCore == nil {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeUnsupported, "CORE_LOADER_UNAVAILABLE", "this executable cannot load a configured CORE snapshot", "runtime_config")))
	}
	ontology, err := loadCore(configPath)
	if err != nil {
		return writeQuestion(stdout, stderr, errorQuestion(err, "CORE_LOAD_ERROR", "configured CORE snapshot could not be loaded", "runtime_config"))
	}
	set, err := facts.Build(ontology, facts.Input{SchemaVersion: facts.SchemaVersion, Context: request.Question.Context, Facts: request.Facts}, facts.DefaultBudget())
	if err != nil {
		return writeQuestion(stdout, stderr, errorQuestion(err, "FACT_SET_INVALID", "request FactSet failed validation", "facts"))
	}
	question, err := reason.Plan(request.Question, ontology, meaning.DefaultBudget())
	if err != nil {
		return writeQuestion(stdout, stderr, errorQuestion(err, "QUESTION_PLAN_FAILED", "typed question could not be planned", "question"))
	}
	answer, err := reason.Execute(set, request.Rules, question, reason.DefaultQueryBudget())
	if err != nil {
		return writeQuestion(stdout, stderr, errorQuestion(err, "QUESTION_EXECUTION_FAILED", "question could not be answered", "question"))
	}
	decision, err := planner.ChooseGoal(planner.DecisionInput{Query: question, Result: &answer})
	if err != nil {
		return writeQuestion(stdout, stderr, errorQuestion(err, "CONTENT_GOAL_FAILED", "response goal could not be selected", "answer"))
	}
	result := questionOutput{
		SchemaVersion: QuestionRequestSchemaVersion,
		Status:        answer.Outcome, Goal: decision.Goal,
		Question: &question, Answer: &answer,
		Evidence:    questionEvidence(answer, set),
		Diagnostics: []status.Diagnostic{},
	}
	return writeQuestion(stdout, stderr, result)
}

func parseQuestionArgs(args []string) (string, string) {
	if len(args) < 2 || args[0] != "question" {
		return "", "usage: question --runtime-config <path> --json"
	}
	var config string
	seenJSON := false
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--json":
			if seenJSON {
				return "", "duplicate --json flag"
			}
			seenJSON = true
		case "--runtime-config":
			if config != "" || index+1 >= len(args) || args[index+1] == "--json" || args[index+1] == "--runtime-config" {
				return "", "--runtime-config requires one non-empty path"
			}
			index++
			config = args[index]
		default:
			return "", fmt.Sprintf("unknown or malformed question argument %q", args[index])
		}
	}
	if !seenJSON || config == "" {
		return "", "question requires --runtime-config <path> and --json"
	}
	return config, ""
}

func writeQuestion(stdout, stderr io.Writer, result questionOutput) int {
	if !result.Status.Valid() {
		return writeResult(stdout, stderr, diagnosticResult(status.OutcomeError, "INVALID_QUESTION_OUTPUT", "question response has an invalid status", "question"))
	}
	data, err := json.Marshal(result)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize question response: %v\n", err)
		return ExitError
	}
	if len(data) > maxQuestionOutputBytes {
		result = questionFailure(diagnosticResult(status.OutcomeResourceLimit, "QUESTION_OUTPUT_SIZE", "question response exceeds the output limit", "question"))
		data, err = json.Marshal(result)
	}
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not serialize question response: %v\n", err)
		return ExitError
	}
	if _, err := fmt.Fprintln(stdout, string(data)); err != nil {
		_, _ = fmt.Fprintf(stderr, "INFRASTRUCTURE_ERROR: could not write question response: %v\n", err)
		return ExitError
	}
	return exitCode(result.Status)
}

func questionFailure(result status.Result) questionOutput {
	return questionOutput{SchemaVersion: QuestionRequestSchemaVersion, Status: result.Status, Evidence: []facts.Fact{}, Diagnostics: append([]status.Diagnostic(nil), result.Diagnostics...)}
}

func errorQuestion(err error, fallbackCode, fallbackMessage, path string) questionOutput {
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		result := typed.Result()
		if result.Validate() == nil && result.Status != status.OutcomeOK {
			return questionFailure(result)
		}
	}
	return questionFailure(diagnosticResult(status.OutcomeError, fallbackCode, fallbackMessage, path))
}

func questionEvidence(answer reason.QuestionResult, set *facts.Set) []facts.Fact {
	ids := map[string]struct{}{}
	var collect func([]reason.ProofNode)
	collect = func(proofs []reason.ProofNode) {
		for _, proof := range proofs {
			for _, id := range proof.PremiseFactIDs {
				ids[id] = struct{}{}
			}
		}
	}
	collect(answer.Proof)
	for _, candidate := range answer.Candidates {
		collect(candidate.Proof)
		for _, fact := range candidate.NegativeEvidence {
			ids[fact.ID] = struct{}{}
		}
	}
	for _, fact := range answer.NegativeEvidence {
		ids[fact.ID] = struct{}{}
	}
	keys := make([]string, 0, len(ids))
	for id := range ids {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	result := make([]facts.Fact, 0, len(keys))
	for _, id := range keys {
		if fact, exists := set.ByID(id); exists {
			result = append(result, fact)
		}
	}
	return result
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var visit func(int) error
	visit = func(depth int) error {
		if depth > maxQuestionJSONDepth {
			return errors.New("question JSON nesting exceeds the depth limit")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				token, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := token.(string)
				if !ok {
					return errors.New("JSON object key is not a string")
				}
				if _, exists := seen[key]; exists {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = struct{}{}
				if err := visit(depth + 1); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim('}') {
				return errors.New("unclosed JSON object")
			}
		case '[':
			for decoder.More() {
				if err := visit(depth + 1); err != nil {
					return err
				}
			}
			closing, err := decoder.Token()
			if err != nil || closing != json.Delim(']') {
				return errors.New("unclosed JSON array")
			}
		default:
			return fmt.Errorf("unexpected JSON delimiter %q", delim)
		}
		return nil
	}
	if err := visit(0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}
