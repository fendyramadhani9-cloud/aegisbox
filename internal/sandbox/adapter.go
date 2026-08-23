package sandbox

import (
	"context"
	"time"

	"github.com/aegisbox/aegisbox/internal/config"
	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
)

// SandboxAdapter adapts ProcessManager to executor.SandboxExecutor interface.
type SandboxAdapter struct {
	pm     *ProcessManager
	cfg    *config.Config
	logger *logging.Logger
}

// NewSandboxAdapter creates an adapter bridging sandbox isolation to the executor manager.
func NewSandboxAdapter(cfg *config.Config, logger *logging.Logger) *SandboxAdapter {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	if logger == nil {
		logger = logging.Default()
	}
	return &SandboxAdapter{
		pm:     NewProcessManager(logger),
		cfg:    cfg,
		logger: logger,
	}
}

// Execute fulfills executor.SandboxExecutor.
func (s *SandboxAdapter) Execute(
	ctx context.Context,
	execID string,
	cmd []string,
	workspaceDir string,
	timeout time.Duration,
	memMB int64,
	pidsMax int64,
) (*executor.ExecutionResult, int64, bool, bool, error) {
	execCfg := &ExecutionConfig{
		ExecutionID:      execID,
		Command:          cmd,
		WorkspaceDir:     workspaceDir,
		RootfsDir:        s.cfg.Sandbox.RootfsDir,
		CgroupParent:     s.cfg.Sandbox.CgroupParent,
		Timeout:          timeout,
		MaxMemoryMB:      memMB,
		MaxProcesses:     pidsMax,
		CPUQuotaPercent:  100,
		DisableNetwork:   true,
		EnableSeccomp:    true,
		EnableNamespaces: true,
		UID:              1000,
		GID:              1000,
	}

	res, usage, err := s.pm.Execute(ctx, execCfg)
	if err != nil {
		return res, 0, false, false, err
	}

	var memPeak int64
	var oomHit, pidsHit bool
	if usage != nil {
		memPeak = usage.MemoryPeakBytes
		oomHit = usage.OOMKilled
		pidsHit = usage.ProcessLimitHit
	}

	return res, memPeak, oomHit, pidsHit, nil
}
