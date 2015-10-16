package nginx

type NginxConfig struct {
	PluginConfig
	Hosts []*Host
}
