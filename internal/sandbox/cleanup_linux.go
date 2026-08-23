//go:build linux

package sandbox

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func unmountPath(target string) error {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil
	}

	// Try lazy unmount with detach first to avoid hanging on open file descriptors
	err := unix.Unmount(target, unix.MNT_DETACH)
	if err != nil && err != unix.EINVAL && err != unix.ENOENT {
		return fmt.Errorf("unix.Unmount %s failed: %w", target, err)
	}
	return nil
}

func removeCgroupDir(cgroupPath string) error {
	if _, err := os.Stat(cgroupPath); os.IsNotExist(err) {
		return nil
	}

	// First kill any lingering processes inside this cgroup using cgroup.kill (Linux 5.14+)
	killPath := filepath.Join(cgroupPath, "cgroup.kill")
	if _, err := os.Stat(killPath); err == nil {
		_ = os.WriteFile(killPath, []byte("1"), 0644)
	}

	// Wait briefly for processes to terminate
	time.Sleep(10 * time.Millisecond)

	// If any processes remain in cgroup.procs, send SIGKILL to each
	procsPath := filepath.Join(cgroupPath, "cgroup.procs")
	if data, err := os.ReadFile(procsPath); err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var pid int
			if _, err := fmt.Sscanf(line, "%d", &pid); err == nil && pid > 0 {
				_ = unix.Kill(pid, unix.SIGKILL)
			}
		}
	}

	// Remove cgroup directory (rmdir is standard cgroup v2 removal mechanism)
	return os.Remove(cgroupPath)
}
