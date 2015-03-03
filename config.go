package interlock

type (
	Config struct {
		SwarmUrl         string `json:"swarmUrl,omitempty"`
		PluginConfigPath string `json:"pluginConfigPath,omitempty"`
	}

	InterlockConfig struct {
		Version string `json:"version,omitempty"`
	}
)
