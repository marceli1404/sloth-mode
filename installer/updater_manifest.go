package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const installerVersion = "0.14.0"
const updateManifestURL = "https://sloth.marceli.dev/updates/stable.json"
const pinnedUpdateKey = "SUBOk4FHKAO+PmMStucXWhuKXpA8/xRkbkAeMVAGyGQ="
const maxInstallerBytes int64 = 90 << 20

// The signature covers all security-critical metadata, using a fixed unambiguous
// wire format; any changes require a new protocol version and pinned key.
type updateManifest struct {
	Format    int    `json:"format"`
	Version   string `json:"version"`
	URL       string `json:"url"`
	SHA256    string `json:"sha256"`
	Size      int64  `json:"size"`
	Signature string `json:"signature"`
}

func manifestPayload(m updateManifest) []byte {
	return []byte(fmt.Sprintf("sloth-mode-update-v1\n%d\n%s\n%s\n%s\n%d\n", m.Format, m.Version, m.URL, m.SHA256, m.Size))
}
func versionParts(v string) ([]int, error) {
	s := strings.Split(v, ".")
	if len(s) != 3 {
		return nil, errors.New("invalid version")
	}
	out := make([]int, 3)
	for i, x := range s {
		if x == "" || len(x) > 8 || strings.Trim(x, "0123456789") != "" {
			return nil, errors.New("invalid version")
		}
		n, e := strconv.Atoi(x)
		if e != nil {
			return nil, e
		}
		out[i] = n
	}
	return out, nil
}
func newerVersion(current, latest string) (bool, error) {
	a, e := versionParts(current)
	if e != nil {
		return false, e
	}
	b, e := versionParts(latest)
	if e != nil {
		return false, e
	}
	for i := range a {
		if b[i] > a[i] {
			return true, nil
		}
		if b[i] < a[i] {
			return false, nil
		}
	}
	return false, nil
}
func verifyManifest(m updateManifest) error {
	if m.Format != 1 {
		return errors.New("unrecognised update metadata")
	}
	if _, e := versionParts(m.Version); e != nil {
		return e
	}
	pk, e := base64.StdEncoding.DecodeString(pinnedUpdateKey)
	if e != nil || len(pk) != ed25519.PublicKeySize {
		return errors.New("invalid updater public key")
	}
	sig, e := base64.StdEncoding.DecodeString(m.Signature)
	if e != nil || !ed25519.Verify(pk, manifestPayload(m), sig) {
		return errors.New("update signature verification failed")
	}
	// Release metadata for the current version may omit a download URL. A newer
	// version is required to provide HTTPS, digest and bounded payload length.
	newer, e := newerVersion(installerVersion, m.Version)
	if e != nil {
		return e
	}
	if !newer {
		return nil
	}
	u, e := url.Parse(m.URL)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return errors.New("update URL must be HTTPS")
	}
	digest, e := hex.DecodeString(m.SHA256)
	if e != nil || len(digest) != sha256.Size {
		return errors.New("invalid update digest")
	}
	if m.Size < 1024 || m.Size > maxInstallerBytes {
		return errors.New("update installer size out of bounds")
	}
	return nil
}
