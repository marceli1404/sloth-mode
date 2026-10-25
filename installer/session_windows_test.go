//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"testing"
)

func TestInstallerSessionExcludesOverlapAndAllowsHandoff(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	name := fmt.Sprintf(`Local\SlothModeInstaller-test-%d`, os.Getpid())
	h, err := acquireInstallerSession(name)
	if err != nil {
		t.Fatal(err)
	}
	defer kernel32.NewProc("CloseHandle").Call(h)
	owned := true
	defer func() {
		if owned {
			kernel32.NewProc("ReleaseMutex").Call(h)
		}
	}()
	probe := func() error {
		result := make(chan error, 1)
		go func() {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			next, err := acquireInstallerSession(name)
			if err == nil {
				kernel32.NewProc("ReleaseMutex").Call(next)
				kernel32.NewProc("CloseHandle").Call(next)
			}
			result <- err
		}()
		return <-result
	}
	if err := probe(); !errors.Is(err, errSetupAlreadyOpen) {
		t.Fatalf("parallel setup was not excluded: %v", err)
	}
	if ok, _, err := kernel32.NewProc("ReleaseMutex").Call(h); ok == 0 {
		t.Fatalf("release failed: %v", err)
	}
	owned = false
	// Our handle remains open, exactly as it does while the successor starts.
	if err := probe(); err != nil {
		t.Fatalf("verified successor could not take over: %v", err)
	}
}
