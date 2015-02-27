package manager

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

type (
	Manager struct {
		client     *dockerclient.DockerClient
		pluginPath string
		plugins    []*interlock.Plugin
	}
)

func getTLSConfig(caCert, cert, key []byte, allowInsecure bool) (*tls.Config, error) {
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
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

func getPluginInfo(pluginPath string, pluginChan chan *interlock.Plugin, errorChan chan error, doneChan chan bool) {
	log.Debugf("loading plugin: %s", pluginPath)
	var plugin interlock.Plugin

	p := plugins.NewPluginCmd(pluginPath)
	input := &interlock.PluginInput{
		Command: "info",
		Data:    nil,
	}

	out, err := p.Exec(input)
	if err != nil {
		errorChan <- err
	}

	if out != nil {
		if out.Error != nil {
			errorChan <- fmt.Errorf(string(out.Error))
		}

		if out.Output != nil {
			if err := json.Unmarshal(out.Output, &plugin); err != nil {
				errorChan <- err
			}

			log.Debugf("loaded plugin: name=%s version=%s", plugin.Name, plugin.Version)
			plugin.Path = pluginPath
			pluginChan <- &plugin
		}
	}

	doneChan <- true
}

func loadPlugins(pluginPath string) ([]*interlock.Plugin, error) {
	log.Debugf("plugin path: %s", pluginPath)

	pluginFiles, err := ioutil.ReadDir(pluginPath)
	if err != nil {
		return nil, err
	}

	plugins := []*interlock.Plugin{}
	pluginChan := make(chan *interlock.Plugin)
	errorChan := make(chan error)
	doneChan := make(chan bool)

	go func() {
		for {
			err := <-errorChan
			log.Error("error loading plugin: %s", err)
		}
	}()

	go func() {
		for {
			plugin := <-pluginChan
			plugins = append(plugins, plugin)
		}
	}()

	for _, fi := range pluginFiles {
		go getPluginInfo(filepath.Join(pluginPath, fi.Name()), pluginChan, errorChan, doneChan)
	}

	for i := 0; i < len(pluginFiles); i++ {
		<-doneChan
	}

	return plugins, nil

}

func NewManager(dockerUrl, tlsCaCert, tlsCert, tlsKey string, allowInsecure bool, pluginPath string) (*Manager, error) {
	var tlsConfig *tls.Config

	// attempt to load the certs from the DOCKER_CERT_PATH
	certPath := os.Getenv("DOCKER_CERT_PATH")
	if certPath != "" {
		tlsCaCert = filepath.Join(certPath, "ca.pem")
		tlsCert = filepath.Join(certPath, "cert.pem")
		tlsKey = filepath.Join(certPath, "key.pem")
	}

	if tlsCaCert != "" && tlsCert != "" && tlsKey != "" {
		caCert, err := ioutil.ReadFile(tlsCaCert)
		if err != nil {
			return nil, fmt.Errorf("error loading tls ca cert: %s", err)
		}

		cert, err := ioutil.ReadFile(tlsCert)
		if err != nil {
			return nil, fmt.Errorf("error loading tls cert: %s", err)
		}

		key, err := ioutil.ReadFile(tlsKey)
		if err != nil {
			return nil, fmt.Errorf("error loading tls key: %s", err)
		}

		cfg, err := getTLSConfig(caCert, cert, key, allowInsecure)
		if err != nil {
			return nil, fmt.Errorf("error configuring tls: %s", err)
		}
		tlsConfig = cfg
	}

	d, err := dockerclient.NewDockerClient(dockerUrl, tlsConfig)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		client:     d,
		pluginPath: pluginPath,
	}

	if err := m.ReloadPlugins(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) ReloadPlugins() error {
	plugins, err := loadPlugins(m.pluginPath)
	if err != nil {
		return err
	}

	m.plugins = plugins
	return nil
}

func (m *Manager) Plugins() ([]*interlock.Plugin, error) {
	if err := m.ReloadPlugins(); err != nil {
		return nil, err
	}
	return m.plugins, nil
}

func (m *Manager) eventHandler(evt *dockerclient.Event, errorChan chan error, args ...interface{}) {
	outputChan := make(chan *interlock.PluginOutput)

	go func() {
		output := <-outputChan
		log.Debugf("response: plugin=%v command=%s output: %s",
			output.Plugin,
			output.Command,
			output.Output,
		)
	}()

	for _, plugin := range m.plugins {
		log.Infof("notifying plugin: name=%s version=%s event=%s", plugin.Name, plugin.Version, evt.Status)
		go m.notifyPlugin(plugin, evt, outputChan, errorChan)
	}
}

func (m *Manager) notifyPlugin(plugin *interlock.Plugin, evt *dockerclient.Event, outputChan chan *interlock.PluginOutput, errorChan chan error) {
	p := plugins.NewPluginCmd(plugin.Path)

	data, err := json.Marshal(evt)
	if err != nil {
		errorChan <- fmt.Errorf("error serializing event: %s", err)
		return
	}

	input := &interlock.PluginInput{
		Command: "event",
		Data:    data,
	}

	out, err := p.Exec(input)
	if err != nil {
		errorChan <- err
		return
	}

	if out != nil {
		outputChan <- out
	}
}

func (m *Manager) Run() {
	errorChan := make(chan error)

	go func() {
		err := <-errorChan
		log.Error(err)
	}()

	m.client.StartMonitorEvents(m.eventHandler, nil)
}
