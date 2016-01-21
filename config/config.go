package config

// the extension config has all options for all load balancer extensions
// the extension itself will use whichever options needed
type ExtensionConfig struct {
	Name                   string
	ConfigPath             string
	PidPath                string
	BackendOverrideAddress string
	ConnectTimeout         int
	ServerTimeout          int
	ClientTimeout          int
	MaxConn                int
	Port                   int
	SyslogAddr             string
	AdminUser              string
	AdminPass              string
	SSLCert                string
	SSLPort                int
	SSLOpts                string
}

type Config struct {
	ListenAddr    string
	DockerURL     string
	TLSCACert     string
	TLSCert       string
	TLSKey        string
	AllowInsecure bool
	Extensions    []ExtensionConfig
}
