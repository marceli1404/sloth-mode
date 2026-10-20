package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type HardwareProfile struct {
	RAMGiB      float64 `json:"ram_gib"`
	CPUCores    int     `json:"cpu_cores"`
	GPUName     string  `json:"gpu_name"`
	GPUVRAMGiB  float64 `json:"gpu_vram_gib"`
	GPUDetected bool    `json:"gpu_detected"`
	Recommended string  `json:"recommended"`
	Backend     string  `json:"backend"`
	Reason      string  `json:"reason"`
}

func recommendQuant(p HardwareProfile) (string, string) {
	// Conservative system RAM gate: model weights plus context, runtime and OS.
	if p.RAMGiB > 0 && p.RAMGiB < 7.5 {
		return "Q3_K_L", "Limited system memory: smaller GGUF; quality may be lower. Close other apps."
	}
	if p.RAMGiB < 15 {
		return "Q4_K_M", "Balanced quantization for modest RAM, with room for model context."
	}
	if p.RAMGiB < 27 {
		return "Q5_K_M", "More RAM available: higher-quality weights without an excessive footprint."
	}
	return "Q6_K", "Ample RAM: higher-quality quantization, while leaving room for system use."
}

var hardwareOnce sync.Once
var hardwareResult HardwareProfile

func profileHardware() HardwareProfile {
	hardwareOnce.Do(func() {
		p := probeHardware()
		if p.CPUCores < 1 {
			p.CPUCores = runtime.NumCPU()
		}
		p.Recommended, p.Reason = recommendQuant(p)
		p.Backend = "cpu" // CPU always works; Vulkan is opt-in because drivers vary.
		if p.GPUDetected && p.GPUVRAMGiB >= 5 {
			p.Reason += " Compatible Vulkan drivers may accelerate inference (optional)."
		}
		hardwareResult = p
	})
	return hardwareResult
}
func preferredQuant(base string, p HardwareProfile) string {
	if data, err := os.ReadFile(filepath.Join(base, "model-preference.json")); err == nil {
		var v struct {
			Quant string `json:"quant"`
		}
		if json.Unmarshal(data, &v) == nil && modelSpec(v.Quant).Name != "" {
			return v.Quant
		}
	}
	// Existing downloaded model wins on upgrade (avoid surprise multi-GB download).
	for _, q := range []string{"Q4_K_M", "Q3_K_L", "Q5_K_M", "Q6_K"} {
		if _, err := os.Stat(modelPathFor(base, q)); err == nil {
			return q
		}
	}
	return p.Recommended
}
func setPreferredQuant(base, quant string) error {
	if modelSpec(quant).Name == "" {
		return errInvalidQuant
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return err
	}
	data, _ := json.Marshal(struct {
		Quant string `json:"quant"`
	}{quant})
	tmp := filepath.Join(base, "model-preference.json.tmp")
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(base, "model-preference.json"))
}
func gpuLabel(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 140 {
		return s[:140]
	}
	return s
}
