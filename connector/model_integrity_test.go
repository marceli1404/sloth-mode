package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func integrityFixture(t *testing.T, contents []byte) (string, ModelSpec) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "model.gguf")
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(contents)
	return path, ModelSpec{Name: "model.gguf", Hash: hex.EncodeToString(digest[:]), Size: int64(len(contents))}
}

func TestVerifyModelFileAcceptsPinnedContent(t *testing.T) {
	path, spec := integrityFixture(t, []byte("small verified GGUF fixture"))
	if err := verifyModelFileWithLimit(context.Background(), path, spec, 128, func(string) {}); err != nil {
		t.Fatalf("valid model rejected: %v", err)
	}
}

func TestVerifyModelFileRejectsSameSizeCorruption(t *testing.T) {
	path, spec := integrityFixture(t, []byte("small verified GGUF fixture"))
	corrupt := []byte("small altered GGUF fixtures")
	if len(corrupt) != int(spec.Size) {
		t.Fatalf("test fixture must preserve file size: old=%d new=%d", spec.Size, len(corrupt))
	}
	if err := os.WriteFile(path, corrupt, 0600); err != nil {
		t.Fatal(err)
	}
	err := verifyModelFileWithLimit(context.Background(), path, spec, 128, nil)
	if !errors.Is(err, errModelDigest) || !strings.Contains(err.Error(), "old file has been kept") {
		t.Fatalf("same-size changed model was not rejected clearly: %v", err)
	}
}

func TestVerifyModelFileRejectsDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "model.gguf")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	err := verifyModelFileWithLimit(context.Background(), path, ModelSpec{Hash: strings.Repeat("0", 64), Size: 1}, 128, nil)
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("directory accepted as model: %v", err)
	}
}

func TestVerifyModelFileHonorsCancellation(t *testing.T) {
	path, spec := integrityFixture(t, []byte("small verified GGUF fixture"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := verifyModelFileWithLimit(ctx, path, spec, 128, func(string) { cancel() })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled checksum continued: %v", err)
	}
}

func TestVerifyModelFileEnforcesSizeBoundWithSmallFixture(t *testing.T) {
	path, spec := integrityFixture(t, []byte("small verified GGUF fixture"))
	err := verifyModelFileWithLimit(context.Background(), path, spec, int64(len("small")), nil)
	if !errors.Is(err, errModelSize) {
		t.Fatalf("oversized fixture accepted: %v", err)
	}
}

func TestFailedModelPublishRestoresPreviousFile(t *testing.T) {
	root := t.TempDir()
	model := filepath.Join(root, "model.gguf")
	staged := filepath.Join(root, "new.gguf")
	if err := os.WriteFile(model, []byte("old model"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new model"), 0600); err != nil {
		t.Fatal(err)
	}
	keepBackup, err := publishVerifiedModel(model, staged, root, func(from, to string) error {
		if from == staged {
			return errors.New("simulated publish failure")
		}
		return os.Rename(from, to)
	})
	if err == nil || keepBackup {
		t.Fatalf("expected a recoverable publish failure, err=%v keepBackup=%v", err, keepBackup)
	}
	got, readErr := os.ReadFile(model)
	if readErr != nil || string(got) != "old model" {
		t.Fatalf("previous model not restored: %q, %v", got, readErr)
	}
}

func TestFailedModelPublishKeepsRecoveryCopyIfRestoreFails(t *testing.T) {
	root := t.TempDir()
	model := filepath.Join(root, "model.gguf")
	staged := filepath.Join(root, "new.gguf")
	if err := os.WriteFile(model, []byte("old model"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new model"), 0600); err != nil {
		t.Fatal(err)
	}
	keepBackup, err := publishVerifiedModel(model, staged, root, func(from, to string) error {
		if from == staged || from == filepath.Join(root, "previous-model.gguf") {
			return errors.New("simulated rename failure")
		}
		return os.Rename(from, to)
	})
	if err == nil || !keepBackup {
		t.Fatalf("expected a retained recovery copy, err=%v keepBackup=%v", err, keepBackup)
	}
	got, readErr := os.ReadFile(filepath.Join(root, "previous-model.gguf"))
	if readErr != nil || string(got) != "old model" {
		t.Fatalf("previous model recovery copy not preserved: %q, %v", got, readErr)
	}
}
