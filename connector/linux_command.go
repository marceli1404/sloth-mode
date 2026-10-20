//go:build linux

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func linuxCommand(base, secret string) (bool, error) {
	args := os.Args[1:]
	if len(args) == 0 {
		return false, nil
	}
	p := profileHardware()
	switch args[0] {
	case "--help", "-h":
		fmt.Println("Sloth Mode Linux connector v" + bridgeVersion)
		fmt.Println("Run without arguments for the background connector; no account required.")
		fmt.Println("--setup            Show the browser pairing key and instructions")
		fmt.Println("--copy-key         Copy local browser key to clipboard")
		fmt.Println("--hardware         Show detected components and recommended Gemma quantization")
		fmt.Println("--install-ai       Download the verified model and llama.cpp CPU runtime")
		fmt.Println("--install-ai-vulkan Download verified Vulkan runtime and model (requires drivers)")
		fmt.Println("--link-account CODE Optional device link to your cloud account")
		fmt.Println("--status           Query local service status")
		return true, nil
	case "--setup":
		linuxSetup(base, secret)
		return true, nil
	case "--copy-key":
		return true, linuxClipboard(secret)
	case "--hardware":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return true, enc.Encode(p)
	case "--install-ai", "--install-ai-vulkan":
		backend := "cpu"
		if args[0] == "--install-ai-vulkan" {
			backend = "vulkan"
		}
		ctx, cancel := context.WithTimeout(context.Background(), 75*time.Minute)
		defer cancel()
		return true, installAI(ctx, base, backend, func(s string) { fmt.Fprintln(os.Stdout, s) })
	case "--link-account":
		code := ""
		if len(args) >= 2 {
			code = args[1]
		} else {
			var e error
			code, e = linuxPromptCode()
			if e != nil {
				return true, e
			}
		}
		if code == "" {
			return true, errors.New("no account code provided; guest mode still works")
		}
		host, _ := os.Hostname()
		if host == "" {
			host = "Linux computer"
		}
		if len(host) > 60 {
			host = host[:60]
		}
		if len(host) < 2 {
			host = "Linux PC"
		}
		err := linkAccount(base, code, host)
		if err == nil {
			fmt.Println("Account linked. Your local tasks remain on this computer unless you explicitly save them to your account.")
			linuxNotify("Account linked successfully")
		}
		return true, err
	case "--status":
		request, _ := http.NewRequest("GET", "http://127.0.0.1:"+bridgePort+"/v1/health", nil)
		request.Host = "127.0.0.1:" + bridgePort
		request.Header.Set("Origin", bridgeOrigin)
		client := &http.Client{Timeout: 3 * time.Second}
		resp, e := client.Do(request)
		if e != nil {
			return true, fmt.Errorf("connector not running: %w", e)
		}
		defer resp.Body.Close()
		return true, json.NewEncoder(os.Stdout).Encode(map[string]any{"status": resp.StatusCode, "reachable": resp.StatusCode == 200})
	default:
		return true, fmt.Errorf("unknown argument: %s (try --help)", strings.Join(args, " "))
	}
}
func linuxServe(server *http.Server, stop chan struct{}, base string) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	_ = stop
	fmt.Fprintf(os.Stderr, "Sloth Mode on %s stopped cleanly. Data preserved at %s\n", runtime.GOOS, filepath.Clean(base))
}
