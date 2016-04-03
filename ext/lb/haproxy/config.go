package haproxy

import (
	"github.com/ehazlett/interlock/config"
)

type ContextRoot struct {
	Name string
	Path string
}

type Host struct {
	Name                string
	ContextRoot         *ContextRoot
	ContextRootRewrite  bool
	Domain              string
	Check               string
	BackendOptions      []string
	Upstreams           []*Upstream
	SSLOnly             bool
	SSLBackend          bool
	SSLBackendTLSVerify string
	BalanceAlgorithm    string
}

type Upstream struct {
	Container     string
	Addr          string
	CheckInterval int
}

type Config struct {
	Hosts    []*Host
	Config   *config.ExtensionConfig
	Networks map[string]string
}
