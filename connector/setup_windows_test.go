//go:build windows

package main

import (
	"runtime"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

func TestSetupLayoutFitsWorkAreasAndControlBounds(t *testing.T) {
	frames := []struct{ width, height int32 }{{16, 39}, {12, 32}}
	areas := []setupRect{{0, 0, 1920, 1040}, {0, 0, 1366, 728}, {0, 0, 1280, 660}, {-1920, 0, 0, 800}}
	for _, area := range areas {
		for _, frame := range frames {
			for _, dpi := range []uint32{96, 120, 144, 192} {
				scale := setupFitScale(dpi, area.right-area.left, area.bottom-area.top, frame.width, frame.height)
				bounds := setupBounds(area, scale, frame.width, frame.height)
				if bounds.left < area.left || bounds.top < area.top || bounds.right > area.right || bounds.bottom > area.bottom {
					t.Fatalf("bounds %+v exceed work area %+v at %ddpi, scale %.2f", bounds, area, dpi, scale)
				}
				if setupPx(638, scale) > setupPx(setupClientWidth, scale) || setupPx(592, scale) > setupPx(setupClientHeight, scale) {
					t.Fatalf("control layout exceeds scaled client at scale %.2f", scale)
				}
			}
		}
	}
}

func TestSetupCopyFitsMeasuredControls(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	getDC := u32.NewProc("GetDC")
	releaseDC := u32.NewProc("ReleaseDC")
	createFont := setupGDI.NewProc("CreateFontW")
	drawText := u32.NewProc("DrawTextW")
	selectObject := setupGDI.NewProc("SelectObject")
	deleteObject := setupGDI.NewProc("DeleteObject")
	dc, _, _ := getDC.Call(0)
	if dc == 0 {
		t.Fatal("GetDC failed")
	}
	defer releaseDC.Call(0, dc)
	for _, scale := range []float64{0.72, 0.85, 1} {
		for _, tc := range []struct {
			text          string
			width, height int32
			fontSize      int32
			fontWeight    uintptr
		}{
			{"Ready on this computer", 610, 31, 23, 600},
			{"01   Pair a browser", 610, 23, 16, 600},
			{"Your connector stays on this PC. Pair it at sloth.marceli.dev.", 610, 22, 15, 400},
			{"Copy this private key, then paste it into the website.", 610, 21, 15, 400},
			{"Select Find local app, paste the key, and allow local access.", 610, 21, 15, 400},
			{"Get a one-time code from your account page, then paste it here.", 610, 21, 15, 400},
			{"Not linked. Browser pairing works without this.", 610, 21, 15, 400},
			{"Approve a model download on your dashboard. No Ollama needed.", 610, 21, 15, 400},
			{"Recommended Gemma 27B · See dashboard to change.", 610, 21, 15, 400},
			{"No tray icon. Keep this window open.", 300, 26, 15, 400},
			{"Start automatically at Windows sign-in", 390, 28, 15, 400},
		} {
			fontName := wide("Segoe UI")
			font, _, _ := createFont.Call(uintptr(uint32(-int32(float64(tc.fontSize)*scale))), 0, 0, 0, tc.fontWeight, 0, 0, 0, 0, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
			old, _, _ := selectObject.Call(dc, font)
			text, _ := syscall.UTF16FromString(tc.text)
			r := setupRect{0, 0, setupPx(tc.width, scale), setupPx(tc.height, scale)}
			drawText.Call(dc, uintptr(unsafe.Pointer(&text[0])), uintptr(len(text)-1), uintptr(unsafe.Pointer(&r)), 0x400|0x10|0x800)
			selectObject.Call(dc, old)
			deleteObject.Call(font)
			runtime.KeepAlive(fontName)
			if r.bottom > setupPx(tc.height, scale) {
				t.Errorf("scale %.2f text %q needs %dpx, available %dpx", scale, strings.TrimSpace(tc.text), r.bottom, setupPx(tc.height, scale))
			}
		}
	}
}

func TestAccountLinkCodeInputIsBounded(t *testing.T) {
	for _, tc := range []struct {
		length uintptr
		accept bool
	}{{0, false}, {1, true}, {21, true}, {64, true}, {65, false}, {65536, false}, {^uintptr(0), false}} {
		if got := validAccountLinkInputLength(tc.length); got != tc.accept {
			t.Fatalf("length %d: got %v, want %v", tc.length, got, tc.accept)
		}
	}
}
