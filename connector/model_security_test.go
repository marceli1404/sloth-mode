package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func makeRuntimeZip(t *testing.T, names []string) string {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "runtime.zip")
	out, err := os.Create(file)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(out)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("dummy executable")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestRuntimeZipSafeExtraction(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "runtime")
	if err := extractRuntime(makeRuntimeZip(t, []string{"bin/llama-server.exe", "bin/ggml.dll"}), dest); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"llama-server.exe", "ggml.dll"} {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRuntimeZipRejectsUnsafeEntries(t *testing.T) {
	cases := [][]string{
		{"../outside", "bin/llama-server.exe"},
		{"..\\outside", "bin/llama-server.exe"},
		{"/outside", "bin/llama-server.exe"},
		{"bin/llama-server.exe", "other/llama-server.exe"},
		{"bin/llama-server.exe", "other/LLAMA-SERVER.exe"},
	}
	for _, files := range cases {
		t.Run(files[0], func(t *testing.T) {
			dir := t.TempDir()
			dest := filepath.Join(dir, "runtime")
			if err := extractRuntime(makeRuntimeZip(t, files), dest); err == nil {
				t.Fatal("unsafe ZIP accepted")
			}
			if _, err := os.Stat(dest); !os.IsNotExist(err) {
				t.Fatalf("failed extraction produced a destination: %v", err)
			}
		})
	}
}

func TestRuntimeZipNeverOverwritesExistingDestination(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "runtime")
	if err := os.Mkdir(dest, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(dest, "do-not-remove")
	if err := os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := extractRuntime(makeRuntimeZip(t, []string{"llama-server.exe"}), dest); err == nil {
		t.Fatal("existing installation overwritten")
	}
	if value, err := os.ReadFile(sentinel); err != nil || string(value) != "keep" {
		t.Fatalf("old data lost: %v", err)
	}
}

func TestRuntimeArchiveExpandedSizeBudget(t *testing.T) {
	const budget int64 = 750 << 20
	for _, tc := range []struct{ used int64; advertised uint64; allowed bool }{
		{0, 150 << 20, true},
		{700 << 20, 50 << 20, true},
		{700 << 20, 51 << 20, false},
		{budget, 1, false},
		{-1, 0, false},
		{budget + 1, 0, false},
		{0, ^uint64(0), false},
	} {
		if got := runtimeExpansionWithinBudget(tc.used, tc.advertised, budget); got != tc.allowed {
			t.Fatalf("total=%d, advertised=%d: expected %v, got %v", tc.used, tc.advertised, tc.allowed, got)
		}
	}
}
