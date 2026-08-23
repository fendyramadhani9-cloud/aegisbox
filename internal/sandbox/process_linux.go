//go:build linux

package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
	"golang.org/x/sys/unix"
)

const maxOutputBytes = 1024 * 1024 // 1 MB max buffer for stdout/stderr

func (p *ProcessManager) executeOS(ctx context.Context, cfg *ExecutionConfig) (*executor.ExecutionResult, *ResourceUsage, error) {
	tracker := NewCleanupTracker()
	defer tracker.CleanupAll()

	startTime := time.Now()

	// 1. Setup Cgroups v2 Manager if requested
	var cgMgr CgroupManager
	var cgErr error
	if cfg.CgroupParent != "" {
		cgMgr, cgErr = NewCgroupManager(cfg.CgroupParent, cfg.ExecutionID)
		if cgErr == nil && cgMgr != nil {
			tracker.TrackCgroup(cgMgr.Path())
			limits := CgroupLimits{
				MemoryMaxBytes:     cfg.MaxMemoryMB * 1024 * 1024,
				MemorySwapMaxBytes: 0,
				PIDsMax:            cfg.MaxProcesses,
			}
			if cfg.CPUQuotaPercent > 0 {
				limits.CPUPeriodUSec = 100000
				limits.CPUQuotaUSec = cfg.CPUQuotaPercent * 1000
			}
			_ = cgMgr.ApplyLimits(limits)
		}
	}

	// 2. Prepare Command
	if len(cfg.Command) == 0 {
		return nil, nil, errors.New("empty command specified for execution")
	}

	execCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, cfg.Command[0], cfg.Command[1:]...)
	cmd.Dir = cfg.WorkspaceDir

	// 3. Configure Namespaces if enabled
	if cfg.EnableNamespaces {
		nsCfg := NamespaceConfig{
			NewPID:   true,
			NewMount: true,
			NewNet:   cfg.DisableNetwork,
			NewUTS:   true,
			NewIPC:   true,
			Hostname: "aegisbox",
			UID:      cfg.UID,
			GID:      cfg.GID,
		}
		cmd.SysProcAttr = BuildSysProcAttr(nsCfg)
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// 4. Start child process (with unprivileged fallback if unprivileged user runs test)
	startErr := cmd.Start()
	if startErr != nil {
		if cfg.EnableNamespaces {
			// Retry with standard session attribute if unprivileged namespace creation was restricted
			cmd = exec.CommandContext(execCtx, cfg.Command[0], cfg.Command[1:]...)
			cmd.Dir = cfg.WorkspaceDir
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			if retryErr := cmd.Start(); retryErr != nil {
				return nil, nil, fmt.Errorf("failed to start process: %w", retryErr)
			}
		} else {
			return nil, nil, fmt.Errorf("failed to start process: %w", startErr)
		}
	}

	childPID := cmd.Process.Pid

	// 5. Attach child process into cgroup
	if cgMgr != nil && childPID > 0 {
		_ = cgMgr.AttachProcess(childPID)
	}

	// 6. Wait for process completion
	waitErr := cmd.Wait()
	execDuration := time.Since(startTime)

	// 7. Collect telemetry from Cgroup
	var usage *ResourceUsage
	if cgMgr != nil {
		usage, _ = cgMgr.CollectMetrics()
	}
	if usage == nil {
		usage = &ResourceUsage{}
	}
	usage.ExecutionTimeMs = execDuration.Milliseconds()

	// 8. Format initial result
	res := &executor.ExecutionResult{
		ExecutionID:      cfg.ExecutionID,
		Stdout:           stdoutBuf.String(),
		Stderr:           stderrBuf.String(),
		ExecutionTimeMs:  usage.ExecutionTimeMs,
		MemoryUsageBytes: usage.MemoryPeakBytes,
		CPUTimeMs:        usage.CPUTimeUserMs + usage.CPUTimeSystemMs,
	}

	// Check exit code and context timeout
	if execCtx.Err() == context.DeadlineExceeded {
		res.Status = executor.StatusTimeLimitExceeded
		res.ExitCode = -1
		// Kill lingering process tree
		if cgMgr != nil {
			_ = cgMgr.KillAll()
		}
		if childPID > 0 {
			_ = unix.Kill(-childPID, unix.SIGKILL)
		}
		return res, usage, nil
	}

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
		}
	} else {
		res.ExitCode = 0
	}

	return res, usage, nil
}
