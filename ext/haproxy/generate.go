package haproxy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ehazlett/interlock/ext"
	"github.com/samalba/dockerclient"
)

func (p *HAProxyLoadBalancer) GenerateProxyConfig() (*Config, error) {
	log().Debug("generating proxy config")

	containers, err := p.client.ListContainers(false, false, "")
	if err != nil {
		return nil, err
	}

	var hosts []*Host

	proxyUpstreams := map[string][]*Upstream{}
	hostChecks := map[string]string{}
	hostBalanceAlgorithms := map[string]string{}
	hostBackendOptions := map[string][]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostSSLBackendTLSVerify := map[string]string{}

	// TODO: instead of setting defaults here use
	// SetDefaultConfig in the utils package
	for _, cnt := range containers {
		cntId := cnt.Id[:12]
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

		if v, ok := cInfo.Config.Labels[ext.InterlockHealthCheckLabel]; ok {
			if val, ok := hostChecks[domain]; ok {
				// check existing host check for different values
				if val != v {
					log().Warnf("conflicting check specified for %s", domain)
				}
			} else {
				hostChecks[domain] = v
				log().Debugf("using custom check for %s: %s", domain, v)
			}
		}

		checkInterval := 5000

		if v, ok := cInfo.Config.Labels[ext.InterlockHealthCheckIntervalLabel]; ok && v != "" {
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			if i != 0 {
				checkInterval = i
				log().Debugf("using custom check interval for %s: %d", domain, checkInterval)
			}
		}

		hostBalanceAlgorithms[domain] = "roundrobin"

		if v, ok := cInfo.Config.Labels[ext.InterlockBalanceAlgorithmLabel]; ok && v != "" {
			hostBalanceAlgorithms[domain] = v
		}

		backendOptions := []string{}
		for l, v := range cInfo.Config.Labels {
			// this is for labels like interlock.backend_option.1=foo
			if strings.Index(l, ext.InterlockBackendOptionLabel) > -1 {
				backendOptions = append(backendOptions, v)
			}
		}

		if len(backendOptions) > 0 {
			hostBackendOptions[domain] = backendOptions
			log().Debugf("using backend options for %s: %s", domain, strings.Join(backendOptions, ","))
		}

		hostSSLOnly[domain] = false
		if _, ok := cInfo.Config.Labels[ext.InterlockSSLOnlyLabel]; ok {
			log().Debugf("configuring ssl redirect for %s", domain)
			hostSSLOnly[domain] = true
		}

		// ssl backend
		hostSSLBackend[domain] = false
		if _, ok := cInfo.Config.Labels[ext.InterlockSSLBackendLabel]; ok {
			hostSSLBackend[domain] = true

			sslBackendTLSVerify := "none"
			if v, ok := cInfo.Config.Labels[ext.InterlockSSLBackendTLSVerifyLabel]; ok {
				sslBackendTLSVerify = v
			}
			hostSSLBackendTLSVerify[domain] = sslBackendTLSVerify

			log().Debugf("configuring ssl backend for %s verify=%s", domain, sslBackendTLSVerify)
		}

		//host := cInfo.NetworkSettings.IpAddress
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
			for k, x := range ports {
				parts := strings.Split(k, "/")
				if parts[0] == v {
					port := x[0]
					log().Debugf("%s: found specified port %s exposed as %s", domain, v, port.HostPort)
					addr = fmt.Sprintf("%s:%s", portDef.HostIp, port.HostPort)
					break
				}
			}
		}

		container_name := cInfo.Name[1:]
		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: checkInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

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
	// generate config
	cfg := &Config{
		Hosts:  hosts,
		Config: p.cfg,
	}

	return cfg, nil
}
