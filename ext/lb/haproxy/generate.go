package haproxy

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/docker/engine-api/types"
	swarmtypes "github.com/docker/engine-api/types/swarm"
	"github.com/ehazlett/interlock/ext"
	"github.com/ehazlett/interlock/ext/lb/utils"
	"golang.org/x/net/context"
)

func (p *HAProxyLoadBalancer) GenerateProxyConfig(containers []types.Container, services []swarmtypes.Service) (interface{}, error) {
	var hosts []*Host

	proxyUpstreams := map[string][]*Upstream{}
	hostChecks := map[string]string{}
	hostBalanceAlgorithms := map[string]string{}
	hostContextRoots := map[string]*ContextRoot{}
	hostContextRootRewrites := map[string]bool{}
	hostBackendOptions := map[string][]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostSSLBackendTLSVerify := map[string]string{}
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
		containerName := ""

		switch t := c.(type) {
		case types.Container:
			labels = t.Labels
			cntID := t.ID[:12]
			// load interlock data
			cInfo, err := p.client.ContainerInspect(context.Background(), t.ID)
			if err != nil {
				return nil, err
			}

			id = cntID
			containerName = cInfo.Name[1:]

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
			labels = t.Spec.Labels
			id = t.ID
			publishedPort := uint32(0)

			// get service address
			if len(t.Endpoint.Spec.Ports) == 0 {
				log().Debugf("service has no published ports: id=%s", t.ID)
				continue
			}

			if v, ok := t.Spec.Labels[ext.InterlockPortLabel]; ok {
				port, err := strconv.Atoi(v)
				if err != nil {
					log().Error(err)
					continue
				}
				for _, p := range t.Endpoint.Ports {
					if p.TargetPort == uint32(port) {
						publishedPort = p.PublishedPort
						break
					}
				}
			} else {
				publishedPort = t.Endpoint.Spec.Ports[0].PublishedPort
			}

			// get the node IP
			ip := ""

			// HACK?: get the local node gateway addr to use as the ip to resolve for the interlock container to access the published port
			network, err := p.client.NetworkInspect(context.Background(), "docker_gwbridge")
			if err != nil {
				log().Error(err)
				continue
			}

			// TODO: what do we do if the IPAM has more than a single definition?
			ipAddr, _, err := net.ParseCIDR(network.IPAM.Config[0].Gateway)
			if err != nil {
				log().Error(err)
				continue
			}

			ip = ipAddr.String()

			// check for override backend address
			if v := p.cfg.BackendOverrideAddress; v != "" {
				ip = v
			}

			addr = fmt.Sprintf("%s:%d", ip, publishedPort)
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

		healthCheck := utils.HealthCheck(labels)
		healthCheckInterval, err := utils.HealthCheckInterval(labels)
		if err != nil {
			log().Errorf("error parsing health check interval: %s", err)
			continue
		}

		if healthCheck != "" {
			if val, ok := hostChecks[domain]; ok {
				// check existing host check for different values
				if val != healthCheck {
					log().Warnf("conflicting check specified for %s", domain)
				}
			} else {
				hostChecks[domain] = healthCheck
				log().Debugf("using custom check for %s: %s", domain, healthCheck)
			}

			log().Debugf("check interval for %s: %d", domain, healthCheckInterval)
		}

		hostBalanceAlgorithms[domain] = utils.BalanceAlgorithm(labels)

		backendOptions := utils.BackendOptions(labels)

		if len(backendOptions) > 0 {
			hostBackendOptions[domain] = backendOptions
			log().Debugf("using backend options for %s: %s", domain, strings.Join(backendOptions, ","))
		}

		hostSSLOnly[domain] = utils.SSLOnly(labels)

		// ssl backend
		hostSSLBackend[domain] = utils.SSLBackend(labels)
		hostSSLBackendTLSVerify[domain] = utils.SSLBackendTLSVerify(labels)

		up := &Upstream{
			Addr:          addr,
			Container:     containerName,
			CheckInterval: healthCheckInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, containerName)

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(labels)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, id)
			proxyUpstreams[alias] = append(proxyUpstreams[alias], up)
		}

		proxyUpstreams[domain] = append(proxyUpstreams[domain], up)
	}

	for k, v := range proxyUpstreams {
		name := strings.Replace(k, ".", "_", -1)
		host := &Host{
			Name:                name,
			ContextRoot:         hostContextRoots[k],
			ContextRootRewrite:  hostContextRootRewrites[k],
			Domain:              k,
			Upstreams:           v,
			Check:               hostChecks[k],
			BalanceAlgorithm:    hostBalanceAlgorithms[k],
			BackendOptions:      hostBackendOptions[k],
			SSLOnly:             hostSSLOnly[k],
			SSLBackend:          hostSSLBackend[k],
			SSLBackendTLSVerify: hostSSLBackendTLSVerify[k],
		}
		log().Debugf("adding host name=%s domain=%s contextroot=%v", host.Name, host.Domain, host.ContextRoot)
		hosts = append(hosts, host)
	}

	cfg := &Config{
		Hosts:    hosts,
		Config:   p.cfg,
		Networks: networks,
	}

	return cfg, nil
}
