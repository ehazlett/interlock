package haproxy

type (
	Host struct {
		Name             string
		Domain           string
		Check            string
		BackendOptions   []string
		Upstreams        []*Upstream
		SSLOnly          bool
		BalanceAlgorithm string
		Mode             string
		PublicPort       int
	}
	Upstream struct {
		Addr          string
		CheckInterval int
	}
	// this is the struct that is used for generation of the proxy config
	ProxyConfig struct {
		Hosts        []*Host
		PluginConfig *PluginConfig
	}
)
