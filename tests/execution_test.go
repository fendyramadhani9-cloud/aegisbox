package tests

import (
	"context"
	"testing"

	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
	"github.com/aegisbox/aegisbox/internal/runtime"
	"github.com/aegisbox/aegisbox/internal/sandbox"
)

func newTestExecutionManager(t *testing.T) *executor.ExecutionManager {
	cfg := config.DefaultConfig()
	logger := logging.Default()
	resolver := runtime.DefaultRegistry()
	sbxAdapter := sandbox.NewSandboxAdapter(cfg, logger)
	return executor.NewExecutionManager(cfg, resolver, sbxAdapter, logger)
}

func TestExecution_HelloWorld(t *testing.T) {
	mgr := newTestExecutionManager(t)

	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "print('Hello, AegisBox!')",
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if res.Status != executor.StatusCompleted {
		t.Fatalf("expected COMPLETED, got %s (err: %s)", res.Status, res.ErrorMessage)
	}

	if res.Stdout != "Hello, AegisBox!\n" {
		t.Errorf("expected stdout 'Hello, AegisBox!\\n', got '%s'", res.Stdout)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
}

func TestExecution_SyntaxError(t *testing.T) {
	mgr := newTestExecutionManager(t)

	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "def broken_syntax(:",
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if res.Status != executor.StatusRuntimeError {
		t.Fatalf("expected RUNTIME_ERROR, got %s", res.Status)
	}
	if res.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code for syntax error")
	}
}

func TestExecution_RuntimeError(t *testing.T) {
	mgr := newTestExecutionManager(t)

	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "raise ValueError('custom error')",
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if res.Status != executor.StatusRuntimeError {
		t.Fatalf("expected RUNTIME_ERROR, got %s", res.Status)
	}
}
