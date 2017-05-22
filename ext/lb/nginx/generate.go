package nginx

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext/lb/utils"
	"golang.org/x/net/context"
)

func (p *NginxLoadBalancer) GenerateProxyConfig(containers []types.Container) (interface{}, error) {
	var hosts []*Host
	upstreamHosts := map[string]struct{}{}
	upstreamServers := map[string][]string{}
	serverNames := map[string][]string{}
	hostContextRoots := map[string]map[string]*ContextRoot{}
	//hostContextUpstreams := map[string][]*ContextRootUpstream{}
	hostSSL := map[string]bool{}
	hostSSLCert := map[string]string{}
	hostSSLCertKey := map[string]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostWebsocketEndpoints := map[string][]string{}
	hostIPHash := map[string]bool{}
	networks := map[string]string{}

	for _, c := range containers {
		cntId := c.ID[:12]
		// load interlock data
		cInfo, err := p.client.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			log().Errorf("unable to inspect container for upstream: %s", err)
			continue
		}

		hostname := utils.Hostname(cInfo.Config)
		domain := utils.Domain(cInfo.Config)

		// context root
		contextRoot := utils.ContextRoot(cInfo.Config)
		contextRootName := fmt.Sprintf("%s_%s", domain, strings.Replace(contextRoot, "/", "_", -1))
		contextRootRewrite := utils.ContextRootRewrite(cInfo.Config)

		if domain == "" && contextRoot == "" {
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

		hostSSL[domain] = utils.SSLEnabled(cInfo.Config)
		hostSSLOnly[domain] = utils.SSLOnly(cInfo.Config)
		hostIPHash[domain] = utils.IPHash(cInfo.Config)
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

			network, err := p.client.NetworkInspect(context.Background(), n, false)
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
			portsExposed := false
			for _, portBindings := range cInfo.NetworkSettings.Ports {
				if len(portBindings) != 0 {
					portsExposed = true
					break
				}
			}
			if !portsExposed {
				log().Warnf("%s: no ports exposed", cntId)
				continue
			}

			addr, err = utils.BackendAddress(cInfo, p.cfg.BackendOverrideAddress)
			if err != nil {
				log().Error(err)
				continue
			}
		}

		if contextRoot != "" {
			if _, ok := hostContextRoots[domain]; !ok {
				hostContextRoots[domain] = map[string]*ContextRoot{}
			}
			hc, ok := hostContextRoots[domain][contextRootName]
			if !ok {
				hostContextRoots[domain][contextRootName] = &ContextRoot{
					Name:      contextRootName,
					Path:      contextRoot,
					Rewrite:   contextRootRewrite,
					Upstreams: []string{},
				}

				hc = hostContextRoots[domain][contextRootName]
			}

			hc.Upstreams = append(hc.Upstreams, addr)

			//log().Debugf("adding contextroot upstream: %s=%s", contextRootName, addr)
			//hostContextUpstreams[domain] = append(hostContextUpstreams[domain], &ContextRootUpstream{
			//	Name:     contextRootName,
			//	Upstream: addr,
			//})
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

		if contextRoot == "" {
			upstreamServers[domain] = append(upstreamServers[domain], addr)
		}

		upstreamHosts[domain] = struct{}{}
		log().Infof("%s: upstream=%s", domain, addr)
	}

	for k, _ := range upstreamHosts {
		log().Debugf("%s contextroots=%+v", k, hostContextRoots[k])
		h := &Host{
			ServerNames:        serverNames[k],
			Port:               p.cfg.Port,
			ContextRoots:       hostContextRoots[k],
			SSLPort:            p.cfg.SSLPort,
			SSL:                hostSSL[k],
			SSLCert:            hostSSLCert[k],
			SSLCertKey:         hostSSLCertKey[k],
			SSLOnly:            hostSSLOnly[k],
			SSLBackend:         hostSSLBackend[k],
			WebsocketEndpoints: hostWebsocketEndpoints[k],
			IPHash:             hostIPHash[k],
		}

		servers := []*Server{}

		for _, s := range upstreamServers[k] {
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
