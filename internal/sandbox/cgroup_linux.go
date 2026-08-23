//go:build linux

package sandbox

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LinuxCgroupManager implements CgroupManager for Linux cgroups v2.
type LinuxCgroupManager struct {
	parentPath string
	cgroupPath string
	execID     string
}

// NewCgroupManager creates a cgroups v2 manager for the given execution ID.
func NewCgroupManager(parentPath string, execID string) (CgroupManager, error) {
	if parentPath == "" {
		parentPath = "/sys/fs/cgroup/aegisbox"
	}

	// Ensure parent cgroup directory exists and has controllers enabled
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent cgroup directory %s: %w", parentPath, err)
	}

	// Enable controllers in parent subtree if not yet enabled
	subtreeFile := filepath.Join(parentPath, "cgroup.subtree_control")
	if _, err := os.Stat(subtreeFile); err == nil {
		_ = os.WriteFile(subtreeFile, []byte("+cpu +memory +pids"), 0644)
	}

	cgroupPath := filepath.Join(parentPath, execID)
	if err := os.MkdirAll(cgroupPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create execution cgroup %s: %w", cgroupPath, err)
	}

	return &LinuxCgroupManager{
		parentPath: parentPath,
		cgroupPath: cgroupPath,
		execID:     execID,
	}, nil
}

func (c *LinuxCgroupManager) Path() string {
	return c.cgroupPath
}

func (c *LinuxCgroupManager) ApplyLimits(limits CgroupLimits) error {
	// 1. Memory limit (memory.max)
	if limits.MemoryMaxBytes > 0 {
		memFile := filepath.Join(c.cgroupPath, "memory.max")
		if err := os.WriteFile(memFile, []byte(strconv.FormatInt(limits.MemoryMaxBytes, 10)), 0644); err != nil {
			return fmt.Errorf("failed to write memory.max: %w", err)
		}
	}

	// 2. Memory swap limit (memory.swap.max) - if supported
	if limits.MemorySwapMaxBytes >= 0 {
		swapFile := filepath.Join(c.cgroupPath, "memory.swap.max")
		if _, err := os.Stat(swapFile); err == nil {
			_ = os.WriteFile(swapFile, []byte(strconv.FormatInt(limits.MemorySwapMaxBytes, 10)), 0644)
		}
	}

	// 3. CPU limit (cpu.max) e.g., "50000 100000"
	if limits.CPUQuotaUSec > 0 && limits.CPUPeriodUSec > 0 {
		cpuFile := filepath.Join(c.cgroupPath, "cpu.max")
		quotaStr := fmt.Sprintf("%d %d", limits.CPUQuotaUSec, limits.CPUPeriodUSec)
		if err := os.WriteFile(cpuFile, []byte(quotaStr), 0644); err != nil {
			return fmt.Errorf("failed to write cpu.max: %w", err)
		}
	}

	// 4. Process limit (pids.max)
	if limits.PIDsMax > 0 {
		pidsFile := filepath.Join(c.cgroupPath, "pids.max")
		if err := os.WriteFile(pidsFile, []byte(strconv.FormatInt(limits.PIDsMax, 10)), 0644); err != nil {
			return fmt.Errorf("failed to write pids.max: %w", err)
		}
	}

	return nil
}

func (c *LinuxCgroupManager) AttachProcess(pid int) error {
	procsFile := filepath.Join(c.cgroupPath, "cgroup.procs")
	if err := os.WriteFile(procsFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to attach PID %d to cgroup %s: %w", pid, c.cgroupPath, err)
	}
	return nil
}

func (c *LinuxCgroupManager) CollectMetrics() (*ResourceUsage, error) {
	metrics := &ResourceUsage{}

	// Read memory peak or current
	peakFile := filepath.Join(c.cgroupPath, "memory.peak")
	if data, err := os.ReadFile(peakFile); err == nil {
		if bytes, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			metrics.MemoryPeakBytes = bytes
		}
	} else {
		currFile := filepath.Join(c.cgroupPath, "memory.current")
		if data, err := os.ReadFile(currFile); err == nil {
			if bytes, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
				metrics.MemoryPeakBytes = bytes
			}
		}
	}

	// Read memory.events for OOM kills
	eventsFile := filepath.Join(c.cgroupPath, "memory.events")
	if file, err := os.Open(eventsFile); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) == 2 {
				val, _ := strconv.ParseInt(fields[1], 10, 64)
				if fields[0] == "oom_kill" && val > 0 {
					metrics.OOMKilled = true
				} else if fields[0] == "oom" && val > 0 {
					metrics.OOMKilled = true
				}
			}
		}
	}

	// Read pids.events for pids max hit
	pidsEventsFile := filepath.Join(c.cgroupPath, "pids.events")
	if file, err := os.Open(pidsEventsFile); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) == 2 && fields[0] == "max" {
				val, _ := strconv.ParseInt(fields[1], 10, 64)
				if val > 0 {
					metrics.ProcessLimitHit = true
				}
			}
		}
	}

	// Read cpu.stat
	cpuStatFile := filepath.Join(c.cgroupPath, "cpu.stat")
	if file, err := os.Open(cpuStatFile); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) == 2 {
				val, _ := strconv.ParseInt(fields[1], 10, 64)
				if fields[0] == "user_usec" {
					metrics.CPUTimeUserMs = val / 1000
				} else if fields[0] == "system_usec" {
					metrics.CPUTimeSystemMs = val / 1000
				}
			}
		}
	}

	return metrics, nil
}

func (c *LinuxCgroupManager) KillAll() error {
	killFile := filepath.Join(c.cgroupPath, "cgroup.kill")
	if _, err := os.Stat(killFile); err == nil {
		_ = os.WriteFile(killFile, []byte("1"), 0644)
	}
	return nil
}

func (c *LinuxCgroupManager) Destroy() error {
	_ = c.KillAll()
	time.Sleep(5 * time.Millisecond)
	return os.Remove(c.cgroupPath)
}
