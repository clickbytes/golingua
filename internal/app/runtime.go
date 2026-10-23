package app

import (
	"context"

	"github.com/clickbytes/golingua/internal/status"
)

type RuntimeValidationRequest struct {
	RuntimeConfig string
}

// RuntimeValidationResult reports whether the selected runtime can be loaded.
// It does not promise support for every request or activate a snapshot.
type RuntimeValidationResult struct {
	Ready  bool
	Result status.Result
}

func (service *Service) ValidateRuntime(ctx context.Context, request RuntimeValidationRequest) RuntimeValidationResult {
	if ctx != nil && ctx.Err() != nil {
		return RuntimeValidationResult{Result: failure(status.OutcomeError, "RUNTIME_CHECK_CANCELED", "runtime_config", "runtime validation was canceled")}
	}
	if request.RuntimeConfig == "" {
		return RuntimeValidationResult{Result: failure(status.OutcomeInvalid, "RUNTIME_CONFIG_PATH_REQUIRED", "runtime_config", "select a runtime config path in F5 Settings")}
	}
	if service == nil || service.loadEngine == nil {
		return RuntimeValidationResult{Result: failure(status.OutcomeUnsupported, "RUNTIME_LOADER_UNAVAILABLE", "runtime_config", "runtime loader is unavailable")}
	}
	loaded, err := service.loadEngine(request.RuntimeConfig)
	if err != nil {
		return RuntimeValidationResult{Result: errorResult(err, "ENGINE_SETUP_ERROR", "runtime_config", "runtime setup failed unexpectedly")}
	}
	if loaded == nil {
		return RuntimeValidationResult{Result: failure(status.OutcomeError, "ENGINE_SETUP_ERROR", "runtime_config", "runtime loader returned no engine")}
	}
	if ctx != nil && ctx.Err() != nil {
		return RuntimeValidationResult{Result: failure(status.OutcomeError, "RUNTIME_CHECK_CANCELED", "runtime_config", "runtime validation was canceled")}
	}
	return RuntimeValidationResult{Ready: true, Result: status.Result{Status: status.OutcomeOK}}
}
