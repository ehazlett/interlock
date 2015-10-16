package nginx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

var (
	eventsErrChan = make(chan error)
	proxyCmd      *exec.Cmd
	jobs          = 0
)

type NginxPlugin struct {
	interlockConfig *interlock.Config
	pluginConfig    *PluginConfig
	client          *dockerclient.DockerClient
	mux             sync.Mutex
}

func init() {
	plugins.Register(
		pluginInfo.Name,
		&plugins.RegisteredPlugin{
			New: NewPlugin,
			Info: func() *interlock.PluginInfo {
				return pluginInfo
			},
		})
}

func logMessage(level log.Level, args ...string) {
	plugins.Log(pluginInfo.Name, level, args...)
}

func loadPluginConfig() (*PluginConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// TODO: finish options in config template
	cfg := &PluginConfig{
		ProxyConfigPath:             "/etc/nginx/nginx.conf",
		ProxyBackendOverrideAddress: "",
		Port:                80,
		PidPath:             filepath.Join(wd, "nginx.pid"),
		MaxConnections:      1024,
		MaxProcesses:        2,
		ProxyConnectTimeout: 600,
		ProxySendTimeout:    600,
		ProxyReadTimeout:    600,
		SendTimeout:         600,
		User:                "www-data",
		RLimitNoFile:        65535,
		SSLCertDir:          "/etc/nginx/ssl",
		SSLPort:             443,
		SSLCiphers:          "HIGH:!aNULL:!MD5",
		SSLProtocols:        "SSLv3 TLSv1 TLSv1.1 TLSv1.2",
	}

	// load custom config via environment
	proxyConfigPath := os.Getenv("NGINX_PROXY_CONFIG_PATH")
	if proxyConfigPath != "" {
		cfg.ProxyConfigPath = proxyConfigPath
	}

	proxyBackendOverrideAddress := os.Getenv("NGINX_PROXY_BACKEND_OVERRIDE_ADDRESS")
	if proxyBackendOverrideAddress != "" {
		cfg.ProxyBackendOverrideAddress = proxyBackendOverrideAddress
	}

	port := os.Getenv("NGINX_PORT")
	if port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		cfg.Port = p
	}

	proxyPidPath := os.Getenv("NGINX_PID_PATH")
	if proxyPidPath != "" {
		cfg.PidPath = proxyPidPath
	}

	maxConn := os.Getenv("NGINX_MAX_CONN")
	if maxConn != "" {
		c, err := strconv.Atoi(maxConn)
		if err != nil {
			return nil, err
		}
		cfg.MaxConnections = c
	}

	maxProcesses := os.Getenv("NGINX_MAX_PROCESSES")
	if maxProcesses != "" {
		c, err := strconv.Atoi(maxProcesses)
		if err != nil {
			return nil, err
		}
		cfg.MaxProcesses = c
	}

	rlimitNoFile := os.Getenv("NGINX_RLIMIT_NOFILE")
	if rlimitNoFile != "" {
		v, err := strconv.Atoi(rlimitNoFile)
		if err != nil {
			return nil, err
		}
		cfg.RLimitNoFile = v
	}

	proxyConnectTimeout := os.Getenv("NGINX_PROXY_CONNECT_TIMEOUT")
	if proxyConnectTimeout != "" {
		c, err := strconv.Atoi(proxyConnectTimeout)
		if err != nil {
			return nil, err
		}
		cfg.ProxyConnectTimeout = c
	}

	proxyReadTimeout := os.Getenv("NGINX_PROXY_READ_TIMEOUT")
	if proxyReadTimeout != "" {
		c, err := strconv.Atoi(proxyReadTimeout)
		if err != nil {
			return nil, err
		}
		cfg.ProxyReadTimeout = c
	}

	proxySendTimeout := os.Getenv("NGINX_PROXY_SEND_TIMEOUT")
	if proxySendTimeout != "" {
		c, err := strconv.Atoi(proxySendTimeout)
		if err != nil {
			return nil, err
		}
		cfg.ProxySendTimeout = c
	}

	sendTimeout := os.Getenv("NGINX_SEND_TIMEOUT")
	if sendTimeout != "" {
		c, err := strconv.Atoi(sendTimeout)
		if err != nil {
			return nil, err
		}
		cfg.SendTimeout = c
	}

	sslPort := os.Getenv("NGINX_SSL_PORT")
	if sslPort != "" {
		p, err := strconv.Atoi(sslPort)
		if err != nil {
			return nil, err
		}
		cfg.SSLPort = p
	}

	sslCertDir := os.Getenv("NGINX_SSL_CERT_DIR")
	if sslCertDir != "" {
		cfg.SSLCertDir = sslCertDir
	}

	sslCiphers := os.Getenv("NGINX_SSL_CIPHERS")
	if sslCiphers != "" {
		cfg.SSLCiphers = sslCiphers
	}

	sslProtocols := os.Getenv("NGINX_SSL_PROTOCOLS")
	if sslProtocols != "" {
		cfg.SSLProtocols = sslProtocols
	}

	user := os.Getenv("NGINX_USER")
	if user != "" {
		cfg.User = user
	}

	return cfg, nil
}

