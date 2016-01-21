package config

import (
	"github.com/BurntSushi/toml"
)

// SetConfigDefaults sets default values if not present
func SetConfigDefaults(c *ExtensionConfig) error {
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 5000
	}

	if c.ServerTimeout == 0 {
		c.ServerTimeout = 10000
	}

	if c.ClientTimeout == 0 {
		c.ClientTimeout = 10000
	}

	if c.MaxConn == 0 {
		c.MaxConn = 1024
	}

	if c.Port == 0 {
		c.Port = 80
	}

	if c.AdminUser == "" {
		c.AdminUser = "admin"
	}

	if c.AdminPass == "" {
		c.AdminPass = ""
	}

	if c.User == "" {
		c.User = "www-data"
	}

	if c.WorkerProcesses == 0 {
		c.WorkerProcesses = 2
	}

	if c.RLimitNoFile == 0 {
		c.RLimitNoFile = 65535
	}

	if c.ProxyConnectTimeout == 0 {
		c.ProxyConnectTimeout = 600
	}

	if c.ProxySendTimeout == 0 {
		c.ProxySendTimeout = 600
	}

	if c.ProxyReadTimeout == 0 {
		c.ProxyReadTimeout = 600
	}

	if c.SendTimeout == 0 {
		c.SendTimeout = 600
	}

	if c.SSLCiphers == "" {
		c.SSLCiphers = "HIGH:!aNULL:!MD5"
	}

	if c.SSLProtocols == "" {
		c.SSLProtocols = "SSLv3 TLSv1 TLSv1.1 TLSv1.2"
	}

	return nil
}

// ParseConfig returns a Config object from a raw string config TOML
func ParseConfig(data string) (*Config, error) {
	var cfg Config
	if _, err := toml.Decode(data, &cfg); err != nil {
		return nil, err
	}

	for _, ext := range cfg.Extensions {
		// setup defaults for missing config entries
		if err := SetConfigDefaults(ext); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}
