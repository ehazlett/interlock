package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"

	"github.com/codegangsta/cli"
	"github.com/ehazlett/interlock"
	"github.com/samalba/dockerclient"
)

var appCommands = []cli.Command{
	{
		Name:   "ls",
		Usage:  "list available plugins",
		Action: listPlugins,
	},
}

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

func getDockerClient(c *cli.Context) (*dockerclient.DockerClient, error) {
	var tlsConfig *tls.Config
	tlsCaCert := c.GlobalString("tls-ca-cert")
	tlsCert := c.GlobalString("tls-cert")
	tlsKey := c.GlobalString("tls-key")
	allowInsecure := c.GlobalBool("allow-insecure")
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

	d, err := dockerclient.NewDockerClient(c.GlobalString("docker"), tlsConfig)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func getPluginInfo(path string, w io.Writer, wg *sync.WaitGroup, ec chan error) {
	defer wg.Done()

	p := NewPluginCmd(path)
	input := &interlock.PluginInput{
		Command: "info",
		Args:    nil,
	}
	out, err := p.Exec(input)
	if err != nil {
		ec <- err
		return
	}

	if out != nil {
		if out.Error != nil {
			ec <- fmt.Errorf(string(out.Error))
			return
		}

		if out.Output != nil {
			var plugin interlock.Plugin
			if err := json.Unmarshal(out.Output, &plugin); err != nil {
				ec <- err
				return
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", plugin.Name,
				plugin.Version, plugin.Author, plugin.Url)
		}
	}
}
func listPlugins(c *cli.Context) {
	pluginPath := c.GlobalString("plugin-path")

	// make sure dir exists
	if err := os.MkdirAll(pluginPath, 0700); err != nil {
		log.Fatal(err)
	}

	plugins, err := ioutil.ReadDir(pluginPath)
	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	fmt.Fprintln(w, "NAME\tVERSION\tAUTHOR\tURL")

	wg := &sync.WaitGroup{}
	ec := make(chan error)

	for _, fi := range plugins {
		wg.Add(1)
		go getPluginInfo(filepath.Join(pluginPath, fi.Name()), w, wg, ec)

	}

	wg.Wait()

	w.Flush()

}
