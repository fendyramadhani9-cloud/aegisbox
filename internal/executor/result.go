package executor

// ClassifyResult determines the definitive ExecutionStatus using kernel metrics and process exit state.
func ClassifyResult(res *ExecutionResult, usageMemBytes int64, oomKilled bool, pidsLimitHit bool, timeoutHit bool, execErr error) *ExecutionResult {
	if res == nil {
		res = &ExecutionResult{
			Status:       StatusSandboxError,
			ExitCode:     -1,
			ErrorMessage: "null execution result generated",
		}
		return res
	}

	// 1. Check if kernel cgroup reported OOM kill
	if oomKilled {
		res.Status = StatusOOMKilled
		res.ExitCode = 137 // Standard 128 + 9 (SIGKILL from OOM)
		res.ErrorMessage = "process was killed by cgroups v2 OOM killer"
		return res
	}

	// 2. Check if process creation hit pids.max
	if pidsLimitHit {
		res.Status = StatusProcessLimitExceeded
		res.ExitCode = -1
		res.ErrorMessage = "process creation limit exceeded (pids.max)"
		return res
	}

	// 3. Check if wall-clock timeout exceeded
	if timeoutHit || res.Status == StatusTimeLimitExceeded {
		res.Status = StatusTimeLimitExceeded
		res.ExitCode = -1
		res.ErrorMessage = "execution exceeded wall-clock timeout limit"
		return res
	}

	// 4. Check for unhandled execution errors
	if execErr != nil {
		res.Status = StatusSandboxError
		res.ErrorMessage = execErr.Error()
		return res
	}

	// 5. Evaluate process exit code
	if res.ExitCode == 0 {
		res.Status = StatusCompleted
	} else {
		res.Status = StatusRuntimeError
	}

	return res
}
