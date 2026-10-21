package main

import (
	"runtime"
	"strings"
	"testing"
)

func assetNamesForPlatform() (string, string) {
	if runtime.GOOS == "linux" {
		arch := "x64"
		if runtime.GOARCH == "arm64" {
			arch = "arm64"
		}
		return "llama-b100-bin-ubuntu-" + arch + ".tar.gz",
			"llama-b099-bin-ubuntu-vulkan-" + arch + ".tar.gz"
	}
	return "llama-b100-bin-win-cpu-x64.zip",
		"llama-b099-bin-win-vulkan-x64.zip"
}

func TestRuntimeReleaseMatchesRequestedBackend(t *testing.T) {
	cpuName, vulkanName := assetNamesForPlatform()
	releases := []release{
		{Tag: "b100", Assets: []asset{{Name: cpuName}}},
		{Tag: "b099", Assets: []asset{{Name: vulkanName}}},
	}
	cpu, err := selectRuntimeRelease(releases, "cpu")
	if err != nil || cpu.Tag != "b100" {
		t.Fatalf("CPU should choose latest CPU release, got %q: %v", cpu.Tag, err)
	}
	vulkan, err := selectRuntimeRelease(releases, "vulkan")
	if err != nil || vulkan.Tag != "b099" {
		t.Fatalf("Vulkan should skip CPU-only release, got %q: %v", vulkan.Tag, err)
	}
}

func TestRuntimeReleaseRefusesUnavailableBackend(t *testing.T) {
	cpuName, _ := assetNamesForPlatform()
	_, err := selectRuntimeRelease([]release{{Tag: "b100", Assets: []asset{{Name: cpuName}}}}, "vulkan")
	if err == nil || !strings.Contains(err.Error(), "vulkan") {
		t.Fatalf("missing Vulkan runtime must not silently fall back to CPU: %v", err)
	}
	if _, err := selectRuntimeRelease(nil, "../vulkan"); err == nil {
		t.Fatal("invalid backend accepted")
	}
}
