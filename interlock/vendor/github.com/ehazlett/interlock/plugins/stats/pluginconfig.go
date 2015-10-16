package stats

import (
	"regexp"
)

type PluginConfig struct {
	CarbonAddress   string
	StatsPrefix     string
	ImageNameFilter *regexp.Regexp
	Interval        int
}
