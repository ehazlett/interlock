package haproxy

type InterlockData struct {
	// these are custom vals for upstreams
	Port             int      `json:"port,omitempty"`
	AliasDomains     []string `json:"alias_domains,omitempty"`
	Warm             bool     `json:"warm,omitempty"`
	SSLOnly          bool     `json:"ssl_only,omitempty"`
	CheckInterval    int      `json:"check_interval,omitempty"`
	Hostname         string   `json:"hostname,omitempty"`
	Domain           string   `json:"domain,omitempty"`
	BalanceAlgorithm string   `json:"balance_algorithm,omitempty"`

	// these are custom vals for hosts
	Check          string   `json:"check,omitempty"`
	BackendOptions []string `json:"backend_options,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	PublicPort     int      `json:"public_port,omitempty"`
}
