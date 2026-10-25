//go:build windows

package main

import (
	"errors"
	"fmt"
	"syscall"
)

var errSetupAlreadyOpen = errors.New("Setup is already open. Return to the existing installer window.")

// The caller must stay on the same OS thread until ReleaseMutex. An existing
// handle may outlive its ownership during the verified-update handoff.
func acquireInstallerSession(name string) (uintptr, error) {
	h, _, err := kernel32.NewProc("CreateMutexW").Call(0, 1, utf16(name))
	if h == 0 {
		return 0, fmt.Errorf("could not reserve installer session: %v", err)
	}
	if err == syscall.ERROR_ALREADY_EXISTS {
		wait, _, _ := kernel32.NewProc("WaitForSingleObject").Call(h, 0)
		if wait != 0 && wait != 0x80 { // WAIT_OBJECT_0 or abandoned owner.
			kernel32.NewProc("CloseHandle").Call(h)
			return 0, errSetupAlreadyOpen
		}
	}
	return h, nil
}
