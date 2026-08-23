//go:build !linux

package sandbox

import "syscall"

func buildSysProcAttrOS(cfg NamespaceConfig) *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
