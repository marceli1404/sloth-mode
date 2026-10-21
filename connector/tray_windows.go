//go:build windows

package main

import (
	"context"
	_ "embed"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	wmDestroy     = 0x0002
	wmClose       = 0x0010
	wmApp         = 0x8000
	wmTray        = wmApp + 33
	wmRButtonUp   = 0x0205
	wmLButtonUp   = 0x0202
	wmContextMenu = 0x007B
	nifMessage    = 0x00000001
	nifIcon       = 0x00000002
	nifTip        = 0x00000004
	nimAdd        = 0x0
	nimDelete     = 0x2
	niVersion4    = 4
	mfString      = 0x0000
	mfSeparator   = 0x0800
	tpmRetCmd     = 0x0100
	tpmRight      = 0x0002
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
	idOpen        = 201
	idPair        = 202
	idQuit        = 203
	idStartup     = 204
	idSetup       = 205
)

type winPoint struct{ X, Y int32 }
type winMessage struct {
	Hwnd           uintptr
	Msg            uint32
	Pad            uint32
	WParam, LParam uintptr
	Time           uint32
	P              winPoint
	Extra          uint32
}
type winClass struct {
	Size                                                               uint32
	Style                                                              uint32
	WndProc                                                            uintptr
	ClassExtra, WindowExtra                                            int32
	Instance, Icon, Cursor, Background, MenuName, ClassName, SmallIcon uintptr
}

// NOTIFYICONDATAW has different layout on x64; Go's native alignment preserves it.
type notifyIcon struct {
	Size            uint32
	Hwnd            uintptr
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            uintptr
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	GuidItem        [16]byte
	BalloonIcon     uintptr
}

var (
	u32           = syscall.NewLazyDLL("user32.dll")
	sh32          = syscall.NewLazyDLL("shell32.dll")
	k32           = syscall.NewLazyDLL("kernel32.dll")
	pRegister     = u32.NewProc("RegisterClassExW")
	pCreate       = u32.NewProc("CreateWindowExW")
	pDef          = u32.NewProc("DefWindowProcW")
	pGet          = u32.NewProc("GetMessageW")
	pTranslate    = u32.NewProc("TranslateMessage")
	pDispatch     = u32.NewProc("DispatchMessageW")
	pDestroy      = u32.NewProc("DestroyWindow")
	pQuit         = u32.NewProc("PostQuitMessage")
	pLoadIcon     = u32.NewProc("LoadIconW")
	pCreateMenu   = u32.NewProc("CreatePopupMenu")
	pAppendMenu   = u32.NewProc("AppendMenuW")
	pDestroyMenu  = u32.NewProc("DestroyMenu")
	pTrack        = u32.NewProc("TrackPopupMenu")
	pGetPos       = u32.NewProc("GetCursorPos")
	pForeground   = u32.NewProc("SetForegroundWindow")
	pMessageBox   = u32.NewProc("MessageBoxW")
	pShellExecute = sh32.NewProc("ShellExecuteW")
	pNotify       = sh32.NewProc("Shell_NotifyIconW")
	pModule       = k32.NewProc("GetModuleHandleW")
	pAlloc        = k32.NewProc("GlobalAlloc")
	pLock         = k32.NewProc("GlobalLock")
	pUnlock       = k32.NewProc("GlobalUnlock")
	pFree         = k32.NewProc("GlobalFree")
	pOpenClip     = u32.NewProc("OpenClipboard")
	pEmptyClip    = u32.NewProc("EmptyClipboard")
	pSetClip      = u32.NewProc("SetClipboardData")
	pCloseClip    = u32.NewProc("CloseClipboard")
)
var winBridge *Bridge
var winServer *http.Server
var trayData notifyIcon
var trayReady bool

func wide(s string) *uint16                   { v, _ := syscall.UTF16PtrFromString(s); return v }
func wint(s string) uintptr                   { return uintptr(unsafe.Pointer(wide(s))) }
func popup(hwnd uintptr, caption, msg string) { pMessageBox.Call(hwnd, wint(msg), wint(caption), 0x40) }
func openSite() {
	pShellExecute.Call(0, wint("open"), wint("https://sloth.marceli.dev/#connector"), 0, 0, 1)
}
func copyToClipboard(hwnd uintptr, s string) bool {
	data := syscall.StringToUTF16(s)
	mem, _, _ := pAlloc.Call(gmemMoveable, uintptr(len(data)*2))
	if mem == 0 {
		return false
	}
	loc, _, _ := pLock.Call(mem)
	if loc == 0 {
		pFree.Call(mem)
		return false
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(loc)), len(data)), data)
	pUnlock.Call(mem)
	ok, _, _ := pOpenClip.Call(hwnd)
	if ok == 0 {
		pFree.Call(mem)
		return false
	}
	defer pCloseClip.Call()
	pEmptyClip.Call()
	res, _, _ := pSetClip.Call(cfUnicodeText, mem)
	if res == 0 {
		pFree.Call(mem)
		return false
	}
	return true
}

