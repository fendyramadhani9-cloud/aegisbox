package sandbox

// CgroupLimits defines resource constraints enforced by Linux cgroups v2.
type CgroupLimits struct {
	MemoryMaxBytes     int64
	MemorySwapMaxBytes int64
	CPUQuotaUSec       int64
	CPUPeriodUSec      int64
	PIDsMax            int64
}

// CgroupManager manages lifecycle, limits, and metrics for execution cgroups.
type CgroupManager interface {
	Path() string
	ApplyLimits(limits CgroupLimits) error
	AttachProcess(pid int) error
	CollectMetrics() (*ResourceUsage, error)
	KillAll() error
	Destroy() error
}
