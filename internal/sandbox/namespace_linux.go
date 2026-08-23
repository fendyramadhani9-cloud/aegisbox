//go:build linux

package sandbox

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func buildSysProcAttrOS(cfg NamespaceConfig) *syscall.SysProcAttr {
	attr := &syscall.SysProcAttr{
		Setsid: true, // Create new session to prevent tty signaling
	}

	if cfg.NewPID {
		attr.Cloneflags |= unix.CLONE_NEWPID
	}
	if cfg.NewMount {
		attr.Cloneflags |= unix.CLONE_NEWNS
	}
	if cfg.NewNet {
		attr.Cloneflags |= unix.CLONE_NEWNET
	}
	if cfg.NewUTS {
		attr.Cloneflags |= unix.CLONE_NEWUTS
	}
	if cfg.NewIPC {
		attr.Cloneflags |= unix.CLONE_NEWIPC
	}

	hostUID := syscall.Getuid()
	hostGID := syscall.Getgid()

	targetUID := cfg.UID
	if targetUID <= 0 {
		targetUID = 1000
	}
	targetGID := cfg.GID
	if targetGID <= 0 {
		targetGID = 1000
	}

	if hostUID == 0 {
		// When host process runs as root (sudo / systemd daemon), drop child process credentials
		// to unprivileged user (UID/GID 1000) to ensure filesystem and capability isolation.
		attr.Credential = &syscall.Credential{
			Uid:         uint32(targetUID),
			Gid:         uint32(targetGID),
			NoSetGroups: true,
		}
	} else {
		// When host process runs as unprivileged user, use CLONE_NEWUSER to allow
		// unprivileged namespace creation (PID, Network, Mount) without root.
		attr.Cloneflags |= unix.CLONE_NEWUSER
		attr.UidMappings = []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: hostUID, Size: 1},
		}
		attr.GidMappings = []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: hostGID, Size: 1},
		}
	}

	return attr
}
