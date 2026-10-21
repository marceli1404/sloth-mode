//go:build windows

package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	_ "embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

//go:embed payload/SlothMode-Connector.exe
var connector []byte

const (
	wsOverlappedWindow = 0x00CF0000
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
	wmConnectorChecked = 0x8000 + 41
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
	user32           = syscall.NewLazyDLL("user32.dll")
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	gdi32            = syscall.NewLazyDLL("gdi32.dll")
	registerClass    = user32.NewProc("RegisterClassExW")
	createWindow     = user32.NewProc("CreateWindowExW")
	defProc          = user32.NewProc("DefWindowProcW")
	getMessage       = user32.NewProc("GetMessageW")
	translateMessage = user32.NewProc("TranslateMessage")
	dispatchMessage  = user32.NewProc("DispatchMessageW")
	showWindow       = user32.NewProc("ShowWindow")
	updateWindow     = user32.NewProc("UpdateWindow")
	destroyWindow    = user32.NewProc("DestroyWindow")
	postQuitMessage  = user32.NewProc("PostQuitMessage")
	setWindowText    = user32.NewProc("SetWindowTextW")
	sendMessage      = user32.NewProc("SendMessageW")
	getFont          = gdi32.NewProc("GetStockObject")
	messageBox       = user32.NewProc("MessageBoxW")
	getModule        = kernel32.NewProc("GetModuleHandleW")
	getMetrics       = user32.NewProc("GetSystemMetrics")
	postMessage      = user32.NewProc("PostMessageW")
	hwnd             uintptr
	installButton    uintptr
	statusLabel      uintptr
	pathLabel        uintptr
	startupButton    uintptr
	installedPath    string
	installed        bool
	launchOnSignIn   bool
	installStage     int
	updateMu         sync.Mutex
	pendingUpdate    updateManifest
	updateError      error
	updateButton     uintptr
	graphicsLabel    uintptr
	graphicsMu       sync.Mutex
	graphicsText     string
	launchMu         sync.Mutex
	launchReport     string
	launchOK         bool
)

