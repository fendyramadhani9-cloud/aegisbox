package executor_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
)

type mockRuntime struct {
	name string
}

func (m *mockRuntime) Name() string {
	return m.name
}

func (m *mockRuntime) Validate(req *executor.ExecutionRequest) error {
	return nil
}

func (m *mockRuntime) Prepare(ctx context.Context, workspaceDir string, req *executor.ExecutionRequest) error {
	return nil
}

func (m *mockRuntime) Command(req *executor.ExecutionRequest) ([]string, error) {
	return []string{"mock-runtime", "/workspace/main.py"}, nil
}

func (m *mockRuntime) Cleanup(workspaceDir string) error {
	return nil
}

type mockResolver struct {
	runtimes map[string]executor.RuntimeAdapter
}

func (m *mockResolver) Get(name string) (executor.RuntimeAdapter, error) {
	rt, exists := m.runtimes[name]
	if !exists {
		return nil, errors.New("runtime not found")
	}
	return rt, nil
}

func (m *mockResolver) SupportedLanguages() []string {
	return []string{"python", "mock"}
}

type mockSandbox struct {
	result  *executor.ExecutionResult
	memPeak int64
	oom     bool
	pids    bool
	err     error
}

func (s *mockSandbox) Execute(
	ctx context.Context,
	execID string,
	cmd []string,
	workspaceDir string,
	timeout time.Duration,
	memMB int64,
	pidsMax int64,
) (*executor.ExecutionResult, int64, bool, bool, error) {
	if s.result != nil {
		s.result.ExecutionID = execID
	}
	return s.result, s.memPeak, s.oom, s.pids, s.err
}

func TestExecutionManager_Lifecycle(t *testing.T) {
	cfg := config.DefaultConfig()
	resolver := &mockResolver{
		runtimes: map[string]executor.RuntimeAdapter{
			"python": &mockRuntime{name: "python"},
		},
	}
	sbx := &mockSandbox{
		result: &executor.ExecutionResult{
			Stdout:   "hello mock\n",
			ExitCode: 0,
		},
		memPeak: 123456,
	}

	logger := logging.Default()
	mgr := executor.NewExecutionManager(cfg, resolver, sbx, logger)

	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "print('hello')",
		TimeoutMs: 1000,
		MaxMemMB:  64,
	}

	result := mgr.Execute(context.Background(), req)
	if result.Status != executor.StatusCompleted {
		t.Fatalf("expected status COMPLETED, got: %s (error: %s)", result.Status, result.ErrorMessage)
	}

	if result.Stdout != "hello mock\n" {
		t.Errorf("expected stdout 'hello mock\\n', got '%s'", result.Stdout)
	}

	if result.ExecutionID == "" {
		t.Errorf("expected non-empty execution ID")
	}
}

func TestExecutionManager_UnsupportedLanguage(t *testing.T) {
	cfg := config.DefaultConfig()
	resolver := &mockResolver{
		runtimes: map[string]executor.RuntimeAdapter{},
	}
	sbx := &mockSandbox{}
	mgr := executor.NewExecutionManager(cfg, resolver, sbx, logging.Default())

	req := &executor.ExecutionRequest{
		Language:  "rust",
		Code:      "fn main() {}",
		TimeoutMs: 1000,
		MaxMemMB:  64,
	}

	result := mgr.Execute(context.Background(), req)
	if result.Status != executor.StatusUnsupportedLanguage {
		t.Fatalf("expected status UNSUPPORTED_LANGUAGE, got: %s", result.Status)
	}
}
