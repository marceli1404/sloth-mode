//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	wmPaint          = 0x000F
	wmEraseBkgnd     = 0x0014
	wmCtlColorStatic = 0x0138
	wmDrawItem       = 0x002B
	transparentBk    = 1
	bsOwnerDraw      = 0x0000000B
)

type rectWin struct{ Left, Top, Right, Bottom int32 }
type paintStruct struct {
	Hdc                uintptr
	Erase              int32
	RcPaint            rectWin
	Restore, IncUpdate int32
	RGBReserved        [32]byte
}
type drawItemStruct struct {
	CtlType, CtlID, ItemID, ItemAction, ItemState uint32
	HwndItem, Hdc                                 uintptr
	RcItem                                        rectWin
	ItemData                                      uintptr
}

var (
	brushDLL     = syscall.NewLazyDLL("gdi32.dll")
	paintBegin   = user32.NewProc("BeginPaint")
	paintEnd     = user32.NewProc("EndPaint")
	createBrush  = brushDLL.NewProc("CreateSolidBrush")
	fillRect     = user32.NewProc("FillRect")
	roundRect    = brushDLL.NewProc("RoundRect")
	selectObj    = brushDLL.NewProc("SelectObject")
	deleteObj    = brushDLL.NewProc("DeleteObject")
	createPen    = brushDLL.NewProc("CreatePen")
	ellipse      = brushDLL.NewProc("Ellipse")
	setBkMode    = brushDLL.NewProc("SetBkMode")
	setTextColor = brushDLL.NewProc("SetTextColor")
	drawText     = user32.NewProc("DrawTextW")
	createFont   = brushDLL.NewProc("CreateFontW")
)

func rgb(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }

var backgroundBrush, panelBrush uintptr
var titleFont, bodyFont, smallFont uintptr

func makeTheme() {
	backgroundBrush, _, _ = createBrush.Call(rgb(14, 35, 31))
	panelBrush, _, _ = createBrush.Call(rgb(14, 35, 31))
	titleFont = createSegoe(29, true)
	bodyFont = createSegoe(16, false)
	smallFont = createSegoe(14, false)
}
func createSegoe(size int32, bold bool) uintptr {
	weight := uintptr(400)
	if bold {
		weight = 700
	}
	f, _, _ := createFont.Call(uintptr(uint32(-size)), 0, 0, 0, weight, 0, 0, 0, 0, 0, 0, 0, 0, utf16("Segoe UI"))
	return f
}
func paintBackground(h uintptr) {
	var ps paintStruct
	hdc, _, _ := paintBegin.Call(h, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer paintEnd.Call(h, uintptr(unsafe.Pointer(&ps)))
	base, _, _ := createBrush.Call(rgb(14, 35, 31))
	fillRect.Call(hdc, uintptr(unsafe.Pointer(&rectWin{0, 0, 760, 680})), base)
	deleteObj.Call(base)
	// Soft block composition. Real native GDI, no browser or web view.
	drawRound(hdc, 20, 15, 738, 182, rgb(31, 75, 60), 0)
	drawRound(hdc, 20, 198, 738, 467, rgb(23, 53, 46), 0)
	drawRound(hdc, 20, 470, 738, 688, rgb(21, 48, 41), 0)
	// Sloth mark: warm cream silhouette on the top-right, with gentle leaf detail.
	outlinePen := mkPen(rgb(31, 75, 60))
	oldPen, _, _ := selectObj.Call(hdc, outlinePen)
	face, _, _ := createBrush.Call(rgb(218, 232, 188))
	oldBrush, _, _ := selectObj.Call(hdc, face)
	ellipse.Call(hdc, uintptr(int32(645)), uintptr(int32(46)), uintptr(int32(709)), uintptr(int32(118)))
	eye, _, _ := createBrush.Call(rgb(39, 76, 57))
	selectObj.Call(hdc, eye)
	ellipse.Call(hdc, 662, 78, 670, 87)
	ellipse.Call(hdc, 686, 78, 694, 87)
	nose, _, _ := createBrush.Call(rgb(130, 91, 69))
	selectObj.Call(hdc, nose)
	ellipse.Call(hdc, 675, 91, 681, 97)
	selectObj.Call(hdc, oldBrush)
	selectObj.Call(hdc, oldPen)
	deleteObj.Call(face)
	deleteObj.Call(eye)
	deleteObj.Call(nose)
	deleteObj.Call(outlinePen)
	// Three discrete progress milestones; Stage 1 -> 2 -> 3.
	for i := 0; i < 3; i++ {
		col := rgb(60, 91, 77)
		if i <= installStage {
			col = rgb(159, 220, 161)
		}
		drawRound(hdc, int32(40+i*78), 668, int32(98+i*78), 675, col, 0)
	}
}
func mkPen(col uintptr) uintptr { p, _, _ := createPen.Call(0, 1, col); return p }
func drawRound(hdc uintptr, x, y, x2, y2 int32, fill uintptr, border uintptr) {
	brush, _, _ := createBrush.Call(fill)
	pen := mkPen(fill)
	oldBrush, _, _ := selectObj.Call(hdc, brush)
	oldPen, _, _ := selectObj.Call(hdc, pen)
	roundRect.Call(hdc, uintptr(x), uintptr(y), uintptr(x2), uintptr(y2), 18, 18)
	selectObj.Call(hdc, oldBrush)
	selectObj.Call(hdc, oldPen)
	deleteObj.Call(brush)
	deleteObj.Call(pen)
}
func setStaticColors(hdc uintptr) uintptr {
	setTextColor.Call(hdc, rgb(211, 233, 213))
	setBkMode.Call(hdc, transparentBk)
	return backgroundBrush
}
func paintButton(l uintptr) uintptr {
	di := (*drawItemStruct)(unsafe.Pointer(l))
	if di == nil {
		return 0
	}
	focused := di.ItemState&0x10 != 0
	pressed := di.ItemState&0x1 != 0
	color := rgb(43, 88, 69)
	if di.CtlID == idInstall {
		color = rgb(172, 224, 151)
	}
	if di.CtlID == idClose {
		color = rgb(57, 69, 64)
	}
	if pressed {
		color = rgb(103, 144, 105)
	}
	drawRound(di.Hdc, di.RcItem.Left, di.RcItem.Top, di.RcItem.Right, di.RcItem.Bottom, color, 0)
	setBkMode.Call(di.Hdc, transparentBk)
	if di.CtlID == idInstall {
		setTextColor.Call(di.Hdc, rgb(24, 53, 43))
	} else {
		setTextColor.Call(di.Hdc, rgb(235, 245, 224))
	}
	prev, _, _ := selectObj.Call(di.Hdc, bodyFont)
	var title [160]uint16
	sendMessage.Call(di.HwndItem, 0x000D, 160, uintptr(unsafe.Pointer(&title[0]))) // WM_GETTEXT
	drawText.Call(di.Hdc, uintptr(unsafe.Pointer(&title[0])), ^uintptr(0), uintptr(unsafe.Pointer(&di.RcItem)), 0x00000001|0x00000004|0x00000020)
	selectObj.Call(di.Hdc, prev)
	if focused {
		setTextColor.Call(di.Hdc, rgb(240, 250, 230))
	}
	return 1
}
