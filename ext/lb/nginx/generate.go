package nginx

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func (p *NginxLoadBalancer) GenerateProxyConfig() (*Config, error) {
	containers, err := p.client.ListContainers(false, false, "")
	if err != nil {
		return nil, err
	}

	var hosts []*Host
	upstreamServers := map[string][]string{}
	serverNames := map[string][]string{}
	//hostBalanceAlgorithms := map[string]string{}
	hostSSL := map[string]bool{}
	hostSSLCert := map[string]string{}
	hostSSLCertKey := map[string]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostWebsocketEndpoints := map[string][]string{}

	for _, c := range containers {
		cntId := c.Id[:12]
		// load interlock data
		cInfo, err := p.client.InspectContainer(cntId)
		if err != nil {
			return nil, err
		}

		hostname := cInfo.Config.Hostname
		domain := cInfo.Config.Domainname

		if v, ok := cInfo.Config.Labels[ext.InterlockHostnameLabel]; ok {
			hostname = v
		}

		if v, ok := cInfo.Config.Labels[ext.InterlockDomainLabel]; ok {
			domain = v
		}

		if domain == "" {
			continue
		}

		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
		}

		// check if the first server name is there; if not, add
		// this happens if there are multiple backend containers
		if _, ok := serverNames[domain]; !ok {
			serverNames[domain] = []string{domain}
		}

		if _, ok := cInfo.Config.Labels[ext.InterlockSSLLabel]; ok {
			hostSSL[domain] = true
		}

		hostSSLOnly[domain] = false

		if _, ok := cInfo.Config.Labels[ext.InterlockSSLOnlyLabel]; ok {
			log().Infof("configuring ssl redirect for %s", domain)
			hostSSLOnly[domain] = true
		}

		// check ssl backend
		hostSSLBackend[domain] = false
		if _, ok := cInfo.Config.Labels[ext.InterlockSSLBackendLabel]; ok {
			log().Debugf("configuring ssl backend for %s", domain)
			hostSSLBackend[domain] = true
		}

		// set cert paths
		baseCertPath := p.cfg.SSLCertPath
		if v, ok := cInfo.Config.Labels[ext.InterlockSSLCertLabel]; ok {
			certPath := filepath.Join(baseCertPath, v)
			log().Infof("ssl cert for %s: %s", domain, certPath)
			hostSSLCert[domain] = certPath
		}

		if v, ok := cInfo.Config.Labels[ext.InterlockSSLCertKeyLabel]; ok {
			keyPath := filepath.Join(baseCertPath, v)
			log().Infof("ssl key for %s: %s", domain, keyPath)
			hostSSLCertKey[domain] = keyPath
		}

		ports := cInfo.NetworkSettings.Ports
		if len(ports) == 0 {
			log().Warnf("%s: no ports exposed", cntId)
			continue
		}

		var portDef dockerclient.PortBinding

		for _, v := range ports {
			if len(v) > 0 {
				portDef = dockerclient.PortBinding{
					HostIp:   v[0].HostIp,
					HostPort: v[0].HostPort,
				}
				break
			}
		}

		if p.cfg.BackendOverrideAddress != "" {
			portDef.HostIp = p.cfg.BackendOverrideAddress
		}

		addr := fmt.Sprintf("%s:%s", portDef.HostIp, portDef.HostPort)

		if v, ok := cInfo.Config.Labels[ext.InterlockPortLabel]; ok {
			interlockPort := v
			for k, x := range ports {
				parts := strings.Split(k, "/")
				if parts[0] == interlockPort {
					port := x[0]
					log().Debugf("%s: found specified port %s exposed as %s", domain, interlockPort, port.HostPort)
					addr = fmt.Sprintf("%s:%s", portDef.HostIp, port.HostPort)
					break
				}
			}
		}

		// "parse" multiple labels for websocket endpoints
		websocketEndpoints := []string{}
		for l, v := range cInfo.Config.Labels {
			// this is for labels like interlock.websocket_endpoint.1=foo
			if strings.Index(l, ext.InterlockWebsocketEndpointLabel) > -1 {
				websocketEndpoints = append(websocketEndpoints, v)
			}
		}

		log().Debugf("websocket endpoints: %v", websocketEndpoints)

		// websocket endpoints
		for _, ws := range websocketEndpoints {
			hostWebsocketEndpoints[domain] = append(hostWebsocketEndpoints[domain], ws)
		}

		// "parse" multiple labels for alias domains
		aliasDomains := []string{}
		for l, v := range cInfo.Config.Labels {
			// this is for labels like interlock.alias_domain.1=foo.local
			if strings.Index(l, ext.InterlockAliasDomainLabel) > -1 {
				aliasDomains = append(aliasDomains, v)
			}
		}

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

	return &Config{
		Hosts:  hosts,
		Config: p.cfg,
	}, nil
}
