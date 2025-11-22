//go:build windows
// +build windows

package health

import "errors"

// getDiskStats returns an error on Windows as disk stats are not implemented
func getDiskStats() (available, total uint64, err error) {
	// Disk stats not implemented on Windows
	return 0, 0, errors.New("disk stats not available on Windows")
}
