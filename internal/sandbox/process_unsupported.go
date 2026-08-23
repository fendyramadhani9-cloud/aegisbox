//go:build !linux

package sandbox

import (
	"context"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func (p *ProcessManager) executeOS(ctx context.Context, cfg *ExecutionConfig) (*executor.ExecutionResult, *ResourceUsage, error) {
	return nil, nil, ErrSandboxUnsupported
}
