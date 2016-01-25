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
		ConfigPath:             "/etc/conf/nginx.conf",
		PidPath:                "/etc/conf/nginx.pid",
		BackendOverrideAddress: "",
	}

	if err := config.SetConfigDefaults(ec); err != nil {
		log.Fatal(err)
	}

	cfg := &config.Config{
		ListenAddr: ":8080",
		DockerURL:  "unix:///var/run/docker.sock",
		Extensions: []*config.ExtensionConfig{
			ec,
		},
	}

	if err := toml.NewEncoder(os.Stdout).Encode(cfg); err != nil {
		log.Fatal(err)
	}
}
