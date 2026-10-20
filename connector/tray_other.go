//go:build !windows

package main

import "net/http"

func runTray(b *Bridge, s *http.Server) {}

func showExistingSetup() bool { return false }
