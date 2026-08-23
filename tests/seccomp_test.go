package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestExecution_SeccompBlockedSyscalls(t *testing.T) {
	mgr := newTestExecutionManager(t)

	// Python script attempting to invoke reboot/mount syscall directly via ctypes
	req := &executor.ExecutionRequest{
		Language: "python",
		Code: `
import ctypes
libc = ctypes.CDLL(None)
# SYS_reboot is syscall number 169 on x86_64
try:
    # Attempting to call reboot syscall (RB_AUTOBOOT)
    ret = libc.syscall(169, 0xfee1dead, 672274793, 0x1234567, 0)
    print(f"REBOOT_RESULT:{ret}")
except Exception as e:
    print(f"REBOOT_BLOCKED:{type(e).__name__}")
`,
		TimeoutMs: 2000,
		MaxMemMB:  64,
	}

	res := mgr.Execute(context.Background(), req)
	// Syscall should either fail with EPERM (-1 / errno 1) or be caught by seccomp
	if strings.Contains(res.Stdout, "REBOOT_RESULT:0") {
		t.Fatalf("security violation: reboot syscall succeeded inside sandbox")
	}
}