func NewPlugin(interlockConfig *interlock.Config, client *dockerclient.DockerClient) (interlock.Plugin, error) {
	pluginConfig, err := loadPluginConfig()
	if err != nil {
		return nil, err
	}

	plugin := NginxPlugin{
		pluginConfig:    pluginConfig,
		interlockConfig: interlockConfig,
		client:          client,
	}

	return plugin, nil
}

func (p NginxPlugin) Info() *interlock.PluginInfo {
	return pluginInfo
}

func (p NginxPlugin) HandleEvent(event *dockerclient.Event) error {
	logMessage(log.InfoLevel,
		fmt.Sprintf("action=received event=%s time=%d",
			event.Id,
			event.Time,
		),
	)

	switch event.Status {
	case "start", "interlock-start":
		jobs = jobs + 1
		if err := p.handleUpdate(event); err != nil {
			return err
		}
	case "stop", "kill", "die":
		jobs = jobs + 1
		// delay to make sure container is removed
		time.Sleep(250 * time.Millisecond)
		if err := p.handleUpdate(event); err != nil {
			return err
		}
	case "interlock-stop":
		// stop haproxy
		if proxyCmd != nil {
			pid, err := p.getProxyPid()
			if err != nil {
				return err
			}
			logMessage(log.DebugLevel, fmt.Sprintf("stopping nginx pid=%d", pid))
			syscall.Kill(pid, syscall.SIGTERM)
		}
		// wait for stop
		time.Sleep(1 * time.Second)
		return nil
	}

	return nil
}

func (p NginxPlugin) handleUpdate(event *dockerclient.Event) error {
	defer p.handleReload(event)

	return nil
}

func (p NginxPlugin) Init() error {
	return nil
}

func (p NginxPlugin) handleReload(event *dockerclient.Event) error {
	jobs -= 1

	logMessage(log.DebugLevel, fmt.Sprintf("jobs: %d", jobs))

	if jobs == 0 {
		logMessage(log.DebugLevel, fmt.Sprintf("reload triggered"))
		if err := p.updateConfig(event); err != nil {
			return err
		}

		logMessage(log.DebugLevel, "reloading nginx process")

		if err := p.reload(); err != nil {
			logMessage(log.ErrorLevel, fmt.Sprintf("error reloading: %s", err))
			return err
		}

		time.Sleep(250 * time.Millisecond)
	}

	return nil
}

