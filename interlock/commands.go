package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/manager"
)

var appCommands = []cli.Command{
	{
		Name:   "ls",
		Usage:  "list available plugins",
		Action: cmdListPlugins,
	},
	{
		Name:   "start",
		Usage:  "start interlock",
		Action: cmdStart,
	},
}

func waitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	for _ = range sigChan {
		os.Exit(0)
	}
}

func getManager(c *cli.Context) (*manager.Manager, error) {
	dockerUrl := c.GlobalString("docker")
	tlsCaCert := c.GlobalString("tls-ca-cert")
	tlsCert := c.GlobalString("tls-cert")
	tlsKey := c.GlobalString("tls-key")
	allowInsecure := c.GlobalBool("allow-insecure")
	pluginPath := c.GlobalString("plugin-path")

	mgr, err := manager.NewManager(dockerUrl, tlsCaCert, tlsCert, tlsKey, allowInsecure, pluginPath)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func cmdListPlugins(c *cli.Context) {
	mgr, err := getManager(c)
	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tAUTHOR\tURL")

	plugins, err := mgr.Plugins()
	if err != nil {
		log.Fatal(err)
	}

	for _, plugin := range plugins {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", plugin.Name,
			plugin.Version, plugin.Author, plugin.Url)
	}

	w.Flush()
}

func cmdStart(c *cli.Context) {
	log.Infof("interlock %s started", interlock.VERSION)
	log.Infof("listening for events: url=%s", c.GlobalString("docker"))

	mgr, err := getManager(c)
	if err != nil {
		log.Fatal(err)
	}

	mgr.Run()
	waitForInterrupt()
}
