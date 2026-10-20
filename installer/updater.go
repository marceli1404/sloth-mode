//go:build windows

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func updaterClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if len(via) >= 3 || r.URL.Scheme != "https" {
			return errors.New("unsafe update redirect")
		}
		return nil
	}}
}
func fetchUpdate(ctx context.Context) (updateManifest, error) {
	var m updateManifest
	request, e := http.NewRequestWithContext(ctx, http.MethodGet, updateManifestURL, nil)
	if e != nil {
		return m, e
	}
	response, e := updaterClient().Do(request)
	if e != nil {
		return m, e
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return m, fmt.Errorf("update server returned HTTP %d", response.StatusCode)
	}
	if e = json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&m); e != nil {
		return m, e
	}
	return m, verifyManifest(m)
}
func downloadInstaller(ctx context.Context, m updateManifest) (string, error) {
	if e := verifyManifest(m); e != nil {
		return "", e
	}
	ok, e := newerVersion(installerVersion, m.Version)
	if e != nil {
		return "", e
	}
	if !ok {
		return "", errors.New("no newer version")
	}
	r, e := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, nil)
	if e != nil {
		return "", e
	}
	response, e := updaterClient().Do(r)
	if e != nil {
		return "", e
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", fmt.Errorf("installer server returned HTTP %d", response.StatusCode)
	}
	dir, e := os.MkdirTemp("", "sloth-update-*")
	if e != nil {
		return "", e
	}
	cleanup := func(err error) (string, error) { _ = os.RemoveAll(dir); return "", err }
	destination := filepath.Join(dir, "SlothMode-Setup-"+m.Version+".exe")
	f, e := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return cleanup(e)
	}
	h := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(f, h), io.LimitReader(response.Body, maxInstallerBytes+1))
	closeErr := f.Close()
	if copyErr != nil {
		return cleanup(copyErr)
	}
	if closeErr != nil {
		return cleanup(closeErr)
	}
	if n != m.Size {
		return cleanup(errors.New("update download length mismatch"))
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), m.SHA256) {
		return cleanup(errors.New("update download checksum mismatch"))
	}
	// Never silently install or overwrite an active executable. User must approve.
	return destination, nil
}
func runVerifiedInstaller(path string) error {
	cmd := exec.Command(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	return cmd.Start()
}
