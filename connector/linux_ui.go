//go:build linux

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func linuxShow(text string) {
	if _, e := exec.LookPath("zenity"); e == nil {
		cmd := exec.Command("zenity", "--text-info", "--title=Sloth Mode Setup & Pairing", "--width=620", "--height=460")
		cmd.Stdin = strings.NewReader(text)
		if e := cmd.Run(); e == nil {
			return
		}
	}
	fmt.Fprintln(os.Stdout, text)
}
func linuxPromptCode() (string, error) {
	if _, e := exec.LookPath("zenity"); e == nil {
		cmd := exec.Command("zenity", "--entry", "--title=Link Sloth Mode account (optional)", "--text=Paste the 21-character device code from your account dashboard:")
		out, e := cmd.Output()
		return strings.TrimSpace(string(out)), e
	}
	fmt.Println("Optional: to link an account, run sloth-mode-connector --link-account YOUR-CODE")
	return "", nil
}
func linuxSetup(base, secret string) {
	msg := "🦥 Sloth Mode — Linux local connector\n\n" +
		"Pair your browser once at https://sloth.marceli.dev/dashboard\n" +
		"Paste this LOCAL pairing key when prompted:\n\n" + secret + "\n\n" +
		"Account sign-in is completely OPTIONAL. Local tasks and Gemma work without it.\n" +
		"To link an account, use the dashboard to generate a device code, then run:\n" +
		"sloth-mode-connector --link-account YOUR-21-CHARACTER-CODE\n\n" +
		"Your notes: " + base + "\n\n" +
		"AI model: optional, downloaded only if you run --install-ai.\n" +
		"Check hardware: --hardware\n" +
		"Local service listens only on 127.0.0.1:39871."
	linuxShow(msg)
}
func linuxNotify(message string) {
	if _, e := exec.LookPath("notify-send"); e == nil {
		cmd := exec.Command("notify-send", "Sloth Mode", message)
		_ = cmd.Run()
	}
}
func linuxClipboard(text string) error {
	for _, command := range [][]string{{"wl-copy"}, {"xclip", "-selection", "clipboard"}, {"xsel", "--clipboard", "--input"}} {
		if _, e := exec.LookPath(command[0]); e != nil {
			continue
		}
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdin = bytes.NewReader([]byte(text))
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("install wl-clipboard, xclip or xsel to copy automatically")
}
