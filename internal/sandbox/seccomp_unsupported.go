//go:build !linux

package sandbox

func ApplySeccompFilter(policy SeccompPolicy) error {
	return nil
}
