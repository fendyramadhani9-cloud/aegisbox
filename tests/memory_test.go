package tests

import (
	"context"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_MemoryAllocation(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// Normal small allocation should complete cleanly
	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "data = [i for i in range(100000)]; print(len(data))",
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if res.Status != executor.StatusCompleted {
		t.Fatalf("expected COMPLETED, got: %s (err: %s)", res.Status, res.ErrorMessage)
	}

	if res.Stdout != "100000\n" {
		t.Errorf("expected stdout '100000\\n', got '%s'", res.Stdout)
	}
}
