package main

import (
	"crypto/tls"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock"
)

func cmdStart(c *cli.Context) {
	swarmUrl := c.GlobalString("swarm-url")
	swarmTlsCaCert := c.GlobalString("swarm-tls-ca-cert")
	swarmTlsCert := c.GlobalString("swarm-tls-cert")
	swarmTlsKey := c.GlobalString("swarm-tls-key")
	allowInsecureTls := c.GlobalBool("swarm-allow-insecure")

	// check environment for docker client config
	envDockerHost := os.Getenv("DOCKER_HOST")
	if envDockerHost != "" {
		swarmUrl = envDockerHost
	}

	envDockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	envDockerTlsVerify := os.Getenv("DOCKER_TLS_VERIFY")
	if envDockerCertPath != "" && envDockerTlsVerify != "" {
		swarmTlsCaCert = filepath.Join(envDockerCertPath, "ca.pem")
		swarmTlsCert = filepath.Join(envDockerCertPath, "cert.pem")
		swarmTlsKey = filepath.Join(envDockerCertPath, "key.pem")
	}

	config := &interlock.Config{}
	config.SwarmUrl = swarmUrl

	// load tlsconfig
	var tlsConfig *tls.Config
	if swarmTlsCaCert != "" && swarmTlsCert != "" && swarmTlsKey != "" {
		log.Infof("using tls for communication with swarm")
		caCert, err := ioutil.ReadFile(swarmTlsCaCert)
		if err != nil {
			log.Fatalf("error loading tls ca cert: %s", err)
		}

		cert, err := ioutil.ReadFile(swarmTlsCert)
		if err != nil {
			log.Fatalf("error loading tls cert: %s", err)
		}

		key, err := ioutil.ReadFile(swarmTlsKey)
		if err != nil {
			log.Fatalf("error loading tls key: %s", err)
		}

		cfg, err := getTLSConfig(caCert, cert, key, allowInsecureTls)
		if err != nil {
			log.Fatalf("error configuring tls: %s", err)
		}
		tlsConfig = cfg
	}

	m := NewManager(config, tlsConfig)

	log.Infof("interlock running version=%s", VERSION)
	if err := m.Run(); err != nil {
		log.Fatal(err)
	}

	waitForInterrupt()
}
