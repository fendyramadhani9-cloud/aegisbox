//go:build linux

package sandbox

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// LinuxMountManager implements MountManager for Linux VFS namespaces.
type LinuxMountManager struct {
	cfg MountConfig
}

// NewMountManager initializes a mount manager.
func NewMountManager(cfg MountConfig) MountManager {
	return &LinuxMountManager{cfg: cfg}
}

// Setup creates mount points and isolates VFS.
func (m *LinuxMountManager) Setup(tracker *CleanupTracker) error {
	newRoot := m.cfg.InstanceRootDir

	// Ensure instance root directory exists
	if err := os.MkdirAll(newRoot, 0755); err != nil {
		return fmt.Errorf("failed to create instance root %s: %w", newRoot, err)
	}
	if tracker != nil {
		tracker.TrackDir(newRoot)
	}

	// 1. Make mount namespace private so mount changes don't leak to the host
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		// Non-fatal if already private
	}

	// 2. Bind mount the template rootfs into the instance root
	if m.cfg.RootfsTemplateDir != "" {
		if _, err := os.Stat(m.cfg.RootfsTemplateDir); err == nil {
			if err := unix.Mount(m.cfg.RootfsTemplateDir, newRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
				return fmt.Errorf("failed to bind mount template rootfs: %w", err)
			}
			if tracker != nil {
				tracker.TrackUnmount(newRoot)
			}

			// Remount as read-only if configured
			if m.cfg.ReadOnlyRootfs {
				if err := unix.Mount("", newRoot, "", unix.MS_BIND|unix.MS_REMOUNT|unix.MS_RDONLY|unix.MS_NOSUID, ""); err != nil {
					return fmt.Errorf("failed to remount rootfs as read-only: %w", err)
				}
			}
		}
	} else {
		// If no rootfs template provided, bind mount newRoot onto itself so it becomes a valid mountpoint
		if err := unix.Mount(newRoot, newRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			return fmt.Errorf("failed to bind mount instance root: %w", err)
		}
		if tracker != nil {
			tracker.TrackUnmount(newRoot)
		}
	}

	// 3. Mount isolated /proc
	if m.cfg.MountProc {
		procDir := filepath.Join(newRoot, "proc")
		_ = os.MkdirAll(procDir, 0755)
		if err := unix.Mount("proc", procDir, "proc", unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC, ""); err == nil {
			if tracker != nil {
				tracker.TrackUnmount(procDir)
			}
		}
	}

	// 4. Mount tmpfs on /tmp
	if m.cfg.MountTmpfsTmp {
		tmpDir := filepath.Join(newRoot, "tmp")
		_ = os.MkdirAll(tmpDir, 0777)
		if err := unix.Mount("tmpfs", tmpDir, "tmpfs", unix.MS_NOSUID|unix.MS_NODEV, "size=64m"); err == nil {
			if tracker != nil {
				tracker.TrackUnmount(tmpDir)
			}
		}
	}

	// 5. Mount writable workspace at /workspace
	if m.cfg.MountWorkspace && m.cfg.WorkspaceHostDir != "" {
		wsMount := filepath.Join(newRoot, "workspace")
		_ = os.MkdirAll(wsMount, 0755)
		if err := unix.Mount(m.cfg.WorkspaceHostDir, wsMount, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			return fmt.Errorf("failed to bind mount workspace to %s: %w", wsMount, err)
		}
		if tracker != nil {
			tracker.TrackUnmount(wsMount)
		}
	}

	// 6. Setup minimal /dev nodes
	m.setupMinimalDev(newRoot, tracker)

	return nil
}

func (m *LinuxMountManager) setupMinimalDev(newRoot string, tracker *CleanupTracker) {
	devDir := filepath.Join(newRoot, "dev")
	_ = os.MkdirAll(devDir, 0755)

	nodes := []string{"null", "zero", "urandom", "random"}
	for _, node := range nodes {
		hostPath := filepath.Join("/dev", node)
		targetPath := filepath.Join(devDir, node)
		if _, err := os.Stat(hostPath); err == nil {
			// Create empty file placeholder for bind mount
			if f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0666); err == nil {
				_ = f.Close()
				if err := unix.Mount(hostPath, targetPath, "", unix.MS_BIND, ""); err == nil {
					if tracker != nil {
						tracker.TrackUnmount(targetPath)
					}
				}
			}
		}
	}
}

// SwitchRoot switches root to the sandboxed filesystem using pivot_root or fallback chroot.
func (m *LinuxMountManager) SwitchRoot() error {
	newRoot := m.cfg.InstanceRootDir

	if m.cfg.UsePivotRoot {
		oldRoot := filepath.Join(newRoot, ".old_root")
		if err := os.MkdirAll(oldRoot, 0700); err != nil {
			return fmt.Errorf("failed to create pivot old_root %s: %w", oldRoot, err)
		}

		if err := unix.PivotRoot(newRoot, oldRoot); err != nil {
			// If pivot_root fails (e.g. rootfs on initramfs or unsupported fs), fall back to chroot
			if err := unix.Chroot(newRoot); err != nil {
				return fmt.Errorf("pivot_root failed and fallback chroot failed: %w", err)
			}
			if err := os.Chdir("/workspace"); err != nil {
				_ = os.Chdir("/")
			}
			return nil
		}

		if err := os.Chdir("/workspace"); err != nil {
			_ = os.Chdir("/")
		}

		// Unmount old root
		_ = unix.Unmount("/.old_root", unix.MNT_DETACH)
		_ = os.Remove("/.old_root")
		return nil
	}

	// Fallback chroot
	if err := unix.Chroot(newRoot); err != nil {
		return fmt.Errorf("chroot to %s failed: %w", newRoot, err)
	}
	if err := os.Chdir("/workspace"); err != nil {
		_ = os.Chdir("/")
	}

	return nil
}
