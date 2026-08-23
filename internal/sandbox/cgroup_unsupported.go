//go:build !linux

package sandbox

import "errors"

type UnsupportedCgroupManager struct{}

func NewCgroupManager(parentPath string, execID string) (CgroupManager, error) {
	return nil, errors.New("cgroups v2 is only supported on Linux")
}

func (u *UnsupportedCgroupManager) Path() string {
	return ""
}

func (u *UnsupportedCgroupManager) ApplyLimits(limits CgroupLimits) error {
	return errors.New("cgroups v2 is only supported on Linux")
}

func (u *UnsupportedCgroupManager) AttachProcess(pid int) error {
	return errors.New("cgroups v2 is only supported on Linux")
}

func (u *UnsupportedCgroupManager) CollectMetrics() (*ResourceUsage, error) {
	return &ResourceUsage{}, nil
}

func (u *UnsupportedCgroupManager) KillAll() error {
	return nil
}

func (u *UnsupportedCgroupManager) Destroy() error {
	return nil
}
