package main

import (
	"runtime"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestWindowsRuntimeArtifactNames(t *testing.T) {
	for _, tc := range []struct{ name, backend, arch string; ok bool }{
		{"llama-b11798-bin-win-cpu-x64.zip", "cpu", "amd64", true},
		{"llama-b11798-bin-win-vulkan-x64.zip", "vulkan", "amd64", true},
		{"llama-b100-bin-win-x64.zip", "cpu", "amd64", true},
		{"llama-b100-bin-win-x64.zip", "vulkan", "amd64", false},
		{"llama-b11798-bin-win-cpu-arm64.zip", "cpu", "amd64", false},
		{"llama-b11798-bin-win-vulkan-arm64.zip", "vulkan", "arm64", true},
		{"llama-b11798-bin-win-cpu-x64.zip", "cpu", "386", false},
		{"llama-b11798-bin-win-cuda-13.4-x64.zip", "cpu", "amd64", false},
		{"llama-b11798-bin-win-vulkan-x64.zip", "cpu", "amd64", false},
		{"llama-b11798-bin-win-cpu-x64.zip", "vulkan", "amd64", false},
		{"cudart-llama-b11798-bin-win-cpu-x64.zip", "cpu", "amd64", false},
		{"evil-bin-win-cpu-x64.zip", "cpu", "amd64", false},
		{"llama-b100-bin-win-cpu-x64.zip.sha256", "cpu", "amd64", false},
		{"llama-b100-bin-win-cpu-x64-debug.zip", "cpu", "amd64", false},
		{"llama-b100-bin-win-cpu-x64.zip", "../cpu", "amd64", false},
	} {
		t.Run(tc.name+"/"+tc.backend+"/"+tc.arch, func(t *testing.T) {
			_, err := pickAssetForPlatform([]asset{{Name: tc.name}}, tc.backend, "windows", tc.arch)
			if (err == nil) != tc.ok { t.Fatalf("accepted=%v, want %v: %v", err == nil, tc.ok, err) }
		})
	}
}

func TestRuntimeReleasePagination(t *testing.T) {
	cpu, vulkan := assetNamesForPlatform()
	for _, backend := range []string{"cpu", "vulkan"} {
		t.Run(backend, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				if r.URL.Query().Get("per_page") != "8" { t.Error("missing page bound") }
				switch r.URL.Query().Get("page") {
				case "1":
					// Stable metadata and incomplete nightlies must not stop fallback.
					rows := make([]release, 8)
					rows[0] = release{Tag: "v0.7.0", Assets: []asset{{Name: "nightly-tag.txt"}}}
					if backend == "vulkan" { rows[1] = release{Tag: "cpu-only", Assets: []asset{{Name: cpu}}} }
					json.NewEncoder(w).Encode(rows)
				case "2":
					name := cpu
					if backend == "vulkan" { name = vulkan }
					json.NewEncoder(w).Encode([]release{
						{Tag: "draft", Draft: true, Assets: []asset{{Name: name}}},
						{Tag: "usable", Assets: []asset{{Name: name}}},
					})
				default: t.Error("unexpected extra request"); http.Error(w, "extra", 500)
				}
			}))
			defer server.Close()
			rel, err := getReleaseFrom(context.Background(), backend, server.Client(), server.URL)
			if err != nil || rel.Tag != "usable" || requests != 2 { t.Fatalf("got %q after %d requests: %v", rel.Tag, requests, err) }
		})
	}
}

func TestRuntimeReleaseSearchFailsClosed(t *testing.T) {
	for _, mode := range []string{"empty", "full", "status", "json"} {
		t.Run(mode, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				switch mode {
				case "empty": json.NewEncoder(w).Encode([]release{})
				case "full": json.NewEncoder(w).Encode(make([]release, 8))
				case "status": http.Error(w, "rate limited", http.StatusForbidden)
				case "json": w.Write([]byte("{"))
				}
			}))
			defer server.Close()
			if _, err := getReleaseFrom(context.Background(), "cpu", server.Client(), server.URL); err == nil { t.Fatal("unavailable release accepted") }
			want := 1
			if mode == "full" { want = 4 }
			if requests != want { t.Fatalf("requests=%d, want %d", requests, want) }
			if _, err := getReleaseFrom(context.Background(), "../cpu", server.Client(), server.URL); err == nil || requests != want { t.Fatal("invalid backend made request") }
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := getReleaseFrom(ctx, "cpu", server.Client(), server.URL); err == nil { t.Fatal("cancellation ignored") }
		})
	}
}
