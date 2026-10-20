package main

// API keys are stored on the PC under per-user DPAPI encryption on Windows.
// The website only communicates with localhost after explicit browser pairing.
import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxCredentialLength = 512

var serviceClient = &http.Client{Timeout: 12 * time.Second}

type integrationVault struct {
	Keys map[string]string `json:"keys"`
}
type integrationInfo struct {
	Provider  string `json:"provider"`
	Connected bool   `json:"connected"`
}

var allowedIntegration = map[string]bool{"serpapi": true, "elevenlabs": true}

func validProvider(p string) bool { return allowedIntegration[p] }
func validCredential(s string) bool {
	return len(s) >= 8 && len(s) <= maxCredentialLength && !strings.ContainsAny(s, "\x00\r\n") && strings.TrimSpace(s) == s
}
func vaultPath(base string) string { return filepath.Join(base, "integrations.vault.json") }
func readVault(base string) (integrationVault, error) {
	v := integrationVault{Keys: make(map[string]string)}
	content, err := os.ReadFile(vaultPath(base))
	if errors.Is(err, os.ErrNotExist) {
		return v, nil
	}
	if err != nil {
		return v, err
	}
	if len(content) > 16384 {
		return v, errors.New("local integrations store is too large")
	}
	if err = json.Unmarshal(content, &v); err != nil {
		return v, errors.New("local integration store could not be read")
	}
	if v.Keys == nil {
		v.Keys = make(map[string]string)
	}
	return v, nil
}
func saveVault(base string, v integrationVault) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	path := vaultPath(base)
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".integrations-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err = temp.Chmod(0600); err != nil {
		temp.Close()
		return err
	}
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}
func getProviderKey(v integrationVault, provider string) (string, error) {
	value := v.Keys[provider]
	if value == "" {
		return "", errors.New("Provider not connected")
	}
	protected, e := base64.StdEncoding.DecodeString(value)
	if e != nil {
		return "", errors.New("Invalid local credential")
	}
	plain, e := unsealAccountToken(protected)
	if e != nil {
		return "", errors.New("Could not unlock local credential")
	}
	return string(plain), nil
}
func putProviderKey(v *integrationVault, p, k string) error {
	if !validProvider(p) || !validCredential(k) {
		return errors.New("Choose a supported provider and enter its API key")
	}
	sealed, e := sealAccountToken([]byte(k))
	if e != nil {
		return errors.New("Local credential protection unavailable; Linux requires an unlocked Secret Service keyring")
	}
	if v.Keys == nil {
		v.Keys = map[string]string{}
	}
	v.Keys[p] = base64.StdEncoding.EncodeToString(sealed)
	return nil
}
func (b *Bridge) serveIntegrations(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/v1/integrations") {
		return false
	}
	b.vaultMu.Lock()
	defer b.vaultMu.Unlock() // Serializes vault writes independently of AI progress state.
	v, err := readVault(b.base)
	if err != nil {
		jsonReply(w, 500, map[string]string{"error": "Cannot open local integration store"})
		return true
	}
	switch r.URL.Path {
	case "/v1/integrations":
		if r.Method != http.MethodGet {
			break
		}
		info := []integrationInfo{}
		for _, p := range []string{"serpapi", "elevenlabs"} {
			info = append(info, integrationInfo{Provider: p, Connected: v.Keys[p] != ""})
		}
		jsonReply(w, 200, map[string]any{"integrations": info, "storage": "Local OS-protected keyring (DPAPI on Windows / Secret Service on Linux)"})
		return true
	case "/v1/integrations/connect":
		if r.Method != http.MethodPost {
			break
		}
		var in struct {
			Provider string `json:"provider"`
			Key      string `json:"key"`
		}
		if !b.readJSON(w, r, &in) {
			return true
		}
		if !validProvider(in.Provider) || !validCredential(in.Key) {
			jsonReply(w, 400, map[string]string{"error": "Invalid API provider or credential"})
			return true
		}
		if err = putProviderKey(&v, in.Provider, in.Key); err == nil {
			err = saveVault(b.base, v)
		}
		if err != nil {
			jsonReply(w, 500, map[string]string{"error": "Could not protect the API key locally"})
			return true
		}
		jsonReply(w, 200, map[string]any{"provider": in.Provider, "connected": true})
		return true
	case "/v1/integrations/disconnect":
		if r.Method != http.MethodPost {
			break
		}
		var in struct {
			Provider string `json:"provider"`
		}
		if !b.readJSON(w, r, &in) {
			return true
		}
		if !validProvider(in.Provider) {
			jsonReply(w, 400, map[string]string{"error": "Unknown provider"})
			return true
		}
		delete(v.Keys, in.Provider)
		if err = saveVault(b.base, v); err != nil {
			jsonReply(w, 500, map[string]string{"error": "Could not update local integration store"})
			return true
		}
		jsonReply(w, 200, map[string]any{"provider": in.Provider, "connected": false})
		return true
	case "/v1/integrations/test":
		if r.Method != http.MethodPost {
			break
		}
		var in struct {
			Provider string `json:"provider"`
		}
		if !b.readJSON(w, r, &in) {
			return true
		}
		if !validProvider(in.Provider) {
			jsonReply(w, 400, map[string]string{"error": "Unknown provider"})
			return true
		}
		key, e := getProviderKey(v, in.Provider)
		if e != nil {
			jsonReply(w, 409, map[string]string{"error": "Connect this provider first"})
			return true
		}
		// Only explicit user action triggers an external request. Never log credentials.
		ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
		defer cancel()
		target := "https://api.elevenlabs.io/v1/user"
		if in.Provider == "serpapi" {
			target = "https://serpapi.com/account?api_key=" + url.QueryEscape(key)
		}
		req, e := http.NewRequestWithContext(ctx, "GET", target, nil)
		if e != nil {
			jsonReply(w, 500, map[string]string{"error": "Could not start provider check"})
			return true
		}
		if in.Provider == "elevenlabs" {
			req.Header.Set("xi-api-key", key)
		}
		resp, e := serviceClient.Do(req)
		if e != nil {
			jsonReply(w, 502, map[string]string{"error": "Provider could not be reached"})
			return true
		}
		io.Copy(io.Discard, io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		if resp.StatusCode != 200 {
			jsonReply(w, 502, map[string]any{"error": "Provider rejected credentials or request", "status": resp.StatusCode})
			return true
		}
		jsonReply(w, 200, map[string]any{"provider": in.Provider, "verified": true})
		return true
	case "/v1/integrations/search":
		if r.Method != http.MethodPost {
			break
		}
		var in struct {
			Query string `json:"query"`
		}
		if !b.readJSON(w, r, &in) {
			return true
		}
		in.Query = strings.TrimSpace(in.Query)
		if len(in.Query) < 2 || len(in.Query) > 160 {
			jsonReply(w, 400, map[string]string{"error": "Use a search between 2 and 160 characters"})
			return true
		}
		key, e := getProviderKey(v, "serpapi")
		if e != nil {
			jsonReply(w, 409, map[string]string{"error": "Connect SerpApi first"})
			return true
		}
		query := url.Values{"api_key": {key}, "engine": {"google"}, "q": {in.Query}, "num": {"5"}}
		ctx, cancel := context.WithTimeout(r.Context(), 18*time.Second)
		defer cancel()
		req, e := http.NewRequestWithContext(ctx, "GET", "https://serpapi.com/search.json?"+query.Encode(), nil)
		if e != nil {
			jsonReply(w, 500, map[string]string{"error": "Invalid search"})
			return true
		}
		resp, e := serviceClient.Do(req)
		if e != nil {
			jsonReply(w, 502, map[string]string{"error": "Search provider unavailable"})
			return true
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			jsonReply(w, 502, map[string]any{"error": "Search provider returned an error", "status": resp.StatusCode})
			return true
		}
		var results struct {
			Organic []struct {
				Title   string `json:"title"`
				Link    string `json:"link"`
				Snippet string `json:"snippet"`
			} `json:"organic_results"`
		}
		if e = json.NewDecoder(io.LimitReader(resp.Body, 512<<10)).Decode(&results); e != nil {
			jsonReply(w, 502, map[string]string{"error": "Unexpected search response"})
			return true
		}
		out := make([]map[string]string, 0, 5)
		for _, result := range results.Organic {
			if len(out) == 5 {
				break
			}
			u, e := url.Parse(result.Link)
			if e != nil || u.Scheme != "https" || u.Host == "" {
				continue
			}
			out = append(out, map[string]string{"title": limitString(result.Title, 180), "url": result.Link, "snippet": limitString(result.Snippet, 320)})
		}
		jsonReply(w, 200, map[string]any{"results": out, "source": "SerpApi", "count": len(out)})
		return true
	}
	jsonReply(w, 405, map[string]string{"error": "Unsupported integration action"})
	return true
}
func limitString(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// The digest is non-secret metadata, useful for human-driven troubleshooting.
func providerKeyFingerprint(key string) string {
	h := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", h[:4])
}
