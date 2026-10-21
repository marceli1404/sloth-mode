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
	wmAccountLinkDone   = 0x8000 + 46 // WM_APP; update native controls only on their owning thread
)

var (
	pShowWindow        = u32.NewProc("ShowWindow")
	pUpdateWindow      = u32.NewProc("UpdateWindow")
	pSetFocus          = u32.NewProc("SetFocus")
	pSendMessage       = u32.NewProc("SendMessageW")
	pGetStockObject    = syscall.NewLazyDLL("gdi32.dll").NewProc("GetStockObject")
	pSetWindowText     = u32.NewProc("SetWindowTextW")
	pIsWindow          = u32.NewProc("IsWindow")
	pFindWindow        = u32.NewProc("FindWindowW")
	pPostSetup         = u32.NewProc("PostMessageW")
	pSetWindowPos      = u32.NewProc("SetWindowPos")
	pGetSystemMetrics  = u32.NewProc("GetSystemMetrics")
	setupHandle        uintptr
	setupKeyField      uintptr
	setupStartupButton uintptr
	setupTrayLabel     uintptr
	setupAccountField  uintptr
	setupAccountStatus uintptr
	setupHardwareLabel uintptr
	setupHardwareMu sync.Mutex
	setupHardwareResult string
	accountLinkMu      sync.Mutex
	accountLinkResult  error
	accountLinkPending bool // accessed only from the Win32 message loop
)

func setupControl(parent uintptr, typ, content string, x, y, width, height int32, style uint32, id uintptr) uintptr {
	instance, _, _ := pModule.Call(0)
	hwnd, _, _ := pCreate.Call(0, wint(typ), wint(content), uintptr(wsVisible|wsChild|style), uintptr(x), uintptr(y), uintptr(width), uintptr(height), parent, id, instance, 0)
	if hwnd != 0 {
		font, _, _ := pGetStockObject.Call(17) // DEFAULT_GUI_FONT
		pSendMessage.Call(hwnd, wmSetFont, font, 1)
	}
	return hwnd
}
func setupHeading(parent uintptr, s string, x, y, w, h int32) uintptr {
	return setupControl(parent, "STATIC", s, x, y, w, h, 0, 0)
}
func refreshStartupLabel() {
	if setupStartupButton == 0 {
		return
	}
	label := "[ ] Start automatically at Windows sign-in"
	if startWithWindowsEnabled() {
		label = "[x] Start automatically at Windows sign-in"
	}
	pSetWindowText.Call(setupStartupButton, wint(label))
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
			if length == 0 || length > 64 {
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
				if e != nil || name == "" { name = "Windows PC" }
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
	wc := winClass{Size: uint32(unsafe.Sizeof(winClass{})), WndProc: setupCallback, Instance: instance, ClassName: wint(cls), Background: 6}
	if r, _, _ := pRegister.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		logStartup("RegisterClassExW failed for setup window")
		return false
	}
	screenX, _, _ := pGetSystemMetrics.Call(0)
	screenY, _, _ := pGetSystemMetrics.Call(1)
	x, y := int32(140), int32(90)
	if screenX > 720 {
		x = (int32(screenX) - 690) / 2
	}
	if screenY > 700 {
		y = (int32(screenY) - 680) / 2
	}
	hwnd, _, _ := pCreate.Call(0, wint(cls), wint("Sloth Mode - Setup & Pairing"), wsOverlappedWindow, uintptr(x), uintptr(y), 690, 680, 0, 0, instance, 0)
	if hwnd == 0 {
		logStartup("CreateWindowExW failed for setup window")
		return false
	}
	setupHandle = hwnd
	setupHeading(hwnd, "SLOTH MODE     /     LOCAL AI COMPANION", 29, 25, 570, 28)
	setupHeading(hwnd, "Ready on this computer", 29, 60, 620, 28)
	setupHeading(hwnd, "The connector runs locally. You can control it through sloth.marceli.dev.", 29, 92, 635, 25)
	setupHeading(hwnd, "STEP 1     Browser pairing key (not an account code)", 29, 139, 610, 22)
	setupHeading(hwnd, "Use Copy key below, then paste into the website. Never share this key.", 29, 165, 610, 22)
	setupKeyField = setupControl(hwnd, "EDIT", winBridge.secret, 29, 193, 475, 33, wsTabstop|esAutoHscroll|esReadOnly, 0)
	setupControl(hwnd, "BUTTON", "Copy key", 516, 193, 134, 33, wsTabstop, idCopyKey)
	setupHeading(hwnd, "STEP 2     Connect the website", 29, 247, 610, 22)
	setupHeading(hwnd, "Open the site, click Find local app, and paste your key. Allow local access if asked.", 29, 272, 635, 22)
	setupControl(hwnd, "BUTTON", "Open sloth.marceli.dev", 29, 302, 254, 37, wsTabstop, idOpenWebsite)
	setupHeading(hwnd, "STEP 3     Enter your one-time account linking code", 29, 351, 610, 22)
	setupHeading(hwnd, "Account dashboard: Link a computer > Copy code > Paste below.", 29, 376, 610, 22)
	setupAccountField = setupControl(hwnd, "EDIT", "", 29, 404, 282, 34, wsTabstop|esAutoHscroll, 0)
	pSendMessage.Call(setupAccountField, 0x00C5, 64, 0) // EM_LIMITTEXT: never buffer unbounded pasted codes
	setupControl(hwnd, "BUTTON", "Paste", 319, 404, 65, 34, wsTabstop, idPasteAccount)
	setupControl(hwnd, "BUTTON", "Link", 392, 404, 103, 34, wsTabstop, idLinkAccount)
	setupControl(hwnd, "BUTTON", "Account page", 504, 404, 148, 34, wsTabstop, idOpenAccount)
	setupAccountStatus = setupHeading(hwnd, "Not linked. Local pairing still works.", 29, 443, 620, 22)
	setupHeading(hwnd, "STEP 4     Optional: install local Gemma", 29, 488, 610, 22)
	setupHeading(hwnd, "Approve a verified GGUF download from your dashboard; no Ollama needed.", 29, 513, 625, 22)
	setupHardwareLabel = setupHeading(hwnd, "Checking your hardware for the best Gemma size...", 29, 534, 625, 20)
	go func(owner uintptr) {
		p := profileHardware()
		setupHardwareMu.Lock()
		setupHardwareResult = "Recommended Gemma " + p.Recommended + " · See dashboard to change."
		setupHardwareMu.Unlock()
		pPostSetup.Call(owner, wmHardwareResult, 0, 0)
	}(hwnd)
	setupStartupButton = setupControl(hwnd, "BUTTON", "", 29, 556, 400, 28, wsTabstop, idAutoStart)
	setupControl(hwnd, "BUTTON", "Hide to tray", 368, 603, 136, 37, wsTabstop, idHide)
	setupControl(hwnd, "BUTTON", "Exit connector", 515, 603, 136, 37, wsTabstop, idCloseApp)
	setupTrayLabel = setupHeading(hwnd, "Closing this window hides the connector in the tray.", 29, 606, 322, 38)
	if !trayReady {
		pSetWindowText.Call(setupTrayLabel, wint("Tray icon unavailable. Keep this window open."))
	}
	refreshStartupLabel()
	pShowWindow.Call(hwnd, swShow)
	pUpdateWindow.Call(hwnd)
	pForeground.Call(hwnd)
	logStartup("Setup window shown; connector is listening on 127.0.0.1:" + bridgePort)
	return true
}
