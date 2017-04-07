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

type ContextRoot struct {
	Name string
	Path string
}

type Host struct {
	ServerNames        []string
	Port               int
	ContextRoot        *ContextRoot
	ContextRootRewrite bool
	SSLPort            int
	SSL                bool
	SSLCert            string
	SSLCertKey         string
	SSLOnly            bool
	SSLBackend         bool
	Upstream           *Upstream
	WebsocketEndpoints []string
	IPHash             bool
	Check              string
    CheckInterval      int
}
type Config struct {
	Hosts    []*Host
	Config   *config.ExtensionConfig
	Networks map[string]string
}
