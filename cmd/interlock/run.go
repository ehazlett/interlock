package main

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock/config"
	"github.com/ehazlett/interlock/server"
)

var cmdRun = cli.Command{
	Name:   "run",
	Action: runAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Path to config file",
			Value: "config.toml",
		},
	},
}

func runAction(c *cli.Context) {
	configPath := c.String("config")

	var data string
	d, err := ioutil.ReadFile(configPath)
	switch {
	case os.IsNotExist(err):
		log.Debug("no config detected; generating local config")
		data = `listenAddr = ":8080"
dockerURL = "unix:///var/run/docker.sock"
`
	case err == nil:
		data = string(d)
	default:
		log.Fatal(err)
	}

	cfg, err := config.ParseConfig(string(data))
	if err != nil {
		log.Fatal(err)
	}

	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}
