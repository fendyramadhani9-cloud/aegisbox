package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aegisbox/aegisbox/internal/api"
	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
)

type mockRuntime struct{}

func (m *mockRuntime) Name() string                                  { return "python" }
func (m *mockRuntime) Validate(req *executor.ExecutionRequest) error { return nil }
func (m *mockRuntime) Prepare(ctx context.Context, workspaceDir string, req *executor.ExecutionRequest) error {
	return nil
}
func (m *mockRuntime) Command(req *executor.ExecutionRequest) ([]string, error) {
	return []string{"python3", "-c", "print(1)"}, nil
}
func (m *mockRuntime) Cleanup(workspaceDir string) error { return nil }

type mockResolver struct{}

func (m *mockResolver) Get(name string) (executor.RuntimeAdapter, error) {
	return &mockRuntime{}, nil
}

func (m *mockResolver) SupportedLanguages() []string {
	return []string{"python"}
}

type mockSandbox struct{}

func (s *mockSandbox) Execute(
	ctx context.Context,
	execID string,
	cmd []string,
	workspaceDir string,
	timeout time.Duration,
	memMB int64,
	pidsMax int64,
) (*executor.ExecutionResult, int64, bool, bool, error) {
	return &executor.ExecutionResult{
		ExecutionID: execID,
		Status:      executor.StatusCompleted,
		Stdout:      "mock output\n",
		ExitCode:    0,
	}, 1024, false, false, nil
}

func TestAPI_Health(t *testing.T) {
	cfg := config.DefaultConfig()
	resolver := &mockResolver{}
	sbx := &mockSandbox{}
	logger := logging.Default()
	mgr := executor.NewExecutionManager(cfg, resolver, sbx, logger)

	router := api.NewRouter(mgr, resolver, logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", rec.Code)
	}

	var health api.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&health); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}

	if health.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", health.Status)
	}
}

func TestAPI_Execute(t *testing.T) {
	cfg := config.DefaultConfig()
	resolver := &mockResolver{}
	sbx := &mockSandbox{}
	logger := logging.Default()
	mgr := executor.NewExecutionManager(cfg, resolver, sbx, logger)

	router := api.NewRouter(mgr, resolver, logger)

	payload := api.ExecuteRequestPayload{
		Language:  "python",
		Code:      "print('test')",
		TimeoutMs: 1000,
		MaxMemMB:  64,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got: %d (body: %s)", rec.Code, rec.Body.String())
	}

	var res executor.ExecutionResult
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode execute response: %v", err)
	}

	if res.Status != executor.StatusCompleted {
		t.Errorf("expected status COMPLETED, got '%s'", res.Status)
	}
	if res.Stdout != "mock output\n" {
		t.Errorf("expected stdout 'mock output\\n', got '%s'", res.Stdout)
	}
}
