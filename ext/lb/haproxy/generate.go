package haproxy

import (
	"fmt"
	"strings"

	"github.com/docker/engine-api/types"
	"github.com/ehazlett/interlock/ext/lb/utils"
	"golang.org/x/net/context"
)

func (p *HAProxyLoadBalancer) GenerateProxyConfig(containers []types.Container) (interface{}, error) {
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

	for _, cnt := range containers {
		cntId := cnt.ID[:12]
		// load interlock data
		cInfo, err := p.client.ContainerInspect(context.Background(), cntId)
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
		hostContextRootRewrites[domain] = utils.ContextRootRewrite(cInfo.Config)

		healthCheck := utils.HealthCheck(cInfo.Config)
		healthCheckInterval, err := utils.HealthCheckInterval(cInfo.Config)
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

		hostBalanceAlgorithms[domain] = utils.BalanceAlgorithm(cInfo.Config)

		backendOptions := utils.BackendOptions(cInfo.Config)

		if len(backendOptions) > 0 {
			hostBackendOptions[domain] = backendOptions
			log().Debugf("using backend options for %s: %s", domain, strings.Join(backendOptions, ","))
		}

		hostSSLOnly[domain] = utils.SSLOnly(cInfo.Config)

		// ssl backend
		hostSSLBackend[domain] = utils.SSLBackend(cInfo.Config)
		hostSSLBackendTLSVerify[domain] = utils.SSLBackendTLSVerify(cInfo.Config)

		addr := ""

		// check for networking
		if n, ok := utils.OverlayEnabled(cInfo.Config); ok {
			log().Debugf("configuring docker network: name=%s", n)

			// FIXME: for some reason the request from dockerclient
			// is not returning a populated Networks object
			// so we hack this by inspecting the Network
			// we should switch to engine-api/client -- hopefully
			// that will fix
			//net, found := cInfo.NetworkSettings.Networks[n]
			//if !found {
			//	log().Errorf("container %s is not connected to the network %s", cInfo.Id, n)
			//	continue
			//}

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
				log().Warnf("%s: no ports exposed", cntId)
				continue
			}

			addr, err = utils.BackendAddress(cInfo, p.cfg.BackendOverrideAddress)
			if err != nil {
				log().Error(err)
				continue
			}
		}

		protocol := utils.Protocol(cInfo.Config);

		container_name := cInfo.Name[1:]
		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: healthCheckInterval,
			Protocol:      protocol,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(cInfo.Config)

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
