package main

// Explicit, revocable account linking. The link token is never the localhost
// pairing key and is never sent to the website's JavaScript or Cloudflare Access.
import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const enrolEndpoint = "https://enroll.sloth.marceli.dev"

var accountCodePattern = regexp.MustCompile(`^[A-Z2-7]{21}$`)
var accountClient = &http.Client{Timeout: 12 * time.Second}

type accountCredential struct {
	DeviceID    string `json:"device_id"`
	Label       string `json:"label"`
	SealedToken string `json:"sealed_token"`
}

func saveAccountCredential(base string, cred accountCredential) error {
	file := filepath.Join(base, "account-link.json")
	b, e := json.MarshalIndent(cred, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.OpenFile(file+".tmp", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	if _, e = f.Write(b); e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		_ = os.Remove(file + ".tmp")
		return e
	}
	if e = os.Rename(file+".tmp", file); e != nil {
		_ = os.Remove(file + ".tmp")
		return e
	}
	return nil
}
func readAccountCredential(base string) (accountCredential, string, error) {
	var c accountCredential
	b, e := os.ReadFile(filepath.Join(base, "account-link.json"))
	if e != nil {
		return c, "", e
	}
	if e = json.Unmarshal(b, &c); e != nil {
		return c, "", e
	}
	sealed, e := base64.StdEncoding.DecodeString(c.SealedToken)
	if e != nil {
		return c, "", e
	}
	plain, e := unsealAccountToken(sealed)
	if e != nil {
		return c, "", e
	}
	return c, string(plain), nil
}
func linkAccount(base, code, label string) error {
	code = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(code), " ", ""), "-", ""))
	if !accountCodePattern.MatchString(code) {
		return errors.New("Enter the 21-character code from your Sloth Mode account page")
	}
	label = strings.TrimSpace(label)
	if len(label) < 2 || len(label) > 60 {
		return errors.New("Invalid computer label")
	}
	body, _ := json.Marshal(map[string]string{"code": code, "label": label})
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	request, e := http.NewRequestWithContext(ctx, http.MethodPost, enrolEndpoint+"/v1/claim", bytes.NewReader(body))
	if e != nil {
		return e
	}
	request.Header.Set("Content-Type", "application/json")
	response, e := accountClient.Do(request)
	if e != nil {
		return fmt.Errorf("Cannot reach Sloth account service: %w", e)
	}
	defer response.Body.Close()
	limited, e := io.ReadAll(io.LimitReader(response.Body, 4096))
	if e != nil {
		return e
	}
	if response.StatusCode != 201 {
		return fmt.Errorf("Account linking failed (HTTP %d). Check that the code is new and has not expired", response.StatusCode)
	}
	var received struct {
		DeviceID string `json:"device_id"`
		Token    string `json:"token"`
	}
	if json.Unmarshal(limited, &received) != nil || received.DeviceID == "" || len(received.Token) != 43 {
		return errors.New("Invalid account response")
	}
	sealed, e := sealAccountToken([]byte(received.Token))
	if e != nil {
		return fmt.Errorf("Could not protect account credential: %w", e)
	}
	return saveAccountCredential(base, accountCredential{DeviceID: received.DeviceID, Label: label, SealedToken: base64.StdEncoding.EncodeToString(sealed)})
}
func heartbeat(base string) error {
	_, token, e := readAccountCredential(base)
	if e != nil {
		return e
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	req, e := http.NewRequestWithContext(ctx, "POST", enrolEndpoint+"/v1/heartbeat", strings.NewReader("{}"))
	if e != nil {
		return e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, e := accountClient.Do(req)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("device heartbeat rejected: %d", resp.StatusCode)
	}
	return nil
}
func runAccountHeartbeat(base string, stop <-chan struct{}) {
	tick := time.NewTicker(90 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if _, e := os.Stat(filepath.Join(base, "account-link.json")); e == nil {
				if err := heartbeat(base); err != nil {
					fmt.Fprintf(os.Stderr, "Sloth account heartbeat: %v\n", err)
				}
			}
		case <-stop:
			return
		}
	}
}
