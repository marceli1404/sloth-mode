//go:build windows

package main

import (
	"runtime"
	"syscall"
	"unsafe"
)

const (
	setupWmCtlColorStatic = 0x0138
	setupWmCtlColorEdit   = 0x0133
	setupWmDrawItem       = 0x002B
	setupBsOwnerDraw      = 0x0000000B
)

type setupDrawItem struct {
	ctlType, ctlID, itemID, itemAction, itemState uint32
	hwndItem, hdc                                 uintptr
	rcItem                                        setupRect
	itemData                                      uintptr
}

var (
	setupGDI          = syscall.NewLazyDLL("gdi32.dll")
	setupCreateBrush  = setupGDI.NewProc("CreateSolidBrush")
	setupCreatePen    = setupGDI.NewProc("CreatePen")
	setupSelectObject = setupGDI.NewProc("SelectObject")
	setupDeleteObject = setupGDI.NewProc("DeleteObject")
	setupRoundRect    = setupGDI.NewProc("RoundRect")
	setupRectangle    = setupGDI.NewProc("Rectangle")
	setupFillRect     = u32.NewProc("FillRect")
	setupDrawFocus    = u32.NewProc("DrawFocusRect")
	setupSetTextColor = setupGDI.NewProc("SetTextColor")
	setupSetBkMode    = setupGDI.NewProc("SetBkMode")
	setupDrawText     = u32.NewProc("DrawTextW")
	setupMoveMemory   = k32.NewProc("RtlMoveMemory")
	setupButtonFont   = uintptr(0)
	setupHeadingFont  = uintptr(0)
	setupTitleFont    = uintptr(0)
	setupCreamBrush   = uintptr(0)
	setupWhiteBrush   = uintptr(0)
)

func setupRGB(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }

