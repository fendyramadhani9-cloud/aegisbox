package sandbox

import (
	"context"
	"errors"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
)

var (
	ErrSandboxUnsupported = errors.New("sandbox isolation is only supported on Linux")
	ErrExecutionTimeout   = errors.New("execution timed out")
	ErrOOMKilled          = errors.New("process terminated by cgroups OOM killer")
	ErrProcessLimit       = errors.New("process limit exceeded (fork bomb prevented)")
)

// ResourceUsage holds telemetry measured by the kernel during process execution.
type ResourceUsage struct {
	ExecutionTimeMs  int64 `json:"execution_time_ms"`
	MemoryPeakBytes  int64 `json:"memory_peak_bytes"`
	CPUTimeUserMs    int64 `json:"cpu_time_user_ms"`
	CPUTimeSystemMs  int64 `json:"cpu_time_system_ms"`
	OOMKilled        bool  `json:"oom_killed"`
	ProcessLimitHit  bool  `json:"process_limit_hit"`
	ProcessCountPeak int64 `json:"process_count_peak"`
}

// ExecutionConfig specifies container isolation parameters for a single execution.
type ExecutionConfig struct {
	ExecutionID      string
	Language         string
	Command          []string
	WorkspaceDir     string
	RootfsDir        string
	CgroupParent     string
	Timeout          time.Duration
	MaxMemoryMB      int64
	MaxProcesses     int64
	CPUQuotaPercent  int64 // e.g. 100 = 1 full CPU core, 50 = 50% core
	DisableNetwork   bool
	EnableSeccomp    bool
	EnableNamespaces bool
	UID              int
	GID              int
}

// Sandbox defines the isolation manager interface.
type Sandbox interface {
	// Execute orchestrates the isolated execution and returns output + resource usage.
	Execute(ctx context.Context, cfg *ExecutionConfig) (*executor.ExecutionResult, *ResourceUsage, error)
}
