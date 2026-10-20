package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testBridge(t *testing.T) *Bridge {
	t.Helper()
	base := t.TempDir()
	s, e := openStore(base)
	if e != nil {
		t.Fatal(e)
	}
	return newBridge(base, s, strings.Repeat("x", 43))
}
func req(method, path, origin, auth, stringBody string) *http.Request {
	r := httptest.NewRequest(method, "http://127.0.0.1:"+bridgePort+path, strings.NewReader(stringBody))
	r.Header.Set("Origin", origin)
	if auth != "" {
		r.Header.Set("Authorization", "Bearer "+auth)
	}
	if stringBody != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	return r
}
func invoke(t *testing.T, b *Bridge, method, path, origin, auth, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	b.handler().ServeHTTP(w, req(method, path, origin, auth, body))
	return w
}
func TestBridgeOriginAndPairing(t *testing.T) {
	b := testBridge(t)
	for _, origin := range []string{"https://evil.example", "http://sloth.marceli.dev", "null", ""} {
		w := invoke(t, b, "GET", "/v1/health", origin, "", "")
		if w.Code != 403 {
			t.Errorf("origin %q: expected 403, got %d", origin, w.Code)
		}
	}
	w := invoke(t, b, "GET", "/v1/health", bridgeOrigin, "", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "pairing_required") {
		t.Fatalf("health: %d %s", w.Code, w.Body.String())
	}
	w = invoke(t, b, "OPTIONS", "/v1/tasks", bridgeOrigin, "", "")
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") != bridgeOrigin {
		t.Fatalf("preflight %d %#v", w.Code, w.Header())
	}
	w = invoke(t, b, "GET", "/v1/tasks", bridgeOrigin, "wrong", "")
	if w.Code != 401 {
		t.Fatal("bad secret accepted")
	}
	w = invoke(t, b, "GET", "/v1/tasks", bridgeOrigin, b.secret, "")
	if w.Code != 200 {
		t.Fatal("paired request rejected")
	}
	w = httptest.NewRecorder()
	r := req("GET", "/v1/health", bridgeOrigin, "", "")
	r.Host = "example.com"
	b.handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("DNS rebinding host accepted")
	}
}
func TestBridgeNotesAndValidation(t *testing.T) {
	b := testBridge(t)
	key := b.secret
	w := invoke(t, b, "POST", "/v1/tasks", bridgeOrigin, key, `{"text":" Send the report "}`)
	if w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	w = invoke(t, b, "GET", "/v1/tasks", bridgeOrigin, key, "")
	var out struct {
		Tasks []Task `json:"tasks"`
	}
	if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil || len(out.Tasks) != 1 || out.Tasks[0].Text != "Send the report" {
		t.Fatal(w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/tasks/complete", bridgeOrigin, key, `{"id":1}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"completed":true`) {
		t.Fatal(w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/tasks/delete", bridgeOrigin, key, `{"id":1}`)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"tasks":[]`) {
		t.Fatal(w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/tasks", bridgeOrigin, key, `{"text":"okay","extra":true}`)
	if w.Code != 400 {
		t.Fatal("unknown fields allowed")
	}
	w = invoke(t, b, "POST", "/v1/tasks", bridgeOrigin, key, `{"text":"one"}{"text":"two"}`)
	if w.Code != 400 {
		t.Fatal("trailing JSON accepted")
	}
	w = invoke(t, b, "POST", "/v1/tasks", bridgeOrigin, key, `{"text":"`+strings.Repeat("q", 40000)+`"}`)
	if w.Code != 400 {
		t.Fatal("oversized payload allowed")
	}
}
func TestBridgeSetupRequiresConsent(t *testing.T) {
	b := testBridge(t)
	k := b.secret
	for _, body := range []string{`{"backend":"cpu","consent":false}`, `{"backend":"invalid","consent":true}`} {
		w := invoke(t, b, "POST", "/v1/setup", bridgeOrigin, k, body)
		if w.Code != 400 {
			t.Fatalf("setup bypass %d %s", w.Code, w.Body.String())
		}
	}
	w := invoke(t, b, "GET", "/v1/status", bridgeOrigin, k, "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"model_ready":false`) {
		t.Fatal(w.Body.String())
	}
}
func TestSecretCreationPersistence(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "bridge-secret")
	a, e := getOrCreateSecret(p)
	if e != nil || len(a) != 43 {
		t.Fatalf("unexpected key %q %v", a, e)
	}
	b, e := getOrCreateSecret(p)
	if e != nil || a != b {
		t.Fatal("secret changed")
	}
	fi, e := os.Stat(p)
	if e != nil || fi.Size() < 43 {
		t.Fatal("secret not persisted")
	}
}
func TestMalformedJSON(t *testing.T) {
	b := testBridge(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "http://127.0.0.1:"+bridgePort+"/v1/tasks", bytes.NewBufferString("{bad"))
	r.Header.Set("Origin", bridgeOrigin)
	r.Header.Set("Authorization", "Bearer "+b.secret)
	b.handler().ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
}
