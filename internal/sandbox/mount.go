package sandbox

// MountConfig defines directory paths and mount options for the sandboxed filesystem.
type MountConfig struct {
	RootfsTemplateDir string
	InstanceRootDir   string
	WorkspaceHostDir  string
	ReadOnlyRootfs    bool
	MountProc         bool
	MountTmpfsTmp     bool
	MountWorkspace    bool
	UsePivotRoot      bool
}

// MountManager manages isolated VFS mount configurations and switches.
type MountManager interface {
	Setup(tracker *CleanupTracker) error
	SwitchRoot() error
}
