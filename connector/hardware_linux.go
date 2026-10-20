//go:build linux

package main

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func probeHardware() HardwareProfile {
	p := HardwareProfile{CPUCores: runtime.NumCPU()}
	file, e := os.Open("/proc/meminfo")
	if e == nil {
		defer file.Close()
		s := bufio.NewScanner(file)
		for s.Scan() {
			var kb uint64
			if _, err := fmtSscanMem(s.Text(), &kb); err == nil && kb > 0 {
				p.RAMGiB = float64(kb) / 1024 / 1024
				break
			}
		}
	}
	// nvidia-smi reports real framebuffer VRAM; lspci is a name-only fallback.
	if _, e = exec.LookPath("nvidia-smi"); e == nil {
		out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader,nounits").Output()
		if err == nil {
			line := strings.Split(strings.TrimSpace(string(out)), "\n")[0]
			col := strings.Split(line, ",")
			if len(col) == 2 {
				p.GPUName = gpuLabel(col[0])
				v, _ := strconv.ParseFloat(strings.TrimSpace(col[1]), 64)
				p.GPUVRAMGiB = v / 1024
				p.GPUDetected = p.GPUName != ""
			}
		}
	}
	if !p.GPUDetected {
		if _, e = exec.LookPath("lspci"); e == nil {
			out, err := exec.Command("lspci", "-mm").Output()
			if err == nil {
				for _, ln := range strings.Split(string(out), "\n") {
					if strings.Contains(strings.ToLower(ln), "vga compatible controller") || strings.Contains(strings.ToLower(ln), "3d controller") {
						p.GPUName = gpuLabel(ln)
						p.GPUDetected = true
						break
					}
				}
			}
		}
	}
	return p
}
func fmtSscanMem(line string, value *uint64) (bool, error) {
	if !strings.HasPrefix(line, "MemTotal:") {
		return false, os.ErrNotExist
	}
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false, os.ErrInvalid
	}
	n, e := strconv.ParseUint(fields[1], 10, 64)
	if e == nil {
		*value = n
	}
	return true, e
}
