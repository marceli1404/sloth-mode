//go:build !windows

package main

import "log"

func startupError(msg string) { log.Println(msg) }
