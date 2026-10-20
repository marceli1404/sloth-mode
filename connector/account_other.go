//go:build !windows && !linux

package main

import "errors"

func sealAccountToken(_ []byte) ([]byte, error) {
	return nil, errors.New("Account linking is supported on Windows only")
}
func unsealAccountToken(_ []byte) ([]byte, error) {
	return nil, errors.New("Account linking is supported on Windows only")
}
