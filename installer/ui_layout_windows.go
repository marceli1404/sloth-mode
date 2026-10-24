//go:build windows

package main

import (
	"math"
	"unsafe"
)

const clientWidth, clientHeight = 820, 628

type controlLayout struct {
	handle              uintptr
	x, y, width, height int32
	role                string
}

var uiScale = 1.0
var controls []controlLayout
var controlRoles = map[uintptr]string{}
var headingLabel, stepLabel uintptr

func px(value int32) int32 { return int32(math.Round(float64(value) * uiScale)) }

// Include space for the title bar and borders when fitting smaller work areas.
func fittedScale(dpi uint32, work rectWin) float64 {
	scale := float64(dpi) / 96
	if scale <= 0 {
		scale = 1
	}
	width, height := float64(work.Right-work.Left), float64(work.Bottom-work.Top)
	if width > 0 && height > 0 {
		frame := rectWin{}
		adjustForDPI(&frame, dpi)
		scale = math.Min(scale, math.Min((width-float64(frame.Right-frame.Left)-16)/clientWidth, (height-float64(frame.Bottom-frame.Top)-16)/clientHeight))
	}
	return scale
}

func windowBounds(work rectWin, dpi uint32) rectWin {
	r := rectWin{0, 0, px(clientWidth), px(clientHeight)}
	adjustForDPI(&r, dpi)
	w, h := r.Right-r.Left, r.Bottom-r.Top
	x, y := work.Left+(work.Right-work.Left-w)/2, work.Top+(work.Bottom-work.Top-h)/2
	return rectWin{x, y, x + w, y + h}
}

func adjustForDPI(r *rectWin, dpi uint32) {
	adjust := user32.NewProc("AdjustWindowRectExForDpi")
	if adjust.Find() == nil {
		adjust.Call(uintptr(unsafe.Pointer(r)), wsOverlappedWindow, 0, 0x10000, uintptr(dpi))
	} else {
		user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(r)), wsOverlappedWindow, 0, 0x10000)
	}
}

func textControl(text, role string, x, y, width, height int32) uintptr {
	h := ctrl("STATIC", text, x, y, width, height, 0x80, 0)
	controlRoles[h] = role
	controls[len(controls)-1].role = role
	sendMessage.Call(h, wmSetFont, fontForRole(role), 1)
	return h
}

func fontForRole(role string) uintptr {
	switch role {
	case "title", "rail-title":
		return titleFont
	case "heading":
		return headingFont
	case "eyebrow", "rail-small", "muted", "rail-body":
		return smallFont
	default:
		return bodyFont
	}
}

func buildInstallerControls() {
	controls = nil
	controlRoles = map[uintptr]string{}
	textControl("SLOTH MODE", "rail-small", 28, 29, 180, 24)
	textControl("A little less\nto carry.", "rail-title", 28, 146, 172, 96)
	textControl("Keep your thoughts close.\nFind one small next step.", "rail-body", 28, 258, 174, 58)
	textControl("ON THIS COMPUTER", "rail-small", 28, 386, 178, 22)
	textControl(installerRecommendation(), "rail-body", 28, 417, 174, 68)
	graphicsLabel = textControl("Checking graphics locally...", "rail-body", 28, 493, 174, 89)
	textControl("UNSIGNED WINDOWS PREVIEW  /  "+installerVersion, "eyebrow", 252, 31, 532, 24)
	headingLabel = textControl("Settle in.", "title", 252, 67, 532, 49)
	textControl("Your private task companion, on this PC.", "body", 252, 127, 532, 28)
	textControl("01", "eyebrow", 252, 187, 34, 25)
	textControl("Install for your Windows account", "heading", 297, 183, 481, 29)
	textControl("No administrator access or cloud account needed.", "muted", 297, 214, 481, 27)
	textControl("02", "eyebrow", 252, 256, 34, 25)
	textControl("Pair a browser you trust", "heading", 297, 252, 481, 29)
	textControl("Copy its key from Setup & Pairing to the website.", "muted", 297, 283, 481, 27)
	textControl("03", "eyebrow", 252, 325, 34, 25)
	textControl("Add Gemma when you are ready", "heading", 297, 321, 481, 29)
	textControl("Several GB to download, only with your approval. No Ollama.", "muted", 297, 352, 481, 40)
	startupButton = ctrl("BUTTON", startChoice(), 252, 403, 532, 30, wsTabstop|0x2, idStartup)
	if launchOnSignIn {
		sendMessage.Call(startupButton, 0x00F1, 1, 0)
	}
	stepLabel = textControl("READY TO INSTALL", "eyebrow", 268, 453, 500, 20)
	statusLabel = textControl("Existing notes, models and pairing keys stay in place.", "status", 268, 478, 500, 38)
	installButton = ctrl("BUTTON", "Install & open", 252, 536, 326, 44, wsTabstop|bsOwnerDraw, idInstall)
	ctrl("BUTTON", "Close", 590, 536, 194, 44, wsTabstop|bsOwnerDraw, idClose)
	ctrl("BUTTON", "Open app folder", 252, 588, 248, 26, wsTabstop|bsOwnerDraw, idOpenLocation)
	updateButton = ctrl("BUTTON", "Check updates", 536, 588, 248, 26, wsTabstop|bsOwnerDraw, idCheckUpdates)
}

func refreshInstallerStage() {
	labels := []string{"READY TO INSTALL", "INSTALLING COMPANION", "APP INSTALLED", "COMPANION READY"}
	i := installStage
	if i < 0 || i >= len(labels) {
		i = 0
	}
	setWindowText.Call(stepLabel, utf16(labels[i]))
	if installStage >= 2 {
		setWindowText.Call(headingLabel, utf16("Make yourself at home."))
	}
	user32.NewProc("InvalidateRect").Call(hwnd, 0, 1)
}

func resizeForDPI(dpi uint32, work rectWin) {
	uiScale = fittedScale(dpi, work)
	oldFonts := []uintptr{titleFont, headingFont, bodyFont, smallFont}
	makeFonts()
	for _, c := range controls {
		user32.NewProc("MoveWindow").Call(c.handle, uintptr(px(c.x)), uintptr(px(c.y)), uintptr(px(c.width)), uintptr(px(c.height)), 1)
		sendMessage.Call(c.handle, wmSetFont, fontForRole(c.role), 1)
	}
	for _, f := range oldFonts {
		if f != 0 {
			deleteObj.Call(f)
		}
	}
	r := windowBounds(work, dpi)
	user32.NewProc("SetWindowPos").Call(hwnd, 0, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right-r.Left), uintptr(r.Bottom-r.Top), 0x14)
	user32.NewProc("InvalidateRect").Call(hwnd, 0, 1)
}

type monitorInfo struct {
	Size          uint32
	Monitor, Work rectWin
	Flags         uint32
}

func monitorWorkArea(window uintptr) rectWin {
	m, _, _ := user32.NewProc("MonitorFromWindow").Call(window, 2)
	info := monitorInfo{Size: uint32(unsafe.Sizeof(monitorInfo{}))}
	if ok, _, _ := user32.NewProc("GetMonitorInfoW").Call(m, uintptr(unsafe.Pointer(&info))); ok != 0 {
		return info.Work
	}
	var work rectWin
	user32.NewProc("SystemParametersInfoW").Call(0x30, 0, uintptr(unsafe.Pointer(&work)), 0)
	return work
}
