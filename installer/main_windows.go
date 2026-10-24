//go:build windows

package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

//go:embed payload/SlothMode-Connector.exe
var connector []byte

const (
	wsOverlappedWindow = 0x00CA0000 // Caption, system menu, minimize; fixed native layout.
	wsVisible          = 0x10000000
	wsChild            = 0x40000000
	wsTabstop          = 0x00010000
	wmDestroy          = 0x0002
	wmClose            = 0x0010
	wmCommand          = 0x0111
	wmSetFont          = 0x0030
	swShow             = 5
	idInstall          = 301
	idClose            = 302
	idOpenLocation     = 303
	idStartup          = 304
	idCheckUpdates     = 305
	wmUpdateReady      = 0x8000 + 22
	wmHardwareReady    = 0x8000 + 40
	wmInstallReady     = 0x8000 + 41
	wmConnectorChecked = 0x8000 + 42
)

type winClass struct {
	Size                                                               uint32
	Style                                                              uint32
	WndProc                                                            uintptr
	ClassExtra, WindowExtra                                            int32
	Instance, Icon, Cursor, Background, MenuName, ClassName, SmallIcon uintptr
}
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

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	registerClass        = user32.NewProc("RegisterClassExW")
	createWindow         = user32.NewProc("CreateWindowExW")
	defProc              = user32.NewProc("DefWindowProcW")
	getMessage           = user32.NewProc("GetMessageW")
	translateMessage     = user32.NewProc("TranslateMessage")
	dispatchMessage      = user32.NewProc("DispatchMessageW")
	showWindow           = user32.NewProc("ShowWindow")
	updateWindow         = user32.NewProc("UpdateWindow")
	destroyWindow        = user32.NewProc("DestroyWindow")
	postQuitMessage      = user32.NewProc("PostQuitMessage")
	setWindowText        = user32.NewProc("SetWindowTextW")
	sendMessage          = user32.NewProc("SendMessageW")
	getFont              = gdi32.NewProc("GetStockObject")
	messageBox           = user32.NewProc("MessageBoxW")
	getModule            = kernel32.NewProc("GetModuleHandleW")
	getMetrics           = user32.NewProc("GetSystemMetrics")
	postMessage          = user32.NewProc("PostMessageW")
	hwnd                 uintptr
	installButton        uintptr
	statusLabel          uintptr
	startupButton        uintptr
	installedPath        string
	installed            bool
	launchOnSignIn       bool
	installStage         int
	updateMu             sync.Mutex
	pendingUpdate        updateManifest
	updateError          error
	updateButton         uintptr
	graphicsLabel        uintptr
	graphicsMu           sync.Mutex
	graphicsText         string
	operationBusy        bool // UI thread owns operation state and all window updates.
	operationCancel      context.CancelFunc
	closeRequested       bool
	installMu            sync.Mutex
	pendingInstall       installResult
	launchMu             sync.Mutex
	launchReport         string
	launchOK             bool
	installerSession     uintptr
	installerOwnsSession bool
)

func utf16(s string) uintptr {
	p, _ := syscall.UTF16PtrFromString(s)
	return uintptr(unsafe.Pointer(p))
}
func notify(title, msg string) { messageBox.Call(hwnd, utf16(msg), utf16(title), 0x40) }
func ctrl(typ, txt string, x, y, w, h int32, style uint32, id uintptr) uintptr {
	instance, _, _ := getModule.Call(0)
	hWnd, _, _ := createWindow.Call(0, utf16(typ), utf16(txt), uintptr(wsVisible|wsChild|style), uintptr(px(x)), uintptr(px(y)), uintptr(px(w)), uintptr(px(h)), hwnd, id, instance, 0)
	if hWnd != 0 {
		controls = append(controls, controlLayout{hWnd, x, y, w, h, "body"})
		sendMessage.Call(hWnd, wmSetFont, bodyFont, 1)
	}
	return hWnd
}
func status(s string) {
	if statusLabel != 0 {
		setWindowText.Call(statusLabel, utf16(s))
		updateWindow.Call(statusLabel)
	}
}
func startChoice() string { return "Launch at Windows sign-in (optional)" }

