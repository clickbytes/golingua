package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/dialogue"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/facts"
	"github.com/clickbytes/golingua/internal/features"
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
	Response      *dialogue.Result       `json:"response,omitempty"`
	Evidence      []facts.Fact           `json:"evidence"`
	Diagnostics   []status.Diagnostic    `json:"diagnostics,omitempty"`
}

type questionOptions struct {
	RuntimeConfig string
	Language      string
}

func runQuestion(args []string, input io.Reader, stdout, stderr io.Writer, loadEngine EngineLoader, loadCore CoreLoader) int {
	options, parseError := parseQuestionArgs(args)
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
	var ontology *core.ConceptGraph
	var configuredEngine *engine.Engine
	if options.Language != "" {
		if loadEngine == nil {
			return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeUnsupported, "RUNTIME_LOADER_UNAVAILABLE", "this executable cannot load configured language runtimes", "runtime_config")))
		}
		configuredEngine, err = loadEngine(options.RuntimeConfig)
		if err != nil {
			return writeQuestion(stdout, stderr, errorQuestion(err, "ENGINE_SETUP_ERROR", "configured language runtime could not be loaded", "runtime_config"))
		}
		ontology = configuredEngine.CoreGraph()
	} else {
		if loadCore == nil {
			return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeUnsupported, "CORE_LOADER_UNAVAILABLE", "this executable cannot load a configured CORE snapshot", "runtime_config")))
		}
		ontology, err = loadCore(options.RuntimeConfig)
		if err != nil {
			return writeQuestion(stdout, stderr, errorQuestion(err, "CORE_LOAD_ERROR", "configured CORE snapshot could not be loaded", "runtime_config"))
		}
	}
	if ontology == nil {
		return writeQuestion(stdout, stderr, questionFailure(diagnosticResult(status.OutcomeUnknown, "RUNTIME_SNAPSHOT_UNAVAILABLE", "configured CORE snapshot could not be loaded", "runtime_config")))
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
	if options.Language != "" {
		contentPlan, planErr := planner.Build(planner.PlanInput{
			Decision:        planner.DecisionInput{Query: question, Result: &answer},
			QuestionMeaning: request.Question.Meaning, Core: ontology, FactSet: set,
		})
		if planErr != nil {
			failure := errorQuestion(planErr, "CONTENT_PLAN_FAILED", "response content could not be planned", "answer")
			result.Status = failure.Status
			result.Diagnostics = append(result.Diagnostics, failure.Diagnostics...)
			return writeQuestion(stdout, stderr, result)
		}
		var target engine.TargetRealizer
		if configuredEngine != nil {
			target, _ = configuredEngine.TargetRealizerFor(features.LanguageID(options.Language))
		}
		responseContext, cancel := context.WithTimeout(context.Background(), engine.DefaultBudget().RequestTimeout)
		response, realizeErr := dialogue.Realize(responseContext, contentPlan, ontology, target)
		cancel()
		if realizeErr != nil {
			failure := failedDialogueResponse(contentPlan, realizeErr)
			result.Response = &failure
			result.Status = failure.Outcome
			result.Diagnostics = append(result.Diagnostics, failure.Diagnostics...)
			return writeQuestion(stdout, stderr, result)
		}
		result.Response = &response
		if response.Outcome != status.OutcomeOK {
			result.Status = response.Outcome
			result.Diagnostics = append(result.Diagnostics, response.Diagnostics...)
		}
	}
	return writeQuestion(stdout, stderr, result)
}

func failedDialogueResponse(plan planner.ContentPlan, err error) dialogue.Result {
	result := dialogue.Result{
		Goal: plan.Goal, Outcome: status.OutcomeError, ResponseMeaning: plan.ResponseMeaning,
		SelectedFactIDs:    append([]string(nil), plan.SelectedFactIDs...),
		ConfidenceEvidence: append([]planner.EvidenceRef(nil), plan.ConfidenceEvidence...),
		ProofRefs:          append([]string(nil), plan.ProofRefs...),
		Constructions:      []dialogue.ConstructionRef{}, Alternatives: []engine.SurfaceAlternative{},
		Diagnostics: []status.Diagnostic{{Code: "DIALOGUE_REALIZATION_FAILED", Path: "response", Message: "response could not be realized", Severity: status.SeverityError}},
	}
	var typed interface{ Result() status.Result }
	if errors.As(err, &typed) {
		failure := typed.Result()
		if failure.Validate() == nil && failure.Status != status.OutcomeOK {
			result.Outcome = failure.Status
			result.Diagnostics = append([]status.Diagnostic(nil), failure.Diagnostics...)
		}
	}
	return result
}

func parseQuestionArgs(args []string) (questionOptions, string) {
	if len(args) < 2 || args[0] != "question" {
		return questionOptions{}, "usage: question --runtime-config <path> [--language pl|en] --json"
	}
	var options questionOptions
	seenJSON := false
	for index := 1; index < len(args); index++ {
		switch args[index] {
		case "--json":
			if seenJSON {
				return questionOptions{}, "duplicate --json flag"
			}
			seenJSON = true
		case "--runtime-config":
			if options.RuntimeConfig != "" || index+1 >= len(args) || args[index+1] == "--json" || args[index+1] == "--runtime-config" || args[index+1] == "--language" {
				return questionOptions{}, "--runtime-config requires one non-empty path"
			}
			index++
			options.RuntimeConfig = args[index]
		case "--language":
			if options.Language != "" || index+1 >= len(args) || args[index+1] == "--json" || args[index+1] == "--runtime-config" || args[index+1] == "--language" {
				return questionOptions{}, "--language requires one language ID: pl or en"
			}
			index++
			options.Language = args[index]
			if options.Language != "pl" && options.Language != "en" {
				return questionOptions{}, "--language must be pl or en"
			}
		default:
			return questionOptions{}, fmt.Sprintf("unknown or malformed question argument %q", args[index])
		}
	}
	if !seenJSON || options.RuntimeConfig == "" {
		return questionOptions{}, "question requires --runtime-config <path> and --json"
	}
	return options, ""
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
