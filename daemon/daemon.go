package daemon

import (
	"codeforge-observer/proxy"
	"context"
	"errors"
	"fmt"
	"log"
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

	server := &http.Server{
		Addr:    listenAddr,
		Handler: pm,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("proxy listening on %s and forwarding to %s", listenAddr, targetAddr)
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
