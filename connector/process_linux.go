//go:build linux

package main

import "os/exec"

func configureHiddenProcess(_ *exec.Cmd) {}
func runNative(_ string, _ *Store)       {}
