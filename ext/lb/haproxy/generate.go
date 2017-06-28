package haproxy

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
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

	for _, c := range containers {
		cntId := c.ID[:12]
		// load interlock data
		hostname := utils.Hostname(c)
		domain := utils.Domain(c)

		// context root
		contextRoot := utils.ContextRoot(c)
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
		hostContextRootRewrites[domain] = utils.ContextRootRewrite(c)

		healthCheck := utils.HealthCheck(c)
		healthCheckInterval, err := utils.HealthCheckInterval(c)
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

		hostBalanceAlgorithms[domain] = utils.BalanceAlgorithm(c)

		backendOptions := utils.BackendOptions(c)

		if len(backendOptions) > 0 {
			hostBackendOptions[domain] = backendOptions
			log().Debugf("using backend options for %s: %s", domain, strings.Join(backendOptions, ","))
		}

		hostSSLOnly[domain] = utils.SSLOnly(c)

		// ssl backend
		hostSSLBackend[domain] = utils.SSLBackend(c)
		hostSSLBackendTLSVerify[domain] = utils.SSLBackendTLSVerify(c)

		addr := ""

		// check for networking
		if n, ok := utils.OverlayEnabled(c); ok {
			log().Debugf("configuring docker network: name=%s", n)

			network, err := p.client.NetworkInspect(context.Background(), n, false)
			if err != nil {
				log().Error(err)
				continue
			}

			addr, err = utils.BackendOverlayAddress(network, c)
			if err != nil {
				log().Error(err)
				continue
			}

			networks[n] = ""
		} else {
			if len(c.Ports) == 0 {
				log().Warnf("%s: no ports exposed", cntId)
				continue
			}

			a, err := utils.BackendAddress(c, p.cfg.BackendOverrideAddress)
			if err != nil {
				log().Error(err)
				continue
			}
			addr = a
		}

		container_name := c.Names[0][1:]
		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: healthCheckInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(c)

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
