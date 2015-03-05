package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/samalba/dockerclient"
)

const (
	haproxyTmpl = `# managed by interlock
global
    {{ if .Config.SyslogAddr }}log {{ .Config.SyslogAddr }} local0
    log-send-hostname{{ end }}
    maxconn {{ .Config.MaxConn }}
    pidfile {{ .Config.PidPath }}

defaults
    mode http
    retries 3
    option redispatch
    option httplog
    option dontlognull
    option http-server-close
    option forwardfor
    timeout connect {{ .Config.ConnectTimeout }}
    timeout client {{ .Config.ClientTimeout }}
    timeout server {{ .Config.ServerTimeout }}

frontend http-default
    bind *:{{ .Config.Port }}
    monitor-uri /haproxy?monitor
    {{ if .Config.StatsUser }}stats realm Stats
    stats auth {{ .Config.StatsUser }}:{{ .Config.StatsPassword }}{{ end }}
    stats enable
    stats uri /haproxy?stats
    stats refresh 5s
    reqadd X-Forwarded-Proto:\ http
    {{ range $host := .Hosts }}acl is_{{ $host.Name }} hdr_beg(host) {{ $host.Domain }}
    use_backend {{ $host.Name }} if is_{{ $host.Name }}
    {{ end }}
{{ if .Config.SSLCert }}
frontend https-default
    bind *:{{ .Config.SSLPort }} ssl crt {{ .Config.SSLCert }} {{ .Config.SSLOpts }}
    reqadd X-Forwarded-Proto:\ https
    {{ range $host := .Hosts }}acl is_{{ $host.Name }} hdr_beg(host) {{ $host.Domain }}
    use_backend {{ $host.Name }} if is_{{ $host.Name }}
    {{ end }}
{{ end }}
{{ range $host := .Hosts }}backend {{ $host.Name }}
    http-response add-header X-Request-Start %Ts.%ms
    balance roundrobin
    {{ range $option := $host.BackendOptions }}option {{ $option }}
    {{ end }}
    {{ if $host.Check }}option {{ $host.Check }}{{ end }}
    {{ if $host.SSLOnly }}redirect scheme https if !{ ssl_fc  }{{ end }}
    {{ range $i,$up := $host.Upstreams }}server {{ $host.Name }}_{{ $i }} {{ $up.Addr }} check inter {{ $up.CheckInterval }}
    {{ end }}
{{ end }}`
)

var (
	eventsErrChan = make(chan error)
)

type (
	Manager struct {
		mux       sync.Mutex
		config    *interlock.Config
		tlsConfig *tls.Config
		proxyCmd  *exec.Cmd
		client    *dockerclient.DockerClient
	}
)

func NewManager(cfg *interlock.Config, tlsConfig *tls.Config) *Manager {
	m := &Manager{
		config:    cfg,
		tlsConfig: tlsConfig,
	}

	return m
}

func (m *Manager) writeConfig(config *interlock.ProxyConfig) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	f, err := os.OpenFile(m.config.ProxyConfigPath, os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		ff, fErr := os.Create(m.config.ProxyConfigPath)
		defer ff.Close()
		if fErr != nil {
			return fErr
		}
		f = ff
	}
	defer f.Close()
	t := template.New("haproxy")
	tmpl, err := t.Parse(haproxyTmpl)
	if err != nil {
		return err
	}
	var c bytes.Buffer
	if err := tmpl.Execute(&c, config); err != nil {
		return err
	}
	_, fErr := f.Write(c.Bytes())
	if fErr != nil {
		return fErr
	}
	f.Sync()
	return nil
}

