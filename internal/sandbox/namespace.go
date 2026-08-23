package sandbox

import (
	"syscall"
)

// NamespaceConfig defines Linux namespace isolation flags.
type NamespaceConfig struct {
	NewPID   bool
	NewMount bool
	NewNet   bool
	NewUTS   bool
	NewIPC   bool
	NewUser  bool
	Hostname string
	UID      int
	GID      int
}

// DefaultNamespaceConfig returns standard namespace isolation settings.
func DefaultNamespaceConfig() NamespaceConfig {
	return NamespaceConfig{
		NewPID:   true,
		NewMount: true,
		NewNet:   true,
		NewUTS:   true,
		NewIPC:   true,
		Hostname: "aegisbox",
		UID:      1000,
		GID:      1000,
	}
}

// BuildSysProcAttr constructs platform-specific process attributes.
func BuildSysProcAttr(cfg NamespaceConfig) *syscall.SysProcAttr {
	return buildSysProcAttrOS(cfg)
}
