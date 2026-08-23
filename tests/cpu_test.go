package tests

import (
	"context"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_CPUHeavyWorkload(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// CPU-intensive calculation that finishes within timeout
	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "total = sum(i * i for i in range(500000)); print(total > 0)",
		TimeoutMs: 3000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if res.Status != executor.StatusCompleted {
		t.Fatalf("expected COMPLETED, got: %s", res.Status)
	}

	if res.Stdout != "True\n" {
		t.Errorf("expected stdout 'True\\n', got '%s'", res.Stdout)
	}
}
