//go:build !windows && !linux

package main

import "runtime"

func probeHardware() HardwareProfile { return HardwareProfile{RAMGiB: 16, CPUCores: runtime.NumCPU()} }
