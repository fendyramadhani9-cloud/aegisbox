//go:build !linux

package sandbox

import "os"

func RunChildInit() {
	os.Exit(0)
}
