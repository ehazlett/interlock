package interlock

import (
	"github.com/citadel/citadel"
)

type (
	InterlockEngine struct {
		Engine         *citadel.Engine `json:"engine,omitempty"`
		SSLCertificate string          `json:"ssl_cert,omitempty"`
		SSLKey         string          `json:"ssl_key,omitempty"`
		CACertificate  string          `json:"ca_cert,omitempty"`
	}

	Config struct {
		ProxyConfigPath  string             `json:"proxy_config_path,omitempty"`
		Port             int                `json:"port,omitempty"`
		ConnectTimeout   int                `json:"connect_timeout,omitempty"`
		ServerTimeout    int                `json:"server_timeout,omitempty"`
		ClientTimeout    int                `json:"client_timeout,omitempty"`
		MaxConn          int                `json:"max_conn,omitempty"`
		SyslogAddr       string             `json:"syslog_addr,omitempty"`
		InterlockEngines []*InterlockEngine `json:"engines,omitempty"`
	}
)
