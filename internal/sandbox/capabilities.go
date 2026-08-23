package sandbox

// CapabilityConfig defines security capability flags for the sandboxed process.
type CapabilityConfig struct {
	NoNewPrivs  bool
	DropAllCaps bool
	UID         int
	GID         int
}

// DefaultCapabilityConfig returns strict capability restrictions.
func DefaultCapabilityConfig() CapabilityConfig {
	return CapabilityConfig{
		NoNewPrivs:  true,
		DropAllCaps: true,
		UID:         1000,
		GID:         1000,
	}
}
