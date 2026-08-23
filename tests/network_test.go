package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_NetworkIsolation(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// Script attempting network socket connection to public IP (e.g. 1.1.1.1:80)
	req := &executor.ExecutionRequest{
		Language: "python",
		Code: `
import socket
try:
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.settimeout(0.5)
    s.connect(("1.1.1.1", 80))
    print("NETWORK_SUCCESS")
except Exception as e:
    print("NETWORK_BLOCKED:", type(e).__name__)
`,
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	if strings.Contains(res.Stdout, "NETWORK_SUCCESS") {
		t.Fatalf("security violation: network access succeeded inside sandbox")
	}
}
