package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	errModelSize   = errors.New("model file size is outside the accepted bounds")
	errModelDigest = errors.New("model checksum does not match")
)

// verifyModelFile streams the complete model through SHA256 while honoring
// cancellation. It is deliberately called during explicit setup and directly
// before engine startup, never from the status polling path.
func verifyModelFile(ctx context.Context, path string, spec ModelSpec, status statusFunc) error {
	return verifyModelFileWithLimit(ctx, path, spec, maxModelBytes, status)
}

func verifyModelFileWithLimit(ctx context.Context, path string, spec ModelSpec, maxBytes int64, status statusFunc) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("could not inspect the Gemma model; use Setup to install it: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("Gemma model path is not a regular file; move it aside and use Setup to install a verified copy")
	}
	if info.Size() < spec.Size*90/100 || info.Size() <= 0 || info.Size() > maxBytes {
		return fmt.Errorf("%w: the existing Gemma model is incomplete or oversized; use Setup to download a verified copy", errModelSize)
	}
	expected, err := hex.DecodeString(spec.Hash)
	if err != nil || len(expected) != sha256.Size {
		return errors.New("the pinned Gemma model checksum is invalid; update Sloth Mode before installing the model")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not read the Gemma model; use Setup to install it: %w", err)
	}
	defer f.Close()
	openedInfo, err := f.Stat()
	if err != nil {
		return fmt.Errorf("could not inspect the Gemma model while reading it: %w", err)
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return errors.New("Gemma model changed while it was being opened; retry Setup")
	}

	h := sha256.New()
	buf := make([]byte, 256<<10)
	var n int64
	lastStatus := time.Now()
	statusSent := false
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		read, readErr := f.Read(buf)
		if read > 0 {
			n += int64(read)
			if n > maxBytes {
				return fmt.Errorf("%w: the Gemma model grew beyond the allowed size while being checked", errModelSize)
			}
			_, _ = h.Write(buf[:read])
			if status != nil && (!statusSent || time.Since(lastStatus) >= time.Second) {
				status(fmt.Sprintf("Checking Gemma model integrity (%.1f/%.1f GB)...", float64(n)/1e9, float64(info.Size())/1e9))
				lastStatus = time.Now()
				statusSent = true
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return fmt.Errorf("could not read the Gemma model checksum: %w", readErr)
		}
	}
	if n != info.Size() {
		return errors.New("Gemma model changed while its checksum was being checked; retry Setup")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !equalBytes(h.Sum(nil), expected) {
		return fmt.Errorf("%w: the existing Gemma model is damaged or changed; use Setup to download a verified replacement. The old file has been kept.", errModelDigest)
	}
	return nil
}

// replaceModelWithVerifiedDownload stages a fully checked download before
// moving the old regular file. If publishing the new file fails, it restores
// the previous file; if restoration also fails, it keeps the recovery copy.
func replaceModelWithVerifiedDownload(ctx context.Context, path string, spec ModelSpec, status statusFunc) error {
	stageDir, err := os.MkdirTemp(filepath.Dir(path), ".sloth-model-replace-*")
	if err != nil {
		return fmt.Errorf("could not stage the verified Gemma replacement: %w", err)
	}
	keepStage := false
	defer func() {
		if !keepStage {
			_ = os.RemoveAll(stageDir)
		}
	}()
	staged := filepath.Join(stageDir, spec.Name)
	if err := download(ctx, spec.URL, staged, maxModelBytes, spec.Hash, status); err != nil {
		return fmt.Errorf("could not download a verified Gemma replacement; the existing model was kept: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	keepStage, err = publishVerifiedModel(path, staged, stageDir, os.Rename)
	if err != nil {
		return err
	}
	return nil
}

func publishVerifiedModel(path, staged, stageDir string, rename func(string, string) error) (keepBackup bool, err error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, fmt.Errorf("could not recheck the existing Gemma model before replacement: %w", err)
	}
	if !info.Mode().IsRegular() {
		return false, errors.New("Gemma model path is no longer a regular file; the verified replacement was not installed")
	}
	backup := filepath.Join(stageDir, "previous-model.gguf")
	if err := rename(path, backup); err != nil {
		return false, fmt.Errorf("could not preserve the existing Gemma model before replacement: %w", err)
	}
	if err := rename(staged, path); err != nil {
		if restoreErr := rename(backup, path); restoreErr != nil {
			return true, fmt.Errorf("verified replacement failed (%v); the previous model is preserved at %s because automatic restoration failed: %w", err, backup, restoreErr)
		}
		return false, fmt.Errorf("verified replacement failed; the previous model was restored: %w", err)
	}
	return false, nil
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}
