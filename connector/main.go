package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const bridgePort = "39871"
const bridgeOrigin = "https://sloth.marceli.dev"
const bridgeVersion = "0.15.0-linux-beta"

func main() {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("fatal panic recovered: %v\n%s", recovered, debug.Stack())
			startupError("Sloth Mode encountered an error. Notes are safe. See connector.log.")
		}
	}()
	base, err := appDataDir()
	if err != nil {
		startupError(err.Error())
		return
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		startupError(err.Error())
		return
	}
	logFile, err := os.OpenFile(filepath.Join(base, "connector.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(io.MultiWriter(logFile, os.Stderr))
	}
	log.Printf("Sloth Mode %s starting; OS=%s", bridgeVersion, runtime.GOOS)
	store, err := openStore(base)
	if err != nil {
		startupError("Could not open notes: " + err.Error())
		return
	}
	secret, err := getOrCreateSecret(filepath.Join(base, "bridge-secret"))
	if err != nil {
		startupError("Could not make pairing key: " + err.Error())
		return
	}
	if runtime.GOOS == "linux" {
		executed, runErr := linuxCommand(base, secret)
		if executed {
			if runErr != nil {
				startupError(runErr.Error())
				os.Exit(1)
			}
			return
		}
	}
	app := newBridge(base, store, secret)
	ln, err := net.Listen("tcp4", "127.0.0.1:"+bridgePort)
	if err != nil {
		if !showExistingSetup() {
			startupError("Connector could not listen on port 39871. Another version may be running. Exit it from the tray or Task Manager and try again. Details: " + err.Error())
		}
		return
	}
	heartbeatStop := make(chan struct{})
	defer close(heartbeatStop)
	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("account heartbeat panic recovered: %v\\n%s", recovered, debug.Stack())
			}
		}()
		runAccountHeartbeat(base, heartbeatStop)
	}()
	srv := &http.Server{Handler: app.handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 8192, ErrorLog: log.Default()}
	go func() {
		e := srv.Serve(ln)
		if !errors.Is(e, http.ErrServerClosed) {
			log.Printf("bridge: %v", e)
		}
	}()
	if runtime.GOOS == "windows" {
		runTray(app, srv)
	} else if runtime.GOOS == "linux" {
		fmt.Fprintln(os.Stdout, "Sloth Mode Linux connector active on 127.0.0.1:"+bridgePort+". Account optional.")
		linuxServe(srv, heartbeatStop, base)
	} else {
		fmt.Fprintln(os.Stdout, "Sloth Mode connector active on localhost ("+strings.ToLower(runtime.GOOS)+")")
		select {}
	}
}
