package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/ehazlett/interlock"
	"github.com/sirupsen/logrus"
)

var (
	configPath         string
	proxyConfigPath    string
	proxyPidPath       string
	proxyPort          int
	syslogAddr         string
	shipyardUrl        string
	shipyardServiceKey string
	sslCert            string
	sslOpts            string
	sslPort            int
	dockerUrl          string
	logger             = logrus.New()
)

func loadConfig() (*interlock.Config, error) {
	var config *interlock.Config
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func init() {
	flag.StringVar(&dockerUrl, "swarm-url", "tcp://127.0.0.1:2375", "Swarm URL")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&proxyConfigPath, "proxy-conf-path", "proxy.conf", "path to proxy file")
	flag.StringVar(&proxyPidPath, "proxy-pid-path", "proxy.pid", "path to proxy pid file")
	flag.StringVar(&syslogAddr, "syslog", "", "address to syslog (optional)")
	flag.IntVar(&proxyPort, "proxy-port", 8080, "proxy listen port")
	flag.StringVar(&sslCert, "ssl-cert", "", "path to ssl cert (enables SSL)")
	flag.IntVar(&sslPort, "ssl-port", 8443, "ssl listen port (must have cert above)")
	flag.StringVar(&sslOpts, "ssl-opts", "", "string of SSL options (eg. ciphers or tls versions)")
	flag.Parse()
}

func main() {
	config := &interlock.Config{}
	config.DockerUrl = dockerUrl
	config.ProxyConfigPath = proxyConfigPath
	config.PidPath = proxyPidPath
	config.Port = proxyPort
	config.SSLPort = sslPort
	config.SSLOpts = sslOpts
	if syslogAddr != "" {
		config.SyslogAddr = syslogAddr
	}
	if sslCert != "" {
		config.SSLCert = sslCert
	}
	m, err := NewManager(config)
	if err != nil {
		logger.Fatalf("unable to init manager: %s", err)
	}
	logger.Infof("Interlock running proxy=:%d", m.config.Port)
	if m.config.SSLCert != "" {
		logger.Infof("SSL listener active=:%d", m.config.SSLPort)
	}
	if err := m.Run(); err != nil {
		logger.Fatal(err)
	}
}
