package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
    {{ if .Config.StatsUser }}stats realm Stats
    stats auth {{ .Config.StatsUser }}:{{ .Config.StatsPassword }}
    stats enable
    stats uri /haproxy?stats{{ end }}
    {{ range $host := .Hosts }}acl is_{{ $host.Name }} hdr_end(host) -i {{ $host.Domain }}
    use_backend {{ $host.Name }} if is_{{ $host.Name }}
    {{ end }}
{{ range $host := .Hosts }}backend {{ $host.Name }}
    balance roundrobin
    option httpclose
    option forwardfor
    {{ range $i,$up := $host.Upstreams }}server {{$host.Name}}_{{$i}} {{$up.Addr}} check
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
		if err := e.Engine.Connect(nil); err != nil {
			return nil, err
		}
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
	c, err := cluster.New(scheduler.NewResourceManager(), m.engines...)
	if err != nil {
		return err
	}
	m.cluster = c
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

func (m *Manager) GenerateProxyConfig() (*interlock.ProxyConfig, error) {
	containers, err := m.cluster.ListContainers()
	if err != nil {
		return nil, err
	}
	var hosts []*interlock.Host
	proxyUpstreams := map[string][]*interlock.Upstream{}
	for _, cnt := range containers {
		if cnt.Image.Domainname == "" {
			continue
		}
		cntId := cnt.ID[:12]
		hostname := cnt.Image.Hostname
		domain := cnt.Image.Domainname
		if hostname != domain && hostname != "" {
			domain = fmt.Sprintf("%s.%s", hostname, domain)
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
		up := &interlock.Upstream{
			Addr: addr,
		}
		proxyUpstreams[domain] = append(proxyUpstreams[domain], up)
	}
	for k, v := range proxyUpstreams {
		name := strings.Replace(k, ".", "_", -1)
		host := &interlock.Host{
			Name:      name,
			Domain:    k,
			Upstreams: v,
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

func (m *Manager) UpdateConfig() error {
	cfg, err := m.GenerateProxyConfig()
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
	//cmd := exec.Command("haproxy", "-f", m.config.ProxyConfigPath, "-p", "/var/run/haproxy.pid", pidKill)
	cmd := exec.Command("haproxy", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	// kill old process
	//if m.proxyCmd != nil {
	//	syscall.Kill(m.proxyCmd.Process.Pid, syscall.SIGKILL)
	//}
	m.proxyCmd = cmd
	logger.Info("reloaded proxy")
	return nil
}

func (m *Manager) Run() error {
	if err := m.UpdateConfig(); err != nil {
		return err
	}
	if err := m.cluster.Events(&EventHandler{Manager: m}); err != nil {
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
