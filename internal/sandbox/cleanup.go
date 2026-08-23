package sandbox

import (
	"fmt"
	"os"
	"sync"
)

// CleanupTracker tracks resources and mounts for guaranteed post-execution teardown.
type CleanupTracker struct {
	mu        sync.Mutex
	unmounts  []string
	cgroups   []string
	dirs      []string
	customFns []func() error
}

// NewCleanupTracker creates a new cleanup manager.
func NewCleanupTracker() *CleanupTracker {
	return &CleanupTracker{
		unmounts:  make([]string, 0),
		cgroups:   make([]string, 0),
		dirs:      make([]string, 0),
		customFns: make([]func() error, 0),
	}
}

// TrackUnmount records a mount path to unmount during cleanup.
func (c *CleanupTracker) TrackUnmount(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unmounts = append(c.unmounts, path)
}

// TrackCgroup records a cgroup path to delete during cleanup.
func (c *CleanupTracker) TrackCgroup(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cgroups = append(c.cgroups, path)
}

// TrackDir records a directory to remove recursively during cleanup.
func (c *CleanupTracker) TrackDir(dir string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirs = append(c.dirs, dir)
}

// TrackFunc records a custom cleanup function.
func (c *CleanupTracker) TrackFunc(fn func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.customFns = append(c.customFns, fn)
}

// CleanupAll executes all tracked cleanup operations in reverse registration order.
func (c *CleanupTracker) CleanupAll() []error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error

	// Execute custom teardown functions in reverse
	for i := len(c.customFns) - 1; i >= 0; i-- {
		if err := c.customFns[i](); err != nil {
			errs = append(errs, fmt.Errorf("custom cleanup failed: %w", err))
		}
	}

	// Perform unmounts in reverse order
	for i := len(c.unmounts) - 1; i >= 0; i-- {
		path := c.unmounts[i]
		if err := unmountPath(path); err != nil {
			errs = append(errs, fmt.Errorf("unmount failed for %s: %w", path, err))
		}
	}

	// Delete cgroups
	for i := len(c.cgroups) - 1; i >= 0; i-- {
		cg := c.cgroups[i]
		if err := removeCgroupDir(cg); err != nil {
			errs = append(errs, fmt.Errorf("cgroup cleanup failed for %s: %w", cg, err))
		}
	}

	// Remove temporary directories
	for i := len(c.dirs) - 1; i >= 0; i-- {
		dir := c.dirs[i]
		if err := os.RemoveAll(dir); err != nil {
			errs = append(errs, fmt.Errorf("dir removal failed for %s: %w", dir, err))
		}
	}

	// Reset tracked lists
	c.unmounts = nil
	c.cgroups = nil
	c.dirs = nil
	c.customFns = nil

	return errs
}
