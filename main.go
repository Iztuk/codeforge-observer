package main

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
	listenAddr = ":8080"
	targetAddr = "http://localhost:8081"
)

func main() {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	logger := log.New(f, "cf-observer: ", log.LstdFlags)

	if err := ensureSingleInstance(); err != nil {
		logger.Fatalf("startup failed: %v", err)
	}

	pid := os.Getpid()
	if err := os.WriteFile(pidFile, fmt.Appendf([]byte{}, "%d", pid), 0644); err != nil {
		logger.Fatalf("failed to write pid file: %v", err)
	}
	defer os.Remove(pidFile)

	logger.Printf("daemon started with pid=%d", pid)

	apiProxy, err := proxy.NewProxyHandler("http://localhost:8081", "api.local", logger)
	if err != nil {
		log.Fatal(err)
	}

	authProxy, err := proxy.NewProxyHandler("http://localhost:8082", "auth.local", logger)
	if err != nil {
		log.Fatal(err)
	}

	pm := &proxy.ProxyManger{
		Hosts: map[string]*proxy.ProxyTarget{
			"api.local":  apiProxy,
			"auth.local": authProxy,
		},
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
