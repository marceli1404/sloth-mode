//go:build windows

package main

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type ramStatus struct {
	Length, MemoryLoad                                                                                   uint32
	TotalPhys, AvailPhys, TotalPageFile, AvailPageFile, TotalVirtual, AvailVirtual, AvailExtendedVirtual uint64
}

// Native Windows RAM probe; local only, no device identifiers, no telemetry.
func installerRecommendation() string {
	ram := ramStatus{Length: uint32(unsafe.Sizeof(ramStatus{}))}
	p := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	gb := float64(0)
	if ok, _, _ := p.Call(uintptr(unsafe.Pointer(&ram))); ok != 0 {
		gb = float64(ram.TotalPhys) / (1 << 30)
	}
	quant := "Q4_K_M"
	if gb > 0 && gb < 7.5 {
		quant = "Q3_K_L"
	} else if gb >= 27 {
		quant = "Q6_K"
	} else if gb >= 15 {
		quant = "Q5_K_M"
	}
	if gb > 0 {
		return fmt.Sprintf("%.0f GB RAM\n%d CPU threads\nGemma %s suggested", gb, runtime.NumCPU(), quant)
	}
	return fmt.Sprintf("%d CPU threads\nGemma Q4_K_M suggested", runtime.NumCPU())
}

// GPU details may be delayed; query off the installer UI thread.
func installerGPUName() string {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", `Get-CimInstance Win32_VideoController | Select-Object -First 1 -ExpandProperty Name`)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	b, err := cmd.Output()
	if err != nil {
		return "Graphics: Windows did not report a GPU (CPU mode is available)."
	}
	name := strings.TrimSpace(string(b))
	if len(name) > 70 {
		name = name[:70]
	}
	if name == "" {
		return "Graphics: Not reported by Windows."
	}
	return "Graphics: " + name + "\nVulkan optional; CPU available"
}
