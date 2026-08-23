package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_PIDIsolation(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// Python script inspecting its own PID and process tree
	req := &executor.ExecutionRequest{
		Language: "python",
		Code: `
import os
pid = os.getpid()
print(f"MY_PID:{pid}")
`,
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if res.Status != executor.StatusCompleted {
		t.Fatalf("expected COMPLETED, got %s", res.Status)
	}

	if !strings.Contains(res.Stdout, "MY_PID:") {
		t.Errorf("expected PID output in stdout, got: %s", res.Stdout)
	}
}
