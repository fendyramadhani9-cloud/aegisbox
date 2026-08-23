package tests

import (
	"encoding/json"
	"runtime"
	"testing"

	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestFoundation_DefaultsAndExecutionContract(t *testing.T) {
	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("foundation failure: default config is invalid: %v", err)
	}

	req := executor.ExecutionRequest{
		Language:  "python",
		Code:      "print('foundation test')",
		TimeoutMs: cfg.Sandbox.DefaultTimeoutMs,
		MaxMemMB:  cfg.Sandbox.DefaultMemoryMB,
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("foundation failure: valid request failed validation: %v", err)
	}

	res := executor.ExecutionResult{
		ExecutionID:     "foundation-1",
		Status:          executor.StatusCompleted,
		Stdout:          "foundation test\n",
		ExitCode:        0,
		ExecutionTimeMs: 15,
	}

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("foundation failure: could not marshal execution result: %v", err)
	}

	if len(data) == 0 {
		t.Fatalf("foundation failure: marshaled output is empty")
	}

	t.Logf("AegisBox Foundation Verification PASSED on %s/%s", runtime.GOOS, runtime.GOARCH)
}
