package proxy

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
