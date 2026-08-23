package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_FilesystemIsolation(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// Script attempting to write to /etc/evil or /root
	req := &executor.ExecutionRequest{
		Language: "python",
		Code: `
try:
    with open("/etc/evil_probe.txt", "w") as f:
        f.write("injected")
    print("WRITE_SUCCEEDED")
except Exception as e:
    print("WRITE_DENIED:", type(e).__name__)
`,
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if strings.Contains(res.Stdout, "WRITE_SUCCEEDED") {
		t.Fatalf("security violation: sandboxed process successfully wrote to /etc")
	}
}
