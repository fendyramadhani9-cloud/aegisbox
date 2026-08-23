//go:build linux

package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
	"golang.org/x/sys/unix"
)

func (p *ProcessManager) executeOS(ctx context.Context, cfg *ExecutionConfig) (*executor.ExecutionResult, *ResourceUsage, error) {
	tracker := NewCleanupTracker()
	defer tracker.CleanupAll()

	startTime := time.Now()

	// 1. Setup Cgroups v2 Manager
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

	// 2. Prepare Command & Payload
	if len(cfg.Command) == 0 {
		return nil, nil, errors.New("empty command specified for execution")
	}

	execCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	payload := ChildPayload{
		Command:          cfg.Command,
		WorkspaceDir:     cfg.WorkspaceDir,
		RootfsDir:        cfg.RootfsDir,
		EnableNamespaces: cfg.EnableNamespaces,
		DisableNetwork:   cfg.DisableNetwork,
		EnableSeccomp:    cfg.EnableSeccomp,
		UID:              cfg.UID,
		GID:              cfg.GID,
		Seccomp:          DefaultSeccompPolicy(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize child payload: %w", err)
	}

	// 3. Create synchronization pipe to eliminate container startup race condition
	rPipe, wPipe, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sync pipe: %w", err)
	}
	defer rPipe.Close()
	defer wPipe.Close()

	var stdoutBuf, stderrBuf bytes.Buffer

	// 4. Resolve binary for execution (aegisbox __init__ or direct command)
	selfExe, exeErr := os.Executable()
	isTestBinary := (exeErr == nil && (strings.HasSuffix(selfExe, ".test") || strings.Contains(selfExe, "go-build")))

	var launcherBinary string
	if !isTestBinary && exeErr == nil && selfExe != "" {
		launcherBinary = selfExe
	} else if binCandidate, err := filepath.Abs("bin/aegisbox"); err == nil {
		if _, statErr := os.Stat(binCandidate); statErr == nil {
			launcherBinary = binCandidate
		}
	}

	var cmd *exec.Cmd
	useReExec := false

	if launcherBinary != "" {
		cmd = exec.CommandContext(execCtx, launcherBinary, "__init__")
		cmd.Stdin = bytes.NewReader(payloadBytes)
		cmd.ExtraFiles = []*os.File{rPipe} // Exposed as fd 3 in child process
		useReExec = true
	} else {
		cmd = exec.CommandContext(execCtx, cfg.Command[0], cfg.Command[1:]...)
	}

	cmd.Dir = cfg.WorkspaceDir
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// 5. Configure Namespaces
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

	// 6. Start container launcher child process
	startErr := cmd.Start()
	if startErr != nil {
		// If unprivileged user namespace clone failed, fallback to direct execution without new namespaces
		if cfg.EnableNamespaces {
			cmd = exec.CommandContext(execCtx, cfg.Command[0], cfg.Command[1:]...)
			cmd.Dir = cfg.WorkspaceDir
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			if retryErr := cmd.Start(); retryErr != nil {
				return nil, nil, fmt.Errorf("failed to start process (fallback): %w", retryErr)
			}
			useReExec = false
		} else {
			return nil, nil, fmt.Errorf("failed to start process: %w", startErr)
		}
	}

	childPID := cmd.Process.Pid

	// 7. Attach child process into cgroup BEFORE user code executes
	if cgMgr != nil && childPID > 0 {
		_ = cgMgr.AttachProcess(childPID)
	}

	// 8. Release child process from synchronization lock
	if useReExec {
		_, _ = wPipe.Write([]byte{1})
		_ = wPipe.Close()
	}

	// 9. Wait for process completion
	waitErr := cmd.Wait()
	execDuration := time.Since(startTime)

	// 10. Collect telemetry from Cgroup
	var usage *ResourceUsage
	if cgMgr != nil {
		usage, _ = cgMgr.CollectMetrics()
	}
	if usage == nil {
		usage = &ResourceUsage{}
	}
	usage.ExecutionTimeMs = execDuration.Milliseconds()

	// 11. Format initial result
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
