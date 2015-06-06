package haproxy

import (
	"time"
)

type (
	Host struct {
		Name             string
		Domain           string
		Check            string
		BackendOptions   []string
		Upstreams        []*Upstream
		SSLOnly          bool
		BalanceAlgorithm string
	}
	Upstream struct {
		Container     string
		Addr          string
		CheckInterval int
		Image         string
		Created       time.Time
	}
	// this is the struct that is used for generation of the proxy config
	ProxyConfig struct {
		Hosts        []*Host
		PluginConfig *PluginConfig
	}
	ByCreatedTime []*Upstream
)

// sort.Interface for ByCreatedTime (aka []Upstream)
func (s ByCreatedTime) Len() int {
	return len(s)
}
func (s ByCreatedTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByCreatedTime) Less(i, j int) bool {
	if s[i] == nil {
		return true
	}
	if s[j] == nil {
		return false
	}
	return s[i].Created.Before(s[j].Created)
}
