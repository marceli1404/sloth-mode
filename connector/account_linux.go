//go:build linux

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Linux credentials live in Secret Service (GNOME Keyring, KWallet with a
// Secret Service provider, etc.). The JSON file contains only an opaque ID.
// No plaintext-on-disk fallback: guest mode remains available without a keyring.
func sealAccountToken(value []byte) ([]byte, error) {
	if _, err := exec.LookPath("secret-tool"); err != nil {
		return nil, errors.New("install libsecret-tools and unlock a desktop keyring to link accounts or API keys (guest mode needs neither)")
	}
	if len(value) == 0 || len(value) > 4096 {
		return nil, errors.New("invalid secret size")
	}
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	id := hex.EncodeToString(random)
	cmd := exec.Command("secret-tool", "store", "--label=Sloth Mode local secret", "application", "sloth-mode", "reference", id)
	cmd.Stdin = bytes.NewReader(value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("could not unlock Secret Service keyring: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return []byte("libsecret:" + id), nil
}
func unsealAccountToken(value []byte) ([]byte, error) {
	id := strings.TrimPrefix(string(value), "libsecret:")
	if len(id) != 48 || string(value) != "libsecret:"+id {
		return nil, errors.New("invalid Secret Service reference")
	}
	for _, ch := range id {
		if !strings.ContainsRune("0123456789abcdef", ch) {
			return nil, errors.New("invalid Secret Service reference")
		}
	}
	cmd := exec.Command("secret-tool", "lookup", "application", "sloth-mode", "reference", id)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve secret from keyring: %w", err)
	}
	out = bytes.TrimSuffix(out, []byte{'\n'})
	if len(out) == 0 {
		return nil, errors.New("keyring secret unavailable")
	}
	return out, nil
}
