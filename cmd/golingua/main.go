package main

import (
	"os"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/cli"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/runtime"
)

func main() {
	loadEngine := func(path string) (*engine.Engine, error) {
		return runtime.LoadEngineFile(path, compiledRuntimeOptions())
	}
	service := app.NewService(app.Config{LoadEngine: loadEngine})
	if len(os.Args) > 1 && os.Args[1] == "forms" {
		os.Exit(runForms(os.Args[2:], os.Stdout, os.Stderr, service))
	}
	if len(os.Args) > 1 && os.Args[1] == "catalog" {
		os.Exit(runCatalog(os.Args[2:], os.Stdout, os.Stderr, service))
	}
	if len(os.Args) > 1 && os.Args[1] == "snapshot" {
		os.Exit(runSnapshot(os.Args[2:], os.Stdout, os.Stderr, service))
	}
	os.Exit(cli.RunWithInputAndService(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, service, loadEngine, runtime.LoadCoreFile))
}
