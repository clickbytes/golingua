package main

import (
	"os"

	"github.com/clickbytes/golingua/internal/cli"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/runtime"
)

func main() {
	loadEngine := func(path string) (*engine.Engine, error) {
		return runtime.LoadEngineFile(path, compiledRuntimeOptions())
	}
	os.Exit(cli.RunWithInputAndEngineLoader(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, loadEngine))
}
