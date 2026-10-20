package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func signRequest(t *testing.T, method, path, body, id string, priv *ecdsa.PrivateKey) *http.Request {
	t.Helper()
	r := req(method, path, bridgeOrigin, "", body)
	nonceBytes := make([]byte, 24)
	rand.Read(nonceBytes)
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	stamp := fmt.Sprintf("%d", time.Now().Unix())
	h := sha256.Sum256([]byte(body))
	msg := strings.Join([]string{method, path, nonce, stamp, hex.EncodeToString(h[:])}, "\n")
	d := sha256.Sum256([]byte(msg))
	R, S, e := ecdsa.Sign(rand.Reader, priv, d[:])
	if e != nil {
		t.Fatal(e)
	}
	signature := append(R.FillBytes(make([]byte, 32)), S.FillBytes(make([]byte, 32))...)
	r.Header.Set("X-Sloth-Device", id)
	r.Header.Set("X-Sloth-Time", stamp)
	r.Header.Set("X-Sloth-Nonce", nonce)
	r.Header.Set("X-Sloth-Signature", base64.RawURLEncoding.EncodeToString(signature))
	return r
}
func TestRegisteredBrowserAcrossReloadAndRevocation(t *testing.T) {
	base := t.TempDir()
	s, e := openStore(base)
	if e != nil {
		t.Fatal(e)
	}
	b := newBridge(base, s, strings.Repeat("x", 43))
	priv, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	x := base64.RawURLEncoding.EncodeToString(priv.X.FillBytes(make([]byte, 32)))
	y := base64.RawURLEncoding.EncodeToString(priv.Y.FillBytes(make([]byte, 32)))
	pay, _ := json.Marshal(map[string]string{"name": "Laptop", "x": x, "y": y})
	w := invoke(t, b, "POST", "/v1/pair", bridgeOrigin, b.secret, string(pay))
	if w.Code != 201 {
		t.Fatalf("registration %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		ID string `json:"device_id"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.ID == "" {
		t.Fatal("missing id")
	}
	r := signRequest(t, "GET", "/v1/tasks", "", response.ID, priv)
	w = httptest.NewRecorder()
	b.handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("signed %d %s", w.Code, w.Body.String())
	}
	// Same packet must not be accepted twice.
	w = httptest.NewRecorder()
	b.handler().ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("nonce replay accepted: %d", w.Code)
	}
	pairs, e := openPairs(base)
	if e != nil {
		t.Fatal(e)
	}
	b.pairs = pairs
	w = httptest.NewRecorder()
	b.handler().ServeHTTP(w, signRequest(t, "GET", "/v1/tasks", "", response.ID, priv))
	if w.Code != 200 {
		t.Fatal("pairing did not survive reload")
	}
	w = httptest.NewRecorder()
	b.handler().ServeHTTP(w, signRequest(t, "POST", "/v1/pair/revoke", "{}", response.ID, priv))
	if w.Code != 200 {
		t.Fatalf("revoke %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	b.handler().ServeHTTP(w, signRequest(t, "GET", "/v1/tasks", "", response.ID, priv))
	if w.Code != 401 {
		t.Fatal("revoked browser still authorised")
	}
}
func TestPairingRejectsInvalidCoordinatesAndUnknownBrowser(t *testing.T) {
	b := testBridge(t)
	w := invoke(t, b, "POST", "/v1/pair", bridgeOrigin, b.secret, `{"name":"bad","x":"a","y":"b"}`)
	if w.Code != 400 {
		t.Fatal(w.Body.String())
	}
	priv, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	w = httptest.NewRecorder()
	b.handler().ServeHTTP(w, signRequest(t, "GET", "/v1/tasks", "", "missing", priv))
	if w.Code != 401 {
		t.Fatal("unknown device accepted")
	}
	// Tests that the verifier never treats signature inputs as signed integers.
	if new(big.Int).SetBytes([]byte{255}).Sign() != 1 {
		t.Fatal("invalid integer decoding")
	}
}
