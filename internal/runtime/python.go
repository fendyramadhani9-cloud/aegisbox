package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aegisbox/aegisbox/internal/executor"
)

// PythonRuntime provides code execution support for Python 3 scripts.
type PythonRuntime struct{}

// NewPythonRuntime creates an instance of the Python 3 runtime adapter.
func NewPythonRuntime() *PythonRuntime {
	return &PythonRuntime{}
}

// Name returns "python".
func (p *PythonRuntime) Name() string {
	return "python"
}

// Validate checks request validity for Python execution.
func (p *PythonRuntime) Validate(req *executor.ExecutionRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}
	if len(req.Code) == 0 {
		return errors.New("python script code cannot be empty")
	}
	return nil
}

// Prepare writes the Python source code to the ephemeral workspace directory.
func (p *PythonRuntime) Prepare(ctx context.Context, workspaceDir string, req *executor.ExecutionRequest) error {
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	scriptPath := filepath.Join(workspaceDir, "main.py")
	if err := os.WriteFile(scriptPath, []byte(req.Code), 0644); err != nil {
		return fmt.Errorf("failed to write python script to workspace: %w", err)
	}

	return nil
}

// Command returns the execution command inside the workspace.
// Using "main.py" relative to working directory works seamlessly both within
// containerized rootfs (/workspace) and unprivileged workspace directories.
func (p *PythonRuntime) Command(req *executor.ExecutionRequest) ([]string, error) {
	cmd := []string{"python3", "-u", "main.py"}
	if len(req.Args) > 0 {
		cmd = append(cmd, req.Args...)
	}
	return cmd, nil
}

// Cleanup performs any language-specific workspace cleanup.
func (p *PythonRuntime) Cleanup(workspaceDir string) error {
	return nil
}
