//go:build linux

package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// ChildPayload carries container setup configuration to the spawned child process.
type ChildPayload struct {
	Command          []string      `json:"command"`
	WorkspaceDir     string        `json:"workspace_dir"`
	RootfsDir        string        `json:"rootfs_dir"`
	EnableNamespaces bool          `json:"enable_namespaces"`
	DisableNetwork   bool          `json:"disable_network"`
	EnableSeccomp    bool          `json:"enable_seccomp"`
	UID              int           `json:"uid"`
	GID              int           `json:"gid"`
	Seccomp          SeccompPolicy `json:"seccomp"`
}

// RunChildInit executes the in-container bootstrap sequence before running user code.
func RunChildInit() {
	// File descriptor 3 is the synchronization pipe from the parent
	syncPipe := os.NewFile(uintptr(3), "sync_pipe")
	if syncPipe == nil {
		fmt.Fprintf(os.Stderr, "aegisbox __init__: missing synchronization pipe fd 3\n")
		os.Exit(1)
	}

	// 1. Read container configuration from stdin
	var payload ChildPayload
	if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
		fmt.Fprintf(os.Stderr, "aegisbox __init__: failed to decode payload: %v\n", err)
		os.Exit(1)
	}

	// 2. Wait for parent to attach PID to cgroup and configure limits
	var syncBuf [1]byte
	if _, err := syncPipe.Read(syncBuf[:]); err != nil {
		fmt.Fprintf(os.Stderr, "aegisbox __init__: parent sync pipe failed: %v\n", err)
		os.Exit(1)
	}
	_ = syncPipe.Close()

	// 3. Configure Mount Namespace if enabled and template rootfs exists
	if payload.EnableNamespaces && payload.RootfsDir != "" {
		if _, err := os.Stat(payload.RootfsDir); err == nil {
			mountCfg := MountConfig{
				RootfsTemplateDir: payload.RootfsDir,
				InstanceRootDir:   payload.RootfsDir,
				WorkspaceHostDir:  payload.WorkspaceDir,
				ReadOnlyRootfs:    true,
				MountProc:         true,
				MountTmpfsTmp:     true,
				MountWorkspace:    true,
				UsePivotRoot:      false, // Fallback chroot inside user namespace
			}
			mountMgr := NewMountManager(mountCfg)
			_ = mountMgr.Setup(nil)
			_ = mountMgr.SwitchRoot()
		}
	}

	// 4. Set working directory to workspace
	if payload.WorkspaceDir != "" {
		if _, err := os.Stat(payload.WorkspaceDir); err == nil {
			_ = os.Chdir(payload.WorkspaceDir)
		} else if _, err := os.Stat("/workspace"); err == nil {
			_ = os.Chdir("/workspace")
		}
	}

	// 5. Apply Security Capabilities and PR_SET_NO_NEW_PRIVS
	capCfg := DefaultCapabilityConfig()
	capCfg.UID = payload.UID
	capCfg.GID = payload.GID
	if err := ApplyLinuxSecurityPrivileges(capCfg); err != nil {
		fmt.Fprintf(os.Stderr, "aegisbox __init__: failed to drop capabilities: %v\n", err)
		os.Exit(1)
	}

	// 6. Apply Seccomp BPF Syscall Filter before executing user code
	if payload.EnableSeccomp {
		if err := ApplySeccompFilter(payload.Seccomp); err != nil {
			// Non-fatal if kernel doesn't support seccomp filter in unprivileged context
		}
	}

	// 7. Resolve target executable
	if len(payload.Command) == 0 {
		fmt.Fprintf(os.Stderr, "aegisbox __init__: empty command\n")
		os.Exit(1)
	}

	binPath, err := exec.LookPath(payload.Command[0])
	if err != nil {
		binPath = payload.Command[0]
	}

	// 8. Execve: replace container launcher process with target executable
	env := []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"LANG=C.UTF-8",
		"PYTHONUNBUFFERED=1",
	}

	if err := syscall.Exec(binPath, payload.Command, env); err != nil {
		fmt.Fprintf(os.Stderr, "aegisbox __init__: execve %s failed: %v\n", binPath, err)
		os.Exit(1)
	}
}
