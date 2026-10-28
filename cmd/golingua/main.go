package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/cli"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/tui"
	"github.com/clickbytes/golingua/languages/en"
	"github.com/clickbytes/golingua/languages/pl"
)

func main() {
	defaultLoadEngine := func(path string) (*engine.Engine, error) {
		return runtime.LoadEngineFile(path, compiledRuntimeOptions())
	}
	if len(os.Args) == 1 && tui.Interactive(os.Stdin, os.Stdout) {
		service := app.NewService(app.Config{LoadEngine: defaultLoadEngine})
		if err := tui.Run(service, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "learning" {
		os.Exit(runLearning(os.Args[2:], os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "forms" {
		args, storePath, useLearning, err := extractLearningStore(os.Args[2:])
		if err != nil {
			os.Exit(writeForms(os.Stdout, os.Stderr, formsResponse{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{formsDiagnostic("INVALID_LEARNING_STORE_OPTION", err.Error(), "learning_store")}}))
		}
		service := app.NewService(app.Config{LoadEngine: defaultLoadEngine})
		if useLearning {
			polishSeed, englishSeed, options, err := prepareLearningRuntime(storePath)
			if err != nil {
				os.Exit(writeLearningResponse(os.Stdout, os.Stderr, learningError("forms", status.OutcomeInvalid, "LEARNING_RUNTIME_SETUP_FAILED", err)))
			}
			service = learningService(options, polishSeed, englishSeed)
		}
		os.Exit(runForms(args, os.Stdout, os.Stderr, service))
	}
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "translate" || args[0] == "algebra" || args[0] == "question") {
		remaining, storePath, useLearning, err := extractLearningStore(args[1:])
		if err != nil {
			os.Exit(writeLearningResponse(os.Stdout, os.Stderr, learningError(args[0], status.OutcomeInvalid, "INVALID_LEARNING_STORE_OPTION", err)))
		}
		args = append([]string{args[0]}, remaining...)
		if useLearning {
			polishSeed, englishSeed, options, err := prepareLearningRuntime(storePath)
			if err != nil {
				os.Exit(writeLearningResponse(os.Stdout, os.Stderr, learningError(args[0], status.OutcomeInvalid, "LEARNING_RUNTIME_SETUP_FAILED", err)))
			}
			loadEngine := func(path string) (*engine.Engine, error) {
				return runtime.LoadEngineFile(path, options)
			}
			service := learningService(options, polishSeed, englishSeed)
			os.Exit(cli.RunWithInputAndService(args, os.Stdin, os.Stdout, os.Stderr, service, loadEngine, runtime.LoadCoreFile))
		}
	}
	service := app.NewService(app.Config{LoadEngine: defaultLoadEngine})
	if len(os.Args) > 1 && os.Args[1] == "catalog" {
		os.Exit(runCatalog(os.Args[2:], os.Stdout, os.Stderr, service))
	}
	if len(os.Args) > 1 && os.Args[1] == "snapshot" {
		os.Exit(runSnapshot(os.Args[2:], os.Stdout, os.Stderr, service))
	}
	os.Exit(cli.RunWithInputAndService(args, os.Stdin, os.Stdout, os.Stderr, service, defaultLoadEngine, runtime.LoadCoreFile))
}

func extractLearningStore(args []string) ([]string, string, bool, error) {
	remaining := make([]string, 0, len(args))
	storePath := ""
	present := false
	for index := 0; index < len(args); index++ {
		if args[index] != "--learning-store" {
			remaining = append(remaining, args[index])
			continue
		}
		if present {
			return nil, "", false, fmt.Errorf("duplicate --learning-store option")
		}
		present = true
		if index+1 >= len(args) || strings.HasPrefix(args[index+1], "--") {
			return nil, "", false, fmt.Errorf("--learning-store requires a file path")
		}
		index++
		storePath = args[index]
		if storePath == "" || len(storePath) > 4096 {
			return nil, "", false, fmt.Errorf("--learning-store path is empty or exceeds 4096 bytes")
		}
	}
	return remaining, storePath, present, nil
}

func learningService(options runtime.Options, polishSeed *pl.Seed, englishSeed *en.Seed) *app.Service {
	return app.NewService(app.Config{
		LoadEngine: func(path string) (*engine.Engine, error) {
			return runtime.LoadEngineFile(path, options)
		},
		LoadMorphology: func(languageID features.LanguageID) (app.Morphology, error) {
			switch languageID {
			case "pl":
				return polishSeed, nil
			case "en":
				return englishSeed, nil
			default:
				return nil, fmt.Errorf("unsupported morphology language %q", languageID)
			}
		},
	})
}
