//go:build !windows
// +build !windows

package health

import "syscall"

// getDiskStats returns disk usage statistics for Unix-like systems
func getDiskStats() (available, total uint64, err error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, 0, err
	}

	available = stat.Bavail * uint64(stat.Bsize)
	total = stat.Blocks * uint64(stat.Bsize)
	return available, total, nil
}
