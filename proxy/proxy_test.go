package proxy

import (
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"testing"
)

func makeTarget(t *testing.T, name, upstream string) *ProxyTarget {
	t.Helper()

	u, err := url.Parse(upstream)
	if err != nil {
		t.Fatalf("failed to parse upstream: %v", err)
	}

	return &ProxyTarget{
		Name:     name,
		Upstream: u,
		Proxy:    httputil.NewSingleHostReverseProxy(u),
		Logger:   log.New(os.Stdout, "", 0),
	}
}

func TestProxyManager_HostOperations(t *testing.T) {
	tests := []struct {
		name          string
		operations    func(pm *ProxyManager)
		expectedCount int
		expectedHosts map[string]string // name -> upstream
	}{
		{
			name: "add single host",
			operations: func(pm *ProxyManager) {
				pm.AddHost("api.local", makeTarget(t, "api.local", "http://127.0.0.1:8081"))
			},
			expectedCount: 1,
			expectedHosts: map[string]string{
				"api.local": "http://127.0.0.1:8081",
			},
		},
		{
			name: "overwrite existing host",
			operations: func(pm *ProxyManager) {
				pm.AddHost("api.local", makeTarget(t, "api.local", "http://127.0.0.1:8081"))
				pm.AddHost("api.local", makeTarget(t, "api.local", "http://127.0.0.1:9090"))
			},
			expectedCount: 1,
			expectedHosts: map[string]string{
				"api.local": "http://127.0.0.1:9090",
			},
		},
		{
			name: "add multiple hosts",
			operations: func(pm *ProxyManager) {
				pm.AddHost("api.local", makeTarget(t, "api.local", "http://127.0.0.1:8081"))
				pm.AddHost("auth.local", makeTarget(t, "auth.local", "http://127.0.0.1:8082"))
			},
			expectedCount: 2,
			expectedHosts: map[string]string{
				"api.local":  "http://127.0.0.1:8081",
				"auth.local": "http://127.0.0.1:8082",
			},
		},
		{
			name: "remove existing host",
			operations: func(pm *ProxyManager) {
				pm.AddHost("api.local", makeTarget(t, "api.local", "http://127.0.0.1:8081"))
				pm.RemoveHost("api.local")
			},
			expectedCount: 0,
			expectedHosts: map[string]string{},
		},
		{
			name: "remove non-existent host does nothing",
			operations: func(pm *ProxyManager) {
				pm.RemoveHost("does-not-exist")
			},
			expectedCount: 0,
			expectedHosts: map[string]string{},
		},
		{
			name: "add then remove one of many",
			operations: func(pm *ProxyManager) {
				pm.AddHost("api.local", makeTarget(t, "api.local", "http://127.0.0.1:8081"))
				pm.AddHost("auth.local", makeTarget(t, "auth.local", "http://127.0.0.1:8082"))
				pm.RemoveHost("api.local")
			},
			expectedCount: 1,
			expectedHosts: map[string]string{
				"auth.local": "http://127.0.0.1:8082",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pm := &ProxyManager{
				Hosts:  make(map[string]*ProxyTarget),
				Logger: log.New(os.Stdout, "", 0),
			}

			tc.operations(pm)

			hosts := pm.ListHosts()

			if len(hosts) != tc.expectedCount {
				t.Fatalf("expected %d hosts, got %d", tc.expectedCount, len(hosts))
			}

			// Convert slice → map for easier comparison
			got := make(map[string]string)
			for _, h := range hosts {
				got[h.Name] = h.Upstream
			}

			for name, upstream := range tc.expectedHosts {
				if got[name] != upstream {
					t.Fatalf("expected host %s -> %s, got %s", name, upstream, got[name])
				}
			}
		})
	}
}
