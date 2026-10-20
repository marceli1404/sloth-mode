//go:build !windows && !linux

package main

import "os/exec"

func configureHiddenProcess(c *exec.Cmd) {}
func runNative(base string, s *Store)    {}
