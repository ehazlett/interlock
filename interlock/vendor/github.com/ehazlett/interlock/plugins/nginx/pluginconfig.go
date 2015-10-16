package nginx

type PluginConfig struct {
	ProxyConfigPath             string `json:"proxy_config_path,omitempty"`
	ProxyBackendOverrideAddress string `json:"proxy_backend_override_address,omitempty"`
	ProxyConnectTimeout         int    `json:"proxy_connect_timeout,omitempty"`
	ProxySendTimeout            int    `json:"proxy_send_timeout,omitempty"`
	ProxyReadTimeout            int    `json:"proxy_read_timeout,omitempty"`
	SendTimeout                 int    `json:"send_timeout,omitempty"`
	MaxConnections              int    `json:"max_connections,omitempty"`
	MaxProcesses                int    `json:"max_processes,omitempty"`
	RLimitNoFile                int    `json:"rlimit_no_file,omitempty"`
	Port                        int    `json:"port,omitempty"`
	PidPath                     string `json:"pid_path,omitempty"`
	SSLCertDir                  string `json:"ssl_cert_dir,omitempty"`
	SSLPort                     int    `json:"ssl_port,omitempty"`
	SSLCiphers                  string `json:"ssl_ciphers,omitempty"`
	SSLProtocols                string `json:"ssl_protocols,omitempty"`
	User                        string `json:"user,omitempty"`
}
