package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "plugin-path, p",
			Usage:  "path for interlock plugins",
			Value:  defaultPluginPath,
			EnvVar: "INTERLOCK_PLUGIN_PATH",
		},
	}
	app.Commands = appCommands

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
