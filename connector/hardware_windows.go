//go:build windows

package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// MEMORYSTATUSEX structure; only total physical RAM is used (nothing uploaded).
type memoryStat struct {
	Length, MemoryLoad                                                                                   uint32
	TotalPhys, AvailPhys, TotalPageFile, AvailPageFile, TotalVirtual, AvailVirtual, AvailExtendedVirtual uint64
}

func probeHardware() HardwareProfile {
	p := HardwareProfile{CPUCores: runtime.NumCPU()}
	k := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	m := memoryStat{Length: uint32(unsafe.Sizeof(memoryStat{}))}
	if v, _, _ := k.Call(uintptr(unsafe.Pointer(&m))); v != 0 {
		p.RAMGiB = float64(m.TotalPhys) / (1 << 30)
	}
	// Optional bounded WMI query, not a hardware fingerprint or network action.
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", `Get-CimInstance Win32_VideoController | Select-Object -First 1 Name,AdapterRAM | ConvertTo-Json -Compress`)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if out, err := cmd.Output(); err == nil {
		var g struct {
			Name       string `json:"Name"`
			AdapterRAM uint64 `json:"AdapterRAM"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(string(out))), &g) == nil {
			p.GPUName = gpuLabel(g.Name)
			p.GPUDetected = p.GPUName != ""
			// WMI AdapterRAM often saturates at 4GB on cards with more VRAM.
			p.GPUVRAMGiB = float64(g.AdapterRAM) / (1 << 30)
		}
	}
	return p
}
