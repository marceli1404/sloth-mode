//go:build linux

package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Official llama.cpp Ubuntu releases are tar.gz bundles. Extract their
// shared libraries alongside llama-server; allow only safe relative symlinks.
func extractLinuxRuntime(archive, dest string) error {
	in, e := os.Open(archive)
	if e != nil {
		return e
	}
	defer in.Close()
	gz, e := gzip.NewReader(in)
	if e != nil {
		return e
	}
	defer gz.Close()
	tmp := dest + "-extract"
	_ = os.RemoveAll(tmp)
	if e = os.MkdirAll(tmp, 0700); e != nil {
		return e
	}
	defer os.RemoveAll(tmp)
	tr := tar.NewReader(gz)
	count := 0
	total := int64(0)
	names := map[string]bool{}
	found := false
	type delayedLink struct{ file, target string }
	symlinks := []delayedLink{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if h.Typeflag == tar.TypeDir {
			continue
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA && h.Typeflag != tar.TypeSymlink {
			return fmt.Errorf("unexpected archive entry type %d", h.Typeflag)
		}
		count++
		if count > 700 {
			return errors.New("too many archive files")
		}
		normalized := filepath.ToSlash(filepath.Clean(h.Name))
		if strings.HasPrefix(normalized, "../") || normalized == ".." || strings.HasPrefix(normalized, "/") || strings.Contains(normalized, ":") {
			return errors.New("unsafe archive member")
		}
		name := filepath.Base(normalized)
		if name == "" || name == "." || name == ".." {
			return errors.New("invalid binary name")
		}
		if names[name] {
			return errors.New("duplicate runtime filename: " + name)
		}
		names[name] = true
		path := filepath.Join(tmp, name)
		if h.Typeflag == tar.TypeSymlink {
			target := filepath.Base(h.Linkname)
			if target != h.Linkname || target == "." || target == ".." || strings.ContainsAny(target, "/\\") {
				return errors.New("unsafe runtime library link")
			}
			symlinks = append(symlinks, delayedLink{path, target})
			continue
		}
		if h.Size < 0 || h.Size > 150<<20 {
			return errors.New("oversize runtime file")
		}
		total += h.Size
		if total > 750<<20 {
			return errors.New("runtime too large")
		}
		mode := os.FileMode(0600)
		if strings.HasPrefix(name, "llama-") || strings.HasPrefix(name, "ggml-") {
			mode = 0700
		}
		destfile, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if e != nil {
			return e
		}
		n, e := io.Copy(destfile, io.LimitReader(tr, h.Size+1))
		ce := destfile.Close()
		if e != nil {
			return e
		}
		if ce != nil {
			return ce
		}
		if n != h.Size {
			return errors.New("incomplete runtime entry")
		}
		if name == "llama-server" {
			found = true
		}
	}
	for _, link := range symlinks {
		if !names[link.target] {
			return errors.New("dangling binary symlink")
		}
		if e = os.Symlink(link.target, link.file); e != nil {
			return e
		}
	}
	if !found {
		return errors.New("llama-server missing from official archive")
	}
	if e = os.MkdirAll(filepath.Dir(dest), 0700); e != nil {
		return e
	}
	if _, e = os.Stat(dest); e == nil {
		return errors.New("runtime destination already exists; refusing overwrite")
	}
	return os.Rename(tmp, dest)
}
