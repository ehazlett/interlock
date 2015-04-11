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
	app.Version = version.FULL_VERSION
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
	}
	// plugin supplied commands
	baseCommands = append(baseCommands, plugins.GetCommands()...)

	app.Commands = baseCommands

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
