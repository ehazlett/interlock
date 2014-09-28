package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	"github.com/citadel/citadel"
	"github.com/citadel/citadel/cluster"
	"github.com/citadel/citadel/scheduler"
	"github.com/ehazlett/interlock"
	"github.com/shipyard/shipyard/client"
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
    {{ range $host := .Hosts }}acl is_{{ $host.Name }} hdr_beg(host) {{ $host.Domain }}
    use_backend {{ $host.Name }} if is_{{ $host.Name }}
    {{ end }}
{{ range $host := .Hosts }}backend {{ $host.Name }}
    http-response add-header X-Request-Start %Ts.%ms
    balance roundrobin
    option forwardfor
    {{ range $option := $host.BackendOptions }}option {{ $option }}
    {{ end }}
    {{ if $host.Check }}option {{ $host.Check }}{{ end }}
    {{ range $i,$up := $host.Upstreams }}server {{ $host.Name }}_{{ $i }} {{ $up.Addr }} check inter {{ $up.CheckInterval }}
    {{ end }}
{{ end }}`
)

type (
	Manager struct {
		mux      sync.Mutex
		config   *interlock.Config
		engines  []*citadel.Engine
		cluster  *cluster.Cluster
		proxyCmd *exec.Cmd
	}
)

func NewManager(cfg *interlock.Config) (*Manager, error) {
	engines := []*citadel.Engine{}
	for _, e := range cfg.InterlockEngines {
		engines = append(engines, e.Engine)
	}
	m := &Manager{
		config:  cfg,
		engines: engines,
	}
	if err := m.init(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) init() error {
	var engines []*citadel.Engine
	if m.config.ShipyardUrl != "" {
		cfg := &client.ShipyardConfig{
			Url:        m.config.ShipyardUrl,
			ServiceKey: m.config.ShipyardServiceKey,
		}
		mgr := client.NewManager(cfg)
		eng, err := mgr.Engines()
		if err != nil {
			return err
		}
		for _, e := range eng {
			engines = append(engines, e.Engine)
		}
	} else {
		engines = m.engines
	}
	for _, e := range engines {
		if err := e.Connect(nil); err != nil {
			return err
		}
		logger.Infof("loaded engine: %s", e.ID)
	}
	c, err := cluster.New(scheduler.NewResourceManager(), engines...)
	if err != nil {
		return err
	}
	m.cluster = c
	// register handler
	if err := m.cluster.Events(&EventHandler{Manager: m}); err != nil {
		return err
	}
	return nil
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
	containers, err := m.cluster.ListContainers(false)
	if err != nil {
		return nil, err
	}
	var hosts []*interlock.Host
	proxyUpstreams := map[string][]*interlock.Upstream{}
	hostChecks := map[string]string{}
	hostBackendOptions := map[string][]string{}
	for _, cnt := range containers {
		cntId := cnt.ID[:12]
		// load interlock data
		env := cnt.Image.Environment
		interlockData := &interlock.InterlockData{}
		if key, ok := env["INTERLOCK_DATA"]; ok {
			b := bytes.NewBufferString(key)
			if err := json.NewDecoder(b).Decode(&interlockData); err != nil {
				logger.Warnf("%s: unable to parse interlock data: %s", cntId, err)
			}
		}
		hostname := cnt.Image.Hostname
		domain := cnt.Image.Domainname
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
					logger.Warnf("conflicting check specified for %s", domain)
				}
			} else {
				hostChecks[domain] = interlockData.Check
				logger.Infof("using custom check for %s: %s", domain, interlockData.Check)
			}
		}
		checkInterval := 5000
		if interlockData.CheckInterval != 0 {
			checkInterval = interlockData.CheckInterval
			logger.Infof("using custom check interval for %s: %d", domain, checkInterval)
		}
		if len(interlockData.BackendOptions) > 0 {
			hostBackendOptions[domain] = interlockData.BackendOptions
			logger.Infof("using backend options for %s: %s", domain, strings.Join(interlockData.BackendOptions, ","))
		}
		hostAddrUrl, err := url.Parse(cnt.Engine.Addr)
		if err != nil {
			logger.Warnf("%s: unable to parse engine addr: %s", cntId, err)
			continue
		}
		host := hostAddrUrl.Host
		hostParts := strings.Split(hostAddrUrl.Host, ":")
		if len(hostParts) != 1 {
			host = hostParts[0]
		}
		if len(cnt.Ports) == 0 {
			logger.Warnf("%s: no ports exposed", cntId)
			continue
		}
		portDef := cnt.Ports[0]
		addr := fmt.Sprintf("%s:%d", host, portDef.Port)
		if interlockData.Port != 0 {
			for _, p := range cnt.Ports {
				if p.ContainerPort == interlockData.Port {
					addr = fmt.Sprintf("%s:%d", host, p.Port)
				}
			}
		}
		up := &interlock.Upstream{
			Addr:          addr,
			CheckInterval: checkInterval,
		}
		for _, alias := range interlockData.AliasDomains {
			logger.Infof("adding alias %s for %s", alias, cntId)
			proxyUpstreams[alias] = append(proxyUpstreams[alias], up)
		}
		proxyUpstreams[domain] = append(proxyUpstreams[domain], up)
		if !isKillEvent && interlockData.Warm {
			logger.Infof("warming %s: %s", cntId, addr)
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
		}
		logger.Infof("adding host name=%s domain=%s", host.Name, host.Domain)
		hosts = append(hosts, host)
	}
	// generate config
	cfg := &interlock.ProxyConfig{
		Hosts:  hosts,
		Config: m.config,
	}
	return cfg, nil
}

func (m *Manager) UpdateConfig(e *citadel.Event) error {
	isKillEvent := false
	if e != nil && e.Type == "kill" {
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
			logger.Error(err)
		}
		pid := strconv.Itoa(p)
		args = append(args, pid)
	}
	cmd := exec.Command("haproxy", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	m.proxyCmd = cmd
	logger.Info("reloaded proxy")
	return nil
}

func (m *Manager) Run() error {
	if err := m.UpdateConfig(nil); err != nil {
		return err
	}
	m.Reload()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		<-ch
		if m.proxyCmd != nil {
			pid, err := m.getProxyPid()
			if err != nil {
				logger.Fatal(err)
			}
			syscall.Kill(pid, syscall.SIGTERM)
		}
		os.Exit(1)
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}
