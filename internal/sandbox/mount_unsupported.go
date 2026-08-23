//go:build !linux

package sandbox

import "errors"

type UnsupportedMountManager struct{}

func NewMountManager(cfg MountConfig) MountManager {
	return &UnsupportedMountManager{}
}

func (u *UnsupportedMountManager) Setup(tracker *CleanupTracker) error {
	return errors.New("mount isolation is only supported on Linux")
}

func (u *UnsupportedMountManager) SwitchRoot() error {
	return errors.New("root switching is only supported on Linux")
}
