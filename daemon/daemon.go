package daemon

import (
	"codeforge-observer/audit"
	"codeforge-observer/proxy"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	pidFile    = "/tmp/cf-observer.pid"
	logFile    = "/tmp/cf-observer.log"
	sockFile   = "/tmp/cf-observer.sock"
	listenAddr = ":8080"
	targetAddr = "http://localhost:8081"
)

func RunDaemon() error {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	logger := log.New(f, "cf-observer: ", log.LstdFlags)

	if err := ensureSingleInstance(); err != nil {
		return fmt.Errorf("startup failed: %w", err)
	}

	pid := os.Getpid()
	if err := os.WriteFile(pidFile, fmt.Appendf([]byte{}, "%d", pid), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}
	defer os.Remove(pidFile)

	logger.Printf("daemon started with pid=%d", pid)

	pm := &proxy.ProxyManager{
		Hosts:  make(map[string]*proxy.ProxyTarget),
		Logger: logger,
	}

	// Remove stale socket file before binding
	if _, err := os.Stat(sockFile); err == nil {
		_ = os.Remove(sockFile)
	}

	controlLn, err := net.Listen("unix", sockFile)
	if err != nil {
		return fmt.Errorf("failed to listen on socket %s: %w", sockFile, err)
	}
	defer func() {
		_ = controlLn.Close()
		_ = os.Remove(sockFile)
	}()

	go acceptControlConnections(controlLn, pm)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: pm,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("proxy listening on %s", listenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Println("shutdown signal received, shutting down proxy")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
	} else {
		logger.Println("proxy stopped cleanly")
	}
	return nil
}

func ensureSingleInstance() error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid pid file contents: %w", err)
	}

	// Signal 0 checks whether process exists without killing it.
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}

	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return fmt.Errorf("daemon already running with pid=%d", pid)
	}

	return nil
}

func acceptControlConnections(ln net.Listener, pm *proxy.ProxyManager) {
	pm.Logger.Printf("control plane listening on %s", sockFile)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// Listener likely closed during shutdown.
			pm.Logger.Printf("control plane accept error: %v", err)
			return
		}

		go handleControlConn(conn, pm)
	}
}

func handleControlConn(conn net.Conn, pm *proxy.ProxyManager) {
	defer conn.Close()

	var cmd proxy.ControlCommand
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
			OK:    false,
			Error: fmt.Sprintf("invalid command: %v", err),
		})
		return
	}

	pm.Logger.Printf("control command received: action=%s host=%s upstream=%s contractFile=%s", cmd.Action, cmd.Host, cmd.Upstream, cmd.Contract)

	switch cmd.Action {
	case "add_host":
		contract, err := audit.ReadOpenApiDoc(cmd.Contract)
		if err != nil {
			_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		target, err := proxy.NewProxyHandler(cmd.Upstream, cmd.Host, pm.Logger, contract)
		if err != nil {
			_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
				OK:    false,
				Error: err.Error(),
			})
			return
		}

		pm.AddHost(cmd.Host, target)

		_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
			OK:      true,
			Message: "host added",
		})
	case "remove_host":
		pm.RemoveHost(cmd.Host)
		_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
			OK:      true,
			Message: "host removed",
		})

	case "list_hosts":
		hosts := pm.ListHosts()

		_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
			OK:    true,
			Hosts: hosts,
		})

	default:
		_ = json.NewEncoder(conn).Encode(proxy.ControlResponse{
			OK:    false,
			Error: fmt.Sprintf("unknown action: %s", cmd.Action),
		})
	}
}

func StopDaemon() error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("daemon is not running")
		}
		return err
	}

	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid pid file contents: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	fmt.Printf("send SIGTERM to daemon pid=%d\n", pid)
	return nil
}
