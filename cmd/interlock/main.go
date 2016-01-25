package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "interlock"
	app.Version = version.FullVersion()
	app.Author = "@ehazlett"
	app.Email = "ejhazlett@gmail.com"
	app.Usage = "an event driven load balancer for docker"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug logging",
		},
	}
	app.Commands = []cli.Command{
		cmdSpec,
		cmdRun,
	}
	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
