package nginx

type InterlockData struct {
	// these are custom vals for upstreams
	Port             int      `json:"port,omitempty"`
	AliasDomains     []string `json:"alias_domains,omitempty"`
	SSLOnly          bool     `json:"ssl_only,omitempty"`
	Hostname         string   `json:"hostname,omitempty"`
	Domain           string   `json:"domain,omitempty"`
	BalanceAlgorithm string   `json:"balance_algorithm,omitempty"`
}
