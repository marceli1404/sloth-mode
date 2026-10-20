package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Public-key trusted-browser linking: the private key never leaves the browser.
// The initial pairing secret is exchanged only once; signed requests authenticate
// that particular browser across page reloads, without a stored bearer token.
type PairedBrowser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	X     string `json:"x"`
	Y     string `json:"y"`
	Added string `json:"added"`
}
type PairManager struct {
	mu       sync.Mutex
	filename string
	devices  map[string]PairedBrowser
	nonces   map[string]int64
}

func openPairs(base string) (*PairManager, error) {
	p := &PairManager{filename: filepath.Join(base, "trusted-browsers.json"), devices: make(map[string]PairedBrowser), nonces: make(map[string]int64)}
	raw, e := os.ReadFile(p.filename)
	if errors.Is(e, os.ErrNotExist) {
		return p, nil
	}
	if e != nil {
		return nil, e
	}
	var all []PairedBrowser
	if e = json.Unmarshal(raw, &all); e != nil {
		return nil, e
	}
	for _, x := range all {
		if x.ID == "" {
			return nil, errors.New("invalid trusted-browser registry")
		}
		p.devices[x.ID] = x
	}
	return p, nil
}
func (p *PairManager) save(next map[string]PairedBrowser) error {
	all := make([]PairedBrowser, 0, len(next))
	for _, d := range next {
		all = append(all, d)
	}
	raw, e := json.MarshalIndent(all, "", "  ")
	if e != nil {
		return e
	}
	tmp := p.filename + ".tmp"
	if e = os.WriteFile(tmp, raw, 0600); e != nil {
		return e
	}
	if e = os.Rename(tmp, p.filename); e != nil {
		os.Remove(tmp)
		return e
	}
	p.devices = next
	return nil
}
func validCoordinates(x, y string) (*big.Int, *big.Int, bool) {
	xb, e := base64.RawURLEncoding.DecodeString(x)
	if e != nil || len(xb) != 32 {
		return nil, nil, false
	}
	yb, e := base64.RawURLEncoding.DecodeString(y)
	if e != nil || len(yb) != 32 {
		return nil, nil, false
	}
	X, Y := new(big.Int).SetBytes(xb), new(big.Int).SetBytes(yb)
	return X, Y, elliptic.P256().IsOnCurve(X, Y)
}
func (p *PairManager) register(name, x, y string) (string, error) {
	if _, _, ok := validCoordinates(x, y); !ok {
		return "", errors.New("invalid P-256 public key")
	}
	if len([]rune(name)) > 80 {
		return "", errors.New("device name too long")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.devices) >= 20 {
		return "", errors.New("device limit reached; revoke an old browser first")
	}
	idbytes := make([]byte, 18)
	if _, e := rand.Read(idbytes); e != nil {
		return "", e
	}
	id := base64.RawURLEncoding.EncodeToString(idbytes)
	next := make(map[string]PairedBrowser, len(p.devices)+1)
	for k, v := range p.devices {
		next[k] = v
	}
	next[id] = PairedBrowser{ID: id, Name: strings.TrimSpace(name), X: x, Y: y, Added: time.Now().UTC().Format(time.RFC3339)}
	return id, p.save(next)
}
func (p *PairManager) revoke(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.devices[id]; !ok {
		return errors.New("browser not found")
	}
	next := make(map[string]PairedBrowser, len(p.devices))
	for k, v := range p.devices {
		if k != id {
			next[k] = v
		}
	}
	return p.save(next)
}
func (p *PairManager) signed(r *http.Request) bool {
	device := r.Header.Get("X-Sloth-Device")
	if device == "" {
		return false
	}
	stamp, e := strconv.ParseInt(r.Header.Get("X-Sloth-Time"), 10, 64)
	if e != nil {
		return false
	}
	now := time.Now().Unix()
	if stamp < now-90 || stamp > now+90 {
		return false
	}
	nonce := r.Header.Get("X-Sloth-Nonce")
	nb, e := base64.RawURLEncoding.DecodeString(nonce)
	if e != nil || len(nb) < 16 || len(nb) > 48 {
		return false
	}
	signature, e := base64.RawURLEncoding.DecodeString(r.Header.Get("X-Sloth-Signature"))
	if e != nil || len(signature) != 64 {
		return false
	}
	body, e := io.ReadAll(io.LimitReader(r.Body, 32<<10+1))
	if e != nil || len(body) > 32<<10 {
		return false
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	bodyHash := sha256.Sum256(body)
	signedText := strings.Join([]string{r.Method, r.URL.Path, nonce, strconv.FormatInt(stamp, 10), hex.EncodeToString(bodyHash[:])}, "\n")
	digest := sha256.Sum256([]byte(signedText))
	p.mu.Lock()
	defer p.mu.Unlock()
	d, ok := p.devices[device]
	if !ok {
		return false
	}
	X, Y, ok := validCoordinates(d.X, d.Y)
	if !ok {
		return false
	}
	key := &ecdsa.PublicKey{Curve: elliptic.P256(), X: X, Y: Y}
	if !ecdsa.Verify(key, digest[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
		return false
	}
	replayKey := device + ":" + nonce
	if _, exists := p.nonces[replayKey]; exists {
		return false
	}
	for k, v := range p.nonces {
		if v < now-120 {
			delete(p.nonces, k)
		}
	}
	if len(p.nonces) > 8000 {
		return false
	}
	p.nonces[replayKey] = now
	return true
}
func (b *Bridge) servePair(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/v1/pair" && r.Method == http.MethodPost {
		if !b.authorized(r) {
			jsonReply(w, 401, map[string]string{"error": "Invalid setup key"})
			return true
		}
		var in struct {
			Name string `json:"name"`
			X    string `json:"x"`
			Y    string `json:"y"`
		}
		if !b.readJSON(w, r, &in) {
			return true
		}
		id, e := b.pairs.register(in.Name, in.X, in.Y)
		if e != nil {
			jsonReply(w, 400, map[string]string{"error": e.Error()})
			return true
		}
		jsonReply(w, 201, map[string]string{"device_id": id})
		return true
	}
	if r.URL.Path == "/v1/pair/revoke" && r.Method == http.MethodPost {
		if !b.pairs.signed(r) {
			jsonReply(w, 401, map[string]string{"error": "Unknown paired device"})
			return true
		}
		id := r.Header.Get("X-Sloth-Device")
		if e := b.pairs.revoke(id); e != nil {
			jsonReply(w, 400, map[string]string{"error": e.Error()})
			return true
		}
		jsonReply(w, 200, map[string]bool{"revoked": true})
		return true
	}
	return false
}
