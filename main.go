package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

type ProxyHandler struct {
	TargetUrl *url.URL
	Logger    *log.Logger
	Proxy     *httputil.ReverseProxy
}

func NewProxyHandler(target string, logger *log.Logger) (*ProxyHandler, error) {
	targetUrl, err := url.Parse(targetAddr)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(targetUrl)

	h := &ProxyHandler{
		TargetUrl: targetUrl,
		Logger:    logger,
		Proxy:     rp,
	}

	originalDirector := rp.Director
	rp.Director = func(r *http.Request) {
		originalDirector(r)
		logger.Printf(
			"proxy request url=%s method=%s path=%s remote=%s host=%s",
			r.URL,
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.Host,
		)
	}

	rp.ModifyResponse = func(r *http.Response) error {
		h.Logger.Printf(
			"proxy response method=%s path=%s status=%d",
			r.Request.Method,
			r.Request.URL.Path,
			r.StatusCode,
		)
		return nil
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.Logger.Printf(
			"proxy error method=%s path=%s err=%v",
			r.Method,
			r.URL.Path,
			err,
		)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	return h, nil
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	h.Proxy.ServeHTTP(w, r)

	h.Logger.Printf(
		"request complete method=%s path=%s duration=%s",
		r.Method,
		r.URL.Path,
		time.Since(start),
	)
}

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

	handler, err := NewProxyHandler(targetAddr, logger)
	if err != nil {
		logger.Fatalf("failed to creat proxy handler: %v", err)
	}

	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
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
