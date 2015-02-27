package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/homedir"
	"github.com/ehazlett/interlock"
)

var (
	defaultPluginPath = filepath.Join(homedir.Get(), ".interlock", "plugins")
)

func main() {
	app := cli.NewApp()
	app.Name = "interlock"
	app.Usage = "event driven docker plugins"
	app.Version = interlock.VERSION
	app.Email = "github.com/ehazlett/interlock"
	app.Author = "@ehazlett"
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.DebugLevel)
			os.Setenv("DEBUG", "1")
		}

		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "plugin-path, p",
			Usage:  "path for interlock plugins",
			Value:  defaultPluginPath,
			EnvVar: "INTERLOCK_PLUGIN_PATH",
		},
		cli.StringFlag{
			Name:   "docker, d",
			Usage:  "url to docker",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},
		cli.StringFlag{
			Name:  "tls-ca-cert",
			Usage: "TLS CA certificate for Docker",
			Value: "",
		},
		cli.StringFlag{
			Name:  "tls-cert",
			Usage: "TLS certificate for Docker",
			Value: "",
		},
		cli.StringFlag{
			Name:  "tls-key",
			Usage: "TLS key for Docker",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "allow-insecure",
			Usage: "allow insecure tls for Docker",
		},
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "enable debug logging",
		},
	}
	app.Commands = appCommands

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
