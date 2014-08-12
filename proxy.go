package interlock

type (
	Host struct {
		Name      string
		Domain    string
		Upstreams []*Upstream
	}
	Upstream struct {
		Addr string
	}
	// this is the struct that is used for generation of the proxy config
	ProxyConfig struct {
		Hosts          []*Host
		Path           string
		PidPath        string
		SyslogAddr     string
		MaxConn        int
		Port           int
		ConnectTimeout int
		ServerTimeout  int
		ClientTimeout  int
	}
)
