package proxy

type HostControlCommand struct {
	Action   string `json:"action"`
	Host     string `json:"host"`
	Upstream string `json:"upstream"`
	Contract string `json:"contract"`
}

func AddHostCommand(host, upstream, contractFile string) error {
	return nil
}

func RemoveHostCommand(host string) error {
	return nil
}

func ListHostsCommand() error {
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

func (pm *ProxyManager) GetHost(host string) (*ProxyTarget, bool) {
	pm.Mu.Lock()
	defer pm.Mu.Unlock()
	target, ok := pm.Hosts[host]
	return target, ok
}
