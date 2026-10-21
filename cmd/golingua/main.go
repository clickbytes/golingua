package main

import (
	"os"

	"github.com/clickbytes/golingua/internal/cli"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/runtime"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "forms" {
		os.Exit(runForms(os.Args[2:], os.Stdout, os.Stderr))
	}
	if len(os.Args) > 1 && os.Args[1] == "catalog" {
		os.Exit(runCatalog(os.Args[2:], os.Stdout, os.Stderr))
	}
	loadEngine := func(path string) (*engine.Engine, error) {
		return runtime.LoadEngineFile(path, compiledRuntimeOptions())
	}
	os.Exit(cli.RunWithInputAndLoaders(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, loadEngine, runtime.LoadCoreFile))
}
