//go:build linux && (amd64 || arm64)

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxRuntimePublishNeverReplacesEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "runtime")
	if err := os.Mkdir(dest, 0700); err != nil {
		t.Fatal(err)
	}
	staged := filepath.Join(root, "staged")
	if err := os.Mkdir(staged, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staged, "llama-server"), []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := publishLinuxRuntimeDir(staged, dest); err == nil {
		t.Fatal("existing empty directory was replaced")
	}
	if entries, err := os.ReadDir(dest); err != nil || len(entries) != 0 {
		t.Fatalf("destination changed: %d entries: %v", len(entries), err)
	}
	if _, err := os.Stat(filepath.Join(staged, "llama-server")); err != nil {
		t.Fatalf("staged source lost: %v", err)
	}
}

func TestLinuxRuntimePublishRefusesSymlinkAndPublishesOnce(t *testing.T) {
	root := t.TempDir()
	staged := filepath.Join(root, "staged")
	dest := filepath.Join(root, "runtime")
	if err := os.Mkdir(staged, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("elsewhere", dest); err != nil {
		t.Fatal(err)
	}
	if err := publishLinuxRuntimeDir(staged, dest); err == nil {
		t.Fatal("symlink target replaced")
	}
	if err := os.Remove(dest); err != nil {
		t.Fatal(err)
	}
	if err := publishLinuxRuntimeDir(staged, dest); err != nil {
		t.Fatal(err)
	}
	if err := publishLinuxRuntimeDir(dest, dest); err == nil {
		t.Fatal("identical source/destination accepted")
	}
}
