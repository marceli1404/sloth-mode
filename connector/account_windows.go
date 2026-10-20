//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// Windows DPAPI user-scope encryption, not plaintext tokens in a JSON file.
type dataBlob struct {
	Size uint32
	Data *byte
}

var crypt32 = syscall.NewLazyDLL("crypt32.dll")
var protect = crypt32.NewProc("CryptProtectData")
var unprotect = crypt32.NewProc("CryptUnprotectData")
var localFree = syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")

func dpapi(in []byte, decrypt bool) ([]byte, error) {
	if len(in) == 0 {
		return nil, syscall.EINVAL
	}
	source := dataBlob{Size: uint32(len(in)), Data: &in[0]}
	var target dataBlob
	fn := protect
	if decrypt {
		fn = unprotect
	}
	r, _, e := fn.Call(uintptr(unsafe.Pointer(&source)), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&target)))
	if r == 0 {
		if e != nil && e != syscall.Errno(0) {
			return nil, e
		}
		return nil, syscall.EINVAL
	}
	defer localFree.Call(uintptr(unsafe.Pointer(target.Data)))
	return append([]byte{}, unsafe.Slice(target.Data, target.Size)...), nil
}
func sealAccountToken(s []byte) ([]byte, error)   { return dpapi(s, false) }
func unsealAccountToken(s []byte) ([]byte, error) { return dpapi(s, true) }
