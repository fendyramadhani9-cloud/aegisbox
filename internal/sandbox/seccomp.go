package sandbox

// SeccompPolicy defines configuration parameters for Linux syscall filtering.
type SeccompPolicy struct {
	Enabled         bool
	DefaultAction   string // "ALLOW" or "ERRNO"
	BlockedSyscalls []string
}

// DefaultSeccompPolicy returns a hardened filter policy blocking dangerous system calls.
func DefaultSeccompPolicy() SeccompPolicy {
	return SeccompPolicy{
		Enabled:       true,
		DefaultAction: "ALLOW",
		BlockedSyscalls: []string{
			"mount", "umount2", "pivot_root",
			"reboot", "ptrace", "process_vm_readv", "process_vm_writev",
			"kexec_load", "kexec_file_load",
			"init_module", "finit_module", "delete_module",
			"bpf", "userfaultfd", "perf_event_open",
			"iopl", "ioperm", "acct", "swapon", "swapoff",
			"settimeofday", "clock_settime",
			"keyctl", "request_key", "add_key",
		},
	}
}
