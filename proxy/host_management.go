package proxy

import (
	"codeforge-observer/audit"
	"codeforge-observer/config"
	"codeforge-observer/storage"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type ControlCommand struct {
	Action   string `json:"action"`
	Host     string `json:"host,omitempty"`
	Upstream string `json:"upstream,omitempty"`
	Contract string `json:"contract,omitempty"`
	Resource string `json:"resource,omitempty"`
}

type ControlResponse struct {
	OK      bool               `json:"ok"`
	Message string             `json:"message,omitempty"`
	Error   string             `json:"error,omitempty"`
	Hosts   []storage.HostInfo `json:"hosts,omitempty"`
}

func AddHostCommand(host, upstream, contractFile, resourceFile string) error {
	conn, err := net.Dial("unix", config.SockFile)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	cmd := ControlCommand{
		Action:   "add_host",
		Host:     host,
		Upstream: upstream,
		Contract: contractFile,
		Resource: resourceFile,
	}

	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	var resp ControlResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}

	fmt.Println(resp.Message)
	return nil
}

func RemoveHostCommand(host string) error {
	conn, err := net.Dial("unix", config.SockFile)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	cmd := ControlCommand{
		Action: "remove_host",
		Host:   host,
	}

	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	var resp ControlResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}

	fmt.Println(resp.Message)
	return nil
}

func ListHostsCommand() error {
	conn, err := net.Dial("unix", config.SockFile)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	cmd := ControlCommand{
		Action: "list_hosts",
	}

	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	var resp ControlResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}

	fmt.Printf("%-20s %-30s\n", "HOST", "UPSTREAM")
	fmt.Println(strings.Repeat("-", 50))

	if len(resp.Hosts) == 0 {
		fmt.Println("no hosts configured")
	} else {
		for _, h := range resp.Hosts {
			fmt.Printf("%-20s %-30s\n", h.Name, h.Upstream)
		}
	}

	return nil
}

func (pm *ProxyManager) AddHost(host string, target *ProxyTarget) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	pm.Hosts[host] = target
}

func (pm *ProxyManager) RemoveHost(host string) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	delete(pm.Hosts, host)
}

func (pm *ProxyManager) ListHosts() []storage.HostInfo {
	pm.Mu.RLock()
	defer pm.Mu.RUnlock()
	hosts := make([]storage.HostInfo, 0, len(pm.Hosts))
	for _, host := range pm.Hosts {
		hosts = append(hosts, storage.HostInfo{
			Name:     host.Name,
			Upstream: host.Upstream.String(),
		})
	}

	return hosts
}

func (pm *ProxyManager) BootstrapHosts(db *sql.DB) error {
	hosts, err := storage.ReadHosts(db)
	if err != nil {
		return err
	}

	for _, host := range hosts {
		var target ProxyTarget = ProxyTarget{
			Name:   host.Name,
			Logger: pm.Logger,
		}

		u, err := url.Parse(host.Upstream)
		if err != nil {
			return err
		}
		target.Upstream = u

		apiDoc, err := audit.ReadOpenApiDoc(host.Contract)
		if err != nil {
			return err
		}
		target.Contracts = &apiDoc

		resourceDoc, err := audit.ReadResourceDoc(host.Resource)
		if err != nil {
			return err
		}
		target.Resource = &resourceDoc

		pm.AddHost(host.Name, &target)
	}

	return nil
}
