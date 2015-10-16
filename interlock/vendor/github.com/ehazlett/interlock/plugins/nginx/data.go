package nginx

type InterlockData struct {
	// these are custom vals for upstreams
	Port               int      `json:"port,omitempty"`
	AliasDomains       []string `json:"alias_domains,omitempty"`
	SSL                bool     `json:"ssl,omitempty"`
	SSLCert            string   `json:"ssl_certificate,omitempty"`
	SSLCertKey         string   `json:"ssl_certificate_key,omitempty"`
	SSLOnly            bool     `json:"ssl_only,omitempty"`
	SSLBackend         bool     `json:"ssl_backend,omitempty"`
	Hostname           string   `json:"hostname,omitempty"`
	Domain             string   `json:"domain,omitempty"`
	BalanceAlgorithm   string   `json:"balance_algorithm,omitempty"`
	WebsocketEndpoints []string `json:"websocket_endpoints,omitempty"`
}
