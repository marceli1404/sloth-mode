//go:build linux

package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestTar(t *testing.T, items []struct {
	name string
	kind byte
	link string
	text string
}) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "assets.tar.gz")
	out, e := os.Create(file)
	if e != nil {
		t.Fatal(e)
	}
	gz := gzip.NewWriter(out)
	tarWriter := tar.NewWriter(gz)
	for _, item := range items {
		hdr := &tar.Header{Name: item.name, Typeflag: item.kind, Mode: 0755, Linkname: item.link, Size: int64(len(item.text))}
		if item.kind == tar.TypeSymlink {
			hdr.Size = 0
		}
		if e = tarWriter.WriteHeader(hdr); e != nil {
			t.Fatal(e)
		}
		if item.kind != tar.TypeSymlink {
			if _, e = io.WriteString(tarWriter, item.text); e != nil {
				t.Fatal(e)
			}
		}
	}
	if e = tarWriter.Close(); e != nil {
		t.Fatal(e)
	}
	if e = gz.Close(); e != nil {
		t.Fatal(e)
	}
	if e = out.Close(); e != nil {
		t.Fatal(e)
	}
	return file
}
func TestLinuxAssetSelection(t *testing.T) {
	a := []asset{{Name: "llama-b999-bin-ubuntu-x64.tar.gz", Digest: "sha256:abc"}, {Name: "llama-b999-bin-ubuntu-vulkan-x64.tar.gz", Digest: "sha256:def"}, {Name: "cudart-llama-b999-bin-ubuntu-x64.tar.gz", Digest: "sha256:xyz"}}
	for _, x := range []struct{ backend, token string }{{"cpu", "ubuntu-x64"}, {"vulkan", "vulkan-x64"}} {
		s, e := pickAsset(a, x.backend)
		if e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(s.Name, x.token) {
			t.Fatalf("unexpected asset %s", s.Name)
		}
	}
}
func TestLinuxRuntimeExtraction(t *testing.T) {
	tarball := writeTestTar(t, []struct {
		name string
		kind byte
		link string
		text string
	}{{name: "build/bin/llama-server", kind: tar.TypeReg, text: "fake binary"}, {name: "build/lib/libggml.so.0", kind: tar.TypeReg, text: "lib"}, {name: "build/lib/libggml.so", kind: tar.TypeSymlink, link: "libggml.so.0"}})
	dest := filepath.Join(t.TempDir(), "runtime")
	if e := extractLinuxRuntime(tarball, dest); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(dest, "llama-server")); e != nil {
		t.Fatal(e)
	}
	if link, e := os.Readlink(filepath.Join(dest, "libggml.so")); e != nil || link != "libggml.so.0" {
		t.Fatalf("link %s, %v", link, e)
	}
}
func TestRejectEscapingRuntimeArchive(t *testing.T) {
	tarball := writeTestTar(t, []struct {
		name string
		kind byte
		link string
		text string
	}{{name: "../escape", kind: tar.TypeReg, text: "owned"}, {name: "bin/llama-server", kind: tar.TypeReg, text: "binary"}})
	if e := extractLinuxRuntime(tarball, filepath.Join(t.TempDir(), "runtime")); e == nil {
		t.Fatal("path traversal accepted")
	}
}
func TestAccountLinkingDoesNotRequireKeyringForGuest(t *testing.T) {
	if _, err := os.Stat(filepath.Join(t.TempDir(), "account-link.json")); !os.IsNotExist(err) {
		t.Fatal("guest unexpectedly linked")
	}
}
