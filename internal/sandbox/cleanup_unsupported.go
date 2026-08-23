//go:build !linux

package sandbox

import (
	"os"
)

func unmountPath(target string) error {
	return nil
}

func removeCgroupDir(cgroupPath string) error {
	return os.RemoveAll(cgroupPath)
}
