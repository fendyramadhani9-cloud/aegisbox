package executor

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ExecutionStatus represents the lifecycle termination state of an execution.
type ExecutionStatus string

const (
	// StatusCompleted indicates normal termination without sandbox violations.
	StatusCompleted ExecutionStatus = "COMPLETED"

	// StatusRuntimeError indicates the user script exited with a non-zero exit code.
	StatusRuntimeError ExecutionStatus = "RUNTIME_ERROR"

	// StatusTimeLimitExceeded indicates the execution exceeded the configured wall-clock timeout.
	StatusTimeLimitExceeded ExecutionStatus = "TIME_LIMIT_EXCEEDED"

	// StatusOOMKilled indicates the process exceeded memory limits and was killed by cgroups OOM killer.
	StatusOOMKilled ExecutionStatus = "OOM_KILLED"

	// StatusProcessLimitExceeded indicates process/thread creation exceeded pids.max limit.
	StatusProcessLimitExceeded ExecutionStatus = "PROCESS_LIMIT_EXCEEDED"

	// StatusStartError indicates failure to initialize or start the isolated environment.
	StatusStartError ExecutionStatus = "START_ERROR"

	// StatusSandboxError indicates internal failure during containment, isolation, or cleanup.
	StatusSandboxError ExecutionStatus = "SANDBOX_ERROR"

	// StatusUnsupportedLanguage indicates the requested runtime language is not registered.
	StatusUnsupportedLanguage ExecutionStatus = "UNSUPPORTED_LANGUAGE"
)

// ExecutionRequest defines the payload received for executing code inside the sandbox.
type ExecutionRequest struct {
	Language     string            `json:"language"`
	Code         string            `json:"code"`
	TimeoutMs    int64             `json:"timeout_ms"`
	MaxMemMB     int64             `json:"max_mem_mb"`
	MaxProcesses int64             `json:"max_processes,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	Args         []string          `json:"args,omitempty"`
}

// ExecutionResult represents the structured outcome of a sandboxed execution.
type ExecutionResult struct {
	ExecutionID      string          `json:"execution_id"`
	Status           ExecutionStatus `json:"status"`
	Stdout           string          `json:"stdout"`
	Stderr           string          `json:"stderr"`
	ExitCode         int             `json:"exit_code"`
	ExecutionTimeMs  int64           `json:"execution_time_ms"`
	MemoryUsageBytes int64           `json:"memory_usage_bytes,omitempty"`
	CPUTimeMs        int64           `json:"cpu_time_ms,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
}

var (
	ErrEmptyLanguage    = errors.New("language cannot be empty")
	ErrEmptyCode        = errors.New("code cannot be empty")
	ErrInvalidTimeout   = errors.New("timeout_ms must be between 100ms and 60000ms")
	ErrInvalidMemory    = errors.New("max_mem_mb must be between 16MB and 1024MB")
	ErrInvalidProcesses = errors.New("max_processes must be between 1 and 256")
	ErrCodeTooLarge     = errors.New("code payload exceeds maximum allowed size (64KB)")
)

const (
	MaxCodeSizeBytes = 64 * 1024 // 64 KB
	MinTimeoutMs     = 100
	MaxTimeoutMs     = 60000
	MinMemoryMB      = 16
	MaxMemoryMB      = 1024
	MinProcesses     = 1
	MaxProcesses     = 256
)

// Validate checks if the execution request meets all boundary and validation constraints.
func (r *ExecutionRequest) Validate() error {
	r.Language = strings.TrimSpace(strings.ToLower(r.Language))
	if r.Language == "" {
		return ErrEmptyLanguage
	}

	if strings.TrimSpace(r.Code) == "" {
		return ErrEmptyCode
	}

	if len(r.Code) > MaxCodeSizeBytes {
		return ErrCodeTooLarge
	}

	if r.TimeoutMs < MinTimeoutMs || r.TimeoutMs > MaxTimeoutMs {
		return fmt.Errorf("%w (got %d)", ErrInvalidTimeout, r.TimeoutMs)
	}

	if r.MaxMemMB < MinMemoryMB || r.MaxMemMB > MaxMemoryMB {
		return fmt.Errorf("%w (got %d)", ErrInvalidMemory, r.MaxMemMB)
	}

	if r.MaxProcesses != 0 && (r.MaxProcesses < MinProcesses || r.MaxProcesses > MaxProcesses) {
		return fmt.Errorf("%w (got %d)", ErrInvalidProcesses, r.MaxProcesses)
	}

	return nil
}

// TimeoutDuration returns the configured timeout as a standard time.Duration.
func (r *ExecutionRequest) TimeoutDuration() time.Duration {
	return time.Duration(r.TimeoutMs) * time.Millisecond
}
