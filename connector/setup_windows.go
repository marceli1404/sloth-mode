//go:build windows

package main

// Setup and pairing is intentionally a *visible native Win32 window*.
// The tray icon is an optional convenience, not the only way to use the app.
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	wsOverlappedWindow = 0x00CF0000
	wsVisible          = 0x10000000
	wsChild            = 0x40000000
	wsTabstop          = 0x00010000
	wsVscroll          = 0x00200000
	esAutoHscroll      = 0x0080
	esReadOnly         = 0x0800
	wmCommand          = 0x0111
	wmSetFont          = 0x0030
	swShow             = 5
	swRestore          = 9
	idCopyKey          = 301
	idOpenWebsite      = 302
	idAutoStart        = 303
	idHide             = 304
	idCloseApp         = 305
	idToggleKey        = 306
	idLinkAccount      = 307
	idOpenAccount      = 308
	idPasteAccount     = 309
	wmHardwareResult   = 0x8000 + 47 // Win32 UI thread only
	wmAccountLinkDone  = 0x8000 + 46 // WM_APP; update native controls only on their owning thread
)

var (
	pShowWindow           = u32.NewProc("ShowWindow")
	pUpdateWindow         = u32.NewProc("UpdateWindow")
	pSetFocus             = u32.NewProc("SetFocus")
	pSendMessage          = u32.NewProc("SendMessageW")
	pGetStockObject       = syscall.NewLazyDLL("gdi32.dll").NewProc("GetStockObject")
	pSetWindowText        = u32.NewProc("SetWindowTextW")
	pIsWindow             = u32.NewProc("IsWindow")
	pFindWindow           = u32.NewProc("FindWindowW")
	pPostSetup            = u32.NewProc("PostMessageW")
	pSetWindowPos         = u32.NewProc("SetWindowPos")
	pGetSystemMetrics     = u32.NewProc("GetSystemMetrics")
	pSystemParametersInfo = u32.NewProc("SystemParametersInfoW")
	pGetDpiForSystem      = u32.NewProc("GetDpiForSystem")
	pInvalidateSetup      = u32.NewProc("InvalidateRect")
	setupHandle           uintptr
	setupKeyField         uintptr
	setupStartupButton    uintptr
	setupTrayLabel        uintptr
	setupAccountField     uintptr
	setupAccountStatus    uintptr
	setupHardwareLabel    uintptr
	setupHardwareMu       sync.Mutex
	setupHardwareResult   string
	accountLinkMu         sync.Mutex
	accountLinkResult     error
	accountLinkPending    bool // accessed only from the Win32 message loop
)

// Bound account-code input before allocating a UTF-16 buffer from UI length.
func validAccountLinkInputLength(length uintptr) bool { return length > 0 && length <= 64 }

