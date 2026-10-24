//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func TestInstallerDPIFitsWorkArea(t *testing.T) {
	oldScale := uiScale
	defer func() { uiScale = oldScale }()
	for _, work := range []rectWin{{0, 0, 1920, 1040}, {0, 0, 1366, 728}, {-1920, 0, 0, 1040}, {0, 0, 3840, 2080}} {
		for _, dpi := range []uint32{96, 120, 144, 192} {
			t.Run(fmt.Sprintf("%dx%d-%ddpi", work.Right-work.Left, work.Bottom-work.Top, dpi), func(t *testing.T) {
				uiScale = fittedScale(dpi, work)
				r := windowBounds(work, dpi)
				if r.Left < work.Left || r.Top < work.Top || r.Right > work.Right || r.Bottom > work.Bottom {
					t.Fatalf("window %+v outside %+v", r, work)
				}
			})
		}
	}
}

func TestInstallerControlTextFits(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	oldHwnd, oldScale := hwnd, uiScale
	defer func() { hwnd, uiScale = oldHwnd, oldScale }()
	for _, scale := range []float64{0.85, 1, 1.25, 1.5, 2} {
		uiScale = scale
		makeTheme()
		h, _, _ := createWindow.Call(0, utf16("STATIC"), utf16("Layout fixture"), wsOverlappedWindow, 0, 0, uintptr(px(clientWidth)), uintptr(px(clientHeight)), 0, 0, 0, 0)
		if h == 0 {
			t.Fatal("fixture creation failed")
		}
		hwnd = h
		buildInstallerControls()
		for _, c := range controls {
			if c.handle == 0 {
				t.Fatal("control creation failed")
			}
			if c.x+c.width > clientWidth || c.y+c.height > clientHeight {
				t.Fatalf("control outside client: %+v", c)
			}
			if c.role == "body" {
				continue
			} // Button labels are single-line and separately sized.
			var text [512]uint16
			sendMessage.Call(c.handle, 0xD, 512, uintptr(unsafe.Pointer(&text[0])))
			dc, _, _ := user32.NewProc("GetDC").Call(c.handle)
			old, _, _ := selectObj.Call(dc, fontForRole(c.role))
			r := rectWin{0, 0, px(c.width), 0}
			drawText.Call(dc, uintptr(unsafe.Pointer(&text[0])), ^uintptr(0), uintptr(unsafe.Pointer(&r)), 0x400|0x10|0x800) // CALCRECT, WORDBREAK, NOPREFIX
			selectObj.Call(dc, old)
			user32.NewProc("ReleaseDC").Call(c.handle, dc)
			if r.Bottom > px(c.height) {
				t.Errorf("scale %.2f text %q needs %dpx, available %dpx", scale, syscall.UTF16ToString(text[:]), r.Bottom, px(c.height))
			}
		}
		destroyWindow.Call(h)
		releaseTheme()
	}
}

// Show the production window without installing, launching a companion,
// changing startup settings or contacting update servers. Separate opt-in
// process: it does not take or bypass the real installer's mutation lock.
func TestInstallerVisualPreviewOptIn(t *testing.T) {
	if os.Getenv("SLOTH_INSTALLER_UI_PREVIEW") != "1" {
		t.Skip("manual UI fixture")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := createInstallerUI(); err != nil {
		t.Fatal(err)
	}
	defer releaseTheme()
	setWindowText.Call(hwnd, utf16("Sloth Mode - Layout preview (no installation)"))
	// Replace only the fixture's procedure: commands that could mutate state
	// are swallowed, while painting, focus, DPI, checkbox and Close stay real.
	guard := syscall.NewCallback(func(h, msg, w, l uintptr) uintptr {
		if msg == wmCommand {
			id := uint16(w & 0xffff)
			if id != idClose && id != 2 && id != idStartup {
				return 0
			}
		}
		r, _, _ := user32.NewProc("CallWindowProcW").Call(wndproc, h, msg, w, l)
		return r
	})
	user32.NewProc("SetWindowLongPtrW").Call(hwnd, ^uintptr(3), guard)
	status("Layout preview only. Install and update actions are inactive.")
	showWindow.Call(hwnd, swShow) // Explicit visible app window after a hidden helper launch.
	owner := hwnd
	go func() { time.Sleep(4 * time.Minute); postMessage.Call(owner, wmClose, 0, 0) }()
	runMessageLoop()
}
