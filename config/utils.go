package config

import (
	"github.com/BurntSushi/toml"
)

// ParseConfig returns a Config object from a raw string config TOML
func ParseConfig(data string) (*Config, error) {
	var cfg Config
	if _, err := toml.Decode(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
