package nginx

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ehazlett/interlock/ext/lb/utils"
	"github.com/samalba/dockerclient"
)

func (p *NginxLoadBalancer) GenerateProxyConfig(containers []dockerclient.Container) (interface{}, error) {
	var hosts []*Host
	upstreamServers := map[string][]string{}
	serverNames := map[string][]string{}
	hostContextRoots := map[string]*ContextRoot{}
	hostSSL := map[string]bool{}
	hostSSLCert := map[string]string{}
	hostSSLCertKey := map[string]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostWebsocketEndpoints := map[string][]string{}

	networks := map[string]string{}

	for _, c := range containers {
		cntId := c.Id[:12]
		// load interlock data
		cInfo, err := p.client.InspectContainer(c.Id)
		if err != nil {
			return nil, err
		}

		hostname := utils.Hostname(cInfo.Config)
		domain := utils.Domain(cInfo.Config)

		// context root
		contextRoot := utils.ContextRoot(cInfo.Config)
		contextRootName := strings.Replace(contextRoot, "/", "_", -1)

		if domain == "" && contextRoot == "" {
			continue
		}

		// we check if a context root is passed and overwrite the
		// domain component
		if contextRoot != "" {
			domain = contextRootName
		} else {
			if hostname != domain && hostname != "" {
				domain = fmt.Sprintf("%s.%s", hostname, domain)
			}
		}

		hostContextRoots[domain] = &ContextRoot{
			Name: contextRootName,
			Path: contextRoot,
		}

		// check if the first server name is there; if not, add
		// this happens if there are multiple backend containers
		if _, ok := serverNames[domain]; !ok {
			serverNames[domain] = []string{domain}
		}

		hostSSL[domain] = utils.SSLEnabled(cInfo.Config)
		hostSSLOnly[domain] = utils.SSLOnly(cInfo.Config)

		// check ssl backend
		hostSSLBackend[domain] = utils.SSLBackend(cInfo.Config)

		// set cert paths
		baseCertPath := p.cfg.SSLCertPath

		certName := utils.SSLCertName(cInfo.Config)

		if certName != "" {
			certPath := filepath.Join(baseCertPath, certName)
			log().Infof("ssl cert for %s: %s", domain, certPath)
			hostSSLCert[domain] = certPath
		}

		certKeyName := utils.SSLCertKey(cInfo.Config)
		if certKeyName != "" {
			keyPath := filepath.Join(baseCertPath, certKeyName)
			log().Infof("ssl key for %s: %s", domain, keyPath)
			hostSSLCertKey[domain] = keyPath
		}

		addr := ""

		// check for networking
		if n, ok := utils.OverlayEnabled(cInfo.Config); ok {
			log().Debugf("configuring docker network: name=%s", n)

			network, err := p.client.InspectNetwork(n)
			if err != nil {
				log().Error(err)
				continue
			}

			addr, err = utils.BackendOverlayAddress(network, cInfo)
			if err != nil {
				log().Error(err)
				continue
			}

			networks[n] = ""
		} else {
			if len(cInfo.NetworkSettings.Ports) == 0 {
				log().Warnf("%s: no ports exposed", cntId)
				continue
			}

			addr, err = utils.BackendAddress(cInfo, p.cfg.BackendOverrideAddress)
			if err != nil {
				log().Error(err)
				continue
			}
		}

		// "parse" multiple labels for websocket endpoints
		websocketEndpoints := utils.WebsocketEndpoints(cInfo.Config)

		log().Debugf("websocket endpoints: %v", websocketEndpoints)

		// websocket endpoints
		for _, ws := range websocketEndpoints {
			hostWebsocketEndpoints[domain] = append(hostWebsocketEndpoints[domain], ws)
		}

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(cInfo.Config)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, cntId)
			serverNames[domain] = append(serverNames[domain], alias)
		}

		log().Infof("%s: upstream=%s", domain, addr)

		upstreamServers[domain] = append(upstreamServers[domain], addr)
	}

	for k, v := range upstreamServers {
		h := &Host{
			ServerNames:        serverNames[k],
			Port:               p.cfg.Port,
			ContextRoot:        hostContextRoots[k],
			SSLPort:            p.cfg.SSLPort,
			SSL:                hostSSL[k],
			SSLCert:            hostSSLCert[k],
			SSLCertKey:         hostSSLCertKey[k],
			SSLOnly:            hostSSLOnly[k],
			SSLBackend:         hostSSLBackend[k],
			WebsocketEndpoints: hostWebsocketEndpoints[k],
		}

		servers := []*Server{}

		for _, s := range v {
			srv := &Server{
				Addr: s,
			}

			servers = append(servers, srv)
		}

		up := &Upstream{
			Name:    k,
			Servers: servers,
		}
		h.Upstream = up

		hosts = append(hosts, h)
	}

	config := &Config{
		Hosts:    hosts,
		Config:   p.cfg,
		Networks: networks,
	}

	return config, nil
}
