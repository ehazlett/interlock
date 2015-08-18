package haproxy

type (
	Host struct {
		Name                string
		Domain              string
		Check               string
		BackendOptions      []string
		Upstreams           []*Upstream
		SSLOnly             bool
		SSLBackend          bool
		SSLBackendTLSVerify string
		BalanceAlgorithm    string
	}
	Upstream struct {
		Container     string
		Addr          string
		CheckInterval int
	}
	// this is the struct that is used for generation of the proxy config
	ProxyConfig struct {
		Hosts        []*Host
		PluginConfig *PluginConfig
	}
)
