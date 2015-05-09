package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

var (
	errChan = make(chan (error))
)

type ExternalPlugin struct {
	interlockConfig *interlock.Config
	client          *dockerclient.DockerClient
	pluginConfig    *PluginConfig
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
	cfg := &PluginConfig{
		Paths: []string{},
	}

	// load custom config via environment
	extPaths := os.Getenv("EXTERNAL_PATHS")
	if extPaths != "" {
		cfg.Paths = strings.Split(extPaths, ",")
	}

	return cfg, nil
}

func NewPlugin(interlockConfig *interlock.Config, client *dockerclient.DockerClient) (interlock.Plugin, error) {
	pluginConfig, err := loadPluginConfig()
	if err != nil {
		return nil, err
	}

	if len(pluginConfig.Paths) == 0 {
		logMessage(log.ErrorLevel, "no external paths specified")
	}

	return ExternalPlugin{
		interlockConfig: interlockConfig,
		client:          client,
		pluginConfig:    pluginConfig,
	}, nil
}

func (p ExternalPlugin) Info() *interlock.PluginInfo {
	return pluginInfo
}

func (p ExternalPlugin) notify(path string, event *dockerclient.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		errChan <- err
		return err
	}

	buf := bytes.NewBuffer(data)

	payload := buf.String()

	logMessage(log.DebugLevel,
		fmt.Sprintf("notifying external: path=%s data=%s", path, payload),
	)

	cmd := exec.Command(path)
	stdin, err := cmd.StdinPipe()
	if err := cmd.Start(); err != nil {
		errChan <- err
		return err
	}

	stdin.Write(data)
	stdin.Write([]byte("\n"))

	cmd.Wait()

	return nil
}

func (p ExternalPlugin) HandleEvent(event *dockerclient.Event) error {
	for _, path := range p.pluginConfig.Paths {
		go p.notify(path, event)
	}

	return nil
}

func (p ExternalPlugin) Init() error {
	go func() {
		err := <-errChan
		logMessage(log.ErrorLevel, fmt.Sprintf("error running external: %s", err))
	}()

	return nil
}
