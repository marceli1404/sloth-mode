package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallApplicationAndUpgrade(t *testing.T) {
	root := t.TempDir()
	dest := filepath.Join(root, "Programs", "Sloth Mode", "SlothMode-Connector.exe")
	notes := filepath.Join(root, "notes.json")
	if err := os.WriteFile(notes, []byte("private notes"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, value := range []byte{1, 2} {
		payload := bytes.Repeat([]byte{value}, 1_000_000)
		if err := installApplication(dest, payload); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(dest)
		if err != nil || !bytes.Equal(got, payload) {
			t.Fatalf("installed bytes differ: %v", err)
		}
		files, err := os.ReadDir(filepath.Dir(dest))
		if err != nil || len(files) != 1 {
			t.Fatalf("staged files left behind: %v, %v", files, err)
		}
	}
	if got, _ := os.ReadFile(notes); string(got) != "private notes" {
		t.Fatal("notes were changed")
	}
}

func TestIncompletePayloadPreservesExistingApplication(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "app.exe")
	if err := os.WriteFile(dest, []byte("old app"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := installApplication(dest, []byte("truncated")); err == nil {
		t.Fatal("incomplete payload accepted")
	}
	if got, _ := os.ReadFile(dest); string(got) != "old app" {
		t.Fatal("old app changed")
	}
}

func TestReplacementRollsBackOnFailure(t *testing.T) {
	for _, failRestore := range []bool{false, true} {
		t.Run(map[bool]string{false: "restore succeeds", true: "restore fails"}[failRestore], func(t *testing.T) {
			root := t.TempDir()
			dest := filepath.Join(root, "app.exe")
			staged := filepath.Join(root, "new.exe")
			if err := os.WriteFile(dest, []byte("old app"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(staged, []byte("new app"), 0600); err != nil {
				t.Fatal(err)
			}
			err := replaceApplication(staged, dest, func(from, to string) error {
				if from == staged || (failRestore && from == dest+".previous") {
					return errors.New("simulated lock")
				}
				return os.Rename(from, to)
			})
			if err == nil {
				t.Fatal("replacement unexpectedly succeeded")
			}
			oldPath := dest
			if failRestore {
				oldPath += ".previous"
				if !strings.Contains(err.Error(), "could not restore") || !strings.Contains(err.Error(), oldPath) {
					t.Fatalf("backup recovery location omitted: %v", err)
				}
			} else if !strings.Contains(err.Error(), "restored") {
				t.Fatalf("rollback not reported: %v", err)
			}
			if got, _ := os.ReadFile(oldPath); string(got) != "old app" {
				t.Fatal("previous app was lost")
			}
		})
	}
}

func TestApplicationDirectoryIsNotMoved(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "app.exe")
	if err := os.Mkdir(dest, 0700); err != nil {
		t.Fatal(err)
	}
	if err := installApplication(dest, bytes.Repeat([]byte{1}, 1_000_000)); err == nil {
		t.Fatal("directory accepted as app")
	}
	if info, err := os.Stat(dest); err != nil || !info.IsDir() {
		t.Fatal("directory moved")
	}
}
