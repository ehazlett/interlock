package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func GetTLSConfig(caCert, cert, key []byte, allowInsecure bool) (*tls.Config, error) {
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

func GetDockerClient(dockerUrl, tlsCaCert, tlsCert, tlsKey string, allowInsecure bool) (*dockerclient.DockerClient, error) {
	// check environment for docker client config
	envDockerHost := os.Getenv("DOCKER_HOST")
	if dockerUrl == "" && envDockerHost != "" {
		dockerUrl = envDockerHost
	}

	envDockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	envDockerTlsVerify := os.Getenv("DOCKER_TLS_VERIFY")
	if tlsCaCert == "" && envDockerCertPath != "" && envDockerTlsVerify != "" {
		tlsCaCert = filepath.Join(envDockerCertPath, "ca.pem")
		tlsCert = filepath.Join(envDockerCertPath, "cert.pem")
		tlsKey = filepath.Join(envDockerCertPath, "key.pem")
	}

	// load tlsconfig
	var tlsConfig *tls.Config
	if tlsCaCert != "" && tlsCert != "" && tlsKey != "" {
		log.Debug("using tls for communication with docker")
		caCert, err := ioutil.ReadFile(tlsCaCert)
		if err != nil {
			log.Fatalf("error loading tls ca cert: %s", err)
		}

		cert, err := ioutil.ReadFile(tlsCert)
		if err != nil {
			log.Fatalf("error loading tls cert: %s", err)
		}

		key, err := ioutil.ReadFile(tlsKey)
		if err != nil {
			log.Fatalf("error loading tls key: %s", err)
		}

		cfg, err := GetTLSConfig(caCert, cert, key, allowInsecure)
		if err != nil {
			log.Fatalf("error configuring tls: %s", err)
		}
		tlsConfig = cfg
	}

	log.Debugf("docker client: url=%s", dockerUrl)

	client, err := dockerclient.NewDockerClient(dockerUrl, tlsConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
