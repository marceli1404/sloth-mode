package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestUpdaterVersions(t *testing.T) {
	for _, x := range []struct {
		a, b string
		want bool
	}{{"0.9.0", "0.10.0", true}, {"0.10.0", "0.10.0", false}, {"0.10.0", "0.9.99", false}} {
		v, e := newerVersion(x.a, x.b)
		if e != nil || v != x.want {
			t.Fatalf("%s vs %s: %v %v", x.a, x.b, v, e)
		}
	}
	if _, e := newerVersion("oops", "0.10.0"); e == nil {
		t.Fatal("bad version accepted")
	}
}
func TestManifestRejectsUnsigned(t *testing.T) {
	if e := verifyManifest(updateManifest{Format: 1, Version: "99.0.0", URL: "https://example.com/update.exe", SHA256: "abc", Signature: "bad"}); e == nil {
		t.Fatal("unsigned update accepted")
	}
}

func TestPublishedChannelSignature(t *testing.T) {
	b, err := os.ReadFile("../signing/current-stable-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var m updateManifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if err := verifyManifest(m); err != nil {
		t.Fatal(err)
	}
	m.Version = "0.10.1"
	if err := verifyManifest(m); err == nil {
		t.Fatal("tampered manifest accepted")
	}
}
