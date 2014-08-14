package interlock

type (
	InterlockData struct {
		// these are custom vals for upstreams
		Port          int      `json:"port,omitempty"`
		AliasDomains  []string `json:"alias_domains,omitempty"`
		Warm          bool     `json:"warm,omitempty"`
		CheckInterval int      `json:"check_interval,omitempty"`

		// these are custom vals for hosts
		Check          string   `json:"check,omitempty"`
		BackendOptions []string `json:"backend_options,omitempty"`
	}
)
