package main

import (
	"os"
	"strings"
	"testing"
)

func TestIntegrationProviderAllowlist(t *testing.T) {
	for _, p := range []string{"serpapi", "elevenlabs"} {
		if !validProvider(p) {
			t.Fatal("missing allowed provider", p)
		}
	}
	for _, p := range []string{"serpapi/../../", "", "https://evil.example", "browserbase", "sentry"} {
		if validProvider(p) {
			t.Fatal("unexpected provider", p)
		}
	}
}
func TestCredentialValidation(t *testing.T) {
	if !validCredential("abcdefghijklmnop") {
		t.Fatal("normal key rejected")
	}
	for _, secret := range []string{"short", "a\nbadcredential", strings.Repeat("x", 513), "    xabcdefghijk", "abcdefghijk\r"} {
		if validCredential(secret) {
			t.Fatal("invalid key allowed")
		}
	}
}
func TestIntegrationEndpointsDoNotRevealKeys(t *testing.T) {
	b := testBridge(t)
	w := invoke(t, b, "GET", "/v1/integrations", bridgeOrigin, b.secret, "")
	if w.Code != 200 || strings.Contains(w.Body.String(), b.secret) {
		t.Fatal("integration status unsafe", w.Code, w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/integrations/connect", bridgeOrigin, b.secret, `{"provider":"invalid","key":"private-key-not-to-leak"}`)
	if w.Code != 400 || strings.Contains(w.Body.String(), "private-key-not-to-leak") {
		t.Fatal("invalid provider response leaking", w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/integrations/connect", "https://not-sloth.example", b.secret, `{"provider":"serpapi","key":"private-key-not-to-leak"}`)
	if w.Code != 403 {
		t.Fatal("invalid origin admitted", w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/integrations/connect", bridgeOrigin, "wrong-key", `{"provider":"serpapi","key":"private-key-not-to-leak"}`)
	if w.Code != 401 {
		t.Fatal("invalid auth admitted", w.Body.String())
	}
	if _, e := os.Stat(vaultPath(b.base)); !os.IsNotExist(e) {
		t.Fatal("invalid request wrote vault")
	}
}
func TestIntegrationNoSearchWithoutApproval(t *testing.T) {
	b := testBridge(t)
	w := invoke(t, b, "POST", "/v1/integrations/search", bridgeOrigin, b.secret, `{"query":"safe query"}`)
	if w.Code != 409 {
		t.Fatal("search possible without key", w.Body.String())
	}
	w = invoke(t, b, "POST", "/v1/integrations/search", bridgeOrigin, b.secret, `{"query":"x"}`)
	if w.Code != 400 {
		t.Fatal("too short search allowed")
	}
}
