//go:build !linux

package sandbox

func ApplyLinuxSecurityPrivileges(cfg CapabilityConfig) error {
	return nil
}
