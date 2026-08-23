package runtime_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/runtime"
)

func TestRegistry_Lookup(t *testing.T) {
	reg := runtime.DefaultRegistry()

	rt, err := reg.Get("python")
	if err != nil {
		t.Fatalf("expected python runtime to be registered, got error: %v", err)
	}

	if rt.Name() != "python" {
		t.Errorf("expected runtime name 'python', got '%s'", rt.Name())
	}

	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Fatalf("expected error for nonexistent runtime, got nil")
	}
}

func TestPythonRuntime_PrepareAndCommand(t *testing.T) {
	rt := runtime.NewPythonRuntime()
	req := &executor.ExecutionRequest{
		Language:  "python",
		Code:      "print('Hello, AegisBox Runtime!')",
		TimeoutMs: 1000,
		MaxMemMB:  64,
		Args:      []string{"arg1", "arg2"},
	}

	if err := rt.Validate(req); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "aegisbox-runtime-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := rt.Prepare(context.Background(), tmpDir, req); err != nil {
		t.Fatalf("failed to prepare workspace: %v", err)
	}

	scriptContent, err := os.ReadFile(filepath.Join(tmpDir, "main.py"))
	if err != nil {
		t.Fatalf("failed to read prepared script: %v", err)
	}

	if string(scriptContent) != req.Code {
		t.Errorf("expected script content '%s', got '%s'", req.Code, string(scriptContent))
	}

	cmd, err := rt.Command(req)
	if err != nil {
		t.Fatalf("failed to generate command: %v", err)
	}

	expectedLen := 5 // python3, -u, main.py, arg1, arg2
	if len(cmd) != expectedLen {
		t.Fatalf("expected command length %d, got %d: %v", expectedLen, len(cmd), cmd)
	}
	if cmd[0] != "python3" || cmd[1] != "-u" || cmd[2] != "main.py" || cmd[3] != "arg1" || cmd[4] != "arg2" {
		t.Errorf("unexpected command structure: %v", cmd)
	}
}
