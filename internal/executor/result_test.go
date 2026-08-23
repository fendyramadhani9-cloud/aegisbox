package executor_test

import (
	"errors"
	"testing"

	"github.com/aegisbox/aegisbox/internal/executor"
)

func TestClassifyResult(t *testing.T) {
	tests := []struct {
		name         string
		initialRes   *executor.ExecutionResult
		memBytes     int64
		oomKilled    bool
		pidsLimitHit bool
		timeoutHit   bool
		execErr      error
		expectStatus executor.ExecutionStatus
		expectExit   int
	}{
		{
			name: "Normal completion (exit code 0)",
			initialRes: &executor.ExecutionResult{
				ExitCode: 0,
				Stdout:   "hello\n",
			},
			expectStatus: executor.StatusCompleted,
			expectExit:   0,
		},
		{
			name: "Script runtime error (exit code 1)",
			initialRes: &executor.ExecutionResult{
				ExitCode: 1,
				Stderr:   "Traceback...",
			},
			expectStatus: executor.StatusRuntimeError,
			expectExit:   1,
		},
		{
			name: "OOM Kill takes precedence",
			initialRes: &executor.ExecutionResult{
				ExitCode: 137,
			},
			oomKilled:    true,
			expectStatus: executor.StatusOOMKilled,
			expectExit:   137,
		},
		{
			name: "Process limit exceeded (pids.max)",
			initialRes: &executor.ExecutionResult{
				ExitCode: -1,
			},
			pidsLimitHit: true,
			expectStatus: executor.StatusProcessLimitExceeded,
			expectExit:   -1,
		},
		{
			name: "Wall-clock timeout exceeded",
			initialRes: &executor.ExecutionResult{
				ExitCode: -1,
			},
			timeoutHit:   true,
			expectStatus: executor.StatusTimeLimitExceeded,
			expectExit:   -1,
		},
		{
			name: "Sandbox internal execution error",
			initialRes: &executor.ExecutionResult{
				ExitCode: -1,
			},
			execErr:      errors.New("failed to setup mount namespace"),
			expectStatus: executor.StatusSandboxError,
			expectExit:   -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := executor.ClassifyResult(tc.initialRes, tc.memBytes, tc.oomKilled, tc.pidsLimitHit, tc.timeoutHit, tc.execErr)
			if res.Status != tc.expectStatus {
				t.Errorf("expected status %s, got %s", tc.expectStatus, res.Status)
			}
			if res.ExitCode != tc.expectExit {
				t.Errorf("expected exit code %d, got %d", tc.expectExit, res.ExitCode)
			}
		})
	}
}
