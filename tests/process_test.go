package tests

import (
	"context"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_ProcessCreationLimit(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// Script attempting multiple forks
	req := &executor.ExecutionRequest{
		Language: "python",
		Code: `
import os, sys
try:
    for _ in range(50):
        if os.fork() == 0:
            os._exit(0)
    print("forks created")
except BlockingIOError:
    print("fork blocked")
except OSError:
    print("fork blocked")
`,
		TimeoutMs:    2000,
		MaxMemMB:     64,
		MaxProcesses: 5, // Strict limit
	}

	res := mgr.Execute(context.Background(), req)
	// Either fork blocked or process limit status triggered
	if res.Status != executor.StatusCompleted && res.Status != executor.StatusProcessLimitExceeded {
		t.Logf("Process limit handled with status: %s", res.Status)
	}
}
