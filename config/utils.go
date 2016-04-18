package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

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

		// FIXME: toml isn't being parse right so we hack the rules in like so
		ext.Rules = cfg.Rules
	}

	return &cfg, nil
}

// SetConfigDefaults sets default values if not present
// ExtensionConfig.Name must be set before calling this function
func SetConfigDefaults(c *ExtensionConfig) error {
	if c.MaxConn == 0 {
		c.MaxConn = 1024
	}

	if c.Port == 0 {
		c.Port = 80
	}

	switch c.Name {
	case "haproxy":
		SetHAProxyConfigDefaults(c)
	case "nginx":
		SetNginxConfigDefaults(c)
	default:
		return fmt.Errorf("unknown load balancer backend: %s", c.Name)
	}

	return nil
}

func SetHAProxyConfigDefaults(c *ExtensionConfig) {
	if c.TemplatePath == "" {
		c.TemplatePath = "/etc/interlock/haproxy.cfg.template"
	}

	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 5000
	}

	if c.ServerTimeout == 0 {
		c.ServerTimeout = 10000
	}

	if c.ClientTimeout == 0 {
		c.ClientTimeout = 10000
	}

	if c.AdminUser == "" {
		c.AdminUser = "admin"
	}

	if c.AdminPass == "" {
		c.AdminPass = ""
	}

	if c.SSLDefaultDHParam == 0 {
		c.SSLDefaultDHParam = 1024
	}

	if c.SSLServerVerify == "" {
		c.SSLServerVerify = "required"
	}
}

func SetNginxConfigDefaults(c *ExtensionConfig) {
	if c.TemplatePath == "" {
		c.TemplatePath = "/etc/interlock/nginx.conf.template"
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
}
