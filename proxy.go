package interlock

type (
	Host struct {
		Name      string
		Domain    string
		Check     string
		Upstreams []*Upstream
	}
	Upstream struct {
		Addr string
	}
	// this is the struct that is used for generation of the proxy config
	ProxyConfig struct {
		Hosts  []*Host
		Config *Config
	}
)