// Keep concurrent checks, downloads and installs from sharing result slots.
func beginOperation() bool {
	if operationBusy {
		return false
	}
	operationBusy = true
	for _, control := range []uintptr{installButton, updateButton, startupButton} {
		user32.NewProc("EnableWindow").Call(control, 0)
	}
	return true
}
func finishOperation() bool {
	if operationCancel != nil {
		operationCancel()
		operationCancel = nil
	}
	operationBusy = false
	for _, control := range []uintptr{installButton, updateButton, startupButton} {
		enabled := uintptr(1)
		if control == startupButton && installed {
			enabled = 0
		}
		user32.NewProc("EnableWindow").Call(control, enabled)
	}
	refreshInstallerStage()
	if closeRequested {
		destroyWindow.Call(hwnd)
		return true
	}
	return false
}
func requestClose() {
	if operationBusy {
		closeRequested = true
		if operationCancel != nil {
			operationCancel()
		}
		status("Closing after the current operation finishes...")
		return
	}
	destroyWindow.Call(hwnd)
}

func appTarget() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", fmt.Errorf("LOCALAPPDATA is missing")
	}
	return filepath.Join(base, "Programs", "Sloth Mode", "SlothMode-Connector.exe"), nil
}
func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
func makeShortcut(exe string) error {
	base := os.Getenv("APPDATA")
	if base == "" {
		return fmt.Errorf("APPDATA is missing")
	}
	folder := filepath.Join(base, "Microsoft", "Windows", "Start Menu", "Programs")
	if err := os.MkdirAll(folder, 0700); err != nil {
		return err
	}
	dest := filepath.Join(folder, "Sloth Mode Connector.lnk")
	script := "$w=New-Object -ComObject WScript.Shell;" +
		"$s=$w.CreateShortcut(" + psQuote(dest) + ");" +
		"$s.TargetPath=" + psQuote(exe) + ";" +
		"$s.WorkingDirectory=" + psQuote(filepath.Dir(exe)) + ";" +
		"$s.Description='Sloth Mode local AI connector';$s.Save()"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Start menu shortcut: %v %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
func priorStartupEnabled() bool {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "reg.exe", "QUERY", key, "/v", "SlothModeConnector")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.Output()
	return err == nil && strings.Contains(strings.ToLower(string(out)), "slothmodeconnector")
}

