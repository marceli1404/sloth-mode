package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// Stage beside the destination so replacement stays on the same filesystem.
// This function writes only the application, never notes, models or pairing keys.
func installApplication(dest string, payload []byte) error {
	if len(payload) < 1_000_000 {
		return fmt.Errorf("embedded application is missing or incomplete")
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(dest), "SlothMode-*.new")
	if err != nil {
		return err
	}
	staged := f.Name()
	defer os.Remove(staged)
	_, writeErr := f.Write(payload)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Chmod(staged, 0700); err != nil {
		return err
	}
	data, err := os.ReadFile(staged)
	if err != nil {
		return err
	}
	if sha256.Sum256(data) != sha256.Sum256(payload) {
		return fmt.Errorf("installed executable did not pass integrity check")
	}
	return replaceApplication(staged, dest, os.Rename)
}

func replaceApplication(staged, dest string, rename func(string, string) error) error {
	backup := dest + ".previous"
	hadOld := false
	if info, err := os.Stat(dest); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("application path is not a regular file: %s", dest)
		}
		if err := os.Remove(backup); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not prepare application backup: %w", err)
		}
		if err := rename(dest, backup); err != nil {
			return fmt.Errorf("close the old connector before upgrading: %w", err)
		}
		hadOld = true
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := rename(staged, dest); err != nil {
		if hadOld {
			if restoreErr := rename(backup, dest); restoreErr != nil {
				return fmt.Errorf("installation failed: %v; could not restore the previous app: %v (backup: %s)", err, restoreErr, backup)
			}
			return fmt.Errorf("installation failed; previous app restored: %w", err)
		}
		return fmt.Errorf("installation failed: %w", err)
	}
	if hadOld {
		_ = os.Remove(backup)
	}
	return nil
}
