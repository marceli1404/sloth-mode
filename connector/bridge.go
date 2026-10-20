package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Bridge struct {
	base     string
	store    *Store
	secret   string
	mu       sync.Mutex
	vaultMu  sync.Mutex
	busy     bool
	progress string
	backend  string
	pairs    *PairManager
}

func getOrCreateSecret(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		if len(strings.TrimSpace(string(b))) >= 40 {
			return strings.TrimSpace(string(b)), nil
		}
		return "", errors.New("invalid pairing secret; safely back up then remove bridge-secret to regenerate")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", err
	}
	value := base64.RawURLEncoding.EncodeToString(raw)
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	// O_EXCL ensures racing connector instances cannot overwrite the key.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if errors.Is(err, os.ErrExist) {
		b, e := os.ReadFile(path)
		return strings.TrimSpace(string(b)), e
	}
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(value + "\n")
	closeErr := f.Close()
	if err != nil {
		return "", err
	}
	return value, closeErr
}
func newBridge(base string, store *Store, secret string) *Bridge {
	pairs, err := openPairs(base)
	if err != nil {
		log.Printf("trusted browsers registry unavailable: %v", err)
		pairs = &PairManager{filename: filepath.Join(base, "trusted-browsers-recovered.json"), devices: map[string]PairedBrowser{}, nonces: map[string]int64{}}
	}
	return &Bridge{base: base, store: store, secret: secret, backend: "cpu", progress: "Local connector ready", pairs: pairs}
}
func (b *Bridge) setProgress(s string) { b.mu.Lock(); b.progress = s; b.mu.Unlock() }
func (b *Bridge) state() (bool, string, string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.busy, b.progress, b.backend
}
func (b *Bridge) startJob(job func()) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.busy {
		return false
	}
	b.busy = true
	go func() {
		defer func() {
			if value := recover(); value != nil {
				log.Printf("background job panic recovered: %v\n%s", value, debug.Stack())
				b.setProgress("Unexpected background error; notes were preserved. Restart the app or open diagnostics.")
			}
			b.mu.Lock()
			b.busy = false
			b.mu.Unlock()
		}()
		job()
	}()
	return true
}
func jsonReply(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
func (b *Bridge) readJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		jsonReply(w, 400, map[string]string{"error": "Invalid request: " + err.Error()})
		return false
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		jsonReply(w, 400, map[string]string{"error": "Unexpected additional JSON"})
		return false
	}
	return true
}
func (b *Bridge) permitted(r *http.Request) bool {
	return r.Header.Get("Origin") == bridgeOrigin && r.Host == "127.0.0.1:"+bridgePort
}
func (b *Bridge) authorized(r *http.Request) bool {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	return len(token) == len(b.secret) && subtle.ConstantTimeCompare([]byte(token), []byte(b.secret)) == 1
}
func (b *Bridge) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reject unexpected hosts and cross-origin requests before they reach any handler.
		if !b.permitted(r) {
			jsonReply(w, 403, map[string]string{"error": "Origin not allowed"})
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", bridgeOrigin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Sloth-Device, X-Sloth-Time, X-Sloth-Nonce, X-Sloth-Signature")
		w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		if r.URL.Path == "/v1/health" && r.Method == http.MethodGet {
			jsonReply(w, 200, map[string]any{"ok": true, "name": "Sloth Mode Connector", "version": bridgeVersion, "pairing_required": true})
			return
		}
		if b.servePair(w, r) {
			return
		}
		if !b.authorized(r) && !b.pairs.signed(r) {
			jsonReply(w, 401, map[string]string{"error": "Pair this browser from Sloth Mode Setup & Pairing"})
			return
		}
		if b.serveIntegrations(w, r) {
			return
		}
		busy, progress, backend := b.state()
		switch r.URL.Path {
		case "/v1/status":
			if r.Method != http.MethodGet {
				break
			}
			jsonReply(w, 200, map[string]any{"connected": true, "model_ready": detectAI(b.base, backend), "busy": busy, "progress": progress, "backend": backend, "version": bridgeVersion, "capabilities": []string{"tasks.edit", "tasks.bulk", "focus.dashboard", "pairing.device-key", "integrations.local-vault", "integrations.serpapi-search"}})
			return
		case "/v1/tasks":
			if r.Method == http.MethodGet {
				jsonReply(w, 200, map[string]any{"tasks": b.store.All()})
				return
			}
			if r.Method == http.MethodPost {
				var in struct {
					Text string `json:"text"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				if err := b.store.Add(in.Text); err != nil {
					jsonReply(w, 400, map[string]string{"error": err.Error()})
					return
				}
				jsonReply(w, 201, map[string]any{"tasks": b.store.All()})
				return
			}
		case "/v1/tasks/edit":
			if r.Method == http.MethodPost {
				var in struct {
					ID   int64  `json:"id"`
					Text string `json:"text"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				if err := b.store.Edit(in.ID, in.Text); err != nil {
					jsonReply(w, 400, map[string]string{"error": err.Error()})
					return
				}
				jsonReply(w, 200, map[string]any{"tasks": b.store.All()})
				return
			}
		case "/v1/tasks/bulk":
			if r.Method == http.MethodPost {
				var in struct {
					Lines []string `json:"lines"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				count, err := b.store.AddBatch(in.Lines)
				if err != nil {
					jsonReply(w, 400, map[string]string{"error": err.Error()})
					return
				}
				jsonReply(w, 201, map[string]any{"count": count, "tasks": b.store.All()})
				return
			}
		case "/v1/tasks/complete":
			if r.Method == http.MethodPost {
				var in struct {
					ID int64 `json:"id"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				if err := b.store.Complete(in.ID); err != nil {
					jsonReply(w, 404, map[string]string{"error": err.Error()})
					return
				}
				jsonReply(w, 200, map[string]any{"tasks": b.store.All()})
				return
			}
		case "/v1/tasks/delete":
			if r.Method == http.MethodPost {
				var in struct {
					ID int64 `json:"id"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				if err := b.store.Delete(in.ID); err != nil {
					jsonReply(w, 404, map[string]string{"error": err.Error()})
					return
				}
				jsonReply(w, 200, map[string]any{"tasks": b.store.All()})
				return
			}
		case "/v1/hardware":
			if r.Method == http.MethodGet {
				p := profileHardware()
				p.Recommended = preferredQuant(b.base, p)
				jsonReply(w, 200, p)
				return
			}
		case "/v1/model/preference":
			if r.Method == http.MethodPost {
				var in struct {
					Quant string `json:"quant"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				if err := setPreferredQuant(b.base, in.Quant); err != nil {
					jsonReply(w, 400, map[string]string{"error": err.Error()})
					return
				}
				jsonReply(w, 200, map[string]string{"selected": in.Quant})
				return
			}
		case "/v1/setup":
			if r.Method == http.MethodPost {
				var in struct {
					Backend string `json:"backend"`
					Consent bool   `json:"consent"`
				}
				if !b.readJSON(w, r, &in) {
					return
				}
				if !in.Consent {
					jsonReply(w, 400, map[string]string{"error": "Download requires explicit approval"})
					return
				}
				if in.Backend != "cpu" && in.Backend != "vulkan" {
					jsonReply(w, 400, map[string]string{"error": "Select CPU or Vulkan"})
					return
				}
				if !b.startJob(func() {
					b.mu.Lock()
					b.backend = in.Backend
					b.mu.Unlock()
					ctx, cancel := context.WithTimeout(context.Background(), 40*time.Minute)
					defer cancel()
					err := installAI(ctx, b.base, in.Backend, b.setProgress)
					if err != nil {
						b.setProgress("Setup failed: " + err.Error())
					}
				}) {
					jsonReply(w, 409, map[string]string{"error": "Another operation is running"})
					return
				}
				jsonReply(w, 202, map[string]any{"started": true})
				return
			}
		case "/v1/plan":
			if r.Method == http.MethodPost {
				if !b.startJob(func() {
					ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
					defer cancel()
					_, _, backend := b.state()
					items := b.store.Pending(10)
					if len(items) == 0 {
						b.setProgress("No unsorted notes. Add tasks or clear them first.")
						return
					}
					b.setProgress("Starting local Gemma...")
					e, err := startEngine(ctx, b.base, backend, b.setProgress)
					if err != nil {
						b.setProgress("AI unavailable: " + err.Error())
						return
					}
					defer e.stop()
					b.setProgress("Gemma is reviewing " + strconv.Itoa(len(items)) + " local notes...")
					out, err := e.triage(ctx, items)
					if err == nil {
						err = b.store.Apply(out, items)
					}
					if err != nil {
						b.setProgress("Gemma result not saved: " + err.Error())
						return
					}
					b.setProgress(fmt.Sprintf("Gemma sorted %d tasks locally", len(out)))
				}) {
					jsonReply(w, 409, map[string]string{"error": "Another operation is running"})
					return
				}
				jsonReply(w, 202, map[string]any{"started": true})
				return
			}
		}
		jsonReply(w, 404, map[string]string{"error": "Unknown endpoint or method"})
	})
}
