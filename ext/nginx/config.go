package nginx

import (
	"github.com/ehazlett/interlock/config"
)

type Server struct {
	Addr string
}

type Upstream struct {
	Name    string
	Servers []*Server
}
type Host struct {
	ServerNames        []string
	Port               int
	SSLPort            int
	SSL                bool
	SSLCert            string
	SSLCertKey         string
	SSLOnly            bool
	SSLBackend         bool
	Upstream           *Upstream
	WebsocketEndpoints []string
}
type Config struct {
	Hosts  []*Host
	Config *config.ExtensionConfig
}
