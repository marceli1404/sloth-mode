//go:build windows

package main

import "math"

// Pairing stays compact enough to fit common laptop work areas. The outer
// frame is measured by Windows; these dimensions describe the client area.
const (
	setupClientWidth  int32 = 660
	setupClientHeight int32 = 610
)

type setupRect struct{ left, top, right, bottom int32 }

var setupScale = 1.0

func setupFitScale(dpi uint32, workWidth, workHeight, frameWidth, frameHeight int32) float64 {
	scale := float64(dpi) / 96
	if scale <= 0 {
		scale = 1
	}
	if workWidth > 0 && workHeight > 0 {
		availableWidth := float64(workWidth - frameWidth - 16)
		availableHeight := float64(workHeight - frameHeight - 16)
		scale = math.Min(scale, math.Min(availableWidth/float64(setupClientWidth), availableHeight/float64(setupClientHeight)))
	}
	if scale > 1 {
		scale = 1 // Keep the established layout size; Windows may scale its fonts.
	}
	if scale < 0.72 {
		scale = 0.72 // Retain readable type when the available area is unusually small.
	}
	return scale
}

func setupBounds(work setupRect, scale float64, frameWidth, frameHeight int32) setupRect {
	w := int32(math.Round(float64(setupClientWidth)*scale)) + frameWidth
	h := int32(math.Round(float64(setupClientHeight)*scale)) + frameHeight
	x := work.left + (work.right-work.left-w)/2
	y := work.top + (work.bottom-work.top-h)/2
	return setupRect{x, y, x + w, y + h}
}

func setupPx(v int32, scale float64) int32 { return int32(math.Round(float64(v) * scale)) }
