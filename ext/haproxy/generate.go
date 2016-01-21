package haproxy

import (
	"fmt"
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
	for _, cnt := range containers {
		cntId := cnt.Id[:12]
		// load interlock data
		cInfo, err := p.client.InspectContainer(cntId)
		if err != nil {
			return nil, err
		}

		hostname := cInfo.Config.Hostname
		domain := cInfo.Config.Domainname

		// TODO: parse labels for custom option overrides

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

		checkInterval := 5000

		hostBalanceAlgorithms[domain] = "roundrobin"

		hostSSLOnly[domain] = false

		// ssl backend
		hostSSLBackend[domain] = false

		//host := cInfo.NetworkSettings.IpAddress
		ports := cInfo.NetworkSettings.Ports
		if len(ports) == 0 {
			log().Warn(fmt.Sprintf("%s: no ports exposed", cntId))
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

		container_name := cInfo.Name[1:]
		up := &Upstream{
			Addr:          addr,
			Container:     container_name,
			CheckInterval: checkInterval,
		}

		log().Infof("%s: upstream=%s container=%s", domain, addr, container_name)

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