func changeStartup(exe string, enabled bool) error {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	if enabled {
		cmd = exec.CommandContext(ctx, "reg.exe", "ADD", key, "/v", "SlothModeConnector", "/t", "REG_SZ", "/d", `"`+exe+`"`, "/f")
	} else {
		cmd = exec.CommandContext(ctx, "reg.exe", "DELETE", key, "/v", "SlothModeConnector", "/f")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Run()
}

type installResult struct {
	path, warning string
	err           error
}

func install() (string, error) {
	dest, err := appTarget()
	if err != nil {
		return "", err
	}
	if c, e := net.DialTimeout("tcp4", "127.0.0.1:39871", 250*time.Millisecond); e == nil {
		c.Close()
		return "", fmt.Errorf("Port 39871 is in use. Exit Sloth Mode from the tray or Task Manager before updating. If it is already closed, another app may be using that port.")
	}
	if err := installApplication(dest, connector); err != nil {
		return "", err
	}
	return dest, nil
}
func beginInstall() {
	if !beginOperation() {
		return
	}
	installStage = 1
	refreshInstallerStage()
	status("Installing Sloth Mode for your Windows account...")
	setWindowText.Call(installButton, utf16("Installing..."))
	startup := launchOnSignIn
	go func() {
		path, err := install()
		result := installResult{path: path, err: err}
		if err == nil {
			if e := makeShortcut(path); e != nil {
				result.warning = "Start menu shortcut not created; app is installed. "
			}
			if startup {
				if e := changeStartup(path, true); e != nil {
					result.warning += "Could not enable launch at sign-in. "
				}
			} else if priorStartupEnabled() {
				if e := changeStartup(path, false); e != nil {
					result.warning += "Could not remove the previous launch at sign-in setting. "
				}
			}
		}
		installMu.Lock()
		pendingInstall = result
		installMu.Unlock()
		postMessage.Call(hwnd, wmInstallReady, 0, 0)
	}()
}
func finishInstall() {
	installMu.Lock()
	result := pendingInstall
	installMu.Unlock()
	if finishOperation() {
		return
	}
	if result.err != nil {
		installStage = 0
		refreshInstallerStage()
		status("Installation failed. Your notes were not changed.")
		notify("Install error", result.err.Error())
		setWindowText.Call(installButton, utf16("Try again"))
		return
	}
	installedPath, installed = result.path, true
	user32.NewProc("EnableWindow").Call(startupButton, 0)
	installStage = 2
	refreshInstallerStage()
	setWindowText.Call(installButton, utf16("Open pairing window"))
	launchInstalled()
	if result.warning != "" {
		notify("Installed with notice", result.warning)
	}
}
func beginCheckUpdates() {
	if !beginOperation() {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	operationCancel = cancel
	setWindowText.Call(stepLabel, utf16("CHECKING UPDATES"))
	status("Checking the signed update channel...")
	setWindowText.Call(updateButton, utf16("Checking..."))
	go func() {
		defer cancel()
		m, e := fetchUpdate(ctx)
		updateMu.Lock()
		pendingUpdate = m
		updateError = e
		updateMu.Unlock()
		postMessage.Call(hwnd, wmUpdateReady, 0, 0)
	}()
}
func finishCheckUpdates() {
	if closeRequested {
		finishOperation()
		return
	}
	updateMu.Lock()
	m, e := pendingUpdate, updateError
	updateMu.Unlock()
	setWindowText.Call(updateButton, utf16("Check updates"))
	if e != nil {
		finishOperation()
		status("Could not verify updates. Your installed connector is unchanged.")
		notify("Updater", e.Error())
		return
	}
	newer, e := newerVersion(installerVersion, m.Version)
	if e != nil || !newer {
		finishOperation()
		status("No newer update in the signed channel (" + m.Version + "). This installer: " + installerVersion + ".")
		return
	}
	answer, _, _ := messageBox.Call(hwnd, utf16("Signed update "+m.Version+" is available. Download and run the verified installer?\n\nThe existing connector and your notes will not be replaced until you approve the installer."), utf16("Sloth Mode update"), 0x24) // MB_ICONQUESTION | MB_YESNO
	if answer != 6 {
		finishOperation()
		status("Update skipped. Your installed connector is unchanged.")
		return
	}
	setWindowText.Call(stepLabel, utf16("DOWNLOADING UPDATE"))
	status("Downloading and validating the update...")
	setWindowText.Call(updateButton, utf16("Downloading..."))
	if operationCancel != nil {
		operationCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	operationCancel = cancel
	go func() {
		defer cancel()
		path, downloadErr := downloadInstaller(ctx, m)
		updateMu.Lock()
		updateError = downloadErr
		updateDownloadPath = path
		updateMu.Unlock()
		postMessage.Call(hwnd, wmDownloadReady, 0, 0)
	}()
}

const wmDownloadReady = 0x8000 + 23

var updateDownloadPath string

func finishDownload() {
	updateMu.Lock()
	path, e := updateDownloadPath, updateError
	updateMu.Unlock()
	if finishOperation() {
		discardDownloadedInstaller(path)
		return
	}
	setWindowText.Call(updateButton, utf16("Check updates"))
	if e != nil {
		status("Update not installed; verification failed.")
		notify("Updater", e.Error())
		discardDownloadedInstaller(path)
		return
	}
	// Let the verified successor take the session before starting it. Keeping
	// our singleton lock here would make every downloaded installer reject itself.
	releaseInstallerSession()
	if e := runVerifiedInstaller(path); e != nil {
		wait, _, _ := kernel32.NewProc("WaitForSingleObject").Call(installerSession, 0)
		installerOwnsSession = wait == 0 || wait == 0x80
		status("Verified update could not launch.")
		notify("Updater", e.Error())
		discardDownloadedInstaller(path)
		if !installerOwnsSession {
			destroyWindow.Call(hwnd)
		}
		return
	}
	destroyWindow.Call(hwnd)
}

func discardDownloadedInstaller(path string) {
	if path != "" {
		_ = os.Remove(path)
		_ = os.Remove(filepath.Dir(path)) // Only the now-empty download directory.
	}
}

func releaseInstallerSession() {
	if installerOwnsSession {
		kernel32.NewProc("ReleaseMutex").Call(installerSession)
		installerOwnsSession = false
	}
}

// Poll the real localhost health route; starting a process is not proof it is ready.
// Work and process waiting are background-only, and UI changes use WM_APP messages.
func probeLocalCompanion() bool {
	return probeCompanionHealth(&http.Client{Timeout: 500 * time.Millisecond}, "http://127.0.0.1:39871/v1/health")
}

// A listening port alone cannot establish that this is our connector.
func probeCompanionHealth(client *http.Client, url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Origin", "https://sloth.marceli.dev")
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return false
	}
	var body struct {
		OK   bool   `json:"ok"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2048)).Decode(&body); err != nil {
		return false
	}
	return body.OK && body.Name == "Sloth Mode Connector"
}
func launchInstalled() {
	connection, err := net.DialTimeout("tcp4", "127.0.0.1:39871", 350*time.Millisecond)
	if err == nil {
		_ = connection.Close()
		status("A companion is already running. Open its Setup & Pairing window from the tray.")
		notify("Previous connector is still running", "Exit the old Sloth Mode connector from the tray or Task Manager, then click Open pairing window. Local notes and keys are preserved.")
		return
	}
	status("Starting connector and checking its local service…")
	cmd := exec.Command(installedPath)
	if e := cmd.Start(); e != nil {
		finishOperation()
		status("Windows could not launch the companion. Check security prompts.")
		notify("Launch error", "The application could not start: "+e.Error())
		return
	}
	owner := hwnd
	go func() {
		exited := make(chan error, 1)
		go func() { exited <- cmd.Wait() }()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		report := "Connector did not become ready. See %LOCALAPPDATA%\\SlothModeNative\\connector.log"
		ready := false
		for !ready {
			select {
			case err := <-exited:
				if err != nil {
					report = "Connector exited during startup: " + err.Error() + ". Check connector.log."
				} else {
					report = "Connector exited before pairing was ready. Check connector.log."
				}
				goto finished
			case <-ticker.C:
				if probeLocalCompanion() {
					report = "Connector is running. Open Setup & Pairing and copy your browser key."
					ready = true
				}
			case <-timer.C:
				goto finished
			}
		}
	finished:
		launchMu.Lock()
		launchReport, launchOK = report, ready
		launchMu.Unlock()
		postMessage.Call(owner, wmConnectorChecked, 0, 0)
	}()
}

func openFolder() {
	dest, e := appTarget()
	if e != nil {
		return
	}
	folder := filepath.Dir(dest)
	cmd := exec.Command("explorer.exe", folder)
	_ = cmd.Start()
}

var wndproc = syscall.NewCallback(func(h, msg, w, l uintptr) uintptr {
	switch msg {
	case 0x02E0: // WM_DPICHANGED; keep all controls within this monitor.
		resizeForDPI(uint32(w&0xffff), monitorWorkArea(h))
		return 0
	case wmPaint:
		paintBackground(h)
		return 0
	case wmEraseBkgnd:
		return 1
	case wmCtlColorStatic, 0x0135: // WM_CTLCOLORBTN for the native checkbox.
		return setStaticColors(w, l)
	case wmDrawItem:
		return paintButton(l)
	case wmInstallReady:
		finishInstall()
		return 0
	case wmConnectorChecked:
		if finishOperation() {
			return 0
		}
		launchMu.Lock()
		report, ready := launchReport, launchOK
		launchMu.Unlock()
		status(report)
		if ready {
			installStage = 3
			refreshInstallerStage()
		} else {
			notify("Connector startup", report)
		}
		return 0
	case wmHardwareReady:
		graphicsMu.Lock()
		v := graphicsText
		graphicsMu.Unlock()
		setWindowText.Call(graphicsLabel, utf16(v))
		return 0
	case wmUpdateReady:
		finishCheckUpdates()
		return 0
	case wmDownloadReady:
		finishDownload()
		return 0
	case wmCommand:
		switch uint16(w & 0xffff) {
		case idInstall, 1: // IDOK from keyboard dialog navigation.
			if operationBusy {
				return 0
			}
			if installed {
				launchInstalled()
			} else {
				beginInstall()
			}
		case idClose, 2: // IDCANCEL (Escape).
			requestClose()
		case idOpenLocation:
			openFolder()
		case idCheckUpdates:
			beginCheckUpdates()
		case idStartup:
			if operationBusy {
				return 0
			}
			launchOnSignIn = !launchOnSignIn
			checked := uintptr(0)
			if launchOnSignIn {
				checked = 1
			}
			sendMessage.Call(startupButton, 0x00F1, checked, 0)
		}
		return 0
	case wmClose:
		requestClose()
		return 0
	case wmDestroy:
		postQuitMessage.Call(0)
		return 0
	}
	r, _, _ := defProc.Call(h, msg, w, l)
	return r
})

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	// A second installer must not race the first while replacing the same app.
	mutex, err := acquireInstallerSession(`Local\SlothModeInstaller`)
	if err != nil {
		notify("Sloth Mode", err.Error())
		return
	}
	defer kernel32.NewProc("CloseHandle").Call(mutex)
	installerSession, installerOwnsSession = mutex, true
	defer releaseInstallerSession()
	if err := createInstallerUI(); err != nil {
		notify("Sloth Mode", err.Error())
		return
	}
	defer releaseTheme()
	runMessageLoop()
}

func createInstallerUI() error {
	// Prefer per-monitor DPI v2. Fallback keeps Windows bitmap scaling available.
	aware := user32.NewProc("SetProcessDpiAwarenessContext")
	if aware.Find() == nil {
		aware.Call(^uintptr(3))
	}
	dpi := uint32(96)
	probe := user32.NewProc("GetDpiForSystem")
	if probe.Find() == nil {
		d, _, _ := probe.Call()
		if d > 0 {
			dpi = uint32(d)
		}
	}
	work := monitorWorkArea(0)
	uiScale = fittedScale(dpi, work)
	makeTheme()
	instance, _, _ := getModule.Call(0)
	cls := "SlothModeInstallWindow"
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	wc := winClass{Size: uint32(unsafe.Sizeof(winClass{})), WndProc: wndproc, Instance: instance, Cursor: cursor, Icon: brandIcon, SmallIcon: brandIcon, ClassName: utf16(cls), Background: backgroundBrush}
	if r, _, e := registerClass.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		return fmt.Errorf("cannot create installer window: %v", e)
	}
	r := windowBounds(work, dpi)
	h, _, e := createWindow.Call(0x10000, utf16(cls), utf16("Sloth Mode - Setup"), wsOverlappedWindow, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0, 0, instance, 0)
	if h == 0 {
		return fmt.Errorf("could not open installer: %v", e)
	}
	hwnd = h
	launchOnSignIn = priorStartupEnabled()
	buildInstallerControls()
	go func() {
		name := installerGPUName()
		graphicsMu.Lock()
		graphicsText = name
		graphicsMu.Unlock()
		postMessage.Call(h, wmHardwareReady, 0, 0)
	}()
	showWindow.Call(hwnd, swShow)
	updateWindow.Call(hwnd)
	user32.NewProc("SetFocus").Call(installButton)
	return nil
}

func runMessageLoop() {
	var message winMessage
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		if handled, _, _ := user32.NewProc("IsDialogMessageW").Call(hwnd, uintptr(unsafe.Pointer(&message))); handled != 0 {
			continue
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&message)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
}
