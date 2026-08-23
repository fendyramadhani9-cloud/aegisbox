package tests

import (
	"context"
	"testing"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_Timeout(t *testing.T) {
	mgr := newTestExecutionManager(t)

	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "import time; time.sleep(5)",
		TimeoutMs: 500, // 500ms timeout
		MaxMemMB:  64,
	}

	start := time.Now()
	res := mgr.Execute(context.Background(), req)
	elapsed := time.Since(start)

	if res.Status != executor.StatusTimeLimitExceeded {
		t.Fatalf("expected TIME_LIMIT_EXCEEDED, got: %s (stdout: %s, stderr: %s)", res.Status, res.Stdout, res.Stderr)
	}

	// Verify the execution was terminated promptly near the timeout limit
	if elapsed > 3*time.Second {
		t.Fatalf("execution took too long to terminate: %v", elapsed)
	}
}
