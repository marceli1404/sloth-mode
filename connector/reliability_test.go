package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecoveredInterruptedStoreUsesBackup(t *testing.T) {
	dir := t.TempDir()
	s, e := openStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Add("Original saved note"); e != nil {
		t.Fatal(e)
	}
	if e = s.Add("Second saved note"); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "tasks.json"), []byte("{broken"), 0600); e != nil {
		t.Fatal(e)
	}
	recovered, e := openStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	if len(recovered.All()) != 1 || recovered.All()[0].Text != "Original saved note" {
		t.Fatalf("backup recovery unexpected: %+v", recovered.All())
	}
	matches, e := filepath.Glob(filepath.Join(dir, "tasks.json.corrupt-*"))
	if e != nil || len(matches) != 1 {
		t.Fatalf("corrupted snapshot not preserved: %v %v", matches, e)
	}
}
func TestMutationsKeepBackupAndProtectNotes(t *testing.T) {
	dir := t.TempDir()
	s, e := openStore(dir)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Add("alpha"); e != nil {
		t.Fatal(e)
	}
	if e = s.Complete(1); e != nil {
		t.Fatal(e)
	}
	if e = s.Edit(1, "beta"); e != nil {
		t.Fatal(e)
	}
	if e = s.Delete(1); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(dir, "tasks.json.bak"))
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(b), "beta") {
		t.Fatalf("backup should have prior note before deletion: %s", b)
	}
}
func TestBackgroundPanicRecoveredAndJobReleased(t *testing.T) {
	s, e := openStore(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	b := newBridge(t.TempDir(), s, "secret")
	if !b.startJob(func() { panic("simulated inference failure") }) {
		t.Fatal("could not start test job")
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		busy, msg, _ := b.state()
		if !busy {
			if !strings.Contains(msg, "Unexpected background error") {
				t.Fatalf("unexpected status %s", msg)
			}
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	t.Fatal("job remained busy after panic")
}
