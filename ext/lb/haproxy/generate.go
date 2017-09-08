package haproxy

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/ehazlett/interlock/ext/lb/utils"
	"golang.org/x/net/context"
	"github.com/docker/docker/api/types/swarm"
	"net"
)

func (p *HAProxyLoadBalancer) GenerateProxyConfig(containers []types.Container, tasks []swarm.Task) (interface{}, error) {
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
	cntId := ""
	labels := map[string]string{}
	container_name := ""

	networks := map[string]string{}

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
			labels = t.Labels
		default:
			log().Warnf("unknown type detected: %v", t)
			continue
		}

		// load interlock data
		hostname := utils.Hostname(labels)
		domain := utils.Domain(labels)

		// context root
		contextRoot := utils.ContextRoot(labels)
		contextRootName := strings.Replace(contextRoot, "/", "_", -1)

		if domain == "" && contextRoot == "" {
			continue
		}

		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
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
			container_name = t.Names[0][1:]
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
			container_name = ""//t.Names[0][1:]
		default:
			log().Warnf("unknown type detected: %v", t)
			continue
		}

		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: healthCheckInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(labels)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, cntId)
			proxyUpstreams[alias] = append(proxyUpstreams[alias], up)
			hostContextRoots[alias] = &ContextRoot{
				Name: contextRootName,
				Path: contextRoot,
			}
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

func (p *HAProxyLoadBalancer) GenerateProxyConfigForTasks(tasks []swarm.Task) (interface{}, error) {
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

	for _, t := range tasks {
		cntId := t.ID[:12]
		labels := t.Labels
		// load interlock data
		hostname := utils.Hostname(labels)
		domain := utils.Domain(labels)

		// context root
		contextRoot := utils.ContextRoot(labels)
		contextRootName := strings.Replace(contextRoot, "/", "_", -1)

		if domain == "" && contextRoot == "" {
			continue
		}

		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
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

		addr := ""

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

		container_name := t.ID
		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: healthCheckInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(labels)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, cntId)
			proxyUpstreams[alias] = append(proxyUpstreams[alias], up)
			hostContextRoots[alias] = &ContextRoot{
				Name: contextRootName,
				Path: contextRoot,
			}
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
