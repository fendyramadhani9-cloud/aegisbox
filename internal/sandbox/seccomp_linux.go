//go:build linux

package sandbox

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	auditArchX86_64  = 0xc000003e
	auditArchAArch64 = 0xc00000b7
)

// SockFilter represents a single BPF instruction.
type SockFilter struct {
	Code uint16
	Jt   uint8
	Jf   uint8
	K    uint32
}

// SockFprog represents the BPF program passed to seccomp.
type SockFprog struct {
	Len    uint16
	Filter *SockFilter
}

// ApplySeccompFilter applies a BPF filter program blocking dangerous syscalls.
func ApplySeccompFilter(policy SeccompPolicy) error {
	if !policy.Enabled {
		return nil
	}

	// PR_SET_NO_NEW_PRIVS is mandatory before loading unprivileged seccomp filters
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("seccomp requires PR_SET_NO_NEW_PRIVS: %w", err)
	}

	// Blacklisted syscall numbers on x86_64
	blockedNumbersX86_64 := []uint32{
		unix.SYS_MOUNT,
		unix.SYS_UMOUNT2,
		unix.SYS_PIVOT_ROOT,
		unix.SYS_REBOOT,
		unix.SYS_PTRACE,
		unix.SYS_PROCESS_VM_READV,
		unix.SYS_PROCESS_VM_WRITEV,
		unix.SYS_KEXEC_LOAD,
		unix.SYS_KEXEC_FILE_LOAD,
		unix.SYS_INIT_MODULE,
		unix.SYS_FINIT_MODULE,
		unix.SYS_DELETE_MODULE,
		unix.SYS_BPF,
		unix.SYS_USERFAULTFD,
		unix.SYS_PERF_EVENT_OPEN,
		unix.SYS_IOPL,
		unix.SYS_IOPERM,
		unix.SYS_ACCT,
		unix.SYS_SWAPON,
		unix.SYS_SWAPOFF,
		unix.SYS_SETTIMEOFDAY,
		unix.SYS_CLOCK_SETTIME,
		unix.SYS_KEYCTL,
		unix.SYS_REQUEST_KEY,
		unix.SYS_ADD_KEY,
	}

	// BPF instruction offsets
	const (
		seccompRetKill  = 0x00000000 // SECCOMP_RET_KILL_PROCESS
		seccompRetErrno = 0x00050001 // SECCOMP_RET_ERRNO | EPERM (1)
		seccompRetAllow = 0x7fff0000 // SECCOMP_RET_ALLOW

		bpfLd  = 0x00
		bpfW   = 0x00
		bpfAbs = 0x20
		bpfJmp = 0x05
		bpfJeq = 0x10
		bpfK   = 0x00
		bpfRet = 0x06
	)

	filter := []SockFilter{
		// 0: Load architecture (offset 4 in seccomp_data)
		{Code: bpfLd | bpfW | bpfAbs, K: 4},
		// 1: Check if x86_64 architecture; if yes jump to 3, else jump to 2 (kill)
		{Code: bpfJmp | bpfJeq | bpfK, Jt: 1, Jf: 0, K: auditArchX86_64},
		// 2: Kill process if architecture doesn't match
		{Code: bpfRet | bpfK, K: seccompRetKill},
		// 3: Load syscall number (offset 0 in seccomp_data)
		{Code: bpfLd | bpfW | bpfAbs, K: 0},
	}

	// Append checks for each blocked syscall
	for _, nr := range blockedNumbersX86_64 {
		filter = append(filter,
			// Check if syscall == nr
			// If match, jump 0 steps (next instruction: return errno EPERM)
			// If not match, jump 1 step (skip the return errno instruction)
			SockFilter{Code: bpfJmp | bpfJeq | bpfK, Jt: 0, Jf: 1, K: nr},
			SockFilter{Code: bpfRet | bpfK, K: seccompRetErrno},
		)
	}

	// Default: Allow all other syscalls
	filter = append(filter, SockFilter{Code: bpfRet | bpfK, K: seccompRetAllow})

	prog := SockFprog{
		Len:    uint16(len(filter)),
		Filter: &filter[0],
	}

	// Load seccomp filter via prctl
	_, _, errno := syscall.RawSyscall(
		syscall.SYS_PRCTL,
		unix.PR_SET_SECCOMP,
		unix.SECCOMP_MODE_FILTER,
		uintptr(unsafe.Pointer(&prog)),
	)
	if errno != 0 {
		return fmt.Errorf("failed to load seccomp BPF filter: %w", errno)
	}

	return nil
}
