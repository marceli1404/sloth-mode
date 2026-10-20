//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func configureHiddenProcess(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
