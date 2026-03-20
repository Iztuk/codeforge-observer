package proxy

import (
	"codeforge-observer/config"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

type ControlCommand struct {
	Action   string `json:"action"`
	Host     string `json:"host,omitempty"`
	Upstream string `json:"upstream,omitempty"`
	Contract string `json:"contract,omitempty"`
}

type ControlResponse struct {
	OK      bool       `json:"ok"`
	Message string     `json:"message,omitempty"`
	Error   string     `json:"error,omitempty"`
	Hosts   []HostInfo `json:"hosts,omitempty"`
}

type HostInfo struct {
	Name     string `json:"name"`
	Upstream string `json:"upstream"`
	// Contract string `json:"contract"`
}

func AddHostCommand(host, upstream, contractFile string) error {
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

func (pm *ProxyManager) ListHosts() []HostInfo {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	hosts := make([]HostInfo, 0, len(pm.Hosts))
	for _, host := range pm.Hosts {
		hosts = append(hosts, HostInfo{
			Name:     host.Name,
			Upstream: host.Upstream.String(),
		})
	}

	return hosts
}
