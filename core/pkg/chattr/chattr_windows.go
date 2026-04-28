//go:build windows

package chattr

import (
	"os"
)

// File attribute flags (no-op on Windows)
const (
	FS_IMMUTABLE_FL = 0x00000010
	FS_APPEND_FL    = 0x00000020
	FS_NODUMP_FL    = 0x00000040
	FS_SYNC_FL      = 0x00000080
	FS_COMPR_FL     = 0x00000400
)

// SetAttr sets the given attributes on the file (no-op on Windows)
func SetAttr(file *os.File, flags int32) error {
	return nil // no-op on Windows
}

// UnsetAttr unsets the given attributes on the file (no-op on Windows)
func UnsetAttr(file *os.File, flags int32) error {
	return nil // no-op on Windows
}