func setupInitTheme(scale float64) {
	setupCreamBrush, _, _ = setupCreateBrush.Call(setupRGB(247, 246, 237))
	setupWhiteBrush, _, _ = setupCreateBrush.Call(setupRGB(255, 255, 252))
	fontProc := setupGDI.NewProc("CreateFontW")
	fontName := wide("Segoe UI")
	fontHeight := int32(15 * scale)
	setupButtonFont, _, _ = fontProc.Call(uintptr(uint32(-fontHeight)), 0, 0, 0, 400, 0, 0, 0, 0, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
	setupHeadingFont, _, _ = fontProc.Call(uintptr(uint32(-int32(16*scale))), 0, 0, 0, 600, 0, 0, 0, 0, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
	setupTitleFont, _, _ = fontProc.Call(uintptr(uint32(-int32(23*scale))), 0, 0, 0, 600, 0, 0, 0, 0, 0, 0, 5, 0, uintptr(unsafe.Pointer(fontName)))
	runtime.KeepAlive(fontName)
}

func setupReleaseTheme() {
	for _, h := range []uintptr{setupCreamBrush, setupWhiteBrush, setupButtonFont, setupHeadingFont, setupTitleFont} {
		if h != 0 {
			setupDeleteObject.Call(h)
		}
	}
	setupCreamBrush, setupWhiteBrush, setupButtonFont, setupHeadingFont, setupTitleFont = 0, 0, 0, 0, 0
}

func setupColorControl(msg, l uintptr) uintptr {
	if msg == setupWmCtlColorEdit {
		setupSetTextColor.Call(l, setupRGB(35, 55, 44))
		setupGDI.NewProc("SetBkColor").Call(l, setupRGB(255, 255, 252))
		return setupWhiteBrush
	}
	setupSetTextColor.Call(l, setupRGB(35, 55, 44))
	setupSetBkMode.Call(l, 1) // Transparent static labels sit directly on the cream surface.
	setupGDI.NewProc("SetBkColor").Call(l, setupRGB(247, 246, 237))
	return setupCreamBrush
}

func setupPaintButton(l uintptr) uintptr {
	if l == 0 {
		return 0
	}
	var di setupDrawItem
	setupMoveMemory.Call(uintptr(unsafe.Pointer(&di)), l, unsafe.Sizeof(di))
	if di.ctlID == idAutoStart {
		brush, _, _ := setupCreateBrush.Call(setupRGB(247, 246, 237))
		setupFillRect.Call(di.hdc, uintptr(unsafe.Pointer(&di.rcItem)), brush)
		setupDeleteObject.Call(brush)
		box := setupRect{di.rcItem.left + 3, di.rcItem.top + (di.rcItem.bottom-di.rcItem.top-17)/2, di.rcItem.left + 20, di.rcItem.top + (di.rcItem.bottom-di.rcItem.top+17)/2}
		boxBrush, _, _ := setupCreateBrush.Call(setupRGB(255, 255, 252))
		boxPen, _, _ := setupCreatePen.Call(0, 1, setupRGB(32, 63, 49))
		oldBrush, _, _ := setupSelectObject.Call(di.hdc, boxBrush)
		oldPen, _, _ := setupSelectObject.Call(di.hdc, boxPen)
		setupRectangle.Call(di.hdc, uintptr(box.left), uintptr(box.top), uintptr(box.right), uintptr(box.bottom))
		setupSelectObject.Call(di.hdc, oldBrush)
		setupSelectObject.Call(di.hdc, oldPen)
		if startWithWindowsEnabled() {
			setupSetTextColor.Call(di.hdc, setupRGB(32, 63, 49))
			var tick [2]uint16
			tick[0] = '✓'
			setupDrawText.Call(di.hdc, uintptr(unsafe.Pointer(&tick[0])), 1, uintptr(unsafe.Pointer(&box)), 0x1|0x4|0x20)
		}
		setupDeleteObject.Call(boxBrush)
		setupDeleteObject.Call(boxPen)
		var label [160]uint16
		pSendMessage.Call(di.hwndItem, 0x000D, uintptr(len(label)), uintptr(unsafe.Pointer(&label[0])))
		setupSetTextColor.Call(di.hdc, setupRGB(35, 55, 44))
		oldFont, _, _ := setupSelectObject.Call(di.hdc, setupButtonFont)
		r := di.rcItem
		r.left += 29
		setupDrawText.Call(di.hdc, uintptr(unsafe.Pointer(&label[0])), ^uintptr(0), uintptr(unsafe.Pointer(&r)), 0x0|0x4|0x20|0x800)
		setupSelectObject.Call(di.hdc, oldFont)
		if di.itemState&0x10 != 0 {
			r = di.rcItem
			r.left += 1
			r.top += 1
			r.right -= 1
			r.bottom -= 1
			setupDrawFocus.Call(di.hdc, uintptr(unsafe.Pointer(&r)))
		}
		return 1
	}
	fill := setupRGB(230, 235, 218)
	ink := setupRGB(35, 55, 44)
	if di.ctlID == idCopyKey || di.ctlID == idOpenWebsite || di.ctlID == idLinkAccount {
		fill, ink = setupRGB(32, 63, 49), setupRGB(247, 246, 237)
	}
	if di.itemState&1 != 0 {
		fill = setupRGB(208, 220, 191)
	}
	brush, _, _ := setupCreateBrush.Call(fill)
	pen, _, _ := setupCreatePen.Call(0, 1, fill)
	oldBrush, _, _ := setupSelectObject.Call(di.hdc, brush)
	oldPen, _, _ := setupSelectObject.Call(di.hdc, pen)
	setupRoundRect.Call(di.hdc, uintptr(di.rcItem.left), uintptr(di.rcItem.top), uintptr(di.rcItem.right), uintptr(di.rcItem.bottom), 10, 10)
	setupSelectObject.Call(di.hdc, oldBrush)
	setupSelectObject.Call(di.hdc, oldPen)
	setupSetBkMode.Call(di.hdc, 1)
	setupSetTextColor.Call(di.hdc, ink)
	oldFont, _, _ := setupSelectObject.Call(di.hdc, setupButtonFont)
	var label [160]uint16
	pSendMessage.Call(di.hwndItem, 0x000D, uintptr(len(label)), uintptr(unsafe.Pointer(&label[0])))
	r := di.rcItem
	setupDrawText.Call(di.hdc, uintptr(unsafe.Pointer(&label[0])), ^uintptr(0), uintptr(unsafe.Pointer(&r)), 0x1|0x4|0x20|0x800)
	setupSelectObject.Call(di.hdc, oldFont)
	if di.itemState&0x10 != 0 {
		r := di.rcItem
		r.left += 4
		r.top += 4
		r.right -= 4
		r.bottom -= 4
		setupDrawFocus.Call(di.hdc, uintptr(unsafe.Pointer(&r)))
	}
	setupDeleteObject.Call(brush)
	setupDeleteObject.Call(pen)
	return 1
}
