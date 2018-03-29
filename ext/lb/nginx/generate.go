package nginx

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/ehazlett/interlock/ext/lb/utils"
	"golang.org/x/net/context"
	"net"
)

func (p *NginxLoadBalancer) GenerateProxyConfig(containers []types.Container, tasks [] swarm.Task) (interface{}, error) {
	var hosts []*Host
	upstreamHosts := map[string]struct{}{}
	upstreamServers := map[string][]string{}
	serverNames := map[string][]string{}
	hostContextRoots := map[string]map[string]*ContextRoot{}
	hostBackendOptions := map[string][]string{}
	hostSSL := map[string]bool{}
	hostSSLCert := map[string]string{}
	hostSSLCertKey := map[string]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostWebsocketEndpoints := map[string][]string{}
	hostIPHash := map[string]bool{}
	networks := map[string]string{}
	cntId := ""
	labels := map[string]string{}

	var backends []interface{}

	for _, i := range containers {
		backends = append(backends, i)
	}

	for _, i := range tasks {
		backends = append(backends, i)
	}

	for _, c := range backends {
		switch t := c.(type) {
		case types.Container:
			cntId = t.ID[:12]
			labels = t.Labels
		case swarm.Task:
			cntId = t.ID[:12]
			labels = t.Spec.ContainerSpec.Labels

			if t.Status.State != swarm.TaskStateRunning {
				continue
			}

		default:
			log().Warnf("unknown type detected: %v", t)
			continue
		}

		// load interlock data
		contextRoot := utils.ContextRoot(labels)

		hostname := utils.Hostname(labels)
		domain := utils.Domain(labels)


		if domain == "local" && contextRoot == "" {
			continue
		}

		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
		}

		// context root
		contextRootName := fmt.Sprintf("%s_%s", domain, strings.Replace(contextRoot, "/", "_", -1))
		contextRootRewrite := utils.ContextRootRewrite(labels)

		// check if the first server name is there; if not, add
		// this happens if there are multiple backend containers
		if _, ok := serverNames[domain]; !ok {
			serverNames[domain] = []string{domain}
		}

		hostSSL[domain] = utils.SSLEnabled(labels)
		hostSSLOnly[domain] = utils.SSLOnly(labels)
		hostIPHash[domain] = utils.IPHash(labels)
		// check ssl backend
		hostSSLBackend[domain] = utils.SSLBackend(labels)

		backendOptions := utils.BackendOptions(c)
		if len(backendOptions) > 0 {
			hostBackendOptions[domain] = backendOptions
			log().Debugf("using backend options for %s: %s", domain, strings.Join(backendOptions, ","))
		}

		// set cert paths
		baseCertPath := p.cfg.SSLCertPath

		certName := utils.SSLCertName(labels)

		if certName != "" {
			certPath := filepath.Join(baseCertPath, certName)
			log().Infof("ssl cert for %s: %s", domain, certPath)
			hostSSLCert[domain] = certPath
		}

		certKeyName := utils.SSLCertKey(labels)
		if certKeyName != "" {
			keyPath := filepath.Join(baseCertPath, certKeyName)
			log().Infof("ssl key for %s: %s", domain, keyPath)
			hostSSLCertKey[domain] = keyPath
		}

		addr := ""

		switch t := c.(type) {
		case types.Container:

		// check for networking
		if n, ok := utils.OverlayEnabled(labels); ok {
			log().Debugf("configuring docker network: name=%s", n)

			network, err := p.client.NetworkInspect(context.Background(), n, false)
			if err != nil {
				log().Error(err)
				continue
			}

			addr, err = utils.BackendOverlayAddress(network, t)
			if err != nil {
				log().Error(err)
				continue
			}

			networks[n] = ""
		} else {
			if len(t.Ports) == 0 {
				log().Warnf("%s: no ports exposed", cntId)
				continue
			}

			a, err := utils.BackendAddress(t, p.cfg.BackendOverrideAddress)
			if err != nil {
				log().Error(err)
				continue
			}

			addr = a
		}
		case swarm.Task:
			interlockPort, err := utils.CustomPort(labels)
			if err != nil {
				log().Error(err)
				continue
			}
			log().Debug(interlockPort)

			// check for networking
			if overlayNetworkName, ok := utils.OverlayEnabled(labels); ok {
				log().Debugf("configuring docker network: name=%s", overlayNetworkName)

				for _, networksAttachment := range t.NetworksAttachments {

					if overlayNetworkName == networksAttachment.Network.Spec.Annotations.Name {
						for _, address := range networksAttachment.Addresses {
							log().Debug(address)

							ip, _, err := net.ParseCIDR(address)
							if err != nil {
								log().Error(err)
								continue
							}

							addr = fmt.Sprintf("%s:%d", ip, interlockPort)
							log().Debug(addr)
						}
					}
				}

				networks[overlayNetworkName] = ""
			} else {

				//addr = fmt.Sprintf("%s:%d", network, interlockPort)
			}
		default:
			log().Warnf("unknown type detected: %v", t)
			continue
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
		}

		// "parse" multiple labels for websocket endpoints
		websocketEndpoints := utils.WebsocketEndpoints(labels)

		log().Debugf("websocket endpoints: %v", websocketEndpoints)

		// websocket endpoints
	outer:
		for _, ws := range websocketEndpoints {
			for _, existingEndpoint := range hostWebsocketEndpoints[domain] {
				if existingEndpoint == ws {
					continue outer
				}
			}
			hostWebsocketEndpoints[domain] = append(hostWebsocketEndpoints[domain], ws)
		}

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(labels)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, cntId)
			serverNames[domain] = append(serverNames[domain], alias)
		}

		if contextRoot == "" {
			log().Debugf("adding upstream %s: upstream=%s", domain, addr)
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
			BackendOptions:     hostBackendOptions[k],
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