func utf16(s string) uintptr {
	p, _ := syscall.UTF16PtrFromString(s)
	return uintptr(unsafe.Pointer(p))
}
func notify(title, msg string) { messageBox.Call(hwnd, utf16(msg), utf16(title), 0x40) }
func ctrl(typ, txt string, x, y, w, h int32, style uint32, id uintptr) uintptr {
	instance, _, _ := getModule.Call(0)
	hWnd, _, _ := createWindow.Call(0, utf16(typ), utf16(txt), uintptr(wsVisible|wsChild|style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), hwnd, id, instance, 0)
	if hWnd != 0 {
		font := bodyFont
		if len(txt) > 0 && (strings.HasPrefix(txt, "Your ") || strings.HasPrefix(txt, "A quieter")) {
			font = titleFont
		}
		if strings.HasPrefix(txt, "STEP") || strings.HasPrefix(txt, "SLOTH MODE") {
			font = smallFont
		}
		sendMessage.Call(hWnd, wmSetFont, font, 1)
	}
	return hWnd
}
func status(s string) {
	if statusLabel != 0 {
		setWindowText.Call(statusLabel, utf16(s))
		updateWindow.Call(statusLabel)
	}
}
func startChoice() string {
	if launchOnSignIn {
		return "[x] Launch at Windows sign-in (optional)"
	}
	return "[ ] Launch at Windows sign-in (optional)"
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
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Start menu shortcut: %v %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
func priorStartupEnabled() bool {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	cmd := exec.Command("reg.exe", "QUERY", key, "/v", "SlothModeConnector")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	out, err := cmd.Output()
	return err == nil && strings.Contains(strings.ToLower(string(out)), "slothmodeconnector")
}

func changeStartup(exe string, enabled bool) error {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	var cmd *exec.Cmd
	if enabled {
		cmd = exec.Command("reg.exe", "ADD", key, "/v", "SlothModeConnector", "/t", "REG_SZ", "/d", `"`+exe+`"`, "/f")
	} else {
		cmd = exec.Command("reg.exe", "DELETE", key, "/v", "SlothModeConnector", "/f")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	return cmd.Run()
}
func install() error {
	dest, err := appTarget()
	if err != nil {
		return err
	}
	// Never clobber a running connector. Windows will not safely replace it.
	if c, e := net.DialTimeout("tcp4", "127.0.0.1:39871", 250*time.Millisecond); e == nil {
		c.Close()
		return fmt.Errorf("Sloth Mode is running. Exit it from the tray or Task Manager before updating.")
	}
	if len(connector) < 1_000_000 {
		return fmt.Errorf("embedded application is missing")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	// Never write to the previous app's private notes / keys, only to Program files.
	newPath := dest + ".new"
	if err := os.WriteFile(newPath, connector, 0700); err != nil {
		return err
	}
	defer os.Remove(newPath)
	data, err := os.ReadFile(newPath)
	if err != nil {
		return err
	}
	if sha256.Sum256(data) != sha256.Sum256(connector) {
		return fmt.Errorf("installed executable did not pass integrity check")
	}
	// Preserve the last installed binary until the new version is in place.
	backup := dest + ".previous"
	hadOld := false
	if _, statErr := os.Stat(dest); statErr == nil {
		_ = os.Remove(backup)
		if e := os.Rename(dest, backup); e != nil {
			return fmt.Errorf("close the old connector before upgrading: %w", e)
		}
		hadOld = true
	}
	if err := os.Rename(newPath, dest); err != nil {
		if hadOld {
			_ = os.Rename(backup, dest)
		}
		return fmt.Errorf("installation failed; previous version restored: %w", err)
	}
	if hadOld {
		_ = os.Remove(backup)
	}
	installedPath = dest
	installed = true
	return nil
}
func beginInstall() {
	installStage = 1
	updateWindow.Call(hwnd)
	status("Installing Sloth Mode for your Windows account...")
	setWindowText.Call(installButton, utf16("Installing..."))
	err := install()
	if err != nil {
		status("Installation failed. Previous local notes were not touched.")
		notify("Install error", err.Error())
		setWindowText.Call(installButton, utf16("Try again"))
		return
	}
	warning := ""
	if e := makeShortcut(installedPath); e != nil {
		warning = "Start menu shortcut not created; app is installed. "
	}
	if launchOnSignIn {
		// Migrate any old Downloads-path startup entry to the installed program.
		if e := changeStartup(installedPath, true); e != nil {
			warning += "Could not enable launch at sign-in. "
		}
	} else if priorStartupEnabled() {
		// The user explicitly deselected startup; remove the old entry.
		if e := changeStartup(installedPath, false); e != nil {
			warning += "Could not remove the previous launch at sign-in setting. "
		}
	}
	installStage = 2
	updateWindow.Call(hwnd)
	status("Installed successfully — opening your pairing window...")
	setWindowText.Call(installButton, utf16("Open pairing window"))
	launchInstalled()
	if warning != "" {
		notify("Installed with notice", warning)
	}
}
func beginCheckUpdates() {
	status("Checking the signed update channel...")
	setWindowText.Call(updateButton, utf16("Checking..."))
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
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
	updateMu.Lock()
	m, e := pendingUpdate, updateError
	updateMu.Unlock()
	setWindowText.Call(updateButton, utf16("Check updates"))
	if e != nil {
		status("Could not verify updates. Your installed connector is unchanged.")
		notify("Updater", e.Error())
		return
	}
	newer, e := newerVersion(installerVersion, m.Version)
	if e != nil || !newer {
		status("You're up to date. Signed channel: " + installerVersion)
		return
	}
	answer, _, _ := messageBox.Call(hwnd, utf16("Signed update "+m.Version+" is available. Download and run the verified installer?\n\nThe existing connector and your notes will not be replaced until you approve the installer."), utf16("Sloth Mode update"), 0x24) // MB_ICONQUESTION | MB_YESNO
	if answer != 6 {
		return
	}
	status("Downloading and validating the update...")
	setWindowText.Call(updateButton, utf16("Downloading..."))
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		defer cancel()
		path, downloadErr := downloadInstaller(ctx, m)
		updateMu.Lock()
		updateError = downloadErr
		if downloadErr == nil {
			updateDownloadPath = path
		}
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
	setWindowText.Call(updateButton, utf16("Check updates"))
	if e != nil {
		status("Update not installed; verification failed.")
		notify("Updater", e.Error())
		return
	}
	if e := runVerifiedInstaller(path); e != nil {
		status("Verified update could not launch.")
		notify("Updater", e.Error())
		return
	}
	status("Verified update installer opened. Close your running connector when prompted.")
}
// Poll the real localhost health route; starting a process is not proof it is ready.
// Work and process waiting are background-only, and UI changes use WM_APP messages.
func probeLocalCompanion() bool {
	return probeCompanionHealth(&http.Client{Timeout: 500 * time.Millisecond}, "http://127.0.0.1:39871/v1/health")
}
// A listening port alone cannot establish that this is our connector.
func probeCompanionHealth(client *http.Client, url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil { return false }
	req.Header.Set("Origin", "https://sloth.marceli.dev")
	resp, err := client.Do(req)
	if err != nil { return false }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(resp.Header.Get("Content-Type"), "application/json") { return false }
	var body struct { OK bool `json:"ok"`; Name string `json:"name"` }
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2048)).Decode(&body); err != nil { return false }
	return body.OK && body.Name == "Sloth Mode Connector"
}
func launchInstalled() {
	connection, err := net.DialTimeout("tcp4", "127.0.0.1:39871", 350*time.Millisecond)
	if err == nil {
		_ = connection.Close()
		status("Another connector is already running. Close it before opening the updated version.")
		notify("Previous connector is still running", "Exit the old Sloth Mode connector from the tray or Task Manager, then click Open pairing window. Local notes and keys are preserved.")
		return
	}
	status("Starting connector and checking its local service…")
	cmd := exec.Command(installedPath)
	if e := cmd.Start(); e != nil {
		status("Windows could not launch the connector. Check security prompts.")
		notify("Launch error", "The application could not start: "+e.Error())
		return
	}
	owner := hwnd
	go func() {
		exited := make(chan error, 1)
		go func(){ exited <- cmd.Wait() }()
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		report := "Connector did not become ready. See %LOCALAPPDATA%\\SlothModeNative\\connector.log"
		ready := false
		for !ready {
			select {
			case err := <-exited:
				if err != nil { report = "Connector exited during startup: " + err.Error() + ". Check connector.log." } else { report = "Connector exited before pairing was ready. Check connector.log." }
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
	case wmPaint:
		paintBackground(h)
		return 0
	case wmEraseBkgnd:
		return 1
	case wmCtlColorStatic:
		return setStaticColors(w)
	case wmDrawItem:
		return paintButton(l)
	case wmConnectorChecked:
		launchMu.Lock()
		report, ready := launchReport, launchOK
		launchMu.Unlock()
		status(report)
		if !ready { notify("Connector startup", report) }
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
		case idInstall:
			if installed {
				launchInstalled()
			} else {
				beginInstall()
			}
		case idClose:
			destroyWindow.Call(h)
		case idOpenLocation:
			openFolder()
		case idCheckUpdates:
			beginCheckUpdates()
		case idStartup:
			launchOnSignIn = !launchOnSignIn
			setWindowText.Call(startupButton, utf16(startChoice()))
		}
		return 0
	case wmClose:
		destroyWindow.Call(h)
		return 0
	case wmDestroy:
		postQuitMessage.Call(0)
		return 0
	}
	r, _, _ := defProc.Call(h, msg, w, l)
	return r
})

func main() {
	// A Win32 window and its GetMessage loop must stay on the creating OS thread.
	// A normal Go goroutine can otherwise migrate to a different OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	makeTheme()
	instance, _, _ := getModule.Call(0)
	cls := "SlothModeInstallWindow"
	wc := winClass{Size: uint32(unsafe.Sizeof(winClass{})), WndProc: wndproc, Instance: instance, ClassName: utf16(cls), Background: 6}
	if r, _, _ := registerClass.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		notify("Sloth Mode", "Cannot create installer window.")
		return
	}
	sx, _, _ := getMetrics.Call(0)
	sy, _, _ := getMetrics.Call(1)
	x, y := int32(90), int32(90)
	if sx > 800 {
		x = (int32(sx) - 760) / 2
	}
	if sy > 760 {
		y = (int32(sy) - 715) / 2
	}
	h, _, _ := createWindow.Call(0, utf16(cls), utf16("Sloth Mode • Setup"), wsOverlappedWindow, uintptr(x), uintptr(y), 760, 715, 0, 0, instance, 0)
	if h == 0 {
		notify("Sloth Mode", "Could not open installer window.")
		return
	}
	hwnd = h
	ctrl("STATIC", "SLOTH MODE "+installerVersion+"   /   DESKTOP COMPANION", 42, 30, 480, 26, 0, 0)
	ctrl("STATIC", "A quieter way to get things done.", 42, 74, 590, 47, 0, 0)
	ctrl("STATIC", "Your own Gemma AI, running on this Windows PC.", 42, 130, 555, 26, 0, 0)
	ctrl("STATIC", "STEP 01   LOCAL-FIRST INSTALL", 42, 219, 625, 23, 0, 0)
	ctrl("STATIC", "Your private companion stays in the notification area.", 42, 250, 625, 27, 0, 0)
	ctrl("STATIC", "STEP 02   PAIR ONCE", 42, 301, 625, 23, 0, 0)
	ctrl("STATIC", "Each trusted browser gets its own key; you can revoke it.", 42, 331, 625, 28, 0, 0)
	ctrl("STATIC", "STEP 03   OPTIONAL GEMMA", 42, 365, 625, 25, 0, 0)
	ctrl("STATIC", "Download the model only when you're ready. No Ollama needed.", 42, 391, 625, 25, 0, 0)
	ctrl("STATIC", installerRecommendation(), 42, 418, 650, 24, 0, 0)
	graphicsLabel = ctrl("STATIC", "Detecting graphics hardware locally...", 42, 440, 650, 22, 0, 0)
	go func() {
		name := installerGPUName()
		graphicsMu.Lock()
		graphicsText = name
		graphicsMu.Unlock()
		postMessage.Call(hwnd, wmHardwareReady, 0, 0)
	}()
	launchOnSignIn = priorStartupEnabled()
	startupButton = ctrl("BUTTON", startChoice(), 42, 478, 430, 37, wsTabstop|bsOwnerDraw, idStartup)
	statusLabel = ctrl("STATIC", "Ready to install. Existing local notes stay untouched.", 42, 520, 665, 24, 0, 0)
	installButton = ctrl("BUTTON", "Install & open", 42, 562, 232, 42, wsTabstop|bsOwnerDraw, idInstall)
	ctrl("BUTTON", "Open app folder", 287, 562, 210, 42, wsTabstop|bsOwnerDraw, idOpenLocation)
	updateButton = ctrl("BUTTON", "Check updates", 509, 562, 210, 42, wsTabstop|bsOwnerDraw, idCheckUpdates)
	ctrl("BUTTON", "Close", 509, 614, 210, 36, wsTabstop|bsOwnerDraw, idClose)
	showWindow.Call(hwnd, swShow)
	updateWindow.Call(hwnd)
	var message winMessage
	for {
		r, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&message)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&message)))
	}
}
