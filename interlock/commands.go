package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/manager"
	"github.com/ehazlett/interlock/plugins"
	"github.com/ehazlett/interlock/version"
)

func waitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	for _ = range sigChan {
		return
	}
}

func cmdStart(c *cli.Context) {
	swarmURL := c.GlobalString("swarm-url")
	swarmTLSCaCert := c.GlobalString("swarm-tls-ca-cert")
	swarmTLSCert := c.GlobalString("swarm-tls-cert")
	swarmTLSKey := c.GlobalString("swarm-tls-key")
	allowInsecureTLS := c.GlobalBool("swarm-allow-insecure")

	// only load env vars if no args
	// check environment for docker client config
	envDockerHost := os.Getenv("DOCKER_HOST")
	if swarmURL == "" && envDockerHost != "" {
		swarmURL = envDockerHost
	}

	// only load env vars if no args
	envDockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	envDockerTLSVerify := os.Getenv("DOCKER_TLS_VERIFY")
	if swarmTLSCaCert == "" && envDockerCertPath != "" && envDockerTLSVerify != "" {
		swarmTLSCaCert = filepath.Join(envDockerCertPath, "ca.pem")
		swarmTLSCert = filepath.Join(envDockerCertPath, "cert.pem")
		swarmTLSKey = filepath.Join(envDockerCertPath, "key.pem")
	}

	config := &interlock.Config{}
	config.SwarmUrl = swarmURL
	config.EnabledPlugins = c.GlobalStringSlice("plugin")

	// load tlsconfig
	var tlsConfig *tls.Config
	if swarmTLSCaCert != "" && swarmTLSCert != "" && swarmTLSKey != "" {
		log.Infof("using tls for communication with swarm")
		caCert, err := ioutil.ReadFile(swarmTLSCaCert)
		if err != nil {
			log.Fatalf("error loading tls ca cert: %s", err)
		}

		cert, err := ioutil.ReadFile(swarmTLSCert)
		if err != nil {
			log.Fatalf("error loading tls cert: %s", err)
		}

		key, err := ioutil.ReadFile(swarmTLSKey)
		if err != nil {
			log.Fatalf("error loading tls key: %s", err)
		}

		cfg, err := getTLSConfig(caCert, cert, key, allowInsecureTLS)
		if err != nil {
			log.Fatalf("error configuring tls: %s", err)
		}
		tlsConfig = cfg
	}

	m := manager.NewManager(config, tlsConfig)

	log.Infof("interlock running version=%s", version.FullVersion())
	if err := m.Run(); err != nil {
		log.Fatal(err)
	}

	waitForInterrupt()

	log.Infof("shutting down")
	if err := m.Stop(); err != nil {
		log.Fatal(err)
	}
}

func cmdListPlugins(c *cli.Context) {
	allPlugins := plugins.GetPlugins()
	w := tabwriter.NewWriter(os.Stdout, 8, 1, 3, ' ', 0)

	fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION\tURL")

	for _, p := range allPlugins {
		i := p.Info()
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			i.Name,
			i.Version,
			i.Description,
			i.Url,
		)
	}
	w.Flush()
}

func cmdInfo(c *cli.Context) {
	haproxyOut, hErr := exec.Command("/usr/sbin/haproxy", "-v").Output()
	if hErr != nil {
		log.Fatal(hErr)
	}

	hData := strings.Split(string(haproxyOut), "\n")

	nginxOut, nErr := exec.Command("/usr/sbin/nginx", "-v").CombinedOutput()
	if nErr != nil {
		log.Fatal(nErr)
	}

	nData := strings.Split(string(nginxOut), "\n")

	fmt.Println("interlock " + version.FullVersion())
	fmt.Println(" " + string(hData[0]))
	fmt.Println(" " + string(nData[0]))
}
