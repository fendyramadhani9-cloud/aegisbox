package executor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/logging"
)

// RuntimeAdapter minimal interface for runtime execution orchestration.
type RuntimeAdapter interface {
	Name() string
	Validate(req *ExecutionRequest) error
	Prepare(ctx context.Context, workspaceDir string, req *ExecutionRequest) error
	Command(req *ExecutionRequest) ([]string, error)
	Cleanup(workspaceDir string) error
}

// RuntimeResolver retrieves runtime adapters by language.
type RuntimeResolver interface {
	Get(name string) (RuntimeAdapter, error)
	SupportedLanguages() []string
}

// SandboxExecutor minimal interface for sandbox execution.
type SandboxExecutor interface {
	Execute(ctx context.Context, execID string, cmd []string, workspaceDir string, timeout time.Duration, memMB int64, pidsMax int64) (*ExecutionResult, int64, bool, bool, error)
}

// ExecutionManager orchestrates the complete execution lifecycle.
type ExecutionManager struct {
	cfg      *config.Config
	resolver RuntimeResolver
	sandbox  SandboxExecutor
	logger   *logging.Logger
}

// NewExecutionManager creates a new ExecutionManager.
func NewExecutionManager(cfg *config.Config, resolver RuntimeResolver, sandbox SandboxExecutor, logger *logging.Logger) *ExecutionManager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	if logger == nil {
		logger = logging.Default()
	}
	return &ExecutionManager{
		cfg:      cfg,
		resolver: resolver,
		sandbox:  sandbox,
		logger:   logger,
	}
}

// GenerateExecutionID creates a cryptographically unique execution identifier.
func GenerateExecutionID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("exec-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("exec-%s", hex.EncodeToString(bytes))
}

// Execute processes an execution request through the sandbox lifecycle.
func (m *ExecutionManager) Execute(ctx context.Context, req *ExecutionRequest) *ExecutionResult {
	execID := GenerateExecutionID()
	startTime := time.Now()

	m.logger.Lifecycle("execution.created", execID, req.Language, "received execution request", map[string]interface{}{
		"timeout_ms": req.TimeoutMs,
		"max_mem_mb": req.MaxMemMB,
	})

	// 1. Validate request
	if err := req.Validate(); err != nil {
		res := &ExecutionResult{
			ExecutionID:  execID,
			Status:       StatusSandboxError,
			ExitCode:     -1,
			ErrorMessage: fmt.Sprintf("invalid request: %v", err),
		}
		m.logger.Lifecycle("execution.classified", execID, req.Language, "validation failed", map[string]interface{}{
			"error": err.Error(),
		})
		return res
	}

	// 2. Resolve language runtime adapter
	rt, err := m.resolver.Get(req.Language)
	if err != nil {
		res := &ExecutionResult{
			ExecutionID:  execID,
			Status:       StatusUnsupportedLanguage,
			ExitCode:     -1,
			ErrorMessage: fmt.Sprintf("unsupported language '%s': %v", req.Language, err),
		}
		m.logger.Lifecycle("execution.classified", execID, req.Language, "runtime not found", map[string]interface{}{
			"error": err.Error(),
		})
		return res
	}

	// 3. Create ephemeral workspace directory
	workspaceDir := filepath.Join(m.cfg.Sandbox.WorkspaceBaseDir, execID)
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		res := &ExecutionResult{
			ExecutionID:  execID,
			Status:       StatusStartError,
			ExitCode:     -1,
			ErrorMessage: fmt.Sprintf("failed to create workspace: %v", err),
		}
		return res
	}
	defer func() {
		_ = os.RemoveAll(workspaceDir)
		_ = rt.Cleanup(workspaceDir)
		m.logger.Lifecycle("sandbox.cleaned", execID, req.Language, "workspace cleaned", nil)
	}()

	m.logger.Lifecycle("sandbox.created", execID, req.Language, "workspace initialized", map[string]interface{}{
		"workspace_dir": workspaceDir,
	})

	// 4. Prepare code inside workspace
	if err := rt.Prepare(ctx, workspaceDir, req); err != nil {
		res := &ExecutionResult{
			ExecutionID:  execID,
			Status:       StatusStartError,
			ExitCode:     -1,
			ErrorMessage: fmt.Sprintf("runtime preparation failed: %v", err),
		}
		return res
	}

	// 5. Generate runtime execution command
	cmd, err := rt.Command(req)
	if err != nil {
		res := &ExecutionResult{
			ExecutionID:  execID,
			Status:       StatusStartError,
			ExitCode:     -1,
			ErrorMessage: fmt.Sprintf("command generation failed: %v", err),
		}
		return res
	}

	pidsLimit := req.MaxProcesses
	if pidsLimit <= 0 {
		pidsLimit = m.cfg.Sandbox.DefaultProcesses
	}

	m.logger.Lifecycle("process.started", execID, req.Language, "spawning sandboxed process", map[string]interface{}{
		"command": cmd,
	})

	// 6. Execute in sandbox
	rawRes, memPeak, oomHit, pidsHit, execErr := m.sandbox.Execute(
		ctx,
		execID,
		cmd,
		workspaceDir,
		req.TimeoutDuration(),
		req.MaxMemMB,
		pidsLimit,
	)

	m.logger.Lifecycle("process.finished", execID, req.Language, "process terminated", map[string]interface{}{
		"duration_ms": time.Since(startTime).Milliseconds(),
		"oom_hit":     oomHit,
		"pids_hit":    pidsHit,
	})

	// 7. Classify result
	timeoutHit := (ctx.Err() == context.DeadlineExceeded)
	classified := ClassifyResult(rawRes, memPeak, oomHit, pidsHit, timeoutHit, execErr)
	classified.ExecutionID = execID
	if classified.ExecutionTimeMs == 0 {
		classified.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	}

	m.logger.Lifecycle("execution.classified", execID, req.Language, "result classified", map[string]interface{}{
		"status":    classified.Status,
		"exit_code": classified.ExitCode,
	})

	return classified
}
