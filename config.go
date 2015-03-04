package interlock

type (
	Config struct {
		SwarmUrl       string   `json:"swarm_url,omitempty"`
		EnabledPlugins []string `json:"enabled_plugins,omitempty"`
	}

	InterlockConfig struct {
		Version string `json:"version,omitempty"`
	}
)
