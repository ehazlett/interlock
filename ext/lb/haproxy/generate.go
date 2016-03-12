package haproxy

import (
	"fmt"
	"strings"

	"github.com/ehazlett/interlock/ext/lb/utils"
	"github.com/samalba/dockerclient"
)

func (p *HAProxyLoadBalancer) GenerateProxyConfig(containers []dockerclient.Container) (interface{}, error) {
	var hosts []*Host

	proxyUpstreams := map[string][]*Upstream{}
	hostChecks := map[string]string{}
	hostBalanceAlgorithms := map[string]string{}
	hostBackendOptions := map[string][]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostSSLBackendTLSVerify := map[string]string{}

	networks := map[string]string{}

	// TODO: instead of setting defaults here use
	// SetDefaultConfig in the utils package
	for _, cnt := range containers {
		cntId := cnt.Id[:12]
		// load interlock data
		cInfo, err := p.client.InspectContainer(cntId)
		if err != nil {
			return nil, err
		}

		hostname := utils.Hostname(cInfo.Config)
		domain := utils.Domain(cInfo.Config)

		if domain == "" {
			continue
		}

		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
		}

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

		container_name := cInfo.Name[1:]
		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: healthCheckInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

		// "parse" multiple labels for alias domains
		aliasDomains := utils.AliasDomains(cInfo.Config)

		log().Debugf("alias domains: %v", aliasDomains)

		for _, alias := range aliasDomains {
			log().Debugf("adding alias %s for %s", alias, cntId)
			proxyUpstreams[alias] = append(proxyUpstreams[alias], up)
		}

		proxyUpstreams[domain] = append(proxyUpstreams[domain], up)
	}

	for k, v := range proxyUpstreams {
		name := strings.Replace(k, ".", "_", -1)
		host := &Host{
			Name:                name,
			Domain:              k,
			Upstreams:           v,
			Check:               hostChecks[k],
			BalanceAlgorithm:    hostBalanceAlgorithms[k],
			BackendOptions:      hostBackendOptions[k],
			SSLOnly:             hostSSLOnly[k],
			SSLBackend:          hostSSLBackend[k],
			SSLBackendTLSVerify: hostSSLBackendTLSVerify[k],
		}
		log().Debugf("adding host name=%s domain=%s", host.Name, host.Domain)
		hosts = append(hosts, host)
	}

	cfg := &Config{
		Hosts:    hosts,
		Config:   p.cfg,
		Networks: networks,
	}

	return cfg, nil
}
