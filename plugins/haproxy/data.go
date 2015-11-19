package haproxy

type InterlockData struct {
	// these are custom vals for upstreams
	Port                int      `json:"port,omitempty"`
	AliasDomains        []string `json:"alias_domains,omitempty"`
	SSLOnly             bool     `json:"ssl_only,omitempty"`
	SSLBackend          bool     `json:"ssl_backend,omitempty"`
	SSLBackendTLSVerify string   `json:"ssl_backend_tls_verify,omitempty"`
	CheckInterval       int      `json:"check_interval,omitempty"`
	Hostname            string   `json:"hostname,omitempty"`
	Domain              string   `json:"domain,omitempty"`
	BalanceAlgorithm    string   `json:"balance_algorithm,omitempty"`
	ListenOnTcpPort     bool     `json:"listen_on_tcp_port,omitempty"`

	// these are custom vals for hosts
	Check          string   `json:"check,omitempty"`
	BackendOptions []string `json:"backend_options,omitempty"`
}
