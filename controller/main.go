package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/ehazlett/interlock"
	"github.com/sirupsen/logrus"
)

var (
	configPath      string
	proxyConfigPath string
	proxyPort       int
	syslogAddr      string
	logger          = logrus.New()
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
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&proxyConfigPath, "proxy-conf-path", "", "path to proxy file")
	flag.StringVar(&syslogAddr, "syslog", "", "address to syslog (optional)")
	flag.IntVar(&proxyPort, "proxy-port", 0, "proxy listen port")
	flag.Parse()
}

func main() {
	config, err := loadConfig()
	if err != nil {
		logger.Fatalf("unable to load config: %s", err)
	}
	if proxyConfigPath != "" {
		config.ProxyConfigPath = proxyConfigPath
	}
	if proxyPort != 0 {
		config.Port = proxyPort
	}
	if syslogAddr != "" {
		config.SyslogAddr = syslogAddr
	}
	m, err := NewManager(config)
	if err != nil {
		logger.Fatalf("unable to init manager: %s", err)
	}
	logger.Infof("Interlock running proxy=:%d config=%s", m.config.Port, configPath)
	if err := m.Run(); err != nil {
		logger.Fatal(err)
	}
}
