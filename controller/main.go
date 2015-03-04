package main

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/homedir"
	"github.com/ehazlett/interlock/plugins"
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
	app.Version = VERSION
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
			Name:   "swarm-url, s",
			Value:  "unix:///var/run/docker.sock",
			Usage:  "swarm addr",
			EnvVar: "DOCKER_HOST",
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
		cli.StringFlag{
			Name:   "plugin-config-path, p",
			Value:  filepath.Join(homedir.Get(), ".interlock"),
			Usage:  "path for plugin specific config files",
			EnvVar: "INTERLOCK_PLUGIN_CONFIG_PATH",
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
	}
	// plugin supplied commands
	baseCommands = append(baseCommands, plugins.GetCommands()...)

	app.Commands = baseCommands

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
