package nginx

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/engine-api/types"
	swarmtypes "github.com/docker/engine-api/types/swarm"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/lb/utils"
	"golang.org/x/net/context"
)

func (p *NginxLoadBalancer) GenerateProxyConfig(containers []types.Container, services []swarmtypes.Service) (interface{}, error) {
	var hosts []*Host
	upstreamServers := map[string][]string{}
	serverNames := map[string][]string{}
	hostContextRoots := map[string]*ContextRoot{}
	hostContextRootRewrites := map[string]bool{}
	hostSSL := map[string]bool{}
	hostSSLCert := map[string]string{}
	hostSSLCertKey := map[string]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostWebsocketEndpoints := map[string][]string{}
	hostIPHash := map[string]bool{}
	networks := map[string]string{}

	var backends []interface{}

	for _, i := range containers {
		backends = append(backends, i)
	}

	for _, i := range services {
		backends = append(backends, i)
	}

	for _, c := range backends {
		var labels map[string]string

		addr := ""
		hostname := ""
		domain := ""
		id := ""

		switch t := c.(type) {
		case types.Container:
			labels = t.Labels
			cntID := t.ID[:12]
			// load interlock data
			cInfo, err := p.client.ContainerInspect(context.Background(), t.ID)
			if err != nil {
				return nil, err
			}

			log().Debugf("checking container: id=%s", cntID)
			id = cntID

			hostname = cInfo.Config.Hostname
			domain = cInfo.Config.Domainname

			if n, ok := utils.OverlayEnabled(labels); ok {
				log().Debugf("configuring docker network: name=%s", n)

				network, err := p.client.NetworkInspect(context.Background(), n)
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
					log().Warnf("%s: no ports exposed", cntID)
					continue
				}

				addr, err = utils.BackendAddress(cInfo, p.cfg.BackendOverrideAddress)
				if err != nil {
					log().Error(err)
					continue
				}
			}
		case swarmtypes.Service:
			log().Debugf("checking service: id=%s", t.ID)
			labels = t.Spec.Labels
			id = t.ID
			servicePort := uint32(0)

			// get service address
			if v, ok := t.Spec.Labels[ext.InterlockPortLabel]; ok {
				port, err := strconv.Atoi(v)
				if err != nil {
					log().Error(err)
					continue
				}
				servicePort = uint32(port)
			} else {
				if len(t.Endpoint.Spec.Ports) == 0 {
					log().Debugf("service has no published ports and no label for port: id=%s", t.ID)
					continue
				}

				servicePort = t.Endpoint.Ports[0].TargetPort
			}

			// get the node IP
			ip := t.Spec.Name

			addr = fmt.Sprintf("%s:%d", ip, servicePort)
		default:
			log().Warnf("unknown type detected: %v", t)
			continue
		}

		if v := utils.Hostname(labels); v != "" {
			hostname = v
		}
		if v := utils.Domain(labels); v != "" {
			domain = v
		}

		// context root
		contextRoot := utils.ContextRoot(labels)
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
		hostContextRootRewrites[domain] = utils.ContextRootRewrite(labels)

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

		// "parse" multiple labels for websocket endpoints
		websocketEndpoints := utils.WebsocketEndpoints(labels)

		log().Debugf("websocket endpoints: %v", websocketEndpoints)

		// websocket endpoints
		for _, ws := range websocketEndpoints {
			hostWebsocketEndpoints[domain] = append(hostWebsocketEndpoints[domain], ws)
		}

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(labels)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, id)
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
			ContextRootRewrite: hostContextRootRewrites[k],
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
