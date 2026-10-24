//go:build windows

package main

import (
	_ "embed"
	"encoding/binary"
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
	setBkMode    = brushDLL.NewProc("SetBkMode")
	setTextColor = brushDLL.NewProc("SetTextColor")
	drawText     = user32.NewProc("DrawTextW")
	drawFocus    = user32.NewProc("DrawFocusRect")
	createFont   = brushDLL.NewProc("CreateFontW")
)

func rgb(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }

//go:embed brand.ico
var brandICO []byte
var brandIcon uintptr

var cream = rgb(247, 246, 237)
var forest = rgb(32, 63, 49)
var ink = rgb(35, 55, 44)
var muted = rgb(90, 103, 88)
var sage = rgb(230, 235, 218)
var backgroundBrush, panelBrush, statusBrush uintptr
var titleFont, headingFont, bodyFont, smallFont uintptr

func makeTheme() {
	brandIcon = loadBrandIcon()
	backgroundBrush, _, _ = createBrush.Call(cream)
	panelBrush, _, _ = createBrush.Call(forest)
	statusBrush, _, _ = createBrush.Call(sage)
	makeFonts()
}
func makeFonts() {
	titleFont = makeFont("Georgia", 33, false)
	headingFont = makeFont("Segoe UI", 18, true)
	bodyFont = makeFont("Segoe UI", 16, false)
	smallFont = makeFont("Segoe UI", 14, false)
}
func makeFont(name string, size int32, bold bool) uintptr {
	weight := uintptr(400)
	if bold {
		weight = 600
	}
	f, _, _ := createFont.Call(uintptr(uint32(-px(size))), 0, 0, 0, weight, 0, 0, 0, 0, 0, 0, 5, 0, utf16(name))
	return f
}
func releaseTheme() {
	if brandIcon != 0 {
		user32.NewProc("DestroyIcon").Call(brandIcon)
		brandIcon = 0
	}
	for _, handle := range []uintptr{backgroundBrush, panelBrush, statusBrush, titleFont, headingFont, bodyFont, smallFont} {
		if handle != 0 {
			deleteObj.Call(handle)
		}
	}
}
func paintBackground(h uintptr) {
	var ps paintStruct
	hdc, _, _ := paintBegin.Call(h, uintptr(unsafe.Pointer(&ps)))
	if hdc == 0 {
		return
	}
	defer paintEnd.Call(h, uintptr(unsafe.Pointer(&ps)))
	fillRect.Call(hdc, uintptr(unsafe.Pointer(&ps.RcPaint)), backgroundBrush)
	r := rectWin{0, 0, px(220), px(clientHeight)}
	fillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), panelBrush)
	drawRound(hdc, 252, 441, 784, 521, sage)
	if brandIcon != 0 {
		user32.NewProc("DrawIconEx").Call(hdc, uintptr(px(28)), uintptr(px(75)), brandIcon, uintptr(px(72)), uintptr(px(72)), 0, 0, 3)
	}
	r = rectWin{px(252), px(396), px(784), px(397)}
	brush, _, _ := createBrush.Call(rgb(215, 220, 205))
	fillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), brush)
	deleteObj.Call(brush)
}
func mkPen(col uintptr) uintptr { p, _, _ := createPen.Call(0, 1, col); return p }
func drawRound(hdc uintptr, x, y, x2, y2 int32, fill uintptr) {
	drawRoundPixels(hdc, rectWin{px(x), px(y), px(x2), px(y2)}, fill)
}
func drawRoundPixels(hdc uintptr, r rectWin, fill uintptr) {
	brush, _, _ := createBrush.Call(fill)
	pen := mkPen(fill)
	oldBrush, _, _ := selectObj.Call(hdc, brush)
	oldPen, _, _ := selectObj.Call(hdc, pen)
	roundRect.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom), uintptr(px(10)), uintptr(px(10)))
	selectObj.Call(hdc, oldBrush)
	selectObj.Call(hdc, oldPen)
	deleteObj.Call(brush)
	deleteObj.Call(pen)
}
func setStaticColors(hdc, control uintptr) uintptr {
	role := controlRoles[control]
	color, bg, brush := ink, cream, backgroundBrush
	switch role {
	case "rail-title":
		color, bg, brush = cream, forest, panelBrush
	case "rail-small", "rail-body":
		color, bg, brush = rgb(208, 220, 191), forest, panelBrush
	case "muted":
		color = muted
	case "eyebrow":
		color = muted
		if control == stepLabel {
			bg, brush = sage, statusBrush
		}
	case "status":
		bg, brush = sage, statusBrush
	}
	setTextColor.Call(hdc, color)
	brushDLL.NewProc("SetBkColor").Call(hdc, bg)
	setBkMode.Call(hdc, 1)
	return brush
}
func paintButton(l uintptr) uintptr {
	if l == 0 {
		return 0
	}
	var di drawItemStruct
	kernel32.NewProc("RtlMoveMemory").Call(uintptr(unsafe.Pointer(&di)), l, unsafe.Sizeof(di))
	focused, pressed, disabled := di.ItemState&0x10 != 0, di.ItemState&1 != 0, di.ItemState&4 != 0
	fillRect.Call(di.Hdc, uintptr(unsafe.Pointer(&di.RcItem)), backgroundBrush)
	color, text := sage, ink
	if di.CtlID == idInstall {
		color, text = forest, cream
	}
	if di.CtlID == idOpenLocation || di.CtlID == idCheckUpdates {
		color, text = cream, muted
	}
	if pressed {
		color = rgb(177, 194, 157)
	}
	if disabled {
		color, text = rgb(221, 226, 211), muted
	}
	drawRoundPixels(di.Hdc, di.RcItem, color)
	setBkMode.Call(di.Hdc, transparentBk)
	setTextColor.Call(di.Hdc, text)
	font := bodyFont
	if di.CtlID == idOpenLocation || di.CtlID == idCheckUpdates {
		font = smallFont
	}
	prev, _, _ := selectObj.Call(di.Hdc, font)
	var title [160]uint16
	sendMessage.Call(di.HwndItem, 0x000D, 160, uintptr(unsafe.Pointer(&title[0])))
	drawText.Call(di.Hdc, uintptr(unsafe.Pointer(&title[0])), ^uintptr(0), uintptr(unsafe.Pointer(&di.RcItem)), 0x1|0x4|0x20|0x800)
	selectObj.Call(di.Hdc, prev)
	if focused {
		r := di.RcItem
		r.Left += px(5)
		r.Top += px(5)
		r.Right -= px(5)
		r.Bottom -= px(5)
		drawFocus.Call(di.Hdc, uintptr(unsafe.Pointer(&r)))
	}
	return 1
}

// The checked-in ICO is the same canonical mark as the companion and website.
func loadBrandIcon() uintptr {
	if len(brandICO) < 6 {
		return 0
	}
	count := int(binary.LittleEndian.Uint16(brandICO[4:6]))
	if count == 0 || len(brandICO) < 6+16*count {
		return 0
	}
	entry := brandICO[6+16*(count-1) : 6+16*count]
	size, offset := int(binary.LittleEndian.Uint32(entry[8:12])), int(binary.LittleEndian.Uint32(entry[12:16]))
	if size <= 0 || offset < 0 || offset > len(brandICO)-size {
		return 0
	}
	icon, _, _ := user32.NewProc("CreateIconFromResourceEx").Call(uintptr(unsafe.Pointer(&brandICO[offset])), uintptr(size), 1, 0x30000, 256, 256, 0)
	return icon
}
