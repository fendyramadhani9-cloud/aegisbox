//go:build linux

package sandbox

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// ApplyLinuxSecurityPrivileges applies PR_SET_NO_NEW_PRIVS and clears ambient capabilities.
func ApplyLinuxSecurityPrivileges(cfg CapabilityConfig) error {
	if cfg.NoNewPrivs {
		if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
			return fmt.Errorf("failed to set PR_SET_NO_NEW_PRIVS: %w", err)
		}
	}

	if cfg.DropAllCaps {
		// Clear ambient capabilities
		_ = unix.Prctl(unix.PR_CAP_AMBIENT, unix.PR_CAP_AMBIENT_CLEAR_ALL, 0, 0, 0)
	}

	return nil
}
