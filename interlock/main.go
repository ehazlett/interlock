/*
NAME:
   interlock - event driven docker plugins

USAGE:
   interlock [global options] command [command options] [arguments...]

VERSION:
   0.2.9 (012be26)

AUTHOR:
  @ehazlett

COMMANDS:
   start
   list-plugins
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --swarm-url, -s "unix:///var/run/docker.sock"	swarm addr
   --swarm-tls-ca-cert 					tls ca certificate
   --swarm-tls-cert 					tls certificate
   --swarm-tls-key 					tls key
   --swarm-allow-insecure				enable insecure tls communication
   --plugin, -p [--plugin option --plugin option]	enable plugin
   --debug, -D						enable debug
   --help, -h						show help
   --version, -v					print the version
*/
package main

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock/plugins"
	"github.com/ehazlett/interlock/version"
)

func getTLSConfig(caCert, cert, key []byte, allowInsecure bool) (*tls.Config, error) {
	// TLS config
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	certPool := x509.NewCertPool()

	certPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = certPool
	keypair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{keypair}
	if allowInsecure {
		tlsConfig.InsecureSkipVerify = true
	}

	return &tlsConfig, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "interlock"
	app.Version = version.FullVersion()
	app.Author = "@ehazlett"
	app.Email = ""
	app.Usage = "event driven docker plugins"
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "swarm-url, s",
			Value: "unix:///var/run/docker.sock",
			Usage: "swarm addr",
		},
		cli.StringFlag{
			Name:  "swarm-tls-ca-cert",
			Value: "",
			Usage: "tls ca certificate",
		},
		cli.StringFlag{
			Name:  "swarm-tls-cert",
			Value: "",
			Usage: "tls certificate",
		},
		cli.StringFlag{
			Name:  "swarm-tls-key",
			Value: "",
			Usage: "tls key",
		},
		cli.BoolFlag{
			Name:  "swarm-allow-insecure",
			Usage: "enable insecure tls communication",
		},
		cli.StringSliceFlag{
			Name:  "plugin, p",
			Usage: "enable plugin",
			Value: &cli.StringSlice{},
		},
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "enable debug",
		},
	}
	// base commands
	baseCommands := []cli.Command{
		{
			Name:   "start",
			Action: cmdStart,
		},
		{
			Name:   "list-plugins",
			Action: cmdListPlugins,
		},
		{
			Name:   "info",
			Action: cmdInfo,
		},
	}
	// plugin supplied commands
	baseCommands = append(baseCommands, plugins.GetCommands()...)

	app.Commands = baseCommands

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