func (m *Manager) GenerateProxyConfig(isKillEvent bool) (*interlock.ProxyConfig, error) {
	containers, err := m.client.ListContainers(false, false, "")
	if err != nil {
		return nil, err
	}
	var hosts []*interlock.Host
	proxyUpstreams := map[string][]*interlock.Upstream{}
	hostChecks := map[string]string{}
	hostBackendOptions := map[string][]string{}
	hostSSLOnly := map[string]bool{}
	for _, cnt := range containers {
		cntId := cnt.Id[:12]
		// load interlock data
		cInfo, err := m.client.InspectContainer(cntId)
		if err != nil {
			return nil, err
		}
		env := cInfo.Config.Env
		interlockData := &interlock.InterlockData{}
		for _, e := range env {

			envParts := strings.Split(e, "=")
			if envParts[0] == "INTERLOCK_DATA" {
				b := bytes.NewBufferString(envParts[1])
				if err := json.NewDecoder(b).Decode(&interlockData); err != nil {
					log.Warnf("%s: unable to parse interlock data: %s", cntId, err)
				}
				break
			}
		}
		hostname := cInfo.Config.Hostname
		domain := cInfo.Config.Domainname
		if interlockData.Hostname != "" {
			hostname = interlockData.Hostname
		}
		if interlockData.Domain != "" {
			domain = interlockData.Domain
		}
		if domain == "" {
			continue
		}
		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
		}
		if interlockData.Check != "" {
			if val, ok := hostChecks[domain]; ok {
				// check existing host check for different values
				if val != interlockData.Check {
					log.Warnf("conflicting check specified for %s", domain)
				}
			} else {
				hostChecks[domain] = interlockData.Check
				log.Infof("using custom check for %s: %s", domain, interlockData.Check)
			}
		}
		checkInterval := 5000
		if interlockData.CheckInterval != 0 {
			checkInterval = interlockData.CheckInterval
			log.Infof("using custom check interval for %s: %d", domain, checkInterval)
		}
		if len(interlockData.BackendOptions) > 0 {
			hostBackendOptions[domain] = interlockData.BackendOptions
			log.Infof("using backend options for %s: %s", domain, strings.Join(interlockData.BackendOptions, ","))
		}
		hostSSLOnly[domain] = false
		if interlockData.SSLOnly {
			log.Infof("configuring ssl redirect for %s", domain)
			hostSSLOnly[domain] = true
		}

		//host := cInfo.NetworkSettings.IpAddress
		ports := cInfo.NetworkSettings.Ports
		if len(ports) == 0 {
			log.Warnf("%s: no ports exposed", cntId)
			continue
		}
		var portDef dockerclient.PortBinding

		for _, v := range ports {
			portDef = dockerclient.PortBinding{
				HostIp:   v[0].HostIp,
				HostPort: v[0].HostPort,
			}
			break
		}

		if m.config.ProxyBackendOverrideAddress != "" {
			portDef.HostIp = m.config.ProxyBackendOverrideAddress
		}

		addr := fmt.Sprintf("%s:%s", portDef.HostIp, portDef.HostPort)

		if interlockData.Port != 0 {
			for k, v := range ports {
				parts := strings.Split(k, "/")
				if parts[0] == string(interlockData.Port) {
					port := v[0]
					log.Infof("using port %s", port.HostPort)
					addr = fmt.Sprintf("%s:%s", port.HostIp, port.HostPort)
					break
				}
			}
		}

		up := &interlock.Upstream{
			Addr:          addr,
			CheckInterval: checkInterval,
		}
		for _, alias := range interlockData.AliasDomains {
			log.Infof("adding alias %s for %s", alias, cntId)
			proxyUpstreams[alias] = append(proxyUpstreams[alias], up)
		}
		proxyUpstreams[domain] = append(proxyUpstreams[domain], up)
		if !isKillEvent && interlockData.Warm {
			log.Infof("warming %s: %s", cntId, addr)
			http.Get(fmt.Sprintf("http://%s", addr))
		}

	}
	for k, v := range proxyUpstreams {
		name := strings.Replace(k, ".", "_", -1)
		host := &interlock.Host{
			Name:           name,
			Domain:         k,
			Upstreams:      v,
			Check:          hostChecks[k],
			BackendOptions: hostBackendOptions[k],
			SSLOnly:        hostSSLOnly[k],
		}
		log.Infof("adding host name=%s domain=%s", host.Name, host.Domain)
		hosts = append(hosts, host)
	}
	// generate config
	cfg := &interlock.ProxyConfig{
		Hosts:  hosts,
		Config: m.config,
	}
	return cfg, nil
}

func (m *Manager) UpdateConfig(e *dockerclient.Event) error {
	isKillEvent := false
	if e != nil && e.Status == "kill" {
		isKillEvent = true
	}
	cfg, err := m.GenerateProxyConfig(isKillEvent)
	if err != nil {
		return err
	}
	if err := m.writeConfig(cfg); err != nil {
		return err
	}
	return nil
}

func (m *Manager) getProxyPid() (int, error) {
	f, err := ioutil.ReadFile(m.config.PidPath)
	if err != nil {
		return -1, err
	}
	buf := bytes.NewBuffer(f)
	p := buf.String()
	p = strings.TrimSpace(p)
	pid, err := strconv.Atoi(p)
	if err != nil {
		return -1, err
	}
	return pid, nil
}

func (m *Manager) Reload() error {
	args := []string{"-D", "-f", m.config.ProxyConfigPath, "-p", m.config.PidPath, "-sf"}
	if m.proxyCmd != nil {
		p, err := m.getProxyPid()
		if err != nil {
			log.Error(err)
		}
		pid := strconv.Itoa(p)
		args = append(args, pid)
	}
	cmd := exec.Command("haproxy", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	m.proxyCmd = cmd
	log.Info("proxy reloaded and ready")
	return nil
}

func (m *Manager) connect() error {
	log.Debugf("connecting to swarm on %s", m.config.SwarmUrl)
	c, err := dockerclient.NewDockerClient(m.config.SwarmUrl, m.tlsConfig)
	if err != nil {
		log.Warn(err)
		return err
	}
	m.client = c
	go m.startEventListener()
	go m.reconnectOnFail()
	return nil
}

func (m *Manager) startEventListener() {
	evt := NewEventHandler(m)
	m.client.StartMonitorEvents(evt.Handle, eventsErrChan)
}

func waitForTCP(addr string) error {
	log.Debugf("waiting for swarm to become available on %s", addr)
	for {
		conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
		if err != nil {
			continue
		}
		conn.Close()
		break
	}
	return nil
}

func (m *Manager) reconnectOnFail() {
	<-eventsErrChan
	for {
		log.Warnf("error receiving events; attempting to reconnect")
		u, err := url.Parse(m.config.SwarmUrl)
		if err != nil {
			log.Warnf("unable to parse Swarm URL: %s", err)
			continue
		}

		if err := waitForTCP(u.Host); err != nil {
			log.Warnf("error connecting to Swarm: %s", err)
			continue
		}

		if err := m.connect(); err == nil {
			log.Debugf("re-connected to Swarm: %s", u.Host)
			// force a reload to make sure we keep things in sync
			if err := m.Reload(); err != nil {
				log.Warnf("error during reload: %s", err)
			}
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func (m *Manager) Run() error {
	if err := m.connect(); err != nil {
		return err
	}

	if err := m.UpdateConfig(nil); err != nil {
		return err
	}

	if err := m.Reload(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		<-ch
		if m.proxyCmd != nil {
			pid, err := m.getProxyPid()
			if err != nil {
				log.Fatal(err)
			}
			syscall.Kill(pid, syscall.SIGTERM)
		}
		os.Exit(1)
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}