func setupControl(parent uintptr, typ, content string, x, y, width, height int32, style uint32, id uintptr) uintptr {
	x, y, width, height = setupPx(x, setupScale), setupPx(y, setupScale), setupPx(width, setupScale), setupPx(height, setupScale)
	instance, _, _ := pModule.Call(0)
	hwnd, _, _ := pCreate.Call(0, wint(typ), wint(content), uintptr(wsVisible|wsChild|style), uintptr(x), uintptr(y), uintptr(width), uintptr(height), parent, id, instance, 0)
	if hwnd != 0 {
		pSendMessage.Call(hwnd, wmSetFont, setupButtonFont, 1)
	}
	return hwnd
}
func setupHeading(parent uintptr, s string, x, y, w, h int32) uintptr {
	hwnd := setupControl(parent, "STATIC", s, x, y, w, h, 0x80, 0)
	font := setupButtonFont
	if strings.HasPrefix(s, "01   ") || strings.HasPrefix(s, "02   ") || strings.HasPrefix(s, "03   ") || strings.HasPrefix(s, "04   ") {
		font = setupHeadingFont
	}
	if s == "Ready on this computer" {
		font = setupTitleFont
	}
	pSendMessage.Call(hwnd, wmSetFont, font, 1)
	return hwnd
}
func refreshStartupLabel() {
	if setupStartupButton == 0 {
		return
	}
	pSetWindowText.Call(setupStartupButton, wint("Start automatically at Windows sign-in"))
	pInvalidateSetup.Call(setupStartupButton, 0, 1)
}
func showSetupWindow() {
	if setupHandle == 0 {
		return
	}
	pShowWindow.Call(setupHandle, swRestore)
	pForeground.Call(setupHandle)
	pSetFocus.Call(setupKeyField)
	refreshStartupLabel()
}
func showExistingSetup() bool {
	hwnd, _, _ := pFindWindow.Call(wint("SlothModeSetupWindow"), 0)
	if hwnd == 0 {
		return false
	}
	pShowWindow.Call(hwnd, swRestore)
	pForeground.Call(hwnd)
	return true
}
func logStartup(s string) {
	if base, err := appDataDir(); err == nil {
		_ = os.MkdirAll(base, 0700)
		f, e := os.OpenFile(filepath.Join(base, "connector.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if e == nil {
			_, _ = fmt.Fprintln(f, s)
			_ = f.Close()
		}
	}
}

var setupCallback = syscall.NewCallback(func(hwnd, msg, w, l uintptr) uintptr {
	switch msg {
	case wmCommand:
		// Only the low word of WPARAM is a control ID.
		switch uint16(w & 0xffff) {
		case idCopyKey:
			if copyToClipboard(0, winBridge.secret) {
				popup(hwnd, "Copied", "Pairing key copied to clipboard. Paste it on sloth.marceli.dev. Keep it private.")
			} else {
				popup(hwnd, "Clipboard error", "Could not copy the pairing key. You can select the key in the text box and press Ctrl+C.")
			}
		case idOpenWebsite:
			openSite()
		case idOpenAccount:
			shell := "https://account.sloth.marceli.dev"
			pShellExecute.Call(0, wint("open"), wint(shell), 0, 0, 1)
		case idPasteAccount:
			// WM_PASTE uses the system clipboard; the text stays in the native
			// edit field and is never logged or sent until Link account is clicked.
			pSendMessage.Call(setupAccountField, 0x0302, 0, 0)
			pSetFocus.Call(setupAccountField)
		case idLinkAccount:
			if accountLinkPending {
				break // Prevent double-submit while a network request is running.
			}
			length, _, _ := pSendMessage.Call(setupAccountField, 0x000E, 0, 0) // WM_GETTEXTLENGTH
			if !validAccountLinkInputLength(length) {
				popup(hwnd, "Account linking", "Paste a valid 21-character code from your account page.")
				break
			}
			buff := make([]uint16, length+1) // bounded even for a malicious clipboard payload
			pGetText := u32.NewProc("GetWindowTextW")
			pGetText.Call(setupAccountField, uintptr(unsafe.Pointer(&buff[0])), uintptr(len(buff)))
			code := strings.TrimSpace(syscall.UTF16ToString(buff))
			accountLinkPending = true
			pSetWindowText.Call(setupAccountStatus, wint("Linking account…"))
			go func(owner uintptr, oneTimeCode string) {
				name, e := os.Hostname()
				if e != nil || name == "" {
					name = "Windows PC"
				}
				result := linkAccount(winBridge.base, oneTimeCode, name)
				accountLinkMu.Lock()
				accountLinkResult = result
				accountLinkMu.Unlock()
				// No SendMessage, SetWindowText or MessageBox from a network goroutine.
				// If the window was closed, posting simply fails and the main loop exits.
				pPostSetup.Call(owner, wmAccountLinkDone, 0, 0)
			}(hwnd, code)
		case idAutoStart:
			if startWithWindowsEnabled() {
				if e := changeStartWithWindows(false); e != nil {
					popup(hwnd, "Sign-in setting", e.Error())
				}
			} else {
				if e := changeStartWithWindows(true); e != nil {
					popup(hwnd, "Sign-in setting", e.Error())
				}
			}
			refreshStartupLabel()
		case idHide:
			if trayReady {
				pShowWindow.Call(hwnd, 0)
			} else {
				popup(hwnd, "Tray icon unavailable", "Windows could not display the tray icon. Keep this window open, or choose Exit connector.")
			}
		case idCloseApp:
			pDestroy.Call(hwnd) // wmDestroy handles server shutdown via hidden tray window.
		}
		return 0
	case wmHardwareResult:
		setupHardwareMu.Lock()
		result := setupHardwareResult
		setupHardwareMu.Unlock()
		if setupHardwareLabel != 0 {
			pSetWindowText.Call(setupHardwareLabel, wint(result))
		}
		return 0
	case setupWmCtlColorStatic, setupWmCtlColorEdit:
		return setupColorControl(msg, w)
	case setupWmDrawItem:
		return setupPaintButton(l)
	case wmAccountLinkDone:
		accountLinkMu.Lock()
		result := accountLinkResult
		accountLinkResult = nil
		accountLinkMu.Unlock()
		accountLinkPending = false
		if result != nil {
			logStartup("Account link failed: " + result.Error())
			pSetWindowText.Call(setupAccountStatus, wint("Link failed. Check the code and retry."))
			popup(hwnd, "Account linking", result.Error())
		} else {
			pSetWindowText.Call(setupAccountStatus, wint("Account linked! Manage this PC online."))
			popup(hwnd, "Sloth Mode", "This computer is linked. No notes were uploaded.")
		}
		return 0
	case wmClose:
		if trayReady {
			pShowWindow.Call(hwnd, 0)
		} else {
			popup(hwnd, "Sloth Mode", "The tray icon is unavailable, so this window must stay open. Select Exit connector to stop the app.")
		}
		return 0
	case wmDestroy:
		setupHandle = 0
		setupHardwareLabel = 0
		setupReleaseTheme()
		// Exiting the visible setup window deliberately quits only when the tray is absent,
		// or the user pressed Exit. The hidden tray owner manages server shutdown.
		if trayData.Hwnd != 0 {
			pDestroy.Call(trayData.Hwnd)
		} else {
			if winServer != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				_ = winServer.Shutdown(ctx)
				cancel()
			}
			pQuit.Call(0)
		}
		return 0
	}
	v, _, _ := pDef.Call(hwnd, msg, w, l)
	return v
})

