package sandbox

import (
	"context"

	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
)

// ProcessManager manages sandboxed execution lifecycle.
type ProcessManager struct {
	logger *logging.Logger
}

// NewProcessManager creates a new ProcessManager.
func NewProcessManager(logger *logging.Logger) *ProcessManager {
	if logger == nil {
		logger = logging.Default()
	}
	return &ProcessManager{logger: logger}
}

// Execute orchestrates the containerized execution on the target host.
func (p *ProcessManager) Execute(ctx context.Context, cfg *ExecutionConfig) (*executor.ExecutionResult, *ResourceUsage, error) {
	return p.executeOS(ctx, cfg)
}