func (p NginxPlugin) generateNginxConfig() (*NginxConfig, error) {
	containers, err := p.client.ListContainers(false, false, "")
	if err != nil {
		return nil, err
	}

	var hosts []*Host
	upstreamServers := map[string][]string{}
	serverNames := map[string][]string{}
	//hostBalanceAlgorithms := map[string]string{}
	hostSSL := map[string]bool{}
	hostSSLCert := map[string]string{}
	hostSSLCertKey := map[string]string{}
	hostSSLOnly := map[string]bool{}
	hostSSLBackend := map[string]bool{}
	hostWebsocketEndpoints := map[string][]string{}

	for _, c := range containers {
		cntId := c.Id[:12]
		// load interlock data
		cInfo, err := p.client.InspectContainer(cntId)
		if err != nil {
			return nil, err
		}

		env := cInfo.Config.Env
		interlockData := &InterlockData{}

		for _, e := range env {
			envParts := strings.Split(e, "=")
			if envParts[0] == "INTERLOCK_DATA" {
				b := bytes.NewBufferString(envParts[1])
				if err := json.NewDecoder(b).Decode(&interlockData); err != nil {
					logMessage(log.WarnLevel,
						fmt.Sprintf("%s: unable to parse interlock data: %s", cntId, err))
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

		// check if the first server name is there; if not, add
		// this happens if there are multiple backend containers
		if _, ok := serverNames[domain]; !ok {
			serverNames[domain] = []string{domain}
		}

		hostSSL[domain] = interlockData.SSL

		hostSSLOnly[domain] = false
		if interlockData.SSLOnly {
			logMessage(log.DebugLevel,
				fmt.Sprintf("configuring ssl redirect for %s", domain))
			hostSSLOnly[domain] = true
		}

		// check ssl backend
		hostSSLBackend[domain] = false
		if interlockData.SSLBackend {
			logMessage(log.DebugLevel,
				fmt.Sprintf("configuring ssl backend for %s", domain))
			hostSSLBackend[domain] = true
		}

		// set cert paths
		baseCertPath := p.pluginConfig.SSLCertDir
		if interlockData.SSLCert != "" {
			certPath := filepath.Join(baseCertPath, interlockData.SSLCert)
			logMessage(log.InfoLevel,
				fmt.Sprintf("ssl cert for %s: %s", domain, certPath))
			hostSSLCert[domain] = certPath
		}

		if interlockData.SSLCertKey != "" {
			keyPath := filepath.Join(baseCertPath, interlockData.SSLCertKey)
			logMessage(log.InfoLevel,
				fmt.Sprintf("ssl key for %s: %s", domain, keyPath))
			hostSSLCertKey[domain] = keyPath
		}

		ports := cInfo.NetworkSettings.Ports
		if len(ports) == 0 {
			logMessage(log.WarnLevel, fmt.Sprintf("%s: no ports exposed", cntId))
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

		if p.pluginConfig.ProxyBackendOverrideAddress != "" {
			portDef.HostIp = p.pluginConfig.ProxyBackendOverrideAddress
		}

		addr := fmt.Sprintf("%s:%s", portDef.HostIp, portDef.HostPort)

		if interlockData.Port != 0 {
			interlockPort := fmt.Sprintf("%d", interlockData.Port)
			for k, v := range ports {
				parts := strings.Split(k, "/")
				if parts[0] == interlockPort {
					port := v[0]
					logMessage(log.DebugLevel,
						fmt.Sprintf("%s: found specified port %s exposed as %s", domain, interlockPort, port.HostPort))
					addr = fmt.Sprintf("%s:%s", portDef.HostIp, port.HostPort)
					break
				}
			}
		}

		// websocket endpoints
		for _, ws := range interlockData.WebsocketEndpoints {
			hostWebsocketEndpoints[domain] = append(hostWebsocketEndpoints[domain], ws)
		}

		logMessage(log.InfoLevel,
			fmt.Sprintf("%s: upstream=%s", domain, addr))

		for _, alias := range interlockData.AliasDomains {
			logMessage(log.DebugLevel,
				fmt.Sprintf("adding alias %s for %s", alias, cntId))
			serverNames[domain] = append(serverNames[domain], alias)
		}

		upstreamServers[domain] = append(upstreamServers[domain], addr)
	}

	for k, v := range upstreamServers {
		h := &Host{
			ServerNames: serverNames[k],
			// TODO: make configurable for TCP via InterlockData
			Port:               p.pluginConfig.Port,
			SSLPort:            p.pluginConfig.SSLPort,
			SSL:                hostSSL[k],
			SSLCert:            hostSSLCert[k],
			SSLCertKey:         hostSSLCertKey[k],
			SSLOnly:            hostSSLOnly[k],
			SSLBackend:         hostSSLBackend[k],
			WebsocketEndpoints: hostWebsocketEndpoints[k],
		}

		servers := []*Server{}

		for _, s := range v {
			srv := &Server{
				Addr: s,
			}

			servers = append(servers, srv)
		}

		up := &Upstream{
			Name:    k,
			Servers: servers,
		}
		h.Upstream = up

		hosts = append(hosts, h)
	}

	return &NginxConfig{
		*p.pluginConfig,
		hosts,
	}, nil
}

func (p NginxPlugin) updateConfig(e *dockerclient.Event) error {
	cfg, err := p.generateNginxConfig()
	if err != nil {
		return err
	}

	logMessage(log.DebugLevel, "writing config: ", p.pluginConfig.ProxyConfigPath)

	if err := p.writeConfig(cfg); err != nil {
		return err
	}

	return nil
}

func (p NginxPlugin) writeConfig(config *NginxConfig) error {
	p.mux.Lock()
	defer p.mux.Unlock()
	f, err := os.OpenFile(p.pluginConfig.ProxyConfigPath, os.O_WRONLY|os.O_TRUNC, 0664)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		ff, fErr := os.Create(p.pluginConfig.ProxyConfigPath)
		defer ff.Close()
		if fErr != nil {
			return fErr
		}
		f = ff
	}
	defer f.Close()

	t := template.New("nginx")
	tmpl := template.Must(t.Parse(nginxConfTemplate))
	// validate parse

	var c bytes.Buffer
	if err := tmpl.Execute(&c, config); err != nil {
		return err
	}

	if _, err := f.Write(c.Bytes()); err != nil {
		return err
	}

	f.Sync()

	return nil
}
func (p NginxPlugin) getProxyPid() (int, error) {
	f, err := ioutil.ReadFile(p.pluginConfig.PidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, nil
		}

		return -1, err
	}
	buf := bytes.NewBuffer(f)
	pd := buf.String()
	pd = strings.TrimSpace(pd)
	pid, err := strconv.Atoi(pd)
	if err != nil {
		return -1, err
	}
	return pid, nil
}

func (p NginxPlugin) reload() error {
	p.mux.Lock()
	defer p.mux.Unlock()

	proxyPid, err := p.getProxyPid()
	if err != nil {
		log.Error(err)
	}

	args := []string{"-HUP", fmt.Sprintf("%d", proxyPid)}

	pCmdPath, err := exec.LookPath("kill")
	if err != nil {
		return err
	}

	// if initial load, spawn nginx
	if proxyCmd == nil {
		cmdPath, err := exec.LookPath("nginx")
		if err != nil {
			return err
		}
		args = []string{"-c", p.pluginConfig.ProxyConfigPath}
		pCmdPath = cmdPath
	}

	cmd := exec.Command(pCmdPath, args...)
	if err := cmd.Run(); err != nil {
		return err
	}

	proxyCmd = cmd

	time.Sleep(100 * time.Millisecond)
	logMessage(log.InfoLevel, "nginx reloaded and ready")
	return nil
}
