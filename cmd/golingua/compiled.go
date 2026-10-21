package main

import "github.com/clickbytes/golingua/internal/runtime"

// compiledRuntimeOptions is the sole compile-time registration point for
// trusted language providers and stage factories in the CLI. The bundled
// PL/EN seeds do not declare sentence capabilities or approved bindings.
func compiledRuntimeOptions() runtime.Options {
	return runtime.Options{}
}
