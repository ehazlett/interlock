package stats

import (
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/interlock"
	"github.com/ehazlett/interlock/plugins"
	"github.com/samalba/dockerclient"
)

const (
	defaultImageNameRegex = ".*"
)

var (
	errorChan = make(chan error)
)

type StatsPlugin struct {
	interlockConfig *interlock.Config
	pluginConfig    *PluginConfig
	client          *dockerclient.DockerClient
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

func loadPluginConfig() (*PluginConfig, error) {
	defaultImageNameFilter := regexp.MustCompile(defaultImageNameRegex)

	cfg := &PluginConfig{
		CarbonAddress:   "",
		StatsPrefix:     "docker.stats",
		ImageNameFilter: defaultImageNameFilter,
		Interval:        10,
	}

	// load custom config via environment
	carbonAddress := os.Getenv("STATS_CARBON_ADDRESS")
	if carbonAddress != "" {
		cfg.CarbonAddress = carbonAddress
	}

	statsPrefix := os.Getenv("STATS_PREFIX")
	if statsPrefix != "" {
		cfg.StatsPrefix = statsPrefix
	}

	imageNameFilter := os.Getenv("STATS_IMAGE_NAME_FILTER")
	if imageNameFilter != "" {
		// validate regex
		r, err := regexp.Compile(imageNameFilter)
		if err != nil {
			return nil, err
		}
		cfg.ImageNameFilter = r
	}

	interval := os.Getenv("STATS_INTERVAL")
	if interval != "" {
		i, err := strconv.Atoi(interval)
		if err != nil {
			return nil, err
		}
		cfg.Interval = i
	}

	return cfg, nil
}

func NewPlugin(interlockConfig *interlock.Config, client *dockerclient.DockerClient) (interlock.Plugin, error) {
	p := StatsPlugin{interlockConfig: interlockConfig, client: client}
	cfg, err := loadPluginConfig()
	if err != nil {
		return nil, err
	}
	p.pluginConfig = cfg

	// handle errorChan
	go func() {
		for {
			err := <-errorChan
			plugins.Log(pluginInfo.Name,
				log.ErrorLevel,
				err.Error(),
			)
		}
	}()

	plugins.Log(pluginInfo.Name, log.InfoLevel, fmt.Sprintf("sending stats every %d seconds", cfg.Interval))

	return p, nil
}

func (p StatsPlugin) initialize() error {
	containers, err := p.client.ListContainers(false, false, "")
	if err != nil {
		return err
	}

	for _, c := range containers {
		if err := p.startStats(c.Id); err != nil {
			errorChan <- err
		}
	}

	return nil
}

func (p StatsPlugin) handleStats(id string, cb dockerclient.StatCallback, ec chan error, args ...interface{}) {
	go p.client.StartMonitorStats(id, cb, ec, args)
}

func (p StatsPlugin) Info() *interlock.PluginInfo {
	return pluginInfo
}

func (p StatsPlugin) sendStat(path string, value interface{}, t *time.Time) error {
	conn, err := net.Dial("tcp", p.pluginConfig.CarbonAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	timestamp := t.Unix()
	v := fmt.Sprintf("%s %v %d", path, value, timestamp)
	plugins.Log(pluginInfo.Name, log.DebugLevel,
		fmt.Sprintf("sending to carbon: %v", v),
	)
	fmt.Fprintf(conn, "%s\n", v)

	return nil
}

func (p StatsPlugin) sendEventStats(id string, stats *dockerclient.Stats, ec chan error, args ...interface{}) {
	timestamp := time.Now()
	// report every n seconds
	rem := math.Mod(float64(timestamp.Second()), float64(p.pluginConfig.Interval))
	if rem != 0 {
		return
	}

	if len(id) > 12 {
		id = id[:12]
	}
	cInfo, err := p.client.InspectContainer(id)
	if err != nil {
		ec <- err
		return
	}
	cName := cInfo.Name[1:]
	cNamePath := fmt.Sprintf(cName)

	if cInfo.Node.ID != "" {
		cNamePath = fmt.Sprintf("nodes.%s.%s", cInfo.Node.Name, cName)
	}

	statBasePath := p.pluginConfig.StatsPrefix + ".containers." + cNamePath
	type containerStat struct {
		Key   string
		Value interface{}
	}

	memPercent := float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0

	statData := []containerStat{
		{
			Key:   "cpu.totalUsage",
			Value: stats.CpuStats.CpuUsage.TotalUsage,
		},
		{
			Key:   "cpu.usageInKernelmode",
			Value: stats.CpuStats.CpuUsage.UsageInKernelmode,
		},
		{
			Key:   "cpu.usageInUsermode",
			Value: stats.CpuStats.CpuUsage.UsageInUsermode,
		},
		{
			Key:   "memory.usage",
			Value: stats.MemoryStats.Usage,
		},
		{
			Key:   "memory.maxUsage",
			Value: stats.MemoryStats.MaxUsage,
		},
		{
			Key:   "memory.failcnt",
			Value: stats.MemoryStats.Failcnt,
		},
		{
			Key:   "memory.limit",
			Value: stats.MemoryStats.Limit,
		},
		{
			Key:   "memory.percent",
			Value: memPercent,
		},
		{
			Key:   "network.rxBytes",
			Value: stats.Network.RxBytes,
		},
		{
			Key:   "network.rxPackets",
			Value: stats.Network.RxPackets,
		},
		{
			Key:   "network.rxErrors",
			Value: stats.Network.RxErrors,
		},
		{
			Key:   "network.rxDropped",
			Value: stats.Network.RxDropped,
		},
		{
			Key:   "network.txBytes",
			Value: stats.Network.TxBytes,
		},
		{
			Key:   "network.txPackets",
			Value: stats.Network.TxPackets,
		},
		{
			Key:   "network.txErrors",
			Value: stats.Network.TxErrors,
		},
		{
			Key:   "network.txDropped",
			Value: stats.Network.TxDropped,
		},
	}

	// send every n seconds
	for _, s := range statData {
		plugins.Log(pluginInfo.Name,
			log.DebugLevel,
			fmt.Sprintf("stat t=%d id=%s key=%s value=%v",
				timestamp.UnixNano(),
				id,
				s.Key,
				s.Value,
			),
		)
		m := fmt.Sprintf("%s.%s", statBasePath, s.Key)
		if err := p.sendStat(m, s.Value, &timestamp); err != nil {
			ec <- err
		}
	}

	return
}

func (p StatsPlugin) startStats(id string) error {
	// get container info for event
	c, err := p.client.InspectContainer(id)
	if err != nil {
		return err
	}
	// match regex to start monitoring
	if p.pluginConfig.ImageNameFilter.MatchString(c.Config.Image) {
		plugins.Log(pluginInfo.Name, log.DebugLevel,
			fmt.Sprintf("gathering stats: image=%s id=%s", c.Image, c.Id[:12]))
		go p.handleStats(id, p.sendEventStats, errorChan, nil)
	}

	return nil
}

func (p StatsPlugin) HandleEvent(event *dockerclient.Event) error {
	// check all containers to see if stats are needed
	if err := p.initialize(); err != nil {
		return err
	}

	t := time.Now()
	if err := p.sendStat(p.pluginConfig.StatsPrefix+".cluster.events", 1, &t); err != nil {
		plugins.Log(pluginInfo.Name, log.ErrorLevel, err.Error())
	}

	if event.Status == "start" {
		if err := p.startStats(event.Id); err != nil {
			return err
		}
	}
	return nil
}
