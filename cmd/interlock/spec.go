package main

import (
	"os"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock/config"
)

var cmdSpec = cli.Command{
	Name:   "spec",
	Usage:  "generate a configuration file",
	Action: specAction,
}

func specAction(c *cli.Context) {
	ec := &config.ExtensionConfig{
		Name:                   "nginx",
		ConfigPath:             "/etc/nginx/nginx.conf",
		PidPath:                "/var/run/nginx.pid",
		TemplatePath:           "/etc/interlock/nginx.conf.template",
		BackendOverrideAddress: "",
	}

	config.SetConfigDefaults(ec)

	cfg := &config.Config{
		ListenAddr:    ":8080",
		DockerURL:     "unix:///var/run/docker.sock",
		EnableMetrics: true,
		Extensions: []*config.ExtensionConfig{
			ec,
		},
	}

	if err := toml.NewEncoder(os.Stdout).Encode(cfg); err != nil {
		log.Fatal(err)
	}
}
