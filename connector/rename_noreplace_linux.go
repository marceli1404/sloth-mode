//go:build linux && (amd64 || arm64)

package main

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

// publishLinuxRuntimeDir is an atomic, fail-closed no-replace publication.
// This is used only for freshly extracted runtime directories inside the
// application's private data directory. In particular, os.Rename alone can
// replace an existing empty directory on Linux.
func publishLinuxRuntimeDir(src, dst string) error {
	from, err := syscall.BytePtrFromString(src)
	if err != nil {
		return err
	}
	to, err := syscall.BytePtrFromString(dst)
	if err != nil {
		return err
	}
	const atFDCWD = ^uintptr(99) // -100 converted to uintptr
	const renameNoReplace = 1
	_, _, errno := syscall.Syscall6(
		renameat2Syscall,
		atFDCWD, uintptr(unsafe.Pointer(from)),
		atFDCWD, uintptr(unsafe.Pointer(to)),
		renameNoReplace, 0,
	)
	if errno != 0 {
		if errors.Is(errno, syscall.EEXIST) {
			return os.ErrExist
		}
		// Deliberately fail closed on kernels/filesystems that do not
		// support atomic no-replace. Never fall back to os.Rename.
		return errno
	}
	return nil
}
