package config

// TODO: Rules allow for filtering metrics
// Note: for use with the Beacon extension only
type Rule struct {
	Type  string
	Regex string
}

// ExtensionConfig has all options for all load balancer extensions
// the extension itself will use whichever options needed
type ExtensionConfig struct {
	Name                   string           // extension name
	ConfigPath             string           // config file path
	ConfigBasePath         string           `toml:"-"` // internal
	PidPath                string           // haproxy, nginx
	BackendOverrideAddress string           // haproxy, nginx
	ConnectTimeout         int              // haproxy
	ServerTimeout          int              // haproxy
	ClientTimeout          int              // haproxy
	MaxConn                int              // haproxy, nginx
	Port                   int              // haproxy, nginx
	SyslogAddr             string           // haproxy
	NginxPlusEnabled       bool             // nginx
	AdminUser              string           // haproxy
	AdminPass              string           // haproxy
	SSLCertPath            string           // haproxy, nginx
	SSLCert                string           // haproxy
	SSLPort                int              // haproxy, nginx
	SSLOpts                string           // haproxy
	SSLDefaultDHParam      int              // haproxy
	SSLServerVerify        string           // haproxy
	User                   string           // nginx
	WorkerProcesses        int              // nginx
	RLimitNoFile           int              // nginx
	ProxyConnectTimeout    int              // nginx
	ProxySendTimeout       int              // nginx
	ProxyReadTimeout       int              // nginx
	SendTimeout            int              // nginx
	SSLCiphers             string           // nginx
	SSLProtocols           string           // nginx
	StatInterval           int              // beacon
	Rules                  map[string]*Rule // beacon FIXME: this isn't loaded properly from toml; we set it as a hack now
}

// Config is the top level configuration
type Config struct {
	ListenAddr        string
	DockerURL         string
	TLSCACert         string
	TLSCert           string
	TLSKey            string
	AllowInsecure     bool
	EnableMetrics     bool
	HeartbeatInterval int // interval to check the system health
	Extensions        []*ExtensionConfig
	Rules             map[string]*Rule // beacon TODO: move to ExtensionConfig
}
