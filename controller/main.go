package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
)

var (
	configPath                  string
	proxyConfigPath             string
	proxyPidPath                string
	proxyPort                   int
	syslogAddr                  string
	shipyardUrl                 string
	shipyardServiceKey          string
	sslCert                     string
	sslOpts                     string
	sslPort                     int
	swarmUrl                    string
	debug                       bool
	swarmTlsCaCert              string
	swarmTlsCert                string
	swarmTlsKey                 string
	allowInsecureTls            bool
	proxyBackendOverrideAddress string
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

func init() {
	flag.StringVar(&swarmUrl, "swarm-url", "tcp://127.0.0.1:2375", "swarm url")
	flag.StringVar(&swarmTlsCaCert, "tlscacert", "", "ca certificate for tls")
	flag.StringVar(&swarmTlsCert, "tlscert", "", "certificate for tls")
	flag.StringVar(&swarmTlsKey, "tlskey", "", "key for tls")
	flag.BoolVar(&allowInsecureTls, "allow-insecure-tls", false, "allow insecure certificates for TLS")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&proxyConfigPath, "proxy-conf-path", "proxy.conf", "path to proxy file")
	flag.StringVar(&proxyPidPath, "proxy-pid-path", "proxy.pid", "path to proxy pid file")
	flag.StringVar(&proxyBackendOverrideAddress, "proxy-backend-override-address", "", "force proxy to use this address in all backends")
	flag.StringVar(&syslogAddr, "syslog", "", "address to syslog")
	flag.IntVar(&proxyPort, "proxy-port", 8080, "proxy listen port")
	flag.StringVar(&sslCert, "ssl-cert", "", "path to ssl cert (enables SSL)")
	flag.IntVar(&sslPort, "ssl-port", 8443, "ssl listen port (must have cert above)")
	flag.StringVar(&sslOpts, "ssl-opts", "", "string of SSL options (eg. ciphers or tls versions)")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.Parse()
}

func main() {
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	config := &interlock.Config{}
	config.SwarmUrl = swarmUrl
	config.ProxyConfigPath = proxyConfigPath
	config.ProxyBackendOverrideAddress = proxyBackendOverrideAddress
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

	// load tlsconfig
	var tlsConfig *tls.Config
	if swarmTlsCaCert != "" && swarmTlsCert != "" && swarmTlsKey != "" {
		log.Infof("using tls for communication with swarm")
		caCert, err := ioutil.ReadFile(swarmTlsCaCert)
		if err != nil {
			log.Fatalf("error loading ca cert: %s", err)
		}

		cert, err := ioutil.ReadFile(swarmTlsCert)
		if err != nil {
			log.Fatalf("error loading cert: %s", err)
		}

		key, err := ioutil.ReadFile(swarmTlsKey)
		if err != nil {
			log.Fatalf("error loading swarm key: %s", err)
		}

		cfg, err := getTLSConfig(caCert, cert, key, allowInsecureTls)
		if err != nil {
			log.Fatalf("error configuring tls: %s", err)
		}
		tlsConfig = cfg

	}

	m := NewManager(config, tlsConfig)

	log.Infof("interlock running proxy=:%d", m.config.Port)
	if m.config.SSLCert != "" {
		log.Infof("ssl listener active=:%d", m.config.SSLPort)
	}
	if err := m.Run(); err != nil {
		log.Fatal(err)
	}
}