func createSetupWindow(instance uintptr) bool {
	cls := "SlothModeSetupWindow"
	setupWork := setupRect{}
	pSystemParametersInfo.Call(0x30, 0, uintptr(unsafe.Pointer(&setupWork)), 0)
	if setupWork.right <= setupWork.left || setupWork.bottom <= setupWork.top {
		screenX, _, _ := pGetSystemMetrics.Call(0)
		screenY, _, _ := pGetSystemMetrics.Call(1)
		setupWork = setupRect{0, 0, int32(screenX), int32(screenY)}
	}
	dpi := uintptr(96)
	if pGetDpiForSystem.Find() == nil {
		dpi, _, _ = pGetDpiForSystem.Call()
	}
	if dpi == 0 {
		dpi = 96
	}
	client := setupRect{0, 0, setupClientWidth, setupClientHeight}
	adjust := u32.NewProc("AdjustWindowRectExForDpi")
	if adjust.Find() == nil {
		adjust.Call(uintptr(unsafe.Pointer(&client)), wsOverlappedWindow, 0, 0, dpi)
	} else {
		u32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&client)), wsOverlappedWindow, 0, 0)
	}
	frameWidth, frameHeight := client.right-client.left-setupClientWidth, client.bottom-client.top-setupClientHeight
	setupScale = setupFitScale(uint32(dpi), setupWork.right-setupWork.left, setupWork.bottom-setupWork.top, frameWidth, frameHeight)
	setupInitTheme(setupScale)
	client = setupRect{0, 0, setupPx(setupClientWidth, setupScale), setupPx(setupClientHeight, setupScale)}
	if adjust.Find() == nil {
		adjust.Call(uintptr(unsafe.Pointer(&client)), wsOverlappedWindow, 0, 0, dpi)
	} else {
		u32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&client)), wsOverlappedWindow, 0, 0)
	}
	frameWidth, frameHeight = client.right-client.left-setupPx(setupClientWidth, setupScale), client.bottom-client.top-setupPx(setupClientHeight, setupScale)
	bounds := setupBounds(setupWork, setupScale, frameWidth, frameHeight)
	wc := winClass{Size: uint32(unsafe.Sizeof(winClass{})), WndProc: setupCallback, Instance: instance, ClassName: wint(cls), Background: setupCreamBrush}
	if r, _, _ := pRegister.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		logStartup("RegisterClassExW failed for setup window")
		setupReleaseTheme()
		return false
	}
	hwnd, _, _ := pCreate.Call(0, wint(cls), wint("Sloth Mode - Setup & Pairing"), wsOverlappedWindow, uintptr(bounds.left), uintptr(bounds.top), uintptr(bounds.right-bounds.left), uintptr(bounds.bottom-bounds.top), 0, 0, instance, 0)
	if hwnd == 0 {
		logStartup("CreateWindowExW failed for setup window")
		setupReleaseTheme()
		return false
	}
	setupHandle = hwnd
	setupHeading(hwnd, "SLOTH MODE   /   LOCAL COMPANION", 28, 20, 590, 22)
	setupHeading(hwnd, "Ready on this computer", 28, 47, 610, 31)
	setupHeading(hwnd, "Your connector stays on this PC. Pair it at sloth.marceli.dev.", 28, 81, 610, 22)
	setupHeading(hwnd, "01   Pair a browser", 28, 119, 610, 23)
	setupHeading(hwnd, "Copy this private key, then paste it into the website.", 28, 143, 610, 21)
	setupKeyField = setupControl(hwnd, "EDIT", winBridge.secret, 28, 168, 474, 32, wsTabstop|esAutoHscroll|esReadOnly, 0)
	setupControl(hwnd, "BUTTON", "Copy key", 512, 168, 126, 32, wsTabstop|setupBsOwnerDraw, idCopyKey)
	setupHeading(hwnd, "02   Connect the website", 28, 211, 610, 23)
	setupHeading(hwnd, "Select Find local app, paste the key, and allow local access.", 28, 235, 610, 21)
	setupControl(hwnd, "BUTTON", "Open sloth.marceli.dev", 28, 260, 245, 36, wsTabstop|setupBsOwnerDraw, idOpenWebsite)
	setupHeading(hwnd, "03   Link your account (optional)", 28, 310, 610, 23)
	setupHeading(hwnd, "Get a one-time code from your account page, then paste it here.", 28, 334, 610, 21)
	setupAccountField = setupControl(hwnd, "EDIT", "", 28, 359, 270, 32, wsTabstop|esAutoHscroll, 0)
	pSendMessage.Call(setupAccountField, 0x00C5, 64, 0) // EM_LIMITTEXT: never buffer unbounded pasted codes
	setupControl(hwnd, "BUTTON", "Paste", 306, 359, 63, 32, wsTabstop|setupBsOwnerDraw, idPasteAccount)
	setupControl(hwnd, "BUTTON", "Link", 377, 359, 78, 32, wsTabstop|setupBsOwnerDraw, idLinkAccount)
	setupControl(hwnd, "BUTTON", "Account page", 463, 359, 145, 32, wsTabstop|setupBsOwnerDraw, idOpenAccount)
	setupAccountStatus = setupHeading(hwnd, "Not linked. Browser pairing works without this.", 28, 394, 610, 21)
	setupHeading(hwnd, "04   Add Gemma (optional)", 28, 434, 610, 23)
	setupHeading(hwnd, "Approve a model download on your dashboard. No Ollama needed.", 28, 458, 610, 21)
	setupHardwareLabel = setupHeading(hwnd, "Checking this PC for a suitable Gemma size…", 28, 481, 610, 21)
	go func(owner uintptr) {
		p := profileHardware()
		setupHardwareMu.Lock()
		setupHardwareResult = "Recommended Gemma " + p.Recommended + " · See dashboard to change."
		setupHardwareMu.Unlock()
		pPostSetup.Call(owner, wmHardwareResult, 0, 0)
	}(hwnd)
	setupStartupButton = setupControl(hwnd, "BUTTON", "", 28, 511, 390, 28, wsTabstop|setupBsOwnerDraw, idAutoStart)
	setupTrayLabel = setupHeading(hwnd, "Close hides the connector in the tray.", 28, 562, 300, 26)
	setupControl(hwnd, "BUTTON", "Hide to tray", 361, 556, 128, 36, wsTabstop|setupBsOwnerDraw, idHide)
	setupControl(hwnd, "BUTTON", "Exit connector", 500, 556, 138, 36, wsTabstop|setupBsOwnerDraw, idCloseApp)
	if !trayReady {
		pSetWindowText.Call(setupTrayLabel, wint("No tray icon. Keep this window open."))
	}
	refreshStartupLabel()
	pShowWindow.Call(hwnd, swShow)
	pUpdateWindow.Call(hwnd)
	pForeground.Call(hwnd)
	logStartup("Setup window shown; connector is listening on 127.0.0.1:" + bridgePort)
	return true
}
