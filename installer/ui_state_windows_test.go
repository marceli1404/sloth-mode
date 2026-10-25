//go:build windows

package main

import (
	"context"
	"runtime"
	"testing"
)

// Exercise native enable/disable state without showing a window, installing an
// app, changing startup settings or contacting the update server.
func TestInstallerOperationControls(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	h, _, err := createWindow.Call(0, utf16("STATIC"), utf16("Installer state fixture"), wsOverlappedWindow, 0, 0, 760, 700, 0, 0, 0, 0)
	if h == 0 {
		t.Fatalf("fixture window: %v", err)
	}
	oldHwnd, oldInstall, oldUpdate, oldStartup := hwnd, installButton, updateButton, startupButton
	hwnd = h
	defer func() {
		destroyWindow.Call(h)
		hwnd, installButton, updateButton, startupButton = oldHwnd, oldInstall, oldUpdate, oldStartup
		operationBusy, operationCancel, closeRequested = false, nil, false
	}()
	installButton = ctrl("BUTTON", "Install", 0, 0, 100, 40, wsTabstop, idInstall)
	updateButton = ctrl("BUTTON", "Update", 110, 0, 100, 40, wsTabstop, idCheckUpdates)
	startupButton = ctrl("BUTTON", startChoice(), 0, 50, 400, 40, wsTabstop|0x2, idStartup)
	for _, c := range []uintptr{installButton, updateButton, startupButton} {
		if c == 0 {
			t.Fatal("control creation failed")
		}
	}
	if !beginOperation() {
		t.Fatal("first operation rejected")
	}
	if beginOperation() {
		t.Fatal("overlapping operation accepted")
	}
	for _, c := range []uintptr{installButton, updateButton, startupButton} {
		if enabled, _, _ := user32.NewProc("IsWindowEnabled").Call(c); enabled != 0 {
			t.Fatal("busy control still enabled")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	operationCancel = cancel
	if finishOperation() {
		t.Fatal("window unexpectedly closed")
	}
	if ctx.Err() == nil || operationBusy {
		t.Fatal("operation not cancelled/cleared")
	}
	for _, c := range []uintptr{installButton, updateButton, startupButton} {
		if enabled, _, _ := user32.NewProc("IsWindowEnabled").Call(c); enabled == 0 {
			t.Fatal("control not re-enabled")
		}
	}
	sendMessage.Call(startupButton, 0x00F1, 1, 0)
	if checked, _, _ := sendMessage.Call(startupButton, 0x00F0, 0, 0); checked != 1 {
		t.Fatal("checkbox checked state missing")
	}
	ctx, cancel = context.WithCancel(context.Background())
	operationCancel = cancel
	if !beginOperation() {
		t.Fatal("second operation rejected")
	}
	requestClose()
	if !closeRequested || ctx.Err() == nil {
		t.Fatal("close did not cancel pending work")
	}
	if valid, _, _ := user32.NewProc("IsWindow").Call(h); valid == 0 {
		t.Fatal("closed before worker could finish")
	}
}
