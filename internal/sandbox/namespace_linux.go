//go:build linux

package sandbox

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func buildSysProcAttrOS(cfg NamespaceConfig) *syscall.SysProcAttr {
	attr := &syscall.SysProcAttr{
		Cloneflags: unix.CLONE_NEWUSER,
		Setsid:     true, // Create new session to prevent tty signaling
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

	// Map container UID/GID 0 to host caller UID/GID for unprivileged user namespaces
	attr.UidMappings = []syscall.SysProcIDMap{
		{ContainerID: 0, HostID: hostUID, Size: 1},
	}
	attr.GidMappings = []syscall.SysProcIDMap{
		{ContainerID: 0, HostID: hostGID, Size: 1},
	}

	return attr
}
