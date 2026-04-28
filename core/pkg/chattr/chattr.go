//go:build !windows

package chattr

import (
	"os"

	"github.com/g0rbe/go-chattr"
)

// File attribute flags (same as go-chattr)
const (
	FS_IMMUTABLE_FL = chattr.FS_IMMUTABLE_FL
	FS_APPEND_FL    = chattr.FS_APPEND_FL
	FS_NODUMP_FL    = chattr.FS_NODUMP_FL
	FS_SYNC_FL      = chattr.FS_SYNC_FL
	FS_COMPR_FL     = chattr.FS_COMPR_FL
)

// SetAttr sets the given attributes on the file
func SetAttr(file *os.File, flags int32) error {
	return chattr.SetAttr(file, flags)
}

// UnsetAttr unsets the given attributes on the file
func UnsetAttr(file *os.File, flags int32) error {
	return chattr.UnsetAttr(file, flags)
}