// Auto-start is opt-in and per-user only. We never create a system service.
const startupKey = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`

func startWithWindowsEnabled() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	cmd := exec.Command("reg.exe", "QUERY", startupKey, "/v", "SlothModeConnector")
	configureHiddenProcess(cmd)
	out, err := cmd.Output()
	return err == nil && strings.Contains(strings.ToLower(string(out)), strings.ToLower(exe))
}
func changeStartWithWindows(enable bool) error {
	var cmd *exec.Cmd
	if enable {
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		cmd = exec.Command("reg.exe", "ADD", startupKey, "/v", "SlothModeConnector", "/t", "REG_SZ", "/d", `"`+exe+`"`, "/f")
	} else {
		cmd = exec.Command("reg.exe", "DELETE", startupKey, "/v", "SlothModeConnector", "/f")
	}
	configureHiddenProcess(cmd)
	return cmd.Run()
}
func trayMenu(hwnd uintptr) {
	menu, _, _ := pCreateMenu.Call()
	if menu == 0 {
		return
	}
	defer pDestroyMenu.Call(menu)
	pAppendMenu.Call(menu, mfString, idSetup, wint("Setup & pairing"))
	pAppendMenu.Call(menu, mfString, idOpen, wint("Open website"))
	pAppendMenu.Call(menu, mfString, idPair, wint("Copy pairing key"))
	startupLabel := "Enable launch at sign-in"
	if startWithWindowsEnabled() {
		startupLabel = "Disable launch at sign-in"
	}
	pAppendMenu.Call(menu, mfString, idStartup, wint(startupLabel))
	pAppendMenu.Call(menu, mfSeparator, 0, 0)
	pAppendMenu.Call(menu, mfString, idQuit, wint("Exit connector"))
	var pt winPoint
	pGetPos.Call(uintptr(unsafe.Pointer(&pt)))
	pForeground.Call(hwnd)
	which, _, _ := pTrack.Call(menu, tpmRetCmd|tpmRight, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	switch which {
	case idOpen:
		openSite()
	case idSetup:
		showSetupWindow()
	case idPair:
		if !copyToClipboard(hwnd, winBridge.secret) {
			popup(hwnd, "Sloth Mode", "Could not copy pairing key. Open Setup & pairing to view it.")
		} else {
			popup(hwnd, "Sloth Mode", "Pairing key copied. Paste it into sloth.marceli.dev, and do not share it.")
		}
	case idStartup:
		enabled := startWithWindowsEnabled()
		if err := changeStartWithWindows(!enabled); err != nil {
			popup(hwnd, "Sloth Mode", "Could not change sign-in setting: "+err.Error())
		} else {
			if enabled {
				popup(hwnd, "Sloth Mode", "Launch at sign-in disabled.")
			} else {
				popup(hwnd, "Sloth Mode", "Sloth Mode will now run at sign-in.")
			}
		}
	case idQuit:
		pDestroy.Call(hwnd)
	}
}

var trayCallback = syscall.NewCallback(func(hwnd, msg, w, l uintptr) uintptr {
	switch msg {
	case wmTray:
		if l == wmLButtonUp {
			showSetupWindow()
			return 0
		}
		if l == wmRButtonUp || l == wmContextMenu {
			trayMenu(hwnd)
			return 0
		}
	case wmClose:
		pDestroy.Call(hwnd)
		return 0
	case wmDestroy:
		if trayReady {
			pNotify.Call(nimDelete, uintptr(unsafe.Pointer(&trayData)))
		}
		if winServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = winServer.Shutdown(ctx)
			cancel()
		}
		pQuit.Call(0)
		return 0
	}
	v, _, _ := pDef.Call(hwnd, msg, w, l)
	return v
})

func startupError(msg string) { popup(0, "Sloth Mode connector", msg) }

//go:embed sloth.ico
var embeddedIcon []byte

func runTray(b *Bridge, s *http.Server) {
	// The native hidden tray window and visible pairing window share a Win32
	// message pump; creating it and GetMessage must run on the same OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	winBridge = b
	winServer = s
	cls := "SlothModeConnectorTray"
	instance, _, _ := pModule.Call(0)
	wc := winClass{Size: uint32(unsafe.Sizeof(winClass{})), WndProc: trayCallback, Instance: instance, ClassName: wint(cls)}
	var hwnd uintptr
	if r, _, _ := pRegister.Call(uintptr(unsafe.Pointer(&wc))); r != 0 {
		hwnd, _, _ = pCreate.Call(0, wint(cls), wint("Sloth Mode Connector"), 0, 0, 0, 0, 0, 0, 0, instance, 0)
	}
	if hwnd == 0 {
		logStartup("Tray owner window unavailable; trying independent Setup & Pairing window")
	} else {
		// The notification area is optional. The visible setup window is the primary UI.
		icon, _, _ := pLoadIcon.Call(0, 32512)
		if err := os.WriteFile(filepath.Join(b.base, "sloth-tray.ico"), embeddedIcon, 0600); err == nil {
			const imageIcon = 1
			const loadFromFile = 0x10
			custom, _, _ := u32.NewProc("LoadImageW").Call(0, wint(filepath.Join(b.base, "sloth-tray.ico")), imageIcon, 0, 0, loadFromFile)
			if custom != 0 {
				icon = custom
			}
		}
		trayData = notifyIcon{Size: uint32(unsafe.Sizeof(notifyIcon{})), Hwnd: hwnd, ID: 1, Flags: nifMessage | nifIcon | nifTip, CallbackMessage: wmTray, Icon: icon}
		copy(trayData.Tip[:], syscall.StringToUTF16("Sloth Mode - local AI connector"))
		res, _, _ := pNotify.Call(nimAdd, uintptr(unsafe.Pointer(&trayData)))
		trayReady = res != 0
		if !trayReady {
			logStartup("Windows could not add the tray icon; setup window remains accessible")
		}
	}
	if !createSetupWindow(instance) {
		if hwnd != 0 {
			pDestroy.Call(hwnd)
		}
		startupError("Could not create the pairing window. See connector.log in AppData.")
		return
	}
	var m winMessage
	for {
		ret, _, _ := pGet.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		pTranslate.Call(uintptr(unsafe.Pointer(&m)))
		pDispatch.Call(uintptr(unsafe.Pointer(&m)))
	}
}
