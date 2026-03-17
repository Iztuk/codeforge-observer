package proxy

import (
	"codeforge-observer/audit"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type ProxyManger struct {
	Hosts  map[string]*ProxyTarget
	Logger *log.Logger
}

type ProxyTarget struct {
	Name     string
	Upstream *url.URL
	Proxy    *httputil.ReverseProxy
	Logger   *log.Logger
}

type contextKey string

const observationKey contextKey = "observation"

func NewProxyHandler(target, hostName string, logger *log.Logger) (*ProxyTarget, error) {
	targetUrl, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(targetUrl)

	h := &ProxyTarget{
		Name:     hostName,
		Upstream: targetUrl,
		Proxy:    rp,
		Logger:   logger,
	}

	originalDirector := rp.Director
	rp.Director = func(r *http.Request) {
		originalDirector(r)

		requestID := getOrCreateRequestID(r)
		r.Header.Set("X-Request-ID", requestID)

		obs := &audit.Observation{
			Timestamp:      time.Now().UTC(),
			Event:          "request_started",
			RequestID:      requestID,
			Host:           r.Host,
			Method:         r.Method,
			Path:           r.URL.Path,
			Query:          r.URL.RawQuery,
			Upstream:       h.Upstream.String(),
			RequestHeaders: cloneHeader(r.Header),
		}

		ctx := context.WithValue(r.Context(), observationKey, obs)
		*r = *r.WithContext(ctx)
		writeObservation(logger, obs)
	}

	rp.ModifyResponse = func(resp *http.Response) error {
		obs, _ := resp.Request.Context().Value(observationKey).(*audit.Observation)
		if obs == nil {
			return nil
		}

		obs.Event = "request_completed"
		obs.Status = resp.StatusCode
		obs.DurationMs = time.Since(obs.Timestamp).Milliseconds()
		obs.ResponseHeaders = cloneHeader(resp.Header)

		writeObservation(logger, obs)
		return nil
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		obs, _ := r.Context().Value(observationKey).(*audit.Observation)
		if obs != nil {
			obs.Error = err.Error()
			obs.DurationMs = time.Since(obs.Timestamp).Milliseconds()
			writeObservation(logger, obs)
		}

		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	return h, nil
}

func (pm *ProxyManger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := normalizeHost(r.Host)

	target, ok := pm.Hosts[host]
	if !ok {
		pm.Logger.Printf("no route found for host=%s rawHost=%s", host, r.Host)
		http.NotFound(w, r)
		return
	}

	pm.Logger.Printf("routing host=%s to upstream=%s", host, target.Upstream.String())
	target.Proxy.ServeHTTP(w, r)
}

func normalizeHost(host string) string {
	if strings.Contains(host, ":") {
		h, _, err := net.SplitHostPort(host)
		if err == nil {
			return strings.ToLower(h)
		}
	}
	return strings.ToLower(host)
}

func getOrCreateRequestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}

	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}

	return hex.EncodeToString(b[:])
}

func cloneHeader(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, v := range h {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

func writeObservation(logger *log.Logger, obs *audit.Observation) {
	b, err := json.Marshal(obs)
	if err != nil {
		logger.Printf(`{"message":"failed to marshal observation","error":%q}`, err.Error())
		return
	}
	logger.Print(string(b))
}
