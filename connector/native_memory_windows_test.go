//go:build windows

package main

import (
	"bytes"
	"testing"
)

func TestNativeMemoryCopyRoundTrip(t *testing.T) {
	want := []byte("Sloth Mode native memory copy round trip")

	handle, _, _ := pAlloc.Call(gmemMoveable, uintptr(len(want)))
	if handle == 0 {
		t.Fatal("GlobalAlloc failed")
	}
	defer pFree.Call(handle)

	locked, _, _ := pLock.Call(handle)
	if locked == 0 {
		t.Fatal("GlobalLock failed")
	}
	defer pUnlock.Call(handle)

	copyToNative(locked, want)
	got := make([]byte, len(want))
	copyFromNative(got, locked)
	if !bytes.Equal(got, want) {
		t.Fatalf("native memory round trip mismatch: got %q, want %q", got, want)
	}
}
