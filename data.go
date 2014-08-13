package interlock

type (
	InterlockData struct {
		Port         int      `json:"port,omitempty"`
		AliasDomains []string `json:"alias_domains,omitempty"`
		Warm         bool     `json:"warm,omitempty"`
	}
)
