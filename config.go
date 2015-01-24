package interlock

type (
	Config struct {
		SwarmUrl        string `json:"swarm_url,omitempty"`
		ProxyConfigPath string `json:"proxy_config_path,omitempty"`
		Port            int    `json:"port,omitempty"`
		PidPath         string `json:"pid_path,omitempty"`
		ConnectTimeout  int    `json:"connect_timeout,omitempty"`
		ServerTimeout   int    `json:"server_timeout,omitempty"`
		ClientTimeout   int    `json:"client_timeout,omitempty"`
		MaxConn         int    `json:"max_conn,omitempty"`
		SyslogAddr      string `json:"syslog_addr,omitempty"`
		StatsUser       string `json:"stats_user,omitempty"`
		StatsPassword   string `json:"stats_password,omitempty"`
		SSLCert         string `json:"ssl_cert,omitempty"`
		SSLPort         int    `json:"ssl_port,omitempty"`
		SSLOpts         string `json:"ssl_opts,omitempty"`
	}
)
